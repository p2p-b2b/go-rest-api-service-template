package config

import "os"

type WriteCloserVar interface {
	WriteCloser()

	// https://pkg.go.dev/gg-scm.io/tool/internal/flag#FlagSet.Var
	// String presents the current value as a string.
	String() string

	// Set is called once, in command line order, for each flag present.
	Set(string) error

	// Get returns the contents of the Value.
	Get() interface{}

	// If IsBoolFlag returns true, then the command-line parser makes
	// -name equivalent to -name=true rather than using the next
	// command-line argument.
	IsBoolFlag() bool
}

var (
	// DefaultLogLevel is the default log level
	DefaultLogLevel = "info"

	// DefaultLogFormat is the default log format
	DefaultLogFormat = "text"

	// DefaultLogOutput is the default log output
	// DefaultLogOutput = "stdout"
	DefaultLogOutput = FileFlag{os.Stdout}
)

// LogConfig is the configuration for the logger
type LogConfig struct {
	Level  Item[string]
	Format Item[string]
	// Output Item[string]
	Output Item[FileFlag]
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
