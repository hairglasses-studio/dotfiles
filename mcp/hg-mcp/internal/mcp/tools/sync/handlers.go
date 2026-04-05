package sync

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	musicsync "github.com/hairglasses-studio/hg-mcp/internal/sync"
	"github.com/hairglasses-studio/mcpkit/sanitize"
)

func handleSyncAll(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := musicsync.DefaultConfig()

	// Parse parameters
	if user := tools.GetStringParam(request, "user"); user != "" {
		var filtered []musicsync.UserConfig
		for _, u := range config.Users {
			if u.Username == user {
				filtered = append(filtered, u)
				break
			}
		}
		if len(filtered) == 0 {
			return tools.ErrorResult(fmt.Errorf("user not found: %s", user)), nil
		}
		config.Users = filtered
	}

	// Dry run defaults to true
	config.DryRun = tools.GetBoolParam(request, "dry_run", true)

	manager, err := musicsync.NewManager(config)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create manager: %w", err)), nil
	}

	results, err := manager.SyncAll(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("sync failed: %w", err)), nil
	}

	return tools.JSONResult(results), nil
}

func handleSyncSoundCloud(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := musicsync.DefaultConfig()

	user := tools.GetStringParam(request, "user")
	if user == "" {
		user = "hairglasses"
	}

	config.DryRun = tools.GetBoolParam(request, "dry_run", true)

	manager, err := musicsync.NewManager(config)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create manager: %w", err)), nil
	}

	results, err := manager.SyncSoundCloud(ctx, user)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("soundcloud sync failed: %w", err)), nil
	}

	return tools.JSONResult(results), nil
}

func handleSyncBeatport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := musicsync.DefaultConfig()

	user := tools.GetStringParam(request, "user")
	if user == "" {
		user = "hairglasses"
	}

	config.DryRun = tools.GetBoolParam(request, "dry_run", true)

	manager, err := musicsync.NewManager(config)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create manager: %w", err)), nil
	}

	results, err := manager.SyncBeatport(ctx, user)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("beatport sync failed: %w", err)), nil
	}

	return tools.JSONResult(results), nil
}

func handleSyncRekordbox(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := musicsync.DefaultConfig()

	config.DryRun = tools.GetBoolParam(request, "dry_run", true)

	manager, err := musicsync.NewManager(config)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create manager: %w", err)), nil
	}

	result, err := manager.SyncRekordbox(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("rekordbox import failed: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleSyncStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := musicsync.DefaultConfig()

	manager, err := musicsync.NewManager(config)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create manager: %w", err)), nil
	}

	state, err := manager.Status()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(state), nil
}

func handleSyncHealth(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check for circuit breaker reset request
	resetService := tools.GetStringParam(request, "reset_circuit_breaker")
	var resetResults []string

	if resetService != "" {
		if resetService == "all" {
			// Reset all known circuit breakers
			for _, svc := range []string{"soundcloud", "beatport", "rekordbox"} {
				cb := musicsync.GlobalCircuitBreakers.Get(svc)
				cb.Reset()
				resetResults = append(resetResults, svc)
			}
		} else {
			// Reset specific service
			cb := musicsync.GlobalCircuitBreakers.Get(resetService)
			cb.Reset()
			resetResults = append(resetResults, resetService)
		}
	}

	config := musicsync.DefaultConfig()
	checker := musicsync.NewHealthChecker(config)
	healthResult := checker.Check(ctx)

	// Build response with optional reset info
	response := map[string]interface{}{
		"overall":          healthResult.Overall,
		"services":         healthResult.Services,
		"circuit_breakers": healthResult.CircuitBreakers,
	}

	if len(resetResults) > 0 {
		response["circuit_breakers_reset"] = resetResults
	}

	return tools.JSONResult(response), nil
}

func handleAddUser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, errResult := tools.RequireStringParam(request, "username")
	if errResult != nil {
		return errResult, nil
	}

	// Validate username to prevent command injection via S3 keys and URLs
	if err := sanitize.Username(username); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid username: %w", err)), nil
	}

	displayName := tools.GetStringParam(request, "display_name")
	if displayName == "" {
		displayName = username
	}

	download := tools.GetBoolParam(request, "download", true)

	// Create S3 folder structure
	config := musicsync.DefaultConfig()
	s3Base := fmt.Sprintf("s3://%s/users/%s/soundcloud/likes/", config.S3Bucket, username)
	cmd := exec.CommandContext(ctx, "aws", "s3api", "put-object", "--bucket", config.S3Bucket, "--key", fmt.Sprintf("users/%s/soundcloud/likes/", username), "--profile", config.AWSProfile)
	if err := cmd.Run(); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create S3 folder: %w", err)), nil
	}

	result := map[string]interface{}{
		"username":     username,
		"display_name": displayName,
		"s3_path":      s3Base,
		"status":       "created",
	}

	if download {
		// Start background download of likes and playlists
		go func() {
			likesCmd := exec.Command("scdl", "-l", fmt.Sprintf("https://soundcloud.com/%s/likes", username), "--path", fmt.Sprintf("/tmp/scdl/%s/likes", username), "--onlymp3")
			likesCmd.Run()
			// Upload to S3
			uploadCmd := exec.Command("aws", "s3", "sync", fmt.Sprintf("/tmp/scdl/%s/likes", username), fmt.Sprintf("s3://%s/users/%s/soundcloud/likes/", config.S3Bucket, username), "--profile", config.AWSProfile)
			uploadCmd.Run()
		}()
		result["download"] = "started in background"
	}

	return tools.JSONResult(result), nil
}

func handleListUsers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := musicsync.DefaultConfig()

	users := make([]map[string]interface{}, 0)
	for _, u := range config.Users {
		user := map[string]interface{}{
			"username":     u.Username,
			"display_name": u.DisplayName,
			"soundcloud":   u.SoundCloud,
			"beatport":     u.Beatport,
		}
		users = append(users, user)
	}

	return tools.JSONResult(users), nil
}

func handleDiscoverPlaylists(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, errResult := tools.RequireStringParam(request, "username")
	if errResult != nil {
		return errResult, nil
	}

	// Validate username to prevent command injection
	if err := sanitize.Username(username); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid username: %w", err)), nil
	}

	// Use scdl to list playlists
	cmd := exec.CommandContext(ctx, "scdl", "-l", fmt.Sprintf("https://soundcloud.com/%s/sets", username), "--list-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fall back to S3 discovery
		config := musicsync.DefaultConfig()
		s3Cmd := exec.CommandContext(ctx, "aws", "s3", "ls", fmt.Sprintf("s3://%s/users/%s/soundcloud/", config.S3Bucket, username), "--profile", config.AWSProfile)
		s3Output, s3Err := s3Cmd.Output()
		if s3Err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to discover playlists: %w", err)), nil
		}
		var playlists []string
		for _, line := range strings.Split(string(s3Output), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "PRE ") {
				name := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(line), "PRE "), "/")
				playlists = append(playlists, name)
			}
		}
		result := map[string]interface{}{"source": "s3", "playlists": playlists}
		return tools.JSONResult(result), nil
	}

	result := map[string]interface{}{"source": "soundcloud", "output": string(output)}
	return tools.JSONResult(result), nil
}
