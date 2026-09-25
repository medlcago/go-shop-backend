package service

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/mapper"

	"github.com/google/uuid"
)

type ProductChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type favoriteService struct {
	productChecker ProductChecker
	favoriteRepo   repository.FavoriteRepository
}

func NewFavoriteService(
	productChecker ProductChecker,
	favoriteRepo repository.FavoriteRepository,
) *favoriteService {
	return &favoriteService{
		productChecker: productChecker,
		favoriteRepo:   favoriteRepo,
	}
}

func (f *favoriteService) AddProduct(ctx context.Context, userID uuid.UUID, productID uuid.UUID) ([]*dto.FavoriteProduct, int64, error) {
	const op = "favoriteService.AddProduct"

	exists, err := f.productChecker.Exists(ctx, productID)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	if !exists {
		return nil, 0, apperror.Wrap(op, apperror.ErrProductNotFound)
	}

	favorite := &models.Favorite{
		UserID:    userID,
		ProductID: productID,
	}

	if err := f.favoriteRepo.Add(ctx, favorite); err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	response, total, err := f.getFavoriteProductsResponse(ctx, userID)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	return response, total, nil
}

func (f *favoriteService) RemoveProduct(ctx context.Context, userID uuid.UUID, productID uuid.UUID) ([]*dto.FavoriteProduct, int64, error) {
	const op = "favoriteService.RemoveProduct"

	err := f.favoriteRepo.Remove(ctx, productID, userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, 0, apperror.Wrap(op, apperror.ErrProductNotFound)
		}

		return nil, 0, apperror.Wrap(op, err)
	}

	response, total, err := f.getFavoriteProductsResponse(ctx, userID)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	return response, total, nil
}

func (f *favoriteService) List(ctx context.Context, userID uuid.UUID) ([]*dto.FavoriteProduct, int64, error) {
	const op = "favoriteService.List"

	response, total, err := f.getFavoriteProductsResponse(ctx, userID)
	if err != nil {
		return nil, 0, apperror.Wrap(op, err)
	}

	return response, total, nil

}

func (f *favoriteService) IsFavorite(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*dto.FavoriteStatusResponse, error) {
	const op = "favoriteService.IsFavorite"

	exists, err := f.favoriteRepo.Exists(
		ctx,
		productID,
		userID,
	)

	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.FavoriteStatusResponse{
		ProductID:  productID,
		IsFavorite: exists,
	}, nil
}

func (f *favoriteService) mapFavoriteProducts(products []*models.FavoriteProductProjection) ([]*dto.FavoriteProduct, error) {
	const op = "favoriteService.mapFavoriteProducts"

	response, err := mapper.MapList[*models.FavoriteProductProjection, *dto.FavoriteProduct](products)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (f *favoriteService) getFavoriteProductsResponse(
	ctx context.Context,
	userID uuid.UUID,
) ([]*dto.FavoriteProduct, int64, error) {
	favoriteProducts, total, err := f.favoriteRepo.GetListByUser(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	response, err := f.mapFavoriteProducts(favoriteProducts)
	if err != nil {
		return nil, 0, err
	}

	return response, total, nil
}
