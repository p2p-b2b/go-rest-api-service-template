package config

import (
	"errors"
	"time"
)

var (
	// ErrInvalidExporter is the error for invalid exporter
	ErrInvalidExporter = errors.New("invalid exporter, must be one of [console, otlp-http]")

	// ErrInvalidSampling is the error for invalid sampling
	ErrInvalidSampling = errors.New("invalid sampling, must be between 0 and 100")

	// ErrInvalidMetricInterval is the error for invalid metric interval
	ErrInvalidMetricInterval = errors.New("invalid metric interval, must be greater than 0")

	// ErrInvalidTracePort is the error for invalid trace port
	ErrInvalidTracePort = errors.New("invalid trace port, must be between 0 and 65535")

	// ErrInvalidMetricPort is the error for invalid metric port
	ErrInvalidMetricPort = errors.New("invalid metric port, must be between 0 and 65535")
)

const (
	TraceExporters  = "console|otlp-http"
	MetricExporters = "console|otlp-http|prometheus"

	DefaultTraceEndpoint             = "localhost"
	DefaultTracePort                 = 4318
	DefaultTraceExporter             = "console"
	DefaultTraceExporterBatchTimeout = 5 * time.Second
	DefaultTraceSampling             = 100

	DefaultMetricEndpoint = "localhost"
	DefaultMetricPort     = 9090
	DefaultMetricExporter = "console"
	DefaultMetricInterval = 15 * time.Second
)

type OpenTelemetryConfig struct {
	TraceEndpoint             Field[string]
	TracePort                 Field[int]
	TraceExporter             Field[string]
	TraceExporterBatchTimeout Field[time.Duration]
	TraceSampling             Field[int]

	MetricEndpoint Field[string]
	MetricPort     Field[int]
	MetricExporter Field[string]
	MetricInterval Field[time.Duration]

	AttributeServiceName    string
	AttributeServiceVersion string
}

func NewOpenTelemetryConfig(appName string, appVersion string) *OpenTelemetryConfig {
	return &OpenTelemetryConfig{
		TraceEndpoint:             NewField("opentelemetry.trace.endpoint", "OPENTELEMETRY_TRACE_ENDPOINT", "OpenTelemetry Endpoint to send traces to", DefaultTraceEndpoint),
		TracePort:                 NewField("opentelemetry.trace.port", "OPENTELEMETRY_TRACE_PORT", "OpenTelemetry Port to send traces to", DefaultTracePort),
		TraceExporter:             NewField("opentelemetry.trace.exporter", "OPENTELEMETRY_TRACE_EXPORTER", "OpenTelemetry Exporter to send traces to ["+TraceExporters+"]", DefaultTraceExporter),
		TraceExporterBatchTimeout: NewField("opentelemetry.trace.exporter.batch.timeout", "OPENTELEMETRY_TRACE_EXPORTER_BATCH_TIMEOUT", "OpenTelemetry Exporter Batch Timeout", DefaultTraceExporterBatchTimeout),
		TraceSampling:             NewField("opentelemetry.trace.sampling", "OPENTELEMETRY_TRACE_SAMPLING", "OpenTelemetry Exporter trace sampling", DefaultTraceSampling),

		MetricEndpoint: NewField("opentelemetry.metric.endpoint", "OPENTELEMETRY_METRIC_ENDPOINT", "OpenTelemetry Endpoint to send metrics to", DefaultMetricEndpoint),
		MetricPort:     NewField("opentelemetry.metric.port", "OPENTELEMETRY_METRIC_PORT", "OpenTelemetry Port to send metrics to", DefaultMetricPort),
		MetricExporter: NewField("opentelemetry.metric.exporter", "OPENTELEMETRY_METRIC_EXPORTER", "OpenTelemetry Exporter to send metrics to ["+MetricExporters+"]", DefaultMetricExporter),
		MetricInterval: NewField("opentelemetry.metric.interval", "OPENTELEMETRY_METRIC_INTERVAL", "OpenTelemetry Interval in to send metrics", DefaultMetricInterval),

		AttributeServiceVersion: appVersion,
		AttributeServiceName:    appName,
	}
}

// ParseEnvVars reads the OpenTracing configuration from environment variables
// and sets the values in the configuration
func (c *OpenTelemetryConfig) ParseEnvVars() {
	c.TraceEndpoint.Value = GetEnv(c.TraceEndpoint.EnVarName, c.TraceEndpoint.Value)
	c.TracePort.Value = GetEnv(c.TracePort.EnVarName, c.TracePort.Value)
	c.TraceExporter.Value = GetEnv(c.TraceExporter.EnVarName, c.TraceExporter.Value)
	c.TraceExporterBatchTimeout.Value = GetEnv(c.TraceExporterBatchTimeout.EnVarName, c.TraceExporterBatchTimeout.Value)
	c.TraceSampling.Value = GetEnv(c.TraceSampling.EnVarName, c.TraceSampling.Value)

	c.MetricEndpoint.Value = GetEnv(c.MetricEndpoint.EnVarName, c.MetricEndpoint.Value)
	c.MetricPort.Value = GetEnv(c.MetricPort.EnVarName, c.MetricPort.Value)
	c.MetricExporter.Value = GetEnv(c.MetricExporter.EnVarName, c.MetricExporter.Value)
	c.MetricInterval.Value = GetEnv(c.MetricInterval.EnVarName, c.MetricInterval.Value)
}

// Validate validates the OpenTracing configuration values
func (c *OpenTelemetryConfig) Validate() error {
	if c.TraceExporter.Value != "console" &&
		c.TraceExporter.Value != "otlp-http" {
		return ErrInvalidExporter
	}

	if c.MetricExporter.Value != "console" &&
		c.MetricExporter.Value != "otlp-http" &&
		c.MetricExporter.Value != "prometheus" {
		return ErrInvalidExporter
	}

	if c.TraceSampling.Value < 0 || c.TraceSampling.Value > 100 {
		return ErrInvalidSampling
	}

	if c.MetricInterval.Value < 0 {
		return ErrInvalidMetricInterval
	}

	if c.MetricPort.Value < 1 || c.MetricPort.Value > 65535 {
		return ErrInvalidMetricPort
	}

	if c.TracePort.Value < 1 || c.TracePort.Value > 65535 {
		return ErrInvalidTracePort
	}

	return nil
}
