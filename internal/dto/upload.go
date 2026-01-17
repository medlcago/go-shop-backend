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
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	ExpireDate  time.Time `json:"expire_date"`

	FormData map[string]string `json:"form_data,omitempty"`
}
