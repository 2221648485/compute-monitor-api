package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Addr     string
	Password string
	DB       int
}

// New 创建 Redis 客户端，并通过 Ping 确认 Redis 当前可用。
func New(ctx context.Context, opts Options) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

// Close 关闭 Redis 客户端。
func Close(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}
