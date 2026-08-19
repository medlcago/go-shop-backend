package dto

import (
	"time"

	"github.com/google/uuid"
)

type UploadEntity struct {
	ID   uuid.UUID `json:"id" validate:"required"`
	Type string    `json:"type" validate:"required"`
}

func NewUploadEntity(id uuid.UUID, entityType string) UploadEntity {
	return UploadEntity{
		ID:   id,
		Type: entityType,
	}
}

type GeneratePresignedURLRequest struct {
	ContentType string       `json:"content_type" validate:"required"`
	Entity      UploadEntity `json:"entity" validate:"required"`
	Ext         string       `json:"ext" validate:"required"`
}

type GeneratePresignedURLResponse struct {
	UploadID    uuid.UUID `json:"upload_id"`
	UploadURL   string    `json:"upload_url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	ExpireDate  time.Time `json:"expire_date"`

	FormData map[string]string `json:"form_data,omitempty"`
}

type AttachFileRequest struct {
	UploadID  uuid.UUID    `json:"upload_id" validate:"required"`
	ObjectKey string       `json:"object_key" validate:"required"`
	Entity    UploadEntity `json:"entity" validate:"required"`
}

type UploadResponse struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	ContentType *string   `json:"content_type"`
	MediaType   string    `json:"media_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
