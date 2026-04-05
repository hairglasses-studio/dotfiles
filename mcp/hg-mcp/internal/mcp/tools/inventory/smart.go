package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SMARTData holds parsed smartctl output.
type SMARTData struct {
	Model          string `json:"model,omitempty"`
	Serial         string `json:"serial,omitempty"`
	Firmware       string `json:"firmware,omitempty"`
	HealthPassed   *bool  `json:"health_passed,omitempty"`
	TemperatureC   *int   `json:"temperature_c,omitempty"`
	PowerOnHours   *int   `json:"power_on_hours,omitempty"`
	PercentageUsed *int   `json:"percentage_used,omitempty"`
	PowerCycles    *int   `json:"power_cycles,omitempty"`

	// NVMe specific
	DataUnitsWritten *int `json:"data_units_written,omitempty"`
	DataUnitsRead    *int `json:"data_units_read,omitempty"`

	// SATA specific
	ReallocatedSectors *int `json:"reallocated_sectors,omitempty"`
	TotalLBAsWritten   *int `json:"total_lbas_written,omitempty"`
}

// CollectSMARTData runs smartctl on a device and parses the JSON output.
// Ported from tools/hw-resale/server.py:469-555
func CollectSMARTData(ctx context.Context, devicePath string) (*SMARTData, error) {
	if devicePath == "" {
		devicePath = "/dev/sda"
	}

	// Validate device path
	if !strings.HasPrefix(devicePath, "/dev/") {
		return nil, fmt.Errorf("invalid device path %q: must start with /dev/", devicePath)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "smartctl", "-a", "-j", devicePath)
	output, err := cmd.Output()
	if err != nil {
		// smartctl returns exit code 1 for warnings, which is OK
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() > 1 {
				return nil, fmt.Errorf("smartctl failed (exit %d): %s", exitErr.ExitCode(), string(exitErr.Stderr))
			}
			// Code 1 is warnings — continue with stdout
			output = append(output, exitErr.Stderr...)
		} else if strings.Contains(err.Error(), "executable file not found") {
			return nil, fmt.Errorf("smartctl not found — install smartmontools: sudo apt install smartmontools")
		} else {
			return nil, fmt.Errorf("smartctl error: %w", err)
		}
	}

	// Parse JSON
	var raw map[string]interface{}
	if err := json.Unmarshal(output, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse smartctl JSON: %w", err)
	}

	parsed := &SMARTData{}

	// Common fields
	if v, ok := raw["model_name"].(string); ok {
		parsed.Model = v
	}
	if v, ok := raw["serial_number"].(string); ok {
		parsed.Serial = v
	}
	if v, ok := raw["firmware_version"].(string); ok {
		parsed.Firmware = v
	}

	// Health
	if status, ok := raw["smart_status"].(map[string]interface{}); ok {
		if passed, ok := status["passed"].(bool); ok {
			parsed.HealthPassed = &passed
		}
	}

	// Temperature
	if temp, ok := raw["temperature"].(map[string]interface{}); ok {
		if v, ok := temp["current"].(float64); ok {
			iv := int(v)
			parsed.TemperatureC = &iv
		}
	}

	// Power-on hours
	if pot, ok := raw["power_on_time"].(map[string]interface{}); ok {
		if v, ok := pot["hours"].(float64); ok {
			iv := int(v)
			parsed.PowerOnHours = &iv
		}
	}

	// NVMe-specific
	if nvme, ok := raw["nvme_smart_health_information_log"].(map[string]interface{}); ok {
		if v, ok := nvme["percentage_used"].(float64); ok {
			iv := int(v)
			parsed.PercentageUsed = &iv
		}
		if v, ok := nvme["data_units_written"].(float64); ok {
			iv := int(v)
			parsed.DataUnitsWritten = &iv
		}
		if v, ok := nvme["data_units_read"].(float64); ok {
			iv := int(v)
			parsed.DataUnitsRead = &iv
		}
		if v, ok := nvme["power_cycles"].(float64); ok {
			iv := int(v)
			parsed.PowerCycles = &iv
		}
	}

	// SATA-specific attributes
	if ataAttrs, ok := raw["ata_smart_attributes"].(map[string]interface{}); ok {
		if table, ok := ataAttrs["table"].([]interface{}); ok {
			for _, entry := range table {
				attr, ok := entry.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := attr["name"].(string)
				rawVal, _ := attr["raw"].(map[string]interface{})
				value, _ := rawVal["value"].(float64)
				iv := int(value)

				switch name {
				case "Reallocated_Sector_Ct":
					parsed.ReallocatedSectors = &iv
				case "Power_On_Hours":
					parsed.PowerOnHours = &iv
				case "Temperature_Celsius":
					parsed.TemperatureC = &iv
				case "Total_LBAs_Written":
					parsed.TotalLBAsWritten = &iv
				}
			}
		}
	}

	return parsed, nil
}
