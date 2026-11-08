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
	ContentType    string       `json:"content_type" validate:"required"`
	Entity         UploadEntity `json:"entity" validate:"required"`
	Ext            string       `json:"ext" validate:"required"`
	Purpose        string       `json:"purpose" validate:"required"`
	ChecksumSHA256 string       `json:"checksum_sha_256" validate:"required,base64"`
}

type GeneratePresignedURLResponse struct {
	UploadID       uuid.UUID `json:"upload_id"`
	UploadURL      string    `json:"upload_url"`
	Method         string    `json:"method"`
	Filename       string    `json:"filename"`
	ContentType    string    `json:"content_type"`
	ExpireDate     time.Time `json:"expire_date"`
	ChecksumSHA256 string    `json:"checksum_sha_256"`

	FormData map[string]string `json:"form_data"`
}

type AttachFileRequest struct {
	UploadID  uuid.UUID    `json:"upload_id" validate:"required"`
	ObjectKey string       `json:"object_key" validate:"required"`
	Entity    UploadEntity `json:"entity" validate:"required"`
	Purpose   string       `json:"purpose" validate:"required"`
}

type AttachFileResponse struct {
	ID  uuid.UUID `json:"id"`
	URL string    `json:"url"`
}

type UploadResponse struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	ContentType *string   `json:"content_type"`
	Purpose     string    `json:"purpose"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
