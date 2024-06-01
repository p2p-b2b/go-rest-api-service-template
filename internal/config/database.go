package config

import (
	"time"
)

const (
	DefaultDatabaseKind     = "postgres"
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
		Kind:     NewField("database.kind", "DATABASE_KIND", "Database Kind. Possible values [postgres|mysql]", DefaultDatabaseKind),
		Address:  NewField("database.address", "DATABASE_ADDRESS", "Database IP Address or Hostname", DefaultDatabaseAddress),
		Port:     NewField("database.port", "DATABASE_PORT", "Database Port", DefaultDatabasePort),
		Username: NewField("database.username", "DATABASE_USERNAME", "Database Username", DefaultDatabaseUsername),
		Password: NewField("database.password", "DATABASE_PASSWORD", "Database Password", DefaultDatabasePassword),
		Name:     NewField("database.name", "DATABASE_NAME", "Database Name", DefaultDatabaseName),
		SSLMode:  NewField("database.ssl.mode", "DATABASE_SSL_MODE", "Database SSL Mode", DefaultDatabaseSSLMode),
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

// PaseEnvVars reads the database configuration from environment variables
// and sets the values in the configuration
func (c *DatabaseConfig) PaseEnvVars() {
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
