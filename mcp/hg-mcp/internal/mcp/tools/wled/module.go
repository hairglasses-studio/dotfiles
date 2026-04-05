// Package wled provides MCP tools for WLED LED controller management.
package wled

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for WLED
type Module struct{}

var getClient = tools.LazyClient(clients.NewWLEDClient)

func (m *Module) Name() string {
	return "wled"
}

func (m *Module) Description() string {
	return "WLED LED controller discovery and control"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_wled_discover",
				mcp.WithDescription("Discover WLED controllers on the network. Scans subnet for WLED devices."),
				mcp.WithString("subnet", mcp.Description("Subnet to scan (e.g., '192.168.1.0/24'). Default: 192.168.1.0/24")),
			),
			Handler:             handleDiscover,
			Category:            "wled",
			Subcategory:         "discovery",
			Tags:                []string{"wled", "discover", "scan", "led", "network"},
			UseCases:            []string{"Find WLED devices", "Network LED discovery"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "wled",
		},
		{
			Tool: mcp.NewTool("aftrs_wled_status",
				mcp.WithDescription("Get status of a WLED device including brightness, effect, and segments."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
			),
			Handler:             handleStatus,
			Category:            "wled",
			Subcategory:         "control",
			Tags:                []string{"wled", "status", "led", "info"},
			UseCases:            []string{"Check LED status", "View current effect"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
		},
		{
			Tool: mcp.NewTool("aftrs_wled_power",
				mcp.WithDescription("Turn WLED device on or off."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
				mcp.WithBoolean("on", mcp.Required(), mcp.Description("true to turn on, false to turn off")),
			),
			Handler:             handlePower,
			Category:            "wled",
			Subcategory:         "control",
			Tags:                []string{"wled", "power", "on", "off", "led"},
			UseCases:            []string{"Turn LEDs on/off", "Power control"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_wled_brightness",
				mcp.WithDescription("Set WLED device brightness."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
				mcp.WithNumber("brightness", mcp.Required(), mcp.Description("Brightness level 0-255")),
			),
			Handler:             handleBrightness,
			Category:            "wled",
			Subcategory:         "control",
			Tags:                []string{"wled", "brightness", "dim", "led"},
			UseCases:            []string{"Adjust LED brightness", "Dim lights"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_wled_effect",
				mcp.WithDescription("Set WLED effect by ID or name."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
				mcp.WithNumber("effect_id", mcp.Description("Effect ID (0-based)")),
				mcp.WithString("effect_name", mcp.Description("Effect name (partial match supported)")),
			),
			Handler:             handleEffect,
			Category:            "wled",
			Subcategory:         "control",
			Tags:                []string{"wled", "effect", "animation", "led"},
			UseCases:            []string{"Change LED effect", "Set animation"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_wled_color",
				mcp.WithDescription("Set WLED primary color."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
				mcp.WithNumber("red", mcp.Required(), mcp.Description("Red value 0-255")),
				mcp.WithNumber("green", mcp.Required(), mcp.Description("Green value 0-255")),
				mcp.WithNumber("blue", mcp.Required(), mcp.Description("Blue value 0-255")),
			),
			Handler:             handleColor,
			Category:            "wled",
			Subcategory:         "control",
			Tags:                []string{"wled", "color", "rgb", "led"},
			UseCases:            []string{"Set LED color", "Change RGB values"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_wled_palette",
				mcp.WithDescription("Set WLED color palette."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
				mcp.WithNumber("palette_id", mcp.Required(), mcp.Description("Palette ID")),
			),
			Handler:             handlePalette,
			Category:            "wled",
			Subcategory:         "control",
			Tags:                []string{"wled", "palette", "colors", "led"},
			UseCases:            []string{"Change color palette", "Set color scheme"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_wled_effects_list",
				mcp.WithDescription("List all available effects on a WLED device."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
			),
			Handler:             handleEffectsList,
			Category:            "wled",
			Subcategory:         "info",
			Tags:                []string{"wled", "effects", "list", "animations"},
			UseCases:            []string{"Browse available effects", "Find effect IDs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
		},
		{
			Tool: mcp.NewTool("aftrs_wled_palettes_list",
				mcp.WithDescription("List all available color palettes on a WLED device."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
			),
			Handler:             handlePalettesList,
			Category:            "wled",
			Subcategory:         "info",
			Tags:                []string{"wled", "palettes", "list", "colors"},
			UseCases:            []string{"Browse color palettes", "Find palette IDs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
		},
		{
			Tool: mcp.NewTool("aftrs_wled_preset_save",
				mcp.WithDescription("Save current WLED state as a preset."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
				mcp.WithNumber("preset_id", mcp.Required(), mcp.Description("Preset slot ID (1-250)")),
				mcp.WithString("name", mcp.Required(), mcp.Description("Preset name")),
			),
			Handler:             handlePresetSave,
			Category:            "wled",
			Subcategory:         "presets",
			Tags:                []string{"wled", "preset", "save", "store"},
			UseCases:            []string{"Save LED configuration", "Create preset"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_wled_preset_load",
				mcp.WithDescription("Load a saved WLED preset."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
				mcp.WithNumber("preset_id", mcp.Required(), mcp.Description("Preset slot ID to load")),
			),
			Handler:             handlePresetLoad,
			Category:            "wled",
			Subcategory:         "presets",
			Tags:                []string{"wled", "preset", "load", "recall"},
			UseCases:            []string{"Recall LED preset", "Load configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "wled",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_wled_artnet_config",
				mcp.WithDescription("Configure Art-Net settings for WLED device."),
				mcp.WithString("ip", mcp.Required(), mcp.Description("WLED device IP address")),
				mcp.WithBoolean("enabled", mcp.Required(), mcp.Description("Enable Art-Net input")),
				mcp.WithNumber("universe", mcp.Description("Art-Net universe (default: 0)")),
				mcp.WithNumber("start_address", mcp.Description("DMX start address (default: 1)")),
			),
			Handler:             handleArtNetConfig,
			Category:            "wled",
			Subcategory:         "artnet",
			Tags:                []string{"wled", "artnet", "dmx", "config"},
			UseCases:            []string{"Configure Art-Net input", "Setup DMX control"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "wled",
			IsWrite:             true,
		},
	}
}

func handleDiscover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	subnet := tools.GetStringParam(req, "subnet")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	devices, err := client.DiscoverDevices(ctx, subnet)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# WLED Device Discovery\n\n")

	if len(devices) == 0 {
		sb.WriteString("No WLED devices found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Found:** %d devices\n\n", len(devices)))
	sb.WriteString("| Name | IP | LEDs | Version | Status |\n")
	sb.WriteString("|------|----|----- |---------|--------|\n")

	for _, d := range devices {
		status := "🔴 Off"
		if d.On {
			status = "🟢 On"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s |\n",
			d.Name, d.IP, d.LEDCount, d.Version, status))
	}

	return tools.TextResult(sb.String()), nil
}

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	device, err := client.GetDevice(ctx, ip)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# WLED: %s\n\n", device.Name))
	sb.WriteString(fmt.Sprintf("**IP:** %s\n", device.IP))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n", device.Version))
	sb.WriteString(fmt.Sprintf("**LEDs:** %d\n", device.LEDCount))

	status := "🔴 Off"
	if device.On {
		status = "🟢 On"
	}
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", status))
	sb.WriteString(fmt.Sprintf("**Brightness:** %d/255 (%.0f%%)\n", device.Brightness, float64(device.Brightness)/255*100))

	if len(device.Segments) > 0 {
		sb.WriteString("\n## Segments\n\n")
		for _, seg := range device.Segments {
			segStatus := "Off"
			if seg.On {
				segStatus = "On"
			}
			sb.WriteString(fmt.Sprintf("- Segment %d: LEDs %d-%d (%s, Effect %d)\n",
				seg.ID, seg.Start, seg.Stop, segStatus, seg.Effect))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handlePower(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")
	on := tools.GetBoolParam(req, "on", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.SetPower(ctx, ip, on); err != nil {
		return tools.ErrorResult(err), nil
	}

	status := "off"
	if on {
		status = "on"
	}
	return tools.TextResult(fmt.Sprintf("WLED %s turned %s.", ip, status)), nil
}

func handleBrightness(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")
	brightness := tools.GetIntParam(req, "brightness", 128)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.SetBrightness(ctx, ip, brightness); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("WLED %s brightness set to %d (%.0f%%).", ip, brightness, float64(brightness)/255*100)), nil
}

func handleEffect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")
	effectID := tools.GetIntParam(req, "effect_id", -1)
	effectName := tools.GetStringParam(req, "effect_name")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If name provided, find matching effect
	if effectName != "" && effectID < 0 {
		effects, err := client.GetEffects(ctx, ip)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		effectName = strings.ToLower(effectName)
		for _, e := range effects {
			if strings.Contains(strings.ToLower(e.Name), effectName) {
				effectID = e.ID
				break
			}
		}
		if effectID < 0 {
			return tools.ErrorResult(fmt.Errorf("effect not found: %s", effectName)), nil
		}
	}

	if err := client.SetEffect(ctx, ip, effectID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("WLED %s effect set to %d.", ip, effectID)), nil
}

func handleColor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")
	r := tools.GetIntParam(req, "red", 255)
	g := tools.GetIntParam(req, "green", 255)
	b := tools.GetIntParam(req, "blue", 255)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.SetColor(ctx, ip, r, g, b); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("WLED %s color set to RGB(%d, %d, %d).", ip, r, g, b)), nil
}

func handlePalette(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")
	paletteID := tools.GetIntParam(req, "palette_id", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.SetPalette(ctx, ip, paletteID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("WLED %s palette set to %d.", ip, paletteID)), nil
}

func handleEffectsList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	effects, err := client.GetEffects(ctx, ip)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# WLED Effects\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d effects\n\n", len(effects)))

	// Group in columns for readability
	sb.WriteString("| ID | Effect | ID | Effect |\n")
	sb.WriteString("|----|--------|----|--------|\n")

	for i := 0; i < len(effects); i += 2 {
		e1 := effects[i]
		if i+1 < len(effects) {
			e2 := effects[i+1]
			sb.WriteString(fmt.Sprintf("| %d | %s | %d | %s |\n", e1.ID, e1.Name, e2.ID, e2.Name))
		} else {
			sb.WriteString(fmt.Sprintf("| %d | %s | | |\n", e1.ID, e1.Name))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handlePalettesList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	palettes, err := client.GetPalettes(ctx, ip)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# WLED Palettes\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d palettes\n\n", len(palettes)))

	sb.WriteString("| ID | Palette | ID | Palette |\n")
	sb.WriteString("|----|---------|----|---------|\n")

	for i := 0; i < len(palettes); i += 2 {
		p1 := palettes[i]
		if i+1 < len(palettes) {
			p2 := palettes[i+1]
			sb.WriteString(fmt.Sprintf("| %d | %s | %d | %s |\n", p1.ID, p1.Name, p2.ID, p2.Name))
		} else {
			sb.WriteString(fmt.Sprintf("| %d | %s | | |\n", p1.ID, p1.Name))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handlePresetSave(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")
	presetID := tools.GetIntParam(req, "preset_id", 1)
	name := tools.GetStringParam(req, "name")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.SavePreset(ctx, ip, presetID, name); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Preset %d '%s' saved on %s.", presetID, name, ip)), nil
}

func handlePresetLoad(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")
	presetID := tools.GetIntParam(req, "preset_id", 1)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.LoadPreset(ctx, ip, presetID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Preset %d loaded on %s.", presetID, ip)), nil
}

func handleArtNetConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ip := tools.GetStringParam(req, "ip")
	enabled := tools.GetBoolParam(req, "enabled", false)
	universe := tools.GetIntParam(req, "universe", 0)
	startAddr := tools.GetIntParam(req, "start_address", 1)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	config := &clients.WLEDArtNetConfig{
		Enabled:   enabled,
		Universe:  universe,
		StartAddr: startAddr,
	}

	if err := client.ConfigureArtNet(ctx, ip, config); err != nil {
		return tools.ErrorResult(err), nil
	}

	status := "disabled"
	if enabled {
		status = fmt.Sprintf("enabled (Universe %d, Address %d)", universe, startAddr)
	}

	return tools.TextResult(fmt.Sprintf("Art-Net %s on %s.", status, ip)), nil
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
