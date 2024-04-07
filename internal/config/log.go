package config

import "os"

var (
	// DefaultLogLevel is the default log level
	DefaultLogLevel = "info"

	// DefaultLogFormat is the default log format
	DefaultLogFormat = "text"

	// DefaultLogOutput is the default log output destination
	DefaultLogOutput = FileVar{os.Stdout}
)

// LogConfig is the configuration for the logger
type LogConfig struct {
	Level  Field[string]
	Format Field[string]
	Output Field[FileVar]
}

// NewLogConfig creates a new logger configuration
func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:  NewField("log.level", "LOG_LEVEL", "Log Level [debug, info, warn, error]", DefaultLogLevel),
		Format: NewField("log.format", "LOG_FORMAT", "Log Format [text, json]", DefaultLogFormat),
		Output: NewField("log.output", "LOG_OUTPUT", "Log Output", DefaultLogOutput),
	}
}

// ParseEnvVars reads the logger configuration from environment variables
// and sets the values in the configuration
func (c *LogConfig) ParseEnvVars() {
	c.Level.Value = GetEnv(c.Level.EnVarName, c.Level.Value)
	c.Format.Value = GetEnv(c.Format.EnVarName, c.Format.Value)
	c.Output.Value = GetEnv(c.Output.EnVarName, c.Output.Value)
}
