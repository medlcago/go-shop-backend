package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/hasher"
	"go-shop-backend/pkg/token"
	"go-shop-backend/pkg/utils"
)

type authService struct {
	userRepo       repository.UserRepository
	tokenManager   token.Manager
	passwordHasher hasher.Hasher
}

func NewAuthService(userRepo repository.UserRepository, tokenManager token.Manager, passwordHasher hasher.Hasher) AuthService {
	return &authService{
		userRepo:       userRepo,
		tokenManager:   tokenManager,
		passwordHasher: passwordHasher,
	}
}

func (a *authService) Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserTokenResponse, error) {
	const op = "authService.Login"

	user, err := a.userRepo.GetByEmailUnscoped(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	match, err := a.passwordHasher.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !match {
		return nil, apperrors.ErrInvalidCredentials
	}

	if user.DeletedAt.Valid {
		return nil, apperrors.ErrUserProfileDeleted
	}

	tokenResponse, err := a.createTokens(user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return buildUserTokenResponse(user, tokenResponse)
}

func (a *authService) Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error) {
	const op = "authService.Register"

	_, err := a.userRepo.GetByEmailUnscoped(ctx, req.Email)
	if err == nil {
		return nil, apperrors.ErrEmailTaken
	}

	passwordHash, err := a.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to hash password: %w", op, err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := a.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	tokenResponse, err := a.createTokens(user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return buildUserTokenResponse(user, tokenResponse)
}

func (a *authService) createTokens(user *models.User) (*dto.TokenResponse, error) {
	const (
		op        = "authService.createTokens"
		tokenType = "Bearer"
	)

	var errs []error
	payload := token.Payload{
		UserID:   user.ID.String(),
		UserRole: string(user.Role),
	}

	accessToken, err := a.tokenManager.GenerateAccessToken(payload)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: failed to generate access token: %w", op, err))
	}

	refreshToken, err := a.tokenManager.GenerateRefreshToken(payload)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: failed to generate refresh token: %w", op, err))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
	}, nil
}

func buildUserTokenResponse(user *models.User, token *dto.TokenResponse) (*dto.UserTokenResponse, error) {
	resp := &dto.UserTokenResponse{
		TokenResponse: token,
		User:          &dto.UserResponse{},
	}

	if err := utils.Copy(resp.User, user); err != nil {
		return nil, fmt.Errorf("failed to copy user: %w", err)
	}

	return resp, nil
}
