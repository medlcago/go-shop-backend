package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/utils"

	"github.com/google/uuid"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (u *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	const op = "userService.GetUserByID"

	user, err := u.userRepo.GetByIDUnscoped(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if user.DeletedAt.Valid {
		return nil, apperrors.ErrUserProfileDeleted
	}

	var resp dto.UserResponse
	if err := utils.Copy(&resp, user); err != nil {
		return nil, fmt.Errorf("%s: failed to copy user: %w", op, err)
	}

	return &resp, nil
}
