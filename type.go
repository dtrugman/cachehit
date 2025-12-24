package cachehit

import (
	"context"
	"errors"
)

var (
	errNotFound = errors.New("not found")
)

type Repository[K comparable, V any] interface {
	Get(ctx context.Context, key K) (V, bool)
}

type Cache[K comparable, V any] interface {
	Repository[K, V]

	Set(ctx context.Context, key K, value V) error
}

type ErrorCallback func(err error)

type syncMap interface {
	LoadOrStore(key, value any) (actual any, loaded bool)
	Delete(key any)
}
