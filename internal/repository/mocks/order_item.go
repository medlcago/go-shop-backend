package mocks

import (
	"context"
	"go-shop-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type OrderItemRepositoryMock struct {
	mock.Mock
}

func (o *OrderItemRepositoryMock) GetItem(ctx context.Context, productID uuid.UUID, orderID uuid.UUID) (*models.OrderItem, error) {
	args := o.Called(ctx, productID, orderID)
	return args.Get(0).(*models.OrderItem), args.Error(1)
}

func (o *OrderItemRepositoryMock) AddItem(ctx context.Context, orderItem *models.OrderItem) error {
	args := o.Called(ctx, orderItem)
	return args.Error(0)
}

func (o *OrderItemRepositoryMock) UpdateQuantity(ctx context.Context, itemID uuid.UUID, qty int) error {
	args := o.Called(ctx, itemID, qty)
	return args.Error(0)
}

func (o *OrderItemRepositoryMock) DeleteItem(ctx context.Context, orderID uuid.UUID, productID uuid.UUID) error {
	args := o.Called(ctx, orderID, productID)
	return args.Error(0)
}
