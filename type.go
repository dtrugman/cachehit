package cachehit

import (
	"context"

	"github.com/dtrugman/cachehit/internal"
)

var (
	ErrNotFound = internal.ErrNotFound
)

type Repository[K comparable, V any] interface {
	Get(ctx context.Context, key K) (V, error)
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
