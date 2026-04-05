package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/secrets"
)

// OnePasswordProvider retrieves secrets from 1Password using the CLI.
type OnePasswordProvider struct {
	mu            sync.RWMutex
	vault         string
	account       string
	priority      int
	available     bool
	lastCheck     time.Time
	checkInterval time.Duration
	opPath        string // Path to op CLI
}

// OnePasswordOption configures the OnePasswordProvider.
type OnePasswordOption func(*OnePasswordProvider)

// WithOnePasswordVault sets the default vault.
func WithOnePasswordVault(vault string) OnePasswordOption {
	return func(p *OnePasswordProvider) {
		p.vault = vault
	}
}

// WithOnePasswordAccount sets the 1Password account.
func WithOnePasswordAccount(account string) OnePasswordOption {
	return func(p *OnePasswordProvider) {
		p.account = account
	}
}

// WithOnePasswordPriority sets the provider priority.
func WithOnePasswordPriority(priority int) OnePasswordOption {
	return func(p *OnePasswordProvider) {
		p.priority = priority
	}
}

// WithOPPath sets a custom path to the op CLI.
func WithOPPath(path string) OnePasswordOption {
	return func(p *OnePasswordProvider) {
		p.opPath = path
	}
}

// NewOnePasswordProvider creates a new 1Password CLI provider.
func NewOnePasswordProvider(opts ...OnePasswordOption) (*OnePasswordProvider, error) {
	p := &OnePasswordProvider{
		priority:      75, // Between AWS (50) and env (100)
		checkInterval: 5 * time.Minute,
		opPath:        "op",
	}
	for _, opt := range opts {
		opt(p)
	}

	// Find op CLI
	if p.opPath == "op" {
		path, err := exec.LookPath("op")
		if err != nil {
			p.available = false
			return p, nil // Return provider but mark as unavailable
		}
		p.opPath = path
	}

	// Check availability
	p.checkAvailability(context.Background())

	return p, nil
}

// checkAvailability tests if 1Password CLI is accessible and signed in.
func (p *OnePasswordProvider) checkAvailability(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	args := []string{"account", "get"}
	if p.account != "" {
		args = append(args, "--account", p.account)
	}
	args = append(args, "--format", "json")

	cmd := exec.CommandContext(ctx, p.opPath, args...)
	err := cmd.Run()
	p.available = err == nil
	p.lastCheck = time.Now()
}

// Name returns the provider identifier.
func (p *OnePasswordProvider) Name() string {
	return "1password"
}

// Get retrieves a secret from 1Password using op read.
func (p *OnePasswordProvider) Get(ctx context.Context, key string) (*secrets.Secret, error) {
	// op read supports secret references like:
	// - op://vault/item/field
	// - op://vault/item/section/field
	// - Just an item title (requires vault to be set)

	reference := key
	if !strings.HasPrefix(key, "op://") {
		// Build reference from vault and key
		if p.vault == "" {
			return nil, fmt.Errorf("%w: no vault specified and key is not a full reference", secrets.ErrInvalidKey)
		}
		// Assume key is in format "item/field" or "item" (uses default field)
		if strings.Contains(key, "/") {
			reference = fmt.Sprintf("op://%s/%s", p.vault, key)
		} else {
			// Assume password field
			reference = fmt.Sprintf("op://%s/%s/password", p.vault, key)
		}
	}

	args := []string{"read", reference}
	if p.account != "" {
		args = append(args, "--account", p.account)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.opPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if strings.Contains(stderrStr, "isn't an item") ||
			strings.Contains(stderrStr, "not found") ||
			strings.Contains(stderrStr, "no item") {
			return nil, secrets.ErrSecretNotFound
		}
		return nil, fmt.Errorf("%w: %s", secrets.ErrProviderError, stderrStr)
	}

	value := strings.TrimSpace(stdout.String())

	return &secrets.Secret{
		Key:    key,
		Value:  value,
		Source: "1password:" + reference,
	}, nil
}

// GetItem retrieves a full item from 1Password.
func (p *OnePasswordProvider) GetItem(ctx context.Context, itemName string) (map[string]string, error) {
	args := []string{"item", "get", itemName, "--format", "json"}
	if p.vault != "" {
		args = append(args, "--vault", p.vault)
	}
	if p.account != "" {
		args = append(args, "--account", p.account)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.opPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", secrets.ErrProviderError, stderr.String())
	}

	var item struct {
		Fields []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"fields"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &item); err != nil {
		return nil, fmt.Errorf("failed to parse item: %w", err)
	}

	result := make(map[string]string)
	for _, field := range item.Fields {
		if field.Value != "" {
			label := field.Label
			if label == "" {
				label = field.ID
			}
			result[label] = field.Value
		}
	}

	return result, nil
}

// List returns all item names from the configured vault.
func (p *OnePasswordProvider) List(ctx context.Context) ([]string, error) {
	args := []string{"item", "list", "--format", "json"}
	if p.vault != "" {
		args = append(args, "--vault", p.vault)
	}
	if p.account != "" {
		args = append(args, "--account", p.account)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.opPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", secrets.ErrProviderError, stderr.String())
	}

	var items []struct {
		Title string `json:"title"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &items); err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}

	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = item.Title
	}

	return keys, nil
}

// Exists checks if an item exists in 1Password.
func (p *OnePasswordProvider) Exists(ctx context.Context, key string) (bool, error) {
	args := []string{"item", "get", key, "--format", "json"}
	if p.vault != "" {
		args = append(args, "--vault", p.vault)
	}
	if p.account != "" {
		args = append(args, "--account", p.account)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.opPath, args...)
	err := cmd.Run()
	return err == nil, nil
}

// Priority returns the provider priority.
func (p *OnePasswordProvider) Priority() int {
	return p.priority
}

// IsAvailable returns true if 1Password CLI is accessible.
func (p *OnePasswordProvider) IsAvailable() bool {
	p.mu.RLock()
	available := p.available
	lastCheck := p.lastCheck
	p.mu.RUnlock()

	// Refresh check if stale
	if time.Since(lastCheck) > p.checkInterval {
		go p.checkAvailability(context.Background())
	}

	return available
}

// Health returns health status for the 1Password provider.
func (p *OnePasswordProvider) Health(ctx context.Context) secrets.ProviderHealth {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	args := []string{"account", "get", "--format", "json"}
	if p.account != "" {
		args = append(args, "--account", p.account)
	}

	cmd := exec.CommandContext(ctx, p.opPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	health := secrets.ProviderHealth{
		Name:      p.Name(),
		Available: err == nil,
		Latency:   time.Since(start),
		LastCheck: time.Now(),
	}

	if err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			health.Error = stderrStr
		} else {
			health.Error = err.Error()
		}
	}

	p.mu.Lock()
	p.available = health.Available
	p.lastCheck = health.LastCheck
	p.mu.Unlock()

	return health
}

// SignIn attempts to sign in to 1Password.
// This is a helper for interactive use; for automated use, rely on service accounts.
func (p *OnePasswordProvider) SignIn(ctx context.Context) error {
	// Check for service account token first
	if os.Getenv("OP_SERVICE_ACCOUNT_TOKEN") != "" {
		// Service account is configured, no sign-in needed
		p.checkAvailability(ctx)
		return nil
	}

	args := []string{"signin"}
	if p.account != "" {
		args = append(args, "--account", p.account)
	}

	cmd := exec.CommandContext(ctx, p.opPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sign-in failed: %w", err)
	}

	p.checkAvailability(ctx)
	return nil
}

// Close is a no-op for the 1Password provider.
func (p *OnePasswordProvider) Close() error {
	return nil
}

// Ensure OnePasswordProvider implements SecretProvider.
var _ secrets.SecretProvider = (*OnePasswordProvider)(nil)
