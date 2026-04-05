// Package supercollider provides SuperCollider audio synthesis tools for hg-mcp.
package supercollider

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for SuperCollider integration.
type Module struct{}

func (m *Module) Name() string {
	return "supercollider"
}

func (m *Module) Description() string {
	return "SuperCollider audio synthesis engine control via OSC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_supercollider_status",
				mcp.WithDescription("Get SuperCollider server status (CPU, synths, sample rate)."),
			),
			Handler:             handleStatus,
			Category:            "audio",
			Subcategory:         "supercollider",
			Tags:                []string{"supercollider", "status", "audio", "synthesis"},
			UseCases:            []string{"Check SuperCollider connectivity", "Monitor server load"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "supercollider",
		},
		{
			Tool: mcp.NewTool("aftrs_supercollider_health",
				mcp.WithDescription("Check SuperCollider health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "audio",
			Subcategory:         "supercollider",
			Tags:                []string{"supercollider", "health", "diagnostics"},
			UseCases:            []string{"Diagnose SuperCollider issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "supercollider",
		},
		{
			Tool: mcp.NewTool("aftrs_supercollider_synth",
				mcp.WithDescription("Create or free synths by name and node ID."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: create or free")),
				mcp.WithString("def_name", mcp.Description("SynthDef name (required for create)")),
				mcp.WithNumber("node_id", mcp.Required(), mcp.Description("Node ID for the synth")),
			),
			Handler:             handleSynth,
			Category:            "audio",
			Subcategory:         "supercollider",
			Tags:                []string{"supercollider", "synth", "create", "free"},
			UseCases:            []string{"Create synth instances", "Stop synths"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "supercollider",
		},
		{
			Tool: mcp.NewTool("aftrs_supercollider_node",
				mcp.WithDescription("Set parameters on a running synth node."),
				mcp.WithNumber("node_id", mcp.Required(), mcp.Description("Target node ID")),
				mcp.WithString("param", mcp.Required(), mcp.Description("Parameter name")),
				mcp.WithNumber("value", mcp.Required(), mcp.Description("Parameter value")),
			),
			Handler:             handleNode,
			Category:            "audio",
			Subcategory:         "supercollider",
			Tags:                []string{"supercollider", "node", "parameter", "control"},
			UseCases:            []string{"Adjust synth parameters in real-time"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "supercollider",
		},
		{
			Tool: mcp.NewTool("aftrs_supercollider_eval",
				mcp.WithDescription("Evaluate SuperCollider code via sclang."),
				mcp.WithString("code", mcp.Required(), mcp.Description("SuperCollider code to evaluate")),
			),
			Handler:             handleEval,
			Category:            "audio",
			Subcategory:         "supercollider",
			Tags:                []string{"supercollider", "eval", "code", "sclang"},
			UseCases:            []string{"Run SuperCollider code", "Execute synthesis scripts"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "supercollider",
		},
	}
}

var getClient = tools.LazyClient(clients.GetSuperColliderClient)

// handleStatus returns SuperCollider server status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SuperCollider client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# SuperCollider Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**scsynth:** %s:%d\n", status.Host, status.ScSynthPort))
		sb.WriteString(fmt.Sprintf("**sclang:** %s:%d\n\n", status.Host, status.ScLangPort))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Start SuperCollider (scsynth or SuperCollider IDE)\n")
		sb.WriteString("2. Ensure scsynth is listening on the configured port\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export SUPERCOLLIDER_HOST=localhost\n")
		sb.WriteString("export SUPERCOLLIDER_SCSYNTH_PORT=57110\n")
		sb.WriteString("export SUPERCOLLIDER_SCLANG_PORT=57120\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| UGens | %d |\n", status.UGens))
	sb.WriteString(fmt.Sprintf("| Synths | %d |\n", status.Synths))
	sb.WriteString(fmt.Sprintf("| Groups | %d |\n", status.Groups))
	sb.WriteString(fmt.Sprintf("| SynthDefs | %d |\n", status.SynthDefs))
	sb.WriteString(fmt.Sprintf("| Avg CPU | %.1f%% |\n", status.AvgCPU))
	sb.WriteString(fmt.Sprintf("| Peak CPU | %.1f%% |\n", status.PeakCPU))
	sb.WriteString(fmt.Sprintf("| Sample Rate | %.0f Hz |\n", status.SampleRate))
	sb.WriteString(fmt.Sprintf("| Actual Rate | %.0f Hz |\n", status.ActualSampleRate))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns SuperCollider health and recommendations.
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
		issues = append(issues, "Not connected to scsynth")
		recommendations = append(recommendations,
			"Start SuperCollider or scsynth",
			fmt.Sprintf("Verify scsynth is listening on %s:%d", status.Host, status.ScSynthPort),
			"Check SUPERCOLLIDER_HOST and SUPERCOLLIDER_SCSYNTH_PORT env vars",
		)
	}

	if status.AvgCPU > 80 {
		score -= 20
		issues = append(issues, fmt.Sprintf("High CPU usage: %.1f%%", status.AvgCPU))
		recommendations = append(recommendations, "Reduce active synths or optimize SynthDefs")
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

// handleSynth creates or frees synth nodes.
func handleSynth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	nodeID, errResult := tools.RequireIntParam(req, "node_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "create":
		defName, errResult := tools.RequireStringParam(req, "def_name")
		if errResult != nil {
			return errResult, nil
		}
		err = client.CreateSynth(ctx, defName, nodeID, nil)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Created synth %s (node %d)", defName, nodeID)), nil

	case "free":
		err = client.FreeSynth(ctx, nodeID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Freed node %d", nodeID)), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use create or free)", action)), nil
	}
}

// handleNode sets parameters on a running synth node.
func handleNode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	nodeID, errResult := tools.RequireIntParam(req, "node_id")
	if errResult != nil {
		return errResult, nil
	}

	param, errResult := tools.RequireStringParam(req, "param")
	if errResult != nil {
		return errResult, nil
	}

	value := tools.GetFloatParam(req, "value", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetNodeParam(ctx, nodeID, param, value)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Set node %d param %s = %g", nodeID, param, value)), nil
}

// handleEval evaluates SuperCollider code via sclang.
func handleEval(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code, errResult := tools.RequireStringParam(req, "code")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.EvalCode(ctx, code)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("eval failed: %w", err)), nil
	}

	if result == "" {
		return tools.TextResult("Code sent to sclang (no response)"), nil
	}
	return tools.TextResult(result), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
