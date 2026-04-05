// Package chataigne provides Chataigne show control tools for hg-mcp.
// Chataigne bridges OSC, MIDI, DMX, ArtNet, sACN, HTTP, WebSocket,
// MQTT, PJLink, and Ableton Link into a unified show control environment.
package chataigne

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Chataigne.
type Module struct{}

var getClient = tools.LazyClient(clients.NewChataigneClient)

func (m *Module) Name() string        { return "chataigne" }
func (m *Module) Description() string { return "Chataigne show control bridge (OSC/MIDI/DMX/HTTP)" }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool:                mcp.NewTool("aftrs_chataigne_status", mcp.WithDescription("Get Chataigne connection status including HTTP and OSC endpoints.")),
			Handler:             handleStatus,
			Category:            "showcontrol",
			Subcategory:         "chataigne",
			Tags:                []string{"chataigne", "showcontrol", "status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "chataigne",
		},
		{
			Tool:                mcp.NewTool("aftrs_chataigne_health", mcp.WithDescription("Check Chataigne system health and get troubleshooting recommendations.")),
			Handler:             handleHealth,
			Category:            "showcontrol",
			Subcategory:         "chataigne",
			Tags:                []string{"chataigne", "health", "diagnostics"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "chataigne",
		},
		{
			Tool:                mcp.NewTool("aftrs_chataigne_modules", mcp.WithDescription("List protocol modules in Chataigne (OSC, MIDI, DMX, HTTP bridges).")),
			Handler:             handleModules,
			Category:            "showcontrol",
			Subcategory:         "chataigne",
			Tags:                []string{"chataigne", "modules", "protocols"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "chataigne",
		},
		{
			Tool: mcp.NewTool("aftrs_chataigne_state",
				mcp.WithDescription("Get or set the state machine state. Use action 'get' to read current state, or 'set' with a state name to transition."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: get or set")),
				mcp.WithString("state", mcp.Description("State name to transition to (required for 'set')")),
			),
			Handler:             handleState,
			Category:            "showcontrol",
			Subcategory:         "chataigne",
			Tags:                []string{"chataigne", "state", "statemachine"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "chataigne",
		},
		{
			Tool: mcp.NewTool("aftrs_chataigne_sequence",
				mcp.WithDescription("Control sequence/timeline playback in Chataigne."),
				mcp.WithString("name", mcp.Description("Sequence name (omit to list all)")),
				mcp.WithString("action", mcp.Description("Action: play or stop (required when name is provided)")),
			),
			Handler:             handleSequence,
			Category:            "showcontrol",
			Subcategory:         "chataigne",
			Tags:                []string{"chataigne", "sequence", "timeline", "playback"},
			IsWrite:             true,
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "chataigne",
		},
	}
}

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
	sb.WriteString("# Chataigne Status\n\n")
	if status.Connected {
		sb.WriteString("**Status:** Connected\n")
	} else {
		sb.WriteString("**Status:** Disconnected\n")
	}
	sb.WriteString(fmt.Sprintf("**Host:** %s\n", status.Host))
	sb.WriteString(fmt.Sprintf("**HTTP Port:** %d\n", status.Port))
	sb.WriteString(fmt.Sprintf("**OSC Port:** %d\n", status.OSCPort))
	sb.WriteString(fmt.Sprintf("**API URL:** `%s`\n", status.URL))

	return tools.TextResult(sb.String()), nil
}

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
	sb.WriteString("# Chataigne Health\n\n")
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

func handleModules(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	modules, err := client.GetModules(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(modules) == 0 {
		return tools.TextResult("# Chataigne Modules\n\nNo modules found."), nil
	}

	var sb strings.Builder
	sb.WriteString("# Chataigne Modules\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d modules\n\n", len(modules)))
	sb.WriteString("| Name | Type | Status | Enabled |\n")
	sb.WriteString("|---|---|---|---|\n")
	for _, mod := range modules {
		enabled := "No"
		if mod.Enabled {
			enabled = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", mod.Name, mod.Type, mod.Status, enabled))
	}

	return tools.TextResult(sb.String()), nil
}

func handleState(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	actionRaw, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}
	action := strings.ToLower(actionRaw)

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	switch action {
	case "get":
		sm, err := client.GetStateMachine(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(map[string]interface{}{
			"current_state": sm.CurrentState,
			"states":        sm.States,
		}), nil

	case "set":
		state, errResult := tools.RequireStringParam(req, "state")
		if errResult != nil {
			return errResult, nil
		}
		if err := client.TriggerState(ctx, state); err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(map[string]interface{}{
			"state":   state,
			"action":  "set",
			"success": true,
		}), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("action must be 'get' or 'set', got '%s'", action)), nil
	}
}

func handleSequence(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	action := strings.ToLower(tools.GetStringParam(req, "action"))

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	// If no name, list sequences
	if name == "" {
		sequences, err := client.GetSequences(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		if len(sequences) == 0 {
			return tools.TextResult("# Chataigne Sequences\n\nNo sequences found."), nil
		}

		var sb strings.Builder
		sb.WriteString("# Chataigne Sequences\n\n")
		sb.WriteString("| Name | Playing | Position | Duration |\n")
		sb.WriteString("|---|---|---|---|\n")
		for _, seq := range sequences {
			playing := "Stopped"
			if seq.Playing {
				playing = "Playing"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %.1fs | %.1fs |\n",
				seq.Name, playing, seq.Position, seq.Duration))
		}
		return tools.TextResult(sb.String()), nil
	}

	// With name, control sequence
	if action == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("action is required when name is provided (play or stop)")), nil
	}

	var play bool
	switch action {
	case "play":
		play = true
	case "stop":
		play = false
	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("action must be 'play' or 'stop', got '%s'", action)), nil
	}

	if err := client.TriggerSequence(ctx, name, play); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"sequence": name,
		"action":   action,
		"success":  true,
	}), nil
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
