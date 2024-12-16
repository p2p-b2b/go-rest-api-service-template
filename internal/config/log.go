package config

import (
	"errors"
	"os"
)

var (
	// ErrInvalidLogLevel is the error for invalid log level
	ErrInvalidLogLevel = errors.New("invalid log level, must be one of [debug, info, warn, error]")

	// ErrInvalidLogFormat is the error for invalid log format
	ErrInvalidLogFormat = errors.New("invalid log format, must be one of [text, json]")
)

var (
	// DefaultLogLevel is the default log level
	DefaultLogLevel = "info"

	// DefaultLogFormat is the default log format
	DefaultLogFormat = "text"

	// DefaultLogOutput is the default log output destination
	DefaultLogOutput = FileVar{os.Stdout, os.O_APPEND | os.O_CREATE | os.O_WRONLY}
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

// Validate validates the logger configuration values
func (c *LogConfig) Validate() error {
	if c.Level.Value != "debug" &&
		c.Level.Value != "info" &&
		c.Level.Value != "warn" &&
		c.Level.Value != "error" {
		return ErrInvalidLogLevel
	}

	if c.Format.Value != "text" &&
		c.Format.Value != "json" {
		return ErrInvalidLogFormat
	}

	return nil
}
