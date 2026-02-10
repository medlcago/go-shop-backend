package repository

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByEmailUnscoped(ctx context.Context, email string) (*models.User, error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, preload bool) (*models.Product, error)
	ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error)
	CreateProduct(ctx context.Context, product *models.Product) error
	UpdateProduct(ctx context.Context, product *models.Product) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	Search(ctx context.Context, req dto.SearchProductRequest) ([]*models.Product, int64, error)
}

type CategoryRepository interface {
	ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*models.Category, int64, error)
}

type UploadRepository interface {
	Save(ctx context.Context, req *models.Upload) error
	Exists(ctx context.Context, objectKey string) (bool, error)
}

type CartRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.Cart, error)
	Create(ctx context.Context, cart *models.Cart) error
	Save(ctx context.Context, cart *models.Cart) error
	DeleteItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID) error
}
