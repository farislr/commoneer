package rdbx

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type cache struct {
	redis redis.UniversalClient
}

func (c *cache) Set(ctx context.Context, key string, value interface{}) error {
	return c.redis.Set(ctx, key, value, 1800*time.Second).Err()
}

func (c *cache) Get(ctx context.Context, key string) ([]byte, error) {
	return c.redis.Get(ctx, key).Bytes()
}

func (c *cache) Append(ctx context.Context, key string, value interface{}) (int64, error) {
	return c.redis.LPush(ctx, key, value).Result()
}

func (c *cache) GetList(ctx context.Context, key string) ([]string, error) {
	return c.redis.LRange(ctx, key, 0, -1).Result() // -1 means all
}
