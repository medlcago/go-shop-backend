package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
)

type uploadRepository struct {
	db database.Provider
}

func NewUploadRepository(db database.Provider) *uploadRepository {
	return &uploadRepository{
		db: db,
	}
}

func (u *uploadRepository) Create(ctx context.Context, req *models.Upload) error {
	db := u.db.GetDB(ctx)

	err := db.Create(req).Error
	return repository.HandleError(err)
}

func (u *uploadRepository) ExistsByObjectKey(ctx context.Context, objectKey string) (bool, error) {
	db := u.db.GetDB(ctx)

	var exists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM uploads WHERE object_key = ?)", objectKey).Scan(&exists).Error; err != nil {
		return false, repository.HandleError(err)
	}

	return exists, nil
}

func (u *uploadRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	db := u.db.GetDB(ctx)

	if err := db.Where("id = ?", id).Delete(&models.Upload{}).Error; err != nil {
		return repository.HandleError(err)
	}

	return nil
}

func (u *uploadRepository) Count(ctx context.Context, entityType string, entityID uuid.UUID) (int64, error) {
	db := u.db.GetDB(ctx)

	var count int64
	if err := db.Model(&models.Upload{}).
		Where("entity_type = ? AND entity_id = ?",
			entityType, entityID).
		Count(&count).Error; err != nil {
		return 0, repository.HandleError(err)
	}

	return count, nil
}
