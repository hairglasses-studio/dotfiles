package tailscale

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getAPIClient = tools.LazyClient(clients.NewTailscaleAPIClient)

// handleCommission generates a pre-authorized auth key and install script for a new machine.
func handleCommission(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetOS := tools.OptionalStringParam(req, "os", "linux")
	if targetOS != "linux" && targetOS != "macos" {
		return tools.ErrorResult(fmt.Errorf("os must be 'linux' or 'macos', got %q", targetOS)), nil
	}

	hostname, errResult := tools.RequireStringParam(req, "hostname")
	if errResult != nil {
		return errResult, nil
	}

	tags := tools.GetStringArrayParam(req, "tags")
	ephemeral := tools.GetBoolParam(req, "ephemeral", false)
	enableSSH := tools.GetBoolParam(req, "ssh", true)

	apiClient, err := getAPIClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Tailscale API not available: %v", err)), nil
	}

	// Create a single-use, pre-authorized auth key
	keyOpts := clients.AuthKeyOpts{
		Reusable:      false,
		Ephemeral:     ephemeral,
		Preauthorized: true,
		Tags:          tags,
		Description:   fmt.Sprintf("commission-%s-%s", targetOS, hostname),
		ExpirySeconds: 3600, // 1 hour
	}

	keyResult, err := apiClient.CreateAuthKey(ctx, keyOpts)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("create auth key: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Commission: %s (%s)\n\n", hostname, targetOS))
	sb.WriteString(fmt.Sprintf("**Auth Key:** `%s`\n", keyResult.Key))
	sb.WriteString(fmt.Sprintf("**Key ID:** %s\n", keyResult.ID))
	sb.WriteString(fmt.Sprintf("**Expires:** %s\n", keyResult.Expires.Format("2006-01-02 15:04:05 MST")))
	sb.WriteString(fmt.Sprintf("**Ephemeral:** %v\n", ephemeral))
	if len(tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(tags, ", ")))
	}

	sb.WriteString("\n## Install Script\n\n")
	sb.WriteString("Copy and paste the following on the target machine:\n\n```bash\n")

	sshFlag := ""
	if enableSSH {
		sshFlag = " --ssh"
	}

	switch targetOS {
	case "linux":
		sb.WriteString("curl -fsSL https://tailscale.com/install.sh | sh\n")
		sb.WriteString(fmt.Sprintf("sudo tailscale up --authkey=%s%s --hostname=%s\n", keyResult.Key, sshFlag, hostname))
	case "macos":
		sb.WriteString("brew install --cask tailscale\n")
		sb.WriteString("open -a Tailscale\n")
		sb.WriteString("# Wait for Tailscale to start, then:\n")
		sb.WriteString(fmt.Sprintf("tailscale up --authkey=%s --accept-routes --hostname=%s\n", keyResult.Key, hostname))
		if enableSSH {
			sb.WriteString("# Note: Tailscale SSH is not available on macOS GUI builds.\n")
			sb.WriteString("# Enable standard SSH instead: System Settings → General → Sharing → Remote Login\n")
		}
	}

	sb.WriteString("```\n")

	sb.WriteString("\n## Verification\n\n")
	sb.WriteString("After running the script, verify with:\n```bash\n")
	sb.WriteString("tailscale status\n")
	sb.WriteString("tailscale ip -4\n")
	sb.WriteString("```\n")

	return tools.TextResult(sb.String()), nil
}

// handleCreateAuthKey creates a raw auth key with full options.
func handleCreateAuthKey(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiClient, err := getAPIClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Tailscale API not available: %v", err)), nil
	}

	opts := clients.AuthKeyOpts{
		Reusable:      tools.GetBoolParam(req, "reusable", false),
		Ephemeral:     tools.GetBoolParam(req, "ephemeral", false),
		Preauthorized: tools.GetBoolParam(req, "preauthorized", true),
		Tags:          tools.GetStringArrayParam(req, "tags"),
		Description:   tools.GetStringParam(req, "description"),
		ExpirySeconds: tools.GetIntParam(req, "expiry_hours", 24) * 3600,
	}

	result, err := apiClient.CreateAuthKey(ctx, opts)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("create auth key: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"id":      result.ID,
		"key":     result.Key,
		"expires": result.Expires.Format("2006-01-02 15:04:05 MST"),
		"options": map[string]interface{}{
			"reusable":      opts.Reusable,
			"ephemeral":     opts.Ephemeral,
			"preauthorized": opts.Preauthorized,
			"tags":          opts.Tags,
		},
	}), nil
}

// handleApprove authorizes a pending device on the tailnet.
func handleApprove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, errResult := tools.RequireStringParam(req, "device_id")
	if errResult != nil {
		return errResult, nil
	}

	apiClient, err := getAPIClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Tailscale API not available: %v", err)), nil
	}

	if err := apiClient.ApproveDevice(ctx, deviceID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Device %s approved successfully.", deviceID)), nil
}

// handleRemove removes a device from the tailnet.
func handleRemove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, errResult := tools.RequireStringParam(req, "device_id")
	if errResult != nil {
		return errResult, nil
	}

	confirm := tools.GetBoolParam(req, "confirm", false)
	if !confirm {
		return tools.ErrorResult(fmt.Errorf("set confirm=true to remove device %s", deviceID)), nil
	}

	apiClient, err := getAPIClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Tailscale API not available: %v", err)), nil
	}

	if err := apiClient.DeleteDevice(ctx, deviceID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Device %s removed from tailnet.", deviceID)), nil
}

// handleTag sets ACL tags on a device.
func handleTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID, errResult := tools.RequireStringParam(req, "device_id")
	if errResult != nil {
		return errResult, nil
	}

	tags := tools.GetStringArrayParam(req, "tags")
	if len(tags) == 0 {
		return tools.ErrorResult(fmt.Errorf("tags is required (array of ACL tags, e.g. [\"tag:server\"])")), nil
	}

	apiClient, err := getAPIClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Tailscale API not available: %v", err)), nil
	}

	if err := apiClient.SetDeviceTags(ctx, deviceID, tags); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Device %s tags set to: %s", deviceID, strings.Join(tags, ", "))), nil
}
