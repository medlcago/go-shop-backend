package service

import (
	"context"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/logger"
	"go-shop-backend/pkg/mapper"
	"go-shop-backend/pkg/paymentprovider"
	"go-shop-backend/pkg/utils"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderQuery interface {
	GetByID(ctx context.Context, id uuid.UUID, preload bool) (*models.Order, error)
	GetByPayment(ctx context.Context, providerName string, paymentID string, preload bool) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
}

type OrderStatusUpdater interface {
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status models.OrderStatus) error
}

type paymentService struct {
	provider              paymentprovider.Provider
	orderQuery            OrderQuery
	orderStatusUpdater    OrderStatusUpdater
	userPaymentMethodRepo repository.UserPaymentMethodRepository
	txManager             database.TxManager
	logger                *slog.Logger
}

func NewPaymentService(
	provider paymentprovider.Provider,
	orderQuery OrderQuery,
	orderStatusUpdater OrderStatusUpdater,
	userPaymentMethodRepo repository.UserPaymentMethodRepository,
	txManager database.TxManager,
	logger *slog.Logger,
) *paymentService {
	return &paymentService{
		provider:              provider,
		orderQuery:            orderQuery,
		orderStatusUpdater:    orderStatusUpdater,
		userPaymentMethodRepo: userPaymentMethodRepo,
		txManager:             txManager,
		logger:                logger,
	}
}

func (p *paymentService) CreatePayment(ctx context.Context, userID uuid.UUID, req dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	const op = "paymentService.CreatePayment"

	order, err := p.getOrderForPayment(ctx, req.OrderID, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	providerName := p.provider.GetName()

	createPaymentRequest, err := p.createPaymentRequest(ctx, order, userID, req, providerName)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	idempotencyKey := order.ID.String()

	payment, err := p.provider.CreatePayment(ctx, createPaymentRequest, idempotencyKey)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	order.SetPaymentInfo(payment.ID, providerName)
	if err := p.orderQuery.Update(ctx, order); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.PaymentResponse{
		Status:            string(payment.Status),
		ConfirmationURL:   payment.ConfirmationURL,
		ConfirmationToken: payment.ConfirmationToken,
	}, nil
}

func (p *paymentService) HandleWebhook(ctx context.Context, body []byte) error {
	const op = "paymentService.HandleWebhook"

	event, err := p.provider.ParseWebhook(body)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	err = p.txManager.Wrap(ctx, func(ctx context.Context) error {
		order, err := p.orderQuery.GetByPayment(ctx, p.provider.GetName(), event.PaymentID, true)
		switch {
		case repository.IsRecordNotFound(err):
			return nil
		case err != nil:
			return err
		}

		if order.Status != models.OrderStatusPending {
			return nil
		}

		switch event.Status {
		case paymentprovider.PaymentStatusSucceeded:
			if event.PaymentMethod.Saved {
				if err := p.savePaymentMethod(ctx, event.Metadata.UserID, event.PaymentMethod); err != nil {
					p.logger.ErrorContext(
						ctx,
						"failed to save user payment method",
						logger.Err(err),
						logger.Op(op),
					)
				}
			}

			return p.orderStatusUpdater.UpdateOrderStatus(ctx, order.ID, models.OrderStatusPaid)
		case paymentprovider.PaymentStatusCanceled:
			if event.CancellationDetails != nil {
				p.logger.InfoContext(
					ctx,
					"payment canceled with cancellation details",
					slog.String("party", event.CancellationDetails.Party),
					slog.String("reason", event.CancellationDetails.Reason),
					logger.Op(op),
				)
			}

			return p.orderStatusUpdater.UpdateOrderStatus(ctx, order.ID, models.OrderStatusCanceled)
		default:
			return nil
		}
	})

	if err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (p *paymentService) GetUserPaymentMethods(ctx context.Context, userID uuid.UUID) ([]*dto.UserPaymentMethodResponse, int64, error) {
	const op = "paymentService.GetUserPaymentMethods"

	userPaymentMethods, total, err := p.userPaymentMethodRepo.GetListByUser(ctx, userID)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	response, err := p.mapUserPaymentMethods(userPaymentMethods)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	return response, total, nil
}

func (p *paymentService) SetDefaultPaymentMethod(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	const op = "paymentService.SetDefaultPaymentMethod"

	switch err := p.userPaymentMethodRepo.SetDefault(ctx, id, userID); {
	case repository.IsRecordNotFound(err):
		return apperror.Wrap(op, apperror.ErrPaymentMethodNotFound)
	case err != nil:
		return apperror.Wrap(op, err)
	default:
		return nil
	}
}

func (p *paymentService) DeletePaymentMethod(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	const op = "paymentService.DeletePaymentMethod"

	switch err := p.userPaymentMethodRepo.Delete(ctx, id, userID); {
	case repository.IsRecordNotFound(err):
		return apperror.Wrap(op, apperror.ErrPaymentMethodNotFound)
	case err != nil:
		return apperror.Wrap(op, err)
	default:
		return nil
	}
}

func (p *paymentService) getOrderForPayment(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (*models.Order, error) {
	const op = "paymentService.getOrderForPayment"

	order, err := p.orderQuery.GetByID(ctx, orderID, false)
	switch {
	case repository.IsRecordNotFound(err):
		return nil, apperror.Wrap(op, apperror.ErrOrderNotFound)
	case err != nil:
		return nil, apperror.Wrap(op, err)
	}

	if !order.IsOwnedBy(userID) {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	if order.Status != models.OrderStatusPending {
		return nil, apperror.Wrap(op, apperror.ErrInvalidOrderStatus)
	}

	if order.ExpiresAt != nil && order.ExpiresAt.Before(time.Now().UTC()) {
		return nil, apperror.Wrap(op, apperror.ErrOrderExpired)
	}

	if order.PaymentID != nil && order.ProviderName != nil {
		return nil, apperror.Wrap(op, apperror.ErrPaymentAlreadyCreated)
	}

	return order, nil
}

func (p *paymentService) createPaymentRequest(
	ctx context.Context,
	order *models.Order,
	userID uuid.UUID,
	req dto.CreatePaymentRequest,
	providerName string,
) (*paymentprovider.CreatePaymentRequest, error) {
	const op = "paymentService.createPaymentRequest"

	rubles := decimal.NewFromInt(order.TotalAmount).Div(decimal.NewFromInt(100))

	createPaymentRequest := &paymentprovider.CreatePaymentRequest{
		Amount: paymentprovider.Amount{
			Value:    rubles.String(),
			Currency: paymentprovider.CurrencyRUB,
		},
		Metadata: paymentprovider.Metadata{
			UserID:  userID,
			OrderID: order.ID,
		},
		Capture:     true,
		Description: fmt.Sprintf("Оплата заказа № %s", order.ID),
	}

	if req.PaymentMethodID != uuid.Nil {
		userPaymentMethod, err := p.getUserPaymentMethod(ctx, userID, req.PaymentMethodID, providerName)
		if err != nil {
			return nil, apperror.Wrap(op, err)
		}

		createPaymentRequest.PaymentMethodID = userPaymentMethod.ProviderPaymentMethodID
	} else {
		createPaymentRequest.PaymentType = paymentprovider.PaymentType(req.PaymentType)
		createPaymentRequest.SavePaymentMethod = req.SavePaymentMethod
	}

	return createPaymentRequest, nil
}

func (p *paymentService) getUserPaymentMethod(
	ctx context.Context,
	userID uuid.UUID,
	paymentMethodID uuid.UUID,
	providerName string,
) (*models.UserPaymentMethod, error) {
	const op = "paymentService.getUserPaymentMethod"

	userPaymentMethod, err := p.userPaymentMethodRepo.GetByID(ctx, paymentMethodID)
	switch {
	case repository.IsRecordNotFound(err):
		return nil, apperror.Wrap(op, apperror.ErrPaymentMethodNotFound)
	case err != nil:
		return nil, apperror.Wrap(op, err)
	}

	if userPaymentMethod.UserID != userID {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	if userPaymentMethod.Provider != providerName {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	return userPaymentMethod, nil
}

func (p *paymentService) savePaymentMethod(
	ctx context.Context,
	userID uuid.UUID,
	method paymentprovider.PaymentMethod,
) error {
	const op = "paymentService.savePaymentMethod"

	userPaymentMethod := &models.UserPaymentMethod{
		UserID:                  userID,
		Title:                   method.Title,
		Provider:                p.provider.GetName(),
		ProviderPaymentMethodID: method.ID,
		Type:                    method.Type,
		Brand:                   utils.StringPtr(method.Card.CardType),
		Last4:                   utils.StringPtr(method.Card.Last4),
		ExpMonth:                utils.ParseInt16Ptr(method.Card.ExpiryMonth),
		ExpYear:                 utils.ParseInt16Ptr(method.Card.ExpiryYear),
	}

	if err := p.userPaymentMethodRepo.Upsert(ctx, userPaymentMethod); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (p *paymentService) mapUserPaymentMethods(methods []*models.UserPaymentMethod) ([]*dto.UserPaymentMethodResponse, error) {
	const op = "paymentService.mapUserPaymentMethods"

	response, err := mapper.MapList[*models.UserPaymentMethod, *dto.UserPaymentMethodResponse](methods)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}
