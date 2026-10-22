package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const userCtxKey contextKey = "user"

type User struct {
	UserID    *uuid.UUID
	SessionID *uuid.UUID
	Role      string
}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userCtxKey).(User)
	return user, ok
}
