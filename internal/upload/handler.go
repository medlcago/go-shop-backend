package upload

import (
	"context"

	"github.com/google/uuid"
)

type Handler interface {
	CanSignURL(ctx context.Context, entityType string, entityID uuid.UUID) error
	CanAttach(ctx context.Context, entityType string, entityID uuid.UUID) error
}
