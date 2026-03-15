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
	"go-shop-backend/pkg/paymentprovider"
	"time"

	"github.com/google/uuid"
)

type orderService struct {
	orderRepo       repository.OrderRepository
	orderItemRepo   repository.OrderItemRepository
	productRepo     repository.ProductRepository
	paymentProvider paymentprovider.Provider
	txManager       database.TxManager
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	productRepo repository.ProductRepository,
	paymentProvider paymentprovider.Provider,
	txManager database.TxManager,
) *orderService {
	return &orderService{
		orderRepo:       orderRepo,
		orderItemRepo:   orderItemRepo,
		productRepo:     productRepo,
		paymentProvider: paymentProvider,
		txManager:       txManager,
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

		if err := product.CanBeAdded(req.Quantity); err != nil {
			return err
		}

		item := &models.OrderItem{
			OrderID:     orderID,
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    req.Quantity,
			UnitPrice:   product.Price,
		}

		if err := o.orderItemRepo.Upsert(ctx, item); err != nil {
			return err
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
	productID uuid.UUID,
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

		item, err := o.orderItemRepo.GetItem(ctx, productID, orderID)
		if err != nil {
			if !errors.Is(err, repository.ErrRecordNotFound) {
				return err
			}
			return apperrors.ErrItemNotFound
		}

		if err = o.orderItemRepo.DeleteItem(ctx, orderID, item.ProductID); err != nil {
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

func (o *orderService) Checkout(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
	orderID uuid.UUID,
) (*dto.OrderCheckoutResponse, error) {
	const op = "orderService.Checkout"

	var confirmationURL string

	checkout := func(ctx context.Context) error {
		order, err := o.getOrder(ctx, orderID, &userID, sessionID, true)
		if err != nil {
			return err
		}

		if err := order.Checkout(userID); err != nil {
			return err
		}

		if err := o.reserveItems(ctx, order.Items); err != nil {
			return err
		}

		req := paymentprovider.NewCreatePaymentRequest(userID, orderID, order.TotalAmount)
		payment, err := o.paymentProvider.CreatePayment(req)
		if err != nil {
			return err
		}

		expiresAt := time.Now().UTC().Add(10 * time.Minute)
		order.SetPaymentInfo(payment.ID, o.paymentProvider.GetName(), expiresAt)

		if err := o.orderRepo.Update(ctx, order); err != nil {
			return err
		}

		confirmationURL = payment.ConfirmationURL
		return nil
	}

	err := o.txManager.Wrap(ctx, checkout)
	if err != nil {
		if unavailableErr, ok := errors.AsType[*ItemsUnavailableError](err); ok {
			unavailableItems := make([]dto.UnavailableItem, len(unavailableErr.Items))

			for i, item := range unavailableErr.Items {
				unavailableItems[i] = dto.UnavailableItem{
					ProductID:    item.ProductID,
					RequestedQty: item.RequestedQty,
					AvailableQty: item.AvailableQty,
					Action:       item.Action,
					Reason:       item.Reason,
				}
			}

			return &dto.OrderCheckoutResponse{
				OrderID:          orderID,
				UnavailableItems: unavailableItems,
			}, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response := &dto.OrderCheckoutResponse{
		OrderID:         orderID,
		ConfirmationURL: confirmationURL,
	}

	return response, nil
}

func (o *orderService) HandlePaymentWebhook(
	ctx context.Context,
	body []byte,
) error {
	const op = "orderService.HandlePaymentWebhook"

	event, err := o.paymentProvider.ParseWebhook(body)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	handle := func(ctx context.Context) error {
		order, err := o.orderRepo.GetByPayment(ctx, o.paymentProvider.GetName(), event.PaymentID, true)
		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		if order.Status != models.OrderStatusPending {
			return nil
		}

		switch event.Status {
		case paymentprovider.PaymentStatusSucceeded:
			if err := o.deductItems(ctx, order.Items); err != nil {
				return err
			}

			if err := order.MarkPaid(); err != nil {
				return err
			}
		case paymentprovider.PaymentStatusCanceled:
			if err := o.releaseItems(ctx, order.Items); err != nil {
				return err
			}

			if err := order.MarkCanceled(); err != nil {
				return err
			}
		default:
			return nil
		}

		return o.orderRepo.Update(ctx, order)
	}

	err = o.txManager.Wrap(ctx, handle)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
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

	order.Recalculate()

	if err := o.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (o *orderService) reserveItems(ctx context.Context, items []models.OrderItem) error {
	return o.applyOnProducts(ctx, items, "reserve",
		func(p *models.Product, qty int) error {
			return p.Reserve(qty)
		},
	)
}

func (o *orderService) releaseItems(ctx context.Context, items []models.OrderItem) error {
	return o.applyOnProducts(ctx, items, "release",
		func(p *models.Product, qty int) error {
			return p.Release(qty)
		},
	)
}

func (o *orderService) deductItems(ctx context.Context, items []models.OrderItem) error {
	return o.applyOnProducts(ctx, items, "deduct",
		func(p *models.Product, qty int) error {
			return p.Deduct(qty)
		},
	)
}

type UnavailableItem struct {
	ProductID    uuid.UUID
	RequestedQty int
	AvailableQty int
	Action       string
	Reason       string
}

type ItemsUnavailableError struct {
	Items []UnavailableItem
}

func (e *ItemsUnavailableError) Error() string {
	return fmt.Sprintf("%d items unavailable", len(e.Items))
}

func (o *orderService) applyOnProducts(
	ctx context.Context,
	items []models.OrderItem,
	actionName string,
	action func(p *models.Product, qty int) error,
) error {
	productIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	products, err := o.productRepo.GetByIDsForUpdate(ctx, productIDs)
	if err != nil {
		return err
	}

	productMap := make(map[uuid.UUID]*models.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	var unavailable []UnavailableItem

	for _, item := range items {
		product := productMap[item.ProductID]
		if product == nil {
			unavailable = append(unavailable, UnavailableItem{
				ProductID:    item.ProductID,
				RequestedQty: item.Quantity,
				Action:       actionName,
				Reason:       "PRODUCT_NOT_FOUND",
			})
			continue
		}

		if err := action(product, item.Quantity); err != nil {
			unavailable = append(unavailable, UnavailableItem{
				ProductID:    product.ID,
				RequestedQty: item.Quantity,
				AvailableQty: product.Available(),
				Action:       actionName,
				Reason:       err.Error(),
			})
			continue
		}
	}

	if len(unavailable) > 0 {
		return &ItemsUnavailableError{
			Items: unavailable,
		}
	}

	return o.productRepo.BulkUpsert(ctx, products)
}
