package database

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

const (
	tableName     = "goose_migrations"
	migrationsDir = "migrations"
)

func Migrate(db *sql.DB, dialect string) error {
	goose.SetLogger(goose.NopLogger())

	goose.SetTableName(tableName)

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("migrate: failed to set dialect: %w", err)
	}

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("migrate: failed to migrate: %w", err)
	}

	return nil
}
