package tasks

import "github.com/hibiken/asynq"

type TaskFactory interface {
	Orders() OrderTask
	Close() error
}

type taskFactory struct {
	client *asynq.Client
}

func NewTaskFactory(redisAddr, redisPassword string) TaskFactory {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	return &taskFactory{
		client: client,
	}
}

func (f *taskFactory) Orders() OrderTask {
	return &orderTask{f.client}
}

func (f *taskFactory) Close() error {
	return f.client.Close()
}
