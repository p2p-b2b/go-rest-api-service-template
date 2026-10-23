//go:build unit

package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/slashdevops/go-rest-api-service-template/internal/config"
	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// A replica that cannot serve stops with one line at ERROR, and that line
// leaves the process before it ends: the exporter batches, and main exits
// straight after. log.Fatalf wrote the error at INFO and exited with the
// batch unsent, so the replica left nothing in the log store, its start
// line included. Read on the exported side only after stopOnError returns,
// with nothing else flushing it.
// appWithExportedLog is an App whose log is exported to exported through the
// console exporter, batched for an hour: only a shutdown sends a record
// there in a test's time.
func appWithExportedLog(t *testing.T, stdout, exported *bytes.Buffer) *App {
	t.Helper()

	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })

	slog.SetDefault(slog.New(slog.NewJSONHandler(stdout, nil)))

	a := &App{
		configs: &Configs{
			Telemetry: config.NewOpenTelemetryConfig("test-app", "0.0.0-test"),
			Log:       config.NewLogConfig(),
		},
	}
	a.configs.Telemetry.TraceExporter.Value = config.ExporterNoop
	a.configs.Telemetry.MetricExporter.Value = config.ExporterNoop
	a.configs.Telemetry.LogExporter.Value = "console"

	telemetry, err := o11y.New(t.Context(), a.configs.Telemetry, a.configs.Log)
	if err != nil {
		t.Fatalf("o11y.New: %v", err)
	}

	telemetry.Logs = o11y.NewOpenTelemetryLogger(t.Context(), &o11y.OpenTelemetryLoggerConfig{
		Name:                    "test-app",
		LogExporter:             "console",
		Level:                   slog.LevelInfo,
		ConsoleWriter:           exported,
		LogExporterBatchTimeout: time.Hour,
	})

	if err := telemetry.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}

	a.telemetry = telemetry
	a.composeLogger()

	return a
}

// A replica that cannot serve stops with one line at ERROR, and that line
// leaves the process before it ends: the exporter batches, and main exits
// straight after. log.Fatalf wrote the error at INFO and exited with the
// batch unsent, so the replica left nothing in the log store, its start
// line included. Read on the exported side only after stopOnError returns,
// with nothing else flushing it.
func TestAReplicaThatCannotServeSaysSoBeforeItStops(t *testing.T) {
	var stdout, exported bytes.Buffer

	a := appWithExportedLog(t, &stdout, &exported)

	cause := errors.New("could not load the revoked access token denylist")

	if got := a.stopOnError(t.Context(), cause); !errors.Is(got, cause) {
		t.Fatalf("stopOnError returned %v, want the cause for main's exit status", got)
	}

	out := exported.String()
	if !strings.Contains(out, stoppedOnAnError) || !strings.Contains(out, cause.Error()) {
		t.Fatalf("the line did not leave the process before the stop: exported %q", out)
	}

	if !strings.Contains(out, "ERROR") {
		t.Errorf("the line left the process, not at ERROR: %s", out)
	}

	if !strings.Contains(stdout.String(), `"level":"ERROR"`) {
		t.Errorf("the standard sink did not get the line at ERROR: %s", stdout.String())
	}
}

// A signal while a first load runs cancels the load, and its error is the
// cancellation: a rollout stopping a replica that was still starting. It is
// the shutdown's INFO line, flushed, and no error for main to exit 1 on.
func TestASignalWhileStartingIsNotAnError(t *testing.T) {
	var stdout, exported bytes.Buffer

	a := appWithExportedLog(t, &stdout, &exported)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if got := a.stopOnError(ctx, fmt.Errorf("could not load the rate limit rules: %w", ctx.Err())); got != nil {
		t.Errorf("stopOnError returned %v for a signal, want nil: main exits 1 on it", got)
	}

	out := exported.String()
	if strings.Contains(out, stoppedOnAnError) || strings.Contains(stdout.String(), `"level":"ERROR"`) {
		t.Errorf("a signal while starting was written as a stop on an error: exported %q, stdout %q", out, stdout.String())
	}

	if !strings.Contains(out, "received shutdown signal") {
		t.Errorf("the shutdown's line did not leave the process: %q", out)
	}
}

// Every failure of the builder after the telemetry phase goes through stop,
// so that it is written at ERROR and flushed: a phase added later with a
// bare return would leave nothing in the log store again.
func TestEveryFailureAfterTheTelemetryIsWrittenAndFlushed(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("builder.go")
	if err != nil {
		t.Fatal(err)
	}

	_, after, found := strings.Cut(string(b), "stop := func(err error)")
	if !found {
		t.Fatal("builder.go has no stop after the telemetry phase")
	}

	if strings.Contains(after, "return nil, err") {
		t.Error("Build returns a failure after the telemetry phase without stop(err): it is not written at ERROR or flushed")
	}

	if strings.Count(after, "return stop(err)") < 7 {
		t.Errorf("%d failures go through stop(err), want the seven phases after the telemetry", strings.Count(after, "return stop(err)"))
	}
}
