package config

import (
	"errors"
	"os"
	"slices"
	"strings"
)

var (
	// ErrInvalidLogLevel is the error for invalid log level
	ErrInvalidLogLevel = errors.New("invalid log level, must be one of [" + ValidLogLevel + "]")

	// ErrInvalidLogFormat is the error for invalid log format
	ErrInvalidLogFormat = errors.New("invalid log format, must be one of [" + ValidLogFormat + "]")
)

const (
	ValidLogLevel  = "debug|info|warn|error"
	ValidLogFormat = "text|json"

	// DefaultLogLevel is the default log level
	DefaultLogLevel = "info"

	// DefaultLogFormat is the default log format
	DefaultLogFormat = "text"
)

// DefaultLogOutput is the default log output destination
var DefaultLogOutput = FileVar{os.Stdout, os.O_APPEND | os.O_CREATE | os.O_WRONLY}

// LogConfig is the configuration for the logger
type LogConfig struct {
	Level  Field[string]
	Format Field[string]
	Output Field[FileVar]
}

// NewLogConfig creates a new logger configuration
func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:  NewField("log.level", "LOG_LEVEL", "Log Level. Possible values ["+ValidLogLevel+"]", DefaultLogLevel),
		Format: NewField("log.format", "LOG_FORMAT", "Log Format. Possible values ["+ValidLogFormat+"]", DefaultLogFormat),
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
	if !slices.Contains(strings.Split(ValidLogLevel, "|"), c.Level.Value) {
		return ErrInvalidLogLevel
	}

	if !slices.Contains(strings.Split(ValidLogFormat, "|"), c.Format.Value) {
		return ErrInvalidLogFormat
	}

	return nil
}
