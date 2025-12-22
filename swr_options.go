package cachehit

import (
	"fmt"
	"time"
)

const (
	DefaultRefreshWorkers    = 3
	DefaultRefreshBufferSize = 256
	DefaultRefreshTimeout    = 15 * time.Second
)

type options struct {
	refreshWorkers    int
	refreshBufferSize int
	refreshTimeout    time.Duration
}

func (o *options) Validate() error {
	if o.refreshWorkers <= 0 {
		return fmt.Errorf("workers count must be positive")
	}

	if o.refreshBufferSize <= 0 {
		return fmt.Errorf("buffer size must be positive")
	}

	if o.refreshTimeout <= time.Duration(0) {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

func defaultOptions() *options {
	return &options{
		refreshWorkers:    DefaultRefreshWorkers,
		refreshBufferSize: DefaultRefreshBufferSize,
		refreshTimeout:    DefaultRefreshTimeout,
	}
}

func compileOptions(opts ...Option) *options {
	o := defaultOptions()

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type Option func(*options)

// WithRefreshWorkers configures the SWR cache to use N worker
// goroutines to refresh value asynchronously.
func WithRefreshWorkers(workers int) Option {
	return func(o *options) {
		o.refreshWorkers = workers
	}
}

// WithRefreshWorkers configures the SWR cache to queue up to N
// async refresh requests.
func WithRefreshBufferSize(size int) Option {
	return func(o *options) {
		o.refreshBufferSize = size
	}
}

// WithRefreshTimeout configures the SWR cache to timeout async
// refresh requests after the specified amount of time.
func WithRefreshTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.refreshTimeout = timeout
	}
}
