package o11y

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otelMetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	otelTrace "go.opentelemetry.io/otel/trace"
)

// OpenTrace represents the tracing of the service
type OpenTelemetry struct {
	ctx context.Context

	traceEndpoint             string
	tracePort                 int
	traceExporter             string
	traceExporterBatchTimeout time.Duration

	metricEndpoint string
	metricPort     int
	metricExporter string
	metricInterval time.Duration

	attributeServiceName    string
	attributeServiceVersion string

	Tracer otelTrace.Tracer
	Meter  otelMetric.Meter
}

func New(ctx context.Context, conf *config.OpenTelemetryConfig) *OpenTelemetry {
	return &OpenTelemetry{
		ctx: ctx,

		traceEndpoint:             conf.TraceEndpoint.Value,
		tracePort:                 conf.TracePort.Value,
		traceExporter:             conf.TraceExporter.Value,
		traceExporterBatchTimeout: conf.TraceExporterBatchTimeout.Value,

		metricEndpoint: conf.MetricEndpoint.Value,
		metricPort:     conf.MetricPort.Value,
		metricExporter: conf.MetricExporter.Value,
		metricInterval: conf.MetricInterval.Value,

		attributeServiceName:    conf.AttributeServiceName,
		attributeServiceVersion: conf.AttributeServiceVersion,

		Tracer: otel.Tracer(conf.AttributeServiceName),
		Meter:  otel.Meter(conf.AttributeServiceName),
	}
}

func (o *OpenTelemetry) SetupOTelSDK() (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(o.ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace exporter.
	tExp, err := o.newTraceExporter(o.ctx)
	if err != nil {
		handleErr(err)
		return
	}

	// Set up trace provider.
	tp, err := o.newTraceProvider(tExp)
	if err != nil {
		handleErr(err)
		return
	}

	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	otel.SetTracerProvider(tp)
	o.Tracer = tp.Tracer(o.attributeServiceName)

	// Set up metric exporter.
	mExp, err := o.newMetricExporter(o.ctx)
	if err != nil {
		handleErr(err)
		return
	}

	// Set up meter provider.
	meterProvider, err := o.newMeterProvider(mExp)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)
	o.Meter = meterProvider.Meter(o.attributeServiceName)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func (o *OpenTelemetry) newMetricExporter(ctx context.Context) (metric.Exporter, error) {
	var exporter metric.Exporter
	var err error

	switch o.metricExporter {
	case "console":
		exporter, err = stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
	case "otlp-http":
		insecureOpt := otlpmetrichttp.WithInsecure()

		endpointOpt := otlpmetrichttp.WithEndpoint(fmt.Sprintf("%s:%d", o.metricEndpoint, o.metricPort))
		exporter, err = otlpmetrichttp.New(ctx, insecureOpt, endpointOpt)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown metric exporter: %s", o.metricExporter)
	}

	return exporter, nil
}

func (o *OpenTelemetry) newTraceExporter(ctx context.Context) (trace.SpanExporter, error) {
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

func (o *OpenTelemetry) newTraceProvider(exp trace.SpanExporter) (*trace.TracerProvider, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(o.attributeServiceName),
			semconv.ServiceVersionKey.String(o.attributeServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	p := trace.NewTracerProvider(
		trace.WithBatcher(
			exp,
			trace.WithBatchTimeout(o.traceExporterBatchTimeout),
		),
		trace.WithResource(r),
	)

	return p, nil
}

func (o *OpenTelemetry) newMeterProvider(exp metric.Exporter) (*metric.MeterProvider, error) {
	// Create resources to set service name and service version

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(o.attributeServiceName),
			semconv.ServiceVersionKey.String(o.attributeServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(r),
		metric.WithReader(
			metric.NewPeriodicReader(exp, metric.WithInterval(o.metricInterval)),
		),
	)

	return meterProvider, nil
}
