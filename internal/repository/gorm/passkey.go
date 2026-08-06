package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
)

type passkeyRepository struct {
	db database.Provider
}

func NewPasskeyRepository(db database.Provider) *passkeyRepository {
	return &passkeyRepository{
		db: db,
	}
}

func (p *passkeyRepository) GetListByUser(ctx context.Context, userID uuid.UUID) ([]models.PasskeyCredential, error) {
	db := p.db.GetDB(ctx)

	var passkeys []models.PasskeyCredential
	if err := db.Where("user_id = ?", userID).Find(&passkeys).Error; err != nil {
		return nil, repository.HandleError(err)
	}

	return passkeys, nil
}

func (p *passkeyRepository) GetByCredentialID(ctx context.Context, credentialID []byte) (*models.PasskeyCredential, error) {
	db := p.db.GetDB(ctx)

	var passkey models.PasskeyCredential
	if err := db.Where("credential_id = ?", credentialID).First(&passkey).Error; err != nil {
		return nil, repository.HandleError(err)
	}

	return &passkey, nil
}

func (p *passkeyRepository) GetByUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.PasskeyCredential, error) {
	db := p.db.GetDB(ctx)

	var passkey models.PasskeyCredential
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&passkey).Error; err != nil {
		return nil, repository.HandleError(err)
	}

	return &passkey, nil
}

func (p *passkeyRepository) Create(ctx context.Context, credential *models.PasskeyCredential) error {
	db := p.db.GetDB(ctx)

	if err := db.Create(&credential).Error; err != nil {
		return repository.HandleError(err)
	}

	return nil
}

func (p *passkeyRepository) Update(ctx context.Context, credential *models.PasskeyCredential) error {
	db := p.db.GetDB(ctx)

	if err := db.Select("*").Updates(credential).Error; err != nil {
		return repository.HandleError(err)
	}

	return nil
}

func (p *passkeyRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	db := p.db.GetDB(ctx)

	result := db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.PasskeyCredential{})

	if result.Error != nil {
		return repository.HandleError(result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrRecordNotFound
	}

	return nil
}
