// Package xlights provides xLights FSEQ file parsing tools for hg-mcp.
package xlights

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for xLights FSEQ integration.
type Module struct{}

func (m *Module) Name() string {
	return "xlights"
}

func (m *Module) Description() string {
	return "xLights FSEQ lighting sequence file parser"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_xlights_info",
				mcp.WithDescription("Parse FSEQ file header and return metadata (channels, frames, FPS, duration, compression)."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the FSEQ file")),
			),
			Handler:             handleInfo,
			Category:            "lighting",
			Subcategory:         "xlights",
			Tags:                []string{"xlights", "fseq", "sequence", "metadata"},
			UseCases:            []string{"Inspect FSEQ file properties", "Check sequence duration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "xlights",
		},
		{
			Tool: mcp.NewTool("aftrs_xlights_frames",
				mcp.WithDescription("Extract frame data from an uncompressed FSEQ file for playback preview."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the FSEQ file")),
				mcp.WithNumber("start_frame", mcp.Description("Starting frame index (default: 0)")),
				mcp.WithNumber("count", mcp.Description("Number of frames to extract (default: 1)")),
			),
			Handler:             handleFrames,
			Category:            "lighting",
			Subcategory:         "xlights",
			Tags:                []string{"xlights", "fseq", "frames", "dmx", "playback"},
			UseCases:            []string{"Preview frame data", "Extract DMX values for playback"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "xlights",
		},
	}
}

var getClient = tools.LazyClient(clients.GetXLightsClient)

// handleInfo parses FSEQ header and returns metadata.
func handleInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	header, err := client.ParseHeader(path)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse FSEQ: %w", err)), nil
	}

	return tools.JSONResult(header), nil
}

// handleFrames extracts frame data from an FSEQ file.
func handleFrames(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	startFrame := tools.GetIntParam(req, "start_frame", 0)
	count := tools.GetIntParam(req, "count", 1)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	frames, err := client.ReadFrames(path, startFrame, count)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to read frames: %w", err)), nil
	}

	result := make([]map[string]interface{}, len(frames))
	for i, f := range frames {
		result[i] = map[string]interface{}{
			"frame_index":   f.FrameIndex,
			"channel_count": f.ChannelCount,
			"preview":       f.Preview,
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"path":        path,
		"start_frame": startFrame,
		"count":       len(frames),
		"frames":      result,
	}), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
