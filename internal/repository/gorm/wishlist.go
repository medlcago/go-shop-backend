package gorm

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/repository/gorm/scopes"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
)

type wishlistRepository struct {
	db database.Provider
}

func NewWishlistRepository(db database.Provider) *wishlistRepository {
	return &wishlistRepository{
		db: db,
	}
}

func (w *wishlistRepository) Create(ctx context.Context, wishlist *models.Wishlist) error {
	db := w.db.GetDB(ctx)

	err := db.Create(wishlist).Error
	return repository.HandleSQLError(err)
}

func (w *wishlistRepository) Update(ctx context.Context, wishlist *models.Wishlist) error {
	db := w.db.GetDB(ctx)

	err := db.Select("*").Updates(wishlist).Error
	return repository.HandleSQLError(err)
}

func (w *wishlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := w.db.GetDB(ctx)

	result := db.Delete(&models.Wishlist{}, id)
	if result.Error != nil {
		return repository.HandleSQLError(result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrRecordNotFound
	}

	return nil
}

func (w *wishlistRepository) GetByID(ctx context.Context, id uuid.UUID, preload bool) (*models.Wishlist, error) {
	db := w.db.GetDB(ctx)

	if preload {
		db = db.Scopes(
			scopes.WishlistWithRelations(),
		)
	}

	var wishlist models.Wishlist
	err := db.First(&wishlist, id).Error
	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &wishlist, nil
}

func (w *wishlistRepository) GetByShareToken(ctx context.Context, token string, preload bool) (*models.Wishlist, error) {
	db := w.db.GetDB(ctx)

	if preload {
		db = db.Scopes(
			scopes.WishlistWithRelations(),
		)
	}

	db = db.Where("share_token = ?", token)

	var wishlist models.Wishlist
	err := db.First(&wishlist).Error
	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return &wishlist, nil
}

func (w *wishlistRepository) GetListByUser(ctx context.Context, userID uuid.UUID, req dto.ListWishlistRequest) ([]*models.Wishlist, int64, error) {
	db := w.db.GetDB(ctx)

	db = db.Where("user_id = ?", userID)

	var total int64
	if err := db.Model(&models.Wishlist{}).Count(&total).Error; err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	var wishlists []*models.Wishlist
	err := db.Scopes(
		scopes.WishlistWithRelations(),
		scopes.Paginate(req.Limit, req.Limit),
	).Find(&wishlists).Error

	if err != nil {
		return nil, 0, repository.HandleSQLError(err)
	}

	return wishlists, total, nil
}
