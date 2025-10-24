package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zuxt268/sales/internal/config"
)

func NewRedisQueue() *redis.Client {
	addr := fmt.Sprintf("%s:%d", config.Env.RedisHost, config.Env.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// 接続確認
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("Failed to connect to Redis", "error", err)
		panic("failed to connect to Redis")
	}

	slog.Info("Redis connection established",
		"host", config.Env.RedisHost,
		"port", config.Env.RedisPort,
	)

	return client
}
