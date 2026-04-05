package samples

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// CuePoint represents a Rekordbox cue point (hot cue or memory cue)
type CuePoint struct {
	Num   int
	Name  string
	Start float64
	Type  int // 0 = hot cue, 1 = memory cue
	Red   int
	Green int
	Blue  int
}

// LoudnessMeasurement holds EBU R128 loudness measurement results
type LoudnessMeasurement struct {
	IntegratedLUFS float64
	LRA            float64
	TruePeak       float64
}

// measureLoudness measures loudness of an audio file using ffmpeg's ebur128 filter.
// Returns integrated LUFS, loudness range (LRA), and true peak.
func measureLoudness(ctx context.Context, filePath string) (*LoudnessMeasurement, error) {
	args := []string{
		"-i", filePath,
		"-af", "ebur128=peak=true",
		"-f", "null", "-",
	}
	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("loudness measurement failed: %w", err)
	}

	m := &LoudnessMeasurement{}

	// Parse the summary block from stderr
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "I:") && strings.Contains(line, "LUFS") {
			fmt.Sscanf(extractAfter(line, "I:"), "%f", &m.IntegratedLUFS)
		}
		if strings.Contains(line, "LRA:") && strings.Contains(line, "LU") {
			fmt.Sscanf(extractAfter(line, "LRA:"), "%f", &m.LRA)
		}
		if strings.Contains(line, "Peak:") && strings.Contains(line, "dBFS") {
			fmt.Sscanf(extractAfter(line, "Peak:"), "%f", &m.TruePeak)
		}
	}

	return m, nil
}

// extractAfter returns the substring after the given prefix in a line, trimmed.
func extractAfter(line, prefix string) string {
	idx := strings.Index(line, prefix)
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(line[idx+len(prefix):])
}

// --- Stub handlers: not yet implemented ---

func handleDetectKey(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}

func handleMeasureLoudness(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}

func handlePreview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}

func handleFrequencyAnalysis(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}

func handleCatalog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}

func handleCatalogSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}

func handleTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}

func handleTagBatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}

func handleFindDuplicates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.TextResult("not yet implemented"), nil
}
