package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	serviceMocks "go-shop-backend/internal/service/mocks"
	"go-shop-backend/internal/tasks"
	tasksMocks "go-shop-backend/internal/tasks/mocks"
	uploadMocks "go-shop-backend/internal/upload/mocks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/database"
	paymentproviderMocks "go-shop-backend/pkg/paymentprovider/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type OrderServiceTestSuite struct {
	suite.Suite
	orderRepo        *repoMocks.MockOrderRepository
	orderItemRepo    *repoMocks.MockOrderItemRepository
	addressRepo      *repoMocks.MockAddressRepository
	txManager        *database.NoopTxManager
	paymentProvider  *paymentproviderMocks.MockProvider
	orderTask        *tasksMocks.MockOrderTask
	publicURLBuilder *uploadMocks.MockPublicURLBuilder
	inventoryService *serviceMocks.MockInventoryService
	orderService     *orderService

	ctx              context.Context
	userID           uuid.UUID
	sessionID        uuid.UUID
	orderID          uuid.UUID
	productID        uuid.UUID
	itemID           uuid.UUID
	addressID        uuid.UUID
	paymentID        string
	providerName     string
	orderCancelDelay time.Duration
}

func (suite *OrderServiceTestSuite) SetupTest() {
	suite.orderCancelDelay = 10 * time.Minute

	suite.orderRepo = repoMocks.NewMockOrderRepository(suite.T())
	suite.orderItemRepo = repoMocks.NewMockOrderItemRepository(suite.T())
	suite.addressRepo = repoMocks.NewMockAddressRepository(suite.T())
	suite.txManager = database.NewNoopTxManager()
	suite.orderTask = tasksMocks.NewMockOrderTask(suite.T())
	suite.publicURLBuilder = uploadMocks.NewMockPublicURLBuilder(suite.T())
	suite.inventoryService = serviceMocks.NewMockInventoryService(suite.T())
	suite.orderService = NewOrderService(
		suite.orderRepo,
		suite.orderItemRepo,
		suite.addressRepo,
		suite.orderTask,
		suite.txManager,
		suite.orderCancelDelay,
		suite.publicURLBuilder,
		suite.inventoryService,
	)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
	suite.sessionID = uuid.New()
	suite.orderID = uuid.New()
	suite.productID = uuid.New()
	suite.itemID = uuid.New()
	suite.addressID = uuid.New()
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

	suite.publicURLBuilder.EXPECT().PublicURL(mock.AnythingOfType("string")).
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

	suite.publicURLBuilder.EXPECT().PublicURL(mock.AnythingOfType("string")).
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

	suite.inventoryService.EXPECT().CheckProduct(suite.ctx, suite.productID, req.Quantity).
		Return(product, nil).Once()

	suite.orderItemRepo.EXPECT().Upsert(suite.ctx, mock.AnythingOfType("*models.OrderItem")).
		Return(nil).Once()

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(order, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(nil).Once()

	suite.publicURLBuilder.EXPECT().PublicURL(mock.AnythingOfType("string")).
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

func (suite *OrderServiceTestSuite) TestAddItem_CheckProductError() {
	order := &models.Order{
		ID:        suite.orderID,
		UserID:    &suite.userID,
		SessionID: suite.sessionID,
		Status:    models.OrderStatusDraft,
	}

	req := dto.AddOrderItemRequest{
		ProductID: suite.productID,
		Quantity:  10,
	}

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, false).
		Return(order, nil).Once()

	suite.inventoryService.EXPECT().CheckProduct(suite.ctx, suite.productID, req.Quantity).
		Return(nil, apperror.ErrInsufficientStock).Once()

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
		Return(nil).Once()

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(orderAfterDelete, nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Return(nil).Once()

	suite.publicURLBuilder.EXPECT().PublicURL(mock.AnythingOfType("string")).
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
		Return(repository.ErrRecordNotFound).Once()

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

	address := &models.Address{
		ID: suite.addressID,
	}

	req := dto.OrderCheckoutRequest{
		AddressID: suite.addressID,
	}

	payload := tasks.CancelOrderPayload{
		OrderID: suite.orderID,
		UserID:  suite.userID,
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(address, nil).Once()

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.inventoryService.EXPECT().ReserveItems(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.AnythingOfType("*models.Order")).
		Run(func(ctx context.Context, o *models.Order) {
			suite.Equal(suite.userID, *o.UserID)
			suite.Equal(order.TotalAmount, o.TotalAmount)
			suite.Equal(models.OrderStatusPending, o.Status)
			suite.Nil(o.PaymentID)
			suite.Nil(o.ProviderName)
			suite.NotNil(o.ExpiresAt)
			suite.NotNil(o.Address)
		}).Return(nil).Once()

	suite.orderTask.EXPECT().EnqueueCancelOrder(suite.ctx, payload, suite.orderCancelDelay).
		Return(nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.NoError(err)
	suite.NotNil(response)

	suite.Equal(suite.orderID, response.ID)
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

	address := &models.Address{
		ID: suite.addressID,
	}

	req := dto.OrderCheckoutRequest{
		AddressID: suite.addressID,
	}

	payload := tasks.CancelOrderPayload{
		OrderID: suite.orderID,
		UserID:  suite.userID,
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(address, nil).Once()

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.inventoryService.EXPECT().ReserveItems(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.MatchedBy(func(o *models.Order) bool {
		return o.UserID != nil && *o.UserID == suite.userID
	})).Return(nil).Once()

	suite.orderTask.EXPECT().EnqueueCancelOrder(suite.ctx, payload, suite.orderCancelDelay).
		Return(nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.NoError(err)
	suite.NotNil(response)

	suite.NotNil(order.UserID)
	suite.Equal(suite.userID, *order.UserID)
	suite.Equal(suite.orderID, response.ID)
}

func (suite *OrderServiceTestSuite) TestCheckout_AddressNotFound() {
	req := dto.OrderCheckoutRequest{
		AddressID: suite.addressID,
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.orderService.Checkout(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrAddressNotFound)
	suite.ErrorContains(err, "orderService.Checkout")
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

	address := &models.Address{
		ID: suite.addressID,
	}

	req := dto.OrderCheckoutRequest{
		AddressID: suite.addressID,
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(address, nil).Once()

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, suite.userID, suite.sessionID, order.ID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
	suite.ErrorContains(err, "orderService.Checkout")
}

func (suite *OrderServiceTestSuite) TestCheckout_EmptyOrder() {
	order := &models.Order{
		ID:          suite.orderID,
		UserID:      &suite.userID,
		SessionID:   suite.sessionID,
		Status:      models.OrderStatusDraft,
		TotalAmount: 4000,
	}

	address := &models.Address{
		ID: suite.addressID,
	}

	req := dto.OrderCheckoutRequest{
		AddressID: suite.addressID,
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(address, nil).Once()

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, suite.userID, suite.sessionID, order.ID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrEmptyOrder)
	suite.ErrorContains(err, "orderService.Checkout")
}

func (suite *OrderServiceTestSuite) TestCheckout_UnavailableItemError() {
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

	unavailableItems := []apperror.UnavailableItem{
		{
			ID:           suite.itemID,
			ProductID:    suite.productID,
			RequestedQty: 10,
			AvailableQty: 5,
			Action:       "reserve",
			Reason:       apperror.ErrInsufficientStock.Error(),
		},
	}

	address := &models.Address{
		ID: suite.addressID,
	}

	req := dto.OrderCheckoutRequest{
		AddressID: suite.addressID,
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(address, nil).Once()

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(order, nil).Once()

	suite.inventoryService.EXPECT().ReserveItems(suite.ctx, mock.Anything).
		Return(apperror.UnavailableItemsError(unavailableItems)).Once()

	response, err := suite.orderService.Checkout(suite.ctx, suite.userID, suite.sessionID, order.ID, req)

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
	suite.ErrorContains(err, "orderService.Checkout")
}

func (suite *OrderServiceTestSuite) TestCheckout_Forbidden() {
	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	address := &models.Address{
		ID: suite.addressID,
	}

	req := dto.OrderCheckoutRequest{
		AddressID: suite.addressID,
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(address, nil).Once()

	response, err := suite.orderService.Checkout(suite.ctx, suite.userID, suite.sessionID, suite.orderID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrForbidden)
	suite.ErrorContains(err, "orderService.Checkout")
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

	address := &models.Address{
		ID: suite.addressID,
	}

	req := dto.OrderCheckoutRequest{
		AddressID: suite.addressID,
	}

	suite.addressRepo.EXPECT().GetByID(suite.ctx, suite.addressID, suite.userID).
		Return(address, nil).Once()

	dbError := errors.New("db error")

	suite.orderRepo.EXPECT().GetByOwner(suite.ctx, suite.orderID, &suite.userID, suite.sessionID, true).
		Return(nil, dbError).Once()

	response, err := suite.orderService.Checkout(suite.ctx, suite.userID, suite.sessionID, order.ID, req)

	suite.Nil(response)
	suite.ErrorIs(err, dbError)
	suite.ErrorContains(err, "orderService.Checkout")
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

	suite.inventoryService.EXPECT().ReleaseItems(suite.ctx, mock.Anything).
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

// ==================== UpdateOrderStatus Tests ====================

func (suite *OrderServiceTestSuite) TestUpdateOrderStatus_Success_OrderStatusPaid() {
	order := &models.Order{
		ID:     suite.orderID,
		Status: models.OrderStatusPending,
	}

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(order, nil).Once()

	suite.inventoryService.EXPECT().DeductItems(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.MatchedBy(func(order *models.Order) bool {
		return order.Status == models.OrderStatusPaid
	})).Return(nil).Once()

	err := suite.orderService.UpdateOrderStatus(suite.ctx, suite.orderID, models.OrderStatusPaid)
	suite.NoError(err)
}

func (suite *OrderServiceTestSuite) TestUpdateOrderStatus_Success_OrderStatusCanceled() {
	order := &models.Order{
		ID:     suite.orderID,
		Status: models.OrderStatusPending,
	}

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(order, nil).Once()

	suite.inventoryService.EXPECT().ReleaseItems(suite.ctx, mock.Anything).
		Return(nil).Once()

	suite.orderRepo.EXPECT().Update(suite.ctx, mock.MatchedBy(func(order *models.Order) bool {
		return order.Status == models.OrderStatusCanceled
	})).Return(nil).Once()

	err := suite.orderService.UpdateOrderStatus(suite.ctx, suite.orderID, models.OrderStatusCanceled)
	suite.NoError(err)
}

func (suite *OrderServiceTestSuite) TestUpdateOrderStatus_InvalidOrderStatus() {
	err := suite.orderService.UpdateOrderStatus(suite.ctx, suite.orderID, "test")

	suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
	suite.ErrorContains(err, "orderService.UpdateOrderStatus")
}

func (suite *OrderServiceTestSuite) TestUpdateOrderStatus_InvalidTransitionStatus() {
	order := &models.Order{
		ID:     suite.orderID,
		Status: models.OrderStatusCanceled,
	}

	suite.orderRepo.EXPECT().GetByID(suite.ctx, suite.orderID, true).
		Return(order, nil).Once()

	err := suite.orderService.UpdateOrderStatus(suite.ctx, suite.orderID, models.OrderStatusPaid)

	suite.ErrorIs(err, apperror.ErrInvalidOrderStatus)
	suite.ErrorContains(err, "orderService.UpdateOrderStatus")
}
