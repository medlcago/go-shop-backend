package service

import (
	"context"
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	wishlist := &models.Wishlist{
		UserID:     userID,
		Title:      req.Title,
		IsPublic:   req.IsPublic,
		ShareToken: shareToken,
	}

	if err := w.wishlistRepo.Create(ctx, wishlist); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := w.mapWishlist(wishlist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (w *wishlistService) GetWishlist(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.GetWishlist"

	wishlist, err := w.getWishlist(ctx, wishlistID, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !wishlist.CanView(userID) {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrForbidden)
	}

	wishlist, err = w.getWishlist(ctx, wishlistID, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := w.mapWishlist(wishlist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	response, err := w.mapWishlists(wishlists)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}

	return response, total, nil
}

func (w *wishlistService) GetSharedWishlist(
	ctx context.Context,
	shareToken string,
) (*dto.WishlistResponse, error) {
	const op = "wishlistService.GetSharedWishlist"

	wishlist, err := w.wishlistRepo.GetByShareToken(ctx, shareToken, true)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, fmt.Errorf("%s: %w", op, apperror.ErrWishlistNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !wishlist.IsPublic {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrWishlistNotFound)
	}

	response, err := w.mapWishlist(wishlist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	wishlist, err := w.getWishlist(ctx, wishlistID, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !wishlist.CanEdit(userID) {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrForbidden)
	}

	if req.Title != nil {
		wishlist.Title = *req.Title
	}

	if req.IsPublic != nil {
		wishlist.IsPublic = *req.IsPublic
	}

	if err := w.wishlistRepo.Update(ctx, wishlist); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	wishlist, err = w.getWishlist(ctx, wishlistID, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := w.mapWishlist(wishlist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (w *wishlistService) DeleteWishlist(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
) error {
	const op = "wishlistService.DeleteWishlist"

	wishlist, err := w.getWishlist(ctx, wishlistID, false)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !wishlist.CanEdit(userID) {
		return fmt.Errorf("%s: %w", op, apperror.ErrForbidden)
	}

	if err := w.wishlistRepo.Delete(ctx, wishlistID); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (w *wishlistService) RegenerateShareToken(
	ctx context.Context,
	userID uuid.UUID,
	wishlistID uuid.UUID,
) (*dto.WishlistShareTokenResponse, error) {
	const op = "wishlistService.RegenerateShareToken"

	wishlist, err := w.getWishlist(ctx, wishlistID, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !wishlist.CanEdit(userID) {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrForbidden)
	}

	shareToken, err := w.newShareToken()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	wishlist.ShareToken = shareToken
	if err := w.wishlistRepo.Update(ctx, wishlist); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	wishlist, err := w.getWishlist(ctx, wishlistID, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !wishlist.CanEdit(userID) {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrForbidden)
	}

	existsInWishlist, err := w.wishlistItemRepo.ProductExistsInWishlist(ctx, wishlistID, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if existsInWishlist {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrProductAlreadyInWishlist)
	}

	productExists, err := w.productRepo.Exists(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !productExists {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrProductNotFound)
	}

	item := &models.WishlistItem{
		WishlistID: wishlist.ID,
		ProductID:  req.ProductID,
		Note:       req.Note,
		Priority:   req.Priority,
	}

	if err := w.wishlistItemRepo.AddItem(ctx, item); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	wishlist, err = w.getWishlist(ctx, wishlistID, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := w.mapWishlist(wishlist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	wishlist, err := w.getWishlist(ctx, wishlistID, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !wishlist.CanEdit(userID) {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrForbidden)
	}

	item, err := w.wishlistItemRepo.GetItem(ctx, wishlistID, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, fmt.Errorf("%s: %w", op, apperror.ErrWishlistItemNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if req.Note != nil {
		item.Note = req.Note
	}

	if req.Priority != nil {
		item.Priority = *req.Priority
	}

	if err := w.wishlistItemRepo.UpdateItem(ctx, item); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	wishlist, err = w.getWishlist(ctx, wishlistID, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := w.mapWishlist(wishlist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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

	wishlist, err := w.getWishlist(ctx, wishlistID, false)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !wishlist.CanEdit(userID) {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrForbidden)
	}

	removed, err := w.wishlistItemRepo.RemoveItem(ctx, wishlistID, itemID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !removed {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrWishlistItemNotFound)
	}

	wishlist, err = w.getWishlist(ctx, wishlistID, true)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	response, err := w.mapWishlist(wishlist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (w *wishlistService) getWishlist(
	ctx context.Context,
	id uuid.UUID,
	preload bool,
) (*models.Wishlist, error) {
	const op = "wishlistService.getWishlist"

	wishlist, err := w.wishlistRepo.GetByID(ctx, id, preload)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, fmt.Errorf("%s: %w", op, apperror.ErrWishlistNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return wishlist, nil
}

func (w *wishlistService) mapWishlist(wishlist *models.Wishlist) (*dto.WishlistResponse, error) {
	const op = "wishlistService.mapWishlist"

	response, err := mapper.MapOne[*models.Wishlist, dto.WishlistResponse](wishlist)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (w *wishlistService) mapWishlists(wishlists []*models.Wishlist) ([]*dto.WishlistResponse, error) {
	const op = "wishlistService.mapWishlists"

	response, err := mapper.MapList[*models.Wishlist, *dto.WishlistResponse](wishlists)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (w *wishlistService) newShareToken() (string, error) {
	return gonanoid.New(21)
}
