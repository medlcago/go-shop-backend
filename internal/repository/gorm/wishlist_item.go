package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
)

type wishlistItemRepository struct {
	db database.Provider
}

func NewWishlistItemRepository(db database.Provider) *wishlistItemRepository {
	return &wishlistItemRepository{
		db: db,
	}
}

func (w *wishlistItemRepository) GetItem(ctx context.Context, wishlistID uuid.UUID, itemID uuid.UUID) (*models.WishlistItem, error) {
	db := w.db.GetDB(ctx)

	var item models.WishlistItem
	err := db.Where("id = ? AND wishlist_id = ?", itemID, wishlistID).
		First(&item).Error

	if err != nil {
		return nil, repository.HandleError(err)
	}

	return &item, nil
}

func (w *wishlistItemRepository) AddItem(ctx context.Context, item *models.WishlistItem) error {
	db := w.db.GetDB(ctx)

	err := db.Create(item).Error
	return repository.HandleError(err)
}

func (w *wishlistItemRepository) UpdateItem(ctx context.Context, wishlistItem *models.WishlistItem) error {
	db := w.db.GetDB(ctx)

	err := db.Select("*").Updates(wishlistItem).Error
	return repository.HandleError(err)
}

func (w *wishlistItemRepository) RemoveItem(ctx context.Context, wishlistID uuid.UUID, itemID uuid.UUID) (bool, error) {
	db := w.db.GetDB(ctx)

	result := db.Where("id = ? AND wishlist_id = ?", itemID, wishlistID).
		Delete(&models.WishlistItem{})

	if result.Error != nil {
		return false, repository.HandleError(result.Error)
	}

	return result.RowsAffected > 0, nil
}

func (w *wishlistItemRepository) ProductExistsInWishlist(ctx context.Context, wishlistID uuid.UUID, productID uuid.UUID) (bool, error) {
	db := w.db.GetDB(ctx)

	var exists bool
	if err := db.Raw(
		"SELECT EXISTS(SELECT 1 FROM wishlist_items WHERE wishlist_id = ? AND product_id = ?)",
		wishlistID, productID,
	).Scan(&exists).Error; err != nil {
		return false, repository.HandleError(err)
	}

	return exists, nil
}
