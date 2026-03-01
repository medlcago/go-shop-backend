package repository

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByEmailUnscoped(ctx context.Context, email string) (*models.User, error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, id uuid.UUID, preload bool) (*models.Product, error)
	ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*models.Product, int64, error)
	Create(ctx context.Context, product *models.Product) error
	Update(ctx context.Context, product *models.Product) error
	BulkUpsert(ctx context.Context, products []*models.Product) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetByIDsForUpdate(ctx context.Context, ids []uuid.UUID) ([]*models.Product, error)
	Search(ctx context.Context, req dto.SearchProductRequest) ([]*models.Product, int64, error)
}

type CategoryRepository interface {
	ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*models.Category, int64, error)
}

type UploadRepository interface {
	Create(ctx context.Context, req *models.Upload) error
	Exists(ctx context.Context, objectKey string) (bool, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	Update(ctx context.Context, order *models.Order) error
	GetByOwner(ctx context.Context, orderID uuid.UUID, userID *uuid.UUID, sessionID uuid.UUID, preload bool) (*models.Order, error)
	GetListByOwner(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, req dto.ListOrderRequest) ([]*models.Order, int64, error)
	GetByPayment(ctx context.Context, providerName string, paymentID string, preload bool) (*models.Order, error)
}

type OrderItemRepository interface {
	GetItem(ctx context.Context, productID uuid.UUID, orderID uuid.UUID) (*models.OrderItem, error)
	AddItem(ctx context.Context, orderItem *models.OrderItem) error
	Upsert(ctx context.Context, orderItem *models.OrderItem) error
	DeleteItem(ctx context.Context, orderID uuid.UUID, productID uuid.UUID) error
	Clear(ctx context.Context, orderID uuid.UUID) error
}
