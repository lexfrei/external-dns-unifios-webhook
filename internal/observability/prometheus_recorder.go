package observability

import (
	"strconv"
	"time"

	"github.com/lexfrei/go-unifi/observability"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusRecorder implements observability.MetricsRecorder using Prometheus.
type PrometheusRecorder struct {
	httpRequests         *prometheus.CounterVec
	httpDuration         *prometheus.HistogramVec
	retries              *prometheus.CounterVec
	rateLimits           *prometheus.CounterVec
	errors               *prometheus.CounterVec
	contextCancellations *prometheus.CounterVec
}

// NewPrometheusRecorder creates a new PrometheusRecorder and registers metrics with the given registry.
//
//nolint:funlen // Constructor initializes multiple metrics - length is acceptable for clarity
func NewPrometheusRecorder(registry *prometheus.Registry, namespace string) *PrometheusRecorder {
	r := &PrometheusRecorder{
		httpRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "unifi_api_requests_total",
				Help:      "Total number of UniFi API HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),
		httpDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "unifi_api_request_duration_seconds",
				Help:      "Duration of UniFi API HTTP requests in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		retries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "unifi_api_retries_total",
				Help:      "Total number of UniFi API retry attempts",
			},
			[]string{"endpoint"},
		),
		rateLimits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "unifi_api_rate_limits_total",
				Help:      "Total number of UniFi API rate limit events",
			},
			[]string{"endpoint"},
		),
		errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "unifi_api_errors_total",
				Help:      "Total number of UniFi API errors",
			},
			[]string{"operation", "error_type"},
		),
		contextCancellations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "unifi_api_context_cancellations_total",
				Help:      "Total number of UniFi API context cancellations",
			},
			[]string{"operation"},
		),
	}

	registry.MustRegister(
		r.httpRequests,
		r.httpDuration,
		r.retries,
		r.rateLimits,
		r.errors,
		r.contextCancellations,
	)

	return r
}

// RecordHTTPRequest records an HTTP request with method, path, status code, and duration.
func (r *PrometheusRecorder) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	r.httpRequests.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
	r.httpDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordRetry records a retry attempt for an endpoint.
func (r *PrometheusRecorder) RecordRetry(_ int, endpoint string) {
	r.retries.WithLabelValues(endpoint).Inc()
}

// RecordRateLimit records a rate limit wait event.
func (r *PrometheusRecorder) RecordRateLimit(endpoint string, _ time.Duration) {
	r.rateLimits.WithLabelValues(endpoint).Inc()
}

// RecordError records an error occurrence.
func (r *PrometheusRecorder) RecordError(operation, errorType string) {
	r.errors.WithLabelValues(operation, errorType).Inc()
}

// RecordContextCancellation records a context cancellation event.
func (r *PrometheusRecorder) RecordContextCancellation(operation string) {
	r.contextCancellations.WithLabelValues(operation).Inc()
}

// Ensure PrometheusRecorder implements observability.MetricsRecorder.
var _ observability.MetricsRecorder = (*PrometheusRecorder)(nil)
