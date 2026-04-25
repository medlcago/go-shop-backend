package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateWishlistRequest struct {
	Title    string `json:"title" validate:"required,min=1,max=255"`
	IsPublic bool   `json:"is_public"`
}

type UpdateWishlistRequest struct {
	Title    *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	IsPublic *bool   `json:"is_public,omitempty"`
}

type AddWishlistItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Note      *string   `json:"note" validate:"omitempty,max=128"`
	Priority  int       `json:"priority" validate:"required,min=0,max=2"`
}

type UpdateWishlistItemRequest struct {
	Note     *string `json:"note" validate:"omitempty,max=128"`
	Priority *int    `json:"priority" validate:"omitempty,min=0,max=2"`
}

type WishlistProductResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     int       `json:"price"`
	Available int       `json:"available"`
	IsActive  bool      `json:"is_active"`
}

type WishlistItemResponse struct {
	ID        uuid.UUID                `json:"id"`
	ProductID uuid.UUID                `json:"product_id"`
	Product   *WishlistProductResponse `json:"product,omitempty"`
	Note      *string                  `json:"note"`
	Priority  int                      `json:"priority"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

type WishlistResponse struct {
	ID         uuid.UUID              `json:"id"`
	Title      string                 `json:"title"`
	IsPublic   bool                   `json:"is_public"`
	ShareToken string                 `json:"share_token"`
	Items      []WishlistItemResponse `json:"items"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type WishlistShareTokenResponse struct {
	ShareToken string `json:"share_token"`
}

type ListWishlistRequest struct {
	Limit  int `query:"limit"`
	Offset int `query:"offset"`
}
