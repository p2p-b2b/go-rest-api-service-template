package config

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestNewCacheConfig(t *testing.T) {
	config := NewCacheConfig()

	if config.Kind.Value != DefaultCacheKind {
		t.Errorf("Expected Kind to be %s, got %s", DefaultCacheKind, config.Kind.Value)
	}
	if config.Addresses.Value.String() != DefaultCacheAddress.String() {
		t.Errorf("Expected Addresses to be %s, got %s", DefaultCacheAddress.String(), config.Addresses.Value.String())
	}
	if config.Username.Value != DefaultCacheUsername {
		t.Errorf("Expected Username to be %s, got %s", DefaultCacheUsername, config.Username.Value)
	}
	if config.Password.Value != DefaultCachePassword {
		t.Errorf("Expected Password to be %s, got %s", DefaultCachePassword, config.Password.Value)
	}
	if config.DB.Value != DefaultCacheDB {
		t.Errorf("Expected DB to be %d, got %d", DefaultCacheDB, config.DB.Value)
	}
	if config.QueryTimeout.Value != DefaultCacheQueryTimeout {
		t.Errorf("Expected QueryTimeout to be %v, got %v", DefaultCacheQueryTimeout, config.QueryTimeout.Value)
	}
	if config.EntitiesTTL.Value != DefaultCacheEntitiesTTL {
		t.Errorf("Expected EntitiesTTL to be %v, got %v", DefaultCacheEntitiesTTL, config.EntitiesTTL.Value)
	}
	if config.Enabled.Value != DefaultCacheEnabled {
		t.Errorf("Expected Enabled to be %v, got %v", DefaultCacheEnabled, config.Enabled.Value)
	}
}

func TestParseEnvVars_cache(t *testing.T) {
	os.Setenv("CACHE_KIND", "valkey")
	os.Setenv("CACHE_ADDRESSES", "redis1:6379,redis2:6380")
	os.Setenv("CACHE_USERNAME", "cacheuser")
	os.Setenv("CACHE_PASSWORD", "cachepass")
	os.Setenv("CACHE_DB", "2")
	os.Setenv("CACHE_QUERY_TIMEOUT", "100ms")
	os.Setenv("CACHE_ENTITIES_TTL", "24h")
	os.Setenv("CACHE_ENABLED", "false")

	config := NewCacheConfig()
	config.ParseEnvVars()

	if config.Kind.Value != "valkey" {
		t.Errorf("Expected Kind to be valkey, got %s", config.Kind.Value)
	}
	if config.Username.Value != "cacheuser" {
		t.Errorf("Expected Username to be cacheuser, got %s", config.Username.Value)
	}
	if config.Password.Value != "cachepass" {
		t.Errorf("Expected Password to be cachepass, got %s", config.Password.Value)
	}
	if config.DB.Value != 2 {
		t.Errorf("Expected DB to be 2, got %d", config.DB.Value)
	}
	if config.QueryTimeout.Value != 100*time.Millisecond {
		t.Errorf("Expected QueryTimeout to be 100ms, got %v", config.QueryTimeout.Value)
	}
	if config.EntitiesTTL.Value != 24*time.Hour {
		t.Errorf("Expected EntitiesTTL to be 24h, got %v", config.EntitiesTTL.Value)
	}
	if config.Enabled.Value != false {
		t.Errorf("Expected Enabled to be false, got %v", config.Enabled.Value)
	}

	// Clean up environment variables
	os.Unsetenv("CACHE_KIND")
	os.Unsetenv("CACHE_ADDRESSES")
	os.Unsetenv("CACHE_USERNAME")
	os.Unsetenv("CACHE_PASSWORD")
	os.Unsetenv("CACHE_DB")
	os.Unsetenv("CACHE_QUERY_TIMEOUT")
	os.Unsetenv("CACHE_ENTITIES_TTL")
	os.Unsetenv("CACHE_ENABLED")
}

func TestValidate_cache(t *testing.T) {
	config := NewCacheConfig()

	// Test valid configuration
	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test invalid Kind
	config.Kind.Value = "invalid"
	err = config.Validate()
	var invalidErr *InvalidConfigurationError
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.kind" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.kind', got %v", err)
	}
	config.Kind.Value = DefaultCacheKind

	// Test invalid Addresses (empty)
	config.Addresses.Value = SliceStringVar{}
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.addresses" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.addresses', got %v", err)
	}
	config.Addresses.Value = DefaultCacheAddress

	// Test invalid Addresses (bad format)
	config.Addresses.Value = SliceStringVar{"invalid-address"}
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.addresses" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.addresses', got %v", err)
	}

	// Test invalid Addresses (bad port)
	config.Addresses.Value = SliceStringVar{"localhost:abc"}
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.addresses" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.addresses', got %v", err)
	}

	// Test invalid Addresses (port out of range)
	config.Addresses.Value = SliceStringVar{"localhost:99999"}
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.addresses" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.addresses', got %v", err)
	}

	// Test invalid Addresses (short hostname)
	config.Addresses.Value = SliceStringVar{"ab:6379"}
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.addresses" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.addresses', got %v", err)
	}
	config.Addresses.Value = DefaultCacheAddress

	// Test invalid DB
	config.DB.Value = -1
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.db" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.db', got %v", err)
	}
	config.DB.Value = DefaultCacheDB

	// Test invalid QueryTimeout (too short)
	config.QueryTimeout.Value = 5 * time.Millisecond
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.query.timeout" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.query.timeout', got %v", err)
	}
	config.QueryTimeout.Value = DefaultCacheQueryTimeout

	// Test invalid EntitiesTTL (too short)
	config.EntitiesTTL.Value = 30 * time.Minute
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "cache.entities.ttl" {
		t.Errorf("Expected InvalidConfigurationError with field 'cache.entities.ttl', got %v", err)
	}
	config.EntitiesTTL.Value = DefaultCacheEntitiesTTL
}
