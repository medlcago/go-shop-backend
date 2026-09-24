package tasks

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

type NotificationTask interface {
	SendEmailConfirmationCode(ctx context.Context, payload SendEmailConfirmationCodePayload) error
}

type SendEmailConfirmationCodePayload struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type notificationTask struct {
	client *asynq.Client
}

func (n *notificationTask) SendEmailConfirmationCode(ctx context.Context, payload SendEmailConfirmationCodePayload) error {
	task, err := newSendEmailConfirmationCodeTask(payload)
	if err != nil {
		return err
	}

	_, err = n.client.EnqueueContext(ctx, task)
	return err
}

func newSendEmailConfirmationCodeTask(payload SendEmailConfirmationCodePayload) (*asynq.Task, error) {
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
