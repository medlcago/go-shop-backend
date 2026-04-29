package models

import (
	"time"

	"github.com/google/uuid"
)

type Wishlist struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID `gorm:"type:uuid;not null"`
	Title      string    `gorm:"type:varchar(255);not null"`
	IsPublic   bool      `gorm:"default:false;not null"`
	ShareToken string    `gorm:"type:varchar(64);uniqueIndex:idx_wishlists_share_token_unique;not null"`
	CreatedAt  time.Time `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;default:now();not null"`

	Items []WishlistItem `gorm:"foreignKey:WishlistID;constraint:OnDelete:CASCADE"`
}

func (w *Wishlist) IsOwnedBy(userID uuid.UUID) bool {
	return w.UserID == userID
}
