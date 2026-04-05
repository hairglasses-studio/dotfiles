package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	smtypes "github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/hairglasses-studio/hg-mcp/pkg/secrets"
)

// AWSProvider retrieves secrets from AWS Secrets Manager.
type AWSProvider struct {
	mu            sync.RWMutex
	client        *secretsmanager.Client
	stsClient     *sts.Client
	profile       string
	region        string
	prefix        string
	priority      int
	available     bool
	lastCheck     time.Time
	checkInterval time.Duration
}

// AWSOption configures the AWSProvider.
type AWSOption func(*AWSProvider)

// WithAWSProfile sets the AWS profile to use.
func WithAWSProfile(profile string) AWSOption {
	return func(p *AWSProvider) {
		p.profile = profile
	}
}

// WithAWSRegion sets the AWS region.
func WithAWSRegion(region string) AWSOption {
	return func(p *AWSProvider) {
		p.region = region
	}
}

// WithAWSPrefix sets a prefix for secret names (e.g., "hg-mcp/").
func WithAWSPrefix(prefix string) AWSOption {
	return func(p *AWSProvider) {
		p.prefix = prefix
	}
}

// WithAWSPriority sets the provider priority.
func WithAWSPriority(priority int) AWSOption {
	return func(p *AWSProvider) {
		p.priority = priority
	}
}

// NewAWSProvider creates a new AWS Secrets Manager provider.
func NewAWSProvider(ctx context.Context, opts ...AWSOption) (*AWSProvider, error) {
	p := &AWSProvider{
		profile:       "cr8", // Default to cr8 profile as per roadmap
		region:        "us-east-1",
		priority:      50, // Higher priority than env (AWS is source of truth)
		checkInterval: 5 * time.Minute,
	}
	for _, opt := range opts {
		opt(p)
	}

	// Build AWS config options
	configOpts := []func(*config.LoadOptions) error{
		config.WithRegion(p.region),
	}
	if p.profile != "" {
		configOpts = append(configOpts, config.WithSharedConfigProfile(p.profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	p.client = secretsmanager.NewFromConfig(cfg)
	p.stsClient = sts.NewFromConfig(cfg)

	// Check availability
	p.checkAvailability(ctx)

	return p, nil
}

// checkAvailability tests if AWS is accessible.
func (p *AWSProvider) checkAvailability(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := p.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	p.available = err == nil
	p.lastCheck = time.Now()
}

// Name returns the provider identifier.
func (p *AWSProvider) Name() string {
	return "aws"
}

// Get retrieves a secret from AWS Secrets Manager.
func (p *AWSProvider) Get(ctx context.Context, key string) (*secrets.Secret, error) {
	secretName := key
	if p.prefix != "" && !strings.HasPrefix(key, p.prefix) {
		secretName = p.prefix + key
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := p.client.GetSecretValue(ctx, input)
	if err != nil {
		var notFoundErr *smtypes.ResourceNotFoundException
		if isResourceNotFound(err, notFoundErr) {
			return nil, secrets.ErrSecretNotFound
		}
		return nil, fmt.Errorf("%w: %v", secrets.ErrProviderError, err)
	}

	value := ""
	if result.SecretString != nil {
		value = *result.SecretString

		// Try to parse as JSON and extract a simple value
		var jsonValue map[string]interface{}
		if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
			// If it's a JSON object with a single "value" or key matching the secret name
			if v, ok := jsonValue["value"]; ok {
				if s, ok := v.(string); ok {
					value = s
				}
			} else if v, ok := jsonValue[key]; ok {
				if s, ok := v.(string); ok {
					value = s
				}
			}
		}
	}

	secret := &secrets.Secret{
		Key:    key,
		Value:  value,
		Source: "aws:" + secretName,
	}

	if result.VersionId != nil {
		secret.Version = *result.VersionId
	}
	if result.CreatedDate != nil {
		secret.CreatedAt = *result.CreatedDate
	}

	return secret, nil
}

// GetJSON retrieves a secret and parses it as JSON into the provided target.
func (p *AWSProvider) GetJSON(ctx context.Context, key string, target interface{}) error {
	secret, err := p.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(secret.Value), target)
}

// List returns all secret names from AWS Secrets Manager.
func (p *AWSProvider) List(ctx context.Context) ([]string, error) {
	var keys []string
	var nextToken *string

	for {
		input := &secretsmanager.ListSecretsInput{
			NextToken:  nextToken,
			MaxResults: aws.Int32(100),
		}

		if p.prefix != "" {
			input.Filters = []smtypes.Filter{
				{
					Key:    smtypes.FilterNameStringTypeName,
					Values: []string{p.prefix},
				},
			}
		}

		result, err := p.client.ListSecrets(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", secrets.ErrProviderError, err)
		}

		for _, secret := range result.SecretList {
			if secret.Name != nil {
				name := *secret.Name
				if p.prefix != "" {
					name = strings.TrimPrefix(name, p.prefix)
				}
				keys = append(keys, name)
			}
		}

		nextToken = result.NextToken
		if nextToken == nil {
			break
		}
	}

	return keys, nil
}

// Exists checks if a secret exists in AWS Secrets Manager.
func (p *AWSProvider) Exists(ctx context.Context, key string) (bool, error) {
	secretName := key
	if p.prefix != "" && !strings.HasPrefix(key, p.prefix) {
		secretName = p.prefix + key
	}

	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretName),
	}

	_, err := p.client.DescribeSecret(ctx, input)
	if err != nil {
		var notFoundErr *smtypes.ResourceNotFoundException
		if isResourceNotFound(err, notFoundErr) {
			return false, nil
		}
		return false, fmt.Errorf("%w: %v", secrets.ErrProviderError, err)
	}

	return true, nil
}

// Priority returns the provider priority.
func (p *AWSProvider) Priority() int {
	return p.priority
}

// IsAvailable returns true if AWS is accessible.
func (p *AWSProvider) IsAvailable() bool {
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

// Health returns health status for the AWS provider.
func (p *AWSProvider) Health(ctx context.Context) secrets.ProviderHealth {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	identity, err := p.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})

	health := secrets.ProviderHealth{
		Name:      p.Name(),
		Available: err == nil,
		Latency:   time.Since(start),
		LastCheck: time.Now(),
	}

	if err != nil {
		health.Error = err.Error()
	} else if identity.Account != nil {
		health.Error = "" // Clear error, could add account info
	}

	p.mu.Lock()
	p.available = health.Available
	p.lastCheck = health.LastCheck
	p.mu.Unlock()

	return health
}

// GetCallerIdentity returns the AWS caller identity.
func (p *AWSProvider) GetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	return p.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
}

// Close is a no-op for the AWS provider.
func (p *AWSProvider) Close() error {
	return nil
}

// Set creates or updates a secret in AWS Secrets Manager.
func (p *AWSProvider) Set(ctx context.Context, key, value string, opts ...secrets.SetOption) error {
	options := &secrets.SetOptions{}
	for _, opt := range opts {
		opt(options)
	}

	secretName := key
	if p.prefix != "" && !strings.HasPrefix(key, p.prefix) {
		secretName = p.prefix + key
	}

	// Check if secret exists
	exists, _ := p.Exists(ctx, key)

	if exists {
		// Update existing secret
		input := &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(value),
		}
		_, err := p.client.PutSecretValue(ctx, input)
		if err != nil {
			return fmt.Errorf("%w: %v", secrets.ErrProviderError, err)
		}
	} else {
		// Create new secret
		input := &secretsmanager.CreateSecretInput{
			Name:         aws.String(secretName),
			SecretString: aws.String(value),
		}
		if options.Description != "" {
			input.Description = aws.String(options.Description)
		}
		if len(options.Tags) > 0 {
			for k, v := range options.Tags {
				input.Tags = append(input.Tags, smtypes.Tag{
					Key:   aws.String(k),
					Value: aws.String(v),
				})
			}
		}
		_, err := p.client.CreateSecret(ctx, input)
		if err != nil {
			return fmt.Errorf("%w: %v", secrets.ErrProviderError, err)
		}
	}

	return nil
}

// Delete removes a secret from AWS Secrets Manager.
func (p *AWSProvider) Delete(ctx context.Context, key string) error {
	secretName := key
	if p.prefix != "" && !strings.HasPrefix(key, p.prefix) {
		secretName = p.prefix + key
	}

	input := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: aws.Bool(false), // Allow recovery by default
	}

	_, err := p.client.DeleteSecret(ctx, input)
	if err != nil {
		var notFoundErr *smtypes.ResourceNotFoundException
		if isResourceNotFound(err, notFoundErr) {
			return secrets.ErrSecretNotFound
		}
		return fmt.Errorf("%w: %v", secrets.ErrProviderError, err)
	}

	return nil
}

// isResourceNotFound checks if an error is a ResourceNotFoundException.
func isResourceNotFound(err error, _ *smtypes.ResourceNotFoundException) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "ResourceNotFoundException") ||
		strings.Contains(errStr, "Secrets Manager can't find") ||
		strings.Contains(errStr, "not found")
}

// Ensure AWSProvider implements SecretProvider and WritableProvider.
var _ secrets.SecretProvider = (*AWSProvider)(nil)
var _ secrets.WritableProvider = (*AWSProvider)(nil)
