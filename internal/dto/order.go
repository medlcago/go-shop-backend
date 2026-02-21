package dto

import (
	"time"

	"github.com/google/uuid"
)

type AddOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type OrderItemResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   int64     `json:"unit_price"`
}

type OrderResponse struct {
	ID          uuid.UUID           `json:"id"`
	Status      string              `json:"status"`
	Items       []OrderItemResponse `json:"items"`
	TotalAmount int64               `json:"total_amount"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	CompletedAt *time.Time          `json:"completed_at"`
}

type ListOrderRequest struct {
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	Status string `query:"status" validate:"omitempty,oneof=draft pending paid canceled completed"`
}
