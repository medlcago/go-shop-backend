package paymentprovider

import (
	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusCanceled  PaymentStatus = "canceled"
)

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type Metadata struct {
	UserID  uuid.UUID `json:"user_id"`
	OrderID uuid.UUID `json:"order_id"`
}

type CreatePaymentRequest struct {
	Metadata `json:"metadata"`

	// Amount in the smallest currency unit (e.g., cents for USD, kopeks for RUB)
	Amount int64 `json:"amount"`
}

type Payment struct {
	ID              string        `json:"id"`
	Metadata        Metadata      `json:"metadata"`
	Status          PaymentStatus `json:"status"`
	Amount          Amount        `json:"amount"`
	Description     string        `json:"description"`
	ConfirmationURL string        `json:"confirmation_url"`
}

type WebhookEvent struct {
	Status    PaymentStatus `json:"status"`
	Amount    Amount        `json:"amount"`
	PaymentID string        `json:"payment_id"`
	Metadata  Metadata      `json:"metadata"`
}
