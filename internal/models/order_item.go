package models

import (
	"time"

	"github.com/google/uuid"
)

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
