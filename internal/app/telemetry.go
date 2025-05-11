package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/o11y"
)

// initTelemetry initializes the observability components (metrics, tracing)
func (a *App) initTelemetry(ctx context.Context) error {
	var err error

	// Create OpenTelemetry instance
	a.telemetry, err = o11y.New(ctx, a.configs.Telemetry)
	if err != nil {
		return fmt.Errorf("error creating OpenTelemetry: %w", err)
	}

	// Start telemetry services
	if err := a.telemetry.Start(); err != nil {
		return fmt.Errorf("error starting telemetry: %w", err)
	}

	slog.Info("telemetry started successfully")
	return nil
}

// startPprofServer starts the pprof server for debugging if enabled
func (a *App) startPprofServer() {
	pprofAddr := fmt.Sprintf("%s:%d",
		a.configs.HTTPServer.PprofAddress.Value,
		a.configs.HTTPServer.PprofPort.Value,
	)

	pprofURL := fmt.Sprintf("http://%s/debug/pprof", pprofAddr)
	slog.Info("starting pprof server", "url", pprofURL)

	pprofRouter := http.NewServeMux()

	// Register pprof handlers
	pprofRouter.HandleFunc("/debug/pprof/", pprof.Index)
	pprofRouter.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofRouter.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofRouter.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofRouter.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Create the server
	a.pprofServer = &http.Server{
		Addr:    pprofAddr,
		Handler: pprofRouter,
	}

	if err := a.pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("pprof server error", "error", err)
	}
}
