package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userPaymentMethodRepository struct {
	db database.Provider
}

func NewUserPaymentMethodRepository(db database.Provider) *userPaymentMethodRepository {
	return &userPaymentMethodRepository{
		db: db,
	}
}

func (p *userPaymentMethodRepository) Upsert(ctx context.Context, paymentMethod *models.UserPaymentMethod) error {
	db := p.db.GetDB(ctx)

	err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "provider"}, {Name: "provider_payment_method_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_used_at": time.Now().UTC(),
			"updated_at":   time.Now().UTC(),
		}),
	}).Create(paymentMethod).Error
	return repository.HandleError(err)
}

func (p *userPaymentMethodRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.UserPaymentMethod, error) {
	db := p.db.GetDB(ctx)

	var paymentMethod models.UserPaymentMethod
	if err := db.First(&paymentMethod, id).Error; err != nil {
		return nil, repository.HandleError(err)
	}

	return &paymentMethod, nil
}

func (p *userPaymentMethodRepository) GetListByUser(ctx context.Context, userID uuid.UUID) ([]*models.UserPaymentMethod, int64, error) {
	db := p.db.GetDB(ctx)

	db = db.Where("user_id = ?", userID)

	var total int64
	if err := db.Model(&models.UserPaymentMethod{}).Count(&total).Error; err != nil {
		return nil, 0, repository.HandleError(err)
	}

	if total == 0 {
		return nil, 0, nil
	}

	var paymentMethods []*models.UserPaymentMethod
	err := db.Find(&paymentMethods).Error

	if err != nil {
		return nil, 0, repository.HandleError(err)
	}

	return paymentMethods, total, nil
}

func (p *userPaymentMethodRepository) SetDefault(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	db := p.db.GetDB(ctx)

	err := db.Transaction(func(tx *gorm.DB) error {
		return SetDefault(tx, id, userID, &models.UserPaymentMethod{})
	})

	return repository.HandleError(err)
}

func (p *userPaymentMethodRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	db := p.db.GetDB(ctx)

	result := db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.UserPaymentMethod{})
	if result.Error != nil {
		return repository.HandleError(result.Error)
	}

	if result.RowsAffected == 0 {
		return repository.ErrRecordNotFound
	}

	return nil
}
