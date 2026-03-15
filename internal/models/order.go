package models

import (
	"database/sql"
	"go-shop-backend/pkg/apperrors"
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusDraft     OrderStatus = "draft"
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCanceled  OrderStatus = "canceled"
	OrderStatusCompleted OrderStatus = "completed"
)

type Order struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	UserID    *uuid.UUID `gorm:"type:uuid;index"`
	SessionID uuid.UUID  `gorm:"type:uuid;index"`

	Status OrderStatus `gorm:"type:order_status;not null;index:idx_orders_status;index:idx_orders_status_expires_at;default:'draft'"`

	TotalAmount int64 `gorm:"not null;check:total_amount >= 0;default:0"`

	PaymentID    *string    `gorm:"type:varchar(255);index:idx_orders_provider_payment_unique"`
	ProviderName *string    `gorm:"type:varchar(255);index:idx_orders_provider_payment_unique"`
	ExpiresAt    *time.Time `gorm:"type:timestamptz;index:idx_orders_status_expires_at"`
	PaidAt       *time.Time `gorm:"type:timestamptz"`
	CanceledAt   *time.Time `gorm:"type:timestamptz"`

	CreatedAt   time.Time    `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt   time.Time    `gorm:"type:timestamptz;default:now();not null"`
	CompletedAt sql.NullTime `gorm:"type:timestamptz"`

	Items []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:RESTRICT;"`
}

func (p *Product) CanBeAdded(qty int) error {
	if !p.IsActive {
		return apperrors.ErrProductNotActive
	}

	if p.Available() < qty {
		return apperrors.ErrInsufficientStock
	}

	return nil
}

func (o *Order) CanEdit() bool {
	return o.Status == OrderStatusDraft
}

func (o *Order) HasItems() bool {
	return len(o.Items) > 0
}

func (o *Order) Checkout(userID uuid.UUID) error {
	if !o.CanEdit() {
		return apperrors.ErrInvalidOrderStatus
	}

	if !o.HasItems() {
		return apperrors.ErrEmptyOrder
	}

	// If the order does not have a user, we link the order to the user
	if o.UserID == nil {
		o.UserID = &userID
	}

	o.Status = OrderStatusPending
	o.Recalculate()
	return nil
}

func (o *Order) MarkPaid() error {
	if o.Status != OrderStatusPending {
		return apperrors.ErrInvalidOrderStatus
	}

	o.Status = OrderStatusPaid
	o.PaidAt = new(time.Now().UTC())
	return nil
}

func (o *Order) MarkCanceled() error {
	if o.Status != OrderStatusPending {
		return apperrors.ErrInvalidOrderStatus
	}

	o.Status = OrderStatusCanceled
	o.CanceledAt = new(time.Now().UTC())
	return nil
}

func (o *Order) Recalculate() {
	var total int64

	for _, item := range o.Items {
		total += int64(item.Quantity) * item.UnitPrice
	}

	o.TotalAmount = total
}

func (o *Order) SetPaymentInfo(paymentID string, providerName string, expiresAt time.Time) {
	o.PaymentID = &paymentID
	o.ProviderName = &providerName
	o.ExpiresAt = &expiresAt
}
