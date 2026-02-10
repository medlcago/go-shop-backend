package dto

import "github.com/google/uuid"

type AddItemRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required,gt=0"`
}

type ItemResponse struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
}

type CartResponse struct {
	ID        uuid.UUID      `json:"id"`
	Items     []ItemResponse `json:"items"`
	TotalCost float64        `json:"total_cost"`
}
