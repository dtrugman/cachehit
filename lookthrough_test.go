package cachehit

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_LookThrough_ValueMissing(t *testing.T) {
	ctx := t.Context()

	key := "key"

	cache := &mockCache[string, string]{}
	cache.On("Get", ctx, key).Return("", false)

	repo := &mockCache[string, string]{}
	repo.On("Get", ctx, key).Return("", false)

	lt := NewLookThrough(cache, repo)
	_, found := lt.Get(ctx, key)
	require.False(t, found)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_LookThrough_ValueInRepository(t *testing.T) {
	ctx := t.Context()

	key := "key"
	expected := "value"

	cache := &mockCache[string, string]{}
	cache.On("Get", ctx, key).Return("", false)
	cache.On("Set", ctx, key, expected)

	repo := &mockCache[string, string]{}
	repo.On("Get", ctx, key).Return(expected, true)

	lt := NewLookThrough(cache, repo)
	actual, found := lt.Get(ctx, key)
	require.True(t, found)
	require.Equal(t, expected, actual)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_LookThrough_ValueInCache(t *testing.T) {
	ctx := t.Context()

	key := "key"
	expected := "value"

	cache := &mockCache[string, string]{}
	cache.On("Get", ctx, key).Return(expected, true)

	repo := &mockCache[string, string]{}

	lt := NewLookThrough(cache, repo)
	actual, found := lt.Get(ctx, key)
	require.True(t, found)
	require.Equal(t, expected, actual)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_LookThrough_ParallelFetch(t *testing.T) {
	ctx := t.Context()

	key := "key"
	expected := "value"

	n := 50

	cache := &mockCache[string, string]{}
	cache.On("Get", ctx, key).Return("", false).Times(n)
	cache.On("Set", ctx, key, expected).Once()

	repo := &mockCache[string, string]{}
	repo.On("Get", ctx, key).
		Run(func(args mock.Arguments) {
			time.Sleep(100 * time.Millisecond)
		}).
		Return(expected, true).
		Once()

	lt := NewLookThrough(cache, repo)

	ready := sync.WaitGroup{}
	ready.Add(n)

	done := sync.WaitGroup{}
	done.Add(n)

	for range n {
		go func() {
			ready.Done()
			ready.Wait()

			actual, found := lt.Get(ctx, key)
			require.True(t, found)
			require.Equal(t, expected, actual)

			done.Done()
		}()
	}

	done.Wait()

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
