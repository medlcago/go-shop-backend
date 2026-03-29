package tasks

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type TaskFactory interface {
	Orders() OrderTask
}

type taskFactory struct {
	client *asynq.Client
}

func NewTaskFactory(rdb redis.UniversalClient) TaskFactory {
	client := asynq.NewClientFromRedisClient(rdb)

	return &taskFactory{
		client: client,
	}
}

func (f *taskFactory) Orders() OrderTask {
	return &orderTask{f.client}
}
