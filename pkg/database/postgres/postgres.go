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

func New(addr string, opts ...Option) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", addr)
	if err != nil {
		return nil, err
	}

	cfg := defaultOptions()

	for _, opt := range opts {
		opt.apply(cfg)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return db, nil
}

func Migrate(db *sql.DB) error {
	return database.ApplyMigrations(db, postgresDialect)
}
