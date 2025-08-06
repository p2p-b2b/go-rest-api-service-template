package config

import (
	"os"
	"slices"
	"strings"
)

const (
	ValidLogLevel  = "debug|info|warn|error"
	ValidLogFormat = "text|json"

	DefaultLogLevel     = "info"
	DefaultLogFormat    = "text"
	DefaultLogDebug     = false
	DefaultLogAddSource = false
)

// DefaultLogOutput is the default log output destination
var DefaultLogOutput = FileVar{os.Stdout, os.O_APPEND | os.O_CREATE | os.O_WRONLY}

// LogConfig is the configuration for the logger
type LogConfig struct {
	Level     Field[string]
	Format    Field[string]
	Output    Field[FileVar]
	Debug     Field[bool]
	AddSource Field[bool]
}

// NewLogConfig creates a new logger configuration
func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:     NewField("log.level", "LOG_LEVEL", "Log Level. Possible values ["+ValidLogLevel+"]", DefaultLogLevel),
		Format:    NewField("log.format", "LOG_FORMAT", "Log Format. Possible values ["+ValidLogFormat+"]", DefaultLogFormat),
		Output:    NewField("log.output", "LOG_OUTPUT", "Log Output", DefaultLogOutput),
		Debug:     NewField("debug", "DEBUG", "Debug mode. Short hand for log.level=debug", DefaultLogDebug),
		AddSource: NewField("log.add.source", "LOG_ADD_SOURCE", "Add source file and line number to log output", DefaultLogAddSource),
	}
}

// ParseEnvVars reads the logger configuration from environment variables
// and sets the values in the configuration
func (c *LogConfig) ParseEnvVars() {
	c.Level.Value = GetEnv(c.Level.EnVarName, c.Level.Value)
	c.Format.Value = GetEnv(c.Format.EnVarName, c.Format.Value)
	c.Output.Value = GetEnv(c.Output.EnVarName, c.Output.Value)
	c.Debug.Value = GetEnv(c.Debug.EnVarName, c.Debug.Value)
	c.AddSource.Value = GetEnv(c.AddSource.EnVarName, c.AddSource.Value)
}

// Validate validates the logger configuration values
func (c *LogConfig) Validate() error {
	if !slices.Contains(strings.Split(ValidLogLevel, "|"), c.Level.Value) {
		return &InvalidConfigurationError{
			Field:   "log.level",
			Value:   c.Level.Value,
			Message: "Log level must be one of [" + ValidLogLevel + "]",
		}
	}

	if !slices.Contains(strings.Split(ValidLogFormat, "|"), c.Format.Value) {
		return &InvalidConfigurationError{
			Field:   "log.format",
			Value:   c.Format.Value,
			Message: "Log format must be one of [" + ValidLogFormat + "]",
		}
	}

	return nil
}
