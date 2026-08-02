package yookassa

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-shop-backend/pkg/paymentprovider"
	"net/http"

	yookassasdk "github.com/rvinnie/yookassa-sdk-go/yookassa"
	yoocommon "github.com/rvinnie/yookassa-sdk-go/yookassa/common"
	yooopts "github.com/rvinnie/yookassa-sdk-go/yookassa/opts"
	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"
	yoowebhook "github.com/rvinnie/yookassa-sdk-go/yookassa/webhook"
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

func New(httpClient *http.Client, cfg *Config) (*Provider, error) {
	if cfg == nil {
		return nil, errors.New("yookassa: config is nil")
	}

	if httpClient == nil {
		return nil, errors.New("yookassa: httpClient is nil")
	}

	client := yookassasdk.NewClient(cfg.AccountId, cfg.SecretKey, yooopts.WithHTTPClient(*httpClient))

	return &Provider{
		client:         client,
		paymentHandler: yookassasdk.NewPaymentHandler(client),
		cfg:            cfg,
	}, nil
}

func (p *Provider) CreatePayment(ctx context.Context, req *paymentprovider.CreatePaymentRequest, idempotencyKey string) (*paymentprovider.Payment, error) {
	if !req.Amount.Currency.IsValid() {
		return nil, fmt.Errorf("yookassa: unsupported currency: %s", req.Amount.Currency)
	}

	paymentHandler := p.paymentHandler.WithIdempotencyKey(idempotencyKey)

	var (
		confirmation  yoopayment.Confirmer
		paymentMethod yoopayment.PaymentMethoder
		err           error
	)

	if req.PaymentType != "" {
		if !req.PaymentType.IsValid() {
			return nil, fmt.Errorf("yookassa: unsupported payment type: %s", req.PaymentType)
		}

		confirmation, err = p.createConfirmation(req.PaymentType)
		if err != nil {
			return nil, fmt.Errorf("yookassa: failed to create confirmation: %w", err)
		}

		paymentMethod = yoopayment.PaymentTypeBankCard
	}

	payment, err := paymentHandler.CreatePayment(ctx, &yoopayment.Payment{
		Amount: &yoocommon.Amount{
			Value:    req.Amount.Value,
			Currency: string(req.Amount.Currency),
		},
		PaymentMethod:     paymentMethod,
		Confirmation:      confirmation,
		Capture:           req.Capture,
		Description:       req.Description,
		Metadata:          req.Metadata,
		SavePaymentMethod: req.SavePaymentMethod,
		PaymentMethodID:   req.PaymentMethodID,
	})

	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to create payment: %w", err)
	}

	response, err := p.buildPaymentResponse(req.PaymentType, payment)
	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to build payment response: %w", err)
	}

	return response, nil
}

func (p *Provider) CancelPayment(ctx context.Context, paymentID string, idempotencyKey string) error {
	paymentHandler := p.paymentHandler.WithIdempotencyKey(idempotencyKey)

	_, err := paymentHandler.CancelPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("yookassa: failed to cancel payment: %w", err)
	}

	return nil
}

func (p *Provider) CapturePayment(ctx context.Context, paymentID string, idempotencyKey string) error {
	paymentHandler := p.paymentHandler.WithIdempotencyKey(idempotencyKey)

	payment, err := paymentHandler.FindPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("yookassa: failed to find payment: %w", err)
	}

	_, err = paymentHandler.CapturePayment(ctx, payment)
	if err != nil {
		return fmt.Errorf("yookassa: failed to capture payment: %w", err)
	}

	return nil
}

func (p *Provider) ParseWebhook(body []byte) (*paymentprovider.WebhookEvent, error) {
	yookassaWebhookEvent, err := p.parseWebhookBody(body)
	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to parse payment webhook body: %w", err)
	}

	switch yookassaWebhookEvent.Type {
	case yoowebhook.WebhookTypeNotification:
		if !p.isSupportedWebhookEvent(yookassaWebhookEvent.Event) {
			return nil, fmt.Errorf("yookassa: unsupported webhook event: %s", yookassaWebhookEvent.Event)
		}
	default:
		return nil, fmt.Errorf("yookassa: webhook type %s not supported", yookassaWebhookEvent.Type)
	}

	paymentMethod, err := p.parsePaymentMethod(yookassaWebhookEvent.Object.PaymentMethod)
	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to parse payment method: %w", err)
	}

	metadata, err := p.parsePaymentMetadata(yookassaWebhookEvent.Object.Metadata)
	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to parse webhook event: failed to extract payment metadata: %w", err)
	}

	status, err := p.parseWebhookEventStatus(yookassaWebhookEvent.Event)
	if err != nil {
		return nil, fmt.Errorf("yookassa: failed to parse payment status: %w", err)
	}

	var cancellationDetails *paymentprovider.CancellationDetails
	if yookassaWebhookEvent.Object.CancellationDetails != nil {
		cancellationDetails = &paymentprovider.CancellationDetails{
			Party:  yookassaWebhookEvent.Object.CancellationDetails.Party,
			Reason: yookassaWebhookEvent.Object.CancellationDetails.Reason,
		}
	}

	webhookEvent := &paymentprovider.WebhookEvent{
		Status:    status,
		PaymentID: yookassaWebhookEvent.Object.ID,
		Metadata:  metadata,
		Amount: paymentprovider.Amount{
			Value:    yookassaWebhookEvent.Object.Amount.Value,
			Currency: paymentprovider.Currency(yookassaWebhookEvent.Object.Amount.Currency),
		},
		PaymentMethod:       paymentMethod,
		CancellationDetails: cancellationDetails,
	}

	return webhookEvent, nil
}
func (p *Provider) GetName() string {
	return ProviderName
}

func (p *Provider) createConfirmation(paymentType paymentprovider.PaymentType) (yoopayment.Confirmer, error) {
	switch paymentType {
	case paymentprovider.PaymentTypeRedirect:
		return yoopayment.Redirect{
			Type:      yoopayment.TypeRedirect,
			ReturnURL: p.cfg.ReturnURL,
		}, nil
	case paymentprovider.PaymentTypeEmbedded:
		return yoopayment.Embedded{
			Type: yoopayment.TypeEmbedded,
		}, nil
	default:
		return nil, fmt.Errorf("invalid payment type: %s", paymentType)
	}
}

func (p *Provider) isSupportedWebhookEvent(event yoowebhook.WebhookEventType) bool {
	switch event {
	case yoowebhook.EventPaymentSucceeded, yoowebhook.EventPaymentWaitingForCapture, yoowebhook.EventPaymentCanceled:
		return true
	}

	return false
}

func (p *Provider) parseWebhookBody(body []byte) (yoowebhook.WebhookEvent[yoopayment.Payment], error) {
	var webhookEvent yoowebhook.WebhookEvent[yoopayment.Payment]
	if err := json.Unmarshal(body, &webhookEvent); err != nil {
		return yoowebhook.WebhookEvent[yoopayment.Payment]{}, fmt.Errorf("failed to unmarshal webhook event: %w", err)
	}

	return webhookEvent, nil
}

func (p *Provider) parseWebhookEventStatus(event yoowebhook.WebhookEventType) (paymentprovider.PaymentStatus, error) {
	switch event {
	case yoowebhook.EventPaymentSucceeded:
		return paymentprovider.PaymentStatusSucceeded, nil
	case yoowebhook.EventPaymentCanceled:
		return paymentprovider.PaymentStatusCanceled, nil
	case yoowebhook.EventPaymentWaitingForCapture:
		return paymentprovider.PaymentStatusWaitingForCapture, nil
	default:
		return "", fmt.Errorf("unsupported webhook event status: %s", event)
	}
}

func (p *Provider) parsePaymentStatus(status yoopayment.Status) (paymentprovider.PaymentStatus, error) {
	switch status {
	case yoopayment.Pending:
		return paymentprovider.PaymentStatusPending, nil
	case yoopayment.WaitingForCapture:
		return paymentprovider.PaymentStatusWaitingForCapture, nil
	case yoopayment.Succeeded:
		return paymentprovider.PaymentStatusSucceeded, nil
	case yoopayment.Canceled:
		return paymentprovider.PaymentStatusCanceled, nil
	default:
		return "", fmt.Errorf("unsupported payment status: %s", status)
	}
}

func (p *Provider) parsePaymentToken(payment *yoopayment.Payment) (string, error) {
	confirmationMap, ok := payment.Confirmation.(map[string]interface{})
	if !ok {
		return "", errors.New("unable to get token")
	}

	token, ok := confirmationMap["confirmation_token"].(string)
	if !ok {
		return "", errors.New("unable to get token")
	}

	return token, nil
}

func (p *Provider) parsePaymentMetadata(metadata any) (paymentprovider.Metadata, error) {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return paymentprovider.Metadata{}, fmt.Errorf("failed to marshal payment metadata: %w", err)
	}

	var paymentMetadata paymentprovider.Metadata
	if err := json.Unmarshal(metadataBytes, &paymentMetadata); err != nil {
		return paymentprovider.Metadata{}, fmt.Errorf("failed to unmarshal payment metadata: %w", err)
	}

	return paymentMetadata, nil
}

func (p *Provider) parsePaymentMethod(method yoopayment.PaymentMethoder) (paymentprovider.PaymentMethod, error) {
	paymentMethodBytes, err := json.Marshal(method)
	if err != nil {
		return paymentprovider.PaymentMethod{}, fmt.Errorf("failed to marshal payment method: %w", err)
	}

	var paymentMethod paymentprovider.PaymentMethod
	if err := json.Unmarshal(paymentMethodBytes, &paymentMethod); err != nil {
		return paymentprovider.PaymentMethod{}, fmt.Errorf("failed to unmarshal payment method: %w", err)
	}

	return paymentMethod, nil
}

func (p *Provider) extractConfirmationData(paymentType paymentprovider.PaymentType, payment *yoopayment.Payment) (confirmationURL string, confirmationToken string, err error) {
	switch paymentType {
	case paymentprovider.PaymentTypeRedirect:
		confirmationURL, err = p.paymentHandler.ParsePaymentLink(payment)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse payment link: %w", err)
		}
		return confirmationURL, "", nil
	case paymentprovider.PaymentTypeEmbedded:
		confirmationToken, err = p.parsePaymentToken(payment)
		if err != nil {
			return "", "", fmt.Errorf("failed to parse payment token: %w", err)
		}
		return "", confirmationToken, nil
	default:
		return "", "", fmt.Errorf("invalid payment type: %s", paymentType)
	}
}

func (p *Provider) buildPaymentResponse(paymentType paymentprovider.PaymentType, payment *yoopayment.Payment) (*paymentprovider.Payment, error) {
	var (
		confirmationURL     string
		confirmationToken   string
		cancellationDetails *paymentprovider.CancellationDetails
		err                 error
	)

	if payment.Confirmation != nil {
		confirmationURL, confirmationToken, err = p.extractConfirmationData(paymentType, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to extract confirmation data: %w", err)
		}
	}

	if payment.CancellationDetails != nil {
		cancellationDetails = &paymentprovider.CancellationDetails{
			Party:  payment.CancellationDetails.Party,
			Reason: payment.CancellationDetails.Reason,
		}
	}

	paymentStatus, err := p.parsePaymentStatus(payment.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payment status: %w", err)
	}

	metadata, err := p.parsePaymentMetadata(payment.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse payment metadata: %w", err)
	}

	return &paymentprovider.Payment{
		ID:     payment.ID,
		Status: paymentStatus,
		Amount: paymentprovider.Amount{
			Value:    payment.Amount.Value,
			Currency: paymentprovider.Currency(payment.Amount.Currency),
		},
		Description:         payment.Description,
		Metadata:            metadata,
		ConfirmationURL:     confirmationURL,
		ConfirmationToken:   confirmationToken,
		CancellationDetails: cancellationDetails,
	}, nil
}
