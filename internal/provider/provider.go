package provider

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/metrics"
	unifi "github.com/lexfrei/go-unifi/api/network"
	"golang.org/x/sync/semaphore"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

const (
	defaultTTL = 300
	// maxConcurrency limits parallel DNS operations to protect UniFi API from overload.
	maxConcurrency = 5
	// operationTimeout is the maximum time allowed for a single DNS operation.
	operationTimeout = 30 * time.Second
)

// UniFiProvider implements the provider.Provider interface for UniFi OS.
type UniFiProvider struct {
	client       unifi.NetworkAPIClient
	site         string
	domainFilter endpoint.DomainFilter
}

// New creates a new UniFiProvider instance with the provided client.
// This constructor accepts an interface to enable dependency injection for testing.
func New(client unifi.NetworkAPIClient, site string, domainFilter endpoint.DomainFilter) *UniFiProvider {
	return &UniFiProvider{
		client:       client,
		site:         site,
		domainFilter: domainFilter,
	}
}

// Records retrieves all DNS records from UniFi that match the domain filter.
func (p *UniFiProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	slog.InfoContext(ctx, "fetching DNS records from UniFi", "site", p.site)

	// Get DNS records from UniFi API using the proper APIClient wrapper
	records, err := p.client.ListDNSRecords(ctx, p.site)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list DNS records from UniFi")
	}

	slog.DebugContext(ctx, "received DNS records from UniFi", "total_count", len(records))

	// Convert UniFi DNS records to endpoints
	// Pre-allocate with capacity to avoid reallocations (most records will match filter)
	endpoints := make([]*endpoint.Endpoint, 0, len(records))

	recordsByType := make(map[string]int)

	for _, record := range records {
		// Skip records that don't match the domain filter
		if !p.domainFilter.Match(record.Key) {
			continue
		}

		endpointRecord := p.unifiToEndpoint(&record)
		if endpointRecord != nil {
			endpoints = append(endpoints, endpointRecord)
			recordsByType[endpointRecord.RecordType]++
		}
	}

	// Update metrics for managed records by type
	for recordType, count := range recordsByType {
		metrics.DNSRecordsManaged.WithLabelValues(recordType).Set(float64(count))
	}

	slog.InfoContext(ctx, "fetched DNS records", "filtered_count", len(endpoints))

	return endpoints, nil
}

// ApplyChanges applies the given changes to UniFi DNS.
func (p *UniFiProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	slog.InfoContext(ctx, "applying DNS changes",
		"create", len(changes.Create),
		"update", len(changes.UpdateNew),
		"delete", len(changes.Delete))

	// Record number of changes
	if len(changes.Delete) > 0 {
		metrics.DNSChangesApplied.WithLabelValues("delete").Observe(float64(len(changes.Delete)))
	}

	if len(changes.UpdateNew) > 0 {
		metrics.DNSChangesApplied.WithLabelValues("update").Observe(float64(len(changes.UpdateNew)))
	}

	if len(changes.Create) > 0 {
		metrics.DNSChangesApplied.WithLabelValues("create").Observe(float64(len(changes.Create)))
	}

	// Handle deletions
	err := p.applyDeletions(ctx, changes.Delete)
	if err != nil {
		return err
	}

	// Handle updates
	err = p.applyUpdates(ctx, changes.UpdateOld, changes.UpdateNew)
	if err != nil {
		return err
	}

	// Handle creations
	err = p.applyCreations(ctx, changes.Create)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "successfully applied DNS changes")

	return nil
}

// AdjustEndpoints adjusts endpoints as needed by the provider.
// For UniFi, we return endpoints as-is.
func (p *UniFiProvider) AdjustEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error) {
	return endpoints, nil
}

// GetDomainFilter returns the domain filter configuration.
//
//nolint:ireturn // Required by external-dns provider interface
func (p *UniFiProvider) GetDomainFilter() endpoint.DomainFilterInterface {
	return &p.domainFilter
}

func (p *UniFiProvider) applyDeletions(ctx context.Context, endpoints []*endpoint.Endpoint) error {
	if len(endpoints) == 0 {
		return nil
	}

	// Build record index ONCE before parallel operations to avoid N API calls
	// This is critical for performance: without this, each goroutine would call
	// ListDNSRecords independently, resulting in N*API_calls instead of 1
	allRecords, err := p.client.ListDNSRecords(ctx, p.site)
	if err != nil {
		return errors.Wrap(err, "failed to list DNS records for deletion")
	}

	recordIndex := buildRecordIndex(allRecords)

	return p.parallelDeleteWithIndex(ctx, endpoints, recordIndex, "delete")
}

func (p *UniFiProvider) applyUpdates(ctx context.Context, oldEndpoints, newEndpoints []*endpoint.Endpoint) error {
	if len(oldEndpoints) == 0 && len(newEndpoints) == 0 {
		return nil
	}

	// Build record index ONCE for all delete operations
	var recordIndex map[string][]unifi.DNSRecord

	if len(oldEndpoints) > 0 {
		allRecords, err := p.client.ListDNSRecords(ctx, p.site)
		if err != nil {
			return errors.Wrap(err, "failed to list DNS records for update")
		}

		recordIndex = buildRecordIndex(allRecords)

		// Delete old records in parallel
		err = p.parallelDeleteWithIndex(ctx, oldEndpoints, recordIndex, "update")
		if err != nil {
			return err
		}
	}

	// Create new records in parallel
	if len(newEndpoints) > 0 {
		err := p.parallelCreate(ctx, newEndpoints, "update")
		if err != nil {
			return err
		}
	}

	return nil
}

// parallelDeleteWithIndex performs parallel deletion using a pre-built record index.
func (p *UniFiProvider) parallelDeleteWithIndex(ctx context.Context, endpoints []*endpoint.Endpoint, recordIndex map[string][]unifi.DNSRecord, operation string) error {
	sem := semaphore.NewWeighted(maxConcurrency)
	errChan := make(chan error, len(endpoints))

	var wg sync.WaitGroup

	for _, endpointToDelete := range endpoints {
		wg.Add(1)

		go func(endpointItem *endpoint.Endpoint) {
			defer wg.Done()

			// Acquire semaphore (blocks if at max concurrency)
			err := sem.Acquire(ctx, 1)
			if err != nil {
				errChan <- errors.Wrap(err, "failed to acquire semaphore")

				return
			}

			defer sem.Release(1)

			opCtx, cancel := context.WithTimeout(ctx, operationTimeout)
			defer cancel()

			start := time.Now()

			deleteErr := p.deleteRecordWithIndex(opCtx, endpointItem, recordIndex)
			if deleteErr != nil {
				metrics.DNSOperationsTotal.WithLabelValues(operation, "error").Inc()

				errChan <- errors.Wrapf(deleteErr, "failed to delete record %s", endpointItem.DNSName)

				return
			}

			metrics.DNSOperationsTotal.WithLabelValues(operation, "success").Inc()
			metrics.DNSOperationDuration.WithLabelValues(operation).Observe(time.Since(start).Seconds())
		}(endpointToDelete)
	}

	wg.Wait()
	close(errChan)

	return collectErrors(errChan, "parallel deletions")
}

// parallelCreate performs parallel creation of DNS records.
func (p *UniFiProvider) parallelCreate(ctx context.Context, endpoints []*endpoint.Endpoint, operation string) error {
	sem := semaphore.NewWeighted(maxConcurrency)
	errChan := make(chan error, len(endpoints))

	var wg sync.WaitGroup

	for _, endpointToCreate := range endpoints {
		wg.Add(1)

		go func(endpointItem *endpoint.Endpoint) {
			defer wg.Done()

			// Acquire semaphore (blocks if at max concurrency)
			err := sem.Acquire(ctx, 1)
			if err != nil {
				errChan <- errors.Wrap(err, "failed to acquire semaphore")

				return
			}

			defer sem.Release(1)

			opCtx, cancel := context.WithTimeout(ctx, operationTimeout)
			defer cancel()

			start := time.Now()

			createErr := p.createRecord(opCtx, endpointItem)
			if createErr != nil {
				metrics.DNSOperationsTotal.WithLabelValues(operation, "error").Inc()

				errChan <- errors.Wrapf(createErr, "failed to create record %s", endpointItem.DNSName)

				return
			}

			metrics.DNSOperationsTotal.WithLabelValues(operation, "success").Inc()
			metrics.DNSOperationDuration.WithLabelValues(operation).Observe(time.Since(start).Seconds())
		}(endpointToCreate)
	}

	wg.Wait()
	close(errChan)

	return collectErrors(errChan, "parallel creations")
}

// collectErrors aggregates errors from an error channel into a single error.
func collectErrors(errChan chan error, operation string) error {
	//nolint:prealloc // Cannot pre-allocate: len(errChan) returns current queue size, not total errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}

	// Use strings.Builder for efficient string concatenation
	var builder strings.Builder
	builder.WriteString(operation)
	builder.WriteString(" failed: ")
	builder.WriteString(strconv.Itoa(len(errs)))
	builder.WriteString(" errors occurred: [")

	for i, err := range errs {
		if i > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString(err.Error())
	}

	builder.WriteString("]")

	return errors.New(builder.String())
}

func (p *UniFiProvider) applyCreations(ctx context.Context, endpoints []*endpoint.Endpoint) error {
	if len(endpoints) == 0 {
		return nil
	}

	return p.parallelCreate(ctx, endpoints, "create")
}

// unifiToEndpoint converts a UniFi DNS record to an endpoint.
func (p *UniFiProvider) unifiToEndpoint(record *unifi.DNSRecord) *endpoint.Endpoint {
	// Map UniFi record types to standard DNS types
	var recordType string

	switch record.RecordType {
	case unifi.DNSRecordRecordTypeA:
		recordType = endpoint.RecordTypeA
	case unifi.DNSRecordRecordTypeAAAA:
		recordType = endpoint.RecordTypeAAAA
	case unifi.DNSRecordRecordTypeCNAME:
		recordType = endpoint.RecordTypeCNAME
	case unifi.DNSRecordRecordTypeMX:
		recordType = endpoint.RecordTypeMX
	case unifi.DNSRecordRecordTypeNS:
		recordType = endpoint.RecordTypeNS
	case unifi.DNSRecordRecordTypeSRV:
		recordType = endpoint.RecordTypeSRV
	case unifi.DNSRecordRecordTypeTXT:
		recordType = endpoint.RecordTypeTXT
	default:
		// Skip unsupported record types
		return nil
	}

	ttl := endpoint.TTL(defaultTTL)
	if record.Ttl != nil {
		ttl = endpoint.TTL(*record.Ttl)
	}

	return &endpoint.Endpoint{
		DNSName:    record.Key,
		RecordType: recordType,
		RecordTTL:  ttl,
		Targets:    []string{record.Value},
	}
}

// endpointToUniFiWithTarget converts an endpoint to UniFi DNS record input format
// for a specific target value.
func (p *UniFiProvider) endpointToUniFiWithTarget(endpointData *endpoint.Endpoint, targetValue string) *unifi.DNSRecordInput {
	// Map standard DNS types to UniFi record types
	var recordType unifi.DNSRecordInputRecordType

	switch endpointData.RecordType {
	case endpoint.RecordTypeA:
		recordType = unifi.DNSRecordInputRecordTypeA
	case endpoint.RecordTypeAAAA:
		recordType = unifi.DNSRecordInputRecordTypeAAAA
	case endpoint.RecordTypeCNAME:
		recordType = unifi.DNSRecordInputRecordTypeCNAME
	case endpoint.RecordTypeMX:
		recordType = unifi.DNSRecordInputRecordTypeMX
	case endpoint.RecordTypeNS:
		recordType = unifi.DNSRecordInputRecordTypeNS
	case endpoint.RecordTypeSRV:
		recordType = unifi.DNSRecordInputRecordTypeSRV
	case endpoint.RecordTypeTXT:
		recordType = unifi.DNSRecordInputRecordTypeTXT
	default:
		// Skip unsupported record types
		return nil
	}

	enabled := true

	// UniFi API does not support TTL for TXT records
	// TXT records must not have ttl, port, weight, or priority fields
	var ttl *int

	if endpointData.RecordType != endpoint.RecordTypeTXT {
		ttlValue := int(endpointData.RecordTTL)
		if ttlValue == 0 {
			ttlValue = defaultTTL
		}

		ttl = &ttlValue
	}

	return &unifi.DNSRecordInput{
		Key:        endpointData.DNSName,
		RecordType: recordType,
		Value:      targetValue,
		Ttl:        ttl,
		Enabled:    &enabled,
	}
}

// createRecord creates DNS records in UniFi.
// For endpoints with multiple targets (e.g., A records with multiple IPs),
// creates separate DNS records with the same name.
func (p *UniFiProvider) createRecord(ctx context.Context, endpointToCreate *endpoint.Endpoint) error {
	if len(endpointToCreate.Targets) == 0 {
		//nolint:wrapcheck // errors.Newf already creates wrapped error
		return errors.Newf("endpoint has no targets: %s", endpointToCreate.DNSName)
	}

	// Create a separate DNS record for each target
	// This enables round-robin DNS for multiple IPs
	for _, target := range endpointToCreate.Targets {
		recordInput := p.endpointToUniFiWithTarget(endpointToCreate, target)
		if recordInput == nil {
			slog.WarnContext(ctx, "skipping unsupported record type",
				"name", endpointToCreate.DNSName,
				"type", endpointToCreate.RecordType)

			continue
		}

		slog.InfoContext(ctx, "creating DNS record",
			"name", endpointToCreate.DNSName,
			"type", endpointToCreate.RecordType,
			"target", target)

		slog.DebugContext(ctx, "DNS record input",
			"record_input", recordInput)

		_, err := p.client.CreateDNSRecord(ctx, p.site, recordInput)
		if err != nil {
			return errors.Wrapf(err, "failed to create DNS record for target %s", target)
		}
	}

	return nil
}

// deleteRecordWithIndex deletes a DNS record using a pre-built index.
// This avoids repeated API calls to list all records, significantly improving
// performance for batch operations (10+ records: 2-5s -> 200-400ms).
func (p *UniFiProvider) deleteRecordWithIndex(ctx context.Context, endpointToDelete *endpoint.Endpoint, recordIndex map[string][]unifi.DNSRecord) error {
	records := recordIndex[endpointToDelete.DNSName]

	if len(records) == 0 {
		slog.WarnContext(ctx, "record not found for deletion", "name", endpointToDelete.DNSName)

		return nil // Record doesn't exist, nothing to delete
	}

	// Delete all matching records
	for _, record := range records {
		slog.InfoContext(ctx, "deleting DNS record",
			"name", endpointToDelete.DNSName,
			"type", endpointToDelete.RecordType,
			"id", record.UnderscoreId)

		err := p.client.DeleteDNSRecord(ctx, p.site, record.UnderscoreId)
		if err != nil {
			return errors.Wrap(err, "failed to delete DNS record")
		}
	}

	return nil
}

// buildRecordIndex creates a map index of DNS records by name.
// This allows O(1) lookup instead of O(N) linear search.
func buildRecordIndex(records []unifi.DNSRecord) map[string][]unifi.DNSRecord {
	index := make(map[string][]unifi.DNSRecord, len(records))

	for _, record := range records {
		if existing := index[record.Key]; existing == nil {
			// First record with this key - pre-allocate capacity for typical case (1-2 targets)
			index[record.Key] = make([]unifi.DNSRecord, 0, 2)
		}

		index[record.Key] = append(index[record.Key], record)
	}

	return index
}
