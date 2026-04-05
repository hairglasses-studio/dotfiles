// Package vimix provides vimix video mixing tools for hg-mcp.
package vimix

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for vimix integration.
type Module struct{}

func (m *Module) Name() string {
	return "vimix"
}

func (m *Module) Description() string {
	return "vimix video mixing for live performance via OSC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_vimix_status",
				mcp.WithDescription("Get vimix connection status."),
			),
			Handler:             handleStatus,
			Category:            "video",
			Subcategory:         "vimix",
			Tags:                []string{"vimix", "status", "video", "mixing"},
			UseCases:            []string{"Check vimix connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vimix",
		},
		{
			Tool: mcp.NewTool("aftrs_vimix_health",
				mcp.WithDescription("Check vimix health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "video",
			Subcategory:         "vimix",
			Tags:                []string{"vimix", "health", "diagnostics"},
			UseCases:            []string{"Diagnose vimix issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vimix",
		},
		{
			Tool: mcp.NewTool("aftrs_vimix_source",
				mcp.WithDescription("Control a video source (set opacity, activate/deactivate)."),
				mcp.WithNumber("source_id", mcp.Required(), mcp.Description("Source index (0-based)")),
				mcp.WithNumber("alpha", mcp.Description("Source opacity (0.0-1.0)")),
				mcp.WithBoolean("active", mcp.Description("Activate or deactivate the source")),
			),
			Handler:             handleSource,
			Category:            "video",
			Subcategory:         "vimix",
			Tags:                []string{"vimix", "source", "opacity", "layer"},
			UseCases:            []string{"Adjust source opacity", "Toggle source visibility"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "vimix",
		},
		{
			Tool: mcp.NewTool("aftrs_vimix_session",
				mcp.WithDescription("Save or load a vimix session."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: save or load")),
				mcp.WithString("filename", mcp.Description("Session filename (required for load)")),
			),
			Handler:             handleSession,
			Category:            "video",
			Subcategory:         "vimix",
			Tags:                []string{"vimix", "session", "save", "load"},
			UseCases:            []string{"Save current session", "Load a session file"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "vimix",
		},
	}
}

var getClient = tools.LazyClient(clients.GetVimixClient)

// handleStatus returns vimix connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create vimix client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# vimix Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**OSC:** %s:%d\n\n", status.Host, status.Port))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open vimix\n")
		sb.WriteString("2. Enable OSC in Preferences (default port 7000)\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export VIMIX_HOST=localhost\n")
		sb.WriteString("export VIMIX_PORT=7000\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**OSC:** %s:%d\n", status.Host, status.Port))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns vimix health and recommendations.
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
		issues = append(issues, "Not connected to vimix")
		recommendations = append(recommendations,
			"Start vimix and enable OSC",
			fmt.Sprintf("Verify vimix is listening on %s:%d", status.Host, status.Port),
			"Check VIMIX_HOST and VIMIX_PORT env vars",
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

// handleSource controls a video source.
func handleSource(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourceID := tools.GetIntParam(req, "source_id", -1)
	if sourceID < 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("source_id is required and must be non-negative")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var actions []string

	if v := tools.GetFloatParam(req, "alpha", -1); v >= 0 {
		err = client.SetSourceAlpha(ctx, sourceID, v)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		actions = append(actions, fmt.Sprintf("alpha=%.2f", v))
	}

	if req.GetArguments()["active"] != nil {
		active := tools.GetBoolParam(req, "active", false)
		err = client.SetSourceActive(ctx, sourceID, active)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		actions = append(actions, fmt.Sprintf("active=%v", active))
	}

	if len(actions) == 0 {
		return tools.TextResult(fmt.Sprintf("Source %d: no changes (provide alpha or active)", sourceID)), nil
	}

	return tools.TextResult(fmt.Sprintf("Source %d: set %s", sourceID, strings.Join(actions, ", "))), nil
}

// handleSession saves or loads a vimix session.
func handleSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "save":
		err = client.SaveSession(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Session saved"), nil

	case "load":
		filename, errResult := tools.RequireStringParam(req, "filename")
		if errResult != nil {
			return errResult, nil
		}
		err = client.LoadSession(ctx, filename)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Session loaded: %s", filename)), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use save or load)", action)), nil
	}
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
