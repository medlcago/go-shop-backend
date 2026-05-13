package service

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/database"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type InventoryServiceTestSuite struct {
	suite.Suite

	productRepo      *repoMocks.MockProductRepository
	txManager        *database.NoopTxManager
	inventoryService *inventoryService

	ctx       context.Context
	productID uuid.UUID
	itemID    uuid.UUID
}

func (suite *InventoryServiceTestSuite) SetupTest() {
	suite.productRepo = repoMocks.NewMockProductRepository(suite.T())
	suite.txManager = database.NewNoopTxManager()
	suite.inventoryService = NewInventoryService(
		suite.productRepo,
		suite.txManager,
	)

	suite.ctx = context.Background()
	suite.productID = uuid.New()
	suite.itemID = uuid.New()
}

func TestInventoryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(InventoryServiceTestSuite))
}

// ==================== CheckProduct Tests ====================

func (suite *InventoryServiceTestSuite) TestCheckProduct_Success() {
	product := &models.Product{
		ID:       suite.productID,
		Name:     "Product",
		Stock:    10,
		Reserved: 0,
		IsActive: true,
	}

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil)

	response, err := suite.inventoryService.CheckProduct(suite.ctx, suite.productID, 5)

	suite.NoError(err)
	suite.Equal(response.ID, product.ID)
}

func (suite *InventoryServiceTestSuite) TestCheckProduct_ProductNotFound() {
	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(nil, repository.ErrRecordNotFound)

	response, err := suite.inventoryService.CheckProduct(suite.ctx, suite.productID, 5)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
	suite.ErrorContains(err, "inventoryService.CheckProduct")
}

func (suite *InventoryServiceTestSuite) TestCheckProduct_ProductNotActive() {
	product := &models.Product{
		ID:       suite.productID,
		Name:     "Product",
		Stock:    10,
		Reserved: 0,
		IsActive: false,
	}

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil)

	response, err := suite.inventoryService.CheckProduct(suite.ctx, suite.productID, 5)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrProductNotActive)
	suite.ErrorContains(err, "inventoryService.CheckProduct")
}

func (suite *InventoryServiceTestSuite) TestCheckProduct_InsufficientStock() {
	product := &models.Product{
		ID:       suite.productID,
		Name:     "Product",
		Stock:    10,
		Reserved: 8,
		IsActive: true,
	}

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil)

	response, err := suite.inventoryService.CheckProduct(suite.ctx, suite.productID, 5)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInsufficientStock)
	suite.ErrorContains(err, "inventoryService.CheckProduct")
}

// ==================== ReserveItems Tests ====================

func (suite *InventoryServiceTestSuite) TestReserveItems_Success() {
	items := []dto.InventoryItem{
		{
			ItemID:    suite.itemID,
			ProductID: suite.productID,
			Quantity:  10,
		},
		{
			ItemID:    uuid.New(),
			ProductID: uuid.New(),
			Quantity:  5,
		},
	}

	productIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	products := []*models.Product{
		{
			ID:       items[0].ProductID,
			Stock:    1000,
			Reserved: 10,
			IsActive: true,
		},
		{
			ID:       items[1].ProductID,
			Stock:    20,
			Reserved: 7,
			IsActive: true,
		},
	}

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, productIDs).
		Return(products, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, products).
		Return(nil).Once()

	err := suite.inventoryService.ReserveItems(suite.ctx, items)

	suite.NoError(err)
	suite.Equal(20, products[0].Reserved)
	suite.Equal(1000, products[0].Stock)
	suite.Equal(12, products[1].Reserved)
	suite.Equal(20, products[1].Stock)
}

func (suite *InventoryServiceTestSuite) TestReserveItems_ProductNotActive() {
	items := []dto.InventoryItem{
		{
			ItemID:    suite.itemID,
			ProductID: suite.productID,
			Quantity:  10,
		},
	}

	product := &models.Product{
		ID:       suite.productID,
		Stock:    20,
		IsActive: false,
	}

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, []uuid.UUID{suite.productID}).
		Return([]*models.Product{product}, nil).Once()

	err := suite.inventoryService.ReserveItems(suite.ctx, items)

	suite.Error(err)
	suite.ErrorContains(err, "inventoryService.ReserveItems")
	unavailableItems, ok := apperror.GetUnavailableItemsFromError(err)
	suite.True(ok)
	suite.Len(unavailableItems, 1)
	suite.Equal(suite.productID, unavailableItems[0].ProductID)
	suite.Equal("reserve", unavailableItems[0].Action)
	suite.Equal(apperror.ErrProductNotActive.Error(), unavailableItems[0].Reason)
}

func (suite *InventoryServiceTestSuite) TestReserveItems_InsufficientStock() {
	items := []dto.InventoryItem{
		{
			ItemID:    suite.itemID,
			ProductID: suite.productID,
			Quantity:  101,
		},
		{
			ItemID:    uuid.New(),
			ProductID: uuid.New(),
			Quantity:  5,
		},
	}

	productIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.ProductID)
	}

	products := []*models.Product{
		{
			ID:       items[0].ProductID,
			Stock:    10,
			Reserved: 0,
			IsActive: true,
		},
		{
			ID:       items[1].ProductID,
			Stock:    20,
			Reserved: 7,
			IsActive: true,
		},
	}

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, productIDs).
		Return(products, nil).Once()

	err := suite.inventoryService.ReserveItems(suite.ctx, items)

	suite.Error(err)
	suite.ErrorContains(err, "inventoryService.ReserveItems")
	unavailableItems, ok := apperror.GetUnavailableItemsFromError(err)
	suite.True(ok)
	suite.Len(unavailableItems, 1)
	suite.Equal(suite.productID, unavailableItems[0].ProductID)
	suite.Equal("reserve", unavailableItems[0].Action)
	suite.Equal(apperror.ErrInsufficientStock.Error(), unavailableItems[0].Reason)
}

// ==================== ReleaseItems Tests ====================

func (suite *InventoryServiceTestSuite) TestReleaseItems_Success() {
	items := []dto.InventoryItem{
		{
			ItemID:    suite.itemID,
			ProductID: suite.productID,
			Quantity:  5,
		},
	}

	product := &models.Product{
		ID:       suite.productID,
		Stock:    100,
		Reserved: 20,
	}

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, []uuid.UUID{suite.productID}).
		Return([]*models.Product{product}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, []*models.Product{product}).
		Return(nil).Once()

	err := suite.inventoryService.ReleaseItems(suite.ctx, items)

	suite.NoError(err)
	suite.Equal(100, product.Stock)
	suite.Equal(15, product.Reserved)
}

func (suite *InventoryServiceTestSuite) TestReleaseItems_InconsistentStock() {
	items := []dto.InventoryItem{
		{
			ItemID:    suite.itemID,
			ProductID: suite.productID,
			Quantity:  50,
		},
	}

	product := &models.Product{
		ID:       suite.productID,
		Stock:    100,
		Reserved: 20,
	}

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, []uuid.UUID{suite.productID}).
		Return([]*models.Product{product}, nil).Once()

	err := suite.inventoryService.ReleaseItems(suite.ctx, items)

	suite.Error(err)
	suite.ErrorContains(err, "inventoryService.ReleaseItems")
	unavailableItems, ok := apperror.GetUnavailableItemsFromError(err)
	suite.True(ok)
	suite.Len(unavailableItems, 1)
	suite.Equal(suite.productID, unavailableItems[0].ProductID)
	suite.Equal("release", unavailableItems[0].Action)
	suite.Equal(apperror.ErrInconsistentStock.Error(), unavailableItems[0].Reason)
}

// ==================== DeductItems Tests ====================

func (suite *InventoryServiceTestSuite) TestDeductItems_Success() {
	items := []dto.InventoryItem{
		{
			ItemID:    suite.itemID,
			ProductID: suite.productID,
			Quantity:  50,
		},
	}

	product := &models.Product{
		ID:       suite.productID,
		Stock:    100,
		Reserved: 70,
	}

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, []uuid.UUID{suite.productID}).
		Return([]*models.Product{product}, nil).Once()

	suite.productRepo.EXPECT().BulkUpsert(suite.ctx, []*models.Product{product}).
		Return(nil).Once()

	err := suite.inventoryService.DeductItems(suite.ctx, items)

	suite.NoError(err)
	suite.Equal(50, product.Stock)
	suite.Equal(20, product.Reserved)
}

func (suite *InventoryServiceTestSuite) TestDeductItems_InconsistentStock() {
	items := []dto.InventoryItem{
		{
			ItemID:    suite.itemID,
			ProductID: suite.productID,
			Quantity:  200,
		},
	}

	product := &models.Product{
		ID:       suite.productID,
		Stock:    100,
		Reserved: 70,
	}

	suite.productRepo.EXPECT().GetByIDsForUpdate(suite.ctx, []uuid.UUID{suite.productID}).
		Return([]*models.Product{product}, nil).Once()

	err := suite.inventoryService.DeductItems(suite.ctx, items)
	suite.Error(err)
	suite.ErrorContains(err, "inventoryService.DeductItems")
	unavailableItems, ok := apperror.GetUnavailableItemsFromError(err)
	suite.True(ok)
	suite.Len(unavailableItems, 1)
	suite.Equal(suite.productID, unavailableItems[0].ProductID)
	suite.Equal("deduct", unavailableItems[0].Action)
	suite.Equal(apperror.ErrInconsistentStock.Error(), unavailableItems[0].Reason)
}
