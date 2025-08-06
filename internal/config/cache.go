package config

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

var DefaultCacheAddress = SliceStringVar{"localhost:6379"}

const (
	ValidCacheMaxDatabaseNumber = 16
	ValidCacheMinDatabaseNumber = 0
	ValidCacheKind              = "valkey"
	ValidCacheMaxPort           = 65535
	ValidCacheMinPort           = 0
	ValidCacheMaxQueryTimeout   = 1000 * time.Millisecond
	ValidCacheMinQueryTimeout   = 10 * time.Millisecond
	ValidCacheMaxEntitiesTTL    = 72 * time.Hour
	ValidCacheMinEntitiesTTL    = 1 * time.Hour

	DefaultCacheKind         = "valkey"
	DefaultCacheUsername     = ""
	DefaultCachePassword     = ""
	DefaultCacheDB           = 0
	DefaultCacheQueryTimeout = 80 * time.Millisecond
	DefaultCacheEnabled      = true
	DefaultCacheEntitiesTTL  = 12 * time.Hour
)

type CacheConfig struct {
	Kind         Field[string]
	Addresses    Field[SliceStringVar]
	Username     Field[string]
	Password     Field[string]
	DB           Field[int]
	QueryTimeout Field[time.Duration]
	EntitiesTTL  Field[time.Duration]
	Enabled      Field[bool]
}

func NewCacheConfig() *CacheConfig {
	return &CacheConfig{
		Kind:         NewField("cache.kind", "CACHE_KIND", "Cache Kind. Possible values ["+ValidCacheKind+"]", DefaultCacheKind),
		Addresses:    NewField("cache.addresses", "CACHE_ADDRESSES", "Cache Server Addresses. List of host:port, Example: --cache.addresses=host1:port1 --cache.addresses=host2:port2", DefaultCacheAddress),
		Username:     NewField("cache.username", "CACHE_USERNAME", "Cache Server Username", DefaultCacheUsername),
		Password:     NewField("cache.password", "CACHE_PASSWORD", "Cache Server Password", DefaultCachePassword),
		DB:           NewField("cache.db", "CACHE_DB", "Cache Server DB number", DefaultCacheDB),
		QueryTimeout: NewField("cache.query.timeout", "CACHE_QUERY_TIMEOUT", "Cache Query Timeout", DefaultCacheQueryTimeout),
		EntitiesTTL:  NewField("cache.entities.ttl", "CACHE_ENTITIES_TTL", "TTl for the cache entities", DefaultCacheEntitiesTTL),
		Enabled:      NewField("cache.enabled", "CACHE_ENABLED", "Cache Server Enabled", DefaultCacheEnabled),
	}
}

func (c *CacheConfig) ParseEnvVars() {
	c.Kind.Value = GetEnv(c.Kind.EnVarName, c.Kind.Value)
	c.Addresses.Value = GetEnv(c.Addresses.EnVarName, c.Addresses.Value)
	c.Username.Value = GetEnv(c.Username.EnVarName, c.Username.Value)
	c.Password.Value = GetEnv(c.Password.EnVarName, c.Password.Value)
	c.DB.Value = GetEnv(c.DB.EnVarName, c.DB.Value)
	c.QueryTimeout.Value = GetEnv(c.QueryTimeout.EnVarName, c.QueryTimeout.Value)
	c.EntitiesTTL.Value = GetEnv(c.EntitiesTTL.EnVarName, c.EntitiesTTL.Value)
	c.Enabled.Value = GetEnv(c.Enabled.EnVarName, c.Enabled.Value)
}

func (c *CacheConfig) Validate() error {
	if !slices.Contains(strings.Split(ValidCacheKind, "|"), c.Kind.Value) {
		return &InvalidConfigurationError{
			Field:   "cache.kind",
			Value:   c.Kind.Value,
			Message: "invalid cache kind, must be one of [" + ValidCacheKind + "]",
		}
	}

	if len(c.Addresses.Value) == 0 {
		return &InvalidConfigurationError{
			Field:   "cache.addresses",
			Value:   c.Addresses.Value.String(),
			Message: "invalid cache addresses, must be a list of host:port",
		}
	}

	if len(c.Addresses.Value) > 0 {
		for _, addr := range c.Addresses.Value {
			parts := strings.Split(addr, ":")

			if len(parts) != 2 {
				return &InvalidConfigurationError{
					Field:   "cache.addresses",
					Value:   c.Addresses.Value.String(),
					Message: "invalid cache address, must be in the format host:port",
				}
			}

			port, err := strconv.Atoi(parts[1])
			if err != nil {
				return &InvalidConfigurationError{
					Field:   "cache.addresses",
					Value:   c.Addresses.Value.String(),
					Message: "invalid cache address port, must be a number",
				}
			}

			if port < ValidCacheMinPort || port > ValidCacheMaxPort {
				return &InvalidConfigurationError{
					Field:   "cache.addresses",
					Value:   c.Addresses.Value.String(),
					Message: fmt.Sprintf("invalid cache address port, must be between %d and %d", ValidCacheMinPort, ValidCacheMaxPort),
				}
			}

			if len(parts[0]) < 3 {
				return &InvalidConfigurationError{
					Field:   "cache.addresses",
					Value:   c.Addresses.Value.String(),
					Message: "invalid cache address, must be at least 3 characters",
				}
			}
		}
	}

	if c.DB.Value < ValidCacheMinDatabaseNumber || c.DB.Value > ValidCacheMaxDatabaseNumber {
		return &InvalidConfigurationError{
			Field:   "cache.db",
			Value:   fmt.Sprintf("%d", c.DB.Value),
			Message: fmt.Sprintf("invalid cache db number, must be between %d and %d", ValidCacheMinDatabaseNumber, ValidCacheMaxDatabaseNumber),
		}
	}

	if c.QueryTimeout.Value < ValidCacheMinQueryTimeout || c.QueryTimeout.Value > ValidCacheMaxQueryTimeout {
		return &InvalidConfigurationError{
			Field:   "cache.query.timeout",
			Value:   fmt.Sprintf("%d", c.QueryTimeout.Value),
			Message: fmt.Sprintf("invalid cache query timeout, must be between %d and %d", ValidCacheMinQueryTimeout, ValidCacheMaxQueryTimeout),
		}
	}

	if c.EntitiesTTL.Value < ValidCacheMinEntitiesTTL || c.EntitiesTTL.Value > ValidCacheMaxEntitiesTTL {
		return &InvalidConfigurationError{
			Field:   "cache.entities.ttl",
			Value:   fmt.Sprintf("%d", c.EntitiesTTL.Value),
			Message: fmt.Sprintf("invalid cache entities ttl, must be between %d and %d", ValidCacheMinEntitiesTTL, ValidCacheMaxEntitiesTTL),
		}
	}

	return nil
}
