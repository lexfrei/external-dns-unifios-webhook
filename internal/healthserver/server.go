package healthserver

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/lexfrei/external-dns-unifios-webhook/api/health"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/metrics"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/provider"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/singleflight"
)

// readinessCache stores the cached result of readiness checks.
type readinessCache struct {
	isReady   bool
	checkedAt time.Time
	mu        sync.RWMutex
}

const readinessCacheTTL = 30 * time.Second

// Server implements the health.ServerInterface for health checks and metrics.
type Server struct {
	provider       *provider.UniFiProvider
	registry       *prometheus.Registry
	readinessCache *readinessCache
	checkGroup     singleflight.Group
}

// New creates a new health server instance with a custom Prometheus registry.
func New(prov *provider.UniFiProvider, registry *prometheus.Registry) *Server {
	return &Server{
		provider: prov,
		registry: registry,
		readinessCache: &readinessCache{
			isReady:   false,
			checkedAt: time.Time{}, // Zero value means cache is cold
		},
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
// Uses caching with 30 second TTL to avoid excessive UniFi API calls.
// Concurrent requests are deduplicated using singleflight to prevent thundering herd.
func (s *Server) Readiness(w http.ResponseWriter, r *http.Request) {
	// Try to use cached result if fresh enough
	if cached, ok := s.getCachedReadiness(); ok {
		s.writeReadinessResponse(w, cached, "")

		return
	}

	// Use singleflight to deduplicate concurrent checks
	// This prevents multiple simultaneous requests from all hitting the UniFi API
	result, _, _ := s.checkGroup.Do("readiness", func() (any, error) {
		isReady := s.checkProviderReadiness(r)

		return isReady, nil
	})

	isReady, ok := result.(bool)
	if !ok {
		// This should never happen, but handle gracefully
		s.writeReadinessResponse(w, false, "")

		return
	}

	s.writeReadinessResponse(w, isReady, "")
}

// Metrics exports Prometheus metrics.
// GET /metrics.
func (s *Server) Metrics(w http.ResponseWriter, r *http.Request) {
	// Use custom prometheus registry with our metrics
	handler := promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, r)
}

// getCachedReadiness returns cached readiness status if cache is still fresh.
func (s *Server) getCachedReadiness() (isReady, ok bool) {
	s.readinessCache.mu.RLock()
	defer s.readinessCache.mu.RUnlock()

	cacheAge := time.Since(s.readinessCache.checkedAt)
	if cacheAge < readinessCacheTTL {
		metrics.ReadinessCacheHits.Inc()
		metrics.ReadinessCacheAge.Set(cacheAge.Seconds())

		return s.readinessCache.isReady, true
	}

	metrics.ReadinessCacheMisses.Inc()

	return false, false
}

// checkProviderReadiness performs actual readiness check and updates cache.
func (s *Server) checkProviderReadiness(r *http.Request) bool {
	_, err := s.provider.Records(r.Context())

	// Update cache with new result
	s.readinessCache.mu.Lock()
	s.readinessCache.isReady = (err == nil)
	s.readinessCache.checkedAt = time.Now()
	s.readinessCache.mu.Unlock()

	return err == nil
}

// writeReadinessResponse writes readiness check response.
func (s *Server) writeReadinessResponse(w http.ResponseWriter, isReady bool, suffix string) {
	var statusCode int

	var statusValue health.HealthStatusStatus

	var message string

	if isReady {
		statusCode = http.StatusOK
		statusValue = health.Ok

		message = "Service is ready"
		if suffix != "" {
			message += " " + suffix
		}
	} else {
		statusCode = http.StatusServiceUnavailable
		statusValue = health.Error

		message = "Service is not ready"
		if suffix != "" {
			message += " " + suffix
		}
	}

	response := health.HealthStatus{
		Status:  &statusValue,
		Message: &message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}
