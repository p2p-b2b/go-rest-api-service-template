package o11y

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	tracer "go.opentelemetry.io/otel/trace"
)

// OpenTrace represents the tracing of the service
type OpenTelemetry struct {
	traceEndpoint           string
	tracePort               int
	metricEndpoint          string
	metricPort              int
	metricInterval          time.Duration
	AttributeServiceName    string
	AttributeServiceVersion string
	ctx                     context.Context
	tracer                  tracer.Tracer
	meterProvider           *metric.MeterProvider
}

func New(ctx context.Context, conf *config.OpenTelemetryConfig) *OpenTelemetry {
	return &OpenTelemetry{
		traceEndpoint:        conf.TraceEndpoint.Value,
		tracePort:            conf.TracePort.Value,
		metricEndpoint:       conf.MetricEndpoint.Value,
		metricPort:           conf.MetricPort.Value,
		metricInterval:       conf.MetricInterval.Value,
		AttributeServiceName: conf.AttributeServiceName,
		ctx:                  ctx,
	}
}

func (o *OpenTelemetry) GetTrace() tracer.Tracer {
	return o.tracer
}

func (o *OpenTelemetry) GetMeterProvider() *metric.MeterProvider {
	return o.meterProvider
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

	// Set up trace provider.
	tracerProvider, err := o.newTraceProvider(o.ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)
	o.tracer = tracerProvider.Tracer(o.AttributeServiceName)

	// Set up meter provider.
	meterProvider, err := o.newMeterProvider(o.ctx)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)
	o.meterProvider = meterProvider

	// Set up logger provider.
	// loggerProvider, err := newLoggerProvider()
	// if err != nil {
	// 	handleErr(err)
	// 	return
	// }
	// shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	// global.SetLoggerProvider(loggerProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func (o *OpenTelemetry) newTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {
	insecureOpt := otlptracehttp.WithInsecure()

	// Create resources to set service name and service version
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(o.AttributeServiceName),
		semconv.ServiceVersionKey.String(o.AttributeServiceVersion),
	)

	// Update default OTLP receiver endpoint
	endpointOpt := otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%d", o.traceEndpoint, o.tracePort))
	spanExporter, _ := otlptracehttp.New(ctx, insecureOpt, endpointOpt)

	traceProvider := trace.NewTracerProvider(trace.WithBatcher(spanExporter), trace.WithResource(res))

	return traceProvider, nil
}

func (o *OpenTelemetry) newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	// Create resources to set service name and service version

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(o.AttributeServiceName),
		semconv.ServiceVersionKey.String(o.AttributeServiceVersion),
	)

	metricExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpointURL(fmt.Sprintf("http://%s:%d/api/v1/otlp/v1/metrics", o.metricEndpoint, o.metricPort)))
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(metricExporter, metric.WithInterval(o.metricInterval)),
		),
	)

	return meterProvider, nil
}

// func newLoggerProvider() (*log.LoggerProvider, error) {
// 	logExporter, err := stdoutlog.New()
// 	if err != nil {
// 		return nil, err
// 	}

// 	loggerProvider := log.NewLoggerProvider(
// 		log.WithProcessor(log.NewBatchProcessor(logExporter)),
// 	)
// 	return loggerProvider, nil
// }
