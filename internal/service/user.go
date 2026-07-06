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
	"go-shop-backend/pkg/crypto"
	"go-shop-backend/pkg/hasher"
	"go-shop-backend/pkg/mapper"
	"go-shop-backend/pkg/token"
	"go-shop-backend/pkg/totp"
	"go-shop-backend/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type UserEmailConfig struct {
	EmailConfirmationCodeLength int
	EmailConfirmationCodeTTL    time.Duration
}

type userService struct {
	userRepo          repository.UserRepository
	tokenManager      token.Manager
	hasher            hasher.Hasher
	totpManager       totp.Manager
	encryptionManager crypto.EncryptionManager
	notificationTask  tasks.NotificationTask
	cache             cache.Cache
	userEmailConfig   *UserEmailConfig
}

func NewUserService(
	userRepo repository.UserRepository,
	tokenManager token.Manager,
	hasher hasher.Hasher,
	totpManager totp.Manager,
	encryptionManager crypto.EncryptionManager,
	notificationTask tasks.NotificationTask,
	cache cache.Cache,
	userEmailConfig *UserEmailConfig,
) *userService {
	if userEmailConfig == nil {
		panic("userService: userEmailConfig is nil")
	}

	return &userService{
		userRepo:          userRepo,
		tokenManager:      tokenManager,
		hasher:            hasher,
		totpManager:       totpManager,
		encryptionManager: encryptionManager,
		notificationTask:  notificationTask,
		cache:             cache,
		userEmailConfig:   userEmailConfig,
	}
}

func (u *userService) Login(ctx context.Context, req dto.UserLoginRequest) (*dto.UserTokenResponse, error) {
	const op = "userService.Login"

	user, err := u.userRepo.GetByEmailIncludingDeleted(ctx, req.Email)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrInvalidCredentials)
		}

		return nil, apperror.Wrap(op, err)
	}

	if err := u.verifyPassword(req.Password, user.PasswordHash); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.TwoFAEnabled && user.TwoFASecret != nil {
		if err := u.verify2FACode(ctx, *user.TwoFASecret, req.Code); err != nil {
			return nil, apperror.Wrap(op, err)
		}
	}

	if user.DeletedAt.Valid {
		return nil, apperror.Wrap(op, apperror.ErrUserProfileDeleted)
	}

	tokens, err := u.createTokens(user)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return buildUserTokenResponse(user, tokens), nil
}

func (u *userService) Register(ctx context.Context, req dto.UserRegisterRequest) (*dto.UserTokenResponse, error) {
	const op = "userService.Register"

	exists, err := u.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if exists {
		return nil, apperror.Wrap(op, apperror.ErrEmailTaken)
	}

	passwordHash, err := u.hasher.Hash(req.Password)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         models.UserRoleCustomer,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	tokens, err := u.createTokens(user)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return buildUserTokenResponse(user, tokens), nil
}

func (u *userService) Setup2FA(ctx context.Context, userID uuid.UUID) (*dto.Setup2FAResponse, error) {
	const op = "userService.Setup2FA"

	user, err := u.getUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.TwoFAEnabled {
		return nil, apperror.Wrap(op, apperror.Err2FAAlreadyEnabled)
	}

	key, err := u.totpManager.GenerateSecret(user.Email)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	encryptedKey, err := u.encryptionManager.Encrypt(ctx, key.Secret)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	user.TwoFASecret = &encryptedKey

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.Setup2FAResponse{
		Secret: key.Secret,
		QRCode: key.QRCode,
	}, nil
}

func (u *userService) Confirm2FA(ctx context.Context, userID uuid.UUID, req dto.Confirm2FARequest) error {
	const op = "userService.Confirm2FA"

	user, err := u.getUserByID(ctx, userID)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if err := u.verifyPassword(req.Password, user.PasswordHash); err != nil {
		return apperror.Wrap(op, err)
	}

	if user.TwoFASecret == nil {
		return apperror.Wrap(op, apperror.Err2FANotInitialized)
	}

	if user.TwoFAEnabled {
		return apperror.Wrap(op, apperror.Err2FAAlreadyEnabled)
	}

	if err := u.verify2FACode(ctx, *user.TwoFASecret, req.Code); err != nil {
		return apperror.Wrap(op, err)
	}

	user.TwoFAEnabled = true
	user.TwoFAConfirmedAt = new(time.Now().UTC())

	if err := u.userRepo.Update(ctx, user); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (u *userService) Disable2FA(ctx context.Context, userID uuid.UUID, req dto.Disable2FARequest) error {
	const op = "userService.Disable2FA"

	user, err := u.getUserByID(ctx, userID)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if err := u.verifyPassword(req.Password, user.PasswordHash); err != nil {
		return apperror.Wrap(op, err)
	}

	if !user.TwoFAEnabled || user.TwoFASecret == nil {
		return apperror.Wrap(op, apperror.Err2FANotEnabled)
	}

	if err := u.verify2FACode(ctx, *user.TwoFASecret, req.Code); err != nil {
		return apperror.Wrap(op, err)
	}

	user.TwoFAEnabled = false
	user.TwoFASecret = nil
	user.TwoFAConfirmedAt = nil

	if err := u.userRepo.Update(ctx, user); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
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

func (u *userService) SendEmailConfirmationCode(ctx context.Context, userID uuid.UUID) (*dto.SendEmailConfirmationResponse, error) {
	const op = "userService.SendEmailConfirmationCode"

	cacheKey := fmt.Sprintf("email_confirmation:%s", userID)
	exists, err := u.cache.Exists(ctx, cacheKey)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if exists {
		return nil, apperror.Wrap(op, apperror.ErrEmailConfirmationCodeAlreadySent)
	}

	user, err := u.getUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.EmailConfirmedAt != nil {
		return nil, apperror.Wrap(op, apperror.ErrEmailAlreadyConfirmed)
	}

	code, err := utils.GenerateCode(u.userEmailConfig.EmailConfirmationCodeLength)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if err := u.cache.Set(ctx, cacheKey, code, u.userEmailConfig.EmailConfirmationCodeTTL); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	payload := tasks.SendEmailConfirmationCodePayload{
		Email: user.Email,
		Code:  code,
	}
	if err := u.notificationTask.SendEmailConfirmationCode(ctx, payload); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.SendEmailConfirmationResponse{
		ExpiresIn: int(u.userEmailConfig.EmailConfirmationCodeTTL.Seconds()),
	}, nil
}

func (u *userService) ConfirmEmail(ctx context.Context, userID uuid.UUID, req dto.ConfirmEmailRequest) (*dto.ConfirmEmailResponse, error) {
	const op = "userService.ConfirmEmail"

	user, err := u.getUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.EmailConfirmedAt != nil {
		return nil, apperror.Wrap(op, apperror.ErrEmailAlreadyConfirmed)
	}

	if err := u.verifyEmailCode(ctx, userID, req.Code); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	now := time.Now().UTC()
	user.EmailConfirmedAt = &now
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return &dto.ConfirmEmailResponse{
		OK:               true,
		EmailConfirmedAt: now.UTC().Format(time.RFC3339),
	}, nil
}

func (u *userService) ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest) error {
	const op = "userService.ChangePassword"

	user, err := u.getUserByID(ctx, userID)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if err := u.verifyPassword(req.Password, user.PasswordHash); err != nil {
		return apperror.Wrap(op, err)
	}

	if user.TwoFAEnabled && user.TwoFASecret != nil {
		if err := u.verify2FACode(ctx, *user.TwoFASecret, req.Code); err != nil {
			return apperror.Wrap(op, err)
		}
	}

	passwordHash, err := u.hasher.Hash(req.NewPassword)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	user.PasswordHash = passwordHash
	if err := u.userRepo.Update(ctx, user); err != nil {
		return apperror.Wrap(op, err)
	}

	return nil
}

func (u *userService) RefreshToken(ctx context.Context, tokenString string) (*dto.UserTokenResponse, error) {
	const op = "userService.RefreshToken"

	claims, err := u.tokenManager.ValidateToken(tokenString)
	if err != nil {
		if token.IsErrInvalidToken(err) {
			return nil, apperror.Wrap(op, apperror.ErrInvalidToken)
		}

		return nil, apperror.Wrap(op, err)
	}

	if claims.TokenType != token.RefreshTokenType {
		return nil, apperror.Wrap(op, apperror.ErrInvalidTokenType)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, apperror.Wrap(op, apperror.ErrInvalidToken)
	}

	if err := u.revokeToken(ctx, claims.ID, claims.TokenType, claims.ExpiresAt.Time); err != nil {
		return nil, apperror.Wrap(op, err)
	}

	user, err := u.getUserByID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	if user.DeletedAt.Valid {
		return nil, apperror.Wrap(op, apperror.ErrUserProfileDeleted)
	}

	tokens, err := u.createTokens(user)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return buildUserTokenResponse(user, tokens), nil
}

func (u *userService) createTokens(user *models.User) (*dto.TokenResponse, error) {
	const (
		op        = "userService.createTokens"
		tokenType = "Bearer"
	)

	payload := token.Payload{
		UserID:         user.ID.String(),
		UserRole:       string(user.Role),
		TwoFAEnabled:   user.TwoFAEnabled,
		EmailConfirmed: user.EmailConfirmed(),
	}

	accessToken, accessClaims, err := u.tokenManager.GenerateAccessToken(payload)
	if err != nil {
		return nil, apperror.Wrap(op, fmt.Errorf("failed to generate access token: %w", err))
	}

	refreshToken, refreshClaims, err := u.tokenManager.GenerateRefreshToken(payload)
	if err != nil {
		return nil, apperror.Wrap(op, fmt.Errorf("failed to generate refresh token: %w", err))
	}

	return &dto.TokenResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time.Unix(),
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time.Unix(),
		TokenType:             tokenType,
	}, nil
}

func (u *userService) revokeToken(ctx context.Context, tokenID string, tokenType string, expiresAt time.Time) error {
	const op = "userService.revokeToken"

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}

	key := fmt.Sprintf("blacklist:%s:%s", tokenType, tokenID)

	ok, err := u.cache.SetNX(ctx, key, "1", ttl)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !ok {
		return apperror.Wrap(op, apperror.ErrInvalidToken)
	}

	return nil
}

func (u *userService) verifyPassword(password, passwordHash string) error {
	const op = "userService.verifyPassword"

	match, err := u.hasher.Verify(password, passwordHash)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !match {
		return apperror.Wrap(op, apperror.ErrInvalidCredentials)
	}

	return nil
}

func (u *userService) verifyEmailCode(ctx context.Context, userID uuid.UUID, inputCode string) error {
	const op = "userService.verifyCode"

	cacheKey := fmt.Sprintf("email_confirmation:%s", userID)
	code, err := u.cache.Get(ctx, cacheKey)
	if err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
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

func (u *userService) verify2FACode(ctx context.Context, secret string, code string) error {
	const op = "userService.verify2FACode"

	if code == "" {
		return apperror.Wrap(op, apperror.Err2FACodeRequired)
	}

	decryptedKey, err := u.encryptionManager.Decrypt(ctx, secret)
	if err != nil {
		return apperror.Wrap(op, err)
	}

	if !u.totpManager.ValidateCode(decryptedKey, code) {
		return apperror.Wrap(op, apperror.ErrInvalid2FACode)
	}

	return nil
}

func (u *userService) getUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const op = "userService.getUserByID"

	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrUserNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return user, nil
}

func (u *userService) getUserByIDIncludingDeleted(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const op = "userService.getUserByIDIncludingDeleted"

	user, err := u.userRepo.GetByIDIncludingDeleted(ctx, userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, apperror.Wrap(op, apperror.ErrUserNotFound)
		}

		return nil, apperror.Wrap(op, err)
	}

	return user, nil
}

func (u *userService) mapUser(user *models.User) (*dto.UserResponse, error) {
	const op = "userService.mapUser"

	response, err := mapper.MapOne[*models.User, dto.UserResponse](user)
	if err != nil {
		return nil, apperror.Wrap(op, err)
	}

	return response, nil
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
