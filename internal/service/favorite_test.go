package service

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	serviceMocks "go-shop-backend/internal/service/mocks"
	"go-shop-backend/pkg/apperror"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type FavoriteServiceTestSuite struct {
	suite.Suite

	productChecker  *serviceMocks.MockProductChecker
	favoriteRepo    *repoMocks.MockFavoriteRepository
	favoriteService *favoriteService

	ctx       context.Context
	userID    uuid.UUID
	productID uuid.UUID
}

func (suite *FavoriteServiceTestSuite) SetupTest() {
	suite.productChecker = serviceMocks.NewMockProductChecker(suite.T())
	suite.favoriteRepo = repoMocks.NewMockFavoriteRepository(suite.T())
	suite.favoriteService = NewFavoriteService(
		suite.productChecker,
		suite.favoriteRepo,
	)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
	suite.productID = uuid.New()

}

func TestFavoriteServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FavoriteServiceTestSuite))
}

// ==================== AddProduct Tests ====================

func (suite *FavoriteServiceTestSuite) TestAddProduct_Success() {
	favorite := &models.Favorite{
		UserID:    suite.userID,
		ProductID: suite.productID,
	}

	favoriteProducts := []*models.FavoriteProductProjection{{ID: uuid.New()}, {ID: uuid.New()}}

	suite.productChecker.EXPECT().Exists(suite.ctx, suite.productID).
		Return(true, nil).Once()

	suite.favoriteRepo.EXPECT().Add(suite.ctx, favorite).
		Return(nil).Once()

	suite.favoriteRepo.EXPECT().GetListByUser(suite.ctx, suite.userID).
		Return(favoriteProducts, 2, nil).Once()

	response, total, err := suite.favoriteService.AddProduct(suite.ctx, suite.userID, suite.productID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(int64(2), total)
	suite.Len(response, 2)
	suite.Equal(favoriteProducts[0].ID, response[0].ID)
	suite.Equal(favoriteProducts[1].ID, response[1].ID)
}

func (suite *FavoriteServiceTestSuite) TestAddProduct_ProductNotFound() {
	suite.productChecker.EXPECT().Exists(suite.ctx, suite.productID).
		Return(false, nil).Once()

	response, total, err := suite.favoriteService.AddProduct(suite.ctx, suite.userID, suite.productID)

	suite.Nil(response)
	suite.Equal(int64(0), total)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
	suite.ErrorContains(err, "favoriteService.AddProduct")
}

// ==================== RemoveProduct Tests ====================

func (suite *FavoriteServiceTestSuite) TestRemoveProduct_Success() {
	favoriteProducts := []*models.FavoriteProductProjection{{ID: uuid.New()}, {ID: uuid.New()}}

	suite.favoriteRepo.EXPECT().Remove(suite.ctx, suite.productID, suite.userID).
		Return(nil).Once()

	suite.favoriteRepo.EXPECT().GetListByUser(suite.ctx, suite.userID).
		Return(favoriteProducts, 2, nil).Once()

	response, total, err := suite.favoriteService.RemoveProduct(suite.ctx, suite.userID, suite.productID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(int64(2), total)
	suite.Len(response, 2)
	suite.Equal(favoriteProducts[0].ID, response[0].ID)
	suite.Equal(favoriteProducts[1].ID, response[1].ID)
}

func (suite *FavoriteServiceTestSuite) TestRemoveProduct_ProductNotFound() {
	suite.favoriteRepo.EXPECT().Remove(suite.ctx, suite.productID, suite.userID).
		Return(repository.ErrRecordNotFound).Once()

	response, total, err := suite.favoriteService.RemoveProduct(suite.ctx, suite.userID, suite.productID)

	suite.Nil(response)
	suite.Equal(int64(0), total)
	suite.ErrorIs(err, apperror.ErrProductNotFound)
	suite.ErrorContains(err, "favoriteService.RemoveProduct")
}

// ==================== List Tests ====================

func (suite *FavoriteServiceTestSuite) TestList_Success() {
	favoriteProducts := []*models.FavoriteProductProjection{{ID: uuid.New()}, {ID: uuid.New()}}

	suite.favoriteRepo.EXPECT().GetListByUser(suite.ctx, suite.userID).
		Return(favoriteProducts, 2, nil).Once()

	response, total, err := suite.favoriteService.List(suite.ctx, suite.userID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(int64(2), total)
	suite.Len(response, 2)
	suite.Equal(favoriteProducts[0].ID, response[0].ID)
	suite.Equal(favoriteProducts[1].ID, response[1].ID)
}

// ==================== IsFavorite Tests ====================

func (suite *FavoriteServiceTestSuite) TestIsFavorite_Success() {
	suite.favoriteRepo.EXPECT().Exists(suite.ctx, suite.productID, suite.userID).
		Return(false, nil).Once()

	response, err := suite.favoriteService.IsFavorite(suite.ctx, suite.userID, suite.productID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(suite.productID, response.ProductID)
	suite.False(response.IsFavorite)
}
