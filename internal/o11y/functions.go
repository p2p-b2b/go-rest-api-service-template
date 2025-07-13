package o11y

import (
	"context"
	"log/slog"
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// RecordError is a utility function to record an error in the span, log it, and update metrics.
// It centralizes error handling across repositories to reduce code duplication.
//
// Parameters:
// - ctx: The context.Context
// - span: The OpenTelemetry trace span
// - err: The error to record
// - metricsCounter: The OpenTelemetry metrics counter to increment
// - metricAttrs: The metrics attributes to include
// - component: The component name (e.g., "repository.Users")
// - operation: The operation name (e.g., "Insert")
// - details: Optional additional details to log
//
// Returns the original error for chainable error handling.
func RecordError(
	ctx context.Context,
	span trace.Span,
	err error,
	metricsCounter metric.Int64Counter,
	metricAttrs []attribute.KeyValue,
	component string,
	details ...any,
) error {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	funcName := runtime.FuncForPC(pc).Name()
	if len(details) > 0 {
		slog.Error(component, append([]any{"error", err, "func", funcName, "file", file, "line", line}, details...)...)
	} else {
		slog.Error(component, "error", err, "func", funcName, "file", file, "line", line)
	}

	if metricsCounter != nil {
		metricsCounter.Add(ctx, 1,
			metric.WithAttributes(
				append(metricAttrs, attribute.String("successful", "false"))...,
			),
		)
	}

	return err
}

// RecordSuccess is a utility function to record a successful operation in the span and update metrics.
// It centralizes success handling across repositories to reduce code duplication.
//
// Parameters:
// - ctx: The context.Context
// - span: The OpenTelemetry trace span
// - metricsCounter: The OpenTelemetry metrics counter to increment
// - metricAttrs: The metrics attributes to include
// - message: The success message
// - attrs: Optional additional span attributes
func RecordSuccess(
	ctx context.Context,
	span trace.Span,
	metricsCounter metric.Int64Counter,
	metricAttrs []attribute.KeyValue,
	message string,
	attrs ...attribute.KeyValue,
) {
	span.SetStatus(codes.Ok, message)

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	if metricsCounter != nil {
		metricsCounter.Add(ctx, 1,
			metric.WithAttributes(
				append(metricAttrs, attribute.String("successful", "true"))...,
			),
		)
	}
}
