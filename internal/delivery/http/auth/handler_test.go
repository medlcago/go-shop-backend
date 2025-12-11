package auth

import (
	"encoding/json"
	"errors"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service/mocks"
	"go-shop-backend/internal/testutils"
	"go-shop-backend/pkg/apperrors"
	"go-shop-backend/pkg/response"
	"net/http"
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
				}).Return(&dto.LoginResponse{}, nil)
			},
			expectedCode: http.StatusOK,
			exceptedBody: response.NewResponse(&dto.LoginResponse{}, ""),
			email:        "superuser@test.com",
			password:     "test123",
		},
		{
			name: "login failed",
			setupMock: func(serviceMock *mocks.AuthServiceMock) {
				serviceMock.On("Login", mock.Anything, dto.UserLoginRequest{
					Email:    "superuser@test.com",
					Password: "123123",
				}).Return(&dto.LoginResponse{}, apperrors.ErrInvalidCredentials)
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
				}).Return(&dto.LoginResponse{}, errors.New("unexpected error"))
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
				}).Return(&dto.LoginResponse{}, apperrors.ErrInvalidCredentials)
			},
			expectedCode: http.StatusUnauthorized,
			exceptedBody: response.NewResponse(struct{}{}, apperrors.ErrInvalidCredentials.Message),
			email:        "test",
			password:     "123123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			ctx, cleanup := testutils.PrepareTestContext(app, "/", p1)
			defer cleanup()

			err := authHandler.Login(ctx)
			if err != nil {
				_ = app.Config().ErrorHandler(ctx, err)
			}

			exceptedBody, _ := json.Marshal(tt.exceptedBody)

			assert.Equal(t, tt.expectedCode, ctx.Response().StatusCode())
			assert.Equal(t, exceptedBody, ctx.Response().Body())

			mockService.AssertExpectations(t)
		})
	}
}
