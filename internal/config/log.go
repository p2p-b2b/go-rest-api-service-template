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
	Level  Item[string]
	Format Item[string]
	// Output Item[string]
	Output Item[FileVar]
}

// NewLogConfig creates a new logger configuration
func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:  NewItem("log.level", "LOG_LEVEL", DefaultLogLevel),
		Format: NewItem("log.format", "LOG_FORMAT", DefaultLogFormat),
		Output: NewItem("log.output", "LOG_OUTPUT", DefaultLogOutput),
	}
}

// ParseEnvVars reads the logger configuration from environment variables
// and sets the values in the configuration
func (c *LogConfig) ParseEnvVars() {
	c.Level.Value = getEnv(c.Level.EnVarName, c.Level.Value)
	c.Format.Value = getEnv(c.Format.EnVarName, c.Format.Value)
	c.Output.Value = getEnv(c.Output.EnVarName, c.Output.Value)
}
