package paymentprovider

type Provider interface {
	CreatePayment(req *CreatePaymentRequest) (*Payment, error)
	CancelPayment(paymentID string) error
	ParseWebhook(body []byte) (*WebhookEvent, error)
	GetName() string
}
