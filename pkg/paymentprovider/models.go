package paymentprovider

import (
	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending           PaymentStatus = "pending"
	PaymentStatusWaitingForCapture PaymentStatus = "waiting_for_capture"
	PaymentStatusSucceeded         PaymentStatus = "succeeded"
	PaymentStatusCanceled          PaymentStatus = "canceled"
)

type PaymentType string

const (
	PaymentTypeRedirect PaymentType = "redirect"
	PaymentTypeEmbedded PaymentType = "embedded"
)

type Currency string

const (
	CurrencyRUB Currency = "RUB"
)

func (c Currency) IsValid() bool {
	switch c {
	case CurrencyRUB:
		return true
	}

	return false
}

type Amount struct {
	Value    string   `json:"value"`
	Currency Currency `json:"currency"`
}

type Metadata struct {
	UserID  uuid.UUID `json:"user_id"`
	OrderID uuid.UUID `json:"order_id"`
}

type CreatePaymentRequest struct {
	Amount   Amount      `json:"amount"`
	Type     PaymentType `json:"type"`
	Capture  bool        `json:"capture"`
	Metadata Metadata    `json:"metadata"`
}

type Payment struct {
	ID                string        `json:"id"`
	Metadata          Metadata      `json:"metadata"`
	Status            PaymentStatus `json:"status"`
	Amount            Amount        `json:"amount"`
	Description       string        `json:"description"`
	ConfirmationURL   string        `json:"confirmation_url,omitempty"`
	ConfirmationToken string        `json:"confirmation_token,omitempty"`
}

type WebhookEvent struct {
	Status    PaymentStatus `json:"status"`
	Amount    Amount        `json:"amount"`
	PaymentID string        `json:"payment_id"`
	Metadata  Metadata      `json:"metadata"`
}
