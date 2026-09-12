package database

import (
	"context"
	"database/sql"
	"fmt"
	"go-shop-backend/migrations"
	"strings"

	"github.com/pressly/goose/v3"
)

const (
	tableName = "goose_migrations"
)

func Migrate(
	ctx context.Context,
	db *sql.DB,
	dialect string,
) (string, error) {
	provider, err := goose.NewProvider(
		goose.Dialect(dialect),
		db,
		migrations.FS,
		goose.WithTableName(tableName),
		goose.WithLogger(goose.NopLogger()),
	)

	if err != nil {
		return "", fmt.Errorf("migrate: failed to create goose provider: %w", err)
	}

	result, err := provider.Up(ctx)
	if err != nil {
		return "", fmt.Errorf("migrate: failed to migrate: %w", err)
	}

	var builder strings.Builder
	for i, mig := range result {
		builder.WriteString(mig.String())
		if i < len(result)-1 {
			builder.WriteString(", ")
		}
	}

	return builder.String(), nil
}
