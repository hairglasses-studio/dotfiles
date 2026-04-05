package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// PipelinePhase represents a phase in the pipeline
type PipelinePhase string

const (
	PhaseDiscover PipelinePhase = "discover"
	PhaseDownload PipelinePhase = "download"
	PhaseUpload   PipelinePhase = "upload"
	PhaseImport   PipelinePhase = "import"
)

// PipelineState holds the state passed between pipeline steps
type PipelineState struct {
	// Input parameters
	Username      string `json:"username"`
	Format        string `json:"format"`
	DryRun        bool   `json:"dry_run"`
	IncludeLikes  bool   `json:"include_likes"`
	SkipGDrive    bool   `json:"skip_gdrive"`
	SkipRekordbox bool   `json:"skip_rekordbox"`

	// Paths
	LocalRoot       string `json:"local_root"`
	GDriveMountPath string `json:"gdrive_mount_path"`
	GDriveBasePath  string `json:"gdrive_base_path"`

	// Discovery results
	Playlists   []string `json:"playlists"`
	TotalTracks int      `json:"total_tracks"`

	// Download results
	DownloadedTracks int      `json:"downloaded_tracks"`
	SkippedTracks    int      `json:"skipped_tracks"`
	DownloadErrors   []string `json:"download_errors,omitempty"`

	// GDrive results
	GDriveResults []GDriveSyncResult `json:"gdrive_results,omitempty"`

	// Rekordbox results
	ImportedTracks   int      `json:"imported_tracks"`
	PlaylistsCreated []string `json:"playlists_created,omitempty"`
	ImportErrors     []string `json:"import_errors,omitempty"`
}

// PipelineStep defines a step in the sync pipeline
type PipelineStep interface {
	Name() string
	Phase() PipelinePhase
	Execute(ctx context.Context, state *PipelineState) error
}

// PipelineProgress reports progress during pipeline execution
type PipelineProgress struct {
	Phase        PipelinePhase `json:"phase"`
	PhaseNum     int           `json:"phase_num"`
	TotalPhases  int           `json:"total_phases"`
	CurrentItem  string        `json:"current_item"`
	ItemsCurrent int           `json:"items_current"`
	ItemsTotal   int           `json:"items_total"`
	OverallPct   float64       `json:"overall_pct"`
	Message      string        `json:"message"`
}

// PipelineResult is the final result of pipeline execution
type PipelineResult struct {
	Status    string         `json:"status"` // "completed", "partial", "failed"
	StartTime time.Time      `json:"start_time"`
	EndTime   time.Time      `json:"end_time"`
	Duration  string         `json:"duration"`
	State     *PipelineState `json:"state"`
	Phases    map[string]any `json:"phases"`
	Summary   string         `json:"summary"`
	Errors    []string       `json:"errors,omitempty"`
}

// Pipeline orchestrates multiple steps
type Pipeline struct {
	name     string
	steps    []PipelineStep
	state    *PipelineState
	progress chan PipelineProgress
}

// NewPipeline creates a new pipeline
func NewPipeline(name string, state *PipelineState) *Pipeline {
	return &Pipeline{
		name:  name,
		state: state,
	}
}

// AddStep adds a step to the pipeline
func (p *Pipeline) AddStep(step PipelineStep) {
	p.steps = append(p.steps, step)
}

// Execute runs all pipeline steps
func (p *Pipeline) Execute(ctx context.Context) *PipelineResult {
	result := &PipelineResult{
		StartTime: time.Now(),
		State:     p.state,
		Phases:    make(map[string]any),
	}

	totalSteps := len(p.steps)
	var allErrors []string

	for i, step := range p.steps {
		select {
		case <-ctx.Done():
			result.Status = "cancelled"
			result.Errors = append(allErrors, "pipeline cancelled")
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime).Round(time.Second).String()
			return result
		default:
		}

		log.Printf("[%s] Step %d/%d: %s", p.name, i+1, totalSteps, step.Name())

		if err := step.Execute(ctx, p.state); err != nil {
			log.Printf("[%s] Step %s failed: %v", p.name, step.Name(), err)
			allErrors = append(allErrors, fmt.Sprintf("%s: %v", step.Name(), err))

			// Continue on non-critical errors for some phases
			if step.Phase() == PhaseDiscover {
				// Discovery failure is fatal
				result.Status = "failed"
				result.Errors = allErrors
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime).Round(time.Second).String()
				return result
			}
			// Other phases can continue with partial results
		}
	}

	// Build phase summaries
	result.Phases["discover"] = map[string]any{
		"playlists": len(p.state.Playlists),
		"tracks":    p.state.TotalTracks,
	}

	result.Phases["download"] = map[string]any{
		"downloaded": p.state.DownloadedTracks,
		"skipped":    p.state.SkippedTracks,
		"errors":     len(p.state.DownloadErrors),
	}

	if !p.state.SkipGDrive {
		totalUploaded := 0
		for _, r := range p.state.GDriveResults {
			totalUploaded += r.FilesUploaded
		}
		result.Phases["gdrive"] = map[string]any{
			"uploaded":  totalUploaded,
			"playlists": len(p.state.GDriveResults),
		}
	}

	if !p.state.SkipRekordbox {
		result.Phases["rekordbox"] = map[string]any{
			"imported":          p.state.ImportedTracks,
			"playlists_created": p.state.PlaylistsCreated,
		}
	}

	// Set final status
	if len(allErrors) > 0 {
		result.Status = "partial"
		result.Errors = allErrors
	} else {
		result.Status = "completed"
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).Round(time.Second).String()
	result.Summary = p.buildSummary()

	return result
}

func (p *Pipeline) buildSummary() string {
	if p.state.DownloadedTracks == 0 {
		return fmt.Sprintf("No new tracks to sync from %d playlists", len(p.state.Playlists))
	}

	parts := []string{
		fmt.Sprintf("Synced %d new tracks from %d playlists", p.state.DownloadedTracks, len(p.state.Playlists)),
	}

	if !p.state.SkipGDrive && len(p.state.GDriveResults) > 0 {
		totalUploaded := 0
		for _, r := range p.state.GDriveResults {
			totalUploaded += r.FilesUploaded
		}
		if totalUploaded > 0 {
			parts = append(parts, fmt.Sprintf("uploaded %d to Google Drive", totalUploaded))
		}
	}

	if !p.state.SkipRekordbox && p.state.ImportedTracks > 0 {
		parts = append(parts, fmt.Sprintf("imported %d to Rekordbox", p.state.ImportedTracks))
	}

	return fmt.Sprintf("%s", parts[0])
}

// JSON returns the pipeline result as JSON
func (r *PipelineResult) JSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}
