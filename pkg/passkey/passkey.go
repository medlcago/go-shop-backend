package passkey

import (
	"context"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type (
	User        = webauthn.User
	Credential  = webauthn.Credential
	SessionData = webauthn.SessionData

	CredentialCreation  = protocol.CredentialCreation
	CredentialAssertion = protocol.CredentialAssertion

	DiscoverableUserHandler = webauthn.DiscoverableUserHandler
)

type SessionStore interface {
	SaveSession(ctx context.Context, sessionID string, data *SessionData) error
	GetSession(ctx context.Context, sessionID string) (*SessionData, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type Manager interface {
	BeginRegistration(ctx context.Context, user User) (*CredentialCreation, string, error)
	FinishRegistration(ctx context.Context, user User, sessionID string, data []byte) (*Credential, error)
	BeginDiscoverableLogin(ctx context.Context) (*CredentialAssertion, string, error)
	FinishDiscoverableLogin(ctx context.Context, handler DiscoverableUserHandler, sessionID string, data []byte) (User, *Credential, error)
}
