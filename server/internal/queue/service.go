package queue

import (
	"fmt"

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
	info, err := s.client.Enqueue(task, opts...)
	if err != nil {
		fmt.Printf("[Queue] Failed to enqueue task %s: %v\n", taskType, err)
		return nil, err
	}
	fmt.Printf("[Queue] Enqueued task: %s, ID: %s, Queue: %s\n", taskType, info.ID, info.Queue)
	return info, nil
}

func (s *queueService) Close() error {
	return s.client.Close()
}
