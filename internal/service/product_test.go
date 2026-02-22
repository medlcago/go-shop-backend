package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	serviceMocks "go-shop-backend/internal/service/mocks"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProductServiceTestSuite struct {
	suite.Suite
	productRepo    *repoMocks.MockProductRepository
	urlBuilder     *serviceMocks.MockPublicURLBuilder
	productService ProductService
}

func (suite *ProductServiceTestSuite) SetupTest() {
	suite.productRepo = repoMocks.NewMockProductRepository(suite.T())
	suite.urlBuilder = serviceMocks.NewMockPublicURLBuilder(suite.T())
	suite.productService = NewProductService(suite.productRepo, suite.urlBuilder)
}

func TestProductServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}

// ==================== GetProductByID Tests ====================

func (suite *ProductServiceTestSuite) TestGetProductByID_Success() {
	ctx := context.Background()
	productID := uuid.New()

	mockProduct := &models.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 10_000,
		Categories: []models.Category{
			{
				ID:   uuid.MustParse("e2f832de-12e7-46af-9e36-2f2df847f43d"),
				Name: "Test Category 1",
			},
			{
				ID:   uuid.MustParse("fc7afa36-c488-4e78-a2a5-852ebfeb06a2"),
				Name: "Test Category 2",
			},
		},
	}

	suite.productRepo.EXPECT().GetByID(ctx, productID, true).
		Return(mockProduct, nil).Once()

	product, err := suite.productService.GetProductByID(ctx, productID)

	suite.NoError(err)
	suite.NotNil(product)
	suite.Equal(productID, product.ID)
	suite.Equal("Test Product", product.Name)
	suite.Equal(int64(10_000), product.Price)
	suite.Len(product.Categories, 2)
	suite.Equal("e2f832de-12e7-46af-9e36-2f2df847f43d", product.Categories[0].ID)
	suite.Equal("fc7afa36-c488-4e78-a2a5-852ebfeb06a2", product.Categories[1].ID)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_NotFound() {
	ctx := context.Background()
	productID := uuid.New()

	suite.productRepo.EXPECT().GetByID(ctx, productID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	product, err := suite.productService.GetProductByID(ctx, productID)

	suite.Nil(product)
	suite.ErrorIs(err, apperrors.ErrProductNotFound)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_RepositoryError() {
	ctx := context.Background()
	productID := uuid.New()

	repoErr := errors.New("database connection failed")
	suite.productRepo.EXPECT().GetByID(ctx, productID, true).
		Return(nil, repoErr).Once()

	product, err := suite.productService.GetProductByID(ctx, productID)

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
			Name:  "Product 1",
			Price: 499_900,
			Categories: []models.Category{
				{
					ID:   uuid.MustParse("e2f832de-12e7-46af-9e36-2f2df847f43d"),
					Name: "Test Category 1",
				},
			},
		},
		{
			Name:  "Product 2",
			Price: 999_900,
			Categories: []models.Category{
				{
					ID:   uuid.MustParse("e2f832de-12e7-46af-9e36-2f2df847f43d"),
					Name: "Test Category 1",
				},
				{
					ID:   uuid.MustParse("fc7afa36-c488-4e78-a2a5-852ebfeb06a2"),
					Name: "Test Category 2",
				},
			},
		},
	}

	suite.productRepo.EXPECT().ListProducts(ctx, req).
		Return(mockProducts, 2, nil).Once()

	products, totalCount, err := suite.productService.ListProducts(ctx, req)

	suite.NoError(err)
	suite.Equal(int64(2), totalCount)
	suite.Len(products, 2)

	suite.Equal("Product 1", products[0].Name)
	suite.Equal(int64(499_900), products[0].Price)
	suite.Len(products[0].Categories, 1)

	suite.Equal("Product 2", products[1].Name)
	suite.Equal(int64(999_900), products[1].Price)
	suite.Len(products[1].Categories, 2)
}

func (suite *ProductServiceTestSuite) TestListProducts_WithCategoryFilter() {
	ctx := context.Background()
	categoryID := uuid.New()
	req := dto.ListProductRequest{
		Limit:      20,
		Offset:     0,
		CategoryID: categoryID,
		OrderBy:    "created_at",
		OrderDesc:  true,
	}

	mockProducts := []*models.Product{
		{
			Name: "Category Product",
			Categories: []models.Category{
				{
					ID: categoryID,
				},
			},
		},
	}

	suite.productRepo.EXPECT().ListProducts(ctx, req).
		Return(mockProducts, 1, nil).Once()

	products, totalCount, err := suite.productService.ListProducts(ctx, req)

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

	suite.productRepo.EXPECT().ListProducts(ctx, req).
		Return([]*models.Product{}, 0, nil).Once()

	products, totalCount, err := suite.productService.ListProducts(ctx, req)

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
	suite.productRepo.EXPECT().ListProducts(ctx, req).
		Return(nil, 0, repoErr).Once()

	products, totalCount, err := suite.productService.ListProducts(ctx, req)

	suite.Nil(products)
	suite.Equal(int64(0), totalCount)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== CreateProduct Tests ====================

func (suite *ProductServiceTestSuite) TestCreateProduct_Success_DefaultIsActive() {
	ctx := context.Background()
	req := dto.ProductCreateRequest{Name: "Product 1"}

	suite.productRepo.EXPECT().Create(ctx, mock.MatchedBy(func(product *models.Product) bool {
		return product.IsActive
	})).Return(nil).Once()

	product, err := suite.productService.CreateProduct(ctx, req)

	suite.NoError(err)
	suite.NotNil(product)
	suite.True(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_Success_IsActiveFalse() {
	ctx := context.Background()
	isActive := false
	req := dto.ProductCreateRequest{Name: "Product 1", IsActive: &isActive}

	suite.productRepo.EXPECT().Create(ctx, mock.MatchedBy(func(product *models.Product) bool {
		return !product.IsActive
	})).Return(nil).Once()

	product, err := suite.productService.CreateProduct(ctx, req)

	suite.NoError(err)
	suite.NotNil(product)
	suite.False(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_RepositoryError() {
	ctx := context.Background()
	req := dto.ProductCreateRequest{}

	repoErr := errors.New("query execution failed")
	suite.productRepo.EXPECT().Create(ctx, mock.Anything).
		Return(repoErr).Once()

	product, err := suite.productService.CreateProduct(ctx, req)

	suite.Nil(product)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== UpdateProduct Tests ====================

func (suite *ProductServiceTestSuite) TestUpdateProduct_Success_UpdateAllFields() {
	ctx := context.Background()
	productID := uuid.New()

	name := "Updated Product"
	description := "updated description"
	price := int64(999_900)
	stock := 20
	isActive := true

	existingProduct := &models.Product{
		ID:          productID,
		Name:        "Old Product",
		Slug:        "old-product",
		Description: utils.Ptr("old description"),
		Price:       50,
		Stock:       5,
		IsActive:    false,
	}

	req := dto.ProductUpdateRequest{
		Name:        &name,
		Description: &description,
		Price:       &price,
		Stock:       &stock,
		IsActive:    &isActive,
	}

	suite.productRepo.
		EXPECT().GetByID(ctx, productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		EXPECT().Update(ctx, existingProduct).
		Return(nil).
		Once()

	product, err := suite.productService.UpdateProduct(ctx, productID, req)

	suite.NoError(err)
	suite.NotNil(product)

	suite.Equal(productID, product.ID)
	suite.Equal(name, product.Name)
	suite.Equal(utils.Slugify(name), product.Slug)
	suite.Equal(&description, product.Description)
	suite.Equal(price, product.Price)
	suite.Equal(stock, product.Stock)
	suite.True(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_Success_PartialUpdate() {
	ctx := context.Background()
	productID := uuid.New()

	name := "New Name"

	existingProduct := &models.Product{
		ID:          productID,
		Name:        "Old Name",
		Slug:        "old-name",
		Description: utils.Ptr("description"),
		Price:       10_000,
		Stock:       3,
		IsActive:    true,
	}

	req := dto.ProductUpdateRequest{
		Name: &name,
	}

	suite.productRepo.
		EXPECT().GetByID(ctx, productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		EXPECT().Update(ctx, existingProduct).
		Return(nil).
		Once()

	product, err := suite.productService.UpdateProduct(ctx, productID, req)

	suite.NoError(err)
	suite.NotNil(product)

	suite.Equal(name, product.Name)
	suite.Equal(utils.Slugify(name), product.Slug)

	suite.Equal(existingProduct.Price, product.Price)
	suite.Equal(existingProduct.Stock, product.Stock)
	suite.Equal(existingProduct.Description, product.Description)
	suite.True(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_ProductNotFound() {
	ctx := context.Background()
	productID := uuid.New()
	req := dto.ProductUpdateRequest{}

	suite.productRepo.
		EXPECT().GetByID(ctx, productID, false).
		Return(nil, repository.ErrRecordNotFound).
		Once()

	product, err := suite.productService.UpdateProduct(ctx, productID, req)

	suite.Nil(product)
	suite.ErrorIs(apperrors.ErrProductNotFound, err)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_UpdateRepositoryError() {
	ctx := context.Background()
	productID := uuid.New()

	existingProduct := &models.Product{ID: productID}
	req := dto.ProductUpdateRequest{}

	repoErr := errors.New("update failed")

	suite.productRepo.
		EXPECT().GetByID(ctx, productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		EXPECT().Update(ctx, existingProduct).
		Return(repoErr).
		Once()

	product, err := suite.productService.UpdateProduct(ctx, productID, req)

	suite.Nil(product)
	suite.ErrorContains(err, repoErr.Error())
}
