package dto

import (
	"time"

	"github.com/google/uuid"
)

type FavoriteProduct struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Price     int64     `json:"price"`
	Available int       `json:"available"`
	CreatedAt time.Time `json:"created_at"`
}

type FavoriteStatusResponse struct {
	ProductID  uuid.UUID `json:"product_id"`
	IsFavorite bool      `json:"is_favorite"`
}
