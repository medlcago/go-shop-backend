package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/internal/upload"
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
	mockProduct := &models.Product{
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
		Return(mockProduct, nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.Anything).
		Return(uuid.NewString()).Times(len(mockProduct.Images))

	product, err := suite.productService.GetProductByID(suite.ctx, suite.productID)

	suite.NoError(err)
	suite.NotNil(product)
	suite.Equal(suite.productID, product.ID)
	suite.Equal("Test Product", product.Name)
	suite.Equal(int64(10_000), product.Price)
	suite.Len(product.Categories, 2)
	suite.Equal(mockProduct.Categories[0].ID, product.Categories[0].ID)
	suite.Equal(mockProduct.Categories[1].ID, product.Categories[1].ID)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_NotFound() {
	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	product, err := suite.productService.GetProductByID(suite.ctx, suite.productID)

	suite.Nil(product)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
}

func (suite *ProductServiceTestSuite) TestGetProductByID_RepositoryError() {
	repoErr := errors.New("database connection failed")
	suite.productRepo.EXPECT().GetByID(suite.ctx, suite.productID, true).
		Return(nil, repoErr).Once()

	product, err := suite.productService.GetProductByID(suite.ctx, suite.productID)

	suite.Nil(product)
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

	mockProducts := []*models.Product{
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
		Return(mockProducts, 2, nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.Anything).
		Return(uuid.NewString()).Times(len(mockProducts[0].Images) + len(mockProducts[1].Images))

	products, totalCount, err := suite.productService.ListProducts(suite.ctx, req)

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
	req := dto.ListProductRequest{
		Limit:      20,
		Offset:     0,
		CategoryID: suite.categoryID,
		OrderBy:    "created_at",
		OrderDesc:  true,
	}

	mockProducts := []*models.Product{
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
		Return(mockProducts, 1, nil).Once()

	suite.uploadManager.EXPECT().PublicURL(mock.Anything).
		Return(uuid.NewString()).Times(len(mockProducts[0].Images))

	products, totalCount, err := suite.productService.ListProducts(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(1), totalCount)
	suite.Len(products, 1)
	suite.Equal("Category Product", products[0].Name)
	suite.Len(products[0].Categories, 1)
}

func (suite *ProductServiceTestSuite) TestListProducts_EmptyResult() {
	req := dto.ListProductRequest{
		Limit:  10,
		Offset: 1000,
	}

	suite.productRepo.EXPECT().ListProducts(suite.ctx, req).
		Return([]*models.Product{}, 0, nil).Once()

	suite.uploadManager.AssertNotCalled(suite.T(), "PublicURL")

	products, totalCount, err := suite.productService.ListProducts(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(0), totalCount)
	suite.Empty(products)
	suite.Len(products, 0)
}

func (suite *ProductServiceTestSuite) TestListProducts_RepositoryError() {
	req := dto.ListProductRequest{
		Limit:  10,
		Offset: 0,
	}

	repoErr := errors.New("query execution failed")
	suite.productRepo.EXPECT().ListProducts(suite.ctx, req).
		Return(nil, 0, repoErr).Once()

	products, totalCount, err := suite.productService.ListProducts(suite.ctx, req)

	suite.Nil(products)
	suite.Equal(int64(0), totalCount)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== CreateProduct Tests ====================

func (suite *ProductServiceTestSuite) TestCreateProduct_Success_DefaultIsActive() {
	req := dto.ProductCreateRequest{Name: "Product 1"}

	suite.productRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(product *models.Product) bool {
		return product.IsActive
	})).Return(nil).Once()

	product, err := suite.productService.CreateProduct(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(product)
	suite.True(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_Success_IsActiveFalse() {
	req := dto.ProductCreateRequest{Name: "Product 1", IsActive: new(false)}

	suite.productRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(product *models.Product) bool {
		return !product.IsActive
	})).Return(nil).Once()

	product, err := suite.productService.CreateProduct(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(product)
	suite.False(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestCreateProduct_RepositoryError() {
	req := dto.ProductCreateRequest{}

	repoErr := errors.New("query execution failed")
	suite.productRepo.EXPECT().Create(suite.ctx, mock.Anything).
		Return(repoErr).Once()

	product, err := suite.productService.CreateProduct(suite.ctx, req)

	suite.Nil(product)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== UpdateProduct Tests ====================

func (suite *ProductServiceTestSuite) TestUpdateProduct_Success_UpdateAllFields() {
	name := "Updated Product"
	description := "updated description"
	price := int64(999_900)
	stock := 20

	existingProduct := &models.Product{
		ID:          suite.productID,
		Name:        "Old Product",
		Slug:        "old-product",
		Description: new("old description"),
		Price:       50,
		Stock:       5,
		IsActive:    false,
	}

	req := dto.ProductUpdateRequest{
		Name:        &name,
		Description: &description,
		Price:       &price,
		Stock:       &stock,
		IsActive:    new(true),
	}

	suite.productRepo.
		EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		EXPECT().Update(suite.ctx, existingProduct).
		Return(nil).
		Once()

	product, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.NoError(err)
	suite.NotNil(product)

	suite.Equal(suite.productID, product.ID)
	suite.Equal(name, product.Name)
	suite.Equal(utils.Slugify(name), product.Slug)
	suite.Equal(&description, product.Description)
	suite.Equal(price, product.Price)
	suite.Equal(stock, product.Stock)
	suite.True(product.IsActive)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_Success_PartialUpdate() {
	name := "New Name"

	existingProduct := &models.Product{
		ID:          suite.productID,
		Name:        "Old Name",
		Slug:        "old-name",
		Description: new("description"),
		Price:       10_000,
		Stock:       3,
		IsActive:    true,
	}

	req := dto.ProductUpdateRequest{
		Name: &name,
	}

	suite.productRepo.
		EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		EXPECT().Update(suite.ctx, existingProduct).
		Return(nil).
		Once()

	product, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

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
	req := dto.ProductUpdateRequest{}

	suite.productRepo.
		EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(nil, repository.ErrRecordNotFound).
		Once()

	product, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.Nil(product)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
}

func (suite *ProductServiceTestSuite) TestUpdateProduct_UpdateRepositoryError() {
	existingProduct := &models.Product{ID: suite.productID}
	req := dto.ProductUpdateRequest{}

	repoErr := errors.New("update failed")

	suite.productRepo.
		EXPECT().GetByID(suite.ctx, suite.productID, false).
		Return(existingProduct, nil).
		Once()

	suite.productRepo.
		EXPECT().Update(suite.ctx, existingProduct).
		Return(repoErr).
		Once()

	product, err := suite.productService.UpdateProduct(suite.ctx, suite.productID, req)

	suite.Nil(product)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== UploadImage Tests ====================

func (suite *ProductServiceTestSuite) TestUploadImage_Success() {
	req := dto.UploadProductImageRequest{
		ContentType: "image/jpeg",
		Ext:         "jpg",
	}

	signUrlReq := upload.SignURLRequest{
		ContentType: req.ContentType,
		Entity:      upload.NewProductEntity(suite.productID),
		Ext:         req.Ext,
	}

	uploadID := uuid.New()

	suite.productRepo.EXPECT().Exists(suite.ctx, suite.productID).
		Return(true, nil).Once()

	suite.uploadManager.EXPECT().SignURL(suite.ctx, signUrlReq, upload.ProductImagePolicy).
		Return(&upload.SignURLResponse{UploadID: uploadID}, nil).Once()

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

	saveUploadReq := upload.SaveUploadRequest{
		UploadID:  req.UploadID,
		ObjectKey: req.ObjectKey,
		Entity:    upload.NewProductEntity(suite.productID),
		IsMain:    false,
	}

	suite.productRepo.EXPECT().Exists(suite.ctx, suite.productID).
		Return(true, nil).Once()

	suite.uploadManager.EXPECT().Save(suite.ctx, saveUploadReq, upload.ProductImagePolicy).
		Return(&upload.ContentResponse{URL: url}, nil).Once()

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
