package service

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	"go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperrors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	userRepo    *mocks.UserRepositoryMock
	userService UserService
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.userRepo = new(mocks.UserRepositoryMock)
	suite.userService = NewUserService(suite.userRepo)
}

func (suite *UserServiceTestSuite) TearDownTest() {
	suite.userRepo.AssertExpectations(suite.T())
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// ==================== GetUserByID Tests ====================

func (suite *UserServiceTestSuite) TestGetUserByID_Success() {
	ctx := context.Background()
	userID := uuid.New()

	mockUser := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	suite.userRepo.On("GetByIDUnscoped", ctx, userID).
		Return(mockUser, nil).Once()

	user, err := suite.userService.GetUserByID(ctx, userID)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(userID, user.ID)
	suite.Equal("test@example.com", user.Email)
	suite.NotZero(user.CreatedAt)
}

func (suite *UserServiceTestSuite) TestGetUserByID_NotFound() {
	ctx := context.Background()
	userID := uuid.New()

	suite.userRepo.On("GetByIDUnscoped", ctx, userID).
		Return(&models.User{}, repository.ErrRecordNotFound).Once()

	user, err := suite.userService.GetUserByID(ctx, userID)

	suite.Error(err)
	suite.Nil(user)
	suite.ErrorIs(err, apperrors.ErrUserNotFound)
}

func (suite *UserServiceTestSuite) TestGetUserByID_ProfileDeleted() {
	ctx := context.Background()
	userID := uuid.New()

	suite.userRepo.On("GetByIDUnscoped", ctx, userID).
		Return(&models.User{ID: userID, DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	user, err := suite.userService.GetUserByID(ctx, userID)

	suite.Error(err)
	suite.Nil(user)
	suite.ErrorIs(err, apperrors.ErrUserProfileDeleted)
}

func (suite *UserServiceTestSuite) TestGetUserByID_RepositoryError() {
	ctx := context.Background()
	userID := uuid.New()

	repoErr := errors.New("database error")
	suite.userRepo.On("GetByIDUnscoped", ctx, userID).
		Return(&models.User{}, repoErr).Once()

	user, err := suite.userService.GetUserByID(ctx, userID)

	suite.Error(err)
	suite.Nil(user)
	suite.ErrorContains(err, repoErr.Error())
}
