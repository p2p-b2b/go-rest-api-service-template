package config

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
	"time"
)

const (
	ValidHTTPServerMaxPort            = 65535
	ValidHTTPServerMinPort            = 0
	ValidHTTPServerMaxShutdownTimeout = 600 * time.Second
	ValidHTTPServerMinShutdownTimeout = 1 * time.Second
	ValidHTTPServerCorsAllowedHeaders = 2
	ValidHTTPServerMaxPprofPort       = 6060
	ValidHTTPServerMinPprofPort       = 6060

	ValidHTTPServerMaxIPRateLimiterLimit       = 1000.0
	ValidHTTPServerMinIPRateLimiterLimit       = 1
	ValidHTTPServerMaxIPRateLimiterBurst       = 5000
	ValidHTTPServerMinIPRateLimiterBurst       = 1
	ValidHTTPServerMaxIPRateLimiterDeleteAfter = 1 * time.Hour
	ValidHTTPServerMinIPRateLimiterDeleteAfter = 2 * time.Second

	DefaultHTTPServerShutdownTimeout = 5 * time.Second
	DefaultHTTPServerAddress         = "localhost"
	DefaultHTTPServerPort            = 8080
	DefaultHTTPServerTLSEnabled      = false

	DefaultHTTPServerPprofPort    = 6060
	DefaultHTTPServerPprofAddress = "localhost"
	DefaultHTTPServerPprofEnabled = false

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

	DefaultHTTPServerIPRateLimiterEnabled     = true
	DefaultHTTPServerIPRateLimiterLimit       = 100.0
	DefaultHTTPServerIPRateLimiterBurst       = 300
	DefaultHTTPServerIPRateLimiterDeleteAfter = 1 * time.Minute
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
	Address                  Field[string]
	Port                     Field[int]
	ShutdownTimeout          Field[time.Duration]
	PrivateKeyFile           Field[FileVar]
	CertificateFile          Field[FileVar]
	CorsAllowedOrigins       Field[string]
	CorsAllowedMethods       Field[string]
	CorsAllowedHeaders       Field[string]
	PprofAddress             Field[string]
	IPRateLimiterLimit       Field[float64]
	IPRateLimiterBurst       Field[int]
	IPRateLimiterDeleteAfter Field[time.Duration]
	PprofPort                Field[int]
	IPRateLimiterEnabled     Field[bool]
	TLSEnabled               Field[bool]
	PprofEnabled             Field[bool]
	CorsEnabled              Field[bool]
	CorsAllowCredentials     Field[bool]
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

		PprofAddress: NewField("http.server.pprof.address", "SERVER_PPROF_ADDRESS", "Pprof Address", DefaultHTTPServerPprofAddress),
		PprofPort:    NewField("http.server.pprof.port", "SERVER_PPROF_PORT", "Pprof Port", DefaultHTTPServerPprofPort),
		PprofEnabled: NewField("http.server.pprof.enabled", "SERVER_PPROF_ENABLED", "Enable pprof. WARNING: Enable this only for debugging, it has performance impact!", DefaultHTTPServerPprofEnabled),

		CorsEnabled:          NewField("http.server.cors.enabled", "SERVER_CORS_ENABLED", "Enable CORS", DefaultHTTPServerCorsEnabled),
		CorsAllowCredentials: NewField("http.server.cors.allow.credentials", "SERVER_CORS_ALLOW_CREDENTIALS", "Allow Credentials for CORS", DefaultHTTPServerCorsAllowCredentials),
		CorsAllowedOrigins:   NewField("http.server.cors.allowed.origins", "SERVER_CORS_ALLOWED_ORIGINS", "Allowed Origins for CORS", DefaultHTTPServerCorsAllowedOrigins),
		CorsAllowedMethods:   NewField("http.server.cors.allowed.methods", "SERVER_CORS_ALLOWED_METHODS", "Allowed Methods for CORS", DefaultHTTPServerCorsAllowedMethods),
		CorsAllowedHeaders:   NewField("http.server.cors.allowed.headers", "SERVER_CORS_ALLOWED_HEADERS", "Allowed Headers for CORS", DefaultHTTPServerCorsAllowedHeaders),

		IPRateLimiterEnabled:     NewField("http.server.ip.rate.limiter.enabled", "SERVER_IP_RATE_LIMITER_ENABLED", "Enable IP Rate Limiter", DefaultHTTPServerIPRateLimiterEnabled),
		IPRateLimiterLimit:       NewField("http.server.ip.rate.limiter.limit", "SERVER_IP_RATE_LIMITER_LIMIT", "IP Rate Limiter Limit.  The number of requests allowed per second per client ip over each API endpoint", DefaultHTTPServerIPRateLimiterLimit),
		IPRateLimiterBurst:       NewField("http.server.ip.rate.limiter.burst", "SERVER_IP_RATE_LIMITER_BURST", "IP Rate Limiter Burst. The number of requests allowed per second per client ip in a short burst when the rate limit is reached", DefaultHTTPServerIPRateLimiterBurst),
		IPRateLimiterDeleteAfter: NewField("http.server.ip.rate.limiter.delete.after", "SERVER_IP_RATE_LIMITER_DELETE_AFTER", "IP Rate Limiter Delete After. The time after which the IP rate limiter will be deleted if no requests are made", DefaultHTTPServerIPRateLimiterDeleteAfter),
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

	c.PprofAddress.Value = GetEnv(c.PprofAddress.EnVarName, c.PprofAddress.Value)
	c.PprofPort.Value = GetEnv(c.PprofPort.EnVarName, c.PprofPort.Value)
	c.PprofEnabled.Value = GetEnv(c.PprofEnabled.EnVarName, c.PprofEnabled.Value)

	c.CorsEnabled.Value = GetEnv(c.CorsEnabled.EnVarName, c.CorsEnabled.Value)
	c.CorsAllowCredentials.Value = GetEnv(c.CorsAllowCredentials.EnVarName, c.CorsAllowCredentials.Value)
	c.CorsAllowedOrigins.Value = GetEnv(c.CorsAllowedOrigins.EnVarName, c.CorsAllowedOrigins.Value)
	c.CorsAllowedMethods.Value = GetEnv(c.CorsAllowedMethods.EnVarName, c.CorsAllowedMethods.Value)
	c.CorsAllowedHeaders.Value = GetEnv(c.CorsAllowedHeaders.EnVarName, c.CorsAllowedHeaders.Value)

	c.IPRateLimiterEnabled.Value = GetEnv(c.IPRateLimiterEnabled.EnVarName, c.IPRateLimiterEnabled.Value)
	c.IPRateLimiterLimit.Value = GetEnv(c.IPRateLimiterLimit.EnVarName, c.IPRateLimiterLimit.Value)
	c.IPRateLimiterBurst.Value = GetEnv(c.IPRateLimiterBurst.EnVarName, c.IPRateLimiterBurst.Value)
	c.IPRateLimiterDeleteAfter.Value = GetEnv(c.IPRateLimiterDeleteAfter.EnVarName, c.IPRateLimiterDeleteAfter.Value)
}

// Validate validates the server configuration values
func (c *HTTPServerConfig) Validate() error {
	if c.Address.Value == "" || (c.Address.Value != "localhost" && net.ParseIP(c.Address.Value) == nil) {
		return &InvalidConfigurationError{
			Field:   "http.server.address",
			Value:   c.Address.Value,
			Message: "invalid http.server.address, must be a valid IP address or hostname",
		}
	}

	// validate the if is a valid IP Address or Hostname

	if c.Port.Value < ValidHTTPServerMinPort || c.Port.Value > ValidHTTPServerMaxPort || c.Port.Value == c.PprofPort.Value {
		return &InvalidConfigurationError{
			Field:   "http.server.port",
			Value:   fmt.Sprintf("%d", c.Port.Value),
			Message: fmt.Sprintf("invalid http.server.port, must be between %d and %d and not equal to http.server.pprof.port", ValidHTTPServerMinPort, ValidHTTPServerMaxPort),
		}
	}

	if c.ShutdownTimeout.Value < ValidHTTPServerMinShutdownTimeout || c.ShutdownTimeout.Value > ValidHTTPServerMaxShutdownTimeout {
		return &InvalidConfigurationError{
			Field:   "http.server.shutdown.timeout",
			Value:   fmt.Sprintf("%d", c.ShutdownTimeout.Value),
			Message: fmt.Sprintf("invalid http.server.shutdown.timeout, must be between %d and %d", ValidHTTPServerMinShutdownTimeout, ValidHTTPServerMaxShutdownTimeout),
		}
	}

	if c.CorsEnabled.Value {
		if c.CorsAllowedOrigins.Value == "" {
			return &InvalidConfigurationError{
				Field:   "http.server.cors.allowed.origins",
				Value:   c.CorsAllowedOrigins.Value,
				Message: "invalid http.server.cors.allowed.origins, must be a non-empty string",
			}
		}

		for method := range strings.SplitSeq(c.CorsAllowedMethods.Value, ",") {
			if !slices.Contains(strings.Split(ValidHTTPServerCorsAllowedMethods, "|"), strings.Trim(method, " ")) {
				return &InvalidConfigurationError{
					Field:   "http.server.cors.allowed.methods",
					Value:   c.CorsAllowedMethods.Value,
					Message: fmt.Sprintf("invalid http.server.cors.allowed.methods, must be one of: %s", ValidHTTPServerCorsAllowedMethods),
				}
			}
		}

		if len(c.CorsAllowedHeaders.Value) < ValidHTTPServerCorsAllowedHeaders {
			return &InvalidConfigurationError{
				Field:   "http.server.cors.allowed.headers",
				Value:   c.CorsAllowedHeaders.Value,
				Message: fmt.Sprintf("invalid http.server.cors.allowed.headers, must be at least %d elements", ValidHTTPServerCorsAllowedHeaders),
			}
		}
	}

	if c.PprofEnabled.Value {
		if c.PprofPort.Value < ValidHTTPServerMinPprofPort || c.PprofPort.Value > ValidHTTPServerMaxPprofPort || c.Port.Value == c.PprofPort.Value {
			return &InvalidConfigurationError{
				Field:   "http.server.pprof.port",
				Value:   fmt.Sprintf("%d", c.PprofPort.Value),
				Message: fmt.Sprintf("invalid http.server.pprof.port, must be between %d and %d and not equal to http.server.port", ValidHTTPServerMinPprofPort, ValidHTTPServerMaxPprofPort),
			}
		}

		if c.PprofAddress.Value == "" || (c.PprofAddress.Value != "localhost" && net.ParseIP(c.PprofAddress.Value) == nil) {
			return &InvalidConfigurationError{
				Field:   "http.server.pprof.address",
				Value:   c.PprofAddress.Value,
				Message: "invalid http.server.pprof.address, must be a valid IP address or hostname",
			}
		}
	}

	if c.IPRateLimiterLimit.Value < ValidHTTPServerMinIPRateLimiterLimit || c.IPRateLimiterLimit.Value > ValidHTTPServerMaxIPRateLimiterLimit {
		return &InvalidConfigurationError{
			Field:   "http.server.ip.rate.limiter.limit",
			Value:   fmt.Sprintf("%.1f", c.IPRateLimiterLimit.Value),
			Message: fmt.Sprintf("invalid http.server.ip.rate.limiter.limit, must be between %v and %v", ValidHTTPServerMinIPRateLimiterLimit, ValidHTTPServerMaxIPRateLimiterLimit),
		}
	}

	if c.IPRateLimiterBurst.Value < ValidHTTPServerMinIPRateLimiterBurst || c.IPRateLimiterBurst.Value > ValidHTTPServerMaxIPRateLimiterBurst {
		return &InvalidConfigurationError{
			Field:   "http.server.ip.rate.limiter.burst",
			Value:   fmt.Sprintf("%d", c.IPRateLimiterBurst.Value),
			Message: fmt.Sprintf("invalid http.server.ip.rate.limiter.burst, must be between %d and %d", ValidHTTPServerMinIPRateLimiterBurst, ValidHTTPServerMaxIPRateLimiterBurst),
		}
	}

	if c.IPRateLimiterDeleteAfter.Value < ValidHTTPServerMinIPRateLimiterDeleteAfter || c.IPRateLimiterDeleteAfter.Value > ValidHTTPServerMaxIPRateLimiterDeleteAfter {
		return &InvalidConfigurationError{
			Field:   "http.server.ip.rate.limiter.delete.after",
			Value:   fmt.Sprintf("%d", c.IPRateLimiterDeleteAfter.Value),
			Message: fmt.Sprintf("invalid http.server.ip.rate.limiter.delete.after, must be between %d and %d", ValidHTTPServerMinIPRateLimiterDeleteAfter, ValidHTTPServerMaxIPRateLimiterDeleteAfter),
		}
	}

	return nil
}
