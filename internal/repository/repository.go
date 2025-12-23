package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"

	"github.com/google/uuid"
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
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
}

type CategoryRepository interface {
	ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*models.ProductCategory, int64, error)
}
