package provider

import (
	"context"

	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
)

// DNSProvider defines the interface for DNS record management operations.
// This interface enables dependency injection and testing with mocks.
type DNSProvider interface {
	// Records retrieves all DNS records that match the domain filter.
	Records(ctx context.Context) ([]*endpoint.Endpoint, error)

	// ApplyChanges applies the given changes to DNS records.
	ApplyChanges(ctx context.Context, changes *plan.Changes) error

	// AdjustEndpoints allows the provider to modify endpoints before they are applied.
	AdjustEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error)
}
