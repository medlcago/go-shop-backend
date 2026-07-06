package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

type orderItemRepository struct {
	db database.Provider
}

func NewOrderItemRepository(db database.Provider) *orderItemRepository {
	return &orderItemRepository{
		db: db,
	}
}

func (o *orderItemRepository) AddItem(ctx context.Context, orderItem *models.OrderItem) error {
	db := o.db.GetDB(ctx)

	err := db.Create(orderItem).Error
	return repository.HandleError(err)
}

func (o *orderItemRepository) Upsert(ctx context.Context, orderItem *models.OrderItem) error {
	db := o.db.GetDB(ctx)

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "order_id"}, {Name: "product_id"}},
		UpdateAll: true,
	}).Create(orderItem).Error

	return repository.HandleError(err)
}

func (o *orderItemRepository) RemoveItem(ctx context.Context, orderID uuid.UUID, itemID uuid.UUID) error {
	db := o.db.GetDB(ctx)

	result := db.Where("id = ? AND order_id = ?", itemID, orderID).
		Delete(&models.OrderItem{})

	if result.Error != nil {
		return repository.HandleError(result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrRecordNotFound
	}

	return nil
}

func (o *orderItemRepository) Clear(ctx context.Context, orderID uuid.UUID) error {
	db := o.db.GetDB(ctx)

	err := db.Where("order_id = ?", orderID).
		Delete(&models.OrderItem{}).Error

	return repository.HandleError(err)
}
