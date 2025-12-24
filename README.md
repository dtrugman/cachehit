# cachehit

![Tests](https://github.com/dtrugman/cachehit/actions/workflows/merge.yml/badge.svg)
[![codecov](https://codecov.io/gh/dtrugman/cachehit/branch/main/graph/badge.svg)](https://codecov.io/gh/dtrugman/cachehit)

A Go library providing high-performance caching constructs with stale-while-revalidate semantics and look-through patterns.

## Cache Constructs

### Stale-While-Revalidate (SWR) Cache

The SWR cache implements a stale-while-revalidate strategy with three distinct data states:

- **Fresh** (< `timeToStale`): Data is served instantly from memory
- **Stale** (`timeToStale` - `timeToDead`): Data is served from memory while being asynchronously refreshed in the background
- **Dead** (> `timeToDead`): Data is synchronously fetched from the underlying repository

#### Benefits

- **Exceptionally low latency**: Requests are served from memory whenever possible, even when data needs refreshing
- **Graceful data refresh**: Stale data is returned immediately while fresh data is fetched in the background, eliminating refresh-induced latency spikes
- **No thundering herd**: Built-in request deduplication prevents multiple concurrent fetches for the same key
- **Increased stability**: If the data source is temporarily unavailable, async refreshes are retried on subsequent requests, keeping stale data available and improving consistency

#### How It Works

The SWR cache uses an LRU (Least Recently Used) eviction policy to manage memory efficiently.
It maintains entries with two timestamps: `staleAt` and `deadAt`.
When you request data:

1. If cached and fresh, returns immediately from memory
2. If cached but stale, returns from memory and queues a background refresh
3. If cached but dead or not cached, fetches synchronously from the repository

Background refreshes are handled by configurable worker goroutines that process keys needing updates.

The cache uses deduplication logic to prevent concurrent requests for the same key, for both sync and async fetches.

#### Options

```go
func NewSWR[K comparable, V any](
    cacheSize int,
    repo Repository[K, V],
    timeToStale time.Duration,
    timeToDead time.Duration,
    opts ...SWROption,
) (*SWR[K, V], error)
```

- `cacheSize`: Maximum number of entries in the LRU cache
- `repo`: Data source implementing `Repository[K, V]` interface
- `timeToStale`: Duration after which data becomes stale
- `timeToDead`: Duration after which data is considered dead

Available options:
- `SWRWithRefreshWorkers(n int)`: Number of background workers for async refreshes (default: 1)
- `SWRWithRefreshBufferSize(size int)`: Channel buffer size for refresh queue (default: 100)
- `SWRWithRefreshTimeout(timeout time.Duration)`: Timeout for background refresh operations (default: 30s)

#### Usage

```go
// repo is any object that implements
// Get(ctx context.Context, key K) (V, bool)
repo := ...

cache, err := cachehit.NewSWR(128, repo, 5*time.Minute, 15*time.Minute)
if err != nil {
    // handle error
}

user, ok := cache.Get(ctx, userID)
if !ok {
    // handle miss
}
```

### LookThrough Cache

A classic caching pattern that automatically populates the cache on misses from the repository (the next layer).

#### How It Works

When you request data:

1. If found in cache, returns immediately
2. If not found, fetches from repository, stores in cache, and returns

The cache uses deduplication logic to prevent concurrent requests for the same key.

LookThrough caches can be chained together to create complex multi-layer fetch patterns.
Since the `Repository` interface is generic, one LookThrough cache can use another LookThrough cache (or any cache construct) as its repository, enabling architectures like: in-memory cache → distributed cache → HTTP API.

#### Options

```go
func NewLookThrough[K comparable, V any](
    cache Cache[K, V],
    repo Repository[K, V],
) *LookThrough[K, V]
```

- `cache`: Any cache implementing the `Cache[K, V]` interface
- `repo`: Data source implementing the `Repository[K, V]` interface

#### Usage

```go
// repo is any object that implements
// Get(ctx context.Context, key K) (V, bool)
repo := ...

// Use Hashicorp's LRU for example with a thin adapter
lru := lru.New[string, Config](1024)
lruCache := adaper.From(lru)

lookThrough := cachehit.NewLookThrough(lruCache, repo)

product, ok := lookThrough.Get(ctx, productID)
if !ok {
    // handle miss
}
```

## Examples

### SWR with a data repository

[`example/swr/main.go`](example/swr/main.go) demonstrates a two-layer architecture:

- **Layer 1**: SWR cache with 128-entry LRU, 10s stale time, 30s dead time
- **Layer 2**: Redis as the data repository

(Time to stale/dead set to a low value as an example. Real values should depend on your use case)

This setup provides ultra-fast in-memory access for fresh data,
serves stale data instantly while refreshing in the background,
and falls back to the data repository for cache misses or dead entries.
Perfect for scenarios where you want to minimize data fetches while keeping data fresh.

### Three-Layer Architecture

[`example/layered/main.go`](example/layered/main.go) demonstrates a sophisticated three-layer caching strategy for GitHub user data:

- **Layer 1**: SWR cache (in-memory, 128 entries, 10s stale, 30s dead)
- **Layer 2**: LookThrough cache with Redis (1m expiration)
- **Layer 3**: GitHub API (source of truth)

(Time to stale/dead set to a low value as an example. Real values should depend on your use case)

The flow works as follows:
1. SWR cache serves fresh/stale data from memory instantly
2. On SWR miss or dead data, LookThrough checks Redis
3. On Redis miss, GitHub API is called and result propagates back through both caches

This architecture minimizes API calls to rate-limited services,
provides sub-millisecond response times for cached data,
and gracefully handles multiple levels of cache invalidation.

## Installation

```bash
go get github.com/dtrugman/cachehit
```

## Interfaces

Both cache constructs work with generic interfaces:

```go
type Cache[K comparable, V any] interface {
    Get(ctx context.Context, key K) (V, bool)
    Set(ctx context.Context, key K, value V)
}

type Repository[K comparable, V any] interface {
    Get(ctx context.Context, key K) (V, bool)
}
```

Some basic adapters are provided for popular cache backends in the `adapter/` directory.
