package database

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

type TxManager interface {
	Wrap(ctx context.Context, fn func(context.Context) error) error
}

type gormTxManager struct {
	db *gorm.DB
}

func NewManager(db *gorm.DB) TxManager {
	return &gormTxManager{db: db}
}

func (m *gormTxManager) Wrap(ctx context.Context, fn func(context.Context) error) (err error) {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.Transaction(func(nestedTx *gorm.DB) error {
			txCtx := context.WithValue(ctx, txKey{}, nestedTx)
			return fn(txCtx)
		})
	}

	return m.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}
