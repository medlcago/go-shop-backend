package user

import (
	"errors"
	"go-shop-backend/internal/dto"
	serviceMocks "go-shop-backend/internal/service/mocks"
	"go-shop-backend/pkg/apperror"
	"go-shop-backend/pkg/response"
	"go-shop-backend/pkg/testutils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *serviceMocks.MockUserService)
		expectedCode int
		expectedBody any
		requestBody  any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().Login(mock.Anything, dto.UserLoginRequest{
					Email:    "superuser@test.com",
					Password: "test123",
				}).Return(&dto.UserTokenResponse{}, nil).Once()
			},
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(&dto.UserTokenResponse{}),
			requestBody: dto.UserLoginRequest{
				Email:    "superuser@test.com",
				Password: "test123",
			},
		},
		{
			name: "invalid credentials",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().Login(mock.Anything, dto.UserLoginRequest{
					Email:    "superuser@test.com",
					Password: "123123",
				}).Return(&dto.UserTokenResponse{}, apperror.ErrInvalidCredentials).Once()
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: response.NewError(apperror.ErrInvalidCredentials.Message),
			requestBody: dto.UserLoginRequest{
				Email:    "superuser@test.com",
				Password: "123123",
			},
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().Login(mock.Anything, dto.UserLoginRequest{
					Email:    "superuser@test.com",
					Password: "123123",
				}).Return(&dto.UserTokenResponse{}, errors.New("unexpected error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(http.StatusText(http.StatusInternalServerError)),
			requestBody: dto.UserLoginRequest{
				Email:    "superuser@test.com",
				Password: "123123",
			},
		},
		{
			name: "invalid credentials",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().Login(mock.Anything, dto.UserLoginRequest{
					Email:    "test@test.com",
					Password: "123123",
				}).Return(&dto.UserTokenResponse{}, apperror.ErrInvalidCredentials).Once()
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: response.NewError(apperror.ErrInvalidCredentials.Message),
			requestBody: dto.UserLoginRequest{
				Email:    "test@test.com",
				Password: "123123",
			},
		},
		{
			name:         "validation error",
			setupMock:    nil,
			expectedCode: http.StatusBadRequest,
			expectedBody: testutils.ValidationError(dto.UserLoginRequest{}),
			requestBody:  dto.UserLoginRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := serviceMocks.NewMockUserService(t)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			authHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Post("/", authHandler.Login)

			body := testutils.StringJSON(tt.requestBody)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

			resp, err := app.Test(req)
			assert.NoError(t, err)

			expectedJSON := testutils.StringJSON(tt.expectedBody)

			testutils.AssertJSONResponse(t, tt.expectedCode, expectedJSON, resp)

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *serviceMocks.MockUserService)
		expectedCode int
		expectedBody any
		requestBody  any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().Register(mock.Anything, dto.UserRegisterRequest{
					Email:    "superuser@test.com",
					Password: "test123",
				}).Return(&dto.UserTokenResponse{}, nil).Once()
			},
			expectedCode: http.StatusCreated,
			expectedBody: response.NewResponse(&dto.UserTokenResponse{}),
			requestBody: dto.UserRegisterRequest{
				Email:    "superuser@test.com",
				Password: "test123",
			},
		},
		{
			name: "email already in use",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().Register(mock.Anything, dto.UserRegisterRequest{
					Email:    "test@test.com",
					Password: "123123",
				}).Return(&dto.UserTokenResponse{}, apperror.ErrEmailTaken).Once()
			},
			expectedCode: http.StatusConflict,
			expectedBody: response.NewError(apperror.ErrEmailTaken.Message),
			requestBody: dto.UserRegisterRequest{
				Email:    "test@test.com",
				Password: "123123",
			},
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().Register(mock.Anything, dto.UserRegisterRequest{
					Email:    "test@test.com",
					Password: "123123",
				}).Return(&dto.UserTokenResponse{}, errors.New("unexpected error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(http.StatusText(http.StatusInternalServerError)),
			requestBody: dto.UserRegisterRequest{
				Email:    "test@test.com",
				Password: "123123",
			},
		},
		{
			name:         "invalid email",
			setupMock:    nil,
			expectedCode: http.StatusBadRequest,
			expectedBody: testutils.ValidationError(
				dto.UserRegisterRequest{
					Email:    "test",
					Password: "123123",
				},
			),
			requestBody: dto.UserRegisterRequest{
				Email:    "test",
				Password: "123123",
			},
		},
		{
			name:         "invalid password",
			setupMock:    nil,
			expectedCode: http.StatusBadRequest,
			expectedBody: testutils.ValidationError(
				dto.UserRegisterRequest{
					Email:    "test@test.com",
					Password: "123",
				},
			),
			requestBody: dto.UserRegisterRequest{
				Email:    "test@test.com",
				Password: "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := serviceMocks.NewMockUserService(t)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			authHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Post("/", authHandler.Register)

			body := testutils.StringJSON(tt.requestBody)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

			resp, err := app.Test(req)
			assert.NoError(t, err)

			expectedJSON := testutils.StringJSON(tt.expectedBody)

			testutils.AssertJSONResponse(t, tt.expectedCode, expectedJSON, resp)

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetMe(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *serviceMocks.MockUserService)
		userID       *uuid.UUID
		expectedCode int
		expectedBody any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().GetUserByID(mock.Anything, uuid.MustParse("c2f72e02-98b6-4cef-9a80-616f820fed31")).
					Return(&dto.UserResponse{
						ID:    uuid.MustParse("c2f72e02-98b6-4cef-9a80-616f820fed31"),
						Email: "test@test.com",
					}, nil).Once()
			},
			userID:       new(uuid.MustParse("c2f72e02-98b6-4cef-9a80-616f820fed31")),
			expectedCode: http.StatusOK,
			expectedBody: response.NewResponse(
				&dto.UserResponse{
					ID:    uuid.MustParse("c2f72e02-98b6-4cef-9a80-616f820fed31"),
					Email: "test@test.com",
				}),
		}, {
			name:         "no user id in context",
			setupMock:    nil,
			userID:       nil,
			expectedCode: http.StatusUnauthorized,
			expectedBody: response.NewError(apperror.ErrInvalidCredentials.Message),
		},
		{
			name: "user not found",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().GetUserByID(mock.Anything, uuid.MustParse("c2f72e02-98b6-4cef-9a80-616f820fed31")).
					Return(&dto.UserResponse{}, apperror.ErrUserNotFound).Once()
			},
			userID:       new(uuid.MustParse("c2f72e02-98b6-4cef-9a80-616f820fed31")),
			expectedCode: http.StatusNotFound,
			expectedBody: response.NewError(apperror.ErrUserNotFound.Message),
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *serviceMocks.MockUserService) {
				serviceMock.EXPECT().GetUserByID(mock.Anything, uuid.MustParse("c2f72e02-98b6-4cef-9a80-616f820fed31")).
					Return(&dto.UserResponse{}, errors.New("unexpected error")).Once()
			},
			userID:       new(uuid.MustParse("c2f72e02-98b6-4cef-9a80-616f820fed31")),
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewError(http.StatusText(http.StatusInternalServerError)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := serviceMocks.NewMockUserService(t)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			userHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Get("/", func(c fiber.Ctx) error {
				if tt.userID != nil {
					c.Locals("userID", *tt.userID)
				}

				return userHandler.GetMe(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			resp, err := app.Test(req)
			assert.NoError(t, err)

			expectedJSON := testutils.StringJSON(tt.expectedBody)

			testutils.AssertJSONResponse(t, tt.expectedCode, expectedJSON, resp)

			mockService.AssertExpectations(t)
		})
	}
}
