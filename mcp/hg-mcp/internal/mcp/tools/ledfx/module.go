// Package ledfx provides LedFX audio-reactive LED control tools for hg-mcp.
package ledfx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for LedFX integration
type Module struct{}

func (m *Module) Name() string {
	return "ledfx"
}

func (m *Module) Description() string {
	return "LedFX audio-reactive LED control via REST API"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ledfx_status",
				mcp.WithDescription("Get LedFX connection status, version, and device counts."),
			),
			Handler:             handleStatus,
			Category:            "ledfx",
			Subcategory:         "status",
			Tags:                []string{"ledfx", "led", "status", "audio"},
			UseCases:            []string{"Check LedFX connection", "View system overview"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_devices",
				mcp.WithDescription("List physical LED devices configured in LedFX."),
			),
			Handler:             handleDevices,
			Category:            "ledfx",
			Subcategory:         "devices",
			Tags:                []string{"ledfx", "devices", "led", "wled"},
			UseCases:            []string{"View configured LED devices", "Check device status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_find_devices",
				mcp.WithDescription("Auto-discover WLED devices on the network."),
			),
			Handler:             handleFindDevices,
			Category:            "ledfx",
			Subcategory:         "devices",
			Tags:                []string{"ledfx", "discovery", "wled", "network"},
			UseCases:            []string{"Find WLED devices", "Setup new LED strips"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ledfx",
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_virtuals",
				mcp.WithDescription("List virtual LED devices with their active effects."),
			),
			Handler:             handleVirtuals,
			Category:            "ledfx",
			Subcategory:         "virtuals",
			Tags:                []string{"ledfx", "virtuals", "effects", "mapping"},
			UseCases:            []string{"View virtual devices", "Check active effects"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_virtual",
				mcp.WithDescription("Control a virtual LED device (activate, set effect, clear)."),
				mcp.WithString("id", mcp.Required(), mcp.Description("Virtual device ID")),
				mcp.WithString("action", mcp.Description("Action: activate, deactivate, set_effect, clear (default: show info)")),
				mcp.WithString("effect", mcp.Description("Effect type (for set_effect action)")),
				mcp.WithString("config", mcp.Description("JSON effect config (for set_effect action)")),
			),
			Handler:             handleVirtual,
			Category:            "ledfx",
			Subcategory:         "virtuals",
			Tags:                []string{"ledfx", "virtual", "control", "effect"},
			UseCases:            []string{"Apply effect to virtual", "Activate/deactivate LEDs"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ledfx",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_effects",
				mcp.WithDescription("List available effect types in LedFX."),
				mcp.WithString("category", mcp.Description("Filter by category (optional)")),
			),
			Handler:             handleEffects,
			Category:            "ledfx",
			Subcategory:         "effects",
			Tags:                []string{"ledfx", "effects", "list", "audio-reactive"},
			UseCases:            []string{"Browse available effects", "Find effect by category"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_effect",
				mcp.WithDescription("Apply or manage an effect on a virtual device."),
				mcp.WithString("virtual_id", mcp.Required(), mcp.Description("Virtual device ID")),
				mcp.WithString("effect", mcp.Required(), mcp.Description("Effect type to apply")),
				mcp.WithString("config", mcp.Description("JSON effect configuration")),
				mcp.WithString("preset", mcp.Description("Preset name to apply")),
			),
			Handler:             handleEffect,
			Category:            "ledfx",
			Subcategory:         "effects",
			Tags:                []string{"ledfx", "effect", "apply", "configure"},
			UseCases:            []string{"Apply audio-reactive effect", "Configure effect parameters"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ledfx",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_presets",
				mcp.WithDescription("List presets for an effect type."),
				mcp.WithString("effect", mcp.Required(), mcp.Description("Effect type to get presets for")),
			),
			Handler:             handlePresets,
			Category:            "ledfx",
			Subcategory:         "effects",
			Tags:                []string{"ledfx", "presets", "effects", "config"},
			UseCases:            []string{"Browse effect presets", "Find preset configurations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_scenes",
				mcp.WithDescription("List saved scenes in LedFX."),
			),
			Handler:             handleScenes,
			Category:            "ledfx",
			Subcategory:         "scenes",
			Tags:                []string{"ledfx", "scenes", "snapshots", "presets"},
			UseCases:            []string{"View saved scenes", "Browse show configurations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_scene",
				mcp.WithDescription("Manage LedFX scenes (activate, save, delete)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: activate, save, delete")),
				mcp.WithString("id", mcp.Description("Scene ID (for activate/delete)")),
				mcp.WithString("name", mcp.Description("Scene name (for save action)")),
			),
			Handler:             handleScene,
			Category:            "ledfx",
			Subcategory:         "scenes",
			Tags:                []string{"ledfx", "scene", "activate", "save"},
			UseCases:            []string{"Switch scenes", "Save current state", "Delete old scenes"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ledfx",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_audio",
				mcp.WithDescription("List or set audio input devices."),
				mcp.WithNumber("device_index", mcp.Description("Device index to set as active (omit to list devices)")),
			),
			Handler:             handleAudio,
			Category:            "ledfx",
			Subcategory:         "audio",
			Tags:                []string{"ledfx", "audio", "input", "microphone"},
			UseCases:            []string{"View audio inputs", "Change audio source"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_health",
				mcp.WithDescription("Get LedFX system health and recommendations."),
			),
			Handler:             handleHealth,
			Category:            "ledfx",
			Subcategory:         "health",
			Tags:                []string{"ledfx", "health", "monitoring", "status"},
			UseCases:            []string{"Check system health", "Diagnose issues"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ledfx",
		},
		// Enhanced LED control tools
		{
			Tool: mcp.NewTool("aftrs_ledfx_segments",
				mcp.WithDescription("List or configure LED segments for a device."),
				mcp.WithString("device_id", mcp.Description("Device ID (omit to list all segments)")),
				mcp.WithString("action", mcp.Description("Action: list, create, delete")),
				mcp.WithNumber("start", mcp.Description("Segment start LED index")),
				mcp.WithNumber("end", mcp.Description("Segment end LED index")),
			),
			Handler:             handleSegments,
			Category:            "ledfx",
			Subcategory:         "segments",
			Tags:                []string{"ledfx", "segments", "led", "zones"},
			UseCases:            []string{"Split LED strip into zones", "Configure segments"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ledfx",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_bpm",
				mcp.WithDescription("Get or set BPM for audio-reactive effects."),
				mcp.WithNumber("bpm", mcp.Description("BPM to set (omit to get current BPM)")),
				mcp.WithString("mode", mcp.Description("Sync mode: manual, auto, tap")),
			),
			Handler:             handleBPM,
			Category:            "ledfx",
			Subcategory:         "audio",
			Tags:                []string{"ledfx", "bpm", "tempo", "sync"},
			UseCases:            []string{"Sync LEDs to music tempo", "Manual BPM override"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_gradient",
				mcp.WithDescription("Create or apply a color gradient to effects."),
				mcp.WithString("virtual_id", mcp.Required(), mcp.Description("Virtual device ID")),
				mcp.WithString("colors", mcp.Required(), mcp.Description("Comma-separated hex colors (e.g., 'FF0000,00FF00,0000FF')")),
				mcp.WithString("name", mcp.Description("Save gradient with this name")),
			),
			Handler:             handleGradient,
			Category:            "ledfx",
			Subcategory:         "effects",
			Tags:                []string{"ledfx", "gradient", "colors", "palette"},
			UseCases:            []string{"Create custom color gradients", "Apply rainbow effects"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ledfx",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ledfx_solid_color",
				mcp.WithDescription("Set a solid color on a virtual device."),
				mcp.WithString("virtual_id", mcp.Required(), mcp.Description("Virtual device ID")),
				mcp.WithString("color", mcp.Required(), mcp.Description("Color (hex like 'FF0000' or name like 'red')")),
			),
			Handler:             handleSolidColor,
			Category:            "ledfx",
			Subcategory:         "effects",
			Tags:                []string{"ledfx", "color", "solid", "rgb"},
			UseCases:            []string{"Quick solid color", "Static LED color"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ledfx",
			IsWrite:             true,
		},
	}
}

var getClient = tools.LazyClient(clients.NewLedFXClient)

// handleStatus handles the aftrs_ledfx_status tool
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# LedFX Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Target:** %s:%d\n\n", status.Host, status.Port))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Install LedFX: `pip install ledfx`\n")
		sb.WriteString("2. Start LedFX: `ledfx`\n")
		sb.WriteString("3. Access web UI at http://localhost:8888\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export LEDFX_HOST=localhost\n")
		sb.WriteString("export LEDFX_PORT=8888\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	if status.Version != "" {
		sb.WriteString(fmt.Sprintf("**Version:** %s\n", status.Version))
	}
	sb.WriteString(fmt.Sprintf("**Host:** %s:%d\n\n", status.Host, status.Port))

	sb.WriteString("## Overview\n\n")
	sb.WriteString("| Metric | Count |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Devices | %d |\n", status.DeviceCount))
	sb.WriteString(fmt.Sprintf("| Virtuals | %d |\n", status.VirtualCount))
	sb.WriteString(fmt.Sprintf("| Active Effects | %d |\n", status.ActiveEffects))

	return tools.TextResult(sb.String()), nil
}

// handleDevices handles the aftrs_ledfx_devices tool
func handleDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	devices, err := client.GetDevices(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# LedFX Devices\n\n")

	if len(devices) == 0 {
		sb.WriteString("No devices configured.\n\n")
		sb.WriteString("Use `aftrs_ledfx_find_devices` to auto-discover WLED devices.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** devices:\n\n", len(devices)))
	sb.WriteString("| ID | Name | Type | Pixels |\n")
	sb.WriteString("|----|------|------|--------|\n")

	for _, dev := range devices {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d |\n", dev.ID, dev.Name, dev.Type, dev.PixelCount))
	}

	return tools.TextResult(sb.String()), nil
}

// handleFindDevices handles the aftrs_ledfx_find_devices tool
func handleFindDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	devices, err := client.FindDevices(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# WLED Device Discovery\n\n")

	if len(devices) == 0 {
		sb.WriteString("No WLED devices found on the network.\n\n")
		sb.WriteString("Make sure WLED devices are:\n")
		sb.WriteString("- Powered on and connected to the same network\n")
		sb.WriteString("- Running WLED firmware\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** WLED devices:\n\n", len(devices)))
	sb.WriteString("| ID | Name | Type |\n")
	sb.WriteString("|----|------|------|\n")

	for _, dev := range devices {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", dev.ID, dev.Name, dev.Type))
	}

	return tools.TextResult(sb.String()), nil
}

// handleVirtuals handles the aftrs_ledfx_virtuals tool
func handleVirtuals(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	virtuals, err := client.GetVirtuals(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# LedFX Virtual Devices\n\n")

	if len(virtuals) == 0 {
		sb.WriteString("No virtual devices configured.\n\n")
		sb.WriteString("Create virtuals in the LedFX web UI to map effects to physical devices.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** virtuals:\n\n", len(virtuals)))
	sb.WriteString("| ID | Name | Pixels | Active | Effect |\n")
	sb.WriteString("|----|------|--------|--------|--------|\n")

	for _, v := range virtuals {
		active := "No"
		if v.IsActive {
			active = "Yes"
		}
		effect := "-"
		if v.Effect != nil {
			effect = v.Effect.Type
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s |\n", v.ID, v.Name, v.PixelCount, active, effect))
	}

	return tools.TextResult(sb.String()), nil
}

// handleVirtual handles the aftrs_ledfx_virtual tool
func handleVirtual(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	virtualID, errResult := tools.RequireStringParam(req, "id")
	if errResult != nil {
		return errResult, nil
	}

	action := tools.GetStringParam(req, "action")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "activate":
		err = client.SetVirtualActive(ctx, virtualID, true)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Activated virtual: %s", virtualID)), nil

	case "deactivate":
		err = client.SetVirtualActive(ctx, virtualID, false)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Deactivated virtual: %s", virtualID)), nil

	case "set_effect":
		effectType, errResult := tools.RequireStringParam(req, "effect")
		if errResult != nil {
			return errResult, nil
		}
		configStr := tools.GetStringParam(req, "config")
		var config map[string]interface{}
		if configStr != "" {
			if err := json.Unmarshal([]byte(configStr), &config); err != nil {
				return tools.ErrorResult(fmt.Errorf("invalid config JSON: %w", err)), nil
			}
		}
		err = client.SetVirtualEffect(ctx, virtualID, effectType, config)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Applied effect '%s' to virtual: %s", effectType, virtualID)), nil

	case "clear":
		err = client.ClearVirtualEffect(ctx, virtualID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Cleared effect from virtual: %s", virtualID)), nil

	default:
		// Show info
		v, err := client.GetVirtual(ctx, virtualID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Virtual: %s\n\n", v.Name))
		sb.WriteString(fmt.Sprintf("**ID:** %s\n", v.ID))
		sb.WriteString(fmt.Sprintf("**Pixels:** %d\n", v.PixelCount))
		sb.WriteString(fmt.Sprintf("**Active:** %t\n", v.IsActive))

		if v.Effect != nil {
			sb.WriteString(fmt.Sprintf("\n## Active Effect\n\n"))
			sb.WriteString(fmt.Sprintf("**Type:** %s\n", v.Effect.Type))
			if v.Effect.Config != nil {
				configJSON, _ := json.MarshalIndent(v.Effect.Config, "", "  ")
				sb.WriteString(fmt.Sprintf("\n**Config:**\n```json\n%s\n```\n", string(configJSON)))
			}
		} else {
			sb.WriteString("\n*No active effect*\n")
		}

		return tools.TextResult(sb.String()), nil
	}
}

// handleEffects handles the aftrs_ledfx_effects tool
func handleEffects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category := tools.GetStringParam(req, "category")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	effects, err := client.GetEffects(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Filter by category if specified
	if category != "" {
		filtered := []clients.LedFXEffectType{}
		for _, e := range effects {
			if strings.EqualFold(e.Category, category) {
				filtered = append(filtered, e)
			}
		}
		effects = filtered
	}

	var sb strings.Builder
	sb.WriteString("# LedFX Effects\n\n")

	if len(effects) == 0 {
		sb.WriteString("No effects found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** effect types:\n\n", len(effects)))
	sb.WriteString("| ID | Name | Category |\n")
	sb.WriteString("|----|------|----------|\n")

	for _, e := range effects {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", e.ID, e.Name, e.Category))
	}

	return tools.TextResult(sb.String()), nil
}

// handleEffect handles the aftrs_ledfx_effect tool
func handleEffect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	virtualID, errResult := tools.RequireStringParam(req, "virtual_id")
	if errResult != nil {
		return errResult, nil
	}

	effectType, errResult := tools.RequireStringParam(req, "effect")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Parse config if provided
	configStr := tools.GetStringParam(req, "config")
	var config map[string]interface{}
	if configStr != "" {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid config JSON: %w", err)), nil
		}
	}

	// Apply the effect
	err = client.SetVirtualEffect(ctx, virtualID, effectType, config)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Applied effect '%s' to virtual '%s'", effectType, virtualID)), nil
}

// handlePresets handles the aftrs_ledfx_presets tool
func handlePresets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	effectType, errResult := tools.RequireStringParam(req, "effect")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	presets, err := client.GetEffectPresets(ctx, effectType)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Presets for %s\n\n", effectType))

	if len(presets) == 0 {
		sb.WriteString("No presets available.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** presets:\n\n", len(presets)))

	for _, p := range presets {
		sb.WriteString(fmt.Sprintf("### %s\n", p.Name))
		if p.Config != nil {
			configJSON, _ := json.MarshalIndent(p.Config, "", "  ")
			sb.WriteString(fmt.Sprintf("```json\n%s\n```\n\n", string(configJSON)))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleScenes handles the aftrs_ledfx_scenes tool
func handleScenes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	scenes, err := client.GetScenes(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# LedFX Scenes\n\n")

	if len(scenes) == 0 {
		sb.WriteString("No scenes saved.\n\n")
		sb.WriteString("Use `aftrs_ledfx_scene action=save name=\"My Scene\"` to save current state.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** scenes:\n\n", len(scenes)))
	sb.WriteString("| ID | Name | Virtuals |\n")
	sb.WriteString("|----|------|----------|\n")

	for _, s := range scenes {
		virtualCount := len(s.Virtuals)
		sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", s.ID, s.Name, virtualCount))
	}

	return tools.TextResult(sb.String()), nil
}

// handleScene handles the aftrs_ledfx_scene tool
func handleScene(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "activate":
		sceneID, errResult := tools.RequireStringParam(req, "id")
		if errResult != nil {
			return errResult, nil
		}
		err = client.ActivateScene(ctx, sceneID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Activated scene: %s", sceneID)), nil

	case "save":
		name, errResult := tools.RequireStringParam(req, "name")
		if errResult != nil {
			return errResult, nil
		}
		scene, err := client.SaveScene(ctx, name)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Saved scene '%s' (ID: %s)", name, scene.ID)), nil

	case "delete":
		sceneID, errResult := tools.RequireStringParam(req, "id")
		if errResult != nil {
			return errResult, nil
		}
		err = client.DeleteScene(ctx, sceneID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Deleted scene: %s", sceneID)), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use activate, save, or delete)", action)), nil
	}
}

// handleAudio handles the aftrs_ledfx_audio tool
func handleAudio(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceIndex := tools.GetIntParam(req, "device_index", -1)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If device index provided, set it
	if deviceIndex >= 0 {
		err = client.SetAudioDevice(ctx, deviceIndex)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set audio device to index: %d", deviceIndex)), nil
	}

	// Otherwise list devices
	devices, err := client.GetAudioDevices(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# LedFX Audio Devices\n\n")

	if len(devices) == 0 {
		sb.WriteString("No audio devices found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** audio devices:\n\n", len(devices)))
	sb.WriteString("| Index | Name | Active |\n")
	sb.WriteString("|-------|------|--------|\n")

	for _, d := range devices {
		active := ""
		if d.Active {
			active = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", d.Index, d.Name, active))
	}

	sb.WriteString("\n*Use `device_index` parameter to change the active audio input.*\n")

	return tools.TextResult(sb.String()), nil
}

// handleHealth handles the aftrs_ledfx_health tool
func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# LedFX Health\n\n")

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100\n", health.Score))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))
	sb.WriteString(fmt.Sprintf("**Connected:** %t\n\n", health.Connected))

	sb.WriteString("## Metrics\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Devices | %d |\n", health.DeviceCount))
	sb.WriteString(fmt.Sprintf("| Virtuals | %d |\n", health.VirtualCount))

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

// handleSegments handles the aftrs_ledfx_segments tool
func handleSegments(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	deviceID := tools.GetStringParam(req, "device_id")
	action := tools.GetStringParam(req, "action")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "create":
		start := tools.GetIntParam(req, "start", 0)
		end := tools.GetIntParam(req, "end", 0)
		if start >= end {
			return tools.ErrorResult(fmt.Errorf("end must be greater than start")), nil
		}
		err = client.CreateSegment(ctx, deviceID, start, end)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Created segment: LEDs %d-%d on device %s", start, end, deviceID)), nil

	case "delete":
		err = client.DeleteSegment(ctx, deviceID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Deleted segment from device: %s", deviceID)), nil

	default:
		// List segments
		segments, err := client.GetSegments(ctx, deviceID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		var sb strings.Builder
		sb.WriteString("# LED Segments\n\n")

		if len(segments) == 0 {
			sb.WriteString("No segments configured.\n")
			return tools.TextResult(sb.String()), nil
		}

		sb.WriteString("| Device | Start | End | Pixels |\n")
		sb.WriteString("|--------|-------|-----|--------|\n")
		for _, seg := range segments {
			sb.WriteString(fmt.Sprintf("| %s | %d | %d | %d |\n",
				seg.DeviceID, seg.Start, seg.End, seg.End-seg.Start))
		}
		return tools.TextResult(sb.String()), nil
	}
}

// handleBPM handles the aftrs_ledfx_bpm tool
func handleBPM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	bpm := tools.GetIntParam(req, "bpm", -1)
	mode := tools.GetStringParam(req, "mode")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If BPM provided, set it
	if bpm > 0 {
		err = client.SetBPM(ctx, float64(bpm))
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set BPM to %d", bpm)), nil
	}

	// If mode provided, set sync mode
	if mode != "" {
		err = client.SetBPMMode(ctx, mode)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set BPM sync mode to: %s", mode)), nil
	}

	// Get current BPM
	currentBPM, err := client.GetBPM(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Current BPM: %.1f", currentBPM)), nil
}

// handleGradient handles the aftrs_ledfx_gradient tool
func handleGradient(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	virtualID, errResult := tools.RequireStringParam(req, "virtual_id")
	if errResult != nil {
		return errResult, nil
	}

	colorsStr, errResult := tools.RequireStringParam(req, "colors")
	if errResult != nil {
		return errResult, nil
	}

	name := tools.GetStringParam(req, "name")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Parse colors
	colors := strings.Split(colorsStr, ",")
	for i, c := range colors {
		colors[i] = strings.TrimSpace(c)
	}

	// Apply gradient
	err = client.SetGradient(ctx, virtualID, colors, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Applied gradient to %s:\n", virtualID))
	for i, c := range colors {
		sb.WriteString(fmt.Sprintf("- Color %d: #%s\n", i+1, c))
	}
	if name != "" {
		sb.WriteString(fmt.Sprintf("\nSaved as: %s\n", name))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSolidColor handles the aftrs_ledfx_solid_color tool
func handleSolidColor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	virtualID, errResult := tools.RequireStringParam(req, "virtual_id")
	if errResult != nil {
		return errResult, nil
	}

	color, errResult := tools.RequireStringParam(req, "color")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Apply solid color effect
	err = client.SetSolidColor(ctx, virtualID, color)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Set solid color on %s: %s", virtualID, color)), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
