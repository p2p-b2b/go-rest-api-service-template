package database

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Migrate runs the database migrations
func Migrate(ctx context.Context, dialect string, db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(dialect); err != nil {
		slog.Error("database dialect not supported", "error", err)
		return err
	}

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		slog.Error("failed to migrate database", "error", err)
		return err
	}

	if err := goose.VersionContext(ctx, db, "migrations"); err != nil {
		slog.Error("failed to get database version", "error", err)
		return err
	}

	slog.Info("database migrated")
	return nil
}
