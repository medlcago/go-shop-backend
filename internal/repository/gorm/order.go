package gorm

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/paging"

	"github.com/google/uuid"
)

type orderRepository struct {
	db database.Provider
}

func NewOrderRepository(db database.Provider) *orderRepository {
	return &orderRepository{
		db: db,
	}
}

func (o *orderRepository) Create(ctx context.Context, order *models.Order) error {
	db := o.db.GetDB(ctx)

	err := db.Create(order).Error
	return repository.HandleSQLError(err)
}

func (o *orderRepository) Update(ctx context.Context, order *models.Order) error {
	db := o.db.GetDB(ctx)

	err := db.Updates(order).Error
	return repository.HandleSQLError(err)
}

func (o *orderRepository) GetByOwner(
	ctx context.Context,
	orderID uuid.UUID,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	preload bool,
) (*models.Order, error) {
	db := o.db.GetDB(ctx)

	db = db.Where("id = ?", orderID)

	if userID != nil {
		db = db.Where(
			"(user_id = ? OR session_id = ?)",
			*userID,
			sessionID,
		)
	} else {
		db = db.Where("session_id = ?", sessionID)
	}

	if preload {
		db = db.Preload("Items")
	}

	var order models.Order
	err := db.First(&order).Error

	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &order, nil
}

func (o *orderRepository) GetListByOwner(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	req dto.ListOrderRequest,
) ([]*models.Order, int64, error) {
	db := o.db.GetDB(ctx)

	if userID != nil {
		db = db.Where(
			"(user_id = ? OR session_id = ?)",
			*userID,
			sessionID,
		)
	} else {
		db = db.Where("session_id = ?", sessionID)
	}

	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	var total int64
	if err := db.Model(&models.Order{}).Count(&total).Error; err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	pagination := paging.New(req.Limit, req.Offset)

	var orders []*models.Order
	err := db.Preload("Items").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&orders).Error

	if err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	return orders, total, nil
}
