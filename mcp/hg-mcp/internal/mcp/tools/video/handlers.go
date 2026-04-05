package video

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleProcess handles the aftrs_video_process tool
func handleProcess(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	videoPath, errResult := tools.RequireStringParam(req, "video_path")
	if errResult != nil {
		return errResult, nil
	}

	processor, errResult := tools.RequireStringParam(req, "processor")
	if errResult != nil {
		return errResult, nil
	}

	// Validate video file exists
	if !fileExists(videoPath) {
		return tools.ErrorResult(fmt.Errorf("video file not found: %s", videoPath)), nil
	}

	// Build command arguments
	args := []string{processor, videoPath}

	// Add device
	args = append(args, "-d", getDevice())

	// Add output directory if specified
	outputDir := tools.GetStringParam(req, "output_dir")
	if outputDir == "" {
		outputDir = getOutputDir()
	}
	if outputDir != "" {
		args = append(args, "-o", outputDir)
	}

	// Parse and add processor-specific parameters
	params := tools.GetStringParam(req, "params")
	if params != "" {
		for _, pair := range strings.Split(params, ",") {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				args = append(args, fmt.Sprintf("--%s", key), value)
			}
		}
	}

	// Execute vidtool
	stdout, stderr, err := runVidtool(ctx, args...)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Video Processing: %s\n\n", processor))
	sb.WriteString(fmt.Sprintf("**Input:** %s\n", videoPath))
	sb.WriteString(fmt.Sprintf("**Processor:** %s\n", processor))
	if params != "" {
		sb.WriteString(fmt.Sprintf("**Parameters:** %s\n", params))
	}
	sb.WriteString("\n---\n\n")

	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString(fmt.Sprintf("Processing failed: %v\n\n", err))
		if stderr != "" {
			sb.WriteString("**Details:**\n```\n")
			sb.WriteString(stderr)
			sb.WriteString("\n```\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Output\n\n")
	sb.WriteString("```\n")
	sb.WriteString(stdout)
	sb.WriteString("\n```\n")

	return tools.TextResult(sb.String()), nil
}

// handlePipeline handles the aftrs_video_pipeline tool
func handlePipeline(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	videoPath, errResult := tools.RequireStringParam(req, "video_path")
	if errResult != nil {
		return errResult, nil
	}

	steps, errResult := tools.RequireStringParam(req, "steps")
	if errResult != nil {
		return errResult, nil
	}

	// Validate video file exists
	if !fileExists(videoPath) {
		return tools.ErrorResult(fmt.Errorf("video file not found: %s", videoPath)), nil
	}

	// Build command arguments
	args := []string{"pipeline", "run", videoPath, "--steps", steps}

	// Add device
	args = append(args, "-d", getDevice())

	// Add output directory if specified
	outputDir := tools.GetStringParam(req, "output_dir")
	if outputDir == "" {
		outputDir = getOutputDir()
	}
	if outputDir != "" {
		args = append(args, "-o", outputDir)
	}

	// Execute vidtool
	stdout, stderr, err := runVidtool(ctx, args...)

	var sb strings.Builder
	sb.WriteString("# Video Pipeline\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** %s\n", videoPath))
	sb.WriteString(fmt.Sprintf("**Steps:** %s\n", steps))
	sb.WriteString("\n---\n\n")

	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString(fmt.Sprintf("Pipeline failed: %v\n\n", err))
		if stderr != "" {
			sb.WriteString("**Details:**\n```\n")
			sb.WriteString(stderr)
			sb.WriteString("\n```\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Output\n\n")
	sb.WriteString("```\n")
	sb.WriteString(stdout)
	sb.WriteString("\n```\n")

	return tools.TextResult(sb.String()), nil
}

// handleRandom handles the aftrs_video_random tool
func handleRandom(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	videoPath, errResult := tools.RequireStringParam(req, "video_path")
	if errResult != nil {
		return errResult, nil
	}

	// Validate video file exists
	if !fileExists(videoPath) {
		return tools.ErrorResult(fmt.Errorf("video file not found: %s", videoPath)), nil
	}

	// Build command arguments
	args := []string{"pipeline", "random", videoPath}

	// Add optional parameters
	minSteps := tools.GetIntParam(req, "min_steps", 0)
	if minSteps > 0 {
		args = append(args, "--min-steps", fmt.Sprintf("%d", minSteps))
	}

	maxSteps := tools.GetIntParam(req, "max_steps", 0)
	if maxSteps > 0 {
		args = append(args, "--max-steps", fmt.Sprintf("%d", maxSteps))
	}

	categories := tools.GetStringParam(req, "categories")
	if categories != "" {
		args = append(args, "--categories", categories)
	}

	exclude := tools.GetStringParam(req, "exclude")
	if exclude != "" {
		args = append(args, "--exclude", exclude)
	}

	seed := tools.GetIntParam(req, "seed", 0)
	if seed > 0 {
		args = append(args, "--seed", fmt.Sprintf("%d", seed))
	}

	preview := tools.GetBoolParam(req, "preview", false)
	if preview {
		args = append(args, "--preview")
	}

	// Add device
	args = append(args, "-d", getDevice())

	// Execute vidtool
	stdout, stderr, err := runVidtool(ctx, args...)

	var sb strings.Builder
	sb.WriteString("# Random Video Pipeline\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** %s\n", videoPath))
	if minSteps > 0 || maxSteps > 0 {
		sb.WriteString(fmt.Sprintf("**Steps:** %d-%d\n", minSteps, maxSteps))
	}
	if categories != "" {
		sb.WriteString(fmt.Sprintf("**Categories:** %s\n", categories))
	}
	if exclude != "" {
		sb.WriteString(fmt.Sprintf("**Excluded:** %s\n", exclude))
	}
	if seed > 0 {
		sb.WriteString(fmt.Sprintf("**Seed:** %d\n", seed))
	}
	if preview {
		sb.WriteString("**Mode:** Preview only\n")
	}
	sb.WriteString("\n---\n\n")

	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString(fmt.Sprintf("Random pipeline failed: %v\n\n", err))
		if stderr != "" {
			sb.WriteString("**Details:**\n```\n")
			sb.WriteString(stderr)
			sb.WriteString("\n```\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Output\n\n")
	sb.WriteString("```\n")
	sb.WriteString(stdout)
	sb.WriteString("\n```\n")

	return tools.TextResult(sb.String()), nil
}

// handleProcessors handles the aftrs_video_processors tool
func handleProcessors(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"pipeline", "list"}

	category := tools.GetStringParam(req, "category")
	if category != "" {
		args = append(args, "--category", category)
	}

	stdout, stderr, err := runVidtool(ctx, args...)

	var sb strings.Builder
	sb.WriteString("# Available Video Processors\n\n")

	if err != nil {
		// If vidtool is not available, return a static list
		sb.WriteString("**Note:** vidtool not available. Showing static processor list.\n\n")
		sb.WriteString(getStaticProcessorList(category))
		return tools.TextResult(sb.String()), nil
	}

	if stderr != "" && stdout == "" {
		sb.WriteString(stderr)
	} else {
		sb.WriteString(stdout)
	}

	return tools.TextResult(sb.String()), nil
}

// handleInfo handles the aftrs_video_info tool
func handleInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	videoPath, errResult := tools.RequireStringParam(req, "video_path")
	if errResult != nil {
		return errResult, nil
	}

	// Validate video file exists
	if !fileExists(videoPath) {
		return tools.ErrorResult(fmt.Errorf("video file not found: %s", videoPath)), nil
	}

	info, err := runFFprobe(ctx, videoPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get video info: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Video Information\n\n")
	sb.WriteString(fmt.Sprintf("**File:** %s\n\n", videoPath))
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Resolution | %dx%d |\n", info.Width, info.Height))
	sb.WriteString(fmt.Sprintf("| FPS | %.2f |\n", info.FPS))
	sb.WriteString(fmt.Sprintf("| Duration | %s (%.1f sec) |\n", formatDuration(info.Duration), info.Duration))
	sb.WriteString(fmt.Sprintf("| Frame Count | %d |\n", info.FrameCount))
	sb.WriteString(fmt.Sprintf("| Codec | %s |\n", info.Codec))
	sb.WriteString(fmt.Sprintf("| Format | %s |\n", info.Format))

	return tools.TextResult(sb.String()), nil
}

// getStaticProcessorList returns a static list of processors when vidtool is unavailable
func getStaticProcessorList(category string) string {
	processors := map[string][]string{
		"enhancement": {
			"- **upscale** (Real-ESRGAN): AI video upscaling (2x, 4x)",
			"- **denoise** (FastDVDnet): Real-time video denoising",
			"- **stabilize** (DeepStab): Video stabilization",
			"- **face** (CodeFormer): Face restoration",
			"- **interpolate** (RIFE): Frame interpolation",
		},
		"analysis": {
			"- **depth** (Video Depth Anything): Depth estimation",
			"- **flow** (RAFT): Optical flow estimation",
			"- **segment** (SAM 2): Object segmentation",
		},
		"creative": {
			"- **style** (AdaIN): Artistic style transfer",
			"- **colorize** (DeOldify): Video colorization",
		},
		"composition": {
			"- **matte** (RobustVideoMatting): Background removal",
			"- **inpaint** (ProPainter): Object removal/inpainting",
		},
		"generation": {
			"- **generate** (Wan 2.1): Text/image to video",
		},
	}

	var sb strings.Builder

	if category != "" {
		if procs, ok := processors[category]; ok {
			sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(category)))
			for _, p := range procs {
				sb.WriteString(p + "\n")
			}
		} else {
			sb.WriteString(fmt.Sprintf("Unknown category: %s\n\n", category))
			sb.WriteString("Available categories: enhancement, analysis, creative, composition, generation\n")
		}
	} else {
		for cat, procs := range processors {
			sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(cat)))
			for _, p := range procs {
				sb.WriteString(p + "\n")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
