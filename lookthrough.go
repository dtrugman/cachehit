package cachehit

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"
)

type LookThrough[K comparable, V any] struct {
	cache Cache[K, V]
	repo  Repository[K, V]

	dedup *singleflight.Group
}

func NewLookThrough[K comparable, V any](
	cache Cache[K, V],
	repo Repository[K, V],
) *LookThrough[K, V] {
	dedup := new(singleflight.Group)

	return &LookThrough[K, V]{
		cache: cache,
		repo:  repo,
		dedup: dedup,
	}
}

func (c *LookThrough[K, V]) get(ctx context.Context, key K) (V, bool) {
	k := fmt.Sprintf("%v", key)
	res, err, _ := c.dedup.Do(k, func() (interface{}, error) {
		value, found := c.repo.Get(ctx, key)
		if !found {
			return nil, errNotFound
		}

		c.cache.Set(ctx, key, value)
		return value, nil
	})

	if err != nil {
		var v V
		return v, false
	}

	value, ok := res.(V)
	if !ok {
		var v V
		return v, false
	}

	return value, true
}

func (c *LookThrough[K, V]) Get(ctx context.Context, key K) (V, bool) {
	value, found := c.cache.Get(ctx, key)
	if found {
		return value, true
	}

	return c.get(ctx, key)
}
