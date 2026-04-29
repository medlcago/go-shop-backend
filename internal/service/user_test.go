package service

import (
	"context"
	"database/sql"
	"errors"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"
	repoMocks "go-shop-backend/internal/repository/mocks"
	"go-shop-backend/pkg/apperror"
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
}

func (suite *UserServiceTestSuite) TestGetUserByID_ProfileDeleted() {
	suite.userRepo.EXPECT().GetByIDIncludingDeleted(suite.ctx, suite.userID).
		Return(&models.User{ID: suite.userID, DeletedAt: gorm.DeletedAt(sql.NullTime{Time: time.Now(), Valid: true})}, nil).Once()

	response, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorIs(err, apperror.ErrUserProfileDeleted)
}

func (suite *UserServiceTestSuite) TestGetUserByID_RepositoryError() {
	repoErr := errors.New("database error")
	suite.userRepo.EXPECT().GetByIDIncludingDeleted(suite.ctx, suite.userID).
		Return(nil, repoErr).Once()

	response, err := suite.userService.GetUserByID(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorContains(err, repoErr.Error())
}
