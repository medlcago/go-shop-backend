package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperrors"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type ProductServiceTestSuite struct {
	suite.Suite
	mockRepo *mocks.ProductRepositoryMock
	service  ProductService
}

func (suite *ProductServiceTestSuite) SetupTest() {
	suite.mockRepo = new(mocks.ProductRepositoryMock)
	suite.service = NewProductService(suite.mockRepo)
}

func (suite *ProductServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

func TestProductServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}

// ==================== GetProductByID Tests ====================

func (suite *ProductServiceTestSuite) TestGetProductByID_Success() {
	ctx := context.Background()
	productID := uuid.NewString()

	mockProduct := &models.Product{
		ID:         uuid.MustParse(productID),
		Name:       "Test Product",
		Price:      99.99,
		Categories: pq.StringArray{"e2f832de-12e7-46af-9e36-2f2df847f43d", "fc7afa36-c488-4e78-a2a5-852ebfeb06a2"},
	}

	suite.mockRepo.On("GetByID", ctx, productID).
		Return(mockProduct, nil).Once()

	product, err := suite.service.GetProductByID(ctx, productID)

	suite.NoError(err)
	suite.NotNil(product)
	suite.Equal(productID, product.ID)
	suite.Equal("Test Product", product.Name)
	suite.Equal(99.99, product.Price)
	suite.Len(product.Categories, 2)
	suite.Equal("e2f832de-12e7-46af-9e36-2f2df847f43d", product.Categories[0])
	suite.Equal("fc7afa36-c488-4e78-a2a5-852ebfeb06a2", product.Categories[1])
}

func (suite *ProductServiceTestSuite) TestGetProductByID_NotFound() {
	ctx := context.Background()
	productID := uuid.NewString()

	suite.mockRepo.On("GetByID", ctx, productID).
		Return(&models.Product{}, repository.ErrRecordNotFound).Once()

	product, err := suite.service.GetProductByID(ctx, productID)

	suite.Error(err)
	suite.Nil(product)
	suite.ErrorIs(err, apperrors.ErrProductNotFound)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_RepositoryError() {
	ctx := context.Background()
	productID := uuid.NewString()

	repoErr := errors.New("database connection failed")
	suite.mockRepo.On("GetByID", ctx, productID).
		Return(&models.Product{}, repoErr).Once()

	product, err := suite.service.GetProductByID(ctx, productID)

	suite.Error(err)
	suite.Nil(product)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== ListProducts Tests ====================

func (suite *ProductServiceTestSuite) TestListProducts_Success() {
	ctx := context.Background()
	req := dto.ListProductRequest{
		Limit:     10,
		Offset:    0,
		OrderBy:   "price",
		OrderDesc: false,
	}

	mockProducts := []*models.Product{
		{
			Name:       "Product 1",
			Price:      49.99,
			Categories: []string{"784ab8fc-f2b5-41ad-b957-41288b547277"},
		},
		{
			Name:       "Product 2",
			Price:      99.99,
			Categories: []string{"784ab8fc-f2b5-41ad-b957-41288b547277", "fc7afa36-c488-4e78-a2a5-852ebfeb06a2"},
		},
	}

	suite.mockRepo.On("ListProducts", ctx, req).
		Return(mockProducts, 2, nil).Once()

	products, totalCount, err := suite.service.ListProducts(ctx, req)

	suite.NoError(err)
	suite.Equal(int64(2), totalCount)
	suite.Len(products, 2)

	suite.Equal("Product 1", products[0].Name)
	suite.Equal(49.99, products[0].Price)
	suite.Len(products[0].Categories, 1)

	suite.Equal("Product 2", products[1].Name)
	suite.Equal(99.99, products[1].Price)
	suite.Len(products[1].Categories, 2)
}

func (suite *ProductServiceTestSuite) TestListProducts_WithCategoryFilter() {
	ctx := context.Background()
	categoryID := "784ab8fc-f2b5-41ad-b957-41288b547277"
	req := dto.ListProductRequest{
		Limit:      20,
		Offset:     0,
		CategoryID: uuid.MustParse(categoryID),
		OrderBy:    "created_at",
		OrderDesc:  true,
	}

	mockProducts := []*models.Product{
		{
			Name:       "Category Product",
			Categories: []string{categoryID},
		},
	}

	suite.mockRepo.On("ListProducts", ctx, req).
		Return(mockProducts, 1, nil).Once()

	products, totalCount, err := suite.service.ListProducts(ctx, req)

	suite.NoError(err)
	suite.Equal(int64(1), totalCount)
	suite.Len(products, 1)
	suite.Equal("Category Product", products[0].Name)
	suite.Len(products[0].Categories, 1)
}

func (suite *ProductServiceTestSuite) TestListProducts_EmptyResult() {
	ctx := context.Background()
	req := dto.ListProductRequest{
		Limit:  10,
		Offset: 1000,
	}

	suite.mockRepo.On("ListProducts", ctx, req).
		Return([]*models.Product{}, 0, nil).Once()

	products, totalCount, err := suite.service.ListProducts(ctx, req)

	suite.NoError(err)
	suite.Equal(int64(0), totalCount)
	suite.Empty(products)
	suite.Len(products, 0)
}

func (suite *ProductServiceTestSuite) TestListProducts_RepositoryError() {
	ctx := context.Background()
	req := dto.ListProductRequest{
		Limit:  10,
		Offset: 0,
	}

	repoErr := errors.New("query execution failed")
	suite.mockRepo.On("ListProducts", ctx, req).
		Return([]*models.Product{}, 0, repoErr).Once()

	products, totalCount, err := suite.service.ListProducts(ctx, req)

	suite.Error(err)
	suite.Nil(products)
	suite.Equal(int64(0), totalCount)
	suite.ErrorContains(err, repoErr.Error())
}
