// Package streamdeck provides MCP tools for Elgato Stream Deck integration.
package streamdeck

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Stream Deck.
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string { return "streamdeck" }
func (m *Module) Description() string {
	return "Elgato Stream Deck physical control surface integration"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool:                mcp.NewTool("aftrs_streamdeck_status", mcp.WithDescription("Get Stream Deck connection status and device info.")),
			Handler:             handleStatus,
			Category:            "streamdeck",
			Subcategory:         "status",
			Tags:                []string{"streamdeck", "status", "devices"},
			UseCases:            []string{"Check connected devices", "Get device info"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
		},
		{
			Tool:                mcp.NewTool("aftrs_streamdeck_devices", mcp.WithDescription("List all connected Stream Deck devices.")),
			Handler:             handleDevices,
			Category:            "streamdeck",
			Subcategory:         "devices",
			Tags:                []string{"streamdeck", "devices", "list"},
			UseCases:            []string{"List Stream Decks", "Get device details"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
		},
		{
			Tool: mcp.NewTool("aftrs_streamdeck_buttons", mcp.WithDescription("Get button states for a device."),
				mcp.WithNumber("device", mcp.Description("Device index (default 0)"))),
			Handler:             handleButtons,
			Category:            "streamdeck",
			Subcategory:         "buttons",
			Tags:                []string{"streamdeck", "buttons", "state"},
			UseCases:            []string{"Get button layout", "Check button states"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
		},
		{
			Tool: mcp.NewTool("aftrs_streamdeck_set_image", mcp.WithDescription("Set a button image from file path."),
				mcp.WithNumber("device", mcp.Description("Device index (default 0)")),
				mcp.WithNumber("button", mcp.Required(), mcp.Description("Button index")),
				mcp.WithString("path", mcp.Required(), mcp.Description("Image file path (PNG recommended)"))),
			Handler:             handleSetImage,
			Category:            "streamdeck",
			Subcategory:         "buttons",
			Tags:                []string{"streamdeck", "image", "button"},
			UseCases:            []string{"Set button icon", "Display image on button"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_streamdeck_set_color", mcp.WithDescription("Set a button to a solid color."),
				mcp.WithNumber("device", mcp.Description("Device index (default 0)")),
				mcp.WithNumber("button", mcp.Required(), mcp.Description("Button index")),
				mcp.WithString("color", mcp.Required(), mcp.Description("Color as hex (#FF0000) or name (red, green, blue, etc.)"))),
			Handler:             handleSetColor,
			Category:            "streamdeck",
			Subcategory:         "buttons",
			Tags:                []string{"streamdeck", "color", "button"},
			UseCases:            []string{"Set button color", "Indicate state with color"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_streamdeck_clear", mcp.WithDescription("Clear button(s) to black."),
				mcp.WithNumber("device", mcp.Description("Device index (default 0)")),
				mcp.WithNumber("button", mcp.Description("Button index (omit to clear all)")),
				mcp.WithBoolean("all", mcp.Description("Clear all buttons"))),
			Handler:             handleClear,
			Category:            "streamdeck",
			Subcategory:         "buttons",
			Tags:                []string{"streamdeck", "clear", "reset"},
			UseCases:            []string{"Clear button", "Clear all buttons"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_streamdeck_brightness", mcp.WithDescription("Set device brightness."),
				mcp.WithNumber("device", mcp.Description("Device index (default 0)")),
				mcp.WithNumber("brightness", mcp.Required(), mcp.Description("Brightness 0-100"))),
			Handler:             handleBrightness,
			Category:            "streamdeck",
			Subcategory:         "device",
			Tags:                []string{"streamdeck", "brightness", "display"},
			UseCases:            []string{"Adjust brightness", "Dim display"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_streamdeck_reset", mcp.WithDescription("Reset device to default logo."),
				mcp.WithNumber("device", mcp.Description("Device index (default 0)"))),
			Handler:             handleReset,
			Category:            "streamdeck",
			Subcategory:         "device",
			Tags:                []string{"streamdeck", "reset", "logo"},
			UseCases:            []string{"Reset to default", "Show Elgato logo"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
			IsWrite:             true,
		},
		{
			Tool:                mcp.NewTool("aftrs_streamdeck_refresh", mcp.WithDescription("Rescan for connected Stream Deck devices.")),
			Handler:             handleRefresh,
			Category:            "streamdeck",
			Subcategory:         "devices",
			Tags:                []string{"streamdeck", "refresh", "scan"},
			UseCases:            []string{"Detect new devices", "Reconnect devices"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
		},
		{
			Tool:                mcp.NewTool("aftrs_streamdeck_health", mcp.WithDescription("Check Stream Deck system health.")),
			Handler:             handleHealth,
			Category:            "streamdeck",
			Subcategory:         "health",
			Tags:                []string{"streamdeck", "health", "status"},
			UseCases:            []string{"Check connection health", "Diagnose issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streamdeck",
		},
	}
}

var getClient = tools.LazyClient(clients.NewStreamDeckClient)

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(status), nil
}

func handleDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	devices, err := client.GetDevices(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	if len(devices) == 0 {
		return tools.TextResult("No Stream Deck devices connected"), nil
	}
	return tools.JSONResult(devices), nil
}

func handleButtons(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	deviceIndex := tools.GetIntParam(req, "device", 0)
	buttons, err := client.GetButtons(ctx, deviceIndex)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(buttons), nil
}

func handleSetImage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	deviceIndex := tools.GetIntParam(req, "device", 0)
	buttonIndex := tools.GetIntParam(req, "button", -1)
	imagePath, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	if buttonIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("button index required")), nil
	}

	if err := client.SetButtonImage(ctx, deviceIndex, buttonIndex, imagePath); err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Set button %d image to: %s", buttonIndex, imagePath)), nil
}

func handleSetColor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	deviceIndex := tools.GetIntParam(req, "device", 0)
	buttonIndex := tools.GetIntParam(req, "button", -1)
	colorStr, errResult := tools.RequireStringParam(req, "color")
	if errResult != nil {
		return errResult, nil
	}

	if buttonIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("button index required")), nil
	}

	r, g, b, err := parseColor(colorStr)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.SetButtonColor(ctx, deviceIndex, buttonIndex, r, g, b); err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Set button %d color to: %s", buttonIndex, colorStr)), nil
}

func handleClear(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	deviceIndex := tools.GetIntParam(req, "device", 0)
	buttonIndex := tools.GetIntParam(req, "button", -1)
	clearAll := tools.GetBoolParam(req, "all", false)

	if clearAll || buttonIndex < 0 {
		if err := client.ClearAllButtons(ctx, deviceIndex); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Cleared all buttons"), nil
	}

	if err := client.ClearButton(ctx, deviceIndex, buttonIndex); err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Cleared button %d", buttonIndex)), nil
}

func handleBrightness(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	deviceIndex := tools.GetIntParam(req, "device", 0)
	brightness := tools.GetIntParam(req, "brightness", 50)

	if err := client.SetBrightness(ctx, deviceIndex, brightness); err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Set brightness to %d%%", brightness)), nil
}

func handleReset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	deviceIndex := tools.GetIntParam(req, "device", 0)

	if err := client.ResetDevice(ctx, deviceIndex); err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult("Device reset to default logo"), nil
}

func handleRefresh(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	devices, err := client.RefreshDevices(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	if len(devices) == 0 {
		return tools.TextResult("No Stream Deck devices found"), nil
	}
	return tools.JSONResult(devices), nil
}

func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(health), nil
}

// parseColor parses a color string (hex or name) to RGB values
func parseColor(s string) (r, g, b uint8, err error) {
	s = strings.TrimSpace(strings.ToLower(s))

	// Named colors
	colors := map[string][3]uint8{
		"red":     {255, 0, 0},
		"green":   {0, 255, 0},
		"blue":    {0, 0, 255},
		"yellow":  {255, 255, 0},
		"cyan":    {0, 255, 255},
		"magenta": {255, 0, 255},
		"white":   {255, 255, 255},
		"black":   {0, 0, 0},
		"orange":  {255, 165, 0},
		"purple":  {128, 0, 128},
		"pink":    {255, 192, 203},
		"gray":    {128, 128, 128},
		"grey":    {128, 128, 128},
	}

	if rgb, ok := colors[s]; ok {
		return rgb[0], rgb[1], rgb[2], nil
	}

	// Hex color
	s = strings.TrimPrefix(s, "#")
	if len(s) == 6 {
		var rr, gg, bb int
		_, err := fmt.Sscanf(s, "%02x%02x%02x", &rr, &gg, &bb)
		if err == nil {
			return uint8(rr), uint8(gg), uint8(bb), nil
		}
	}
	if len(s) == 3 {
		var rr, gg, bb int
		_, err := fmt.Sscanf(s, "%1x%1x%1x", &rr, &gg, &bb)
		if err == nil {
			return uint8(rr * 17), uint8(gg * 17), uint8(bb * 17), nil
		}
	}

	return 0, 0, 0, fmt.Errorf("invalid color: %s (use hex like #FF0000 or name like red)", s)
}
