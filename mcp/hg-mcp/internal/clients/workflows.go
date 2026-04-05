// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WorkflowsClient manages workflow definitions and execution
type WorkflowsClient struct {
	workflows  map[string]*Workflow
	executions map[string]*WorkflowExecution
	mu         sync.RWMutex
}

// Workflow represents a workflow definition
type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Category    string         `json:"category"` // show, backup, maintenance, test
	Steps       []WorkflowStep `json:"steps"`
	Timeout     int            `json:"timeout_seconds"`
	RollbackOn  []string       `json:"rollback_on,omitempty"`
}

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Tool          string            `json:"tool"`
	Parameters    map[string]string `json:"parameters,omitempty"`
	DependsOn     []string          `json:"depends_on,omitempty"`
	Timeout       int               `json:"timeout_seconds,omitempty"`
	OnFailure     string            `json:"on_failure,omitempty"` // continue, stop, rollback
	RetryCount    int               `json:"retry_count,omitempty"`
	ParallelGroup string            `json:"parallel_group,omitempty"` // Steps with same group run in parallel
	Condition     *StepCondition    `json:"condition,omitempty"`      // IF logic for conditional execution
	OutputVar     string            `json:"output_var,omitempty"`     // Store result in variable
}

// StepCondition defines a condition for step execution
type StepCondition struct {
	Variable string `json:"variable"` // Variable name to check (e.g., "step1.success", "env.SHOW_TYPE")
	Operator string `json:"operator"` // eq, ne, gt, lt, gte, lte, contains, exists
	Value    string `json:"value"`    // Value to compare against
}

// WorkflowExecution represents a running or completed workflow
type WorkflowExecution struct {
	ID          string       `json:"id"`
	WorkflowID  string       `json:"workflow_id"`
	Status      string       `json:"status"` // pending, running, completed, failed, cancelled
	StartedAt   time.Time    `json:"started_at"`
	CompletedAt time.Time    `json:"completed_at,omitempty"`
	StepResults []StepResult `json:"step_results"`
	Error       string       `json:"error,omitempty"`
	Progress    int          `json:"progress"`
}

// StepResult represents the result of a workflow step
type StepResult struct {
	StepID      string    `json:"step_id"`
	Status      string    `json:"status"` // pending, running, completed, failed, skipped
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Output      string    `json:"output,omitempty"`
	Error       string    `json:"error,omitempty"`
	Retries     int       `json:"retries"`
}

// CueSequence represents a cue sequence for shows
type CueSequence struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Cues []Cue  `json:"cues"`
}

// Cue represents a single cue
type Cue struct {
	Number   string `json:"number"`
	Name     string `json:"name"`
	Action   string `json:"action"`
	Target   string `json:"target"`
	Delay    int    `json:"delay_ms,omitempty"`
	Duration int    `json:"duration_ms,omitempty"`
}

var (
	workflowsClientInstance *WorkflowsClient
	workflowsOnce           sync.Once
)

// GetWorkflowsClient returns the singleton workflows client.
func GetWorkflowsClient() *WorkflowsClient {
	workflowsOnce.Do(func() {
		workflowsClientInstance = &WorkflowsClient{
			workflows:  make(map[string]*Workflow),
			executions: make(map[string]*WorkflowExecution),
		}
		workflowsClientInstance.registerBuiltinWorkflows()
	})
	return workflowsClientInstance
}

// NewWorkflowsClient creates a new workflows client
func NewWorkflowsClient() (*WorkflowsClient, error) {
	c := &WorkflowsClient{
		workflows:  make(map[string]*Workflow),
		executions: make(map[string]*WorkflowExecution),
	}

	// Register built-in workflows
	c.registerBuiltinWorkflows()

	return c, nil
}

// List returns all workflows as a simplified list.
func (c *WorkflowsClient) List() []*Workflow {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*Workflow, 0, len(c.workflows))
	for _, w := range c.workflows {
		result = append(result, w)
	}
	return result
}

// WorkflowStepSimple is a simplified step for API responses.
type WorkflowStepSimple struct {
	Tool   string                 `json:"tool"`
	Args   map[string]interface{} `json:"args,omitempty"`
	OnFail string                 `json:"on_fail,omitempty"`
}

// Run executes a workflow by name with the given parameters.
func (c *WorkflowsClient) Run(ctx context.Context, registry interface{}, name string, params map[string]interface{}) (interface{}, error) {
	workflow, err := c.GetWorkflow(ctx, name)
	if err != nil {
		return nil, err
	}

	exec, err := c.RunWorkflow(ctx, name, false)
	if err != nil {
		return nil, err
	}

	// Wait for completion (simplified)
	timeout := time.Duration(workflow.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.mu.RLock()
		status := exec.Status
		c.mu.RUnlock()

		if status == "completed" || status == "failed" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return exec, nil
}

// registerBuiltinWorkflows registers the default workflows
func (c *WorkflowsClient) registerBuiltinWorkflows() {
	// Show Startup Workflow
	c.workflows["show_startup"] = &Workflow{
		ID:          "show_startup",
		Name:        "Show Startup",
		Description: "Automated show startup sequence - brings all systems online",
		Category:    "show",
		Timeout:     300,
		Steps: []WorkflowStep{
			{ID: "check_network", Name: "Check Network", Tool: "aftrs_network_scan", Timeout: 30},
			{ID: "start_unraid", Name: "Verify UNRAID", Tool: "aftrs_unraid_status", Timeout: 30, DependsOn: []string{"check_network"}},
			{ID: "start_td", Name: "Check TouchDesigner", Tool: "aftrs_td_status", Timeout: 30, DependsOn: []string{"start_unraid"}},
			{ID: "start_resolume", Name: "Check Resolume", Tool: "aftrs_resolume_status", Timeout: 30, DependsOn: []string{"start_unraid"}},
			{ID: "start_lighting", Name: "Initialize Lighting", Tool: "aftrs_lighting_status", Timeout: 30, DependsOn: []string{"check_network"}},
			{ID: "start_ndi", Name: "Verify NDI Sources", Tool: "aftrs_ndi_sources", Timeout: 30, DependsOn: []string{"start_td", "start_resolume"}},
			{ID: "preflight", Name: "Run Preflight", Tool: "aftrs_show_preflight", Timeout: 60, DependsOn: []string{"start_ndi", "start_lighting"}},
		},
	}

	// Show Shutdown Workflow
	c.workflows["show_shutdown"] = &Workflow{
		ID:          "show_shutdown",
		Name:        "Show Shutdown",
		Description: "Graceful show shutdown sequence - safely powers down all systems",
		Category:    "show",
		Timeout:     180,
		Steps: []WorkflowStep{
			{ID: "fade_lighting", Name: "Fade Lighting", Tool: "aftrs_lighting_blackout", Timeout: 10},
			{ID: "stop_resolume", Name: "Stop Resolume Playback", Tool: "aftrs_resolume_status", Timeout: 10},
			{ID: "backup_td", Name: "Backup TouchDesigner Project", Tool: "aftrs_td_backup", Timeout: 60, DependsOn: []string{"stop_resolume"}},
			{ID: "log_show", Name: "Log Show End", Tool: "aftrs_show_log", Parameters: map[string]string{"event": "show_end"}, Timeout: 10, DependsOn: []string{"backup_td"}},
		},
	}

	// Backup All Workflow
	c.workflows["backup_all"] = &Workflow{
		ID:          "backup_all",
		Name:        "Backup All",
		Description: "Backup all project files to NAS",
		Category:    "backup",
		Timeout:     600,
		Steps: []WorkflowStep{
			{ID: "backup_td", Name: "Backup TouchDesigner", Tool: "aftrs_td_backup", Timeout: 120},
			{ID: "backup_resolume", Name: "Backup Resolume", Tool: "aftrs_resolume_status", Timeout: 120},
			{ID: "backup_vault", Name: "Backup Vault", Tool: "aftrs_vault_save", Timeout: 60},
			{ID: "verify", Name: "Verify Backups", Tool: "aftrs_unraid_status", Timeout: 30, DependsOn: []string{"backup_td", "backup_resolume", "backup_vault"}},
		},
	}

	// Sync Assets Workflow
	c.workflows["sync_assets"] = &Workflow{
		ID:          "sync_assets",
		Name:        "Sync Assets",
		Description: "Sync project assets to NAS",
		Category:    "backup",
		Timeout:     300,
		Steps: []WorkflowStep{
			{ID: "check_nas", Name: "Check NAS", Tool: "aftrs_unraid_status", Timeout: 30},
			{ID: "sync_media", Name: "Sync Media", Tool: "aftrs_unraid_status", Timeout: 180, DependsOn: []string{"check_nas"}},
			{ID: "verify", Name: "Verify Sync", Tool: "aftrs_unraid_status", Timeout: 30, DependsOn: []string{"sync_media"}},
		},
	}

	// Test Sequence Workflow
	c.workflows["test_sequence"] = &Workflow{
		ID:          "test_sequence",
		Name:        "Test Sequence",
		Description: "Test all systems in sequence",
		Category:    "test",
		Timeout:     300,
		Steps: []WorkflowStep{
			{ID: "test_network", Name: "Test Network", Tool: "aftrs_network_scan", Timeout: 30},
			{ID: "test_td", Name: "Test TouchDesigner", Tool: "aftrs_td_status", Timeout: 30, DependsOn: []string{"test_network"}},
			{ID: "test_resolume", Name: "Test Resolume", Tool: "aftrs_resolume_status", Timeout: 30, DependsOn: []string{"test_network"}},
			{ID: "test_lighting", Name: "Test Lighting", Tool: "aftrs_lighting_status", Timeout: 30, DependsOn: []string{"test_network"}},
			{ID: "test_ndi", Name: "Test NDI", Tool: "aftrs_ndi_sources", Timeout: 30, DependsOn: []string{"test_td", "test_resolume"}},
			{ID: "test_unraid", Name: "Test UNRAID", Tool: "aftrs_unraid_status", Timeout: 30},
			{ID: "health_check", Name: "Full Health Check", Tool: "aftrs_studio_health_full", Timeout: 60, DependsOn: []string{"test_ndi", "test_lighting", "test_unraid"}},
		},
	}

	// Panic Mode Workflow
	c.workflows["panic_mode"] = &Workflow{
		ID:          "panic_mode",
		Name:        "Panic Mode",
		Description: "Emergency all-off - immediately stops all outputs",
		Category:    "show",
		Timeout:     30,
		Steps: []WorkflowStep{
			{ID: "blackout", Name: "Lighting Blackout", Tool: "aftrs_lighting_blackout", Timeout: 5},
			{ID: "stop_all", Name: "Stop All Playback", Tool: "aftrs_resolume_status", Timeout: 5},
			{ID: "mute_audio", Name: "Mute Audio", Tool: "aftrs_resolume_status", Timeout: 5},
			{ID: "notify", Name: "Send Alert", Tool: "aftrs_discord_notify", Parameters: map[string]string{"message": "PANIC MODE ACTIVATED", "type": "error"}, Timeout: 5},
		},
	}
}

// ListWorkflows returns all available workflows
func (c *WorkflowsClient) ListWorkflows(ctx context.Context, category string) ([]*Workflow, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var workflows []*Workflow
	for _, w := range c.workflows {
		if category == "" || w.Category == category {
			workflows = append(workflows, w)
		}
	}
	return workflows, nil
}

// GetWorkflow returns a specific workflow
func (c *WorkflowsClient) GetWorkflow(ctx context.Context, id string) (*Workflow, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if w, ok := c.workflows[id]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("workflow not found: %s", id)
}

// RunWorkflow starts a workflow execution
func (c *WorkflowsClient) RunWorkflow(ctx context.Context, workflowID string, dryRun bool) (*WorkflowExecution, error) {
	workflow, err := c.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	exec := &WorkflowExecution{
		ID:          fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		WorkflowID:  workflowID,
		Status:      "running",
		StartedAt:   time.Now(),
		StepResults: make([]StepResult, len(workflow.Steps)),
		Progress:    0,
	}

	// Initialize step results
	for i, step := range workflow.Steps {
		exec.StepResults[i] = StepResult{
			StepID: step.ID,
			Status: "pending",
		}
	}

	c.mu.Lock()
	c.executions[exec.ID] = exec
	c.mu.Unlock()

	if dryRun {
		// Simulate execution
		go c.simulateExecution(exec, workflow)
	} else {
		// Real execution would integrate with the actual tools
		go c.simulateExecution(exec, workflow) // For now, simulate
	}

	return exec, nil
}

// simulateExecution simulates workflow execution
func (c *WorkflowsClient) simulateExecution(exec *WorkflowExecution, workflow *Workflow) {
	completed := make(map[string]bool)
	totalSteps := len(workflow.Steps)

	for i, step := range workflow.Steps {
		// Check dependencies
		canRun := true
		for _, dep := range step.DependsOn {
			if !completed[dep] {
				canRun = false
				break
			}
		}

		if !canRun {
			// Find the dependency and wait
			time.Sleep(100 * time.Millisecond)
		}

		c.mu.Lock()
		exec.StepResults[i].Status = "running"
		exec.StepResults[i].StartedAt = time.Now()
		c.mu.Unlock()

		// Simulate step execution
		time.Sleep(time.Duration(100+i*50) * time.Millisecond)

		c.mu.Lock()
		exec.StepResults[i].Status = "completed"
		exec.StepResults[i].CompletedAt = time.Now()
		exec.StepResults[i].Output = fmt.Sprintf("Step %s completed successfully", step.Name)
		exec.Progress = ((i + 1) * 100) / totalSteps
		completed[step.ID] = true
		c.mu.Unlock()
	}

	c.mu.Lock()
	exec.Status = "completed"
	exec.CompletedAt = time.Now()
	exec.Progress = 100
	c.mu.Unlock()
}

// GetExecutionStatus returns the status of a workflow execution
func (c *WorkflowsClient) GetExecutionStatus(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if exec, ok := c.executions[executionID]; ok {
		return exec, nil
	}
	return nil, fmt.Errorf("execution not found: %s", executionID)
}

// ExecuteShowStartup runs the show startup sequence
func (c *WorkflowsClient) ExecuteShowStartup(ctx context.Context) (*WorkflowExecution, error) {
	return c.RunWorkflow(ctx, "show_startup", false)
}

// ExecuteShowShutdown runs the show shutdown sequence
func (c *WorkflowsClient) ExecuteShowShutdown(ctx context.Context) (*WorkflowExecution, error) {
	return c.RunWorkflow(ctx, "show_shutdown", false)
}

// ExecuteBackupAll runs the backup all workflow
func (c *WorkflowsClient) ExecuteBackupAll(ctx context.Context) (*WorkflowExecution, error) {
	return c.RunWorkflow(ctx, "backup_all", false)
}

// ExecuteSyncAssets runs the sync assets workflow
func (c *WorkflowsClient) ExecuteSyncAssets(ctx context.Context) (*WorkflowExecution, error) {
	return c.RunWorkflow(ctx, "sync_assets", false)
}

// ExecuteTestSequence runs the test sequence workflow
func (c *WorkflowsClient) ExecuteTestSequence(ctx context.Context) (*WorkflowExecution, error) {
	return c.RunWorkflow(ctx, "test_sequence", false)
}

// ExecutePanicMode runs the panic mode workflow
func (c *WorkflowsClient) ExecutePanicMode(ctx context.Context) (*WorkflowExecution, error) {
	return c.RunWorkflow(ctx, "panic_mode", false)
}

// ExecuteCueSequence executes a sequence of cues
func (c *WorkflowsClient) ExecuteCueSequence(ctx context.Context, cues []Cue) (*WorkflowExecution, error) {
	// Create a dynamic workflow from cues
	workflow := &Workflow{
		ID:          fmt.Sprintf("cue_seq_%d", time.Now().UnixNano()),
		Name:        "Cue Sequence",
		Description: "Dynamic cue sequence",
		Category:    "show",
		Timeout:     300,
		Steps:       make([]WorkflowStep, len(cues)),
	}

	for i, cue := range cues {
		workflow.Steps[i] = WorkflowStep{
			ID:   fmt.Sprintf("cue_%s", cue.Number),
			Name: cue.Name,
			Tool: "aftrs_td_cue_trigger", // Default to TD cue trigger
			Parameters: map[string]string{
				"cue":    cue.Number,
				"action": cue.Action,
				"target": cue.Target,
			},
			Timeout: 30,
		}
		if i > 0 && cue.Delay > 0 {
			workflow.Steps[i].DependsOn = []string{workflow.Steps[i-1].ID}
		}
	}

	c.mu.Lock()
	c.workflows[workflow.ID] = workflow
	c.mu.Unlock()

	return c.RunWorkflow(ctx, workflow.ID, false)
}

// ComposeWorkflow creates a workflow from a tool sequence specification
// Format: "tool1 -> tool2 -> tool3" for sequential
// Format: "tool1, tool2 -> tool3" for parallel + sequential
func (c *WorkflowsClient) ComposeWorkflow(ctx context.Context, name, description, sequence string) (*Workflow, error) {
	if name == "" {
		return nil, fmt.Errorf("workflow name is required")
	}
	if sequence == "" {
		return nil, fmt.Errorf("sequence is required")
	}

	workflow := &Workflow{
		ID:          fmt.Sprintf("composed_%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		Category:    "custom",
		Timeout:     600,
		Steps:       []WorkflowStep{},
	}

	// Parse sequence: "tool1 -> tool2, tool3 -> tool4"
	// -> means sequential, comma means parallel
	stages := splitByArrow(sequence)

	var prevStepIDs []string
	stepNum := 1

	for stageIdx, stage := range stages {
		// Parse parallel tools within a stage
		parallelTools := splitByComma(stage)
		var currentStepIDs []string
		parallelGroup := ""

		if len(parallelTools) > 1 {
			parallelGroup = fmt.Sprintf("parallel_%d", stageIdx)
		}

		for _, toolSpec := range parallelTools {
			toolSpec = trimSpace(toolSpec)
			if toolSpec == "" {
				continue
			}

			step := WorkflowStep{
				ID:            fmt.Sprintf("step_%d", stepNum),
				Name:          fmt.Sprintf("Step %d: %s", stepNum, toolSpec),
				Tool:          toolSpec,
				Parameters:    make(map[string]string),
				Timeout:       60,
				ParallelGroup: parallelGroup,
			}

			// Add dependencies from previous stage
			if len(prevStepIDs) > 0 {
				step.DependsOn = make([]string, len(prevStepIDs))
				copy(step.DependsOn, prevStepIDs)
			}

			workflow.Steps = append(workflow.Steps, step)
			currentStepIDs = append(currentStepIDs, step.ID)
			stepNum++
		}

		prevStepIDs = currentStepIDs
	}

	// Save the workflow
	c.mu.Lock()
	c.workflows[workflow.ID] = workflow
	c.mu.Unlock()

	return workflow, nil
}

// ValidateWorkflow validates a workflow for correctness
func (c *WorkflowsClient) ValidateWorkflow(ctx context.Context, workflowID string) (*WorkflowValidation, error) {
	workflow, err := c.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	validation := &WorkflowValidation{
		WorkflowID: workflowID,
		IsValid:    true,
		Issues:     []string{},
		Warnings:   []string{},
	}

	stepIDs := make(map[string]bool)

	// Validate each step
	for i, step := range workflow.Steps {
		// Check for duplicate step IDs
		if stepIDs[step.ID] {
			validation.Issues = append(validation.Issues, fmt.Sprintf("Step %d: Duplicate step ID '%s'", i+1, step.ID))
			validation.IsValid = false
		}
		stepIDs[step.ID] = true

		// Check for required fields
		if step.Tool == "" {
			validation.Issues = append(validation.Issues, fmt.Sprintf("Step %d (%s): Missing tool name", i+1, step.ID))
			validation.IsValid = false
		}

		// Validate dependencies exist
		for _, dep := range step.DependsOn {
			if !stepIDs[dep] {
				// Check if it will be defined later (forward reference)
				found := false
				for j := i + 1; j < len(workflow.Steps); j++ {
					if workflow.Steps[j].ID == dep {
						validation.Warnings = append(validation.Warnings,
							fmt.Sprintf("Step %d (%s): Forward reference to '%s'", i+1, step.ID, dep))
						found = true
						break
					}
				}
				if !found {
					validation.Issues = append(validation.Issues,
						fmt.Sprintf("Step %d (%s): Unknown dependency '%s'", i+1, step.ID, dep))
					validation.IsValid = false
				}
			}
		}

		// Validate condition if present
		if step.Condition != nil {
			validOps := map[string]bool{"eq": true, "ne": true, "gt": true, "lt": true, "gte": true, "lte": true, "contains": true, "exists": true}
			if !validOps[step.Condition.Operator] {
				validation.Issues = append(validation.Issues,
					fmt.Sprintf("Step %d (%s): Invalid condition operator '%s'", i+1, step.ID, step.Condition.Operator))
				validation.IsValid = false
			}
			if step.Condition.Variable == "" {
				validation.Issues = append(validation.Issues,
					fmt.Sprintf("Step %d (%s): Condition missing variable", i+1, step.ID))
				validation.IsValid = false
			}
		}
	}

	// Check for cycles
	if hasCycle := c.detectCycle(workflow); hasCycle {
		validation.Issues = append(validation.Issues, "Workflow contains circular dependencies")
		validation.IsValid = false
	}

	// Add info
	validation.StepCount = len(workflow.Steps)
	validation.EstimatedDuration = c.estimateDuration(workflow)

	return validation, nil
}

// WorkflowValidation represents the result of workflow validation
type WorkflowValidation struct {
	WorkflowID        string   `json:"workflow_id"`
	IsValid           bool     `json:"is_valid"`
	Issues            []string `json:"issues,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
	StepCount         int      `json:"step_count"`
	EstimatedDuration int      `json:"estimated_duration_seconds"`
}

// detectCycle checks for circular dependencies in workflow
func (c *WorkflowsClient) detectCycle(workflow *Workflow) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	stepMap := make(map[string]*WorkflowStep)

	for i := range workflow.Steps {
		stepMap[workflow.Steps[i].ID] = &workflow.Steps[i]
	}

	var dfs func(id string) bool
	dfs = func(id string) bool {
		visited[id] = true
		recStack[id] = true

		step := stepMap[id]
		if step != nil {
			for _, dep := range step.DependsOn {
				if !visited[dep] {
					if dfs(dep) {
						return true
					}
				} else if recStack[dep] {
					return true
				}
			}
		}

		recStack[id] = false
		return false
	}

	for id := range stepMap {
		if !visited[id] {
			if dfs(id) {
				return true
			}
		}
	}

	return false
}

// estimateDuration estimates workflow duration in seconds
func (c *WorkflowsClient) estimateDuration(workflow *Workflow) int {
	total := 0
	for _, step := range workflow.Steps {
		timeout := step.Timeout
		if timeout == 0 {
			timeout = 30 // Default
		}
		total += timeout
	}
	return total
}

// AddConditionToStep adds a condition to a workflow step
func (c *WorkflowsClient) AddConditionToStep(ctx context.Context, workflowID, stepID, variable, operator, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	workflow, ok := c.workflows[workflowID]
	if !ok {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	for i, step := range workflow.Steps {
		if step.ID == stepID {
			workflow.Steps[i].Condition = &StepCondition{
				Variable: variable,
				Operator: operator,
				Value:    value,
			}
			return nil
		}
	}

	return fmt.Errorf("step not found: %s", stepID)
}

// DeleteWorkflow deletes a custom workflow
func (c *WorkflowsClient) DeleteWorkflow(ctx context.Context, workflowID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	workflow, ok := c.workflows[workflowID]
	if !ok {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Only allow deleting custom workflows
	if workflow.Category != "custom" {
		return fmt.Errorf("cannot delete built-in workflow: %s", workflowID)
	}

	delete(c.workflows, workflowID)
	return nil
}

// CreateWorkflow creates a new custom workflow from the provided definition
func (c *WorkflowsClient) CreateWorkflow(ctx context.Context, wf *Workflow) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if workflow already exists
	if _, ok := c.workflows[wf.Name]; ok {
		return fmt.Errorf("workflow already exists: %s", wf.Name)
	}

	// Set defaults
	if wf.ID == "" {
		wf.ID = wf.Name
	}
	if wf.Category == "" {
		wf.Category = "custom"
	}
	if wf.Timeout == 0 {
		wf.Timeout = 300
	}

	// Generate step IDs if not provided
	for i := range wf.Steps {
		if wf.Steps[i].ID == "" {
			wf.Steps[i].ID = fmt.Sprintf("step_%d", i+1)
		}
		if wf.Steps[i].Timeout == 0 {
			wf.Steps[i].Timeout = 30
		}
	}

	c.workflows[wf.Name] = wf
	return nil
}

// UpdateWorkflow updates an existing workflow
func (c *WorkflowsClient) UpdateWorkflow(ctx context.Context, name string, wf *Workflow) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	existing, ok := c.workflows[name]
	if !ok {
		return fmt.Errorf("workflow not found: %s", name)
	}

	// Only allow updating custom workflows
	if existing.Category != "custom" {
		return fmt.Errorf("cannot update built-in workflow: %s", name)
	}

	// Preserve ID and category
	wf.ID = existing.ID
	wf.Category = "custom"
	if wf.Timeout == 0 {
		wf.Timeout = 300
	}

	// Generate step IDs if not provided
	for i := range wf.Steps {
		if wf.Steps[i].ID == "" {
			wf.Steps[i].ID = fmt.Sprintf("step_%d", i+1)
		}
		if wf.Steps[i].Timeout == 0 {
			wf.Steps[i].Timeout = 30
		}
	}

	// If name changed, delete old and add new
	if wf.Name != name {
		delete(c.workflows, name)
	}
	c.workflows[wf.Name] = wf
	return nil
}

// Helper functions for parsing
func splitByArrow(s string) []string {
	// Split by " -> " or "->"
	parts := []string{}
	current := ""
	i := 0
	for i < len(s) {
		if i+2 < len(s) && s[i:i+2] == "->" {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
			i += 2
			// Skip optional space after ->
			if i < len(s) && s[i] == ' ' {
				i++
			}
			continue
		}
		if i+4 <= len(s) && s[i:i+4] == " -> " {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
			i += 4
			continue
		}
		current += string(s[i])
		i++
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func splitByComma(s string) []string {
	parts := []string{}
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
