package config

const (
	// Default Database Configuration
	DefaultOTLPEndpoint = "localhost"
	DefaultOTLPPort     = 4318
)

type OpenTraceConfig struct {
	OTLPEndpoint Field[string]
	OTLPPort     Field[int]
}

func NewOpenTracingConfig() *OpenTraceConfig {
	return &OpenTraceConfig{
		OTLPEndpoint: NewField("opentrace.oltpendpoint", "OTLP_ENDPOINT", "OTLP Endoint to send traces to", DefaultOTLPEndpoint),
		OTLPPort:     NewField("opentrace.port", "OTLP_PORT", "OTLP Port to send traces to", DefaultOTLPPort),
	}
}

// PaseEnvVars reads the OpenTracing configuration from environment variables
// and sets the values in the configuration
func (c *OpenTraceConfig) PaseEnvVars() {
	c.OTLPEndpoint.Value = GetEnv(c.OTLPEndpoint.EnVarName, c.OTLPEndpoint.Value)
	c.OTLPPort.Value = GetEnv(c.OTLPPort.EnVarName, c.OTLPPort.Value)
}
