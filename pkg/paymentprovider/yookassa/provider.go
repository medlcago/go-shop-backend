package yookassa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-shop-backend/pkg/paymentprovider"

	"github.com/google/uuid"
	yookassasdk "github.com/rvinnie/yookassa-sdk-go/yookassa"
	yoocommon "github.com/rvinnie/yookassa-sdk-go/yookassa/common"
	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"
	yoowebhook "github.com/rvinnie/yookassa-sdk-go/yookassa/webhook"
	"github.com/shopspring/decimal"
)

const (
	ProviderName = "yookassa"
)

type Provider struct {
	client         *yookassasdk.Client
	paymentHandler *yookassasdk.PaymentHandler
	cfg            *Config
}

type Config struct {
	AccountId string
	SecretKey string
	ReturnURL string
}

func NewConfig(accountId string, secretKey string, returnURL string) *Config {
	return &Config{
		AccountId: accountId,
		SecretKey: secretKey,
		ReturnURL: returnURL,
	}
}

func New(cfg *Config) (*Provider, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	client := yookassasdk.NewClient(cfg.AccountId, cfg.SecretKey)

	return &Provider{
		client:         client,
		paymentHandler: yookassasdk.NewPaymentHandler(client),
		cfg:            cfg,
	}, nil
}

func (p *Provider) CreatePayment(ctx context.Context, req *paymentprovider.CreatePaymentRequest) (*paymentprovider.Payment, error) {
	paymentHandler := p.paymentHandler.WithIdempotencyKey(uuid.NewString())

	rubles := decimal.NewFromInt(req.Amount).Div(decimal.NewFromInt(100))

	payment, err := paymentHandler.CreatePayment(ctx, &yoopayment.Payment{
		Amount: &yoocommon.Amount{
			Value:    rubles.StringFixed(2),
			Currency: "RUB",
		},
		PaymentMethod: yoopayment.PaymentTypeBankCard,
		Confirmation: yoopayment.Redirect{
			Type:      yoopayment.TypeRedirect,
			ReturnURL: p.cfg.ReturnURL,
		},
		Capture:     true,
		Description: fmt.Sprintf("Оплата заказа № %s", req.Metadata.OrderID),
		Metadata:    req.Metadata,
	})

	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to create payment: %w", err)
	}

	paymentLink, err := paymentHandler.ParsePaymentLink(payment)
	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to create payment link: %w", err)
	}

	return &paymentprovider.Payment{
		ID:     payment.ID,
		Status: paymentprovider.PaymentStatusPending,
		Amount: paymentprovider.Amount{
			Value:    payment.Amount.Value,
			Currency: payment.Amount.Currency,
		},
		Description:     payment.Description,
		Metadata:        req.Metadata,
		ConfirmationURL: paymentLink,
	}, nil
}

func (p *Provider) CancelPayment(ctx context.Context, paymentID string) error {
	paymentHandler := p.paymentHandler.WithIdempotencyKey(uuid.NewString())

	_, err := paymentHandler.CancelPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("yookassa: failed to cancel payment: %w", err)
	}
	return nil
}

func (p *Provider) ParseWebhook(body []byte) (*paymentprovider.WebhookEvent, error) {
	var yookassaWebhookEvent yoowebhook.WebhookEvent[yoopayment.Payment]
	if err := json.Unmarshal(body, &yookassaWebhookEvent); err != nil {
		return nil, fmt.Errorf("yookassa: failed to parse webhook event: %w", err)
	}

	switch yookassaWebhookEvent.Type {
	case yoowebhook.WebhookTypeNotification:
		switch yookassaWebhookEvent.Event {
		case yoowebhook.EventPaymentSucceeded, yoowebhook.EventPaymentWaitingForCapture, yoowebhook.EventPaymentCanceled:
		default:
			return nil, fmt.Errorf("yookassa: webhook event type %s not supported", yookassaWebhookEvent.Event)
		}
	default:
		return nil, fmt.Errorf("yookassa: webhook type %s not supported", yookassaWebhookEvent.Type)
	}

	metadataBytes, err := json.Marshal(yookassaWebhookEvent.Object.Metadata)
	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to parse webhook event: %w", err)
	}

	var metadata paymentprovider.Metadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, fmt.Errorf("yookassa: failed to parse webhook event: invalid metadata format: %w", err)
	}

	webhookEvent := &paymentprovider.WebhookEvent{
		Status:    p.parsePaymentStatus(yookassaWebhookEvent.Event),
		PaymentID: yookassaWebhookEvent.Object.ID,
		Metadata:  metadata,
		Amount: paymentprovider.Amount{
			Value:    yookassaWebhookEvent.Object.Amount.Value,
			Currency: yookassaWebhookEvent.Object.Amount.Currency,
		},
	}

	return webhookEvent, nil
}
func (p *Provider) GetName() string {
	return ProviderName
}

func (p *Provider) parsePaymentStatus(event yoowebhook.WebhookEventType) paymentprovider.PaymentStatus {
	switch event {
	case yoowebhook.EventPaymentSucceeded:
		return paymentprovider.PaymentStatusSucceeded
	case yoowebhook.EventPaymentCanceled:
		return paymentprovider.PaymentStatusCanceled
	case yoowebhook.EventPaymentWaitingForCapture:
		return paymentprovider.PaymentStatusPending
	default:
		return paymentprovider.PaymentStatusCanceled
	}
}
