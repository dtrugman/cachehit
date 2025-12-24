package adapter

import (
	"context"
	"testing"

	"github.com/dtrugman/cachehit/internal"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/stretchr/testify/require"
)

func TestFrom(t *testing.T) {
	cache, err := lru.New[string, int](5)
	require.NoError(t, err)

	adapter := From(cache)
	require.NotNil(t, adapter)
	require.Equal(t, cache, adapter.underlying)
}

func TestLRU_Get(t *testing.T) {
	cache, err := lru.New[string, string](10)
	require.NoError(t, err)

	adapter := From(cache)
	ctx := context.Background()

	cache.Add("key1", "value1")

	value, err := adapter.Get(ctx, "key1")
	require.NoError(t, err)
	require.Equal(t, "value1", value)

	value, err = adapter.Get(ctx, "nonexistent")
	require.ErrorIs(t, err, internal.ErrNotFound)
	require.Equal(t, "", value)
}

func TestLRU_Set(t *testing.T) {
	cache, err := lru.New[string, string](10)
	require.NoError(t, err)

	adapter := From(cache)
	ctx := context.Background()

	adapter.Set(ctx, "key1", "value1")

	value, ok := cache.Get("key1")
	require.True(t, ok)
	require.Equal(t, "value1", value)

	adapter.Set(ctx, "key1", "value2")

	value, ok = cache.Get("key1")
	require.True(t, ok)
	require.Equal(t, "value2", value)
}

func TestLRU_SetEviction(t *testing.T) {
	cache, err := lru.New[int, string](2)
	require.NoError(t, err)

	adapter := From(cache)
	ctx := context.Background()

	adapter.Set(ctx, 1, "value1")
	adapter.Set(ctx, 2, "value2")
	adapter.Set(ctx, 3, "value3")

	_, ok := cache.Get(1)
	require.False(t, ok)

	value, ok := cache.Get(2)
	require.True(t, ok)
	require.Equal(t, "value2", value)

	value, ok = cache.Get(3)
	require.True(t, ok)
	require.Equal(t, "value3", value)
}
