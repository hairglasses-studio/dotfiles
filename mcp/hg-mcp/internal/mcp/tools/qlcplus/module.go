// Package qlcplus provides QLC+ open source lighting control tools for hg-mcp.
// QLC+ is an open source lighting control application that supports DMX, ArtNet,
// sACN, and many other protocols. It exposes a WebSocket API for remote control.
package qlcplus

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for QLC+.
type Module struct{}

// getClient returns the singleton QLC+ client (thread-safe via LazyClient).
var getClient = tools.LazyClient(clients.NewQLCPlusClient)

func (m *Module) Name() string {
	return "qlcplus"
}

func (m *Module) Description() string {
	return "QLC+ open source lighting control via WebSocket API"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_qlcplus_status",
				mcp.WithDescription("Get QLC+ connection status and WebSocket endpoint info."),
			),
			Handler:             handleStatus,
			Category:            "lighting",
			Subcategory:         "qlcplus",
			Tags:                []string{"qlcplus", "lighting", "status", "dmx"},
			UseCases:            []string{"Check QLC+ connection", "Verify WebSocket endpoint"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "qlcplus",
		},
		{
			Tool: mcp.NewTool("aftrs_qlcplus_health",
				mcp.WithDescription("Check QLC+ system health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "lighting",
			Subcategory:         "qlcplus",
			Tags:                []string{"qlcplus", "health", "diagnostics"},
			UseCases:            []string{"Diagnose QLC+ issues", "Check lighting health"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "qlcplus",
		},
		{
			Tool: mcp.NewTool("aftrs_qlcplus_widgets",
				mcp.WithDescription("List virtual console widgets in QLC+."),
			),
			Handler:             handleWidgets,
			Category:            "lighting",
			Subcategory:         "qlcplus",
			Tags:                []string{"qlcplus", "widgets", "virtual-console"},
			UseCases:            []string{"List QLC+ widgets", "Enumerate virtual console"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "qlcplus",
		},
		{
			Tool: mcp.NewTool("aftrs_qlcplus_widget_get",
				mcp.WithDescription("Get the current value of a virtual console widget."),
				mcp.WithString("widget_id", mcp.Required(), mcp.Description("Widget ID to query")),
			),
			Handler:             handleWidgetGet,
			Category:            "lighting",
			Subcategory:         "qlcplus",
			Tags:                []string{"qlcplus", "widget", "read", "value"},
			UseCases:            []string{"Read widget value", "Check slider position"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "qlcplus",
		},
		{
			Tool: mcp.NewTool("aftrs_qlcplus_widget_set",
				mcp.WithDescription("Set the value of a virtual console widget (slider, button, etc)."),
				mcp.WithString("widget_id", mcp.Required(), mcp.Description("Widget ID to control")),
				mcp.WithString("value", mcp.Required(), mcp.Description("Value to set (0-255 for sliders, 0/255 for buttons)")),
			),
			Handler:             handleWidgetSet,
			Category:            "lighting",
			Subcategory:         "qlcplus",
			Tags:                []string{"qlcplus", "widget", "write", "control"},
			UseCases:            []string{"Set widget value", "Control slider", "Press button"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "qlcplus",
		},
		{
			Tool: mcp.NewTool("aftrs_qlcplus_function",
				mcp.WithDescription("Start or stop a QLC+ function (scene, chaser, show, etc)."),
				mcp.WithString("function_id", mcp.Required(), mcp.Description("Function ID to control")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start or stop")),
			),
			Handler:             handleFunction,
			Category:            "lighting",
			Subcategory:         "qlcplus",
			Tags:                []string{"qlcplus", "function", "scene", "chaser", "control"},
			UseCases:            []string{"Start scene", "Stop chaser", "Trigger show function"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "qlcplus",
		},
		{
			Tool: mcp.NewTool("aftrs_qlcplus_channels",
				mcp.WithDescription("Get or set DMX channel values in a QLC+ universe."),
				mcp.WithNumber("universe", mcp.Description("DMX universe number (default: 0)")),
				mcp.WithNumber("start_channel", mcp.Description("Starting channel (default: 1)")),
				mcp.WithNumber("count", mcp.Description("Number of channels to read (default: 16)")),
				mcp.WithString("values", mcp.Description("Comma-separated channel:value pairs to set (e.g., '1:255,2:128,3:0')")),
			),
			Handler:             handleChannels,
			Category:            "lighting",
			Subcategory:         "qlcplus",
			Tags:                []string{"qlcplus", "dmx", "channels", "universe"},
			UseCases:            []string{"Read DMX values", "Set channel levels", "Control universe"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "qlcplus",
		},
	}
}

// handleStatus returns QLC+ connection status.
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
	sb.WriteString("# QLC+ Status\n\n")

	if status.Connected {
		sb.WriteString("**Status:** Connected\n")
	} else {
		sb.WriteString("**Status:** Disconnected\n")
	}
	sb.WriteString(fmt.Sprintf("**Host:** %s\n", status.Host))
	sb.WriteString(fmt.Sprintf("**Port:** %d\n", status.Port))
	sb.WriteString(fmt.Sprintf("**WebSocket URL:** `%s`\n", status.URL))

	if !status.Connected {
		sb.WriteString("\nStart QLC+ and enable the WebSocket server to connect.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns QLC+ system health.
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
	sb.WriteString("# QLC+ Health\n\n")
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

// handleWidgets lists virtual console widgets.
func handleWidgets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	widgets, err := client.GetWidgets(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(widgets) == 0 {
		return tools.TextResult("# QLC+ Widgets\n\nNo widgets found. Open QLC+ and create a virtual console layout."), nil
	}

	var sb strings.Builder
	sb.WriteString("# QLC+ Virtual Console Widgets\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d widgets\n\n", len(widgets)))
	sb.WriteString("| ID | Type | Name |\n")
	sb.WriteString("|---|---|---|\n")
	for _, w := range widgets {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", w.ID, w.Type, w.Name))
	}

	return tools.TextResult(sb.String()), nil
}

// handleWidgetGet reads the current value of a widget.
func handleWidgetGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widgetID, errResult := tools.RequireStringParam(req, "widget_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	value, err := client.GetWidgetValue(ctx, widgetID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"widget_id": widgetID,
		"value":     value,
	}), nil
}

// handleWidgetSet sets the value of a widget.
func handleWidgetSet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widgetID, errResult := tools.RequireStringParam(req, "widget_id")
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

	if err := client.SetWidgetValue(ctx, widgetID, value); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"widget_id": widgetID,
		"value":     value,
		"success":   true,
	}), nil
}

// handleFunction starts or stops a QLC+ function.
func handleFunction(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	functionID, errResult := tools.RequireStringParam(req, "function_id")
	if errResult != nil {
		return errResult, nil
	}

	actionRaw, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}
	action := strings.ToLower(actionRaw)

	var running bool
	switch action {
	case "start":
		running = true
	case "stop":
		running = false
	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("action must be 'start' or 'stop', got '%s'", action)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	if err := client.SetFunctionStatus(ctx, functionID, running); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"function_id": functionID,
		"action":      action,
		"success":     true,
	}), nil
}

// handleChannels gets or sets DMX channel values.
func handleChannels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	universe := tools.GetIntParam(req, "universe", 0)
	startChannel := tools.GetIntParam(req, "start_channel", 1)
	count := tools.GetIntParam(req, "count", 16)
	values := tools.GetStringParam(req, "values")

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	// If values provided, set channels
	if values != "" {
		channelValues, parseErr := parseChannelValuePairs(values)
		if parseErr != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, parseErr), nil
		}

		if err := client.SetChannelsValues(ctx, universe, channelValues); err != nil {
			return tools.ErrorResult(err), nil
		}

		return tools.JSONResult(map[string]interface{}{
			"universe": universe,
			"channels": channelValues,
			"action":   "set",
			"success":  true,
		}), nil
	}

	// Otherwise, read channels
	channelData, err := client.GetChannelsValues(ctx, universe, startChannel, count)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	channels := make([]map[string]interface{}, len(channelData))
	for i, val := range channelData {
		channels[i] = map[string]interface{}{
			"channel": startChannel + i,
			"value":   val,
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"universe":      universe,
		"start_channel": startChannel,
		"count":         len(channelData),
		"channels":      channels,
	}), nil
}

// parseChannelValuePairs parses "1:255,2:128,3:0" format.
func parseChannelValuePairs(input string) (map[int]int, error) {
	result := make(map[int]int)
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid channel:value pair '%s', expected format '1:255'", pair)
		}
		ch := 0
		val := 0
		if _, err := fmt.Sscanf(parts[0], "%d", &ch); err != nil {
			return nil, fmt.Errorf("invalid channel number '%s': %w", parts[0], err)
		}
		if _, err := fmt.Sscanf(parts[1], "%d", &val); err != nil {
			return nil, fmt.Errorf("invalid value '%s': %w", parts[1], err)
		}
		if ch < 1 || ch > 512 {
			return nil, fmt.Errorf("channel %d out of range (1-512)", ch)
		}
		if val < 0 || val > 255 {
			return nil, fmt.Errorf("value %d out of range (0-255)", val)
		}
		result[ch] = val
	}
	return result, nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
