package dto

import (
	"time"

	"github.com/google/uuid"
)

type UploadEntity struct {
	ID   uuid.UUID
	Type string
}

func NewUploadEntity(id uuid.UUID, entityType string) UploadEntity {
	return UploadEntity{
		ID:   id,
		Type: entityType,
	}
}

type UploadSignURLRequest struct {
	ContentType string
	Entity      UploadEntity
	Ext         string
}

type UploadSignURLResponse struct {
	UploadID    uuid.UUID `json:"upload_id"`
	UploadURL   string    `json:"upload_url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	ExpireDate  time.Time `json:"expire_date"`

	FormData map[string]string `json:"form_data,omitempty"`
}

type UploadSaveRequest struct {
	UploadID  uuid.UUID
	ObjectKey string
	Entity    UploadEntity
	IsMain    bool
}

type UploadResponse struct {
	URL         string    `json:"url"`
	ContentType *string   `json:"content_type"`
	MediaType   string    `json:"media_type"`
	Variant     string    `json:"variant"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
