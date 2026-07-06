package service

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/mapper"

	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type wishlistService struct {
	wishlistRepo     repository.WishlistRepository
	wishlistItemRepo repository.WishlistItemRepository
	productRepo      repository.ProductRepository
}

func NewWishlistService(
	wishlistRepo repository.WishlistRepository,
	wishlistItemRepo repository.WishlistItemRepository,
	productRepo repository.ProductRepository,
) *wishlistService {
	return &wishlistService{
		wishlistRepo:     wishlistRepo,
		wishlistItemRepo: wishlistItemRepo,
		productRepo:      productRepo,
	}
}

func (w *wishlistService) CreateWishlist(
	ctx context.Context,
	userID uuid.UUID,
	req dto.CreateWishlistRequest,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.CreateWishlist"

	shareToken, err := w.newShareToken()
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	wl := &models.Wishlist{
		UserID:     userID,
		Title:      req.Title,
		IsPublic:   req.IsPublic,
		ShareToken: shareToken,
	}

	if err := w.wishlistRepo.Create(ctx, wl); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := w.mapWishlist(wl)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) GetWishlist(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.GetWishlist"

	wl, err := w.getWishlistByID(ctx, wishlistID, false)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if !wl.IsOwnedBy(userID) && !wl.IsPublic {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	response, err := w.getWishlistResponse(ctx, wishlistID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) GetWishlists(
	ctx context.Context,
	userID uuid.UUID,
	req dto.ListWishlistRequest,
) ([]*dto.WishlistResponse, int64, error) {
	const op = "wishlistService.GetWishlists"

	wishlists, total, err := w.wishlistRepo.GetListByUser(ctx, userID, req)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	response, err := w.mapWishlists(wishlists)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	return response, total, nil
}

func (w *wishlistService) GetSharedWishlist(
	ctx context.Context,
	shareToken string,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.GetSharedWishlist"

	wl, err := w.wishlistRepo.GetByShareToken(ctx, shareToken, true)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrWishlistNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	if !wl.IsPublic {
		return nil, apperror.Wrap(op, apperror.ErrWishlistNotFound)
	}

	response, err := w.mapWishlist(wl)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) UpdateWishlist(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
	req dto.UpdateWishlistRequest,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.UpdateWishlist"

	wl, err := w.getWishlistByID(ctx, wishlistID, false)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if !wl.IsOwnedBy(userID) {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	applyWishlistUpdates(wl, req)

	if err := w.wishlistRepo.Update(ctx, wl); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := w.getWishlistResponse(ctx, wishlistID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) DeleteWishlist(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
) error {
	const op = "wishlistService.DeleteWishlist"

	wl, err := w.getWishlistByID(ctx, wishlistID, false)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !wl.IsOwnedBy(userID) {
		return apperror.Wrap(op, apperror.ErrForbidden)
	}

	if err := w.wishlistRepo.Delete(ctx, wishlistID); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (w *wishlistService) RegenerateShareToken(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
) (*dto.WishlistShareTokenResponse, error) {
	const op = "wishlistService.RegenerateShareToken"

	wl, err := w.getWishlistByID(ctx, wishlistID, false)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if !wl.IsOwnedBy(userID) {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	shareToken, err := w.newShareToken()
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	wl.ShareToken = shareToken
	if err := w.wishlistRepo.Update(ctx, wl); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.WishlistShareTokenResponse{
		ShareToken: shareToken,
	}, nil
}

func (w *wishlistService) AddItem(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
	req dto.AddWishlistItemRequest,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.AddItem"

	wl, err := w.getWishlistByID(ctx, wishlistID, false)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if !wl.IsOwnedBy(userID) {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	if err := w.ensureProductCanBeAdded(ctx, wishlistID, req.ProductID); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	item := &models.WishlistItem{
		WishlistID: wl.ID,
		ProductID:  req.ProductID,
		Note:       req.Note,
		Priority:   req.Priority,
	}

	if err := w.wishlistItemRepo.AddItem(ctx, item); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := w.getWishlistResponse(ctx, wishlistID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) UpdateItem(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
	itemID uuid.UUID,
	req dto.UpdateWishlistItemRequest,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.UpdateItem"

	wl, err := w.getWishlistByID(ctx, wishlistID, false)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if !wl.IsOwnedBy(userID) {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	item, err := w.getWishlistItem(ctx, wishlistID, itemID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	applyWishlistItemUpdates(item, req)

	if err := w.wishlistItemRepo.UpdateItem(ctx, item); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	response, err := w.getWishlistResponse(ctx, wishlistID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) RemoveItem(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
	itemID uuid.UUID,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.RemoveItem"

	wl, err := w.getWishlistByID(ctx, wishlistID, false)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if !wl.IsOwnedBy(userID) {
		return nil, apperror.Wrap(op, apperror.ErrForbidden)
	}

	if err := w.wishlistItemRepo.RemoveItem(ctx, wishlistID, itemID); err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrWishlistItemNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	response, err := w.getWishlistResponse(ctx, wishlistID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) getWishlistByID(
	ctx context.Context,
	id uuid.UUID,
	preload bool,
) (*models.Wishlist, error) {
	const op = "wishlistService.getWishlistByID"

	wl, err := w.wishlistRepo.GetByID(ctx, id, preload)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrWishlistNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return wl, nil
}

func (w *wishlistService) getWishlistItem(
	ctx context.Context,
	wishlistID uuid.UUID,
	itemID uuid.UUID,
) (*models.WishlistItem, error) {
	const op = "wishlistService.getWishlistItem"

	item, err := w.wishlistItemRepo.GetItem(ctx, wishlistID, itemID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrWishlistItemNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return item, nil
}

func (w *wishlistService) getWishlistResponse(
	ctx context.Context,
	wishlistID uuid.UUID,
) (*dto.WishlistResponse, error) {
	wl, err := w.getWishlistByID(ctx, wishlistID, true)
	if err != nil {
		return nil, err
	}

	return w.mapWishlist(wl)
}

func (w *wishlistService) ensureProductCanBeAdded(
	ctx context.Context,
	wishlistID uuid.UUID,
	productID uuid.UUID,
) error {
	const op = "wishlistService.ensureProductCanBeAdded"

	exists, err := w.wishlistItemRepo.ProductExistsInWishlist(ctx, wishlistID, productID)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if exists {
		return apperror.Wrap(op, apperror.ErrProductAlreadyInWishlist)
	}

	productExists, err := w.productRepo.Exists(ctx, productID)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !productExists {
		return apperror.Wrap(op, apperror.ErrProductNotFound)
	}

	return nil
}

func (w *wishlistService) mapWishlist(wishlist *models.Wishlist) (*dto.WishlistResponse, error) {
	const op = "wishlistService.mapWishlist"

	response, err := mapper.MapOne[*models.Wishlist, dto.WishlistResponse](wishlist)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) mapWishlists(wishlists []*models.Wishlist) ([]*dto.WishlistResponse, error) {
	const op = "wishlistService.mapWishlists"

	response, err := mapper.MapList[*models.Wishlist, *dto.WishlistResponse](wishlists)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (w *wishlistService) newShareToken() (string, error) {
	return gonanoid.New(21)
}

func applyWishlistUpdates(wl *models.Wishlist, req dto.UpdateWishlistRequest) {
	if req.Title != nil {
		wl.Title = *req.Title
	}

	if req.IsPublic != nil {
		wl.IsPublic = *req.IsPublic
	}
}

func applyWishlistItemUpdates(wlItem *models.WishlistItem, req dto.UpdateWishlistItemRequest) {
	if req.Note != nil {
		wlItem.Note = req.Note
	}

	if req.Priority != nil {
		wlItem.Priority = *req.Priority
	}
}
