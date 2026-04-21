package dto

import (
	"go-shop-backend/pkg/apperror"
	"time"

	"github.com/google/uuid"
)

type AddOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type OrderItemResponse struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   int64     `json:"unit_price"`
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
}

type UnavailableItem struct {
	ProductID    uuid.UUID `json:"product_id"`
	RequestedQty int       `json:"requested_qty"`
	AvailableQty int       `json:"available_qty"`
	Action       string    `json:"action"`
	Reason       string    `json:"reason"`
}

type UnavailableItems []UnavailableItem

func (u *UnavailableItems) FromErr(unavailableErr *apperror.ItemsUnavailableError) UnavailableItems {
	items := make([]UnavailableItem, len(unavailableErr.Items))

	for i, item := range unavailableErr.Items {
		items[i] = UnavailableItem{
			ProductID:    item.ProductID,
			RequestedQty: item.RequestedQty,
			AvailableQty: item.AvailableQty,
			Action:       item.Action,
			Reason:       item.Reason,
		}
	}

	return items
}

type OrderCheckoutResponse struct {
	OrderID          uuid.UUID         `json:"order_id"`
	ConfirmationURL  string            `json:"confirmation_url,omitempty"`
	UnavailableItems []UnavailableItem `json:"unavailable_items,omitempty"`
}

type ListOrderRequest struct {
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
	Status string `query:"status" validate:"omitempty,oneof=draft pending paid canceled completed"`
}
