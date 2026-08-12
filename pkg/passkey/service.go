package passkey

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type Config struct {
	RPDisplayName string
	RPID          string
	RPOrigins     []string
}

type Service struct {
	wa       *webauthn.WebAuthn
	sessions SessionStore
}

func NewService(cfg Config, sessions SessionStore) (*Service, error) {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: cfg.RPDisplayName,
		RPID:          cfg.RPID,
		RPOrigins:     cfg.RPOrigins,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize webauthn: %w", err)
	}

	return &Service{
		wa:       wa,
		sessions: sessions,
	}, nil
}

func newSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *Service) BeginRegistration(ctx context.Context, user User) (*CredentialCreation, string, error) {
	credentials := user.WebAuthnCredentials()

	excludeList := make([]protocol.CredentialDescriptor, 0, len(credentials))
	for _, cred := range credentials {
		excludeList = append(excludeList, cred.Descriptor())
	}

	opts := []webauthn.RegistrationOption{
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			AuthenticatorAttachment: protocol.Platform,
			UserVerification:        protocol.VerificationRequired,
			ResidentKey:             protocol.ResidentKeyRequirementRequired,
		}),
		webauthn.WithExclusions(excludeList),
	}

	credential, session, err := s.wa.BeginRegistration(user, opts...)
	if err != nil {
		return nil, "", fmt.Errorf("passkey: begin registration: %w", err)
	}

	id, err := newSessionID()
	if err != nil {
		return nil, "", fmt.Errorf("passkey: new session id: %w", err)
	}

	if err := s.sessions.SaveSession(ctx, id, session); err != nil {
		return nil, "", fmt.Errorf("passkey: save session: %w", err)
	}

	return credential, id, err
}

func (s *Service) FinishRegistration(
	ctx context.Context,
	user User,
	sessionID string,
	data []byte,
) (*Credential, error) {
	session, err := s.sessions.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("passkey: get session: %w", err)
	}

	defer func() {
		_ = s.sessions.DeleteSession(ctx, sessionID)
	}()

	parsed, err := protocol.ParseCredentialCreationResponseBytes(data)
	if err != nil {
		return nil, fmt.Errorf("passkey: parse credential response: %w", err)
	}

	credential, err := s.wa.CreateCredential(user, *session, parsed)
	if err != nil {
		return nil, fmt.Errorf("passkey: save credential: %w", err)
	}

	return credential, nil
}

func (s *Service) BeginDiscoverableLogin(ctx context.Context) (*CredentialAssertion, string, error) {
	credential, session, err := s.wa.BeginDiscoverableLogin()
	if err != nil {
		return nil, "", fmt.Errorf("passkey: begin discoverable login: %w", err)
	}

	id, err := newSessionID()
	if err != nil {
		return nil, "", err
	}

	if err := s.sessions.SaveSession(ctx, id, session); err != nil {
		return nil, "", fmt.Errorf("passkey: save session: %w", err)
	}

	return credential, id, nil
}

func (s *Service) FinishDiscoverableLogin(
	ctx context.Context,
	handler DiscoverableUserHandler,
	sessionID string,
	data []byte,
) (User, *Credential, error) {
	session, err := s.sessions.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("passkey: get session: %w", err)
	}

	defer func() {
		_ = s.sessions.DeleteSession(ctx, sessionID)
	}()

	parsed, err := protocol.ParseCredentialRequestResponseBytes(data)
	if err != nil {
		return nil, nil, fmt.Errorf("passkey: parse credential response: %w", err)
	}

	user, credential, err := s.wa.ValidatePasskeyLogin(handler, *session, parsed)
	if err != nil {
		return nil, nil, fmt.Errorf("passkey: validate passkey login: %w", err)
	}

	return user, credential, nil
}
