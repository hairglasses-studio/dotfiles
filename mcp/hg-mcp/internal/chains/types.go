// Package chains provides workflow chain execution for multi-step operations.
package chains

import (
	"time"
)

// ChainStatus represents the current state of a chain execution
type ChainStatus string

const (
	ChainStatusPending   ChainStatus = "pending"
	ChainStatusRunning   ChainStatus = "running"
	ChainStatusPaused    ChainStatus = "paused" // Waiting for gate approval
	ChainStatusCompleted ChainStatus = "completed"
	ChainStatusFailed    ChainStatus = "failed"
	ChainStatusCancelled ChainStatus = "cancelled"
)

// StepType defines the type of chain step
type StepType string

const (
	StepTypeTool     StepType = "tool"     // Execute a single tool
	StepTypeChain    StepType = "chain"    // Execute another chain
	StepTypeParallel StepType = "parallel" // Execute multiple steps in parallel
	StepTypeBranch   StepType = "branch"   // Conditional branching
	StepTypeGate     StepType = "gate"     // Requires approval to continue
	StepTypeDelay    StepType = "delay"    // Wait for duration
)

// OnErrorAction defines what to do when a step fails
type OnErrorAction string

const (
	OnErrorStop     OnErrorAction = "stop"     // Stop chain execution
	OnErrorContinue OnErrorAction = "continue" // Continue to next step
	OnErrorRetry    OnErrorAction = "retry"    // Retry the step
)

// TriggerType defines how a chain can be triggered
type TriggerType string

const (
	TriggerManual TriggerType = "manual" // Manual execution
	TriggerCron   TriggerType = "cron"   // Cron schedule
	TriggerEvent  TriggerType = "event"  // Event-driven
)

// Chain represents a workflow with multiple steps
type Chain struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Category    string           `json:"category"` // show, stream, backup, etc.
	Steps       []ChainStep      `json:"steps"`
	Parameters  []ChainParameter `json:"parameters"`
	Triggers    []ChainTrigger   `json:"triggers"`
	Timeout     time.Duration    `json:"timeout"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// ChainStep represents a single step in a chain
type ChainStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       StepType               `json:"type"`
	Tool       string                 `json:"tool,omitempty"`      // Tool name for StepTypeTool
	ChainID    string                 `json:"chain_id,omitempty"`  // Chain ID for StepTypeChain
	Inputs     map[string]interface{} `json:"inputs,omitempty"`    // Input parameters
	Condition  string                 `json:"condition,omitempty"` // Condition expression for branches
	OnError    OnErrorAction          `json:"on_error"`
	MaxRetries int                    `json:"max_retries,omitempty"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
	DelayAfter time.Duration          `json:"delay_after,omitempty"` // Delay after step completes

	// For parallel steps
	ParallelSteps []ChainStep `json:"parallel_steps,omitempty"`

	// For branch steps
	Branches []BranchCase `json:"branches,omitempty"`

	// For gate steps
	GateMessage       string        `json:"gate_message,omitempty"`
	GateApprovers     []string      `json:"gate_approvers,omitempty"`
	GateTimeout       time.Duration `json:"gate_timeout,omitempty"`        // Auto-resolve after this duration (0 = wait forever)
	GateDefaultAction string        `json:"gate_default_action,omitempty"` // "approve" or "reject" on timeout (default: reject)
}

// BranchCase represents a conditional branch
type BranchCase struct {
	Condition string      `json:"condition"` // Expression to evaluate
	Steps     []ChainStep `json:"steps"`     // Steps to execute if condition is true
}

// ChainParameter defines an input parameter for a chain
type ChainParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // string, int, bool, etc.
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// ChainTrigger defines how a chain can be triggered
type ChainTrigger struct {
	Type     TriggerType `json:"type"`
	Schedule string      `json:"schedule,omitempty"` // Cron expression for TriggerCron
	Event    string      `json:"event,omitempty"`    // Event name for TriggerEvent
	Enabled  bool        `json:"enabled"`
}

// ChainExecution represents a running or completed chain execution
type ChainExecution struct {
	ID          string                 `json:"id"`
	ChainID     string                 `json:"chain_id"`
	ChainName   string                 `json:"chain_name"`
	Status      ChainStatus            `json:"status"`
	Parameters  map[string]interface{} `json:"parameters"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	CurrentStep int                    `json:"current_step"`
	TotalSteps  int                    `json:"total_steps"`
	StepResults []StepResult           `json:"step_results"`
	Error       string                 `json:"error,omitempty"`
	TriggeredBy string                 `json:"triggered_by"` // user, cron, event
	DryRun      bool                   `json:"dry_run,omitempty"`
}

// StepResult represents the result of executing a step
type StepResult struct {
	StepID      string                 `json:"step_id"`
	StepName    string                 `json:"step_name"`
	Status      ChainStatus            `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Retries     int                    `json:"retries"`
}

// GateApproval represents an approval for a gate step
type GateApproval struct {
	ExecutionID string    `json:"execution_id"`
	StepID      string    `json:"step_id"`
	Approved    bool      `json:"approved"`
	ApprovedBy  string    `json:"approved_by"`
	ApprovedAt  time.Time `json:"approved_at"`
	Comment     string    `json:"comment,omitempty"`
}

// PendingGate represents a gate waiting for approval
type PendingGate struct {
	ExecutionID string    `json:"execution_id"`
	ChainName   string    `json:"chain_name"`
	StepID      string    `json:"step_id"`
	StepName    string    `json:"step_name"`
	Message     string    `json:"message"`
	Approvers   []string  `json:"approvers"`
	CreatedAt   time.Time `json:"created_at"`
}
