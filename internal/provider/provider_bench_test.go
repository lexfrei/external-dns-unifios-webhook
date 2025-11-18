//nolint:testpackage // Benchmarks test internal functions
package provider

import (
	"testing"

	unifi "github.com/lexfrei/go-unifi/api/network"
)

// generateMockRecords creates a slice of mock DNS records for benchmarking.
func generateMockRecords(count int) []unifi.DNSRecord {
	records := make([]unifi.DNSRecord, count)
	ttl := 300

	for i := range records {
		id := string(rune('a' + (i % 26)))
		records[i] = unifi.DNSRecord{
			UnderscoreId: id + "507f1f77bcf86cd799439011",
			Key:          "test" + id + ".example.com",
			RecordType:   unifi.DNSRecordRecordTypeA,
			Value:        "192.168.1." + string(rune('1'+i%254)),
			Ttl:          &ttl,
		}
	}

	return records
}

func BenchmarkBuildRecordIndex(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"10_records", 10},
		{"50_records", 50},
		{"100_records", 100},
		{"500_records", 500},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			records := generateMockRecords(bench.count)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				_ = buildRecordIndex(records)
			}
		})
	}
}

func BenchmarkCollectErrors(b *testing.B) {
	benchmarks := []struct {
		name       string
		errorCount int
	}{
		{"no_errors", 0},
		{"5_errors", 5},
		{"10_errors", 10},
		{"20_errors", 20},
	}

	for _, bench := range benchmarks {
		b.Run(bench.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				errChan := make(chan error, bench.errorCount)

				for i := range bench.errorCount {
					errChan <- &testError{msg: "error " + string(rune('0'+i))}
				}

				close(errChan)

				_ = collectErrors(errChan, "test operation")
			}
		})
	}
}

// testError is a simple error implementation for benchmarking.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
