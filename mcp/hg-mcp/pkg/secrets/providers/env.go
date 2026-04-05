// Package providers implements secret providers for various backends.
package providers

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/secrets"
)

// EnvProvider reads secrets from environment variables.
type EnvProvider struct {
	prefix   string
	priority int
}

// EnvOption configures the EnvProvider.
type EnvOption func(*EnvProvider)

// WithPrefix sets a prefix filter for environment variables.
// Only variables with this prefix will be considered secrets.
func WithPrefix(prefix string) EnvOption {
	return func(p *EnvProvider) {
		p.prefix = prefix
	}
}

// WithEnvPriority sets the provider priority.
func WithEnvPriority(priority int) EnvOption {
	return func(p *EnvProvider) {
		p.priority = priority
	}
}

// NewEnvProvider creates a new environment variable provider.
func NewEnvProvider(opts ...EnvOption) *EnvProvider {
	p := &EnvProvider{
		priority: 100, // Default priority (env vars often used for local dev overrides)
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Name returns the provider identifier.
func (p *EnvProvider) Name() string {
	return "env"
}

// Get retrieves a secret from environment variables.
func (p *EnvProvider) Get(ctx context.Context, key string) (*secrets.Secret, error) {
	lookupKey := key
	if p.prefix != "" && !strings.HasPrefix(key, p.prefix) {
		lookupKey = p.prefix + key
	}

	value := os.Getenv(lookupKey)
	if value == "" {
		// Try uppercase version
		value = os.Getenv(strings.ToUpper(lookupKey))
		if value == "" {
			return nil, secrets.ErrSecretNotFound
		}
		lookupKey = strings.ToUpper(lookupKey)
	}

	return &secrets.Secret{
		Key:    key,
		Value:  value,
		Source: "env:" + lookupKey,
	}, nil
}

// List returns all environment variable keys that match the prefix.
func (p *EnvProvider) List(ctx context.Context) ([]string, error) {
	var keys []string
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		if p.prefix == "" {
			keys = append(keys, key)
		} else if strings.HasPrefix(key, p.prefix) {
			// Remove prefix from the returned key
			keys = append(keys, strings.TrimPrefix(key, p.prefix))
		}
	}
	return keys, nil
}

// Exists checks if an environment variable exists.
func (p *EnvProvider) Exists(ctx context.Context, key string) (bool, error) {
	lookupKey := key
	if p.prefix != "" && !strings.HasPrefix(key, p.prefix) {
		lookupKey = p.prefix + key
	}

	_, exists := os.LookupEnv(lookupKey)
	if !exists {
		// Try uppercase version
		_, exists = os.LookupEnv(strings.ToUpper(lookupKey))
	}
	return exists, nil
}

// Priority returns the provider priority.
func (p *EnvProvider) Priority() int {
	return p.priority
}

// IsAvailable always returns true for env provider.
func (p *EnvProvider) IsAvailable() bool {
	return true
}

// Health returns health status for the env provider.
func (p *EnvProvider) Health(ctx context.Context) secrets.ProviderHealth {
	return secrets.ProviderHealth{
		Name:      p.Name(),
		Available: true,
		Latency:   0,
		LastCheck: time.Now(),
	}
}

// Close is a no-op for the env provider.
func (p *EnvProvider) Close() error {
	return nil
}

// Ensure EnvProvider implements SecretProvider.
var _ secrets.SecretProvider = (*EnvProvider)(nil)
