package o11y

import (
	"context"
	"fmt"
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

func (o *OpenTelemetryTracer) SetupTraces() error {
	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace exporter.
	tExp, err := o.newTraceExporter(o.ctx)
	if err != nil {
		return err
	}

	// Set up trace provider.
	tp, err := o.newTraceProvider(tExp)
	if err != nil {
		return err
	}
	o.tp = tp

	otel.SetTracerProvider(tp)
	o.Tracer = tp.Tracer(o.name)

	return nil
}

func (o *OpenTelemetryTracer) Shutdown() {
	if o.Tracer != nil {
		o.tp.Shutdown(o.ctx)
	}
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func (o *OpenTelemetryTracer) newTraceExporter(ctx context.Context) (trace.SpanExporter, error) {
	var exporter trace.SpanExporter
	var err error

	switch o.traceExporter {
	case "console":
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
	case "otlp-http":
		insecureOpt := otlptracehttp.WithInsecure()

		endpointOpt := otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%d", o.traceEndpoint, o.tracePort))
		exporter, err = otlptracehttp.New(ctx, insecureOpt, endpointOpt)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown trace exporter: %s", o.traceExporter)
	}

	return exporter, nil
}

func (o *OpenTelemetryTracer) newTraceProvider(exp trace.SpanExporter) (*trace.TracerProvider, error) {
	p := trace.NewTracerProvider(
		trace.WithResource(o.res),
		trace.WithBatcher(
			exp,
			trace.WithBatchTimeout(o.traceExporterBatchTimeout),
		),
	)

	return p, nil
}
