package rdbx

import "context"

type Cache interface {
	Set(ctx context.Context, key string, value interface{}) error
	Append(ctx context.Context, key string, value interface{}) (int64, error)
	Get(ctx context.Context, key string) ([]byte, error)
	GetList(ctx context.Context, key string) ([]string, error)
}
