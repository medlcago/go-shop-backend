package tasks

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Factory struct {
	client *asynq.Client
}

func NewFactory(rdb redis.UniversalClient) *Factory {
	client := asynq.NewClientFromRedisClient(rdb)

	return &Factory{
		client: client,
	}
}

func (f *Factory) Orders() OrderTask {
	return &orderTask{f.client}
}

func (f *Factory) Notifications() NotificationTask {
	return &notificationTask{f.client}
}
