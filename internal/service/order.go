package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/tasks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/mapper"
	"go-shop-backend/pkg/paymentprovider"
	"time"

	"github.com/google/uuid"
)

type orderService struct {
	orderRepo            repository.OrderRepository
	orderItemRepo        repository.OrderItemRepository
	productRepo          repository.ProductRepository
	paymentProvider      paymentprovider.Provider
	orderTask            tasks.OrderTask
	txManager            database.TxManager
	orderCancelDelay     time.Duration
	orderCheckoutTimeout time.Duration
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	productRepo repository.ProductRepository,
	paymentProvider paymentprovider.Provider,
	orderTask tasks.OrderTask,
	txManager database.TxManager,
	orderCancelDelay time.Duration,
	orderCheckoutTimeout time.Duration,
) *orderService {
	return &orderService{
		orderRepo:            orderRepo,
		orderItemRepo:        orderItemRepo,
		productRepo:          productRepo,
		paymentProvider:      paymentProvider,
		orderTask:            orderTask,
		txManager:            txManager,
		orderCancelDelay:     orderCancelDelay,
		orderCheckoutTimeout: orderCheckoutTimeout,
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

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	response, err := o.mapOrders(orders)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
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
		return nil, apperror.ErrInvalidQuantity
	}

	var order *models.Order

	handle := func(ctx context.Context) error {
		var err error

		order, err = o.getOrder(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return err
		}

		if !order.CanEdit() {
			return apperror.ErrInvalidOrderStatus
		}

		product, err := o.productRepo.GetByID(ctx, req.ProductID, false)
		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return apperror.ErrProductNotFound
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

	err := o.txManager.Wrap(ctx, handle)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	handle := func(ctx context.Context) error {
		var err error

		order, err = o.getOrder(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return err
		}

		if !order.CanEdit() {
			return apperror.ErrInvalidOrderStatus
		}

		item, err := o.orderItemRepo.GetItem(ctx, productID, orderID)
		if err != nil {
			if !errors.Is(err, repository.ErrRecordNotFound) {
				return err
			}
			return apperror.ErrItemNotFound
		}

		if err = o.orderItemRepo.DeleteItem(ctx, orderID, item.ProductID); err != nil {
			return err
		}

		order, err = o.recalculateOrder(ctx, orderID, userID, sessionID)
		return err
	}

	err := o.txManager.Wrap(ctx, handle)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	handle := func(ctx context.Context) error {
		var err error

		order, err = o.getOrder(ctx, orderID, userID, sessionID, false)
		if err != nil {
			return err
		}

		if !order.CanEdit() {
			return apperror.ErrInvalidOrderStatus
		}

		if err := o.orderItemRepo.Clear(ctx, orderID); err != nil {
			return err
		}

		order, err = o.recalculateOrder(ctx, orderID, userID, sessionID)
		return err
	}

	err := o.txManager.Wrap(ctx, handle)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := o.mapOrder(order)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	ctx, cancel := context.WithTimeout(ctx, o.orderCheckoutTimeout)
	defer cancel()

	response := &dto.OrderCheckoutResponse{
		OrderID: orderID,
	}

	handle := func(ctx context.Context) error {
		order, err := o.getOrder(ctx, orderID, &userID, sessionID, true)
		if err != nil {
			return err
		}

		if err := o.checkoutOrder(ctx, order, userID); err != nil {
			if unavailableErr, ok := errors.AsType[*apperror.ItemsUnavailableError](err); ok {
				var unavailableItems dto.UnavailableItems
				response.UnavailableItems = unavailableItems.FromErr(unavailableErr)
				return nil
			}

			return err
		}

		payment, err := o.createPayment(ctx, userID, orderID, order.TotalAmount)
		if err != nil {
			return err
		}

		expiresAt := time.Now().UTC().Add(o.orderCancelDelay)
		order.SetPaymentInfo(payment.ID, o.paymentProvider.GetName(), expiresAt)

		if err := o.orderRepo.Update(ctx, order); err != nil {
			return err
		}

		if err := o.orderTask.EnqueueCancelOrder(ctx, userID, orderID, o.orderCancelDelay); err != nil {
			return err
		}

		response.ConfirmationURL = payment.ConfirmationURL
		return nil
	}

	err := o.txManager.Wrap(ctx, handle)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%s: %w", op, apperror.ErrGatewayTimeout)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (o *orderService) HandlePaymentWebhook(ctx context.Context, body []byte) error {
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

		if err := o.processWebhookEvent(ctx, event, order); err != nil {
			return err
		}

		return o.orderRepo.Update(ctx, order)
	}

	err = o.txManager.Wrap(ctx, handle)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) CancelOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error {
	const op = "orderService.CancelOrder"

	handle := func(ctx context.Context) error {
		order, err := o.orderRepo.GetByID(ctx, orderID, true)
		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				return apperror.ErrOrderNotFound
			}

			return fmt.Errorf("%s: %w", op, err)
		}

		if err := o.cancelOrder(ctx, order, userID); err != nil {
			return err
		}

		return o.orderRepo.Update(ctx, order)
	}

	err := o.txManager.Wrap(ctx, handle)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) processWebhookEvent(ctx context.Context, event *paymentprovider.WebhookEvent, order *models.Order) error {
	const op = "orderService.processWebhookEvent"

	var err error

	switch event.Status {
	case paymentprovider.PaymentStatusSucceeded:
		err = o.payOrder(ctx, order, event.Metadata.UserID)
	case paymentprovider.PaymentStatusCanceled:
		err = o.cancelOrder(ctx, order, event.Metadata.UserID)
	default:
		return nil
	}

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) checkoutOrder(ctx context.Context, order *models.Order, userID uuid.UUID) error {
	const op = "orderService.checkoutOrder"

	if err := order.Checkout(userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := o.reserveItems(ctx, order.Items); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) payOrder(ctx context.Context, order *models.Order, userID uuid.UUID) error {
	const op = "orderService.payOrder"

	if err := order.Pay(userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := o.deductItems(ctx, order.Items); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) cancelOrder(ctx context.Context, order *models.Order, userID uuid.UUID) error {
	const op = "orderService.cancelOrder"

	if err := order.Cancel(userID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := o.releaseItems(ctx, order.Items); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) recalculateOrder(
	ctx context.Context,
	orderID uuid.UUID,
	userID *uuid.UUID,
	sessionID uuid.UUID,
) (*models.Order, error) {
	const op = "orderService.recalculateOrder"

	order, err := o.getOrder(ctx, orderID, userID, sessionID, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	order.Recalculate()

	if err := o.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}

func (o *orderService) getOrder(
	ctx context.Context,
	orderID uuid.UUID,
	userID *uuid.UUID,
	sessionID uuid.UUID,
	preload bool,
) (*models.Order, error) {
	const op = "orderService.getOrder"

	order, err := o.orderRepo.GetByOwner(ctx, orderID, userID, sessionID, preload)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, fmt.Errorf("%s: %w", op, apperror.ErrForbidden)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return order, nil
}

func (o *orderService) createPayment(
	ctx context.Context,
	userID, orderID uuid.UUID,
	amount int64,
) (*paymentprovider.Payment, error) {
	const op = "orderService.createPayment"

	req := paymentprovider.NewCreatePaymentRequest(userID, orderID, amount)
	payment, err := o.paymentProvider.CreatePayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return payment, nil
}

func (o *orderService) mapOrder(order *models.Order) (*dto.OrderResponse, error) {
	const op = "orderService.mapOrder"

	response, err := mapper.MapOne[*models.Order, dto.OrderResponse](order)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (o *orderService) mapOrders(orders []*models.Order) ([]*dto.OrderResponse, error) {
	const op = "orderService.mapOrders"

	response, err := mapper.MapList[*models.Order, *dto.OrderResponse](orders)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (o *orderService) reserveItems(ctx context.Context, items []models.OrderItem) error {
	const op = "orderService.reserveItems"

	err := o.applyOnProducts(ctx, items, "reserve",
		func(p *models.Product, qty int) error {
			return p.Reserve(qty)
		},
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) releaseItems(ctx context.Context, items []models.OrderItem) error {
	const op = "orderService.releaseItems"

	err := o.applyOnProducts(ctx, items, "release",
		func(p *models.Product, qty int) error {
			return p.Release(qty)
		},
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) deductItems(ctx context.Context, items []models.OrderItem) error {
	const op = "orderService.deductItems"

	err := o.applyOnProducts(ctx, items, "deduct",
		func(p *models.Product, qty int) error {
			return p.Deduct(qty)
		},
	)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (o *orderService) applyOnProducts(
	ctx context.Context,
	items []models.OrderItem,
	actionName string,
	action func(p *models.Product, qty int) error,
) error {
	const op = "orderService.applyOnProducts"

	productIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	products, err := o.productRepo.GetByIDsForUpdate(ctx, productIDs)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	productMap := make(map[uuid.UUID]*models.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	var unavailable []apperror.UnavailableItem

	for _, item := range items {
		product := productMap[item.ProductID]
		if product == nil {
			unavailable = append(unavailable, apperror.UnavailableItem{
				ProductID:    item.ProductID,
				RequestedQty: item.Quantity,
				Action:       actionName,
				Reason:       "PRODUCT_NOT_FOUND",
			})
			continue
		}

		if err := action(product, item.Quantity); err != nil {
			unavailable = append(unavailable, apperror.UnavailableItem{
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
		err := &apperror.ItemsUnavailableError{
			Items: unavailable,
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	if err := o.productRepo.BulkUpsert(ctx, products); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
