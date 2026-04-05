// Package grandma3 provides grandMA3 lighting console control tools for hg-mcp.
package grandma3

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for grandMA3 integration
type Module struct{}

func (m *Module) Name() string {
	return "grandma3"
}

func (m *Module) Description() string {
	return "grandMA3 lighting console control via OSC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_gma3_status",
				mcp.WithDescription("Get grandMA3 console connection status."),
			),
			Handler:     handleStatus,
			Category:    "grandma3",
			Subcategory: "status",
			Tags:        []string{"grandma3", "lighting", "status", "osc"},
			UseCases:    []string{"Check console connection", "Verify OSC settings"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_command",
				mcp.WithDescription("Send a command line instruction to grandMA3."),
				mcp.WithString("command", mcp.Required(), mcp.Description("Command to execute (e.g., 'Go+ Sequence 1', 'Fixture 1 At 50')")),
			),
			Handler:     handleCommand,
			Category:    "grandma3",
			Subcategory: "command",
			Tags:        []string{"grandma3", "command", "raw"},
			UseCases:    []string{"Execute any grandMA3 command", "Advanced control"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_executor",
				mcp.WithDescription("Control an executor (trigger, stop, set fader)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: go, stop, flash_on, flash_off, fader")),
				mcp.WithNumber("page", mcp.Description("Executor page (default: 1)")),
				mcp.WithNumber("executor", mcp.Required(), mcp.Description("Executor number")),
				mcp.WithNumber("value", mcp.Description("Fader value 0-100 (for fader action)")),
			),
			Handler:     handleExecutor,
			Category:    "grandma3",
			Subcategory: "executor",
			Tags:        []string{"grandma3", "executor", "playback", "fader"},
			UseCases:    []string{"Trigger cue", "Adjust fader", "Flash effect"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_sequence",
				mcp.WithDescription("Control a sequence (go, stop, pause, go to cue)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: go, go_next, go_prev, stop, pause, goto_cue")),
				mcp.WithNumber("sequence", mcp.Required(), mcp.Description("Sequence number")),
				mcp.WithNumber("cue", mcp.Description("Cue number (for goto_cue action)")),
			),
			Handler:     handleSequence,
			Category:    "grandma3",
			Subcategory: "sequence",
			Tags:        []string{"grandma3", "sequence", "cue", "playback"},
			UseCases:    []string{"Advance cue", "Jump to specific cue", "Stop sequence"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_master",
				mcp.WithDescription("Control grand master or speed masters."),
				mcp.WithString("type", mcp.Required(), mcp.Description("Master type: grand, speed")),
				mcp.WithNumber("number", mcp.Description("Speed master number (for speed type)")),
				mcp.WithNumber("value", mcp.Description("Value 0-100 (percentage)")),
				mcp.WithNumber("bpm", mcp.Description("BPM value (for speed master)")),
				mcp.WithBoolean("tap", mcp.Description("Tap tempo (for speed master)")),
			),
			Handler:     handleMaster,
			Category:    "grandma3",
			Subcategory: "master",
			Tags:        []string{"grandma3", "master", "dimmer", "speed", "bpm"},
			UseCases:    []string{"Adjust grand master", "Set BPM", "Tap tempo"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_fixture",
				mcp.WithDescription("Control fixtures (select, set dimmer, set attribute)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: select, dimmer, attribute, clear")),
				mcp.WithString("fixtures", mcp.Description("Fixture selection (e.g., '1 Thru 10', '1+2+3')")),
				mcp.WithString("attribute", mcp.Description("Attribute name (for attribute action)")),
				mcp.WithNumber("value", mcp.Description("Value 0-100 (for dimmer/attribute)")),
			),
			Handler:     handleFixture,
			Category:    "grandma3",
			Subcategory: "fixture",
			Tags:        []string{"grandma3", "fixture", "dimmer", "attribute"},
			UseCases:    []string{"Select fixtures", "Set dimmer levels", "Adjust colors"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_preset",
				mcp.WithDescription("Store or recall presets."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: store, call")),
				mcp.WithString("type", mcp.Required(), mcp.Description("Preset type (e.g., Dimmer, Color, Position, Gobo, All)")),
				mcp.WithNumber("number", mcp.Required(), mcp.Description("Preset number")),
			),
			Handler:     handlePreset,
			Category:    "grandma3",
			Subcategory: "preset",
			Tags:        []string{"grandma3", "preset", "store", "recall"},
			UseCases:    []string{"Save looks", "Recall presets"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_macro",
				mcp.WithDescription("Trigger a macro."),
				mcp.WithNumber("macro", mcp.Required(), mcp.Description("Macro number")),
			),
			Handler:     handleMacro,
			Category:    "grandma3",
			Subcategory: "macro",
			Tags:        []string{"grandma3", "macro", "automation"},
			UseCases:    []string{"Run automated sequences", "Trigger complex actions"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_timecode",
				mcp.WithDescription("Control timecode playback."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, goto")),
				mcp.WithNumber("slot", mcp.Description("Timecode slot number (default: 1)")),
				mcp.WithString("time", mcp.Description("Timecode position HH:MM:SS.FF (for goto action)")),
			),
			Handler:     handleTimecode,
			Category:    "grandma3",
			Subcategory: "timecode",
			Tags:        []string{"grandma3", "timecode", "sync", "playback"},
			UseCases:    []string{"Start timecode show", "Jump to position"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_blackout",
				mcp.WithDescription("Toggle blackout on/off."),
				mcp.WithBoolean("on", mcp.Description("True for blackout, false to release (default: toggle)")),
			),
			Handler:     handleBlackout,
			Category:    "grandma3",
			Subcategory: "control",
			Tags:        []string{"grandma3", "blackout", "safety"},
			UseCases:    []string{"Emergency blackout", "Show start/end"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_clear",
				mcp.WithDescription("Clear programmer or all playback."),
				mcp.WithString("scope", mcp.Description("Scope: programmer (default), all")),
			),
			Handler:     handleClear,
			Category:    "grandma3",
			Subcategory: "control",
			Tags:        []string{"grandma3", "clear", "programmer"},
			UseCases:    []string{"Clear programmer", "Reset all playback"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_gma3_health",
				mcp.WithDescription("Get grandMA3 system health and recommendations."),
			),
			Handler:     handleHealth,
			Category:    "grandma3",
			Subcategory: "health",
			Tags:        []string{"grandma3", "health", "monitoring", "status"},
			UseCases:    []string{"Check system health", "Diagnose issues"},
			Complexity:  tools.ComplexityModerate,
		},
	}

	// Apply circuit breaker to all tools — network-dependent (WebSocket)
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "grandma3"
	}

	return allTools
}

// getClient creates a new grandMA3 client
func getClient() (*clients.GrandMA3Client, error) {
	return clients.NewGrandMA3Client()
}

// handleStatus handles the aftrs_gma3_status tool
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
	sb.WriteString("# grandMA3 Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Target:** %s:%d (%s)\n\n", status.Host, status.Port, status.Protocol))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open grandMA3 console or onPC\n")
		sb.WriteString("2. Go to Menu > In & Out > OSC\n")
		sb.WriteString("3. Enable Input and configure port (default: 8000)\n")
		sb.WriteString("4. Set prefix if needed (or leave empty)\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export GRANDMA3_HOST=192.168.1.100\n")
		sb.WriteString("export GRANDMA3_OSC_PORT=8000\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**Host:** %s\n", status.Host))
	sb.WriteString(fmt.Sprintf("**Port:** %d\n", status.Port))
	sb.WriteString(fmt.Sprintf("**Protocol:** %s\n", status.Protocol))

	return tools.TextResult(sb.String()), nil
}

// handleCommand handles the aftrs_gma3_command tool
func handleCommand(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	command, errResult := tools.RequireStringParam(req, "command")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SendCommand(ctx, command)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Sent command: `%s`", command)), nil
}

// handleExecutor handles the aftrs_gma3_executor tool
func handleExecutor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)
	executor, errResult := tools.RequireIntParam(req, "executor")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "go":
		err = client.GoExecutor(ctx, page, executor)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Triggered Go on Executor %d.%d", page, executor)), nil

	case "stop":
		err = client.StopExecutor(ctx, page, executor)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Stopped Executor %d.%d", page, executor)), nil

	case "flash_on":
		err = client.FlashExecutor(ctx, page, executor, true)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Flash ON Executor %d.%d", page, executor)), nil

	case "flash_off":
		err = client.FlashExecutor(ctx, page, executor, false)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Flash OFF Executor %d.%d", page, executor)), nil

	case "fader":
		value := float32(tools.GetIntParam(req, "value", 0))
		err = client.SetExecutorFader(ctx, page, executor, value)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set Executor %d.%d fader to %.0f%%", page, executor, value)), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use go, stop, flash_on, flash_off, fader)", action)), nil
	}
}

// handleSequence handles the aftrs_gma3_sequence tool
func handleSequence(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	sequence, errResult := tools.RequireIntParam(req, "sequence")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "go", "go_next":
		err = client.GoNextCue(ctx, sequence)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Go+ on Sequence %d", sequence)), nil

	case "go_prev":
		err = client.GoPrevCue(ctx, sequence)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Go- on Sequence %d", sequence)), nil

	case "stop":
		err = client.StopSequence(ctx, sequence)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Stopped Sequence %d", sequence)), nil

	case "pause":
		err = client.PauseSequence(ctx, sequence)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Paused Sequence %d", sequence)), nil

	case "goto_cue":
		cue := float64(tools.GetIntParam(req, "cue", 0))
		if cue == 0 {
			return tools.ErrorResult(fmt.Errorf("cue number is required for goto_cue action")), nil
		}
		err = client.GoCue(ctx, sequence, cue)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Jumped to Cue %.0f in Sequence %d", cue, sequence)), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use go, go_next, go_prev, stop, pause, goto_cue)", action)), nil
	}
}

// handleMaster handles the aftrs_gma3_master tool
func handleMaster(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	masterType, errResult := tools.RequireStringParam(req, "type")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch masterType {
	case "grand":
		value := float32(tools.GetIntParam(req, "value", 100))
		err = client.SetGrandMaster(ctx, value)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set Grand Master to %.0f%%", value)), nil

	case "speed":
		number := tools.GetIntParam(req, "number", 1)
		tap := tools.GetBoolParam(req, "tap", false)

		if tap {
			err = client.TapTempo(ctx, number)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			return tools.TextResult(fmt.Sprintf("Tapped tempo on Speed Master %d", number)), nil
		}

		bpm := float32(tools.GetIntParam(req, "bpm", 0))
		if bpm > 0 {
			err = client.SetBPMMaster(ctx, number, bpm)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			return tools.TextResult(fmt.Sprintf("Set Speed Master %d to %.1f BPM", number, bpm)), nil
		}

		value := float32(tools.GetIntParam(req, "value", 100))
		err = client.SetSpeedMaster(ctx, number, value)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set Speed Master %d to %.0f%%", number, value)), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid type: %s (use grand or speed)", masterType)), nil
	}
}

// handleFixture handles the aftrs_gma3_fixture tool
func handleFixture(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "select":
		fixtures, errResult := tools.RequireStringParam(req, "fixtures")
		if errResult != nil {
			return errResult, nil
		}
		err = client.SelectFixtures(ctx, fixtures)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Selected Fixture %s", fixtures)), nil

	case "dimmer":
		value := float32(tools.GetIntParam(req, "value", 0))
		err = client.SetDimmer(ctx, value)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set dimmer to %.0f%%", value)), nil

	case "attribute":
		attribute, errResult := tools.RequireStringParam(req, "attribute")
		if errResult != nil {
			return errResult, nil
		}
		value := float32(tools.GetIntParam(req, "value", 0))
		err = client.SetFixtureAttribute(ctx, attribute, value)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set %s to %.0f%%", attribute, value)), nil

	case "clear":
		err = client.ClearProgrammer(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Cleared programmer"), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use select, dimmer, attribute, clear)", action)), nil
	}
}

// handlePreset handles the aftrs_gma3_preset tool
func handlePreset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	presetType, errResult2 := tools.RequireStringParam(req, "type")
	if errResult2 != nil {
		return errResult2, nil
	}

	number, errResult := tools.RequireIntParam(req, "number")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "store":
		err = client.StorePreset(ctx, presetType, number)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Stored Preset %d.%s", number, presetType)), nil

	case "call":
		err = client.CallPreset(ctx, presetType, number)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Called Preset %d.%s", number, presetType)), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use store or call)", action)), nil
	}
}

// handleMacro handles the aftrs_gma3_macro tool
func handleMacro(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	macro, errResult := tools.RequireIntParam(req, "macro")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.CallMacro(ctx, macro)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Triggered Macro %d", macro)), nil
}

// handleTimecode handles the aftrs_gma3_timecode tool
func handleTimecode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	slot := tools.GetIntParam(req, "slot", 1)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err = client.SetTimecodeEnabled(ctx, slot, true)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Started Timecode %d", slot)), nil

	case "stop":
		err = client.SetTimecodeEnabled(ctx, slot, false)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Stopped Timecode %d", slot)), nil

	case "goto":
		timeStr, errResult := tools.RequireStringParam(req, "time")
		if errResult != nil {
			return errResult, nil
		}
		// Parse HH:MM:SS.FF format
		var hours, minutes, seconds, frames int
		_, parseErr := fmt.Sscanf(timeStr, "%d:%d:%d.%d", &hours, &minutes, &seconds, &frames)
		if parseErr != nil {
			// Try without frames
			_, parseErr = fmt.Sscanf(timeStr, "%d:%d:%d", &hours, &minutes, &seconds)
			if parseErr != nil {
				return tools.ErrorResult(fmt.Errorf("invalid time format: %s (use HH:MM:SS or HH:MM:SS.FF)", timeStr)), nil
			}
		}
		err = client.JumpTimecode(ctx, slot, hours, minutes, seconds, frames)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Jumped Timecode %d to %s", slot, timeStr)), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use start, stop, goto)", action)), nil
	}
}

// handleBlackout handles the aftrs_gma3_blackout tool
func handleBlackout(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	on := tools.GetBoolParam(req, "on", true)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.Blackout(ctx, on)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status := "ON"
	if !on {
		status = "OFF"
	}
	return tools.TextResult(fmt.Sprintf("Blackout %s", status)), nil
}

// handleClear handles the aftrs_gma3_clear tool
func handleClear(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scope := tools.OptionalStringParam(req, "scope", "programmer")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch scope {
	case "programmer":
		err = client.ClearProgrammer(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Cleared programmer"), nil

	case "all":
		err = client.ClearAll(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Cleared all playback and programmer"), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid scope: %s (use programmer or all)", scope)), nil
	}
}

// handleHealth handles the aftrs_gma3_health tool
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
	sb.WriteString("# grandMA3 Health\n\n")

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100\n", health.Score))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))
	sb.WriteString(fmt.Sprintf("**Connected:** %t\n", health.Connected))

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

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
