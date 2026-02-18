package mocks

import (
	"context"
	"go-shop-backend/internal/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type OrderServiceMock struct {
	mock.Mock
}

func (o *OrderServiceMock) CreateOrder(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID) (*dto.OrderResponse, error) {
	args := o.Called(ctx, userID, sessionID)
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (o *OrderServiceMock) GetOrder(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID) (*dto.OrderResponse, error) {
	args := o.Called(ctx, userID, sessionID, orderID)
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (o *OrderServiceMock) GetOrders(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, req dto.ListOrderRequest) ([]*dto.OrderResponse, int64, error) {
	args := o.Called(ctx, userID, sessionID, req)
	return args.Get(0).([]*dto.OrderResponse), args.Get(1).(int64), args.Error(2)
}

func (o *OrderServiceMock) AddItem(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID, req dto.AddOrderItemRequest) (*dto.OrderResponse, error) {
	args := o.Called(ctx, userID, sessionID, orderID, req)
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}

func (o *OrderServiceMock) DeleteItem(ctx context.Context, userID *uuid.UUID, sessionID uuid.UUID, orderID uuid.UUID, itemID uuid.UUID) (*dto.OrderResponse, error) {
	args := o.Called(ctx, userID, sessionID, orderID, itemID)
	return args.Get(0).(*dto.OrderResponse), args.Error(1)
}
