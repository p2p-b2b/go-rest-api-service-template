package config

import (
	"errors"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	ErrHTTPServerInvalidAddress            = errors.New("invalid server address, must not be empty and a valid IP Address or Hostname")
	ErrHTTPServerInvalidPort               = errors.New("invalid server port, must be between [" + strconv.Itoa(ValidHTTPServerMinPort) + "] and [" + strconv.Itoa(ValidHTTPServerMaxPort) + "]")
	ErrHTTPServerInvalidShutdownTimeout    = errors.New("invalid server shutdown timeout, must be between [" + ValidHTTPServerMinShutdownTimeout.String() + "] and [" + ValidHTTPServerMaxShutdownTimeout.String() + "]")
	ErrHTTPServerInvalidCorsAllowedOrigins = errors.New("invalid CORS allowed origins. Must not be empty")
	ErrHTTPServerInvalidCorsAllowedMethods = errors.New("invalid CORS allowed methods. Must be one of [" + ValidHTTPServerCorsAllowedMethods + "]")
	ErrHTTPServerInvalidCorsAllowedHeaders = errors.New("invalid CORS allowed headers. Must be at least [" + strconv.Itoa(ValidHTTPServerCorsAllowedHeaders) + "]")
)

const (
	ValidHTTPServerMaxPort            = 65535
	ValidHTTPServerMinPort            = 0
	ValidHTTPServerMaxShutdownTimeout = 600 * time.Second
	ValidHTTPServerMinShutdownTimeout = 1 * time.Second
	ValidHTTPServerCorsAllowedHeaders = 2

	DefaultHTTPServerShutdownTimeout = 5 * time.Second
	DefaultHTTPServerAddress         = "localhost"
	DefaultHTTPServerPort            = 8080
	DefaultHTTPServerTLSEnabled      = false
	DefaultHTTPServerPprofEnabled    = false

	// DefaultHTTPServerCorsEnabled is the default value for enabling CORS
	// If enabled, the server will use the following values for CORS
	// - AllowedOrigins: "*"
	// - AllowedMethods: "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD"
	// - AllowedHeaders: "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-Requested-With, X-Api-Version"
	// Remember to change the values if you need to restrict the allowed origins, methods or headers
	DefaultHTTPServerCorsEnabled          = false
	DefaultHTTPServerCorsAllowCredentials = true

	// DefaultHTTPServerCorsAllowedOrigins is the default value for allowed origins
	// Could be a comma separated list of origins. Example: "http://localhost:3000, http://localhost:8080"
	DefaultHTTPServerCorsAllowedOrigins = "*" // allow all origins

	DefaultHTTPServerCorsAllowedMethods = "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD"
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
	if c.Address.Value == "" || (c.Address.Value != "localhost" && net.ParseIP(c.Address.Value) == nil) {
		return ErrHTTPServerInvalidAddress
	}

	// validate the if is a valid IP Address or Hostname

	if c.Port.Value < ValidHTTPServerMinPort || c.Port.Value > ValidHTTPServerMaxPort {
		return ErrHTTPServerInvalidPort
	}

	if c.ShutdownTimeout.Value < ValidHTTPServerMinShutdownTimeout || c.ShutdownTimeout.Value > ValidHTTPServerMaxShutdownTimeout {
		return ErrHTTPServerInvalidShutdownTimeout
	}

	if c.CorsEnabled.Value {
		if c.CorsAllowedOrigins.Value == "" {
			return ErrHTTPServerInvalidCorsAllowedOrigins
		}

		for _, method := range strings.Split(c.CorsAllowedMethods.Value, ",") {
			if !slices.Contains(strings.Split(ValidHTTPServerCorsAllowedMethods, "|"), strings.Trim(method, " ")) {
				return ErrHTTPServerInvalidCorsAllowedMethods
			}
		}

		if len(c.CorsAllowedHeaders.Value) < ValidHTTPServerCorsAllowedHeaders {
			return ErrHTTPServerInvalidCorsAllowedHeaders
		}

	}

	return nil
}
