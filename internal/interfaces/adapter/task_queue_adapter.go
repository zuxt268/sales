package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/zuxt268/sales/internal/domain"
)

type TaskQueueAdapter interface {
	Enqueue(ctx context.Context, task domain.Task) error
}

type taskQueueAdapter struct {
	redisClient *redis.Client
}

func NewTaskQueueAdapter(
	redisClient *redis.Client,
) TaskQueueAdapter {
	return &taskQueueAdapter{
		redisClient: redisClient,
	}
}

const queueNameTask = "task"

func (a *taskQueueAdapter) Enqueue(ctx context.Context, task domain.Task) error {
	jsonData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = a.redisClient.RPush(ctx, queueNameTask, jsonData).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue: %w", err)
	}

	return nil
}
