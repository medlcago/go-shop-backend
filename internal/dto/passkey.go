package dto

import (
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
)

type UpdatePasskeyNameRequest struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
}

type PasskeyResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       *string    `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

type BeginPasskeyRegistrationResponse struct {
	*protocol.CredentialCreation
	SessionID string `json:"session_id"`
}

type BeginPasskeyDiscoverableLoginResponse struct {
	*protocol.CredentialAssertion
	SessionID string `json:"session_id"`
}
