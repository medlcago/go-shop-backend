package upload

import (
	"time"

	"github.com/google/uuid"
)

type EntityType string

const (
	EntityTypeProduct EntityType = "products"
)

type Entity struct {
	ID   uuid.UUID
	Type EntityType
}

func NewEntity(id uuid.UUID, entityType EntityType) Entity {
	return Entity{
		ID:   id,
		Type: entityType,
	}
}

func NewProductEntity(id uuid.UUID) Entity {
	return NewEntity(id, EntityTypeProduct)
}

type SignURLRequest struct {
	ContentType string
	Entity      Entity
	Ext         string
}

type SignURLResponse struct {
	UploadID    uuid.UUID `json:"upload_id"`
	UploadURL   string    `json:"upload_url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	ExpireDate  time.Time `json:"expire_date"`

	FormData map[string]string `json:"form_data,omitempty"`
}

type SaveUploadRequest struct {
	UploadID  uuid.UUID
	ObjectKey string
	Entity    Entity
	IsMain    bool
}

type ContentResponse struct {
	URL         string    `json:"url"`
	ContentType *string   `json:"content_type"`
	IsMain      bool      `json:"is_main"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
