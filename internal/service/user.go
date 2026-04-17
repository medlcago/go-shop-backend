package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/mapper"

	"github.com/google/uuid"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *userService {
	return &userService{
		userRepo: userRepo,
	}
}

func (u *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	const op = "userService.GetUserByID"

	user, err := u.userRepo.GetByIDUnscoped(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, fmt.Errorf("%s: %w", op, apperror.ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if user.DeletedAt.Valid {
		return nil, fmt.Errorf("%s: %w", op, apperror.ErrUserProfileDeleted)
	}

	response, err := u.mapUser(user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}

func (u *userService) mapUser(user *models.User) (*dto.UserResponse, error) {
	const op = "userService.mapUser"

	response, err := mapper.MapOne[*models.User, dto.UserResponse](user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return response, nil
}
