package adapters

import (
	"net/http"
	"time"
)

// HTTPClient abstracts HTTP operations for testing.
// This interface is satisfied by *http.Client and can be mocked in tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient returns a configured HTTP client for production use.
func DefaultHTTPClient() HTTPClient {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// adapterOptions holds optional dependencies for adapters.
type adapterOptions struct {
	httpClient HTTPClient
	getEnv     func(string) string
}

// AdapterOption configures optional adapter dependencies.
type AdapterOption func(*adapterOptions)

// WithHTTPClient sets a custom HTTP client for the adapter.
// Use this in tests to inject a mock HTTP client.
func WithHTTPClient(client HTTPClient) AdapterOption {
	return func(o *adapterOptions) {
		o.httpClient = client
	}
}

// WithEnvGetter sets a custom environment variable getter.
// Use this in tests to avoid depending on actual environment variables.
func WithEnvGetter(fn func(string) string) AdapterOption {
	return func(o *adapterOptions) {
		o.getEnv = fn
	}
}

// defaultOptions returns adapter options with production defaults.
func defaultOptions() *adapterOptions {
	return &adapterOptions{
		httpClient: DefaultHTTPClient(),
		getEnv:     defaultGetEnv,
	}
}

// applyOptions applies option functions to the options struct.
func applyOptions(opts []AdapterOption) *adapterOptions {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}
