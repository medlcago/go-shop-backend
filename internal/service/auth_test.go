package service

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/password"
	"go-shop-backend/pkg/token"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AuthServiceTestSuite struct {
	suite.Suite
	mockRepo     *mocks.UserRepositoryMock
	tokenManager *token.ManagerMock
	service      AuthService
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.mockRepo = new(mocks.UserRepositoryMock)
	suite.tokenManager = new(token.ManagerMock)
	suite.service = NewAuthService(suite.mockRepo, suite.tokenManager)
}

func (suite *AuthServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
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

	passwordHash, _ := password.HashPassword(req.Password)
	expectedUser := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         models.UserRoleCustomer,
	}

	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(expectedUser, nil).Once()

	payload := map[string]interface{}{
		"user_id": expectedUser.ID,
		"role":    expectedUser.Role,
	}

	suite.tokenManager.On("GenerateAccessToken", payload).Return("test_access_token", nil).Once()
	suite.tokenManager.On("GenerateRefreshToken", payload).Return("test_refresh_token", nil).Once()

	result, err := suite.service.Login(ctx, req)

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

	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(&models.User{}, repository.ErrRecordNotFound).Once()

	tokenResp, err := suite.service.Login(ctx, req)

	suite.Error(err)
	suite.Nil(tokenResp)
	suite.ErrorIs(err, apperrors.ErrInvalidCredentials)
}

func (suite *AuthServiceTestSuite) TestLogin_ProfileDeleted() {
	ctx := context.Background()
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(&models.User{DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	tokenResp, err := suite.service.Login(ctx, req)

	suite.Error(err)
	suite.Nil(tokenResp)
	suite.ErrorIs(err, apperrors.ErrUserProfileDeleted)
}

func (suite *AuthServiceTestSuite) TestLogin_InvalidPassword() {
	ctx := context.Background()
	req := dto.UserLoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	passwordHash, _ := password.HashPassword("test123")
	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(user, nil).Once()

	tokenResp, err := suite.service.Login(ctx, req)

	suite.Error(err)
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
	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(&models.User{}, repoErr).Once()

	result, err := suite.service.Login(ctx, req)

	suite.Error(err)
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

	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(&models.User{}, repository.ErrRecordNotFound).Once()

	suite.mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		user.ID = userID
		user.Role = userRole
		return user.Email == req.Email && user.PasswordHash != ""
	})).Return(nil).Once()

	payload := map[string]interface{}{
		"user_id": userID,
		"role":    userRole,
	}

	suite.tokenManager.On("GenerateAccessToken", payload).Return("test_access_token", nil).Once()
	suite.tokenManager.On("GenerateRefreshToken", payload).Return("test_refresh_token", nil).Once()

	result, err := suite.service.Register(ctx, req)

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

	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(user, nil).Once()

	result, err := suite.service.Register(ctx, req)

	suite.Error(err)
	suite.Nil(result)
	suite.ErrorIs(err, apperrors.ErrEmailTaken)
}

func (suite *AuthServiceTestSuite) TestRegister_ProfileDeleted() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "testuser@example.com",
		Password: "password123",
	}

	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(&models.User{DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	result, err := suite.service.Register(ctx, req)

	suite.Error(err)
	suite.Nil(result)
	suite.ErrorIs(err, apperrors.ErrEmailTaken)
}

func (suite *AuthServiceTestSuite) TestRegister_RepositoryError() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	suite.mockRepo.On("GetByEmailUnscoped", mock.Anything, req.Email).
		Return(&models.User{}, repository.ErrRecordNotFound).Once()

	repoErr := errors.New("database error")
	suite.mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).
		Return(repoErr).Once()

	result, err := suite.service.Register(ctx, req)

	suite.Error(err)
	suite.Nil(result)
	suite.ErrorContains(err, repoErr.Error())
}
