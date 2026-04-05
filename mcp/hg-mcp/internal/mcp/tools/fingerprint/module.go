// Package fingerprint provides MCP tools for audio fingerprinting using Chromaprint/AcoustID.
package fingerprint

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Fingerprint tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "fingerprint"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Audio fingerprinting using Chromaprint and AcoustID for track identification"
}

// Tools returns the Fingerprint tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_fingerprint_generate",
				mcp.WithDescription("Generate an audio fingerprint for a track using Chromaprint"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the audio file to fingerprint")),
			),
			Handler:             handleFingerprintGenerate,
			Category:            "fingerprint",
			Subcategory:         "generate",
			Tags:                []string{"fingerprint", "chromaprint", "audio", "identify"},
			UseCases:            []string{"Generate fingerprint", "Prepare for matching", "Create audio signature"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "fingerprint",
		},
		{
			Tool: mcp.NewTool("aftrs_fingerprint_match",
				mcp.WithDescription("Match a fingerprint against AcoustID database to identify a track"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the audio file to identify")),
			),
			Handler:             handleFingerprintMatch,
			Category:            "fingerprint",
			Subcategory:         "match",
			Tags:                []string{"fingerprint", "acoustid", "identify", "match"},
			UseCases:            []string{"Identify unknown track", "Find track metadata", "Music recognition"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "fingerprint",
		},
		{
			Tool: mcp.NewTool("aftrs_fingerprint_duplicates",
				mcp.WithDescription("Find duplicate tracks in a directory using fingerprint comparison"),
				mcp.WithString("directory", mcp.Required(), mcp.Description("Directory to scan for duplicate tracks")),
				mcp.WithNumber("threshold", mcp.Description("Similarity threshold (0.0-1.0, default: 0.9)")),
			),
			Handler:             handleFingerprintDuplicates,
			Category:            "fingerprint",
			Subcategory:         "analysis",
			Tags:                []string{"fingerprint", "duplicates", "library", "cleanup"},
			UseCases:            []string{"Find duplicate tracks", "Clean up library", "Detect same songs"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "fingerprint",
		},
		{
			Tool: mcp.NewTool("aftrs_fingerprint_batch",
				mcp.WithDescription("Generate fingerprints for multiple audio files"),
				mcp.WithArray("file_paths", mcp.Required(), mcp.Description("List of audio file paths to fingerprint")),
			),
			Handler:             handleFingerprintBatch,
			Category:            "fingerprint",
			Subcategory:         "generate",
			Tags:                []string{"fingerprint", "batch", "bulk", "generate"},
			UseCases:            []string{"Fingerprint multiple tracks", "Batch processing", "Library analysis"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "fingerprint",
		},
		{
			Tool: mcp.NewTool("aftrs_fingerprint_health",
				mcp.WithDescription("Check fingerprinting service health and configuration"),
			),
			Handler:             handleFingerprintHealth,
			Category:            "fingerprint",
			Subcategory:         "status",
			Tags:                []string{"fingerprint", "health", "diagnostics", "status"},
			UseCases:            []string{"Check fpcalc installation", "Verify AcoustID key", "Diagnose issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "fingerprint",
		},
	}
}

var getFingerprintClient = tools.LazyClient(clients.NewFingerprintClient)

func handleFingerprintGenerate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getFingerprintClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create fingerprint client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	fp, err := client.GenerateFingerprint(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to generate fingerprint: %w", err)), nil
	}

	return tools.JSONResult(fp), nil
}

func handleFingerprintMatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getFingerprintClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create fingerprint client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	result, err := client.IdentifyTrack(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to identify track: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleFingerprintDuplicates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getFingerprintClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create fingerprint client: %w", err)), nil
	}

	directory, errResult := tools.RequireStringParam(req, "directory")
	if errResult != nil {
		return errResult, nil
	}

	threshold := tools.GetFloatParam(req, "threshold", 0.9)

	duplicates, err := client.FindDuplicates(ctx, directory, threshold)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to find duplicates: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"directory":  directory,
		"threshold":  threshold,
		"duplicates": duplicates,
		"count":      len(duplicates),
	}), nil
}

func handleFingerprintBatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getFingerprintClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create fingerprint client: %w", err)), nil
	}

	// Get file paths from arguments
	filePaths, errResult := tools.RequireStringArrayParam(req, "file_paths")
	if errResult != nil {
		return errResult, nil
	}

	fingerprints, err := client.BatchGenerate(ctx, filePaths)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to generate fingerprints: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"fingerprints": fingerprints,
		"count":        len(fingerprints),
		"requested":    len(filePaths),
	}), nil
}

func handleFingerprintHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getFingerprintClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create fingerprint client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
