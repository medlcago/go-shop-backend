package service

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperrors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	userRepo    *repoMocks.MockUserRepository
	userService *userService

	ctx    context.Context
	userID uuid.UUID
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.userRepo = repoMocks.NewMockUserRepository(suite.T())
	suite.userService = NewUserService(suite.userRepo)

	suite.ctx = context.Background()
	suite.userID = uuid.New()
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// ==================== GetUserByID Tests ====================

func (suite *UserServiceTestSuite) TestGetUserByID_Success() {
	mockUser := &models.User{
		ID:        suite.userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	suite.userRepo.EXPECT().GetByIDUnscoped(suite.ctx, suite.userID).
		Return(mockUser, nil).Once()

	user, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.NoError(err)
	suite.NotNil(user)
	suite.Equal(suite.userID, user.ID)
	suite.Equal(mockUser.Email, user.Email)
	suite.NotZero(user.CreatedAt)
}

func (suite *UserServiceTestSuite) TestGetUserByID_NotFound() {
	suite.userRepo.EXPECT().GetByIDUnscoped(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	user, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.Nil(user)
	suite.ErrorIs(err, apperrors.ErrUserNotFound)
}

func (suite *UserServiceTestSuite) TestGetUserByID_ProfileDeleted() {
	suite.userRepo.EXPECT().GetByIDUnscoped(suite.ctx, suite.userID).
		Return(&models.User{ID: suite.userID, DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	user, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.Nil(user)
	suite.ErrorIs(err, apperrors.ErrUserProfileDeleted)
}

func (suite *UserServiceTestSuite) TestGetUserByID_RepositoryError() {
	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().GetByIDUnscoped(suite.ctx, suite.userID).
		Return(nil, repoErr).Once()

	user, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.Nil(user)
	suite.ErrorContains(err, repoErr.Error())
}
