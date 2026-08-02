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

func (p PaymentType) IsValid() bool {
	switch p {
	case PaymentTypeRedirect, PaymentTypeEmbedded:
		return true
	}

	return false
}

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

type Card struct {
	Last4       string `json:"last4"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear  string `json:"expiry_year"`
	CardType    string `json:"card_type"`
}

type PaymentMethod struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Title string `json:"title"`
	Saved bool   `json:"saved"`
	Card  Card   `json:"card"`
}

type CreatePaymentRequest struct {
	Amount            Amount      `json:"amount"`
	PaymentType       PaymentType `json:"payment_type"`
	Capture           bool        `json:"capture"`
	PaymentMethodID   string      `json:"payment_method_id"`
	SavePaymentMethod bool        `json:"save_payment_method"`
	Description       string      `json:"description"`
	Metadata          Metadata    `json:"metadata"`
}

type CancellationDetails struct {
	Party  string `json:"party,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type Payment struct {
	ID                  string               `json:"id"`
	Metadata            Metadata             `json:"metadata"`
	Status              PaymentStatus        `json:"status"`
	Amount              Amount               `json:"amount"`
	Description         string               `json:"description"`
	ConfirmationURL     string               `json:"confirmation_url,omitempty"`
	ConfirmationToken   string               `json:"confirmation_token,omitempty"`
	CancellationDetails *CancellationDetails `json:"cancellation_details,omitempty"`
}

type WebhookEvent struct {
	Status              PaymentStatus        `json:"status"`
	Amount              Amount               `json:"amount"`
	PaymentID           string               `json:"payment_id"`
	Metadata            Metadata             `json:"metadata"`
	PaymentMethod       PaymentMethod        `json:"payment_method"`
	CancellationDetails *CancellationDetails `json:"cancellation_details,omitempty"`
}
