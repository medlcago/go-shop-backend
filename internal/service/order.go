package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/tasks"
	"go-shop-backend/internal/upload"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/mapper"
	"time"

	"github.com/google/uuid"
)

type orderService struct {
	orderRepo        repository.OrderRepository
	orderItemRepo    repository.OrderItemRepository
	productRepo      repository.ProductRepository
	addressRepo      repository.AddressRepository
	orderTask        tasks.OrderTask
	txManager        database.TxManager
	orderCancelDelay time.Duration
	publicURLBuilder upload.PublicURLBuilder
	inventoryService InventoryService
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	productRepo repository.ProductRepository,
	addressRepo repository.AddressRepository,
	orderTask tasks.OrderTask,
	txManager database.TxManager,
	orderCancelDelay time.Duration,
	publicURLBuilder upload.PublicURLBuilder,
	inventoryService InventoryService,
) *orderService {
	return &orderService{
		orderRepo:        orderRepo,
		orderItemRepo:    orderItemRepo,
		productRepo:      productRepo,
		addressRepo:      addressRepo,
		orderTask:        orderTask,
		txManager:        txManager,
		orderCancelDelay: orderCancelDelay,
		publicURLBuilder: publicURLBuilder,
		inventoryService: inventoryService,
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
		return nil, apperror.Wrap(op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (o *orderService) GetOrder(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	orderID uuid.UUID,
) (*dto.OrderResponse, error) {
	const op = "orderService.GetOrder"

	order, err := o.getOrderByOwner(ctx, orderID, userID, sessionID, true)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (o *orderService) GetOrders(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	req dto.ListOrderRequest,
) ([]*dto.OrderResponse, int64, error) {
	const op = "orderService.GetOrders"

	orders, total, err := o.orderRepo.GetListByOwner(ctx, userID, sessionID, req)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	response, err := o.mapOrders(orders)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
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
		return nil, apperror.Wrap(op, apperror.ErrInvalidQuantity)
	}

	order, err := database.Transaction(ctx, o.txManager, func(ctx context.Context) (*models.Order, error) {
		order, err := o.getOrderByOwner(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return nil, err
		}

		if order.Status != models.OrderStatusDraft {
			return nil, apperror.ErrInvalidOrderStatus
		}

		product, err := o.inventoryService.CheckProduct(ctx, req.ProductID, req.Quantity)
		if err != nil {
			return nil, err
		}

		item := &models.OrderItem{
			OrderID:     orderID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    req.Quantity,
			UnitPrice:   product.Price,
		}

		if err := o.orderItemRepo.Upsert(ctx, item); err != nil {
			return nil, err
		}

		return o.recalculateOrder(ctx, orderID)
	})

	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (o *orderService) RemoveItem(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	orderID uuid.UUID,
	itemID uuid.UUID,
) (*dto.OrderResponse, error) {
	const op = "orderService.RemoveItem"

	order, err := database.Transaction(ctx, o.txManager, func(ctx context.Context) (*models.Order, error) {
		order, err := o.getOrderByOwner(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return nil, err
		}

		if order.Status != models.OrderStatusDraft {
			return nil, apperror.ErrInvalidOrderStatus
		}

		removed, err := o.orderItemRepo.RemoveItem(ctx, orderID, itemID)
		if err != nil {
			return nil, err
		}

		if !removed {
			return nil, apperror.ErrItemNotFound
		}

		return o.recalculateOrder(ctx, orderID)
	})

	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, apperror.Wrap(op, err)
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

	order, err := database.Transaction(ctx, o.txManager, func(ctx context.Context) (*models.Order, error) {
		order, err := o.getOrderByOwner(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return nil, err
		}

		if order.Status != models.OrderStatusDraft {
			return nil, apperror.ErrInvalidOrderStatus
		}

		if err := o.orderItemRepo.Clear(ctx, orderID); err != nil {
			return nil, err
		}

		return o.recalculateOrder(ctx, orderID)
	})

	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (o *orderService) Checkout(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
	orderID uuid.UUID,
	req dto.OrderCheckoutRequest,
) (*dto.OrderResponse, error) {
	const op = "orderService.Checkout"

	order, err := database.Transaction(ctx, o.txManager, func(ctx context.Context) (*models.Order, error) {
		address, err := o.addressRepo.GetByID(ctx, req.AddressID, userID)
		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return nil, apperror.ErrAddressNotFound
			}

			return nil, err
		}

		order, err := o.getOrderByOwner(ctx, orderID, &userID, sessionID, true)
		if err != nil {
			return nil, err
		}

		// If the order does not have a user, we link the order to the user
		if order.UserID == nil {
			order.UserID = &userID
		}

		if !order.IsOwnedBy(userID) {
			return nil, apperror.ErrForbidden
		}

		if err := order.Checkout(); err != nil {
			return nil, err
		}

		if err := o.inventoryService.ReserveItems(ctx, o.mapOrderItemsToInventoryItems(order.Items)); err != nil {
			return nil, err
		}

		order.ExpiresAt = new(time.Now().UTC().Add(o.orderCancelDelay))
		order.Address = address

		if err := o.orderRepo.Update(ctx, order); err != nil {
			return nil, err
		}

		if err := o.orderTask.EnqueueCancelOrder(ctx, userID, orderID, o.orderCancelDelay); err != nil {
			return nil, err
		}

		return order, nil
	})

	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (o *orderService) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status models.OrderStatus) error {
	const op = "orderService.UpdateOrderStatus"

	err := o.txManager.Wrap(ctx, func(ctx context.Context) error {
		order, err := o.getOrderByID(ctx, orderID, true)
		if err != nil {
			return err
		}

		if !status.IsValid() {
			return apperror.ErrInvalidOrderStatus
		}

		if !order.Status.CanTransitionTo(status) {
			return apperror.ErrInvalidOrderStatus
		}

		switch status {
		case models.OrderStatusPaid:
			if err := order.Pay(); err != nil {
				return err
			}
			if err := o.inventoryService.DeductItems(ctx, o.mapOrderItemsToInventoryItems(order.Items)); err != nil {
				return err
			}
		case models.OrderStatusCanceled:
			if err := order.Cancel(); err != nil {
				return err
			}
			if err := o.inventoryService.ReleaseItems(ctx, o.mapOrderItemsToInventoryItems(order.Items)); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported status: %s", status)
		}

		return o.orderRepo.Update(ctx, order)
	})

	if err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (o *orderService) CancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error {
	const op = "orderService.CancelOrder"

	err := o.txManager.Wrap(ctx, func(ctx context.Context) error {
		order, err := o.getOrderByID(ctx, orderID, true)
		if err != nil {
			return err
		}

		if !order.IsOwnedBy(userID) {
			return apperror.ErrForbidden
		}

		if err := order.Cancel(); err != nil {
			return err
		}

		if err := o.inventoryService.ReleaseItems(ctx, o.mapOrderItemsToInventoryItems(order.Items)); err != nil {
			return err
		}

		return o.orderRepo.Update(ctx, order)
	})

	if err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (o *orderService) recalculateOrder(
	ctx context.Context,
	orderID uuid.UUID,
) (*models.Order, error) {
	const op = "orderService.recalculateOrder"

	order, err := o.getOrderByID(ctx, orderID, true)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	order.Recalculate()

	if err := o.orderRepo.Update(ctx, order); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return order, nil
}

func (o *orderService) getOrderByOwner(
	ctx context.Context,
	orderID uuid.UUID,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	preload bool,
) (*models.Order, error) {
	const op = "orderService.getOrderByOwner"

	order, err := o.orderRepo.GetByOwner(ctx, orderID, userID, sessionID, preload)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrForbidden)
		}

		return nil, apperror.Wrap(op, err)
	}

	return order, nil
}

func (o *orderService) getOrderByID(
	ctx context.Context,
	orderID uuid.UUID,
	preload bool,
) (*models.Order, error) {
	const op = "orderService.getOrderByID"

	order, err := o.orderRepo.GetByID(ctx, orderID, preload)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrOrderNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return order, nil
}

func (o *orderService) mapOrder(order *models.Order) (*dto.OrderResponse, error) {
	const op = "orderService.mapOrder"

	for _, item := range order.Items {
		upload.AssignPublicURLs(item.Product.Images, o.publicURLBuilder)
	}

	response, err := mapper.MapOne[*models.Order, dto.OrderResponse](order)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (o *orderService) mapOrders(orders []*models.Order) ([]*dto.OrderResponse, error) {
	const op = "orderService.mapOrders"

	for _, order := range orders {
		for _, item := range order.Items {
			upload.AssignPublicURLs(item.Product.Images, o.publicURLBuilder)
		}
	}

	response, err := mapper.MapList[*models.Order, *dto.OrderResponse](orders)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (o *orderService) mapOrderItemsToInventoryItems(orderItems []models.OrderItem) []dto.InventoryItem {
	inventoryItems := make([]dto.InventoryItem, 0, len(orderItems))

	for _, item := range orderItems {
		inventoryItems = append(inventoryItems, dto.InventoryItem{
			ItemID:    item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return inventoryItems
}
