package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateAddressRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=100"`
	Street  string `json:"street" validate:"required,min=1,max=255"`
	House   string `json:"house" validate:"required,min=1,max=50"`
	City    string `json:"city" validate:"required,min=1,max=255"`
	Country string `json:"country" validate:"required,min=1,max=100"`

	Floor     *string `json:"floor" validate:"omitempty,min=1,max=50"`
	Entrance  *string `json:"entrance" validate:"omitempty,min=1,max=50"`
	Apartment *string `json:"apartment" validate:"omitempty,min=1,max=50"`
	Comment   *string `json:"comment" validate:"omitempty,min=1,max=500"`
}

type UpdateAddressRequest struct {
	CreateAddressRequest
}

type AddressResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Street  string    `json:"street"`
	House   string    `json:"house"`
	City    string    `json:"city"`
	Country string    `json:"country"`

	Floor     *string   `json:"floor"`
	Entrance  *string   `json:"entrance"`
	Comment   *string   `json:"comment"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
