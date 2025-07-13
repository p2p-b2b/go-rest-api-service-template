package config

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	ErrCacheInvalidKind           = errors.New("invalid cache kind, must be one of [" + ValidCacheKind + "]")
	ErrCacheInvalidAddress        = errors.New("invalid cache address. Must be in the format host:port and port must be between [" + strconv.Itoa(ValidCacheMinPort) + "] and [" + strconv.Itoa(ValidCacheMaxPort) + "]")
	ErrCacheInvalidDatabaseNumber = errors.New("invalid cache database number. Must be between [" + strconv.Itoa(ValidCacheMinDatabaseNumber) + "] and [" + strconv.Itoa(ValidCacheMaxDatabaseNumber) + "]")
	ErrCacheInvalidQueryTimeout   = errors.New("invalid cache query timeout. Must be between [" + ValidCacheMinQueryTimeout.String() + "] and [" + ValidCacheMaxQueryTimeout.String() + "]")
	ErrCacheInvalidEntitiesTTL    = errors.New("invalid cache entities TTL. Must be between [" + ValidCacheMinEntitiesTTL.String() + "] and [" + ValidCacheMaxEntitiesTTL.String() + "]")
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
		return ErrCacheInvalidKind
	}

	if len(c.Addresses.Value) == 0 {
		return ErrCacheInvalidAddress
	}

	if len(c.Addresses.Value) > 0 {
		for _, addr := range c.Addresses.Value {
			parts := strings.Split(addr, ":")

			if len(parts) != 2 {
				return ErrCacheInvalidAddress
			}

			port, err := strconv.Atoi(parts[1])
			if err != nil {
				return ErrCacheInvalidAddress
			}

			if port < ValidCacheMinPort || port > ValidCacheMaxPort {
				return ErrCacheInvalidAddress
			}

			if len(parts[0]) < 3 {
				return ErrCacheInvalidAddress
			}
		}
	}

	if c.DB.Value < ValidCacheMinDatabaseNumber || c.DB.Value > ValidCacheMaxDatabaseNumber {
		return ErrCacheInvalidDatabaseNumber
	}

	if c.QueryTimeout.Value < ValidCacheMinQueryTimeout || c.QueryTimeout.Value > ValidCacheMaxQueryTimeout {
		return ErrCacheInvalidQueryTimeout
	}

	if c.EntitiesTTL.Value < ValidCacheMinEntitiesTTL || c.EntitiesTTL.Value > ValidCacheMaxEntitiesTTL {
		return ErrCacheInvalidEntitiesTTL
	}

	return nil
}
