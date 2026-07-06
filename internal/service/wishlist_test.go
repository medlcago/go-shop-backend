package service

import (
	"context"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperror"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type WishlistServiceTestSuite struct {
	suite.Suite
	wishlistRepo     *repoMocks.MockWishlistRepository
	wishlistItemRepo *repoMocks.MockWishlistItemRepository
	productRepo      *repoMocks.MockProductRepository
	wishlistService  *wishlistService

	ctx        context.Context
	userID     uuid.UUID
	wishlistID uuid.UUID
	itemID     uuid.UUID
	shareToken string
}

func (suite *WishlistServiceTestSuite) SetupTest() {
	suite.wishlistRepo = repoMocks.NewMockWishlistRepository(suite.T())
	suite.wishlistItemRepo = repoMocks.NewMockWishlistItemRepository(suite.T())
	suite.productRepo = repoMocks.NewMockProductRepository(suite.T())
	suite.wishlistService = NewWishlistService(
		suite.wishlistRepo,
		suite.wishlistItemRepo,
		suite.productRepo,
	)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
	suite.wishlistID = uuid.New()
	suite.itemID = uuid.New()
	suite.shareToken = uuid.NewString()
}

func TestWishlistServiceTestSuite(t *testing.T) {
	suite.Run(t, new(WishlistServiceTestSuite))
}

// ==================== CreateWishlist Tests ====================

func (suite *WishlistServiceTestSuite) TestCreateWishlist_Success() {
	req := dto.CreateWishlistRequest{
		Title:    "wishlist",
		IsPublic: true,
	}

	suite.wishlistRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(wishlist *models.Wishlist) bool {
		return wishlist.UserID == suite.userID &&
			wishlist.Title == req.Title &&
			wishlist.IsPublic == req.IsPublic &&
			wishlist.ShareToken != ""
	})).Return(nil).Once()

	response, err := suite.wishlistService.CreateWishlist(suite.ctx, suite.userID, req)

	suite.NoError(err)
	suite.NotNil(response)
}

func (suite *WishlistServiceTestSuite) TestCreateWishlist_RepositoryError() {
	req := dto.CreateWishlistRequest{
		Title:    "wishlist",
		IsPublic: true,
	}

	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().Create(suite.ctx, mock.AnythingOfType("*models.Wishlist")).
		Return(dbErr).Once()

	response, err := suite.wishlistService.CreateWishlist(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.CreateWishlist")
	suite.ErrorIs(err, dbErr)
}

// ==================== GetWishlist Tests ====================

func (suite *WishlistServiceTestSuite) TestGetWishlist_Success() {
	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
		Title:  "wishlist",
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.GetWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(response.ID, wishlist.ID)
}

func (suite *WishlistServiceTestSuite) TestGetWishlist_WishlistNotFound() {
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.GetWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.GetWishlist")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestGetWishlist_PublicWishlist_UserIsNotOwner() {
	wishlist := &models.Wishlist{
		ID:       suite.wishlistID,
		UserID:   uuid.New(),
		Title:    "public wishlist",
		IsPublic: true,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.GetWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(response.ID, wishlist.ID)
}

func (suite *WishlistServiceTestSuite) TestGetWishlist_PublicWishlist_UserIsOwner() {
	wishlist := &models.Wishlist{
		ID:       suite.wishlistID,
		UserID:   suite.userID,
		Title:    "public wishlist",
		IsPublic: true,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.GetWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(response.ID, wishlist.ID)
}

func (suite *WishlistServiceTestSuite) TestGetWishlist_PrivateWishlist_UserIsNotOwner() {
	wishlist := &models.Wishlist{
		ID:       suite.wishlistID,
		UserID:   uuid.New(),
		Title:    "private wishlist",
		IsPublic: false,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.GetWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.GetWishlist")
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *WishlistServiceTestSuite) TestGetWishlist_PrivateWishlist_UserIsOwner() {
	wishlist := &models.Wishlist{
		ID:       suite.wishlistID,
		UserID:   suite.userID,
		Title:    "private wishlist",
		IsPublic: false,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.GetWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.NoError(err)
	suite.NotNil(response)
}

func (suite *WishlistServiceTestSuite) TestGetWishlist_RepositoryError() {
	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, dbErr).Once()

	response, err := suite.wishlistService.GetWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.GetWishlist")
	suite.ErrorIs(err, dbErr)
}

// ==================== GetWishlists Tests ====================

func (suite *WishlistServiceTestSuite) TestGetWishlists_Success() {
	req := dto.ListWishlistRequest{}
	wishlists := []*models.Wishlist{{ID: uuid.New()}, {ID: uuid.New()}}

	suite.wishlistRepo.EXPECT().GetListByUser(suite.ctx, suite.userID, req).
		Return(wishlists, 3, nil).Once()

	response, total, err := suite.wishlistService.GetWishlists(suite.ctx, suite.userID, req)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(int64(3), total)
	suite.Len(response, 2)
}

func (suite *WishlistServiceTestSuite) TestGetWishlists_EmptyList() {
	req := dto.ListWishlistRequest{}

	suite.wishlistRepo.EXPECT().GetListByUser(suite.ctx, suite.userID, req).
		Return([]*models.Wishlist{}, 0, nil).Once()

	response, total, err := suite.wishlistService.GetWishlists(suite.ctx, suite.userID, req)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(int64(0), total)
	suite.Len(response, 0)
}

func (suite *WishlistServiceTestSuite) TestGetWishlists_RepositoryError() {
	req := dto.ListWishlistRequest{}
	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetListByUser(suite.ctx, suite.userID, req).
		Return([]*models.Wishlist{}, 0, dbErr).Once()

	response, total, err := suite.wishlistService.GetWishlists(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.Equal(int64(0), total)
	suite.ErrorContains(err, "wishlistService.GetWishlists")
	suite.ErrorIs(err, dbErr)
}

// ==================== GetSharedWishlist Tests ====================

func (suite *WishlistServiceTestSuite) TestGetSharedWishlist_Success() {
	wishlist := &models.Wishlist{
		ID:         suite.wishlistID,
		ShareToken: suite.shareToken,
		IsPublic:   true,
	}

	suite.wishlistRepo.EXPECT().GetByShareToken(suite.ctx, suite.shareToken, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.GetSharedWishlist(suite.ctx, suite.shareToken)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(response.ID, wishlist.ID)
}

func (suite *WishlistServiceTestSuite) TestGetSharedWishlist_PrivateWishlist() {
	wishlist := &models.Wishlist{
		ID:         suite.wishlistID,
		ShareToken: suite.shareToken,
		IsPublic:   false,
	}

	suite.wishlistRepo.EXPECT().GetByShareToken(suite.ctx, suite.shareToken, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.GetSharedWishlist(suite.ctx, suite.shareToken)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.GetSharedWishlist")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestGetSharedWishlist_WishlistNotFound() {
	suite.wishlistRepo.EXPECT().GetByShareToken(suite.ctx, suite.shareToken, true).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.GetSharedWishlist(suite.ctx, suite.shareToken)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.GetSharedWishlist")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestGetSharedWishlist_RepositoryError() {
	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetByShareToken(suite.ctx, suite.shareToken, true).
		Return(nil, dbErr).Once()

	response, err := suite.wishlistService.GetSharedWishlist(suite.ctx, suite.shareToken)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.GetSharedWishlist")
	suite.ErrorIs(err, dbErr)
}

// ==================== UpdateWishlist Tests ====================

func (suite *WishlistServiceTestSuite) TestUpdateWishlist_Success() {
	req := dto.UpdateWishlistRequest{
		Title:    new("new title"),
		IsPublic: new(true),
	}

	wishlist := &models.Wishlist{
		ID:       suite.wishlistID,
		UserID:   suite.userID,
		Title:    "title",
		IsPublic: false,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistRepo.EXPECT().Update(suite.ctx, wishlist).
		Return(nil).Once()

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.UpdateWishlist(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.NotNil(response)
	suite.NoError(err)
	suite.Equal(response.ID, wishlist.ID)
	suite.Equal("new title", wishlist.Title)
	suite.True(wishlist.IsPublic)
}

func (suite *WishlistServiceTestSuite) TestUpdateWishlist_WishlistNotFound() {
	req := dto.UpdateWishlistRequest{}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.UpdateWishlist(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.UpdateWishlist")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestUpdateWishlist_UserIsNotOwner() {
	req := dto.UpdateWishlistRequest{}

	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: uuid.New(),
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.UpdateWishlist(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.UpdateWishlist")
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *WishlistServiceTestSuite) TestUpdateWishlist_RepositoryError() {
	req := dto.UpdateWishlistRequest{}

	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, dbErr).Once()

	response, err := suite.wishlistService.UpdateWishlist(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.UpdateWishlist")
	suite.ErrorIs(err, dbErr)
}

// ==================== DeleteWishlist Tests ====================

func (suite *WishlistServiceTestSuite) TestDeleteWishlist_Success() {
	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistRepo.EXPECT().Delete(suite.ctx, wishlist.ID).
		Return(nil).Once()

	err := suite.wishlistService.DeleteWishlist(suite.ctx, suite.userID, suite.wishlistID)
	suite.NoError(err)
}

func (suite *WishlistServiceTestSuite) TestDeleteWishlist_WishlistNotFound() {
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	err := suite.wishlistService.DeleteWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.ErrorContains(err, "wishlistService.DeleteWishlist")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestDeleteWishlist_UserIsNotOwner() {
	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: uuid.New(),
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	err := suite.wishlistService.DeleteWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.ErrorContains(err, "wishlistService.DeleteWishlist")
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *WishlistServiceTestSuite) TestDeleteWishlist_RepositoryError() {
	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, dbErr).Once()

	err := suite.wishlistService.DeleteWishlist(suite.ctx, suite.userID, suite.wishlistID)

	suite.ErrorContains(err, "wishlistService.DeleteWishlist")
	suite.ErrorIs(err, dbErr)
}

// ==================== RegenerateShareToken Tests ====================

func (suite *WishlistServiceTestSuite) TestRegenerateShareToken_Success() {
	wishlist := &models.Wishlist{
		ID:         suite.wishlistID,
		UserID:     suite.userID,
		ShareToken: suite.shareToken,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistRepo.EXPECT().Update(suite.ctx, wishlist).
		Return(nil).Once()

	response, err := suite.wishlistService.RegenerateShareToken(suite.ctx, suite.userID, suite.wishlistID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(wishlist.ShareToken, response.ShareToken)
}

func (suite *WishlistServiceTestSuite) TestRegenerateShareToken_WishlistNotFound() {
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.RegenerateShareToken(suite.ctx, suite.userID, suite.wishlistID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.RegenerateShareToken")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestRegenerateShareToken_UserIsNotOwner() {
	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: uuid.New(),
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.RegenerateShareToken(suite.ctx, suite.userID, suite.wishlistID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.RegenerateShareToken")
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *WishlistServiceTestSuite) TestRegenerateShareToken_RepositoryError() {
	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, dbErr).Once()

	response, err := suite.wishlistService.RegenerateShareToken(suite.ctx, suite.userID, suite.wishlistID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.RegenerateShareToken")
	suite.ErrorIs(err, dbErr)
}

// ==================== AddItem Tests ====================

func (suite *WishlistServiceTestSuite) TestAddItem_Success() {
	req := dto.AddWishlistItemRequest{
		ProductID: uuid.New(),
		Priority:  2,
	}

	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistItemRepo.EXPECT().ProductExistsInWishlist(suite.ctx, suite.wishlistID, req.ProductID).
		Return(false, nil).Once()

	suite.productRepo.EXPECT().Exists(suite.ctx, req.ProductID).
		Return(true, nil).Once()

	suite.wishlistItemRepo.EXPECT().AddItem(suite.ctx, mock.MatchedBy(func(item *models.WishlistItem) bool {
		return item.WishlistID == suite.wishlistID &&
			item.ProductID == req.ProductID &&
			item.Note == req.Note &&
			item.Priority == req.Priority
	})).Return(nil).Once()

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.AddItem(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(wishlist.ID, response.ID)
}

func (suite *WishlistServiceTestSuite) TestAddItem_WishlistNotFound() {
	req := dto.AddWishlistItemRequest{}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.AddItem(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.AddItem")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestAddItem_UserIsNotOwner() {
	req := dto.AddWishlistItemRequest{
		ProductID: uuid.New(),
		Priority:  2,
	}

	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: uuid.New(),
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.AddItem(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.AddItem")
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *WishlistServiceTestSuite) TestAddItem_ProductExistsInWishlist() {
	req := dto.AddWishlistItemRequest{
		ProductID: uuid.New(),
		Priority:  2,
	}

	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistItemRepo.EXPECT().ProductExistsInWishlist(suite.ctx, suite.wishlistID, req.ProductID).
		Return(true, nil).Once()

	response, err := suite.wishlistService.AddItem(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.AddItem")
	suite.ErrorIs(err, apperror.ErrProductAlreadyInWishlist)
}

func (suite *WishlistServiceTestSuite) TestAddItem_ProductNotFound() {
	req := dto.AddWishlistItemRequest{
		ProductID: uuid.New(),
		Priority:  2,
	}

	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistItemRepo.EXPECT().ProductExistsInWishlist(suite.ctx, suite.wishlistID, req.ProductID).
		Return(false, nil).Once()

	suite.productRepo.EXPECT().Exists(suite.ctx, req.ProductID).
		Return(false, nil).Once()

	response, err := suite.wishlistService.AddItem(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.AddItem")
	suite.ErrorIs(err, apperror.ErrProductNotFound)
}

func (suite *WishlistServiceTestSuite) TestAddItem_RepositoryError() {
	req := dto.AddWishlistItemRequest{
		ProductID: uuid.New(),
		Priority:  2,
	}

	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, dbErr).Once()

	response, err := suite.wishlistService.AddItem(suite.ctx, suite.userID, suite.wishlistID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.AddItem")
	suite.ErrorIs(err, dbErr)
}

// ==================== UpdateItem Tests ====================

func (suite *WishlistServiceTestSuite) TestUpdateItem_Success() {
	req := dto.UpdateWishlistItemRequest{
		Note:     new("new note"),
		Priority: new(2),
	}

	item := &models.WishlistItem{
		Note:     new("old note"),
		Priority: 1,
	}

	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
		Items:  []models.WishlistItem{*item},
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistItemRepo.EXPECT().GetItem(suite.ctx, suite.wishlistID, suite.itemID).
		Return(item, nil).Once()

	suite.wishlistItemRepo.EXPECT().UpdateItem(suite.ctx, item).
		Return(nil).Once()

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.UpdateItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Len(response.Items, 1)
	suite.NotNil(item.Note)
	suite.Equal("new note", *item.Note)
	suite.Equal(2, item.Priority)
	suite.Equal(wishlist.ID, response.ID)
}

func (suite *WishlistServiceTestSuite) TestUpdateItem_WishlistNotFound() {
	req := dto.UpdateWishlistItemRequest{}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.UpdateItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.UpdateItem")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestUpdateItem_UserIsNotOwner() {
	req := dto.UpdateWishlistItemRequest{}

	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: uuid.New(),
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.UpdateItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.UpdateItem")
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *WishlistServiceTestSuite) TestUpdateItem_ItemNotFound() {
	req := dto.UpdateWishlistItemRequest{}

	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistItemRepo.EXPECT().GetItem(suite.ctx, suite.wishlistID, suite.itemID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.UpdateItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.UpdateItem")
	suite.ErrorIs(err, apperror.ErrWishlistItemNotFound)
}

func (suite *WishlistServiceTestSuite) TestUpdateItem_RepositoryError() {
	req := dto.UpdateWishlistItemRequest{}

	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, dbErr).Once()

	response, err := suite.wishlistService.UpdateItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.UpdateItem")
	suite.ErrorIs(err, dbErr)
}

// ==================== RemoveItem Tests ====================

func (suite *WishlistServiceTestSuite) TestRemoveItem_Success() {
	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistItemRepo.EXPECT().RemoveItem(suite.ctx, suite.wishlistID, suite.itemID).
		Return(nil).Once()

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, true).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.RemoveItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(wishlist.ID, response.ID)
}

func (suite *WishlistServiceTestSuite) TestRemoveItem_WishlistNotFound() {
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.RemoveItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.RemoveItem")
	suite.ErrorIs(err, apperror.ErrWishlistNotFound)
}

func (suite *WishlistServiceTestSuite) TestRemoveItem_UserIsNotOwner() {
	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: uuid.New(),
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	response, err := suite.wishlistService.RemoveItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.RemoveItem")
	suite.ErrorIs(err, apperror.ErrForbidden)
}

func (suite *WishlistServiceTestSuite) TestRemoveItem_WishlistItemNotFound() {
	wishlist := &models.Wishlist{
		ID:     suite.wishlistID,
		UserID: suite.userID,
	}

	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(wishlist, nil).Once()

	suite.wishlistItemRepo.EXPECT().RemoveItem(suite.ctx, suite.wishlistID, suite.itemID).
		Return(repository.ErrRecordNotFound).Once()

	response, err := suite.wishlistService.RemoveItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.RemoveItem")
	suite.ErrorIs(err, apperror.ErrWishlistItemNotFound)
}

func (suite *WishlistServiceTestSuite) TestRemoveItem_RepositoryError() {
	dbErr := errors.New("db error")
	suite.wishlistRepo.EXPECT().GetByID(suite.ctx, suite.wishlistID, false).
		Return(nil, dbErr).Once()

	response, err := suite.wishlistService.RemoveItem(suite.ctx, suite.userID, suite.wishlistID, suite.itemID)

	suite.Nil(response)
	suite.ErrorContains(err, "wishlistService.RemoveItem")
	suite.ErrorIs(err, dbErr)
}
