// Package ndicv provides NDI computer vision tools for hg-mcp.
package ndicv

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for NDI Computer Vision
type Module struct{}

func (m *Module) Name() string {
	return "ndicv"
}

func (m *Module) Description() string {
	return "NDI computer vision and frame analysis tools"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ndi_capture_frame",
				mcp.WithDescription("Capture a single frame from an NDI source."),
				mcp.WithString("source", mcp.Required(), mcp.Description("NDI source name to capture from")),
				mcp.WithString("output_path", mcp.Description("Output file path (optional, auto-generated if not provided)")),
				mcp.WithBoolean("base64", mcp.Description("Return frame as base64 instead of file path")),
			),
			Handler:             handleCaptureFrame,
			Category:            "ndicv",
			Subcategory:         "capture",
			Tags:                []string{"ndi", "video", "capture", "frame", "snapshot"},
			UseCases:            []string{"Capture video frame", "Take NDI snapshot", "Get video still"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ndicv",
		},
		{
			Tool: mcp.NewTool("aftrs_ndi_detect_faces",
				mcp.WithDescription("Detect faces in an NDI video frame using OpenCV."),
				mcp.WithString("source", mcp.Description("NDI source name (captures new frame if provided)")),
				mcp.WithString("frame_path", mcp.Description("Path to existing frame image (used if source not provided)")),
			),
			Handler:             handleDetectFaces,
			Category:            "ndicv",
			Subcategory:         "detection",
			Tags:                []string{"ndi", "video", "face", "detection", "opencv", "cv"},
			UseCases:            []string{"Detect faces in video", "Count people on camera", "Face tracking"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ndicv",
		},
		{
			Tool: mcp.NewTool("aftrs_ndi_detect_motion",
				mcp.WithDescription("Detect motion between two video frames."),
				mcp.WithString("source", mcp.Description("NDI source name (captures two frames with delay)")),
				mcp.WithString("frame1_path", mcp.Description("Path to first frame (used if source not provided)")),
				mcp.WithString("frame2_path", mcp.Description("Path to second frame (used if source not provided)")),
				mcp.WithNumber("threshold", mcp.Description("Motion detection threshold percentage (default: 5%)")),
				mcp.WithNumber("delay_ms", mcp.Description("Delay between frame captures in ms (default: 500)")),
			),
			Handler:             handleDetectMotion,
			Category:            "ndicv",
			Subcategory:         "detection",
			Tags:                []string{"ndi", "video", "motion", "detection", "security", "cv"},
			UseCases:            []string{"Detect motion in video", "Security monitoring", "Activity detection"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ndicv",
		},
		{
			Tool: mcp.NewTool("aftrs_ndi_detect_qr",
				mcp.WithDescription("Detect and decode QR codes and barcodes in an NDI frame."),
				mcp.WithString("source", mcp.Description("NDI source name (captures new frame if provided)")),
				mcp.WithString("frame_path", mcp.Description("Path to existing frame image (used if source not provided)")),
			),
			Handler:             handleDetectQR,
			Category:            "ndicv",
			Subcategory:         "detection",
			Tags:                []string{"ndi", "video", "qr", "barcode", "detection", "scan"},
			UseCases:            []string{"Scan QR codes from video", "Read barcodes", "Decode visual data"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ndicv",
		},
		{
			Tool: mcp.NewTool("aftrs_ndi_ocr",
				mcp.WithDescription("Extract text from an NDI video frame using OCR."),
				mcp.WithString("source", mcp.Description("NDI source name (captures new frame if provided)")),
				mcp.WithString("frame_path", mcp.Description("Path to existing frame image (used if source not provided)")),
				mcp.WithString("language", mcp.Description("OCR language (default: eng)")),
			),
			Handler:             handleOCR,
			Category:            "ndicv",
			Subcategory:         "detection",
			Tags:                []string{"ndi", "video", "ocr", "text", "extraction", "tesseract"},
			UseCases:            []string{"Read text from video", "Extract on-screen text", "Video OCR"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ndicv",
		},
		{
			Tool: mcp.NewTool("aftrs_ndi_scene_change",
				mcp.WithDescription("Detect scene changes in NDI video stream."),
				mcp.WithString("source", mcp.Description("NDI source name (captures two frames with delay)")),
				mcp.WithString("frame1_path", mcp.Description("Path to first frame (used if source not provided)")),
				mcp.WithString("frame2_path", mcp.Description("Path to second frame (used if source not provided)")),
				mcp.WithNumber("threshold", mcp.Description("Scene change threshold percentage (default: 30%)")),
				mcp.WithNumber("delay_ms", mcp.Description("Delay between frame captures in ms (default: 1000)")),
			),
			Handler:             handleSceneChange,
			Category:            "ndicv",
			Subcategory:         "detection",
			Tags:                []string{"ndi", "video", "scene", "change", "detection", "cv"},
			UseCases:            []string{"Detect scene changes", "Video segmentation", "Shot detection"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ndicv",
		},
	}
}

var getClient = tools.LazyClient(clients.NewNDIClient)

// handleCaptureFrame captures a frame from an NDI source
func handleCaptureFrame(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	asBase64 := tools.GetBoolParam(req, "base64", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var frame *clients.NDIFrame
	if asBase64 {
		frame, err = client.CaptureFrameBase64(ctx, source)
	} else {
		frame, err = client.CaptureFrame(ctx, source, outputPath)
	}

	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# NDI Frame Capture\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", frame.SourceName))
	sb.WriteString(fmt.Sprintf("**Timestamp:** %s\n", frame.Timestamp))
	sb.WriteString(fmt.Sprintf("**Dimensions:** %dx%d\n", frame.Width, frame.Height))
	sb.WriteString(fmt.Sprintf("**Format:** %s\n", frame.Format))
	sb.WriteString(fmt.Sprintf("**Size:** %d bytes\n", frame.SizeBytes))

	if frame.FilePath != "" {
		sb.WriteString(fmt.Sprintf("\n**Saved to:** `%s`\n", frame.FilePath))
	}
	if frame.Base64 != "" {
		// Show truncated base64 for display
		preview := frame.Base64
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		sb.WriteString(fmt.Sprintf("\n**Base64 Preview:** `%s`\n", preview))
		sb.WriteString(fmt.Sprintf("**Full Base64 Length:** %d chars\n", len(frame.Base64)))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDetectFaces detects faces in an NDI frame
func handleDetectFaces(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	framePath := tools.GetStringParam(req, "frame_path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Capture frame if source provided
	if source != "" {
		frame, err := client.CaptureFrame(ctx, source, "")
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to capture frame: %w", err)), nil
		}
		framePath = frame.FilePath
		defer os.Remove(framePath) // Clean up temp frame
	}

	if framePath == "" {
		return tools.ErrorResult(fmt.Errorf("either source or frame_path is required")), nil
	}

	result, err := client.DetectFaces(ctx, framePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Face Detection Results\n\n")
	sb.WriteString(fmt.Sprintf("**Faces Detected:** %d\n\n", result.FaceCount))

	if result.FaceCount > 0 {
		sb.WriteString("| # | Position | Size | Confidence |\n")
		sb.WriteString("|---|----------|------|------------|\n")
		for i, face := range result.Faces {
			sb.WriteString(fmt.Sprintf("| %d | (%d, %d) | %dx%d | %.0f%% |\n",
				i+1, face.X, face.Y, face.Width, face.Height, face.Confidence*100))
		}
	} else {
		sb.WriteString("No faces detected in the frame.\n")
		sb.WriteString("\n*Requires Python + OpenCV with haarcascade_frontalface_default.xml*\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleDetectMotion detects motion between frames
func handleDetectMotion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	frame1Path := tools.GetStringParam(req, "frame1_path")
	frame2Path := tools.GetStringParam(req, "frame2_path")
	threshold := tools.GetFloatParam(req, "threshold", 5.0)
	delayMs := tools.GetIntParam(req, "delay_ms", 500)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Capture frames if source provided
	var cleanupFrames []string
	if source != "" {
		// Capture first frame
		frame1, err := client.CaptureFrame(ctx, source, "")
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to capture frame 1: %w", err)), nil
		}
		frame1Path = frame1.FilePath
		cleanupFrames = append(cleanupFrames, frame1Path)

		// Wait for delay
		time.Sleep(time.Duration(delayMs) * time.Millisecond)

		// Capture second frame
		frame2, err := client.CaptureFrame(ctx, source, "")
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to capture frame 2: %w", err)), nil
		}
		frame2Path = frame2.FilePath
		cleanupFrames = append(cleanupFrames, frame2Path)
	}

	// Clean up temp frames on exit
	defer func() {
		for _, p := range cleanupFrames {
			os.Remove(p)
		}
	}()

	if frame1Path == "" || frame2Path == "" {
		return tools.ErrorResult(fmt.Errorf("either source or both frame1_path and frame2_path are required")), nil
	}

	result, err := client.DetectMotion(ctx, frame1Path, frame2Path, threshold)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Motion Detection Results\n\n")

	if result.MotionDetected {
		sb.WriteString("**Status:** ⚠️ Motion Detected\n\n")
	} else {
		sb.WriteString("**Status:** ✅ No Significant Motion\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Motion Level:** %.2f%%\n", result.MotionPercent))
	sb.WriteString(fmt.Sprintf("**Threshold:** %.2f%%\n", result.Threshold))

	if len(result.Regions) > 0 {
		sb.WriteString("\n## Motion Regions\n\n")
		sb.WriteString("| Region | Position | Size | Score |\n")
		sb.WriteString("|--------|----------|------|-------|\n")
		for i, region := range result.Regions {
			sb.WriteString(fmt.Sprintf("| %d | (%d, %d) | %dx%d | %.2f |\n",
				i+1, region.X, region.Y, region.Width, region.Height, region.Score))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleDetectQR detects QR codes and barcodes
func handleDetectQR(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	framePath := tools.GetStringParam(req, "frame_path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Capture frame if source provided
	if source != "" {
		frame, err := client.CaptureFrame(ctx, source, "")
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to capture frame: %w", err)), nil
		}
		framePath = frame.FilePath
		defer os.Remove(framePath)
	}

	if framePath == "" {
		return tools.ErrorResult(fmt.Errorf("either source or frame_path is required")), nil
	}

	result, err := client.DetectQR(ctx, framePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# QR/Barcode Detection Results\n\n")

	if result.Found {
		sb.WriteString(fmt.Sprintf("**Codes Found:** %d\n\n", len(result.Codes)))
		sb.WriteString("| Type | Data | Position |\n")
		sb.WriteString("|------|------|----------|\n")
		for _, code := range result.Codes {
			data := code.Data
			if len(data) > 50 {
				data = data[:50] + "..."
			}
			sb.WriteString(fmt.Sprintf("| %s | `%s` | (%d, %d) |\n",
				code.Type, data, code.X, code.Y))
		}
	} else {
		sb.WriteString("**No QR codes or barcodes detected.**\n\n")
		sb.WriteString("*Requires zbarimg or Python + pyzbar*\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleOCR performs OCR on an NDI frame
func handleOCR(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	framePath := tools.GetStringParam(req, "frame_path")
	// language := tools.GetStringParam(req, "language") // Future: pass to tesseract

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Capture frame if source provided
	if source != "" {
		frame, err := client.CaptureFrame(ctx, source, "")
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to capture frame: %w", err)), nil
		}
		framePath = frame.FilePath
		defer os.Remove(framePath)
	}

	if framePath == "" {
		return tools.ErrorResult(fmt.Errorf("either source or frame_path is required")), nil
	}

	result, err := client.DetectText(ctx, framePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# OCR Results\n\n")

	if result.Text != "" {
		sb.WriteString(fmt.Sprintf("**Confidence:** %.0f%%\n\n", result.Confidence*100))
		sb.WriteString("## Extracted Text\n\n")
		sb.WriteString("```\n")
		sb.WriteString(result.Text)
		sb.WriteString("\n```\n")

		if len(result.Blocks) > 0 {
			sb.WriteString("\n## Text Blocks\n\n")
			for i, block := range result.Blocks {
				sb.WriteString(fmt.Sprintf("**Block %d** at (%d, %d): %s\n",
					i+1, block.X, block.Y, block.Text))
			}
		}
	} else {
		sb.WriteString("**No text detected in the frame.**\n\n")
		sb.WriteString("*Requires Tesseract OCR installed*\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleSceneChange detects scene changes
func handleSceneChange(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	frame1Path := tools.GetStringParam(req, "frame1_path")
	frame2Path := tools.GetStringParam(req, "frame2_path")
	threshold := tools.GetFloatParam(req, "threshold", 30.0)
	delayMs := tools.GetIntParam(req, "delay_ms", 1000)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Capture frames if source provided
	var cleanupFrames []string
	if source != "" {
		// Create temp directory for frames
		tmpDir := filepath.Join(os.TempDir(), "ndi_scene_change")
		os.MkdirAll(tmpDir, 0755)

		// Capture first frame
		frame1, err := client.CaptureFrame(ctx, source, filepath.Join(tmpDir, "frame1.jpg"))
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to capture frame 1: %w", err)), nil
		}
		frame1Path = frame1.FilePath
		cleanupFrames = append(cleanupFrames, frame1Path)

		// Wait for delay
		time.Sleep(time.Duration(delayMs) * time.Millisecond)

		// Capture second frame
		frame2, err := client.CaptureFrame(ctx, source, filepath.Join(tmpDir, "frame2.jpg"))
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to capture frame 2: %w", err)), nil
		}
		frame2Path = frame2.FilePath
		cleanupFrames = append(cleanupFrames, frame2Path)
	}

	// Clean up temp frames on exit
	defer func() {
		for _, p := range cleanupFrames {
			os.Remove(p)
		}
	}()

	if frame1Path == "" || frame2Path == "" {
		return tools.ErrorResult(fmt.Errorf("either source or both frame1_path and frame2_path are required")), nil
	}

	result, err := client.DetectSceneChange(ctx, frame1Path, frame2Path, threshold)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Scene Change Detection Results\n\n")

	if result.SceneChanged {
		sb.WriteString("**Status:** 🎬 Scene Change Detected\n\n")
	} else {
		sb.WriteString("**Status:** ✅ Same Scene\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Change Score:** %.2f%%\n", result.ChangeScore))
	sb.WriteString(fmt.Sprintf("**Threshold:** %.2f%%\n", result.Threshold))

	sb.WriteString("\n## Interpretation\n\n")
	if result.ChangeScore < 10 {
		sb.WriteString("- Very low change (static scene)\n")
	} else if result.ChangeScore < 30 {
		sb.WriteString("- Minor changes (camera movement, lighting)\n")
	} else if result.ChangeScore < 60 {
		sb.WriteString("- Significant change (new shot, major movement)\n")
	} else {
		sb.WriteString("- Major scene change (cut to new scene)\n")
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
