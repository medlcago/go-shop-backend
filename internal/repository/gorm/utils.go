package gorm

import (
	"go-shop-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SetDefault(tx *gorm.DB, id uuid.UUID, userID uuid.UUID, value any) error {
	if err := tx.Model(value).Where("user_id = ?", userID).
		Update("is_default", false).Error; err != nil {
		return err
	}

	result := tx.Model(value).Where("id = ? AND user_id = ?", id, userID).
		Update("is_default", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrRecordNotFound
	}

	return nil
}
