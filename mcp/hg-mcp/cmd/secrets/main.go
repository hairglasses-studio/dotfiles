// Package main provides a CLI for testing the secrets manager.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hairglasses-studio/hg-mcp/pkg/secrets"
	"github.com/hairglasses-studio/hg-mcp/pkg/secrets/providers"
)

var (
	awsProfile   string
	awsRegion    string
	awsPrefix    string
	opVault      string
	opAccount    string
	envPrefix    string
	envFiles     []string
	outputFormat string
	cacheTTL     time.Duration
	enableAWS    bool
	enable1Pass  bool
	enableEnv    bool
	enableFile   bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "secrets",
		Short: "AFTRS secrets manager CLI",
		Long:  "CLI for testing and managing secrets from multiple providers (env, file, AWS, 1Password).",
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&awsProfile, "aws-profile", "cr8", "AWS profile to use")
	rootCmd.PersistentFlags().StringVar(&awsRegion, "aws-region", "us-east-1", "AWS region")
	rootCmd.PersistentFlags().StringVar(&awsPrefix, "aws-prefix", "hg-mcp/", "AWS secret name prefix")
	rootCmd.PersistentFlags().StringVar(&opVault, "op-vault", "", "1Password vault")
	rootCmd.PersistentFlags().StringVar(&opAccount, "op-account", "", "1Password account")
	rootCmd.PersistentFlags().StringVar(&envPrefix, "env-prefix", "", "Environment variable prefix")
	rootCmd.PersistentFlags().StringSliceVar(&envFiles, "env-file", []string{".env"}, "Environment files to load")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json")
	rootCmd.PersistentFlags().DurationVar(&cacheTTL, "cache-ttl", 5*time.Minute, "Cache TTL")
	rootCmd.PersistentFlags().BoolVar(&enableAWS, "aws", true, "Enable AWS Secrets Manager provider")
	rootCmd.PersistentFlags().BoolVar(&enable1Pass, "1password", true, "Enable 1Password provider")
	rootCmd.PersistentFlags().BoolVar(&enableEnv, "env", true, "Enable environment variable provider")
	rootCmd.PersistentFlags().BoolVar(&enableFile, "file", true, "Enable file provider")

	// Commands
	rootCmd.AddCommand(getCmd(), listCmd(), healthCmd(), whoamiCmd(), sanitizeCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// createManager creates a secrets manager with configured providers.
func createManager(ctx context.Context) (*secrets.Manager, error) {
	var providersList []secrets.SecretProvider

	// Environment provider (highest priority for local overrides)
	if enableEnv {
		envOpts := []providers.EnvOption{providers.WithEnvPriority(100)}
		if envPrefix != "" {
			envOpts = append(envOpts, providers.WithPrefix(envPrefix))
		}
		providersList = append(providersList, providers.NewEnvProvider(envOpts...))
	}

	// File provider
	if enableFile {
		fileOpts := []providers.FileOption{providers.WithFilePriority(200)}
		if len(envFiles) > 0 {
			fileOpts = append(fileOpts, providers.WithFiles(envFiles...))
		}
		providersList = append(providersList, providers.NewFileProvider(fileOpts...))
	}

	// 1Password provider
	if enable1Pass {
		opOpts := []providers.OnePasswordOption{providers.WithOnePasswordPriority(75)}
		if opVault != "" {
			opOpts = append(opOpts, providers.WithOnePasswordVault(opVault))
		}
		if opAccount != "" {
			opOpts = append(opOpts, providers.WithOnePasswordAccount(opAccount))
		}
		op, err := providers.NewOnePasswordProvider(opOpts...)
		if err == nil {
			providersList = append(providersList, op)
		}
	}

	// AWS provider (source of truth)
	if enableAWS {
		awsOpts := []providers.AWSOption{
			providers.WithAWSProfile(awsProfile),
			providers.WithAWSRegion(awsRegion),
			providers.WithAWSPriority(50),
		}
		if awsPrefix != "" {
			awsOpts = append(awsOpts, providers.WithAWSPrefix(awsPrefix))
		}
		aws, err := providers.NewAWSProvider(ctx, awsOpts...)
		if err == nil {
			providersList = append(providersList, aws)
		}
	}

	return secrets.NewManager(
		secrets.WithCacheTTL(cacheTTL),
		secrets.WithProviders(providersList...),
	), nil
}

// getCmd returns the get command.
func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a secret by key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			mgr, err := createManager(ctx)
			if err != nil {
				return err
			}
			defer mgr.Close()

			secret, err := mgr.Get(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to get secret: %w", err)
			}

			if outputFormat == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(secret)
			}

			fmt.Printf("Key:    %s\n", secret.Key)
			fmt.Printf("Value:  %s\n", secret.Value)
			fmt.Printf("Source: %s\n", secret.Source)
			if !secret.ExpiresAt.IsZero() {
				fmt.Printf("Expires: %s\n", secret.ExpiresAt.Format(time.RFC3339))
			}
			if secret.Version != "" {
				fmt.Printf("Version: %s\n", secret.Version)
			}
			return nil
		},
	}
}

// listCmd returns the list command.
func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available secret keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			mgr, err := createManager(ctx)
			if err != nil {
				return err
			}
			defer mgr.Close()

			keys, err := mgr.List(ctx)
			if err != nil {
				return fmt.Errorf("failed to list secrets: %w", err)
			}

			if outputFormat == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{"keys": keys, "count": len(keys)})
			}

			fmt.Printf("Found %d secrets:\n", len(keys))
			for _, key := range keys {
				fmt.Printf("  - %s\n", key)
			}
			return nil
		},
	}
}

// healthCmd returns the health command.
func healthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check health of all providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			mgr, err := createManager(ctx)
			if err != nil {
				return err
			}
			defer mgr.Close()

			health := mgr.Health(ctx)
			cacheStats := mgr.CacheStats()

			if outputFormat == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]any{"providers": health, "cache": cacheStats})
			}

			fmt.Println("Provider Health:")
			for _, h := range health {
				status := "OK"
				if !h.Available {
					status = "UNAVAILABLE"
				}
				fmt.Printf("  %s: %s (latency: %v)\n", h.Name, status, h.Latency)
				if h.Error != "" {
					fmt.Printf("    Error: %s\n", h.Error)
				}
			}
			fmt.Printf("\nCache:\n")
			fmt.Printf("  Size: %d entries\n", cacheStats.Size)
			fmt.Printf("  Valid: %d, Expired: %d\n", cacheStats.Valid, cacheStats.Expired)
			fmt.Printf("  TTL: %v\n", cacheStats.TTL)
			return nil
		},
	}
}

// whoamiCmd returns the whoami command.
func whoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current identity information",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			identity, err := secrets.Whoami(ctx)
			if err != nil {
				return fmt.Errorf("failed to get identity: %w", err)
			}

			if outputFormat == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(identity)
			}

			fmt.Print(identity.String())
			fmt.Println("\nAccess:")
			fmt.Printf("  AWS: %v\n", identity.HasAWSAccess())
			fmt.Printf("  GitHub: %v\n", identity.HasGitHubAccess())
			fmt.Printf("  Kubernetes: %v\n", identity.HasKubernetesAccess())
			fmt.Printf("  1Password: %v\n", identity.Has1PasswordAccess())
			return nil
		},
	}
}

// sanitizeCmd returns the sanitize command.
func sanitizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sanitize <text>",
		Short: "Sanitize sensitive content from text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sanitized := secrets.SanitizeString(args[0])
			fmt.Println(sanitized)
			return nil
		},
	}
}

// checkCmd checks if specific keys are sensitive.
func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <keys...>",
		Short: "Check if keys are considered sensitive",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results := make(map[string]bool)
			for _, key := range args {
				results[key] = secrets.IsSensitiveKey(key)
			}

			if outputFormat == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(results)
			}

			for key, sensitive := range results {
				status := "not sensitive"
				if sensitive {
					status = "SENSITIVE"
				}
				fmt.Printf("%s: %s\n", key, status)
			}
			return nil
		},
	}
}

// init adds additional commands that may be useful.
func init() {
	// Add banner
	fmt.Fprintln(os.Stderr, strings.Repeat("-", 50))
	fmt.Fprintln(os.Stderr, "AFTRS Secrets Manager CLI")
	fmt.Fprintln(os.Stderr, strings.Repeat("-", 50))
}
