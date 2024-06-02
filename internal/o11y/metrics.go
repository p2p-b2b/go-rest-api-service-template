package o11y

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	otelMetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// OpenTelemetryMeterConfig represents the configuration of the OpenTelemetry meter.
type OpenTelemetryMeterConfig struct {
	Name           string
	Resources      *resource.Resource
	MetricEndpoint string
	MetricPort     int
	MetricExporter string
	MetricInterval time.Duration
}

// OpenTrace represents the tracing of the service
type OpenTelemetryMeter struct {
	ctx  context.Context
	name string

	metricEndpoint string
	metricPort     int
	metricExporter string
	metricInterval time.Duration

	// Resource is the OpenTelemetry resource.
	res *resource.Resource

	// MeterProvider is the OpenTelemetry metric meter provider.
	mp *metric.MeterProvider

	// Meter is the OpenTelemetry metric meter.
	Meter otelMetric.Meter
}

func NewOpenTelemetryMeter(ctx context.Context, conf *OpenTelemetryMeterConfig) *OpenTelemetryMeter {
	return &OpenTelemetryMeter{
		ctx:  ctx,
		name: conf.Name,

		metricEndpoint: conf.MetricEndpoint,
		metricPort:     conf.MetricPort,
		metricExporter: conf.MetricExporter,
		metricInterval: conf.MetricInterval,

		res: conf.Resources,

		Meter: otel.Meter(conf.Name),
	}
}

func (o *OpenTelemetryMeter) SetupMetrics() error {
	// Set up metric exporter.
	mExp, err := o.newMetricExporter(o.ctx)
	if err != nil {
		return err
	}

	// Set up meter provider.
	mp, err := o.newMeterProvider(mExp)
	if err != nil {
		return err
	}
	o.mp = mp

	// Register the meter provider with the global provider.
	otel.SetMeterProvider(mp)
	o.Meter = mp.Meter(o.name)

	return nil
}

func (o *OpenTelemetryMeter) Shutdown() {
	if o.mp != nil {
		o.mp.Shutdown(o.ctx)
	}
}

// newMetricExporter creates a new metric exporter based on the configuration.
func (o *OpenTelemetryMeter) newMetricExporter(ctx context.Context) (metric.Exporter, error) {
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

		endpointOpt := otlpmetrichttp.WithEndpointURL(fmt.Sprintf("http://%s:%d/api/v1/otlp/v1/metrics", o.metricEndpoint, o.metricPort))
		exporter, err = otlpmetrichttp.New(ctx, insecureOpt, endpointOpt)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown metric exporter: %s", o.metricExporter)
	}

	return exporter, nil
}

// newMeterProvider creates a new MeterProvider with the given exporter.
func (o *OpenTelemetryMeter) newMeterProvider(exp metric.Exporter) (*metric.MeterProvider, error) {
	// Create resources to set service name and service version

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(o.res),
		metric.WithReader(
			metric.NewPeriodicReader(exp, metric.WithInterval(o.metricInterval)),
		),
	)

	return meterProvider, nil
}
