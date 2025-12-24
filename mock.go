package cachehit

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockCache[K comparable, V any] struct {
	mock.Mock
}

func (m *mockCache[K, V]) Get(ctx context.Context, key K) (V, bool) {
	args := m.Called(ctx, key)
	return args.Get(0).(V), args.Bool(1)
}

func (m *mockCache[K, V]) Set(ctx context.Context, key K, value V) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

type mockRepo[K comparable, V any] struct {
	mock.Mock
}

func (m *mockRepo[K, V]) Get(ctx context.Context, key K) (V, bool) {
	args := m.Called(ctx, key)
	return args.Get(0).(V), args.Bool(1)
}

type mockSyncMap struct {
	mock.Mock
}

func (m *mockSyncMap) LoadOrStore(key, value any) (actual any, loaded bool) {
	args := m.Called(key, value)
	return args.Get(0), args.Bool(1)
}

func (m *mockSyncMap) Delete(key any) {
	m.Called(key)
}
