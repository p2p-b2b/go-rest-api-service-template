package config

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	ErrTraceInvalidExporter  = errors.New("invalid trace exporter, must be one of [" + ValidTraceExporters + "]")
	ErrMetricInvalidExporter = errors.New("invalid metric exporter, must be one of [" + ValidMetricExporters + "]")
	ErrMetricInvalidSampling = errors.New("invalid sampling, must be between [" + strconv.Itoa(ValidTaceSamplingMin) + "] and [" + strconv.Itoa(ValidTaceSamplingMax) + "]")
	ErrMetricInvalidInterval = errors.New("invalid metric interval, must be greater than [" + ValidMetricMinInterval.String() + "]")
	ErrTraceInvalidPort      = errors.New("invalid trace port, must be between [" + strconv.Itoa(ValidTraceMinPort) + "] and [" + strconv.Itoa(ValidTraceMaxPort) + "]")
	ErrMetricInvalidPort     = errors.New("invalid metric port, must be between [" + strconv.Itoa(ValidMetricMinPort) + "] and [" + strconv.Itoa(ValidMetricMaxPort) + "]")
)

const (
	ValidTraceExporters    = "console|otlp-http|noop"
	ValidMetricExporters   = "console|otlp-http|prometheus|noop"
	ValidTaceSamplingMin   = 0
	ValidTaceSamplingMax   = 100
	ValidMetricMinInterval = 1 * time.Second
	ValidMetricMinPort     = 1
	ValidMetricMaxPort     = 65535
	ValidTraceMinPort      = 1
	ValidTraceMaxPort      = 65535

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
		TraceExporter:             NewField("opentelemetry.trace.exporter", "OPENTELEMETRY_TRACE_EXPORTER", "OpenTelemetry Exporter to send traces to. Possible values ["+ValidTraceExporters+"]", DefaultTraceExporter),
		TraceExporterBatchTimeout: NewField("opentelemetry.trace.exporter.batch.timeout", "OPENTELEMETRY_TRACE_EXPORTER_BATCH_TIMEOUT", "OpenTelemetry Exporter Batch Timeout", DefaultTraceExporterBatchTimeout),
		TraceSampling:             NewField("opentelemetry.trace.sampling", "OPENTELEMETRY_TRACE_SAMPLING", "OpenTelemetry Exporter trace sampling", DefaultTraceSampling),

		MetricEndpoint: NewField("opentelemetry.metric.endpoint", "OPENTELEMETRY_METRIC_ENDPOINT", "OpenTelemetry Endpoint to send metrics to", DefaultMetricEndpoint),
		MetricPort:     NewField("opentelemetry.metric.port", "OPENTELEMETRY_METRIC_PORT", "OpenTelemetry Port to send metrics to", DefaultMetricPort),
		MetricExporter: NewField("opentelemetry.metric.exporter", "OPENTELEMETRY_METRIC_EXPORTER", "OpenTelemetry Exporter to send metrics to. Possible values ["+ValidMetricExporters+"]", DefaultMetricExporter),
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
	if !slices.Contains(strings.Split(ValidTraceExporters, "|"), c.TraceExporter.Value) {
		return ErrTraceInvalidExporter
	}

	if !slices.Contains(strings.Split(ValidMetricExporters, "|"), c.MetricExporter.Value) {
		return ErrMetricInvalidExporter
	}

	if c.TraceSampling.Value < ValidTaceSamplingMin || c.TraceSampling.Value > ValidTaceSamplingMax {
		return ErrMetricInvalidSampling
	}

	if c.MetricInterval.Value < ValidMetricMinInterval {
		return ErrMetricInvalidInterval
	}

	if c.MetricPort.Value < ValidMetricMinPort || c.MetricPort.Value > ValidMetricMaxPort {
		return ErrMetricInvalidPort
	}

	if c.TracePort.Value < ValidTraceMinPort || c.TracePort.Value > ValidTraceMaxPort {
		return ErrTraceInvalidPort
	}

	return nil
}
