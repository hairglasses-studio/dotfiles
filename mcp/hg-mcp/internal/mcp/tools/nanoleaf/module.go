// Package nanoleaf provides MCP tools for controlling Nanoleaf light panels.
package nanoleaf

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements ToolModule for Nanoleaf panels.
type Module struct{}

var getClient = tools.LazyClient(clients.GetNanoleafClient)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "nanoleaf" }
func (m *Module) Description() string { return "Nanoleaf light panel control and monitoring" }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_nanoleaf_status",
				mcp.WithDescription("Get Nanoleaf panel status including model, firmware, power, brightness, and active effect."),
			),
			Handler:             handleStatus,
			Category:            "nanoleaf",
			Tags:                []string{"nanoleaf", "status", "lighting", "panels"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "nanoleaf",
		},
		{
			Tool: mcp.NewTool("aftrs_nanoleaf_health",
				mcp.WithDescription("Check Nanoleaf panel health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "nanoleaf",
			Tags:                []string{"nanoleaf", "health", "diagnostics"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "nanoleaf",
		},
		{
			Tool: mcp.NewTool("aftrs_nanoleaf_power",
				mcp.WithDescription("Turn Nanoleaf panels on or off."),
				mcp.WithBoolean("on", mcp.Required(), mcp.Description("true to turn on, false to turn off")),
			),
			Handler:             handlePower,
			Category:            "nanoleaf",
			Tags:                []string{"nanoleaf", "power", "lighting"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "nanoleaf",
		},
		{
			Tool: mcp.NewTool("aftrs_nanoleaf_brightness",
				mcp.WithDescription("Set Nanoleaf panel brightness (0-100)."),
				mcp.WithNumber("level", mcp.Required(), mcp.Description("Brightness level 0-100")),
			),
			Handler:             handleBrightness,
			Category:            "nanoleaf",
			Tags:                []string{"nanoleaf", "brightness", "lighting"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "nanoleaf",
		},
		{
			Tool: mcp.NewTool("aftrs_nanoleaf_color",
				mcp.WithDescription("Set Nanoleaf panel color. Accepts hex (e.g., 'FF0000') or color name (red, blue, green, etc.)."),
				mcp.WithString("color", mcp.Required(), mcp.Description("Color as hex ('FF0000') or name ('red', 'blue', 'green', 'white', 'purple', 'orange', 'cyan', 'yellow')")),
			),
			Handler:             handleColor,
			Category:            "nanoleaf",
			Tags:                []string{"nanoleaf", "color", "lighting"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "nanoleaf",
		},
		{
			Tool: mcp.NewTool("aftrs_nanoleaf_effect",
				mcp.WithDescription("Set or list Nanoleaf effects. Omit name to list available effects."),
				mcp.WithString("name", mcp.Description("Effect name to activate (omit to list all)")),
			),
			Handler:             handleEffect,
			Category:            "nanoleaf",
			Tags:                []string{"nanoleaf", "effect", "animation", "lighting"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "nanoleaf",
		},
		{
			Tool: mcp.NewTool("aftrs_nanoleaf_panels",
				mcp.WithDescription("List Nanoleaf panel IDs and layout positions."),
			),
			Handler:             handlePanels,
			Category:            "nanoleaf",
			Tags:                []string{"nanoleaf", "panels", "layout"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "nanoleaf",
		},
		{
			Tool: mcp.NewTool("aftrs_nanoleaf_discover",
				mcp.WithDescription("Discover Nanoleaf devices on the network."),
			),
			Handler:             handleDiscover,
			Category:            "nanoleaf",
			Tags:                []string{"nanoleaf", "discover", "network"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "nanoleaf",
		},
	}
}

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get Nanoleaf status: %w", err)), nil
	}
	return tools.JSONResult(status), nil
}

func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.JSONResult(map[string]interface{}{
			"connected":       false,
			"score":           0,
			"status":          "unconfigured",
			"recommendations": []string{"Set NANOLEAF_HOST and NANOLEAF_AUTH_TOKEN environment variables"},
		}), nil
	}
	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(health), nil
}

func handlePower(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	on := tools.GetBoolParam(req, "on", false)
	if err := client.SetPower(ctx, on); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set power: %w", err)), nil
	}
	state := "off"
	if on {
		state = "on"
	}
	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"power":   state,
	}), nil
}

func handleBrightness(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	level := tools.GetIntParam(req, "level", 50)
	if err := client.SetBrightness(ctx, level); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set brightness: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"success":    true,
		"brightness": level,
	}), nil
}

// colorNameToHSB converts a color name to hue (0-360), saturation (0-100), brightness (0-100).
func colorNameToHSB(name string) (int, int, int, bool) {
	colors := map[string][3]int{
		"red":    {0, 100, 100},
		"green":  {120, 100, 100},
		"blue":   {240, 100, 100},
		"white":  {0, 0, 100},
		"purple": {280, 100, 100},
		"orange": {30, 100, 100},
		"cyan":   {180, 100, 100},
		"yellow": {60, 100, 100},
		"pink":   {330, 80, 100},
		"warm":   {30, 50, 100},
	}
	if c, ok := colors[strings.ToLower(name)]; ok {
		return c[0], c[1], c[2], true
	}
	return 0, 0, 0, false
}

// hexToHSB converts a hex color string to Nanoleaf HSB values.
func hexToHSB(hex string) (int, int, int, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hex)
	}
	var r, g, b int
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, 0, err
	}
	// Convert RGB to HSB
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	max := rf
	if gf > max {
		max = gf
	}
	if bf > max {
		max = bf
	}
	min := rf
	if gf < min {
		min = gf
	}
	if bf < min {
		min = bf
	}
	d := max - min
	var h float64
	if d == 0 {
		h = 0
	} else if max == rf {
		h = 60 * float64(int((gf-bf)/d)%6)
	} else if max == gf {
		h = 60 * ((bf-rf)/d + 2)
	} else {
		h = 60 * ((rf-gf)/d + 4)
	}
	if h < 0 {
		h += 360
	}
	s := 0.0
	if max > 0 {
		s = d / max
	}
	return int(h), int(s * 100), int(max * 100), nil
}

func handleColor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	color, errResult := tools.RequireStringParam(req, "color")
	if errResult != nil {
		return errResult, nil
	}

	var hue, sat, bri int
	if h, s, b, ok := colorNameToHSB(color); ok {
		hue, sat, bri = h, s, b
	} else {
		var err error
		hue, sat, bri, err = hexToHSB(color)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid color %q: %w", color, err)), nil
		}
	}

	if err := client.SetColor(ctx, hue, sat, bri); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set color: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"color":   color,
		"hsb":     []int{hue, sat, bri},
	}), nil
}

func handleEffect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	name := tools.GetStringParam(req, "name")
	if name == "" {
		// List effects
		effects, err := client.ListEffects(ctx)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to list effects: %w", err)), nil
		}
		return tools.JSONResult(map[string]interface{}{
			"effects": effects,
			"count":   len(effects),
		}), nil
	}
	if err := client.SetEffect(ctx, name); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set effect: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"effect":  name,
	}), nil
}

func handlePanels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	layout, err := client.GetLayout(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get panel layout: %w", err)), nil
	}
	return tools.JSONResult(layout), nil
}

func handleDiscover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		// Still try basic discovery without auth
		return tools.JSONResult(map[string]interface{}{
			"devices": []interface{}{},
			"note":    "Set NANOLEAF_HOST and NANOLEAF_AUTH_TOKEN to connect",
		}), nil
	}
	devices, err := client.Discover(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("discovery failed: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"devices": devices,
		"count":   len(devices),
	}), nil
}
