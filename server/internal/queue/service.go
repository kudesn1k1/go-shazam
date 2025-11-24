package queue

import (
	"github.com/hibiken/asynq"
)

type QueueService interface {
	Enqueue(taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error)
	Close() error
}

type queueService struct {
	client *asynq.Client
}

func NewQueueService(cfg *Config) QueueService {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &queueService{
		client: client,
	}
}

func (s *queueService) Enqueue(taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	task := asynq.NewTask(taskType, payload)
	return s.client.Enqueue(task, opts...)
}

func (s *queueService) Close() error {
	return s.client.Close()
}
