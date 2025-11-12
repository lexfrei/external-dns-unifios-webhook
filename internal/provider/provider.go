package provider

import (
	"context"
	"log/slog"

	"github.com/cockroachdb/errors"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/config"
	unifi "github.com/lexfrei/go-unifi/api/network"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

const defaultTTL = 300

// UniFiProvider implements the provider.Provider interface for UniFi OS.
type UniFiProvider struct {
	client       *unifi.APIClient
	site         string
	domainFilter endpoint.DomainFilter
}

// New creates a new UniFiProvider instance.
func New(cfg config.UniFiConfig, domainFilter endpoint.DomainFilter) (*UniFiProvider, error) {
	// Create UniFi API client using the official constructor
	// This properly configures the base URL with /proxy/network prefix and API key authentication
	client, err := unifi.NewWithConfig(&unifi.ClientConfig{
		ControllerURL:      cfg.Host,
		APIKey:             cfg.APIKey,
		InsecureSkipVerify: cfg.SkipTLSVerify,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create UniFi API client")
	}

	if cfg.SkipTLSVerify {
		slog.Warn("TLS certificate verification is disabled")
	}

	return &UniFiProvider{
		client:       client,
		site:         cfg.Site,
		domainFilter: domainFilter,
	}, nil
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
	var endpoints []*endpoint.Endpoint

	for _, record := range records {
		// Skip records that don't match the domain filter
		if !p.domainFilter.Match(record.Key) {
			continue
		}

		endpointRecord := p.unifiToEndpoint(&record)
		if endpointRecord != nil {
			endpoints = append(endpoints, endpointRecord)
		}
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

	// Handle deletions
	for _, endpointToDelete := range changes.Delete {
		err := p.deleteRecord(ctx, endpointToDelete)
		if err != nil {
			return errors.Wrapf(err, "failed to delete record %s", endpointToDelete.DNSName)
		}
	}

	// Handle updates (delete old, create new)
	for _, oldEndpoint := range changes.UpdateOld {
		err := p.deleteRecord(ctx, oldEndpoint)
		if err != nil {
			return errors.Wrapf(err, "failed to delete old record %s", oldEndpoint.DNSName)
		}
	}

	for _, newEndpoint := range changes.UpdateNew {
		err := p.createRecord(ctx, newEndpoint)
		if err != nil {
			return errors.Wrapf(err, "failed to create updated record %s", newEndpoint.DNSName)
		}
	}

	// Handle creations
	for _, endpointToCreate := range changes.Create {
		err := p.createRecord(ctx, endpointToCreate)
		if err != nil {
			return errors.Wrapf(err, "failed to create record %s", endpointToCreate.DNSName)
		}
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

// deleteRecord deletes a DNS record from UniFi.
func (p *UniFiProvider) deleteRecord(ctx context.Context, endpointToDelete *endpoint.Endpoint) error {
	// First, we need to find the record ID
	records, err := p.findRecordsByName(ctx, endpointToDelete.DNSName)
	if err != nil {
		return errors.Wrap(err, "failed to find records")
	}

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

// findRecordsByName finds all DNS records with the given name.
func (p *UniFiProvider) findRecordsByName(ctx context.Context, name string) ([]unifi.DNSRecord, error) {
	records, err := p.client.ListDNSRecords(ctx, p.site)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list DNS records")
	}

	var matching []unifi.DNSRecord

	for _, record := range records {
		if record.Key == name {
			matching = append(matching, record)
		}
	}

	return matching, nil
}
