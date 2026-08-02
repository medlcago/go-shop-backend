package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreatePaymentRequest struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`

	PaymentType string `json:"payment_type" validate:"required_without=PaymentMethodID,excluded_with=PaymentMethodID,omitempty,oneof=redirect embedded"`

	SavePaymentMethod bool `json:"save_payment_method"`

	PaymentMethodID uuid.UUID `json:"payment_method_id" validate:"required_without=PaymentType,excluded_with=PaymentType"`
}

type PaymentResponse struct {
	Status            string `json:"status,omitempty"`
	ConfirmationURL   string `json:"confirmation_url,omitempty"`
	ConfirmationToken string `json:"confirmation_token,omitempty"`
}

type UserPaymentMethodResponse struct {
	ID                      uuid.UUID `json:"id"`
	UserID                  uuid.UUID `json:"user_id"`
	Title                   string    `json:"title"`
	Provider                string    `json:"provider"`
	ProviderCustomerID      *string   `json:"provider_customer_id"`
	ProviderPaymentMethodID string    `json:"provider_payment_method_id"`
	Type                    string    `json:"type"`
	Brand                   *string   `json:"brand"`
	Last4                   *string   `json:"last4"`
	ExpMonth                *int16    `json:"exp_month"`
	ExpYear                 *int16    `json:"exp_year"`
	IsDefault               bool      `json:"is_default"`
	LastUsedAt              time.Time `json:"last_used_at"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}
