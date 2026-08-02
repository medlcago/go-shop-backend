package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/paymentprovider"
	"go-shop-backend/pkg/testutils"
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
	userPaymentRepo    *repoMocks.MockUserPaymentMethodRepository
	txManager          *database.NoopTxManager
	paymentService     *paymentService

	ctx             context.Context
	userID          uuid.UUID
	orderID         uuid.UUID
	paymentMethodID uuid.UUID
	providerName    string
}

func (suite *PaymentServiceTestSuite) SetupTest() {
	suite.provider = paymentproviderMocks.NewMockProvider(suite.T())
	suite.orderQuery = serviceMocks.NewMockOrderQuery(suite.T())
	suite.orderStatusUpdater = serviceMocks.NewMockOrderStatusUpdater(suite.T())
	suite.userPaymentRepo = repoMocks.NewMockUserPaymentMethodRepository(suite.T())
	suite.txManager = database.NewNoopTxManager()
	suite.paymentService = NewPaymentService(
		suite.provider,
		suite.orderQuery,
		suite.orderStatusUpdater,
		suite.userPaymentRepo,
		suite.txManager,
		testutils.NewSlogLogger(),
	)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
	suite.orderID = uuid.New()
	suite.paymentMethodID = uuid.New()
	suite.providerName = "yookassa"
}

func TestPaymentServiceTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentServiceTestSuite))
}

// ==================== CreatePayment Tests ====================

func (suite *PaymentServiceTestSuite) TestCreatePayment_Success() {
	req := dto.CreatePaymentRequest{
		OrderID:           suite.orderID,
		PaymentType:       "redirect",
		SavePaymentMethod: true,
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
		Status:          paymentprovider.PaymentStatusPending,
	}

	idempotencyKey := order.ID.String()

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.provider.EXPECT().CreatePayment(suite.ctx, &paymentprovider.CreatePaymentRequest{
		Amount: paymentprovider.Amount{
			Value:    decimal.NewFromInt(order.TotalAmount).Div(decimal.NewFromInt(100)).String(),
			Currency: paymentprovider.CurrencyRUB,
		},
		Metadata: paymentprovider.Metadata{
			UserID:  suite.userID,
			OrderID: order.ID,
		},
		PaymentType:       paymentprovider.PaymentType(req.PaymentType),
		Capture:           true,
		Description:       fmt.Sprintf("Оплата заказа № %s", order.ID),
		SavePaymentMethod: req.SavePaymentMethod,
	}, idempotencyKey).Return(payment, nil).Once()

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
	suite.Equal(payment.ConfirmationToken, response.ConfirmationToken)
	suite.Equal(string(payment.Status), response.Status)
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_UseSavedPaymentMethod_Success() {
	req := dto.CreatePaymentRequest{
		OrderID:         suite.orderID,
		PaymentMethodID: suite.paymentMethodID,
	}

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		Status:      models.OrderStatusPending,
		ExpiresAt:   new(time.Now().UTC().Add(10 * time.Minute)),
		TotalAmount: 100_000,
	}

	userPaymentMethod := &models.UserPaymentMethod{
		ID:                      suite.paymentMethodID,
		UserID:                  suite.userID,
		Provider:                suite.providerName,
		ProviderPaymentMethodID: "pm_provider_123",
	}

	payment := &paymentprovider.Payment{
		ID:              uuid.NewString(),
		ConfirmationURL: "https://test.com",
		Status:          paymentprovider.PaymentStatusPending,
	}

	idempotencyKey := order.ID.String()

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.userPaymentRepo.EXPECT().GetByID(suite.ctx, req.PaymentMethodID).
		Return(userPaymentMethod, nil).Once()

	suite.provider.EXPECT().CreatePayment(suite.ctx, &paymentprovider.CreatePaymentRequest{
		Amount: paymentprovider.Amount{
			Value:    decimal.NewFromInt(order.TotalAmount).Div(decimal.NewFromInt(100)).String(),
			Currency: paymentprovider.CurrencyRUB,
		},
		Metadata: paymentprovider.Metadata{
			UserID:  suite.userID,
			OrderID: order.ID,
		},
		PaymentMethodID: userPaymentMethod.ProviderPaymentMethodID,
		Capture:         true,
		Description:     fmt.Sprintf("Оплата заказа № %s", order.ID),
	}, idempotencyKey).Return(payment, nil).Once()

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
	suite.Equal(string(payment.Status), response.Status)
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_UseSavedPaymentMethod_PaymentMethodNotFound() {
	req := dto.CreatePaymentRequest{
		OrderID:         suite.orderID,
		PaymentMethodID: suite.paymentMethodID,
	}

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		Status:      models.OrderStatusPending,
		ExpiresAt:   new(time.Now().UTC().Add(10 * time.Minute)),
		TotalAmount: 100_000,
	}

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.userPaymentRepo.EXPECT().GetByID(suite.ctx, req.PaymentMethodID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrPaymentMethodNotFound)
	suite.ErrorContains(err, "paymentService.CreatePayment")
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_UseSavedPaymentMethod_UserNotOwner() {
	req := dto.CreatePaymentRequest{
		OrderID:         suite.orderID,
		PaymentMethodID: suite.paymentMethodID,
	}

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		Status:      models.OrderStatusPending,
		ExpiresAt:   new(time.Now().UTC().Add(10 * time.Minute)),
		TotalAmount: 100_000,
	}

	userPaymentMethod := &models.UserPaymentMethod{
		ID:     suite.paymentMethodID,
		UserID: uuid.New(),
	}

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.userPaymentRepo.EXPECT().GetByID(suite.ctx, req.PaymentMethodID).
		Return(userPaymentMethod, nil).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
	suite.ErrorContains(err, "paymentService.CreatePayment")
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_UseSavedPaymentMethod_InvalidProvider() {
	req := dto.CreatePaymentRequest{
		OrderID:         suite.orderID,
		PaymentMethodID: suite.paymentMethodID,
	}

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		Status:      models.OrderStatusPending,
		ExpiresAt:   new(time.Now().UTC().Add(10 * time.Minute)),
		TotalAmount: 100_000,
	}

	userPaymentMethod := &models.UserPaymentMethod{
		ID:       suite.paymentMethodID,
		UserID:   suite.userID,
		Provider: "test123",
	}

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(order, nil).Once()

	suite.userPaymentRepo.EXPECT().GetByID(suite.ctx, req.PaymentMethodID).
		Return(userPaymentMethod, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
	suite.ErrorContains(err, "paymentService.CreatePayment")
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_OrderNotFound() {
	req := dto.CreatePaymentRequest{
		OrderID:     suite.orderID,
		PaymentType: "redirect",
	}

	suite.orderQuery.EXPECT().GetByID(suite.ctx, req.OrderID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.paymentService.CreatePayment(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrOrderNotFound)
	suite.ErrorContains(err, "paymentService.CreatePayment")
}

func (suite *PaymentServiceTestSuite) TestCreatePayment_UserNotOwner() {
	req := dto.CreatePaymentRequest{
		OrderID:     suite.orderID,
		PaymentType: "redirect",
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
				OrderID:     suite.orderID,
				PaymentType: "redirect",
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
		OrderID:     suite.orderID,
		PaymentType: "redirect",
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
		OrderID:     suite.orderID,
		PaymentType: "redirect",
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

func (suite *PaymentServiceTestSuite) TestHandleWebhook_PaymentStatusSucceeded_SavePaymentMethod() {
	event := &paymentprovider.WebhookEvent{
		Status:    paymentprovider.PaymentStatusSucceeded,
		PaymentID: uuid.NewString(),
		Metadata: paymentprovider.Metadata{
			UserID:  suite.userID,
			OrderID: suite.orderID,
		},
		PaymentMethod: paymentprovider.PaymentMethod{
			ID:    uuid.NewString(),
			Title: "Bank card *4444",
			Saved: true,
			Type:  "bank_card",
		},
	}

	order := &models.Order{
		ID:           suite.orderID,
		PaymentID:    &event.PaymentID,
		ProviderName: &suite.providerName,
		Status:       models.OrderStatusPending,
	}

	userPaymentMethod := &models.UserPaymentMethod{
		UserID:                  event.Metadata.UserID,
		Title:                   event.PaymentMethod.Title,
		Provider:                suite.providerName,
		ProviderPaymentMethodID: event.PaymentMethod.ID,
		Type:                    event.PaymentMethod.Type,
	}

	suite.provider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(event, nil).Once()

	suite.provider.EXPECT().GetName().
		Return(suite.providerName).Times(2)

	suite.orderQuery.EXPECT().GetByPayment(suite.ctx, suite.providerName, event.PaymentID, true).
		Return(order, nil).Once()

	suite.userPaymentRepo.EXPECT().Upsert(suite.ctx, userPaymentMethod).
		Return(nil).Once()

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

// ==================== GetUserPaymentMethods Tests ====================

func (suite *PaymentServiceTestSuite) TestGetUserPaymentMethods_Success() {
	userPaymentMethods := []*models.UserPaymentMethod{
		{ID: uuid.New()}, {ID: uuid.New()},
	}

	suite.userPaymentRepo.EXPECT().GetListByUser(suite.ctx, suite.userID).
		Return(userPaymentMethods, 5, nil).Once()

	response, total, err := suite.paymentService.GetUserPaymentMethods(suite.ctx, suite.userID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Len(response, 2)
	suite.Equal(int64(5), total)
}

// ==================== SetDefaultPaymentMethod Tests ====================

func (suite *PaymentServiceTestSuite) TestSetDefaultPaymentMethod_Success() {
	suite.userPaymentRepo.EXPECT().SetDefault(suite.ctx, suite.paymentMethodID, suite.userID).
		Return(nil).Once()

	err := suite.paymentService.SetDefaultPaymentMethod(suite.ctx, suite.paymentMethodID, suite.userID)
	suite.NoError(err)
}

func (suite *PaymentServiceTestSuite) TestSetDefaultPaymentMethod_PaymentMethodNotFound() {
	suite.userPaymentRepo.EXPECT().SetDefault(suite.ctx, suite.paymentMethodID, suite.userID).
		Return(repository.ErrRecordNotFound).Once()

	err := suite.paymentService.SetDefaultPaymentMethod(suite.ctx, suite.paymentMethodID, suite.userID)

	suite.ErrorIs(err, apperror.ErrPaymentMethodNotFound)
	suite.ErrorContains(err, "paymentService.SetDefaultPaymentMethod")
}

// ==================== DeletePaymentMethod Tests ====================

func (suite *PaymentServiceTestSuite) TestDeletePaymentMethod_Success() {
	suite.userPaymentRepo.EXPECT().Delete(suite.ctx, suite.paymentMethodID, suite.userID).
		Return(nil).Once()

	err := suite.paymentService.DeletePaymentMethod(suite.ctx, suite.paymentMethodID, suite.userID)
	suite.NoError(err)
}

func (suite *PaymentServiceTestSuite) TestDeletePaymentMethod_PaymentMethodNotFoundd() {
	suite.userPaymentRepo.EXPECT().Delete(suite.ctx, suite.paymentMethodID, suite.userID).
		Return(repository.ErrRecordNotFound).Once()

	err := suite.paymentService.DeletePaymentMethod(suite.ctx, suite.paymentMethodID, suite.userID)

	suite.ErrorIs(err, apperror.ErrPaymentMethodNotFound)
	suite.ErrorContains(err, "paymentService.DeletePaymentMethod")
}
