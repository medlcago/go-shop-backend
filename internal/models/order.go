package models

import (
	"database/sql"
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

	Status OrderStatus `gorm:"type:order_status;not null;index;default:'draft'"`

	TotalAmount int64 `gorm:"not null;check:total_amount >= 0;default:0"`

	CreatedAt   time.Time    `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt   time.Time    `gorm:"type:timestamptz;default:now();not null"`
	CompletedAt sql.NullTime `gorm:"type:timestamptz"`

	Items []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:RESTRICT;"`
}

func (o Order) CanEdit() bool {
	return o.Status == OrderStatusDraft
}

type OrderItem struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	OrderID uuid.UUID `gorm:"type:uuid;not null;index"`
	Order   Order     `gorm:"foreignKey:OrderID;constraint:OnDelete:RESTRICT;"`

	ProductID   uuid.UUID `gorm:"type:uuid;not null;index"`
	ProductName string    `gorm:"type:varchar(255);not null"`
	Product     *Product  `gorm:"constraint:OnUpdate:RESTRICT;"`

	Quantity  int   `gorm:"not null;check:quantity > 0"`
	UnitPrice int64 `gorm:"not null;check:unit_price >= 0"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`
}
