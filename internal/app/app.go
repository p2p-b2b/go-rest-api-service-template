package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/http/server"
	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
	"github.com/p2p-b2b/mailer"
)

const (
	appName    = "go-rest-api-service-template"
	apiVersion = "v1"
)

var apiPrefix = fmt.Sprintf("api/%s", apiVersion)

// App represents the application
type App struct {
	// Configuration
	configs *Configs

	// Core components
	telemetry *o11y.OpenTelemetry
	dbPool    *pgxpool.Pool

	// HTTP servers
	httpServer  *server.HTTPServer
	mailServer  *mailer.MailService
	pprofServer *http.Server

	// Services and repositories (could be further grouped by domain)
	repositories *Repositories
	services     *Services
	handlers     *Handlers

	// Lifecycle management
	shutdownCh   chan struct{}
	shutdownOnce sync.Once
}

// NewApp creates a new application instance
func NewApp(ctx context.Context) (*App, error) {
	app := &App{
		shutdownCh: make(chan struct{}),
	}

	var err error
	app.configs, err = LoadConfigs()
	if err != nil {
		return nil, err
	}

	// Initialize components
	if err := app.initTelemetry(ctx); err != nil {
		return nil, err
	}

	if err := app.initDatabase(ctx); err != nil {
		return nil, err
	}

	if err := app.initRepositories(); err != nil {
		return nil, err
	}

	// must be before initServices
	if err := app.initMailService(ctx); err != nil {
		return nil, err
	}

	if err := app.initServices(ctx); err != nil {
		return nil, err
	}

	if err := app.initHandlers(); err != nil {
		return nil, err
	}

	if err := app.initHTTPServer(ctx); err != nil {
		return nil, err
	}

	return app, nil
}

// Run starts the application
func (a *App) Run() error {
	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	go a.httpServer.Start()
	go a.mailServer.Start()

	// Start pprof server if enabled
	if a.configs.HTTPServer.PprofEnabled.Value {
		go a.startPprofServer()
	}

	// Wait for shutdown signal
	select {
	case <-sigCh:
		slog.Info("received shutdown signal")
	case <-a.shutdownCh:
		slog.Info("shutdown requested")
	}

	return a.Shutdown()
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown() error {
	var shutdownErr error

	a.shutdownOnce.Do(func() {
		// 1. Shutdown HTTP server with a timeout context for graceful shutdown
		slog.Info("shutting down HTTP server")

		// Setup a timeout context for shutdown operations
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Stop the HTTP server
		a.httpServer.Stop()

		// Wait for server to shut down completely with timeout
		select {
		case <-a.httpServer.Wait():
			slog.Info("HTTP server shut down successfully")
		case <-ctx.Done():
			slog.Warn("HTTP server shutdown timed out")
		}

		// 2. Shutdown pprof server if running
		if a.pprofServer != nil {
			slog.Info("shutting down pprof server")
			if err := a.pprofServer.Shutdown(context.Background()); err != nil {
				slog.Error("error shutting down pprof server", "error", err)
			}
		}

		// 3. Close database connection
		slog.Info("closing database connection")
		a.dbPool.Close()

		// Stop the mail service
		slog.Info("stopping mail service")
		a.mailServer.Stop()

		// 4. Shutdown telemetry
		slog.Info("shutting down telemetry")
		a.telemetry.Shutdown()

		close(a.shutdownCh)
	})

	return shutdownErr
}
