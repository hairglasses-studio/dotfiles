// Package stems provides MCP tools for stem separation using Demucs.
package stems

import (
	"context"
	"fmt"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Stems tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "stems"
}

// Description returns the module description
func (m *Module) Description() string {
	return "AI-powered stem separation using Demucs for vocals, drums, bass, and other"
}

// Tools returns the Stems tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_stems_status",
				mcp.WithDescription("Get stem separation service status including GPU availability and active jobs"),
			),
			Handler:             handleStemsStatus,
			Category:            "stems",
			Subcategory:         "status",
			Tags:                []string{"stems", "demucs", "status", "gpu"},
			UseCases:            []string{"Check service status", "Verify GPU availability", "View active jobs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "stems",
		},
		{
			Tool: mcp.NewTool("aftrs_stems_separate",
				mcp.WithDescription("Separate an audio track into stems (vocals, drums, bass, other)"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the audio file to separate")),
			),
			Handler:             handleStemsSeparate,
			Category:            "stems",
			Subcategory:         "separation",
			Tags:                []string{"stems", "demucs", "separate", "vocals", "drums"},
			UseCases:            []string{"Extract vocals", "Isolate drums", "Create acapella", "Get instrumental"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "stems",
			IsWrite:             true,
			Timeout:             10 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_stems_queue",
				mcp.WithDescription("Queue a track for background stem separation"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the audio file to separate")),
			),
			Handler:             handleStemsQueue,
			Category:            "stems",
			Subcategory:         "separation",
			Tags:                []string{"stems", "queue", "background", "async"},
			UseCases:            []string{"Queue for later", "Background processing", "Batch separation"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "stems",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_stems_job",
				mcp.WithDescription("Get status of a stem separation job"),
				mcp.WithString("job_id", mcp.Required(), mcp.Description("Job ID to check status for")),
			),
			Handler:             handleStemsJob,
			Category:            "stems",
			Subcategory:         "jobs",
			Tags:                []string{"stems", "job", "status", "progress"},
			UseCases:            []string{"Check job status", "Get job result", "Monitor progress"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "stems",
		},
		{
			Tool: mcp.NewTool("aftrs_stems_list",
				mcp.WithDescription("List available stems for a track or all separated tracks"),
				mcp.WithString("track_name", mcp.Description("Track name to list stems for (optional - lists all if not specified)")),
			),
			Handler:             handleStemsList,
			Category:            "stems",
			Subcategory:         "library",
			Tags:                []string{"stems", "list", "browse", "library"},
			UseCases:            []string{"Browse available stems", "Find separated tracks", "List stem files"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "stems",
		},
		{
			Tool: mcp.NewTool("aftrs_stems_models",
				mcp.WithDescription("List available Demucs models for stem separation"),
			),
			Handler:             handleStemsModels,
			Category:            "stems",
			Subcategory:         "configuration",
			Tags:                []string{"stems", "models", "demucs", "config"},
			UseCases:            []string{"View available models", "Choose separation quality", "Model comparison"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "stems",
		},
		{
			Tool: mcp.NewTool("aftrs_stems_health",
				mcp.WithDescription("Check stem separation service health and get troubleshooting recommendations"),
			),
			Handler:             handleStemsHealth,
			Category:            "stems",
			Subcategory:         "status",
			Tags:                []string{"stems", "health", "diagnostics", "troubleshooting"},
			UseCases:            []string{"Diagnose issues", "Check demucs installation", "Verify GPU setup"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "stems",
		},
	}
}

var getStemsClient = tools.LazyClient(clients.NewStemsClient)

func handleStemsStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getStemsClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create stems client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleStemsSeparate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getStemsClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create stems client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	result, err := client.SeparateStems(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to separate stems: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleStemsQueue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getStemsClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create stems client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	job, err := client.QueueSeparation(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to queue separation: %w", err)), nil
	}

	result := map[string]interface{}{
		"job_id":    job.ID,
		"file_path": job.FilePath,
		"status":    job.Status,
		"message":   fmt.Sprintf("Stem separation queued with job ID: %s", job.ID),
	}
	return tools.JSONResult(result), nil
}

func handleStemsJob(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getStemsClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create stems client: %w", err)), nil
	}

	jobID, errResult := tools.RequireStringParam(req, "job_id")
	if errResult != nil {
		return errResult, nil
	}

	job, err := client.GetJob(ctx, jobID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get job: %w", err)), nil
	}

	return tools.JSONResult(job), nil
}

func handleStemsList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getStemsClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create stems client: %w", err)), nil
	}

	trackName := tools.GetStringParam(req, "track_name")

	if trackName != "" {
		// List stems for specific track
		stems, err := client.ListStems(ctx, trackName)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to list stems: %w", err)), nil
		}

		result := map[string]interface{}{
			"track_name": trackName,
			"stems":      stems,
			"count":      len(stems),
		}
		return tools.JSONResult(result), nil
	}

	// List all available stems
	allStems, err := client.ListAllStems(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list stems: %w", err)), nil
	}

	result := map[string]interface{}{
		"tracks":      allStems,
		"track_count": len(allStems),
	}
	return tools.JSONResult(result), nil
}

func handleStemsModels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getStemsClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create stems client: %w", err)), nil
	}

	models, err := client.GetModels(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get models: %w", err)), nil
	}

	result := map[string]interface{}{
		"models": models,
		"count":  len(models),
		"descriptions": map[string]string{
			"htdemucs":    "Default hybrid transformer model - best quality",
			"htdemucs_ft": "Fine-tuned version - slightly better on some genres",
			"htdemucs_6s": "6 stems including piano and guitar",
			"hdemucs_mmi": "Hybrid demucs with MMI training",
			"mdx":         "MDX-Net - faster processing",
			"mdx_extra":   "MDX-Net extra - better quality than mdx",
			"mdx_q":       "MDX-Net quantized - fastest",
			"mdx_extra_q": "MDX-Net extra quantized - fast with good quality",
		},
	}
	return tools.JSONResult(result), nil
}

func handleStemsHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getStemsClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create stems client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
