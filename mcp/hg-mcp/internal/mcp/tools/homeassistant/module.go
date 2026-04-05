// Package homeassistant provides Home Assistant integration tools for hg-mcp.
package homeassistant

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Home Assistant
type Module struct{}

func (m *Module) Name() string {
	return "homeassistant"
}

func (m *Module) Description() string {
	return "Home Assistant smart home integration"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_hass_status",
				mcp.WithDescription("Get Home Assistant connection status."),
			),
			Handler:     handleStatus,
			Category:    "homeassistant",
			Subcategory: "status",
			Tags:        []string{"homeassistant", "hass", "status", "smarthome"},
			UseCases:    []string{"Check HA connection", "View HA status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "homeassistant",
		},
		{
			Tool: mcp.NewTool("aftrs_hass_entities",
				mcp.WithDescription("List Home Assistant entities."),
				mcp.WithString("domain", mcp.Description("Filter by domain (light, switch, sensor, etc.)")),
				mcp.WithString("search", mcp.Description("Search entities by name")),
			),
			Handler:     handleEntities,
			Category:    "homeassistant",
			Subcategory: "entities",
			Tags:        []string{"homeassistant", "entities", "devices", "list"},
			UseCases:    []string{"List smart devices", "Find entity IDs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "homeassistant",
		},
		{
			Tool: mcp.NewTool("aftrs_hass_control",
				mcp.WithDescription("Control a Home Assistant entity (turn on/off/toggle)."),
				mcp.WithString("entity_id", mcp.Required(), mcp.Description("Entity ID to control")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: on, off, toggle")),
				mcp.WithNumber("brightness", mcp.Description("Brightness for lights (0-255)")),
				mcp.WithString("color", mcp.Description("Color for lights (hex or name)")),
			),
			Handler:     handleControl,
			Category:    "homeassistant",
			Subcategory: "control",
			Tags:        []string{"homeassistant", "control", "switch", "light"},
			UseCases:    []string{"Control lights", "Toggle switches"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "homeassistant",
		},
		{
			Tool: mcp.NewTool("aftrs_hass_scenes",
				mcp.WithDescription("List and activate Home Assistant scenes."),
				mcp.WithString("activate", mcp.Description("Scene ID to activate")),
			),
			Handler:     handleScenes,
			Category:    "homeassistant",
			Subcategory: "scenes",
			Tags:        []string{"homeassistant", "scenes", "presets", "automation"},
			UseCases:    []string{"List scenes", "Activate scene"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "homeassistant",
		},
		{
			Tool: mcp.NewTool("aftrs_hass_automations",
				mcp.WithDescription("List and trigger Home Assistant automations."),
				mcp.WithString("trigger", mcp.Description("Automation ID to trigger")),
			),
			Handler:     handleAutomations,
			Category:    "homeassistant",
			Subcategory: "automations",
			Tags:        []string{"homeassistant", "automations", "trigger", "workflows"},
			UseCases:    []string{"List automations", "Trigger automation"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "homeassistant",
		},
	}
}

var getClient = tools.LazyClient(clients.GetHomeAssistantClient)

// handleStatus handles the aftrs_hass_status tool
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Home Assistant not configured: %v", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Home Assistant Status\n\n")

	if status.Connected {
		sb.WriteString("**Status:** Connected\n")
		sb.WriteString(fmt.Sprintf("**Version:** %s\n", status.Version))
		sb.WriteString(fmt.Sprintf("**URL:** %s\n", status.BaseURL))
		sb.WriteString(fmt.Sprintf("**Entities:** %d\n", status.EntityCount))
		if status.LocationName != "" {
			sb.WriteString(fmt.Sprintf("**Location:** %s\n", status.LocationName))
		}
	} else {
		sb.WriteString("**Status:** Disconnected\n\n")
		sb.WriteString("Ensure HASS_URL and HASS_TOKEN environment variables are set.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleEntities handles the aftrs_hass_entities tool
func handleEntities(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Home Assistant not configured: %v", err)), nil
	}

	domain := tools.GetStringParam(req, "domain")
	search := strings.ToLower(tools.GetStringParam(req, "search"))

	entities, err := client.GetEntities(ctx, domain)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Filter by search term
	if search != "" {
		filtered := []clients.HAEntity{}
		for _, e := range entities {
			name := e.EntityID
			if friendlyName, ok := e.Attributes["friendly_name"].(string); ok {
				name = friendlyName
			}
			if strings.Contains(strings.ToLower(e.EntityID), search) ||
				strings.Contains(strings.ToLower(name), search) {
				filtered = append(filtered, e)
			}
		}
		entities = filtered
	}

	var sb strings.Builder
	sb.WriteString("# Home Assistant Entities\n\n")

	if len(entities) == 0 {
		sb.WriteString("No entities found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** entities", len(entities)))
	if domain != "" {
		sb.WriteString(fmt.Sprintf(" in domain `%s`", domain))
	}
	sb.WriteString(":\n\n")

	// Group by domain
	byDomain := make(map[string][]clients.HAEntity)
	for _, e := range entities {
		d := "unknown"
		if idx := strings.Index(e.EntityID, "."); idx > 0 {
			d = e.EntityID[:idx]
		}
		byDomain[d] = append(byDomain[d], e)
	}

	for d, ents := range byDomain {
		sb.WriteString(fmt.Sprintf("## %s (%d)\n\n", strings.Title(d), len(ents)))
		sb.WriteString("| Entity ID | Name | State |\n")
		sb.WriteString("|-----------|------|-------|\n")

		for _, e := range ents {
			name := "-"
			if friendlyName, ok := e.Attributes["friendly_name"].(string); ok {
				name = friendlyName
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", e.EntityID, name, e.State))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleControl handles the aftrs_hass_control tool
func handleControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	entityID, errResult := tools.RequireStringParam(req, "entity_id")
	if errResult != nil {
		return errResult, nil
	}

	actionRaw, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}
	action := strings.ToLower(actionRaw)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Home Assistant not configured: %v", err)), nil
	}

	// Build service data
	data := make(map[string]interface{})

	if brightness := tools.GetIntParam(req, "brightness", 0); brightness > 0 {
		data["brightness"] = brightness
	}
	if color := tools.GetStringParam(req, "color"); color != "" {
		if strings.HasPrefix(color, "#") {
			data["rgb_color"] = hexToRGB(color)
		} else {
			data["color_name"] = color
		}
	}

	var actionErr error
	switch action {
	case "on":
		actionErr = client.TurnOn(ctx, entityID, data)
	case "off":
		actionErr = client.TurnOff(ctx, entityID)
	case "toggle":
		actionErr = client.Toggle(ctx, entityID)
	default:
		return tools.ErrorResult(fmt.Errorf("unknown action: %s (use on, off, toggle)", action)), nil
	}

	if actionErr != nil {
		return tools.ErrorResult(actionErr), nil
	}

	var sb strings.Builder
	sb.WriteString("# Entity Control\n\n")
	sb.WriteString(fmt.Sprintf("**Entity:** `%s`\n", entityID))
	sb.WriteString(fmt.Sprintf("**Action:** %s\n", action))
	sb.WriteString("**Result:** Success\n")

	return tools.TextResult(sb.String()), nil
}

// handleScenes handles the aftrs_hass_scenes tool
func handleScenes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Home Assistant not configured: %v", err)), nil
	}

	activate := tools.GetStringParam(req, "activate")

	// Activate scene if specified
	if activate != "" {
		if err := client.ActivateScene(ctx, activate); err != nil {
			return tools.ErrorResult(err), nil
		}

		return tools.TextResult(fmt.Sprintf("# Scene Activated\n\n**Scene:** `%s`\n", activate)), nil
	}

	// List scenes
	scenes, err := client.GetScenes(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Home Assistant Scenes\n\n")

	if len(scenes) == 0 {
		sb.WriteString("No scenes found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** scenes:\n\n", len(scenes)))
	sb.WriteString("| Scene ID | Name |\n")
	sb.WriteString("|----------|------|\n")

	for _, scene := range scenes {
		sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", scene.EntityID, scene.Name))
	}

	sb.WriteString("\n*Use `activate` parameter to activate a scene.*\n")

	return tools.TextResult(sb.String()), nil
}

// handleAutomations handles the aftrs_hass_automations tool
func handleAutomations(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Home Assistant not configured: %v", err)), nil
	}

	trigger := tools.GetStringParam(req, "trigger")

	// Trigger automation if specified
	if trigger != "" {
		if err := client.TriggerAutomation(ctx, trigger); err != nil {
			return tools.ErrorResult(err), nil
		}

		return tools.TextResult(fmt.Sprintf("# Automation Triggered\n\n**Automation:** `%s`\n", trigger)), nil
	}

	// List automations
	automations, err := client.GetAutomations(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Home Assistant Automations\n\n")

	if len(automations) == 0 {
		sb.WriteString("No automations found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** automations:\n\n", len(automations)))
	sb.WriteString("| ID | Name | State | Last Triggered |\n")
	sb.WriteString("|----|------|-------|----------------|\n")

	for _, auto := range automations {
		lastTriggered := "-"
		if auto.LastTriggered != "" {
			lastTriggered = auto.LastTriggered
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
			auto.ID, auto.Alias, auto.State, lastTriggered))
	}

	sb.WriteString("\n*Use `trigger` parameter to trigger an automation.*\n")

	return tools.TextResult(sb.String()), nil
}

// hexToRGB converts a hex color to RGB array
func hexToRGB(hex string) []int {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return []int{255, 255, 255}
	}

	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return []int{r, g, b}
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
