package config

const (
	// Default Database Configuration
	DefaultOTLPEndpoint           = "localhost"
	DefaultOTLPPort               = 4318
	DefaultOTLAttr_ServiceVersion = "0.1.0"
)

type OpenTraceConfig struct {
	OTLPEndpoint           Field[string]
	OTLPPort               Field[int]
	OTLAttr_ServiceName    string
	OTLAttr_ServiceVersion Field[string]
}

func NewOpenTracingConfig(appName string) *OpenTraceConfig {
	return &OpenTraceConfig{
		OTLPEndpoint:           NewField("opentrace.oltpendpoint", "OTLP_ENDPOINT", "OTLP Endoint to send traces to", DefaultOTLPEndpoint),
		OTLPPort:               NewField("opentrace.port", "OTLP_PORT", "OTLP Port to send traces to", DefaultOTLPPort),
		OTLAttr_ServiceName:    appName,
		OTLAttr_ServiceVersion: NewField("opentrace.service_version", "OTLP_SERVICE_VERSION", "OTLP service version to show on traces", DefaultOTLAttr_ServiceVersion),
	}
}

// PaseEnvVars reads the OpenTracing configuration from environment variables
// and sets the values in the configuration
func (c *OpenTraceConfig) PaseEnvVars() {
	c.OTLPEndpoint.Value = GetEnv(c.OTLPEndpoint.EnVarName, c.OTLPEndpoint.Value)
	c.OTLPPort.Value = GetEnv(c.OTLPPort.EnVarName, c.OTLPPort.Value)
	c.OTLAttr_ServiceVersion.Value = GetEnv(c.OTLAttr_ServiceVersion.EnVarName, c.OTLAttr_ServiceVersion.Value)
}
