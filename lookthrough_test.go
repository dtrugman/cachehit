package cachehit

import (
	"errors"
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
	repo := &mockCache[string, string]{}

	cache.On("Get", ctx, key).Return("", ErrNotFound)
	repo.On("Get", ctx, key).Return("", ErrNotFound)

	lt, err := NewLookThrough(cache, repo)
	require.NoError(t, err)

	_, err = lt.Get(ctx, key)
	require.ErrorIs(t, err, ErrNotFound)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_LookThrough_ValueInRepository(t *testing.T) {
	ctx := t.Context()

	key := "key"
	expected := "value"

	cache := &mockCache[string, string]{}
	repo := &mockCache[string, string]{}

	cache.On("Get", ctx, key).Return("", ErrNotFound)
	repo.On("Get", ctx, key).Return(expected, nil)
	cache.On("Set", ctx, key, expected).Return(nil)

	lt, err := NewLookThrough(cache, repo)
	require.NoError(t, err)

	actual, err := lt.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_LookThrough_ValueInCache(t *testing.T) {
	ctx := t.Context()

	key := "key"
	expected := "value"

	cache := &mockCache[string, string]{}
	repo := &mockCache[string, string]{}

	cache.On("Get", ctx, key).Return(expected, nil)

	lt, err := NewLookThrough(cache, repo)
	require.NoError(t, err)

	actual, err := lt.Get(ctx, key)
	require.NoError(t, err)
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
	repo := &mockCache[string, string]{}

	cache.On("Get", ctx, key).Return("", ErrNotFound).Times(n)
	repo.On("Get", ctx, key).
		Run(func(args mock.Arguments) {
			time.Sleep(100 * time.Millisecond)
		}).
		Return(expected, nil).
		Once()
	cache.On("Set", ctx, key, expected).Return(nil).Once()

	lt, err := NewLookThrough(cache, repo)
	require.NoError(t, err)

	ready := sync.WaitGroup{}
	ready.Add(n)

	done := sync.WaitGroup{}
	done.Add(n)

	for range n {
		go func() {
			ready.Done()
			ready.Wait()

			actual, err := lt.Get(ctx, key)
			require.NoError(t, err)
			require.Equal(t, expected, actual)

			done.Done()
		}()
	}

	done.Wait()

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_LookThrough_ErrorCallbackCalled_CacheSet(t *testing.T) {
	ctx := t.Context()

	key := "key"
	value := "value"
	cacheSetErr := errors.New("failed")

	cache := &mockCache[string, string]{}
	repo := &mockCache[string, string]{}

	cache.On("Get", ctx, key).Return("", ErrNotFound)
	repo.On("Get", ctx, key).Return(value, nil)
	cache.On("Set", ctx, key, value).Return(cacheSetErr)

	var capturedErr error
	errorCallback := func(err error) {
		capturedErr = err
	}

	lt, err := NewLookThrough(cache, repo, LookThroughWithErrorCallback(errorCallback))
	require.NoError(t, err)

	actual, err := lt.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, actual)

	require.NotNil(t, capturedErr)
	require.ErrorIs(t, capturedErr, cacheSetErr)
	require.ErrorContains(t, capturedErr, key)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_LookThrough_ErrorCallbackCalled_CacheGet(t *testing.T) {
	ctx := t.Context()

	key := "key"
	value := "value"
	cacheGetErr := errors.New("failed")

	cache := &mockCache[string, string]{}
	repo := &mockCache[string, string]{}

	cache.On("Get", ctx, key).Return("", cacheGetErr)
	repo.On("Get", ctx, key).Return(value, nil)
	cache.On("Set", ctx, key, value).Return(nil)

	var capturedErr error
	errorCallback := func(err error) {
		capturedErr = err
	}

	lt, err := NewLookThrough(cache, repo, LookThroughWithErrorCallback(errorCallback))
	require.NoError(t, err)

	actual, err := lt.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, actual)

	require.NotNil(t, capturedErr)
	require.ErrorIs(t, capturedErr, cacheGetErr)
	require.ErrorContains(t, capturedErr, key)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
