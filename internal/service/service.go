package service

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/pkg/notification"

	"github.com/google/uuid"
)

type UserService interface {
	Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserTokenResponse, error)
	Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error)
	Setup2FA(ctx context.Context, userID uuid.UUID) (*dto.Setup2FAResponse, error)
	Confirm2FA(ctx context.Context, userID uuid.UUID, req dto.Confirm2FARequest) error
	Disable2FA(ctx context.Context, userID uuid.UUID, req dto.Disable2FARequest) error
	GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
	SendEmailConfirmationCode(ctx context.Context, userID uuid.UUID) (*dto.SendEmailConfirmationResponse, error)
	ConfirmEmail(ctx context.Context, userID uuid.UUID, req dto.ConfirmEmailRequest) (*dto.ConfirmEmailResponse, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error
	RefreshToken(ctx context.Context, tokenString string) (*dto.UserTokenResponse, error)
}

type ProductService interface {
	GetProductByID(ctx context.Context, productID uuid.UUID) (*dto.ProductResponse, error)
	ListProducts(ctx context.Context, req dto.ListProductRequest) ([]*dto.ProductResponse, int64, error)
	CreateProduct(ctx context.Context, req dto.ProductCreateRequest) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, productID uuid.UUID, req dto.ProductUpdateRequest) (*dto.ProductResponse, error)
	Search(ctx context.Context, req dto.SearchProductRequest) ([]*dto.ProductResponse, int64, error)
	UploadImage(ctx context.Context, productID uuid.UUID, req dto.UploadProductImageRequest) (*dto.SignURLResponse, error)
	ConfirmUploadImage(ctx context.Context, productID uuid.UUID, req dto.ConfirmUploadProductImageRequest) (*dto.UploadResponse, error)
}

type CategoryService interface {
	ListCategories(ctx context.Context, req dto.ListCategoryRequest) ([]*dto.CategoryResponse, int64, error)
}

type OrderService interface {
	CreateOrder(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID) (*dto.OrderResponse, error)
	GetOrder(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error)
	GetOrders(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, req dto.ListOrderRequest) ([]*dto.OrderResponse, int64, error)
	AddItem(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID, req dto.AddOrderItemRequest) (*dto.OrderResponse, error)
	RemoveItem(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID, itemID uuid.UUID) (*dto.OrderResponse, error)
	Clear(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error)
	Checkout(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status models.OrderStatus) error
	CancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error
}

type WishlistService interface {
	CreateWishlist(ctx context.Context, userID uuid.UUID, req dto.CreateWishlistRequest) (*dto.WishlistResponse, error)
	GetWishlist(ctx context.Context, userID uuid.UUID, wishlistID uuid.UUID) (*dto.WishlistResponse, error)
	GetWishlists(ctx context.Context, userID uuid.UUID, req dto.ListWishlistRequest) ([]*dto.WishlistResponse, int64, error)
	GetSharedWishlist(ctx context.Context, shareToken string) (*dto.WishlistResponse, error)
	UpdateWishlist(ctx context.Context, userID uuid.UUID, wishlistID uuid.UUID, req dto.UpdateWishlistRequest) (*dto.WishlistResponse, error)
	DeleteWishlist(ctx context.Context, userID uuid.UUID, wishlistID uuid.UUID) error
	RegenerateShareToken(ctx context.Context, userID uuid.UUID, wishlistID uuid.UUID) (*dto.WishlistShareTokenResponse, error)
	AddItem(ctx context.Context, userID uuid.UUID, wishlistID uuid.UUID, req dto.AddWishlistItemRequest) (*dto.WishlistResponse, error)
	UpdateItem(ctx context.Context, userID uuid.UUID, wishlistID uuid.UUID, itemID uuid.UUID, req dto.UpdateWishlistItemRequest) (*dto.WishlistResponse, error)
	RemoveItem(ctx context.Context, userID, wishlistID, itemID uuid.UUID) (*dto.WishlistResponse, error)
}

type NotificationService interface {
	SendEmailConfirmationCode(ctx context.Context, to string, code string, channel notification.Channel) error
}

type InventoryService interface {
	CheckProduct(ctx context.Context, productID uuid.UUID, quantity int) (*models.Product, error)
	ReserveItems(ctx context.Context, items []dto.InventoryItem) error
	ReleaseItems(ctx context.Context, items []dto.InventoryItem) error
	DeductItems(ctx context.Context, items []dto.InventoryItem) error
}

type PaymentService interface {
	CreatePayment(ctx context.Context, userID uuid.UUID, req dto.CreatePaymentRequest) (*dto.PaymentResponse, error)
	HandleWebhook(ctx context.Context, body []byte) error
}
