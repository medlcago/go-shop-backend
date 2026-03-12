package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/paymentprovider"
	paymentproviderMocks "go-shop-backend/pkg/paymentprovider/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type txManager struct {
}

func (t txManager) Wrap(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type OrderServiceTestSuite struct {
	suite.Suite
	orderRepo       *repoMocks.MockOrderRepository
	orderItemRepo   *repoMocks.MockOrderItemRepository
	productRepo     *repoMocks.MockProductRepository
	txManager       *txManager
	paymentProvider *paymentproviderMocks.MockProvider
	orderService    *orderService

	ctx          context.Context
	userID       *uuid.UUID
	sessionID    uuid.UUID
	orderID      uuid.UUID
	productID    uuid.UUID
	itemID       uuid.UUID
	paymentID    string
	providerName string
}

func (suite *OrderServiceTestSuite) SetupTest() {
	suite.orderRepo = repoMocks.NewMockOrderRepository(suite.T())
	suite.orderItemRepo = repoMocks.NewMockOrderItemRepository(suite.T())
	suite.productRepo = repoMocks.NewMockProductRepository(suite.T())
	suite.txManager = new(txManager)
	suite.paymentProvider = paymentproviderMocks.NewMockProvider(suite.T())
	suite.orderService = NewOrderService(suite.orderRepo, suite.orderItemRepo, suite.productRepo, suite.paymentProvider, suite.txManager)

	suite.ctx = context.Background()
	suite.userID = new(uuid.New())
	suite.sessionID = uuid.New()
	suite.orderID = uuid.New()
	suite.productID = uuid.New()
	suite.itemID = uuid.New()
	suite.paymentID = uuid.NewString()
	suite.providerName = "test provider"
}

func TestOrderServiceTestSuite(t *testing.T) {
	suite.Run(t, new(OrderServiceTestSuite))
}

// ==================== CreateOrder Tests ====================

func (suite *OrderServiceTestSuite) TestCreateOrder_Success() {
	order := &models.Order{
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	suite.orderRepo.EXPECT().Create(suite.ctx, order).
		Return(nil).Once()

	response, err := suite.orderService.CreateOrder(suite.ctx, suite.userID, suite.sessionID)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(order.ID, response.ID)
	suite.Equal(models.OrderStatusDraft, models.OrderStatus(response.Status))
}

func (suite *OrderServiceTestSuite) TestCreateOrder_WithNilUserID() {
	suite.orderRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(order *models.Order) bool {
		return order.UserID == nil && order.SessionID == suite.sessionID
	})).
		Return(nil).Once()

	response, err := suite.orderService.CreateOrder(suite.ctx, nil, suite.sessionID)

	suite.NoError(err)
	suite.NotNil(response)
}

func (suite *OrderServiceTestSuite) TestCreateOrder_RepositoryError() {
	dbError := errors.New("db error")
	suite.orderRepo.EXPECT().Create(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(dbError).Once()

	response, err := suite.orderService.CreateOrder(suite.ctx, suite.userID, suite.sessionID)

	suite.Nil(response)
	suite.ErrorContains(err, dbError.Error())
}

// ==================== GetOrder Tests ====================

func (suite *OrderServiceTestSuite) TestGetOrder_Success() {
	order := &models.Order{ID: suite.orderID}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	response, err := suite.orderService.GetOrder(suite.ctx, suite.userID, suite.sessionID, suite.orderID)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(order.ID, response.ID)
}

func (suite *OrderServiceTestSuite) TestGetOrder_NotFound() {
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.GetOrder(suite.ctx, suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestGetOrder_RepositoryError() {
	dbError := errors.New("db error")
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(nil, dbError).Once()

	response, err := suite.orderService.GetOrder(suite.ctx, suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorContains(err, dbError.Error())
}

// ==================== GetOrders Tests ====================

func (suite *OrderServiceTestSuite) TestGetOrders_Success() {
	orders := []*models.Order{
		{
			ID:        suite.orderID,
			UserID:    suite.userID,
			SessionID: suite.sessionID,
			Status:    models.OrderStatusCanceled,
		},
		{
			ID:        uuid.New(),
			UserID:    suite.userID,
			SessionID: suite.sessionID,
			Status:    models.OrderStatusCanceled,
		},
	}
	req := dto.ListOrderRequest{Limit: 10, Offset: 0, Status: string(models.OrderStatusCanceled)}

	suite.orderRepo.EXPECT().GetListByOwner(suite.ctx, suite.userID, suite.sessionID, req).
		Return(orders, int64(5), nil).Once()

	response, total, err := suite.orderService.GetOrders(suite.ctx, suite.userID, suite.sessionID, req)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(int64(5), total)
	suite.Len(response, 2)
}

func (suite *OrderServiceTestSuite) TestGetOrders_Empty() {
	req := dto.ListOrderRequest{Limit: 10, Offset: 0}

	suite.orderRepo.EXPECT().GetListByOwner(suite.ctx, suite.userID, suite.sessionID, req).
		Return([]*models.Order{}, int64(0), nil).Once()

	response, total, err := suite.orderService.GetOrders(suite.ctx, suite.userID, suite.sessionID, req)

	suite.NoError(err)
	suite.Empty(response)
	suite.Equal(int64(0), total)
}

func (suite *OrderServiceTestSuite) TestGetOrders_RepositoryError() {
	req := dto.ListOrderRequest{}

	dbError := errors.New("db error")
	suite.orderRepo.EXPECT().GetListByOwner(suite.ctx, suite.userID, suite.sessionID, req).
		Return(nil, int64(0), dbError).Once()

	response, total, err := suite.orderService.GetOrders(suite.ctx, suite.userID, suite.sessionID, req)

	suite.Nil(response)
	suite.ErrorContains(err, dbError.Error())
	suite.Equal(int64(0), total)
}

// ==================== AddItem Tests ====================

func (suite *OrderServiceTestSuite) TestAddItem_Success() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
		Items:     []models.OrderItem{},
	}

	product := &models.Product{
		ID:       suite.productID,
		Name:     "Test Product",
		Price:    1000,
		Stock:    10,
		Reserved: 0,
		IsActive: true,
	}

	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	orderWithItems := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 2000,
		Items: []models.OrderItem{
			{
				ID:          suite.itemID,
				OrderID:     suite.orderID,
				ProductID:   suite.productID,
				ProductName: "Test Product",
				Quantity:    2,
				UnitPrice:   1000,
			},
		},
	}

	// First call - for checking order status (preload=false)
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	suite.orderItemRepo.EXPECT().Upsert(suite.ctx, mock.AnythingOfType("*models.OrderItem")).
		Return(nil).Once()

	// Second call - for recalculation (preload=true)
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(orderWithItems, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(nil).Once()

	response, err := suite.orderService.AddItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(suite.orderID, response.ID)
}

func (suite *OrderServiceTestSuite) TestAddItem_InvalidQuantity() {
	tests := []struct {
		name string
		req  dto.AddOrderItemRequest
		err  error
	}{
		{
			name: "zero quantity",
			req: dto.AddOrderItemRequest{
				ProductID: suite.productID,
				Quantity:  0,
			},
			err: apperrors.ErrInvalidQuantity,
		},
		{
			name: "negative quantity",
			req: dto.AddOrderItemRequest{
				ProductID: suite.productID,
				Quantity:  -10,
			},
			err: apperrors.ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			response, err := suite.orderService.AddItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, tt.req)

			suite.Nil(response)
			suite.ErrorIs(err, tt.err)
		})
	}
}

func (suite *OrderServiceTestSuite) TestAddItem_OrderNotFound() {
	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.AddItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestAddItem_InvalidOrderStatus() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusCompleted,
	}

	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	response, err := suite.orderService.AddItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestAddItem_ProductNotFound() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.AddItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrProductNotFound)
}

func (suite *OrderServiceTestSuite) TestAddItem_ProductNotActive() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	product := &models.Product{
		ID:       suite.productID,
		Name:     "Test Product",
		Price:    1000,
		Stock:    10,
		IsActive: false,
	}

	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	response, err := suite.orderService.AddItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrProductNotActive)
}

func (suite *OrderServiceTestSuite) TestAddItem_InsufficientStock() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	product := &models.Product{
		ID:       suite.productID,
		Name:     "Test Product",
		Price:    1000,
		Stock:    5,
		Reserved: 3,
		IsActive: true,
	}

	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  10,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	response, err := suite.orderService.AddItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrInsufficientStock)
}

func (suite *OrderServiceTestSuite) TestAddItem_RepositoryError() {
	dbError := errors.New("db error")
	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(nil, dbError).Once()

	response, err := suite.orderService.AddItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorContains(err, dbError.Error())
}

// ==================== DeleteItem  Tests ====================

func (suite *OrderServiceTestSuite) TestDeleteItem_Success() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 2000,
		Items: []models.OrderItem{
			{
				ID:        suite.itemID,
				ProductID: suite.productID,
				UnitPrice: 1000,
				Quantity:  2,
			},
		},
	}

	orderAfterDelete := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 0,
		Items:       []models.OrderItem{},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.orderItemRepo.EXPECT().GetItem(suite.ctx, suite.productID, suite.orderID).
		Return(&models.OrderItem{ID: suite.itemID, ProductID: suite.productID}, nil).Once()

	suite.orderItemRepo.EXPECT().DeleteItem(suite.ctx, suite.orderID, suite.productID).
		Return(nil).Once()

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(orderAfterDelete, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(nil).Once()

	response, err := suite.orderService.DeleteItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, suite.productID)

	suite.NoError(err)
	suite.NotNil(response)
}

func (suite *OrderServiceTestSuite) TestDeleteItem_OrderNotFound() {
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.DeleteItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, suite.productID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestDeleteItem_InvalidOrderStatus() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusCompleted,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	response, err := suite.orderService.DeleteItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, suite.productID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestDeleteItem_ItemNotFound() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 2000,
		Items: []models.OrderItem{
			{
				ID:        uuid.New(),
				ProductID: suite.productID,
				UnitPrice: 1000,
				Quantity:  2,
			},
		},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.orderItemRepo.EXPECT().GetItem(suite.ctx, suite.productID, suite.orderID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.DeleteItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, suite.productID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrItemNotFound)
}

func (suite *OrderServiceTestSuite) TestDeleteItem_RepositoryError() {
	dbError := errors.New("db error")
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(nil, dbError).Once()

	response, err := suite.orderService.DeleteItem(suite.ctx, suite.userID, suite.sessionID, suite.orderID, suite.productID)

	suite.Nil(response)
	suite.ErrorContains(err, dbError.Error())
}

// ==================== Clear Tests ====================

func (suite *OrderServiceTestSuite) TestClear_Success() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 2000,
		Items: []models.OrderItem{
			{ID: suite.itemID, Quantity: 2, UnitPrice: 1000},
		},
	}

	clearedOrder := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 0,
		Items:       []models.OrderItem{},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.orderItemRepo.EXPECT().Clear(suite.ctx, suite.orderID).
		Return(nil).Once()

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(clearedOrder, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(nil).Once()

	response, err := suite.orderService.Clear(suite.ctx, suite.userID, suite.sessionID, suite.orderID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(int64(0), response.TotalAmount)
	suite.Len(response.Items, 0)
}

func (suite *OrderServiceTestSuite) TestClear_OrderNotFound() {
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.Clear(suite.ctx, suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestClear_InvalidOrderStatus() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusPending,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	response, err := suite.orderService.Clear(suite.ctx, suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestClear_RepositoryError() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	dbError := errors.New("database error")

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.orderItemRepo.EXPECT().Clear(suite.ctx, suite.orderID).
		Return(dbError).Once()

	result, err := suite.orderService.Clear(suite.ctx, suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(result)
	suite.ErrorContains(err, dbError.Error())
}

// ==================== CalculateTotal Tests ====================

func (suite *OrderServiceTestSuite) TestCalculateTotal() {
	tests := []struct {
		name  string
		order *models.Order
		total int64
	}{
		{
			name: "single item",
			order: &models.Order{
				Items: []models.OrderItem{
					{Quantity: 2, UnitPrice: 1000},
				},
			},
			total: 2000,
		},
		{
			name: "multiple items",
			order: &models.Order{
				Items: []models.OrderItem{
					{Quantity: 2, UnitPrice: 1000},
					{Quantity: 3, UnitPrice: 500},
					{Quantity: 1, UnitPrice: 2500},
				},
			},
			total: 6000,
		},
		{
			name: "empty order",
			order: &models.Order{
				Items: []models.OrderItem{},
			},
			total: 0,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			total := suite.orderService.calculateTotal(tt.order)

			suite.Equal(tt.total, total)
		})
	}
}

// ==================== Checkout Tests ====================

func (suite *OrderServiceTestSuite) TestCheckout_Success() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 12_000,
		Items: []models.OrderItem{
			{
				ID:        suite.itemID,
				ProductID: suite.productID,
				UnitPrice: 2000,
				Quantity:  5,
			},
			{
				ID:        uuid.New(),
				ProductID: uuid.New(),
				UnitPrice: 1000,
				Quantity:  2,
			},
		},
	}

	confirmationURL := "https://test.com"

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 0, Stock: 100, IsActive: true},
			{ID: order.Items[1].ProductID, Reserved: 0, Stock: 50, IsActive: true},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.paymentProvider.EXPECT().CreatePayment(mock.MatchedBy(func(req *paymentprovider.CreatePaymentRequest) bool {
		return req.Metadata.OrderID == suite.orderID &&
			req.Metadata.UserID == *suite.userID &&
			req.Amount > 0 &&
			req.Amount == order.TotalAmount
	})).
		Return(&paymentprovider.Payment{
			ID:              suite.paymentID,
			ConfirmationURL: confirmationURL,
		}, nil).Once()

	suite.paymentProvider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Run(func(ctx context.Context, o *models.Order) {
			suite.Equal(*suite.userID, *o.UserID)
			suite.Equal(order.TotalAmount, o.TotalAmount)
			suite.Equal(models.OrderStatusPending, o.Status)
			suite.Equal(suite.paymentID, *o.PaymentID)
			suite.Equal(suite.providerName, *o.ProviderName)
			suite.NotNil(o.ExpiresAt)
		}).Return(nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, suite.orderID)

	suite.NoError(err)
	suite.NotNil(response)

	suite.Equal(suite.orderID, response.OrderID)
	suite.Equal(confirmationURL, response.ConfirmationURL)
}

func (suite *OrderServiceTestSuite) TestCheckout_LinkUser() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      nil,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 4000,
		Items: []models.OrderItem{
			{
				ID:        suite.itemID,
				ProductID: suite.productID,
				UnitPrice: 1000,
				Quantity:  4,
			},
		},
	}

	confirmationURL := "https://test.com"

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 0, Stock: 100, IsActive: true},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.paymentProvider.EXPECT().CreatePayment(mock.Anything).
		Return(&paymentprovider.Payment{
			ID:              suite.paymentID,
			ConfirmationURL: confirmationURL,
		}, nil).Once()

	suite.paymentProvider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.MatchedBy(func(o *models.Order) bool {
		return o.UserID != nil && *o.UserID == *suite.userID
	})).Return(nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, suite.orderID)

	suite.NoError(err)
	suite.NotNil(response)

	suite.NotNil(order.UserID)
	suite.Equal(*suite.userID, *order.UserID)
	suite.Equal(order.ID, response.OrderID)
	suite.Equal(confirmationURL, response.ConfirmationURL)
}

func (suite *OrderServiceTestSuite) TestCheckout_InvalidOrderStatus() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusPending,
		TotalAmount: 4000,
		Items: []models.OrderItem{
			{
				ID:        suite.itemID,
				ProductID: suite.productID,
				UnitPrice: 1000,
				Quantity:  4,
			},
		},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestCheckout_EmptyOrder() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 4000,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrEmptyOrder)
}

func (suite *OrderServiceTestSuite) TestCheckout_InsufficientStock() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 4000,
		Items: []models.OrderItem{
			{
				ID:        suite.itemID,
				ProductID: suite.productID,
				UnitPrice: 1000,
				Quantity:  10,
			},
		},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 3, Stock: 8, IsActive: true},
		}, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, order.ID)

	suite.Nil(err)
	suite.NotNil(response)

	suite.Equal(order.ID, response.OrderID)
	suite.Len(response.UnavailableItems, 1)
	suite.Equal(suite.productID, response.UnavailableItems[0].ProductID)
	suite.Equal(10, response.UnavailableItems[0].RequestedQty)
	suite.Equal(5, response.UnavailableItems[0].AvailableQty)
	suite.Equal("reserve", response.UnavailableItems[0].Action)
	suite.Equal(apperrors.ErrInsufficientStock.Error(), response.UnavailableItems[0].Reason)
}

func (suite *OrderServiceTestSuite) TestCheckout_ProductNotActive() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 4000,
		Items: []models.OrderItem{
			{
				ID:        suite.itemID,
				ProductID: suite.productID,
				UnitPrice: 1000,
				Quantity:  10,
			},
		},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil)

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 3, Stock: 8, IsActive: false},
		}, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, order.ID)

	suite.Nil(err)
	suite.NotNil(response)

	suite.Equal(order.ID, response.OrderID)
	suite.Len(response.UnavailableItems, 1)
	suite.Equal(suite.productID, response.UnavailableItems[0].ProductID)
	suite.Equal(10, response.UnavailableItems[0].RequestedQty)
	suite.Equal(5, response.UnavailableItems[0].AvailableQty)
	suite.Equal("reserve", response.UnavailableItems[0].Action)
	suite.Equal(apperrors.ErrProductNotActive.Error(), response.UnavailableItems[0].Reason)
}

func (suite *OrderServiceTestSuite) TestCheckout_InsufficientStock_And_ProductNotActive() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 4000,
		Items: []models.OrderItem{
			{
				ID:        suite.itemID,
				ProductID: suite.productID,
				UnitPrice: 1000,
				Quantity:  10,
			}, {
				ID:        uuid.New(),
				ProductID: uuid.New(),
				UnitPrice: 1000,
				Quantity:  10,
			},
		},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil)

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 3, Stock: 8, IsActive: true},
			{ID: order.Items[1].ProductID, Reserved: 0, Stock: 100, IsActive: false},
		}, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, order.ID)

	suite.Nil(err)
	suite.NotNil(response)

	suite.Equal(order.ID, response.OrderID)
	suite.Len(response.UnavailableItems, 2)
	suite.Equal(order.Items[0].ProductID, response.UnavailableItems[0].ProductID)
	suite.Equal(order.Items[1].ProductID, response.UnavailableItems[1].ProductID)
	suite.Equal(apperrors.ErrInsufficientStock.Error(), response.UnavailableItems[0].Reason)
	suite.Equal(apperrors.ErrProductNotActive.Error(), response.UnavailableItems[1].Reason)
}

func (suite *OrderServiceTestSuite) TestCheckout_Forbidden() {
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, apperrors.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestCheckout_InternalError() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 4000,
		Items: []models.OrderItem{
			{
				ID:        suite.itemID,
				ProductID: suite.productID,
				UnitPrice: 1000,
				Quantity:  10,
			},
		},
	}

	dbError := errors.New("db error")

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 0, Stock: 1111, IsActive: true},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, mock.Anything).
		Return(dbError).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)
	suite.ErrorContains(err, dbError.Error())
}
