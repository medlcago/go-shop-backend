package dto

import (
	"time"

	"github.com/google/uuid"
)

type EntityType string

const (
	EntityProduct EntityType = "products"
)

type Entity struct {
	ID   uuid.UUID  `json:"id" validate:"required"`
	Type EntityType `json:"type" validate:"required"`
}

type SignURLRequest struct {
	ContentType string `json:"content_type" validate:"required"`
	Entity      Entity `json:"entity" validate:"required"`
	Ext         string `json:"ext" validate:"required,oneof=jpg png"`
}

type SignURLResponse struct {
	UploadID    uuid.UUID `json:"upload_id"`
	UploadURL   string    `json:"upload_url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	ExpireDate  time.Time `json:"expire_date"`

	FormData map[string]string `json:"form_data,omitempty"`
}

type UploadRequest struct {
	UploadID  uuid.UUID `json:"upload_id" validate:"required"`
	ObjectKey string    `json:"object_key" validate:"required"`
	Entity    Entity    `json:"entity" validate:"required"`
}

type UploadResponse struct {
	URL         string    `json:"url"`
	ContentType *string   `json:"content_type"`
	IsMain      bool      `json:"is_main"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
