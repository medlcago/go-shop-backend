package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

type Manager interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Wrap(ctx context.Context, f func(context.Context) error) error
}

type sqlxManager struct {
	db *sqlx.DB
}

func NewManager(db *sqlx.DB) Manager {
	return &sqlxManager{db: db}
}

func (m *sqlxManager) Begin(ctx context.Context) (context.Context, error) {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return context.WithValue(ctx, txKey{}, tx), nil
}

func (m *sqlxManager) Commit(ctx context.Context) error {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx.Commit()
	}
	return errors.New("no transaction in context")
}

func (m *sqlxManager) Rollback(ctx context.Context) error {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx.Rollback()
	}
	return errors.New("no transaction in context")
}

func (m *sqlxManager) Wrap(ctx context.Context, f func(context.Context) error) error {
	txCtx, err := m.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = m.Rollback(txCtx)
	}()

	defer func() {
		if p := recover(); p != nil {
			defer func() {
				_ = m.Rollback(txCtx)
			}()
			panic(p) // re-panic after rollback
		}
	}()

	if err := f(txCtx); err != nil {
		return fmt.Errorf("failed to wrap transaction: %w", err)
	}
	return nil
}

func GetQueryer(ctx context.Context, queryer Queryer) Queryer {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return queryer
}
