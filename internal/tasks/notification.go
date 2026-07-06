package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

type NotificationTask interface {
	SendEmailConfirmationCode(ctx context.Context, payload SendEmailConfirmationCodePayload) error
}

type notificationTask struct {
	client *asynq.Client
}

func (n *notificationTask) SendEmailConfirmationCode(ctx context.Context, payload SendEmailConfirmationCodePayload) error {
	task, err := NewSendEmailConfirmationCodeTask(payload)
	if err != nil {
		return err
	}

	_, err = n.client.EnqueueContext(ctx, task)
	return err
}
