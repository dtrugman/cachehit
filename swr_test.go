package cachehit

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var nilEntry *entry[string]

func isTimeoutContext(ctx context.Context) bool {
	ref, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return reflect.TypeOf(ctx) == reflect.TypeOf(ref)
}

func makeAliveEntry(value string) *entry[string] {
	now := time.Now()
	return &entry[string]{
		staleAt: now.Add(time.Hour),
		deadAt:  now.Add(2 * time.Hour),
		value:   value,
	}
}

func makeStaleEntry(value string) *entry[string] {
	now := time.Now()
	return &entry[string]{
		staleAt: now.Add(-time.Hour),
		deadAt:  now.Add(time.Hour),
		value:   value,
	}
}

func makeDeadEntry(value string) *entry[string] {
	now := time.Now()
	return &entry[string]{
		staleAt: now.Add(-2 * time.Hour),
		deadAt:  now.Add(-time.Hour),
		value:   value,
	}
}

type entryMatcher func(entry *entry[string]) bool

func getEntryMatcher(expected string, timeToStale, timeToDead time.Duration) entryMatcher {
	return func(entry *entry[string]) bool {
		now := time.Now()
		expectedStaleAt := now.Add(timeToStale).After(entry.staleAt)
		expectedDeadAt := now.Add(timeToDead).After(entry.deadAt)
		expectedValue := entry.value == expected
		return expectedStaleAt && expectedDeadAt && expectedValue
	}
}

func Test_SWR_New(t *testing.T) {
	repo := &mockRepo[string, string]{}
	cacheSize := 100
	timeToStale := time.Minute
	timeToDead := 2 * time.Minute

	swr, err := NewSWR(cacheSize, repo, timeToStale, timeToDead)
	require.NoError(t, err)
	require.NotNil(t, swr)
}

func Test_SWR_New_WithAllOptions(t *testing.T) {
	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}

	swr, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{},
		SWRWithRefreshWorkers(5),
		SWRWithRefreshBufferSize(128),
		SWRWithRefreshTimeout(30*time.Second),
	)
	require.NoError(t, err)
	require.NotNil(t, swr)
}

func Test_SWR_New_WithInvalidOptions(t *testing.T) {
	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}

	t.Run("zero refresh workers", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{}, SWRWithRefreshWorkers(0))
		require.Error(t, err)
		require.Contains(t, err.Error(), "workers count must be positive")
	})

	t.Run("negative refresh workers", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{}, SWRWithRefreshWorkers(-1))
		require.Error(t, err)
		require.Contains(t, err.Error(), "workers count must be positive")
	})

	t.Run("zero refresh buffer", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{}, SWRWithRefreshBufferSize(0))
		require.Error(t, err)
		require.Contains(t, err.Error(), "buffer size must be positive")
	})

	t.Run("negative refresh buffer", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{}, SWRWithRefreshBufferSize(-1))
		require.Error(t, err)
		require.Contains(t, err.Error(), "buffer size must be positive")
	})

	t.Run("zero refresh timeout", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{}, SWRWithRefreshTimeout(0))
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout must be positive")
	})

	t.Run("negative refresh timeout", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{}, SWRWithRefreshTimeout(-1))
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout must be positive")
	})

	t.Run("zero time to stale", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Duration(0), 2*time.Minute, &sync.Map{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "time to stale")
	})

	t.Run("negative time to stale", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Duration(-1), 2*time.Minute, &sync.Map{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "time to stale")
	})

	t.Run("zero time to dead", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Minute, time.Duration(0), &sync.Map{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "time to dead")
	})

	t.Run("negative time to dead", func(t *testing.T) {
		_, err := newSWR(repo, cache, time.Minute, time.Duration(-1), &sync.Map{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "time to dead")
	})

	t.Run("nil cache", func(t *testing.T) {
		_, err := newSWR(repo, nil, time.Minute, 2*time.Minute, &sync.Map{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "nil cache")
	})

	t.Run("nil repo", func(t *testing.T) {
		_, err := newSWR(nil, cache, time.Minute, 2*time.Minute, &sync.Map{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "nil repo")
	})
}

func Test_SWR_ValueMissing_NotInRepository(t *testing.T) {
	ctx := t.Context()

	key := "key"

	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}

	cache.On("Get", ctx, key).Return(nilEntry, false)
	repo.On("Get", ctx, key).Return("", false)

	swr, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{})
	require.NoError(t, err)

	_, found := swr.Get(ctx, key)
	require.False(t, found)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_SWR_ValueMissing_InRepository(t *testing.T) {
	ctx := t.Context()

	key := "key"
	expected := "value"

	timeToStale := time.Minute
	timeToDead := 2 * time.Minute

	entryMatcher := getEntryMatcher(expected, timeToStale, timeToDead)

	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}

	cache.On("Get", ctx, key).Return(nilEntry, false)

	repo.On("Get", ctx, key).Return(expected, true)

	cache.On("Set", ctx, key, mock.MatchedBy(entryMatcher))

	swr, err := newSWR(repo, cache, timeToStale, timeToDead, &sync.Map{})
	require.NoError(t, err)

	actual, found := swr.Get(ctx, key)
	require.True(t, found)
	require.Equal(t, expected, actual)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_SWR_ValueStale_InRepository(t *testing.T) {
	timeout := 1 * time.Second
	ctx := t.Context()

	key := "key"
	oldValue := "value"
	newValue := "new"

	timeToStale := time.Minute
	timeToDead := 2 * time.Minute

	repoGetCalled := make(chan struct{})

	staleEntry := makeStaleEntry(oldValue)

	entryMatcher := getEntryMatcher(newValue, timeToStale, timeToDead)

	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}

	cache.On("Get", ctx, key).Return(staleEntry, true)

	repo.On("Get", mock.MatchedBy(isTimeoutContext), key).
		Run(func(args mock.Arguments) {
			close(repoGetCalled)
		}).
		Return(newValue, true)

	cache.On("Set", mock.MatchedBy(isTimeoutContext), key, mock.MatchedBy(entryMatcher))

	swr, err := newSWR(repo, cache, timeToStale, timeToDead, &sync.Map{})
	require.NoError(t, err)

	actual, found := swr.Get(ctx, key)
	require.True(t, found)
	require.Equal(t, oldValue, actual)

	select {
	case <-repoGetCalled: // Background fetch completed
	case <-time.After(timeout): // Background fetch failed, expectations should fail
	}

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_SWR_ValueAlive(t *testing.T) {
	ctx := t.Context()

	key := "key"
	expected := "value"

	aliveEntry := makeAliveEntry(expected)

	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}

	cache.On("Get", ctx, key).Return(aliveEntry, true)

	swr, err := newSWR(repo, cache, time.Minute, 2*time.Minute, &sync.Map{})
	require.NoError(t, err)

	actual, found := swr.Get(ctx, key)
	require.True(t, found)
	require.Equal(t, expected, actual)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_SWR_ParallelFetchWhenMissing(t *testing.T) {
	ctx := t.Context()

	key := "key"
	expected := "value"

	timeToStale := time.Minute
	timeToDead := 2 * time.Minute

	entryMatcher := getEntryMatcher(expected, timeToStale, timeToDead)

	n := 50

	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}

	cache.On("Get", ctx, key).Return(nilEntry, false).Times(n)

	repo.On("Get", ctx, key).
		Run(func(args mock.Arguments) {
			time.Sleep(100 * time.Millisecond)
		}).
		Return(expected, true).
		Once()

	cache.On("Set", ctx, key, mock.MatchedBy(entryMatcher)).Once()

	swr, err := newSWR(repo, cache, timeToStale, timeToDead, &sync.Map{})
	require.NoError(t, err)

	ready := sync.WaitGroup{}
	ready.Add(n)

	done := sync.WaitGroup{}
	done.Add(n)

	for range n {
		go func() {
			ready.Done()
			ready.Wait()

			actual, found := swr.Get(ctx, key)
			require.True(t, found)
			require.Equal(t, expected, actual)

			done.Done()
		}()
	}

	done.Wait()

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_SWR_ParallelFetchWhenStale(t *testing.T) {
	timeout := 1 * time.Second
	ctx := t.Context()

	key := "key"
	oldValue := "stale_value"
	newValue := "fresh_value"

	timeToStale := time.Minute
	timeToDead := 2 * time.Minute

	entryMatcher := getEntryMatcher(newValue, timeToStale, timeToDead)

	repoGetCalled := make(chan struct{})

	staleEntry := makeStaleEntry(oldValue)

	n := 50

	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}

	cache.On("Get", ctx, key).Return(staleEntry, true).Times(n)

	repo.On("Get", mock.MatchedBy(isTimeoutContext), key).
		Run(func(args mock.Arguments) {
			time.Sleep(100 * time.Millisecond)
			close(repoGetCalled)
		}).
		Return(newValue, true).
		Once()

	cache.On("Set", mock.MatchedBy(isTimeoutContext), key, mock.MatchedBy(entryMatcher)).Once()

	swr, err := newSWR(repo, cache, timeToStale, timeToDead, &sync.Map{})
	require.NoError(t, err)

	ready := sync.WaitGroup{}
	ready.Add(n)

	done := sync.WaitGroup{}
	done.Add(n)

	for range n {
		go func() {
			ready.Done()
			ready.Wait()

			actual, found := swr.Get(ctx, key)
			require.True(t, found)
			require.Equal(t, oldValue, actual)

			done.Done()
		}()
	}

	done.Wait()

	select {
	case <-repoGetCalled: // Background fetch completed
	case <-time.After(timeout): // Background fetch failed, expectations should fail
	}

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_SWR_ValueDead_RefreshFromRepository(t *testing.T) {
	ctx := t.Context()

	key := "key"
	oldValue := "dead_value"
	newValue := "fresh_value"

	deadEntry := makeDeadEntry(oldValue)

	timeToStale := time.Minute
	timeToDead := 2 * time.Minute

	entryMatcher := getEntryMatcher(newValue, timeToStale, timeToDead)

	cache := &mockCache[string, *entry[string]]{}
	cache.On("Get", ctx, key).Return(deadEntry, true)
	cache.On("Set", ctx, key, mock.MatchedBy(entryMatcher))

	repo := &mockRepo[string, string]{}
	repo.On("Get", ctx, key).Return(newValue, true)

	swr, err := newSWR(repo, cache, timeToStale, timeToDead, &sync.Map{})
	require.NoError(t, err)

	actual, found := swr.Get(ctx, key)
	require.True(t, found)
	require.Equal(t, newValue, actual)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func Test_SWR_RefreshKey_GracefulHandleFullChannel(t *testing.T) {
	ctx := t.Context()

	timeToStale := time.Millisecond
	timeToDead := time.Hour

	cache := &mockCache[string, *entry[string]]{}
	repo := &mockRepo[string, string]{}
	syncMap := &mockSyncMap{}

	workerBlocked := sync.WaitGroup{}
	workerBlocked.Add(1)

	workerUnblocked := sync.WaitGroup{}
	workerUnblocked.Add(1)

	workerDone := sync.WaitGroup{}
	workerDone.Add(1)

	refreshCancelled := sync.WaitGroup{}
	refreshCancelled.Add(1)

	key1 := "key1"
	value1 := "value1"

	key2 := "key2"
	value2 := "value2"

	key3 := "key3"

	value := "old_value"
	staleEntry := makeStaleEntry(value)

	entryMatcher1 := getEntryMatcher(value1, timeToStale, timeToDead)
	entryMatcher2 := getEntryMatcher(value2, timeToStale, timeToDead)

	// The first stale key refresh request goes through and the worker
	// tries to get the value from the repository
	cache.On("Get", ctx, key1).Return(staleEntry, true).Once()
	syncMap.On("LoadOrStore", key1, struct{}{}).Return("", false).Once()
	repo.On("Get", mock.MatchedBy(isTimeoutContext), key1).
		Run(func(args mock.Arguments) {
			workerBlocked.Done()
			workerUnblocked.Wait()
		}).
		Return(value1, true).Once()

	// Once the fetch is finished, the value is set and the key is
	// deleted from the map
	cache.On("Set", mock.MatchedBy(isTimeoutContext), key1, mock.MatchedBy(entryMatcher1)).
		Run(func(args mock.Arguments) {
			workerDone.Done()
		}).Once()
	syncMap.On("Delete", key1).Once()

	// The second stale key refresh request shoud be pending in the channel,
	// that is now empty because the first request is being processed by
	// the worker
	cache.On("Get", ctx, "key2").Return(staleEntry, true).Once()
	syncMap.On("LoadOrStore", key2, struct{}{}).Return("", false).Once()

	// The third stale key refresh request should fail gracefully,
	// as the channel is full
	cache.On("Get", ctx, "key3").Return(staleEntry, true).Once()
	syncMap.On("LoadOrStore", key3, struct{}{}).Return("", false).Once()
	syncMap.On("Delete", key3).
		Run(func(args mock.Arguments) {
			refreshCancelled.Done()
		}).Once()

	// The refresh flow for key2 may or may not complete
	repo.On("Get", mock.MatchedBy(isTimeoutContext), key2).
		Return(value2, true).Maybe()
	cache.On("Set", mock.MatchedBy(isTimeoutContext), key2, mock.MatchedBy(entryMatcher2)).Maybe()
	syncMap.On("Delete", key2).Maybe()

	swr, err := newSWR(repo, cache, timeToStale, timeToDead, syncMap,
		SWRWithRefreshBufferSize(1),
		SWRWithRefreshWorkers(1),
	)
	require.NoError(t, err)

	swr.Get(ctx, key1)
	workerBlocked.Wait()

	swr.Get(ctx, key2)
	swr.Get(ctx, key3)
	refreshCancelled.Wait()

	workerUnblocked.Done()
	workerDone.Wait()

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}
