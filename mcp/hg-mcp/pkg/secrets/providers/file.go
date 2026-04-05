package providers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/secrets"
)

// FileProvider reads secrets from .env or similar files.
type FileProvider struct {
	mu         sync.RWMutex
	files      []string
	secrets    map[string]string
	priority   int
	loaded     bool
	lastLoaded time.Time
	watchFiles bool
}

// FileOption configures the FileProvider.
type FileOption func(*FileProvider)

// WithFiles sets the files to load secrets from.
func WithFiles(files ...string) FileOption {
	return func(p *FileProvider) {
		p.files = append(p.files, files...)
	}
}

// WithFilePriority sets the provider priority.
func WithFilePriority(priority int) FileOption {
	return func(p *FileProvider) {
		p.priority = priority
	}
}

// WithFileWatch enables watching files for changes.
func WithFileWatch(watch bool) FileOption {
	return func(p *FileProvider) {
		p.watchFiles = watch
	}
}

// NewFileProvider creates a new file-based secret provider.
func NewFileProvider(opts ...FileOption) *FileProvider {
	p := &FileProvider{
		secrets:  make(map[string]string),
		priority: 200, // Lower priority than env (env overrides files)
	}
	for _, opt := range opts {
		opt(p)
	}

	// Default to common .env locations
	if len(p.files) == 0 {
		p.files = []string{".env", ".env.local", ".env.secrets"}
	}
	return p
}

// Name returns the provider identifier.
func (p *FileProvider) Name() string {
	return "file"
}

// loadFiles loads secrets from configured files.
func (p *FileProvider) loadFiles() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.secrets = make(map[string]string)
	var lastErr error

	for _, file := range p.files {
		path := file
		if !filepath.IsAbs(path) {
			// Try current directory and home directory
			if _, err := os.Stat(path); os.IsNotExist(err) {
				home, _ := os.UserHomeDir()
				if home != "" {
					altPath := filepath.Join(home, path)
					if _, err := os.Stat(altPath); err == nil {
						path = altPath
					}
				}
			}
		}

		if err := p.loadFile(path); err != nil {
			if !os.IsNotExist(err) {
				lastErr = err
			}
		}
	}

	p.loaded = true
	p.lastLoaded = time.Now()
	return lastErr
}

// loadFile parses a single .env file.
func (p *FileProvider) loadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		idx := strings.Index(line, "=")
		if idx == -1 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		// Handle escape sequences in double-quoted strings
		value = strings.ReplaceAll(value, "\\n", "\n")
		value = strings.ReplaceAll(value, "\\t", "\t")

		p.secrets[key] = value
	}

	return scanner.Err()
}

// ensureLoaded loads files if not already loaded.
func (p *FileProvider) ensureLoaded() {
	p.mu.RLock()
	loaded := p.loaded
	p.mu.RUnlock()

	if !loaded {
		p.loadFiles()
	}
}

// Get retrieves a secret from loaded files.
func (p *FileProvider) Get(ctx context.Context, key string) (*secrets.Secret, error) {
	p.ensureLoaded()

	p.mu.RLock()
	value, ok := p.secrets[key]
	p.mu.RUnlock()

	if !ok {
		// Try uppercase
		p.mu.RLock()
		value, ok = p.secrets[strings.ToUpper(key)]
		p.mu.RUnlock()
		if !ok {
			return nil, secrets.ErrSecretNotFound
		}
		key = strings.ToUpper(key)
	}

	return &secrets.Secret{
		Key:    key,
		Value:  value,
		Source: "file",
	}, nil
}

// List returns all secret keys from loaded files.
func (p *FileProvider) List(ctx context.Context) ([]string, error) {
	p.ensureLoaded()

	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]string, 0, len(p.secrets))
	for key := range p.secrets {
		keys = append(keys, key)
	}
	return keys, nil
}

// Exists checks if a secret exists in loaded files.
func (p *FileProvider) Exists(ctx context.Context, key string) (bool, error) {
	p.ensureLoaded()

	p.mu.RLock()
	defer p.mu.RUnlock()

	_, ok := p.secrets[key]
	if !ok {
		_, ok = p.secrets[strings.ToUpper(key)]
	}
	return ok, nil
}

// Priority returns the provider priority.
func (p *FileProvider) Priority() int {
	return p.priority
}

// IsAvailable returns true if at least one file exists.
func (p *FileProvider) IsAvailable() bool {
	for _, file := range p.files {
		if _, err := os.Stat(file); err == nil {
			return true
		}
		// Try home directory
		home, _ := os.UserHomeDir()
		if home != "" {
			if _, err := os.Stat(filepath.Join(home, file)); err == nil {
				return true
			}
		}
	}
	return false
}

// Reload reloads secrets from files.
func (p *FileProvider) Reload() error {
	p.mu.Lock()
	p.loaded = false
	p.mu.Unlock()
	return p.loadFiles()
}

// Health returns health status for the file provider.
func (p *FileProvider) Health(ctx context.Context) secrets.ProviderHealth {
	start := time.Now()
	available := p.IsAvailable()

	var errMsg string
	if !available {
		errMsg = fmt.Sprintf("no files found: %v", p.files)
	}

	return secrets.ProviderHealth{
		Name:      p.Name(),
		Available: available,
		Latency:   time.Since(start),
		Error:     errMsg,
		LastCheck: time.Now(),
	}
}

// Close is a no-op for the file provider.
func (p *FileProvider) Close() error {
	return nil
}

// Ensure FileProvider implements SecretProvider.
var _ secrets.SecretProvider = (*FileProvider)(nil)
