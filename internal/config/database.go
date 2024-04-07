package config

import (
	"time"
)

const (
	// Default Database Configuration
	DefaultDatabaseAddress         = "localhost"
	DefaultDatabasePort            = 5432
	DefaultDatabaseUsername        = "dbadmin"
	DefaultDatabasePassword        = "dbadmin"
	DefaultDatabaseName            = "go-template-service"
	DefaultDatabaseSSLMode         = "disable"
	DefaultDatabaseMaxPingTimeout  = 5 * time.Second
	DefaultDatabaseMaxQueryTimeout = 5 * time.Second
	DefaultDatabaseConnMaxLifetime = 30 * time.Minute
	DefaultDatabaseMaxIdleConns    = 10
	DefaultDatabaseMaxOpenConns    = 100
)

type DatabaseConfig struct {
	Address         Field[string]
	Username        Field[string]
	Password        Field[string]
	Name            Field[string]
	SSLMode         Field[string]
	Port            Field[int]
	MaxIdleConns    Field[int]
	MaxOpenConns    Field[int]
	ConnMaxLifetime Field[time.Duration]
	MaxQueryTimeout Field[time.Duration]
	MaxPingTimeout  Field[time.Duration]
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Address:         NewField("database.address", "DATABASE_ADDRESS", "Database IP Address or Hostname", DefaultDatabaseAddress),
		Port:            NewField("database.port", "DATABASE_PORT", "Database Port", DefaultDatabasePort),
		Username:        NewField("database.username", "DATABASE_USERNAME", "Database Username", DefaultDatabaseUsername),
		Password:        NewField("database.password", "DATABASE_PASSWORD", "Database Password", DefaultDatabasePassword),
		Name:            NewField("database.name", "DATABASE_NAME", "Database Name", DefaultDatabaseName),
		SSLMode:         NewField("database.sslmode", "DATABASE_SSL_MODE", "Database SSL Mode", DefaultDatabaseSSLMode),
		MaxPingTimeout:  NewField("database.maxpingtimeout", "DATABASE_MAX_PING_TIMEOUT", "Database Max Ping Timeout", DefaultDatabaseMaxPingTimeout),
		MaxQueryTimeout: NewField("database.maxquerytimeout", "DATABASE_MAX_QUERY_TIMEOUT", "Database Max Query Timeout", DefaultDatabaseMaxQueryTimeout),
		ConnMaxLifetime: NewField("database.connmaxlifetime", "DATABASE_CONN_MAX_LIFETIME", "Database Connection Max Lifetime", DefaultDatabaseConnMaxLifetime),
		MaxIdleConns:    NewField("database.maxidleconns", "DATABASE_MAX_IDLE_CONNS", "Database Max Idle Connections", DefaultDatabaseMaxIdleConns),
		MaxOpenConns:    NewField("database.maxopenconns", "DATABASE_MAX_OPEN_CONNS", "Database Max Open Connections", DefaultDatabaseMaxOpenConns),
	}
}

// PaseEnvVars reads the database configuration from environment variables
// and sets the values in the configuration
func (c *DatabaseConfig) PaseEnvVars() {
	c.Address.Value = GetEnv(c.Address.EnVarName, c.Address.Value)
	c.Port.Value = GetEnv(c.Port.EnVarName, c.Port.Value)
	c.Username.Value = GetEnv(c.Username.EnVarName, c.Username.Value)
	c.Password.Value = GetEnv(c.Password.EnVarName, c.Password.Value)
	c.Name.Value = GetEnv(c.Name.EnVarName, c.Name.Value)
	c.SSLMode.Value = GetEnv(c.SSLMode.EnVarName, c.SSLMode.Value)
	c.MaxPingTimeout.Value = GetEnv(c.MaxPingTimeout.EnVarName, c.MaxPingTimeout.Value)
	c.MaxQueryTimeout.Value = GetEnv(c.MaxQueryTimeout.EnVarName, c.MaxQueryTimeout.Value)
	c.ConnMaxLifetime.Value = GetEnv(c.ConnMaxLifetime.EnVarName, c.ConnMaxLifetime.Value)
	c.MaxIdleConns.Value = GetEnv(c.MaxIdleConns.EnVarName, c.MaxIdleConns.Value)
	c.MaxOpenConns.Value = GetEnv(c.MaxOpenConns.EnVarName, c.MaxOpenConns.Value)
}
