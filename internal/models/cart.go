package models

import "github.com/google/uuid"

type Cart struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    *uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_carts_user_id_unique,where:user_id IS NOT NULL;index:idx_carts_user_id"`
	SessionID *uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_carts_session_id_unique,where:session_id IS NOT NULL;index:idx_carts_session_id"`
	Items     []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE;"`
}

type CartItem struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CartID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_cart_items_cart_product_unique;index:idx_cart_items_cart_id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_cart_items_cart_product_unique;index:idx_cart_items_product_id"`
	Quantity  int       `gorm:"not null;default:1;check:quantity > 0"`
	UnitPrice float64   `gorm:"type:decimal(10,2);not null;check:unit_price >= 0"`
}
