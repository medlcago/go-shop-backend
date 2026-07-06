package tasks

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeCancelOrder               = "cancel:order"
	TypeSendEmailConfirmationCode = "send:email_confirmation_code"
)

type CancelOrderPayload struct {
	UserID  uuid.UUID `json:"user_id"`
	OrderID uuid.UUID `json:"order_id"`
}

type SendEmailConfirmationCodePayload struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func NewCancelOrderTask(payload CancelOrderPayload, delay time.Duration) (*asynq.Task, error) {
	data, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeCancelOrder, data, asynq.ProcessIn(delay)), nil
}

func NewSendEmailConfirmationCodeTask(payload SendEmailConfirmationCodePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		TypeSendEmailConfirmationCode,
		data,
		asynq.MaxRetry(3),
	), nil
}
