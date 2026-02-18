package mocks

import (
	"context"
	"go-shop-backend/internal/dto"
	"go-shop-backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type OrderRepositoryMock struct {
	mock.Mock
}

func (o *OrderRepositoryMock) Create(ctx context.Context, order *models.Order) error {
	args := o.Called(ctx, order)
	return args.Error(0)
}

func (o *OrderRepositoryMock) Update(ctx context.Context, order *models.Order) error {
	args := o.Called(ctx, order)
	return args.Error(0)
}

func (o *OrderRepositoryMock) GetByOwner(ctx context.Context, orderID uuid.UUID, userID *uuid.UUID, sessionID uuid.UUID, preload bool) (*models.Order, error) {
	args := o.Called(ctx, orderID, userID, sessionID, preload)
	return args.Get(0).(*models.Order), args.Error(1)
}

func (o *OrderRepositoryMock) GetListByOwner(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, req dto.ListOrderRequest) ([]*models.Order, int64, error) {
	args := o.Called(ctx, userID, sessionID, req)
	return args.Get(0).([]*models.Order), args.Get(1).(int64), args.Error(2)

}
