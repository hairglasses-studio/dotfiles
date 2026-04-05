// Package atem provides MCP tools for Blackmagic ATEM switcher control.
package atem

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the ATEM tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "atem"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Blackmagic ATEM video switcher control for live production"
}

// Tools returns the ATEM tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_atem_status",
				mcp.WithDescription("Get ATEM switcher status including model, connection state, and current program/preview inputs"),
			),
			Handler:     handleATEMStatus,
			Category:    "atem",
			Subcategory: "status",
			Tags:        []string{"atem", "switcher", "video", "broadcast", "status"},
			UseCases:    []string{"Check ATEM connection", "Monitor switcher state", "Verify program/preview sources"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_atem_inputs",
				mcp.WithDescription("List all available video inputs on the ATEM switcher with names and types"),
			),
			Handler:     handleATEMInputs,
			Category:    "atem",
			Subcategory: "inputs",
			Tags:        []string{"atem", "inputs", "video", "sources"},
			UseCases:    []string{"List available sources", "Find input IDs", "Check input availability"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_atem_program",
				mcp.WithDescription("Set the program (live) output to a specific input source"),
				mcp.WithNumber("input_id",
					mcp.Required(),
					mcp.Description("Input ID to set as program output (1=Input1, 2=Input2, etc.)"),
				),
			),
			Handler:     handleATEMProgram,
			Category:    "atem",
			Subcategory: "switching",
			Tags:        []string{"atem", "program", "live", "output"},
			UseCases:    []string{"Switch live output", "Change program source", "Direct cut to input"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_atem_preview",
				mcp.WithDescription("Set the preview output to a specific input source"),
				mcp.WithNumber("input_id",
					mcp.Required(),
					mcp.Description("Input ID to set as preview output"),
				),
			),
			Handler:     handleATEMPreview,
			Category:    "atem",
			Subcategory: "switching",
			Tags:        []string{"atem", "preview", "prepare"},
			UseCases:    []string{"Prepare next shot", "Set preview source", "Stage next transition"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_atem_cut",
				mcp.WithDescription("Execute an instant cut transition from preview to program"),
			),
			Handler:     handleATEMCut,
			Category:    "atem",
			Subcategory: "transitions",
			Tags:        []string{"atem", "cut", "transition", "instant"},
			UseCases:    []string{"Instant switch", "Hard cut", "Quick transition"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_atem_auto",
				mcp.WithDescription("Execute an auto transition from preview to program using current transition settings"),
			),
			Handler:     handleATEMAuto,
			Category:    "atem",
			Subcategory: "transitions",
			Tags:        []string{"atem", "auto", "transition", "mix", "dissolve"},
			UseCases:    []string{"Smooth transition", "Auto mix", "Dissolve to preview"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_atem_transition",
				mcp.WithDescription("Configure transition type and rate for auto transitions"),
				mcp.WithString("style",
					mcp.Required(),
					mcp.Description("Transition style: mix, dip, wipe, dve, or sting"),
					mcp.Enum("mix", "dip", "wipe", "dve", "sting"),
				),
				mcp.WithNumber("rate",
					mcp.Description("Transition duration in frames (e.g., 30 for 1 second at 30fps)"),
				),
			),
			Handler:     handleATEMTransition,
			Category:    "atem",
			Subcategory: "transitions",
			Tags:        []string{"atem", "transition", "settings", "mix", "wipe"},
			UseCases:    []string{"Set transition type", "Configure mix duration", "Setup wipe effect"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_atem_health",
				mcp.WithDescription("Check ATEM switcher connection health and get troubleshooting recommendations"),
			),
			Handler:     handleATEMHealth,
			Category:    "atem",
			Subcategory: "status",
			Tags:        []string{"atem", "health", "diagnostics", "troubleshooting"},
			UseCases:    []string{"Diagnose connection issues", "Check switcher health", "Get troubleshooting tips"},
			Complexity:  tools.ComplexitySimple,
		},
	}

	// Apply circuit breaker to all tools — network-dependent (TCP)
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "atem"
	}

	return allTools
}

var getATEMClient = tools.LazyClient(clients.NewATEMClient)

func handleATEMStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getATEMClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ATEM client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleATEMInputs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getATEMClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ATEM client: %w", err)), nil
	}

	inputs, err := client.GetInputs(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get inputs: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"inputs": inputs,
		"count":  len(inputs),
	}), nil
}

func handleATEMProgram(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getATEMClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ATEM client: %w", err)), nil
	}

	inputID, errResult := tools.RequireIntParam(req, "input_id")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.SetProgram(ctx, inputID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set program: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":  true,
		"input_id": inputID,
		"message":  fmt.Sprintf("Program output set to input %d", inputID),
	}), nil
}

func handleATEMPreview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getATEMClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ATEM client: %w", err)), nil
	}

	inputID, errResult := tools.RequireIntParam(req, "input_id")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.SetPreview(ctx, inputID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set preview: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":  true,
		"input_id": inputID,
		"message":  fmt.Sprintf("Preview output set to input %d", inputID),
	}), nil
}

func handleATEMCut(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getATEMClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ATEM client: %w", err)), nil
	}

	if err := client.Cut(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to execute cut: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "Cut transition executed",
	}), nil
}

func handleATEMAuto(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getATEMClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ATEM client: %w", err)), nil
	}

	if err := client.Auto(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to execute auto transition: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "Auto transition executed",
	}), nil
}

func handleATEMTransition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getATEMClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ATEM client: %w", err)), nil
	}

	style, errResult := tools.RequireStringParam(req, "style")
	if errResult != nil {
		return errResult, nil
	}

	rate := tools.GetIntParam(req, "rate", 30)

	if err := client.SetTransition(ctx, style, rate); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set transition: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"style":   style,
		"rate":    rate,
		"message": fmt.Sprintf("Transition set to %s with %d frame duration", style, rate),
	}), nil
}

func handleATEMHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getATEMClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ATEM client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
