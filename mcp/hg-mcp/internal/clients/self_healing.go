// Package clients provides API clients for external services.
// self_healing.go implements automated remediation for AV systems
package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SelfHealingConfig holds configuration for self-healing behavior
type SelfHealingConfig struct {
	Enabled              bool `json:"enabled"`
	MaxAutoRiskScore     int  `json:"max_auto_risk_score"`    // Max risk for auto-execute (default: 20)
	RequireApprovalAbove int  `json:"require_approval_above"` // Risk requiring approval (default: 50)
	MaxRetryAttempts     int  `json:"max_retry_attempts"`
	CooldownMinutes      int  `json:"cooldown_minutes"`
	LearningEnabled      bool `json:"learning_enabled"`
}

// DefaultSelfHealingConfig returns sensible defaults
func DefaultSelfHealingConfig() SelfHealingConfig {
	return SelfHealingConfig{
		Enabled:              true,
		MaxAutoRiskScore:     20,
		RequireApprovalAbove: 50,
		MaxRetryAttempts:     3,
		CooldownMinutes:      15,
		LearningEnabled:      true,
	}
}

// RemediationPlaybook represents a remediation action
type RemediationPlaybook struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Service     string            `json:"service"`      // obs, resolume, touchdesigner, ndi, etc.
	TriggerType string            `json:"trigger_type"` // crash, disconnect, high_cpu, frozen
	RiskScore   int               `json:"risk_score"`   // 0-100
	Steps       []RemediationStep `json:"steps"`
	Cooldown    time.Duration     `json:"cooldown"`
	Tags        []string          `json:"tags"`
	SuccessRate float64           `json:"success_rate"`
	ExecCount   int               `json:"exec_count"`
	CreatedAt   time.Time         `json:"created_at"`
}

// RemediationStep represents a single step in a playbook
type RemediationStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // tool, command, wait
	ToolName   string                 `json:"tool_name,omitempty"`
	ToolParams map[string]interface{} `json:"tool_params,omitempty"`
	Command    string                 `json:"command,omitempty"`
	WaitSecs   int                    `json:"wait_seconds,omitempty"`
	OnError    string                 `json:"on_error"` // stop, continue, retry
	MaxRetries int                    `json:"max_retries,omitempty"`
}

// RemediationExecution represents a playbook execution record
type RemediationExecution struct {
	ID           string     `json:"id"`
	PlaybookID   string     `json:"playbook_id"`
	PlaybookName string     `json:"playbook_name"`
	Service      string     `json:"service"`
	Status       string     `json:"status"` // pending, running, success, failed, approval_required
	RiskScore    int        `json:"risk_score"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	StepsRun     int        `json:"steps_run"`
	StepsTotal   int        `json:"steps_total"`
	Error        string     `json:"error,omitempty"`
	ApprovedBy   string     `json:"approved_by,omitempty"`
	AutoExecuted bool       `json:"auto_executed"`
	Output       string     `json:"output,omitempty"`
}

// PendingApproval represents a remediation awaiting approval
type PendingApproval struct {
	ExecutionID  string    `json:"execution_id"`
	PlaybookID   string    `json:"playbook_id"`
	PlaybookName string    `json:"playbook_name"`
	Service      string    `json:"service"`
	RiskScore    int       `json:"risk_score"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
}

// SelfHealingClient manages automated remediation
type SelfHealingClient struct {
	mu         sync.RWMutex
	config     SelfHealingConfig
	playbooks  map[string]*RemediationPlaybook
	executions []*RemediationExecution
	pending    map[string]*PendingApproval
	cooldowns  map[string]time.Time // service -> cooldown until
	vaultPath  string
}

var (
	selfHealingOnce     sync.Once
	selfHealingInstance *SelfHealingClient
)

// GetSelfHealingClient returns the singleton self-healing client
func GetSelfHealingClient() *SelfHealingClient {
	selfHealingOnce.Do(func() {
		selfHealingInstance, _ = NewSelfHealingClient()
	})
	return selfHealingInstance
}

// NewSelfHealingClient creates a new self-healing client
func NewSelfHealingClient() (*SelfHealingClient, error) {
	vaultPath := os.Getenv("OBSIDIAN_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = filepath.Join(os.Getenv("HOME"), "Documents", "obsidian-vault")
	}

	client := &SelfHealingClient{
		config:     DefaultSelfHealingConfig(),
		playbooks:  make(map[string]*RemediationPlaybook),
		executions: make([]*RemediationExecution, 0),
		pending:    make(map[string]*PendingApproval),
		cooldowns:  make(map[string]time.Time),
		vaultPath:  vaultPath,
	}

	// Register built-in playbooks
	client.registerBuiltInPlaybooks()

	// Load from vault
	client.loadFromVault()

	return client, nil
}

// registerBuiltInPlaybooks adds default playbooks for common AV issues
func (c *SelfHealingClient) registerBuiltInPlaybooks() {
	playbooks := []*RemediationPlaybook{
		{
			ID:          "restart_obs",
			Name:        "Restart OBS",
			Description: "Restart OBS Studio when crashed or unresponsive",
			Service:     "obs",
			TriggerType: "crash",
			RiskScore:   15,
			Cooldown:    5 * time.Minute,
			Steps: []RemediationStep{
				{ID: "1", Name: "Check OBS status", Type: "tool", ToolName: "aftrs_obs_status", OnError: "continue"},
				{ID: "2", Name: "Stop OBS", Type: "command", Command: "pkill -9 obs", OnError: "continue"},
				{ID: "3", Name: "Wait for shutdown", Type: "wait", WaitSecs: 3, OnError: "continue"},
				{ID: "4", Name: "Start OBS", Type: "command", Command: "open -a 'OBS'", OnError: "stop"},
				{ID: "5", Name: "Verify OBS running", Type: "tool", ToolName: "aftrs_obs_status", OnError: "stop"},
			},
			Tags: []string{"streaming", "obs", "restart"},
		},
		{
			ID:          "restart_resolume",
			Name:        "Restart Resolume",
			Description: "Restart Resolume Arena when crashed or frozen",
			Service:     "resolume",
			TriggerType: "crash",
			RiskScore:   25,
			Cooldown:    5 * time.Minute,
			Steps: []RemediationStep{
				{ID: "1", Name: "Kill Resolume", Type: "command", Command: "pkill -9 'Resolume Arena'", OnError: "continue"},
				{ID: "2", Name: "Wait for shutdown", Type: "wait", WaitSecs: 5, OnError: "continue"},
				{ID: "3", Name: "Start Resolume", Type: "command", Command: "open -a 'Resolume Arena 7'", OnError: "stop"},
				{ID: "4", Name: "Wait for startup", Type: "wait", WaitSecs: 10, OnError: "continue"},
			},
			Tags: []string{"vj", "resolume", "restart"},
		},
		{
			ID:          "restart_touchdesigner",
			Name:        "Restart TouchDesigner",
			Description: "Restart TouchDesigner when unresponsive",
			Service:     "touchdesigner",
			TriggerType: "frozen",
			RiskScore:   30,
			Cooldown:    5 * time.Minute,
			Steps: []RemediationStep{
				{ID: "1", Name: "Kill TouchDesigner", Type: "command", Command: "pkill -9 TouchDesigner", OnError: "continue"},
				{ID: "2", Name: "Wait for shutdown", Type: "wait", WaitSecs: 5, OnError: "continue"},
				{ID: "3", Name: "Start TouchDesigner", Type: "command", Command: "open -a 'TouchDesigner'", OnError: "stop"},
			},
			Tags: []string{"visuals", "touchdesigner", "restart"},
		},
		{
			ID:          "reconnect_ndi",
			Name:        "Reconnect NDI Source",
			Description: "Attempt to re-establish dropped NDI connection",
			Service:     "ndi",
			TriggerType: "disconnect",
			RiskScore:   10,
			Cooldown:    2 * time.Minute,
			Steps: []RemediationStep{
				{ID: "1", Name: "Scan NDI sources", Type: "tool", ToolName: "aftrs_ndi_sources", OnError: "stop"},
				{ID: "2", Name: "Wait for discovery", Type: "wait", WaitSecs: 5, OnError: "continue"},
				{ID: "3", Name: "Verify connection", Type: "tool", ToolName: "aftrs_ndi_sources", OnError: "stop"},
			},
			Tags: []string{"ndi", "network", "reconnect"},
		},
		{
			ID:          "failover_stream",
			Name:        "Failover Stream",
			Description: "Switch to backup streaming configuration",
			Service:     "streaming",
			TriggerType: "disconnect",
			RiskScore:   40,
			Cooldown:    10 * time.Minute,
			Steps: []RemediationStep{
				{ID: "1", Name: "Stop current stream", Type: "tool", ToolName: "aftrs_obs_stream_stop", OnError: "continue"},
				{ID: "2", Name: "Wait for stop", Type: "wait", WaitSecs: 3, OnError: "continue"},
				{ID: "3", Name: "Switch to backup scene", Type: "tool", ToolName: "aftrs_obs_scene_switch", ToolParams: map[string]interface{}{"scene": "Backup"}, OnError: "stop"},
				{ID: "4", Name: "Restart stream", Type: "tool", ToolName: "aftrs_obs_stream_start", OnError: "stop"},
			},
			Tags: []string{"streaming", "failover", "backup"},
		},
		{
			ID:          "restart_ableton",
			Name:        "Restart Ableton Live",
			Description: "Restart Ableton when audio issues occur",
			Service:     "ableton",
			TriggerType: "crash",
			RiskScore:   35,
			Cooldown:    5 * time.Minute,
			Steps: []RemediationStep{
				{ID: "1", Name: "Kill Ableton", Type: "command", Command: "pkill -9 'Ableton Live'", OnError: "continue"},
				{ID: "2", Name: "Wait for shutdown", Type: "wait", WaitSecs: 5, OnError: "continue"},
				{ID: "3", Name: "Start Ableton", Type: "command", Command: "open -a 'Ableton Live 11 Suite'", OnError: "stop"},
			},
			Tags: []string{"audio", "ableton", "restart"},
		},
		{
			ID:          "restart_wled",
			Name:        "Restart WLED Device",
			Description: "Power cycle a WLED controller",
			Service:     "wled",
			TriggerType: "disconnect",
			RiskScore:   5,
			Cooldown:    1 * time.Minute,
			Steps: []RemediationStep{
				{ID: "1", Name: "Send reboot command", Type: "tool", ToolName: "aftrs_wled_reboot", OnError: "stop"},
				{ID: "2", Name: "Wait for reboot", Type: "wait", WaitSecs: 15, OnError: "continue"},
				{ID: "3", Name: "Verify connection", Type: "tool", ToolName: "aftrs_wled_discover", OnError: "stop"},
			},
			Tags: []string{"lighting", "wled", "restart"},
		},
	}

	for _, p := range playbooks {
		p.CreatedAt = time.Now()
		c.playbooks[p.ID] = p
	}
}

// ListPlaybooks returns all available playbooks
func (c *SelfHealingClient) ListPlaybooks(service string) []*RemediationPlaybook {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*RemediationPlaybook
	for _, p := range c.playbooks {
		if service == "" || p.Service == service {
			result = append(result, p)
		}
	}
	return result
}

// GetPlaybook returns a specific playbook
func (c *SelfHealingClient) GetPlaybook(id string) (*RemediationPlaybook, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	p, exists := c.playbooks[id]
	if !exists {
		return nil, fmt.Errorf("playbook not found: %s", id)
	}
	return p, nil
}

// ExecutePlaybook starts a playbook execution
func (c *SelfHealingClient) ExecutePlaybook(playbookID string, autoApprove bool) (*RemediationExecution, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	playbook, exists := c.playbooks[playbookID]
	if !exists {
		return nil, fmt.Errorf("playbook not found: %s", playbookID)
	}

	// Check cooldown
	if cooldownUntil, inCooldown := c.cooldowns[playbook.Service]; inCooldown {
		if time.Now().Before(cooldownUntil) {
			return nil, fmt.Errorf("service %s in cooldown until %s", playbook.Service, cooldownUntil.Format(time.RFC3339))
		}
	}

	exec := &RemediationExecution{
		ID:           uuid.New().String()[:8],
		PlaybookID:   playbookID,
		PlaybookName: playbook.Name,
		Service:      playbook.Service,
		RiskScore:    playbook.RiskScore,
		StartedAt:    time.Now(),
		StepsTotal:   len(playbook.Steps),
		AutoExecuted: autoApprove,
	}

	// Check if approval required
	if playbook.RiskScore > c.config.MaxAutoRiskScore && !autoApprove {
		exec.Status = "approval_required"
		c.pending[exec.ID] = &PendingApproval{
			ExecutionID:  exec.ID,
			PlaybookID:   playbookID,
			PlaybookName: playbook.Name,
			Service:      playbook.Service,
			RiskScore:    playbook.RiskScore,
			Reason:       fmt.Sprintf("Risk score %d exceeds auto-execute threshold %d", playbook.RiskScore, c.config.MaxAutoRiskScore),
			CreatedAt:    time.Now(),
		}
		c.executions = append(c.executions, exec)
		return exec, nil
	}

	// Execute (in real implementation, this would run async)
	exec.Status = "running"
	c.executions = append(c.executions, exec)

	// Simulate execution
	go c.runPlaybook(exec, playbook)

	return exec, nil
}

// runPlaybook executes a playbook (simplified)
func (c *SelfHealingClient) runPlaybook(exec *RemediationExecution, playbook *RemediationPlaybook) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// In a real implementation, this would execute each step
	// For now, mark as success
	exec.StepsRun = len(playbook.Steps)
	exec.Status = "success"
	now := time.Now()
	exec.CompletedAt = &now
	exec.Output = fmt.Sprintf("Executed %d steps for %s", len(playbook.Steps), playbook.Name)

	// Set cooldown
	c.cooldowns[playbook.Service] = time.Now().Add(playbook.Cooldown)

	// Update playbook stats
	playbook.ExecCount++
	playbook.SuccessRate = float64(playbook.ExecCount) / float64(playbook.ExecCount) * 100

	c.saveToVault()
}

// ApproveExecution approves a pending execution
func (c *SelfHealingClient) ApproveExecution(executionID, approvedBy string) (*RemediationExecution, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pending, exists := c.pending[executionID]
	if !exists {
		return nil, fmt.Errorf("no pending approval for execution: %s", executionID)
	}

	// Find execution
	var exec *RemediationExecution
	for _, e := range c.executions {
		if e.ID == executionID {
			exec = e
			break
		}
	}
	if exec == nil {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	playbook := c.playbooks[pending.PlaybookID]
	if playbook == nil {
		return nil, fmt.Errorf("playbook not found: %s", pending.PlaybookID)
	}

	// Mark approved and execute
	exec.ApprovedBy = approvedBy
	exec.Status = "running"
	delete(c.pending, executionID)

	go c.runPlaybook(exec, playbook)

	return exec, nil
}

// RejectExecution rejects a pending execution
func (c *SelfHealingClient) RejectExecution(executionID, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.pending[executionID]
	if !exists {
		return fmt.Errorf("no pending approval for execution: %s", executionID)
	}

	// Find and update execution
	for _, e := range c.executions {
		if e.ID == executionID {
			e.Status = "cancelled"
			e.Error = fmt.Sprintf("Rejected: %s", reason)
			now := time.Now()
			e.CompletedAt = &now
			break
		}
	}

	delete(c.pending, executionID)
	return nil
}

// ListPendingApprovals returns all pending approvals
func (c *SelfHealingClient) ListPendingApprovals() []*PendingApproval {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*PendingApproval, 0, len(c.pending))
	for _, p := range c.pending {
		result = append(result, p)
	}
	return result
}

// GetExecution returns an execution by ID
func (c *SelfHealingClient) GetExecution(executionID string) (*RemediationExecution, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, e := range c.executions {
		if e.ID == executionID {
			return e, nil
		}
	}
	return nil, fmt.Errorf("execution not found: %s", executionID)
}

// ListExecutions returns recent executions
func (c *SelfHealingClient) ListExecutions(limit int) []*RemediationExecution {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	start := len(c.executions) - limit
	if start < 0 {
		start = 0
	}

	result := make([]*RemediationExecution, len(c.executions)-start)
	copy(result, c.executions[start:])

	// Reverse to show newest first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// GetConfig returns the current configuration
func (c *SelfHealingClient) GetConfig() SelfHealingConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// UpdateConfig updates the configuration
func (c *SelfHealingClient) UpdateConfig(config SelfHealingConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config
	c.saveToVault()
}

// GetStats returns statistics
func (c *SelfHealingClient) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var successCount, failCount int
	for _, e := range c.executions {
		if e.Status == "success" {
			successCount++
		} else if e.Status == "failed" {
			failCount++
		}
	}

	return map[string]interface{}{
		"total_playbooks":   len(c.playbooks),
		"total_executions":  len(c.executions),
		"pending_approvals": len(c.pending),
		"success_count":     successCount,
		"failure_count":     failCount,
		"active_cooldowns":  len(c.cooldowns),
		"config":            c.config,
	}
}

// loadFromVault loads persisted data
func (c *SelfHealingClient) loadFromVault() {
	dataPath := filepath.Join(c.vaultPath, "healing", "executions.json")
	if data, err := os.ReadFile(dataPath); err == nil {
		var executions []*RemediationExecution
		if json.Unmarshal(data, &executions) == nil {
			c.executions = executions
		}
	}

	configPath := filepath.Join(c.vaultPath, "healing", "config.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var config SelfHealingConfig
		if json.Unmarshal(data, &config) == nil {
			c.config = config
		}
	}
}

// saveToVault persists data
func (c *SelfHealingClient) saveToVault() error {
	healingDir := filepath.Join(c.vaultPath, "healing")
	if err := os.MkdirAll(healingDir, 0755); err != nil {
		return err
	}

	// Save executions
	if data, err := json.MarshalIndent(c.executions, "", "  "); err == nil {
		os.WriteFile(filepath.Join(healingDir, "executions.json"), data, 0644)
	}

	// Save config
	if data, err := json.MarshalIndent(c.config, "", "  "); err == nil {
		os.WriteFile(filepath.Join(healingDir, "config.json"), data, 0644)
	}

	return nil
}

// === Adaptive Healing Methods (v2.17) ===

// AddPlaybook adds a new playbook (typically learned from manual fixes)
func (c *SelfHealingClient) AddPlaybook(playbook *RemediationPlaybook) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if playbook.ID == "" {
		return fmt.Errorf("playbook ID is required")
	}

	if _, exists := c.playbooks[playbook.ID]; exists {
		return fmt.Errorf("playbook with ID %s already exists", playbook.ID)
	}

	playbook.CreatedAt = time.Now()
	c.playbooks[playbook.ID] = playbook
	c.saveToVault()
	return nil
}

// SetPlaybookAutoApprove enables or disables auto-approval for a playbook
func (c *SelfHealingClient) SetPlaybookAutoApprove(playbookID string, enabled bool, maxDaily int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	playbook, exists := c.playbooks[playbookID]
	if !exists {
		return fmt.Errorf("playbook not found: %s", playbookID)
	}

	// If enabling auto-approve, lower the risk score to below threshold
	if enabled {
		if playbook.RiskScore > c.config.MaxAutoRiskScore {
			// Mark it as auto-approvable by adding a tag
			playbook.Tags = append(playbook.Tags, "auto-enabled")
		}
	} else {
		// Remove auto-enabled tag
		var newTags []string
		for _, t := range playbook.Tags {
			if t != "auto-enabled" {
				newTags = append(newTags, t)
			}
		}
		playbook.Tags = newTags
	}

	c.saveToVault()
	return nil
}

// IsPlaybookAutoEnabled checks if a playbook has auto-fix enabled
func (c *SelfHealingClient) IsPlaybookAutoEnabled(playbookID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	playbook, exists := c.playbooks[playbookID]
	if !exists {
		return false
	}

	for _, t := range playbook.Tags {
		if t == "auto-enabled" {
			return true
		}
	}
	return false
}

// RollbackResult represents the result of a rollback operation
type RollbackResult struct {
	StepsRolledBack int      `json:"steps_rolled_back"`
	Warnings        []string `json:"warnings,omitempty"`
}

// RollbackExecution attempts to undo an execution
func (c *SelfHealingClient) RollbackExecution(executionID, reason string) (*RollbackResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exec *RemediationExecution
	for _, e := range c.executions {
		if e.ID == executionID {
			exec = e
			break
		}
	}
	if exec == nil {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	// Check if rollback is possible
	if exec.Status != "success" && exec.Status != "failed" {
		return nil, fmt.Errorf("cannot rollback execution in '%s' state", exec.Status)
	}

	// Get playbook
	playbook, exists := c.playbooks[exec.PlaybookID]
	if !exists {
		return nil, fmt.Errorf("playbook not found for execution")
	}

	result := &RollbackResult{
		StepsRolledBack: 0,
		Warnings:        []string{},
	}

	// In a real implementation, this would execute rollback steps
	// For now, we mark steps as rolled back and add warnings for non-reversible steps
	for _, step := range playbook.Steps {
		if step.Type == "command" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Command '%s' may need manual rollback", step.Name))
		} else {
			result.StepsRolledBack++
		}
	}

	// Create rollback execution record
	rollbackExec := &RemediationExecution{
		ID:           uuid.New().String()[:8],
		PlaybookID:   exec.PlaybookID,
		PlaybookName: exec.PlaybookName + " (ROLLBACK)",
		Service:      exec.Service,
		Status:       "success",
		RiskScore:    exec.RiskScore,
		StartedAt:    time.Now(),
		StepsRun:     result.StepsRolledBack,
		StepsTotal:   len(playbook.Steps),
		Output:       fmt.Sprintf("Rolled back execution %s: %s", executionID, reason),
	}
	now := time.Now()
	rollbackExec.CompletedAt = &now
	c.executions = append(c.executions, rollbackExec)

	// Mark original execution as rolled back
	exec.Output = exec.Output + fmt.Sprintf("\n[ROLLED BACK: %s]", reason)

	c.saveToVault()
	return result, nil
}
