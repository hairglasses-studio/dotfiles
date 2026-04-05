// Package ardour provides Ardour DAW control tools for hg-mcp.
package ardour

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Ardour integration.
type Module struct{}

func (m *Module) Name() string {
	return "ardour"
}

func (m *Module) Description() string {
	return "Ardour DAW control via OSC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ardour_status",
				mcp.WithDescription("Get Ardour connection status and transport state."),
			),
			Handler:             handleStatus,
			Category:            "audio",
			Subcategory:         "ardour",
			Tags:                []string{"ardour", "status", "daw", "transport"},
			UseCases:            []string{"Check Ardour connectivity", "View transport state"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ardour",
		},
		{
			Tool: mcp.NewTool("aftrs_ardour_health",
				mcp.WithDescription("Check Ardour health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "audio",
			Subcategory:         "ardour",
			Tags:                []string{"ardour", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Ardour issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ardour",
		},
		{
			Tool: mcp.NewTool("aftrs_ardour_transport",
				mcp.WithDescription("Control Ardour transport (play, stop, record, locate)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: play, stop, record, or locate")),
				mcp.WithNumber("frame", mcp.Description("Frame position for locate action")),
			),
			Handler:             handleTransport,
			Category:            "audio",
			Subcategory:         "ardour",
			Tags:                []string{"ardour", "transport", "play", "stop", "record"},
			UseCases:            []string{"Control playback", "Start recording", "Seek to position"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "ardour",
		},
		{
			Tool: mcp.NewTool("aftrs_ardour_track",
				mcp.WithDescription("Get or set track fader, mute, solo, and pan by strip ID."),
				mcp.WithNumber("strip_id", mcp.Required(), mcp.Description("Mixer strip ID (0-based)")),
				mcp.WithNumber("fader", mcp.Description("Fader level (0.0-1.0)")),
				mcp.WithBoolean("mute", mcp.Description("Set mute state")),
				mcp.WithBoolean("solo", mcp.Description("Set solo state")),
				mcp.WithNumber("pan", mcp.Description("Pan position (0.0=left, 0.5=center, 1.0=right)")),
			),
			Handler:             handleTrack,
			Category:            "audio",
			Subcategory:         "ardour",
			Tags:                []string{"ardour", "track", "mixer", "fader", "pan"},
			UseCases:            []string{"Adjust track levels", "Mute/solo tracks", "Set pan position"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "ardour",
		},
		{
			Tool: mcp.NewTool("aftrs_ardour_meter",
				mcp.WithDescription("Read track meter levels."),
				mcp.WithNumber("strip_id", mcp.Required(), mcp.Description("Mixer strip ID (0-based)")),
			),
			Handler:             handleMeter,
			Category:            "audio",
			Subcategory:         "ardour",
			Tags:                []string{"ardour", "meter", "level", "monitoring"},
			UseCases:            []string{"Monitor track levels", "Check signal presence"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ardour",
		},
	}
}

var getClient = tools.LazyClient(clients.GetArdourClient)

// handleStatus returns Ardour connection and transport status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ardour client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Ardour Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Target:** %s:%d\n\n", status.Host, status.Port))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open Ardour\n")
		sb.WriteString("2. Go to Edit → Preferences → Control Surfaces\n")
		sb.WriteString("3. Enable OSC and set the port\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export ARDOUR_HOST=localhost\n")
		sb.WriteString("export ARDOUR_PORT=3819\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n\n")

	transport := "Stopped"
	if status.Playing {
		transport = "Playing"
	}
	if status.Recording {
		transport = "Recording"
	}
	sb.WriteString(fmt.Sprintf("**Transport:** %s\n", transport))
	sb.WriteString(fmt.Sprintf("**Frame:** %d\n", status.Frame))
	sb.WriteString(fmt.Sprintf("**Speed:** %.1f\n", status.Speed))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns Ardour health and recommendations.
func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("health check failed: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("health check failed: %w", err)), nil
	}

	score := 100
	var issues []string
	var recommendations []string

	if !status.Connected {
		score -= 50
		issues = append(issues, "Not connected to Ardour")
		recommendations = append(recommendations,
			"Start Ardour and enable OSC control surface",
			fmt.Sprintf("Verify Ardour is listening on %s:%d", status.Host, status.Port),
			"Check ARDOUR_HOST and ARDOUR_PORT env vars",
		)
	}

	healthStatus := "healthy"
	if score < 80 {
		healthStatus = "degraded"
	}
	if score < 50 {
		healthStatus = "critical"
	}

	health := map[string]interface{}{
		"score":           score,
		"status":          healthStatus,
		"connected":       status.Connected,
		"issues":          issues,
		"recommendations": recommendations,
	}
	return tools.JSONResult(health), nil
}

// handleTransport controls Ardour transport.
func handleTransport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "play":
		err = client.TransportPlay(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Transport: play"), nil

	case "stop":
		err = client.TransportStop(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Transport: stop"), nil

	case "record":
		err = client.TransportRecord(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Transport: record enabled"), nil

	case "locate":
		frame := int64(tools.GetIntParam(req, "frame", 0))
		err = client.TransportLocate(ctx, frame)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Transport: located to frame %d", frame)), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use play, stop, record, or locate)", action)), nil
	}
}

// handleTrack gets or sets track mixer parameters.
func handleTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stripID := tools.GetIntParam(req, "strip_id", -1)
	if stripID < 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("strip_id is required and must be non-negative")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Check if any set params were provided
	hasSet := false
	var actions []string

	if v := tools.GetFloatParam(req, "fader", -1); v >= 0 {
		err = client.SetStripFader(ctx, stripID, v)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		actions = append(actions, fmt.Sprintf("fader=%.2f", v))
		hasSet = true
	}

	if req.GetArguments()["mute"] != nil {
		mute := tools.GetBoolParam(req, "mute", false)
		err = client.SetStripMute(ctx, stripID, mute)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		actions = append(actions, fmt.Sprintf("mute=%v", mute))
		hasSet = true
	}

	if req.GetArguments()["solo"] != nil {
		solo := tools.GetBoolParam(req, "solo", false)
		err = client.SetStripSolo(ctx, stripID, solo)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		actions = append(actions, fmt.Sprintf("solo=%v", solo))
		hasSet = true
	}

	if v := tools.GetFloatParam(req, "pan", -1); v >= 0 {
		err = client.SetStripPan(ctx, stripID, v)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		actions = append(actions, fmt.Sprintf("pan=%.2f", v))
		hasSet = true
	}

	if hasSet {
		return tools.TextResult(fmt.Sprintf("Strip %d: set %s", stripID, strings.Join(actions, ", "))), nil
	}

	// No set params — return current info (stub)
	return tools.TextResult(fmt.Sprintf("Strip %d: no changes (provide fader, mute, solo, or pan to modify)", stripID)), nil
}

// handleMeter reads track meter levels.
func handleMeter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stripID := tools.GetIntParam(req, "strip_id", -1)
	if stripID < 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("strip_id is required and must be non-negative")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	meter, err := client.GetStripMeter(ctx, stripID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"strip_id": stripID,
		"meter_db": meter,
	}), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
