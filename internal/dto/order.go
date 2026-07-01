package dto

import (
	"time"

	"github.com/google/uuid"
)

type AddOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type OrderCheckoutRequest struct {
	AddressID uuid.UUID `json:"address_id" validate:"required"`
}

type OrderItemProductResponse struct {
	ID     uuid.UUID `json:"id"`
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
}

type OrderItemResponse struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   int64     `json:"unit_price"`

	Product OrderItemProductResponse `json:"product"`
}

type OrderResponse struct {
	ID           uuid.UUID           `json:"id"`
	Status       string              `json:"status"`
	Items        []OrderItemResponse `json:"items"`
	TotalAmount  int64               `json:"total_amount"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	PaymentID    *string             `json:"payment_id"`
	ProviderName *string             `json:"provider_name"`
	PaidAt       *time.Time          `json:"paid_at"`
	ExpiresAt    *time.Time          `json:"expires_at"`
	CanceledAt   *time.Time          `json:"canceled_at"`
	CompletedAt  *time.Time          `json:"completed_at"`
	IsGuestOrder bool                `json:"is_guest_order"`
	Address      *AddressResponse    `json:"address"`
}

type ListOrderRequest struct {
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	Status string `query:"status" validate:"omitempty,oneof=draft pending paid canceled completed"`
}
