package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type favoriteRepository struct {
	db database.Provider
}

func NewFavoriteRepository(db database.Provider) *favoriteRepository {
	return &favoriteRepository{
		db: db,
	}
}

func (f *favoriteRepository) Add(ctx context.Context, favorite *models.Favorite) error {
	db := f.db.GetDB(ctx)

	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "product_id"},
		},
		DoNothing: true,
	}).
		Create(favorite).Error

	return repository.HandleError(err)
}

func (f *favoriteRepository) GetListByUser(ctx context.Context, userID uuid.UUID) ([]*models.FavoriteProductProjection, int64, error) {
	db := f.db.GetDB(ctx)

	db = db.Where("user_id = ?", userID)

	var total int64
	if err := db.Model(&models.Favorite{}).Count(&total).Error; err != nil {
		return nil, 0, repository.HandleError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	var favorites []*models.Favorite
	if err := db.
		Select("id", "product_id", "created_at").
		Preload("Product", func(tx *gorm.DB) *gorm.DB {
			return tx.Select("id", "name", "price", "stock", "reserved")
		}).
		Order("created_at DESC").
		Find(&favorites).Error; err != nil {
		return nil, 0, repository.HandleError(err)
	}

	items := make([]*models.FavoriteProductProjection, 0, len(favorites))
	for _, fav := range favorites {
		items = append(items, &models.FavoriteProductProjection{
			ID:        fav.ProductID,
			Name:      fav.Product.Name,
			Price:     fav.Product.Price,
			Available: fav.Product.Available(),
			CreatedAt: fav.CreatedAt,
		})
	}

	return items, total, nil
}

func (f *favoriteRepository) Exists(ctx context.Context, productID uuid.UUID, userID uuid.UUID) (bool, error) {
	db := f.db.GetDB(ctx)

	var exists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = ? AND product_id = ?)", userID, productID).
		Scan(&exists).Error; err != nil {
		return false, repository.HandleError(err)
	}

	return exists, nil
}

func (f *favoriteRepository) Remove(ctx context.Context, productID uuid.UUID, userID uuid.UUID) error {
	db := f.db.GetDB(ctx)

	result := db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.Favorite{})
	if result.Error != nil {
		return repository.HandleError(result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrRecordNotFound
	}

	return nil
}
