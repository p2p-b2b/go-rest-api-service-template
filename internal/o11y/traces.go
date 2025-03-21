package o11y

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	otelTrace "go.opentelemetry.io/otel/trace"
)

// OpenTelemetryTracerConfig represents the configuration of the OpenTelemetry tracer.
type OpenTelemetryTracerConfig struct {
	Name                      string
	Resources                 *resource.Resource
	TraceEndpoint             string
	TracePort                 int
	TraceExporter             string
	TraceExporterBatchTimeout time.Duration
}

// OpenTrace represents the tracing of the service
type OpenTelemetryTracer struct {
	ctx  context.Context
	name string

	traceEndpoint             string
	tracePort                 int
	traceExporter             string
	traceExporterBatchTimeout time.Duration

	// Resource is the OpenTelemetry resource.
	res *resource.Resource

	// TracerProvider is the OpenTelemetry trace provider.
	tp *trace.TracerProvider

	// Tracer is the OpenTelemetry trace tracer.
	Tracer otelTrace.Tracer
}

func NewOpenTelemetryTracer(ctx context.Context, conf *OpenTelemetryTracerConfig) *OpenTelemetryTracer {
	return &OpenTelemetryTracer{
		ctx:  ctx,
		name: conf.Name,

		traceEndpoint:             conf.TraceEndpoint,
		tracePort:                 conf.TracePort,
		traceExporter:             conf.TraceExporter,
		traceExporterBatchTimeout: conf.TraceExporterBatchTimeout,

		res: conf.Resources,

		Tracer: otel.Tracer(conf.Name),
	}
}

func (ref *OpenTelemetryTracer) SetupTraces() error {
	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace exporter.
	tExp, err := ref.newTraceExporter(ref.ctx)
	if err != nil {
		return err
	}

	// Set up trace provider.
	tp, err := ref.newTraceProvider(tExp)
	if err != nil {
		return err
	}
	ref.tp = tp

	otel.SetTracerProvider(tp)
	ref.Tracer = tp.Tracer(ref.name)

	return nil
}

func (ref *OpenTelemetryTracer) Shutdown() {
	if ref.Tracer != nil {
		if err := ref.tp.Shutdown(ref.ctx); err != nil {
			slog.Error("failed to shutdown OpenTelemetry tracer", "error", err)
		}
	}
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func (ref *OpenTelemetryTracer) newTraceExporter(ctx context.Context) (trace.SpanExporter, error) {
	var exporter trace.SpanExporter
	var err error

	switch ref.traceExporter {
	case "console":
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
	case "otlp-http":
		insecureOpt := otlptracehttp.WithInsecure()
		withCompression := otlptracehttp.WithCompression(otlptracehttp.GzipCompression)
		endpointOpt := otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%d", ref.traceEndpoint, ref.tracePort))
		exporter, err = otlptracehttp.New(ctx, insecureOpt, endpointOpt, withCompression)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown trace exporter: %s", ref.traceExporter)
	}

	return exporter, nil
}

func (ref *OpenTelemetryTracer) newTraceProvider(exp trace.SpanExporter) (*trace.TracerProvider, error) {
	sampler := trace.TraceIDRatioBased(0.5)

	p := trace.NewTracerProvider(
		trace.WithResource(ref.res),
		trace.WithBatcher(
			exp,
			trace.WithBatchTimeout(ref.traceExporterBatchTimeout),
		),
		trace.WithSampler(sampler),
	)

	return p, nil
}
