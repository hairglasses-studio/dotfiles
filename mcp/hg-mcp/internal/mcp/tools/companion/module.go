// Package companion provides Bitfocus Companion control tools for hg-mcp.
// Companion is a modular stream deck / button box controller supporting
// 400+ device modules for professional AV control (lighting, video, audio).
package companion

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Bitfocus Companion.
type Module struct{}

// getClient returns the singleton Companion client (thread-safe via LazyClient).
var getClient = tools.LazyClient(clients.NewCompanionClient)

func (m *Module) Name() string {
	return "companion"
}

func (m *Module) Description() string {
	return "Bitfocus Companion control surface automation via HTTP API"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_companion_status",
				mcp.WithDescription("Get Bitfocus Companion connection status and API endpoint info."),
			),
			Handler:             handleStatus,
			Category:            "automation",
			Subcategory:         "companion",
			Tags:                []string{"companion", "bitfocus", "streamdeck", "status"},
			UseCases:            []string{"Check Companion connection", "Verify API endpoint"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "companion",
		},
		{
			Tool: mcp.NewTool("aftrs_companion_health",
				mcp.WithDescription("Check Bitfocus Companion system health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "automation",
			Subcategory:         "companion",
			Tags:                []string{"companion", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Companion issues", "Check system health"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "companion",
		},
		{
			Tool: mcp.NewTool("aftrs_companion_button_press",
				mcp.WithDescription("Press a button on a specific page and bank in Companion."),
				mcp.WithNumber("page", mcp.Required(), mcp.Description("Page number (1-99)")),
				mcp.WithNumber("bank", mcp.Required(), mcp.Description("Bank/button number (1-32)")),
			),
			Handler:             handleButtonPress,
			Category:            "automation",
			Subcategory:         "companion",
			Tags:                []string{"companion", "button", "press", "trigger"},
			UseCases:            []string{"Trigger button action", "Fire macro", "Control device"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "companion",
		},
		{
			Tool: mcp.NewTool("aftrs_companion_variable_get",
				mcp.WithDescription("Get the value of a Companion custom variable."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Variable name")),
			),
			Handler:             handleVariableGet,
			Category:            "automation",
			Subcategory:         "companion",
			Tags:                []string{"companion", "variable", "read"},
			UseCases:            []string{"Read variable value", "Check state"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "companion",
		},
		{
			Tool: mcp.NewTool("aftrs_companion_variable_set",
				mcp.WithDescription("Set the value of a Companion custom variable."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Variable name")),
				mcp.WithString("value", mcp.Required(), mcp.Description("Value to set")),
			),
			Handler:             handleVariableSet,
			Category:            "automation",
			Subcategory:         "companion",
			Tags:                []string{"companion", "variable", "write"},
			UseCases:            []string{"Set variable value", "Update state"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "companion",
		},
		{
			Tool: mcp.NewTool("aftrs_companion_instances",
				mcp.WithDescription("List connected module instances in Companion (device modules)."),
			),
			Handler:             handleInstances,
			Category:            "automation",
			Subcategory:         "companion",
			Tags:                []string{"companion", "instances", "modules", "devices"},
			UseCases:            []string{"List device modules", "Check instance status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "companion",
		},
		{
			Tool: mcp.NewTool("aftrs_companion_surfaces",
				mcp.WithDescription("List connected control surfaces (Stream Decks, button boxes)."),
			),
			Handler:             handleSurfaces,
			Category:            "automation",
			Subcategory:         "companion",
			Tags:                []string{"companion", "surfaces", "streamdeck", "hardware"},
			UseCases:            []string{"List connected surfaces", "Check hardware"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "companion",
		},
	}
}

// handleStatus returns Companion connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Bitfocus Companion Status\n\n")
	if status.Connected {
		sb.WriteString("**Status:** Connected\n")
	} else {
		sb.WriteString("**Status:** Disconnected\n")
	}
	sb.WriteString(fmt.Sprintf("**Host:** %s\n", status.Host))
	sb.WriteString(fmt.Sprintf("**Port:** %d\n", status.Port))
	sb.WriteString(fmt.Sprintf("**API URL:** `%s`\n", status.URL))

	if !status.Connected {
		sb.WriteString("\nStart Bitfocus Companion to connect.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns Companion system health.
func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Bitfocus Companion Health\n\n")
	sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
		}
	}

	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleButtonPress presses a button.
func handleButtonPress(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	page := tools.GetIntParam(req, "page", 0)
	if page < 1 || page > 99 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("page must be 1-99, got %d", page)), nil
	}

	bank := tools.GetIntParam(req, "bank", 0)
	if bank < 1 || bank > 32 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("bank must be 1-32, got %d", bank)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	if err := client.PressButton(ctx, page, bank); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"page":    page,
		"bank":    bank,
		"action":  "press",
		"success": true,
	}), nil
}

// handleVariableGet reads a custom variable.
func handleVariableGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	value, err := client.GetVariable(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"name":  name,
		"value": value,
	}), nil
}

// handleVariableSet sets a custom variable.
func handleVariableSet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	value, errResult := tools.RequireStringParam(req, "value")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	if err := client.SetVariable(ctx, name, value); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"name":    name,
		"value":   value,
		"success": true,
	}), nil
}

// handleInstances lists connected module instances.
func handleInstances(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	instances, err := client.GetInstances(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(instances) == 0 {
		return tools.TextResult("# Companion Instances\n\nNo module instances found."), nil
	}

	var sb strings.Builder
	sb.WriteString("# Companion Module Instances\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d instances\n\n", len(instances)))
	sb.WriteString("| Label | Module | Status | Enabled |\n")
	sb.WriteString("|---|---|---|---|\n")
	for _, inst := range instances {
		enabled := "No"
		if inst.Enabled {
			enabled = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			inst.Label, inst.Module, inst.Status, enabled))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSurfaces lists connected control surfaces.
func handleSurfaces(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	surfaces, err := client.GetSurfaces(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(surfaces) == 0 {
		return tools.TextResult("# Companion Surfaces\n\nNo control surfaces connected."), nil
	}

	var sb strings.Builder
	sb.WriteString("# Companion Control Surfaces\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d surfaces\n\n", len(surfaces)))
	sb.WriteString("| Name | Type | Online |\n")
	sb.WriteString("|---|---|---|\n")
	for _, surf := range surfaces {
		online := "Offline"
		if surf.Online {
			online = "Online"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", surf.Name, surf.Type, online))
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
