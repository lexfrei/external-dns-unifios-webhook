package metrics

import "github.com/prometheus/client_golang/prometheus"

const namespace = "external_dns_unifi"

//nolint:gochecknoglobals // Prometheus metrics must be global
var (
	// DNSOperationsTotal tracks the total number of DNS operations (create/update/delete).
	DNSOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "dns_operations_total",
			Help:      "Total number of DNS operations",
		},
		[]string{"operation", "status"}, // operation: create/update/delete, status: success/error
	)

	// DNSOperationDuration tracks the duration of DNS operations.
	DNSOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "dns_operation_duration_seconds",
			Help:      "Duration of DNS operations in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"operation"}, // operation: create/update/delete
	)

	// DNSRecordsManaged tracks the number of DNS records currently managed by type.
	DNSRecordsManaged = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "dns_records_managed",
			Help:      "Number of DNS records currently managed",
		},
		[]string{"record_type"}, // record_type: A/AAAA/CNAME/TXT
	)

	// DNSChangesApplied tracks the number of DNS changes applied per batch.
	DNSChangesApplied = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "dns_changes_applied",
			Help:      "Number of DNS changes applied in a single ApplyChanges call",
			Buckets:   []float64{1, 5, 10, 25, 50, 100},
		},
		[]string{"change_type"}, // change_type: create/update/delete
	)

	// ReadinessCacheHits tracks the number of readiness cache hits.
	ReadinessCacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "readiness_cache_hits_total",
			Help:      "Total number of readiness cache hits",
		},
	)

	// ReadinessCacheMisses tracks the number of readiness cache misses.
	ReadinessCacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "readiness_cache_misses_total",
			Help:      "Total number of readiness cache misses",
		},
	)

	// ReadinessCacheAge tracks the age of the readiness cache in seconds.
	ReadinessCacheAge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "readiness_cache_age_seconds",
			Help:      "Age of the readiness cache in seconds",
		},
	)
)

// Register registers all custom metrics with the given Prometheus registry.
func Register(registry *prometheus.Registry) {
	registry.MustRegister(
		DNSOperationsTotal,
		DNSOperationDuration,
		DNSRecordsManaged,
		DNSChangesApplied,
		ReadinessCacheHits,
		ReadinessCacheMisses,
		ReadinessCacheAge,
	)
}
