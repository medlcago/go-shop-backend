package service

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperrors"
	hasherMocks "go-shop-backend/pkg/hasher/mocks"
	"go-shop-backend/pkg/token"
	tokenMocks "go-shop-backend/pkg/token/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AuthServiceTestSuite struct {
	suite.Suite
	userRepo       *repoMocks.MockUserRepository
	tokenManager   *tokenMocks.MockManager
	passwordHasher *hasherMocks.MockHasher
	authService    AuthService
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.userRepo = repoMocks.NewMockUserRepository(suite.T())
	suite.tokenManager = tokenMocks.NewMockManager(suite.T())
	suite.passwordHasher = hasherMocks.NewMockHasher(suite.T())
	suite.authService = NewAuthService(suite.userRepo, suite.tokenManager, suite.passwordHasher)
}

func TestAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

// ==================== Login Tests ====================

func (suite *AuthServiceTestSuite) TestLogin_Success() {
	ctx := context.Background()
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "test123",
	}

	expectedUser := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: "test123",
		Role:         models.UserRoleCustomer,
	}

	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(expectedUser, nil).Once()

	suite.passwordHasher.EXPECT().Verify(req.Password, expectedUser.PasswordHash).
		Return(true, nil).Once()

	payload := token.Payload{
		UserID:   expectedUser.ID.String(),
		UserRole: string(expectedUser.Role),
	}

	suite.tokenManager.EXPECT().GenerateAccessToken(payload).
		Return("test_access_token", nil).Once()

	suite.tokenManager.EXPECT().GenerateRefreshToken(payload).
		Return("test_refresh_token", nil).Once()

	result, err := suite.authService.Login(ctx, req)

	suite.NoError(err)
	suite.NotNil(result)
	suite.NotNil(result.User)
	suite.Equal(result.User.ID, expectedUser.ID)
	suite.Equal(result.User.Email, expectedUser.Email)
	suite.NotNil(result.TokenResponse)
	suite.Equal(result.AccessToken, "test_access_token")
	suite.Equal(result.RefreshToken, "test_refresh_token")
	suite.Equal("Bearer", result.TokenResponse.TokenType)
}

func (suite *AuthServiceTestSuite) TestLogin_UserNotFound() {
	ctx := context.Background()
	req := dto.UserLoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(nil, repository.ErrRecordNotFound).Once()

	tokenResp, err := suite.authService.Login(ctx, req)

	suite.Nil(tokenResp)
	suite.ErrorIs(err, apperrors.ErrInvalidCredentials)
}

func (suite *AuthServiceTestSuite) TestLogin_ProfileDeleted() {
	ctx := context.Background()
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(&models.User{DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	suite.passwordHasher.EXPECT().Verify(req.Password, mock.Anything).
		Return(true, nil).Once()

	tokenResp, err := suite.authService.Login(ctx, req)

	suite.Nil(tokenResp)
	suite.ErrorIs(err, apperrors.ErrUserProfileDeleted)
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidPassword() {
	ctx := context.Background()
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: "test123",
	}

	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(user, nil).Once()

	suite.passwordHasher.EXPECT().Verify(req.Password, user.PasswordHash).
		Return(false, nil).Once()

	tokenResp, err := suite.authService.Login(ctx, req)

	suite.Nil(tokenResp)
	suite.ErrorIs(err, apperrors.ErrInvalidCredentials)
}

func (suite *AuthServiceTestSuite) TestLogin_RepositoryError() {
	ctx := context.Background()
	req := dto.UserLoginRequest{
		Email:    "test@test.test",
		Password: "test123",
	}

	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(nil, repoErr).Once()

	result, err := suite.authService.Login(ctx, req)

	suite.Nil(result)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== Register Tests ====================

func (suite *AuthServiceTestSuite) TestRegister_Success() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	userID := uuid.New()
	userRole := models.UserRoleSeller
	passwordHash := "test123"

	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(nil, repository.ErrRecordNotFound).Once()

	suite.passwordHasher.EXPECT().Hash(req.Password).
		Return(passwordHash, nil).Once()

	suite.userRepo.EXPECT().Create(ctx, mock.MatchedBy(func(user *models.User) bool {
		user.ID = userID
		user.Role = userRole
		return user.Email == req.Email && user.PasswordHash == passwordHash
	})).Return(nil).Once()

	payload := token.Payload{
		UserID:   userID.String(),
		UserRole: string(userRole),
	}

	suite.tokenManager.EXPECT().GenerateAccessToken(payload).
		Return("test_access_token", nil).Once()

	suite.tokenManager.EXPECT().GenerateRefreshToken(payload).
		Return("test_refresh_token", nil).Once()

	result, err := suite.authService.Register(ctx, req)

	suite.NoError(err)
	suite.NotNil(result)
	suite.NotNil(result.User)
	suite.Equal(result.User.ID, userID)
	suite.Equal(result.User.Email, req.Email)
	suite.NotNil(result.TokenResponse)
	suite.Equal(result.AccessToken, "test_access_token")
	suite.Equal(result.RefreshToken, "test_refresh_token")
	suite.Equal("Bearer", result.TokenResponse.TokenType)
}

func (suite *AuthServiceTestSuite) TestRegister_EmailAlreadyExists() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	user := &models.User{
		Email: req.Email,
	}

	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(user, nil).Once()

	result, err := suite.authService.Register(ctx, req)

	suite.Nil(result)
	suite.ErrorIs(err, apperrors.ErrEmailTaken)
}

func (suite *AuthServiceTestSuite) TestRegister_ProfileDeleted() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "testuser@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(&models.User{DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	result, err := suite.authService.Register(ctx, req)

	suite.Nil(result)
	suite.ErrorIs(err, apperrors.ErrEmailTaken)
}

func (suite *AuthServiceTestSuite) TestRegister_RepositoryError() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	suite.userRepo.EXPECT().GetByEmailUnscoped(ctx, req.Email).
		Return(nil, repository.ErrRecordNotFound).Once()

	suite.passwordHasher.EXPECT().Hash(req.Password).
		Return("test123", nil).Once()

	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().Create(ctx, mock.AnythingOfType("*models.User")).
		Return(repoErr).Once()

	result, err := suite.authService.Register(ctx, req)

	suite.Nil(result)
	suite.ErrorContains(err, repoErr.Error())
}
