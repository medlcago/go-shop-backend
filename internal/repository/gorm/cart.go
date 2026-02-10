package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cartRepository struct {
	db database.Provider
}

func NewCartRepository(db database.Provider) repository.CartRepository {
	return &cartRepository{
		db: db,
	}
}

func (c *cartRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error) {
	db := c.db.GetDB(ctx)

	var cart models.Cart
	if err := db.Where("user_id = ?", userID).
		Preload("Items").
		First(&cart).Error; err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &cart, nil
}

func (c *cartRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.Cart, error) {
	db := c.db.GetDB(ctx)

	var cart models.Cart
	if err := db.Where("session_id = ?", sessionID).
		Preload("Items").
		First(&cart).Error; err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &cart, nil
}

func (c *cartRepository) Create(ctx context.Context, cart *models.Cart) error {
	db := c.db.GetDB(ctx)

	err := db.Create(cart).Error
	return repository.HandleSQLError(err)
}

func (c *cartRepository) Save(ctx context.Context, cart *models.Cart) error {
	db := c.db.GetDB(ctx).Session(&gorm.Session{FullSaveAssociations: true})

	err := db.Save(cart).Error
	return repository.HandleSQLError(err)
}

func (c *cartRepository) DeleteItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID) error {
	db := c.db.GetDB(ctx)

	err := db.Where("cart_id = ? AND product_id = ?", cartID, productID).
		Delete(&models.CartItem{}).Error

	return repository.HandleSQLError(err)
}
