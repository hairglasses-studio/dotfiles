// Package unraid provides UNRAID server management tools for hg-mcp.
package unraid

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewUNRAIDClient)

// Module implements the ToolModule interface for UNRAID integration
type Module struct{}

func (m *Module) Name() string {
	return "unraid"
}

func (m *Module) Description() string {
	return "UNRAID server management including Docker, VMs, and array control"
}

func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		// Status tools
		{
			Tool: mcp.NewTool("aftrs_unraid_status",
				mcp.WithDescription("Get UNRAID server status including array, CPU, memory, and disk usage."),
			),
			Handler:     handleStatus,
			Category:    "unraid",
			Subcategory: "status",
			Tags:        []string{"unraid", "status", "server", "health"},
			UseCases:    []string{"Check server status", "Monitor resources"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_disks",
				mcp.WithDescription("List array disks with status, temperature, and health."),
			),
			Handler:     handleDisks,
			Category:    "unraid",
			Subcategory: "array",
			Tags:        []string{"unraid", "disks", "array", "health"},
			UseCases:    []string{"Check disk health", "Monitor temperatures"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_shares",
				mcp.WithDescription("List user shares with usage information."),
			),
			Handler:     handleShares,
			Category:    "unraid",
			Subcategory: "array",
			Tags:        []string{"unraid", "shares", "storage"},
			UseCases:    []string{"Check share usage", "View storage allocation"},
			Complexity:  tools.ComplexitySimple,
		},
		// Docker management
		{
			Tool: mcp.NewTool("aftrs_docker_containers",
				mcp.WithDescription("List all Docker containers with status."),
			),
			Handler:     handleDockerContainers,
			Category:    "unraid",
			Subcategory: "docker",
			Tags:        []string{"docker", "containers", "list", "status"},
			UseCases:    []string{"View running containers", "Check container status"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_docker_start",
				mcp.WithDescription("Start a Docker container."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Container name to start")),
			),
			Handler:     handleDockerStart,
			Category:    "unraid",
			Subcategory: "docker",
			Tags:        []string{"docker", "start", "container"},
			UseCases:    []string{"Start stopped container", "Launch service"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_docker_stop",
				mcp.WithDescription("Stop a Docker container."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Container name to stop")),
			),
			Handler:     handleDockerStop,
			Category:    "unraid",
			Subcategory: "docker",
			Tags:        []string{"docker", "stop", "container"},
			UseCases:    []string{"Stop running container", "Shutdown service"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_docker_restart",
				mcp.WithDescription("Restart a Docker container."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Container name to restart")),
			),
			Handler:     handleDockerRestart,
			Category:    "unraid",
			Subcategory: "docker",
			Tags:        []string{"docker", "restart", "container"},
			UseCases:    []string{"Restart service", "Apply config changes"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_docker_logs",
				mcp.WithDescription("View Docker container logs."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Container name")),
				mcp.WithNumber("lines", mcp.Description("Number of lines (default 50)")),
			),
			Handler:     handleDockerLogs,
			Category:    "unraid",
			Subcategory: "docker",
			Tags:        []string{"docker", "logs", "container", "debug"},
			UseCases:    []string{"Debug container", "Check service logs"},
			Complexity:  tools.ComplexitySimple,
		},
		// VM management
		{
			Tool: mcp.NewTool("aftrs_vm_list",
				mcp.WithDescription("List virtual machines with status."),
			),
			Handler:     handleVMList,
			Category:    "unraid",
			Subcategory: "vm",
			Tags:        []string{"vm", "virtual", "list", "status"},
			UseCases:    []string{"View VMs", "Check VM status"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_vm_control",
				mcp.WithDescription("Control virtual machine (start/stop/pause/resume)."),
				mcp.WithString("name", mcp.Required(), mcp.Description("VM name")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, pause, resume, force_stop")),
			),
			Handler:     handleVMControl,
			Category:    "unraid",
			Subcategory: "vm",
			Tags:        []string{"vm", "control", "start", "stop"},
			UseCases:    []string{"Start/stop VMs", "Manage VM state"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		// Plugin management
		{
			Tool: mcp.NewTool("aftrs_unraid_plugins",
				mcp.WithDescription("List installed UNRAID plugins."),
			),
			Handler:     handlePlugins,
			Category:    "unraid",
			Subcategory: "plugins",
			Tags:        []string{"unraid", "plugins", "list"},
			UseCases:    []string{"View installed plugins", "Check for updates"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_diagnostics",
				mcp.WithDescription("Generate UNRAID diagnostics bundle."),
			),
			Handler:     handleDiagnostics,
			Category:    "unraid",
			Subcategory: "maintenance",
			Tags:        []string{"unraid", "diagnostics", "debug", "support"},
			UseCases:    []string{"Generate diagnostics", "Prepare support request"},
			Complexity:  tools.ComplexityModerate,
		},
		// Array operations
		{
			Tool: mcp.NewTool("aftrs_unraid_array_status",
				mcp.WithDescription("Get detailed array status including parity and mover."),
			),
			Handler:     handleArrayStatus,
			Category:    "unraid",
			Subcategory: "array",
			Tags:        []string{"unraid", "array", "parity", "status"},
			UseCases:    []string{"Check array health", "Monitor parity status"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_mover",
				mcp.WithDescription("Trigger the mover to transfer cache to array."),
			),
			Handler:     handleMover,
			Category:    "unraid",
			Subcategory: "maintenance",
			Tags:        []string{"unraid", "mover", "cache", "array"},
			UseCases:    []string{"Manual mover run", "Clear cache"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
	}

	// Apply circuit breaker to all tools — network-dependent (HTTP/SSH)
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "unraid"
	}

	return allTools
}

// getClient creates or returns the UNRAID client

// handleStatus handles the aftrs_unraid_status tool
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		sb.WriteString("**Status:** Disconnected\n\n")
		sb.WriteString(fmt.Sprintf("**Host:** %s\n\n", client.Host()))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("Configure UNRAID connection:\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export UNRAID_HOST=tower.local\n")
		sb.WriteString("export UNRAID_API_KEY=your-api-key\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**Host:** %s\n", client.Host()))
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
		sb.WriteString(fmt.Sprintf("| CPU Temp | %d°C |\n", status.Temperature))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDisks handles the aftrs_unraid_disks tool
func handleDisks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	disks, err := client.ListDisks(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# UNRAID Array Disks\n\n")

	if len(disks) == 0 {
		sb.WriteString("No disks found or array not started.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Name | Device | Size | Used | Temp | Health |\n")
	sb.WriteString("|------|--------|------|------|------|--------|\n")

	for _, d := range disks {
		sizeGB := float64(d.Size) / (1024 * 1024 * 1024)
		healthIcon := "✅"
		if d.Health == "warning" {
			healthIcon = "⚠️"
		} else if d.Health == "failed" {
			healthIcon = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %.0f GB | %.1f%% | %d°C | %s %s |\n",
			d.Name, d.Device, sizeGB, d.UsedPercent, d.Temperature, healthIcon, d.Health))
	}

	return tools.TextResult(sb.String()), nil
}

// handleShares handles the aftrs_unraid_shares tool
func handleShares(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	shares, err := client.ListShares(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# UNRAID User Shares\n\n")

	if len(shares) == 0 {
		sb.WriteString("No shares configured.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Name | Path | Used | Free | Security |\n")
	sb.WriteString("|------|------|------|------|----------|\n")

	for _, s := range shares {
		usedGB := float64(s.Used) / (1024 * 1024 * 1024)
		freeGB := float64(s.Free) / (1024 * 1024 * 1024)
		sb.WriteString(fmt.Sprintf("| %s | %s | %.1f GB | %.1f GB | %s |\n",
			s.Name, s.Path, usedGB, freeGB, s.Security))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDockerContainers handles the aftrs_docker_containers tool
func handleDockerContainers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	containers, err := client.ListDockers(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Docker Containers\n\n")

	if len(containers) == 0 {
		sb.WriteString("No containers found.\n\n")
		sb.WriteString("Ensure Docker is enabled on UNRAID and containers are configured.\n")
		return tools.TextResult(sb.String()), nil
	}

	running := 0
	for _, c := range containers {
		if c.State == "running" {
			running++
		}
	}

	sb.WriteString(fmt.Sprintf("**%d** containers (%d running):\n\n", len(containers), running))
	sb.WriteString("| Name | Image | State | Uptime | Network |\n")
	sb.WriteString("|------|-------|-------|--------|----------|\n")

	for _, c := range containers {
		stateIcon := "🔴"
		if c.State == "running" {
			stateIcon = "🟢"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s %s | %s | %s |\n",
			c.Name, c.Image, stateIcon, c.State, c.Uptime, c.Network))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDockerStart handles the aftrs_docker_start tool
func handleDockerStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.StartDocker(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Started container: %s", name)), nil
}

// handleDockerStop handles the aftrs_docker_stop tool
func handleDockerStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.StopDocker(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Stopped container: %s", name)), nil
}

// handleDockerRestart handles the aftrs_docker_restart tool
func handleDockerRestart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.RestartDocker(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Restarted container: %s", name)), nil
}

// handleDockerLogs handles the aftrs_docker_logs tool
func handleDockerLogs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	lines := tools.GetIntParam(req, "lines", 50)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	logs, err := client.GetDockerLogs(ctx, name, lines)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Logs: %s\n\n", name))
	sb.WriteString(fmt.Sprintf("Last %d lines:\n\n", lines))
	sb.WriteString("```\n")
	if logs == "" {
		sb.WriteString("(no logs available)\n")
	} else {
		sb.WriteString(logs)
	}
	sb.WriteString("```\n")

	return tools.TextResult(sb.String()), nil
}

// handleVMList handles the aftrs_vm_list tool
func handleVMList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	vms, err := client.ListVMs(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Virtual Machines\n\n")

	if len(vms) == 0 {
		sb.WriteString("No VMs configured.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Name | State | CPUs | Memory | Autostart |\n")
	sb.WriteString("|------|-------|------|--------|----------|\n")

	for _, vm := range vms {
		stateIcon := "🔴"
		if vm.State == "running" {
			stateIcon = "🟢"
		} else if vm.State == "paused" {
			stateIcon = "⏸️"
		}
		autostart := "No"
		if vm.AutoStart {
			autostart = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s %s | %d | %d MB | %s |\n",
			vm.Name, stateIcon, vm.State, vm.CPUCores, vm.MemoryMB, autostart))
	}

	return tools.TextResult(sb.String()), nil
}

// handleVMControl handles the aftrs_vm_control tool
func handleVMControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err = client.StartVM(ctx, name)
	case "stop":
		err = client.StopVM(ctx, name, false)
	case "force_stop":
		err = client.StopVM(ctx, name, true)
	case "pause":
		err = client.PauseVM(ctx, name)
	case "resume":
		err = client.ResumeVM(ctx, name)
	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use start, stop, force_stop, pause, resume)", action)), nil
	}

	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("VM %s: %s completed", name, action)), nil
}

// handlePlugins handles the aftrs_unraid_plugins tool
func handlePlugins(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	plugins, err := client.ListPlugins(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# UNRAID Plugins\n\n")

	if len(plugins) == 0 {
		sb.WriteString("No plugins installed or unable to fetch plugin list.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**%d** plugins installed:\n\n", len(plugins)))
	sb.WriteString("| Name | Version | Author | Update |\n")
	sb.WriteString("|------|---------|--------|--------|\n")

	for _, p := range plugins {
		update := ""
		if p.UpdateAvail {
			update = "⬆️ Available"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			p.Name, p.Version, p.Author, update))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDiagnostics handles the aftrs_unraid_diagnostics tool
func handleDiagnostics(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	path, err := client.GetDiagnostics(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# UNRAID Diagnostics\n\n")

	if path == "" {
		sb.WriteString("Diagnostics generation initiated.\n\n")
		sb.WriteString("The diagnostics bundle will be available in:\n")
		sb.WriteString("`/boot/logs/diagnostics-YYYYMMDD-HHMMSS.zip`\n")
	} else {
		sb.WriteString(fmt.Sprintf("Diagnostics bundle created: `%s`\n", path))
	}

	return tools.TextResult(sb.String()), nil
}

// handleArrayStatus handles the aftrs_unraid_array_status tool
func handleArrayStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetArrayStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# UNRAID Array Status\n\n")

	state := "unknown"
	if s, ok := status["state"].(string); ok {
		state = s
	}
	sb.WriteString(fmt.Sprintf("**State:** %s\n", state))

	if parity, ok := status["parity_valid"].(bool); ok {
		parityStatus := "✅ Valid"
		if !parity {
			parityStatus = "❌ Invalid"
		}
		sb.WriteString(fmt.Sprintf("**Parity:** %s\n", parityStatus))
	}

	if mover, ok := status["mover_active"].(bool); ok {
		moverStatus := "Idle"
		if mover {
			moverStatus = "Running"
		}
		sb.WriteString(fmt.Sprintf("**Mover:** %s\n", moverStatus))
	}

	return tools.TextResult(sb.String()), nil
}

// handleMover handles the aftrs_unraid_mover tool
func handleMover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.TriggerMover(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult("Mover triggered - transferring cache to array"), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
