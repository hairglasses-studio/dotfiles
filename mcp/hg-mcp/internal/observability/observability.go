// Package observability wraps mcpkit/observability with a global provider
// so that registry.go and other callers can use package-level functions.
package observability

import (
	"context"
	"sync"
	"time"

	mcpobs "github.com/hairglasses-studio/mcpkit/observability"
	"go.opentelemetry.io/otel/trace"
)

var (
	provider *mcpobs.Provider
	mu       sync.RWMutex
)

// Config re-exports mcpkit's observability config type.
type Config = mcpobs.Config

// DefaultConfig returns a sensible default config for hg-mcp.
func DefaultConfig() Config {
	return Config{
		ServiceName:    "hg-mcp",
		ServiceVersion: "0.1.0",
		OTLPEndpoint:   "localhost:4317",
		PrometheusPort: "9091",
		EnableTracing:  true,
		EnableMetrics:  true,
	}
}

// Init initializes observability and stores the global provider.
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	p, shutdown, err := mcpobs.Init(ctx, cfg)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	provider = p
	mu.Unlock()
	return shutdown, nil
}

func getProvider() *mcpobs.Provider {
	mu.RLock()
	defer mu.RUnlock()
	return provider
}

// StartSpan starts a tracing span. Returns a no-op span if not initialized.
func StartSpan(ctx context.Context, toolName string) (context.Context, trace.Span) {
	if p := getProvider(); p != nil {
		return p.StartSpan(ctx, toolName)
	}
	return ctx, nil
}

// StartToolExecution records a tool starting.
func StartToolExecution(ctx context.Context, toolName, category string) {
	if p := getProvider(); p != nil {
		p.StartToolExecution(ctx, toolName, category)
	}
}

// EndToolExecution records a tool finishing.
func EndToolExecution(ctx context.Context, toolName, category string) {
	if p := getProvider(); p != nil {
		p.EndToolExecution(ctx, toolName, category)
	}
}

// RecordToolInvocation records metrics for a completed invocation.
func RecordToolInvocation(ctx context.Context, toolName, category string, duration time.Duration, err error) {
	if p := getProvider(); p != nil {
		p.RecordToolInvocation(ctx, toolName, category, duration, err)
	}
}
