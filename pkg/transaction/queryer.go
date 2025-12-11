package transaction

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type QueryerFunc func(ctx context.Context) Queryer

type Queryer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Rebind(query string) string
}

var _ Queryer = (*sqlx.DB)(nil)
var _ Queryer = (*sqlx.Tx)(nil)
