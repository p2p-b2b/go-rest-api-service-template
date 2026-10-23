package notifieremail

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/slashdevops/mailer"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/slashdevops/go-rest-api-service-template/internal/o11y"
)

// fakeMailer is a mailer.MailerService whose Send returns a fixed error.
type fakeMailer struct{ err error }

func (f fakeMailer) Send(_ context.Context, _ mailer.MailContent) error { return f.err }

// counterValue sums the email_send_total data points whose result label
// matches want, across the collected metrics.
func counterValue(t *testing.T, rm metricdata.ResourceMetrics, name, want string) int64 {
	t.Helper()
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("metric %s is not an int64 Sum", name)
			}
			for _, dp := range sum.DataPoints {
				if v, present := dp.Attributes.Value("result"); present && v.AsString() == want {
					total += dp.Value
				}
			}
		}
	}
	return total
}

func TestInstrumentedMailer_Send_recordsResult(t *testing.T) {
	tests := []struct {
		name    string
		sendErr error
		want    string
	}{
		{"ok", nil, resultOK},
		{"error", errors.New("smtp down"), resultError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := metric.NewManualReader()
			mp := metric.NewMeterProvider(metric.WithReader(reader))
			meter := mp.Meter("test")

			m := NewInstrumentedMailer(fakeMailer{err: tc.sendErr}, meter)
			err := m.Send(context.Background(), mailer.MailContent{})

			if (err != nil) != (tc.sendErr != nil) {
				t.Fatalf("Send() error = %v, want error presence %v", err, tc.sendErr != nil)
			}

			var rm metricdata.ResourceMetrics
			if err := reader.Collect(context.Background(), &rm); err != nil {
				t.Fatalf("Collect: %v", err)
			}
			if got := counterValue(t, rm, metricEmailSend, tc.want); got != 1 {
				t.Errorf("email_send_total{result=%q} = %d, want 1", tc.want, got)
			}
		})
	}
}

func TestNewInstrumentedMailer_nilMeter_returnsInner(t *testing.T) {
	inner := fakeMailer{}
	if got := NewInstrumentedMailer(inner, nil); got != mailer.MailerService(inner) {
		t.Error("with nil meter, expected the inner mailer returned unchanged")
	}
}

// TestTheMailLayerSaysHowACallEnded: mail mirrors its operations onto the
// shared layer instruments itself, so it says the outcome itself; without
// it no error ratio could see a mail transport that was failing.
func TestTheMailLayerSaysHowACallEnded(t *testing.T) {
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))
	t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })

	layer, err := o11y.NewLayerMetrics(provider.Meter("test"), "")
	if err != nil {
		t.Fatalf("NewLayerMetrics: %v", err)
	}

	recordLayer(context.Background(), layer, "send", "verification", resultOK, time.Millisecond)
	recordLayer(context.Background(), layer, "send", "verification", resultError, time.Millisecond)
	recordLayer(context.Background(), layer, "send", "verification", resultError, time.Millisecond)

	var collected metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &collected); err != nil {
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
				outcome, _ := point.Attributes.Value(o11y.AttrOutcome)
				got[outcome.AsString()] += point.Value
			}
		}
	}

	if got[o11y.OutcomeOK] != 1 || got[o11y.OutcomeFailed] != 2 || len(got) != 2 {
		t.Errorf("outcomes %v, want ok=1 failed=2", got)
	}
}
