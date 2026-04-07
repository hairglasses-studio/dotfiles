// Package prefetch provides an aggregated system context tool that agents can
// call once to get a snapshot of the host environment. This replaces multiple
// individual tool calls for hostname, uptime, load average, memory, disk, etc.
package prefetch

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements tools.ToolModule for the prefetch system context tool.
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "prefetch" }
func (m *Module) Description() string { return "Pre-fetch aggregated system context" }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_prefetch_system_context",
				mcp.WithDescription("Pre-fetch aggregated system context in a single call. Returns hostname, OS, kernel, uptime, load average, memory usage, disk usage, Go version, and active systemd services count. Use this at the start of a session to avoid multiple individual system queries."),
			),
			Handler:    handleSystemContext,
			Category:   "platform",
			Tags:       []string{"prefetch", "system", "context", "overview", "startup"},
			UseCases:   []string{"Session startup context", "System overview", "Environment discovery"},
			Complexity: tools.ComplexitySimple,
		},
	}
}

// systemContext holds the aggregated system state.
type systemContext struct {
	Hostname         string      `json:"hostname"`
	OS               string      `json:"os"`
	Arch             string      `json:"arch"`
	Kernel           string      `json:"kernel,omitempty"`
	Uptime           string      `json:"uptime,omitempty"`
	LoadAverage      string      `json:"load_average,omitempty"`
	Memory           *memoryInfo `json:"memory,omitempty"`
	Disk             []diskInfo  `json:"disk,omitempty"`
	GoVersion        string      `json:"go_version"`
	NumCPU           int         `json:"num_cpu"`
	NumGoroutine     int         `json:"num_goroutine"`
	ActiveServices   int         `json:"active_systemd_services"`
	FailedServices   int         `json:"failed_systemd_services"`
	CollectedAt      time.Time   `json:"collected_at"`
	CollectionTimeMs int64       `json:"collection_time_ms"`
}

type memoryInfo struct {
	TotalMB     int     `json:"total_mb"`
	UsedMB      int     `json:"used_mb"`
	AvailableMB int     `json:"available_mb"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotalMB int     `json:"swap_total_mb,omitempty"`
	SwapUsedMB  int     `json:"swap_used_mb,omitempty"`
}

type diskInfo struct {
	Filesystem  string `json:"filesystem"`
	MountPoint  string `json:"mount_point"`
	SizeGB      string `json:"size_gb"`
	UsedGB      string `json:"used_gb"`
	AvailGB     string `json:"avail_gb"`
	UsedPercent string `json:"used_percent"`
}

func handleSystemContext(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	start := time.Now()

	sc := &systemContext{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		CollectedAt:  start,
	}

	// Hostname
	if h, err := os.Hostname(); err == nil {
		sc.Hostname = h
	}

	// Kernel version
	sc.Kernel = runCmd(ctx, "uname", "-r")

	// Uptime (human-readable)
	sc.Uptime = runCmd(ctx, "uptime", "-p")

	// Load average
	if loadRaw := runCmd(ctx, "cat", "/proc/loadavg"); loadRaw != "" {
		// /proc/loadavg: "0.52 0.34 0.28 1/567 12345"
		parts := strings.Fields(loadRaw)
		if len(parts) >= 3 {
			sc.LoadAverage = strings.Join(parts[:3], " ")
		}
	}

	// Memory from /proc/meminfo
	sc.Memory = parseMemoryInfo(ctx)

	// Disk usage
	sc.Disk = parseDiskUsage(ctx)

	// Systemd services
	sc.ActiveServices = countSystemdUnits(ctx, "active")
	sc.FailedServices = countSystemdUnits(ctx, "failed")

	sc.CollectionTimeMs = time.Since(start).Milliseconds()

	return tools.JSONResult(sc), nil
}

// runCmd executes a command and returns its trimmed stdout, or empty string on error.
func runCmd(ctx context.Context, name string, args ...string) string {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// parseMemoryInfo reads /proc/meminfo and returns parsed memory stats.
func parseMemoryInfo(ctx context.Context) *memoryInfo {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil
	}

	values := make(map[string]int)
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		var val int
		if _, err := fmt.Sscanf(parts[1], "%d", &val); err == nil {
			values[key] = val // in kB
		}
	}

	total := values["MemTotal"]
	available := values["MemAvailable"]
	if total == 0 {
		return nil
	}

	used := total - available
	usedPct := float64(used) / float64(total) * 100

	mi := &memoryInfo{
		TotalMB:     total / 1024,
		UsedMB:      used / 1024,
		AvailableMB: available / 1024,
		UsedPercent: float64(int(usedPct*10)) / 10, // one decimal
	}

	if swapTotal := values["SwapTotal"]; swapTotal > 0 {
		swapFree := values["SwapFree"]
		mi.SwapTotalMB = swapTotal / 1024
		mi.SwapUsedMB = (swapTotal - swapFree) / 1024
	}

	return mi
}

// parseDiskUsage runs df and parses the output for real filesystems.
func parseDiskUsage(ctx context.Context) []diskInfo {
	out := runCmd(ctx, "df", "-h", "--output=source,size,used,avail,pcent,target", "-x", "tmpfs", "-x", "devtmpfs", "-x", "efivarfs", "-x", "squashfs")
	if out == "" {
		return nil
	}

	lines := strings.Split(out, "\n")
	if len(lines) < 2 {
		return nil
	}

	var disks []diskInfo
	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		disks = append(disks, diskInfo{
			Filesystem:  fields[0],
			SizeGB:      fields[1],
			UsedGB:      fields[2],
			AvailGB:     fields[3],
			UsedPercent: fields[4],
			MountPoint:  fields[5],
		})
	}
	return disks
}

// countSystemdUnits counts systemd service units in the given state.
func countSystemdUnits(ctx context.Context, state string) int {
	out := runCmd(ctx, "systemctl", "list-units", "--type=service", "--state="+state, "--no-legend", "--no-pager")
	if out == "" {
		return 0
	}
	return len(strings.Split(out, "\n"))
}
