package config

import (
	"os"
	"testing"
	"time"
)

func TestNewHTTPServerConfig(t *testing.T) {
	config := NewHTTPServerConfig()

	if config.Address.Value != DefaultHTTPServerAddress {
		t.Errorf("Expected Address to be %s, got %s", DefaultHTTPServerAddress, config.Address.Value)
	}
	if config.Port.Value != DefaultHTTPServerPort {
		t.Errorf("Expected Port to be %d, got %d", DefaultHTTPServerPort, config.Port.Value)
	}
	if config.ShutdownTimeout.Value != DefaultHTTPServerShutdownTimeout {
		t.Errorf("Expected ShutdownTimeout to be %v, got %v", DefaultHTTPServerShutdownTimeout, config.ShutdownTimeout.Value)
	}
	if config.TLSEnabled.Value != DefaultHTTPServerTLSEnabled {
		t.Errorf("Expected TLSEnabled to be %v, got %v", DefaultHTTPServerTLSEnabled, config.TLSEnabled.Value)
	}
	if config.PprofEnabled.Value != DefaultHTTPServerPprofEnabled {
		t.Errorf("Expected PprofEnabled to be %v, got %v", DefaultHTTPServerPprofEnabled, config.PprofEnabled.Value)
	}
	if config.CorsEnabled.Value != DefaultHTTPServerCorsEnabled {
		t.Errorf("Expected CorsEnabled to be %v, got %v", DefaultHTTPServerCorsEnabled, config.CorsEnabled.Value)
	}
	if config.CorsAllowCredentials.Value != DefaultHTTPServerCorsAllowCredentials {
		t.Errorf("Expected CorsAllowCredentials to be %v, got %v", DefaultHTTPServerCorsAllowCredentials, config.CorsAllowCredentials.Value)
	}
	if config.CorsAllowedOrigins.Value != DefaultHTTPServerCorsAllowedOrigins {
		t.Errorf("Expected CorsAllowedOrigins to be %s, got %s", DefaultHTTPServerCorsAllowedOrigins, config.CorsAllowedOrigins.Value)
	}
	if config.CorsAllowedMethods.Value != DefaultHTTPServerCorsAllowedMethods {
		t.Errorf("Expected CorsAllowedMethods to be %s, got %s", DefaultHTTPServerCorsAllowedMethods, config.CorsAllowedMethods.Value)
	}
	if config.CorsAllowedHeaders.Value != DefaultHTTPServerCorsAllowedHeaders {
		t.Errorf("Expected CorsAllowedHeaders to be %s, got %s", DefaultHTTPServerCorsAllowedHeaders, config.CorsAllowedHeaders.Value)
	}
}

func TestParseEnvVars_httpserver(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", "127.0.0.1")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_SHUTDOWN_TIMEOUT", "10s")
	os.Setenv("SERVER_TLS_ENABLED", "true")
	os.Setenv("SERVER_PPROF_ENABLED", "true")
	os.Setenv("SERVER_CORS_ENABLED", "true")
	os.Setenv("SERVER_CORS_ALLOW_CREDENTIALS", "false")
	os.Setenv("SERVER_CORS_ALLOWED_ORIGINS", "http://example.com")
	os.Setenv("SERVER_CORS_ALLOWED_METHODS", "GET,POST")
	os.Setenv("SERVER_CORS_ALLOWED_HEADERS", "Content-Type,Authorization")

	config := NewHTTPServerConfig()
	config.ParseEnvVars()

	if config.Address.Value != "127.0.0.1" {
		t.Errorf("Expected Address to be 127.0.0.1, got %s", config.Address.Value)
	}
	if config.Port.Value != 9090 {
		t.Errorf("Expected Port to be 9090, got %d", config.Port.Value)
	}
	if config.ShutdownTimeout.Value != 10*time.Second {
		t.Errorf("Expected ShutdownTimeout to be 10s, got %v", config.ShutdownTimeout.Value)
	}
	if config.TLSEnabled.Value != true {
		t.Errorf("Expected TLSEnabled to be true, got %v", config.TLSEnabled.Value)
	}
	if config.PprofEnabled.Value != true {
		t.Errorf("Expected PprofEnabled to be true, got %v", config.PprofEnabled.Value)
	}
	if config.CorsEnabled.Value != true {
		t.Errorf("Expected CorsEnabled to be true, got %v", config.CorsEnabled.Value)
	}
	if config.CorsAllowCredentials.Value != false {
		t.Errorf("Expected CorsAllowCredentials to be false, got %v", config.CorsAllowCredentials.Value)
	}
	if config.CorsAllowedOrigins.Value != "http://example.com" {
		t.Errorf("Expected CorsAllowedOrigins to be http://example.com, got %s", config.CorsAllowedOrigins.Value)
	}
	if config.CorsAllowedMethods.Value != "GET,POST" {
		t.Errorf("Expected CorsAllowedMethods to be GET,POST, got %s", config.CorsAllowedMethods.Value)
	}
	if config.CorsAllowedHeaders.Value != "Content-Type,Authorization" {
		t.Errorf("Expected CorsAllowedHeaders to be Content-Type,Authorization, got %s", config.CorsAllowedHeaders.Value)
	}

	// Clean up environment variables
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_SHUTDOWN_TIMEOUT")
	os.Unsetenv("SERVER_TLS_ENABLED")
	os.Unsetenv("SERVER_PPROF_ENABLED")
	os.Unsetenv("SERVER_CORS_ENABLED")
	os.Unsetenv("SERVER_CORS_ALLOW_CREDENTIALS")
	os.Unsetenv("SERVER_CORS_ALLOWED_ORIGINS")
	os.Unsetenv("SERVER_CORS_ALLOWED_METHODS")
	os.Unsetenv("SERVER_CORS_ALLOWED_HEADERS")
}

func TestValidate_httpserver(t *testing.T) {
	config := NewHTTPServerConfig()

	// Test valid configuration
	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test invalid Address
	config.Address.Value = ""
	err = config.Validate()
	if err != ErrHTTPServerInvalidAddress {
		t.Errorf("Expected error %v, got %v", ErrHTTPServerInvalidAddress, err)
	}
	config.Address.Value = DefaultHTTPServerAddress

	// Test invalid Port
	config.Port.Value = -1
	err = config.Validate()
	if err != ErrHTTPServerInvalidPort {
		t.Errorf("Expected error %v, got %v", ErrHTTPServerInvalidPort, err)
	}
	config.Port.Value = DefaultHTTPServerPort

	// Test invalid ShutdownTimeout
	config.ShutdownTimeout.Value = 0
	err = config.Validate()
	if err != ErrHTTPServerInvalidShutdownTimeout {
		t.Errorf("Expected error %v, got %v", ErrHTTPServerInvalidShutdownTimeout, err)
	}
	config.ShutdownTimeout.Value = DefaultHTTPServerShutdownTimeout

	// Test invalid CORS configuration
	config.CorsEnabled.Value = true
	config.CorsAllowedOrigins.Value = ""
	err = config.Validate()
	if err != ErrHTTPServerInvalidCorsAllowedOrigins {
		t.Errorf("Expected error %v, got %v", ErrHTTPServerInvalidCorsAllowedOrigins, err)
	}
	config.CorsAllowedOrigins.Value = DefaultHTTPServerCorsAllowedOrigins

	config.CorsAllowedMethods.Value = "INVALID"
	err = config.Validate()
	if err != ErrHTTPServerInvalidCorsAllowedMethods {
		t.Errorf("Expected error %v, got %v", ErrHTTPServerInvalidCorsAllowedMethods, err)
	}
	config.CorsAllowedMethods.Value = DefaultHTTPServerCorsAllowedMethods

	config.CorsAllowedHeaders.Value = "A"
	err = config.Validate()
	if err != ErrHTTPServerInvalidCorsAllowedHeaders {
		t.Errorf("Expected error %v, got %v", ErrHTTPServerInvalidCorsAllowedHeaders, err)
	}
	config.CorsAllowedHeaders.Value = DefaultHTTPServerCorsAllowedHeaders
}
