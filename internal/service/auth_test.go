package service

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperror"
	cryptoMocks "go-shop-backend/pkg/crypto/mocks"
	"go-shop-backend/pkg/database"
	hasherMocks "go-shop-backend/pkg/hasher/mocks"
	"go-shop-backend/pkg/token"
	tokenMocks "go-shop-backend/pkg/token/mocks"
	"go-shop-backend/pkg/totp"
	totpMocks "go-shop-backend/pkg/totp/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AuthServiceTestSuite struct {
	suite.Suite
	userRepo          *repoMocks.MockUserRepository
	tokenManager      *tokenMocks.MockManager
	hasher            *hasherMocks.MockHasher
	totpManager       *totpMocks.MockManager
	encryptionManager *cryptoMocks.MockEncryptionManager
	txManager         *database.NoopTxManager
	authService       *authService

	ctx    context.Context
	userID uuid.UUID
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.userRepo = repoMocks.NewMockUserRepository(suite.T())
	suite.tokenManager = tokenMocks.NewMockManager(suite.T())
	suite.hasher = hasherMocks.NewMockHasher(suite.T())
	suite.totpManager = totpMocks.NewMockManager(suite.T())
	suite.encryptionManager = cryptoMocks.NewMockEncryptionManager(suite.T())
	suite.authService = NewAuthService(
		suite.userRepo,
		suite.tokenManager,
		suite.hasher,
		suite.totpManager,
		suite.encryptionManager,
		suite.txManager,
	)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

// ==================== Login Tests ====================

func (suite *AuthServiceTestSuite) TestLogin_Success() {
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

	suite.tokenManager.EXPECT().GenerateAccessToken(payload).
		Return("test_access_token", nil).Once()

	suite.tokenManager.EXPECT().GenerateRefreshToken(payload).
		Return("test_refresh_token", nil).Once()

	response, err := suite.authService.Login(suite.ctx, req)

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
}

func (suite *AuthServiceTestSuite) TestLogin_UserNotFound() {
	req := dto.UserLoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.authService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
}

func (suite *AuthServiceTestSuite) TestLogin_ProfileDeleted() {
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(&models.User{DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	suite.hasher.EXPECT().Verify(req.Password, mock.Anything).
		Return(true, nil).Once()

	response, err := suite.authService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserProfileDeleted)
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidPassword() {
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

	response, err := suite.authService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
}

func (suite *AuthServiceTestSuite) TestLogin_RepositoryError() {
	req := dto.UserLoginRequest{
		Email:    "test@test.test",
		Password: "test123",
	}

	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().GetByEmailIncludingDeleted(suite.ctx, req.Email).
		Return(nil, repoErr).Once()

	response, err := suite.authService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, repoErr)
}

func (suite *AuthServiceTestSuite) TestLogin_2FAEnabled_Success() {
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "test123",
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

	payload := token.Payload{
		UserID:       user.ID.String(),
		UserRole:     string(user.Role),
		TwoFAEnabled: user.TwoFAEnabled,
	}

	suite.tokenManager.EXPECT().GeneratePartialToken(payload).
		Return("partial_token", nil).Once()

	response, err := suite.authService.Login(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Nil(response.User)
	suite.True(response.Requires2FA)
	suite.Equal("partial_token", response.PartialToken)
	suite.Empty(response.AccessToken)
	suite.Empty(response.RefreshToken)
	suite.Empty(response.TokenResponse.TokenType)
}

func (suite *AuthServiceTestSuite) TestLogin_2FAEnabled_PartialTokenError() {
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "test123",
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

	payload := token.Payload{
		UserID:       user.ID.String(),
		UserRole:     string(user.Role),
		TwoFAEnabled: user.TwoFAEnabled,
	}

	tokenErr := errors.New("token generation failed")
	suite.tokenManager.EXPECT().GeneratePartialToken(payload).
		Return("", tokenErr).Once()

	response, err := suite.authService.Login(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorContains(err, tokenErr.Error())
}

// ==================== Register Tests ====================

func (suite *AuthServiceTestSuite) TestRegister_Success() {
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

	suite.tokenManager.EXPECT().GenerateAccessToken(payload).
		Return("test_access_token", nil).Once()

	suite.tokenManager.EXPECT().GenerateRefreshToken(payload).
		Return("test_refresh_token", nil).Once()

	response, err := suite.authService.Register(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.NotNil(response.User)
	suite.Equal(response.User.ID, suite.userID)
	suite.Equal(response.User.Email, req.Email)
	suite.NotNil(response.TokenResponse)
	suite.Equal(response.AccessToken, "test_access_token")
	suite.Equal(response.RefreshToken, "test_refresh_token")
	suite.Equal("Bearer", response.TokenResponse.TokenType)
}

func (suite *AuthServiceTestSuite) TestRegister_EmailAlreadyExists() {
	req := dto.UserRegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().ExistsByEmail(suite.ctx, req.Email).
		Return(true, nil).Once()

	response, err := suite.authService.Register(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrEmailTaken)
}

func (suite *AuthServiceTestSuite) TestRegister_RepositoryError() {
	req := dto.UserRegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().ExistsByEmail(suite.ctx, req.Email).
		Return(false, repoErr).Once()

	response, err := suite.authService.Register(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, repoErr)
}

// ==================== Setup2FA Tests ====================

func (suite *AuthServiceTestSuite) TestSetup2FA_Success() {
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

	response, err := suite.authService.Setup2FA(suite.ctx, suite.userID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(secret, response.Secret)
	suite.Equal("qr", response.QRCode)
}

func (suite *AuthServiceTestSuite) TestSetup2FA_2FAAlreadyEnabled() {
	user := &models.User{
		ID:               suite.userID,
		TwoFAEnabled:     true,
		TwoFASecret:      new("secret"),
		TwoFAConfirmedAt: new(time.Now().UTC()),
	}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	response, err := suite.authService.Setup2FA(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.Err2FAAlreadyEnabled)
}

func (suite *AuthServiceTestSuite) TestSetup2FA_EncryptionError() {
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

	response, err := suite.authService.Setup2FA(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, encryptionErr)
}

// ==================== Confirm2FA Tests ====================

func (suite *AuthServiceTestSuite) TestConfirm2FA_Success() {
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

	err := suite.authService.Confirm2FA(suite.ctx, suite.userID, req)

	suite.NoError(err)
	suite.True(user.TwoFAEnabled)
	suite.NotNil(user.TwoFAConfirmedAt)
}

func (suite *AuthServiceTestSuite) TestConfirm2FA_2FANotInitialized() {
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

	err := suite.authService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.Err2FANotInitialized)
}

func (suite *AuthServiceTestSuite) TestConfirm2FA_2FAAlreadyEnabled() {
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

	err := suite.authService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.Err2FAAlreadyEnabled)
}

func (suite *AuthServiceTestSuite) TestConfirm2FA_InvalidPassword() {
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

	err := suite.authService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
}

func (suite *AuthServiceTestSuite) TestConfirm2FA_InvalidCode() {
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

	err := suite.authService.Confirm2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrInvalid2FACode)
}

// ==================== Disable2FA Tests ====================

func (suite *AuthServiceTestSuite) TestDisable2FA_Success() {
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

	err := suite.authService.Disable2FA(suite.ctx, suite.userID, req)

	suite.NoError(err)
	suite.False(user.TwoFAEnabled)
	suite.Nil(user.TwoFASecret)
	suite.Nil(user.TwoFAConfirmedAt)
}

func (suite *AuthServiceTestSuite) TestDisable2FA_2FANotEnabled() {
	user := &models.User{}

	req := dto.Disable2FARequest{Password: "123123"}

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.hasher.EXPECT().
		Verify(req.Password, user.PasswordHash).
		Return(true, nil).Once()

	err := suite.authService.Disable2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.Err2FANotEnabled)
}

func (suite *AuthServiceTestSuite) TestDisable2FA_InvalidPassword() {
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

	err := suite.authService.Disable2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrInvalidCredentials)
}

func (suite *AuthServiceTestSuite) TestDisable2FA_InvalidCode() {
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

	err := suite.authService.Disable2FA(suite.ctx, suite.userID, req)
	suite.ErrorIs(err, apperror.ErrInvalid2FACode)
}

// ==================== Verify2FA Tests ====================

func (suite *AuthServiceTestSuite) TestVerify2FA_Success() {
	user := &models.User{
		ID:               suite.userID,
		TwoFAEnabled:     true,
		TwoFASecret:      new("secret"),
		TwoFAConfirmedAt: new(time.Now().UTC()),
		Role:             models.UserRoleCustomer,
	}

	req := dto.Verify2FARequest{
		Token: "partial",
		Code:  "123456",
	}

	claims := &token.UserClaims{
		UserID:       user.ID.String(),
		UserRole:     string(user.Role),
		TwoFAEnabled: true,
		TokenType:    token.PartialTokenType,
	}

	suite.tokenManager.EXPECT().
		ValidateToken(req.Token).
		Return(claims, nil).Once()

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.encryptionManager.EXPECT().
		Decrypt(suite.ctx, "secret").
		Return("real", nil).Once()

	suite.totpManager.EXPECT().
		ValidateCode("real", req.Code).
		Return(true).Once()

	suite.tokenManager.EXPECT().
		GenerateAccessToken(mock.Anything).
		Return("access", nil).Once()

	suite.tokenManager.EXPECT().
		GenerateRefreshToken(mock.Anything).
		Return("refresh", nil).Once()

	response, err := suite.authService.Verify2FA(suite.ctx, req)

	suite.NoError(err)
	suite.NotNil(response)
	suite.False(response.Requires2FA)
	suite.NotNil(response.User)
	suite.Equal(user.ID, response.User.ID)
	suite.Equal("access", response.AccessToken)
	suite.Equal("refresh", response.RefreshToken)
	suite.Equal("Bearer", response.TokenType)
	suite.Empty(response.PartialToken)
}

func (suite *AuthServiceTestSuite) TestVerify2FA_InvalidTokenType() {
	req := dto.Verify2FARequest{Token: "bad"}

	claims := &token.UserClaims{
		TokenType: token.AccessTokenType,
	}

	suite.tokenManager.EXPECT().
		ValidateToken(req.Token).
		Return(claims, nil).Once()

	response, err := suite.authService.Verify2FA(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalidToken)
}

func (suite *AuthServiceTestSuite) TestVerify2FA_InvalidCode() {
	user := &models.User{
		ID:               suite.userID,
		TwoFAEnabled:     true,
		TwoFASecret:      new("secret"),
		TwoFAConfirmedAt: new(time.Now().UTC()),
	}

	req := dto.Verify2FARequest{
		Token: "partial",
		Code:  "000000",
	}

	claims := &token.UserClaims{
		UserID:       user.ID.String(),
		UserRole:     string(user.Role),
		TwoFAEnabled: true,
		TokenType:    token.PartialTokenType,
	}

	suite.tokenManager.EXPECT().
		ValidateToken(req.Token).
		Return(claims, nil).Once()

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.encryptionManager.EXPECT().
		Decrypt(suite.ctx, "secret").
		Return("real", nil).Once()

	suite.totpManager.EXPECT().
		ValidateCode("real", req.Code).
		Return(false).Once()

	response, err := suite.authService.Verify2FA(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrInvalid2FACode)
}

func (suite *AuthServiceTestSuite) TestVerify2FA_2FANotEnabled() {
	user := &models.User{
		ID:           suite.userID,
		PasswordHash: "password_hash",
	}

	req := dto.Verify2FARequest{
		Token: "partial",
	}

	claims := &token.UserClaims{
		UserID:    user.ID.String(),
		TokenType: token.PartialTokenType,
	}

	suite.tokenManager.EXPECT().
		ValidateToken(req.Token).
		Return(claims, nil).Once()

	suite.userRepo.EXPECT().
		GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	response, err := suite.authService.Verify2FA(suite.ctx, req)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.Err2FANotEnabled)
}
