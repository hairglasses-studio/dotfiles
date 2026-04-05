// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VideoAIClient wraps the video-ai-toolkit (vidtool) CLI
type VideoAIClient struct {
	vidtoolPath string
	modelsPath  string
	outputDir   string
}

// VideoAIResult represents the result of a video processing operation
type VideoAIResult struct {
	Success    bool          `json:"success"`
	InputFile  string        `json:"input_file"`
	OutputFile string        `json:"output_file,omitempty"`
	Duration   time.Duration `json:"duration"`
	Model      string        `json:"model,omitempty"`
	Message    string        `json:"message,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// VideoAIPipeline represents a processing pipeline configuration
type VideoAIPipeline struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Steps       []string  `json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
}

// VideoAICapabilities lists available processing capabilities
type VideoAICapabilities struct {
	Enhancement  []string `json:"enhancement"`
	Segmentation []string `json:"segmentation"`
	Creative     []string `json:"creative"`
	Installed    bool     `json:"installed"`
	GPUAvailable bool     `json:"gpu_available"`
	ModelsPath   string   `json:"models_path"`
}

// NewVideoAIClient creates a new video AI client
func NewVideoAIClient() (*VideoAIClient, error) {
	// Find vidtool in PATH or common locations
	vidtoolPath := os.Getenv("VIDTOOL_PATH")
	if vidtoolPath == "" {
		// Try common locations
		paths := []string{
			"vidtool",
			"/usr/local/bin/vidtool",
			filepath.Join(os.Getenv("HOME"), ".local/bin/vidtool"),
			filepath.Join(os.Getenv("HOME"), "aftrs-studio/video-ai-toolkit/vidtool"),
		}
		for _, p := range paths {
			if _, err := exec.LookPath(p); err == nil {
				vidtoolPath = p
				break
			}
		}
	}

	modelsPath := os.Getenv("HF_MODEL_PATH")
	if modelsPath == "" {
		modelsPath = filepath.Join(os.Getenv("HOME"), ".cache/huggingface")
	}

	outputDir := os.Getenv("VIDTOOL_OUTPUT_DIR")
	if outputDir == "" {
		outputDir = filepath.Join(os.Getenv("HOME"), "Videos/processed")
	}

	// Ensure output directory exists
	_ = os.MkdirAll(outputDir, 0755)

	return &VideoAIClient{
		vidtoolPath: vidtoolPath,
		modelsPath:  modelsPath,
		outputDir:   outputDir,
	}, nil
}

// runVidtool executes a vidtool command
func (c *VideoAIClient) runVidtool(ctx context.Context, args ...string) (string, error) {
	if c.vidtoolPath == "" {
		return "", fmt.Errorf("vidtool not found - install video-ai-toolkit or set VIDTOOL_PATH")
	}

	cmd := exec.CommandContext(ctx, c.vidtoolPath, args...)
	cmd.Env = append(os.Environ(),
		"HF_MODEL_PATH="+c.modelsPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetCapabilities returns available video AI capabilities
func (c *VideoAIClient) GetCapabilities(ctx context.Context) (*VideoAICapabilities, error) {
	caps := &VideoAICapabilities{
		Enhancement: []string{
			"upscale", "denoise", "interpolate", "stabilize", "face",
		},
		Segmentation: []string{
			"segment", "matte", "depth", "inpaint",
		},
		Creative: []string{
			"colorize", "style", "flow", "generate",
		},
		Installed:    c.vidtoolPath != "",
		ModelsPath:   c.modelsPath,
		GPUAvailable: c.checkGPU(ctx),
	}
	return caps, nil
}

// checkGPU checks if CUDA GPU is available
func (c *VideoAIClient) checkGPU(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	return cmd.Run() == nil
}

// Upscale upscales video using AI models
func (c *VideoAIClient) Upscale(ctx context.Context, inputPath string, scale int, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "realesrgan"
	}
	if scale == 0 {
		scale = 2
	}

	outputPath := c.generateOutputPath(inputPath, fmt.Sprintf("upscale_%dx", scale))

	args := []string{"upscale", inputPath, "-o", outputPath, "-s", fmt.Sprintf("%d", scale), "-m", model}
	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Upscaled %dx using %s", scale, model)
	}

	return result, nil
}

// Denoise removes noise from video
func (c *VideoAIClient) Denoise(ctx context.Context, inputPath string, strength float64, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "fastdvdnet"
	}

	outputPath := c.generateOutputPath(inputPath, "denoise")

	args := []string{"denoise", inputPath, "-o", outputPath, "-m", model}
	if strength > 0 {
		args = append(args, "--strength", fmt.Sprintf("%.2f", strength))
	}

	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Denoised using %s", model)
	}

	return result, nil
}

// Interpolate increases frame rate using AI frame interpolation
func (c *VideoAIClient) Interpolate(ctx context.Context, inputPath string, multiplier int, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "rife"
	}
	if multiplier == 0 {
		multiplier = 2
	}

	outputPath := c.generateOutputPath(inputPath, fmt.Sprintf("interp_%dx", multiplier))

	args := []string{"interpolate", inputPath, "-o", outputPath, "-x", fmt.Sprintf("%d", multiplier), "-m", model}
	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Interpolated %dx using %s", multiplier, model)
	}

	return result, nil
}

// Stabilize stabilizes shaky video
func (c *VideoAIClient) Stabilize(ctx context.Context, inputPath string, smoothing float64) (*VideoAIResult, error) {
	start := time.Now()

	outputPath := c.generateOutputPath(inputPath, "stabilized")

	args := []string{"stabilize", inputPath, "-o", outputPath}
	if smoothing > 0 {
		args = append(args, "--smoothing", fmt.Sprintf("%.2f", smoothing))
	}

	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = "Video stabilized"
	}

	return result, nil
}

// FaceRestore restores faces in video
func (c *VideoAIClient) FaceRestore(ctx context.Context, inputPath string, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "gfpgan"
	}

	outputPath := c.generateOutputPath(inputPath, "face_restored")

	args := []string{"face", inputPath, "-o", outputPath, "-m", model}
	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Faces restored using %s", model)
	}

	return result, nil
}

// Segment extracts objects from video using SAM2
func (c *VideoAIClient) Segment(ctx context.Context, inputPath string, concepts []string, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "sam2"
	}

	outputPath := c.generateOutputPath(inputPath, "segmented")

	args := []string{"segment", inputPath, "-o", outputPath, "-m", model}
	if len(concepts) > 0 {
		args = append(args, "-c", strings.Join(concepts, ","))
	}

	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Segmented objects: %v", concepts)
	}

	return result, nil
}

// Matte removes background from video
func (c *VideoAIClient) Matte(ctx context.Context, inputPath string, bgColor string, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "rvm"
	}

	outputPath := c.generateOutputPath(inputPath, "matted")

	args := []string{"matte", inputPath, "-o", outputPath, "-m", model}
	if bgColor != "" {
		args = append(args, "-b", bgColor)
	}

	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Background removed using %s", model)
	}

	return result, nil
}

// Depth generates depth map from video
func (c *VideoAIClient) Depth(ctx context.Context, inputPath string, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "depth_anything"
	}

	outputPath := c.generateOutputPath(inputPath, "depth")

	args := []string{"depth", inputPath, "-o", outputPath, "-m", model}
	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Depth map generated using %s", model)
	}

	return result, nil
}

// Inpaint removes objects from video
func (c *VideoAIClient) Inpaint(ctx context.Context, inputPath, maskPath string, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "propainter"
	}

	outputPath := c.generateOutputPath(inputPath, "inpainted")

	args := []string{"inpaint", inputPath, "-o", outputPath, "-m", model, "--mask", maskPath}
	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Objects inpainted using %s", model)
	}

	return result, nil
}

// Colorize colorizes black and white video
func (c *VideoAIClient) Colorize(ctx context.Context, inputPath string, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "deoldify"
	}

	outputPath := c.generateOutputPath(inputPath, "colorized")

	args := []string{"colorize", inputPath, "-o", outputPath, "-m", model}
	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Colorized using %s", model)
	}

	return result, nil
}

// StyleTransfer applies artistic style to video
func (c *VideoAIClient) StyleTransfer(ctx context.Context, inputPath, stylePath string, strength float64) (*VideoAIResult, error) {
	start := time.Now()

	outputPath := c.generateOutputPath(inputPath, "styled")

	args := []string{"style", inputPath, "-o", outputPath, "--style", stylePath}
	if strength > 0 {
		args = append(args, "--strength", fmt.Sprintf("%.2f", strength))
	}

	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = "Style transfer applied"
	}

	return result, nil
}

// OpticalFlow generates optical flow visualization
func (c *VideoAIClient) OpticalFlow(ctx context.Context, inputPath string, model string) (*VideoAIResult, error) {
	start := time.Now()

	if model == "" {
		model = "raft"
	}

	outputPath := c.generateOutputPath(inputPath, "flow")

	args := []string{"flow", inputPath, "-o", outputPath, "-m", model}
	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
		Model:      model,
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Optical flow generated using %s", model)
	}

	return result, nil
}

// RunPipeline executes a multi-step processing pipeline
func (c *VideoAIClient) RunPipeline(ctx context.Context, inputPath string, steps []string) (*VideoAIResult, error) {
	start := time.Now()

	outputPath := c.generateOutputPath(inputPath, "pipeline")

	args := []string{"pipeline", "run", inputPath, "-o", outputPath, "--steps", strings.Join(steps, ",")}
	_, err := c.runVidtool(ctx, args...)

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Message = fmt.Sprintf("Pipeline completed: %s", strings.Join(steps, " → "))
	}

	return result, nil
}

// BatchProcess processes multiple files
func (c *VideoAIClient) BatchProcess(ctx context.Context, inputPaths []string, operation string, workers int) ([]*VideoAIResult, error) {
	if workers == 0 {
		workers = 4
	}

	args := []string{operation}
	args = append(args, inputPaths...)
	args = append(args, "-j", fmt.Sprintf("%d", workers), "-o", c.outputDir)

	output, err := c.runVidtool(ctx, args...)
	if err != nil {
		return nil, err
	}

	// Parse batch results
	var results []*VideoAIResult
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		// Fallback: create basic results
		for _, path := range inputPaths {
			results = append(results, &VideoAIResult{
				Success:   true,
				InputFile: path,
				Message:   fmt.Sprintf("Batch %s completed", operation),
			})
		}
	}

	return results, nil
}

// Enhance runs a combined enhancement pipeline
func (c *VideoAIClient) Enhance(ctx context.Context, inputPath string, opts EnhanceOptions) (*VideoAIResult, error) {
	var steps []string

	if opts.Denoise {
		steps = append(steps, "denoise")
	}
	if opts.Upscale > 1 {
		steps = append(steps, fmt.Sprintf("upscale:scale=%d", opts.Upscale))
	}
	if opts.Interpolate > 1 {
		steps = append(steps, fmt.Sprintf("interpolate:multiplier=%d", opts.Interpolate))
	}
	if opts.FaceRestore {
		steps = append(steps, "face")
	}
	if opts.Stabilize {
		steps = append(steps, "stabilize")
	}

	if len(steps) == 0 {
		steps = []string{"denoise", "upscale:scale=2"}
	}

	return c.RunPipeline(ctx, inputPath, steps)
}

// EnhanceOptions configures the enhance operation
type EnhanceOptions struct {
	Denoise     bool
	Upscale     int
	Interpolate int
	FaceRestore bool
	Stabilize   bool
}

// generateOutputPath creates an output path for processed video
func (c *VideoAIClient) generateOutputPath(inputPath, suffix string) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(c.outputDir, fmt.Sprintf("%s_%s%s", name, suffix, ext))
}

// VideoAIJob represents a queued processing job
type VideoAIJob struct {
	ID         string        `json:"id"`
	Input      string        `json:"input"`
	Output     string        `json:"output,omitempty"`
	Operation  string        `json:"operation"`
	Status     string        `json:"status"` // queued, running, completed, failed
	Progress   float64       `json:"progress"`
	StartedAt  *time.Time    `json:"started_at,omitempty"`
	FinishedAt *time.Time    `json:"finished_at,omitempty"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration,omitempty"`
}

// ConvertToDXV3 converts video to Resolume DXV3 format
func (c *VideoAIClient) ConvertToDXV3(ctx context.Context, inputPath string, quality string, alpha bool) (*VideoAIResult, error) {
	start := time.Now()

	if quality == "" {
		quality = "high" // high, medium, low
	}

	// DXV3 requires .mov container
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	outputPath := filepath.Join(c.outputDir, fmt.Sprintf("%s_dxv3.mov", name))

	// Use ffmpeg with DXV3 codec
	// DXV3 is NotchLC compatible on Windows, or use Resolume's alley tool
	args := []string{
		"-i", inputPath,
		"-c:v", "dxv",
	}

	// DXV quality settings
	switch quality {
	case "high":
		args = append(args, "-dxv_type", "hq")
	case "medium":
		args = append(args, "-dxv_type", "normal")
	case "low":
		args = append(args, "-dxv_type", "lq")
	}

	// Alpha channel support
	if alpha {
		args = append(args, "-pix_fmt", "yuva444p")
	} else {
		args = append(args, "-pix_fmt", "yuv422p")
	}

	args = append(args, "-c:a", "copy", "-y", outputPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &VideoAIResult{
		Success:    err == nil,
		InputFile:  inputPath,
		OutputFile: outputPath,
		Duration:   time.Since(start),
	}

	if err != nil {
		result.Error = fmt.Sprintf("DXV3 conversion failed: %s", stderr.String())
	} else {
		result.Message = fmt.Sprintf("Converted to DXV3 (quality: %s, alpha: %v)", quality, alpha)
	}

	return result, nil
}

// QueueJob adds a job to the processing queue
func (c *VideoAIClient) QueueJob(ctx context.Context, inputPath, operation string, params map[string]interface{}) (*VideoAIJob, error) {
	args := []string{"queue", "add", inputPath, "--operation", operation}

	// Add parameters
	for k, v := range params {
		args = append(args, fmt.Sprintf("--%s", k), fmt.Sprintf("%v", v))
	}

	output, err := c.runVidtool(ctx, args...)
	if err != nil {
		return nil, err
	}

	var job VideoAIJob
	if err := json.Unmarshal([]byte(output), &job); err != nil {
		// Fallback: create job manually
		job = VideoAIJob{
			ID:        fmt.Sprintf("job_%d", time.Now().UnixNano()),
			Input:     inputPath,
			Operation: operation,
			Status:    "queued",
		}
	}

	return &job, nil
}

// GetQueue returns all jobs in the queue
func (c *VideoAIClient) GetQueue(ctx context.Context) ([]*VideoAIJob, error) {
	output, err := c.runVidtool(ctx, "queue", "list", "--json")
	if err != nil {
		// Return empty queue if vidtool not available
		return []*VideoAIJob{}, nil
	}

	var jobs []*VideoAIJob
	if err := json.Unmarshal([]byte(output), &jobs); err != nil {
		return []*VideoAIJob{}, nil
	}

	return jobs, nil
}

// GetJobStatus returns status of a specific job
func (c *VideoAIClient) GetJobStatus(ctx context.Context, jobID string) (*VideoAIJob, error) {
	output, err := c.runVidtool(ctx, "queue", "status", jobID, "--json")
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	var job VideoAIJob
	if err := json.Unmarshal([]byte(output), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

// CancelJob cancels a queued or running job
func (c *VideoAIClient) CancelJob(ctx context.Context, jobID string) error {
	_, err := c.runVidtool(ctx, "queue", "cancel", jobID)
	return err
}

// ClearQueue removes all completed/failed jobs from queue
func (c *VideoAIClient) ClearQueue(ctx context.Context) error {
	_, err := c.runVidtool(ctx, "queue", "clear")
	return err
}
