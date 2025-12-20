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
}

type CategoryService interface {
	ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*dto.ProductCategoryResponse, int64, error)
}
