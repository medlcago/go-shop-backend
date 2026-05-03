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
	Email string `json:"to"`
	Code  string `json:"code"`
}

func NewCancelOrderTask(userID, orderID uuid.UUID, delay time.Duration) (*asynq.Task, error) {
	payload, err := json.Marshal(CancelOrderPayload{
		UserID:  userID,
		OrderID: orderID,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeCancelOrder, payload, asynq.ProcessIn(delay)), nil
}

func NewSendEmailConfirmationCodeTask(email string, code string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendEmailConfirmationCodePayload{
		Email: email,
		Code:  code,
	})

	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		TypeSendEmailConfirmationCode,
		payload,
		asynq.MaxRetry(3),
	), nil
}
