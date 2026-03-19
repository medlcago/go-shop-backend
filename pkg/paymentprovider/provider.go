package paymentprovider

import "context"

type Provider interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*Payment, error)
	CancelPayment(ctx context.Context, paymentID string) error
	ParseWebhook(body []byte) (*WebhookEvent, error)
	GetName() string
}
