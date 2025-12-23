package resource

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	redis_container "github.com/testcontainers/testcontainers-go/modules/redis"
)

type RedisInstance struct {
	DSN     string
	Cleanup func() error
}

func RedisRun(ctx context.Context) (*RedisInstance, error) {
	redisVersion := "redis:7"
	redisContainer, err := redis_container.Run(ctx,
		redisVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("run container: %w", err)
	}

	dsn, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("get dsn: %w", err)
	}

	cleanup := func() error {
		return testcontainers.TerminateContainer(redisContainer)
	}

	harness := &RedisInstance{
		DSN:     dsn,
		Cleanup: cleanup,
	}
	return harness, nil
}

func RedisConn(ctx context.Context, dsn string) (*redis.Client, error) {
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	client := redis.NewClient(opt)

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("parse url: %w", err)
	}

	return client, nil
}
