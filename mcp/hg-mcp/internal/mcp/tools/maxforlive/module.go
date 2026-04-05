// Package maxforlive provides MCP tools for Max for Live device control.
package maxforlive

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Max for Live tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "maxforlive"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Max for Live device control and parameter automation via OSC"
}

// Tools returns the Max for Live tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_m4l_status",
				mcp.WithDescription("Get Max for Live bridge connection status and device count"),
			),
			Handler:             handleM4LStatus,
			Category:            "maxforlive",
			Subcategory:         "status",
			Tags:                []string{"m4l", "maxforlive", "ableton", "status"},
			UseCases:            []string{"Check M4L connection", "View device count", "Verify bridge"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "maxforlive",
		},
		{
			Tool: mcp.NewTool("aftrs_m4l_devices",
				mcp.WithDescription("List all Max for Live devices connected via the bridge"),
			),
			Handler:             handleM4LDevices,
			Category:            "maxforlive",
			Subcategory:         "devices",
			Tags:                []string{"m4l", "devices", "list"},
			UseCases:            []string{"List M4L devices", "Find device IDs", "View connected devices"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "maxforlive",
		},
		{
			Tool: mcp.NewTool("aftrs_m4l_parameters",
				mcp.WithDescription("Get parameters for a Max for Live device"),
				mcp.WithString("device_id", mcp.Required(), mcp.Description("Device ID to get parameters for")),
			),
			Handler:             handleM4LParameters,
			Category:            "maxforlive",
			Subcategory:         "parameters",
			Tags:                []string{"m4l", "parameters", "get"},
			UseCases:            []string{"View device parameters", "Get parameter values", "List controls"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "maxforlive",
		},
		{
			Tool: mcp.NewTool("aftrs_m4l_parameter_set",
				mcp.WithDescription("Set a parameter value on a Max for Live device"),
				mcp.WithString("device_id", mcp.Required(), mcp.Description("Device ID")),
				mcp.WithString("param_name", mcp.Required(), mcp.Description("Parameter name")),
				mcp.WithNumber("value", mcp.Required(), mcp.Description("Value to set")),
				mcp.WithBoolean("normalized", mcp.Description("If true, treat value as 0-1 normalized")),
			),
			Handler:             handleM4LParameterSet,
			Category:            "maxforlive",
			Subcategory:         "parameters",
			Tags:                []string{"m4l", "parameter", "set", "automation"},
			UseCases:            []string{"Set parameter value", "Automate device", "Control effect"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "maxforlive",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_m4l_send",
				mcp.WithDescription("Send a custom OSC message to a Max for Live device"),
				mcp.WithString("address", mcp.Required(), mcp.Description("OSC address (e.g., /m4l/device/custom)")),
				mcp.WithArray("args", mcp.Description("OSC message arguments")),
			),
			Handler:             handleM4LSend,
			Category:            "maxforlive",
			Subcategory:         "custom",
			Tags:                []string{"m4l", "osc", "send", "custom"},
			UseCases:            []string{"Send custom message", "Trigger custom action", "Direct OSC control"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "maxforlive",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_m4l_mappings",
				mcp.WithDescription("List parameter mappings between M4L devices"),
			),
			Handler:             handleM4LMappings,
			Category:            "maxforlive",
			Subcategory:         "mappings",
			Tags:                []string{"m4l", "mappings", "modulation"},
			UseCases:            []string{"View mappings", "Check modulation sources", "List connections"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "maxforlive",
		},
		{
			Tool: mcp.NewTool("aftrs_m4l_macro",
				mcp.WithDescription("Trigger, store, or recall a macro/preset on an M4L device"),
				mcp.WithString("device_id", mcp.Required(), mcp.Description("Device ID")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action to perform"), mcp.Enum("trigger", "store", "recall", "list")),
				mcp.WithString("macro_id", mcp.Description("Macro ID or name (for trigger/store/recall)")),
			),
			Handler:             handleM4LMacro,
			Category:            "maxforlive",
			Subcategory:         "macros",
			Tags:                []string{"m4l", "macro", "preset", "recall"},
			UseCases:            []string{"Trigger macro", "Store preset", "Recall settings"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "maxforlive",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_m4l_health",
				mcp.WithDescription("Check Max for Live bridge health and get troubleshooting recommendations"),
			),
			Handler:             handleM4LHealth,
			Category:            "maxforlive",
			Subcategory:         "status",
			Tags:                []string{"m4l", "health", "diagnostics", "troubleshooting"},
			UseCases:            []string{"Check connection", "Diagnose issues", "Verify bridge device"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "maxforlive",
		},
	}
}

var getM4LClient = tools.LazyClient(clients.NewMaxForLiveClient)

func handleM4LStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getM4LClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create M4L client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleM4LDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getM4LClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create M4L client: %w", err)), nil
	}

	devices, err := client.GetDevices(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get devices: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"devices": devices,
		"count":   len(devices),
	}), nil
}

func handleM4LParameters(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getM4LClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create M4L client: %w", err)), nil
	}

	deviceID, errResult := tools.RequireStringParam(req, "device_id")
	if errResult != nil {
		return errResult, nil
	}

	params, err := client.GetParameters(ctx, deviceID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get parameters: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"device_id":  deviceID,
		"parameters": params,
		"count":      len(params),
	}), nil
}

func handleM4LParameterSet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getM4LClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create M4L client: %w", err)), nil
	}

	deviceID, errResult := tools.RequireStringParam(req, "device_id")
	if errResult != nil {
		return errResult, nil
	}

	paramName, errResult := tools.RequireStringParam(req, "param_name")
	if errResult != nil {
		return errResult, nil
	}

	value := tools.GetFloatParam(req, "value", 0)
	normalized := tools.GetBoolParam(req, "normalized", false)

	var setErr error
	if normalized {
		setErr = client.SetParameterNormalized(ctx, deviceID, paramName, value)
	} else {
		setErr = client.SetParameter(ctx, deviceID, paramName, value)
	}

	if setErr != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set parameter: %w", setErr)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":    true,
		"device_id":  deviceID,
		"param_name": paramName,
		"value":      value,
		"normalized": normalized,
		"message":    fmt.Sprintf("Parameter '%s' set to %v", paramName, value),
	}), nil
}

func handleM4LSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getM4LClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create M4L client: %w", err)), nil
	}

	address, errResult := tools.RequireStringParam(req, "address")
	if errResult != nil {
		return errResult, nil
	}

	// Parse args
	var args []interface{}
	if argsMap, ok := req.Params.Arguments.(map[string]interface{}); ok {
		if argsRaw, exists := argsMap["args"]; exists {
			if argsArray, ok := argsRaw.([]interface{}); ok {
				args = argsArray
			}
		}
	}

	if err := client.SendCustomMessage(ctx, address, args...); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to send message: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"address": address,
		"args":    args,
		"message": fmt.Sprintf("OSC message sent to %s", address),
	}), nil
}

func handleM4LMappings(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getM4LClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create M4L client: %w", err)), nil
	}

	mappings, err := client.GetMappings(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get mappings: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"mappings": mappings,
		"count":    len(mappings),
	}), nil
}

func handleM4LMacro(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getM4LClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create M4L client: %w", err)), nil
	}

	deviceID, errResult := tools.RequireStringParam(req, "device_id")
	if errResult != nil {
		return errResult, nil
	}

	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	macroID := tools.GetStringParam(req, "macro_id")

	switch action {
	case "list":
		macros, err := client.GetMacros(ctx, deviceID)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to list macros: %w", err)), nil
		}

		return tools.JSONResult(map[string]interface{}{
			"device_id": deviceID,
			"macros":    macros,
			"count":     len(macros),
		}), nil

	case "trigger":
		if macroID == "" {
			return tools.ErrorResult(fmt.Errorf("macro_id is required for trigger")), nil
		}
		if err := client.TriggerMacro(ctx, deviceID, macroID); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to trigger macro: %w", err)), nil
		}

	case "store":
		if macroID == "" {
			return tools.ErrorResult(fmt.Errorf("macro_id (name) is required for store")), nil
		}
		if err := client.StoreMacro(ctx, deviceID, macroID); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to store macro: %w", err)), nil
		}

	case "recall":
		if macroID == "" {
			return tools.ErrorResult(fmt.Errorf("macro_id is required for recall")), nil
		}
		if err := client.RecallMacro(ctx, deviceID, macroID); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to recall macro: %w", err)), nil
		}

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s", action)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":   true,
		"device_id": deviceID,
		"action":    action,
		"macro_id":  macroID,
		"message":   fmt.Sprintf("Macro action '%s' executed", action),
	}), nil
}

func handleM4LHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getM4LClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create M4L client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
