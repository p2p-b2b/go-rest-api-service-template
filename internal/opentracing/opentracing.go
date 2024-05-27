package opentracing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// OpenTrace represents the tracing of the service
type OpenTrace struct {
	otlpendpoint           string
	otlport                int
	ctx                    context.Context
	otTraceProvider        *trace.TracerProvider
	otlAttr_ServiceName    string
	otlAttr_ServiceVersion string
}

func NewOpentracing(conf *config.OpenTraceConfig) *OpenTrace {
	return &OpenTrace{
		otlpendpoint:           conf.OTLPEndpoint.Value,
		otlport:                conf.OTLPPort.Value,
		otlAttr_ServiceName:    conf.OTLAttr_ServiceName,
		otlAttr_ServiceVersion: conf.OTLAttr_ServiceVersion.Value,
	}
}

func (o *OpenTrace) SetContext(ctx context.Context) {
	o.ctx = ctx
}

func (o *OpenTrace) GetTracerProvider() *trace.TracerProvider {
	return o.otTraceProvider
}

func (o *OpenTrace) SetupOTelSDK() (shutdown func(context.Context) error, err error) {
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
	o.otTraceProvider = tracerProvider

	// Set up meter provider.
	// meterProvider, err := newMeterProvider()
	// if err != nil {
	// 	handleErr(err)
	// 	return
	// }
	// shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	// otel.SetMeterProvider(meterProvider)

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

func (o *OpenTrace) newTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {

	insecureOpt := otlptracehttp.WithInsecure()

	//Create resources to set service name and service version
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(o.otlAttr_ServiceName),
		semconv.ServiceVersionKey.String(o.otlAttr_ServiceVersion),
	)

	// Update default OTLP reciver endpoint
	endpointOpt := otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%d", o.otlpendpoint, o.otlport))
	spanExporter, _ := otlptracehttp.New(ctx, insecureOpt, endpointOpt)

	traceProvider := trace.NewTracerProvider(trace.WithBatcher(spanExporter), trace.WithResource(res))

	return traceProvider, nil
}

func newMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	return meterProvider, nil
}

func newLoggerProvider() (*log.LoggerProvider, error) {
	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
