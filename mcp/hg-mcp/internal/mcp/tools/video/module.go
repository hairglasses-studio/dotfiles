package video

import (
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for video processing
type Module struct{}

func (m *Module) Name() string {
	return "video"
}

func (m *Module) Description() string {
	return "AI video processing tools using video-ai-toolkit"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_video_process",
				mcp.WithDescription("Process video with an AI model (denoise, upscale, depth, style, etc.)"),
				mcp.WithString("video_path",
					mcp.Required(),
					mcp.Description("Path to input video file"),
				),
				mcp.WithString("processor",
					mcp.Required(),
					mcp.Description("Processor ID: denoise, upscale, depth, style, colorize, stabilize, face, flow, interpolate, matte, segment, inpaint, generate"),
				),
				mcp.WithString("params",
					mcp.Description("Processor-specific parameters (format: key=value,key=value). Examples: scale=4,preset=anime or sigma=25 or style=/path/to/art.jpg,alpha=0.8"),
				),
				mcp.WithString("output_dir",
					mcp.Description("Output directory for processed video (uses VIDTOOL_OUTPUT_DIR if not specified)"),
				),
			),
			Handler:             handleProcess,
			Category:            "video",
			Subcategory:         "processing",
			Tags:                []string{"video", "ai", "process", "denoise", "upscale", "depth", "style"},
			UseCases:            []string{"Upscale low-resolution video", "Remove noise", "Generate depth maps", "Apply artistic styles"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "video",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_video_pipeline",
				mcp.WithDescription("Run a multi-step video processing pipeline. Chain multiple processors in sequence."),
				mcp.WithString("video_path",
					mcp.Required(),
					mcp.Description("Path to input video file"),
				),
				mcp.WithString("steps",
					mcp.Required(),
					mcp.Description("Pipeline steps (format: processor1,processor2:param=value). Example: denoise,upscale:scale=4,style:style=/path/art.jpg"),
				),
				mcp.WithString("output_dir",
					mcp.Description("Output directory for processed video"),
				),
			),
			Handler:             handlePipeline,
			Category:            "video",
			Subcategory:         "pipeline",
			Tags:                []string{"video", "pipeline", "chain", "workflow"},
			UseCases:            []string{"Chain multiple processors", "Create processing workflows", "Batch enhance videos"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "video",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_video_random",
				mcp.WithDescription("Generate and run a random processing pipeline for experimentation. Creates unexpected combinations of processors."),
				mcp.WithString("video_path",
					mcp.Required(),
					mcp.Description("Path to input video file"),
				),
				mcp.WithNumber("min_steps",
					mcp.Description("Minimum number of processing steps (default: 2)"),
				),
				mcp.WithNumber("max_steps",
					mcp.Description("Maximum number of processing steps (default: 4)"),
				),
				mcp.WithString("categories",
					mcp.Description("Filter by categories (comma-separated): enhancement, analysis, creative, composition, generation"),
				),
				mcp.WithString("exclude",
					mcp.Description("Exclude processors (comma-separated): e.g., generate,inpaint"),
				),
				mcp.WithNumber("seed",
					mcp.Description("Random seed for reproducibility"),
				),
				mcp.WithBoolean("preview",
					mcp.Description("Preview the pipeline without running it"),
				),
			),
			Handler:             handleRandom,
			Category:            "video",
			Subcategory:         "pipeline",
			Tags:                []string{"video", "random", "experiment", "creative"},
			UseCases:            []string{"Experiment with random effects", "Discover new processing combinations", "Creative exploration"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "video",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_processors",
				mcp.WithDescription("List available video processors and their capabilities."),
				mcp.WithString("category",
					mcp.Description("Filter by category: enhancement, analysis, creative, composition, generation"),
				),
			),
			Handler:             handleProcessors,
			Category:            "video",
			Subcategory:         "discovery",
			Tags:                []string{"video", "list", "processors", "help"},
			UseCases:            []string{"Discover available processors", "Learn about capabilities", "Find the right tool"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "video",
		},
		{
			Tool: mcp.NewTool("aftrs_video_info",
				mcp.WithDescription("Get video file information (resolution, fps, duration, codec)."),
				mcp.WithString("video_path",
					mcp.Required(),
					mcp.Description("Path to video file"),
				),
			),
			Handler:             handleInfo,
			Category:            "video",
			Subcategory:         "info",
			Tags:                []string{"video", "info", "metadata", "probe"},
			UseCases:            []string{"Check video properties", "Verify resolution and fps", "Inspect video metadata"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "video",
		},
	}
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
