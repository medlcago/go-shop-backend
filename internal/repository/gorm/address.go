package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type addressRepository struct {
	db database.Provider
}

func NewAddressRepository(db database.Provider) *addressRepository {
	return &addressRepository{
		db: db,
	}
}

func (a *addressRepository) Create(ctx context.Context, address *models.Address) error {
	db := a.db.GetDB(ctx)

	err := db.Create(address).Error
	return repository.HandleError(err)
}

func (a *addressRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Address, error) {
	db := a.db.GetDB(ctx)

	var address models.Address
	err := db.Model(&models.Address{}).Where("id = ? AND user_id = ?", id, userID).First(&address).Error
	if err != nil {
		return nil, repository.HandleError(err)
	}

	return &address, nil
}

func (a *addressRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Address, error) {
	db := a.db.GetDB(ctx)

	var addresses []*models.Address
	err := db.Model(&models.Address{}).Where("user_id = ?", userID).Find(&addresses).Error
	if err != nil {
		return nil, repository.HandleError(err)
	}

	return addresses, nil
}

func (a *addressRepository) Update(ctx context.Context, address *models.Address) error {
	db := a.db.GetDB(ctx)

	err := db.Select("*").Updates(address).Error
	return repository.HandleError(err)
}

func (a *addressRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	db := a.db.GetDB(ctx)

	result := db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Address{})
	if result.Error != nil {
		return repository.HandleError(result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrRecordNotFound
	}

	return nil
}

func (a *addressRepository) SetDefault(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	db := a.db.GetDB(ctx)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Address{}).Where("user_id = ?", userID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		result := tx.Model(&models.Address{}).Where("id = ? AND user_id = ?", id, userID).
			Update("is_default", true)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return repository.ErrRecordNotFound
		}

		return nil
	})

	return repository.HandleError(err)
}
