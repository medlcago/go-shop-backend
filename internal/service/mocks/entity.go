package mocks

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

var _ service.EntityService = (*EntityServiceMock)(nil)

type EntityServiceMock struct {
	mock.Mock
}

func (e *EntityServiceMock) Exists(ctx context.Context, entityType dto.EntityType, id uuid.UUID) (bool, error) {
	args := e.Called(ctx, entityType, id)
	return args.Bool(0), args.Error(1)
}
