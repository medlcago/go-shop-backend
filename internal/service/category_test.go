package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type CategoryServiceTestSuite struct {
	suite.Suite
	categoryRepo    *repoMocks.MockCategoryRepository
	categoryService *categoryService

	ctx      context.Context
	parentID uuid.UUID
}

func (suite *CategoryServiceTestSuite) SetupTest() {
	suite.categoryRepo = repoMocks.NewMockCategoryRepository(suite.T())
	suite.categoryService = NewCategoryService(suite.categoryRepo)

	suite.ctx = context.Background()
	suite.parentID = uuid.New()
}

func TestCategoryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryServiceTestSuite))
}

// ==================== ListCategories Tests ====================

func (suite *CategoryServiceTestSuite) TestListCategories_Success_RootCategories() {
	req := dto.ListCategoryRequest{
		Limit:  2,
		Offset: 0,
	}

	mockCategories := []*models.Category{
		{
			Name: "Electronics",
		},
		{
			Name: "Clothing",
		},
	}

	suite.categoryRepo.EXPECT().ListCategories(suite.ctx, req).
		Return(mockCategories, 5, nil).Once()

	categories, totalCount, err := suite.categoryService.ListCategories(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(5), totalCount)
	suite.Len(categories, 2)

	suite.Equal("Electronics", categories[0].Name)
	suite.Equal("Clothing", categories[1].Name)
}

func (suite *CategoryServiceTestSuite) TestListCategories_Success_Subcategories() {
	req := dto.ListCategoryRequest{
		ID:     suite.parentID,
		Limit:  5,
		Offset: 0,
	}

	mockCategories := []*models.Category{
		{
			Name:     "Laptops",
			ParentID: new(suite.parentID),
		},
		{
			Name:     "Smartphones",
			ParentID: new(suite.parentID),
		},
	}

	suite.categoryRepo.EXPECT().ListCategories(suite.ctx, req).
		Return(mockCategories, 2, nil).Once()

	categories, totalCount, err := suite.categoryService.ListCategories(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(2), totalCount)
	suite.Len(categories, 2)

	suite.Equal(suite.parentID, *categories[0].ParentID)
	suite.Equal(suite.parentID, *categories[1].ParentID)
}

func (suite *CategoryServiceTestSuite) TestListCategories_RepositoryError() {
	req := dto.ListCategoryRequest{
		Limit:  10,
		Offset: 0,
	}

	repoErr := errors.New("database connection failed")
	suite.categoryRepo.EXPECT().ListCategories(suite.ctx, req).
		Return(nil, 0, repoErr).Once()

	categories, totalCount, err := suite.categoryService.ListCategories(suite.ctx, req)

	suite.Nil(categories)
	suite.Equal(int64(0), totalCount)
	suite.ErrorContains(err, repoErr.Error())
}

func (suite *CategoryServiceTestSuite) TestListCategories_EmptyResult() {
	req := dto.ListCategoryRequest{
		Limit:  10,
		Offset: 100,
	}

	suite.categoryRepo.EXPECT().ListCategories(suite.ctx, req).
		Return([]*models.Category{}, 0, nil).Once()

	categories, totalCount, err := suite.categoryService.ListCategories(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(0), totalCount)
	suite.Empty(categories)
	suite.Len(categories, 0)
}
