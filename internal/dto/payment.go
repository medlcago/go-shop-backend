package dto

import "github.com/google/uuid"

type CreatePaymentRequest struct {
	OrderID uuid.UUID `json:"order_id" validate:"required"`
	Type    string    `json:"type" validate:"required,oneof=redirect embedded"`
}

type PaymentResponse struct {
	ConfirmationURL   string `json:"confirmation_url,omitempty"`
	ConfirmationToken string `json:"confirmation_token,omitempty"`
}
