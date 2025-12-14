package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service/mocks"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/response"
	"go-shop-backend/pkg/testutils"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *mocks.AuthServiceMock)
		expectedCode int
		exceptedBody any
		email        string
		password     string
	}{
		{
			name: "login success",
			setupMock: func(serviceMock *mocks.AuthServiceMock) {
				serviceMock.On("Login", mock.Anything, dto.UserLoginRequest{
					Email:    "superuser@test.com",
					Password: "test123",
				}).Return(&dto.LoginResponse{}, nil).Once()
			},
			expectedCode: http.StatusOK,
			exceptedBody: response.NewResponse(&dto.LoginResponse{}),
			email:        "superuser@test.com",
			password:     "test123",
		},
		{
			name: "login failed",
			setupMock: func(serviceMock *mocks.AuthServiceMock) {
				serviceMock.On("Login", mock.Anything, dto.UserLoginRequest{
					Email:    "superuser@test.com",
					Password: "123123",
				}).Return(&dto.LoginResponse{}, apperrors.ErrInvalidCredentials).Once()
			},
			expectedCode: http.StatusUnauthorized,
			exceptedBody: response.NewResponse(struct{}{}, apperrors.ErrInvalidCredentials.Message),
			email:        "superuser@test.com",
			password:     "123123",
		},
		{
			name: "login failed (internal server error)",
			setupMock: func(serviceMock *mocks.AuthServiceMock) {
				serviceMock.On("Login", mock.Anything, dto.UserLoginRequest{
					Email:    "superuser@test.com",
					Password: "123123",
				}).Return(&dto.LoginResponse{}, errors.New("unexpected error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			exceptedBody: response.NewResponse(struct{}{}, http.StatusText(http.StatusInternalServerError)),
			email:        "superuser@test.com",
			password:     "123123",
		},
		{
			name: "login failed (invalid email)",
			setupMock: func(serviceMock *mocks.AuthServiceMock) {
				serviceMock.On("Login", mock.Anything, dto.UserLoginRequest{
					Email:    "test",
					Password: "123123",
				}).Return(&dto.LoginResponse{}, apperrors.ErrInvalidCredentials).Once()
			},
			expectedCode: http.StatusUnauthorized,
			exceptedBody: response.NewResponse(struct{}{}, apperrors.ErrInvalidCredentials.Message),
			email:        "test",
			password:     "123123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p1 := &dto.UserLoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}

			mockService := new(mocks.AuthServiceMock)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			authHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Post("/", authHandler.Login)

			body, _ := json.Marshal(p1)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

			resp, err := app.Test(req)
			assert.NoError(t, err)

			exceptedBody, _ := json.Marshal(tt.exceptedBody)
			actualBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Equal(t, exceptedBody, actualBody)

			mockService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *mocks.AuthServiceMock)
		expectedCode int
		exceptedBody any
		email        string
		password     string
	}{
		{
			name: "success",
			setupMock: func(serviceMock *mocks.AuthServiceMock) {
				serviceMock.On("Register", mock.Anything, dto.UserRegisterRequest{
					Email:    "superuser@test.com",
					Password: "test123",
				}).Return(&dto.RegisterResponse{}, nil).Once()
			},
			expectedCode: http.StatusCreated,
			exceptedBody: response.NewResponse(&dto.RegisterResponse{}),
			email:        "superuser@test.com",
			password:     "test123",
		},
		{
			name: "email already in use",
			setupMock: func(serviceMock *mocks.AuthServiceMock) {
				serviceMock.On("Register", mock.Anything, dto.UserRegisterRequest{
					Email:    "test@test.com",
					Password: "123123",
				}).Return(&dto.RegisterResponse{}, apperrors.ErrEmailTaken).Once()
			},
			expectedCode: http.StatusConflict,
			exceptedBody: response.NewResponse(struct{}{}, apperrors.ErrEmailTaken.Message),
			email:        "test@test.com",
			password:     "123123",
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *mocks.AuthServiceMock) {
				serviceMock.On("Register", mock.Anything, dto.UserRegisterRequest{
					Email:    "test@test.com",
					Password: "123123",
				}).Return(&dto.RegisterResponse{}, errors.New("unexpected error")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			exceptedBody: response.NewResponse(struct{}{}, http.StatusText(http.StatusInternalServerError)),
			email:        "test@test.com",
			password:     "123123",
		},
		{
			name:         "invalid email",
			setupMock:    nil,
			expectedCode: http.StatusBadRequest,
			exceptedBody: response.NewResponse(struct{}{}, "Key: 'UserRegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"),
			email:        "test",
			password:     "123123",
		},
		{
			name:         "invalid password",
			setupMock:    nil,
			expectedCode: http.StatusBadRequest,
			exceptedBody: response.NewResponse(struct{}{}, "Key: 'UserRegisterRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag"),
			email:        "test@test.com",
			password:     "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p1 := &dto.UserRegisterRequest{
				Email:    tt.email,
				Password: tt.password,
			}

			mockService := new(mocks.AuthServiceMock)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			authHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Post("/", authHandler.Register)

			body, _ := json.Marshal(p1)

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

			resp, err := app.Test(req)
			assert.NoError(t, err)

			exceptedBody, _ := json.Marshal(tt.exceptedBody)
			actualBody, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			assert.Equal(t, exceptedBody, actualBody)

			mockService.AssertExpectations(t)
		})
	}
}
