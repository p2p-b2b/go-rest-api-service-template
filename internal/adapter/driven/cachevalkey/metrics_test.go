//go:build unit

package cachevalkey

import (
	"testing"
	"time"

	"github.com/slashdevops/c3e"
	sdkMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// TestTheCacheLayerSaysHowACallEnded: the cache mirrors its operations onto
// the shared layer instruments itself, not through o11y's recorder, so it
// has to say the outcome itself. Without it the cache layer's calls carried
// no app.outcome and no error ratio could see a cache that was failing.
func TestTheCacheLayerSaysHowACallEnded(t *testing.T) {
	t.Parallel()

	reader := sdkMetric.NewManualReader()
	provider := sdkMetric.NewMeterProvider(sdkMetric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

	instruments, err := NewInstruments(provider.Meter("test"))
	if err != nil {
		t.Fatalf("NewInstruments: %v", err)
	}

	instruments.recordLayer(t.Context(), "get", "users", resultOK, time.Millisecond)
	instruments.recordLayer(t.Context(), "get", "users", "miss", time.Millisecond)
	instruments.recordLayer(t.Context(), "get", "users", resultError, time.Millisecond)
	// A read past the query timeout falls back to the database: the request
	// is answered, and the cache did not work.
	instruments.recordLayer(t.Context(), "get", "users", string(c3e.ResultTimeout), time.Millisecond)

	var collected metricdata.ResourceMetrics
	if err := reader.Collect(t.Context(), &collected); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	got := map[string]int64{}

	for _, scope := range collected.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != o11y.MetricCalls {
				continue
			}

			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("%s is not an int64 sum", m.Name)
			}

			for _, point := range sum.DataPoints {
				layer, _ := point.Attributes.Value(o11y.AttrLayer)
				if layer.AsString() != o11y.LayerCache {
					t.Errorf("layer %q on a cache call", layer.AsString())
				}

				outcome, _ := point.Attributes.Value(o11y.AttrOutcome)
				got[outcome.AsString()] += point.Value
			}
		}
	}

	// A miss is the cache working: an error and a timeout are failures, and
	// nothing a caller sent reaches the cache as a request it can refuse.
	if got[o11y.OutcomeOK] != 2 || got[o11y.OutcomeFailed] != 2 || len(got) != 2 {
		t.Errorf("outcomes %v, want ok=2 failed=2", got)
	}
}
