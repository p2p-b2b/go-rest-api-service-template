package o11y

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"slices"
	"testing"
	"time"
	"uuid"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/attribute"
	sdkLog "go.opentelemetry.io/otel/sdk/log"
)

// attributeRecorder keeps the attributes of every record the bridge emits,
// as the exporter receives them.
type attributeRecorder struct {
	attrs map[string]attribute.Value
}

func (p *attributeRecorder) OnEmit(_ context.Context, r *sdkLog.Record) error {
	r.WalkAttributes(func(kv attribute.KeyValue) bool {
		p.attrs[string(kv.Key)] = kv.Value

		return true
	})

	return nil
}
func (p *attributeRecorder) Enabled(context.Context, sdkLog.EnabledParameters) bool { return true }
func (p *attributeRecorder) Shutdown(context.Context) error                         { return nil }
func (p *attributeRecorder) ForceFlush(context.Context) error                       { return nil }

// An id reaches the log store as the id. The bridge converts a value of kind
// Any by reflection and exported a uuid.UUID, a [16]byte, as sixteen
// integers; stdout, which asks a value for its text, wrote the id, so
// nothing local showed it, and a filter on the id in the log store found
// nothing.
func TestAnIDIsExportedAsItsText(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("01a16800-0000-7000-8000-000000000001")

	export := func(wrap bool) map[string]attribute.Value {
		rec := &attributeRecorder{attrs: map[string]attribute.Value{}}
		lp := sdkLog.NewLoggerProvider(sdkLog.WithProcessor(rec))

		var handler slog.Handler = otelslog.NewHandler("test", otelslog.WithLoggerProvider(lp))
		if wrap {
			handler = textValuesHandler{Handler: handler}
		}

		slog.New(handler).With("project_id", id).InfoContext(t.Context(), "provider identity linked",
			"idp_id", id, "account", slog.GroupValue(slog.Any("user_id", id)), "count", 3)

		return rec.attrs
	}

	// What the bridge does alone: the measurement this handler answers.
	if got := export(false)["idp_id"]; got.Type() == attribute.STRING {
		t.Fatalf("the bridge alone exported idp_id as a string (%q); the premise of this test changed", got.AsString())
	}

	attrs := export(true)

	for _, key := range []string{"idp_id", "project_id"} {
		if got := attrs[key]; got.Type() != attribute.STRING || got.AsString() != id.String() {
			t.Errorf("%s exported as %s %v, want the id's text", key, got.Type(), got.String())
		}
	}

	var member attribute.Value
	for _, kv := range attrs["account"].AsMap() {
		if kv.Key == "user_id" {
			member = kv.Value
		}
	}

	if member.Type() != attribute.STRING || member.AsString() != id.String() {
		t.Errorf("a group's member exported as %s %v, want the id's text", member.Type(), member.String())
	}

	if got := attrs["count"]; got.Type() != attribute.INT64 || got.AsInt64() != 3 {
		t.Errorf("a number was changed: %s %v", got.Type(), got.String())
	}
}

// The pipeline the service runs hands its records through the handler above:
// a test of the handler alone would pass with the pipeline built without it.
func TestTheExportedLogIsWrappedInTheTextHandler(t *testing.T) {
	t.Parallel()

	// Not t.Context(): the cleanup's Shutdown runs after it is cancelled,
	// and the batch processor then reports the cancel as a failed shutdown.
	logger := NewOpenTelemetryLogger(context.Background(), &OpenTelemetryLoggerConfig{
		Name:          "test",
		LogExporter:   "console",
		Level:         slog.LevelInfo,
		ConsoleWriter: io.Discard,
	})

	if err := logger.SetupLogs(); err != nil {
		t.Fatalf("SetupLogs: %v", err)
	}

	t.Cleanup(logger.Shutdown)

	if _, ok := logger.Handler.(textValuesHandler); !ok {
		t.Errorf("the exported log's handler is %T, want it wrapped in textValuesHandler", logger.Handler)
	}
}

// panickingText's MarshalText panics: a value of someone else's type.
type panickingText struct{}

func (panickingText) MarshalText() ([]byte, error) { panic("not today") }

// A value whose text cannot be read is left as it is, and the line is
// written: an optional id that was not set, a nil *uuid.UUID, is logged all
// the time, and its MarshalText panicked inside the request that logged it,
// which was then answered 500 (63 tests of the integration suite).
func TestAValueWithoutReadableTextIsLeftAsItIs(t *testing.T) {
	t.Parallel()

	rec := &attributeRecorder{attrs: map[string]attribute.Value{}}
	lp := sdkLog.NewLoggerProvider(sdkLog.WithProcessor(rec))
	logger := slog.New(textValuesHandler{Handler: otelslog.NewHandler("test", otelslog.WithLoggerProvider(lp))})

	var unset *uuid.UUID

	logger.InfoContext(t.Context(), "an optional id that was not set", "project_id", unset, "other", panickingText{}, "count", 3)

	if got := rec.attrs["count"]; got.Type() != attribute.INT64 || got.AsInt64() != 3 {
		t.Fatalf("the line was not written whole: %v", rec.attrs)
	}

	if got := rec.attrs["project_id"]; got.Type() == attribute.STRING {
		t.Errorf("a nil id was given a text: %q", got.AsString())
	}
}

// textError is an error of a named string type that also has a text form:
// it is still the bridge's, which makes it the record's error.
type textError string

func (e textError) Error() string                { return string(e) }
func (e textError) MarshalText() ([]byte, error) { return []byte(e), nil }

// scopeKind is a named scalar type, as domain.ResourcesLimitsScopeType is.
type scopeKind string

// What the bridge matches only by exact built-in type is given as it reads:
// a named string type, which the bridge wrote as "unhandled: (...) system",
// and a slice of ids, which it wrote as arrays of integers. An error stays
// the bridge's, which makes it the record's error.
func TestWhatTheBridgeCannotReadIsGivenAsItReads(t *testing.T) {
	t.Parallel()

	first := uuid.MustParse("01a16800-0000-7000-8000-000000000001")
	second := uuid.MustParse("01a16800-0000-7000-8000-000000000002")

	rec := &attributeRecorder{attrs: map[string]attribute.Value{}}
	lp := sdkLog.NewLoggerProvider(sdkLog.WithProcessor(rec))
	logger := slog.New(textValuesHandler{Handler: otelslog.NewHandler("test", otelslog.WithLoggerProvider(lp))})

	var unsetAt *time.Time

	logger.InfoContext(t.Context(), "usage checked",
		"scope_type", scopeKind("system"), "user_ids", []uuid.UUID{first, second}, "error", errors.New("the database is gone"), "at", unsetAt)

	if got := rec.attrs["scope_type"]; got.Type() != attribute.STRING || got.AsString() != "system" {
		t.Errorf("a named string exported as %s %q, want its value", got.Type(), got.String())
	}

	var texts []string
	for _, element := range rec.attrs["user_ids"].AsSlice() {
		texts = append(texts, element.AsString())
	}

	if !slices.Equal(texts, []string{first.String(), second.String()}) {
		t.Errorf("a slice of ids exported as %s, want their texts", rec.attrs["user_ids"].String())
	}

	// The bridge makes an error the record's error; it is not a value to
	// give as text.
	if _, given := exportable(errors.New("the database is gone")); given {
		t.Error("an error was taken from the bridge")
	}

	if _, given := exportable(textError("a named error with a text form")); given {
		t.Error("an error with a text form was taken from the bridge")
	}

	if got := rec.attrs["at"]; got.Type() == attribute.STRING {
		t.Errorf("a nil *time.Time was given a text: %q", got.AsString())
	}
}
