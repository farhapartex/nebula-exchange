package redisclient

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const connectTimeout = 5 * time.Second

func New(ctx context.Context, redisURL string) (*redis.Client, error) {
	connectionOptions, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(connectionOptions)

	pingContext, cancelPing := context.WithTimeout(ctx, connectTimeout)
	defer cancelPing()
	if err := client.Ping(pingContext).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
