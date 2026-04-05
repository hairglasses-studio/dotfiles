// Package httpclient re-exports mcpkit/client HTTP pool functions.
package httpclient

import (
	"net/http"
	"time"

	"github.com/hairglasses-studio/mcpkit/client"
)

// Fast returns an HTTP client with 5s timeout for LAN/local calls.
func Fast() *http.Client { return client.Fast() }

// Standard returns an HTTP client with 30s timeout for cloud APIs.
func Standard() *http.Client { return client.Standard() }

// Slow returns an HTTP client with 2m timeout for uploads.
func Slow() *http.Client { return client.Slow() }

// WithTimeout returns an HTTP client with a custom timeout.
func WithTimeout(d time.Duration) *http.Client { return client.WithTimeout(d) }

// Transport returns the shared connection-pooling transport.
func Transport() http.RoundTripper { return client.Standard().Transport }
