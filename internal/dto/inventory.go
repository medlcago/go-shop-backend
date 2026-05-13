package dto

import "github.com/google/uuid"

type InventoryItem struct {
	ItemID    uuid.UUID `json:"item_id"`
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}
