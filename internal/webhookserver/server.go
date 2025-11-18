package webhookserver

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/lexfrei/external-dns-unifios-webhook/api/webhook"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/provider"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

const (
	// maxRequestBodySize limits the maximum size of HTTP request body to prevent memory exhaustion.
	// 1MB is more than sufficient for any DNS record changes request.
	maxRequestBodySize = 1 << 20 // 1MB
)

// Server implements the webhook.ServerInterface for external-dns webhook protocol.
type Server struct {
	provider provider.DNSProvider
	filter   endpoint.DomainFilter
}

// New creates a new webhook server instance.
func New(prov provider.DNSProvider, filter endpoint.DomainFilter) *Server {
	return &Server{
		provider: prov,
		filter:   filter,
	}
}

// Negotiate returns the domain filter configuration.
// GET /.
func (s *Server) Negotiate(w http.ResponseWriter, r *http.Request, _ webhook.NegotiateParams) {
	slog.InfoContext(r.Context(), "negotiate called")

	// Return configured domain filters
	filters := s.filter.Filters

	response := webhook.Filters{
		Filters: &filters,
	}

	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetRecords returns all DNS records from UniFi.
// GET /records.
func (s *Server) GetRecords(w http.ResponseWriter, r *http.Request, _ webhook.GetRecordsParams) {
	slog.InfoContext(r.Context(), "get records called")

	// Fetch records from provider
	endpoints, err := s.provider.Records(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to get records", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})

		return
	}

	// Convert external-dns endpoints to webhook API endpoints
	// Pre-allocate exact size to avoid slice growth
	webhookEndpoints := make(webhook.Endpoints, len(endpoints))
	for idx, ep := range endpoints {
		webhookEndpoints[idx] = convertToWebhookEndpoint(ep)
	}

	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(webhookEndpoints)
}

// SetRecords applies DNS record changes.
// POST /records.
func (s *Server) SetRecords(w http.ResponseWriter, r *http.Request, _ webhook.SetRecordsParams) {
	slog.InfoContext(r.Context(), "set records called", "content_type", r.Header.Get("Content-Type"))

	// Limit request body size to prevent memory exhaustion attacks
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	changes, err := s.decodeChanges(r)
	if err != nil {
		// Check if error is due to body size limit
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			slog.WarnContext(r.Context(), "request body too large", "limit_bytes", maxRequestBodySize)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "request body too large"})

			return
		}

		slog.ErrorContext(r.Context(), "failed to decode changes", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})

		return
	}

	// Convert webhook changes to external-dns plan
	planChanges := convertToPlan(changes)

	// Apply changes using provider
	err = s.provider.ApplyChanges(r.Context(), planChanges)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to apply changes", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AdjustRecords allows the provider to modify endpoints before they are applied.
// POST /adjustendpoints.
func (s *Server) AdjustRecords(w http.ResponseWriter, r *http.Request, _ webhook.AdjustRecordsParams) {
	slog.InfoContext(r.Context(), "adjust records called")

	var endpoints webhook.Endpoints

	err := json.NewDecoder(r.Body).Decode(&endpoints)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to bind endpoints", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})

		return
	}

	// Convert to external-dns endpoints
	externalEndpoints := make([]*endpoint.Endpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		externalEndpoints = append(externalEndpoints, convertFromWebhookEndpoint(&ep))
	}

	// Provider's AdjustEndpoints (currently just returns the same endpoints)
	adjusted, err := s.provider.AdjustEndpoints(externalEndpoints)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to adjust endpoints", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})

		return
	}

	// Convert back to webhook format
	result := make(webhook.Endpoints, 0, len(adjusted))
	for _, ep := range adjusted {
		result = append(result, convertToWebhookEndpoint(ep))
	}

	w.Header().Set("Content-Type", "application/external.dns.webhook+json;version=1")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// convertToWebhookEndpoint converts external-dns endpoint to webhook endpoint.
func convertToWebhookEndpoint(externalEndpoint *endpoint.Endpoint) webhook.Endpoint {
	ttl := int64(externalEndpoint.RecordTTL)

	return webhook.Endpoint{
		DnsName:          &externalEndpoint.DNSName,
		RecordType:       &externalEndpoint.RecordType,
		RecordTTL:        &ttl,
		Targets:          (*webhook.Targets)(&externalEndpoint.Targets),
		SetIdentifier:    &externalEndpoint.SetIdentifier,
		Labels:           (*map[string]string)(&externalEndpoint.Labels),
		ProviderSpecific: convertProviderSpecific(externalEndpoint.ProviderSpecific),
	}
}

// convertFromWebhookEndpoint converts webhook endpoint to external-dns endpoint.
func convertFromWebhookEndpoint(webhookEndpoint *webhook.Endpoint) *endpoint.Endpoint {
	externalEndpoint := &endpoint.Endpoint{}

	if webhookEndpoint.DnsName != nil {
		externalEndpoint.DNSName = *webhookEndpoint.DnsName
	}

	if webhookEndpoint.RecordType != nil {
		externalEndpoint.RecordType = *webhookEndpoint.RecordType
	}

	if webhookEndpoint.RecordTTL != nil {
		externalEndpoint.RecordTTL = endpoint.TTL(*webhookEndpoint.RecordTTL)
	}

	if webhookEndpoint.Targets != nil {
		externalEndpoint.Targets = *webhookEndpoint.Targets
	}

	if webhookEndpoint.SetIdentifier != nil {
		externalEndpoint.SetIdentifier = *webhookEndpoint.SetIdentifier
	}

	if webhookEndpoint.Labels != nil {
		externalEndpoint.Labels = *webhookEndpoint.Labels
	}

	if webhookEndpoint.ProviderSpecific != nil {
		externalEndpoint.ProviderSpecific = convertFromProviderSpecific(webhookEndpoint.ProviderSpecific)
	}

	return externalEndpoint
}

// convertProviderSpecific converts external-dns ProviderSpecific to webhook format.
func convertProviderSpecific(providerSpecific endpoint.ProviderSpecific) *[]webhook.ProviderSpecificProperty {
	if len(providerSpecific) == 0 {
		return nil
	}

	// Pre-allocate exact size and use indexed assignment to eliminate bounds checks
	result := make([]webhook.ProviderSpecificProperty, len(providerSpecific))
	for i := range providerSpecific {
		name := providerSpecific[i].Name
		value := providerSpecific[i].Value
		result[i] = webhook.ProviderSpecificProperty{
			Name:  &name,
			Value: &value,
		}
	}

	return &result
}

// convertFromProviderSpecific converts webhook ProviderSpecific to external-dns format.
func convertFromProviderSpecific(wps *[]webhook.ProviderSpecificProperty) endpoint.ProviderSpecific {
	if wps == nil || len(*wps) == 0 {
		return nil
	}

	result := make(endpoint.ProviderSpecific, 0, len(*wps))
	for _, prop := range *wps {
		if prop.Name != nil && prop.Value != nil {
			result = append(result, endpoint.ProviderSpecificProperty{
				Name:  *prop.Name,
				Value: *prop.Value,
			})
		}
	}

	return result
}

// convertToPlan converts webhook changes to external-dns plan.
func convertToPlan(changes *webhook.Changes) *plan.Changes {
	planChanges := &plan.Changes{}

	// Pre-allocate slices with exact sizes to avoid growth
	if changes.Create != nil && len(*changes.Create) > 0 {
		planChanges.Create = make([]*endpoint.Endpoint, len(*changes.Create))
		for idx, ep := range *changes.Create {
			planChanges.Create[idx] = convertFromWebhookEndpoint(&ep)
		}
	}

	if changes.UpdateOld != nil && len(*changes.UpdateOld) > 0 {
		planChanges.UpdateOld = make([]*endpoint.Endpoint, len(*changes.UpdateOld))
		for idx, ep := range *changes.UpdateOld {
			planChanges.UpdateOld[idx] = convertFromWebhookEndpoint(&ep)
		}
	}

	if changes.UpdateNew != nil && len(*changes.UpdateNew) > 0 {
		planChanges.UpdateNew = make([]*endpoint.Endpoint, len(*changes.UpdateNew))
		for idx, ep := range *changes.UpdateNew {
			planChanges.UpdateNew[idx] = convertFromWebhookEndpoint(&ep)
		}
	}

	if changes.Delete != nil && len(*changes.Delete) > 0 {
		planChanges.Delete = make([]*endpoint.Endpoint, len(*changes.Delete))
		for idx, ep := range *changes.Delete {
			planChanges.Delete[idx] = convertFromWebhookEndpoint(&ep)
		}
	}

	return planChanges
}

// decodeChanges reads and decodes the changes from request body.
// Uses streaming JSON decoder to reduce memory allocations.
func (s *Server) decodeChanges(r *http.Request) (*webhook.Changes, error) {
	var changes webhook.Changes

	// Use streaming decoder instead of reading entire body into memory
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict validation

	err := decoder.Decode(&changes)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to decode changes",
			"error", err,
			"content_type", r.Header.Get("Content-Type"))

		return nil, errors.Wrap(err, "failed to decode changes")
	}

	// Log only counts, not full body content (saves memory in debug mode)
	createCount := 0
	if changes.Create != nil {
		createCount = len(*changes.Create)
	}

	updateCount := 0
	if changes.UpdateNew != nil {
		updateCount = len(*changes.UpdateNew)
	}

	deleteCount := 0
	if changes.Delete != nil {
		deleteCount = len(*changes.Delete)
	}

	slog.DebugContext(r.Context(), "decoded changes successfully",
		"create_count", createCount,
		"update_count", updateCount,
		"delete_count", deleteCount)

	return &changes, nil
}
