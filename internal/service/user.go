package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/tasks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/cache"
	"go-shop-backend/pkg/mapper"
	"go-shop-backend/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type userService struct {
	userRepo                    repository.UserRepository
	notificationTask            tasks.NotificationTask
	cache                       cache.Cache
	emailConfirmationCodeLength int
	emailConfirmationCodeTTL    time.Duration
}

func NewUserService(
	userRepo repository.UserRepository,
	notificationTask tasks.NotificationTask,
	cache cache.Cache,
	emailConfirmationCodeLength int,
	emailConfirmationCodeTTL time.Duration,
) *userService {
	return &userService{
		userRepo:                    userRepo,
		notificationTask:            notificationTask,
		cache:                       cache,
		emailConfirmationCodeLength: emailConfirmationCodeLength,
		emailConfirmationCodeTTL:    emailConfirmationCodeTTL,
	}
}

func (u *userService) GetUserByID(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	const op = "userService.GetUserByID"

	user, err := u.getUserByIDIncludingDeleted(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.DeletedAt.Valid {
		return nil, apperror.Wrap(op, apperror.ErrUserProfileDeleted)
	}

	response, err := u.mapUser(user)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}

func (u *userService) EmailConfirmation(ctx context.Context, userID uuid.UUID) (*dto.EmailConfirmationResponse, error) {
	const op = "userService.EmailConfirmation"

	cacheKey := fmt.Sprintf("email_confirmation:%s", userID)
	err := u.cache.Exists(ctx, cacheKey)
	if err == nil {
		return nil, apperror.Wrap(op, apperror.ErrEmailConfirmationCodeAlreadySent)
	}
	if !errors.Is(err, cache.ErrNotFound) {
		return nil, apperror.Wrap(op, err)
	}

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.EmailConfirmedAt != nil {
		return nil, apperror.Wrap(op, apperror.ErrEmailAlreadyConfirmed)
	}

	code, err := utils.GenerateCode(u.emailConfirmationCodeLength)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := u.cache.Set(ctx, cacheKey, code, u.emailConfirmationCodeTTL); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := u.notificationTask.SendEmailConfirmationCode(ctx, user.Email, code); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.EmailConfirmationResponse{
		ExpiresIn: int(u.emailConfirmationCodeTTL.Seconds()),
	}, nil
}

func (u *userService) ConfirmEmail(ctx context.Context, userID uuid.UUID, req dto.ConfirmEmailRequest) (*dto.ConfirmEmailResponse, error) {
	const op = "userService.ConfirmEmail"

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.EmailConfirmedAt != nil {
		return nil, apperror.Wrap(op, apperror.ErrEmailAlreadyConfirmed)
	}

	if err := u.verifyEmailCode(ctx, userID, req.Code); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	user.EmailConfirmedAt = new(time.Now().UTC())
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.ConfirmEmailResponse{OK: true}, nil

}

func (u *userService) getUserByIDIncludingDeleted(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const op = "userService.getUserByIDIncludingDeleted"

	user, err := u.userRepo.GetByIDIncludingDeleted(ctx, userID)

	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrUserNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return user, nil
}

func (u *userService) verifyEmailCode(ctx context.Context, userID uuid.UUID, inputCode string) error {
	const op = "userService.verifyCode"

	cacheKey := fmt.Sprintf("email_confirmation:%s", userID)
	code, err := u.cache.Get(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return apperror.Wrap(op, apperror.ErrInvalidCode)
		}

		return apperror.Wrap(op, err)
	}

	if inputCode != code {
		return apperror.Wrap(op, apperror.ErrInvalidCode)
	}

	_ = u.cache.Delete(ctx, cacheKey)

	return nil
}

func (u *userService) mapUser(user *models.User) (*dto.UserResponse, error) {
	const op = "userService.mapUser"

	response, err := mapper.MapOne[*models.User, dto.UserResponse](user)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
}
