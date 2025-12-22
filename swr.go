package cachehit

import (
	"context"
	"fmt"
	"sync"
	"time"

	lru_adapter "github.com/dtrugman/cachehit/adapter/hashicorp/golang-lru/v2"
	lru "github.com/hashicorp/golang-lru/v2"

	"golang.org/x/sync/singleflight"
)

type entry[V any] struct {
	staleAt time.Time
	deadAt  time.Time
	value   V
}

type SWR[K comparable, V any] struct {
	cache Cache[K, *entry[V]]
	repo  Repository[K, V]

	timeToStale time.Duration
	timeToDead  time.Duration

	dedup *singleflight.Group

	refreshChan    chan K
	refreshTimeout time.Duration
	refreshKeys    syncMap
}

func newSWR[K comparable, V any](
	repo Repository[K, V],
	cache Cache[K, *entry[V]],
	timeToStale time.Duration,
	timeToDead time.Duration,
	syncMap syncMap,
	opts ...Option,
) (*SWR[K, V], error) {
	if cache == nil {
		return nil, fmt.Errorf("nil cache")
	}

	if repo == nil {
		return nil, fmt.Errorf("nil repo")
	}

	if timeToStale <= time.Duration(0) {
		return nil, fmt.Errorf("time to stale must be positive")
	}

	if timeToDead <= time.Duration(0) {
		return nil, fmt.Errorf("time to dead must be positive")
	}

	o := compileOptions(opts...)
	if err := o.Validate(); err != nil {
		return nil, fmt.Errorf("options: %w", err)
	}

	dedup := new(singleflight.Group)

	refreshChan := make(chan K, o.refreshBufferSize)

	swr := &SWR[K, V]{
		cache: cache,
		repo:  repo,

		timeToStale: timeToStale,
		timeToDead:  timeToDead,

		dedup: dedup,

		refreshChan:    refreshChan,
		refreshTimeout: o.refreshTimeout,
		refreshKeys:    syncMap,
	}

	for range o.refreshWorkers {
		go swr.refreshWorker()
	}

	return swr, nil
}

func NewSWR[K comparable, V any](
	cacheSize int,
	repo Repository[K, V],
	timeToStale time.Duration,
	timeToDead time.Duration,
	opts ...Option,
) (*SWR[K, V], error) {
	cache, err := lru.New[K, *entry[V]](cacheSize)
	if err != nil {
		return nil, fmt.Errorf("cache: %w", err)
	}
	adapter := lru_adapter.From(cache)

	syncMap := &sync.Map{}

	return newSWR(repo, adapter, timeToStale, timeToDead, syncMap, opts...)
}

func (c *SWR[K, V]) refreshWorker() {
	for key := range c.refreshChan {
		ctx, cancel := context.WithTimeout(context.Background(), c.refreshTimeout)
		_, _ = c.get(ctx, key)
		cancel()

		c.refreshKeys.Delete(key)
	}
}

func (c *SWR[K, V]) refreshKey(key K) {
	if _, exists := c.refreshKeys.LoadOrStore(key, struct{}{}); exists {
		return
	}

	select {
	case c.refreshChan <- key:
		// Successfully queued

	default:
		// Channel full, handle gracefully
		c.refreshKeys.Delete(key)
	}
}

func (c *SWR[K, V]) get(ctx context.Context, key K) (V, bool) {
	k := fmt.Sprintf("%v", key)
	res, err, _ := c.dedup.Do(k, func() (interface{}, error) {
		value, found := c.repo.Get(ctx, key)
		if !found {
			return nil, errNotFound
		}

		now := time.Now()
		staleAt := now.Add(c.timeToStale)
		deadAt := now.Add(c.timeToDead)

		entry := &entry[V]{
			staleAt: staleAt,
			deadAt:  deadAt,
			value:   value,
		}

		c.cache.Set(ctx, key, entry)
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

func (c *SWR[K, V]) Get(ctx context.Context, key K) (V, bool) {
	entry, ok := c.cache.Get(ctx, key)
	if !ok {
		return c.get(ctx, key)
	}

	now := time.Now()
	if now.Before(entry.staleAt) {
		return entry.value, true
	} else if now.Before(entry.deadAt) {
		c.refreshKey(key)
		return entry.value, true
	} else {
		return c.get(ctx, key)
	}
}
