package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

type NotificationTask interface {
	SendEmailConfirmationCode(ctx context.Context, email string, code string) error
}

type notificationTask struct {
	client *asynq.Client
}

func (n *notificationTask) SendEmailConfirmationCode(ctx context.Context, email string, code string) error {
	task, err := NewSendEmailConfirmationCodeTask(email, code)
	if err != nil {
		return err
	}

	_, err = n.client.EnqueueContext(ctx, task)
	return err
}
