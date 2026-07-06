package models

import (
	"go-shop-backend/pkg/apperror"
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

var (
	orderStatusTransitions = map[OrderStatus][]OrderStatus{
		OrderStatusDraft:     {OrderStatusPending},
		OrderStatusPending:   {OrderStatusPaid, OrderStatusCanceled},
		OrderStatusPaid:      {OrderStatusCompleted},
		OrderStatusCanceled:  {},
		OrderStatusCompleted: {},
	}
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusDraft, OrderStatusPending, OrderStatusPaid, OrderStatusCanceled, OrderStatusCompleted:
		return true
	}
	return false
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	if !s.IsValid() || !next.IsValid() {
		return false
	}

	if s == next {
		return false
	}

	allowed, ok := orderStatusTransitions[s]
	if !ok {
		return false
	}

	for _, allowedNext := range allowed {
		if next == allowedNext {
			return true
		}
	}
	return false
}

type Order struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	UserID    *uuid.UUID `gorm:"type:uuid;index:idx_orders_user_id"`
	SessionID uuid.UUID  `gorm:"type:uuid;index:idx_orders_session_id"`

	Status OrderStatus `gorm:"type:order_status;not null;index:idx_orders_status;index:idx_orders_status_expires_at;default:'draft'"`

	TotalAmount int64 `gorm:"not null;check:total_amount >= 0;default:0"`

	PaymentID    *string    `gorm:"type:varchar(255);index:idx_orders_provider_payment_unique"`
	ProviderName *string    `gorm:"type:varchar(255);index:idx_orders_provider_payment_unique"`
	ExpiresAt    *time.Time `gorm:"type:timestamptz;index:idx_orders_status_expires_at"`
	PaidAt       *time.Time `gorm:"type:timestamptz"`
	CanceledAt   *time.Time `gorm:"type:timestamptz"`
	Address      *Address   `gorm:"serializer:json"`

	CreatedAt   time.Time  `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt   time.Time  `gorm:"type:timestamptz;default:now();not null"`
	CompletedAt *time.Time `gorm:"type:timestamptz"`

	Items []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:RESTRICT;"`
}

func (o *Order) IsOwnedBy(userID uuid.UUID) bool {
	return o.UserID != nil && *o.UserID == userID
}

func (o *Order) IsGuestOrder() bool {
	return o.UserID == nil
}

func (o *Order) HasItems() bool {
	return len(o.Items) > 0
}

func (o *Order) Checkout() error {
	if !o.Status.CanTransitionTo(OrderStatusPending) {
		return apperror.ErrInvalidOrderStatus
	}

	if !o.HasItems() {
		return apperror.ErrEmptyOrder
	}

	o.Status = OrderStatusPending
	o.Recalculate()
	return nil
}

func (o *Order) Pay() error {
	if !o.Status.CanTransitionTo(OrderStatusPaid) {
		return apperror.ErrInvalidOrderStatus
	}

	o.Status = OrderStatusPaid
	o.PaidAt = new(time.Now().UTC())
	return nil
}

func (o *Order) Cancel() error {
	if !o.Status.CanTransitionTo(OrderStatusCanceled) {
		return apperror.ErrInvalidOrderStatus
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

func (o *Order) SetPaymentInfo(paymentID string, providerName string) {
	o.PaymentID = &paymentID
	o.ProviderName = &providerName
}
