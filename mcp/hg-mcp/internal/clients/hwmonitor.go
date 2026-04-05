// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HWMonitorClient provides hardware monitoring capabilities
type HWMonitorClient struct {
	alertThresholds AlertThresholds
	mu              sync.RWMutex
	lastReadings    map[string]TemperatureReading
}

// AlertThresholds defines temperature alert thresholds
type AlertThresholds struct {
	CPUWarning  float64 `json:"cpu_warning"`
	CPUCritical float64 `json:"cpu_critical"`
	GPUWarning  float64 `json:"gpu_warning"`
	GPUCritical float64 `json:"gpu_critical"`
}

// TemperatureReading represents a temperature reading
type TemperatureReading struct {
	Component   string    `json:"component"`
	Temperature float64   `json:"temperature_c"`
	Status      string    `json:"status"` // normal, warning, critical
	Timestamp   time.Time `json:"timestamp"`
}

// CPUInfo represents CPU information
type CPUInfo struct {
	Model       string               `json:"model"`
	Cores       int                  `json:"cores"`
	Threads     int                  `json:"threads"`
	Temperature float64              `json:"temperature_c"`
	Usage       float64              `json:"usage_percent"`
	Frequencies []float64            `json:"frequencies_mhz,omitempty"`
	PerCore     []TemperatureReading `json:"per_core,omitempty"`
}

// GPUInfo represents GPU information
type GPUInfo struct {
	Name        string  `json:"name"`
	Driver      string  `json:"driver"`
	Temperature float64 `json:"temperature_c"`
	Usage       float64 `json:"usage_percent"`
	MemoryTotal int64   `json:"memory_total_mb"`
	MemoryUsed  int64   `json:"memory_used_mb"`
	PowerDraw   float64 `json:"power_draw_w"`
	FanSpeed    int     `json:"fan_speed_percent"`
}

// PowerInfo represents power consumption information
type PowerInfo struct {
	CPUPower   float64 `json:"cpu_power_w"`
	GPUPower   float64 `json:"gpu_power_w"`
	TotalPower float64 `json:"total_power_w"`
	Estimated  bool    `json:"estimated"` // True if values are estimated
}

// ThermalAlert represents a thermal alert
type ThermalAlert struct {
	Component   string    `json:"component"`
	Temperature float64   `json:"temperature_c"`
	Threshold   float64   `json:"threshold_c"`
	Level       string    `json:"level"` // warning, critical
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// ThermalStatus represents overall thermal status
type ThermalStatus struct {
	Status   string               `json:"status"` // normal, warning, critical
	Alerts   []ThermalAlert       `json:"alerts,omitempty"`
	Readings []TemperatureReading `json:"readings"`
}

// NewHWMonitorClient creates a new hardware monitor client
func NewHWMonitorClient() (*HWMonitorClient, error) {
	thresholds := AlertThresholds{
		CPUWarning:  75.0,
		CPUCritical: 90.0,
		GPUWarning:  80.0,
		GPUCritical: 95.0,
	}

	// Override from environment
	if v := os.Getenv("CPU_TEMP_WARNING"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			thresholds.CPUWarning = f
		}
	}
	if v := os.Getenv("CPU_TEMP_CRITICAL"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			thresholds.CPUCritical = f
		}
	}
	if v := os.Getenv("GPU_TEMP_WARNING"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			thresholds.GPUWarning = f
		}
	}
	if v := os.Getenv("GPU_TEMP_CRITICAL"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			thresholds.GPUCritical = f
		}
	}

	return &HWMonitorClient{
		alertThresholds: thresholds,
		lastReadings:    make(map[string]TemperatureReading),
	}, nil
}

// GetCPUTemperature returns CPU temperature information
func (c *HWMonitorClient) GetCPUTemperature(ctx context.Context) (*CPUInfo, error) {
	info := &CPUInfo{
		Cores:   runtime.NumCPU(),
		Threads: runtime.NumCPU(),
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS - use osx-cpu-temp if available, otherwise estimate
		output, err := exec.CommandContext(ctx, "sysctl", "-n", "machdep.cpu.brand_string").Output()
		if err == nil {
			info.Model = strings.TrimSpace(string(output))
		}

		// Try osx-cpu-temp
		output, err = exec.CommandContext(ctx, "osx-cpu-temp").Output()
		if err == nil {
			// Parse "CPU: 45.0°C"
			parts := strings.Split(string(output), ":")
			if len(parts) >= 2 {
				tempStr := strings.TrimSpace(parts[1])
				tempStr = strings.Replace(tempStr, "°C", "", 1)
				if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
					info.Temperature = temp
				}
			}
		} else {
			// Fallback: use powermetrics (requires sudo) or estimate
			info.Temperature = -1 // Unknown
		}

	case "linux":
		// Linux - read from /sys/class/thermal or use lm-sensors
		output, err := exec.CommandContext(ctx, "cat", "/proc/cpuinfo").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "model name") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						info.Model = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}

		// Read temperature from thermal zones
		data, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
		if err == nil {
			if temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64); err == nil {
				info.Temperature = temp / 1000.0 // Convert from millidegrees
			}
		}

	case "windows":
		// Windows - use wmic or OpenHardwareMonitor
		output, err := exec.CommandContext(ctx, "wmic", "cpu", "get", "name").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) >= 2 {
				info.Model = strings.TrimSpace(lines[1])
			}
		}
		// Temperature requires third-party tools on Windows
		info.Temperature = -1
	}

	// Store reading
	c.mu.Lock()
	c.lastReadings["cpu"] = TemperatureReading{
		Component:   "CPU",
		Temperature: info.Temperature,
		Status:      c.getStatus(info.Temperature, c.alertThresholds.CPUWarning, c.alertThresholds.CPUCritical),
		Timestamp:   time.Now(),
	}
	c.mu.Unlock()

	return info, nil
}

// GetGPUTemperature returns GPU temperature information
func (c *HWMonitorClient) GetGPUTemperature(ctx context.Context) (*GPUInfo, error) {
	info := &GPUInfo{}

	switch runtime.GOOS {
	case "darwin":
		// macOS - check for Apple Silicon GPU or discrete GPU
		output, err := exec.CommandContext(ctx, "system_profiler", "SPDisplaysDataType").Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Chipset Model:") {
					info.Name = strings.TrimPrefix(line, "Chipset Model:")
					info.Name = strings.TrimSpace(info.Name)
				}
			}
		}
		// Temperature not easily available on macOS without third-party tools
		info.Temperature = -1

	case "linux", "windows":
		// Try nvidia-smi for NVIDIA GPUs
		output, err := exec.CommandContext(ctx, "nvidia-smi",
			"--query-gpu=name,driver_version,temperature.gpu,utilization.gpu,memory.total,memory.used,power.draw,fan.speed",
			"--format=csv,noheader,nounits").Output()
		if err == nil {
			parts := strings.Split(strings.TrimSpace(string(output)), ", ")
			if len(parts) >= 8 {
				info.Name = parts[0]
				info.Driver = parts[1]
				if temp, err := strconv.ParseFloat(parts[2], 64); err == nil {
					info.Temperature = temp
				}
				if usage, err := strconv.ParseFloat(parts[3], 64); err == nil {
					info.Usage = usage
				}
				if mem, err := strconv.ParseInt(parts[4], 10, 64); err == nil {
					info.MemoryTotal = mem
				}
				if mem, err := strconv.ParseInt(parts[5], 10, 64); err == nil {
					info.MemoryUsed = mem
				}
				if power, err := strconv.ParseFloat(parts[6], 64); err == nil {
					info.PowerDraw = power
				}
				if fan, err := strconv.Atoi(parts[7]); err == nil {
					info.FanSpeed = fan
				}
			}
		} else {
			// Try AMD ROCm for AMD GPUs
			output, err = exec.CommandContext(ctx, "rocm-smi", "--showtemp", "--csv").Output()
			if err == nil {
				lines := strings.Split(string(output), "\n")
				if len(lines) >= 2 {
					parts := strings.Split(lines[1], ",")
					if len(parts) >= 2 {
						if temp, err := strconv.ParseFloat(parts[1], 64); err == nil {
							info.Temperature = temp
						}
					}
				}
			}
		}
	}

	// Store reading
	c.mu.Lock()
	c.lastReadings["gpu"] = TemperatureReading{
		Component:   "GPU",
		Temperature: info.Temperature,
		Status:      c.getStatus(info.Temperature, c.alertThresholds.GPUWarning, c.alertThresholds.GPUCritical),
		Timestamp:   time.Now(),
	}
	c.mu.Unlock()

	return info, nil
}

// GetPowerConsumption returns power consumption information
func (c *HWMonitorClient) GetPowerConsumption(ctx context.Context) (*PowerInfo, error) {
	info := &PowerInfo{
		Estimated: true,
	}

	// Try to get GPU power from nvidia-smi
	output, err := exec.CommandContext(ctx, "nvidia-smi",
		"--query-gpu=power.draw",
		"--format=csv,noheader,nounits").Output()
	if err == nil {
		if power, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64); err == nil {
			info.GPUPower = power
			info.Estimated = false
		}
	}

	// CPU power is harder to get - estimate based on usage
	// Most desktop CPUs: 65-125W TDP
	// We'd need platform-specific tools like Intel RAPL
	if runtime.GOOS == "linux" {
		// Try reading from powercap/intel-rapl
		data, err := os.ReadFile("/sys/class/powercap/intel-rapl/intel-rapl:0/energy_uj")
		if err == nil {
			// This gives cumulative energy, would need two readings to calculate power
			_ = data // Placeholder for more complex implementation
		}
	}

	// Estimate CPU power (rough estimate)
	info.CPUPower = 50.0 // Default estimate
	info.TotalPower = info.CPUPower + info.GPUPower

	return info, nil
}

// GetThermalStatus returns overall thermal status with alerts
func (c *HWMonitorClient) GetThermalStatus(ctx context.Context) (*ThermalStatus, error) {
	status := &ThermalStatus{
		Status:   "normal",
		Readings: []TemperatureReading{},
	}

	// Get CPU temperature
	cpuInfo, _ := c.GetCPUTemperature(ctx)
	if cpuInfo != nil && cpuInfo.Temperature > 0 {
		reading := TemperatureReading{
			Component:   "CPU",
			Temperature: cpuInfo.Temperature,
			Status:      c.getStatus(cpuInfo.Temperature, c.alertThresholds.CPUWarning, c.alertThresholds.CPUCritical),
			Timestamp:   time.Now(),
		}
		status.Readings = append(status.Readings, reading)

		// Check for alerts
		if cpuInfo.Temperature >= c.alertThresholds.CPUCritical {
			status.Status = "critical"
			status.Alerts = append(status.Alerts, ThermalAlert{
				Component:   "CPU",
				Temperature: cpuInfo.Temperature,
				Threshold:   c.alertThresholds.CPUCritical,
				Level:       "critical",
				Message:     fmt.Sprintf("CPU temperature %.1f°C exceeds critical threshold %.1f°C", cpuInfo.Temperature, c.alertThresholds.CPUCritical),
				Timestamp:   time.Now(),
			})
		} else if cpuInfo.Temperature >= c.alertThresholds.CPUWarning {
			if status.Status != "critical" {
				status.Status = "warning"
			}
			status.Alerts = append(status.Alerts, ThermalAlert{
				Component:   "CPU",
				Temperature: cpuInfo.Temperature,
				Threshold:   c.alertThresholds.CPUWarning,
				Level:       "warning",
				Message:     fmt.Sprintf("CPU temperature %.1f°C exceeds warning threshold %.1f°C", cpuInfo.Temperature, c.alertThresholds.CPUWarning),
				Timestamp:   time.Now(),
			})
		}
	}

	// Get GPU temperature
	gpuInfo, _ := c.GetGPUTemperature(ctx)
	if gpuInfo != nil && gpuInfo.Temperature > 0 {
		reading := TemperatureReading{
			Component:   "GPU",
			Temperature: gpuInfo.Temperature,
			Status:      c.getStatus(gpuInfo.Temperature, c.alertThresholds.GPUWarning, c.alertThresholds.GPUCritical),
			Timestamp:   time.Now(),
		}
		status.Readings = append(status.Readings, reading)

		// Check for alerts
		if gpuInfo.Temperature >= c.alertThresholds.GPUCritical {
			status.Status = "critical"
			status.Alerts = append(status.Alerts, ThermalAlert{
				Component:   "GPU",
				Temperature: gpuInfo.Temperature,
				Threshold:   c.alertThresholds.GPUCritical,
				Level:       "critical",
				Message:     fmt.Sprintf("GPU temperature %.1f°C exceeds critical threshold %.1f°C", gpuInfo.Temperature, c.alertThresholds.GPUCritical),
				Timestamp:   time.Now(),
			})
		} else if gpuInfo.Temperature >= c.alertThresholds.GPUWarning {
			if status.Status != "critical" {
				status.Status = "warning"
			}
			status.Alerts = append(status.Alerts, ThermalAlert{
				Component:   "GPU",
				Temperature: gpuInfo.Temperature,
				Threshold:   c.alertThresholds.GPUWarning,
				Level:       "warning",
				Message:     fmt.Sprintf("GPU temperature %.1f°C exceeds warning threshold %.1f°C", gpuInfo.Temperature, c.alertThresholds.GPUWarning),
				Timestamp:   time.Now(),
			})
		}
	}

	return status, nil
}

// SetAlertThresholds updates the alert thresholds
func (c *HWMonitorClient) SetAlertThresholds(thresholds AlertThresholds) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.alertThresholds = thresholds
}

// GetAlertThresholds returns current alert thresholds
func (c *HWMonitorClient) GetAlertThresholds() AlertThresholds {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.alertThresholds
}

// getStatus determines status based on temperature and thresholds
func (c *HWMonitorClient) getStatus(temp, warning, critical float64) string {
	if temp < 0 {
		return "unknown"
	}
	if temp >= critical {
		return "critical"
	}
	if temp >= warning {
		return "warning"
	}
	return "normal"
}
