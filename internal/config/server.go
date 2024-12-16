package config

import (
	"errors"
	"net"
	"os"
	"time"
)

var (
	// ErrInvalidAddress is the error for invalid server address
	ErrInvalidAddress = errors.New("invalid server address, must not be empty and a valid IP Address or Hostname")

	// ErrInvalidPort is the error for invalid server port
	ErrInvalidPort = errors.New("invalid server port, must be between 1 and 65535")

	// ErrInvalidShutdownTimeout is the error for invalid server shutdown timeout
	ErrInvalidShutdownTimeout = errors.New("invalid server shutdown timeout, must be between 1s and 600s")
)

const (
	// DefaultShutdownTimeout is the default time to wait for the server to shutdown
	DefaultShutdownTimeout = 5 * time.Second

	// DefaultServerAddress is the default address for the server
	DefaultServerAddress = "localhost"

	// DefaultServerPort is the default port for the server
	DefaultServerPort = 8080

	// DefaultServerTLSEnabled is the default value for enabling TLS
	DefaultServerTLSEnabled = false

	// DefaultServerPprofEnabled is the default value for enabling pprof
	DefaultServerPprofEnabled = false
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
	PprofEnabled    Field[bool]
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
		PprofEnabled:    NewField("server.pprof.enabled", "SERVER_PPROF_ENABLED", "Enable pprof", DefaultServerPprofEnabled),
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
	c.PprofEnabled.Value = GetEnv(c.PprofEnabled.EnVarName, c.PprofEnabled.Value)
}

// Validate validates the server configuration values
func (c *ServerConfig) Validate() error {
	if c.Address.Value == "" || net.ParseIP(c.Address.Value) == nil {
		return ErrInvalidAddress
	}

	// validate the if is a valid IP Address or Hostname

	if c.Port.Value < 1 || c.Port.Value > 65535 {
		return ErrInvalidPort
	}

	if c.ShutdownTimeout.Value < 1*time.Second || c.ShutdownTimeout.Value > 600*time.Second {
		return ErrInvalidShutdownTimeout
	}

	return nil
}
