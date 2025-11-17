//nolint:testpackage // Testing private functions and types requires same-package tests
package provider

import (
	"context"
	"fmt"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"

	unifi "github.com/lexfrei/go-unifi/api/network"
)

// MockNetworkClient is a mock implementation of unifi.NetworkAPIClient for testing.
type MockNetworkClient struct {
	mock.Mock
}

// ListDNSRecords mocks the ListDNSRecords method.
func (m *MockNetworkClient) ListDNSRecords(ctx context.Context, site unifi.Site) ([]unifi.DNSRecord, error) {
	args := m.Called(ctx, site)
	if args.Get(0) == nil {
		//nolint:wrapcheck // Test mock: errors from testify/mock don't need wrapping
		return nil, args.Error(1)
	}

	//nolint:forcetypeassert,wrapcheck // Test mock: type assertion is safe, errors don't need wrapping
	return args.Get(0).([]unifi.DNSRecord), args.Error(1)
}

// CreateDNSRecord mocks the CreateDNSRecord method.
func (m *MockNetworkClient) CreateDNSRecord(ctx context.Context, site unifi.Site, record *unifi.DNSRecordInput) (*unifi.DNSRecord, error) {
	args := m.Called(ctx, site, record)
	if args.Get(0) == nil {
		//nolint:wrapcheck // Test mock: errors from testify/mock don't need wrapping
		return nil, args.Error(1)
	}

	//nolint:forcetypeassert,wrapcheck // Test mock: type assertion is safe, errors don't need wrapping
	return args.Get(0).(*unifi.DNSRecord), args.Error(1)
}

// DeleteDNSRecord mocks the DeleteDNSRecord method.
func (m *MockNetworkClient) DeleteDNSRecord(ctx context.Context, site unifi.Site, recordID unifi.RecordId) error {
	args := m.Called(ctx, site, recordID)

	//nolint:wrapcheck // Test mock: errors from testify/mock don't need wrapping
	return args.Error(0)
}

// UpdateDNSRecord mocks the UpdateDNSRecord method.
func (m *MockNetworkClient) UpdateDNSRecord(ctx context.Context, site unifi.Site, recordID unifi.RecordId, record *unifi.DNSRecordInput) (*unifi.DNSRecord, error) {
	args := m.Called(ctx, site, recordID, record)
	if args.Get(0) == nil {
		//nolint:wrapcheck // Test mock: errors from testify/mock don't need wrapping
		return nil, args.Error(1)
	}

	//nolint:forcetypeassert,wrapcheck // Test mock: type assertion is safe, errors don't need wrapping
	return args.Get(0).(*unifi.DNSRecord), args.Error(1)
}

// Stub implementations for unused interface methods.
//
//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) ListSites(context.Context, *unifi.ListSitesParams) (*unifi.SitesResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) ListSiteDevices(context.Context, unifi.SiteId, *unifi.ListSiteDevicesParams) (*unifi.DevicesResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) GetDeviceByID(context.Context, unifi.SiteId, unifi.DeviceId) (*unifi.Device, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) ListSiteClients(context.Context, unifi.SiteId, *unifi.ListSiteClientsParams) (*unifi.ClientsResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) GetClientByID(context.Context, unifi.SiteId, unifi.ClientId) (*unifi.NetworkClient, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) ListHotspotVouchers(context.Context, unifi.SiteId, *unifi.ListHotspotVouchersParams) (*unifi.HotspotVouchersResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) CreateHotspotVouchers(context.Context, unifi.SiteId, *unifi.CreateVouchersRequest) (*unifi.HotspotVouchersResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) GetHotspotVoucher(context.Context, unifi.SiteId, openapi_types.UUID) (*unifi.HotspotVoucher, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) DeleteHotspotVoucher(context.Context, unifi.SiteId, openapi_types.UUID) error {
	return fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) ListFirewallPolicies(context.Context, unifi.Site) ([]unifi.FirewallPolicy, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) CreateFirewallPolicy(context.Context, unifi.Site, *unifi.FirewallPolicyInput) (*unifi.FirewallPolicy, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) UpdateFirewallPolicy(context.Context, unifi.Site, unifi.PolicyId, *unifi.FirewallPolicyInput) (*unifi.FirewallPolicy, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) DeleteFirewallPolicy(context.Context, unifi.Site, unifi.PolicyId) error {
	return fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) ListTrafficRules(context.Context, unifi.Site) ([]unifi.TrafficRule, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) CreateTrafficRule(context.Context, unifi.Site, *unifi.TrafficRuleInput) (*unifi.TrafficRule, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) UpdateTrafficRule(context.Context, unifi.Site, unifi.RuleId, *unifi.TrafficRuleInput) (*unifi.TrafficRule, error) {
	return nil, fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) DeleteTrafficRule(context.Context, unifi.Site, unifi.RuleId) error {
	return fmt.Errorf("not implemented")
}

//nolint:err113,perfsprint // Test mock: static errors are acceptable for unimplemented methods
func (m *MockNetworkClient) GetAggregatedDashboard(context.Context, unifi.Site, *unifi.GetAggregatedDashboardParams) (*unifi.AggregatedDashboard, error) {
	return nil, fmt.Errorf("not implemented")
}

// Test helpers.

func createMockDNSRecord(key, value string, recordType unifi.DNSRecordRecordType) unifi.DNSRecord {
	ttl := 300

	return unifi.DNSRecord{
		UnderscoreId: "test-id-" + key,
		Key:          key,
		Value:        value,
		RecordType:   recordType,
		Ttl:          &ttl,
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	mockClient := new(MockNetworkClient)
	domainFilter := endpoint.DomainFilter{}

	provider := New(mockClient, "default", domainFilter)

	assert.NotNil(t, provider)
	assert.Equal(t, "default", provider.site)
	assert.Equal(t, domainFilter, provider.domainFilter)
}

func TestRecords_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockNetworkClient)
	domainFilter := endpoint.DomainFilter{}

	mockRecords := []unifi.DNSRecord{
		createMockDNSRecord("example.com", "192.168.1.1", unifi.DNSRecordRecordTypeA),
		createMockDNSRecord("test.com", "192.168.1.2", unifi.DNSRecordRecordTypeA),
	}

	mockClient.On("ListDNSRecords", mock.Anything, unifi.Site("default")).
		Return(mockRecords, nil)

	provider := New(mockClient, "default", domainFilter)
	endpoints, err := provider.Records(context.Background())

	//nolint:testifylint // Using assert for consistency with other tests in this file
	assert.NoError(t, err)
	assert.Len(t, endpoints, 2)
	assert.Equal(t, "example.com", endpoints[0].DNSName)
	assert.Equal(t, "test.com", endpoints[1].DNSName)
	mockClient.AssertExpectations(t)
}

func TestRecords_WithDomainFilter(t *testing.T) {
	t.Parallel()

	mockClient := new(MockNetworkClient)
	domainFilter := endpoint.NewDomainFilter([]string{"example.com"})

	mockRecords := []unifi.DNSRecord{
		createMockDNSRecord("example.com", "192.168.1.1", unifi.DNSRecordRecordTypeA),
		createMockDNSRecord("test.com", "192.168.1.2", unifi.DNSRecordRecordTypeA),
	}

	mockClient.On("ListDNSRecords", mock.Anything, unifi.Site("default")).
		Return(mockRecords, nil)

	provider := New(mockClient, "default", *domainFilter)
	endpoints, err := provider.Records(context.Background())

	//nolint:testifylint // Using assert for consistency with other tests in this file
	assert.NoError(t, err)
	assert.Len(t, endpoints, 1, "should filter to only example.com")
	assert.Equal(t, "example.com", endpoints[0].DNSName)
	mockClient.AssertExpectations(t)
}

func TestRecords_APIError(t *testing.T) {
	t.Parallel()

	mockClient := new(MockNetworkClient)
	domainFilter := endpoint.DomainFilter{}

	//nolint:err113,perfsprint // Test case: intentionally using dynamic error for realistic API error simulation
	mockClient.On("ListDNSRecords", mock.Anything, unifi.Site("default")).
		Return(nil, fmt.Errorf("API connection timeout"))

	provider := New(mockClient, "default", domainFilter)
	endpoints, err := provider.Records(context.Background())

	//nolint:testifylint // Using assert for consistency with other tests in this file
	assert.Error(t, err)
	assert.Nil(t, endpoints)
	assert.Contains(t, err.Error(), "failed to list DNS records")
	mockClient.AssertExpectations(t)
}

func TestApplyChanges_Create(t *testing.T) {
	t.Parallel()

	mockClient := new(MockNetworkClient)
	domainFilter := endpoint.DomainFilter{}

	ttl := 300
	createdRecord := &unifi.DNSRecord{
		UnderscoreId: "new-record-id",
		Key:          "new.example.com",
		Value:        "192.168.1.10",
		RecordType:   unifi.DNSRecordRecordTypeA,
		Ttl:          &ttl,
	}

	mockClient.On("CreateDNSRecord", mock.Anything, unifi.Site("default"), mock.MatchedBy(func(input *unifi.DNSRecordInput) bool {
		return input.Key == "new.example.com" && input.Value == "192.168.1.10"
	})).Return(createdRecord, nil)

	provider := New(mockClient, "default", domainFilter)

	changes := &plan.Changes{
		Create: []*endpoint.Endpoint{
			{
				DNSName:    "new.example.com",
				RecordType: endpoint.RecordTypeA,
				Targets:    []string{"192.168.1.10"},
			},
		},
	}

	err := provider.ApplyChanges(context.Background(), changes)

	//nolint:testifylint // Using assert for consistency with other tests in this file
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestApplyChanges_Delete(t *testing.T) {
	t.Parallel()

	mockClient := new(MockNetworkClient)
	domainFilter := endpoint.DomainFilter{}

	existingRecords := []unifi.DNSRecord{
		createMockDNSRecord("delete.example.com", "192.168.1.20", unifi.DNSRecordRecordTypeA),
	}

	mockClient.On("ListDNSRecords", mock.Anything, unifi.Site("default")).
		Return(existingRecords, nil)

	mockClient.On("DeleteDNSRecord", mock.Anything, unifi.Site("default"), unifi.RecordId("test-id-delete.example.com")).
		Return(nil)

	provider := New(mockClient, "default", domainFilter)

	changes := &plan.Changes{
		Delete: []*endpoint.Endpoint{
			{
				DNSName:    "delete.example.com",
				RecordType: endpoint.RecordTypeA,
				Targets:    []string{"192.168.1.20"},
			},
		},
	}

	err := provider.ApplyChanges(context.Background(), changes)

	//nolint:testifylint // Using assert for consistency with other tests in this file
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestApplyChanges_Update(t *testing.T) {
	t.Parallel()

	mockClient := new(MockNetworkClient)
	domainFilter := endpoint.DomainFilter{}

	existingRecords := []unifi.DNSRecord{
		createMockDNSRecord("update.example.com", "192.168.1.30", unifi.DNSRecordRecordTypeA),
	}

	ttl := 300
	newRecord := &unifi.DNSRecord{
		UnderscoreId: "updated-record-id",
		Key:          "update.example.com",
		Value:        "192.168.1.31",
		RecordType:   unifi.DNSRecordRecordTypeA,
		Ttl:          &ttl,
	}

	mockClient.On("ListDNSRecords", mock.Anything, unifi.Site("default")).
		Return(existingRecords, nil)

	mockClient.On("DeleteDNSRecord", mock.Anything, unifi.Site("default"), unifi.RecordId("test-id-update.example.com")).
		Return(nil)

	mockClient.On("CreateDNSRecord", mock.Anything, unifi.Site("default"), mock.MatchedBy(func(input *unifi.DNSRecordInput) bool {
		return input.Key == "update.example.com" && input.Value == "192.168.1.31"
	})).Return(newRecord, nil)

	provider := New(mockClient, "default", domainFilter)

	changes := &plan.Changes{
		UpdateOld: []*endpoint.Endpoint{
			{
				DNSName:    "update.example.com",
				RecordType: endpoint.RecordTypeA,
				Targets:    []string{"192.168.1.30"},
			},
		},
		UpdateNew: []*endpoint.Endpoint{
			{
				DNSName:    "update.example.com",
				RecordType: endpoint.RecordTypeA,
				Targets:    []string{"192.168.1.31"},
			},
		},
	}

	err := provider.ApplyChanges(context.Background(), changes)

	//nolint:testifylint // Using assert for consistency with other tests in this file
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestUnifiToEndpoint(t *testing.T) {
	t.Parallel()

	mockClient := new(MockNetworkClient)
	domainFilter := endpoint.DomainFilter{}
	provider := New(mockClient, "default", domainFilter)

	tests := []struct {
		name         string
		record       unifi.DNSRecord
		expectedType string
		shouldBeNil  bool
	}{
		{
			name:         "A record",
			record:       createMockDNSRecord("a.example.com", "192.168.1.1", unifi.DNSRecordRecordTypeA),
			expectedType: endpoint.RecordTypeA,
			shouldBeNil:  false,
		},
		{
			name:         "AAAA record",
			record:       createMockDNSRecord("aaaa.example.com", "2001:db8::1", unifi.DNSRecordRecordTypeAAAA),
			expectedType: endpoint.RecordTypeAAAA,
			shouldBeNil:  false,
		},
		{
			name:         "CNAME record",
			record:       createMockDNSRecord("cname.example.com", "target.example.com", unifi.DNSRecordRecordTypeCNAME),
			expectedType: endpoint.RecordTypeCNAME,
			shouldBeNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			//nolint:varnamelen // Variable 'ep' is idiomatic abbreviation for endpoint in this context
			ep := provider.unifiToEndpoint(&tt.record)

			if tt.shouldBeNil {
				assert.Nil(t, ep)
			} else {
				assert.NotNil(t, ep)
				assert.Equal(t, tt.record.Key, ep.DNSName)
				assert.Equal(t, tt.expectedType, ep.RecordType)
				assert.Equal(t, endpoint.Targets{tt.record.Value}, ep.Targets)
			}
		})
	}
}

func TestBuildRecordIndex(t *testing.T) {
	t.Parallel()

	records := []unifi.DNSRecord{
		createMockDNSRecord("example.com", "192.168.1.1", unifi.DNSRecordRecordTypeA),
		createMockDNSRecord("example.com", "192.168.1.2", unifi.DNSRecordRecordTypeA),
		createMockDNSRecord("test.com", "192.168.1.3", unifi.DNSRecordRecordTypeA),
	}

	index := buildRecordIndex(records)

	assert.Len(t, index, 2)
	assert.Len(t, index["example.com"], 2)
	assert.Len(t, index["test.com"], 1)
	assert.Equal(t, "192.168.1.1", index["example.com"][0].Value)
	assert.Equal(t, "192.168.1.2", index["example.com"][1].Value)
}
