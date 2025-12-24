package cachehit

import (
	"fmt"
	"time"
)

const (
	SWRDefaultRefreshWorkers    = 3
	SWRDefaultRefreshBufferSize = 256
	SWRDefaultRefreshTimeout    = 15 * time.Second
)

type swrOptions struct {
	refreshWorkers    int
	refreshBufferSize int
	refreshTimeout    time.Duration

	errorCallback ErrorCallback
}

func (o *swrOptions) Validate() error {
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

func swrDefaultOptions() *swrOptions {
	return &swrOptions{
		refreshWorkers:    SWRDefaultRefreshWorkers,
		refreshBufferSize: SWRDefaultRefreshBufferSize,
		refreshTimeout:    SWRDefaultRefreshTimeout,
	}
}

func swrCompileOptions(opts ...SWROption) *swrOptions {
	o := swrDefaultOptions()

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type SWROption func(*swrOptions)

// SWRWithRefreshWorkers configures the SWR cache to use N worker
// goroutines to refresh value asynchronously.
func SWRWithRefreshWorkers(workers int) SWROption {
	return func(o *swrOptions) {
		o.refreshWorkers = workers
	}
}

// SWRWithRefreshBufferSize configures the SWR cache to queue up to N
// async refresh requests.
func SWRWithRefreshBufferSize(size int) SWROption {
	return func(o *swrOptions) {
		o.refreshBufferSize = size
	}
}

// SWRWithRefreshTimeout configures the SWR cache to timeout async
// refresh requests after the specified amount of time.
func SWRWithRefreshTimeout(timeout time.Duration) SWROption {
	return func(o *swrOptions) {
		o.refreshTimeout = timeout
	}
}

// SWRWithErrorCallback configures the look through cache to call the
// specified callback synchronously when an error happens during internal operations.
func SWRWithErrorCallback(errorCallback ErrorCallback) SWROption {
	return func(o *swrOptions) {
		o.errorCallback = errorCallback
	}
}
