package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/paymentprovider"
	"testing"
	"time"

	serviceMocks "go-shop-backend/internal/service/mocks"
	"go-shop-backend/pkg/database"
	paymentproviderMocks "go-shop-backend/pkg/paymentprovider/mocks"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type PaymentServiceTestSuite struct {
	suite.Suite
	provider           *paymentproviderMocks.MockProvider
	orderQuery         *serviceMocks.MockOrderQuery
	orderStatusUpdater *serviceMocks.MockOrderStatusUpdater
	txManager          *database.NoopTxManager
	paymentService     *paymentService

	ctx          context.Context
	userID       uuid.UUID
	orderID      uuid.UUID
	providerName string
}

func (suite *PaymentServiceTestSuite) SetupTest() {
	suite.provider = paymentproviderMocks.NewMockProvider(suite.T())
	suite.orderQuery = serviceMocks.NewMockOrderQuery(suite.T())
	suite.orderStatusUpdater = serviceMocks.NewMockOrderStatusUpdater(suite.T())
	suite.txManager = database.NewNoopTxManager()
	suite.paymentService = NewPaymentService(
		suite.provider,
		suite.orderQuery,
		suite.orderStatusUpdater,
		suite.txManager,
	)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
	suite.orderID = uuid.New()
	suite.providerName = "yookassa"
}

func TestPaymentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentServiceTestSuite))
}

// ==================== CreatePayment Tests ====================

func (suite *PaymentServiceTestSuite) TestCreatePayment_Success() {
	req := dto.CreatePaymentRequest{
		OrderID: suite.orderID,
		Type:    "redirect",
	}

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		Status:      models.OrderStatusPending,
		ExpiresAt:   new(time.Now().UTC().Add(10 * time.Minute)),
		TotalAmount: 100_000,
	}

	payment := &paymentprovider.Payment{
		ID:              uuid.NewString(),
		ConfirmationURL: "https://test.com",
	}

	idempotencyKey := order.ID.String()

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	suite.provider.EXPECT().CreatePayment(suite.ctx, &paymentprovider.CreatePaymentRequest{
		Amount: paymentprovider.Amount{
			Value:    decimal.NewFromInt(order.TotalAmount).Div(decimal.NewFromInt(100)).String(),
			Currency: paymentprovider.CurrencyRUB,
		},
		Metadata: paymentprovider.Metadata{
			UserID:  suite.userID,
			OrderID: order.ID,
		},
		Type:    paymentprovider.PaymentType(req.Type),
		Capture: true,
	}, idempotencyKey).Return(payment, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.orderQuery.EXPECT().Update(suite.ctx, order).
		Return(nil).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.NoError(err)
	suite.NotNil(response)

	suite.NotNil(order.PaymentID)
	suite.Equal(payment.ID, *order.PaymentID)
	suite.NotNil(order.ProviderName)
	suite.Equal(suite.providerName, *order.ProviderName)
	suite.Equal(payment.ConfirmationURL, response.ConfirmationURL)
	suite.Equal(payment.ConfirmationToken, response.ConfirmationToken)
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_OrderNotFound() {
	req := dto.CreatePaymentRequest{
		OrderID: suite.orderID,
		Type:    "redirect",
	}

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrOrderNotFound)
	suite.ErrorContains(err, "paymentService.CreatePayment")
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_Forbidden() {
	req := dto.CreatePaymentRequest{
		OrderID: suite.orderID,
		Type:    "redirect",
	}

	order := &models.Order{
		ID:     suite.orderID,
		UserID: new(uuid.New()),
	}

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
	suite.ErrorContains(err, "paymentService.CreatePayment")
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_InvalidOrderStatus() {
	tests := []struct {
		name   string
		status models.OrderStatus
	}{
		{
			name:   "OrderStatusDraft",
			status: models.OrderStatusDraft,
		},
		{
			name:   "OrderStatusPaid",
			status: models.OrderStatusPaid,
		},
		{
			name:   "OrderStatusCanceled",
			status: models.OrderStatusCanceled,
		},
		{
			name:   "OrderStatusCompleted",
			status: models.OrderStatusCompleted,
		},
		{
			name:   "OrderStatusUnknown",
			status: models.OrderStatus("unknown"),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := dto.CreatePaymentRequest{
				OrderID: suite.orderID,
				Type:    "redirect",
			}

			order := &models.Order{
				ID:     suite.orderID,
				UserID: &suite.userID,
				Status: tt.status,
			}

			suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
				Return(order, nil).Once()

			response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

			suite.Nil(response)
			suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
			suite.ErrorContains(err, "paymentService.CreatePayment")
		})
	}

}

func (suite *PaymentServiceTestSuite) TestCreatePayment_OrderExpired() {
	req := dto.CreatePaymentRequest{
		OrderID: suite.orderID,
		Type:    "redirect",
	}

	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
		Status:    models.OrderStatusPending,
		ExpiresAt: new(time.Now().UTC().Add(-10 * time.Minute)),
	}

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrOrderExpired)
	suite.ErrorContains(err, "paymentService.CreatePayment")
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_PaymentAlreadyCreated() {
	req := dto.CreatePaymentRequest{
		OrderID: suite.orderID,
		Type:    "redirect",
	}

	order := &models.Order{
		ID:           suite.orderID,
		UserID:       &suite.userID,
		Status:       models.OrderStatusPending,
		ExpiresAt:    new(time.Now().UTC().Add(10 * time.Minute)),
		PaymentID:    new(uuid.NewString()),
		ProviderName: &suite.providerName,
	}

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrPaymentAlreadyCreated)
	suite.ErrorContains(err, "paymentService.CreatePayment")
}

// ==================== HandleWebhook Tests ====================

func (suite *PaymentServiceTestSuite) TestHandleWebhook_PaymentStatusSucceeded() {
	event := &paymentprovider.WebhookEvent{
		Status:    paymentprovider.PaymentStatusSucceeded,
		PaymentID: uuid.NewString(),
	}

	order := &models.Order{
		ID:           suite.orderID,
		PaymentID:    &event.PaymentID,
		ProviderName: &suite.providerName,
		Status:       models.OrderStatusPending,
	}

	suite.provider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(event, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.orderQuery.EXPECT().GetByPayment(suite.ctx, suite.providerName, event.PaymentID, true).
		Return(order, nil).Once()

	suite.orderStatusUpdater.EXPECT().UpdateOrderStatus(suite.ctx, order.ID, models.OrderStatusPaid).
		Return(nil).Once()

	err := suite.paymentService.HandleWebhook(suite.ctx, []byte("test"))
	suite.NoError(err)
}

func (suite *PaymentServiceTestSuite) TestHandleWebhook_PaymentStatusCanceled() {
	event := &paymentprovider.WebhookEvent{
		Status:    paymentprovider.PaymentStatusCanceled,
		PaymentID: uuid.NewString(),
	}

	order := &models.Order{
		ID:           suite.orderID,
		PaymentID:    &event.PaymentID,
		ProviderName: &suite.providerName,
		Status:       models.OrderStatusPending,
	}

	suite.provider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(event, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.orderQuery.EXPECT().GetByPayment(suite.ctx, suite.providerName, event.PaymentID, true).
		Return(order, nil).Once()

	suite.orderStatusUpdater.EXPECT().UpdateOrderStatus(suite.ctx, order.ID, models.OrderStatusCanceled).
		Return(nil).Once()

	err := suite.paymentService.HandleWebhook(suite.ctx, []byte("test"))
	suite.NoError(err)
}

func (suite *PaymentServiceTestSuite) TestHandleWebhook_OrderNotFound() {
	event := &paymentprovider.WebhookEvent{
		Status:    paymentprovider.PaymentStatusSucceeded,
		PaymentID: uuid.NewString(),
	}

	suite.provider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(event, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.orderQuery.EXPECT().GetByPayment(suite.ctx, suite.providerName, event.PaymentID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	err := suite.paymentService.HandleWebhook(suite.ctx, []byte("test"))
	suite.NoError(err)
}

func (suite *PaymentServiceTestSuite) TestHandleWebhook_OrderStatusNotPending() {
	tests := []struct {
		name   string
		status models.OrderStatus
	}{
		{
			name:   "OrderStatusDraft",
			status: models.OrderStatusDraft,
		},
		{
			name:   "OrderStatusPaid",
			status: models.OrderStatusPaid,
		},
		{
			name:   "OrderStatusCanceled",
			status: models.OrderStatusCanceled,
		},
		{
			name:   "OrderStatusCompleted",
			status: models.OrderStatusCompleted,
		},
		{
			name:   "OrderStatusUnknown",
			status: models.OrderStatus("unknown"),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			event := &paymentprovider.WebhookEvent{
				PaymentID: uuid.NewString(),
			}

			order := &models.Order{
				ID:           suite.orderID,
				PaymentID:    &event.PaymentID,
				ProviderName: &suite.providerName,
				Status:       tt.status,
			}

			suite.provider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
				Return(event, nil).Once()

			suite.provider.EXPECT().GetName().
				Return(suite.providerName).Once()

			suite.orderQuery.EXPECT().GetByPayment(suite.ctx, suite.providerName, event.PaymentID, true).
				Return(order, nil).Once()

			err := suite.paymentService.HandleWebhook(suite.ctx, []byte("test"))
			suite.NoError(err)
		})
	}
}

func (suite *PaymentServiceTestSuite) TestHandleWebhook_ParseWebhookError() {
	parseErr := errors.New("parse webhook error")

	suite.provider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(nil, parseErr).Once()

	err := suite.paymentService.HandleWebhook(suite.ctx, []byte("test"))
	suite.ErrorIs(err, parseErr)
	suite.ErrorContains(err, "paymentService.HandleWebhook")
}
