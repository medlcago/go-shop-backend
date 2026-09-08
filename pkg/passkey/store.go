package passkey

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/redis/go-redis/v9"
)

var _ SessionStore = (*sessionStore)(nil)

var (
	ErrSessionNotFound = errors.New("passkey: session not found")
)

const (
	DefaultSessionTTL = 5 * time.Minute
)

type sessionStore struct {
	rdb redis.UniversalClient
	ttl time.Duration
}

func NewSessionStore(rdb redis.UniversalClient, ttl time.Duration) *sessionStore {
	if ttl <= 0 {
		ttl = DefaultSessionTTL
	}

	return &sessionStore{
		rdb: rdb,
		ttl: ttl,
	}
}

func (s *sessionStore) SaveSession(ctx context.Context, sessionID string, data *SessionData) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal session data: %w", err)
	}

	if err := s.rdb.Set(ctx, sessionID, jsonBytes, s.ttl).Err(); err != nil {
		return fmt.Errorf("save session %q: %w", sessionID, err)
	}

	return nil
}

func (s *sessionStore) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	jsonBytes, err := s.rdb.Get(ctx, sessionID).Bytes()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, fmt.Errorf("get session %q: %w", sessionID, ErrSessionNotFound)
	case err != nil:
		return nil, fmt.Errorf("get session %q: %w", sessionID, err)
	}

	var sd webauthn.SessionData
	if err := json.Unmarshal(jsonBytes, &sd); err != nil {
		return nil, fmt.Errorf("unmarshal session data: %w", err)
	}

	return &sd, nil
}

func (s *sessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	if err := s.rdb.Del(ctx, sessionID).Err(); err != nil {
		return fmt.Errorf("delete session %q: %w", sessionID, err)
	}

	return nil
}
