package models

import (
	"go-shop-backend/pkg/apperror"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

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
	tests := []struct {
		name           string
		order          *Order
		expectedErr    error
		expectedStatus OrderStatus
	}{
		{
			name: "successful checkout",
			order: &Order{
				ID:     uuid.New(),
				Status: OrderStatusDraft,
				Items:  []OrderItem{{Quantity: 1, UnitPrice: 1000}},
			},
			expectedStatus: OrderStatusPending,
		},
		{
			name: "checkout with non-draft order should fail",
			order: &Order{
				ID:     uuid.New(),
				Status: OrderStatusPending,
				Items:  []OrderItem{{Quantity: 1, UnitPrice: 1000}},
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "checkout with empty order should fail",
			order: &Order{
				ID:     uuid.New(),
				Status: OrderStatusDraft,
				Items:  []OrderItem{},
			},
			expectedErr: apperror.ErrEmptyOrder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Checkout()

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, tt.order.Status)
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
	now := time.Now().UTC()

	tests := []struct {
		name        string
		order       *Order
		expectedErr error
		expected    *Order
	}{
		{
			name: "successful payment of pending order",
			order: &Order{
				ID:          uuid.New(),
				Status:      OrderStatusPending,
				TotalAmount: 1000,
			},
			expectedErr: nil,
			expected: &Order{
				Status: OrderStatusPaid,
				PaidAt: &now,
			},
		},
		{
			name: "payment of draft order should fail",
			order: &Order{
				ID:     uuid.New(),
				Status: OrderStatusDraft,
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "payment of already paid order should fail",
			order: &Order{
				ID:     uuid.New(),
				Status: OrderStatusPaid,
				PaidAt: &now,
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "payment of canceled order should fail",
			order: &Order{
				ID:         uuid.New(),
				Status:     OrderStatusCanceled,
				CanceledAt: &now,
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "payment of completed order should fail",
			order: &Order{
				ID:          uuid.New(),
				Status:      OrderStatusCompleted,
				CompletedAt: &now,
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Pay()

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
	now := time.Now().UTC()

	tests := []struct {
		name        string
		order       *Order
		expectedErr error
		expected    *Order
	}{
		{
			name: "successful cancellation of pending order",
			order: &Order{
				ID:          uuid.New(),
				Status:      OrderStatusPending,
				TotalAmount: 1000,
			},
			expectedErr: nil,
			expected: &Order{
				Status:     OrderStatusCanceled,
				CanceledAt: &now,
			},
		},
		{
			name: "cancellation of draft order should fail",
			order: &Order{
				ID:     uuid.New(),
				Status: OrderStatusDraft,
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "cancellation of paid order should fail",
			order: &Order{
				ID:     uuid.New(),
				Status: OrderStatusPaid,
				PaidAt: &now,
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "cancellation of already canceled order should fail",
			order: &Order{
				ID:         uuid.New(),
				Status:     OrderStatusCanceled,
				CanceledAt: &now,
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
		{
			name: "cancellation of completed order should fail",
			order: &Order{
				ID:          uuid.New(),
				Status:      OrderStatusCompleted,
				CompletedAt: &now,
			},
			expectedErr: apperror.ErrInvalidOrderStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Cancel()

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

	order := &Order{}
	order.SetPaymentInfo(paymentID, providerName)

	assert.NotNil(t, order.PaymentID)
	assert.Equal(t, paymentID, *order.PaymentID)

	assert.NotNil(t, order.ProviderName)
	assert.Equal(t, providerName, *order.ProviderName)
}

func TestOrderStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected bool
	}{
		{
			name:     "valid status: draft",
			status:   OrderStatusDraft,
			expected: true,
		},
		{
			name:     "valid status: pending",
			status:   OrderStatusPending,
			expected: true,
		},
		{
			name:     "valid status: paid",
			status:   OrderStatusPaid,
			expected: true,
		},
		{
			name:     "valid status: canceled",
			status:   OrderStatusCanceled,
			expected: true,
		},
		{
			name:     "valid status: completed",
			status:   OrderStatusCompleted,
			expected: true,
		},
		{
			name:     "invalid status: empty string",
			status:   OrderStatus(""),
			expected: false,
		},
		{
			name:     "invalid status: unknown value",
			status:   OrderStatus("unknown"),
			expected: false,
		},
		{
			name:     "invalid status: random string",
			status:   OrderStatus("some_random_value"),
			expected: false,
		},
		{
			name:     "invalid status: uppercase value",
			status:   OrderStatus("DRAFT"),
			expected: false,
		},
		{
			name:     "invalid status: partial value",
			status:   OrderStatus("pend"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestOrderStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		current  OrderStatus
		next     OrderStatus
		expected bool
	}{
		// Valid transitions
		{
			name:     "draft -> pending",
			current:  OrderStatusDraft,
			next:     OrderStatusPending,
			expected: true,
		},
		{
			name:     "pending -> paid",
			current:  OrderStatusPending,
			next:     OrderStatusPaid,
			expected: true,
		},
		{
			name:     "pending -> canceled",
			current:  OrderStatusPending,
			next:     OrderStatusCanceled,
			expected: true,
		},
		{
			name:     "paid -> completed",
			current:  OrderStatusPaid,
			next:     OrderStatusCompleted,
			expected: true,
		},

		// Same status transitions (forbidden)
		{
			name:     "draft -> draft",
			current:  OrderStatusDraft,
			next:     OrderStatusDraft,
			expected: false,
		},
		{
			name:     "pending -> pending",
			current:  OrderStatusPending,
			next:     OrderStatusPending,
			expected: false,
		},
		{
			name:     "paid -> paid",
			current:  OrderStatusPaid,
			next:     OrderStatusPaid,
			expected: false,
		},
		{
			name:     "canceled -> canceled",
			current:  OrderStatusCanceled,
			next:     OrderStatusCanceled,
			expected: false,
		},
		{
			name:     "completed -> completed",
			current:  OrderStatusCompleted,
			next:     OrderStatusCompleted,
			expected: false,
		},

		// Invalid transitions
		{
			name:     "draft -> paid (skip pending)",
			current:  OrderStatusDraft,
			next:     OrderStatusPaid,
			expected: false,
		},
		{
			name:     "draft -> canceled",
			current:  OrderStatusDraft,
			next:     OrderStatusCanceled,
			expected: false,
		},
		{
			name:     "draft -> completed",
			current:  OrderStatusDraft,
			next:     OrderStatusCompleted,
			expected: false,
		},
		{
			name:     "pending -> completed (skip paid)",
			current:  OrderStatusPending,
			next:     OrderStatusCompleted,
			expected: false,
		},
		{
			name:     "paid -> canceled",
			current:  OrderStatusPaid,
			next:     OrderStatusCanceled,
			expected: false,
		},
		{
			name:     "paid -> pending (backwards)",
			current:  OrderStatusPaid,
			next:     OrderStatusPending,
			expected: false,
		},
		{
			name:     "paid -> draft (backwards)",
			current:  OrderStatusPaid,
			next:     OrderStatusDraft,
			expected: false,
		},
		{
			name:     "canceled -> any status",
			current:  OrderStatusCanceled,
			next:     OrderStatusPending,
			expected: false,
		},
		{
			name:     "canceled -> paid",
			current:  OrderStatusCanceled,
			next:     OrderStatusPaid,
			expected: false,
		},
		{
			name:     "completed -> any status",
			current:  OrderStatusCompleted,
			next:     OrderStatusPending,
			expected: false,
		},
		{
			name:     "completed -> paid",
			current:  OrderStatusCompleted,
			next:     OrderStatusPaid,
			expected: false,
		},

		// Invalid current status
		{
			name:     "invalid current status (empty) -> pending",
			current:  OrderStatus(""),
			next:     OrderStatusPending,
			expected: false,
		},
		{
			name:     "invalid current status (unknown) -> pending",
			current:  OrderStatus("unknown"),
			next:     OrderStatusPending,
			expected: false,
		},

		// Invalid next status
		{
			name:     "pending -> invalid next status (empty)",
			current:  OrderStatusPending,
			next:     OrderStatus(""),
			expected: false,
		},
		{
			name:     "pending -> invalid next status (unknown)",
			current:  OrderStatusPending,
			next:     OrderStatus("unknown"),
			expected: false,
		},

		// Both invalid
		{
			name:     "both invalid",
			current:  OrderStatus("invalid1"),
			next:     OrderStatus("invalid2"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.current.CanTransitionTo(tt.next))
		})
	}
}
