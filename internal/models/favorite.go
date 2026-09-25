package models

import (
	"time"

	"github.com/google/uuid"
)

type FavoriteProductProjection struct {
	ID        uuid.UUID
	Name      string
	Price     int64
	Available int
	CreatedAt time.Time
}

type Favorite struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"not null;uniqueIndex:idx_favorites_user_product_unique"`
	ProductID uuid.UUID `gorm:"not null;uniqueIndex:idx_favorites_user_product_unique"`
	CreatedAt time.Time `gorm:"type:timestamptz;default:now();not null"`

	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;"`
}
