//nolint:testpackage // Benchmarks test internal functions
package webhookserver

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/lexfrei/external-dns-unifios-webhook/api/webhook"
	"sigs.k8s.io/external-dns/endpoint"
)

// generateWebhookEndpoints creates a slice of webhook endpoints for benchmarking.
func generateWebhookEndpoints(count int) webhook.Endpoints {
	endpoints := make(webhook.Endpoints, count)

	for i := range endpoints {
		id := string(rune('a' + (i % 26)))
		dnsName := "test" + id + ".example.com"
		recordType := "A"
		ttl := int64(300)
		target := "192.168.1." + string(rune('1'+i%254))
		setIdentifier := ""

		endpoints[i] = webhook.Endpoint{
			DnsName:       &dnsName,
			RecordType:    &recordType,
			RecordTTL:     &ttl,
			Targets:       &webhook.Targets{target},
			SetIdentifier: &setIdentifier,
		}
	}

	return endpoints
}

// generateExternalEndpoints creates a slice of external-dns endpoints for benchmarking.
func generateExternalEndpoints(count int) []*endpoint.Endpoint {
	endpoints := make([]*endpoint.Endpoint, count)

	for i := range endpoints {
		id := string(rune('a' + (i % 26)))
		endpoints[i] = &endpoint.Endpoint{
			DNSName:    "test" + id + ".example.com",
			RecordType: endpoint.RecordTypeA,
			RecordTTL:  300,
			Targets:    []string{"192.168.1." + string(rune('1'+i%254))},
		}
	}

	return endpoints
}

func BenchmarkConvertToWebhookEndpoint(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"10_endpoints", 10},
		{"50_endpoints", 50},
		{"100_endpoints", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			externalEndpoints := generateExternalEndpoints(bm.count)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				for _, ep := range externalEndpoints {
					_ = convertToWebhookEndpoint(ep)
				}
			}
		})
	}
}

func BenchmarkConvertFromWebhookEndpoint(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"10_endpoints", 10},
		{"50_endpoints", 50},
		{"100_endpoints", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			webhookEndpoints := generateWebhookEndpoints(bm.count)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				for i := range webhookEndpoints {
					_ = convertFromWebhookEndpoint(&webhookEndpoints[i])
				}
			}
		})
	}
}

func BenchmarkConvertProviderSpecific(b *testing.B) {
	providerSpecific := endpoint.ProviderSpecific{
		{Name: "key1", Value: "value1"},
		{Name: "key2", Value: "value2"},
		{Name: "key3", Value: "value3"},
		{Name: "key4", Value: "value4"},
		{Name: "key5", Value: "value5"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = convertProviderSpecific(providerSpecific)
	}
}

func BenchmarkJSONEncoder(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"10_endpoints", 10},
		{"50_endpoints", 50},
		{"100_endpoints", 100},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			endpoints := generateWebhookEndpoints(bench.count)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				var buf bytes.Buffer

				_ = json.NewEncoder(&buf).Encode(endpoints)
			}
		})
	}
}

func BenchmarkJSONDecoder(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"10_endpoints", 10},
		{"50_endpoints", 50},
		{"100_endpoints", 100},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			endpoints := generateWebhookEndpoints(bench.count)

			// Pre-encode the test data
			var buf bytes.Buffer

			_ = json.NewEncoder(&buf).Encode(endpoints)

			data := buf.Bytes()

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				var decoded webhook.Endpoints

				reader := bytes.NewReader(data)

				_ = json.NewDecoder(reader).Decode(&decoded)
			}
		})
	}
}

func BenchmarkConvertToPlan(b *testing.B) {
	create := generateWebhookEndpoints(10)
	updateOld := generateWebhookEndpoints(10)
	updateNew := generateWebhookEndpoints(10)
	deleteEndpoints := generateWebhookEndpoints(10)

	changes := &webhook.Changes{
		Create:    &create,
		UpdateOld: &updateOld,
		UpdateNew: &updateNew,
		Delete:    &deleteEndpoints,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_ = convertToPlan(changes)
	}
}
