// Package linuxshowplayer provides Linux Show Player cue control tools for hg-mcp.
package linuxshowplayer

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Linux Show Player integration.
type Module struct{}

func (m *Module) Name() string {
	return "linuxshowplayer"
}

func (m *Module) Description() string {
	return "Linux Show Player cue-based show control via OSC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_lsp_status",
				mcp.WithDescription("Get Linux Show Player connection status."),
			),
			Handler:             handleStatus,
			Category:            "automation",
			Subcategory:         "linuxshowplayer",
			Tags:                []string{"lsp", "status", "show-control", "cue"},
			UseCases:            []string{"Check Linux Show Player connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "linuxshowplayer",
		},
		{
			Tool: mcp.NewTool("aftrs_lsp_health",
				mcp.WithDescription("Check Linux Show Player health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "automation",
			Subcategory:         "linuxshowplayer",
			Tags:                []string{"lsp", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Linux Show Player issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "linuxshowplayer",
		},
		{
			Tool: mcp.NewTool("aftrs_lsp_cue",
				mcp.WithDescription("Trigger, stop, or pause cues in Linux Show Player."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: go (next cue), stop, pause, start (specific cue), stop_cue")),
				mcp.WithNumber("cue", mcp.Description("Cue number (required for start and stop_cue actions)")),
			),
			Handler:             handleCue,
			Category:            "automation",
			Subcategory:         "linuxshowplayer",
			Tags:                []string{"lsp", "cue", "trigger", "playback"},
			UseCases:            []string{"Fire cues", "Stop playback", "Pause show"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "linuxshowplayer",
		},
	}
}

var getClient = tools.LazyClient(clients.GetLinuxShowPlayerClient)

// handleStatus returns Linux Show Player connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create LSP client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Linux Show Player Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**OSC:** %s:%d\n\n", status.Host, status.Port))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open Linux Show Player\n")
		sb.WriteString("2. Go to Application Settings → Plugins → OSC\n")
		sb.WriteString("3. Enable OSC and configure the port\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export LSP_HOST=localhost\n")
		sb.WriteString("export LSP_PORT=9000\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**OSC:** %s:%d\n", status.Host, status.Port))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns Linux Show Player health and recommendations.
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
		issues = append(issues, "Not connected to Linux Show Player")
		recommendations = append(recommendations,
			"Start Linux Show Player and enable OSC plugin",
			fmt.Sprintf("Verify LSP is listening on %s:%d", status.Host, status.Port),
			"Check LSP_HOST and LSP_PORT env vars",
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

// handleCue controls cue playback.
func handleCue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "go":
		err = client.Go(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Cue: go (next cue triggered)"), nil

	case "stop":
		err = client.Stop(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Cue: all stopped"), nil

	case "pause":
		err = client.Pause(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Cue: all paused"), nil

	case "start":
		cueNum := tools.GetIntParam(req, "cue", -1)
		if cueNum < 0 {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("cue number is required for start action")), nil
		}
		err = client.StartCue(ctx, cueNum)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Cue %d: started", cueNum)), nil

	case "stop_cue":
		cueNum := tools.GetIntParam(req, "cue", -1)
		if cueNum < 0 {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("cue number is required for stop_cue action")), nil
		}
		err = client.StopCue(ctx, cueNum)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Cue %d: stopped", cueNum)), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use go, stop, pause, start, or stop_cue)", action)), nil
	}
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
