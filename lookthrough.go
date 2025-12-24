package cachehit

import (
	"context"
	"errors"
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

func (c *LookThrough[K, V]) get(ctx context.Context, key K) (V, error) {
	k := fmt.Sprintf("%v", key)
	res, err, _ := c.dedup.Do(k, func() (interface{}, error) {
		value, err := c.repo.Get(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("repo get: %w", err)
		}

		if err := c.cache.Set(ctx, key, value); err != nil {
			c.reportError(fmt.Errorf("cache set: %v: %w", key, err))
		}
		return value, nil
	})

	if err != nil {
		var v V
		return v, err
	}

	value, ok := res.(V)
	if !ok {
		var v V
		return v, fmt.Errorf("value type: expected %T: found %T", v, res)
	}

	return value, nil
}

func (c *LookThrough[K, V]) Get(ctx context.Context, key K) (V, error) {
	value, err := c.cache.Get(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return c.get(ctx, key)
	} else if err != nil {
		c.reportError(fmt.Errorf("cache get: %v: %w", key, err))
		return c.get(ctx, key)
	}

	return value, nil
}
