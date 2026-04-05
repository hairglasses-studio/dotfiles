// Package system provides MCP system monitoring and management tools.
package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for system tools
type Module struct{}

var getClient = tools.LazyClient(clients.NewSystemClient)

func (m *Module) Name() string {
	return "system"
}

func (m *Module) Description() string {
	return "Cross-platform system monitoring and management tools"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// === Disk & Storage (1) ===
		{
			Tool: mcp.NewTool("aftrs_system_disk_usage",
				mcp.WithDescription("Get disk space usage for specified paths or all mounted volumes."),
				mcp.WithString("paths", mcp.Description("Comma-separated paths to check (default: system drives)")),
			),
			Handler:             handleDiskUsage,
			Category:            "system",
			Subcategory:         "storage",
			Tags:                []string{"disk", "storage", "space", "usage"},
			UseCases:            []string{"Check free space", "Monitor disk usage"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "system",
		},
		// === Memory (1) ===
		{
			Tool: mcp.NewTool("aftrs_system_memory",
				mcp.WithDescription("Get memory (RAM) usage information including swap."),
			),
			Handler:             handleMemory,
			Category:            "system",
			Subcategory:         "memory",
			Tags:                []string{"memory", "ram", "swap", "usage"},
			UseCases:            []string{"Check memory usage", "Diagnose memory issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "system",
		},
		// === Processes (1) ===
		{
			Tool: mcp.NewTool("aftrs_system_processes",
				mcp.WithDescription("List top processes by CPU or memory usage."),
				mcp.WithNumber("limit", mcp.Description("Number of processes to show (default: 10)")),
				mcp.WithString("sort_by", mcp.Description("Sort by: cpu, memory (default: cpu)")),
			),
			Handler:             handleProcesses,
			Category:            "system",
			Subcategory:         "processes",
			Tags:                []string{"processes", "cpu", "memory", "top"},
			UseCases:            []string{"Find resource hogs", "Debug performance"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "system",
		},
		// === Thermal (1) ===
		{
			Tool: mcp.NewTool("aftrs_system_thermal",
				mcp.WithDescription("Get CPU/GPU temperatures and thermal throttling status."),
			),
			Handler:             handleThermal,
			Category:            "system",
			Subcategory:         "thermal",
			Tags:                []string{"temperature", "thermal", "cpu", "gpu", "throttling"},
			UseCases:            []string{"Monitor temps", "Check for throttling"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "system",
		},
		// === Battery (1) ===
		{
			Tool: mcp.NewTool("aftrs_system_battery",
				mcp.WithDescription("Get battery status and health information (laptops only)."),
			),
			Handler:             handleBattery,
			Category:            "system",
			Subcategory:         "power",
			Tags:                []string{"battery", "power", "laptop", "charging"},
			UseCases:            []string{"Check battery status", "Monitor battery health"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "system",
		},
		// === Cache Cleaning (1) ===
		{
			Tool: mcp.NewTool("aftrs_system_cache_clean",
				mcp.WithDescription("Clean developer caches (Go, npm, pip, Docker, Homebrew)."),
				mcp.WithString("categories", mcp.Description("Comma-separated: go, npm, pip, docker, homebrew (default: all)")),
			),
			Handler:             handleCacheClean,
			Category:            "system",
			Subcategory:         "cleanup",
			Tags:                []string{"cache", "clean", "cleanup", "disk", "space"},
			UseCases:            []string{"Free up disk space", "Clean dev caches"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "system",
			IsWrite:             true,
		},
		// === Docker Cleanup (1) ===
		{
			Tool: mcp.NewTool("aftrs_system_docker_prune",
				mcp.WithDescription("Clean up Docker resources (containers, images, volumes, networks)."),
				mcp.WithBoolean("all", mcp.Description("Remove all unused images and volumes (default: false)")),
			),
			Handler:             handleDockerPrune,
			Category:            "system",
			Subcategory:         "docker",
			Tags:                []string{"docker", "prune", "cleanup", "containers", "images"},
			UseCases:            []string{"Free Docker disk space", "Clean unused containers"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "system",
			IsWrite:             true,
		},
		// === System Overview (1) ===
		{
			Tool: mcp.NewTool("aftrs_system_overview",
				mcp.WithDescription("Get a comprehensive system status overview (disk, memory, CPU, thermal)."),
			),
			Handler:             handleOverview,
			Category:            "system",
			Subcategory:         "overview",
			Tags:                []string{"system", "status", "overview", "health"},
			UseCases:            []string{"Quick system check", "Health monitoring"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "system",
		},
	}
}

// === Handler Functions ===

func handleDiskUsage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathsStr := tools.GetStringParam(req, "paths")

	var paths []string
	if pathsStr != "" {
		for _, p := range strings.Split(pathsStr, ",") {
			paths = append(paths, strings.TrimSpace(p))
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	disks, err := client.GetDiskUsage(ctx, paths)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Disk Usage\n\n")

	if len(disks) == 0 {
		sb.WriteString("*No disk information available*\n")
	} else {
		sb.WriteString("| Path | Total | Used | Free | Used % |\n")
		sb.WriteString("|------|-------|------|------|--------|\n")
		for _, d := range disks {
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %.1f%% |\n",
				d.Path,
				formatBytes(d.Total),
				formatBytes(d.Used),
				formatBytes(d.Free),
				d.UsedPct,
			))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleMemory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	mem, err := client.GetMemoryInfo(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Memory Usage\n\n")

	sb.WriteString("## RAM\n")
	sb.WriteString(fmt.Sprintf("- **Total:** %s\n", formatBytes(mem.Total)))
	sb.WriteString(fmt.Sprintf("- **Used:** %s (%.1f%%)\n", formatBytes(mem.Used), mem.UsedPct))
	sb.WriteString(fmt.Sprintf("- **Free:** %s\n", formatBytes(mem.Free)))
	sb.WriteString(fmt.Sprintf("- **Available:** %s\n", formatBytes(mem.Available)))

	if mem.SwapTotal > 0 {
		sb.WriteString("\n## Swap\n")
		sb.WriteString(fmt.Sprintf("- **Total:** %s\n", formatBytes(mem.SwapTotal)))
		sb.WriteString(fmt.Sprintf("- **Used:** %s\n", formatBytes(mem.SwapUsed)))
		sb.WriteString(fmt.Sprintf("- **Free:** %s\n", formatBytes(mem.SwapFree)))
	}

	// Memory status indicator
	sb.WriteString("\n## Status\n")
	if mem.UsedPct > 90 {
		sb.WriteString("⚠️ **Critical:** Memory usage above 90%\n")
	} else if mem.UsedPct > 80 {
		sb.WriteString("⚠️ **Warning:** Memory usage above 80%\n")
	} else {
		sb.WriteString("✅ Memory usage normal\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleProcesses(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 10)
	sortBy := tools.GetStringParam(req, "sort_by")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	procs, err := client.GetTopProcesses(ctx, limit, sortBy)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Top Processes\n\n")

	if sortBy == "memory" {
		sb.WriteString("*Sorted by memory usage*\n\n")
	} else {
		sb.WriteString("*Sorted by CPU usage*\n\n")
	}

	if len(procs) == 0 {
		sb.WriteString("*No process information available*\n")
	} else {
		sb.WriteString("| PID | Name | CPU % | Memory % |\n")
		sb.WriteString("|-----|------|-------|----------|\n")
		for _, p := range procs {
			sb.WriteString(fmt.Sprintf("| %d | %s | %.1f | %.1f |\n",
				p.PID, p.Name, p.CPU, p.Memory))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleThermal(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	thermal, err := client.GetThermalInfo(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Thermal Status\n\n")

	if thermal.CPUTemp > 0 {
		sb.WriteString(fmt.Sprintf("**CPU Temperature:** %.1f°C", thermal.CPUTemp))
		if thermal.CPUTemp > 85 {
			sb.WriteString(" ⚠️ Hot")
		} else if thermal.CPUTemp > 75 {
			sb.WriteString(" ⚠️ Warm")
		} else {
			sb.WriteString(" ✅")
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("**CPU Temperature:** Not available\n")
	}

	if thermal.GPUTemp > 0 {
		sb.WriteString(fmt.Sprintf("**GPU Temperature:** %.1f°C", thermal.GPUTemp))
		if thermal.GPUTemp > 80 {
			sb.WriteString(" ⚠️ Hot")
		} else if thermal.GPUTemp > 70 {
			sb.WriteString(" ⚠️ Warm")
		} else {
			sb.WriteString(" ✅")
		}
		sb.WriteString("\n")
	}

	if thermal.Throttling {
		sb.WriteString("\n⚠️ **Thermal throttling detected!** CPU is running below maximum frequency.\n")
	} else {
		sb.WriteString("\n✅ No thermal throttling\n")
	}

	if thermal.FanSpeed > 0 {
		sb.WriteString(fmt.Sprintf("\n**Fan Speed:** %d RPM (%.0f%%)\n", thermal.FanSpeed, thermal.FanSpeedPct))
	}

	return tools.TextResult(sb.String()), nil
}

func handleBattery(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	battery, err := client.GetBatteryInfo(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Battery Status\n\n")

	if !battery.Present {
		sb.WriteString("*No battery detected (desktop system)*\n")
		return tools.TextResult(sb.String()), nil
	}

	// Battery icon based on level
	icon := "🔋"
	if battery.Charging {
		icon = "🔌"
	} else if battery.Percentage < 20 {
		icon = "🪫"
	}

	sb.WriteString(fmt.Sprintf("%s **Charge:** %.0f%%\n", icon, battery.Percentage))

	if battery.Charging {
		sb.WriteString("**Status:** Charging\n")
		if battery.TimeToFull != "" {
			sb.WriteString(fmt.Sprintf("**Time to full:** %s\n", battery.TimeToFull))
		}
	} else {
		sb.WriteString("**Status:** Discharging\n")
		if battery.TimeToEmpty != "" {
			sb.WriteString(fmt.Sprintf("**Time remaining:** %s\n", battery.TimeToEmpty))
		}
	}

	if battery.Health > 0 {
		sb.WriteString(fmt.Sprintf("**Health:** %.0f%%\n", battery.Health))
	}
	if battery.CycleCount > 0 {
		sb.WriteString(fmt.Sprintf("**Cycle count:** %d\n", battery.CycleCount))
	}

	return tools.TextResult(sb.String()), nil
}

func handleCacheClean(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	categoriesStr := tools.GetStringParam(req, "categories")

	var categories []string
	if categoriesStr != "" {
		for _, c := range strings.Split(categoriesStr, ",") {
			categories = append(categories, strings.TrimSpace(c))
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	results, err := client.CleanCaches(ctx, categories)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Cache Cleanup Results\n\n")

	var totalFreed uint64
	for _, r := range results {
		if r.Error != "" {
			sb.WriteString(fmt.Sprintf("## %s ❌\n", r.Category))
			sb.WriteString(fmt.Sprintf("Error: %s\n\n", r.Error))
		} else {
			sb.WriteString(fmt.Sprintf("## %s ✅\n", r.Category))
			sb.WriteString(fmt.Sprintf("- **Path:** `%s`\n", r.Path))
			sb.WriteString(fmt.Sprintf("- **Before:** %s\n", formatBytes(r.SizeBefore)))
			sb.WriteString(fmt.Sprintf("- **After:** %s\n", formatBytes(r.SizeAfter)))
			sb.WriteString(fmt.Sprintf("- **Freed:** %s\n\n", formatBytes(r.SpaceFreed)))
			totalFreed += r.SpaceFreed
		}
	}

	sb.WriteString(fmt.Sprintf("---\n**Total space freed:** %s\n", formatBytes(totalFreed)))

	return tools.TextResult(sb.String()), nil
}

func handleDockerPrune(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	all := tools.GetBoolParam(req, "all", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	output, err := client.DockerPrune(ctx, all)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Docker Prune Results\n\n")

	if all {
		sb.WriteString("*Removed all unused images and volumes*\n\n")
	} else {
		sb.WriteString("*Removed dangling images and stopped containers*\n\n")
	}

	sb.WriteString("```\n")
	sb.WriteString(output)
	sb.WriteString("```\n")

	return tools.TextResult(sb.String()), nil
}

func handleOverview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# System Overview\n\n")

	// Disk
	disks, _ := client.GetDiskUsage(ctx, nil)
	if len(disks) > 0 {
		sb.WriteString("## Disk\n")
		for _, d := range disks {
			status := "✅"
			if d.UsedPct > 90 {
				status = "🔴"
			} else if d.UsedPct > 80 {
				status = "🟡"
			}
			sb.WriteString(fmt.Sprintf("- %s `%s`: %s free (%.0f%% used)\n",
				status, d.Path, formatBytes(d.Free), d.UsedPct))
		}
		sb.WriteString("\n")
	}

	// Memory
	mem, _ := client.GetMemoryInfo(ctx)
	if mem != nil {
		sb.WriteString("## Memory\n")
		status := "✅"
		if mem.UsedPct > 90 {
			status = "🔴"
		} else if mem.UsedPct > 80 {
			status = "🟡"
		}
		sb.WriteString(fmt.Sprintf("- %s RAM: %s / %s (%.0f%% used)\n",
			status, formatBytes(mem.Used), formatBytes(mem.Total), mem.UsedPct))
		if mem.SwapTotal > 0 {
			swapPct := float64(mem.SwapUsed) / float64(mem.SwapTotal) * 100
			sb.WriteString(fmt.Sprintf("- Swap: %s / %s (%.0f%% used)\n",
				formatBytes(mem.SwapUsed), formatBytes(mem.SwapTotal), swapPct))
		}
		sb.WriteString("\n")
	}

	// Thermal
	thermal, _ := client.GetThermalInfo(ctx)
	if thermal != nil && (thermal.CPUTemp > 0 || thermal.GPUTemp > 0) {
		sb.WriteString("## Thermal\n")
		if thermal.CPUTemp > 0 {
			status := "✅"
			if thermal.CPUTemp > 85 {
				status = "🔴"
			} else if thermal.CPUTemp > 75 {
				status = "🟡"
			}
			sb.WriteString(fmt.Sprintf("- %s CPU: %.0f°C\n", status, thermal.CPUTemp))
		}
		if thermal.GPUTemp > 0 {
			status := "✅"
			if thermal.GPUTemp > 80 {
				status = "🔴"
			} else if thermal.GPUTemp > 70 {
				status = "🟡"
			}
			sb.WriteString(fmt.Sprintf("- %s GPU: %.0f°C\n", status, thermal.GPUTemp))
		}
		if thermal.Throttling {
			sb.WriteString("- ⚠️ Thermal throttling detected\n")
		}
		sb.WriteString("\n")
	}

	// Top processes
	procs, _ := client.GetTopProcesses(ctx, 5, "cpu")
	if len(procs) > 0 {
		sb.WriteString("## Top Processes (CPU)\n")
		for _, p := range procs {
			sb.WriteString(fmt.Sprintf("- %s: %.1f%% CPU, %.1f%% mem\n", p.Name, p.CPU, p.Memory))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// formatBytes formats bytes to human-readable format
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
