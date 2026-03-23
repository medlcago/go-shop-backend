package gorm

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/repository/gorm/scopes"
	"go-shop-backend/pkg/database"
	"time"

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

	err := db.Select("*").Updates(order).Error
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

	db = db.Where("id = ?", orderID).
		Scopes(
			scopes.OrderOwner(userID, sessionID),
		)

	if preload {
		db = db.Scopes(
			scopes.OrderWithRelations(),
		)
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

	db = db.Scopes(
		scopes.OrderOwner(userID, sessionID),
		scopes.OrderStatus(req.Status),
	)

	var total int64
	if err := db.Model(&models.Order{}).Count(&total).Error; err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	var orders []*models.Order
	err := db.
		Scopes(
			scopes.OrderWithRelations(),
			scopes.Paginate(req.Limit, req.Offset),
		).
		Find(&orders).Error

	if err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	return orders, total, nil
}

func (o *orderRepository) GetByPayment(
	ctx context.Context,
	providerName string,
	paymentID string,
	preload bool,
) (*models.Order, error) {
	db := o.db.GetDB(ctx)

	db = db.Where(
		"provider_name = ? AND payment_id = ?",
		providerName, paymentID,
	)

	if preload {
		db = db.Scopes(
			scopes.OrderWithRelations(),
		)
	}

	var order models.Order
	err := db.First(&order).Error

	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &order, nil
}

func (o *orderRepository) CancelExpiredPending(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]uuid.UUID, error) {

	db := o.db.GetDB(ctx)

	var ids []uuid.UUID

	err := db.Raw(`
		WITH expired_orders AS (
			SELECT id
			FROM orders
			WHERE status = 'pending'
			  AND expires_at < ?
			ORDER BY expires_at ASC
			LIMIT ?
			FOR UPDATE SKIP LOCKED
		)
		UPDATE orders o
		SET status = 'canceled',
		    updated_at = ?,
			canceled_at = ?
		FROM expired_orders
		WHERE o.id = expired_orders.id
		RETURNING o.id;
	`, now, limit, now, now).Scan(&ids).Error

	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return ids, nil
}
