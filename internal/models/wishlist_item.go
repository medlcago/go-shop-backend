package models

import (
	"time"

	"github.com/google/uuid"
)

type WishlistItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WishlistID uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_wishlist_items_product_unique;not null"`
	ProductID  uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_wishlist_items_product_unique;not null"`
	Note       *string   `gorm:"type:varchar(128)"`
	Priority   int       `gorm:"default:0;not null"` // 0=normal, 1=high, 2=critical
	CreatedAt  time.Time `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;default:now();not null"`

	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE"`
}
