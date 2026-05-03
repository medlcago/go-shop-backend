package service

import (
	"context"
	"errors"
	"go-shop-backend/pkg/notification"
	notificationMocks "go-shop-backend/pkg/notification/mocks"
	templateMocks "go-shop-backend/pkg/template/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type NotificationServiceTestSuite struct {
	suite.Suite

	registry            *notificationMocks.MockSenderRegistry
	sender              *notificationMocks.MockSender
	templateManager     *templateMocks.MockManager
	notificationService *notificationService

	ctx     context.Context
	to      string
	code    string
	channel notification.Channel
}

func (suite *NotificationServiceTestSuite) SetupTest() {
	suite.registry = notificationMocks.NewMockSenderRegistry(suite.T())
	suite.sender = notificationMocks.NewMockSender(suite.T())
	suite.templateManager = templateMocks.NewMockManager(suite.T())
	suite.notificationService = NewNotificationService(
		suite.registry,
		suite.templateManager,
	)

	suite.ctx = context.Background()
	suite.to = "test@example.com"
	suite.code = uuid.NewString()
	suite.channel = notification.ChannelEmail
}

func TestNotificationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceTestSuite))
}

// ==================== SendEmailConfirmationCode Tests ====================

func (suite *NotificationServiceTestSuite) TestSendEmailConfirmationCode_Success() {
	data := map[string]string{
		"Code": suite.code,
	}

	suite.registry.EXPECT().For(notification.ChannelEmail).
		Return(suite.sender, true).Once()

	suite.templateManager.EXPECT().Render("email_confirmation_code.gohtml", data).
		Return("test html body", nil).Once()

	suite.sender.EXPECT().Send(suite.ctx, mock.AnythingOfType("notification.Notification")).
		Return(nil).Once()

	err := suite.notificationService.SendEmailConfirmationCode(suite.ctx, suite.to, suite.code, suite.channel)
	suite.NoError(err)
}

func (suite *NotificationServiceTestSuite) TestSendEmailConfirmationCode_ToIsEmpty() {
	err := suite.notificationService.SendEmailConfirmationCode(suite.ctx, "", suite.code, suite.channel)

	suite.Error(err)
	suite.ErrorContains(err, "notificationService.SendEmailConfirmationCode")
	suite.ErrorContains(err, "to is empty")
}

func (suite *NotificationServiceTestSuite) TestSendEmailConfirmationCode_NoSenderForChannel() {
	suite.registry.EXPECT().For(notification.Channel("unknown")).
		Return(nil, false).Once()

	err := suite.notificationService.SendEmailConfirmationCode(suite.ctx, suite.to, suite.code, "unknown")

	suite.Error(err)
	suite.ErrorContains(err, "notificationService.SendEmailConfirmationCode")
	suite.ErrorContains(err, "no sender for channel")
}

func (suite *NotificationServiceTestSuite) TestSendEmailConfirmationCode_RenderTemplateError() {
	data := map[string]string{
		"Code": suite.code,
	}

	renderErr := errors.New("render error")

	suite.registry.EXPECT().For(notification.ChannelEmail).
		Return(suite.sender, true).Once()

	suite.templateManager.EXPECT().Render("email_confirmation_code.gohtml", data).
		Return("", renderErr).Once()

	err := suite.notificationService.SendEmailConfirmationCode(suite.ctx, suite.to, suite.code, suite.channel)

	suite.ErrorIs(err, renderErr)
	suite.ErrorContains(err, "notificationService.SendEmailConfirmationCode")
}

func (suite *NotificationServiceTestSuite) TestSendEmailConfirmationCode_SendNotificationError() {
	data := map[string]string{
		"Code": suite.code,
	}

	internalErr := errors.New("internal error")

	suite.registry.EXPECT().For(notification.ChannelEmail).
		Return(suite.sender, true).Once()

	suite.templateManager.EXPECT().Render("email_confirmation_code.gohtml", data).
		Return("test html body", nil).Once()

	suite.sender.EXPECT().Send(suite.ctx, mock.AnythingOfType("notification.Notification")).
		Return(internalErr).Once()

	err := suite.notificationService.SendEmailConfirmationCode(suite.ctx, suite.to, suite.code, suite.channel)

	suite.ErrorIs(err, internalErr)
	suite.ErrorContains(err, "notificationService.SendEmailConfirmationCode")
}
