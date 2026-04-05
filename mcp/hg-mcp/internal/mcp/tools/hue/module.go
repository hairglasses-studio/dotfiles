// Package hue provides MCP tools for controlling Philips Hue lights.
package hue

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements ToolModule for Philips Hue.
type Module struct{}

var getClient = tools.LazyClient(clients.GetHueClient)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "hue" }
func (m *Module) Description() string { return "Philips Hue light control via bridge" }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_hue_status",
				mcp.WithDescription("Get Hue bridge status including model, firmware, and light count."),
			),
			Handler:             handleStatus,
			Category:            "hue",
			Tags:                []string{"hue", "status", "bridge", "lighting"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_health",
				mcp.WithDescription("Check Hue bridge health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "hue",
			Tags:                []string{"hue", "health", "diagnostics"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_lights",
				mcp.WithDescription("List all Hue lights with current state (on/off, brightness, color)."),
			),
			Handler:             handleLights,
			Category:            "hue",
			Tags:                []string{"hue", "lights", "list"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_light_control",
				mcp.WithDescription("Control a Hue light: power, brightness, color, color temperature."),
				mcp.WithString("light_id", mcp.Required(), mcp.Description("Light ID (from aftrs_hue_lights)")),
				mcp.WithBoolean("on", mcp.Description("Turn on (true) or off (false)")),
				mcp.WithNumber("brightness", mcp.Description("Brightness 0-254")),
				mcp.WithString("color", mcp.Description("Color as hex (e.g., 'FF0000') or name ('red', 'blue', etc.)")),
				mcp.WithNumber("ct", mcp.Description("Color temperature in mireds (153=cold, 500=warm)")),
			),
			Handler:             handleLightControl,
			Category:            "hue",
			Tags:                []string{"hue", "light", "control", "brightness", "color"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_rooms",
				mcp.WithDescription("List all Hue rooms/groups with light counts and state."),
			),
			Handler:             handleRooms,
			Category:            "hue",
			Tags:                []string{"hue", "rooms", "groups"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_room_control",
				mcp.WithDescription("Control all lights in a Hue room/group."),
				mcp.WithString("group_id", mcp.Required(), mcp.Description("Group/room ID (from aftrs_hue_rooms)")),
				mcp.WithBoolean("on", mcp.Description("Turn on (true) or off (false)")),
				mcp.WithNumber("brightness", mcp.Description("Brightness 0-254")),
				mcp.WithString("color", mcp.Description("Color as hex or name")),
			),
			Handler:             handleRoomControl,
			Category:            "hue",
			Tags:                []string{"hue", "room", "group", "control"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_scenes",
				mcp.WithDescription("List all saved Hue scenes."),
			),
			Handler:             handleScenes,
			Category:            "hue",
			Tags:                []string{"hue", "scenes", "list"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_scene_activate",
				mcp.WithDescription("Activate a saved Hue scene."),
				mcp.WithString("scene_id", mcp.Required(), mcp.Description("Scene ID to activate (from aftrs_hue_scenes)")),
			),
			Handler:             handleSceneActivate,
			Category:            "hue",
			Tags:                []string{"hue", "scene", "activate"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_discover",
				mcp.WithDescription("Discover Hue bridges on the network via NUPNP/mDNS."),
			),
			Handler:             handleDiscover,
			Category:            "hue",
			Tags:                []string{"hue", "discover", "bridge", "network"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
		{
			Tool: mcp.NewTool("aftrs_hue_entertainment",
				mcp.WithDescription("List Hue entertainment groups/zones for streaming."),
			),
			Handler:             handleEntertainment,
			Category:            "hue",
			Tags:                []string{"hue", "entertainment", "streaming", "zones"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "hue",
		},
	}
}

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	status, err := client.GetBridgeStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get bridge status: %w", err)), nil
	}
	// Augment with light/group/scene counts
	if lights, err := client.GetLights(ctx); err == nil {
		status.LightCount = len(lights)
	}
	if rooms, err := client.GetRooms(ctx); err == nil {
		status.GroupCount = len(rooms)
	}
	if scenes, err := client.GetScenes(ctx); err == nil {
		status.SceneCount = len(scenes)
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
			"recommendations": []string{"Set HUE_BRIDGE_IP and HUE_USERNAME environment variables"},
		}), nil
	}
	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(health), nil
}

func handleLights(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	lights, err := client.GetLights(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get lights: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"lights": lights,
		"count":  len(lights),
	}), nil
}

// hexToHueXY converts hex color to Hue xy color space (simplified).
// Returns hue (0-65535) and sat (0-254) for the Hue API.
func hexToHueSat(hex string) (int, int, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, fmt.Errorf("invalid hex color: %s", hex)
	}
	var r, g, b int
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return 0, 0, err
	}
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
	// Convert to Hue API range: hue 0-65535, sat 0-254
	return int(h / 360 * 65535), int(s * 254), nil
}

func colorNameToHueSat(name string) (int, int, bool) {
	colors := map[string][2]int{
		"red":    {0, 254},
		"green":  {21845, 254},
		"blue":   {43690, 254},
		"white":  {0, 0},
		"purple": {50971, 254},
		"orange": {5461, 254},
		"cyan":   {32768, 254},
		"yellow": {10922, 254},
		"pink":   {59932, 200},
		"warm":   {5461, 127},
	}
	if c, ok := colors[strings.ToLower(name)]; ok {
		return c[0], c[1], true
	}
	return 0, 0, false
}

func buildLightState(req mcp.CallToolRequest) map[string]interface{} {
	state := make(map[string]interface{})

	// Get "on" as raw argument since it's a boolean with valid false value
	if args, ok := req.Params.Arguments.(map[string]interface{}); ok {
		if v, exists := args["on"]; exists {
			if b, ok := v.(bool); ok {
				state["on"] = b
			}
		}
	}

	if bri := tools.GetIntParam(req, "brightness", -1); bri >= 0 {
		if bri > 254 {
			bri = 254
		}
		state["bri"] = bri
	}

	if ct := tools.GetIntParam(req, "ct", -1); ct > 0 {
		state["ct"] = ct
	}

	color := tools.GetStringParam(req, "color")
	if color != "" {
		if h, s, ok := colorNameToHueSat(color); ok {
			state["hue"] = h
			state["sat"] = s
		} else if h, s, err := hexToHueSat(color); err == nil {
			state["hue"] = h
			state["sat"] = s
		}
	}

	return state
}

func handleLightControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	lightID, errResult := tools.RequireStringParam(req, "light_id")
	if errResult != nil {
		return errResult, nil
	}

	state := buildLightState(req)
	if len(state) == 0 {
		return tools.ErrorResult(fmt.Errorf("at least one control parameter required (on, brightness, color, ct)")), nil
	}

	if err := client.SetLightState(ctx, lightID, state); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to control light: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"success":  true,
		"light_id": lightID,
		"state":    state,
	}), nil
}

func handleRooms(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	rooms, err := client.GetRooms(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get rooms: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"rooms": rooms,
		"count": len(rooms),
	}), nil
}

func handleRoomControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	groupID, errResult := tools.RequireStringParam(req, "group_id")
	if errResult != nil {
		return errResult, nil
	}
	// Validate numeric group ID
	if _, err := strconv.Atoi(groupID); err != nil {
		return tools.ErrorResult(fmt.Errorf("group_id must be numeric")), nil
	}

	state := buildLightState(req)
	if len(state) == 0 {
		return tools.ErrorResult(fmt.Errorf("at least one control parameter required (on, brightness, color)")), nil
	}

	if err := client.SetRoomState(ctx, groupID, state); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to control room: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"success":  true,
		"group_id": groupID,
		"state":    state,
	}), nil
}

func handleScenes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	scenes, err := client.GetScenes(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get scenes: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"scenes": scenes,
		"count":  len(scenes),
	}), nil
}

func handleSceneActivate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	sceneID, errResult := tools.RequireStringParam(req, "scene_id")
	if errResult != nil {
		return errResult, nil
	}
	if err := client.ActivateScene(ctx, sceneID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to activate scene: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"success":  true,
		"scene_id": sceneID,
	}), nil
}

func handleDiscover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.JSONResult(map[string]interface{}{
			"bridges": []interface{}{},
			"note":    "Set HUE_BRIDGE_IP and HUE_USERNAME to connect",
		}), nil
	}
	bridges, err := client.Discover(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("discovery failed: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"bridges": bridges,
		"count":   len(bridges),
	}), nil
}

func handleEntertainment(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	groups, err := client.GetEntertainmentGroups(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get entertainment groups: %w", err)), nil
	}
	return tools.JSONResult(map[string]interface{}{
		"entertainment_groups": groups,
		"count":                len(groups),
	}), nil
}
