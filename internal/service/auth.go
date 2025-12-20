package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/jtoken"
	"go-shop-backend/pkg/password"
	"go-shop-backend/pkg/utils"
)

type authService struct {
	userRepo  repository.UserRepository
	secretKey string
}

func NewAuthService(userRepo repository.UserRepository, secretKey string) AuthService {
	return &authService{
		userRepo:  userRepo,
		secretKey: secretKey,
	}
}

func (a *authService) Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserTokenResponse, error) {
	const op = "authService.Login"

	user, err := a.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	match := password.VerifyPassword(req.Password, user.PasswordHash)
	if !match {
		return nil, apperrors.ErrInvalidCredentials
	}

	token, err := a.createTokens(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return buildUserTokenResponse(user, token)
}

func (a *authService) Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error) {
	const op = "authService.Register"

	_, err := a.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, apperrors.ErrEmailTaken
	}

	passwordHash, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to hash password: %w", op, err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := a.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	token, err := a.createTokens(user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return buildUserTokenResponse(user, token)
}

func (a *authService) createTokens(userID string) (*dto.TokenResponse, error) {
	const (
		op        = "authService.createTokens"
		tokenType = "Bearer"
	)

	var errs []error
	payload := map[string]interface{}{
		"user_id": userID,
	}

	accessToken, err := jtoken.GenerateAccessToken(payload, a.secretKey)
	if err != nil {
		errs = append(errs, fmt.Errorf("%s: failed to generate access token: %w", op, err))
	}

	refreshToken, err := jtoken.GenerateRefreshToken(payload, a.secretKey)
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

	if err := utils.Copy(&resp.User, user); err != nil {
		return nil, fmt.Errorf("failed to copy user: %w", err)
	}

	return resp, nil
}
