package cachehit

type lookThroughOptions struct {
	errorCallback ErrorCallback
}

func (o *lookThroughOptions) Validate() error {
	return nil
}

func lookThroughDefaultOptions() *lookThroughOptions {
	return &lookThroughOptions{}
}

func lookThroughCompileOptions(opts ...LookThroughOption) *lookThroughOptions {
	o := lookThroughDefaultOptions()

	for _, opt := range opts {
		opt(o)
	}

	return o
}

type LookThroughOption func(*lookThroughOptions)

// LookThroughWithErrorCallback configures the look through cache to call the
// specified callback synchronously when an error happens during internal operations.
func LookThroughWithErrorCallback(errorCallback ErrorCallback) LookThroughOption {
	return func(o *lookThroughOptions) {
		o.errorCallback = errorCallback
	}
}
