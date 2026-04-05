# Go Secrets Management Patterns and Best Practices

> Deep research on Go secrets management patterns relevant to the hg-mcp Go secrets manager implementation.

## Table of Contents

1. [12-Factor App Secrets Management](#1-12-factor-app-secrets-management)
2. [AWS Secrets Manager Go SDK Patterns](#2-aws-secrets-manager-go-sdk-patterns)
3. [1Password CLI Integration Patterns](#3-1password-cli-integration-patterns)
4. [Environment Variables vs Secrets Manager Tradeoffs](#4-environment-variables-vs-secrets-manager-tradeoffs)
5. [Secret Rotation and Hot-Reload Patterns](#5-secret-rotation-and-hot-reload-patterns)
6. [Caching Strategies and TTL Management](#6-caching-strategies-and-ttl-management)
7. [Sensitive Data Sanitization in Logs/Traces](#7-sensitive-data-sanitization-in-logstraces)
8. [External-Secrets Operator Patterns (Kubernetes)](#8-external-secrets-operator-patterns-kubernetes)
9. [HashiCorp Vault Integration Patterns](#9-hashicorp-vault-integration-patterns)
10. [Security Best Practices (Memory Protection, Secure Comparison)](#10-security-best-practices)
11. [Portable Abstraction Patterns](#11-portable-abstraction-patterns)
12. [Library Recommendations](#12-library-recommendations)

---

## 1. 12-Factor App Secrets Management

The [12-Factor App methodology](https://12factor.net/config) recommends storing configuration in environment variables, including secrets. However, this approach requires careful consideration for production systems.

### Core Principles

**Factor III (Config)** states:
- Strict separation of config from code
- Config stored in environment variables
- Config includes credentials and secrets

### Go Implementation Pattern

```go
package config

import (
    "os"
    "github.com/kelseyhightower/envconfig"
)

// Config holds application configuration following 12-factor principles
type Config struct {
    DatabaseURL    string `envconfig:"DATABASE_URL" required:"true"`
    APIKey         string `envconfig:"API_KEY" required:"true"`
    RedisPassword  string `envconfig:"REDIS_PASSWORD"`
    JWTSecret      string `envconfig:"JWT_SECRET" required:"true"`
    Port           int    `envconfig:"PORT" default:"8080"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
```

### Modern 12-Factor Approach: Delegation Pattern

The recommended approach is to let external systems handle secret synchronization:

```go
// Your Go app reads secrets from ENV vars as if they're already there
// The synchronization from secrets manager to ENV is handled externally

package main

import (
    "log"
    "os"
)

func main() {
    // Secrets are expected to be injected by:
    // 1. Kubernetes secrets mounted as ENV vars
    // 2. Doppler CLI: doppler run -- ./myapp
    // 3. Vault Agent sidecar
    // 4. AWS ECS task definition secrets

    dbPassword := os.Getenv("DATABASE_PASSWORD")
    if dbPassword == "" {
        log.Fatal("DATABASE_PASSWORD not set")
    }

    // Use the secret...
}
```

### References

- [12 Factor App Config](https://12factor.net/config)
- [Mark's Musings: 12 Factor App Secrets Management](https://mark.smithson.me/12-factor-app-secrets-management/)
- [Xenit: Twelve-factor app anno 2022](https://xenitab.github.io/blog/2022/02/23/12factor/)

---

## 2. AWS Secrets Manager Go SDK Patterns

AWS SDK for Go v2 is the recommended approach (v1 end of support: July 31, 2025).

### Basic Secret Retrieval

```go
package secrets

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AWSSecretsManager struct {
    client *secretsmanager.Client
}

func NewAWSSecretsManager(ctx context.Context, region string) (*AWSSecretsManager, error) {
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return nil, fmt.Errorf("unable to load SDK config: %w", err)
    }

    return &AWSSecretsManager{
        client: secretsmanager.NewFromConfig(cfg),
    }, nil
}

func (m *AWSSecretsManager) GetSecret(ctx context.Context, secretName string) (string, error) {
    input := &secretsmanager.GetSecretValueInput{
        SecretId:     aws.String(secretName),
        VersionStage: aws.String("AWSCURRENT"),
    }

    result, err := m.client.GetSecretValue(ctx, input)
    if err != nil {
        return "", fmt.Errorf("failed to retrieve secret %s: %w", secretName, err)
    }

    return *result.SecretString, nil
}

// GetSecretJSON retrieves and parses a JSON secret
func (m *AWSSecretsManager) GetSecretJSON(ctx context.Context, secretName string, v interface{}) error {
    secretString, err := m.GetSecret(ctx, secretName)
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(secretString), v)
}

// DatabaseCredentials represents typical DB secret structure
type DatabaseCredentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
    Host     string `json:"host"`
    Port     int    `json:"port"`
    DBName   string `json:"dbname"`
}
```

### Batch Secret Retrieval

```go
func (m *AWSSecretsManager) GetSecrets(ctx context.Context, secretNames []string) (map[string]string, error) {
    filters := make([]types.Filter, 0, len(secretNames))
    for _, name := range secretNames {
        filters = append(filters, types.Filter{
            Key:    types.FilterNameStringTypeName,
            Values: []string{name},
        })
    }

    input := &secretsmanager.BatchGetSecretValueInput{
        SecretIdList: secretNames,
    }

    result, err := m.client.BatchGetSecretValue(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("batch get failed: %w", err)
    }

    secrets := make(map[string]string)
    for _, sv := range result.SecretValues {
        secrets[*sv.Name] = *sv.SecretString
    }

    return secrets, nil
}
```

### References

- [AWS SDK Go v2 Secrets Manager Package](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/secretsmanager)
- [AWS Documentation: Retrieving secrets with Go SDK](https://docs.aws.amazon.com/secretsmanager/latest/userguide/retrieving-secrets-go-sdk.html)
- [Gruntwork go-commons secretsmanager](https://github.com/gruntwork-io/go-commons/blob/master/awscommons/v2/secretsmanager.go)

---

## 3. 1Password CLI Integration Patterns

1Password offers multiple integration approaches for Go applications.

### Option 1: Official 1Password Go SDK (Recommended)

```go
package secrets

import (
    "context"
    "fmt"

    "github.com/1password/onepassword-sdk-go"
)

type OnePasswordManager struct {
    client *onepassword.Client
}

func NewOnePasswordManager(ctx context.Context, serviceAccountToken string) (*OnePasswordManager, error) {
    client, err := onepassword.NewClient(
        ctx,
        onepassword.WithServiceAccountToken(serviceAccountToken),
        onepassword.WithIntegrationInfo("hg-mcp", "1.0.0"),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create 1Password client: %w", err)
    }

    return &OnePasswordManager{client: client}, nil
}

// GetSecret retrieves a secret using a secret reference
// Format: op://vault-name/item-name/field-name
func (m *OnePasswordManager) GetSecret(ctx context.Context, secretRef string) (string, error) {
    secret, err := m.client.Secrets.Resolve(ctx, secretRef)
    if err != nil {
        return "", fmt.Errorf("failed to resolve secret %s: %w", secretRef, err)
    }
    return secret, nil
}

// ListItems lists items in a vault
func (m *OnePasswordManager) ListItems(ctx context.Context, vaultID string) ([]onepassword.ItemOverview, error) {
    items, err := m.client.Items.ListAll(ctx, vaultID)
    if err != nil {
        return nil, fmt.Errorf("failed to list items: %w", err)
    }
    return items, nil
}
```

### Option 2: 1Password Connect SDK (Self-Hosted)

For self-hosted 1Password Connect servers:

```go
package secrets

import (
    "os"

    "github.com/1Password/connect-sdk-go/connect"
    "github.com/1Password/connect-sdk-go/onepassword"
)

func NewConnectClient() (connect.Client, error) {
    // Uses OP_CONNECT_HOST and OP_CONNECT_TOKEN env vars
    return connect.NewClientFromEnvironment()
}

func GetItemByTitle(client connect.Client, vaultUUID, title string) (*onepassword.Item, error) {
    return client.GetItemByTitle(title, vaultUUID)
}
```

### Option 3: CLI Wrapper (Development/Scripts)

For simpler use cases or shell script integration:

```go
package secrets

import (
    "encoding/json"
    "fmt"
    "os/exec"
)

type OPCLIWrapper struct{}

func (o *OPCLIWrapper) GetField(vault, item, field string) (string, error) {
    cmd := exec.Command("op", "read", fmt.Sprintf("op://%s/%s/%s", vault, item, field))
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("op read failed: %w", err)
    }
    return string(output), nil
}

func (o *OPCLIWrapper) GetItemJSON(vault, item string) (map[string]interface{}, error) {
    cmd := exec.Command("op", "item", "get", item, "--vault", vault, "--format", "json")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("op item get failed: %w", err)
    }

    var result map[string]interface{}
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, err
    }
    return result, nil
}
```

### References

- [1Password Go SDK](https://github.com/1Password/onepassword-sdk-go)
- [1Password Connect SDK for Go](https://github.com/1Password/connect-sdk-go)
- [1Password Developer Docs: SDKs](https://developer.1password.com/docs/sdks/)
- [1Password CLI Documentation](https://developer.1password.com/docs/cli/get-started/)

---

## 4. Environment Variables vs Secrets Manager Tradeoffs

### Comparison Matrix

| Aspect | Environment Variables | Secrets Manager |
|--------|----------------------|-----------------|
| **Setup Complexity** | Simple | Moderate to Complex |
| **Security at Rest** | Plain text | Encrypted |
| **Access Control** | OS-level | Fine-grained IAM/RBAC |
| **Audit Trail** | None | Full audit logs |
| **Rotation** | Manual, requires restart | Automatic, hot-reload capable |
| **Cost** | Free | Per-secret pricing |
| **Multi-environment** | Requires careful management | Built-in environment support |
| **Development Experience** | Excellent | Requires local setup |

### When to Use Environment Variables

```go
// Good for: Non-sensitive configuration, local development
type AppConfig struct {
    LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
    Port        int    `envconfig:"PORT" default:"8080"`
    Environment string `envconfig:"ENVIRONMENT" default:"development"`
    Debug       bool   `envconfig:"DEBUG" default:"false"`
}
```

### When to Use Secrets Manager

```go
// Good for: Production secrets, credentials, API keys
type SecureConfig struct {
    DatabasePassword string // From secrets manager
    APIKey           string // From secrets manager
    JWTSigningKey    []byte // From secrets manager
    EncryptionKey    []byte // From secrets manager
}
```

### Hybrid Approach (Recommended)

```go
package config

import (
    "context"
    "os"
)

type Config struct {
    // Non-sensitive: from environment
    Port        int    `envconfig:"PORT" default:"8080"`
    Environment string `envconfig:"ENVIRONMENT" required:"true"`
    LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

    // Sensitive: from secrets manager
    DBPassword    string
    APIKey        string
    JWTSecret     []byte
}

func Load(ctx context.Context, secretsClient SecretsClient) (*Config, error) {
    cfg := &Config{}

    // Load non-sensitive from env
    if err := envconfig.Process("", cfg); err != nil {
        return nil, err
    }

    // Load sensitive from secrets manager (production)
    // or fall back to env vars (development)
    if cfg.Environment == "production" {
        var err error
        cfg.DBPassword, err = secretsClient.GetSecret(ctx, "db/password")
        if err != nil {
            return nil, err
        }
        cfg.APIKey, err = secretsClient.GetSecret(ctx, "api/key")
        if err != nil {
            return nil, err
        }
    } else {
        // Development: allow env var fallback
        cfg.DBPassword = os.Getenv("DB_PASSWORD")
        cfg.APIKey = os.Getenv("API_KEY")
    }

    return cfg, nil
}
```

### References

- [Medium: Secret Manager or Environment Variables?](https://medium.com/smallcase-engineering/decoding-security-secret-manager-or-environment-variables-9b9beb7c35b7)
- [Tolu Banji: Secrets Management in Go Applications](https://tolubanji.com/posts/secrets-management-in-go/)
- [Arcjet: Storing secrets in env vars considered harmful](https://blog.arcjet.com/storing-secrets-in-env-vars-considered-harmful/)

---

## 5. Secret Rotation and Hot-Reload Patterns

### Pattern 1: SIGHUP Signal Handler

```go
package secrets

import (
    "context"
    "log"
    "os"
    "os/signal"
    "sync"
    "syscall"
)

type HotReloadableSecrets struct {
    mu           sync.RWMutex
    secrets      map[string]string
    loader       func(context.Context) (map[string]string, error)
    reloadChan   chan os.Signal
}

func NewHotReloadableSecrets(loader func(context.Context) (map[string]string, error)) *HotReloadableSecrets {
    hrs := &HotReloadableSecrets{
        secrets:    make(map[string]string),
        loader:     loader,
        reloadChan: make(chan os.Signal, 1),
    }

    // Register SIGHUP handler
    signal.Notify(hrs.reloadChan, syscall.SIGHUP)

    return hrs
}

func (hrs *HotReloadableSecrets) Start(ctx context.Context) error {
    // Initial load
    if err := hrs.reload(ctx); err != nil {
        return err
    }

    // Watch for reload signals
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case <-hrs.reloadChan:
                log.Println("Received SIGHUP, reloading secrets...")
                if err := hrs.reload(ctx); err != nil {
                    log.Printf("Failed to reload secrets: %v", err)
                } else {
                    log.Println("Secrets reloaded successfully")
                }
            }
        }
    }()

    return nil
}

func (hrs *HotReloadableSecrets) reload(ctx context.Context) error {
    newSecrets, err := hrs.loader(ctx)
    if err != nil {
        return err
    }

    hrs.mu.Lock()
    hrs.secrets = newSecrets
    hrs.mu.Unlock()

    return nil
}

func (hrs *HotReloadableSecrets) Get(key string) (string, bool) {
    hrs.mu.RLock()
    defer hrs.mu.RUnlock()
    val, ok := hrs.secrets[key]
    return val, ok
}
```

### Pattern 2: Periodic Refresh with slok/reload

```go
package secrets

import (
    "context"
    "sync"
    "time"
)

type PeriodicRefreshSecrets struct {
    mu            sync.RWMutex
    secrets       map[string]string
    client        SecretsClient
    refreshPeriod time.Duration
    secretNames   []string
    listeners     []func(map[string]string)
}

func (prs *PeriodicRefreshSecrets) Start(ctx context.Context) {
    ticker := time.NewTicker(prs.refreshPeriod)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            prs.refresh(ctx)
        }
    }
}

func (prs *PeriodicRefreshSecrets) refresh(ctx context.Context) {
    newSecrets := make(map[string]string)

    for _, name := range prs.secretNames {
        val, err := prs.client.GetSecret(ctx, name)
        if err != nil {
            // Log error but don't fail - keep using old value
            continue
        }
        newSecrets[name] = val
    }

    prs.mu.Lock()
    prs.secrets = newSecrets
    prs.mu.Unlock()

    // Notify listeners
    for _, listener := range prs.listeners {
        listener(newSecrets)
    }
}

func (prs *PeriodicRefreshSecrets) OnChange(listener func(map[string]string)) {
    prs.listeners = append(prs.listeners, listener)
}
```

### Pattern 3: File-Based Secret Mounting (Kubernetes)

```go
package secrets

import (
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/fsnotify/fsnotify"
)

type FileMountedSecrets struct {
    mu         sync.RWMutex
    secrets    map[string]string
    mountPath  string
    watcher    *fsnotify.Watcher
}

func NewFileMountedSecrets(mountPath string) (*FileMountedSecrets, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    fms := &FileMountedSecrets{
        secrets:   make(map[string]string),
        mountPath: mountPath,
        watcher:   watcher,
    }

    // Initial load
    if err := fms.loadAll(); err != nil {
        return nil, err
    }

    // Watch for changes
    if err := watcher.Add(mountPath); err != nil {
        return nil, err
    }

    return fms, nil
}

func (fms *FileMountedSecrets) Watch(ctx context.Context) {
    // Debounce to handle multiple rapid changes
    var debounceTimer *time.Timer

    for {
        select {
        case <-ctx.Done():
            return
        case event := <-fms.watcher.Events:
            if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
                if debounceTimer != nil {
                    debounceTimer.Stop()
                }
                debounceTimer = time.AfterFunc(100*time.Millisecond, func() {
                    fms.loadAll()
                })
            }
        case err := <-fms.watcher.Errors:
            // Log error
            _ = err
        }
    }
}

func (fms *FileMountedSecrets) loadAll() error {
    entries, err := os.ReadDir(fms.mountPath)
    if err != nil {
        return err
    }

    newSecrets := make(map[string]string)
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        content, err := os.ReadFile(filepath.Join(fms.mountPath, entry.Name()))
        if err != nil {
            continue
        }
        newSecrets[entry.Name()] = string(content)
    }

    fms.mu.Lock()
    fms.secrets = newSecrets
    fms.mu.Unlock()

    return nil
}
```

### References

- [rossedman.io: Hot Reload SIGHUP with Go](https://rossedman.io/blog/computers/hot-reload-sighup-with-go/)
- [ITNEXT: Clean and simple hot-reloading](https://itnext.io/clean-and-simple-hot-reloading-on-uninterrupted-go-applications-5974230ab4c5)
- [HashiCorp: Refresh Secrets for Kubernetes with Vault Agent](https://www.hashicorp.com/en/blog/refresh-secrets-for-kubernetes-applications-with-vault-agent)
- [Stakater Reloader](https://github.com/stakater/Reloader)

---

## 6. Caching Strategies and TTL Management

### AWS Secrets Manager Caching Client

```go
package secrets

import (
    "context"
    "time"

    "github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

func NewCachedAWSSecretsManager() (*secretcache.Cache, error) {
    config := secretcache.CacheConfig{
        MaxCacheSize: 1000,                           // Max cached secrets
        CacheItemTTL: time.Hour.Nanoseconds(),        // Default: 1 hour
        VersionStage: "AWSCURRENT",
    }

    cache, err := secretcache.New(
        func(c *secretcache.Cache) { c.CacheConfig = config },
    )
    if err != nil {
        return nil, err
    }

    return cache, nil
}

// Usage
func GetCachedSecret(cache *secretcache.Cache, secretID string) (string, error) {
    return cache.GetSecretString(secretID)
}
```

### Custom TTL Cache Implementation

```go
package secrets

import (
    "context"
    "math/rand"
    "sync"
    "time"
)

type CacheEntry struct {
    Value     string
    ExpiresAt time.Time
}

type TTLCache struct {
    mu       sync.RWMutex
    entries  map[string]*CacheEntry
    ttl      time.Duration
    jitter   time.Duration // Prevents thundering herd
    client   SecretsClient
}

func NewTTLCache(client SecretsClient, ttl, jitter time.Duration) *TTLCache {
    cache := &TTLCache{
        entries: make(map[string]*CacheEntry),
        ttl:     ttl,
        jitter:  jitter,
        client:  client,
    }

    // Start background cleanup
    go cache.cleanup()

    return cache
}

func (c *TTLCache) Get(ctx context.Context, key string) (string, error) {
    c.mu.RLock()
    entry, exists := c.entries[key]
    c.mu.RUnlock()

    if exists && time.Now().Before(entry.ExpiresAt) {
        return entry.Value, nil
    }

    // Cache miss or expired - fetch from source
    value, err := c.client.GetSecret(ctx, key)
    if err != nil {
        // On error, return stale value if available (fail-open)
        if exists {
            return entry.Value, nil
        }
        return "", err
    }

    // Add jitter to prevent thundering herd
    jitterDuration := time.Duration(rand.Int63n(int64(c.jitter)))
    expiresAt := time.Now().Add(c.ttl + jitterDuration)

    c.mu.Lock()
    c.entries[key] = &CacheEntry{
        Value:     value,
        ExpiresAt: expiresAt,
    }
    c.mu.Unlock()

    return value, nil
}

func (c *TTLCache) Invalidate(key string) {
    c.mu.Lock()
    delete(c.entries, key)
    c.mu.Unlock()
}

func (c *TTLCache) InvalidateAll() {
    c.mu.Lock()
    c.entries = make(map[string]*CacheEntry)
    c.mu.Unlock()
}

func (c *TTLCache) cleanup() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        c.mu.Lock()
        now := time.Now()
        for key, entry := range c.entries {
            if now.After(entry.ExpiresAt) {
                delete(c.entries, key)
            }
        }
        c.mu.Unlock()
    }
}
```

### Using jellydator/ttlcache

```go
package secrets

import (
    "context"
    "time"

    "github.com/jellydator/ttlcache/v3"
)

type CachedSecretsManager struct {
    cache  *ttlcache.Cache[string, string]
    client SecretsClient
}

func NewCachedSecretsManager(client SecretsClient, ttl time.Duration) *CachedSecretsManager {
    cache := ttlcache.New[string, string](
        ttlcache.WithTTL[string, string](ttl),
        ttlcache.WithCapacity[string, string](1000),
    )

    // Start automatic cleanup
    go cache.Start()

    return &CachedSecretsManager{
        cache:  cache,
        client: client,
    }
}

func (m *CachedSecretsManager) GetSecret(ctx context.Context, key string) (string, error) {
    item := m.cache.Get(key)
    if item != nil && !item.IsExpired() {
        return item.Value(), nil
    }

    // Fetch from source
    value, err := m.client.GetSecret(ctx, key)
    if err != nil {
        return "", err
    }

    m.cache.Set(key, value, ttlcache.DefaultTTL)
    return value, nil
}

func (m *CachedSecretsManager) Close() {
    m.cache.Stop()
}
```

### References

- [AWS Secrets Manager Caching Go](https://github.com/aws/aws-secretsmanager-caching-go)
- [jellydator/ttlcache](https://github.com/jellydator/ttlcache)
- [patrickmn/go-cache](https://github.com/patrickmn/go-cache)
- [Alex Edwards: Implementing in-memory cache in Go](https://www.alexedwards.net/blog/implementing-an-in-memory-cache-in-go)

---

## 7. Sensitive Data Sanitization in Logs/Traces

### Pattern 1: Custom Types with LogValuer Interface

```go
package secrets

import (
    "log/slog"
)

// Secret is a string that redacts itself when logged
type Secret string

func (s Secret) String() string {
    return "[REDACTED]"
}

func (s Secret) GoString() string {
    return "[REDACTED]"
}

func (s Secret) LogValue() slog.Value {
    return slog.StringValue("[REDACTED]")
}

// Unwrap returns the actual secret value
func (s Secret) Unwrap() string {
    return string(s)
}

// Usage
type DatabaseConfig struct {
    Host     string
    Port     int
    Username string
    Password Secret // Will be redacted in logs
}

func Example() {
    cfg := DatabaseConfig{
        Host:     "localhost",
        Port:     5432,
        Username: "admin",
        Password: Secret("super-secret-password"),
    }

    slog.Info("database config", "config", cfg)
    // Output: database config config={Host:localhost Port:5432 Username:admin Password:[REDACTED]}
}
```

### Pattern 2: ReplaceAttr Handler for slog

```go
package logging

import (
    "log/slog"
    "os"
    "strings"
)

var sensitiveKeys = map[string]bool{
    "password":      true,
    "token":         true,
    "authorization": true,
    "bearer":        true,
    "secret":        true,
    "api_key":       true,
    "apikey":        true,
    "access_token":  true,
    "refresh_token": true,
    "private_key":   true,
    "ssn":           true,
    "credit_card":   true,
}

func NewSanitizedLogger() *slog.Logger {
    opts := &slog.HandlerOptions{
        ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
            // Check if key is sensitive
            keyLower := strings.ToLower(a.Key)
            if sensitiveKeys[keyLower] {
                return slog.String(a.Key, "[REDACTED]")
            }

            // Check for partial matches
            for sensitive := range sensitiveKeys {
                if strings.Contains(keyLower, sensitive) {
                    return slog.String(a.Key, "[REDACTED]")
                }
            }

            return a
        },
    }

    return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
```

### Pattern 3: Using logfusc Library

```go
package secrets

import (
    "fmt"

    "github.com/AngusGMorrison/logfusc"
)

type Credentials struct {
    Username string
    Password logfusc.Secret[string] // Automatically redacted
    APIKey   logfusc.Secret[string]
}

func (c Credentials) String() string {
    return fmt.Sprintf("Credentials{Username: %s, Password: %s, APIKey: %s}",
        c.Username, c.Password, c.APIKey)
}

func Example() {
    creds := Credentials{
        Username: "admin",
        Password: logfusc.NewSecret("super-secret"),
        APIKey:   logfusc.NewSecret("api-key-12345"),
    }

    fmt.Println(creds)
    // Output: Credentials{Username: admin, Password: [REDACTED], APIKey: [REDACTED]}

    // Access actual value when needed
    actualPassword := creds.Password.Expose()
}
```

### Pattern 4: Using masq Library

```go
package logging

import (
    "log/slog"
    "os"
    "regexp"

    "github.com/m-mizutani/masq"
)

type AccessToken string

func NewMaskedLogger() *slog.Logger {
    // Configure masq with multiple redaction strategies
    filter := masq.New(
        // Redact custom types
        masq.WithType[AccessToken](),

        // Redact by struct tag
        masq.WithTag("secret"),

        // Redact by regex pattern (e.g., phone numbers)
        masq.WithRegex(regexp.MustCompile(`\d{3}-\d{3}-\d{4}`)),

        // Redact by field name prefix
        masq.WithFieldPrefix("password"),
        masq.WithFieldPrefix("secret"),
    )

    opts := &slog.HandlerOptions{
        ReplaceAttr: filter.ReplaceAttr,
    }

    return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

type UserData struct {
    Email    string
    Phone    string      // Will be masked if matches phone regex
    Password string      `masq:"secret"` // Tagged for masking
    Token    AccessToken // Custom type, always masked
}
```

### Pattern 5: Allow-List Approach (Recommended)

```go
package logging

import (
    "log/slog"
    "reflect"
)

// Loggable interface marks types as safe to log
type Loggable interface {
    LoggableFields() map[string]any
}

type SafeUserData struct {
    ID       string
    Email    string
    Password string // Never logged
    APIKey   string // Never logged
}

// Only expose safe fields for logging
func (u SafeUserData) LoggableFields() map[string]any {
    return map[string]any{
        "id":    u.ID,
        "email": u.Email,
        // Password and APIKey intentionally omitted
    }
}

func LogSafe(logger *slog.Logger, msg string, data Loggable) {
    fields := data.LoggableFields()
    attrs := make([]any, 0, len(fields)*2)
    for k, v := range fields {
        attrs = append(attrs, k, v)
    }
    logger.Info(msg, attrs...)
}
```

### References

- [Arcjet: Redacting sensitive data from logs with slog](https://blog.arcjet.com/redacting-sensitive-data-from-logs-with-go-log-slog/)
- [logfusc Library](https://github.com/AngusGMorrison/logfusc)
- [masq Library](https://github.com/m-mizutani/masq)
- [Go Blog: Structured Logging with slog](https://go.dev/blog/slog)

---

## 8. External-Secrets Operator Patterns (Kubernetes)

### Overview

External Secrets Operator synchronizes secrets from external APIs (AWS Secrets Manager, Vault, etc.) into Kubernetes Secrets.

### SecretStore Configuration

```yaml
# AWS Secrets Manager SecretStore
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets
  namespace: production
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa
```

### ExternalSecret Resource

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database-credentials
  namespace: production
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets
    kind: SecretStore
  target:
    name: database-secret
    creationPolicy: Owner
  data:
    - secretKey: username
      remoteRef:
        key: production/database
        property: username
    - secretKey: password
      remoteRef:
        key: production/database
        property: password
```

### Go Pattern: Reading Kubernetes Secrets

```go
package secrets

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
)

// KubernetesSecretReader reads secrets mounted as files or env vars
type KubernetesSecretReader struct {
    mountPath string
}

func NewKubernetesSecretReader(mountPath string) *KubernetesSecretReader {
    if mountPath == "" {
        mountPath = "/var/run/secrets/app"
    }
    return &KubernetesSecretReader{mountPath: mountPath}
}

// GetSecret reads a secret from a mounted file
func (r *KubernetesSecretReader) GetSecret(ctx context.Context, key string) (string, error) {
    // First try file mount
    filePath := filepath.Join(r.mountPath, key)
    if data, err := os.ReadFile(filePath); err == nil {
        return string(data), nil
    }

    // Fall back to environment variable
    if val := os.Getenv(key); val != "" {
        return val, nil
    }

    return "", fmt.Errorf("secret %s not found", key)
}
```

### Go Templating in ExternalSecrets

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: app-config
spec:
  secretStoreRef:
    name: aws-secrets
    kind: SecretStore
  target:
    name: app-config-secret
    template:
      engineVersion: v2
      data:
        # Use Go template to construct connection string
        DATABASE_URL: "postgres://{{ .username }}:{{ .password }}@{{ .host }}:{{ .port }}/{{ .database }}"
  dataFrom:
    - extract:
        key: production/database
```

### References

- [External Secrets Operator Documentation](https://external-secrets.io/)
- [External Secrets GitHub](https://github.com/external-secrets/external-secrets)
- [InfoQ: Managing Kubernetes Secrets with ESO](https://www.infoq.com/articles/k8s-external-secrets-operator/)

---

## 9. HashiCorp Vault Integration Patterns

### Using the New vault-client-go Library (Recommended)

```go
package secrets

import (
    "context"
    "fmt"
    "time"

    "github.com/hashicorp/vault-client-go"
    "github.com/hashicorp/vault-client-go/schema"
)

type VaultClient struct {
    client *vault.Client
}

func NewVaultClient(address, token string) (*VaultClient, error) {
    client, err := vault.New(
        vault.WithAddress(address),
        vault.WithRequestTimeout(30*time.Second),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create vault client: %w", err)
    }

    if err := client.SetToken(token); err != nil {
        return nil, fmt.Errorf("failed to set token: %w", err)
    }

    return &VaultClient{client: client}, nil
}

// GetKVSecret retrieves a secret from KV v2 secrets engine
func (v *VaultClient) GetKVSecret(ctx context.Context, mount, path string) (map[string]interface{}, error) {
    resp, err := v.client.Secrets.KvV2Read(ctx, path, vault.WithMountPath(mount))
    if err != nil {
        return nil, fmt.Errorf("failed to read secret: %w", err)
    }

    return resp.Data.Data, nil
}

// GetKVSecretField retrieves a specific field from a KV v2 secret
func (v *VaultClient) GetKVSecretField(ctx context.Context, mount, path, field string) (string, error) {
    data, err := v.GetKVSecret(ctx, mount, path)
    if err != nil {
        return "", err
    }

    value, ok := data[field]
    if !ok {
        return "", fmt.Errorf("field %s not found in secret", field)
    }

    strValue, ok := value.(string)
    if !ok {
        return "", fmt.Errorf("field %s is not a string", field)
    }

    return strValue, nil
}
```

### AppRole Authentication

```go
package secrets

import (
    "context"
    "fmt"
    "time"

    "github.com/hashicorp/vault-client-go"
    "github.com/hashicorp/vault-client-go/schema"
)

type VaultAppRoleClient struct {
    client   *vault.Client
    roleID   string
    secretID string
    mount    string
}

func NewVaultAppRoleClient(address, roleID, secretID, mount string) (*VaultAppRoleClient, error) {
    client, err := vault.New(
        vault.WithAddress(address),
        vault.WithRequestTimeout(30*time.Second),
    )
    if err != nil {
        return nil, err
    }

    if mount == "" {
        mount = "approle"
    }

    return &VaultAppRoleClient{
        client:   client,
        roleID:   roleID,
        secretID: secretID,
        mount:    mount,
    }, nil
}

func (v *VaultAppRoleClient) Authenticate(ctx context.Context) error {
    resp, err := v.client.Auth.AppRoleLogin(
        ctx,
        schema.AppRoleLoginRequest{
            RoleId:   v.roleID,
            SecretId: v.secretID,
        },
        vault.WithMountPath(v.mount),
    )
    if err != nil {
        return fmt.Errorf("AppRole login failed: %w", err)
    }

    if err := v.client.SetToken(resp.Auth.ClientToken); err != nil {
        return fmt.Errorf("failed to set token: %w", err)
    }

    return nil
}
```

### Using Traditional api Package

```go
package secrets

import (
    "fmt"

    "github.com/hashicorp/vault/api"
)

func NewVaultAPIClient(address, token string) (*api.Client, error) {
    config := api.DefaultConfig()
    config.Address = address

    client, err := api.NewClient(config)
    if err != nil {
        return nil, err
    }

    client.SetToken(token)

    return client, nil
}

func ReadKVV2Secret(client *api.Client, mount, path string) (map[string]interface{}, error) {
    secret, err := client.KVv2(mount).Get(context.Background(), path)
    if err != nil {
        return nil, fmt.Errorf("unable to read secret: %w", err)
    }

    return secret.Data, nil
}
```

### Dynamic Database Credentials

```go
package secrets

import (
    "context"
    "fmt"
    "time"

    "github.com/hashicorp/vault-client-go"
)

type DatabaseCredentials struct {
    Username  string
    Password  string
    ExpiresAt time.Time
}

func (v *VaultClient) GetDatabaseCredentials(ctx context.Context, role string) (*DatabaseCredentials, error) {
    resp, err := v.client.Secrets.DatabaseGenerateCredentials(ctx, role)
    if err != nil {
        return nil, fmt.Errorf("failed to generate credentials: %w", err)
    }

    return &DatabaseCredentials{
        Username:  resp.Data["username"].(string),
        Password:  resp.Data["password"].(string),
        ExpiresAt: time.Now().Add(time.Duration(resp.LeaseDuration) * time.Second),
    }, nil
}
```

### References

- [vault-client-go GitHub](https://github.com/hashicorp/vault-client-go)
- [vault/api Package](https://pkg.go.dev/github.com/hashicorp/vault/api)
- [hello-vault-go Examples](https://github.com/hashicorp/hello-vault-go)
- [HashiCorp Blog: Vault Client Libraries](https://www.hashicorp.com/en/blog/vault-client-libraries-for-go-and-net-are-now-in-public-beta)

---

## 10. Security Best Practices

### Memory Protection with memguard

```go
package secrets

import (
    "github.com/awnumar/memguard"
)

func init() {
    // Catch interrupt signals to safely purge secrets
    memguard.CatchInterrupt()
}

// SecureSecret wraps sensitive data with memory protection
type SecureSecret struct {
    enclave *memguard.Enclave
}

func NewSecureSecret(data []byte) *SecureSecret {
    // Create encrypted enclave
    enclave := memguard.NewEnclave(data)

    // Wipe the original data
    memguard.WipeBytes(data)

    return &SecureSecret{enclave: enclave}
}

// Use decrypts the secret and calls the function with it
// The secret is automatically destroyed after use
func (s *SecureSecret) Use(fn func([]byte) error) error {
    // Decrypt into a locked buffer
    buf, err := s.enclave.Open()
    if err != nil {
        return err
    }
    defer buf.Destroy()

    // Make immutable while in use
    buf.Freeze()

    return fn(buf.Bytes())
}

// Example usage
func Example() {
    defer memguard.Purge() // Clean up on exit

    secret := NewSecureSecret([]byte("my-api-key"))

    err := secret.Use(func(key []byte) error {
        // Use the key here
        // It will be securely destroyed after this function returns
        return nil
    })
    if err != nil {
        memguard.SafePanic(err)
    }
}
```

### HashiCorp mlock for Memory Locking

```go
package secrets

import (
    "log"

    "github.com/hashicorp/go-secure-stdlib/mlock"
)

func init() {
    // Prevent memory from being swapped to disk
    if mlock.Supported() {
        if err := mlock.LockMemory(); err != nil {
            log.Printf("Warning: failed to lock memory: %v", err)
        }
    }
}
```

### Constant-Time Comparison

```go
package secrets

import (
    "crypto/subtle"
)

// SecureCompare performs constant-time comparison of two secrets
// Returns true if the secrets are equal
func SecureCompare(a, b []byte) bool {
    // First check lengths to avoid timing leak on length
    if len(a) != len(b) {
        return false
    }
    return subtle.ConstantTimeCompare(a, b) == 1
}

// SecureCompareStrings compares two string secrets in constant time
func SecureCompareStrings(a, b string) bool {
    return SecureCompare([]byte(a), []byte(b))
}

// ValidateAPIKey validates an API key in constant time
func ValidateAPIKey(provided, expected string) bool {
    // Ensure both keys are the same length to prevent length oracle
    if len(provided) != len(expected) {
        // Still perform comparison to maintain constant time
        subtle.ConstantTimeCompare([]byte(provided), []byte(provided))
        return false
    }
    return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
```

### Secure Secret Zeroing

```go
package secrets

import (
    "unsafe"
)

// WipeBytes securely zeros a byte slice
// Uses a compiler barrier to prevent optimization
func WipeBytes(b []byte) {
    for i := range b {
        b[i] = 0
    }
    // Memory barrier to prevent compiler optimization
    _ = *(*byte)(unsafe.Pointer(&b[0]))
}

// WipeString converts string to bytes and wipes it
// Note: Go strings are immutable, so this creates a copy
func WipeString(s *string) {
    b := []byte(*s)
    WipeBytes(b)
    *s = ""
}

// SecureBuffer wraps a byte slice with automatic wiping
type SecureBuffer struct {
    data []byte
}

func NewSecureBuffer(size int) *SecureBuffer {
    return &SecureBuffer{data: make([]byte, size)}
}

func (sb *SecureBuffer) Bytes() []byte {
    return sb.data
}

func (sb *SecureBuffer) Wipe() {
    WipeBytes(sb.data)
}
```

### Prevent Core Dumps

```go
package main

import (
    "syscall"
)

func disableCoreDumps() {
    // Set RLIMIT_CORE to 0 to prevent core dumps
    var rLimit syscall.Rlimit
    rLimit.Cur = 0
    rLimit.Max = 0
    syscall.Setrlimit(syscall.RLIMIT_CORE, &rLimit)
}

func main() {
    disableCoreDumps()
    // ... rest of application
}
```

### References

- [memguard Package](https://pkg.go.dev/github.com/awnumar/memguard)
- [spacetime.dev: Memory Security in Go](https://spacetime.dev/memory-security-go)
- [crypto/subtle Package](https://pkg.go.dev/crypto/subtle)
- [HashiCorp go-secure-stdlib/mlock](https://pkg.go.dev/github.com/hashicorp/go-secure-stdlib/mlock)
- [Go Issue #18645: Securely wipe sensitive data](https://github.com/golang/go/issues/18645)

---

## 11. Portable Abstraction Patterns

### Go CDK Secrets Package

```go
package secrets

import (
    "context"

    "gocloud.dev/secrets"
    _ "gocloud.dev/secrets/awskms"       // AWS KMS
    _ "gocloud.dev/secrets/gcpkms"       // GCP KMS
    _ "gocloud.dev/secrets/azurekeyvault" // Azure Key Vault
    _ "gocloud.dev/secrets/hashivault"   // HashiCorp Vault
    _ "gocloud.dev/secrets/localsecrets" // Local development
)

type PortableSecretsManager struct {
    keeper *secrets.Keeper
}

func NewPortableSecretsManager(ctx context.Context, url string) (*PortableSecretsManager, error) {
    // URL formats:
    // - awskms://alias/my-key?region=us-east-1
    // - gcpkms://projects/PROJECT/locations/LOCATION/keyRings/RING/cryptoKeys/KEY
    // - azurekeyvault://vault-name.vault.azure.net/keys/key-name
    // - hashivault://my-key
    // - base64key://smGbjm71Nxd1Ig5FS0wj9SlbzAIrnolCz9bQQ6uAhl4=

    keeper, err := secrets.OpenKeeper(ctx, url)
    if err != nil {
        return nil, err
    }

    return &PortableSecretsManager{keeper: keeper}, nil
}

func (m *PortableSecretsManager) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
    return m.keeper.Encrypt(ctx, plaintext)
}

func (m *PortableSecretsManager) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
    return m.keeper.Decrypt(ctx, ciphertext)
}

func (m *PortableSecretsManager) Close() error {
    return m.keeper.Close()
}
```

### Custom Provider Interface

```go
package secrets

import (
    "context"
    "fmt"
)

// SecretsProvider defines a portable interface for secrets management
type SecretsProvider interface {
    GetSecret(ctx context.Context, key string) (string, error)
    SetSecret(ctx context.Context, key, value string) error
    DeleteSecret(ctx context.Context, key string) error
    ListSecrets(ctx context.Context, prefix string) ([]string, error)
    Close() error
}

// ProviderFactory creates secrets providers based on configuration
type ProviderFactory struct {
    providers map[string]func(config map[string]string) (SecretsProvider, error)
}

func NewProviderFactory() *ProviderFactory {
    return &ProviderFactory{
        providers: make(map[string]func(config map[string]string) (SecretsProvider, error)),
    }
}

func (f *ProviderFactory) Register(name string, factory func(config map[string]string) (SecretsProvider, error)) {
    f.providers[name] = factory
}

func (f *ProviderFactory) Create(name string, config map[string]string) (SecretsProvider, error) {
    factory, ok := f.providers[name]
    if !ok {
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
    return factory(config)
}

// Usage
func SetupFactory() *ProviderFactory {
    factory := NewProviderFactory()

    factory.Register("aws", func(config map[string]string) (SecretsProvider, error) {
        return NewAWSSecretsManager(context.Background(), config["region"])
    })

    factory.Register("vault", func(config map[string]string) (SecretsProvider, error) {
        return NewVaultClient(config["address"], config["token"])
    })

    factory.Register("1password", func(config map[string]string) (SecretsProvider, error) {
        return NewOnePasswordManager(context.Background(), config["service_account_token"])
    })

    factory.Register("env", func(config map[string]string) (SecretsProvider, error) {
        return NewEnvProvider(), nil
    })

    return factory
}
```

### Multi-Source Secrets Manager

```go
package secrets

import (
    "context"
    "fmt"
    "sync"
)

// MultiSourceManager tries multiple providers in order
type MultiSourceManager struct {
    providers []SecretsProvider
    cache     sync.Map
}

func NewMultiSourceManager(providers ...SecretsProvider) *MultiSourceManager {
    return &MultiSourceManager{
        providers: providers,
    }
}

func (m *MultiSourceManager) GetSecret(ctx context.Context, key string) (string, error) {
    // Check cache first
    if cached, ok := m.cache.Load(key); ok {
        return cached.(string), nil
    }

    // Try each provider in order
    var lastErr error
    for _, provider := range m.providers {
        value, err := provider.GetSecret(ctx, key)
        if err == nil {
            m.cache.Store(key, value)
            return value, nil
        }
        lastErr = err
    }

    return "", fmt.Errorf("secret %s not found in any provider: %w", key, lastErr)
}

// Usage: Priority order for secret resolution
func Example() {
    manager := NewMultiSourceManager(
        NewEnvProvider(),           // 1. Check environment variables
        NewFileMountProvider(),     // 2. Check mounted files
        NewVaultProvider(),         // 3. Check Vault
        NewAWSSecretsManager(),     // 4. Check AWS Secrets Manager
    )

    secret, err := manager.GetSecret(context.Background(), "DATABASE_PASSWORD")
    if err != nil {
        log.Fatal(err)
    }
}
```

### References

- [Go CDK Secrets](https://gocloud.dev/howto/secrets/)
- [Go CDK Runtimevar](https://gocloud.dev/howto/runtimevar/)
- [infrahq/secrets Package](https://pkg.go.dev/github.com/infrahq/secrets)

---

## 12. Library Recommendations

### Core Libraries

| Library | Purpose | Recommendation |
|---------|---------|----------------|
| [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) | AWS Secrets Manager | **Production Ready** |
| [aws-secretsmanager-caching-go](https://github.com/aws/aws-secretsmanager-caching-go) | AWS SM Caching | **Highly Recommended** |
| [vault-client-go](https://github.com/hashicorp/vault-client-go) | HashiCorp Vault | **New Standard** |
| [onepassword-sdk-go](https://github.com/1Password/onepassword-sdk-go) | 1Password | **Official SDK** |
| [gocloud.dev/secrets](https://gocloud.dev/howto/secrets/) | Portable Abstraction | **Multi-Cloud** |

### Security Libraries

| Library | Purpose | Recommendation |
|---------|---------|----------------|
| [memguard](https://github.com/awnumar/memguard) | Secure Memory | **Essential for Sensitive Data** |
| [go-secure-stdlib/mlock](https://pkg.go.dev/github.com/hashicorp/go-secure-stdlib/mlock) | Memory Locking | **Vault-Grade Security** |
| crypto/subtle | Constant-Time Ops | **Standard Library** |

### Logging & Sanitization

| Library | Purpose | Recommendation |
|---------|---------|----------------|
| log/slog | Structured Logging | **Standard Library (Go 1.21+)** |
| [logfusc](https://github.com/AngusGMorrison/logfusc) | Auto-Redaction | **Simple & Effective** |
| [masq](https://github.com/m-mizutani/masq) | slog Redaction | **Flexible Patterns** |

### Caching

| Library | Purpose | Recommendation |
|---------|---------|----------------|
| [ttlcache](https://github.com/jellydator/ttlcache) | TTL Cache | **Modern, Generic** |
| [go-cache](https://github.com/patrickmn/go-cache) | Simple Cache | **Battle-Tested** |

### Kubernetes

| Tool | Purpose | Notes |
|------|---------|-------|
| [External Secrets Operator](https://external-secrets.io/) | Secret Sync | Industry Standard |
| [Stakater Reloader](https://github.com/stakater/Reloader) | Auto-Restart | Hot Reload |
| [Vault Agent](https://developer.hashicorp.com/vault/docs/agent-and-proxy/agent) | Sidecar | Dynamic Secrets |

---

## Performance Considerations

### Caching Strategy Matrix

| Scenario | TTL | Jitter | Strategy |
|----------|-----|--------|----------|
| API Keys | 1 hour | 10% | Refresh on access |
| DB Credentials | 15 min | 5 min | Periodic refresh |
| Feature Flags | 5 min | 1 min | Background refresh |
| Encryption Keys | 24 hours | 1 hour | Manual rotation |

### Benchmarking Template

```go
package secrets

import (
    "context"
    "testing"
)

func BenchmarkGetSecret(b *testing.B) {
    manager := setupTestManager()
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = manager.GetSecret(ctx, "test-secret")
    }
}

func BenchmarkGetSecretCached(b *testing.B) {
    manager := setupCachedManager()
    ctx := context.Background()

    // Warm cache
    manager.GetSecret(ctx, "test-secret")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = manager.GetSecret(ctx, "test-secret")
    }
}
```

---

## Security Recommendations Summary

1. **Never hardcode secrets** - Use secrets managers or environment variables
2. **Use constant-time comparison** - Prevent timing attacks with `crypto/subtle`
3. **Protect memory** - Use `memguard` for highly sensitive data
4. **Sanitize logs** - Implement `LogValuer` interface or use redaction libraries
5. **Enable audit logging** - Use secrets managers with built-in audit trails
6. **Rotate regularly** - Implement automated rotation with hot-reload
7. **Least privilege** - Grant minimal permissions to access secrets
8. **Separate environments** - Never share secrets between dev/staging/prod
9. **Cache wisely** - Balance security (short TTL) with performance
10. **Use mlock** - Prevent secrets from being swapped to disk

---

## Additional Resources

- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [GitGuardian: State of Secrets Sprawl 2024](https://www.gitguardian.com/state-of-secrets-sprawl-report-2024)
- [12 Factor App](https://12factor.net/)
- [NIST Cryptographic Key Management](https://csrc.nist.gov/projects/key-management)
