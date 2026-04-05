// Package midi provides MIDI control tools for hg-mcp.
package midi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for MIDI integration
type Module struct{}

func (m *Module) Name() string {
	return "midi"
}

func (m *Module) Description() string {
	return "MIDI control for AV equipment and software"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_midi_status",
				mcp.WithDescription("Get MIDI system status including connected devices."),
			),
			Handler:             handleStatus,
			Category:            "midi",
			Subcategory:         "status",
			Tags:                []string{"midi", "status", "devices"},
			UseCases:            []string{"Check MIDI devices", "View MIDI configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
		},
		{
			Tool: mcp.NewTool("aftrs_midi_devices",
				mcp.WithDescription("List all MIDI input and output devices."),
			),
			Handler:             handleDevices,
			Category:            "midi",
			Subcategory:         "devices",
			Tags:                []string{"midi", "devices", "list"},
			UseCases:            []string{"View connected MIDI devices", "Find device names"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
		},
		{
			Tool: mcp.NewTool("aftrs_midi_note",
				mcp.WithDescription("Send a MIDI note on/off message."),
				mcp.WithNumber("channel", mcp.Required(), mcp.Description("MIDI channel (1-16)")),
				mcp.WithNumber("note", mcp.Required(), mcp.Description("Note number (0-127, C4=60)")),
				mcp.WithNumber("velocity", mcp.Description("Velocity (0-127, default 100)")),
				mcp.WithBoolean("off", mcp.Description("Send note off instead of note on")),
				mcp.WithString("device", mcp.Description("Output device name (uses default if omitted)")),
			),
			Handler:             handleNote,
			Category:            "midi",
			Subcategory:         "messages",
			Tags:                []string{"midi", "note", "trigger"},
			UseCases:            []string{"Trigger MIDI notes", "Control instruments"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_cc",
				mcp.WithDescription("Send a MIDI Control Change message."),
				mcp.WithNumber("channel", mcp.Required(), mcp.Description("MIDI channel (1-16)")),
				mcp.WithNumber("controller", mcp.Required(), mcp.Description("Controller number (0-127)")),
				mcp.WithNumber("value", mcp.Required(), mcp.Description("Value (0-127)")),
				mcp.WithString("device", mcp.Description("Output device name (uses default if omitted)")),
			),
			Handler:             handleCC,
			Category:            "midi",
			Subcategory:         "messages",
			Tags:                []string{"midi", "cc", "control", "fader"},
			UseCases:            []string{"Control parameters", "Send fader values"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_program",
				mcp.WithDescription("Send a MIDI Program Change message."),
				mcp.WithNumber("channel", mcp.Required(), mcp.Description("MIDI channel (1-16)")),
				mcp.WithNumber("program", mcp.Required(), mcp.Description("Program number (0-127)")),
				mcp.WithString("device", mcp.Description("Output device name (uses default if omitted)")),
			),
			Handler:             handleProgram,
			Category:            "midi",
			Subcategory:         "messages",
			Tags:                []string{"midi", "program", "preset", "patch"},
			UseCases:            []string{"Change presets", "Select patches"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_pitch",
				mcp.WithDescription("Send a MIDI Pitch Bend message."),
				mcp.WithNumber("channel", mcp.Required(), mcp.Description("MIDI channel (1-16)")),
				mcp.WithNumber("value", mcp.Required(), mcp.Description("Pitch bend value (-8192 to 8191, 0=center)")),
				mcp.WithString("device", mcp.Description("Output device name (uses default if omitted)")),
			),
			Handler:             handlePitch,
			Category:            "midi",
			Subcategory:         "messages",
			Tags:                []string{"midi", "pitch", "bend"},
			UseCases:            []string{"Pitch bend control", "Modulation"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_panic",
				mcp.WithDescription("Send MIDI panic (all notes off) to stop stuck notes."),
				mcp.WithString("device", mcp.Description("Output device name (uses default if omitted)")),
			),
			Handler:             handlePanic,
			Category:            "midi",
			Subcategory:         "control",
			Tags:                []string{"midi", "panic", "stop", "reset"},
			UseCases:            []string{"Stop stuck notes", "Emergency reset"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_transport",
				mcp.WithDescription("Send MIDI transport control (start, stop, continue)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, or continue")),
				mcp.WithString("device", mcp.Description("Output device name (uses default if omitted)")),
			),
			Handler:             handleTransport,
			Category:            "midi",
			Subcategory:         "transport",
			Tags:                []string{"midi", "transport", "start", "stop"},
			UseCases:            []string{"Control playback", "Sync devices"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_mappings",
				mcp.WithDescription("List configured MIDI control mappings."),
			),
			Handler:             handleMappings,
			Category:            "midi",
			Subcategory:         "mappings",
			Tags:                []string{"midi", "mapping", "config"},
			UseCases:            []string{"View control mappings", "Check configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
		},
		{
			Tool: mcp.NewTool("aftrs_midi_learn",
				mcp.WithDescription("Enter MIDI learn mode to capture incoming messages."),
				mcp.WithNumber("timeout", mcp.Description("Timeout in seconds (default 10)")),
			),
			Handler:             handleLearn,
			Category:            "midi",
			Subcategory:         "mappings",
			Tags:                []string{"midi", "learn", "capture"},
			UseCases:            []string{"Capture MIDI input", "Create mappings"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "midi",
		},
		{
			Tool: mcp.NewTool("aftrs_midi_health",
				mcp.WithDescription("Get MIDI system health and recommendations."),
			),
			Handler:             handleHealth,
			Category:            "midi",
			Subcategory:         "health",
			Tags:                []string{"midi", "health", "monitoring"},
			UseCases:            []string{"Check system health", "Diagnose issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
		},
		// New Phase 16B tools for MIDI→tool mapping
		{
			Tool: mcp.NewTool("aftrs_midi_map_create",
				mcp.WithDescription("Create a MIDI→tool mapping. Maps MIDI CC/note/program to invoke a tool with optional value transformation."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Human-readable name for this mapping")),
				mcp.WithNumber("channel", mcp.Required(), mcp.Description("MIDI channel (1-16)")),
				mcp.WithString("type", mcp.Required(), mcp.Description("Message type: cc, note, or program")),
				mcp.WithNumber("number", mcp.Required(), mcp.Description("CC number, note number, or program number (0-127)")),
				mcp.WithString("tool_name", mcp.Required(), mcp.Description("Target tool to invoke (e.g., aftrs_resolume_bpm)")),
				mcp.WithObject("parameters", mcp.Description("Static parameters to pass to the tool")),
				mcp.WithString("target_param", mcp.Description("Parameter name to receive the mapped MIDI value")),
				mcp.WithNumber("output_min", mcp.Description("Output range minimum (default 0)")),
				mcp.WithNumber("output_max", mcp.Description("Output range maximum (default 127)")),
				mcp.WithBoolean("invert", mcp.Description("Invert the value mapping")),
			),
			Handler:             handleMapCreate,
			Category:            "midi",
			Subcategory:         "mappings",
			Tags:                []string{"midi", "mapping", "tool", "automation"},
			UseCases:            []string{"Map faders to tool parameters", "Create MIDI automation"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_map_delete",
				mcp.WithDescription("Delete a MIDI→tool mapping by ID."),
				mcp.WithString("id", mcp.Required(), mcp.Description("Mapping ID to delete")),
			),
			Handler:             handleMapDelete,
			Category:            "midi",
			Subcategory:         "mappings",
			Tags:                []string{"midi", "mapping", "delete"},
			UseCases:            []string{"Remove mappings", "Clean up configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_map_list",
				mcp.WithDescription("List all MIDI→tool mappings with their configurations."),
			),
			Handler:             handleMapList,
			Category:            "midi",
			Subcategory:         "mappings",
			Tags:                []string{"midi", "mapping", "list"},
			UseCases:            []string{"View current mappings", "Check configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
		},
		{
			Tool: mcp.NewTool("aftrs_midi_profile_save",
				mcp.WithDescription("Save current MIDI mappings to a named profile for later recall."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Profile name")),
				mcp.WithString("description", mcp.Description("Profile description")),
			),
			Handler:             handleProfileSave,
			Category:            "midi",
			Subcategory:         "profiles",
			Tags:                []string{"midi", "profile", "save", "backup"},
			UseCases:            []string{"Save mapping configurations", "Create presets"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_profile_load",
				mcp.WithDescription("Load MIDI mappings from a saved profile, replacing current mappings."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Profile name to load")),
			),
			Handler:             handleProfileLoad,
			Category:            "midi",
			Subcategory:         "profiles",
			Tags:                []string{"midi", "profile", "load", "recall"},
			UseCases:            []string{"Switch between configurations", "Recall presets"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_midi_profiles",
				mcp.WithDescription("List all saved MIDI mapping profiles."),
			),
			Handler:             handleProfiles,
			Category:            "midi",
			Subcategory:         "profiles",
			Tags:                []string{"midi", "profile", "list"},
			UseCases:            []string{"View saved profiles", "Find presets"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "midi",
		},
		{
			Tool: mcp.NewTool("aftrs_midi_batch_cc",
				mcp.WithDescription("Send multiple MIDI CC messages in a single call. Reduces latency for multi-parameter updates."),
				mcp.WithString("messages",
					mcp.Required(),
					mcp.Description("JSON array of CC messages: [{\"channel\":1,\"controller\":7,\"value\":100},{\"channel\":1,\"controller\":10,\"value\":64}]"),
				),
				mcp.WithString("device",
					mcp.Description("Output device name (uses default if omitted)"),
				),
			),
			Handler:             handleBatchCC,
			Category:            "midi",
			Subcategory:         "output",
			Tags:                []string{"midi", "cc", "batch", "control"},
			UseCases:            []string{"Send multiple CCs at once", "Batch parameter updates"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "midi",
			IsWrite:             true,
		},
	}
}

var getClient = tools.LazyClient(clients.NewMIDIClient)

// handleStatus handles the aftrs_midi_status tool
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
	sb.WriteString("# MIDI Status\n\n")

	totalDevices := len(status.InputDevices) + len(status.OutputDevices)
	if totalDevices == 0 {
		sb.WriteString("**Devices:** No MIDI devices detected\n\n")
		sb.WriteString("## Setup\n\n")
		sb.WriteString("Connect a MIDI controller or interface to use MIDI tools.\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export MIDI_OUTPUT=\"My MIDI Device\"\n")
		sb.WriteString("export MIDI_INPUT=\"My MIDI Controller\"\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Input Devices:** %d\n", len(status.InputDevices)))
	sb.WriteString(fmt.Sprintf("**Output Devices:** %d\n", len(status.OutputDevices)))

	if status.DefaultInput != "" {
		sb.WriteString(fmt.Sprintf("**Default Input:** %s\n", status.DefaultInput))
	}
	if status.DefaultOutput != "" {
		sb.WriteString(fmt.Sprintf("**Default Output:** %s\n", status.DefaultOutput))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDevices handles the aftrs_midi_devices tool
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
	sb.WriteString("# MIDI Devices\n\n")

	if len(devices) == 0 {
		sb.WriteString("No MIDI devices detected.\n\n")
		sb.WriteString("*Note: Connect a MIDI interface or controller to see devices.*\n")
		return tools.TextResult(sb.String()), nil
	}

	// Separate inputs and outputs
	var inputs, outputs []clients.MIDIDevice
	for _, d := range devices {
		if d.Type == "input" {
			inputs = append(inputs, d)
		} else {
			outputs = append(outputs, d)
		}
	}

	if len(inputs) > 0 {
		sb.WriteString("## Input Devices\n\n")
		sb.WriteString("| ID | Name | Connected |\n")
		sb.WriteString("|----|------|----------|\n")
		for _, d := range inputs {
			conn := "Yes"
			if !d.Connected {
				conn = "No"
			}
			sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", d.ID, d.Name, conn))
		}
		sb.WriteString("\n")
	}

	if len(outputs) > 0 {
		sb.WriteString("## Output Devices\n\n")
		sb.WriteString("| ID | Name | Connected |\n")
		sb.WriteString("|----|------|----------|\n")
		for _, d := range outputs {
			conn := "Yes"
			if !d.Connected {
				conn = "No"
			}
			sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", d.ID, d.Name, conn))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleNote handles the aftrs_midi_note tool
func handleNote(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channel := tools.GetIntParam(req, "channel", 0)
	note := tools.GetIntParam(req, "note", -1)
	velocity := tools.GetIntParam(req, "velocity", 100)
	off := tools.GetBoolParam(req, "off", false)
	device := tools.GetStringParam(req, "device")

	if channel == 0 || note < 0 {
		return tools.ErrorResult(fmt.Errorf("channel and note are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if off {
		err = client.SendNoteOff(ctx, channel, note, device)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Sent Note Off: ch%d note%d", channel, note)), nil
	}

	err = client.SendNoteOn(ctx, channel, note, velocity, device)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Sent Note On: ch%d note%d vel%d", channel, note, velocity)), nil
}

// handleCC handles the aftrs_midi_cc tool
func handleCC(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channel := tools.GetIntParam(req, "channel", 0)
	controller := tools.GetIntParam(req, "controller", -1)
	value := tools.GetIntParam(req, "value", -1)
	device := tools.GetStringParam(req, "device")

	if channel == 0 || controller < 0 || value < 0 {
		return tools.ErrorResult(fmt.Errorf("channel, controller, and value are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SendCC(ctx, channel, controller, value, device)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Sent CC: ch%d cc%d=%d", channel, controller, value)), nil
}

// handleProgram handles the aftrs_midi_program tool
func handleProgram(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channel := tools.GetIntParam(req, "channel", 0)
	program := tools.GetIntParam(req, "program", -1)
	device := tools.GetStringParam(req, "device")

	if channel == 0 || program < 0 {
		return tools.ErrorResult(fmt.Errorf("channel and program are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SendProgramChange(ctx, channel, program, device)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Sent Program Change: ch%d prog%d", channel, program)), nil
}

// handlePitch handles the aftrs_midi_pitch tool
func handlePitch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channel := tools.GetIntParam(req, "channel", 0)
	value := tools.GetIntParam(req, "value", 0)
	device := tools.GetStringParam(req, "device")

	if channel == 0 {
		return tools.ErrorResult(fmt.Errorf("channel is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SendPitchBend(ctx, channel, value, device)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Sent Pitch Bend: ch%d value%d", channel, value)), nil
}

// handlePanic handles the aftrs_midi_panic tool
func handlePanic(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	device := tools.GetStringParam(req, "device")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SendPanic(ctx, device)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult("MIDI Panic sent (All Notes Off on all channels)"), nil
}

// handleTransport handles the aftrs_midi_transport tool
func handleTransport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}
	device := tools.GetStringParam(req, "device")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err = client.SendStart(ctx, device)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Sent MIDI Start"), nil

	case "stop":
		err = client.SendStop(ctx, device)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Sent MIDI Stop"), nil

	case "continue":
		err = client.SendContinue(ctx, device)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Sent MIDI Continue"), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use start, stop, or continue)", action)), nil
	}
}

// handleMappings handles the aftrs_midi_mappings tool
func handleMappings(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	mappings, err := client.GetMappings(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# MIDI Mappings\n\n")

	if len(mappings) == 0 {
		sb.WriteString("No MIDI mappings configured.\n\n")
		sb.WriteString("Use `aftrs_midi_learn` to capture MIDI input and create mappings.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** mappings:\n\n", len(mappings)))
	sb.WriteString("| Name | Type | Channel | CC/Note | Target |\n")
	sb.WriteString("|------|------|---------|---------|--------|\n")

	for _, m := range mappings {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %s |\n", m.Name, m.Type, m.Channel, m.Controller, m.Target))
	}

	return tools.TextResult(sb.String()), nil
}

// handleLearn handles the aftrs_midi_learn tool
func handleLearn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	timeout := tools.GetIntParam(req, "timeout", 10)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.StartLearn(ctx, timeout)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# MIDI Learn\n\n")

	if !result.Success {
		sb.WriteString("**Status:** No MIDI input received within timeout\n\n")
		sb.WriteString("Try again and move a control on your MIDI device.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** MIDI message captured!\n\n")
	sb.WriteString("## Captured Message\n\n")
	sb.WriteString(fmt.Sprintf("- **Type:** %s\n", result.Message.Type))
	sb.WriteString(fmt.Sprintf("- **Channel:** %d\n", result.Message.Channel))
	sb.WriteString(fmt.Sprintf("- **Data 1:** %d\n", result.Message.Data1))
	sb.WriteString(fmt.Sprintf("- **Data 2:** %d\n", result.Message.Data2))

	if result.Suggestion != nil {
		sb.WriteString("\n## Suggested Mapping\n\n")
		sb.WriteString(fmt.Sprintf("- **Name:** %s\n", result.Suggestion.Name))
		sb.WriteString(fmt.Sprintf("- **Type:** %s\n", result.Suggestion.Type))
	}

	return tools.TextResult(sb.String()), nil
}

// handleHealth handles the aftrs_midi_health tool
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
	sb.WriteString("# MIDI Health\n\n")

	// Status emoji
	statusEmoji := ""
	if health.Status == "degraded" {
		statusEmoji = ""
	} else if health.Status == "critical" {
		statusEmoji = ""
	}

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## Metrics\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Input Devices | %d |\n", health.InputCount))
	sb.WriteString(fmt.Sprintf("| Output Devices | %d |\n", health.OutputCount))

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

// handleMapCreate handles the aftrs_midi_map_create tool
func handleMapCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	channel := tools.GetIntParam(req, "channel", 0)
	msgType := tools.GetStringParam(req, "type")
	number := tools.GetIntParam(req, "number", -1)
	toolName := tools.GetStringParam(req, "tool_name")
	targetParam := tools.GetStringParam(req, "target_param")
	outputMin := tools.GetFloatParam(req, "output_min", 0)
	outputMax := tools.GetFloatParam(req, "output_max", 127)
	invert := tools.GetBoolParam(req, "invert", false)

	if name == "" || channel == 0 || msgType == "" || number < 0 || toolName == "" {
		return tools.ErrorResult(fmt.Errorf("name, channel, type, number, and tool_name are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Build value mapping if target_param is specified
	var valueMap *clients.ValueMapping
	if targetParam != "" {
		valueMap = &clients.ValueMapping{
			InputMin:    0,
			InputMax:    127,
			OutputMin:   outputMin,
			OutputMax:   outputMax,
			TargetParam: targetParam,
			Invert:      invert,
		}
	}

	// Get static parameters
	params := make(map[string]interface{})
	if args, ok := req.Params.Arguments.(map[string]interface{}); ok {
		if p, ok := args["parameters"].(map[string]interface{}); ok {
			params = p
		}
	}

	mapping := &clients.ToolMapping{
		Name:       name,
		Channel:    channel,
		Type:       msgType,
		Number:     number,
		ToolName:   toolName,
		Parameters: params,
		ValueMap:   valueMap,
	}

	if err := client.CreateToolMapping(ctx, mapping); err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# MIDI Mapping Created\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** %s\n", mapping.ID))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", mapping.Name))
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", mapping.Type))
	sb.WriteString(fmt.Sprintf("**Channel:** %d\n", mapping.Channel))
	sb.WriteString(fmt.Sprintf("**Number:** %d\n", mapping.Number))
	sb.WriteString(fmt.Sprintf("**Tool:** %s\n", mapping.ToolName))

	if valueMap != nil {
		sb.WriteString(fmt.Sprintf("\n**Value Mapping:**\n"))
		sb.WriteString(fmt.Sprintf("- Target Param: %s\n", valueMap.TargetParam))
		sb.WriteString(fmt.Sprintf("- Range: %d-%d → %.1f-%.1f\n", valueMap.InputMin, valueMap.InputMax, valueMap.OutputMin, valueMap.OutputMax))
		if valueMap.Invert {
			sb.WriteString("- Inverted: Yes\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleMapDelete handles the aftrs_midi_map_delete tool
func handleMapDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, errResult := tools.RequireStringParam(req, "id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.DeleteToolMapping(ctx, id); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Mapping deleted: %s", id)), nil
}

// handleMapList handles the aftrs_midi_map_list tool
func handleMapList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	mappings, err := client.GetToolMappings(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# MIDI Tool Mappings\n\n")

	if len(mappings) == 0 {
		sb.WriteString("No MIDI→tool mappings configured.\n\n")
		sb.WriteString("Use `aftrs_midi_map_create` to create a mapping.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** mappings:\n\n", len(mappings)))
	sb.WriteString("| ID | Name | Type | Ch | # | Tool | Enabled |\n")
	sb.WriteString("|----|------|------|----|----|------|--------|\n")

	for _, m := range mappings {
		enabled := "Yes"
		if !m.Enabled {
			enabled = "No"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %d | %s | %s |\n",
			m.ID[:8], m.Name, m.Type, m.Channel, m.Number, m.ToolName, enabled))
	}

	return tools.TextResult(sb.String()), nil
}

// handleProfileSave handles the aftrs_midi_profile_save tool
func handleProfileSave(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	description := tools.GetStringParam(req, "description")

	if name == "" {
		return tools.ErrorResult(fmt.Errorf("profile name is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.SaveProfile(ctx, name, description); err != nil {
		return tools.ErrorResult(err), nil
	}

	mappings, _ := client.GetToolMappings(ctx)
	return tools.TextResult(fmt.Sprintf("Profile '%s' saved with %d mappings", name, len(mappings))), nil
}

// handleProfileLoad handles the aftrs_midi_profile_load tool
func handleProfileLoad(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.LoadProfile(ctx, name); err != nil {
		return tools.ErrorResult(err), nil
	}

	mappings, _ := client.GetToolMappings(ctx)
	return tools.TextResult(fmt.Sprintf("Profile '%s' loaded with %d mappings", name, len(mappings))), nil
}

// handleProfiles handles the aftrs_midi_profiles tool
func handleProfiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# MIDI Profiles\n\n")

	if len(profiles) == 0 {
		sb.WriteString("No profiles saved.\n\n")
		sb.WriteString("Use `aftrs_midi_profile_save` to save current mappings to a profile.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** profiles:\n\n", len(profiles)))
	sb.WriteString("| Name | Mappings | Description | Updated |\n")
	sb.WriteString("|------|----------|-------------|--------|\n")

	for _, p := range profiles {
		desc := p.Description
		if len(desc) > 30 {
			desc = desc[:30] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s |\n",
			p.Name, len(p.Mappings), desc, p.UpdatedAt.Format("2006-01-02")))
	}

	return tools.TextResult(sb.String()), nil
}

// handleBatchCC sends multiple MIDI CC messages in one call
func handleBatchCC(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	msgsStr, errResult := tools.RequireStringParam(req, "messages")
	if errResult != nil {
		return errResult, nil
	}
	device := tools.GetStringParam(req, "device")

	var msgs []struct {
		Channel    int `json:"channel"`
		Controller int `json:"controller"`
		Value      int `json:"value"`
	}
	if err := json.Unmarshal([]byte(msgsStr), &msgs); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid JSON: %w", err)), nil
	}
	if len(msgs) == 0 {
		return tools.ErrorResult(fmt.Errorf("messages array is empty")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# MIDI Batch CC\n\n")

	sent := 0
	var errors []string
	for i, m := range msgs {
		if m.Channel < 1 || m.Channel > 16 {
			errors = append(errors, fmt.Sprintf("msg %d: channel %d out of range (1-16)", i, m.Channel))
			continue
		}
		if m.Controller < 0 || m.Controller > 127 {
			errors = append(errors, fmt.Sprintf("msg %d: controller %d out of range (0-127)", i, m.Controller))
			continue
		}
		if m.Value < 0 || m.Value > 127 {
			errors = append(errors, fmt.Sprintf("msg %d: value %d out of range (0-127)", i, m.Value))
			continue
		}

		if err := client.SendCC(ctx, m.Channel, m.Controller, m.Value, device); err != nil {
			errors = append(errors, fmt.Sprintf("msg %d: %v", i, err))
			continue
		}
		sent++
	}

	sb.WriteString(fmt.Sprintf("**Sent:** %d/%d messages\n", sent, len(msgs)))
	if len(errors) > 0 {
		sb.WriteString(fmt.Sprintf("\n**Errors:** %d\n", len(errors)))
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
