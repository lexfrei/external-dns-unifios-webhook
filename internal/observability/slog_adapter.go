package observability

import (
	"log/slog"

	"github.com/lexfrei/go-unifi/observability"
)

// SlogAdapter implements observability.Logger using Go's structured logger (slog).
type SlogAdapter struct {
	logger *slog.Logger
}

// NewSlogAdapter creates a new SlogAdapter that delegates to the given slog.Logger.
func NewSlogAdapter(logger *slog.Logger) *SlogAdapter {
	return &SlogAdapter{logger: logger}
}

// Debug logs a debug-level message with optional structured fields.
func (a *SlogAdapter) Debug(msg string, fields ...observability.Field) {
	a.logger.Debug(msg, a.convertFields(fields)...)
}

// Info logs an info-level message with optional structured fields.
func (a *SlogAdapter) Info(msg string, fields ...observability.Field) {
	a.logger.Info(msg, a.convertFields(fields)...)
}

// Warn logs a warning-level message with optional structured fields.
func (a *SlogAdapter) Warn(msg string, fields ...observability.Field) {
	a.logger.Warn(msg, a.convertFields(fields)...)
}

// Error logs an error-level message with optional structured fields.
func (a *SlogAdapter) Error(msg string, fields ...observability.Field) {
	a.logger.Error(msg, a.convertFields(fields)...)
}

// With returns a new logger with the given fields pre-populated.
//
//nolint:ireturn // Interface return is required by observability.Logger contract
func (a *SlogAdapter) With(fields ...observability.Field) observability.Logger {
	return &SlogAdapter{
		logger: a.logger.With(a.convertFields(fields)...),
	}
}

// convertFields converts observability.Field slice to slog attributes (key-value pairs).
func (a *SlogAdapter) convertFields(fields []observability.Field) []any {
	args := make([]any, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key, f.Value)
	}

	return args
}
