package service

import (
	"context"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/pkg/apperror"
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

	user, err := a.userRepo.GetByEmailIncludingDeleted(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, apperror.Wrap(op, apperror.ErrInvalidCredentials)
		}

		return nil, apperror.Wrap(op, err)
	}

	if err := a.verifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.TwoFAEnabled && user.TwoFASecret != nil {
		if req.Code == "" {
			return nil, apperror.Wrap(op, apperror.Err2FACodeRequired)
		}

		decryptedKey, err := a.encryptionManager.Decrypt(ctx, *user.TwoFASecret)
		if err != nil {
			return nil, apperror.Wrap(op, err)
		}

		if !a.totpManager.ValidateCode(decryptedKey, req.Code) {
			return nil, apperror.Wrap(op, apperror.ErrInvalid2FACode)
		}
	}

	if user.DeletedAt.Valid {
		return nil, apperror.Wrap(op, apperror.ErrUserProfileDeleted)
	}

	tokens, err := a.createTokens(user)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return buildUserTokenResponse(user, tokens), nil
}

func (a *authService) Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error) {
	const op = "authService.Register"

	exists, err := a.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if exists {
		return nil, apperror.Wrap(op, apperror.ErrEmailTaken)
	}

	passwordHash, err := a.hasher.Hash(req.Password)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         models.UserRoleCustomer,
	}

	if err := a.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	tokens, err := a.createTokens(user)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return buildUserTokenResponse(user, tokens), nil
}

func (a *authService) Setup2FA(ctx context.Context, userID uuid.UUID) (*dto.Setup2FAResponse, error) {
	const op = "authService.Setup2FA"

	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.TwoFAEnabled {
		return nil, apperror.Wrap(op, apperror.Err2FAAlreadyEnabled)
	}

	key, err := a.totpManager.GenerateSecret(user.Email)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	encryptedKey, err := a.encryptionManager.Encrypt(ctx, key.Secret)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	user.TwoFASecret = &encryptedKey

	if err := a.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Wrap(op, err)
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
		return apperror.Wrap(op, err)
	}

	if err := a.verifyPassword(req.Password, user.PasswordHash); err != nil {
		return apperror.Wrap(op, err)
	}

	if user.TwoFASecret == nil {
		return apperror.Wrap(op, apperror.Err2FANotInitialized)
	}

	if user.TwoFAEnabled {
		return apperror.Wrap(op, apperror.Err2FAAlreadyEnabled)
	}

	decryptedKey, err := a.encryptionManager.Decrypt(ctx, *user.TwoFASecret)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !a.totpManager.ValidateCode(decryptedKey, req.Code) {
		return apperror.Wrap(op, apperror.ErrInvalid2FACode)
	}

	user.TwoFAEnabled = true
	user.TwoFAConfirmedAt = new(time.Now().UTC())

	if err := a.userRepo.Update(ctx, user); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (a *authService) Disable2FA(ctx context.Context, userID uuid.UUID, req dto.Disable2FARequest) error {
	const op = "authService.Disable2FA"

	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if err := a.verifyPassword(req.Password, user.PasswordHash); err != nil {
		return apperror.Wrap(op, err)
	}

	if !user.TwoFAEnabled {
		return apperror.Wrap(op, apperror.Err2FANotEnabled)
	}

	decryptedKey, err := a.encryptionManager.Decrypt(ctx, *user.TwoFASecret)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !a.totpManager.ValidateCode(decryptedKey, req.Code) {
		return apperror.Wrap(op, apperror.ErrInvalid2FACode)
	}

	user.TwoFAEnabled = false
	user.TwoFASecret = nil
	user.TwoFAConfirmedAt = nil

	if err := a.userRepo.Update(ctx, user); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (a *authService) createTokens(user *models.User) (*dto.TokenResponse, error) {
	const (
		op        = "authService.createTokens"
		tokenType = "Bearer"
	)

	payload := token.Payload{
		UserID:         user.ID.String(),
		UserRole:       string(user.Role),
		TwoFAEnabled:   user.TwoFAEnabled,
		EmailConfirmed: user.EmailConfirmed(),
	}

	accessToken, err := a.tokenManager.GenerateAccessToken(payload)
	if err != nil {
		return nil, apperror.Wrap(op, fmt.Errorf("failed to generate access token: %w", err))
	}

	refreshToken, err := a.tokenManager.GenerateRefreshToken(payload)
	if err != nil {
		return nil, apperror.Wrap(op, fmt.Errorf("failed to generate refresh token: %w", err))
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
	}, nil
}

func (a *authService) verifyPassword(password, passwordHash string) error {
	const op = "authService.verifyPassword"

	match, err := a.hasher.Verify(password, passwordHash)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !match {
		return apperror.Wrap(op, apperror.ErrInvalidCredentials)
	}

	return nil
}

func buildUserTokenResponse(user *models.User, token *dto.TokenResponse) *dto.UserTokenResponse {
	response := &dto.UserTokenResponse{
		TokenResponse: token,
	}

	response.User = &dto.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		CreatedAt:      user.CreatedAt,
		Role:           string(user.Role),
		TwoFAEnabled:   user.TwoFAEnabled,
		EmailConfirmed: user.EmailConfirmed(),
	}

	return response
}
