//go:build unit

package o11y

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"go.opentelemetry.io/otel/codes"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// recordedSpan runs one call that returns err through RecordError (or
// RecordSuccess for nil) and returns the span it ended.
func recordedSpan(t *testing.T, err error) sdkTrace.ReadOnlySpan {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	provider := sdkTrace.NewTracerProvider(sdkTrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

	ctx, span, attrs := SetupTrace(WithFailure(t.Context()), provider.Tracer("test"), Metadata{Layer: LayerRepository, Domain: "Things"}, "List")
	if err != nil {
		_ = RecordError(ctx, span, time.Now(), err, nil, attrs)
	} else {
		RecordSuccess(ctx, span, time.Now(), nil, attrs, "listed")
	}
	span.End()

	ended := recorder.Ended()
	if len(ended) != 1 {
		t.Fatalf("ended %d spans, want 1", len(ended))
	}

	return ended[0]
}

func spanOutcome(span sdkTrace.ReadOnlySpan) string {
	for _, attr := range span.Attributes() {
		if string(attr.Key) == AttrOutcome {
			return attr.Value.AsString()
		}
	}

	return ""
}

func eventNames(span sdkTrace.ReadOnlySpan) []string {
	names := make([]string, 0, len(span.Events()))
	for _, event := range span.Events() {
		names = append(names, event.Name)
	}

	return names
}

// TestARefusedCallIsNotARedSpan: every call that returned an error was
// marked Error, so a request answered 404 was a trace with a red span at
// every layer and an error search returned the requests clients got wrong.
// The status follows the outcome the metrics record.
func TestARefusedCallIsNotARedSpan(t *testing.T) {
	captureLogs(t)
	classifyRefusals(t)

	notFound := &errRefusal{what: "no such thing"}

	for name, err := range map[string]error{
		"plain":                    notFound,
		"wrapped_by_a_layer_above": fmt.Errorf("getting the thing: %w", notFound),
	} {
		t.Run(name, func(t *testing.T) {
			span := recordedSpan(t, err)

			if got := span.Status().Code; got != codes.Unset {
				t.Errorf("status %v for a refused call, want it left unset", got)
			}

			if got := spanOutcome(span); got != OutcomeRefused {
				t.Errorf("app.outcome %q on the span, want %q", got, OutcomeRefused)
			}

			// The refusal is on the span, as an event of its own: not as an
			// exception, which is what a fault is.
			events := span.Events()
			if len(events) != 1 || events[0].Name != EventRefused {
				t.Fatalf("events %v, want one %q", eventNames(span), EventRefused)
			}

			said := map[string]string{}
			for _, attr := range events[0].Attributes {
				said[string(attr.Key)] = attr.Value.AsString()
			}

			if said["error.type"] != "*o11y.errRefusal" || said["error.message"] != err.Error() {
				t.Errorf("the event says %v, want the error's type and text", said)
			}
		})
	}
}

// TestAFailedCallIsARedSpan: a fault is marked as it always was, with the
// error as an exception event.
func TestAFailedCallIsARedSpan(t *testing.T) {
	captureLogs(t)
	classifyRefusals(t)

	span := recordedSpan(t, errors.New("connection refused"))

	if got := span.Status(); got.Code != codes.Error || got.Description != "connection refused" {
		t.Errorf("status %v %q, want Error with the error's text", got.Code, got.Description)
	}

	if got := spanOutcome(span); got != OutcomeFailed {
		t.Errorf("app.outcome %q, want %q", got, OutcomeFailed)
	}

	if events := eventNames(span); len(events) != 1 || events[0] != "exception" {
		t.Errorf("events %v, want the exception", events)
	}
}

// TestACallThatWorkedIsOk, with its outcome on the span like the other two.
func TestACallThatWorkedIsOk(t *testing.T) {
	classifyRefusals(t)

	span := recordedSpan(t, nil)

	if got := span.Status().Code; got != codes.Ok {
		t.Errorf("status %v, want Ok", got)
	}

	if got := spanOutcome(span); got != OutcomeOK {
		t.Errorf("app.outcome %q, want %q", got, OutcomeOK)
	}

	if events := eventNames(span); len(events) != 0 {
		t.Errorf("events %v on a call that worked", events)
	}
}

// TestWithNoClassifierAnErrorIsARedSpan: nothing says the error is a
// refusal, so it is a failure, on the span as in the metric.
func TestWithNoClassifierAnErrorIsARedSpan(t *testing.T) {
	captureLogs(t)
	SetRefusalClassifier(nil)

	span := recordedSpan(t, &errRefusal{what: "no such thing"})

	if got := span.Status().Code; got != codes.Error {
		t.Errorf("status %v with no classifier, want Error", got)
	}

	if events := eventNames(span); len(events) != 1 || events[0] != "exception" {
		t.Errorf("events %v, want the exception", events)
	}
}
