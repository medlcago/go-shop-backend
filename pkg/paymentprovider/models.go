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
	Card  Card   `json:"card"`
}

type CreatePaymentRequest struct {
	Amount            string      `json:"amount"`
	Currency          Currency    `json:"currency"`
	PaymentType       PaymentType `json:"payment_type"`
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
	Amount              string               `json:"amount"`
	Currency            Currency             `json:"currency"`
	Metadata            Metadata             `json:"metadata"`
	Status              PaymentStatus        `json:"status"`
	Description         string               `json:"description"`
	ConfirmationURL     string               `json:"confirmation_url,omitempty"`
	ConfirmationToken   string               `json:"confirmation_token,omitempty"`
	CancellationDetails *CancellationDetails `json:"cancellation_details,omitempty"`
}

type WebhookEvent struct {
	Event               string               `json:"event"`
	Status              PaymentStatus        `json:"status"`
	Amount              string               `json:"amount"`
	Currency            Currency             `json:"currency"`
	PaymentID           string               `json:"payment_id"`
	Metadata            Metadata             `json:"metadata"`
	PaymentMethod       PaymentMethod        `json:"payment_method"`
	CancellationDetails *CancellationDetails `json:"cancellation_details,omitempty"`
}
