package mocks

import (
	"context"
	"go-shop-backend/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type EntityServiceMock struct {
	mock.Mock
}

func (e *EntityServiceMock) Exists(ctx context.Context, entityType dto.EntityType, id uuid.UUID) error {
	args := e.Called(ctx, entityType, id)
	return args.Error(0)
}
