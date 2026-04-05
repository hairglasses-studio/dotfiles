// Package hwmonitor provides hardware monitoring tools for hg-mcp.
package hwmonitor

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for hardware monitoring
type Module struct{}

func (m *Module) Name() string {
	return "hwmonitor"
}

func (m *Module) Description() string {
	return "Hardware monitoring for CPU, GPU temperature and power consumption"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_cpu_temperature",
				mcp.WithDescription("Get CPU temperature and information."),
			),
			Handler:             handleCPUTemperature,
			Category:            "hwmonitor",
			Subcategory:         "temperature",
			Tags:                []string{"cpu", "temperature", "monitoring", "hardware"},
			UseCases:            []string{"Monitor CPU heat", "Check thermal status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hwmonitor",
		},
		{
			Tool: mcp.NewTool("aftrs_gpu_temperature",
				mcp.WithDescription("Get GPU temperature and information."),
			),
			Handler:             handleGPUTemperature,
			Category:            "hwmonitor",
			Subcategory:         "temperature",
			Tags:                []string{"gpu", "temperature", "monitoring", "hardware", "nvidia", "amd"},
			UseCases:            []string{"Monitor GPU heat", "Check graphics card status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hwmonitor",
		},
		{
			Tool: mcp.NewTool("aftrs_power_consumption",
				mcp.WithDescription("Get power consumption information for CPU and GPU."),
			),
			Handler:             handlePowerConsumption,
			Category:            "hwmonitor",
			Subcategory:         "power",
			Tags:                []string{"power", "consumption", "watts", "energy"},
			UseCases:            []string{"Monitor power usage", "Check energy consumption"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hwmonitor",
		},
		{
			Tool: mcp.NewTool("aftrs_thermal_alert",
				mcp.WithDescription("Get thermal status and any active temperature alerts."),
				mcp.WithString("set_cpu_warning", mcp.Description("Set CPU warning threshold (Celsius)")),
				mcp.WithString("set_cpu_critical", mcp.Description("Set CPU critical threshold (Celsius)")),
				mcp.WithString("set_gpu_warning", mcp.Description("Set GPU warning threshold (Celsius)")),
				mcp.WithString("set_gpu_critical", mcp.Description("Set GPU critical threshold (Celsius)")),
			),
			Handler:             handleThermalAlert,
			Category:            "hwmonitor",
			Subcategory:         "alerts",
			Tags:                []string{"thermal", "alert", "warning", "critical", "temperature"},
			UseCases:            []string{"Check for overheating", "Configure thermal alerts"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "hwmonitor",
		},
	}
}

var getClient = tools.LazyClient(clients.NewHWMonitorClient)

// handleCPUTemperature handles the aftrs_cpu_temperature tool
func handleCPUTemperature(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.GetCPUTemperature(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# CPU Temperature\n\n")

	if info.Model != "" {
		sb.WriteString(fmt.Sprintf("**Model:** %s\n", info.Model))
	}
	sb.WriteString(fmt.Sprintf("**Cores:** %d\n", info.Cores))
	sb.WriteString(fmt.Sprintf("**Threads:** %d\n\n", info.Threads))

	if info.Temperature > 0 {
		sb.WriteString(fmt.Sprintf("**Temperature:** %.1f°C\n", info.Temperature))

		// Add status indicator
		thresholds := client.GetAlertThresholds()
		if info.Temperature >= thresholds.CPUCritical {
			sb.WriteString("**Status:** CRITICAL\n")
		} else if info.Temperature >= thresholds.CPUWarning {
			sb.WriteString("**Status:** Warning\n")
		} else {
			sb.WriteString("**Status:** Normal\n")
		}
	} else {
		sb.WriteString("**Temperature:** Not available\n")
		sb.WriteString("\n*Temperature monitoring may require additional tools:*\n")
		sb.WriteString("- macOS: Install `osx-cpu-temp` via Homebrew\n")
		sb.WriteString("- Linux: Ensure `lm-sensors` is installed\n")
		sb.WriteString("- Windows: Use OpenHardwareMonitor or similar\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleGPUTemperature handles the aftrs_gpu_temperature tool
func handleGPUTemperature(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.GetGPUTemperature(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# GPU Temperature\n\n")

	if info.Name != "" {
		sb.WriteString(fmt.Sprintf("**GPU:** %s\n", info.Name))
	}
	if info.Driver != "" {
		sb.WriteString(fmt.Sprintf("**Driver:** %s\n", info.Driver))
	}

	sb.WriteString("\n## Metrics\n\n")

	if info.Temperature > 0 {
		sb.WriteString(fmt.Sprintf("**Temperature:** %.1f°C\n", info.Temperature))

		thresholds := client.GetAlertThresholds()
		if info.Temperature >= thresholds.GPUCritical {
			sb.WriteString("**Status:** CRITICAL\n")
		} else if info.Temperature >= thresholds.GPUWarning {
			sb.WriteString("**Status:** Warning\n")
		} else {
			sb.WriteString("**Status:** Normal\n")
		}
	} else {
		sb.WriteString("**Temperature:** Not available\n")
	}

	if info.Usage > 0 {
		sb.WriteString(fmt.Sprintf("**GPU Usage:** %.1f%%\n", info.Usage))
	}

	if info.MemoryTotal > 0 {
		sb.WriteString(fmt.Sprintf("**Memory:** %d / %d MB (%.1f%%)\n",
			info.MemoryUsed, info.MemoryTotal,
			float64(info.MemoryUsed)/float64(info.MemoryTotal)*100))
	}

	if info.PowerDraw > 0 {
		sb.WriteString(fmt.Sprintf("**Power Draw:** %.1f W\n", info.PowerDraw))
	}

	if info.FanSpeed > 0 {
		sb.WriteString(fmt.Sprintf("**Fan Speed:** %d%%\n", info.FanSpeed))
	}

	if info.Name == "" && info.Temperature <= 0 {
		sb.WriteString("\n*GPU monitoring requires:*\n")
		sb.WriteString("- NVIDIA: `nvidia-smi` (comes with drivers)\n")
		sb.WriteString("- AMD: `rocm-smi` (ROCm toolkit)\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handlePowerConsumption handles the aftrs_power_consumption tool
func handlePowerConsumption(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.GetPowerConsumption(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Power Consumption\n\n")

	sb.WriteString("| Component | Power (W) |\n")
	sb.WriteString("|-----------|----------|\n")
	sb.WriteString(fmt.Sprintf("| CPU | %.1f |\n", info.CPUPower))
	sb.WriteString(fmt.Sprintf("| GPU | %.1f |\n", info.GPUPower))
	sb.WriteString(fmt.Sprintf("| **Total** | **%.1f** |\n", info.TotalPower))

	if info.Estimated {
		sb.WriteString("\n*Note: Some values are estimated. For accurate CPU power:*\n")
		sb.WriteString("- Linux: Intel RAPL (requires root)\n")
		sb.WriteString("- Windows: HWiNFO or similar\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleThermalAlert handles the aftrs_thermal_alert tool
func handleThermalAlert(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Check for threshold updates
	thresholds := client.GetAlertThresholds()
	updated := false

	if v := tools.GetStringParam(req, "set_cpu_warning"); v != "" {
		if f, err := parseFloat(v); err == nil {
			thresholds.CPUWarning = f
			updated = true
		}
	}
	if v := tools.GetStringParam(req, "set_cpu_critical"); v != "" {
		if f, err := parseFloat(v); err == nil {
			thresholds.CPUCritical = f
			updated = true
		}
	}
	if v := tools.GetStringParam(req, "set_gpu_warning"); v != "" {
		if f, err := parseFloat(v); err == nil {
			thresholds.GPUWarning = f
			updated = true
		}
	}
	if v := tools.GetStringParam(req, "set_gpu_critical"); v != "" {
		if f, err := parseFloat(v); err == nil {
			thresholds.GPUCritical = f
			updated = true
		}
	}

	if updated {
		client.SetAlertThresholds(thresholds)
	}

	// Get thermal status
	status, err := client.GetThermalStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Thermal Status\n\n")

	// Overall status with emoji
	switch status.Status {
	case "critical":
		sb.WriteString("**Status:** CRITICAL\n\n")
	case "warning":
		sb.WriteString("**Status:** Warning\n\n")
	default:
		sb.WriteString("**Status:** Normal\n\n")
	}

	// Current readings
	if len(status.Readings) > 0 {
		sb.WriteString("## Current Readings\n\n")
		sb.WriteString("| Component | Temperature | Status |\n")
		sb.WriteString("|-----------|-------------|--------|\n")

		for _, reading := range status.Readings {
			tempStr := fmt.Sprintf("%.1f°C", reading.Temperature)
			if reading.Temperature < 0 {
				tempStr = "N/A"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
				reading.Component, tempStr, reading.Status))
		}
	}

	// Active alerts
	if len(status.Alerts) > 0 {
		sb.WriteString("\n## Active Alerts\n\n")
		for _, alert := range status.Alerts {
			if alert.Level == "critical" {
				sb.WriteString(fmt.Sprintf("- **CRITICAL:** %s\n", alert.Message))
			} else {
				sb.WriteString(fmt.Sprintf("- **Warning:** %s\n", alert.Message))
			}
		}
	}

	// Show thresholds
	sb.WriteString("\n## Alert Thresholds\n\n")
	sb.WriteString("| Component | Warning | Critical |\n")
	sb.WriteString("|-----------|---------|----------|\n")
	sb.WriteString(fmt.Sprintf("| CPU | %.0f°C | %.0f°C |\n", thresholds.CPUWarning, thresholds.CPUCritical))
	sb.WriteString(fmt.Sprintf("| GPU | %.0f°C | %.0f°C |\n", thresholds.GPUWarning, thresholds.GPUCritical))

	if updated {
		sb.WriteString("\n*Thresholds updated.*\n")
	}

	return tools.TextResult(sb.String()), nil
}

// parseFloat parses a float from string
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
