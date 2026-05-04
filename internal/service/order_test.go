package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	tasksMocks "go-shop-backend/internal/tasks/mocks"
	uploadMocks "go-shop-backend/internal/upload/mocks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/paymentprovider"
	paymentproviderMocks "go-shop-backend/pkg/paymentprovider/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type OrderServiceTestSuite struct {
	suite.Suite
	orderRepo       *repoMocks.MockOrderRepository
	orderItemRepo   *repoMocks.MockOrderItemRepository
	productRepo     *repoMocks.MockProductRepository
	txManager       *database.NoopTxManager
	paymentProvider *paymentproviderMocks.MockProvider
	orderTask       *tasksMocks.MockOrderTask
	orderService    *orderService

	ctx                  context.Context
	userID               uuid.UUID
	sessionID            uuid.UUID
	orderID              uuid.UUID
	productID            uuid.UUID
	itemID               uuid.UUID
	paymentID            string
	providerName         string
	orderCancelDelay     time.Duration
	orderCheckoutTimeout time.Duration
	uploadManager        *uploadMocks.MockManager
}

func (suite *OrderServiceTestSuite) SetupTest() {
	suite.orderCancelDelay = 10 * time.Minute
	suite.orderCheckoutTimeout = 5 * time.Second

	suite.orderRepo = repoMocks.NewMockOrderRepository(suite.T())
	suite.orderItemRepo = repoMocks.NewMockOrderItemRepository(suite.T())
	suite.productRepo = repoMocks.NewMockProductRepository(suite.T())
	suite.txManager = database.NewNoopTxManager()
	suite.paymentProvider = paymentproviderMocks.NewMockProvider(suite.T())
	suite.orderTask = tasksMocks.NewMockOrderTask(suite.T())
	suite.uploadManager = uploadMocks.NewMockManager(suite.T())
	suite.orderService = NewOrderService(
		suite.orderRepo,
		suite.orderItemRepo,
		suite.productRepo,
		suite.paymentProvider,
		suite.orderTask,
		suite.txManager,
		suite.orderCancelDelay,
		suite.orderCheckoutTimeout,
		suite.uploadManager,
	)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
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
		UserID:    &suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	suite.orderRepo.EXPECT().Create(suite.ctx, order).
		Return(nil).Once()

	response, err := suite.orderService.CreateOrder(suite.ctx, &suite.userID, suite.sessionID)

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

	response, err := suite.orderService.CreateOrder(suite.ctx, &suite.userID, suite.sessionID)

	suite.Nil(response)
	suite.ErrorContains(err, dbError.Error())
}

// ==================== GetOrder Tests ====================

func (suite *OrderServiceTestSuite) TestGetOrder_Success() {
	order := &models.Order{
		ID: suite.orderID,
		Items: []models.OrderItem{
			{
				ID: suite.itemID,
				Product: models.Product{
					ID: suite.productID,
					Images: []*models.Upload{
						{
							ObjectKey: "key1",
						},
						{
							ObjectKey: "key2",
						},
					},
				},
			},
		},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.AnythingOfType("string")).
		Return("key").Times(2)

	response, err := suite.orderService.GetOrder(suite.ctx, &suite.userID, suite.sessionID, suite.orderID)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(order.ID, response.ID)
}

func (suite *OrderServiceTestSuite) TestGetOrder_NotFound() {
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.GetOrder(suite.ctx, &suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestGetOrder_RepositoryError() {
	dbError := errors.New("db error")
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(nil, dbError).Once()

	response, err := suite.orderService.GetOrder(suite.ctx, &suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, dbError)
}

// ==================== GetOrders Tests ====================

func (suite *OrderServiceTestSuite) TestGetOrders_Success() {
	orders := []*models.Order{
		{
			ID:        suite.orderID,
			UserID:    &suite.userID,
			SessionID: suite.sessionID,
			Status:    models.OrderStatusCanceled,
			Items: []models.OrderItem{
				{
					ID: suite.itemID,
					Product: models.Product{
						ID: suite.productID,
						Images: []*models.Upload{
							{
								ObjectKey: "key",
							},
						},
					},
				},
			},
		},
		{
			ID:        uuid.New(),
			UserID:    &suite.userID,
			SessionID: suite.sessionID,
			Status:    models.OrderStatusCanceled,
		},
	}
	req := dto.ListOrderRequest{Limit: 10, Offset: 0, Status: string(models.OrderStatusCanceled)}

	suite.orderRepo.EXPECT().GetListByOwner(suite.ctx, &suite.userID, suite.sessionID, req).
		Return(orders, int64(5), nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.AnythingOfType("string")).
		Return("key").Once()

	response, total, err := suite.orderService.GetOrders(suite.ctx, &suite.userID, suite.sessionID, req)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(int64(5), total)
	suite.Len(response, 2)
	suite.Len(response[0].Items, 1)
	suite.Len(response[1].Items, 0)
}

func (suite *OrderServiceTestSuite) TestGetOrders_Empty() {
	req := dto.ListOrderRequest{Limit: 10, Offset: 0}

	suite.orderRepo.EXPECT().GetListByOwner(suite.ctx, &suite.userID, suite.sessionID, req).
		Return([]*models.Order{}, int64(0), nil).Once()

	response, total, err := suite.orderService.GetOrders(suite.ctx, &suite.userID, suite.sessionID, req)

	suite.NoError(err)
	suite.Empty(response)
	suite.Equal(int64(0), total)
}

func (suite *OrderServiceTestSuite) TestGetOrders_RepositoryError() {
	req := dto.ListOrderRequest{}

	dbError := errors.New("db error")
	suite.orderRepo.EXPECT().GetListByOwner(suite.ctx, &suite.userID, suite.sessionID, req).
		Return(nil, int64(0), dbError).Once()

	response, total, err := suite.orderService.GetOrders(suite.ctx, &suite.userID, suite.sessionID, req)

	suite.Nil(response)
	suite.ErrorContains(err, dbError.Error())
	suite.Equal(int64(0), total)
}

// ==================== AddItem Tests ====================

func (suite *OrderServiceTestSuite) TestAddItem_Success() {
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

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
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
				Product: models.Product{
					ID: suite.productID,
					Images: []*models.Upload{
						{
							ObjectKey: "test1.png",
						},
						{
							ObjectKey: "test2.jpg",
						},
					},
				},
			},
		},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	suite.orderItemRepo.EXPECT().Upsert(suite.ctx, mock.AnythingOfType("*models.OrderItem")).
		Return(nil).Once()

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(order, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.AnythingOfType("string")).
		Return("key").Times(2)

	response, err := suite.orderService.AddItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(suite.orderID, response.ID)
	suite.Len(response.Items, 1)
	suite.Equal(response.Items[0].ProductID, suite.productID)
	suite.Len(response.Items[0].Product.Images, 2)
	suite.NotEmpty(response.Items[0].Product.Images[0].URL)
	suite.NotEmpty(response.Items[0].Product.Images[1].URL)
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
			err: apperror.ErrInvalidQuantity,
		},
		{
			name: "negative quantity",
			req: dto.AddOrderItemRequest{
				ProductID: suite.productID,
				Quantity:  -10,
			},
			err: apperror.ErrInvalidQuantity,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			response, err := suite.orderService.AddItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, tt.req)

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

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.AddItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestAddItem_InvalidOrderStatus() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusCompleted,
	}

	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	response, err := suite.orderService.AddItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestAddItem_ProductNotFound() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.AddItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
}

func (suite *OrderServiceTestSuite) TestAddItem_ProductNotActive() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	response, err := suite.orderService.AddItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrProductNotActive)
}

func (suite *OrderServiceTestSuite) TestAddItem_InsufficientStock() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	response, err := suite.orderService.AddItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInsufficientStock)
}

func (suite *OrderServiceTestSuite) TestAddItem_RepositoryError() {
	dbError := errors.New("db error")
	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  2,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(nil, dbError).Once()

	response, err := suite.orderService.AddItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, dbError)
}

// ==================== RemoveItem Tests ====================

func (suite *OrderServiceTestSuite) TestRemoveItem_Success() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
		Items: []models.OrderItem{
			{
				ID: suite.itemID,
			},
			{
				ID: uuid.New(),
				Product: models.Product{
					Images: []*models.Upload{
						{
							ObjectKey: "test.png",
						},
					},
				},
			},
		},
	}

	orderAfterDelete := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 0,
		Items:       []models.OrderItem{order.Items[1]},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.orderItemRepo.EXPECT().RemoveItem(suite.ctx, suite.orderID, suite.itemID).
		Return(true, nil).Once()

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(orderAfterDelete, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.AnythingOfType("string")).
		Return("key").Once()

	response, err := suite.orderService.RemoveItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, suite.itemID)

	suite.NoError(err)
	suite.NotNil(response)
}

func (suite *OrderServiceTestSuite) TestRemoveItem_OrderNotFound() {
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.RemoveItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, suite.itemID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestRemoveItem_InvalidOrderStatus() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusCompleted,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	response, err := suite.orderService.RemoveItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, suite.itemID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestRemoveItem_ItemNotFound() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.orderItemRepo.EXPECT().RemoveItem(suite.ctx, suite.orderID, suite.itemID).
		Return(false, nil).Once()

	response, err := suite.orderService.RemoveItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, suite.itemID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrItemNotFound)
}

func (suite *OrderServiceTestSuite) TestRemoveItem_RepositoryError() {
	dbError := errors.New("db error")
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(nil, dbError).Once()

	response, err := suite.orderService.RemoveItem(suite.ctx, &suite.userID, suite.sessionID, suite.orderID, suite.itemID)

	suite.Nil(response)
	suite.ErrorIs(err, dbError)
}

// ==================== Clear Tests ====================

func (suite *OrderServiceTestSuite) TestClear_Success() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 2000,
		Items: []models.OrderItem{
			{ID: suite.itemID, Quantity: 2, UnitPrice: 1000},
		},
	}

	clearedOrder := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 0,
		Items:       []models.OrderItem{},
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.orderItemRepo.EXPECT().Clear(suite.ctx, suite.orderID).
		Return(nil).Once()

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(clearedOrder, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(nil).Once()

	response, err := suite.orderService.Clear(suite.ctx, &suite.userID, suite.sessionID, suite.orderID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(int64(0), response.TotalAmount)
	suite.Len(response.Items, 0)
}

func (suite *OrderServiceTestSuite) TestClear_OrderNotFound() {
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.Clear(suite.ctx, &suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestClear_InvalidOrderStatus() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusPending,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	response, err := suite.orderService.Clear(suite.ctx, &suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestClear_RepositoryError() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	dbError := errors.New("database error")

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.orderItemRepo.EXPECT().Clear(suite.ctx, suite.orderID).
		Return(dbError).Once()

	response, err := suite.orderService.Clear(suite.ctx, &suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, dbError)
}

// ==================== Checkout Tests ====================

func (suite *OrderServiceTestSuite) TestCheckout_Success() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(mock.Anything, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 0, Stock: 100, IsActive: true},
			{ID: order.Items[1].ProductID, Reserved: 0, Stock: 50, IsActive: true},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(mock.Anything, mock.Anything).
		Return(nil).Once()

	suite.paymentProvider.EXPECT().CreatePayment(mock.Anything, mock.MatchedBy(func(req *paymentprovider.CreatePaymentRequest) bool {
		return req.Metadata.OrderID == suite.orderID &&
			req.Metadata.UserID == *&suite.userID &&
			req.Amount > 0 &&
			req.Amount == order.TotalAmount
	})).
		Return(&paymentprovider.Payment{
			ID:              suite.paymentID,
			ConfirmationURL: confirmationURL,
		}, nil).Once()

	suite.paymentProvider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.orderRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*models.Order")).
		Run(func(ctx context.Context, o *models.Order) {
			suite.Equal(*&suite.userID, *o.UserID)
			suite.Equal(order.TotalAmount, o.TotalAmount)
			suite.Equal(models.OrderStatusPending, o.Status)
			suite.Equal(suite.paymentID, *o.PaymentID)
			suite.Equal(suite.providerName, *o.ProviderName)
			suite.NotNil(o.ExpiresAt)
		}).Return(nil).Once()

	suite.orderTask.EXPECT().EnqueueCancelOrder(mock.Anything, *&suite.userID, suite.orderID, suite.orderCancelDelay).
		Return(nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, suite.orderID)

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

	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(mock.Anything, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 0, Stock: 100, IsActive: true},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(mock.Anything, mock.Anything).
		Return(nil).Once()

	suite.paymentProvider.EXPECT().CreatePayment(mock.Anything, mock.Anything).
		Return(&paymentprovider.Payment{
			ID:              suite.paymentID,
			ConfirmationURL: confirmationURL,
		}, nil).Once()

	suite.paymentProvider.EXPECT().GetName().
		Return(suite.providerName).Once()

	suite.orderRepo.EXPECT().Update(mock.Anything, mock.MatchedBy(func(o *models.Order) bool {
		return o.UserID != nil && *o.UserID == *&suite.userID
	})).Return(nil).Once()

	suite.orderTask.EXPECT().EnqueueCancelOrder(mock.Anything, *&suite.userID, suite.orderID, suite.orderCancelDelay).
		Return(nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, suite.orderID)

	suite.NoError(err)
	suite.NotNil(response)

	suite.NotNil(order.UserID)
	suite.Equal(*&suite.userID, *order.UserID)
	suite.Equal(order.ID, response.OrderID)
	suite.Equal(confirmationURL, response.ConfirmationURL)
}

func (suite *OrderServiceTestSuite) TestCheckout_InvalidOrderStatus() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestCheckout_EmptyOrder() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 4000,
	}

	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrEmptyOrder)
}

func (suite *OrderServiceTestSuite) TestCheckout_InsufficientStock() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(mock.Anything, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 3, Stock: 8, IsActive: true},
		}, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)

	items, ok := apperror.GetUnavailableItemsFromError(err)
	suite.True(ok)
	suite.Len(items, 1)
	suite.Equal(suite.itemID, items[0].ID)
	suite.Equal(suite.productID, items[0].ProductID)
	suite.Equal(10, items[0].RequestedQty)
	suite.Equal(5, items[0].AvailableQty)
	suite.Equal("reserve", items[0].Action)
	suite.Equal(apperror.ErrInsufficientStock.Error(), items[0].Reason)
}

func (suite *OrderServiceTestSuite) TestCheckout_ProductNotActive() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil)

	suite.productRepo.EXPECT().GetByIDsForUpdate(mock.Anything, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 3, Stock: 8, IsActive: false},
		}, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)

	items, ok := apperror.GetUnavailableItemsFromError(err)
	suite.True(ok)
	suite.Len(items, 1)
	suite.Equal(suite.itemID, items[0].ID)
	suite.Equal(suite.productID, items[0].ProductID)
	suite.Equal(10, items[0].RequestedQty)
	suite.Equal(5, items[0].AvailableQty)
	suite.Equal("reserve", items[0].Action)
	suite.Equal(apperror.ErrProductNotActive.Error(), items[0].Reason)
}

func (suite *OrderServiceTestSuite) TestCheckout_InsufficientStock_And_ProductNotActive() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil)

	suite.productRepo.EXPECT().GetByIDsForUpdate(mock.Anything, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 3, Stock: 8, IsActive: true},
			{ID: order.Items[1].ProductID, Reserved: 0, Stock: 100, IsActive: false},
		}, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)

	items, ok := apperror.GetUnavailableItemsFromError(err)
	suite.True(ok)
	suite.Len(items, 2)
	suite.Equal(order.Items[0].ID, items[0].ID)
	suite.Equal(order.Items[1].ID, items[1].ID)
	suite.Equal(order.Items[0].ProductID, items[0].ProductID)
	suite.Equal(order.Items[1].ProductID, items[1].ProductID)
	suite.Equal(apperror.ErrInsufficientStock.Error(), items[0].Reason)
	suite.Equal(apperror.ErrProductNotActive.Error(), items[1].Reason)
}

func (suite *OrderServiceTestSuite) TestCheckout_Forbidden() {
	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, suite.orderID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *OrderServiceTestSuite) TestCheckout_InternalError() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
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

	suite.orderRepo.EXPECT().GetByOwner(mock.Anything, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(mock.Anything, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 0, Stock: 1111, IsActive: true},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(mock.Anything, mock.Anything).
		Return(dbError).Once()

	response, err := suite.orderService.Checkout(suite.ctx, *&suite.userID, suite.sessionID, order.ID)

	suite.Nil(response)
	suite.ErrorIs(err, dbError)
}

// ==================== HandlePaymentWebhook Tests ====================

func (suite *OrderServiceTestSuite) TestHandlePaymentWebhook_Success_PayOrder() {
	webhookEvent := &paymentprovider.WebhookEvent{
		PaymentID: suite.paymentID,
		Status:    paymentprovider.PaymentStatusSucceeded,
		Metadata: paymentprovider.Metadata{
			UserID:  suite.userID,
			OrderID: suite.orderID,
		},
	}

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusPending,
		TotalAmount: 2500,
		Items: []models.OrderItem{
			{ProductID: uuid.New(), Quantity: 2, UnitPrice: 1000},
			{ProductID: uuid.New(), Quantity: 1, UnitPrice: 500},
		},
	}

	suite.paymentProvider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(webhookEvent, nil).Once()

	suite.paymentProvider.EXPECT().GetName().Return("yookassa").Once()

	suite.orderRepo.EXPECT().GetByPayment(suite.ctx, "yookassa", suite.paymentID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 3, Stock: 8},
			{ID: order.Items[1].ProductID, Reserved: 5, Stock: 100},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, order).Return(nil).Once()

	err := suite.orderService.HandlePaymentWebhook(suite.ctx, []byte("data"))

	suite.NoError(err)
	suite.Equal(models.OrderStatusPaid, order.Status)
}

func (suite *OrderServiceTestSuite) TestHandlePaymentWebhook_Success_CancelOrder() {
	webhookEvent := &paymentprovider.WebhookEvent{
		PaymentID: suite.paymentID,
		Status:    paymentprovider.PaymentStatusCanceled,
		Metadata: paymentprovider.Metadata{
			UserID:  suite.userID,
			OrderID: suite.orderID,
		},
	}

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusPending,
		TotalAmount: 2500,
		Items: []models.OrderItem{
			{ProductID: uuid.New(), Quantity: 2, UnitPrice: 1000},
			{ProductID: uuid.New(), Quantity: 1, UnitPrice: 500},
		},
	}

	suite.paymentProvider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(webhookEvent, nil).Once()

	suite.paymentProvider.EXPECT().GetName().Return("yookassa").Once()

	suite.orderRepo.EXPECT().GetByPayment(suite.ctx, "yookassa", suite.paymentID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 3, Stock: 8},
			{ID: order.Items[1].ProductID, Reserved: 5, Stock: 100},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, order).Return(nil).Once()

	err := suite.orderService.HandlePaymentWebhook(suite.ctx, []byte("data"))

	suite.NoError(err)
	suite.Equal(models.OrderStatusCanceled, order.Status)
}

func (suite *OrderServiceTestSuite) TestHandlePaymentWebhook_OrderNotFound_ReturnsNil() {
	webhookEvent := &paymentprovider.WebhookEvent{
		PaymentID: suite.paymentID,
		Status:    paymentprovider.PaymentStatusSucceeded,
	}

	suite.paymentProvider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(webhookEvent, nil).Once()

	suite.paymentProvider.EXPECT().GetName().Return("yookassa").Once()

	suite.orderRepo.EXPECT().GetByPayment(suite.ctx, "yookassa", suite.paymentID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	err := suite.orderService.HandlePaymentWebhook(suite.ctx, []byte("data"))
	suite.NoError(err)
}

func (suite *OrderServiceTestSuite) TestHandlePaymentWebhook_OrderAlreadyProcessed_ShouldSkip() {
	webhookEvent := &paymentprovider.WebhookEvent{
		PaymentID: suite.paymentID,
		Status:    paymentprovider.PaymentStatusSucceeded,
	}

	order := &models.Order{
		Status: models.OrderStatusCanceled,
	}

	suite.paymentProvider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(webhookEvent, nil).Once()

	suite.paymentProvider.EXPECT().GetName().Return("yookassa").Once()

	suite.orderRepo.EXPECT().GetByPayment(suite.ctx, "yookassa", suite.paymentID, true).
		Return(order, nil).Once()

	err := suite.orderService.HandlePaymentWebhook(suite.ctx, []byte("data"))
	suite.NoError(err)
}

func (suite *OrderServiceTestSuite) TestHandlePaymentWebhook_UnknownEventStatus_ShouldSkip() {
	webhookEvent := &paymentprovider.WebhookEvent{
		PaymentID: suite.paymentID,
		Status:    "unknown_status",
	}

	order := &models.Order{
		Status: models.OrderStatusPending,
	}

	suite.paymentProvider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(webhookEvent, nil).Once()

	suite.paymentProvider.EXPECT().GetName().Return("yookassa").Once()

	suite.orderRepo.EXPECT().GetByPayment(suite.ctx, "yookassa", suite.paymentID, true).
		Return(order, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, order).Return(nil).Once()

	err := suite.orderService.HandlePaymentWebhook(suite.ctx, []byte("data"))
	suite.NoError(err)
}

func (suite *OrderServiceTestSuite) TestHandlePaymentWebhook_ParseError_ReturnsError() {
	parseErr := errors.New("parse error")
	suite.paymentProvider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(nil, parseErr).Once()

	err := suite.orderService.HandlePaymentWebhook(suite.ctx, []byte("data"))

	suite.ErrorContains(err, "orderService.HandlePaymentWebhook")
	suite.ErrorContains(err, parseErr.Error())
}

func (suite *OrderServiceTestSuite) TestHandlePaymentWebhook_GetOrderError_ReturnsError() {
	webhookEvent := &paymentprovider.WebhookEvent{
		PaymentID: suite.paymentID,
		Status:    paymentprovider.PaymentStatusSucceeded,
	}

	dbErr := errors.New("database error")

	suite.paymentProvider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(webhookEvent, nil).Once()

	suite.paymentProvider.EXPECT().GetName().Return("yookassa").Once()

	suite.orderRepo.EXPECT().GetByPayment(suite.ctx, "yookassa", suite.paymentID, true).
		Return(nil, dbErr).Once()

	err := suite.orderService.HandlePaymentWebhook(suite.ctx, []byte("data"))

	suite.ErrorContains(err, "orderService.HandlePaymentWebhook")
	suite.ErrorContains(err, dbErr.Error())
}

func (suite *OrderServiceTestSuite) TestHandlePaymentWebhook_PayOrderFails_ReturnsItemsUnavailableError() {
	webhookEvent := &paymentprovider.WebhookEvent{
		PaymentID: suite.paymentID,
		Status:    paymentprovider.PaymentStatusSucceeded,
		Metadata: paymentprovider.Metadata{
			UserID:  suite.userID,
			OrderID: suite.orderID,
		},
	}

	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusPending,
		TotalAmount: 2500,
		Items: []models.OrderItem{
			{ID: suite.itemID, ProductID: uuid.New(), Quantity: 10, UnitPrice: 1000},
			{ID: suite.itemID, ProductID: uuid.New(), Quantity: 1, UnitPrice: 500},
		},
	}

	suite.paymentProvider.EXPECT().ParseWebhook(mock.AnythingOfType("[]uint8")).
		Return(webhookEvent, nil).Once()

	suite.paymentProvider.EXPECT().GetName().Return("yookassa").Once()

	suite.orderRepo.EXPECT().GetByPayment(suite.ctx, "yookassa", suite.paymentID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 0, Stock: 8},
			{ID: order.Items[1].ProductID, Reserved: 5, Stock: 100},
		}, nil).Once()

	err := suite.orderService.HandlePaymentWebhook(suite.ctx, []byte("data"))

	items, ok := apperror.GetUnavailableItemsFromError(err)
	suite.True(ok)
	suite.Len(items, 1)
	suite.Equal(items[0].ID, order.Items[0].ID)
	suite.Equal(items[0].ProductID, order.Items[0].ProductID)
	suite.Equal(items[0].Action, "deduct")

}

// ==================== CancelOrder Tests ====================

func (suite *OrderServiceTestSuite) TestCancelOrder_Success() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusPending,
		TotalAmount: 2500,
		Items: []models.OrderItem{
			{ProductID: uuid.New(), Quantity: 5, UnitPrice: 1000},
			{ProductID: uuid.New(), Quantity: 1, UnitPrice: 500},
		},
	}

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(order, nil).Once()

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, mock.Anything).
		Return([]*models.Product{
			{ID: order.Items[0].ProductID, Reserved: 6, Stock: 10},
			{ID: order.Items[1].ProductID, Reserved: 5, Stock: 100},
		}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, order).
		Return(nil).Once()

	err := suite.orderService.CancelOrder(suite.ctx, suite.userID, suite.orderID)

	suite.NoError(err)
	suite.Equal(models.OrderStatusCanceled, order.Status)
}

func (suite *OrderServiceTestSuite) TestCancelOrder_OrderNotFound_ReturnsErrOrderNotFound() {
	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	err := suite.orderService.CancelOrder(suite.ctx, suite.userID, suite.orderID)

	suite.ErrorContains(err, "orderService.CancelOrder")
	suite.ErrorIs(err, apperror.ErrOrderNotFound)
}

func (suite *OrderServiceTestSuite) TestCancelOrder_GetByIDDatabaseError_ReturnsError() {
	dbErr := errors.New("database error")

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(nil, dbErr).Once()

	err := suite.orderService.CancelOrder(suite.ctx, suite.userID, suite.orderID)

	suite.ErrorContains(err, "orderService.CancelOrder")
	suite.ErrorIs(err, dbErr)
}

func (suite *OrderServiceTestSuite) TestCancelOrder_InvalidOrderStatus_ReturnsErrInvalidOrderStatus() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusCanceled,
		TotalAmount: 2500,
		Items: []models.OrderItem{
			{ProductID: uuid.New(), Quantity: 5, UnitPrice: 1000},
			{ProductID: uuid.New(), Quantity: 1, UnitPrice: 500},
		},
	}

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(order, nil).Once()

	err := suite.orderService.CancelOrder(suite.ctx, suite.userID, suite.orderID)

	suite.ErrorContains(err, "orderService.CancelOrder")
	suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
}

func (suite *OrderServiceTestSuite) TestCancelOrder_DifferentUserCancelsOrder_ReturnsErrForbidden() {
	order := &models.Order{
		ID:     suite.orderID,
		UserID: new(uuid.New()),
	}

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(order, nil).Once()

	err := suite.orderService.CancelOrder(suite.ctx, suite.userID, suite.orderID)
	suite.ErrorContains(err, "orderService.CancelOrder")
	suite.ErrorIs(err, apperror.ErrForbidden)
}
