package service

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type CategoryServiceTestSuite struct {
	suite.Suite
	mockRepo *mocks.CategoryRepositoryMock
	service  CategoryService
}

func (suite *CategoryServiceTestSuite) SetupTest() {
	suite.mockRepo = new(mocks.CategoryRepositoryMock)
	suite.service = NewCategoryService(suite.mockRepo)
}

func (suite *CategoryServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

func TestCategoryServiceTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryServiceTestSuite))
}

// ==================== ListCategories Tests ====================

func (suite *CategoryServiceTestSuite) TestListCategories_Success_RootCategories() {
	ctx := context.Background()
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

	suite.mockRepo.On("ListCategories", ctx, req).
		Return(mockCategories, 5, nil).Once()

	categories, totalCount, err := suite.service.ListCategories(ctx, req)

	suite.NoError(err)
	suite.Equal(int64(5), totalCount)
	suite.Len(categories, 2)

	suite.Equal("Electronics", categories[0].Name)

	suite.Equal("Clothing", categories[1].Name)
}

func (suite *CategoryServiceTestSuite) TestListCategories_Success_Subcategories() {
	ctx := context.Background()
	parentID := uuid.New()
	req := dto.ListCategoryRequest{
		ID:     parentID,
		Limit:  5,
		Offset: 0,
	}

	parentUUID := sql.Null[uuid.UUID]{V: parentID, Valid: true}
	mockCategories := []*models.Category{
		{
			Name:     "Laptops",
			ParentID: parentUUID,
		},
		{
			Name:     "Smartphones",
			ParentID: parentUUID,
		},
	}

	suite.mockRepo.On("ListCategories", ctx, req).
		Return(mockCategories, 2, nil).Once()

	categories, totalCount, err := suite.service.ListCategories(ctx, req)

	suite.NoError(err)
	suite.Equal(int64(2), totalCount)
	suite.Len(categories, 2)

	suite.Equal(parentID.String(), categories[0].ParentID)
	suite.Equal(parentID.String(), categories[1].ParentID)
}

func (suite *CategoryServiceTestSuite) TestListCategories_RepositoryError() {
	ctx := context.Background()
	req := dto.ListCategoryRequest{
		Limit:  10,
		Offset: 0,
	}

	repoErr := errors.New("database connection failed")
	suite.mockRepo.On("ListCategories", ctx, req).
		Return([]*models.Category{}, 0, repoErr).Once()

	categories, totalCount, err := suite.service.ListCategories(ctx, req)

	suite.Nil(categories)
	suite.Equal(int64(0), totalCount)
	suite.ErrorContains(err, repoErr.Error())
}

func (suite *CategoryServiceTestSuite) TestListCategories_EmptyResult() {
	ctx := context.Background()
	req := dto.ListCategoryRequest{
		Limit:  10,
		Offset: 100,
	}

	suite.mockRepo.On("ListCategories", ctx, req).
		Return([]*models.Category{}, 0, nil).Once()

	categories, totalCount, err := suite.service.ListCategories(ctx, req)

	suite.NoError(err)
	suite.Equal(int64(0), totalCount)
	suite.Empty(categories)
	suite.Len(categories, 0)
}
