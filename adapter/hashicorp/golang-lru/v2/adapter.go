package adapter

import (
	"context"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/dtrugman/cachehit/internal"
)

type LRU[K comparable, V any] struct {
	underlying *lru.Cache[K, V]
}

func From[K comparable, V any](underlying *lru.Cache[K, V]) *LRU[K, V] {
	return &LRU[K, V]{underlying: underlying}
}

func (a *LRU[K, V]) Get(_ context.Context, key K) (V, error) {
	if v, ok := a.underlying.Get(key); !ok {
		return v, internal.ErrNotFound
	} else {
		return v, nil
	}
}

func (a *LRU[K, V]) Set(_ context.Context, key K, value V) error {
	// Discard the eviction bool, use the callbacks if needed
	_ = a.underlying.Add(key, value)
	return nil
}
