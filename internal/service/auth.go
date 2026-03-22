package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/crypto"
	"go-shop-backend/pkg/database"
	"go-shop-backend/pkg/hasher"
	"go-shop-backend/pkg/token"
	"go-shop-backend/pkg/totp"
	"time"

	"github.com/google/uuid"
)

type authService struct {
	userRepo          repository.UserRepository
	tokenManager      token.Manager
	hasher            hasher.Hasher
	totpManager       totp.Manager
	encryptionManager crypto.EncryptionManager
	txManager         database.TxManager
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenManager token.Manager,
	hasher hasher.Hasher,
	totpManager totp.Manager,
	encryptionManager crypto.EncryptionManager,
	txManager database.TxManager,
) *authService {
	return &authService{
		userRepo:          userRepo,
		tokenManager:      tokenManager,
		hasher:            hasher,
		totpManager:       totpManager,
		encryptionManager: encryptionManager,
		txManager:         txManager,
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

	match, err := a.hasher.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !match {
		return nil, apperrors.ErrInvalidCredentials
	}

	if user.DeletedAt.Valid {
		return nil, apperrors.ErrUserProfileDeleted
	}

	if user.TwoFAEnabled {
		partialToken, err := a.createPartialToken(user)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return buildUserTokenResponse(user, partialToken, true), nil
	}

	tokens, err := a.createTokens(user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return buildUserTokenResponse(user, tokens, false), nil
}

func (a *authService) Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error) {
	const op = "authService.Register"

	_, err := a.userRepo.GetByEmailUnscoped(ctx, req.Email)
	if err == nil {
		return nil, apperrors.ErrEmailTaken
	}

	passwordHash, err := a.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to hash password: %w", op, err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         models.UserRoleCustomer,
	}

	if err := a.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := a.createTokens(user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return buildUserTokenResponse(user, tokens, false), nil
}

func (a *authService) Setup2FA(ctx context.Context, userID uuid.UUID) (*dto.Setup2FAResponse, error) {
	const op = "authService.Setup2FA"

	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if user.TwoFASecret != nil && user.TwoFAEnabled {
		return nil, apperrors.Err2FAAlreadyEnabled
	}

	key, err := a.totpManager.GenerateSecret(user.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	encryptedKey, err := a.encryptionManager.Encrypt(ctx, key.Secret)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.TwoFASecret = &encryptedKey

	if err := a.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &dto.Setup2FAResponse{
		Secret: key.Secret,
		QRCode: key.QRCode,
	}, nil
}

func (a *authService) Confirm2FA(ctx context.Context, userID uuid.UUID, req dto.Confirm2FARequest) error {
	const op = "authService.Confirm2FA"

	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	match, err := a.hasher.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !match {
		return apperrors.ErrInvalidPassword
	}

	if user.TwoFASecret == nil {
		return apperrors.Err2FANotInitialized
	}

	if user.TwoFAEnabled {
		return apperrors.Err2FAAlreadyEnabled
	}

	decryptionKey, err := a.encryptionManager.Decrypt(ctx, *user.TwoFASecret)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !a.totpManager.ValidateCode(decryptionKey, req.Code) {
		return apperrors.ErrInvalid2FACode
	}

	user.TwoFAEnabled = true
	user.TwoFaConfirmedAt = new(time.Now().UTC())

	if err := a.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *authService) Disable2FA(ctx context.Context, userID uuid.UUID, req dto.Disable2FARequest) error {
	const op = "authService.Disable2FA"

	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	match, err := a.hasher.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !match {
		return apperrors.ErrInvalidPassword
	}

	if !user.TwoFAEnabled || user.TwoFASecret == nil {
		return apperrors.Err2FANotEnabled
	}

	decryptionKey, err := a.encryptionManager.Decrypt(ctx, *user.TwoFASecret)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !a.totpManager.ValidateCode(decryptionKey, req.Code) {
		return apperrors.ErrInvalid2FACode
	}

	user.TwoFAEnabled = false
	user.TwoFASecret = nil
	user.TwoFaConfirmedAt = nil

	if err := a.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *authService) Verify2FA(ctx context.Context, req dto.Verify2FARequest) (*dto.UserTokenResponse, error) {
	const op = "authService.Verify2FA"

	claims, err := a.tokenManager.ValidateToken(req.Token)
	if err != nil {
		return nil, apperrors.ErrInvalidToken
	}

	if claims.TokenType != token.PartialTokenType {
		return nil, apperrors.ErrInvalidToken
	}

	uid, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.userRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !user.TwoFAEnabled || user.TwoFASecret == nil {
		return nil, apperrors.Err2FANotEnabled
	}

	decryptionKey, err := a.encryptionManager.Decrypt(ctx, *user.TwoFASecret)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !a.totpManager.ValidateCode(decryptionKey, req.Code) {
		return nil, apperrors.ErrInvalid2FACode
	}

	tokens, err := a.createTokens(user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return buildUserTokenResponse(user, tokens, false), nil
}

func (a *authService) createTokens(user *models.User) (*dto.TokenResponse, error) {
	const (
		op        = "authService.createTokens"
		tokenType = "Bearer"
	)

	payload := token.Payload{
		UserID:       user.ID.String(),
		UserRole:     string(user.Role),
		TwoFAEnabled: user.TwoFAEnabled,
	}

	accessToken, err := a.tokenManager.GenerateAccessToken(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate access token: %w", op, err)
	}

	refreshToken, err := a.tokenManager.GenerateRefreshToken(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate refresh token: %w", op, err)
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
	}, nil
}

func (a *authService) createPartialToken(user *models.User) (*dto.TokenResponse, error) {
	const op = "authService.createTempToken"

	partialToken, err := a.tokenManager.GeneratePartialToken(token.Payload{
		UserID:       user.ID.String(),
		UserRole:     string(user.Role),
		TwoFAEnabled: user.TwoFAEnabled,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to generate partial token: %w", op, err)
	}

	return &dto.TokenResponse{
		PartialToken: partialToken,
	}, nil
}

func buildUserTokenResponse(user *models.User, token *dto.TokenResponse, requires2FA bool) *dto.UserTokenResponse {
	response := &dto.UserTokenResponse{
		TokenResponse: token,
		Requires2FA:   requires2FA,
	}

	if !requires2FA {
		response.User = &dto.UserResponse{
			ID:           user.ID,
			Email:        user.Email,
			CreatedAt:    user.CreatedAt,
			Role:         string(user.Role),
			TwoFaEnabled: user.TwoFAEnabled,
		}
	}

	return response
}
