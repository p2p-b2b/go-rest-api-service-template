package config

const (
	// Default Database Configuration
	DefaultOTLPTraceEndpoint      = "localhost"
	DefaultOTLPTracePort          = 4318
	DefaultOTLPMetricEndpoint     = "localhost"
	DefaultOTLPMetricPort         = 9090
	DefaultOTLPMetricInterval     = 15
	DefaultOTLAttr_ServiceVersion = "0.1.0"
)

type OpenTelemetryConfig struct {
	OTLPTraceEndpoint      Field[string]
	OTLPTracePort          Field[int]
	OTLPMetricEndpoint     Field[string]
	OTLPMetricPort         Field[int]
	OTLAttr_ServiceName    string
	OTLAttr_ServiceVersion Field[string]
	OTLPMetricInterval     Field[int]
}

func NewOpenTelemetryConfig(appName string) *OpenTelemetryConfig {
	return &OpenTelemetryConfig{
		OTLPTraceEndpoint:      NewField("opentelemetry.oltptraceendpoint", "OTLP_TRACE_ENDPOINT", "OTLP Endoint to send traces to", DefaultOTLPTraceEndpoint),
		OTLPTracePort:          NewField("opentelemetry.oltptraceport", "OTLP_TRACE_PORT", "OTLP Port to send traces to", DefaultOTLPTracePort),
		OTLPMetricEndpoint:     NewField("opentelemetry.oltpmetricendpoint", "OTLP_METRIC_ENDPOINT", "OTLP Endoint to send metrics to", DefaultOTLPMetricEndpoint),
		OTLPMetricPort:         NewField("opentelemetry.oltpmetricport", "OTLP_METRIC_PORT", "OTLP Port to send metrics to", DefaultOTLPMetricPort),
		OTLAttr_ServiceVersion: NewField("opentelemetry.service_version", "OTLP_SERVICE_VERSION", "OTLP service version to show on traces", DefaultOTLAttr_ServiceVersion),
		OTLPMetricInterval:     NewField("opentelemetry.oltpmetricinterval", "OTLP_METRIC_INTERVAL", "OTLP metric interval to push", DefaultOTLPMetricInterval),
		OTLAttr_ServiceName:    appName,
	}
}

// PaseEnvVars reads the OpenTracing configuration from environment variables
// and sets the values in the configuration
func (c *OpenTelemetryConfig) PaseEnvVars() {
	c.OTLPTraceEndpoint.Value = GetEnv(c.OTLPTraceEndpoint.EnVarName, c.OTLPTraceEndpoint.Value)
	c.OTLPTracePort.Value = GetEnv(c.OTLPTracePort.EnVarName, c.OTLPTracePort.Value)
	c.OTLPMetricEndpoint.Value = GetEnv(c.OTLPMetricEndpoint.EnVarName, c.OTLPMetricEndpoint.Value)
	c.OTLPMetricPort.Value = GetEnv(c.OTLPMetricPort.EnVarName, c.OTLPMetricPort.Value)
	c.OTLPMetricInterval.Value = GetEnv(c.OTLPMetricInterval.EnVarName, c.OTLPMetricInterval.Value)
	c.OTLAttr_ServiceVersion.Value = GetEnv(c.OTLAttr_ServiceVersion.EnVarName, c.OTLAttr_ServiceVersion.Value)
}
