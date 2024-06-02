package o11y

import (
	"context"

	"github.com/p2p-b2b/go-rest-api-service-template/internal/config"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

type OpenTelemetry struct {
	Traces  *OpenTelemetryTracer
	Metrics *OpenTelemetryMeter
}

func New(ctx context.Context, conf *config.OpenTelemetryConfig) (*OpenTelemetry, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(conf.AttributeServiceName),
			semconv.ServiceVersionKey.String(conf.AttributeServiceVersion),
		),
	)
	if err != nil {
		return nil, err
	}

	tracerConf := &OpenTelemetryTracerConfig{
		Name:                      conf.AttributeServiceName,
		Resources:                 res,
		TraceEndpoint:             conf.TraceEndpoint.Value,
		TracePort:                 conf.TracePort.Value,
		TraceExporter:             conf.TraceExporter.Value,
		TraceExporterBatchTimeout: conf.TraceExporterBatchTimeout.Value,
	}

	meterConf := &OpenTelemetryMeterConfig{
		Name:           conf.AttributeServiceName,
		Resources:      res,
		MetricEndpoint: conf.MetricEndpoint.Value,
		MetricPort:     conf.MetricPort.Value,
		MetricExporter: conf.MetricExporter.Value,
		MetricInterval: conf.MetricInterval.Value,
	}

	op := &OpenTelemetry{
		Traces:  NewOpenTelemetryTracer(ctx, tracerConf),
		Metrics: NewOpenTelemetryMeter(ctx, meterConf),
	}

	return op, nil
}

func (o *OpenTelemetry) Start() error {
	if err := o.Traces.SetupTraces(); err != nil {
		return err
	}

	if err := o.Metrics.SetupMetrics(); err != nil {
		return err
	}

	return nil
}

func (o *OpenTelemetry) Shutdown() {
	o.Traces.Shutdown()
	o.Metrics.Shutdown()
}
