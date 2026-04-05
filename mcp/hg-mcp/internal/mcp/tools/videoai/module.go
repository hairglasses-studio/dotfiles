// Package videoai provides MCP video AI processing tools.
package videoai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for video AI
type Module struct{}

var getClient = tools.LazyClient(clients.NewVideoAIClient)

func (m *Module) Name() string {
	return "videoai"
}

func (m *Module) Description() string {
	return "AI-powered video processing and enhancement"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// === Enhancement Tools (6) ===
		{
			Tool: mcp.NewTool("aftrs_video_upscale",
				mcp.WithDescription("Upscale video resolution using AI models (RealESRGAN, Video2X)."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithNumber("scale", mcp.Description("Scale factor: 2 or 4 (default: 2)")),
				mcp.WithString("model", mcp.Description("Model: realesrgan, video2x (default: realesrgan)")),
			),
			Handler:             handleUpscale,
			Category:            "videoai",
			Subcategory:         "enhancement",
			Tags:                []string{"video", "upscale", "ai", "resolution", "quality"},
			UseCases:            []string{"Upscale low-res footage", "Enhance video quality"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_video_denoise",
				mcp.WithDescription("Remove noise from video using AI denoising models."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithNumber("strength", mcp.Description("Denoising strength 0.0-1.0 (default: auto)")),
				mcp.WithString("model", mcp.Description("Model: fastdvdnet, videnn (default: fastdvdnet)")),
			),
			Handler:             handleDenoise,
			Category:            "videoai",
			Subcategory:         "enhancement",
			Tags:                []string{"video", "denoise", "ai", "noise", "quality"},
			UseCases:            []string{"Clean up noisy footage", "Reduce grain"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_video_interpolate",
				mcp.WithDescription("Increase frame rate using AI frame interpolation (RIFE, FILM)."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithNumber("multiplier", mcp.Description("Frame rate multiplier: 2, 4, 8 (default: 2)")),
				mcp.WithString("model", mcp.Description("Model: rife, film (default: rife)")),
			),
			Handler:             handleInterpolate,
			Category:            "videoai",
			Subcategory:         "enhancement",
			Tags:                []string{"video", "interpolate", "framerate", "slowmo", "ai"},
			UseCases:            []string{"Create slow motion", "Increase FPS"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_video_stabilize",
				mcp.WithDescription("Stabilize shaky video footage."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithNumber("smoothing", mcp.Description("Smoothing factor 0.0-1.0 (default: 0.5)")),
			),
			Handler:             handleStabilize,
			Category:            "videoai",
			Subcategory:         "enhancement",
			Tags:                []string{"video", "stabilize", "shake", "smooth"},
			UseCases:            []string{"Fix shaky footage", "Smooth handheld video"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_video_face_restore",
				mcp.WithDescription("Restore and enhance faces in video using GFPGAN or CodeFormer."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("model", mcp.Description("Model: gfpgan, codeformer (default: gfpgan)")),
			),
			Handler:             handleFaceRestore,
			Category:            "videoai",
			Subcategory:         "enhancement",
			Tags:                []string{"video", "face", "restore", "ai", "enhancement"},
			UseCases:            []string{"Enhance faces in video", "Restore old footage"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_video_enhance",
				mcp.WithDescription("Combined enhancement pipeline: denoise + upscale + optional face restore."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithBoolean("denoise", mcp.Description("Apply denoising (default: true)")),
				mcp.WithNumber("upscale", mcp.Description("Upscale factor: 1, 2, 4 (default: 2)")),
				mcp.WithBoolean("face_restore", mcp.Description("Restore faces (default: false)")),
				mcp.WithBoolean("stabilize", mcp.Description("Stabilize video (default: false)")),
			),
			Handler:             handleEnhance,
			Category:            "videoai",
			Subcategory:         "enhancement",
			Tags:                []string{"video", "enhance", "ai", "quality", "pipeline"},
			UseCases:            []string{"One-click video enhancement", "Batch quality improvement"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		// === Segmentation Tools (4) ===
		{
			Tool: mcp.NewTool("aftrs_video_segment",
				mcp.WithDescription("Extract objects from video using SAM2 or Grounded SAM2 with text prompts."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("concepts", mcp.Description("Comma-separated objects to extract (e.g., 'person,car')")),
				mcp.WithString("model", mcp.Description("Model: sam2, grounded_sam2 (default: sam2)")),
			),
			Handler:             handleSegment,
			Category:            "videoai",
			Subcategory:         "segmentation",
			Tags:                []string{"video", "segment", "sam", "object", "extract"},
			UseCases:            []string{"Extract objects for compositing", "Create masks"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_matte",
				mcp.WithDescription("Remove background from video using RobustVideoMatting or MODNet."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("background", mcp.Description("Background color (e.g., '0,255,0' for green) or 'transparent'")),
				mcp.WithString("model", mcp.Description("Model: rvm, modnet (default: rvm)")),
			),
			Handler:             handleMatte,
			Category:            "videoai",
			Subcategory:         "segmentation",
			Tags:                []string{"video", "matte", "background", "greenscreen", "alpha"},
			UseCases:            []string{"Remove background", "Create alpha matte"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_depth",
				mcp.WithDescription("Generate depth map from video using Depth Anything or Video Depth Anything."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("model", mcp.Description("Model: depth_anything, video_depth (default: depth_anything)")),
			),
			Handler:             handleDepth,
			Category:            "videoai",
			Subcategory:         "segmentation",
			Tags:                []string{"video", "depth", "3d", "map", "ai"},
			UseCases:            []string{"Create depth maps for VFX", "3D video effects"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_inpaint",
				mcp.WithDescription("Remove objects from video using ProPainter or E2FGVI inpainting."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("mask", mcp.Required(), mcp.Description("Mask video file path (white = remove)")),
				mcp.WithString("model", mcp.Description("Model: propainter, e2fgvi (default: propainter)")),
			),
			Handler:             handleInpaint,
			Category:            "videoai",
			Subcategory:         "segmentation",
			Tags:                []string{"video", "inpaint", "remove", "object", "ai"},
			UseCases:            []string{"Remove unwanted objects", "Clean up footage"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		// === Creative Tools (4) ===
		{
			Tool: mcp.NewTool("aftrs_video_colorize",
				mcp.WithDescription("Colorize black and white video using DeOldify."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("model", mcp.Description("Model: deoldify (default: deoldify)")),
			),
			Handler:             handleColorize,
			Category:            "videoai",
			Subcategory:         "creative",
			Tags:                []string{"video", "colorize", "bw", "restore", "ai"},
			UseCases:            []string{"Colorize old footage", "Add color to B&W video"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_style_transfer",
				mcp.WithDescription("Apply artistic style transfer to video."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("style", mcp.Required(), mcp.Description("Style image path or preset name")),
				mcp.WithNumber("strength", mcp.Description("Style strength 0.0-1.0 (default: 0.8)")),
			),
			Handler:             handleStyleTransfer,
			Category:            "videoai",
			Subcategory:         "creative",
			Tags:                []string{"video", "style", "transfer", "artistic", "ai"},
			UseCases:            []string{"Apply artistic styles", "Create stylized video"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_flow",
				mcp.WithDescription("Generate optical flow visualization using RAFT or UniMatch."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("model", mcp.Description("Model: raft, unimatch (default: raft)")),
			),
			Handler:             handleFlow,
			Category:            "videoai",
			Subcategory:         "creative",
			Tags:                []string{"video", "flow", "optical", "motion", "visualization"},
			UseCases:            []string{"Visualize motion", "Create flow effects"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_capabilities",
				mcp.WithDescription("List available video AI capabilities and check GPU status."),
			),
			Handler:             handleCapabilities,
			Category:            "videoai",
			Subcategory:         "info",
			Tags:                []string{"video", "capabilities", "status", "gpu"},
			UseCases:            []string{"Check available features", "Verify GPU setup"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "videoai",
		},
		// === Pipeline Tools (4) ===
		{
			Tool: mcp.NewTool("aftrs_video_pipeline_run",
				mcp.WithDescription("Run a multi-step video processing pipeline."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("steps", mcp.Required(), mcp.Description("Comma-separated steps (e.g., 'denoise,upscale:scale=4,face')")),
			),
			Handler:             handlePipelineRun,
			Category:            "videoai",
			Subcategory:         "pipeline",
			Tags:                []string{"video", "pipeline", "batch", "workflow"},
			UseCases:            []string{"Run custom processing chain", "Automate video workflows"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_batch",
				mcp.WithDescription("Batch process multiple videos with parallel workers."),
				mcp.WithString("inputs", mcp.Required(), mcp.Description("Comma-separated input file paths or glob pattern")),
				mcp.WithString("operation", mcp.Required(), mcp.Description("Operation: upscale, denoise, interpolate, etc.")),
				mcp.WithNumber("workers", mcp.Description("Parallel workers (default: 4)")),
			),
			Handler:             handleBatch,
			Category:            "videoai",
			Subcategory:         "pipeline",
			Tags:                []string{"video", "batch", "parallel", "bulk"},
			UseCases:            []string{"Process multiple videos", "Overnight batch jobs"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		// === DXV3 Conversion (1) ===
		{
			Tool: mcp.NewTool("aftrs_video_to_dxv3",
				mcp.WithDescription("Convert video to Resolume DXV3 format for optimal VJ playback performance."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("quality", mcp.Description("Quality: high, medium, low (default: high)")),
				mcp.WithBoolean("alpha", mcp.Description("Preserve alpha channel (default: false)")),
			),
			Handler:             handleToDXV3,
			Category:            "videoai",
			Subcategory:         "conversion",
			Tags:                []string{"video", "dxv3", "resolume", "vj", "conversion"},
			UseCases:            []string{"Convert for Resolume", "Optimize VJ clips", "Export with alpha"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
			Timeout:             5 * time.Minute,
		},
		// === Queue Management (4) ===
		{
			Tool: mcp.NewTool("aftrs_video_queue_add",
				mcp.WithDescription("Add a video processing job to the queue for background processing."),
				mcp.WithString("input", mcp.Required(), mcp.Description("Input video file path")),
				mcp.WithString("operation", mcp.Required(), mcp.Description("Operation: upscale, denoise, segment, etc.")),
				mcp.WithString("params", mcp.Description("JSON string of operation parameters")),
			),
			Handler:             handleQueueAdd,
			Category:            "videoai",
			Subcategory:         "queue",
			Tags:                []string{"video", "queue", "job", "background"},
			UseCases:            []string{"Queue jobs for later", "Background processing"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_queue_list",
				mcp.WithDescription("List all jobs in the video processing queue."),
			),
			Handler:             handleQueueList,
			Category:            "videoai",
			Subcategory:         "queue",
			Tags:                []string{"video", "queue", "list", "status"},
			UseCases:            []string{"Check processing status", "View pending jobs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "videoai",
		},
		{
			Tool: mcp.NewTool("aftrs_video_queue_status",
				mcp.WithDescription("Get status of a specific processing job."),
				mcp.WithString("job_id", mcp.Required(), mcp.Description("Job ID to check")),
			),
			Handler:             handleQueueStatus,
			Category:            "videoai",
			Subcategory:         "queue",
			Tags:                []string{"video", "queue", "status", "job"},
			UseCases:            []string{"Check job progress", "Get job details"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "videoai",
		},
		{
			Tool: mcp.NewTool("aftrs_video_queue_cancel",
				mcp.WithDescription("Cancel a queued or running job."),
				mcp.WithString("job_id", mcp.Required(), mcp.Description("Job ID to cancel")),
			),
			Handler:             handleQueueCancel,
			Category:            "videoai",
			Subcategory:         "queue",
			Tags:                []string{"video", "queue", "cancel", "job"},
			UseCases:            []string{"Stop processing", "Cancel job"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_queue_clear",
				mcp.WithDescription("Clear completed and failed jobs from the queue."),
			),
			Handler:             handleQueueClear,
			Category:            "videoai",
			Subcategory:         "queue",
			Tags:                []string{"video", "queue", "clear", "cleanup"},
			UseCases:            []string{"Clean up queue", "Remove old jobs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "videoai",
			IsWrite:             true,
		},
	}
}

// === Handler Functions ===

func handleUpscale(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	scale := tools.GetIntParam(req, "scale", 2)
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Upscale(ctx, input, scale, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleDenoise(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	strength := tools.GetFloatParam(req, "strength", 0)
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Denoise(ctx, input, strength, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleInterpolate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	multiplier := tools.GetIntParam(req, "multiplier", 2)
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Interpolate(ctx, input, multiplier, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleStabilize(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	smoothing := tools.GetFloatParam(req, "smoothing", 0.5)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Stabilize(ctx, input, smoothing)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleFaceRestore(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.FaceRestore(ctx, input, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleEnhance(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	denoise := tools.GetBoolParam(req, "denoise", true)
	upscale := tools.GetIntParam(req, "upscale", 2)
	faceRestore := tools.GetBoolParam(req, "face_restore", false)
	stabilize := tools.GetBoolParam(req, "stabilize", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	opts := clients.EnhanceOptions{
		Denoise:     denoise,
		Upscale:     upscale,
		FaceRestore: faceRestore,
		Stabilize:   stabilize,
	}

	result, err := client.Enhance(ctx, input, opts)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleSegment(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	conceptsStr := tools.GetStringParam(req, "concepts")
	model := tools.GetStringParam(req, "model")

	var concepts []string
	if conceptsStr != "" {
		concepts = strings.Split(conceptsStr, ",")
		for i := range concepts {
			concepts[i] = strings.TrimSpace(concepts[i])
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Segment(ctx, input, concepts, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleMatte(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	background := tools.GetStringParam(req, "background")
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Matte(ctx, input, background, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleDepth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Depth(ctx, input, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleInpaint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	mask := tools.GetStringParam(req, "mask")
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Inpaint(ctx, input, mask, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleColorize(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Colorize(ctx, input, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleStyleTransfer(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	style := tools.GetStringParam(req, "style")
	strength := tools.GetFloatParam(req, "strength", 0.8)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.StyleTransfer(ctx, input, style, strength)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleFlow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	model := tools.GetStringParam(req, "model")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.OpticalFlow(ctx, input, model)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleCapabilities(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	caps, err := client.GetCapabilities(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Video AI Capabilities\n\n")

	if caps.Installed {
		sb.WriteString("**Status:** ✅ Installed\n")
	} else {
		sb.WriteString("**Status:** ❌ Not installed (set VIDTOOL_PATH or install video-ai-toolkit)\n")
	}

	if caps.GPUAvailable {
		sb.WriteString("**GPU:** ✅ CUDA available\n")
	} else {
		sb.WriteString("**GPU:** ⚠️ Not available (CPU mode)\n")
	}

	sb.WriteString(fmt.Sprintf("**Models Path:** `%s`\n\n", caps.ModelsPath))

	sb.WriteString("## Enhancement\n")
	for _, cap := range caps.Enhancement {
		sb.WriteString(fmt.Sprintf("- %s\n", cap))
	}

	sb.WriteString("\n## Segmentation\n")
	for _, cap := range caps.Segmentation {
		sb.WriteString(fmt.Sprintf("- %s\n", cap))
	}

	sb.WriteString("\n## Creative\n")
	for _, cap := range caps.Creative {
		sb.WriteString(fmt.Sprintf("- %s\n", cap))
	}

	return tools.TextResult(sb.String()), nil
}

func handlePipelineRun(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	stepsStr := tools.GetStringParam(req, "steps")

	steps := strings.Split(stepsStr, ",")
	for i := range steps {
		steps[i] = strings.TrimSpace(steps[i])
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.RunPipeline(ctx, input, steps)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleBatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	inputsStr := tools.GetStringParam(req, "inputs")
	operation := tools.GetStringParam(req, "operation")
	workers := tools.GetIntParam(req, "workers", 4)

	inputs := strings.Split(inputsStr, ",")
	for i := range inputs {
		inputs[i] = strings.TrimSpace(inputs[i])
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	results, err := client.BatchProcess(ctx, inputs, operation, workers)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Batch Processing Results\n\n")
	sb.WriteString(fmt.Sprintf("**Operation:** %s\n", operation))
	sb.WriteString(fmt.Sprintf("**Workers:** %d\n", workers))
	sb.WriteString(fmt.Sprintf("**Files:** %d\n\n", len(results)))

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
			sb.WriteString(fmt.Sprintf("✅ %s\n", r.InputFile))
		} else {
			sb.WriteString(fmt.Sprintf("❌ %s: %s\n", r.InputFile, r.Error))
		}
	}

	sb.WriteString(fmt.Sprintf("\n**Summary:** %d/%d successful\n", successCount, len(results)))

	return tools.TextResult(sb.String()), nil
}

func handleToDXV3(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	quality := tools.GetStringParam(req, "quality")
	alpha := tools.GetBoolParam(req, "alpha", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.ConvertToDXV3(ctx, input, quality, alpha)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return formatResult(result), nil
}

func handleQueueAdd(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	input := tools.GetStringParam(req, "input")
	operation := tools.GetStringParam(req, "operation")
	paramsStr := tools.GetStringParam(req, "params")

	var params map[string]interface{}
	if paramsStr != "" {
		if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid params JSON: %v", err)), nil
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	job, err := client.QueueJob(ctx, input, operation, params)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Job Queued\n\n")
	sb.WriteString(fmt.Sprintf("**Job ID:** `%s`\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", job.Input))
	sb.WriteString(fmt.Sprintf("**Operation:** %s\n", job.Operation))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", job.Status))

	return tools.TextResult(sb.String()), nil
}

func handleQueueList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	jobs, err := client.GetQueue(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Video Processing Queue\n\n")

	if len(jobs) == 0 {
		sb.WriteString("*Queue is empty*\n")
	} else {
		sb.WriteString(fmt.Sprintf("**Total Jobs:** %d\n\n", len(jobs)))

		// Group by status
		statusCounts := map[string]int{}
		for _, job := range jobs {
			statusCounts[job.Status]++
		}
		for status, count := range statusCounts {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", status, count))
		}
		sb.WriteString("\n")

		// List jobs
		sb.WriteString("| ID | Operation | Status | Progress |\n")
		sb.WriteString("|---|---|---|---|\n")
		for _, job := range jobs {
			progress := fmt.Sprintf("%.0f%%", job.Progress*100)
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", job.ID, job.Operation, job.Status, progress))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleQueueStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := tools.GetStringParam(req, "job_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	job, err := client.GetJobStatus(ctx, jobID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Job Status\n\n")
	sb.WriteString(fmt.Sprintf("**Job ID:** `%s`\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", job.Input))
	if job.Output != "" {
		sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", job.Output))
	}
	sb.WriteString(fmt.Sprintf("**Operation:** %s\n", job.Operation))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", job.Status))
	sb.WriteString(fmt.Sprintf("**Progress:** %.0f%%\n", job.Progress*100))

	if job.StartedAt != nil {
		sb.WriteString(fmt.Sprintf("**Started:** %s\n", job.StartedAt.Format(time.RFC3339)))
	}
	if job.FinishedAt != nil {
		sb.WriteString(fmt.Sprintf("**Finished:** %s\n", job.FinishedAt.Format(time.RFC3339)))
	}
	if job.Duration > 0 {
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", job.Duration.Round(time.Second)))
	}
	if job.Error != "" {
		sb.WriteString(fmt.Sprintf("\n**Error:** %s\n", job.Error))
	}

	return tools.TextResult(sb.String()), nil
}

func handleQueueCancel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := tools.GetStringParam(req, "job_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.CancelJob(ctx, jobID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Job `%s` cancelled successfully.", jobID)), nil
}

func handleQueueClear(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.ClearQueue(ctx); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult("Queue cleared successfully. Completed and failed jobs removed."), nil
}

// formatResult formats a VideoAIResult as markdown
func formatResult(result *clients.VideoAIResult) *mcp.CallToolResult {
	var sb strings.Builder

	if result.Success {
		sb.WriteString("# Video Processing Complete\n\n")
		sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", result.InputFile))
		sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", result.OutputFile))
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", result.Duration.Round(100*time.Millisecond)))
		if result.Model != "" {
			sb.WriteString(fmt.Sprintf("**Model:** %s\n", result.Model))
		}
		if result.Message != "" {
			sb.WriteString(fmt.Sprintf("\n%s\n", result.Message))
		}
	} else {
		sb.WriteString("# Processing Failed\n\n")
		sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", result.InputFile))
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", result.Error))
	}

	return tools.TextResult(sb.String())
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
