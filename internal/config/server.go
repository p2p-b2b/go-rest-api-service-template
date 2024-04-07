package config

import "time"

const (
	// DefaultShutdownTimeout is the default time to wait for the server to shutdown
	DefaultShutdownTimeout = 5 * time.Second

	// DefaultServerAddress is the default address for the server
	DefaultServerAddress = "localhost"

	// DefaultServerPort is the default port for the server
	DefaultServerPort = 8080
)

// ServerConfig is the configuration for the server
type ServerConfig struct {
	Address         Field[string]
	Port            Field[int]
	ShutdownTimeout Field[time.Duration]
}

// NewServerConfig creates a new server configuration
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:         NewField("server.address", "SERVER_ADDRESS", "Server IP Address or Hostname", DefaultServerAddress),
		Port:            NewField("server.port", "SERVER_PORT", "Server Port", 8080),
		ShutdownTimeout: NewField("server.shutdown.timeout", "SERVER_SHUTDOWN_TIMEOUT", "Server Shutdown Timeout", DefaultShutdownTimeout),
	}
}

// ParseEnvVars reads the server configuration from environment variables
// and sets the values in the configuration
func (c *ServerConfig) ParseEnvVars() {
	c.Address.Value = GetEnv(c.Address.EnVarName, c.Address.Value)
	c.Port.Value = GetEnv(c.Port.EnVarName, c.Port.Value)
	c.ShutdownTimeout.Value = GetEnv(c.ShutdownTimeout.EnVarName, c.ShutdownTimeout.Value)
}
