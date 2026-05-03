package dto

import (
	"time"

	"github.com/google/uuid"
)

type ProductResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Price       int64     `json:"price"`
	Stock       int       `json:"stock"`
	Available   int       `json:"available"`
	Reserved    int       `json:"reserved"`
	Slug        string    `json:"slug"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Categories []ProductCategoryResponse `json:"categories"`
	Images     []UploadResponse          `json:"images"`
}

type ListProductRequest struct {
	Limit      int       `query:"limit" validate:"omitempty,gte=1,lte=50"`
	Offset     int       `query:"offset" validate:"omitempty,gte=0"`
	OrderBy    string    `query:"order_by"`
	OrderDesc  bool      `query:"order_desc"`
	CategoryID uuid.UUID `query:"category_id"`
	MinPrice   int64     `query:"min_price"  validate:"omitempty,gte=0"`
	MaxPrice   int64     `query:"max_price"  validate:"omitempty,gte=0,gtefield=MinPrice"`
}

type ProductCreateRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,min=5,max=4096"`
	Price       int64   `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"required,gt=0"`
	IsActive    *bool   `json:"is_active"`
}

type ProductUpdateRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,min=5,max=4096"`
	Price       *int64  `json:"price" validate:"omitempty,gt=0"`
	Stock       *int    `json:"stock" validate:"omitempty,gte=0"`
	IsActive    *bool   `json:"is_active"`
}

type SearchProductRequest struct {
	Query  string `query:"q" validate:"required"`
	Limit  int    `query:"limit" validate:"omitempty,gte=1,lte=50"`
	Offset int    `query:"offset" validate:"omitempty,gte=0"`
}

type UploadProductImageRequest struct {
	ContentType string `json:"content_type" validate:"required"`
	Ext         string `json:"ext" validate:"required,oneof=jpg png"`
}

type ConfirmUploadProductImageRequest struct {
	UploadID  uuid.UUID `json:"upload_id" validate:"required"`
	ObjectKey string    `json:"object_key" validate:"required"`
}
