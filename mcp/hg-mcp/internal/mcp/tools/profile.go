package tools

import (
	"os"
	"strings"
)

func hgToolProfile() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("HG_MCP_PROFILE"))) {
	case "", "default":
		return "default"
	case "ops":
		return "ops"
	case "full":
		return "full"
	default:
		return "default"
	}
}

func shouldDeferTool(profile string, td ToolDefinition) bool {
	switch profile {
	case "full":
		return false
	case "ops":
		return td.RuntimeGroup != RuntimeGroupPlatform &&
			td.RuntimeGroup != RuntimeGroupInfrastructure &&
			td.RuntimeGroup != RuntimeGroupShowControl
	default:
		if td.RuntimeGroup == RuntimeGroupPlatform {
			return false
		}
		switch td.Category {
		case "consolidated", "workflows", "workflow_automation", "studio", "dashboard", "gateway":
			return false
		default:
			return true
		}
	}
}
