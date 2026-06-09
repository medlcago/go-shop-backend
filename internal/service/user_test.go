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
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	userRepo                    *repoMocks.MockUserRepository
	notificationTask            *tasksMocks.MockNotificationTask
	cache                       *cacheMocks.MockCache
	emailConfirmationCodeLength int
	emailConfirmationCodeTTL    time.Duration
	userService                 *userService

	ctx    context.Context
	userID uuid.UUID
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.userRepo = repoMocks.NewMockUserRepository(suite.T())
	suite.notificationTask = tasksMocks.NewMockNotificationTask(suite.T())
	suite.cache = cacheMocks.NewMockCache(suite.T())
	suite.emailConfirmationCodeLength = 6
	suite.emailConfirmationCodeTTL = 2 * time.Minute
	suite.userService = NewUserService(
		suite.userRepo,
		suite.notificationTask,
		suite.cache,
		suite.emailConfirmationCodeLength,
		suite.emailConfirmationCodeTTL,
	)

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
	suite.ErrorContains(err, "userService.GetUserByID")
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

// ==================== EmailConfirmation Tests ====================

func (suite *UserServiceTestSuite) TestEmailConfirmation_Success() {
	user := &models.User{
		ID:    suite.userID,
		Email: "test@example.com",
	}

	cacheKey := fmt.Sprintf("email_confirmation:%s", suite.userID)

	suite.cache.EXPECT().Exists(suite.ctx, cacheKey).
		Return(cache.ErrNotFound).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	suite.cache.EXPECT().Set(suite.ctx, cacheKey, mock.AnythingOfType("string"), suite.emailConfirmationCodeTTL).
		Return(nil).Once()

	suite.notificationTask.EXPECT().SendEmailConfirmationCode(suite.ctx, user.Email, mock.AnythingOfType("string")).
		Return(nil).Once()

	response, err := suite.userService.EmailConfirmation(suite.ctx, suite.userID)

	suite.NoError(err)
	suite.NotNil(response)
	suite.Equal(response.ExpiresIn, int(suite.emailConfirmationCodeTTL.Seconds()))
}

func (suite *UserServiceTestSuite) TestEmailConfirmation_EmailConfirmationCodeAlreadySent() {
	cacheKey := fmt.Sprintf("email_confirmation:%s", suite.userID)

	suite.cache.EXPECT().Exists(suite.ctx, cacheKey).
		Return(nil).Once()

	response, err := suite.userService.EmailConfirmation(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorContains(err, "userService.EmailConfirmation")
	suite.ErrorIs(err, apperror.ErrEmailConfirmationCodeAlreadySent)
}

func (suite *UserServiceTestSuite) TestEmailConfirmation_EmailAlreadyConfirmed() {
	user := &models.User{
		ID:               suite.userID,
		Email:            "test@example.com",
		EmailConfirmedAt: new(time.Now()),
	}

	cacheKey := fmt.Sprintf("email_confirmation:%s", suite.userID)

	suite.cache.EXPECT().Exists(suite.ctx, cacheKey).
		Return(cache.ErrNotFound).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(user, nil).Once()

	response, err := suite.userService.EmailConfirmation(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorContains(err, "userService.EmailConfirmation")
	suite.ErrorIs(err, apperror.ErrEmailAlreadyConfirmed)
}

func (suite *UserServiceTestSuite) TestEmailConfirmation_UserNotFound() {
	suite.cache.EXPECT().Exists(suite.ctx, mock.AnythingOfType("string")).
		Return(cache.ErrNotFound).Once()

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.userService.EmailConfirmation(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorContains(err, "userService.EmailConfirmation")
	suite.ErrorIs(err, apperror.ErrUserNotFound)
}

func (suite *UserServiceTestSuite) TestEmailConfirmation_InternalError() {
	cacheErr := errors.New("cache error")
	suite.cache.EXPECT().Exists(suite.ctx, mock.AnythingOfType("string")).
		Return(cacheErr).Once()

	response, err := suite.userService.EmailConfirmation(suite.ctx, suite.userID)

	suite.Nil(response)
	suite.ErrorContains(err, "userService.EmailConfirmation")
	suite.ErrorIs(err, cacheErr)
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
	suite.ErrorContains(err, "userService.ConfirmEmail")
	suite.ErrorIs(err, apperror.ErrEmailAlreadyConfirmed)
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
		Return("", cache.ErrNotFound).Once()

	response, err := suite.userService.ConfirmEmail(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "userService.ConfirmEmail")
	suite.ErrorIs(err, apperror.ErrInvalidCode)
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
	suite.ErrorContains(err, "userService.ConfirmEmail")
	suite.ErrorIs(err, apperror.ErrInvalidCode)
}

func (suite *UserServiceTestSuite) TestConfirmEmail_UserNotFound() {
	req := dto.ConfirmEmailRequest{
		Code: "123456",
	}

	suite.userRepo.EXPECT().GetByID(suite.ctx, suite.userID).
		Return(nil, repository.ErrRecordNotFound).Once()

	response, err := suite.userService.ConfirmEmail(suite.ctx, suite.userID, req)

	suite.Nil(response)
	suite.ErrorContains(err, "userService.ConfirmEmail")
	suite.ErrorIs(err, apperror.ErrUserNotFound)
}
