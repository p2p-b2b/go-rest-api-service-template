package config

import (
	"errors"
	"os"
	"slices"
	"strings"
)

var (
	ErrLogInvalidLevel  = errors.New("invalid log level. Must be one of [" + ValidLogLevel + "]")
	ErrLogInvalidFormat = errors.New("invalid log format. Must be one of [" + ValidLogFormat + "]")
)

const (
	ValidLogLevel  = "debug|info|warn|error"
	ValidLogFormat = "text|json"

	DefaultLogLevel  = "info"
	DefaultLogFormat = "text"
	DefaultLogDebug  = false
)

// DefaultLogOutput is the default log output destination
var DefaultLogOutput = FileVar{os.Stdout, os.O_APPEND | os.O_CREATE | os.O_WRONLY}

// LogConfig is the configuration for the logger
type LogConfig struct {
	Level  Field[string]
	Format Field[string]
	Output Field[FileVar]
	Debug  Field[bool]
}

// NewLogConfig creates a new logger configuration
func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:  NewField("log.level", "LOG_LEVEL", "Log Level. Possible values ["+ValidLogLevel+"]", DefaultLogLevel),
		Format: NewField("log.format", "LOG_FORMAT", "Log Format. Possible values ["+ValidLogFormat+"]", DefaultLogFormat),
		Output: NewField("log.output", "LOG_OUTPUT", "Log Output", DefaultLogOutput),
		Debug:  NewField("debug", "DEBUG", "Debug mode. Short hand for log.level=debug", DefaultLogDebug),
	}
}

// ParseEnvVars reads the logger configuration from environment variables
// and sets the values in the configuration
func (c *LogConfig) ParseEnvVars() {
	c.Level.Value = GetEnv(c.Level.EnVarName, c.Level.Value)
	c.Format.Value = GetEnv(c.Format.EnVarName, c.Format.Value)
	c.Output.Value = GetEnv(c.Output.EnVarName, c.Output.Value)
	c.Debug.Value = GetEnv(c.Debug.EnVarName, c.Debug.Value)
}

// Validate validates the logger configuration values
func (c *LogConfig) Validate() error {
	if !slices.Contains(strings.Split(ValidLogLevel, "|"), c.Level.Value) {
		return ErrLogInvalidLevel
	}

	if !slices.Contains(strings.Split(ValidLogFormat, "|"), c.Format.Value) {
		return ErrLogInvalidFormat
	}

	return nil
}
