package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/p2p-b2b/go-rest-api-service-template/database"
)

// initDatabase sets up the database connection and runs migrations if enabled
func (a *App) initDatabase(ctx context.Context) error {
	// Create DSN string
	dbDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		a.configs.Database.Address.Value,
		a.configs.Database.Port.Value,
		a.configs.Database.Username.Value,
		a.configs.Database.Password.Value,
		a.configs.Database.Name.Value,
		a.configs.Database.SSLMode.Value,
		a.configs.Database.TimeZone.Value,
	)

	// Parse config
	dbCfg, err := pgxpool.ParseConfig(dbDSN)
	if err != nil {
		return fmt.Errorf("error parsing pgx pool config: %w", err)
	}

	// Set connection pool parameters
	dbCfg.MaxConns = int32(a.configs.Database.MaxConns.Value)
	dbCfg.MinConns = int32(a.configs.Database.MinConns.Value)
	dbCfg.MaxConnLifetime = a.configs.Database.ConnMaxLifetime.Value
	dbCfg.MaxConnIdleTime = a.configs.Database.ConnMaxIdleTime.Value

	// This pool will be used first to run migrations
	// Then will be closed and created again to register pgvector types
	initializationPool, err := pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}
	defer initializationPool.Close()

	slog.Debug("database connection established",
		"dsn", dbDSN,
		"kind", a.configs.Database.Kind.Value,
		"address", a.configs.Database.Address.Value,
		"port", a.configs.Database.Port.Value,
		"username", a.configs.Database.Username.Value,
		"name", a.configs.Database.Name.Value,
		"ssl_mode", a.configs.Database.SSLMode.Value,
		"max_conns", a.configs.Database.MaxConns.Value,
		"min_conns", a.configs.Database.MinConns.Value,
		"conn_max_lifetime", a.configs.Database.ConnMaxLifetime.Value,
		"conn_max_idle_time", a.configs.Database.ConnMaxIdleTime.Value,
	)

	// Test database connection
	dbPingCtx, cancel := context.WithTimeout(ctx, a.configs.Database.MaxPingTimeout.Value)
	defer cancel()

	if err := initializationPool.Ping(dbPingCtx); err != nil {
		return fmt.Errorf("database ping error: %w", err)
	}

	// Run migrations if enabled
	if a.configs.Database.MigrationEnable.Value {
		slog.Info("running database migrations")

		db := stdlib.OpenDBFromPool(initializationPool)
		if err := database.Migrate(ctx, "pgx", db); err != nil {
			return fmt.Errorf("database migration error: %w", err)
		}
	}

	// After migrations, register pgvector types
	// First, we need to close the initialization pool
	initializationPool.Close()

	// Create the final connection pool with pgvector support
	a.dbPool, err = pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("failed to create database pool with pgvector support: %w", err)
	}

	// Verify connection with the new pool
	if err := a.dbPool.Ping(dbPingCtx); err != nil {
		return fmt.Errorf("database ping error with pgvector pool: %w", err)
	}

	return nil
}
