package postgres

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/transaction"

	"github.com/google/uuid"
)

type userRepository struct {
	getQueryer transaction.QueryerFunc
}

func NewUserRepository(getQueryer transaction.QueryerFunc) repository.UserRepository {
	return &userRepository{
		getQueryer: getQueryer,
	}
}

func (u userRepository) Save(ctx context.Context, user *models.User) error {
	db := u.getQueryer(ctx)
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING *`

	err := db.GetContext(ctx, user, query, user.Email, user.PasswordHash)
	return repository.HandleSQLError(err)
}

func (u userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	db := u.getQueryer(ctx)
	query := `SELECT * FROM users WHERE id=$1`

	user := new(models.User)
	err := db.GetContext(ctx, user, query, id)

	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return user, nil

}

func (u userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	db := u.getQueryer(ctx)
	query := `SELECT * FROM users WHERE LOWER(email)=LOWER($1)`

	user := new(models.User)
	err := db.GetContext(ctx, user, query, email)

	if err != nil {
		return nil, repository.HandleSQLError(err)
	}

	return user, nil
}

func (u userRepository) IsDeleted(ctx context.Context, id uuid.UUID) (bool, error) {
	db := u.getQueryer(ctx)

	query := `SELECT deleted_at IS NOT NULL as is_deleted FROM users WHERE id = $1`

	var isDeleted bool
	err := db.GetContext(ctx, &isDeleted, query, id)
	if err != nil {
		return false, repository.HandleSQLError(err)
	}

	return isDeleted, nil
}
