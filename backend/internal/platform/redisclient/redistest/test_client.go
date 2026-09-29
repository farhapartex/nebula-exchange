package redistest

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"

	"nebula-exchange/backend/internal/platform/redisclient"
)

const (
	redisURLEnvironmentKey     = "TEST_REDIS_URL"
	requireRedisEnvironmentKey = "REQUIRE_DATABASE_TESTS"
	defaultTestRedisURL        = "redis://localhost:6379/15"
)

func NewClient(t *testing.T) *redis.Client {
	t.Helper()
	redisURL := os.Getenv(redisURLEnvironmentKey)
	if redisURL == "" {
		redisURL = defaultTestRedisURL
	}

	client, err := redisclient.New(context.Background(), redisURL)
	if err != nil {
		if os.Getenv(requireRedisEnvironmentKey) == "true" {
			t.Fatalf("connect to test redis: %v", err)
		}
		t.Skipf("skipping redis test, start redis with make docker-up: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}
