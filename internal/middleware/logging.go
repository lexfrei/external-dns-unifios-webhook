package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type responseWriter struct {
	http.ResponseWriter

	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"method", r.Method,
					"path", r.URL.Path)
				http.Error(wrapped, "Internal Server Error", http.StatusInternalServerError)

				return
			}

			level := slog.LevelInfo
			if wrapped.status >= http.StatusInternalServerError {
				level = slog.LevelError
			} else if wrapped.status >= http.StatusBadRequest {
				level = slog.LevelWarn
			}

			slog.Log(r.Context(), level, "request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.status,
				"duration", time.Since(start))
		}()

		next.ServeHTTP(wrapped, r)
	})
}
