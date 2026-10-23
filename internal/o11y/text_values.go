package o11y

import (
	"context"
	"encoding"
	"fmt"
	"log/slog"
	"reflect"
)

// textValuesHandler hands the exported record a value's text where the slog
// bridge would hand it the value's memory.
//
// The OpenTelemetry bridge converts an attribute of kind Any by reflection,
// and has no case for a value that knows its own text: a uuid.UUID, a
// [16]byte, was exported as an array of sixteen integers. stdout, whose text
// and JSON handlers ask a value for its text (encoding.TextMarshaler), wrote
// the id; the log store held numbers, and a filter on `idp_id` or `user_id`
// found nothing. So every attribute whose value has a text form is given as
// that text before the bridge sees it -- at every call site, the ones
// written after this included, which a guard over the call sites would only
// have asked each of them to remember.
type textValuesHandler struct {
	slog.Handler
}

// Handle implements slog.Handler.
func (h textValuesHandler) Handle(ctx context.Context, r slog.Record) error {
	out := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	r.Attrs(func(a slog.Attr) bool {
		out.AddAttrs(asText(a))

		return true
	})

	return h.Handler.Handle(ctx, out)
}

// WithAttrs implements slog.Handler: attributes bound to a logger are given
// as text too.
func (h textValuesHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	converted := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		converted[i] = asText(a)
	}

	return textValuesHandler{Handler: h.Handler.WithAttrs(converted)}
}

// WithGroup implements slog.Handler.
func (h textValuesHandler) WithGroup(name string) slog.Handler {
	return textValuesHandler{Handler: h.Handler.WithGroup(name)}
}

// asText is a with its value, and the values of a group's members, given as
// the bridge can carry them (exportable).
func asText(a slog.Attr) slog.Attr {
	value := a.Value.Resolve()

	switch value.Kind() {
	case slog.KindGroup:
		members := value.Group()
		converted := make([]slog.Attr, len(members))

		for i, member := range members {
			converted[i] = asText(member)
		}

		return slog.Attr{Key: a.Key, Value: slog.GroupValue(converted...)}
	case slog.KindAny:
		if exported, ok := exportable(value.Any()); ok {
			return slog.Attr{Key: a.Key, Value: exported}
		}
	}

	return slog.Attr{Key: a.Key, Value: value}
}

// exportable is v as a value the bridge carries as it reads, when the
// bridge would not:
//
//   - a value with a text form (encoding.TextMarshaler), a uuid.UUID among
//     them, as that text: the bridge reflects a [16]byte into sixteen
//     integers;
//   - a value of a named scalar type (domain.ResourcesLimitsScopeType, a
//     named string) as its String(), or else as its underlying value: the
//     bridge matches the built-in types exactly and wrote such a value as
//     "unhandled: (domain.ResourcesLimitsScopeType) system";
//   - a slice or an array of values with a text form, as their texts.
//
// An error is left to the bridge, which makes it the record's error. So is a
// value whose text cannot be read (see readText).
func exportable(v any) (slog.Value, bool) {
	if _, isError := v.(error); isError || v == nil {
		return slog.Value{}, false
	}

	if marshaler, ok := v.(encoding.TextMarshaler); ok {
		text, read := readText(v, func() (string, error) {
			raw, err := marshaler.MarshalText()

			return string(raw), err
		})

		return slog.StringValue(text), read
	}

	rv := reflect.ValueOf(v)

	if rv.Type().PkgPath() != "" {
		if scalar, ok := namedScalar(v, rv); ok {
			return scalar, true
		}
	}

	if (rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array) && rv.Type().Elem().Implements(textMarshalerType) {
		texts := make([]string, rv.Len())

		for i := range rv.Len() {
			element := rv.Index(i).Interface()

			text, read := readText(element, func() (string, error) {
				raw, err := element.(encoding.TextMarshaler).MarshalText()

				return string(raw), err
			})
			if !read {
				return slog.Value{}, false
			}

			texts[i] = text
		}

		return slog.AnyValue(texts), true
	}

	return slog.Value{}, false
}

var textMarshalerType = reflect.TypeFor[encoding.TextMarshaler]()

// namedScalar is a value of a named type whose kind is a scalar, as its
// String() or its underlying value.
func namedScalar(v any, rv reflect.Value) (slog.Value, bool) {
	if stringer, ok := v.(fmt.Stringer); ok {
		text, read := readText(v, func() (string, error) { return stringer.String(), nil })

		return slog.StringValue(text), read
	}

	switch rv.Kind() {
	case reflect.String:
		return slog.StringValue(rv.String()), true
	case reflect.Bool:
		return slog.BoolValue(rv.Bool()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return slog.Int64Value(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return slog.Uint64Value(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return slog.Float64Value(rv.Float()), true
	}

	return slog.Value{}, false
}

// readText is what read answers about v, when it can be read. A nil pointer
// is left to the bridge, as slog's own handlers leave it: an optional id
// that was not set (a nil *uuid.UUID) is logged all the time, and its value
// receiver panicked on the nil pointer inside the request that logged it,
// which Recovery then answered 500. A read that panics or fails for any
// other reason leaves the value as it was too: a line of the log is never
// worth a request.
func readText(v any, read func() (string, error)) (text string, ok bool) {
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Pointer && rv.IsNil() {
		return "", false
	}

	defer func() {
		if recover() != nil {
			text, ok = "", false
		}
	}()

	text, err := read()
	if err != nil {
		return "", false
	}

	return text, true
}
