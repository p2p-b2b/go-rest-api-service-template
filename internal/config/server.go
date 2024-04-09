package config

import (
	"os"
	"time"
)

const (
	// DefaultShutdownTimeout is the default time to wait for the server to shutdown
	DefaultShutdownTimeout = 5 * time.Second

	// DefaultServerAddress is the default address for the server
	DefaultServerAddress = "localhost"

	// DefaultServerPort is the default port for the server
	DefaultServerPort = 8443

	// DefaultServerTLSEnabled is the default value for enabling TLS
	DefaultServerTLSEnabled = true
)

var (
	// DefaultServerPrivateKeyFile is the default private key file for the server
	// DefaultServerPrivateKeyFile = "tls.key"
	DefaultServerPrivateKeyFile = FileVar{os.NewFile(0, "server.key"), os.O_RDONLY}

	// DefaultServerCertificateFile is the default certificate file for the server
	// DefaultServerCertificateFile = "tls.crt"
	DefaultServerCertificateFile = FileVar{os.NewFile(0, "server.crt"), os.O_RDONLY}
)

// ServerConfig is the configuration for the server
type ServerConfig struct {
	Address         Field[string]
	Port            Field[int]
	ShutdownTimeout Field[time.Duration]
	PrivateKeyFile  Field[FileVar]
	CertificateFile Field[FileVar]
	TLSEnabled      Field[bool]
}

// NewServerConfig creates a new server configuration
func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		Address:         NewField("server.address", "SERVER_ADDRESS", "Server IP Address or Hostname", DefaultServerAddress),
		Port:            NewField("server.port", "SERVER_PORT", "Server Port", DefaultServerPort),
		ShutdownTimeout: NewField("server.shutdown.timeout", "SERVER_SHUTDOWN_TIMEOUT", "Server Shutdown Timeout", DefaultShutdownTimeout),
		PrivateKeyFile:  NewField("server.private.key.file", "SERVER_PRIVATE_KEY_FILE", "Server Private Key File", DefaultServerPrivateKeyFile),
		CertificateFile: NewField("server.certificate.file", "SERVER_CERTIFICATE_FILE", "Server Certificate File", DefaultServerCertificateFile),
		TLSEnabled:      NewField("server.tls.enabled", "SERVER_TLS_ENABLED", "Enable TLS", DefaultServerTLSEnabled),
	}
}

// ParseEnvVars reads the server configuration from environment variables
// and sets the values in the configuration
func (c *ServerConfig) ParseEnvVars() {
	c.Address.Value = GetEnv(c.Address.EnVarName, c.Address.Value)
	c.Port.Value = GetEnv(c.Port.EnVarName, c.Port.Value)
	c.ShutdownTimeout.Value = GetEnv(c.ShutdownTimeout.EnVarName, c.ShutdownTimeout.Value)
	c.PrivateKeyFile.Value = GetEnv(c.PrivateKeyFile.EnVarName, c.PrivateKeyFile.Value)
	c.CertificateFile.Value = GetEnv(c.CertificateFile.EnVarName, c.CertificateFile.Value)
	c.TLSEnabled.Value = GetEnv(c.TLSEnabled.EnVarName, c.TLSEnabled.Value)
}
