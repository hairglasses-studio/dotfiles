package chains

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Executor manages chain execution
type Executor struct {
	mu         sync.RWMutex
	chains     map[string]*Chain          // Chain definitions by ID
	executions map[string]*ChainExecution // Active executions by ID
	pending    map[string]*PendingGate    // Pending gate approvals
	dataDir    string                     // Directory for persisting chains
	registry   *tools.ToolRegistry
}

// NewExecutor creates a new chain executor
func NewExecutor(dataDir string, registry *tools.ToolRegistry) *Executor {
	e := &Executor{
		chains:     make(map[string]*Chain),
		executions: make(map[string]*ChainExecution),
		pending:    make(map[string]*PendingGate),
		dataDir:    dataDir,
		registry:   registry,
	}
	e.loadBuiltinChains()
	e.loadChainsFromDisk()
	return e
}

// loadBuiltinChains loads the default chain definitions
func (e *Executor) loadBuiltinChains() {
	builtins := []Chain{
		{
			ID:          "show_startup",
			Name:        "Show Startup",
			Description: "Power on systems, check health, load project, test outputs",
			Category:    "show",
			Steps: []ChainStep{
				{ID: "1", Name: "Check Studio Health", Type: StepTypeTool, Tool: "hairglasses_studio_health", OnError: OnErrorStop},
				{ID: "2", Name: "Check AV Systems", Type: StepTypeTool, Tool: "aftrs_av", Inputs: map[string]interface{}{"software": "resolume", "action": "status"}, OnError: OnErrorContinue},
				{ID: "3", Name: "Check Lighting", Type: StepTypeTool, Tool: "aftrs_lighting", Inputs: map[string]interface{}{"system": "grandma3", "action": "status"}, OnError: OnErrorContinue},
				{ID: "4", Name: "Check Audio", Type: StepTypeTool, Tool: "aftrs_audio", Inputs: map[string]interface{}{"system": "ableton", "action": "status"}, OnError: OnErrorContinue},
				{ID: "5", Name: "Approve Startup", Type: StepTypeGate, GateMessage: "All systems checked. Approve to continue with show startup?", OnError: OnErrorStop},
			},
			Triggers: []ChainTrigger{{Type: TriggerManual, Enabled: true}},
			Timeout:  10 * time.Minute,
		},
		{
			ID:          "stream_start",
			Name:        "Start Stream",
			Description: "Configure OBS, go live on platforms, notify Discord",
			Category:    "stream",
			Parameters: []ChainParameter{
				{Name: "title", Type: "string", Description: "Stream title", Required: true},
				{Name: "platform", Type: "string", Description: "Platform: twitch, youtube, both", Required: false, Default: "twitch"},
			},
			Steps: []ChainStep{
				{ID: "1", Name: "Check OBS Status", Type: StepTypeTool, Tool: "aftrs_av", Inputs: map[string]interface{}{"software": "obs", "action": "status"}, OnError: OnErrorStop},
				{ID: "2", Name: "Set Stream Title", Type: StepTypeTool, Tool: "aftrs_streaming", Inputs: map[string]interface{}{"platform": "{{platform}}", "action": "title", "value": "{{title}}"}, OnError: OnErrorContinue},
				{ID: "3", Name: "Approve Go Live", Type: StepTypeGate, GateMessage: "Ready to go live. Approve to start streaming?", OnError: OnErrorStop},
				{ID: "4", Name: "Go Live", Type: StepTypeTool, Tool: "aftrs_streaming", Inputs: map[string]interface{}{"platform": "{{platform}}", "action": "go_live"}, OnError: OnErrorStop},
				{ID: "5", Name: "Notify Discord", Type: StepTypeTool, Tool: "aftrs_discord_send", Inputs: map[string]interface{}{"channel": "announcements", "message": "We're live! {{title}}"}, OnError: OnErrorContinue},
			},
			Triggers: []ChainTrigger{{Type: TriggerManual, Enabled: true}},
			Timeout:  5 * time.Minute,
		},
		{
			ID:          "stream_end",
			Name:        "End Stream",
			Description: "End stream, post stats, thank viewers",
			Category:    "stream",
			Steps: []ChainStep{
				{ID: "1", Name: "Get Viewer Stats", Type: StepTypeTool, Tool: "aftrs_streaming", Inputs: map[string]interface{}{"platform": "twitch", "action": "viewers"}, OnError: OnErrorContinue},
				{ID: "2", Name: "End Stream", Type: StepTypeTool, Tool: "aftrs_streaming", Inputs: map[string]interface{}{"platform": "twitch", "action": "end_stream"}, OnError: OnErrorStop},
				{ID: "3", Name: "Post Stats to Discord", Type: StepTypeTool, Tool: "aftrs_discord_send", Inputs: map[string]interface{}{"channel": "stream-stats", "message": "Stream ended. Thanks for watching!"}, OnError: OnErrorContinue},
			},
			Triggers: []ChainTrigger{{Type: TriggerManual, Enabled: true}},
			Timeout:  5 * time.Minute,
		},
		{
			ID:          "backup_projects",
			Name:        "Backup Projects",
			Description: "Backup Ableton, Resolume, and TouchDesigner projects",
			Category:    "backup",
			Steps: []ChainStep{
				{ID: "1", Name: "Stop Ableton", Type: StepTypeTool, Tool: "aftrs_audio", Inputs: map[string]interface{}{"system": "ableton", "action": "stop"}, OnError: OnErrorContinue},
				{ID: "2", Name: "Backup Ableton Projects", Type: StepTypeTool, Tool: "aftrs_backup_create", Inputs: map[string]interface{}{"source": "ableton"}, OnError: OnErrorContinue},
				{ID: "3", Name: "Backup Resolume Projects", Type: StepTypeTool, Tool: "aftrs_backup_create", Inputs: map[string]interface{}{"source": "resolume"}, OnError: OnErrorContinue},
				{ID: "4", Name: "Backup TouchDesigner Projects", Type: StepTypeTool, Tool: "aftrs_backup_create", Inputs: map[string]interface{}{"source": "touchdesigner"}, OnError: OnErrorContinue},
				{ID: "5", Name: "Verify Backups", Type: StepTypeTool, Tool: "aftrs_backup_verify", OnError: OnErrorStop},
			},
			Triggers: []ChainTrigger{
				{Type: TriggerManual, Enabled: true},
				{Type: TriggerCron, Schedule: "0 3 * * *", Enabled: true}, // 3 AM daily
			},
			Timeout: 30 * time.Minute,
		},
		{
			ID:          "dj_set_prep",
			Name:        "DJ Set Preparation",
			Description: "Sync libraries, check audio routing, prepare setlist",
			Category:    "dj",
			Parameters: []ChainParameter{
				{Name: "software", Type: "string", Description: "DJ software: serato, rekordbox, traktor", Required: false, Default: "rekordbox"},
			},
			Steps: []ChainStep{
				{ID: "1", Name: "Check DJ Software", Type: StepTypeTool, Tool: "aftrs_dj", Inputs: map[string]interface{}{"software": "{{software}}", "action": "status"}, OnError: OnErrorStop},
				{ID: "2", Name: "Check Dante Routing", Type: StepTypeTool, Tool: "aftrs_audio", Inputs: map[string]interface{}{"system": "dante", "action": "routing"}, OnError: OnErrorContinue},
				{ID: "3", Name: "Load Recent Playlists", Type: StepTypeTool, Tool: "aftrs_dj", Inputs: map[string]interface{}{"software": "{{software}}", "action": "playlists"}, OnError: OnErrorContinue},
			},
			Triggers: []ChainTrigger{{Type: TriggerManual, Enabled: true}},
			Timeout:  5 * time.Minute,
		},
		{
			ID:          "lighting_test",
			Name:        "Lighting System Test",
			Description: "Test all lighting systems - WLED, DMX, grandMA3",
			Category:    "lighting",
			Steps: []ChainStep{
				{ID: "1", Name: "WLED Status", Type: StepTypeTool, Tool: "aftrs_lighting", Inputs: map[string]interface{}{"system": "wled", "action": "status"}, OnError: OnErrorContinue},
				{ID: "2", Name: "DMX Status", Type: StepTypeTool, Tool: "aftrs_lighting", Inputs: map[string]interface{}{"system": "dmx", "action": "status"}, OnError: OnErrorContinue},
				{ID: "3", Name: "grandMA3 Status", Type: StepTypeTool, Tool: "aftrs_lighting", Inputs: map[string]interface{}{"system": "grandma3", "action": "status"}, OnError: OnErrorContinue},
				{ID: "4", Name: "Test Pattern", Type: StepTypeGate, GateMessage: "Run test pattern on all fixtures?", OnError: OnErrorStop},
				{ID: "5", Name: "WLED Test Color", Type: StepTypeTool, Tool: "aftrs_lighting", Inputs: map[string]interface{}{"system": "wled", "action": "color", "value": "255,0,0"}, DelayAfter: 2 * time.Second, OnError: OnErrorContinue},
				{ID: "6", Name: "Blackout", Type: StepTypeTool, Tool: "aftrs_lighting", Inputs: map[string]interface{}{"system": "dmx", "action": "blackout"}, OnError: OnErrorContinue},
			},
			Triggers: []ChainTrigger{{Type: TriggerManual, Enabled: true}},
			Timeout:  5 * time.Minute,
		},
	}

	// Video AI processing pipeline using branch steps
	builtins = append(builtins, Chain{
		ID:          "video_ai_pipeline",
		Name:        "Video AI Pipeline",
		Description: "Detect source format, analyze, process (upscale/denoise), convert DXV3, ingest to Resolume",
		Category:    "video",
		Parameters: []ChainParameter{
			{Name: "input_path", Type: "string", Description: "Path to input video file", Required: true},
			{Name: "output_dir", Type: "string", Description: "Output directory", Required: false, Default: ""},
		},
		Steps: []ChainStep{
			{ID: "1", Name: "Analyze Source", Type: StepTypeTool, Tool: "aftrs_video_info", Inputs: map[string]interface{}{"video_path": "{{input_path}}"}, OnError: OnErrorStop},
			{ID: "2", Name: "Denoise", Type: StepTypeTool, Tool: "aftrs_video_denoise", Inputs: map[string]interface{}{"input": "{{input_path}}"}, OnError: OnErrorContinue},
			{ID: "3", Name: "Format Branch", Type: StepTypeBranch, OnError: OnErrorContinue,
				Branches: []BranchCase{
					{Condition: "steps.1.text == 4k", Steps: []ChainStep{
						{ID: "3a", Name: "Skip Upscale (4K)", Type: StepTypeDelay, Inputs: map[string]interface{}{"duration": time.Millisecond}, OnError: OnErrorContinue},
					}},
					{Condition: "true", Steps: []ChainStep{
						{ID: "3b", Name: "Upscale to 4K", Type: StepTypeTool, Tool: "aftrs_video_upscale", Inputs: map[string]interface{}{"input": "{{input_path}}", "scale": 2}, OnError: OnErrorContinue},
					}},
				},
			},
			{ID: "4", Name: "Convert DXV3", Type: StepTypeTool, Tool: "aftrs_video_to_dxv3", Inputs: map[string]interface{}{"input": "{{input_path}}"}, OnError: OnErrorStop},
			{ID: "5", Name: "Approve Ingest", Type: StepTypeGate, GateMessage: "Video processed. Ingest to Resolume?", GateTimeout: 5 * time.Minute, GateDefaultAction: "reject", OnError: OnErrorStop},
			{ID: "6", Name: "Load to Resolume", Type: StepTypeTool, Tool: "aftrs_resolume_clip_load", Inputs: map[string]interface{}{"layer": 1, "column": 1, "file": "{{input_path}}"}, OnError: OnErrorContinue},
		},
		Triggers: []ChainTrigger{{Type: TriggerManual, Enabled: true}},
		Timeout:  30 * time.Minute,
	})

	// Pre-record production chain
	builtins = append(builtins, Chain{
		ID:          "pre_record_production",
		Name:        "Pre-Record Production",
		Description: "Health check, sync DJ libraries, arm cameras, set lighting, gate approval, start recording",
		Category:    "show",
		Steps: []ChainStep{
			{ID: "1", Name: "System Health Checks", Type: StepTypeParallel, OnError: OnErrorContinue,
				ParallelSteps: []ChainStep{
					{ID: "1a", Name: "Check Ableton", Type: StepTypeTool, Tool: "aftrs_ableton_health", OnError: OnErrorContinue},
					{ID: "1b", Name: "Check Resolume", Type: StepTypeTool, Tool: "aftrs_resolume_health", OnError: OnErrorContinue},
					{ID: "1c", Name: "Check OBS", Type: StepTypeTool, Tool: "aftrs_obs_health", OnError: OnErrorContinue},
					{ID: "1d", Name: "Check Lighting", Type: StepTypeTool, Tool: "aftrs_lighting_health", OnError: OnErrorContinue},
				},
			},
			{ID: "2", Name: "Sync DJ Library", Type: StepTypeTool, Tool: "aftrs_rekordbox_s3_sync", Inputs: map[string]interface{}{"direction": "down", "dry_run": false}, OnError: OnErrorContinue},
			{ID: "3", Name: "Set Lighting Scene", Type: StepTypeTool, Tool: "aftrs_scene_recall", Inputs: map[string]interface{}{"scene": "studio_record"}, OnError: OnErrorContinue},
			{ID: "4", Name: "Approve Recording", Type: StepTypeGate, GateMessage: "All systems checked. Start recording?", GateTimeout: 10 * time.Minute, GateDefaultAction: "reject", OnError: OnErrorStop},
			{ID: "5", Name: "Start OBS Recording", Type: StepTypeTool, Tool: "aftrs_obs_record", Inputs: map[string]interface{}{"action": "start"}, OnError: OnErrorStop},
			{ID: "6", Name: "Notify Discord", Type: StepTypeTool, Tool: "aftrs_discord_send", Inputs: map[string]interface{}{"message": "Recording started"}, OnError: OnErrorContinue},
		},
		Triggers: []ChainTrigger{{Type: TriggerManual, Enabled: true}},
		Timeout:  20 * time.Minute,
	})

	// Automated mix archive chain
	builtins = append(builtins, Chain{
		ID:          "automated_mix_archive",
		Name:        "Automated Mix Archive",
		Description: "Stop recording, normalize, tag metadata, stem separation, upload S3, notify Discord",
		Category:    "archive",
		Parameters: []ChainParameter{
			{Name: "recording_path", Type: "string", Description: "Path to the recording file", Required: true},
			{Name: "artist", Type: "string", Description: "Artist name for tagging", Required: false, Default: "DJ Set"},
			{Name: "title", Type: "string", Description: "Mix title", Required: false, Default: "Live Mix"},
		},
		Steps: []ChainStep{
			{ID: "1", Name: "Stop Recording", Type: StepTypeTool, Tool: "aftrs_obs_record", Inputs: map[string]interface{}{"action": "stop"}, OnError: OnErrorContinue},
			{ID: "2", Name: "Wait for File", Type: StepTypeDelay, Inputs: map[string]interface{}{"duration": 3 * time.Second}, OnError: OnErrorContinue},
			{ID: "3", Name: "Normalize Audio", Type: StepTypeTool, Tool: "aftrs_samples_process", Inputs: map[string]interface{}{"audio_path": "{{recording_path}}", "normalize": true, "target_lufs": -14}, OnError: OnErrorContinue},
			{ID: "4", Name: "Tag Metadata", Type: StepTypeTool, Tool: "aftrs_samples_tag", Inputs: map[string]interface{}{"audio_path": "{{recording_path}}", "artist": "{{artist}}", "title": "{{title}}", "genre": "DJ Mix"}, OnError: OnErrorContinue},
			{ID: "5", Name: "Queue Stem Separation", Type: StepTypeTool, Tool: "aftrs_stems_queue", Inputs: map[string]interface{}{"file_path": "{{recording_path}}"}, OnError: OnErrorContinue},
			{ID: "6", Name: "Generate Rekordbox XML", Type: StepTypeTool, Tool: "aftrs_samples_rekordbox_xml", Inputs: map[string]interface{}{"directory": "{{recording_path}}"}, OnError: OnErrorContinue},
			{ID: "7", Name: "Notify Discord", Type: StepTypeTool, Tool: "aftrs_discord_send", Inputs: map[string]interface{}{"message": "Mix archived: {{title}} by {{artist}}"}, OnError: OnErrorContinue},
		},
		Triggers: []ChainTrigger{{Type: TriggerManual, Enabled: true}},
		Timeout:  30 * time.Minute,
	})

	for i := range builtins {
		builtins[i].CreatedAt = time.Now()
		builtins[i].UpdatedAt = time.Now()
		e.chains[builtins[i].ID] = &builtins[i]
	}
}

// loadChainsFromDisk loads custom chain definitions from disk
func (e *Executor) loadChainsFromDisk() {
	if e.dataDir == "" {
		return
	}

	chainsDir := filepath.Join(e.dataDir, "chains")
	if _, err := os.Stat(chainsDir); os.IsNotExist(err) {
		return
	}

	files, err := os.ReadDir(chainsDir)
	if err != nil {
		return
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(chainsDir, f.Name()))
		if err != nil {
			continue
		}

		var chain Chain
		if err := json.Unmarshal(data, &chain); err != nil {
			continue
		}

		e.chains[chain.ID] = &chain
	}
}

// SaveChain saves a chain definition to disk
func (e *Executor) SaveChain(chain *Chain) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	chain.UpdatedAt = time.Now()
	if chain.CreatedAt.IsZero() {
		chain.CreatedAt = time.Now()
	}

	e.chains[chain.ID] = chain

	if e.dataDir == "" {
		return nil
	}

	chainsDir := filepath.Join(e.dataDir, "chains")
	if err := os.MkdirAll(chainsDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(chain, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(chainsDir, chain.ID+".json"), data, 0644)
}

// GetChain returns a chain by ID
func (e *Executor) GetChain(id string) (*Chain, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	chain, ok := e.chains[id]
	return chain, ok
}

// ListChains returns all chain definitions
func (e *Executor) ListChains(category string) []*Chain {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*Chain
	for _, chain := range e.chains {
		if category == "" || chain.Category == category {
			result = append(result, chain)
		}
	}
	return result
}

// Execute starts a chain execution. If dryRun is true, tool steps log
// their parameters but skip actual execution, and gate steps auto-approve.
func (e *Executor) Execute(ctx context.Context, chainID string, params map[string]interface{}, triggeredBy string, dryRun ...bool) (*ChainExecution, error) {
	chain, ok := e.GetChain(chainID)
	if !ok {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}

	// Validate required parameters
	for _, p := range chain.Parameters {
		if p.Required {
			if _, ok := params[p.Name]; !ok {
				if p.Default != nil {
					params[p.Name] = p.Default
				} else {
					return nil, fmt.Errorf("missing required parameter: %s", p.Name)
				}
			}
		}
	}

	// Apply defaults for optional parameters
	for _, p := range chain.Parameters {
		if _, ok := params[p.Name]; !ok && p.Default != nil {
			params[p.Name] = p.Default
		}
	}

	isDryRun := len(dryRun) > 0 && dryRun[0]

	exec := &ChainExecution{
		ID:          uuid.New().String(),
		ChainID:     chain.ID,
		ChainName:   chain.Name,
		Status:      ChainStatusRunning,
		Parameters:  params,
		StartedAt:   time.Now(),
		CurrentStep: 0,
		TotalSteps:  len(chain.Steps),
		StepResults: make([]StepResult, 0),
		TriggeredBy: triggeredBy,
		DryRun:      isDryRun,
	}

	e.mu.Lock()
	e.executions[exec.ID] = exec
	e.mu.Unlock()

	// Execute in background
	go e.runChain(ctx, chain, exec)

	return exec, nil
}

// runChain executes a chain's steps
func (e *Executor) runChain(ctx context.Context, chain *Chain, exec *ChainExecution) {
	// Apply chain timeout
	if chain.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, chain.Timeout)
		defer cancel()
	}

	for i, step := range chain.Steps {
		select {
		case <-ctx.Done():
			e.updateExecution(exec, ChainStatusFailed, "chain timeout exceeded")
			return
		default:
		}

		exec.CurrentStep = i + 1
		result := e.executeStep(ctx, &step, exec)
		exec.StepResults = append(exec.StepResults, result)

		if result.Status == ChainStatusFailed {
			switch step.OnError {
			case OnErrorStop:
				e.updateExecution(exec, ChainStatusFailed, result.Error)
				return
			case OnErrorRetry:
				for retry := 0; retry < step.MaxRetries; retry++ {
					result = e.executeStep(ctx, &step, exec)
					result.Retries = retry + 1
					if result.Status == ChainStatusCompleted {
						break
					}
				}
				if result.Status == ChainStatusFailed {
					e.updateExecution(exec, ChainStatusFailed, result.Error)
					return
				}
			case OnErrorContinue:
				// Continue to next step
			}
		}

		if result.Status == ChainStatusPaused {
			// Gate waiting for approval
			e.updateExecution(exec, ChainStatusPaused, "")
			return
		}

		// Apply delay after step
		if step.DelayAfter > 0 {
			time.Sleep(step.DelayAfter)
		}
	}

	e.updateExecution(exec, ChainStatusCompleted, "")
}

// executeStep executes a single step
func (e *Executor) executeStep(ctx context.Context, step *ChainStep, exec *ChainExecution) StepResult {
	result := StepResult{
		StepID:    step.ID,
		StepName:  step.Name,
		Status:    ChainStatusRunning,
		StartedAt: time.Now(),
	}

	// Apply step timeout
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}

	switch step.Type {
	case StepTypeTool:
		result = e.executeToolStep(ctx, step, exec, result)
	case StepTypeGate:
		result = e.executeGateStep(ctx, step, exec, result)
	case StepTypeDelay:
		if delay, ok := step.Inputs["duration"].(time.Duration); ok {
			time.Sleep(delay)
		}
		result.Status = ChainStatusCompleted
	case StepTypeParallel:
		result = e.executeParallelSteps(ctx, step, exec, result)
	case StepTypeBranch:
		result = e.executeBranchStep(ctx, step, exec, result)
	case StepTypeChain:
		result = e.executeChainStep(ctx, step, exec, result)
	default:
		result.Status = ChainStatusFailed
		result.Error = fmt.Sprintf("unsupported step type: %s", step.Type)
	}

	now := time.Now()
	result.CompletedAt = &now
	return result
}

// executeToolStep executes a tool step
func (e *Executor) executeToolStep(ctx context.Context, step *ChainStep, exec *ChainExecution, result StepResult) StepResult {
	toolDef, ok := e.registry.GetTool(step.Tool)
	if !ok {
		result.Status = ChainStatusFailed
		result.Error = fmt.Sprintf("tool not found: %s", step.Tool)
		return result
	}

	// Substitute parameters and step outputs in inputs
	stepOutputs := e.collectStepOutputs(exec)
	inputs := e.substituteParamsWithOutputs(step.Inputs, exec.Parameters, stepOutputs)

	// Create tool request
	req := mcp.CallToolRequest{}
	req.Params.Name = step.Tool
	req.Params.Arguments = inputs

	// Dry-run: skip actual execution, log parameters
	if exec.DryRun {
		result.Status = ChainStatusCompleted
		result.Output = map[string]interface{}{
			"dry_run":    true,
			"tool":       step.Tool,
			"inputs":     inputs,
			"would_call": step.Tool,
		}
		return result
	}

	// Execute tool
	toolResult, err := toolDef.Handler(ctx, req)
	if err != nil {
		result.Status = ChainStatusFailed
		result.Error = err.Error()
		return result
	}

	if toolResult.IsError {
		result.Status = ChainStatusFailed
		if len(toolResult.Content) > 0 {
			if tc, ok := toolResult.Content[0].(mcp.TextContent); ok {
				result.Error = tc.Text
			}
		}
		return result
	}

	// Extract output
	result.Status = ChainStatusCompleted
	result.Output = make(map[string]interface{})
	if len(toolResult.Content) > 0 {
		if tc, ok := toolResult.Content[0].(mcp.TextContent); ok {
			result.Output["text"] = tc.Text
		}
	}

	return result
}

// executeGateStep creates a pending gate and pauses execution.
// If GateTimeout is set, a background goroutine auto-resolves the gate after the deadline.
func (e *Executor) executeGateStep(ctx context.Context, step *ChainStep, exec *ChainExecution, result StepResult) StepResult {
	// Dry-run: auto-approve gate steps
	if exec.DryRun {
		result.Status = ChainStatusCompleted
		result.Output = map[string]interface{}{
			"dry_run":       true,
			"gate":          step.Name,
			"auto_approved": true,
			"message":       step.GateMessage,
		}
		return result
	}

	gate := &PendingGate{
		ExecutionID: exec.ID,
		ChainName:   exec.ChainName,
		StepID:      step.ID,
		StepName:    step.Name,
		Message:     step.GateMessage,
		Approvers:   step.GateApprovers,
		CreatedAt:   time.Now(),
	}

	e.mu.Lock()
	e.pending[exec.ID] = gate
	e.mu.Unlock()

	// Start gate timeout goroutine if configured.
	// Uses context.Background() because the parent runChain context is cancelled
	// when runChain returns after setting status to paused.
	if step.GateTimeout > 0 {
		execID := exec.ID
		timeout := step.GateTimeout
		defaultAction := step.GateDefaultAction
		if defaultAction == "" {
			defaultAction = "reject"
		}
		go func() {
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			<-timer.C

			// Check if gate is still pending
			e.mu.RLock()
			_, stillPending := e.pending[execID]
			e.mu.RUnlock()

			if !stillPending {
				return // Gate was already resolved manually
			}

			approved := defaultAction == "approve"
			_ = e.ApproveGate(context.Background(), execID, approved, "system",
				fmt.Sprintf("gate timeout after %s (default: %s)", timeout, defaultAction))
		}()
	}

	result.Status = ChainStatusPaused
	return result
}

// executeParallelSteps executes multiple steps in parallel
func (e *Executor) executeParallelSteps(ctx context.Context, step *ChainStep, exec *ChainExecution, result StepResult) StepResult {
	var wg sync.WaitGroup
	results := make([]StepResult, len(step.ParallelSteps))

	for i, ps := range step.ParallelSteps {
		wg.Add(1)
		go func(idx int, s ChainStep) {
			defer wg.Done()
			results[idx] = e.executeStep(ctx, &s, exec)
		}(i, ps)
	}

	wg.Wait()

	// Check if any failed
	allSucceeded := true
	for _, r := range results {
		if r.Status == ChainStatusFailed {
			allSucceeded = false
			result.Error = r.Error
			break
		}
	}

	if allSucceeded {
		result.Status = ChainStatusCompleted
	} else {
		result.Status = ChainStatusFailed
	}

	result.Output = map[string]interface{}{"parallel_results": results}
	return result
}

// executeBranchStep evaluates branch conditions and executes the first matching branch.
// Conditions are simple expressions: "key == value" or "key != value" evaluated against
// prior step outputs. The special key format "steps.<stepID>.<outputKey>" references
// outputs from earlier steps.
func (e *Executor) executeBranchStep(ctx context.Context, step *ChainStep, exec *ChainExecution, result StepResult) StepResult {
	if len(step.Branches) == 0 {
		result.Status = ChainStatusCompleted
		result.Output = map[string]interface{}{"branch": "none", "reason": "no branches defined"}
		return result
	}

	// Build a lookup of all prior step outputs
	stepOutputs := e.collectStepOutputs(exec)

	for i, branch := range step.Branches {
		if e.evaluateCondition(branch.Condition, exec.Parameters, stepOutputs) {
			// Execute the matching branch's steps sequentially
			for _, subStep := range branch.Steps {
				select {
				case <-ctx.Done():
					result.Status = ChainStatusFailed
					result.Error = "branch execution timeout"
					return result
				default:
				}

				subResult := e.executeStep(ctx, &subStep, exec)
				exec.StepResults = append(exec.StepResults, subResult)

				if subResult.Status == ChainStatusFailed && subStep.OnError == OnErrorStop {
					result.Status = ChainStatusFailed
					result.Error = subResult.Error
					return result
				}
				if subResult.Status == ChainStatusPaused {
					result.Status = ChainStatusPaused
					return result
				}
			}

			result.Status = ChainStatusCompleted
			result.Output = map[string]interface{}{
				"branch":    fmt.Sprintf("branch_%d", i),
				"condition": branch.Condition,
			}
			return result
		}
	}

	// No branch matched — complete without executing any sub-steps
	result.Status = ChainStatusCompleted
	result.Output = map[string]interface{}{"branch": "none", "reason": "no condition matched"}
	return result
}

// collectStepOutputs builds a map of step outputs keyed by "steps.<stepID>.<outputKey>".
func (e *Executor) collectStepOutputs(exec *ChainExecution) map[string]string {
	outputs := make(map[string]string)
	for _, sr := range exec.StepResults {
		for k, v := range sr.Output {
			key := fmt.Sprintf("steps.%s.%s", sr.StepID, k)
			outputs[key] = fmt.Sprintf("%v", v)
		}
	}
	return outputs
}

// evaluateCondition checks a simple condition string against parameters and step outputs.
// Supported formats:
//   - "key == value"  — true if lookup(key) equals value
//   - "key != value"  — true if lookup(key) does not equal value
//   - "true" or ""    — always true (default/fallback branch)
//
// Keys are looked up first in stepOutputs (e.g., "steps.1.text"), then in params.
func (e *Executor) evaluateCondition(condition string, params map[string]interface{}, stepOutputs map[string]string) bool {
	condition = trimSpaces(condition)
	if condition == "" || condition == "true" {
		return true
	}

	// Try "!=" first (before "==" since "==" is a substring)
	if key, value, ok := splitCondition(condition, "!="); ok {
		actual := lookupValue(key, params, stepOutputs)
		return actual != value
	}
	if key, value, ok := splitCondition(condition, "=="); ok {
		actual := lookupValue(key, params, stepOutputs)
		return actual == value
	}

	// If it's just a key name, check if it's truthy (non-empty)
	val := lookupValue(condition, params, stepOutputs)
	return val != "" && val != "false" && val != "0"
}

func splitCondition(s, op string) (key, value string, ok bool) {
	idx := indexOf(s, op)
	if idx < 0 {
		return "", "", false
	}
	key = trimSpaces(s[:idx])
	value = trimSpaces(s[idx+len(op):])
	return key, value, true
}

func lookupValue(key string, params map[string]interface{}, stepOutputs map[string]string) string {
	// Check step outputs first
	if v, ok := stepOutputs[key]; ok {
		return v
	}
	// Then check params
	if v, ok := params[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func trimSpaces(s string) string {
	// Simple trim without importing strings
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// executeChainStep executes a sub-chain inline by looking up the chain by ID.
func (e *Executor) executeChainStep(ctx context.Context, step *ChainStep, exec *ChainExecution, result StepResult) StepResult {
	if step.ChainID == "" {
		result.Status = ChainStatusFailed
		result.Error = "chain step missing chain_id"
		return result
	}

	subChain, ok := e.GetChain(step.ChainID)
	if !ok {
		result.Status = ChainStatusFailed
		result.Error = fmt.Sprintf("sub-chain not found: %s", step.ChainID)
		return result
	}

	// Merge current exec params with step inputs for the sub-chain
	subParams := make(map[string]interface{})
	for k, v := range exec.Parameters {
		subParams[k] = v
	}
	if step.Inputs != nil {
		resolved := e.substituteParams(step.Inputs, exec.Parameters)
		for k, v := range resolved {
			subParams[k] = v
		}
	}

	// Execute sub-chain steps inline (not as a separate execution)
	for _, subStep := range subChain.Steps {
		select {
		case <-ctx.Done():
			result.Status = ChainStatusFailed
			result.Error = "chain step timeout"
			return result
		default:
		}

		subResult := e.executeStep(ctx, &subStep, exec)
		exec.StepResults = append(exec.StepResults, subResult)

		if subResult.Status == ChainStatusFailed && subStep.OnError == OnErrorStop {
			result.Status = ChainStatusFailed
			result.Error = fmt.Sprintf("sub-chain %s failed: %s", step.ChainID, subResult.Error)
			return result
		}
		if subResult.Status == ChainStatusPaused {
			result.Status = ChainStatusPaused
			return result
		}
	}

	result.Status = ChainStatusCompleted
	result.Output = map[string]interface{}{
		"sub_chain": step.ChainID,
		"steps_run": len(subChain.Steps),
	}
	return result
}

// substituteParams replaces {{param}} and {{steps.<stepID>.<key>}} placeholders with actual values.
// The exec parameter is optional — when non-nil, step output references are resolved.
func (e *Executor) substituteParams(inputs map[string]interface{}, params map[string]interface{}) map[string]interface{} {
	return e.substituteParamsWithOutputs(inputs, params, nil)
}

// substituteParamsWithOutputs replaces {{param}} and {{steps.<stepID>.<key>}} placeholders.
func (e *Executor) substituteParamsWithOutputs(inputs map[string]interface{}, params map[string]interface{}, stepOutputs map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range inputs {
		if s, ok := v.(string); ok {
			// Substitute step output references: {{steps.<stepID>.<key>}}
			for outKey, outVal := range stepOutputs {
				placeholder := fmt.Sprintf("{{%s}}", outKey)
				s = replaceAll(s, placeholder, outVal)
			}
			// Substitute parameter references: {{param}}
			for pk, pv := range params {
				placeholder := fmt.Sprintf("{{%s}}", pk)
				if pvStr, ok := pv.(string); ok {
					s = replaceAll(s, placeholder, pvStr)
				}
			}
			result[k] = s
		} else {
			result[k] = v
		}
	}
	return result
}

func replaceAll(s, old, new string) string {
	for {
		i := indexOf(s, old)
		if i < 0 {
			break
		}
		s = s[:i] + new + s[i+len(old):]
	}
	return s
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// updateExecution updates the execution status
func (e *Executor) updateExecution(exec *ChainExecution, status ChainStatus, errMsg string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	exec.Status = status
	if errMsg != "" {
		exec.Error = errMsg
	}
	if status == ChainStatusCompleted || status == ChainStatusFailed || status == ChainStatusCancelled {
		now := time.Now()
		exec.CompletedAt = &now
	}
}

// ApproveGate approves a pending gate
func (e *Executor) ApproveGate(ctx context.Context, executionID string, approved bool, approvedBy, comment string) error {
	e.mu.Lock()
	gate, ok := e.pending[executionID]
	if !ok {
		e.mu.Unlock()
		return fmt.Errorf("no pending gate for execution: %s", executionID)
	}

	exec, ok := e.executions[executionID]
	if !ok {
		e.mu.Unlock()
		return fmt.Errorf("execution not found: %s", executionID)
	}

	chain, ok := e.chains[exec.ChainID]
	if !ok {
		e.mu.Unlock()
		return fmt.Errorf("chain not found: %s", exec.ChainID)
	}

	delete(e.pending, executionID)
	e.mu.Unlock()

	if !approved {
		e.updateExecution(exec, ChainStatusCancelled, fmt.Sprintf("gate rejected by %s: %s", approvedBy, comment))
		return nil
	}

	// Record approval in step result
	exec.StepResults[len(exec.StepResults)-1].Status = ChainStatusCompleted
	exec.StepResults[len(exec.StepResults)-1].Output = map[string]interface{}{
		"approved_by": approvedBy,
		"comment":     comment,
	}
	now := time.Now()
	exec.StepResults[len(exec.StepResults)-1].CompletedAt = &now

	// Find the gate step index and continue from there
	gateIdx := -1
	for i, step := range chain.Steps {
		if step.ID == gate.StepID {
			gateIdx = i
			break
		}
	}

	if gateIdx < 0 || gateIdx >= len(chain.Steps)-1 {
		e.updateExecution(exec, ChainStatusCompleted, "")
		return nil
	}

	// Resume execution from after the gate
	exec.Status = ChainStatusRunning
	go e.resumeChain(ctx, chain, exec, gateIdx+1)

	return nil
}

// resumeChain continues chain execution from a specific step
func (e *Executor) resumeChain(ctx context.Context, chain *Chain, exec *ChainExecution, fromStep int) {
	for i := fromStep; i < len(chain.Steps); i++ {
		step := chain.Steps[i]

		select {
		case <-ctx.Done():
			e.updateExecution(exec, ChainStatusFailed, "chain timeout exceeded")
			return
		default:
		}

		exec.CurrentStep = i + 1
		result := e.executeStep(ctx, &step, exec)
		exec.StepResults = append(exec.StepResults, result)

		if result.Status == ChainStatusFailed && step.OnError == OnErrorStop {
			e.updateExecution(exec, ChainStatusFailed, result.Error)
			return
		}

		if result.Status == ChainStatusPaused {
			e.updateExecution(exec, ChainStatusPaused, "")
			return
		}

		if step.DelayAfter > 0 {
			time.Sleep(step.DelayAfter)
		}
	}

	e.updateExecution(exec, ChainStatusCompleted, "")
}

// GetExecution returns an execution by ID
func (e *Executor) GetExecution(id string) (*ChainExecution, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	exec, ok := e.executions[id]
	return exec, ok
}

// ListExecutions returns recent executions
func (e *Executor) ListExecutions(limit int) []*ChainExecution {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*ChainExecution, 0, len(e.executions))
	for _, exec := range e.executions {
		result = append(result, exec)
	}

	// Sort by start time descending
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].StartedAt.After(result[i].StartedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

// ListPendingGates returns all pending gate approvals
func (e *Executor) ListPendingGates() []*PendingGate {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*PendingGate, 0, len(e.pending))
	for _, gate := range e.pending {
		result = append(result, gate)
	}
	return result
}

// CancelExecution cancels a running execution
func (e *Executor) CancelExecution(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	exec, ok := e.executions[id]
	if !ok {
		return fmt.Errorf("execution not found: %s", id)
	}

	if exec.Status != ChainStatusRunning && exec.Status != ChainStatusPaused {
		return fmt.Errorf("execution is not running or paused")
	}

	exec.Status = ChainStatusCancelled
	now := time.Now()
	exec.CompletedAt = &now

	delete(e.pending, id)
	return nil
}

// Global executor instance
var (
	globalExecutor     *Executor
	globalExecutorOnce sync.Once
)

// GetExecutor returns the global chain executor
func GetExecutor() *Executor {
	globalExecutorOnce.Do(func() {
		globalExecutor = NewExecutor(config.GetOrLoad().AftrsDataDir, tools.GetRegistry())
	})
	return globalExecutor
}
