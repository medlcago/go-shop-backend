package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/mapper"

	"github.com/google/uuid"
)

type orderService struct {
	orderRepo     repository.OrderRepository
	orderItemRepo repository.OrderItemRepository
	productRepo   repository.ProductRepository
	txManager     database.TxManager
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	productRepo repository.ProductRepository,
	txManager database.TxManager,
) *orderService {
	return &orderService{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		productRepo:   productRepo,
		txManager:     txManager,
	}
}

func (o *orderService) CreateOrder(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
) (*dto.OrderResponse, error) {
	const op = "orderService.CreateOrder"

	order := &models.Order{
		UserID:    userID,
		SessionID: sessionID,
		Status:    models.OrderStatusDraft,
	}

	if err := o.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*models.Order, dto.OrderResponse](order)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to map order: %w", op, err)
	}

	return response, nil

}

func (o *orderService) GetOrder(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	orderID uuid.UUID,
) (*dto.OrderResponse, error) {
	const op = "orderService.GerOrder"

	order, err := o.getOrder(ctx, orderID, userID, sessionID, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*models.Order, dto.OrderResponse](order)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to map order: %w", op, err)
	}

	return response, nil

}

func (o *orderService) GetOrders(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	req dto.ListOrderRequest,
) ([]*dto.OrderResponse, int64, error) {
	const op = "orderService.GetDraftOrder"

	orders, total, err := o.orderRepo.GetListByOwner(ctx, userID, sessionID, req)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapList[*models.Order, *dto.OrderResponse](orders)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: failed to map orders: %w", op, err)
	}

	return response, total, nil

}

func (o *orderService) AddItem(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	orderID uuid.UUID,
	req dto.AddOrderItemRequest,
) (*dto.OrderResponse, error) {
	const op = "orderService.AddItem"

	if req.Quantity <= 0 {
		return nil, apperrors.ErrInvalidQuantity
	}

	var order *models.Order

	addItem := func(ctx context.Context) error {
		var err error

		order, err = o.getOrder(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return err
		}

		if !order.CanEdit() {
			return apperrors.ErrInvalidOrderStatus
		}

		product, err := o.productRepo.GetByID(ctx, req.ProductID, false)
		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return apperrors.ErrProductNotFound
			}
			return err
		}

		if !product.IsActive {
			return apperrors.ErrProductNotActive
		}

		if product.Available() < req.Quantity {
			return apperrors.ErrInsufficientStock
		}

		item, err := o.orderItemRepo.GetItem(ctx, product.ID, orderID)

		if err != nil {
			if !errors.Is(err, repository.ErrRecordNotFound) {
				return err
			}

			// item not found, create new
			item = &models.OrderItem{
				OrderID:     orderID,
				ProductID:   product.ID,
				ProductName: product.Name,
				Quantity:    req.Quantity,
				UnitPrice:   product.Price,
			}

			if err := o.orderItemRepo.AddItem(ctx, item); err != nil {
				return err
			}
		} else {
			// item found, update quantity
			if err = o.orderItemRepo.UpdateQuantity(ctx, item.ID, req.Quantity); err != nil {
				return err
			}
		}

		order, err = o.recalculateOrder(ctx, orderID, userID, sessionID)
		return err
	}

	err := o.txManager.Wrap(ctx, addItem)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*models.Order, dto.OrderResponse](order)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to map order: %w", op, err)
	}

	return response, nil
}

func (o *orderService) DeleteItem(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	orderID uuid.UUID,
	itemID uuid.UUID,
) (*dto.OrderResponse, error) {
	const op = "orderService.DeleteItem"

	var order *models.Order

	deleteItem := func(ctx context.Context) error {
		var err error

		order, err = o.getOrder(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return err
		}

		if !order.CanEdit() {
			return apperrors.ErrInvalidOrderStatus
		}

		if err = o.orderItemRepo.DeleteItem(ctx, orderID, itemID); err != nil {
			return err
		}

		order, err = o.recalculateOrder(ctx, orderID, userID, sessionID)
		return err
	}

	err := o.txManager.Wrap(ctx, deleteItem)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*models.Order, dto.OrderResponse](order)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to map order: %w", op, err)
	}

	return response, nil
}

func (o *orderService) Clear(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	orderID uuid.UUID,
) (*dto.OrderResponse, error) {
	const op = "orderService.Clear"

	var order *models.Order

	clearItems := func(ctx context.Context) error {
		var err error

		order, err = o.getOrder(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return err
		}

		if !order.CanEdit() {
			return apperrors.ErrInvalidOrderStatus
		}

		if err := o.orderItemRepo.Clear(ctx, orderID); err != nil {
			return err
		}

		order, err = o.recalculateOrder(ctx, orderID, userID, sessionID)
		return err
	}

	err := o.txManager.Wrap(ctx, clearItems)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := mapper.MapOne[*models.Order, dto.OrderResponse](order)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to map order: %w", op, err)
	}

	return response, nil
}

func (o *orderService) calculateTotal(order *models.Order) int64 {
	var total int64

	for _, item := range order.Items {
		total += int64(item.Quantity) * item.UnitPrice
	}

	return total
}

func (o *orderService) getOrder(
	ctx context.Context,
	orderID uuid.UUID,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	preload bool,
) (*models.Order, error) {
	order, err := o.orderRepo.GetByOwner(ctx, orderID, userID, sessionID, preload)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperrors.ErrForbidden
		}
		return nil, err
	}

	return order, nil
}

func (o *orderService) recalculateOrder(
	ctx context.Context,
	orderID uuid.UUID,
	userID *uuid.UUID,
	sessionID uuid.UUID,
) (*models.Order, error) {
	order, err := o.getOrder(ctx, orderID, userID, sessionID, true)
	if err != nil {
		return nil, err
	}

	order.TotalAmount = o.calculateTotal(order)

	if err := o.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}
