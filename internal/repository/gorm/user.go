package gorm

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/database"

	"github.com/google/uuid"
)

type userRepository struct {
	db database.Provider
}

func NewUserRepository(db database.Provider) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) Create(ctx context.Context, user *models.User) error {
	db := u.db.GetDB(ctx)

	err := db.Create(user).Error
	return repository.HandleError(err)
}

func (u *userRepository) GetByID(ctx context.Context, id uuid.UUID, filter repository.GetUserFilter) (*models.User, error) {
	db := u.db.GetDB(ctx)

	if filter.IncludeDeleted {
		db = db.Unscoped()
	}

	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		return nil, repository.HandleError(err)
	}

	return &user, nil
}

func (u *userRepository) GetByEmail(ctx context.Context, email string, filter repository.GetUserFilter) (*models.User, error) {
	db := u.db.GetDB(ctx)

	if filter.IncludeDeleted {
		db = db.Unscoped()
	}

	var user models.User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, repository.HandleError(err)
	}

	return &user, nil
}

func (u *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	db := u.db.GetDB(ctx)

	var exists bool
	if err := db.Raw("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists).Error; err != nil {
		return false, repository.HandleError(err)
	}

	return exists, nil

}

func (u *userRepository) Update(ctx context.Context, user *models.User) error {
	db := u.db.GetDB(ctx)

	err := db.Select("*").Updates(user).Error
	return repository.HandleError(err)
}
