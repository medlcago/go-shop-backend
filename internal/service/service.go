package service

import (
	"context"
	"go-shop-backend/internal/dto"

	"github.com/google/uuid"
)

type AuthService interface {
	Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserTokenResponse, error)
	Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error)
}

type UserService interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}

type ProductService interface {
	GetProductByID(ctx context.Context, productID uuid.UUID) (*dto.ProductResponse, error)
	ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*dto.ProductResponse, int64, error)
	CreateProduct(ctx context.Context, req dto.ProductCreateRequest) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, productID uuid.UUID, req dto.ProductUpdateRequest) (*dto.ProductResponse, error)
	Search(ctx context.Context, req dto.SearchProductRequest) ([]*dto.ProductResponse, int64, error)
}

type CategoryService interface {
	ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*dto.ProductCategoryResponse, int64, error)
}

type UploadService interface {
	SignURL(ctx context.Context, req dto.SignURLRequest) (*dto.SignURLResponse, error)
	Save(ctx context.Context, req dto.UploadRequest) (*dto.UploadResponse, error)
	PublicURL(ctx context.Context, objectKey string) string
}

type PublicURLBuilder interface {
	PublicURL(ctx context.Context, objectKey string) string
}

type EntityService interface {
	Exists(ctx context.Context, entityType dto.EntityType, id uuid.UUID) error
}

type OrderService interface {
	CreateOrder(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID) (*dto.OrderResponse, error)
	GetOrder(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error)
	GetOrders(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, req dto.ListOrderRequest) ([]*dto.OrderResponse, int64, error)
	AddItem(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID, req dto.AddOrderItemRequest) (*dto.OrderResponse, error)
	DeleteItem(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID, productID uuid.UUID) (*dto.OrderResponse, error)
	Clear(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error)
	Checkout(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID) (*dto.OrderCheckoutResponse, error)
	HandlePaymentWebhook(ctx context.Context, body []byte) error
}
