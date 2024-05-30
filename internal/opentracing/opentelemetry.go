package opentelemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelMetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	trc "go.opentelemetry.io/otel/trace"
)

type Metrics struct {
	HttpCnt otelMetric.Int64Counter
}

// OpenTrace represents the tracing of the service
type Opentelemetry struct {
	otlptraceendpoint      string
	otltraceport           int
	otlpmetricendpoint     string
	otlmetricport          int
	otlmetricinterval      int
	ctx                    context.Context
	otlAttr_ServiceName    string
	otlAttr_ServiceVersion string
	trace                  trc.Tracer
	appName                string
	meterProvider          *metric.MeterProvider
}

func NewOpenTelemetry(ctx context.Context, appName string, conf *config.OpenTelemetryConfig) *Opentelemetry {
	return &Opentelemetry{
		otlptraceendpoint:      conf.OTLPTraceEndpoint.Value,
		otltraceport:           conf.OTLPTracePort.Value,
		otlpmetricendpoint:     conf.OTLPMetricEndpoint.Value,
		otlmetricport:          conf.OTLPMetricPort.Value,
		otlmetricinterval:      conf.OTLPMetricInterval.Value,
		otlAttr_ServiceName:    conf.OTLAttr_ServiceName,
		otlAttr_ServiceVersion: conf.OTLAttr_ServiceVersion.Value,
		ctx:                    ctx,
		appName:                appName,
	}
}

func (o *Opentelemetry) GetTrace() trc.Tracer {
	return o.trace
}

func (o *Opentelemetry) GetMeterProdider() *metric.MeterProvider {
	return o.meterProvider
}

func (o *Opentelemetry) SetupOTelSDK() (shutdown func(context.Context) error, err error) {
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
	o.trace = tracerProvider.Tracer(o.appName)

	//Set up meter provider.
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

func (o *Opentelemetry) newTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {

	insecureOpt := otlptracehttp.WithInsecure()

	//Create resources to set service name and service version
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(o.otlAttr_ServiceName),
		semconv.ServiceVersionKey.String(o.otlAttr_ServiceVersion),
	)

	// Update default OTLP reciver endpoint
	endpointOpt := otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%d", o.otlptraceendpoint, o.otltraceport))
	spanExporter, _ := otlptracehttp.New(ctx, insecureOpt, endpointOpt)

	traceProvider := trace.NewTracerProvider(trace.WithBatcher(spanExporter), trace.WithResource(res))

	return traceProvider, nil
}

func (o *Opentelemetry) newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {

	//Create resources to set service name and service version

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(o.otlAttr_ServiceName),
		semconv.ServiceVersionKey.String(o.otlAttr_ServiceVersion),
	)

	metricExporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpointURL(fmt.Sprintf("http://%s:%d/api/v1/otlp/v1/metrics", o.otlpmetricendpoint, o.otlmetricport)))

	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(time.Duration(o.otlmetricinterval)*time.Second))),
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