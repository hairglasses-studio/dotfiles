// Package tailscale provides Tailscale VPN management tools for hg-mcp.
package tailscale

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Tailscale
type Module struct{}

func (m *Module) Name() string {
	return "tailscale"
}

func (m *Module) Description() string {
	return "Tailscale VPN network management and device connectivity"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_tailscale_status",
				mcp.WithDescription("Get Tailscale network status and connected devices."),
			),
			Handler:             handleStatus,
			Category:            "tailscale",
			Subcategory:         "status",
			Tags:                []string{"tailscale", "vpn", "network", "status"},
			UseCases:            []string{"Check VPN status", "View connected devices"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tailscale",
		},
		{
			Tool: mcp.NewTool("aftrs_tailscale_devices",
				mcp.WithDescription("List all devices in the Tailscale network."),
				mcp.WithBoolean("online_only", mcp.Description("Show only online devices")),
			),
			Handler:             handleDevices,
			Category:            "tailscale",
			Subcategory:         "devices",
			Tags:                []string{"tailscale", "devices", "peers", "network"},
			UseCases:            []string{"List network devices", "Find device IPs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tailscale",
		},
		{
			Tool: mcp.NewTool("aftrs_tailscale_ping",
				mcp.WithDescription("Ping a device in the Tailscale network."),
				mcp.WithString("target", mcp.Required(), mcp.Description("Device hostname or IP to ping")),
			),
			Handler:             handlePing,
			Category:            "tailscale",
			Subcategory:         "network",
			Tags:                []string{"tailscale", "ping", "connectivity", "latency"},
			UseCases:            []string{"Test device connectivity", "Measure latency"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tailscale",
		},
		// --- Commissioning tools (API-based) ---
		{
			Tool: mcp.NewTool("aftrs_tailscale_commission",
				mcp.WithDescription("Generate a pre-authorized auth key and install script to commission a new machine onto the tailnet."),
				mcp.WithString("os", mcp.Description("Target OS: 'linux' or 'macos' (default: linux)")),
				mcp.WithString("hostname", mcp.Required(), mcp.Description("Hostname for the new device")),
				mcp.WithArray("tags", mcp.Description("ACL tags to apply (e.g. ['tag:server'])")),
				mcp.WithBoolean("ephemeral", mcp.Description("Auto-remove device when offline (default: false)")),
				mcp.WithBoolean("ssh", mcp.Description("Enable Tailscale SSH on the device (default: true)")),
			),
			Handler:             handleCommission,
			Category:            "tailscale",
			Subcategory:         "commissioning",
			Tags:                []string{"tailscale", "commission", "onboard", "auth-key", "install"},
			UseCases:            []string{"Add new machine to tailnet", "Generate install script"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "tailscale-api",
		},
		{
			Tool: mcp.NewTool("aftrs_tailscale_create_auth_key",
				mcp.WithDescription("Create a Tailscale auth key for device registration."),
				mcp.WithBoolean("reusable", mcp.Description("Allow multiple uses (default: false)")),
				mcp.WithBoolean("ephemeral", mcp.Description("Devices auto-removed when offline (default: false)")),
				mcp.WithBoolean("preauthorized", mcp.Description("Skip manual approval (default: true)")),
				mcp.WithArray("tags", mcp.Description("ACL tags for devices using this key")),
				mcp.WithString("description", mcp.Description("Human-readable label for the key")),
				mcp.WithNumber("expiry_hours", mcp.Description("Key lifetime in hours (default: 24)")),
			),
			Handler:             handleCreateAuthKey,
			Category:            "tailscale",
			Subcategory:         "commissioning",
			Tags:                []string{"tailscale", "auth-key", "registration"},
			UseCases:            []string{"Create auth key", "Device pre-authorization"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tailscale-api",
		},
		{
			Tool: mcp.NewTool("aftrs_tailscale_approve",
				mcp.WithDescription("Approve a pending device on the tailnet."),
				mcp.WithString("device_id", mcp.Required(), mcp.Description("Device ID to approve")),
			),
			Handler:             handleApprove,
			Category:            "tailscale",
			Subcategory:         "commissioning",
			Tags:                []string{"tailscale", "approve", "authorize", "device"},
			UseCases:            []string{"Approve pending device", "Authorize new node"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tailscale-api",
		},
		{
			Tool: mcp.NewTool("aftrs_tailscale_remove",
				mcp.WithDescription("Remove a device from the tailnet."),
				mcp.WithString("device_id", mcp.Required(), mcp.Description("Device ID to remove")),
				mcp.WithBoolean("confirm", mcp.Description("Must be true to confirm removal")),
			),
			Handler:             handleRemove,
			Category:            "tailscale",
			Subcategory:         "commissioning",
			Tags:                []string{"tailscale", "remove", "delete", "device"},
			UseCases:            []string{"Remove stale device", "Decommission node"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tailscale-api",
		},
		{
			Tool: mcp.NewTool("aftrs_tailscale_tag",
				mcp.WithDescription("Set ACL tags on a device in the tailnet."),
				mcp.WithString("device_id", mcp.Required(), mcp.Description("Device ID to tag")),
				mcp.WithArray("tags", mcp.Required(), mcp.Description("ACL tags to set (e.g. ['tag:server', 'tag:prod'])")),
			),
			Handler:             handleTag,
			Category:            "tailscale",
			Subcategory:         "commissioning",
			Tags:                []string{"tailscale", "tags", "acl", "device"},
			UseCases:            []string{"Tag device for ACL", "Classify device role"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tailscale-api",
		},
	}
}

var getClient = tools.LazyClient(clients.NewTailscaleClient)

// handleStatus handles the aftrs_tailscale_status tool
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Tailscale not available: %v", err)), nil
	}

	info, err := client.GetNetworkInfo(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tailscale Status\n\n")

	if info.Connected {
		sb.WriteString("**Status:** Connected\n\n")
	} else {
		sb.WriteString("**Status:** Disconnected\n\n")
		sb.WriteString("Run `tailscale up` to connect.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Network\n\n")
	if info.TailnetName != "" {
		sb.WriteString(fmt.Sprintf("**Tailnet:** %s\n", info.TailnetName))
	}
	sb.WriteString(fmt.Sprintf("**Magic DNS:** %s\n", info.MagicDNS))
	sb.WriteString(fmt.Sprintf("**This Device:** %s\n", info.SelfHostname))

	if len(info.SelfIPs) > 0 {
		sb.WriteString(fmt.Sprintf("**IP Address:** %s\n", info.SelfIPs[0]))
	}

	sb.WriteString("\n## Devices\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d\n", info.DeviceCount))
	sb.WriteString(fmt.Sprintf("**Online:** %d\n", info.OnlineCount))

	if info.ExitNode != "" {
		sb.WriteString(fmt.Sprintf("\n**Exit Node:** %s\n", info.ExitNode))
	}

	// Get version
	if version, err := client.GetVersion(ctx); err == nil {
		sb.WriteString(fmt.Sprintf("\n**Version:** %s\n", strings.Split(version, "\n")[0]))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDevices handles the aftrs_tailscale_devices tool
func handleDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Tailscale not available: %v", err)), nil
	}

	onlineOnly := tools.GetBoolParam(req, "online_only", false)

	devices, err := client.ListDevices(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tailscale Devices\n\n")

	if len(devices) == 0 {
		sb.WriteString("No devices found.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Filter if needed
	filtered := devices
	if onlineOnly {
		filtered = []clients.TailscaleDevice{}
		for _, d := range devices {
			if d.Online {
				filtered = append(filtered, d)
			}
		}
	}

	sb.WriteString(fmt.Sprintf("Found **%d** devices", len(filtered)))
	if onlineOnly {
		sb.WriteString(" (online only)")
	}
	sb.WriteString(":\n\n")

	sb.WriteString("| Hostname | OS | IP | Status | Connection |\n")
	sb.WriteString("|----------|----|----|--------|------------|\n")

	for _, device := range filtered {
		status := "Offline"
		if device.Online {
			status = "Online"
		}
		if device.IsCurrentNode {
			status = "This Device"
		}

		ip := "-"
		if len(device.IPs) > 0 {
			ip = device.IPs[0]
		}

		exitNode := ""
		if device.IsExitNode {
			exitNode = " (Exit Node)"
		}

		sb.WriteString(fmt.Sprintf("| %s%s | %s | %s | %s | %s |\n",
			device.Hostname, exitNode, device.OS, ip, status, device.Connection))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePing handles the aftrs_tailscale_ping tool
func handlePing(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target, errResult := tools.RequireStringParam(req, "target")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Tailscale not available: %v", err)), nil
	}

	result, err := client.Ping(ctx, target)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Ping: %s\n\n", target))

	if result.Success {
		sb.WriteString("**Status:** Success\n")
		if result.Latency != "" {
			sb.WriteString(fmt.Sprintf("**Latency:** %s\n", result.Latency))
		}
		sb.WriteString("\n## Output\n\n```\n")
		sb.WriteString(result.Output)
		sb.WriteString("```\n")
	} else {
		sb.WriteString("**Status:** Failed\n\n")
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", result.Error))
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
