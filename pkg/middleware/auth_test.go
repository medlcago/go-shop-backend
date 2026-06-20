package middleware

import (
	"go-shop-backend/pkg/token"
	tokenMocks "go-shop-backend/pkg/token/mocks"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIdentityUser(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	tests := []struct {
		name              string
		authHeader        string
		sessionHeader     string
		mockSetup         func(*tokenMocks.MockManager)
		expectedUserID    *uuid.UUID
		expectedRole      string
		expectedToken     string
		expectedType      string
		expectedSessionID *uuid.UUID
	}{
		{
			name: "no authorization header",
			mockSetup: func(m *tokenMocks.MockManager) {
			},
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			mockSetup: func(m *tokenMocks.MockManager) {
				m.EXPECT().
					ValidateToken("invalid-token").
					Return(nil, assert.AnError).Once()
			},
		},
		{
			name:       "invalid user id in claims",
			authHeader: "Bearer valid-token",
			mockSetup: func(m *tokenMocks.MockManager) {
				m.EXPECT().
					ValidateToken("valid-token").
					Return(&token.UserClaims{
						Payload: token.Payload{
							UserID:   "not-a-uuid",
							UserRole: "customer",
						},
						TokenType: token.AccessTokenType,
					}, nil)
			},
		},
		{
			name:          "valid token",
			authHeader:    "Bearer valid-token",
			sessionHeader: sessionID.String(),
			mockSetup: func(m *tokenMocks.MockManager) {
				m.EXPECT().
					ValidateToken("valid-token").
					Return(&token.UserClaims{
						Payload: token.Payload{
							UserID:   userID.String(),
							UserRole: "customer",
						},
						TokenType: token.AccessTokenType,
					}, nil).Once()
			},
			expectedUserID:    &userID,
			expectedRole:      "customer",
			expectedToken:     "valid-token",
			expectedType:      token.AccessTokenType,
			expectedSessionID: &sessionID,
		},
		{
			name:          "session id only",
			sessionHeader: sessionID.String(),
			mockSetup: func(m *tokenMocks.MockManager) {
			},
			expectedSessionID: &sessionID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenManager := tokenMocks.NewMockManager(t)

			if tt.mockSetup != nil {
				tt.mockSetup(tokenManager)
			}

			app := fiber.New()

			var userCtx UserContext

			app.Use(IdentityUser(tokenManager))

			app.Get("/", func(ctx fiber.Ctx) error {
				userCtx = GetUserContext(ctx)
				return ctx.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			if tt.sessionHeader != "" {
				req.Header.Set("X-Session-ID", tt.sessionHeader)
			}

			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusOK, resp.StatusCode)

			assert.Equal(t, tt.expectedUserID, userCtx.UserID)
			assert.Equal(t, tt.expectedRole, userCtx.Role)
			assert.Equal(t, tt.expectedToken, userCtx.Token)
			assert.Equal(t, tt.expectedType, userCtx.TokenType)
			assert.Equal(t, tt.expectedSessionID, userCtx.SessionID)
		})
	}
}
