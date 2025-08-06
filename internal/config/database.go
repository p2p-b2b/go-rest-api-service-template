package config

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	ValidDatabaseKind           = "pgxpool"
	ValidDatabaseSSLModes       = "disable|allow|prefer|require|verify-ca|verify-full"
	ValidDatabaseMaxPort        = 65535
	ValidDatabaseMinPort        = 0
	ValidDatabaseUsernameMaxLen = 32
	ValidDatabaseUsernameMinLen = 2
	ValidDatabasePasswordMaxLen = 128
	ValidDatabasePasswordMinLen = 2
	ValidDatabaseNameMaxLen     = 32
	ValidDatabaseNameMinLen     = 2
	ValidDatabaseTimeZoneMaxLen = 32
	ValidDatabaseTimeZoneMinLen = 2

	ValidDatabaseMaxMaxConns = 200
	ValidDatabaseMinMaxConns = 10
	ValidDatabaseMaxMinConns = 10
	ValidDatabaseMinMinConns = 0

	ValidDatabaseMaxPingTimeout = 30 * time.Second
	ValidDatabaseMinPingTimeout = 1 * time.Second

	ValidDatabaseMaxQueryTimeout = 30 * time.Second
	ValidDatabaseMinQueryTimeout = 1 * time.Second

	ValidDatabaseConnMaxIdleTime = 8 * time.Hour
	ValidDatabaseConnMinIdleTime = 1 * time.Minute

	ValidDatabaseConnMaxLifetime = 8 * time.Hour
	ValidDatabaseConnMinLifetime = 1 * time.Minute

	DefaultDatabaseKind     = "pgxpool"
	DefaultDatabaseAddress  = "localhost"
	DefaultDatabasePort     = 5432
	DefaultDatabaseUsername = "username"
	DefaultDatabasePassword = "password"
	DefaultDatabaseName     = "svc-qu3ry-core"
	DefaultDatabaseSSLMode  = "disable"
	DefaultDatabaseTimeZone = "UTC"

	DefaultDatabaseMaxPingTimeout  = 5 * time.Second
	DefaultDatabaseMaxQueryTimeout = 5 * time.Second

	DefaultDatabaseMaxConns = 20
	DefaultDatabaseMinConns = 5

	DefaultDatabaseConnMaxIdleTime = 30 * time.Minute
	DefaultDatabaseConnMaxLifetime = 5 * time.Minute

	DefaultDatabaseMigrationEnable = true
)

type DatabaseConfig struct {
	Kind     Field[string]
	Address  Field[string]
	Username Field[string]
	Password Field[string]
	Name     Field[string]
	SSLMode  Field[string]
	Port     Field[int]
	TimeZone Field[string]

	MaxConns Field[int]
	MinConns Field[int]

	MaxQueryTimeout Field[time.Duration]
	MaxPingTimeout  Field[time.Duration]

	ConnMaxIdleTime Field[time.Duration]
	ConnMaxLifetime Field[time.Duration]

	MigrationEnable Field[bool]
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Kind:     NewField("database.kind", "DATABASE_KIND", "Database Kind. Possible values ["+ValidDatabaseKind+"]", DefaultDatabaseKind),
		Address:  NewField("database.address", "DATABASE_ADDRESS", "Database IP Address or Hostname", DefaultDatabaseAddress),
		Port:     NewField("database.port", "DATABASE_PORT", "Database Port", DefaultDatabasePort),
		Username: NewField("database.username", "DATABASE_USERNAME", "Database Username", DefaultDatabaseUsername),
		Password: NewField("database.password", "DATABASE_PASSWORD", "Database Password", DefaultDatabasePassword),
		Name:     NewField("database.name", "DATABASE_NAME", "Database Name", DefaultDatabaseName),
		SSLMode:  NewField("database.ssl.mode", "DATABASE_SSL_MODE", "Database SSL Mode. Possible values ["+ValidDatabaseSSLModes+"]", DefaultDatabaseSSLMode),
		TimeZone: NewField("database.time.zone", "DATABASE_TIME_ZONE", "Database Time Zone", DefaultDatabaseTimeZone),

		MaxPingTimeout:  NewField("database.max.ping.timeout", "DATABASE_MAX_PING_TIMEOUT", "Database Max Ping Timeout", DefaultDatabaseMaxPingTimeout),
		MaxQueryTimeout: NewField("database.max.query.timeout", "DATABASE_MAX_QUERY_TIMEOUT", "Database Max Query Timeout", DefaultDatabaseMaxQueryTimeout),

		MaxConns: NewField("database.max.conns", "DATABASE_MAX_CONNS", "Database Max Idle Connections", DefaultDatabaseMaxConns),
		MinConns: NewField("database.min.conns", "DATABASE_MIN_CONNS", "Database Max Open Connections", DefaultDatabaseMinConns),

		ConnMaxIdleTime: NewField("database.conn.max.idle.time", "DATABASE_CONN_MAX_IDLE_TIME", "Database Connection Max Idle Time", DefaultDatabaseConnMaxIdleTime),
		ConnMaxLifetime: NewField("database.conn.max.lifetime", "DATABASE_CONN_MAX_LIFETIME", "Database Connection Max Lifetime", DefaultDatabaseConnMaxLifetime),

		MigrationEnable: NewField("database.migration.enable", "DATABASE_MIGRATION_ENABLE", "Database migration is enables?", DefaultDatabaseMigrationEnable),
	}
}

// ParseEnvVars reads the database configuration from environment variables
// and sets the values in the configuration
func (c *DatabaseConfig) ParseEnvVars() {
	c.Kind.Value = GetEnv(c.Kind.EnVarName, c.Kind.Value)
	c.Address.Value = GetEnv(c.Address.EnVarName, c.Address.Value)
	c.Port.Value = GetEnv(c.Port.EnVarName, c.Port.Value)
	c.Username.Value = GetEnv(c.Username.EnVarName, c.Username.Value)
	c.Password.Value = GetEnv(c.Password.EnVarName, c.Password.Value)
	c.Name.Value = GetEnv(c.Name.EnVarName, c.Name.Value)
	c.SSLMode.Value = GetEnv(c.SSLMode.EnVarName, c.SSLMode.Value)
	c.TimeZone.Value = GetEnv(c.TimeZone.EnVarName, c.TimeZone.Value)

	c.MaxPingTimeout.Value = GetEnv(c.MaxPingTimeout.EnVarName, c.MaxPingTimeout.Value)
	c.MaxQueryTimeout.Value = GetEnv(c.MaxQueryTimeout.EnVarName, c.MaxQueryTimeout.Value)

	c.MaxConns.Value = GetEnv(c.MaxConns.EnVarName, c.MaxConns.Value)
	c.MinConns.Value = GetEnv(c.MinConns.EnVarName, c.MinConns.Value)

	c.ConnMaxIdleTime.Value = GetEnv(c.ConnMaxIdleTime.EnVarName, c.ConnMaxIdleTime.Value)
	c.ConnMaxLifetime.Value = GetEnv(c.ConnMaxLifetime.EnVarName, c.ConnMaxLifetime.Value)

	c.MigrationEnable.Value = GetEnv(c.MigrationEnable.EnVarName, c.MigrationEnable.Value)
}

// Validate validates the database configuration values
func (c *DatabaseConfig) Validate() error {
	if !slices.Contains(strings.Split(ValidDatabaseKind, "|"), c.Kind.Value) {
		return &InvalidConfigurationError{
			Field:   "database.kind",
			Value:   c.Kind.Value,
			Message: fmt.Sprintf("invalid database kind, must be one of: %s", ValidDatabaseKind),
		}
	}

	if c.Port.Value <= ValidDatabaseMinPort || c.Port.Value >= ValidDatabaseMaxPort {
		return &InvalidConfigurationError{
			Field:   "database.port",
			Value:   fmt.Sprintf("%d", c.Port.Value),
			Message: fmt.Sprintf("invalid database port, must be between %d and %d", ValidDatabaseMinPort, ValidDatabaseMaxPort),
		}
	}

	if c.Username.Value == "" || len(c.Username.Value) < ValidDatabaseUsernameMinLen || len(c.Username.Value) > ValidDatabaseUsernameMaxLen {
		return &InvalidConfigurationError{
			Field:   "database.username",
			Value:   c.Username.Value,
			Message: fmt.Sprintf("invalid database username, must be between %d and %d characters", ValidDatabaseUsernameMinLen, ValidDatabaseUsernameMaxLen),
		}
	}

	if c.Password.Value == "" || len(c.Password.Value) < ValidDatabasePasswordMinLen || len(c.Password.Value) > ValidDatabasePasswordMaxLen {
		return &InvalidConfigurationError{
			Field:   "database.password",
			Value:   c.Password.Value,
			Message: fmt.Sprintf("invalid database password, must be between %d and %d characters", ValidDatabasePasswordMinLen, ValidDatabasePasswordMaxLen),
		}
	}

	if c.Name.Value == "" || len(c.Name.Value) < ValidDatabaseNameMinLen || len(c.Name.Value) > ValidDatabaseNameMaxLen {
		return &InvalidConfigurationError{
			Field:   "database.name",
			Value:   c.Name.Value,
			Message: fmt.Sprintf("invalid database name, must be between %d and %d characters", ValidDatabaseNameMinLen, ValidDatabaseNameMaxLen),
		}
	}

	if !slices.Contains(strings.Split(ValidDatabaseSSLModes, "|"), c.SSLMode.Value) {
		return &InvalidConfigurationError{
			Field:   "database.sslmode",
			Value:   c.SSLMode.Value,
			Message: fmt.Sprintf("invalid database SSL mode, must be one of: %s", ValidDatabaseSSLModes),
		}
	}

	if c.TimeZone.Value == "" || len(c.TimeZone.Value) < ValidDatabaseTimeZoneMinLen || len(c.TimeZone.Value) > ValidDatabaseTimeZoneMaxLen {
		return &InvalidConfigurationError{
			Field:   "database.timezone",
			Value:   c.TimeZone.Value,
			Message: fmt.Sprintf("invalid database timezone, must be between %d and %d characters", ValidDatabaseTimeZoneMinLen, ValidDatabaseTimeZoneMaxLen),
		}
	}

	if c.MaxConns.Value < ValidDatabaseMinMaxConns || c.MaxConns.Value > ValidDatabaseMaxMaxConns {
		return &InvalidConfigurationError{
			Field:   "database.max_conns",
			Value:   fmt.Sprintf("%d", c.MaxConns.Value),
			Message: fmt.Sprintf("invalid database max connections, must be between %d and %d", ValidDatabaseMinMaxConns, ValidDatabaseMaxMaxConns),
		}
	}

	if c.MinConns.Value < ValidDatabaseMinMinConns || c.MinConns.Value > ValidDatabaseMaxMinConns {
		return &InvalidConfigurationError{
			Field:   "database.min_conns",
			Value:   fmt.Sprintf("%d", c.MinConns.Value),
			Message: fmt.Sprintf("invalid database min connections, must be between %d and %d", ValidDatabaseMinMinConns, ValidDatabaseMaxMinConns),
		}
	}

	if c.MaxPingTimeout.Value < ValidDatabaseMinPingTimeout || c.MaxPingTimeout.Value > ValidDatabaseMaxPingTimeout {
		return &InvalidConfigurationError{
			Field:   "database.max_ping_timeout",
			Value:   fmt.Sprintf("%d", c.MaxPingTimeout.Value),
			Message: fmt.Sprintf("invalid database max ping timeout, must be between %d and %d", ValidDatabaseMinPingTimeout, ValidDatabaseMaxPingTimeout),
		}
	}

	if c.MaxQueryTimeout.Value < ValidDatabaseMinQueryTimeout || c.MaxQueryTimeout.Value > ValidDatabaseMaxQueryTimeout {
		return &InvalidConfigurationError{
			Field:   "database.max_query_timeout",
			Value:   fmt.Sprintf("%d", c.MaxQueryTimeout.Value),
			Message: fmt.Sprintf("invalid database max query timeout, must be between %d and %d", ValidDatabaseMinQueryTimeout, ValidDatabaseMaxQueryTimeout),
		}
	}

	if c.ConnMaxIdleTime.Value < ValidDatabaseConnMinIdleTime || c.ConnMaxIdleTime.Value > ValidDatabaseConnMaxIdleTime {
		return &InvalidConfigurationError{
			Field:   "database.conn_max_idle_time",
			Value:   fmt.Sprintf("%d", c.ConnMaxIdleTime.Value),
			Message: fmt.Sprintf("invalid database max idle time, must be between %d and %d", ValidDatabaseConnMinIdleTime, ValidDatabaseConnMaxIdleTime),
		}
	}

	if c.ConnMaxLifetime.Value < ValidDatabaseConnMinLifetime || c.ConnMaxLifetime.Value > ValidDatabaseConnMaxLifetime {
		return &InvalidConfigurationError{
			Field:   "database.conn_max_lifetime",
			Value:   fmt.Sprintf("%d", c.ConnMaxLifetime.Value),
			Message: fmt.Sprintf("invalid database max connection lifetime, must be between %d and %d", ValidDatabaseConnMinLifetime, ValidDatabaseConnMaxLifetime),
		}
	}

	return nil
}
