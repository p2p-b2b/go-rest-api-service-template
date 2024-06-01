package config

import "time"

const (
	TraceExporters  = "console|otlp-http"
	MetricExporters = "console|otlp-http|prometheus"

	DefaultTraceEndpoint             = "localhost"
	DefaultTracePort                 = 4318
	DefaultTraceExporter             = "console"
	DefaultTraceExporterBatchTimeout = 5 * time.Second

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

		MetricEndpoint: NewField("opentelemetry.metric.endpoint", "OPENTELEMETRY_METRIC_ENDPOINT", "OpenTelemetry Endpoint to send metrics to", DefaultMetricEndpoint),
		MetricPort:     NewField("opentelemetry.metric.port", "OPENTELEMETRY_METRIC_PORT", "OpenTelemetry Port to send metrics to", DefaultMetricPort),
		MetricExporter: NewField("opentelemetry.metric.exporter", "OPENTELEMETRY_METRIC_EXPORTER", "OpenTelemetry Exporter to send metrics to ["+MetricExporters+"]", DefaultMetricExporter),
		MetricInterval: NewField("opentelemetry.metric.interval", "OPENTELEMETRY_METRIC_INTERVAL", "OpenTelemetry Interval in to send metrics", DefaultMetricInterval),

		AttributeServiceVersion: appVersion,
		AttributeServiceName:    appName,
	}
}

// PaseEnvVars reads the OpenTracing configuration from environment variables
// and sets the values in the configuration
func (c *OpenTelemetryConfig) PaseEnvVars() {
	c.TraceEndpoint.Value = GetEnv(c.TraceEndpoint.EnVarName, c.TraceEndpoint.Value)
	c.TracePort.Value = GetEnv(c.TracePort.EnVarName, c.TracePort.Value)
	c.TraceExporter.Value = GetEnv(c.TraceExporter.EnVarName, c.TraceExporter.Value)
	c.TraceExporterBatchTimeout.Value = GetEnv(c.TraceExporterBatchTimeout.EnVarName, c.TraceExporterBatchTimeout.Value)

	c.MetricEndpoint.Value = GetEnv(c.MetricEndpoint.EnVarName, c.MetricEndpoint.Value)
	c.MetricPort.Value = GetEnv(c.MetricPort.EnVarName, c.MetricPort.Value)
	c.MetricExporter.Value = GetEnv(c.MetricExporter.EnVarName, c.MetricExporter.Value)
	c.MetricInterval.Value = GetEnv(c.MetricInterval.EnVarName, c.MetricInterval.Value)
}
