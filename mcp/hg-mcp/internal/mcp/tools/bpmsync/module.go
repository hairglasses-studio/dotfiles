// Package bpmsync provides MCP tools for unified BPM synchronization across systems.
package bpmsync

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the BPM Sync tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "bpmsync"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Unified BPM synchronization across Ableton, Resolume, grandMA3, and other systems"
}

// Tools returns the BPM Sync tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_sync_bpm_status",
				mcp.WithDescription("Get current BPM across all systems and sync status"),
			),
			Handler:             handleBPMStatus,
			Category:            "bpmsync",
			Subcategory:         "status",
			Tags:                []string{"bpm", "sync", "tempo", "status"},
			UseCases:            []string{"Check BPM across systems", "View sync status", "Detect tempo drift"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "bpmsync",
		},
		{
			Tool: mcp.NewTool("aftrs_sync_bpm_master",
				mcp.WithDescription("Set the master BPM source (ableton, resolume, or manual)"),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Master source: ableton, resolume, or manual"),
					mcp.Enum("ableton", "resolume", "manual"),
				),
			),
			Handler:             handleBPMMaster,
			Category:            "bpmsync",
			Subcategory:         "config",
			Tags:                []string{"bpm", "master", "source", "config"},
			UseCases:            []string{"Set Ableton as master", "Switch to manual tempo", "Change sync source"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "bpmsync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_sync_bpm_link",
				mcp.WithDescription("Link or unlink a system to receive BPM updates from master"),
				mcp.WithString("system",
					mcp.Required(),
					mcp.Description("System to link/unlink: ableton, resolume, grandma3"),
					mcp.Enum("ableton", "resolume", "grandma3"),
				),
				mcp.WithBoolean("linked",
					mcp.Description("True to link, false to unlink"),
				),
			),
			Handler:             handleBPMLink,
			Category:            "bpmsync",
			Subcategory:         "config",
			Tags:                []string{"bpm", "link", "sync", "system"},
			UseCases:            []string{"Link Resolume to master", "Unlink grandMA3", "Configure sync targets"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "bpmsync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_sync_bpm_push",
				mcp.WithDescription("Push a BPM value to all linked systems, or sync from master"),
				mcp.WithNumber("bpm",
					mcp.Description("BPM to push (20-999). Omit to sync from current master."),
				),
			),
			Handler:             handleBPMPush,
			Category:            "bpmsync",
			Subcategory:         "control",
			Tags:                []string{"bpm", "push", "sync", "tempo"},
			UseCases:            []string{"Set tempo to 128 BPM", "Sync all systems from master", "Force tempo sync"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "bpmsync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_sync_tap_tempo",
				mcp.WithDescription("Record a tap for tap tempo calculation across all linked systems"),
			),
			Handler:             handleTapTempo,
			Category:            "bpmsync",
			Subcategory:         "control",
			Tags:                []string{"bpm", "tap", "tempo"},
			UseCases:            []string{"Tap to set tempo", "Manual tempo detection", "Live tempo adjustment"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "bpmsync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_sync_bpm_health",
				mcp.WithDescription("Check BPM sync health and get troubleshooting recommendations"),
			),
			Handler:             handleBPMHealth,
			Category:            "bpmsync",
			Subcategory:         "status",
			Tags:                []string{"bpm", "health", "diagnostics"},
			UseCases:            []string{"Check sync health", "Diagnose sync issues", "Verify configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "bpmsync",
		},
	}
}

var getBPMSyncClient = tools.LazyClient(clients.NewBPMSyncClient)

func handleBPMStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getBPMSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create BPM sync client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleBPMMaster(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getBPMSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create BPM sync client: %w", err)), nil
	}

	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.SetMaster(ctx, source); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to set master: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"master":  source,
		"message": fmt.Sprintf("Master source set to '%s'", source),
	}), nil
}

func handleBPMLink(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getBPMSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create BPM sync client: %w", err)), nil
	}

	system, errResult := tools.RequireStringParam(req, "system")
	if errResult != nil {
		return errResult, nil
	}

	linked := tools.GetBoolParam(req, "linked", true)

	var linkErr error
	if linked {
		linkErr = client.LinkSystem(ctx, system)
	} else {
		linkErr = client.UnlinkSystem(ctx, system)
	}

	if linkErr != nil {
		return tools.ErrorResult(fmt.Errorf("failed to update link: %w", linkErr)), nil
	}

	action := "linked"
	if !linked {
		action = "unlinked"
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"system":  system,
		"linked":  linked,
		"message": fmt.Sprintf("System '%s' %s", system, action),
	}), nil
}

func handleBPMPush(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getBPMSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create BPM sync client: %w", err)), nil
	}

	bpm := tools.GetFloatParam(req, "bpm", 0)

	var pushErr error
	if bpm > 0 {
		pushErr = client.PushBPM(ctx, bpm)
	} else {
		pushErr = client.SyncFromMaster(ctx)
	}

	if pushErr != nil {
		return tools.ErrorResult(fmt.Errorf("failed to push BPM: %w", pushErr)), nil
	}

	status, _ := client.GetStatus(ctx)

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"bpm":     status.CurrentBPM,
		"master":  status.MasterSource,
		"in_sync": status.InSync,
		"message": fmt.Sprintf("BPM %.2f pushed to all linked systems", status.CurrentBPM),
	}), nil
}

func handleTapTempo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getBPMSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create BPM sync client: %w", err)), nil
	}

	bpm, err := client.TapTempo(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to tap tempo: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"bpm":     bpm,
		"message": fmt.Sprintf("Tap recorded, BPM: %.2f", bpm),
	}), nil
}

func handleBPMHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getBPMSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create BPM sync client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
