// Package clients provides API clients for external services.
package clients

import (
	"context"
	"net"
	"net/url"
	"os"
	"sync"
	"time"
)

// GPUShareClient provides access to Syphon/Spout GPU texture sharing sources
// via a local helper process REST API.
type GPUShareClient struct {
	helperURL string
}

// GPUShareStatus represents GPU texture sharing detection status.
type GPUShareStatus struct {
	Connected bool   `json:"connected"`
	HelperURL string `json:"helper_url"`
	Platform  string `json:"platform"` // "macos" (Syphon) or "windows" (Spout) or "unknown"
}

// GPUShareSource represents an active Syphon or Spout source.
type GPUShareSource struct {
	Name        string `json:"name"`
	Application string `json:"application"`
	Type        string `json:"type"` // "syphon" or "spout"
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

var (
	gpuShareClientSingleton *GPUShareClient
	gpuShareClientOnce      sync.Once
	gpuShareClientErr       error

	// TestOverrideGPUShareClient, when non-nil, is returned by GetGPUShareClient.
	TestOverrideGPUShareClient *GPUShareClient
)

// GetGPUShareClient returns the singleton GPU share client.
func GetGPUShareClient() (*GPUShareClient, error) {
	if TestOverrideGPUShareClient != nil {
		return TestOverrideGPUShareClient, nil
	}
	gpuShareClientOnce.Do(func() {
		gpuShareClientSingleton, gpuShareClientErr = NewGPUShareClient()
	})
	return gpuShareClientSingleton, gpuShareClientErr
}

// NewTestGPUShareClient creates an in-memory test client.
func NewTestGPUShareClient() *GPUShareClient {
	return &GPUShareClient{
		helperURL: "http://localhost:9876",
	}
}

// NewGPUShareClient creates a new GPU share client from environment.
func NewGPUShareClient() (*GPUShareClient, error) {
	helperURL := os.Getenv("GPUSHARE_HELPER_URL")
	if helperURL == "" {
		helperURL = "http://localhost:9876"
	}

	return &GPUShareClient{
		helperURL: helperURL,
	}, nil
}

// isReachable checks if the helper process is accepting connections.
func (c *GPUShareClient) isReachable() bool {
	u, err := url.Parse(c.helperURL)
	if err != nil {
		return false
	}
	host := u.Host
	if host == "" {
		return false
	}
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns GPU texture sharing detection status.
func (c *GPUShareClient) GetStatus(ctx context.Context) (*GPUShareStatus, error) {
	return &GPUShareStatus{
		Connected: c.isReachable(),
		HelperURL: c.helperURL,
		Platform:  "unknown",
	}, nil
}

// GetSources returns active Syphon/Spout sources.
func (c *GPUShareClient) GetSources(ctx context.Context) ([]GPUShareSource, error) {
	return []GPUShareSource{}, nil
}

// HelperURL returns the configured helper URL.
func (c *GPUShareClient) HelperURL() string {
	return c.helperURL
}
