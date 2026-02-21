package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
)

type orderItemRepository struct {
	db database.Provider
}

func NewOrderItemRepository(db database.Provider) *orderItemRepository {
	return &orderItemRepository{
		db: db,
	}
}
func (o *orderItemRepository) GetItem(ctx context.Context, productID uuid.UUID, orderID uuid.UUID) (*models.OrderItem, error) {
	db := o.db.GetDB(ctx)

	var item models.OrderItem
	if err := db.Where("order_id = ? AND product_id = ?", orderID, productID).First(&item).Error; err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &item, nil
}

func (o *orderItemRepository) AddItem(ctx context.Context, orderItem *models.OrderItem) error {
	db := o.db.GetDB(ctx)

	err := db.Create(orderItem).Error
	return repository.HandleSQLError(err)
}

func (o *orderItemRepository) UpdateQuantity(ctx context.Context, itemID uuid.UUID, qty int) error {
	db := o.db.GetDB(ctx)

	err := db.Model(&models.OrderItem{}).
		Where("id = ?", itemID).
		Update("quantity", qty).Error

	return repository.HandleSQLError(err)
}

func (o *orderItemRepository) DeleteItem(ctx context.Context, orderID uuid.UUID, productID uuid.UUID) error {
	db := o.db.GetDB(ctx)

	err := db.Where("order_id = ? AND product_id = ?", orderID, productID).
		Delete(&models.OrderItem{}).Error

	return repository.HandleSQLError(err)
}

func (o *orderItemRepository) Clear(ctx context.Context, orderID uuid.UUID) error {
	db := o.db.GetDB(ctx)

	err := db.Where("order_id = ?", orderID).
		Delete(&models.OrderItem{}).Error

	return repository.HandleSQLError(err)
}
