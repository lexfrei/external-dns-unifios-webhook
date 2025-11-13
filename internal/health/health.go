package health

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"sigs.k8s.io/external-dns/provider"
)

// Handler handles health check endpoints.
type Handler struct {
	provider provider.Provider
}

// New creates a new health check handler.
func New(prov provider.Provider) *Handler {
	return &Handler{
		provider: prov,
	}
}

// Liveness handles the liveness probe endpoint.
// Returns 200 OK if the application is running.
func (h *Handler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}

// Readiness handles the readiness probe endpoint.
// Returns 200 OK if the application is ready to serve traffic.
// Checks connectivity to UniFi API.
func (h *Handler) Readiness(c echo.Context) error {
	ctx := c.Request().Context()

	// Try to fetch records to verify UniFi API connectivity
	if err := h.checkUniFiConnectivity(ctx); err != nil {
		slog.ErrorContext(ctx, "readiness check failed", "error", err)
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"error":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// checkUniFiConnectivity verifies that the UniFi API is accessible.
func (h *Handler) checkUniFiConnectivity(ctx context.Context) error {
	// Try to get records as a health check
	_, err := h.provider.Records(ctx)
	return err
}
