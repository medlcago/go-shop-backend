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

	categories := []*models.Category{
		{
			Name: "Electronics",
		},
		{
			Name: "Clothing",
		},
	}

	suite.categoryRepo.EXPECT().ListCategories(suite.ctx, req).
		Return(categories, 5, nil).Once()

	response, total, err := suite.categoryService.ListCategories(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(5), total)
	suite.Len(response, 2)

	suite.Equal("Electronics", response[0].Name)
	suite.True(response[0].IsRoot)
	suite.Equal("Clothing", response[1].Name)
	suite.True(response[1].IsRoot)
}

func (suite *CategoryServiceTestSuite) TestListCategories_Success_Subcategories() {
	req := dto.ListCategoryRequest{
		ID:     suite.parentID,
		Limit:  5,
		Offset: 0,
	}

	categories := []*models.Category{
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
		Return(categories, 2, nil).Once()

	response, total, err := suite.categoryService.ListCategories(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(2), total)
	suite.Len(response, 2)

	suite.Equal(suite.parentID, *response[0].ParentID)
	suite.False(response[0].IsRoot)
	suite.Equal(suite.parentID, *response[1].ParentID)
	suite.False(response[1].IsRoot)
}

func (suite *CategoryServiceTestSuite) TestListCategories_RepositoryError() {
	req := dto.ListCategoryRequest{
		Limit:  10,
		Offset: 0,
	}

	repoErr := errors.New("database connection failed")
	suite.categoryRepo.EXPECT().ListCategories(suite.ctx, req).
		Return(nil, 0, repoErr).Once()

	response, total, err := suite.categoryService.ListCategories(suite.ctx, req)

	suite.Nil(response)
	suite.Equal(int64(0), total)
	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "categoryService.ListCategories")
}

func (suite *CategoryServiceTestSuite) TestListCategories_EmptyResult() {
	req := dto.ListCategoryRequest{
		Limit:  10,
		Offset: 100,
	}

	suite.categoryRepo.EXPECT().ListCategories(suite.ctx, req).
		Return([]*models.Category{}, 0, nil).Once()

	response, total, err := suite.categoryService.ListCategories(suite.ctx, req)

	suite.NoError(err)
	suite.Equal(int64(0), total)
	suite.Empty(response)
}
