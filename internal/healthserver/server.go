package healthserver

import (
	"encoding/json"
	"net/http"

	"github.com/lexfrei/external-dns-unifios-webhook/api/health"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/provider"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server implements the health.ServerInterface for health checks and metrics.
type Server struct {
	provider *provider.UniFiProvider
	registry *prometheus.Registry
}

// New creates a new health server instance with a custom Prometheus registry.
func New(prov *provider.UniFiProvider, registry *prometheus.Registry) *Server {
	return &Server{
		provider: prov,
		registry: registry,
	}
}

// Liveness returns OK if the service is alive.
// GET /healthz.
func (s *Server) Liveness(w http.ResponseWriter, _ *http.Request) {
	status := health.Ok
	message := "Service is alive"

	response := health.HealthStatus{
		Status:  &status,
		Message: &message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Readiness checks if the service is ready to accept traffic.
// GET /readyz.
func (s *Server) Readiness(w http.ResponseWriter, r *http.Request) {
	// Check if provider can connect to UniFi
	// We can do a simple health check by trying to list records
	_, err := s.provider.Records(r.Context())
	if err != nil {
		status := health.Error
		message := "Service is not ready: " + err.Error()

		response := health.HealthStatus{
			Status:  &status,
			Message: &message,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(response)

		return
	}

	status := health.Ok
	message := "Service is ready"

	response := health.HealthStatus{
		Status:  &status,
		Message: &message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// Metrics exports Prometheus metrics.
// GET /metrics.
func (s *Server) Metrics(w http.ResponseWriter, r *http.Request) {
	// Use custom prometheus registry with our metrics
	handler := promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, r)
}
