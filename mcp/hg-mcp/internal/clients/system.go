// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/hairglasses-studio/mcpkit/sanitize"
)

// SystemClient provides cross-platform system utilities
type SystemClient struct {
	platform string
}

// DiskUsage represents disk space information
type DiskUsage struct {
	Path       string  `json:"path"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	Free       uint64  `json:"free"`
	UsedPct    float64 `json:"used_pct"`
	Filesystem string  `json:"filesystem,omitempty"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Free      uint64  `json:"free"`
	Available uint64  `json:"available"`
	UsedPct   float64 `json:"used_pct"`
	SwapTotal uint64  `json:"swap_total"`
	SwapUsed  uint64  `json:"swap_used"`
	SwapFree  uint64  `json:"swap_free"`
}

// ProcessInfo represents a running process
type ProcessInfo struct {
	PID     int     `json:"pid"`
	Name    string  `json:"name"`
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Command string  `json:"command,omitempty"`
}

// ThermalInfo represents thermal/temperature data
type ThermalInfo struct {
	CPUTemp     float64 `json:"cpu_temp,omitempty"`
	GPUTemp     float64 `json:"gpu_temp,omitempty"`
	Throttling  bool    `json:"throttling"`
	FanSpeed    int     `json:"fan_speed,omitempty"`
	FanSpeedPct float64 `json:"fan_speed_pct,omitempty"`
}

// BatteryInfo represents battery status (laptops)
type BatteryInfo struct {
	Present     bool    `json:"present"`
	Charging    bool    `json:"charging"`
	Percentage  float64 `json:"percentage"`
	TimeToEmpty string  `json:"time_to_empty,omitempty"`
	TimeToFull  string  `json:"time_to_full,omitempty"`
	Health      float64 `json:"health,omitempty"`
	CycleCount  int     `json:"cycle_count,omitempty"`
}

// CacheCleanResult represents cache cleaning results
type CacheCleanResult struct {
	Category     string `json:"category"`
	Path         string `json:"path"`
	SizeBefore   uint64 `json:"size_before"`
	SizeAfter    uint64 `json:"size_after"`
	SpaceFreed   uint64 `json:"space_freed"`
	ItemsRemoved int    `json:"items_removed"`
	Error        string `json:"error,omitempty"`
}

// NewSystemClient creates a new system client
func NewSystemClient() (*SystemClient, error) {
	return &SystemClient{
		platform: runtime.GOOS,
	}, nil
}

// GetDiskUsage returns disk usage for specified paths
func (c *SystemClient) GetDiskUsage(ctx context.Context, paths []string) ([]*DiskUsage, error) {
	if len(paths) == 0 {
		// Default paths based on platform
		switch c.platform {
		case "windows":
			paths = []string{"C:\\", "D:\\"}
		case "darwin":
			paths = []string{"/", "/Users"}
		default:
			paths = []string{"/", "/home"}
		}
	}

	var results []*DiskUsage
	for _, path := range paths {
		du, err := c.getDiskUsageSingle(ctx, path)
		if err != nil {
			continue // Skip inaccessible paths
		}
		results = append(results, du)
	}

	return results, nil
}

func (c *SystemClient) getDiskUsageSingle(ctx context.Context, path string) (*DiskUsage, error) {
	var cmd *exec.Cmd

	switch c.platform {
	case "windows":
		// Validate drive letter to prevent PowerShell injection
		driveName := strings.TrimSuffix(path, ":\\")
		if err := sanitize.DriveLetter(driveName); err != nil {
			return nil, fmt.Errorf("invalid drive path %q: %w", path, err)
		}
		ps := fmt.Sprintf(`Get-PSDrive -Name '%s' | Select-Object Used,Free | ConvertTo-Json`, driveName)
		cmd = exec.CommandContext(ctx, "powershell", "-Command", ps)
	default:
		cmd = exec.CommandContext(ctx, "df", "-k", path)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	du := &DiskUsage{Path: path}

	if c.platform == "windows" {
		// Parse PowerShell JSON output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Used") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val := strings.TrimSpace(strings.Trim(parts[1], ","))
					du.Used, _ = strconv.ParseUint(val, 10, 64)
				}
			}
			if strings.Contains(line, "Free") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val := strings.TrimSpace(strings.Trim(parts[1], ","))
					du.Free, _ = strconv.ParseUint(val, 10, 64)
				}
			}
		}
		du.Total = du.Used + du.Free
	} else {
		// Parse df output
		lines := strings.Split(string(output), "\n")
		if len(lines) >= 2 {
			fields := strings.Fields(lines[1])
			if len(fields) >= 4 {
				du.Filesystem = fields[0]
				total, _ := strconv.ParseUint(fields[1], 10, 64)
				used, _ := strconv.ParseUint(fields[2], 10, 64)
				free, _ := strconv.ParseUint(fields[3], 10, 64)
				du.Total = total * 1024
				du.Used = used * 1024
				du.Free = free * 1024
			}
		}
	}

	if du.Total > 0 {
		du.UsedPct = float64(du.Used) / float64(du.Total) * 100
	}

	return du, nil
}

// GetMemoryInfo returns memory usage information
func (c *SystemClient) GetMemoryInfo(ctx context.Context) (*MemoryInfo, error) {
	var cmd *exec.Cmd

	switch c.platform {
	case "windows":
		ps := `Get-CimInstance Win32_OperatingSystem | Select-Object TotalVisibleMemorySize,FreePhysicalMemory | ConvertTo-Json`
		cmd = exec.CommandContext(ctx, "powershell", "-Command", ps)
	case "darwin":
		cmd = exec.CommandContext(ctx, "vm_stat")
	default:
		cmd = exec.CommandContext(ctx, "free", "-b")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	mem := &MemoryInfo{}

	switch c.platform {
	case "windows":
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "TotalVisibleMemorySize") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val := strings.TrimSpace(strings.Trim(parts[1], ","))
					kb, _ := strconv.ParseUint(val, 10, 64)
					mem.Total = kb * 1024
				}
			}
			if strings.Contains(line, "FreePhysicalMemory") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val := strings.TrimSpace(strings.Trim(parts[1], ","))
					kb, _ := strconv.ParseUint(val, 10, 64)
					mem.Free = kb * 1024
				}
			}
		}
		mem.Used = mem.Total - mem.Free
		mem.Available = mem.Free

	case "darwin":
		// Parse vm_stat output
		pageSize := uint64(4096)
		var pagesFree, pagesActive, pagesInactive, pagesWired uint64
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Pages free:") {
				val := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "Pages free:"), "."))
				pagesFree, _ = strconv.ParseUint(val, 10, 64)
			} else if strings.HasPrefix(line, "Pages active:") {
				val := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "Pages active:"), "."))
				pagesActive, _ = strconv.ParseUint(val, 10, 64)
			} else if strings.HasPrefix(line, "Pages inactive:") {
				val := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "Pages inactive:"), "."))
				pagesInactive, _ = strconv.ParseUint(val, 10, 64)
			} else if strings.HasPrefix(line, "Pages wired down:") {
				val := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "Pages wired down:"), "."))
				pagesWired, _ = strconv.ParseUint(val, 10, 64)
			}
		}
		mem.Free = pagesFree * pageSize
		mem.Used = (pagesActive + pagesInactive + pagesWired) * pageSize
		mem.Total = mem.Free + mem.Used
		mem.Available = (pagesFree + pagesInactive) * pageSize

	default:
		// Parse free output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Mem:") {
				fields := strings.Fields(line)
				if len(fields) >= 7 {
					mem.Total, _ = strconv.ParseUint(fields[1], 10, 64)
					mem.Used, _ = strconv.ParseUint(fields[2], 10, 64)
					mem.Free, _ = strconv.ParseUint(fields[3], 10, 64)
					mem.Available, _ = strconv.ParseUint(fields[6], 10, 64)
				}
			} else if strings.HasPrefix(line, "Swap:") {
				fields := strings.Fields(line)
				if len(fields) >= 4 {
					mem.SwapTotal, _ = strconv.ParseUint(fields[1], 10, 64)
					mem.SwapUsed, _ = strconv.ParseUint(fields[2], 10, 64)
					mem.SwapFree, _ = strconv.ParseUint(fields[3], 10, 64)
				}
			}
		}
	}

	if mem.Total > 0 {
		mem.UsedPct = float64(mem.Used) / float64(mem.Total) * 100
	}

	return mem, nil
}

// GetTopProcesses returns top processes by resource usage
func (c *SystemClient) GetTopProcesses(ctx context.Context, limit int, sortBy string) ([]*ProcessInfo, error) {
	if limit == 0 {
		limit = 10
	}
	if sortBy == "" {
		sortBy = "cpu"
	}

	// Validate sortBy to prevent PowerShell injection
	sortAllowed := map[string]bool{"cpu": true, "memory": true, "workingset": true}
	if !sortAllowed[strings.ToLower(sortBy)] {
		return nil, fmt.Errorf("invalid sort field %q (allowed: cpu, memory, workingset)", sortBy)
	}

	var cmd *exec.Cmd

	switch c.platform {
	case "windows":
		ps := fmt.Sprintf(`Get-Process | Sort-Object -Property %s -Descending | Select-Object -First %d Id,ProcessName,CPU,WorkingSet | ConvertTo-Json`,
			strings.ToUpper(sortBy[:1])+sortBy[1:], limit)
		cmd = exec.CommandContext(ctx, "powershell", "-Command", ps)
	case "darwin":
		if sortBy == "memory" {
			cmd = exec.CommandContext(ctx, "ps", "aux", "-m")
		} else {
			cmd = exec.CommandContext(ctx, "ps", "aux", "-r")
		}
	default:
		if sortBy == "memory" {
			cmd = exec.CommandContext(ctx, "ps", "aux", "--sort=-%mem")
		} else {
			cmd = exec.CommandContext(ctx, "ps", "aux", "--sort=-%cpu")
		}
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var processes []*ProcessInfo

	if c.platform == "windows" {
		// Parse PowerShell JSON - simplified parsing
		lines := strings.Split(string(output), "\n")
		var currentProc *ProcessInfo
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "\"Id\":") {
				if currentProc != nil {
					processes = append(processes, currentProc)
				}
				currentProc = &ProcessInfo{}
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val := strings.TrimSpace(strings.Trim(parts[1], ","))
					currentProc.PID, _ = strconv.Atoi(val)
				}
			}
			if currentProc != nil {
				if strings.Contains(line, "\"ProcessName\":") {
					parts := strings.Split(line, ":")
					if len(parts) == 2 {
						currentProc.Name = strings.Trim(strings.TrimSpace(strings.Trim(parts[1], ",")), "\"")
					}
				}
				if strings.Contains(line, "\"CPU\":") {
					parts := strings.Split(line, ":")
					if len(parts) == 2 {
						val := strings.TrimSpace(strings.Trim(parts[1], ","))
						currentProc.CPU, _ = strconv.ParseFloat(val, 64)
					}
				}
				if strings.Contains(line, "\"WorkingSet\":") {
					parts := strings.Split(line, ":")
					if len(parts) == 2 {
						val := strings.TrimSpace(strings.Trim(parts[1], ","))
						ws, _ := strconv.ParseFloat(val, 64)
						currentProc.Memory = ws / 1024 / 1024 // Convert to MB
					}
				}
			}
		}
		if currentProc != nil {
			processes = append(processes, currentProc)
		}
	} else {
		// Parse ps output
		lines := strings.Split(string(output), "\n")
		for i, line := range lines {
			if i == 0 || line == "" {
				continue
			}
			if len(processes) >= limit {
				break
			}
			fields := strings.Fields(line)
			if len(fields) >= 11 {
				pid, _ := strconv.Atoi(fields[1])
				cpu, _ := strconv.ParseFloat(fields[2], 64)
				mem, _ := strconv.ParseFloat(fields[3], 64)
				processes = append(processes, &ProcessInfo{
					PID:     pid,
					Name:    fields[10],
					CPU:     cpu,
					Memory:  mem,
					Command: strings.Join(fields[10:], " "),
				})
			}
		}
	}

	return processes, nil
}

// GetThermalInfo returns thermal/temperature information
func (c *SystemClient) GetThermalInfo(ctx context.Context) (*ThermalInfo, error) {
	thermal := &ThermalInfo{}

	switch c.platform {
	case "windows":
		// Try to get GPU temp via nvidia-smi
		if gpuTemp, err := c.getNvidiaTemp(ctx); err == nil {
			thermal.GPUTemp = gpuTemp
		}
		// CPU temp requires WMI or third-party tools on Windows

	case "darwin":
		// Use osx-cpu-temp if available
		if out, err := exec.CommandContext(ctx, "osx-cpu-temp").Output(); err == nil {
			val := strings.TrimSpace(string(out))
			val = strings.TrimSuffix(val, "°C")
			thermal.CPUTemp, _ = strconv.ParseFloat(val, 64)
		}
		// Try to get GPU temp
		if gpuTemp, err := c.getNvidiaTemp(ctx); err == nil {
			thermal.GPUTemp = gpuTemp
		}

	default:
		// Linux - read from sysfs
		if data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp"); err == nil {
			tempMilli, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			thermal.CPUTemp = tempMilli / 1000
		}
		// Check for throttling
		if data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq"); err == nil {
			curFreq, _ := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
			if maxData, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq"); err == nil {
				maxFreq, _ := strconv.ParseUint(strings.TrimSpace(string(maxData)), 10, 64)
				thermal.Throttling = curFreq < maxFreq*90/100
			}
		}
		// GPU temp
		if gpuTemp, err := c.getNvidiaTemp(ctx); err == nil {
			thermal.GPUTemp = gpuTemp
		}
	}

	return thermal, nil
}

func (c *SystemClient) getNvidiaTemp(ctx context.Context) (float64, error) {
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=temperature.gpu", "--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

// GetBatteryInfo returns battery status (for laptops)
func (c *SystemClient) GetBatteryInfo(ctx context.Context) (*BatteryInfo, error) {
	battery := &BatteryInfo{}

	switch c.platform {
	case "windows":
		ps := `Get-CimInstance Win32_Battery | Select-Object BatteryStatus,EstimatedChargeRemaining,DesignCapacity,FullChargeCapacity | ConvertTo-Json`
		out, err := exec.CommandContext(ctx, "powershell", "-Command", ps).Output()
		if err != nil {
			return battery, nil // No battery
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "EstimatedChargeRemaining") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val := strings.TrimSpace(strings.Trim(parts[1], ","))
					battery.Percentage, _ = strconv.ParseFloat(val, 64)
					battery.Present = true
				}
			}
			if strings.Contains(line, "BatteryStatus") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val := strings.TrimSpace(strings.Trim(parts[1], ","))
					status, _ := strconv.Atoi(val)
					battery.Charging = status == 2
				}
			}
		}

	case "darwin":
		out, err := exec.CommandContext(ctx, "pmset", "-g", "batt").Output()
		if err != nil {
			return battery, nil
		}
		output := string(out)
		if strings.Contains(output, "InternalBattery") {
			battery.Present = true
			// Parse percentage
			if idx := strings.Index(output, "%"); idx > 0 {
				start := idx - 1
				for start > 0 && (output[start] >= '0' && output[start] <= '9') {
					start--
				}
				battery.Percentage, _ = strconv.ParseFloat(output[start+1:idx], 64)
			}
			battery.Charging = strings.Contains(output, "charging")
		}

	default:
		// Linux - check /sys/class/power_supply
		files, err := filepath.Glob("/sys/class/power_supply/BAT*/capacity")
		if err != nil || len(files) == 0 {
			return battery, nil
		}
		battery.Present = true
		if data, err := os.ReadFile(files[0]); err == nil {
			battery.Percentage, _ = strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		}
		statusPath := strings.Replace(files[0], "capacity", "status", 1)
		if data, err := os.ReadFile(statusPath); err == nil {
			status := strings.TrimSpace(string(data))
			battery.Charging = status == "Charging"
		}
	}

	return battery, nil
}

// CleanCaches cleans developer caches
func (c *SystemClient) CleanCaches(ctx context.Context, categories []string) ([]*CacheCleanResult, error) {
	if len(categories) == 0 {
		categories = []string{"go", "npm", "pip", "docker"}
	}

	var results []*CacheCleanResult
	homeDir, _ := os.UserHomeDir()

	for _, cat := range categories {
		result := &CacheCleanResult{Category: cat}

		switch cat {
		case "go":
			result.Path = filepath.Join(homeDir, "go", "pkg", "mod", "cache")
			result.SizeBefore = c.getDirSize(result.Path)
			cmd := exec.CommandContext(ctx, "go", "clean", "-modcache")
			if err := cmd.Run(); err != nil {
				result.Error = err.Error()
			}
			result.SizeAfter = c.getDirSize(result.Path)

		case "npm":
			// Get npm cache dir
			out, _ := exec.CommandContext(ctx, "npm", "config", "get", "cache").Output()
			result.Path = strings.TrimSpace(string(out))
			if result.Path == "" {
				result.Path = filepath.Join(homeDir, ".npm")
			}
			result.SizeBefore = c.getDirSize(result.Path)
			cmd := exec.CommandContext(ctx, "npm", "cache", "clean", "--force")
			if err := cmd.Run(); err != nil {
				result.Error = err.Error()
			}
			result.SizeAfter = c.getDirSize(result.Path)

		case "pip":
			result.Path = filepath.Join(homeDir, ".cache", "pip")
			result.SizeBefore = c.getDirSize(result.Path)
			cmd := exec.CommandContext(ctx, "pip", "cache", "purge")
			if err := cmd.Run(); err != nil {
				result.Error = err.Error()
			}
			result.SizeAfter = c.getDirSize(result.Path)

		case "docker":
			result.Path = "docker system"
			// Get docker disk usage before
			out, _ := exec.CommandContext(ctx, "docker", "system", "df", "--format", "{{.Size}}").Output()
			result.SizeBefore = c.parseDockerSize(string(out))
			cmd := exec.CommandContext(ctx, "docker", "system", "prune", "-f")
			if err := cmd.Run(); err != nil {
				result.Error = err.Error()
			}
			out, _ = exec.CommandContext(ctx, "docker", "system", "df", "--format", "{{.Size}}").Output()
			result.SizeAfter = c.parseDockerSize(string(out))

		case "homebrew":
			if c.platform == "darwin" {
				result.Path = "/usr/local/Cellar"
				result.SizeBefore = c.getDirSize(result.Path)
				cmd := exec.CommandContext(ctx, "brew", "cleanup", "-s")
				if err := cmd.Run(); err != nil {
					result.Error = err.Error()
				}
				result.SizeAfter = c.getDirSize(result.Path)
			}
		}

		result.SpaceFreed = result.SizeBefore - result.SizeAfter
		results = append(results, result)
	}

	return results, nil
}

func (c *SystemClient) getDirSize(path string) uint64 {
	var size uint64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	return size
}

func (c *SystemClient) parseDockerSize(output string) uint64 {
	// Parse docker size output like "1.5GB"
	output = strings.TrimSpace(output)
	var multiplier uint64 = 1
	if strings.HasSuffix(output, "GB") {
		multiplier = 1024 * 1024 * 1024
		output = strings.TrimSuffix(output, "GB")
	} else if strings.HasSuffix(output, "MB") {
		multiplier = 1024 * 1024
		output = strings.TrimSuffix(output, "MB")
	} else if strings.HasSuffix(output, "KB") {
		multiplier = 1024
		output = strings.TrimSuffix(output, "KB")
	}
	val, _ := strconv.ParseFloat(output, 64)
	return uint64(val * float64(multiplier))
}

// DockerPrune performs Docker cleanup
func (c *SystemClient) DockerPrune(ctx context.Context, all bool) (string, error) {
	var stderr bytes.Buffer

	args := []string{"system", "prune", "-f"}
	if all {
		args = append(args, "-a", "--volumes")
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}

	return string(out), nil
}
