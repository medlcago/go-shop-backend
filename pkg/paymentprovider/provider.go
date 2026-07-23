package paymentprovider

import "context"

type Provider interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest, idempotencyKey string) (*Payment, error)
	CancelPayment(ctx context.Context, paymentID string, idempotencyKey string) error
	ParseWebhook(body []byte) (*WebhookEvent, error)
	GetName() string
}
