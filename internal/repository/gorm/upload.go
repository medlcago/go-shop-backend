package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"
)

type uploadRepository struct {
	db database.Provider
}

func NewUploadRepository(db database.Provider) repository.UploadRepository {
	return &uploadRepository{
		db: db,
	}
}

func (u *uploadRepository) Save(ctx context.Context, req *models.Upload) error {
	db := u.db.GetDB(ctx)

	err := db.Create(req).Error
	return repository.HandleSQLError(err)
}

func (u *uploadRepository) Exists(ctx context.Context, objectKey string) (bool, error) {
	db := u.db.GetDB(ctx)

	var exists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM uploads WHERE object_key = ?)", objectKey).Scan(&exists).Error; err != nil {
		return false, repository.HandleSQLError(err)
	}

	return exists, nil
}
