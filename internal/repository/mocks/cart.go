package mocks

import (
	"context"
	"go-shop-backend/internal/models"
	"go-shop-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

var _ repository.CartRepository = (*CartRepositoryMock)(nil)

type CartRepositoryMock struct {
	mock.Mock
}

func (c *CartRepositoryMock) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Cart, error) {
	args := c.Called(ctx, userID)
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (c *CartRepositoryMock) GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.Cart, error) {
	args := c.Called(ctx, sessionID)
	return args.Get(0).(*models.Cart), args.Error(1)
}

func (c *CartRepositoryMock) Create(ctx context.Context, cart *models.Cart) error {
	args := c.Called(ctx, cart)
	return args.Error(0)
}

func (c *CartRepositoryMock) Save(ctx context.Context, cart *models.Cart) error {
	args := c.Called(ctx, cart)
	return args.Error(0)
}

func (c *CartRepositoryMock) DeleteItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID) error {
	args := c.Called(ctx, cartID, productID)
	return args.Error(0)
}
