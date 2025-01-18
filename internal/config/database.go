package config

import (
	"errors"
	"slices"
	"strings"
	"time"
)

var (
	// ErrInvalidDatabaseKind is returned when an invalid database kind is provided
	ErrInvalidDatabaseKind = errors.New("invalid database kind, must be one of [" + ValidDatabaseKind + "]")

	// ErrInvalidDatabasePort is returned when an invalid database port is provided
	ErrInvalidDatabasePort = errors.New("invalid database port, must be between 0 and 65535")

	// ErrUserInvalidUsername is returned when an invalid username is provided
	ErrUserInvalidUsername = errors.New("invalid username, must be between 2 and 32 characters")

	// ErrInvalidDatabaseName is returned when an invalid database name is provided
	ErrInvalidDatabaseName = errors.New("invalid database name, must be between 2 and 32 characters")

	// ErrInvalidSSLMode is returned when an invalid SSL mode is provided
	ErrInvalidSSLMode = errors.New("invalid SSL mode, must be one of [" + ValidSSLModes + "]")

	// ErrInvalidTimeZone is returned when an invalid timezone is provided
	ErrInvalidTimeZone = errors.New("invalid timezone, must be between 2 and 32 characters")

	// ErrInvalidPassword is returned when an invalid password is provided
	ErrInvalidPassword = errors.New("invalid password, must be between 2 and 128 characters")

	// ErrInvalidMaxIdleConns is returned when an invalid max idle connections is provided
	ErrInvalidMaxIdleConns = errors.New("invalid max idle connections, must be between 0 and 100")

	// ErrInvalidMaxOpenConns is returned when an invalid max open connections is provided
	ErrInvalidMaxOpenConns = errors.New("invalid max open connections, must be between 0 and 100")

	// ErrDBInvalidMaxPingTimeout is returned when an invalid max ping timeout is provided
	ErrDBInvalidMaxPingTimeout = errors.New("invalid max ping timeout, must be between 1s and 30s")

	// ErrDBInvalidMaxQueryTimeout is returned when an invalid max query timeout is provided
	ErrDBInvalidMaxQueryTimeout = errors.New("invalid max query timeout, must be between 1s and 30s")

	// ErrInvalidConnMaxIdleTime is returned when an invalid connection max idle time is provided
	ErrInvalidConnMaxIdleTime = errors.New("invalid connection max idle time, must be between 1s and 60m")

	// ErrInvalidConnMaxLifetime is returned when an invalid connection max lifetime is provided
	ErrInvalidConnMaxLifetime = errors.New("invalid connection max lifetime, must be between 1s and 600s")
)

const (
	ValidDatabaseKind = "pgx|postgres"
	ValidSSLModes     = "disable|allow|prefer|require|verify-ca|verify-full"

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
		SSLMode:  NewField("database.ssl.mode", "DATABASE_SSL_MODE", "Database SSL Mode. Possible values ["+ValidSSLModes+"]", DefaultDatabaseSSLMode),
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
		return ErrInvalidDatabaseKind
	}

	if c.Port.Value <= 0 || c.Port.Value >= 65535 {
		return ErrInvalidDatabasePort
	}

	if c.Username.Value == "" || len(c.Username.Value) < 2 || len(c.Username.Value) > 32 {
		return ErrUserInvalidUsername
	}

	if c.Password.Value == "" || len(c.Password.Value) < 2 || len(c.Password.Value) > 128 {
		return ErrInvalidPassword
	}

	if c.Name.Value == "" || len(c.Name.Value) < 2 || len(c.Name.Value) > 32 {
		return ErrInvalidDatabaseName
	}

	if !slices.Contains(strings.Split(ValidSSLModes, "|"), c.SSLMode.Value) {
		return ErrInvalidSSLMode
	}

	if c.TimeZone.Value == "" || len(c.TimeZone.Value) < 2 || len(c.TimeZone.Value) > 32 {
		return ErrInvalidTimeZone
	}

	if c.MaxIdleConns.Value < 0 || c.MaxIdleConns.Value > 100 {
		return ErrInvalidMaxIdleConns
	}

	if c.MaxOpenConns.Value < 0 || c.MaxOpenConns.Value > 100 {
		return ErrInvalidMaxOpenConns
	}

	if c.MaxPingTimeout.Value < 1*time.Second || c.MaxPingTimeout.Value > 30*time.Second {
		return ErrDBInvalidMaxPingTimeout
	}

	if c.MaxQueryTimeout.Value < 1*time.Second || c.MaxQueryTimeout.Value > 30*time.Second {
		return ErrDBInvalidMaxQueryTimeout
	}

	if c.ConnMaxIdleTime.Value < 1*time.Second || c.ConnMaxIdleTime.Value > 600*time.Minute {
		return ErrInvalidConnMaxIdleTime
	}

	if c.ConnMaxLifetime.Value < 1*time.Second || c.ConnMaxLifetime.Value > 600*time.Second {
		return ErrInvalidConnMaxLifetime
	}

	return nil
}
