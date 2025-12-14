package user

import (
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

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_GetMe(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(serviceMock *mocks.UserServiceMock)
		userID       string
		expectedCode int
		exceptedBody any
	}{
		{
			name: "success",
			setupMock: func(serviceMock *mocks.UserServiceMock) {
				serviceMock.On("GetUserByID", mock.Anything, "c2f72e02-98b6-4cef-9a80-616f820fed31").
					Return(&dto.UserResponse{
						ID:    "c2f72e02-98b6-4cef-9a80-616f820fed31",
						Email: "test@test.com",
					}, nil).Once()
			},
			userID:       "c2f72e02-98b6-4cef-9a80-616f820fed31",
			expectedCode: http.StatusOK,
			exceptedBody: response.NewResponse(
				&dto.UserResponse{
					ID:    "c2f72e02-98b6-4cef-9a80-616f820fed31",
					Email: "test@test.com",
				}),
		}, {
			name:         "no user id in context",
			setupMock:    nil,
			userID:       "",
			expectedCode: http.StatusUnauthorized,
			exceptedBody: response.NewResponse(struct{}{}, apperrors.ErrInvalidCredentials.Message),
		},
		{
			name: "user not found",
			setupMock: func(serviceMock *mocks.UserServiceMock) {
				serviceMock.On("GetUserByID", mock.Anything, "c2f72e02-98b6-4cef-9a80-616f820fed31").
					Return(&dto.UserResponse{}, apperrors.ErrUserNotFound).Once()
			},
			userID:       "c2f72e02-98b6-4cef-9a80-616f820fed31",
			expectedCode: http.StatusNotFound,
			exceptedBody: response.NewResponse(struct{}{}, apperrors.ErrUserNotFound.Message),
		},
		{
			name: "internal server error",
			setupMock: func(serviceMock *mocks.UserServiceMock) {
				serviceMock.On("GetUserByID", mock.Anything, "c2f72e02-98b6-4cef-9a80-616f820fed31").
					Return(&dto.UserResponse{}, errors.New("unexpected error")).Once()
			},
			userID:       "c2f72e02-98b6-4cef-9a80-616f820fed31",
			expectedCode: http.StatusInternalServerError,
			exceptedBody: response.NewResponse(struct{}{}, http.StatusText(http.StatusInternalServerError)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockService := new(mocks.UserServiceMock)
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			userHandler := NewHandler(mockService)

			app := testutils.CreateTestApp()

			app.Get("/", func(c fiber.Ctx) error {
				c.Locals("userID", tt.userID)
				return userHandler.GetMe(c)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)

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
