// Package studio provides studio automation tools for hg-mcp.
package studio

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewUNRAIDClient)

// Module implements the ToolModule interface for studio automation
type Module struct{}

func (m *Module) Name() string {
	return "studio"
}

func (m *Module) Description() string {
	return "Studio automation including UNRAID, network, and hardware management"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_unraid_status",
				mcp.WithDescription("Get UNRAID server status including array health and resource usage."),
			),
			Handler:             handleUNRAIDStatus,
			Category:            "studio",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "server", "status", "storage"},
			UseCases:            []string{"Check server health", "Monitor storage"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "studio",
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_vms",
				mcp.WithDescription("List virtual machines on UNRAID with their status."),
			),
			Handler:             handleUNRAIDVMs,
			Category:            "studio",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "vm", "virtualization"},
			UseCases:            []string{"List VMs", "Check VM status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "studio",
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_docker",
				mcp.WithDescription("List Docker containers on UNRAID with their status."),
			),
			Handler:             handleUNRAIDDocker,
			Category:            "studio",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "docker", "containers"},
			UseCases:            []string{"List containers", "Check container status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "studio",
		},
		{
			Tool: mcp.NewTool("aftrs_network_scan",
				mcp.WithDescription("Scan the studio network for devices."),
				mcp.WithString("subnet",
					mcp.Description("Subnet to scan (default: auto-detect)"),
				),
			),
			Handler:             handleNetworkScan,
			Category:            "studio",
			Subcategory:         "network",
			Tags:                []string{"network", "scan", "devices", "discovery"},
			UseCases:            []string{"Find network devices", "Check device connectivity"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "studio",
		},
		{
			Tool: mcp.NewTool("aftrs_hardware_status",
				mcp.WithDescription("Get aggregated hardware status including GPUs, capture cards, and audio interfaces."),
			),
			Handler:             handleHardwareStatus,
			Category:            "studio",
			Subcategory:         "hardware",
			Tags:                []string{"hardware", "gpu", "capture", "audio"},
			UseCases:            []string{"Check hardware health", "Pre-show verification"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "studio",
		},
		{
			Tool: mcp.NewTool("hairglasses_studio_health",
				mcp.WithDescription("Get comprehensive studio health score across all systems."),
			),
			Handler:             handleStudioHealth,
			Category:            "studio",
			Subcategory:         "health",
			Tags:                []string{"health", "studio", "overview", "consolidated"},
			UseCases:            []string{"Full health check", "Morning verification"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "studio",
		},
	}
}

// handleUNRAIDStatus handles the aftrs_unraid_status tool
func handleUNRAIDStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# UNRAID Server Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** ❌ Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Target:** %s\n\n", client.Host()))
		sb.WriteString("## Setup\n\n")
		sb.WriteString("Set environment variables:\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export UNRAID_HOST=tower.local\n")
		sb.WriteString("export UNRAID_API_KEY=your-api-key  # Optional\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** ✅ Connected\n")
	sb.WriteString(fmt.Sprintf("**Array:** %s\n", status.ArrayStatus))
	sb.WriteString(fmt.Sprintf("**Uptime:** %s\n\n", status.Uptime))

	sb.WriteString("## Resource Usage\n\n")
	sb.WriteString("| Resource | Usage |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| CPU | %.1f%% |\n", status.CPUUsage))
	sb.WriteString(fmt.Sprintf("| Memory | %.1f%% |\n", status.MemoryUsed))
	sb.WriteString(fmt.Sprintf("| Array | %.1f%% |\n", status.ArrayUsed))
	sb.WriteString(fmt.Sprintf("| Cache | %.1f%% |\n", status.CacheUsed))

	if status.Temperature > 0 {
		sb.WriteString(fmt.Sprintf("\n**Temperature:** %d°C\n", status.Temperature))
	}

	return tools.TextResult(sb.String()), nil
}

// handleUNRAIDVMs handles the aftrs_unraid_vms tool
func handleUNRAIDVMs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	vms, err := client.ListVMs(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# UNRAID Virtual Machines\n\n")

	if len(vms) == 0 {
		sb.WriteString("No virtual machines found.\n\n")
		sb.WriteString("This may indicate:\n")
		sb.WriteString("- No VMs configured on UNRAID\n")
		sb.WriteString("- UNRAID API not accessible\n")
		sb.WriteString("- Authentication required\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Name | State | CPU | Memory | AutoStart |\n")
	sb.WriteString("|------|-------|-----|--------|----------|\n")

	for _, vm := range vms {
		autostart := "No"
		if vm.AutoStart {
			autostart = "Yes"
		}
		stateEmoji := "⚫"
		if vm.State == "running" {
			stateEmoji = "🟢"
		} else if vm.State == "paused" {
			stateEmoji = "🟡"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s %s | %d | %d MB | %s |\n",
			vm.Name, stateEmoji, vm.State, vm.CPUCores, vm.MemoryMB, autostart))
	}

	return tools.TextResult(sb.String()), nil
}

// handleUNRAIDDocker handles the aftrs_unraid_docker tool
func handleUNRAIDDocker(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	containers, err := client.ListDockers(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# UNRAID Docker Containers\n\n")

	if len(containers) == 0 {
		sb.WriteString("No Docker containers found.\n\n")
		sb.WriteString("This may indicate:\n")
		sb.WriteString("- No containers configured on UNRAID\n")
		sb.WriteString("- Docker service not running\n")
		sb.WriteString("- API access required\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Name | Image | State | Uptime |\n")
	sb.WriteString("|------|-------|-------|--------|\n")

	for _, container := range containers {
		stateEmoji := "⚫"
		if container.State == "running" {
			stateEmoji = "🟢"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s %s | %s |\n",
			container.Name, container.Image, stateEmoji, container.State, container.Uptime))
	}

	return tools.TextResult(sb.String()), nil
}

// handleNetworkScan handles the aftrs_network_scan tool
func handleNetworkScan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	subnet := tools.GetStringParam(req, "subnet")
	if subnet == "" {
		subnet = "192.168.1.0/24" // Default
	}

	devices, err := clients.ScanNetwork(ctx, subnet)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Network Scan: %s\n\n", subnet))

	if len(devices) == 0 {
		sb.WriteString("No devices found.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Count online devices
	online := 0
	for _, d := range devices {
		if d.Online {
			online++
		}
	}

	sb.WriteString(fmt.Sprintf("Found **%d** devices (%d online):\n\n", len(devices), online))
	sb.WriteString("| Hostname | IP | Status |\n")
	sb.WriteString("|----------|----|---------|\n")

	for _, device := range devices {
		status := "🔴 Offline"
		if device.Online {
			status = "🟢 Online"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", device.Hostname, device.IP, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleHardwareStatus handles the aftrs_hardware_status tool
func handleHardwareStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status, err := clients.GetHardwareStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Hardware Status\n\n")

	statusEmoji := "✅"
	if status.Status == "degraded" {
		statusEmoji = "⚠️"
	} else if status.Status == "critical" {
		statusEmoji = "❌"
	}

	sb.WriteString(fmt.Sprintf("**Overall Score:** %d/100 %s\n", status.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", status.Status))

	sb.WriteString("## Devices\n\n")
	sb.WriteString("| Device | Status |\n")
	sb.WriteString("|--------|--------|\n")

	if len(status.Devices) == 0 {
		sb.WriteString("| (Hardware detection not implemented) | - |\n")
	} else {
		for name, info := range status.Devices {
			sb.WriteString(fmt.Sprintf("| %s | %v |\n", name, info))
		}
	}

	if len(status.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range status.Issues {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", issue))
		}
	}

	sb.WriteString("\n---\n")
	sb.WriteString("*Note: Full hardware detection requires platform-specific implementations.*\n")

	return tools.TextResult(sb.String()), nil
}

// handleStudioHealth handles the hairglasses_studio_health tool
func handleStudioHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	health, err := clients.GetStudioHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Studio Health Overview\n\n")

	statusEmoji := "✅"
	if health.Status == "degraded" {
		statusEmoji = "⚠️"
	} else if health.Status == "critical" {
		statusEmoji = "❌"
	}

	sb.WriteString(fmt.Sprintf("**Overall Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## Component Health\n\n")
	sb.WriteString("| Component | Score |\n")
	sb.WriteString("|-----------|-------|\n")

	for component, score := range health.Components {
		emoji := "✅"
		if score < 80 {
			emoji = "⚠️"
		}
		if score < 50 {
			emoji = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s | %d%% %s |\n", strings.Title(component), score, emoji))
	}

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", issue))
		}
	}

	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- 💡 %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
