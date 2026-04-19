package models

import (
	"database/sql"
	"go-shop-backend/pkg/apperror"
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
			expectedErr: apperror.ErrInvalidOrderStatus,
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
			expectedErr: apperror.ErrEmptyOrder,
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

func TestOrder_Pay(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name        string
		order       *Order
		userID      uuid.UUID
		expectedErr error
		expected    *Order
	}{
		{
			name: "successful payment of pending order",
			order: &Order{
				ID:          uuid.New(),
				UserID:      &userID,
				Status:      OrderStatusPending,
				TotalAmount: 1000,
			},
			userID:      userID,
			expectedErr: nil,
			expected: &Order{
				Status: OrderStatusPaid,
				PaidAt: &now,
			},
		},
		{
			name: "payment with wrong user should fail",
			order: &Order{
				ID:     uuid.New(),
				UserID: &userID,
				Status: OrderStatusPending,
			},
			userID:      otherUserID,
			expectedErr: apperror.ErrForbidden,
		},
		{
			name: "payment of draft order should fail",
			order: &Order{
				ID:     uuid.New(),
				UserID: &userID,
				Status: OrderStatusDraft,
			},
			userID:      userID,
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "payment of already paid order should fail",
			order: &Order{
				ID:     uuid.New(),
				UserID: &userID,
				Status: OrderStatusPaid,
				PaidAt: &now,
			},
			userID:      userID,
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "payment of canceled order should fail",
			order: &Order{
				ID:         uuid.New(),
				UserID:     &userID,
				Status:     OrderStatusCanceled,
				CanceledAt: &now,
			},
			userID:      userID,
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "payment of completed order should fail",
			order: &Order{
				ID:          uuid.New(),
				UserID:      &userID,
				Status:      OrderStatusCompleted,
				CompletedAt: sql.NullTime{Time: now, Valid: true},
			},
			userID:      userID,
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "payment with nil UserID should fail",
			order: &Order{
				ID:     uuid.New(),
				UserID: nil,
				Status: OrderStatusPending,
			},
			userID:      userID,
			expectedErr: apperror.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Pay(tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Status, tt.order.Status)
				assert.NotNil(t, tt.expected.PaidAt)
				assert.Equal(t, now, *tt.expected.PaidAt)
			}
		})
	}
}

func TestOrder_Cancel(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	now := time.Now().UTC()

	tests := []struct {
		name        string
		order       *Order
		userID      uuid.UUID
		expectedErr error
		expected    *Order
	}{
		{
			name: "successful cancellation of pending order",
			order: &Order{
				ID:          uuid.New(),
				UserID:      &userID,
				Status:      OrderStatusPending,
				TotalAmount: 1000,
			},
			userID:      userID,
			expectedErr: nil,
			expected: &Order{
				Status:     OrderStatusCanceled,
				CanceledAt: &now,
			},
		},
		{
			name: "cancellation with wrong user should fail",
			order: &Order{
				ID:     uuid.New(),
				UserID: &userID,
				Status: OrderStatusPending,
			},
			userID:      otherUserID,
			expectedErr: apperror.ErrForbidden,
		},
		{
			name: "cancellation of draft order should fail",
			order: &Order{
				ID:     uuid.New(),
				UserID: &userID,
				Status: OrderStatusDraft,
			},
			userID:      userID,
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "cancellation of paid order should fail",
			order: &Order{
				ID:     uuid.New(),
				UserID: &userID,
				Status: OrderStatusPaid,
				PaidAt: &now,
			},
			userID:      userID,
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "cancellation of already canceled order should fail",
			order: &Order{
				ID:         uuid.New(),
				UserID:     &userID,
				Status:     OrderStatusCanceled,
				CanceledAt: &now,
			},
			userID:      userID,
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "cancellation of completed order should fail",
			order: &Order{
				ID:          uuid.New(),
				UserID:      &userID,
				Status:      OrderStatusCompleted,
				CompletedAt: sql.NullTime{Time: now, Valid: true},
			},
			userID:      userID,
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "cancellation with nil UserID should fail",
			order: &Order{
				ID:     uuid.New(),
				UserID: nil,
				Status: OrderStatusPending,
			},
			userID:      userID,
			expectedErr: apperror.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Cancel(tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Status, tt.order.Status)
				assert.NotNil(t, tt.expected.CanceledAt)
				assert.Equal(t, now, *tt.expected.CanceledAt)
			}
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
