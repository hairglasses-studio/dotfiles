// Package timecodesync provides MCP timecode synchronization tools.
package timecodesync

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for timecode sync
type Module struct{}

var getClient = tools.LazyClient(clients.NewTimecodeSyncClient)

func (m *Module) Name() string {
	return "timecodesync"
}

func (m *Module) Description() string {
	return "Timecode synchronization across systems (Showkontrol, grandMA3, Ableton)"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_timecode_status",
				mcp.WithDescription("Get current timecode status across all systems. Shows master source, current position, linked systems, and sync status."),
			),
			Handler:             handleTimecodeStatus,
			Category:            "sync",
			Subcategory:         "timecode",
			Tags:                []string{"timecode", "sync", "status", "smpte", "mtc"},
			UseCases:            []string{"Check timecode position", "Verify sync status", "Monitor linked systems"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "timecodesync",
		},
		{
			Tool: mcp.NewTool("aftrs_timecode_sync",
				mcp.WithDescription("Configure timecode synchronization. Set master source and link/unlink systems."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: set_master, link, unlink, sync_now")),
				mcp.WithString("system", mcp.Description("System name: showkontrol, grandma3, ableton")),
				mcp.WithString("format", mcp.Description("Timecode format: smpte, mtc, ltc (for set_master)")),
				mcp.WithNumber("frame_rate", mcp.Description("Frame rate: 24, 25, 29.97, 30, etc.")),
				mcp.WithBoolean("drop_frame", mcp.Description("Use drop frame timecode (for 29.97/59.94)")),
			),
			Handler:             handleTimecodeSync,
			Category:            "sync",
			Subcategory:         "timecode",
			Tags:                []string{"timecode", "sync", "link", "master"},
			UseCases:            []string{"Set master timecode source", "Link systems for sync", "Configure frame rate"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "timecodesync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_timecode_goto",
				mcp.WithDescription("Jump all linked systems to a specific timecode position."),
				mcp.WithString("position", mcp.Required(), mcp.Description("Timecode position in HH:MM:SS:FF format (e.g., 01:30:45:15)")),
			),
			Handler:             handleTimecodeGoto,
			Category:            "sync",
			Subcategory:         "timecode",
			Tags:                []string{"timecode", "goto", "position", "jump"},
			UseCases:            []string{"Jump to show position", "Sync all systems to mark", "Rehearse from point"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "timecodesync",
			IsWrite:             true,
		},
	}
}

func handleTimecodeStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Timecode Sync Status\n\n")

	// Master source and current TC
	sb.WriteString(fmt.Sprintf("**Master Source:** %s\n", status.MasterSource))
	sb.WriteString(fmt.Sprintf("**Current Timecode:** `%s`\n", status.CurrentTC.String()))
	sb.WriteString(fmt.Sprintf("**Format:** %s @ %.2f fps", status.Format, status.CurrentTC.FrameRate))
	if status.CurrentTC.DropFrame {
		sb.WriteString(" (drop frame)")
	}
	sb.WriteString("\n")

	// Sync status
	if status.InSync {
		sb.WriteString("**Sync Status:** ✅ In Sync\n")
	} else {
		sb.WriteString(fmt.Sprintf("**Sync Status:** ⚠️ Out of Sync (drift: %d frames)\n", status.DriftFrames))
	}
	sb.WriteString("\n")

	// Systems table
	sb.WriteString("## Systems\n\n")
	sb.WriteString("| System | Type | Connected | Linked | Current TC |\n")
	sb.WriteString("|--------|------|-----------|--------|------------|\n")

	for _, sys := range status.Systems {
		connected := "❌"
		if sys.Connected {
			connected = "✅"
		}
		linked := "No"
		if sys.Linked {
			linked = "Yes"
		}
		tc := "-"
		if sys.CurrentTC != nil {
			tc = sys.CurrentTC.String()
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			sys.Name, sys.Type, connected, linked, tc))
	}

	return tools.TextResult(sb.String()), nil
}

func handleTimecodeSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.GetStringParam(req, "action")
	system := tools.GetStringParam(req, "system")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var result string

	switch action {
	case "set_master":
		if system == "" {
			return tools.ErrorResult(fmt.Errorf("system is required for set_master action")), nil
		}
		if err := client.SetMaster(ctx, system); err != nil {
			return tools.ErrorResult(err), nil
		}

		// Also set format if provided
		formatStr := tools.GetStringParam(req, "format")
		frameRate := tools.GetFloatParam(req, "frame_rate", 0)
		dropFrame := tools.GetBoolParam(req, "drop_frame", false)

		if formatStr != "" || frameRate > 0 {
			if frameRate == 0 {
				frameRate = 30
			}
			format := clients.TimecodeFormatSMPTE
			switch formatStr {
			case "mtc":
				format = clients.TimecodeFormatMTC
			case "ltc":
				format = clients.TimecodeFormatLTC
			case "ms":
				format = clients.TimecodeFormatMS
			}
			if err := client.SetFormat(ctx, format, frameRate, dropFrame); err != nil {
				return tools.ErrorResult(err), nil
			}
		}

		result = fmt.Sprintf("✅ Set **%s** as master timecode source.", system)

	case "link":
		if system == "" {
			return tools.ErrorResult(fmt.Errorf("system is required for link action")), nil
		}
		if err := client.LinkSystem(ctx, system); err != nil {
			return tools.ErrorResult(err), nil
		}
		result = fmt.Sprintf("✅ Linked **%s** to receive timecode from master.", system)

	case "unlink":
		if system == "" {
			return tools.ErrorResult(fmt.Errorf("system is required for unlink action")), nil
		}
		if err := client.UnlinkSystem(ctx, system); err != nil {
			return tools.ErrorResult(err), nil
		}
		result = fmt.Sprintf("✅ Unlinked **%s** from timecode sync.", system)

	case "sync_now":
		if err := client.SyncFromMaster(ctx); err != nil {
			return tools.ErrorResult(err), nil
		}
		result = "✅ Synced all linked systems from master timecode."

	default:
		return tools.ErrorResult(fmt.Errorf("unknown action: %s (valid: set_master, link, unlink, sync_now)", action)), nil
	}

	return tools.TextResult(result), nil
}

func handleTimecodeGoto(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	position, errResult := tools.RequireStringParam(req, "position")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.GotoPosition(ctx, position); err != nil {
		return tools.ErrorResult(err), nil
	}

	// Get updated status to show result
	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("✅ Jumped to timecode position **%s**\n\n(Unable to verify status: %v)", position, err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Timecode Jump to %s\n\n", position))
	sb.WriteString("✅ Successfully jumped all linked systems.\n\n")

	sb.WriteString("## System Status\n\n")
	sb.WriteString("| System | Linked | Status |\n")
	sb.WriteString("|--------|--------|--------|\n")

	for _, sys := range status.Systems {
		if sys.Linked {
			statusEmoji := "✅"
			if !sys.Connected {
				statusEmoji = "⚠️ Not connected"
			}
			sb.WriteString(fmt.Sprintf("| %s | Yes | %s |\n", sys.Name, statusEmoji))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
