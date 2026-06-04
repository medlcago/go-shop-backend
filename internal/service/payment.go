package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/paymentprovider"
	"time"

	"github.com/google/uuid"
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
	provider           paymentprovider.Provider
	orderQuery         OrderQuery
	orderStatusUpdater OrderStatusUpdater
	txManager          database.TxManager
}

func NewPaymentService(
	provider paymentprovider.Provider,
	orderQuery OrderQuery,
	orderStatusUpdater OrderStatusUpdater,
	txManager database.TxManager,
) *paymentService {
	return &paymentService{
		provider:           provider,
		orderQuery:         orderQuery,
		orderStatusUpdater: orderStatusUpdater,
		txManager:          txManager,
	}
}

func (p *paymentService) CreatePayment(ctx context.Context, userID uuid.UUID, req dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
	const op = "paymentService.CreatePayment"

	order, err := p.orderQuery.GetByID(ctx, req.OrderID, false)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrOrderNotFound)
		}

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

	payment, err := p.provider.CreatePayment(ctx, &paymentprovider.CreatePaymentRequest{
		Amount: order.TotalAmount,
		Metadata: paymentprovider.Metadata{
			UserID:  userID,
			OrderID: req.OrderID,
		},
	})

	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	order.SetPaymentInfo(payment.ID, p.provider.GetName())
	if err := p.orderQuery.Update(ctx, order); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.PaymentResponse{
		ConfirmationURL: payment.ConfirmationURL,
	}, nil
}

func (p *paymentService) HandleWebhook(ctx context.Context, body []byte) error {
	const op = "paymentService.HandleWebhook"

	event, err := p.provider.ParseWebhook(body)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	err = p.txManager.Wrap(ctx, func(ctx context.Context) error {
		var err error

		order, err := p.orderQuery.GetByPayment(ctx, p.provider.GetName(), event.PaymentID, true)
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
			err = p.orderStatusUpdater.UpdateOrderStatus(ctx, order.ID, models.OrderStatusPaid)
		case paymentprovider.PaymentStatusCanceled:
			err = p.orderStatusUpdater.UpdateOrderStatus(ctx, order.ID, models.OrderStatusCanceled)
		default:
			return nil
		}

		return err
	})

	if err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}
