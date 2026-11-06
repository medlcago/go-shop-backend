package gorm

import (
	"go-shop-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SetDefault[T any](db *gorm.DB, id uuid.UUID, userID uuid.UUID) error {
	return db.Transaction(func(tx *gorm.DB) error {
		model := new(T)

		if err := tx.Model(model).Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		result := tx.Model(model).Where("id = ? AND user_id = ?", id, userID).
			Update("is_default", true)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return repository.ErrRecordNotFound
		}

		return nil
	})
}
