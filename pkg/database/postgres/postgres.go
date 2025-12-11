package postgres

import (
	"database/sql"
	"go-shop-backend/pkg/database"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	postgresDialect = "postgres"
)

func New(addr string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", addr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	return database.ApplyMigrations(db, postgresDialect)
}
