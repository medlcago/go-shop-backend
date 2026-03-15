package models

import (
	"go-shop-backend/pkg/apperrors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrder_CanEdit(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected bool
	}{
		{
			name:     "draft order should be editable",
			status:   OrderStatusDraft,
			expected: true,
		},
		{
			name:     "pending order should not be editable",
			status:   OrderStatusPending,
			expected: false,
		},
		{
			name:     "paid order should not be editable",
			status:   OrderStatusPaid,
			expected: false,
		},
		{
			name:     "canceled order should not be editable",
			status:   OrderStatusCanceled,
			expected: false,
		},
		{
			name:     "completed order should not be editable",
			status:   OrderStatusCompleted,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{Status: tt.status}
			assert.Equal(t, tt.expected, order.CanEdit())
		})
	}
}

func TestOrder_HasItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []OrderItem
		expected bool
	}{
		{
			name:     "order with items should return true",
			items:    []OrderItem{{ID: uuid.New()}},
			expected: true,
		},
		{
			name:     "empty order should return false",
			items:    []OrderItem{},
			expected: false,
		},
		{
			name:     "nil items should return false",
			items:    nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &Order{Items: tt.items}
			assert.Equal(t, tt.expected, order.HasItems())
		})
	}
}

func TestOrder_Checkout(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	tests := []struct {
		name        string
		order       *Order
		userID      uuid.UUID
		expectedErr error
		expected    *Order
	}{
		{
			name: "successful checkout with existing user",
			order: &Order{
				ID:        uuid.New(),
				SessionID: sessionID,
				UserID:    &userID,
				Status:    OrderStatusDraft,
				Items:     []OrderItem{{Quantity: 1, UnitPrice: 1000}},
			},
			userID:      userID,
			expectedErr: nil,
			expected: &Order{
				UserID: &userID,
				Status: OrderStatusPending,
			},
		},
		{
			name: "successful checkout without user",
			order: &Order{
				ID:        uuid.New(),
				SessionID: sessionID,
				UserID:    nil,
				Status:    OrderStatusDraft,
				Items:     []OrderItem{{Quantity: 1, UnitPrice: 1000}},
			},
			userID:      userID,
			expectedErr: nil,
			expected: &Order{
				UserID: &userID,
				Status: OrderStatusPending,
			},
		},
		{
			name: "checkout with non-draft order should fail",
			order: &Order{
				ID:        uuid.New(),
				SessionID: sessionID,
				UserID:    &userID,
				Status:    OrderStatusPending,
				Items:     []OrderItem{{Quantity: 1, UnitPrice: 1000}},
			},
			userID:      userID,
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
		{
			name: "checkout with empty order should fail",
			order: &Order{
				ID:        uuid.New(),
				SessionID: sessionID,
				UserID:    &userID,
				Status:    OrderStatusDraft,
				Items:     []OrderItem{},
			},
			userID:      userID,
			expectedErr: apperrors.ErrEmptyOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Checkout(tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Status, tt.order.Status)
				assert.Equal(t, tt.expected.UserID, tt.order.UserID)
			}
		})
	}
}

func TestOrder_MarkPaid(t *testing.T) {
	tests := []struct {
		name        string
		order       *Order
		expectedErr error
		validate    func(t *testing.T, order *Order)
	}{
		{
			name: "successfully mark pending order as paid",
			order: &Order{
				Status: OrderStatusPending,
			},
			expectedErr: nil,
			validate: func(t *testing.T, order *Order) {
				assert.Equal(t, OrderStatusPaid, order.Status)
				assert.NotNil(t, order.PaidAt)
				assert.WithinDuration(t, time.Now(), *order.PaidAt, time.Second)
			},
		},
		{
			name: "mark draft order as paid should fail",
			order: &Order{
				Status: OrderStatusDraft,
			},
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
		{
			name: "mark paid order as paid should fail",
			order: &Order{
				Status: OrderStatusPaid,
			},
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
		{
			name: "mark canceled order as paid should fail",
			order: &Order{
				Status: OrderStatusCanceled,
			},
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
		{
			name: "mark completed order as paid should fail",
			order: &Order{
				Status: OrderStatusCompleted,
			},
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.MarkPaid()

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.order)
				}
			}
		})
	}
}

func TestOrder_MarkCanceled(t *testing.T) {
	tests := []struct {
		name        string
		order       *Order
		expectedErr error
		validate    func(t *testing.T, order *Order)
	}{
		{
			name: "successfully cancel pending order",
			order: &Order{
				Status: OrderStatusPending,
			},
			expectedErr: nil,
			validate: func(t *testing.T, order *Order) {
				assert.Equal(t, OrderStatusCanceled, order.Status)
				assert.NotNil(t, order.CanceledAt)
				assert.WithinDuration(t, time.Now(), *order.CanceledAt, time.Second)
			},
		},
		{
			name: "cancel draft order should fail",
			order: &Order{
				Status: OrderStatusDraft,
			},
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
		{
			name: "cancel paid order should fail",
			order: &Order{
				Status: OrderStatusPaid,
			},
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
		{
			name: "cancel already canceled order should fail",
			order: &Order{
				Status: OrderStatusCanceled,
			},
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
		{
			name: "cancel completed order should fail",
			order: &Order{
				Status: OrderStatusCompleted,
			},
			expectedErr: apperrors.ErrInvalidOrderStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.MarkCanceled()

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.order)
				}
			}
		})
	}
}

func TestOrder_Recalculate(t *testing.T) {
	tests := []struct {
		name  string
		order *Order
		total int64
	}{
		{
			name: "single item",
			order: &Order{
				Items: []OrderItem{
					{Quantity: 2, UnitPrice: 1000},
				},
			},
			total: 2000,
		},
		{
			name: "multiple items",
			order: &Order{
				Items: []OrderItem{
					{Quantity: 2, UnitPrice: 1000},
					{Quantity: 3, UnitPrice: 500},
					{Quantity: 1, UnitPrice: 2500},
				},
			},
			total: 6000,
		},
		{
			name: "empty order",
			order: &Order{
				Items: []OrderItem{},
			},
			total: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.order.Recalculate()

			assert.Equal(t, tt.total, tt.order.TotalAmount)
		})
	}
}

func TestOrder_SetPaymentInfo(t *testing.T) {
	paymentID := "pay_123456"
	providerName := "stripe"
	expiresAt := time.Now().Add(24 * time.Hour)

	order := &Order{}
	order.SetPaymentInfo(paymentID, providerName, expiresAt)

	assert.NotNil(t, order.PaymentID)
	assert.Equal(t, paymentID, *order.PaymentID)

	assert.NotNil(t, order.ProviderName)
	assert.Equal(t, providerName, *order.ProviderName)

	assert.NotNil(t, order.ExpiresAt)
	assert.Equal(t, expiresAt, *order.ExpiresAt)
}

func TestOrder_CompleteLifecycle(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	order := &Order{
		ID:        uuid.New(),
		SessionID: sessionID,
		Status:    OrderStatusDraft,
		Items: []OrderItem{
			{Quantity: 2, UnitPrice: 1000},
			{Quantity: 1, UnitPrice: 500},
		},
	}

	assert.True(t, order.CanEdit())
	assert.True(t, order.HasItems())

	err := order.Checkout(userID)
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusPending, order.Status)
	assert.Equal(t, &userID, order.UserID)

	order.Recalculate()
	assert.Equal(t, int64(2500), order.TotalAmount)

	paymentID := "pay_123456"
	providerName := "yookassa"
	expiresAt := time.Now().Add(1 * time.Hour)
	order.SetPaymentInfo(paymentID, providerName, expiresAt)

	err = order.MarkPaid()
	assert.NoError(t, err)
	assert.Equal(t, OrderStatusPaid, order.Status)
	assert.NotNil(t, order.PaidAt)

	assert.False(t, order.CanEdit())
}
