package config

import (
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	ErrDatabaseInvalidKind            = errors.New("invalid database kind, must be one of [" + ValidDatabaseKind + "]")
	ErrDatabaseInvalidPort            = errors.New("invalid database port, must be between [" + strconv.Itoa(ValidDatabaseMinPort) + "] and [" + strconv.Itoa(ValidDatabaseMaxPort) + "]")
	ErrDatabaseInvalidUsername        = errors.New("invalid username, must be between [" + strconv.Itoa(ValidDatabaseUsernameMinLen) + "] and [" + strconv.Itoa(ValidDatabaseUsernameMaxLen) + "] characters")
	ErrDatabaseInvalidDatabaseName    = errors.New("invalid database name, must be between [" + strconv.Itoa(ValidDatabaseNameMinLen) + "] and [" + strconv.Itoa(ValidDatabaseNameMaxLen) + "] characters")
	ErrDatabaseInvalidSSLMode         = errors.New("invalid SSL mode, must be one of [" + ValidDatabaseSSLModes + "]")
	ErrDatabaseInvalidTimeZone        = errors.New("invalid timezone, must be between [" + strconv.Itoa(ValidDatabaseTimeZoneMinLen) + "] and [" + strconv.Itoa(ValidDatabaseTimeZoneMaxLen) + "] characters")
	ErrDatabaseInvalidPassword        = errors.New("invalid password, must be between [" + strconv.Itoa(ValidDatabasePasswordMinLen) + "] and [" + strconv.Itoa(ValidDatabasePasswordMaxLen) + "] characters")
	ErrDatabaseInvalidMaxIdleConns    = errors.New("invalid max idle connections, must be between [" + strconv.Itoa(ValidDatabaseMinIdleConns) + "] and [" + strconv.Itoa(ValidDatabaseMaxIdleConns) + "]")
	ErrDatabaseInvalidMaxOpenConns    = errors.New("invalid max open connections, must be between [" + strconv.Itoa(ValidDatabaseMinOpenConns) + "] and [" + strconv.Itoa(ValidDatabaseMaxOpenConns) + "]")
	ErrDatabaseInvalidMaxPingTimeout  = errors.New("invalid max ping timeout, must be between [" + ValidDatabaseMinPingTimeout.String() + "] and [" + ValidDatabaseMaxPingTimeout.String() + "]")
	ErrDatabaseInvalidMaxQueryTimeout = errors.New("invalid max query timeout, must be between [" + ValidDatabaseMinQueryTimeout.String() + "] and [" + ValidDatabaseMaxQueryTimeout.String() + "]")
	ErrDatabaseInvalidConnMaxIdleTime = errors.New("invalid connection max idle time, must be between [" + ValidDatabaseConnMinIdleTime.String() + "] and [" + ValidDatabaseConnMaxIdleTime.String() + "]")
	ErrDatabaseInvalidConnMaxLifetime = errors.New("invalid connection max lifetime, must be between [" + ValidDatabaseConnMinLifetime.String() + "] and [" + ValidDatabaseConnMaxLifetime.String() + "]")
)

const (
	ValidDatabaseKind            = "pgx|postgres"
	ValidDatabaseSSLModes        = "disable|allow|prefer|require|verify-ca|verify-full"
	ValidDatabaseMaxPort         = 65535
	ValidDatabaseMinPort         = 0
	ValidDatabaseUsernameMaxLen  = 32
	ValidDatabaseUsernameMinLen  = 2
	ValidDatabasePasswordMaxLen  = 128
	ValidDatabasePasswordMinLen  = 2
	ValidDatabaseNameMaxLen      = 32
	ValidDatabaseNameMinLen      = 2
	ValidDatabaseTimeZoneMaxLen  = 32
	ValidDatabaseTimeZoneMinLen  = 2
	ValidDatabaseMaxIdleConns    = 100
	ValidDatabaseMinIdleConns    = 0
	ValidDatabaseMaxOpenConns    = 100
	ValidDatabaseMinOpenConns    = 0
	ValidDatabaseMaxPingTimeout  = 30 * time.Second
	ValidDatabaseMinPingTimeout  = 1 * time.Second
	ValidDatabaseMaxQueryTimeout = 30 * time.Second
	ValidDatabaseMinQueryTimeout = 1 * time.Second
	ValidDatabaseConnMaxIdleTime = 600 * time.Minute
	ValidDatabaseConnMinIdleTime = 1 * time.Second
	ValidDatabaseConnMaxLifetime = 600 * time.Second
	ValidDatabaseConnMinLifetime = 1 * time.Second

	DefaultDatabaseKind     = "pgx"
	DefaultDatabaseAddress  = "localhost"
	DefaultDatabasePort     = 5432
	DefaultDatabaseUsername = "username"
	DefaultDatabasePassword = "password"
	DefaultDatabaseName     = "go-rest-api-service-template"
	DefaultDatabaseSSLMode  = "disable"
	DefaultDatabaseTimeZone = "UTC"

	DefaultDatabaseMaxPingTimeout  = 5 * time.Second
	DefaultDatabaseMaxQueryTimeout = 5 * time.Second

	DefaultDatabaseMaxIdleConns = 10
	DefaultDatabaseMaxOpenConns = 100

	DefaultDatabaseConnMaxIdleTime = 30 * time.Minute
	DefaultDatabaseConnMaxLifetime = 15 * time.Second

	DefaultDatabaseMigrationEnable = false
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

	MaxIdleConns Field[int]
	MaxOpenConns Field[int]

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

		MaxIdleConns: NewField("database.max.idle.conns", "DATABASE_MAX_IDLE_CONNS", "Database Max Idle Connections", DefaultDatabaseMaxIdleConns),
		MaxOpenConns: NewField("database.max.open.conns", "DATABASE_MAX_OPEN_CONNS", "Database Max Open Connections", DefaultDatabaseMaxOpenConns),

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

	c.MaxIdleConns.Value = GetEnv(c.MaxIdleConns.EnVarName, c.MaxIdleConns.Value)
	c.MaxOpenConns.Value = GetEnv(c.MaxOpenConns.EnVarName, c.MaxOpenConns.Value)

	c.ConnMaxIdleTime.Value = GetEnv(c.ConnMaxIdleTime.EnVarName, c.ConnMaxIdleTime.Value)
	c.ConnMaxLifetime.Value = GetEnv(c.ConnMaxLifetime.EnVarName, c.ConnMaxLifetime.Value)

	c.MigrationEnable.Value = GetEnv(c.MigrationEnable.EnVarName, c.MigrationEnable.Value)
}

// Validate validates the database configuration values
func (c *DatabaseConfig) Validate() error {
	if !slices.Contains(strings.Split(ValidDatabaseKind, "|"), c.Kind.Value) {
		return ErrDatabaseInvalidKind
	}

	if c.Port.Value <= ValidDatabaseMinPort || c.Port.Value >= ValidDatabaseMaxPort {
		return ErrDatabaseInvalidPort
	}

	if c.Username.Value == "" || len(c.Username.Value) < ValidDatabaseUsernameMinLen || len(c.Username.Value) > ValidDatabaseUsernameMaxLen {
		return ErrDatabaseInvalidUsername
	}

	if c.Password.Value == "" || len(c.Password.Value) < ValidDatabasePasswordMinLen || len(c.Password.Value) > ValidDatabasePasswordMaxLen {
		return ErrDatabaseInvalidPassword
	}

	if c.Name.Value == "" || len(c.Name.Value) < ValidDatabaseNameMinLen || len(c.Name.Value) > ValidDatabaseNameMaxLen {
		return ErrDatabaseInvalidDatabaseName
	}

	if !slices.Contains(strings.Split(ValidDatabaseSSLModes, "|"), c.SSLMode.Value) {
		return ErrDatabaseInvalidSSLMode
	}

	if c.TimeZone.Value == "" || len(c.TimeZone.Value) < ValidDatabaseTimeZoneMinLen || len(c.TimeZone.Value) > ValidDatabaseTimeZoneMaxLen {
		return ErrDatabaseInvalidTimeZone
	}

	if c.MaxIdleConns.Value < ValidDatabaseMinIdleConns || c.MaxIdleConns.Value > ValidDatabaseMaxIdleConns {
		return ErrDatabaseInvalidMaxIdleConns
	}

	if c.MaxOpenConns.Value < ValidDatabaseMinOpenConns || c.MaxOpenConns.Value > ValidDatabaseMaxOpenConns {
		return ErrDatabaseInvalidMaxOpenConns
	}

	if c.MaxPingTimeout.Value < ValidDatabaseMinPingTimeout || c.MaxPingTimeout.Value > ValidDatabaseMaxPingTimeout {
		return ErrDatabaseInvalidMaxPingTimeout
	}

	if c.MaxQueryTimeout.Value < ValidDatabaseMinQueryTimeout || c.MaxQueryTimeout.Value > ValidDatabaseMaxQueryTimeout {
		return ErrDatabaseInvalidMaxQueryTimeout
	}

	if c.ConnMaxIdleTime.Value < ValidDatabaseConnMinIdleTime || c.ConnMaxIdleTime.Value > ValidDatabaseConnMaxIdleTime {
		return ErrDatabaseInvalidConnMaxIdleTime
	}

	if c.ConnMaxLifetime.Value < ValidDatabaseConnMinLifetime || c.ConnMaxLifetime.Value > ValidDatabaseConnMaxLifetime {
		return ErrDatabaseInvalidConnMaxLifetime
	}

	return nil
}
