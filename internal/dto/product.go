package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Slug        string    `json:"slug"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Categories []string `json:"categories"`
}

type ListProductRequest struct {
	Limit      int       `query:"limit"`
	Offset     int       `query:"offset"`
	OrderBy    string    `query:"order_by"`
	OrderDesc  bool      `query:"order_desc"`
	CategoryID uuid.UUID `query:"category_id"`
}
