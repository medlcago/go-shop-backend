package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	tasksMocks "go-shop-backend/internal/tasks/mocks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/cache"
	cacheMocks "go-shop-backend/pkg/cache/mocks"
	cryptoMocks "go-shop-backend/pkg/crypto/mocks"
	hasherMocks "go-shop-backend/pkg/hasher/mocks"
	"go-shop-backend/pkg/token"
	tokenMocks "go-shop-backend/pkg/token/mocks"
	"go-shop-backend/pkg/totp"
	totpMocks "go-shop-backend/pkg/totp/mocks"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	userRepo          *repoMocks.MockUserRepository
	tokenManager      *tokenMocks.MockManager
	hasher            *hasherMocks.MockHasher
	totpManager       *totpMocks.MockManager
	encryptionManager *cryptoMocks.MockEncryptionManager
	notificationTask  *tasksMocks.MockNotificationTask
	cache             *cacheMocks.MockCache
	userEmailConfig   *UserEmailConfig
	userService       *userService

	ctx    context.Context
	userID uuid.UUID
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.userRepo = repoMocks.NewMockUserRepository(suite.T())
	suite.tokenManager = tokenMocks.NewMockManager(suite.T())
	suite.hasher = hasherMocks.NewMockHasher(suite.T())
	suite.totpManager = totpMocks.NewMockManager(suite.T())
	suite.encryptionManager = cryptoMocks.NewMockEncryptionManager(suite.T())
	suite.notificationTask = tasksMocks.NewMockNotificationTask(suite.T())
	suite.cache = cacheMocks.NewMockCache(suite.T())
	suite.userEmailConfig = &UserEmailConfig{
		EmailConfirmationCodeLength: 6,
		EmailConfirmationCodeTTL:    2 * time.Minute,
	}
	suite.userService = NewUserService(
		suite.userRepo,
		suite.tokenManager,
		suite.hasher,
		suite.totpManager,
		suite.encryptionManager,
		suite.notificationTask,
		suite.cache,
		suite.userEmailConfig,
	)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// ==================== Login Tests ====================

func (suite *UserServiceTestSuite) TestLogin_Success() {
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "test123",
	}

	user := &models.User{
		ID:               uuid.New(),
		Email:            req.Email,
		PasswordHash:     "test123",
		Role:             models.UserRoleCustomer,
		TwoFAEnabled:     false,
		EmailConfirmedAt: new(time.Now()),
	}

	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	payload := token.Payload{
		UserID:         user.ID.String(),
		UserRole:       string(user.Role),
		EmailConfirmed: user.EmailConfirmed(),
	}

	claims := &token.UserClaims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute))}}

	suite.tokenManager.EXPECT().GenerateAccessToken(payload).
		Return("test_access_token", claims, nil).Once()

	suite.tokenManager.EXPECT().GenerateRefreshToken(payload).
		Return("test_refresh_token", claims, nil).Once()

	response, err := suite.userService.Login(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.NotNil(response.User)
	suite.Equal(response.User.ID, user.ID)
	suite.Equal(response.User.Email, user.Email)
	suite.True(response.User.EmailConfirmed)
	suite.NotNil(response.TokenResponse)
	suite.Equal(response.AccessToken, "test_access_token")
	suite.Equal(response.RefreshToken, "test_refresh_token")
	suite.Equal("Bearer", response.TokenResponse.TokenType)
	suite.Equal(claims.ExpiresAt.Unix(), response.TokenResponse.AccessTokenExpiresAt)
	suite.Equal(claims.ExpiresAt.Unix(), response.TokenResponse.RefreshTokenExpiresAt)
}

func (suite *UserServiceTestSuite) TestLogin_UserNotFound() {
	req := dto.UserLoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.userService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
	suite.ErrorContains(err, "userService.Login")
}

func (suite *UserServiceTestSuite) TestLogin_ProfileDeleted() {
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := &models.User{
		DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true}),
	}

	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, mock.AnythingOfType("string")).
		Return(true, nil).Once()

	response, err := suite.userService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserProfileDeleted)
	suite.ErrorContains(err, "userService.Login")
}

func (suite *UserServiceTestSuite) TestLogin_ProfileDeleted_2FAEnabled() {
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		Code:     "123456",
	}

	user := &models.User{
		DeletedAt:        gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true}),
		TwoFAEnabled:     true,
		TwoFASecret:      new("secret"),
		TwoFAConfirmedAt: new(time.Now()),
	}

	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, mock.AnythingOfType("string")).
		Return(true, nil).Once()

	suite.encryptionManager.EXPECT().Decrypt(suite.ctx, *user.TwoFASecret).
		Return("decryptedKey", nil).Once()

	suite.totpManager.EXPECT().ValidateCode("decryptedKey", req.Code).
		Return(true).Once()

	response, err := suite.userService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserProfileDeleted)
	suite.ErrorContains(err, "userService.Login")
}

func (suite *UserServiceTestSuite) TestLogin_InvalidPassword() {
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: "test123",
	}

	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, user.PasswordHash).
		Return(false, nil).Once()

	response, err := suite.userService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
	suite.ErrorContains(err, "userService.Login")
}

func (suite *UserServiceTestSuite) TestLogin_RepositoryError() {
	req := dto.UserLoginRequest{
		Email:    "test@test.test",
		Password: "test123",
	}

	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(nil, repoErr).Once()

	response, err := suite.userService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "userService.Login")
}

func (suite *UserServiceTestSuite) TestLogin_2FAEnabled_Success() {
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "test123",
		Code:     "123456",
	}

	user := &models.User{
		ID:               suite.userID,
		Email:            req.Email,
		PasswordHash:     "hashed_password",
		Role:             models.UserRoleCustomer,
		TwoFAEnabled:     true,
		TwoFASecret:      new("encrypted_secret"),
		TwoFAConfirmedAt: new(time.Now().UTC()),
	}

	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	suite.encryptionManager.EXPECT().Decrypt(suite.ctx, *user.TwoFASecret).
		Return("decrypted_secret", nil).Once()

	suite.totpManager.EXPECT().ValidateCode("decrypted_secret", req.Code).
		Return(true).Once()

	payload := token.Payload{
		UserID:       user.ID.String(),
		UserRole:     string(user.Role),
		TwoFAEnabled: user.TwoFAEnabled,
	}

	claims := &token.UserClaims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute))}}

	suite.tokenManager.EXPECT().GenerateAccessToken(payload).
		Return("test_access_token", claims, nil).Once()

	suite.tokenManager.EXPECT().GenerateRefreshToken(payload).
		Return("test_refresh_token", claims, nil).Once()

	response, err := suite.userService.Login(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.NotNil(response.User)
	suite.Equal(response.User.ID, user.ID)
	suite.Equal(response.User.Email, user.Email)
	suite.NotNil(response.TokenResponse)
	suite.Equal(response.AccessToken, "test_access_token")
	suite.Equal(response.RefreshToken, "test_refresh_token")
	suite.Equal("Bearer", response.TokenResponse.TokenType)
	suite.Equal(claims.ExpiresAt.Unix(), response.TokenResponse.AccessTokenExpiresAt)
	suite.Equal(claims.ExpiresAt.Unix(), response.TokenResponse.RefreshTokenExpiresAt)
}

func (suite *UserServiceTestSuite) TestLogin_2FAEnabled_Invalid2FACode() {
	tests := []struct {
		name string
		req  dto.UserLoginRequest
		err  error
	}{
		{
			name: "empty 2fa code",
			req: dto.UserLoginRequest{
				Email:    "test@example.com",
				Password: "test123",
			},
			err: apperror.Err2FACodeRequired,
		},
		{
			name: "invalid 2fa code",
			req: dto.UserLoginRequest{
				Email:    "test@example.com",
				Password: "test123",
				Code:     "123456",
			},
			err: apperror.ErrInvalid2FACode,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			user := &models.User{
				ID:               suite.userID,
				Email:            tt.req.Email,
				PasswordHash:     "hashed_password",
				Role:             models.UserRoleCustomer,
				TwoFAEnabled:     true,
				TwoFASecret:      new("encrypted_secret"),
				TwoFAConfirmedAt: new(time.Now().UTC()),
			}

			suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, tt.req.Email).
				Return(user, nil).Once()

			suite.hasher.EXPECT().Verify(tt.req.Password, user.PasswordHash).
				Return(true, nil).Once()

			if tt.req.Code != "" {
				suite.encryptionManager.EXPECT().Decrypt(suite.ctx, *user.TwoFASecret).
					Return("decrypted_secret", nil).Once()

				suite.totpManager.EXPECT().ValidateCode("decrypted_secret", tt.req.Code).
					Return(false).Once()
			}

			response, err := suite.userService.Login(suite.ctx, tt.req)

			suite.Nil(response)
			suite.ErrorIs(err, tt.err)
			suite.ErrorContains(err, "userService.Login")
		})
	}
}

// ==================== Register Tests ====================

func (suite *UserServiceTestSuite) TestRegister_Success() {
	req := dto.UserRegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().ExistsByEmail(suite.ctx, req.Email).
		Return(false, nil).Once()

	suite.hasher.EXPECT().Hash(req.Password).
		Return("test123", nil).Once()

	suite.userRepo.EXPECT().Create(suite.ctx, mock.MatchedBy(func(user *models.User) bool {
		user.ID = suite.userID
		return user.Email == req.Email && user.PasswordHash == "test123" && user.Role == models.UserRoleCustomer
	})).Return(nil).Once()

	payload := token.Payload{
		UserID:   suite.userID.String(),
		UserRole: string(models.UserRoleCustomer),
	}

	claims := &token.UserClaims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute))}}

	suite.tokenManager.EXPECT().GenerateAccessToken(payload).
		Return("test_access_token", claims, nil).Once()

	suite.tokenManager.EXPECT().GenerateRefreshToken(payload).
		Return("test_refresh_token", claims, nil).Once()

	response, err := suite.userService.Register(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.NotNil(response.User)
	suite.Equal(response.User.ID, suite.userID)
	suite.Equal(response.User.Email, req.Email)
	suite.NotNil(response.TokenResponse)
	suite.Equal(response.AccessToken, "test_access_token")
	suite.Equal(response.RefreshToken, "test_refresh_token")
	suite.Equal("Bearer", response.TokenResponse.TokenType)
	suite.Equal(claims.ExpiresAt.Unix(), response.TokenResponse.AccessTokenExpiresAt)
	suite.Equal(claims.ExpiresAt.Unix(), response.TokenResponse.RefreshTokenExpiresAt)
}

func (suite *UserServiceTestSuite) TestRegister_EmailAlreadyExists() {
	req := dto.UserRegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().ExistsByEmail(suite.ctx, req.Email).
		Return(true, nil).Once()

	response, err := suite.userService.Register(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrEmailTaken)
	suite.ErrorContains(err, "userService.Register")
}

func (suite *UserServiceTestSuite) TestRegister_RepositoryError() {
	req := dto.UserRegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().ExistsByEmail(suite.ctx, req.Email).
		Return(false, repoErr).Once()

	response, err := suite.userService.Register(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "userService.Register")
}

// ==================== Setup2FA Tests ====================

func (suite *UserServiceTestSuite) TestSetup2FA_Success() {
	user := &models.User{
		ID:    suite.userID,
		Email: "test@example.com",
	}

	secret := "secret123"
	encrypted := "encrypted_secret"

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.totpManager.EXPECT().
		GenerateSecret(user.Email).
		Return(&totp.Secret{Secret: secret, QRCode: "qr"}, nil).Once()

	suite.encryptionManager.EXPECT().
		Encrypt(suite.ctx, secret).
		Return(encrypted, nil).Once()

	suite.userRepo.EXPECT().
		Update(suite.ctx, mock.MatchedBy(func(user *models.User) bool {
			return user.TwoFASecret != nil && *user.TwoFASecret == encrypted
		})).
		Return(nil).Once()

	response, err := suite.userService.Setup2FA(suite.ctx, suite.userID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(secret, response.Secret)
	suite.Equal("qr", response.QRCode)
}

func (suite *UserServiceTestSuite) TestSetup2FA_2FAAlreadyEnabled() {
	user := &models.User{
		ID:               suite.userID,
		TwoFAEnabled:     true,
		TwoFASecret:      new("secret"),
		TwoFAConfirmedAt: new(time.Now().UTC()),
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	response, err := suite.userService.Setup2FA(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.Err2FAAlreadyEnabled)
	suite.ErrorContains(err, "userService.Setup2FA")
}

func (suite *UserServiceTestSuite) TestSetup2FA_EncryptionError() {
	user := &models.User{
		ID:    suite.userID,
		Email: "test@test.com",
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.totpManager.EXPECT().
		GenerateSecret(user.Email).
		Return(&totp.Secret{Secret: "secret", QRCode: "qr"}, nil).Once()

	encryptionErr := errors.New("encryption error")
	suite.encryptionManager.EXPECT().
		Encrypt(suite.ctx, "secret").
		Return("", encryptionErr).Once()

	response, err := suite.userService.Setup2FA(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, encryptionErr)
	suite.ErrorContains(err, "userService.Setup2FA")
}

func (suite *UserServiceTestSuite) TestSetup2FA_UserNotFound() {
	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.userService.Setup2FA(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserNotFound)
	suite.ErrorContains(err, "userService.Setup2FA")
}

// ==================== Confirm2FA Tests ====================

func (suite *UserServiceTestSuite) TestConfirm2FA_Success() {
	secret := "encrypted"
	decrypted := "decrypted"

	user := &models.User{
		ID:           suite.userID,
		PasswordHash: "password_hash",
		TwoFASecret:  new(secret),
	}

	req := dto.Confirm2FARequest{
		Password: "password123",
		Code:     "123456",
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	suite.encryptionManager.EXPECT().
		Decrypt(suite.ctx, secret).
		Return(decrypted, nil).Once()

	suite.totpManager.EXPECT().
		ValidateCode(decrypted, req.Code).
		Return(true).Once()

	suite.userRepo.EXPECT().
		Update(suite.ctx, mock.MatchedBy(func(user *models.User) bool {
			return user.TwoFAEnabled && user.TwoFAConfirmedAt != nil
		})).
		Return(nil).Once()

	err := suite.userService.Confirm2FA(suite.ctx, suite.userID, req)

	suite.NoError(err)
	suite.True(user.TwoFAEnabled)
	suite.NotNil(user.TwoFAConfirmedAt)
}

func (suite *UserServiceTestSuite) TestConfirm2FA_2FANotInitialized() {
	user := &models.User{
		PasswordHash: "password_hash",
	}

	req := dto.Confirm2FARequest{Password: "123123"}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	err := suite.userService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.Err2FANotInitialized)
	suite.ErrorContains(err, "userService.Confirm2FA")
}

func (suite *UserServiceTestSuite) TestConfirm2FA_2FAAlreadyEnabled() {
	user := &models.User{
		PasswordHash:     "password_hash",
		TwoFASecret:      new("secret"),
		TwoFAEnabled:     true,
		TwoFAConfirmedAt: new(time.Now().UTC()),
	}

	req := dto.Confirm2FARequest{Password: "123123"}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	err := suite.userService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.Err2FAAlreadyEnabled)
	suite.ErrorContains(err, "userService.Confirm2FA")
}

func (suite *UserServiceTestSuite) TestConfirm2FA_InvalidPassword() {
	user := &models.User{
		PasswordHash: "password_hash",
	}

	req := dto.Confirm2FARequest{Password: "wrong"}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(false, nil).Once()

	err := suite.userService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
	suite.ErrorContains(err, "userService.Confirm2FA")
}

func (suite *UserServiceTestSuite) TestConfirm2FA_InvalidCode() {
	secret := "enc"

	user := &models.User{
		PasswordHash: "password_hash",
		TwoFASecret:  new(secret),
	}

	req := dto.Confirm2FARequest{
		Password: "pass",
		Code:     "000000",
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	suite.encryptionManager.EXPECT().
		Decrypt(suite.ctx, secret).
		Return("real", nil).Once()

	suite.totpManager.EXPECT().
		ValidateCode("real", req.Code).
		Return(false).Once()

	err := suite.userService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrInvalid2FACode)
	suite.ErrorContains(err, "userService.Confirm2FA")
}

func (suite *UserServiceTestSuite) TestConfirm2FA_UserNotFound() {
	req := dto.Confirm2FARequest{
		Password: "pass",
		Code:     "000000",
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	err := suite.userService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrUserNotFound)
	suite.ErrorContains(err, "userService.Confirm2FA")
}

// ==================== Disable2FA Tests ====================

func (suite *UserServiceTestSuite) TestDisable2FA_Success() {
	secret := "enc"

	user := &models.User{
		PasswordHash: "password_hash",
		TwoFAEnabled: true,
		TwoFASecret:  new(secret),
	}

	req := dto.Disable2FARequest{
		Password: "password123",
		Code:     "123456",
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	suite.encryptionManager.EXPECT().
		Decrypt(suite.ctx, secret).
		Return("real", nil).Once()

	suite.totpManager.EXPECT().
		ValidateCode("real", req.Code).
		Return(true).Once()

	suite.userRepo.EXPECT().
		Update(suite.ctx, mock.MatchedBy(func(user *models.User) bool {
			return !user.TwoFAEnabled && user.TwoFASecret == nil && user.TwoFAConfirmedAt == nil
		})).
		Return(nil).Once()

	err := suite.userService.Disable2FA(suite.ctx, suite.userID, req)

	suite.NoError(err)
	suite.False(user.TwoFAEnabled)
	suite.Nil(user.TwoFASecret)
	suite.Nil(user.TwoFAConfirmedAt)
}

func (suite *UserServiceTestSuite) TestDisable2FA_2FANotEnabled() {
	user := &models.User{}

	req := dto.Disable2FARequest{Password: "123123"}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	err := suite.userService.Disable2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.Err2FANotEnabled)
	suite.ErrorContains(err, "userService.Disable2FA")
}

func (suite *UserServiceTestSuite) TestDisable2FA_InvalidPassword() {
	user := &models.User{
		PasswordHash: "password_hash",
	}

	req := dto.Disable2FARequest{Password: "wrong"}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(false, nil).Once()

	err := suite.userService.Disable2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
	suite.ErrorContains(err, "userService.Disable2FA")
}

func (suite *UserServiceTestSuite) TestDisable2FA_InvalidCode() {
	secret := "secret"
	decrypted := "decrypted"

	user := &models.User{
		ID:               suite.userID,
		PasswordHash:     "password_hash",
		TwoFAEnabled:     true,
		TwoFASecret:      new(secret),
		TwoFAConfirmedAt: new(time.Now().UTC()),
	}

	req := dto.Disable2FARequest{
		Password: "password123",
		Code:     "123456",
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	suite.encryptionManager.EXPECT().
		Decrypt(suite.ctx, secret).
		Return(decrypted, nil).Once()

	suite.totpManager.EXPECT().
		ValidateCode(decrypted, req.Code).
		Return(false).Once()

	err := suite.userService.Disable2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrInvalid2FACode)
	suite.ErrorContains(err, "userService.Disable2FA")
}

func (suite *UserServiceTestSuite) TestDisable2FA_UserNotFound() {
	req := dto.Disable2FARequest{
		Password: "password123",
		Code:     "123456",
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	err := suite.userService.Disable2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrUserNotFound)
	suite.ErrorContains(err, "userService.Disable2FA")
}

// ==================== GetUserByID Tests ====================

func (suite *UserServiceTestSuite) TestGetUserByID_Success() {
	mockUser := &models.User{
		ID:        suite.userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	suite.userRepo.EXPECT().GetByIDIncludingDeleted(suite.ctx, suite.userID).
		Return(mockUser, nil).Once()

	response, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(suite.userID, response.ID)
	suite.Equal(mockUser.Email, response.Email)
	suite.NotZero(response.CreatedAt)
}

func (suite *UserServiceTestSuite) TestGetUserByID_NotFound() {
	suite.userRepo.EXPECT().GetByIDIncludingDeleted(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserNotFound)
	suite.ErrorContains(err, "userService.GetUserByID")
}

func (suite *UserServiceTestSuite) TestGetUserByID_ProfileDeleted() {
	suite.userRepo.EXPECT().GetByIDIncludingDeleted(suite.ctx, suite.userID).
		Return(&models.User{ID: suite.userID, DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	response, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserProfileDeleted)
	suite.ErrorContains(err, "userService.GetUserByID")
}

func (suite *UserServiceTestSuite) TestGetUserByID_RepositoryError() {
	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().GetByIDIncludingDeleted(suite.ctx, suite.userID).
		Return(nil, repoErr).Once()

	response, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, repoErr)
	suite.ErrorContains(err, "userService.GetUserByID")
}

// ==================== SendEmailConfirmationCode Tests ====================

func (suite *UserServiceTestSuite) TestSendEmailConfirmationCode_Success() {
	user := &models.User{
		ID:    suite.userID,
		Email: "test@example.com",
	}

	cacheKey := fmt.Sprintf("email_confirmation:%s", suite.userID)

	suite.cache.EXPECT().Exists(suite.ctx, cacheKey).
		Return(false, nil).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.cache.EXPECT().Set(suite.ctx, cacheKey, mock.AnythingOfType("string"), suite.userEmailConfig.EmailConfirmationCodeTTL).
		Return(nil).Once()

	suite.notificationTask.EXPECT().SendEmailConfirmationCode(suite.ctx, user.Email, mock.AnythingOfType("string")).
		Return(nil).Once()

	response, err := suite.userService.SendEmailConfirmationCode(suite.ctx, suite.userID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(response.ExpiresIn, int(suite.userEmailConfig.EmailConfirmationCodeTTL.Seconds()))
}

func (suite *UserServiceTestSuite) TestSendEmailConfirmationCode_EmailConfirmationCodeAlreadySent() {
	cacheKey := fmt.Sprintf("email_confirmation:%s", suite.userID)

	suite.cache.EXPECT().Exists(suite.ctx, cacheKey).
		Return(true, nil).Once()

	response, err := suite.userService.SendEmailConfirmationCode(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrEmailConfirmationCodeAlreadySent)
	suite.ErrorContains(err, "userService.SendEmailConfirmationCode")
}

func (suite *UserServiceTestSuite) TestSendEmailConfirmationCode_EmailAlreadyConfirmed() {
	user := &models.User{
		ID:               suite.userID,
		Email:            "test@example.com",
		EmailConfirmedAt: new(time.Now()),
	}

	cacheKey := fmt.Sprintf("email_confirmation:%s", suite.userID)

	suite.cache.EXPECT().Exists(suite.ctx, cacheKey).
		Return(false, nil).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	response, err := suite.userService.SendEmailConfirmationCode(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrEmailAlreadyConfirmed)
	suite.ErrorContains(err, "userService.SendEmailConfirmationCode")
}

func (suite *UserServiceTestSuite) TestSendEmailConfirmationCode_UserNotFound() {
	suite.cache.EXPECT().Exists(suite.ctx, mock.AnythingOfType("string")).
		Return(false, nil).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.userService.SendEmailConfirmationCode(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserNotFound)
	suite.ErrorContains(err, "userService.SendEmailConfirmationCode")
}

func (suite *UserServiceTestSuite) TestSendEmailConfirmationCode_InternalError() {
	cacheErr := errors.New("cache error")
	suite.cache.EXPECT().Exists(suite.ctx, mock.AnythingOfType("string")).
		Return(false, cacheErr).Once()

	response, err := suite.userService.SendEmailConfirmationCode(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, cacheErr)
	suite.ErrorContains(err, "userService.SendEmailConfirmationCode")
}

// ==================== ConfirmEmail Tests ====================

func (suite *UserServiceTestSuite) TestConfirmEmail_Success() {
	req := dto.ConfirmEmailRequest{
		Code: "123456",
	}

	user := &models.User{
		ID:    suite.userID,
		Email: "test@example.com",
	}

	cacheKey := fmt.Sprintf("email_confirmation:%s", suite.userID)

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.cache.EXPECT().Get(suite.ctx, cacheKey).
		Return(req.Code, nil).Once()

	suite.cache.EXPECT().Delete(suite.ctx, cacheKey).
		Return(nil).Once()

	suite.userRepo.EXPECT().Update(suite.ctx, mock.MatchedBy(func(user *models.User) bool {
		return user.EmailConfirmed()
	})).
		Return(nil).Once()

	response, err := suite.userService.ConfirmEmail(suite.ctx, suite.userID, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.True(response.OK)
	suite.NotNil(user.EmailConfirmedAt)
	suite.Equal(user.EmailConfirmedAt.Format(time.RFC3339), response.EmailConfirmedAt)
}

func (suite *UserServiceTestSuite) TestConfirmEmail_EmailAlreadyConfirmed() {
	req := dto.ConfirmEmailRequest{
		Code: "123456",
	}

	user := &models.User{
		ID:               suite.userID,
		Email:            "test@example.com",
		EmailConfirmedAt: new(time.Now()),
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	response, err := suite.userService.ConfirmEmail(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrEmailAlreadyConfirmed)
	suite.ErrorContains(err, "userService.ConfirmEmail")
}

func (suite *UserServiceTestSuite) TestConfirmEmail_InvalidCode_NotFoundInCache() {
	req := dto.ConfirmEmailRequest{
		Code: "123456",
	}

	user := &models.User{
		ID:    suite.userID,
		Email: "test@example.com",
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.cache.EXPECT().Get(suite.ctx, mock.AnythingOfType("string")).
		Return("", cache.ErrCacheMiss).Once()

	response, err := suite.userService.ConfirmEmail(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidCode)
	suite.ErrorContains(err, "userService.ConfirmEmail")
}

func (suite *UserServiceTestSuite) TestConfirmEmail_InvalidCode() {
	req := dto.ConfirmEmailRequest{
		Code: "123456",
	}

	user := &models.User{
		ID:    suite.userID,
		Email: "test@example.com",
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.cache.EXPECT().Get(suite.ctx, mock.AnythingOfType("string")).
		Return("123321", nil).Once()

	response, err := suite.userService.ConfirmEmail(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidCode)
	suite.ErrorContains(err, "userService.ConfirmEmail")
}

func (suite *UserServiceTestSuite) TestConfirmEmail_UserNotFound() {
	req := dto.ConfirmEmailRequest{
		Code: "123456",
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.userService.ConfirmEmail(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserNotFound)
	suite.ErrorContains(err, "userService.ConfirmEmail")
}

// ==================== ChangePassword Tests ====================

func (suite *UserServiceTestSuite) TestChangePassword_Success() {
	req := dto.ChangePasswordRequest{
		Password:    "oldPassword123",
		NewPassword: "newPassword123",
	}

	user := &models.User{
		ID:           suite.userID,
		PasswordHash: "passwordHash",
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	suite.hasher.EXPECT().Hash(req.NewPassword).
		Return("newPasswordHash", nil).Once()

	suite.userRepo.EXPECT().Update(suite.ctx, mock.MatchedBy(func(user *models.User) bool {
		return user.PasswordHash == "newPasswordHash"
	})).Return(nil).Once()

	err := suite.userService.ChangePassword(suite.ctx, suite.userID, req)
	suite.NoError(err)
}

func (suite *UserServiceTestSuite) TestChangePassword_UserNotFound() {
	req := dto.ChangePasswordRequest{
		Password:    "oldPassword123",
		NewPassword: "newPassword123",
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	err := suite.userService.ChangePassword(suite.ctx, suite.userID, req)

	suite.ErrorIs(err, apperror.ErrUserNotFound)
	suite.ErrorContains(err, "userService.ChangePassword")
}

func (suite *UserServiceTestSuite) TestChangePassword_InvalidCurrentPassword() {
	req := dto.ChangePasswordRequest{
		Password:    "oldPassword123",
		NewPassword: "newPassword123",
	}

	user := &models.User{
		ID:           suite.userID,
		PasswordHash: "passwordHash",
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, user.PasswordHash).
		Return(false, nil).Once()

	err := suite.userService.ChangePassword(suite.ctx, suite.userID, req)

	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
	suite.ErrorContains(err, "userService.ChangePassword")
}

func (suite *UserServiceTestSuite) TestChangePassword_2FAEnabled_Success() {
	req := dto.ChangePasswordRequest{
		Password:    "oldPassword123",
		NewPassword: "newPassword123",
		Code:        "123456",
	}

	user := &models.User{
		ID:           suite.userID,
		PasswordHash: "passwordHash",
		TwoFAEnabled: true,
		TwoFASecret:  new("secret"),
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	suite.encryptionManager.EXPECT().Decrypt(suite.ctx, *user.TwoFASecret).
		Return("decryptedSecret", nil).Once()

	suite.totpManager.EXPECT().ValidateCode("decryptedSecret", req.Code).
		Return(true).Once()

	suite.hasher.EXPECT().Hash(req.NewPassword).
		Return("newPasswordHash", nil).Once()

	suite.userRepo.EXPECT().Update(suite.ctx, mock.MatchedBy(func(user *models.User) bool {
		return user.PasswordHash == "newPasswordHash"
	})).Return(nil).Once()

	err := suite.userService.ChangePassword(suite.ctx, suite.userID, req)
	suite.NoError(err)
}

func (suite *UserServiceTestSuite) TestChangePassword_2FAEnabled_InvalidCode() {
	req := dto.ChangePasswordRequest{
		Password:    "oldPassword123",
		NewPassword: "newPassword123",
		Code:        "123456",
	}

	user := &models.User{
		ID:           suite.userID,
		PasswordHash: "passwordHash",
		TwoFAEnabled: true,
		TwoFASecret:  new("secret"),
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	suite.encryptionManager.EXPECT().Decrypt(suite.ctx, *user.TwoFASecret).
		Return("decryptedSecret", nil).Once()

	suite.totpManager.EXPECT().ValidateCode("decryptedSecret", req.Code).
		Return(false).Once()

	err := suite.userService.ChangePassword(suite.ctx, suite.userID, req)

	suite.ErrorIs(err, apperror.ErrInvalid2FACode)
	suite.ErrorContains(err, "userService.ChangePassword")
}

// ==================== RefreshToken Tests ====================

func (suite *UserServiceTestSuite) TestRefreshToken_Success() {
	tokenString := "refresh-token"

	user := &models.User{
		ID:        suite.userID,
		Email:     "user@example.com",
		CreatedAt: time.Now(),
		Role:      models.UserRoleCustomer,
	}

	payload := token.Payload{
		UserID:         user.ID.String(),
		UserRole:       string(user.Role),
		TwoFAEnabled:   user.TwoFAEnabled,
		EmailConfirmed: user.EmailConfirmed(),
	}

	userClaims := &token.UserClaims{
		Payload:   payload,
		TokenType: token.RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	claims := &token.UserClaims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute))}}

	key := fmt.Sprintf("blacklist:%s:%s", token.RefreshTokenType, userClaims.ID)

	suite.tokenManager.EXPECT().ValidateToken(tokenString).
		Return(userClaims, nil).Once()

	suite.cache.EXPECT().SetNX(suite.ctx, key, "1", mock.AnythingOfType("time.Duration")).
		Return(true, nil).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.tokenManager.EXPECT().GenerateAccessToken(payload).
		Return("access-token", claims, nil).Once()

	suite.tokenManager.EXPECT().GenerateRefreshToken(payload).
		Return("refresh-token", claims, nil).Once()

	response, err := suite.userService.RefreshToken(suite.ctx, tokenString)

	suite.NoError(err)
	suite.NotNil(response)
	suite.NotNil(response.TokenResponse)
	suite.Equal(response.TokenResponse.AccessToken, "access-token")
	suite.Equal(response.TokenResponse.RefreshToken, "refresh-token")
	suite.Equal(response.TokenResponse.TokenType, "Bearer")
	suite.Equal(claims.ExpiresAt.Unix(), response.TokenResponse.AccessTokenExpiresAt)
	suite.Equal(claims.ExpiresAt.Unix(), response.TokenResponse.RefreshTokenExpiresAt)
	suite.NotNil(response.User)
	suite.Equal(response.User.ID, user.ID)
}

func (suite *UserServiceTestSuite) TestRefreshToken_InvalidToken() {
	tokenString := "invalid-token"

	suite.tokenManager.EXPECT().ValidateToken(tokenString).
		Return(nil, &token.ErrInvalidToken{}).Once()

	response, err := suite.userService.RefreshToken(suite.ctx, tokenString)

	suite.Error(err)
	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidToken)
	suite.ErrorContains(err, "userService.RefreshToken")
}

func (suite *UserServiceTestSuite) TestRefreshToken_InvalidTokenType() {
	tokenString := "access-token"

	claims := &token.UserClaims{
		TokenType: token.AccessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: uuid.NewString(),
		},
	}

	suite.tokenManager.EXPECT().ValidateToken(tokenString).
		Return(claims, nil).Once()

	response, err := suite.userService.RefreshToken(suite.ctx, tokenString)

	suite.Error(err)
	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidTokenType)
	suite.ErrorContains(err, "userService.RefreshToken")
}

func (suite *UserServiceTestSuite) TestRefreshToken_InvalidUserID() {
	tokenString := "refresh-token"

	claims := &token.UserClaims{
		Payload: token.Payload{
			UserID: "invalid-uuid",
		},
		TokenType: token.RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID: uuid.NewString(),
		},
	}

	suite.tokenManager.EXPECT().ValidateToken(tokenString).
		Return(claims, nil).Once()

	response, err := suite.userService.RefreshToken(suite.ctx, tokenString)

	suite.Error(err)
	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidToken)
	suite.ErrorContains(err, "userService.RefreshToken")
}

func (suite *UserServiceTestSuite) TestRefreshToken_TokenAlreadyRevoked() {
	tokenString := "refresh-token"

	claims := &token.UserClaims{
		Payload: token.Payload{
			UserID: suite.userID.String(),
		},
		TokenType: token.RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	key := fmt.Sprintf("blacklist:%s:%s", token.RefreshTokenType, claims.ID)

	suite.tokenManager.EXPECT().ValidateToken(tokenString).
		Return(claims, nil).Once()

	suite.cache.EXPECT().SetNX(suite.ctx, key, "1", mock.AnythingOfType("time.Duration")).
		Return(false, nil).Once()

	response, err := suite.userService.RefreshToken(suite.ctx, tokenString)

	suite.Error(err)
	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidToken)
	suite.ErrorContains(err, "userService.RefreshToken")
}

func (suite *UserServiceTestSuite) TestRefreshToken_UserNotFound() {
	tokenString := "refresh-token"

	claims := &token.UserClaims{
		Payload: token.Payload{
			UserID: suite.userID.String(),
		},
		TokenType: token.RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	key := fmt.Sprintf("blacklist:%s:%s", token.RefreshTokenType, claims.ID)

	suite.tokenManager.EXPECT().ValidateToken(tokenString).
		Return(claims, nil).Once()

	suite.cache.EXPECT().SetNX(suite.ctx, key, "1", mock.AnythingOfType("time.Duration")).
		Return(true, nil).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.userService.RefreshToken(suite.ctx, tokenString)

	suite.Error(err)
	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserNotFound)
	suite.ErrorContains(err, "userService.RefreshToken")
}

func (suite *UserServiceTestSuite) TestRefreshToken_UserProfileDeleted() {
	tokenString := "refresh-token"

	claims := &token.UserClaims{
		Payload: token.Payload{
			UserID: suite.userID.String(),
		},
		TokenType: token.RefreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	user := &models.User{
		ID:        suite.userID,
		DeletedAt: gorm.DeletedAt{Time: time.Now(), Valid: true},
	}

	key := fmt.Sprintf("blacklist:%s:%s", token.RefreshTokenType, claims.ID)

	suite.tokenManager.EXPECT().ValidateToken(tokenString).
		Return(claims, nil).Once()

	suite.cache.EXPECT().SetNX(suite.ctx, key, "1", mock.AnythingOfType("time.Duration")).
		Return(true, nil).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	response, err := suite.userService.RefreshToken(suite.ctx, tokenString)

	suite.Error(err)
	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserProfileDeleted)
	suite.ErrorContains(err, "userService.RefreshToken")
}
