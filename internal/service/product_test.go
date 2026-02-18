package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type URLBuilder struct {
	mock.Mock
}

func (u *URLBuilder) PublicURL(ctx context.Context, objectKey string) string {
	args := u.Called(ctx, objectKey)
	return args.String(0)
}

type ProductServiceTestSuite struct {
	suite.Suite
	productRepo    *mocks.ProductRepositoryMock
	urlBuilder     *URLBuilder
	productService ProductService
}

func (suite *ProductServiceTestSuite) SetupTest() {
	suite.productRepo = new(mocks.ProductRepositoryMock)
	suite.urlBuilder = new(URLBuilder)
	suite.productService = NewProductService(suite.productRepo, suite.urlBuilder)
}

func (suite *ProductServiceTestSuite) TearDownTest() {
	suite.productRepo.AssertExpectations(suite.T())
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

	suite.productRepo.On("GetByID", ctx, productID, true).
		Return(mockProduct, nil).Once()

	product, err := suite.productService.GetProductByID(ctx, productID)

	suite.NoError(err)
	suite.NotNil(product)
	suite.Equal(productID.String(), product.ID)
	suite.Equal("Test Product", product.Name)
	suite.Equal(int64(10_000), product.Price)
	suite.Len(product.Categories, 2)
	suite.Equal("e2f832de-12e7-46af-9e36-2f2df847f43d", product.Categories[0].ID)
	suite.Equal("fc7afa36-c488-4e78-a2a5-852ebfeb06a2", product.Categories[1].ID)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_NotFound() {
	ctx := context.Background()
	productID := uuid.New()

	suite.productRepo.On("GetByID", ctx, productID, true).
		Return(&models.Product{}, repository.ErrRecordNotFound).Once()

	product, err := suite.productService.GetProductByID(ctx, productID)

	suite.Error(err)
	suite.Nil(product)
	suite.ErrorIs(err, apperrors.ErrProductNotFound)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_RepositoryError() {
	ctx := context.Background()
	productID := uuid.New()

	repoErr := errors.New("database connection failed")
	suite.productRepo.On("GetByID", ctx, productID, true).
		Return(&models.Product{}, repoErr).Once()

	product, err := suite.productService.GetProductByID(ctx, productID)

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

	suite.productRepo.On("ListProducts", ctx, req).
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
			Name: "Category Product",
			Categories: []models.Category{
				{
					ID: uuid.MustParse(categoryID),
				},
			},
		},
	}

	suite.productRepo.On("ListProducts", ctx, req).
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

	suite.productRepo.On("ListProducts", ctx, req).
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
	suite.productRepo.On("ListProducts", ctx, req).
		Return([]*models.Product{}, 0, repoErr).Once()

	products, totalCount, err := suite.productService.ListProducts(ctx, req)

	suite.Error(err)
	suite.Nil(products)
	suite.Equal(int64(0), totalCount)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== CreateProduct Tests ====================

func (suite *ProductServiceTestSuite) TestCreateProduct_Success_DefaultIsActive() {
	ctx := context.Background()
	description := "product 1 description"
	req := dto.ProductCreateRequest{
		Name:        "Product 1",
		Description: &description,
		Price:       499_900,
		Stock:       10,
	}

	slug := utils.Slugify(req.Name)
	productModel := &models.Product{
		Name:        "Product 1",
		Description: &description,
		Price:       499_900,
		Stock:       10,
		IsActive:    true,
		Slug:        slug,
	}
	suite.productRepo.On("CreateProduct", ctx, productModel).Return(nil).Once()

	product, err := suite.productService.CreateProduct(ctx, req)

	suite.NoError(err)
	suite.NotNil(product)

	suite.Equal(product.ID, productModel.ID.String())
	suite.Equal(product.Name, productModel.Name)
	suite.Equal(product.Description, productModel.Description)
	suite.Equal(product.Price, productModel.Price)
	suite.Equal(product.Stock, productModel.Stock)
	suite.Equal(product.Slug, slug)
	suite.True(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_Success_IsActiveFalse() {
	ctx := context.Background()
	isActive := false
	req := dto.ProductCreateRequest{
		Name:     "Product 1",
		Price:    50,
		Stock:    5,
		IsActive: &isActive,
	}

	suite.productRepo.On("CreateProduct", ctx, mock.MatchedBy(func(product *models.Product) bool {
		return !product.IsActive
	})).Return(nil).Once()

	product, err := suite.productService.CreateProduct(ctx, req)

	suite.NoError(err)
	suite.NotNil(product)
	suite.False(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_RepositoryError() {
	ctx := context.Background()
	req := dto.ProductCreateRequest{
		Name:  "Product 1",
		Price: 50,
		Stock: 5,
	}

	repoErr := errors.New("query execution failed")
	suite.productRepo.On("CreateProduct", ctx, mock.Anything).Return(repoErr).Once()

	product, err := suite.productService.CreateProduct(ctx, req)

	suite.Error(err)
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
		On("GetByID", ctx, productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		On("UpdateProduct", ctx, existingProduct).
		Return(nil).
		Once()

	resp, err := suite.productService.UpdateProduct(ctx, productID, req)

	suite.NoError(err)
	suite.NotNil(resp)

	suite.Equal(productID.String(), resp.ID)
	suite.Equal(name, resp.Name)
	suite.Equal(utils.Slugify(name), resp.Slug)
	suite.Equal(&description, resp.Description)
	suite.Equal(price, resp.Price)
	suite.Equal(stock, resp.Stock)
	suite.True(resp.IsActive)
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
		On("GetByID", ctx, productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		On("UpdateProduct", ctx, existingProduct).
		Return(nil).
		Once()

	resp, err := suite.productService.UpdateProduct(ctx, productID, req)

	suite.NoError(err)
	suite.NotNil(resp)

	suite.Equal(name, resp.Name)
	suite.Equal(utils.Slugify(name), resp.Slug)

	suite.Equal(existingProduct.Price, resp.Price)
	suite.Equal(existingProduct.Stock, resp.Stock)
	suite.Equal(existingProduct.Description, resp.Description)
	suite.True(resp.IsActive)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_ProductNotFound() {
	ctx := context.Background()
	productID := uuid.New()

	req := dto.ProductUpdateRequest{}

	suite.productRepo.
		On("GetByID", ctx, productID, false).
		Return(&models.Product{}, repository.ErrRecordNotFound).
		Once()

	resp, err := suite.productService.UpdateProduct(ctx, productID, req)

	suite.Error(err)
	suite.Nil(resp)
	suite.Equal(apperrors.ErrProductNotFound, err)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_UpdateRepositoryError() {
	ctx := context.Background()
	productID := uuid.New()

	price := int64(100000)

	existingProduct := &models.Product{
		ID:    productID,
		Name:  "Product",
		Price: 50_000,
	}

	req := dto.ProductUpdateRequest{
		Price: &price,
	}

	repoErr := errors.New("update failed")

	suite.productRepo.
		On("GetByID", ctx, productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		On("UpdateProduct", ctx, existingProduct).
		Return(repoErr).
		Once()

	resp, err := suite.productService.UpdateProduct(ctx, productID, req)

	suite.Error(err)
	suite.Nil(resp)
	suite.ErrorContains(err, repoErr.Error())
}
