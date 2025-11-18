package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof" // Register pprof handlers
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/lexfrei/external-dns-unifios-webhook/api/health"
	"github.com/lexfrei/external-dns-unifios-webhook/api/webhook"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/config"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/healthserver"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/metrics"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/observability"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/provider"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/webhookserver"
	unifi "github.com/lexfrei/go-unifi/api/network"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/external-dns/endpoint"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	// Setup logging
	setupLogging(cfg.Logging)

	slog.Info("starting external-dns-unifios-webhook",
		"unifi_host", cfg.UniFi.Host,
		"unifi_site", cfg.UniFi.Site,
		"webhook_addr", joinHostPort(cfg.Server.Host, cfg.Server.Port),
		"health_addr", joinHostPort(cfg.Health.Host, cfg.Health.Port))

	// Create Prometheus registry for custom metrics
	registry := prometheus.NewRegistry()

	// Register custom DNS metrics
	metrics.Register(registry)

	// Create observability components
	logger := observability.NewSlogAdapter(slog.Default())
	metricsRecorder := observability.NewPrometheusRecorder(registry, "external_dns_unifi")

	// Create domain filter
	domainFilter := endpoint.NewDomainFilterWithExclusions(
		cfg.DomainFilter.Filters,
		cfg.DomainFilter.ExcludeFilters,
	)

	// Create UniFi API client
	client, err := unifi.NewWithConfig(&unifi.ClientConfig{
		ControllerURL:      cfg.UniFi.Host,
		APIKey:             cfg.UniFi.APIKey,
		InsecureSkipVerify: cfg.UniFi.SkipTLSVerify,
		Logger:             logger,
		Metrics:            metricsRecorder,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create UniFi API client")
	}

	if cfg.UniFi.SkipTLSVerify {
		slog.Warn("TLS certificate verification is disabled")
	}

	// Create UniFi provider with dependency injection
	prov := provider.New(client, cfg.UniFi.Site, *domainFilter)

	// Create webhook server
	webhookSrv := webhookserver.New(prov, *domainFilter)
	webhookRouter := chi.NewRouter()
	webhookRouter.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{}))
	webhookRouter.Use(middleware.Recoverer)

	// Custom error handler with detailed logging
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		slog.ErrorContext(r.Context(), "webhook request validation failed",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
			"content_type", r.Header.Get("Content-Type"),
			"accept", r.Header.Get("Accept"))
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	webhook.HandlerWithOptions(webhookSrv, webhook.ChiServerOptions{
		BaseRouter:       webhookRouter,
		ErrorHandlerFunc: errorHandler,
	})

	webhookHTTPServer := &http.Server{
		Addr:              joinHostPort(cfg.Server.Host, cfg.Server.Port),
		Handler:           webhookRouter,
		ReadTimeout:       30 * time.Second,  // Time to read request headers and body
		ReadHeaderTimeout: 5 * time.Second,   // Time to read request headers only
		WriteTimeout:      60 * time.Second,  // Time to write response (allows batch operations)
		IdleTimeout:       120 * time.Second, // Keep-alive timeout
		MaxHeaderBytes:    64 << 10,          // 64 KB (headers are small, body is in request body)
	}

	// Create health server with custom registry
	healthSrv := healthserver.New(prov, registry)
	healthRouter := chi.NewRouter()
	healthRouter.Use(httplog.RequestLogger(slog.Default(), &httplog.Options{}))
	healthRouter.Use(middleware.Recoverer)
	health.HandlerFromMux(healthSrv, healthRouter)

	healthHTTPServer := &http.Server{
		Addr:              joinHostPort(cfg.Health.Host, cfg.Health.Port),
		Handler:           healthRouter,
		ReadTimeout:       5 * time.Second,  // Health checks are quick
		ReadHeaderTimeout: 2 * time.Second,  // Headers should arrive fast
		WriteTimeout:      10 * time.Second, // Response writing timeout
		IdleTimeout:       30 * time.Second, // Shorter idle for health endpoint
		MaxHeaderBytes:    64 << 10,         // 64 KB (health checks have small headers)
	}

	// Start pprof debug server if enabled
	// pprofHTTPServer is declared outside the block to allow graceful shutdown
	var pprofHTTPServer *http.Server
	if cfg.Debug.PprofEnabled {
		slog.Warn("pprof profiling enabled - DO NOT use in production",
			"port", cfg.Debug.PprofPort,
			"endpoints", []string{
				fmt.Sprintf("http://localhost:%s/debug/pprof/", cfg.Debug.PprofPort),
				fmt.Sprintf("http://localhost:%s/debug/pprof/heap", cfg.Debug.PprofPort),
				fmt.Sprintf("http://localhost:%s/debug/pprof/goroutine", cfg.Debug.PprofPort),
			})

		pprofHTTPServer = &http.Server{
			Addr:              "127.0.0.1:" + cfg.Debug.PprofPort, // Bind to localhost only for security
			Handler:           http.DefaultServeMux,               // pprof handlers
			ReadTimeout:       30 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      30 * time.Second,
		}

		go func() {
			if err := pprofHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("pprof server error", "error", err)
			}
		}()
	}

	// Start servers
	errChan := make(chan error, 2)

	// Start webhook server
	go func() {
		slog.Info("starting webhook server", "address", webhookHTTPServer.Addr)
		if err := webhookHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- errors.Wrap(err, "webhook server error")
		}
	}()

	// Start health server
	go func() {
		slog.Info("starting health server", "address", healthHTTPServer.Addr)
		if err := healthHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- errors.Wrap(err, "health server error")
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case sig := <-sigChan:
		slog.Info("received shutdown signal", "signal", sig.String())
		cancel()

		// Shutdown servers with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := webhookHTTPServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("webhook server shutdown error", "error", err)
		}
		if err := healthHTTPServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("health server shutdown error", "error", err)
		}
		if pprofHTTPServer != nil {
			if err := pprofHTTPServer.Shutdown(shutdownCtx); err != nil {
				slog.Error("pprof server shutdown error", "error", err)
			}
		}
	}

	return nil
}

func setupLogging(cfg config.LoggingConfig) {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// joinHostPort efficiently concatenates host and port using strings.Builder.
// This is more efficient than fmt.Sprintf for simple string concatenation.
func joinHostPort(host, port string) string {
	var sb strings.Builder
	sb.Grow(len(host) + 1 + len(port)) // Pre-allocate: host + ":" + port
	sb.WriteString(host)
	sb.WriteByte(':')
	sb.WriteString(port)

	return sb.String()
}
