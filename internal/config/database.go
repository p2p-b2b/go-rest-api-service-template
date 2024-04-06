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
	Address         Item[string]
	Port            Item[int]
	Username        Item[string]
	Password        Item[string]
	Name            Item[string]
	SSLMode         Item[string]
	MaxPingTimeout  Item[time.Duration]
	MaxQueryTimeout Item[time.Duration]
	ConnMaxLifetime Item[time.Duration]
	MaxIdleConns    Item[int]
	MaxOpenConns    Item[int]
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Address:         NewItem("database.address", "DATABASE_ADDRESS", DefaultDatabaseAddress),
		Port:            NewItem("database.port", "DATABASE_PORT", DefaultDatabasePort),
		Username:        NewItem("database.username", "DATABASE_USERNAME", DefaultDatabaseUsername),
		Password:        NewItem("database.password", "DATABASE_PASSWORD", DefaultDatabasePassword),
		Name:            NewItem("database.name", "DATABASE_NAME", DefaultDatabaseName),
		SSLMode:         NewItem("database.sslmode", "DATABASE_SSL_MODE", DefaultDatabaseSSLMode),
		MaxPingTimeout:  NewItem("database.maxpingtimeout", "DATABASE_MAX_PING_TIMEOUT", DefaultDatabaseMaxPingTimeout),
		MaxQueryTimeout: NewItem("database.maxquerytimeout", "DATABASE_MAX_QUERY_TIMEOUT", DefaultDatabaseMaxQueryTimeout),
		ConnMaxLifetime: NewItem("database.connmaxlifetime", "DATABASE_CONN_MAX_LIFETIME", DefaultDatabaseConnMaxLifetime),
		MaxIdleConns:    NewItem("database.maxidleconns", "DATABASE_MAX_IDLE_CONNS", DefaultDatabaseMaxIdleConns),
		MaxOpenConns:    NewItem("database.maxopenconns", "DATABASE_MAX_OPEN_CONNS", DefaultDatabaseMaxOpenConns),
	}
}

// PaseEnvVars reads the database configuration from environment variables
// and sets the values in the configuration
func (c *DatabaseConfig) PaseEnvVars() {
	c.Address.Value = getEnv(c.Address.EnVarName, c.Address.Value)
	c.Port.Value = getEnv(c.Port.EnVarName, c.Port.Value)
	c.Username.Value = getEnv(c.Username.EnVarName, c.Username.Value)
	c.Password.Value = getEnv(c.Password.EnVarName, c.Password.Value)
	c.Name.Value = getEnv(c.Name.EnVarName, c.Name.Value)
	c.SSLMode.Value = getEnv(c.SSLMode.EnVarName, c.SSLMode.Value)
	c.MaxPingTimeout.Value = getEnv(c.MaxPingTimeout.EnVarName, c.MaxPingTimeout.Value)
	c.MaxQueryTimeout.Value = getEnv(c.MaxQueryTimeout.EnVarName, c.MaxQueryTimeout.Value)
	c.ConnMaxLifetime.Value = getEnv(c.ConnMaxLifetime.EnVarName, c.ConnMaxLifetime.Value)
	c.MaxIdleConns.Value = getEnv(c.MaxIdleConns.EnVarName, c.MaxIdleConns.Value)
	c.MaxOpenConns.Value = getEnv(c.MaxOpenConns.EnVarName, c.MaxOpenConns.Value)
}
