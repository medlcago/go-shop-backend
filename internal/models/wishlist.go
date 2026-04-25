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
	ShareToken string    `gorm:"type:varchar(64);index:idx_wishlist_share_token;not null"`
	CreatedAt  time.Time `gorm:"type:timestamptz;default:now();not null"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;default:now();not null"`

	Items []WishlistItem `gorm:"foreignKey:WishlistID;constraint:OnDelete:CASCADE"`
}

func (w *Wishlist) CanView(userID uuid.UUID) bool {
	return w.IsPublic || w.UserID == userID
}

func (w *Wishlist) CanEdit(userID uuid.UUID) bool {
	return w.UserID == userID
}
