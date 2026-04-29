package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	OrderID uuid.UUID `gorm:"type:uuid;not null;index:idx_order_items_order_id;uniqueIndex:uniq_order_items_order_product"`
	Order   Order     `gorm:"foreignKey:OrderID;constraint:OnDelete:RESTRICT;"`

	ProductID   uuid.UUID `gorm:"type:uuid;not null;index:idx_order_items_product_id;uniqueIndex:uniq_order_items_order_product"`
	Product     Product   `gorm:"foreignKey:ProductID;constraint:OnDelete:RESTRICT;"`
	ProductName string    `gorm:"type:varchar(255);not null"`

	Quantity  int   `gorm:"not null;check:quantity > 0"`
	UnitPrice int64 `gorm:"not null;check:unit_price >= 0"`

	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`
}
