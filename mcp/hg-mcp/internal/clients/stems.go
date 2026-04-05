// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// StemsClient provides stem separation using Demucs
type StemsClient struct {
	demucsPath string
	outputDir  string
	model      string
	useGPU     bool
	mu         sync.Mutex
	jobs       map[string]*StemJob
}

// StemJob represents a stem separation job
type StemJob struct {
	ID        string    `json:"id"`
	FilePath  string    `json:"file_path"`
	Status    string    `json:"status"` // pending, processing, completed, failed
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	OutputDir string    `json:"output_dir,omitempty"`
	Stems     []string  `json:"stems,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// StemResult represents the result of stem separation
type StemResult struct {
	FilePath  string            `json:"file_path"`
	OutputDir string            `json:"output_dir"`
	Stems     map[string]string `json:"stems"` // stem name -> file path
	Duration  float64           `json:"duration_seconds"`
	Model     string            `json:"model"`
}

// StemsStatus represents service status
type StemsStatus struct {
	Installed    bool   `json:"installed"`
	DemucsPath   string `json:"demucs_path,omitempty"`
	Model        string `json:"model"`
	GPUAvailable bool   `json:"gpu_available"`
	GPUEnabled   bool   `json:"gpu_enabled"`
	OutputDir    string `json:"output_dir"`
	ActiveJobs   int    `json:"active_jobs"`
	QueuedJobs   int    `json:"queued_jobs"`
}

// StemsHealth represents health status
type StemsHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	DemucsInstalled bool     `json:"demucs_installed"`
	GPUAvailable    bool     `json:"gpu_available"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// AvailableStem represents an available stem file
type AvailableStem struct {
	Name     string `json:"name"`
	FilePath string `json:"file_path"`
	Size     int64  `json:"size_bytes"`
}

// NewStemsClient creates a new stems client
func NewStemsClient() (*StemsClient, error) {
	// Find demucs binary
	demucsPath := os.Getenv("DEMUCS_PATH")
	if demucsPath == "" {
		// Try to find in PATH
		if path, err := exec.LookPath("demucs"); err == nil {
			demucsPath = path
		}
	}

	// Output directory
	outputDir := os.Getenv("STEMS_OUTPUT_DIR")
	if outputDir == "" {
		home, _ := os.UserHomeDir()
		outputDir = filepath.Join(home, "Music", "Stems")
	}

	// Model selection
	model := os.Getenv("DEMUCS_MODEL")
	if model == "" {
		model = "htdemucs" // Default model
	}

	// GPU usage
	useGPU := os.Getenv("DEMUCS_USE_GPU") != "false"

	return &StemsClient{
		demucsPath: demucsPath,
		outputDir:  outputDir,
		model:      model,
		useGPU:     useGPU,
		jobs:       make(map[string]*StemJob),
	}, nil
}

// GetStatus returns the service status
func (c *StemsClient) GetStatus(ctx context.Context) (*StemsStatus, error) {
	status := &StemsStatus{
		Installed:  c.demucsPath != "",
		DemucsPath: c.demucsPath,
		Model:      c.model,
		GPUEnabled: c.useGPU,
		OutputDir:  c.outputDir,
	}

	// Check GPU availability
	status.GPUAvailable = c.checkGPUAvailable(ctx)

	// Count jobs
	c.mu.Lock()
	for _, job := range c.jobs {
		switch job.Status {
		case "processing":
			status.ActiveJobs++
		case "pending":
			status.QueuedJobs++
		}
	}
	c.mu.Unlock()

	return status, nil
}

// checkGPUAvailable checks if GPU is available for processing
func (c *StemsClient) checkGPUAvailable(ctx context.Context) bool {
	// Check for CUDA availability via Python
	cmd := exec.CommandContext(ctx, "python3", "-c", "import torch; print(torch.cuda.is_available())")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "True"
}

// SeparateStems separates a track into stems
func (c *StemsClient) SeparateStems(ctx context.Context, filePath string) (*StemResult, error) {
	if c.demucsPath == "" {
		return nil, fmt.Errorf("demucs not installed - install with: pip install demucs")
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// Create output directory
	if err := os.MkdirAll(c.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	startTime := time.Now()

	// Build demucs command
	args := []string{
		"-n", c.model,
		"-o", c.outputDir,
	}

	if !c.useGPU {
		args = append(args, "-d", "cpu")
	}

	args = append(args, filePath)

	cmd := exec.CommandContext(ctx, c.demucsPath, args...)
	cmd.Dir = c.outputDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("demucs failed: %w - output: %s", err, string(output))
	}

	duration := time.Since(startTime).Seconds()

	// Find output stems
	baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	stemDir := filepath.Join(c.outputDir, c.model, baseName)

	stems := make(map[string]string)
	stemNames := []string{"vocals", "drums", "bass", "other"}

	for _, stemName := range stemNames {
		stemPath := filepath.Join(stemDir, stemName+".wav")
		if _, err := os.Stat(stemPath); err == nil {
			stems[stemName] = stemPath
		}
	}

	return &StemResult{
		FilePath:  filePath,
		OutputDir: stemDir,
		Stems:     stems,
		Duration:  duration,
		Model:     c.model,
	}, nil
}

// QueueSeparation queues a track for stem separation
func (c *StemsClient) QueueSeparation(ctx context.Context, filePath string) (*StemJob, error) {
	if c.demucsPath == "" {
		return nil, fmt.Errorf("demucs not installed")
	}

	// Create job ID
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())

	job := &StemJob{
		ID:       jobID,
		FilePath: filePath,
		Status:   "pending",
	}

	c.mu.Lock()
	c.jobs[jobID] = job
	c.mu.Unlock()

	// Start processing in background
	go c.processJob(context.Background(), job)

	return job, nil
}

// processJob processes a stem separation job
func (c *StemsClient) processJob(ctx context.Context, job *StemJob) {
	c.mu.Lock()
	job.Status = "processing"
	job.StartTime = time.Now()
	c.mu.Unlock()

	result, err := c.SeparateStems(ctx, job.FilePath)

	c.mu.Lock()
	defer c.mu.Unlock()

	job.EndTime = time.Now()

	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
	} else {
		job.Status = "completed"
		job.OutputDir = result.OutputDir
		for name := range result.Stems {
			job.Stems = append(job.Stems, name)
		}
	}
}

// GetJob returns a job by ID
func (c *StemsClient) GetJob(ctx context.Context, jobID string) (*StemJob, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	job, ok := c.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs returns all jobs
func (c *StemsClient) ListJobs(ctx context.Context) ([]*StemJob, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var jobs []*StemJob
	for _, job := range c.jobs {
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// ListStems lists available stems for a track
func (c *StemsClient) ListStems(ctx context.Context, trackName string) ([]AvailableStem, error) {
	// Look in output directory for stems
	stemDir := filepath.Join(c.outputDir, c.model, trackName)

	if _, err := os.Stat(stemDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no stems found for: %s", trackName)
	}

	var stems []AvailableStem
	stemNames := []string{"vocals", "drums", "bass", "other"}

	for _, name := range stemNames {
		stemPath := filepath.Join(stemDir, name+".wav")
		info, err := os.Stat(stemPath)
		if err == nil {
			stems = append(stems, AvailableStem{
				Name:     name,
				FilePath: stemPath,
				Size:     info.Size(),
			})
		}
	}

	return stems, nil
}

// ListAllStems lists all available stem directories
func (c *StemsClient) ListAllStems(ctx context.Context) (map[string][]AvailableStem, error) {
	modelDir := filepath.Join(c.outputDir, c.model)

	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		return nil, nil // No stems yet
	}

	result := make(map[string][]AvailableStem)

	entries, err := os.ReadDir(modelDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read stems directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		trackName := entry.Name()
		stems, err := c.ListStems(ctx, trackName)
		if err == nil && len(stems) > 0 {
			result[trackName] = stems
		}
	}

	return result, nil
}

// GetHealth returns health status
func (c *StemsClient) GetHealth(ctx context.Context) (*StemsHealth, error) {
	health := &StemsHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check demucs installation
	if c.demucsPath == "" {
		health.DemucsInstalled = false
		health.Score -= 50
		health.Issues = append(health.Issues, "demucs not found")
		health.Recommendations = append(health.Recommendations,
			"Install demucs: pip install demucs")
	} else {
		health.DemucsInstalled = true

		// Test demucs
		cmd := exec.CommandContext(ctx, c.demucsPath, "--help")
		if err := cmd.Run(); err != nil {
			health.Score -= 25
			health.Issues = append(health.Issues, "demucs exists but failed to run")
		}
	}

	// Check GPU
	health.GPUAvailable = c.checkGPUAvailable(ctx)
	if !health.GPUAvailable && c.useGPU {
		health.Score -= 10
		health.Issues = append(health.Issues, "GPU requested but not available - using CPU")
		health.Recommendations = append(health.Recommendations,
			"Install PyTorch with CUDA for faster processing")
	}

	// Check output directory
	if err := os.MkdirAll(c.outputDir, 0755); err != nil {
		health.Score -= 20
		health.Issues = append(health.Issues, "Cannot create output directory")
	}

	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// GetModels returns available demucs models
func (c *StemsClient) GetModels(ctx context.Context) ([]string, error) {
	// Standard demucs models
	return []string{
		"htdemucs",    // Default hybrid transformer model
		"htdemucs_ft", // Fine-tuned version
		"htdemucs_6s", // 6 stems (piano, guitar included)
		"hdemucs_mmi", // Hybrid demucs
		"mdx",         // MDX-Net
		"mdx_extra",   // MDX-Net extra
		"mdx_q",       // MDX-Net quantized
		"mdx_extra_q", // MDX-Net extra quantized
	}, nil
}

// SetModel sets the model to use
func (c *StemsClient) SetModel(model string) {
	c.model = model
}

// GetOutputDir returns the output directory
func (c *StemsClient) GetOutputDir() string {
	return c.outputDir
}

// SetOutputDir sets the output directory
func (c *StemsClient) SetOutputDir(dir string) {
	c.outputDir = dir
}

// BatchSeparate separates multiple tracks
func (c *StemsClient) BatchSeparate(ctx context.Context, filePaths []string) ([]*StemJob, error) {
	var jobs []*StemJob

	for _, filePath := range filePaths {
		job, err := c.QueueSeparation(ctx, filePath)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

// CleanupOldJobs removes completed jobs older than the specified duration
func (c *StemsClient) CleanupOldJobs(ctx context.Context, olderThan time.Duration) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	removed := 0

	for id, job := range c.jobs {
		if job.Status == "completed" || job.Status == "failed" {
			if job.EndTime.Before(cutoff) {
				delete(c.jobs, id)
				removed++
			}
		}
	}

	return removed
}

// ExportStemInfo exports stem information as JSON
func (c *StemsClient) ExportStemInfo(ctx context.Context, trackName string) (string, error) {
	stems, err := c.ListStems(ctx, trackName)
	if err != nil {
		return "", err
	}

	info := map[string]interface{}{
		"track":      trackName,
		"model":      c.model,
		"output_dir": filepath.Join(c.outputDir, c.model, trackName),
		"stems":      stems,
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
