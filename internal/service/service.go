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
}

type CategoryService interface {
	ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*dto.ProductCategoryResponse, int64, error)
}

type UploadService interface {
	SignURL(ctx context.Context, req dto.SignURLRequest) (*dto.SignURLResponse, error)
}

type EntityService interface {
	Exists(ctx context.Context, entityType dto.EntityType, id uuid.UUID) (bool, error)
}
