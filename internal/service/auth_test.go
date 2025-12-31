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
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	suite.Suite
	mockRepo  *mocks.UserRepositoryMock
	service   AuthService
	secretKey string
}

func (suite *AuthServiceTestSuite) SetupTest() {
	suite.mockRepo = new(mocks.UserRepositoryMock)
	suite.secretKey = "test-secret-key"
	suite.service = NewAuthService(suite.mockRepo, suite.secretKey)
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
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
		Return(expectedUser, nil).Once()

	token, err := suite.service.Login(ctx, req)

	suite.NoError(err)
	suite.NotNil(token)
	suite.NotNil(token.TokenResponse)
	suite.NotNil(token.User)
	suite.Equal(req.Email, token.User.Email)
	suite.NotEmpty(token.TokenResponse.AccessToken)
	suite.NotEmpty(token.TokenResponse.RefreshToken)
	suite.Equal("Bearer", token.TokenResponse.TokenType)
}

func (suite *AuthServiceTestSuite) TestLogin_UserNotFound() {
	ctx := context.Background()
	req := dto.UserLoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	}

	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
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

	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
		Return(&models.User{DeletedAt: sql.NullTime{Time: time.Now(), Valid: true}}, nil).Once()

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

	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
		Return(user, nil).Once()

	tokenResp, err := suite.service.Login(ctx, req)

	suite.Error(err)
	suite.Nil(tokenResp)
	suite.ErrorIs(err, apperrors.ErrInvalidCredentials)
}

func (suite *AuthServiceTestSuite) TestLogin_RepositoryError() {
	req := dto.UserLoginRequest{
		Email:    "test@test.test",
		Password: "test123",
	}

	repoErr := errors.New("database error")
	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
		Return(&models.User{}, repoErr).Once()

	token, err := suite.service.Login(context.Background(), req)

	suite.Error(err)
	suite.Nil(token)
	suite.ErrorContains(err, repoErr.Error())
}

// ==================== Register Tests ====================

func (suite *AuthServiceTestSuite) TestRegister_Success() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
		Return(&models.User{}, repository.ErrRecordNotFound).Once()

	suite.mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(user *models.User) bool {
		return user.Email == req.Email && user.PasswordHash != ""
	})).Return(nil).Once()

	token, err := suite.service.Register(ctx, req)

	suite.NoError(err)
	suite.NotNil(token)
	suite.NotNil(token.TokenResponse)
	suite.NotNil(token.User)
	suite.Equal(req.Email, token.User.Email)
	suite.NotEmpty(token.TokenResponse.AccessToken)
	suite.NotEmpty(token.TokenResponse.RefreshToken)
	suite.Equal("Bearer", token.TokenResponse.TokenType)
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

	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
		Return(user, nil).Once()

	token, err := suite.service.Register(ctx, req)

	suite.Error(err)
	suite.Nil(token)
	suite.ErrorIs(err, apperrors.ErrEmailTaken)
}

func (suite *AuthServiceTestSuite) TestRegister_ProfileDeleted() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "testuser@example.com",
		Password: "password123",
	}

	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
		Return(&models.User{DeletedAt: sql.NullTime{Time: time.Now(), Valid: true}}, nil).Once()

	token, err := suite.service.Register(ctx, req)

	suite.Error(err)
	suite.Nil(token)
	suite.ErrorIs(err, apperrors.ErrEmailTaken)
}

func (suite *AuthServiceTestSuite) TestRegister_RepositoryError() {
	ctx := context.Background()
	req := dto.UserRegisterRequest{
		Email:    "newuser@example.com",
		Password: "password123",
	}

	suite.mockRepo.On("GetByEmail", mock.Anything, req.Email).
		Return(&models.User{}, repository.ErrRecordNotFound).Once()

	repoErr := errors.New("database error")
	suite.mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*models.User")).
		Return(repoErr).Once()

	token, err := suite.service.Register(ctx, req)

	suite.Error(err)
	suite.Nil(token)
	suite.ErrorContains(err, repoErr.Error())
}
