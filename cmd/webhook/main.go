package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lexfrei/external-dns-unifios-webhook/api/health"
	"github.com/lexfrei/external-dns-unifios-webhook/api/webhook"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/config"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/healthserver"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/provider"
	"github.com/lexfrei/external-dns-unifios-webhook/internal/webhookserver"
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
		"webhook_addr", fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		"health_addr", fmt.Sprintf("%s:%s", cfg.Health.Host, cfg.Health.Port))

	// Create domain filter
	domainFilter := endpoint.NewDomainFilterWithExclusions(
		cfg.DomainFilter.Filters,
		cfg.DomainFilter.ExcludeFilters,
	)

	// Create UniFi provider
	prov, err := provider.New(cfg.UniFi, *domainFilter)
	if err != nil {
		return errors.Wrap(err, "failed to create provider")
	}

	// Create webhook server
	webhookSrv := webhookserver.New(prov, *domainFilter)
	webhookRouter := chi.NewRouter()
	webhookRouter.Use(middleware.Logger)
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
		Addr:    fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler: webhookRouter,
	}

	// Create health server
	healthSrv := healthserver.New(prov)
	healthRouter := chi.NewRouter()
	healthRouter.Use(middleware.Logger)
	healthRouter.Use(middleware.Recoverer)
	health.HandlerFromMux(healthSrv, healthRouter)

	healthHTTPServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Health.Host, cfg.Health.Port),
		Handler: healthRouter,
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
