package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	uploadMocks "go-shop-backend/internal/upload/mocks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProductServiceTestSuite struct {
	suite.Suite
	productRepo    *repoMocks.MockProductRepository
	uploadManager  *uploadMocks.MockManager
	productService *productService

	ctx        context.Context
	productID  uuid.UUID
	categoryID uuid.UUID
}

func (suite *ProductServiceTestSuite) SetupTest() {
	suite.productRepo = repoMocks.NewMockProductRepository(suite.T())
	suite.uploadManager = uploadMocks.NewMockManager(suite.T())
	suite.productService = NewProductService(suite.productRepo, suite.uploadManager)

	suite.ctx = context.Background()
	suite.productID = uuid.New()
	suite.categoryID = uuid.New()
}

func TestProductServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ProductServiceTestSuite))
}

// ==================== GetProductByID Tests ====================

func (suite *ProductServiceTestSuite) TestGetProductByID_Success() {
	product := &models.Product{
		ID:    suite.productID,
		Name:  "Test Product",
		Price: 10_000,
		Categories: []*models.Category{
			{
				ID:   uuid.New(),
				Name: "Test Category 1",
			},
			{
				ID:   uuid.New(),
				Name: "Test Category 2",
			},
		},
		Images: []*models.Upload{
			{
				ObjectKey: uuid.NewString(),
			},
			{
				ObjectKey: uuid.NewString(),
			},
		},
	}

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, true).
		Return(product, nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.Anything).
		Return(uuid.NewString()).Times(len(product.Images))

	response, err := suite.productService.GetProductByID(suite.ctx, suite.productID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(suite.productID, response.ID)
	suite.Equal("Test Product", response.Name)
	suite.Equal(int64(10_000), response.Price)
	suite.Len(response.Categories, 2)
	suite.Equal(product.Categories[0].ID, response.Categories[0].ID)
	suite.Equal(product.Categories[1].ID, response.Categories[1].ID)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_NotFound() {
	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.productService.GetProductByID(suite.ctx, suite.productID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_RepositoryError() {
	repoErr := errors.New("database connection failed")
	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, true).
		Return(nil, repoErr).Once()

	response, err := suite.productService.GetProductByID(suite.ctx, suite.productID)

	suite.Nil(response)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== ListProducts Tests ====================

func (suite *ProductServiceTestSuite) TestListProducts_Success() {
	req := dto.ListProductRequest{
		Limit:     10,
		Offset:    0,
		OrderBy:   "price",
		OrderDesc: false,
	}

	products := []*models.Product{
		{
			Name:  "Product 1",
			Price: 499_900,
			Categories: []*models.Category{
				{
					ID:   uuid.New(),
					Name: "Test Category 1",
				},
			},
			Images: []*models.Upload{
				{
					ObjectKey: uuid.NewString(),
				},
				{
					ObjectKey: uuid.NewString(),
				},
			},
		},
		{
			Name:  "Product 2",
			Price: 999_900,
			Categories: []*models.Category{
				{
					ID:   uuid.New(),
					Name: "Test Category 1",
				},
				{
					ID:   uuid.New(),
					Name: "Test Category 2",
				},
			},
			Images: []*models.Upload{
				{
					ObjectKey: uuid.NewString(),
				},
			},
		},
	}

	suite.productRepo.EXPECT().ListProducts(suite.ctx, req).
		Return(products, 2, nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.Anything).
		Return(uuid.NewString()).Times(len(products[0].Images) + len(products[1].Images))

	response, total, err := suite.productService.ListProducts(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(response, 2)

	suite.Equal("Product 1", response[0].Name)
	suite.Equal(int64(499_900), response[0].Price)
	suite.Len(response[0].Categories, 1)

	suite.Equal("Product 2", response[1].Name)
	suite.Equal(int64(999_900), response[1].Price)
	suite.Len(response[1].Categories, 2)
}

func (suite *ProductServiceTestSuite) TestListProducts_WithCategoryFilter() {
	req := dto.ListProductRequest{
		Limit:      20,
		Offset:     0,
		CategoryID: suite.categoryID,
		OrderBy:    "created_at",
		OrderDesc:  true,
	}

	products := []*models.Product{
		{
			Name: "Category Product",
			Categories: []*models.Category{
				{
					ID: suite.categoryID,
				},
			},
			Images: []*models.Upload{
				{
					ObjectKey: uuid.NewString(),
				},
				{
					ObjectKey: uuid.NewString(),
				},
			},
		},
	}

	suite.productRepo.EXPECT().ListProducts(suite.ctx, req).
		Return(products, 1, nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.Anything).
		Return(uuid.NewString()).Times(len(products[0].Images))

	response, total, err := suite.productService.ListProducts(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(1), total)
	suite.Len(response, 1)
	suite.Equal("Category Product", response[0].Name)
	suite.Len(response[0].Categories, 1)
}

func (suite *ProductServiceTestSuite) TestListProducts_EmptyResult() {
	req := dto.ListProductRequest{
		Limit:  10,
		Offset: 1000,
	}

	suite.productRepo.EXPECT().ListProducts(suite.ctx, req).
		Return([]*models.Product{}, 0, nil).Once()

	suite.uploadManager.AssertNotCalled(suite.T(), "PublicURL")

	response, total, err := suite.productService.ListProducts(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(0), total)
	suite.Empty(response)
	suite.Len(response, 0)
}

func (suite *ProductServiceTestSuite) TestListProducts_RepositoryError() {
	req := dto.ListProductRequest{
		Limit:  10,
		Offset: 0,
	}

	repoErr := errors.New("query execution failed")
	suite.productRepo.EXPECT().ListProducts(suite.ctx, req).
		Return(nil, 0, repoErr).Once()

	response, total, err := suite.productService.ListProducts(suite.ctx, req)

	suite.Nil(response)
	suite.Equal(int64(0), total)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== CreateProduct Tests ====================

func (suite *ProductServiceTestSuite) TestCreateProduct_Success_DefaultIsActive() {
	req := dto.ProductCreateRequest{Name: "Product 1"}

	suite.productRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(product *models.Product) bool {
		return product.IsActive
	})).Return(nil).Once()

	response, err := suite.productService.CreateProduct(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.True(response.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_Success_IsActiveFalse() {
	req := dto.ProductCreateRequest{Name: "Product 1", IsActive: new(false)}

	suite.productRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(product *models.Product) bool {
		return !product.IsActive
	})).Return(nil).Once()

	response, err := suite.productService.CreateProduct(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.False(response.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_RepositoryError() {
	req := dto.ProductCreateRequest{}

	repoErr := errors.New("query execution failed")
	suite.productRepo.EXPECT().Create(suite.ctx, mock.Anything).
		Return(repoErr).Once()

	response, err := suite.productService.CreateProduct(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== UpdateProduct Tests ====================

func (suite *ProductServiceTestSuite) TestUpdateProduct_Success_UpdateAllFields() {
	name := "Updated Product"
	description := "updated description"
	price := int64(999_900)
	stock := 20

	product := &models.Product{
		ID:          suite.productID,
		Name:        "Old Product",
		Slug:        "old-product",
		Description: new("old description"),
		Price:       50,
		Stock:       5,
		Reserved:    3,
		IsActive:    false,
	}

	req := dto.ProductUpdateRequest{
		Name:        &name,
		Description: &description,
		Price:       &price,
		Stock:       &stock,
		IsActive:    new(true),
	}

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	suite.productRepo.EXPECT().Update(suite.ctx, product).
		Return(nil).Once()

	response, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.NoError(err)
	suite.NotNil(response)

	suite.Equal(suite.productID, response.ID)
	suite.Equal(name, response.Name)
	suite.Equal(utils.Slugify(name), response.Slug)
	suite.Equal(&description, response.Description)
	suite.Equal(price, response.Price)
	suite.Equal(stock, response.Stock)
	suite.True(response.IsActive)
	suite.Equal(17, product.Available())
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_Success_PartialUpdate() {
	name := "New Name"

	product := &models.Product{
		ID:          suite.productID,
		Name:        "Old Name",
		Slug:        "old-name",
		Description: new("description"),
		Price:       10_000,
		Stock:       3,
		Reserved:    1,
		IsActive:    true,
	}

	req := dto.ProductUpdateRequest{
		Name: &name,
	}

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	suite.productRepo.EXPECT().Update(suite.ctx, product).
		Return(nil).Once()

	response, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.NoError(err)
	suite.NotNil(response)

	suite.Equal(name, response.Name)
	suite.Equal(utils.Slugify(name), response.Slug)

	suite.Equal(product.Price, response.Price)
	suite.Equal(product.Stock, response.Stock)
	suite.Equal(product.Description, response.Description)
	suite.True(response.IsActive)
	suite.Equal(2, product.Available())
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_StockLessThanReserved() {
	product := &models.Product{
		ID:          suite.productID,
		Name:        "Test Product",
		Slug:        "test-product",
		Description: new("description"),
		Price:       10_000,
		Stock:       10,
		Reserved:    5,
		IsActive:    true,
	}

	req := dto.ProductUpdateRequest{
		Stock: new(2),
	}

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	response, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrStockLessThanReserved)

	suite.Equal(10, product.Stock)
	suite.Equal(5, product.Reserved)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_StockEqualToReserved() {
	newStock := 5

	product := &models.Product{
		ID:          suite.productID,
		Name:        "Test Product",
		Slug:        "test-product",
		Description: new("description"),
		Price:       10_000,
		Stock:       10,
		Reserved:    5,
		IsActive:    true,
	}

	req := dto.ProductUpdateRequest{
		Stock: &newStock,
	}

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(product, nil).Once()

	suite.productRepo.EXPECT().Update(suite.ctx, product).
		Return(nil).Once()

	response, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(newStock, response.Stock)
	suite.Equal(0, product.Available())
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_ProductNotFound() {
	req := dto.ProductUpdateRequest{}

	suite.productRepo.
		EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(nil, repository.ErrRecordNotFound).
		Once()

	response, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_UpdateRepositoryError() {
	existingProduct := &models.Product{ID: suite.productID}
	req := dto.ProductUpdateRequest{}

	dbErr := errors.New("db error")

	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(existingProduct, nil).Once()

	suite.productRepo.EXPECT().Update(suite.ctx, existingProduct).
		Return(dbErr).Once()

	response, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.Nil(response)
	suite.ErrorIs(err, dbErr)
}

// ==================== UploadImage Tests ====================

func (suite *ProductServiceTestSuite) TestUploadImage_Success() {
	req := dto.UploadProductImageRequest{
		ContentType: "image/jpeg",
		Ext:         "jpg",
	}

	signRequest := dto.UploadSignURLRequest{
		ContentType: req.ContentType,
		Entity:      dto.NewUploadEntity(suite.productID, string(models.EntityTypeProduct)),
		Ext:         req.Ext,
	}

	uploadID := uuid.New()

	suite.productRepo.EXPECT().Exists(suite.ctx, suite.productID).
		Return(true, nil).Once()

	suite.uploadManager.EXPECT().SignURL(suite.ctx, signRequest, ProductImageType).
		Return(&dto.UploadSignURLResponse{UploadID: uploadID}, nil).Once()

	response, err := suite.productService.UploadImage(suite.ctx, suite.productID, req)

	suite.NotNil(response)
	suite.NoError(err)

	suite.Equal(uploadID, response.UploadID)
}

func (suite *ProductServiceTestSuite) TestUploadImage_ProductNotFound() {
	req := dto.UploadProductImageRequest{
		ContentType: "image/jpeg",
		Ext:         "jpg",
	}

	suite.productRepo.EXPECT().Exists(suite.ctx, suite.productID).
		Return(false, nil).Once()

	response, err := suite.productService.UploadImage(suite.ctx, suite.productID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
}

// ==================== ConfirmUploadImage Tests ====================

func (suite *ProductServiceTestSuite) TestConfirmUploadImage_Success() {
	req := dto.ConfirmUploadProductImageRequest{
		UploadID:  uuid.New(),
		ObjectKey: "key",
	}

	url := "https://example.com/"

	saveRequest := dto.UploadSaveRequest{
		UploadID:  req.UploadID,
		ObjectKey: req.ObjectKey,
		Entity:    dto.NewUploadEntity(suite.productID, string(models.EntityTypeProduct)),
		IsMain:    false,
	}

	suite.productRepo.EXPECT().Exists(suite.ctx, suite.productID).
		Return(true, nil).Once()

	suite.uploadManager.EXPECT().Save(suite.ctx, saveRequest, ProductImageType).
		Return(&dto.UploadResponse{URL: url}, nil).Once()

	response, err := suite.productService.ConfirmUploadImage(suite.ctx, suite.productID, req)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(url, response.URL)
}

func (suite *ProductServiceTestSuite) TestConfirmUploadImage_ProductNotFound() {
	req := dto.ConfirmUploadProductImageRequest{
		UploadID:  uuid.New(),
		ObjectKey: "key",
	}

	suite.productRepo.EXPECT().Exists(suite.ctx, suite.productID).
		Return(false, nil).Once()

	response, err := suite.productService.ConfirmUploadImage(suite.ctx, suite.productID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
}
