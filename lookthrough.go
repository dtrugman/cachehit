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

	errorCallback ErrorCallback
}

func NewLookThrough[K comparable, V any](
	cache Cache[K, V],
	repo Repository[K, V],
	opts ...LookThroughOption,
) (*LookThrough[K, V], error) {
	dedup := new(singleflight.Group)

	o := lookThroughCompileOptions(opts...)
	if err := o.Validate(); err != nil {
		return nil, fmt.Errorf("options: %w", err)
	}

	return &LookThrough[K, V]{
		cache:         cache,
		repo:          repo,
		dedup:         dedup,
		errorCallback: o.errorCallback,
	}, nil
}

func (c *LookThrough[K, V]) reportError(err error) {
	if c.errorCallback != nil {
		c.errorCallback(err)
	}
}

func (c *LookThrough[K, V]) get(ctx context.Context, key K) (V, bool) {
	k := fmt.Sprintf("%v", key)
	res, err, _ := c.dedup.Do(k, func() (interface{}, error) {
		value, found := c.repo.Get(ctx, key)
		if !found {
			return nil, errNotFound
		}

		if err := c.cache.Set(ctx, key, value); err != nil {
			c.reportError(fmt.Errorf("cache set: %v: %w", key, err))
		}
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
