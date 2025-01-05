package config

import (
	"errors"
	"net"
	"os"
	"slices"
	"strings"
	"time"
)

var (
	// ErrInvalidAddress is the error for invalid server address
	ErrInvalidHTTPServerConfigAddress = errors.New("invalid server address, must not be empty and a valid IP Address or Hostname")

	// ErrInvalidPort is the error for invalid server port
	ErrInvalidHTTPServerConfigPort = errors.New("invalid server port, must be between 1 and 65535")

	// ErrInvalidShutdownTimeout is the error for invalid server shutdown timeout
	ErrInvalidHTTPServerConfigShutdownTimeout = errors.New("invalid server shutdown timeout, must be between 1s and 600s")

	// ErrInvalidCorsAllowedOrigins is the error for invalid CORS allowed origins
	ErrInvalidHTTPServerConfigCorsAllowedOrigins = errors.New("invalid CORS allowed origins. Must not be empty")

	// ErrInvalidCorsAllowedMethods is the error for invalid CORS allowed methods
	ErrInvalidHTTPServerConfigCorsAllowedMethods = errors.New("invalid CORS allowed methods. Must be one of [" + ValidHTTPServerCorsAllowedMethods + "]")

	// ErrInvalidCorsAllowedHeaders is the error for invalid CORS allowed headers
	ErrInvalidHTTPServerConfigCorsAllowedHeaders = errors.New("invalid CORS allowed headers. Must be at least 2 characters long")
)

const (
	// DefaultHTTPServerShutdownTimeout is the default time to wait for the server to shutdown
	DefaultHTTPServerShutdownTimeout = 5 * time.Second

	// DefaultHTTPServerAddress is the default address for the server
	DefaultHTTPServerAddress = "localhost"

	// DefaultHTTPServerPort is the default port for the server
	DefaultHTTPServerPort = 8080

	// DefaultHTTPServerTLSEnabled is the default value for enabling TLS
	DefaultHTTPServerTLSEnabled = false

	// DefaultHTTPServerPprofEnabled is the default value for enabling pprof
	DefaultHTTPServerPprofEnabled = false

	// DefaultHTTPServerCorsEnabled is the default value for enabling CORS
	// If enabled, the server will use the following values for CORS
	// - AllowedOrigins: "*"
	// - AllowedMethods: "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD"
	// - AllowedHeaders: "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-Requested-With, X-Api-Version"
	// Remember to change the values if you need to restrict the allowed origins, methods or headers
	DefaultHTTPServerCorsEnabled = false

	// DefaultHTTPServerCorsAllowCredentials is the default value for allowing credentials
	DefaultHTTPServerCorsAllowCredentials = true

	// DefaultHTTPServerCorsAllowedOrigins is the default value for allowed origins
	// Could be a comma separated list of origins. Example: "http://localhost:3000, http://localhost:8080"
	DefaultHTTPServerCorsAllowedOrigins = "*" // allow all origins

	// DefaultHTTPServerCorsAllowedMethods is the default value for allowed methods
	DefaultHTTPServerCorsAllowedMethods = "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD"

	// DefaultHTTPServerCorsAllowedHeaders is the default value for allowed headers
	DefaultHTTPServerCorsAllowedHeaders = "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-Requested-With, X-Api-Version, Access-Control-Allow-Headers"
)

const (
	ValidHTTPServerCorsAllowedMethods = "GET|POST|PUT|DELETE|OPTIONS|PATCH|HEAD"
)

var (
	// DefaultHTTPServerPrivateKeyFile is the default private key file for the server
	// DefaultHTTPServerPrivateKeyFile = "tls.key"
	DefaultHTTPServerPrivateKeyFile = FileVar{os.NewFile(0, "server.key"), os.O_RDONLY}

	// DefaultHTTPServerCertificateFile is the default certificate file for the server
	// DefaultHTTPServerCertificateFile = "tls.crt"
	DefaultHTTPServerCertificateFile = FileVar{os.NewFile(0, "server.crt"), os.O_RDONLY}
)

// HTTPServerConfig is the configuration for the server
type HTTPServerConfig struct {
	Address              Field[string]
	Port                 Field[int]
	ShutdownTimeout      Field[time.Duration]
	PrivateKeyFile       Field[FileVar]
	CertificateFile      Field[FileVar]
	CorsAllowedOrigins   Field[string]
	CorsAllowedMethods   Field[string]
	CorsAllowedHeaders   Field[string]
	TLSEnabled           Field[bool]
	PprofEnabled         Field[bool]
	CorsEnabled          Field[bool]
	CorsAllowCredentials Field[bool]
}

// NewHTTPServerConfig creates a new server configuration
func NewHTTPServerConfig() *HTTPServerConfig {
	return &HTTPServerConfig{
		Address:         NewField("http.server.address", "SERVER_ADDRESS", "Server IP Address or Hostname", DefaultHTTPServerAddress),
		Port:            NewField("http.server.port", "SERVER_PORT", "Server Port", DefaultHTTPServerPort),
		ShutdownTimeout: NewField("http.server.shutdown.timeout", "SERVER_SHUTDOWN_TIMEOUT", "Server Shutdown Timeout", DefaultHTTPServerShutdownTimeout),
		PrivateKeyFile:  NewField("http.server.private.key.file", "SERVER_PRIVATE_KEY_FILE", "Server Private Key File", DefaultHTTPServerPrivateKeyFile),
		CertificateFile: NewField("http.server.certificate.file", "SERVER_CERTIFICATE_FILE", "Server Certificate File", DefaultHTTPServerCertificateFile),
		TLSEnabled:      NewField("http.server.tls.enabled", "SERVER_TLS_ENABLED", "Enable TLS", DefaultHTTPServerTLSEnabled),
		PprofEnabled:    NewField("http.server.pprof.enabled", "SERVER_PPROF_ENABLED", "Enable pprof", DefaultHTTPServerPprofEnabled),

		CorsEnabled:          NewField("http.server.cors.enabled", "SERVER_CORS_ENABLED", "Enable CORS", DefaultHTTPServerCorsEnabled),
		CorsAllowCredentials: NewField("http.server.cors.allow.credentials", "SERVER_CORS_ALLOW_CREDENTIALS", "Allow Credentials for CORS", DefaultHTTPServerCorsAllowCredentials),
		CorsAllowedOrigins:   NewField("http.server.cors.allowed.origins", "SERVER_CORS_ALLOWED_ORIGINS", "Allowed Origins for CORS", DefaultHTTPServerCorsAllowedOrigins),
		CorsAllowedMethods:   NewField("http.server.cors.allowed.methods", "SERVER_CORS_ALLOWED_METHODS", "Allowed Methods for CORS", DefaultHTTPServerCorsAllowedMethods),
		CorsAllowedHeaders:   NewField("http.server.cors.allowed.headers", "SERVER_CORS_ALLOWED_HEADERS", "Allowed Headers for CORS", DefaultHTTPServerCorsAllowedHeaders),
	}
}

// ParseEnvVars reads the server configuration from environment variables
// and sets the values in the configuration
func (c *HTTPServerConfig) ParseEnvVars() {
	c.Address.Value = GetEnv(c.Address.EnVarName, c.Address.Value)
	c.Port.Value = GetEnv(c.Port.EnVarName, c.Port.Value)
	c.ShutdownTimeout.Value = GetEnv(c.ShutdownTimeout.EnVarName, c.ShutdownTimeout.Value)
	c.PrivateKeyFile.Value = GetEnv(c.PrivateKeyFile.EnVarName, c.PrivateKeyFile.Value)
	c.CertificateFile.Value = GetEnv(c.CertificateFile.EnVarName, c.CertificateFile.Value)
	c.TLSEnabled.Value = GetEnv(c.TLSEnabled.EnVarName, c.TLSEnabled.Value)
	c.PprofEnabled.Value = GetEnv(c.PprofEnabled.EnVarName, c.PprofEnabled.Value)

	c.CorsEnabled.Value = GetEnv(c.CorsEnabled.EnVarName, c.CorsEnabled.Value)
	c.CorsAllowCredentials.Value = GetEnv(c.CorsAllowCredentials.EnVarName, c.CorsAllowCredentials.Value)
	c.CorsAllowedOrigins.Value = GetEnv(c.CorsAllowedOrigins.EnVarName, c.CorsAllowedOrigins.Value)
	c.CorsAllowedMethods.Value = GetEnv(c.CorsAllowedMethods.EnVarName, c.CorsAllowedMethods.Value)
	c.CorsAllowedHeaders.Value = GetEnv(c.CorsAllowedHeaders.EnVarName, c.CorsAllowedHeaders.Value)
}

// Validate validates the server configuration values
func (c *HTTPServerConfig) Validate() error {
	if c.Address.Value == "" || c.Address.Value != "localhost" && net.ParseIP(c.Address.Value) == nil {
		return ErrInvalidHTTPServerConfigAddress
	}

	// validate the if is a valid IP Address or Hostname

	if c.Port.Value < 1 || c.Port.Value > 65535 {
		return ErrInvalidHTTPServerConfigPort
	}

	if c.ShutdownTimeout.Value < 1*time.Second || c.ShutdownTimeout.Value > 600*time.Second {
		return ErrInvalidHTTPServerConfigShutdownTimeout
	}

	if c.CorsEnabled.Value {
		if c.CorsAllowedOrigins.Value == "" {
			return ErrInvalidHTTPServerConfigCorsAllowedOrigins
		}

		for _, method := range strings.Split(c.CorsAllowedMethods.Value, ",") {
			if !slices.Contains(strings.Split(ValidHTTPServerCorsAllowedMethods, "|"), strings.Trim(method, " ")) {
				return ErrInvalidHTTPServerConfigCorsAllowedMethods
			}
		}

		if len(c.CorsAllowedHeaders.Value) < 2 {
			return ErrInvalidHTTPServerConfigCorsAllowedHeaders
		}

	}

	return nil
}
