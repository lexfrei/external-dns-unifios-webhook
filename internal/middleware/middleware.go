package middleware

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const requestIDHeader = "X-Request-ID"

// RequestID adds a unique request ID to each request.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Check if request ID exists in header
			rid := req.Header.Get(requestIDHeader)
			if rid == "" {
				rid = uuid.New().String()
			}

			// Set request ID in response header
			res.Header().Set(requestIDHeader, rid)

			// Store request ID in context
			c.Set("request_id", rid)

			return next(c)
		}
	}
}

// Logger creates a structured logging middleware using slog.
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()
			res := c.Response()

			// Get request ID from context
			requestID, _ := c.Get("request_id").(string)

			// Process request
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			// Log request
			latency := time.Since(start)
			status := res.Status
			method := req.Method
			path := req.URL.Path

			attrs := []any{
				"request_id", requestID,
				"method", method,
				"path", path,
				"status", status,
				"latency_ms", latency.Milliseconds(),
				"remote_ip", c.RealIP(),
				"user_agent", req.UserAgent(),
			}

			if err != nil {
				attrs = append(attrs, "error", err.Error())
				slog.ErrorContext(req.Context(), "request completed with error", attrs...)
			} else {
				slog.InfoContext(req.Context(), "request completed", attrs...)
			}

			return err
		}
	}
}

// Recovery returns a middleware which recovers from panics and logs the error.
func Recovery() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			requestID, _ := c.Get("request_id").(string)
			slog.ErrorContext(c.Request().Context(), "panic recovered",
				"request_id", requestID,
				"error", err.Error(),
				"stack", string(stack))
			return err
		},
	})
}

// Metrics returns a middleware that records metrics for each request.
func Metrics() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)

			// Record metrics
			duration := time.Since(start)
			status := c.Response().Status
			method := c.Request().Method
			path := c.Path()

			// Metrics are now recorded by OpenTelemetry middleware (otelecho)
			slog.DebugContext(c.Request().Context(), "request metrics",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds())

			return err
		}
	}
}

// WebhookContentType sets the correct content-type for external-dns webhook protocol.
// External-DNS expects application/external.dns.webhook+json;version=1 content type.
func WebhookContentType() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Set the webhook protocol content-type before processing the request
			c.Response().Header().Set(echo.HeaderContentType, "application/external.dns.webhook+json;version=1")
			return next(c)
		}
	}
}
