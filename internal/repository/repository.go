package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/models"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

func HandleSQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrRecordNotFound
	}

	return err
}

type UserRepository interface {
	Save(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}
