// Command ralph runs autonomous DRY refactoring cycles using mcpkit's ralph.Loop.
//
// Each cycle: pick next task from dry_tasks.md → read context → refactor file →
// verify (build/vet/test) → mark done. After each cycle, commits changes and
// optionally pushes. Runs until budget, duration, or task queue is exhausted.
//
// Required: ANTHROPIC_API_KEY environment variable.
//
// Optional env vars:
//
//	RALPH_DURATION   — max runtime (default: 24h)
//	RALPH_BUDGET     — total $ budget (default: 100.0)
//	RALPH_MODEL      — Claude model (default: claude-sonnet-4-6)
//	RALPH_SPEC       — spec file (default: .ralph/specs/dry_refactor.json)
//	RALPH_STATE      — state file (default: .ralph/.ralph_state.json)
//	RALPH_AUTO_PUSH  — push every N commits (default: 10, 0=disabled)
//	RALPH_MAX_CYCLES — max cycles / DRY tasks (default: 0=unlimited)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hairglasses-studio/mcpkit/finops"
	"github.com/hairglasses-studio/mcpkit/ralph"
	"github.com/hairglasses-studio/mcpkit/registry"
	"github.com/hairglasses-studio/mcpkit/sampling"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmsgprefix)
	log.SetPrefix("[ralph] ")

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "ANTHROPIC_API_KEY is required")
		os.Exit(1)
	}

	cfg := loadConfig()

	// Resolve project root.
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Fatalf("find project root: %v", err)
	}

	sampler := &sampling.APISamplingClient{
		APIKey:       apiKey,
		DefaultModel: cfg.Model,
		HTTPClient:   &http.Client{Timeout: 10 * time.Minute},
	}

	runner := &Runner{
		cfg:         cfg,
		projectRoot: projectRoot,
		sampler:     sampler,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Printf("starting (model=%s, budget=$%.0f, duration=%s, spec=%s, auto_push=%d)",
		cfg.Model, cfg.Budget, cfg.Duration, cfg.SpecFile, cfg.AutoPushInterval)

	go func() {
		<-ctx.Done()
		log.Println("signal received, finishing current cycle...")
		runner.Stop()
	}()

	if err := runner.Run(ctx); err != nil && err != context.Canceled {
		log.Printf("runner exited: %v", err)
	}

	// Final push on exit.
	if runner.commitsSincePush > 0 && cfg.AutoPushInterval > 0 {
		log.Printf("final push: %d unpushed commits", runner.commitsSincePush)
		runner.gitPush()
	}

	state := runner.State()
	log.Printf("=== Final Summary ===")
	log.Printf("  Cycles completed: %d", state.CyclesCompleted)
	log.Printf("  Total commits:    %d", state.TotalCommits)
	log.Printf("  Total cost:       $%.4f", state.TotalCost)
	log.Printf("  Runtime:          %s", time.Since(state.StartedAt).Truncate(time.Second))
}

// Config holds all configuration for the ralph runner.
type Config struct {
	Model            string
	Duration         time.Duration
	Budget           float64
	SpecFile         string
	StatePath        string
	AutoPushInterval int
	MaxCycles        int
	MaxIterPerCycle  int
}

func loadConfig() Config {
	return Config{
		Model:            envOr("RALPH_MODEL", "claude-sonnet-4-6"),
		Duration:         envDuration("RALPH_DURATION", 24*time.Hour),
		Budget:           envFloat("RALPH_BUDGET", 100.0),
		SpecFile:         envOr("RALPH_SPEC", ".ralph/specs/dry_refactor.json"),
		StatePath:        envOr("RALPH_STATE", ".ralph/.ralph_state.json"),
		AutoPushInterval: envInt("RALPH_AUTO_PUSH", 10),
		MaxCycles:        envInt("RALPH_MAX_CYCLES", 0),
		MaxIterPerCycle:  envInt("RALPH_MAX_ITER", 50),
	}
}

// RunnerState persists across crashes for resumability.
type RunnerState struct {
	CyclesCompleted int       `json:"cycles_completed"`
	TotalCommits    int       `json:"total_commits"`
	TotalCost       float64   `json:"total_cost"`
	StartedAt       time.Time `json:"started_at"`
	LastCycleAt     time.Time `json:"last_cycle_at"`
	ConsecutiveFails int      `json:"consecutive_fails"`
}

// Runner manages the multi-cycle DRY refactoring loop.
type Runner struct {
	cfg         Config
	projectRoot string
	sampler     sampling.SamplingClient
	mu          sync.Mutex
	state       RunnerState
	stopCh      chan struct{}
	stopped     bool
	commitsSincePush int
}

func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.stopped {
		r.stopped = true
		if r.stopCh != nil {
			close(r.stopCh)
		}
	}
}

func (r *Runner) State() RunnerState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}

// Run executes DRY refactoring cycles until budget, duration, task queue, or signal.
func (r *Runner) Run(ctx context.Context) error {
	r.stopCh = make(chan struct{})

	// Load existing state for resumability.
	state, err := loadState(r.cfg.StatePath)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	r.mu.Lock()
	r.state = state
	if r.state.StartedAt.IsZero() {
		r.state.StartedAt = time.Now()
	}
	r.mu.Unlock()

	deadline := r.state.StartedAt.Add(r.cfg.Duration)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-r.stopCh:
			log.Println("stop signal received")
			return nil
		default:
		}

		if time.Now().After(deadline) {
			log.Println("duration limit reached")
			return nil
		}

		r.mu.Lock()
		if r.state.TotalCost >= r.cfg.Budget {
			r.mu.Unlock()
			log.Printf("budget exhausted ($%.2f >= $%.2f)", r.state.TotalCost, r.cfg.Budget)
			return nil
		}
		if r.cfg.MaxCycles > 0 && r.state.CyclesCompleted >= r.cfg.MaxCycles {
			r.mu.Unlock()
			log.Printf("max cycles reached (%d)", r.cfg.MaxCycles)
			return nil
		}
		cycleNum := r.state.CyclesCompleted + 1
		r.mu.Unlock()

		// Check if tasks remain.
		if !r.hasPendingTasks() {
			log.Println("all tasks complete — no unchecked items in dry_tasks.md")
			return nil
		}

		// Pause between cycles.
		if cycleNum > 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(5 * time.Second):
			}
		}

		log.Printf("=== Cycle %d starting ===", cycleNum)
		cycleStart := time.Now()
		headBefore := r.gitHead()

		cycleCost, err := r.runOneCycle(ctx, cycleNum)
		cycleDuration := time.Since(cycleStart)

		// Detect commit.
		headAfter := r.gitHead()
		hadCommit := headBefore != headAfter && headAfter != ""

		if !hadCommit && err == nil {
			// No commit but no error — git add + commit ourselves if there are changes.
			if r.hasUncommittedChanges() {
				if commitErr := r.gitCommitChanges(cycleNum); commitErr == nil {
					hadCommit = true
				}
			}
		}

		r.mu.Lock()
		r.state.TotalCost += cycleCost
		r.state.LastCycleAt = time.Now()

		if err != nil {
			r.state.ConsecutiveFails++
			log.Printf("cycle %d failed after %s: %v (consecutive fails: %d)",
				cycleNum, cycleDuration.Truncate(time.Second), err, r.state.ConsecutiveFails)
		} else {
			r.state.CyclesCompleted++
			r.state.ConsecutiveFails = 0
			if hadCommit {
				r.state.TotalCommits++
				r.commitsSincePush++
			}
			log.Printf("cycle %d done in %s (cost=$%.4f, commit=%v)",
				cycleNum, cycleDuration.Truncate(time.Second), cycleCost, hadCommit)
		}
		r.mu.Unlock()

		// Save state after each cycle.
		if saveErr := saveState(r.cfg.StatePath, r.state); saveErr != nil {
			log.Printf("warning: failed to save state: %v", saveErr)
		}

		// Auto-push.
		if r.cfg.AutoPushInterval > 0 && r.commitsSincePush >= r.cfg.AutoPushInterval {
			r.gitPush()
		}

		// Back off on consecutive failures.
		if r.state.ConsecutiveFails >= 5 {
			log.Printf("%d consecutive failures, stopping", r.state.ConsecutiveFails)
			return fmt.Errorf("%d consecutive cycle failures", r.state.ConsecutiveFails)
		}
		if r.state.ConsecutiveFails > 0 {
			backoff := time.Duration(r.state.ConsecutiveFails) * 30 * time.Second
			if backoff > 5*time.Minute {
				backoff = 5 * time.Minute
			}
			log.Printf("backing off %s before next cycle", backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
}

// runOneCycle executes a single DRY refactoring cycle via ralph.Loop.
func (r *Runner) runOneCycle(ctx context.Context, cycleNum int) (float64, error) {
	reg := registry.NewToolRegistry()

	// Register file tools pointed at hg-mcp project root.
	fileMod := &ralph.FileToolModule{Root: r.projectRoot}
	reg.RegisterModule(fileMod)

	// Per-cycle cost tracking.
	tracker, _, _ := buildFinOpsStack(r.cfg.Budget - r.state.TotalCost)

	// Circuit breaker: trip after 5 no-progress iterations.
	cb := ralph.NewCircuitBreaker(ralph.CircuitBreakerConfig{
		NoProgressThreshold: 5,
		SameErrorThreshold:  3,
		CooldownDuration:    5 * time.Minute,
	})

	// Cost governor: per-cycle budget, halt after unproductive streaks.
	cg := ralph.NewCostGovernor(ralph.CostGovernorConfig{
		HardBudgetTokens:  500_000, // ~$0.75 per cycle at sonnet pricing
		VelocityWindow:    5,
		VelocityAlarmRate: 0.6,
		UnproductiveMax:   5,
	})

	progressFile := fmt.Sprintf(".ralph/.cycle_%d.progress.json", cycleNum)
	var totalCost float64

	loopCfg := ralph.Config{
		SpecFile:        r.cfg.SpecFile,
		ProgressFile:    progressFile,
		ToolRegistry:    reg,
		Sampler:         r.sampler,
		MaxIterations:   r.cfg.MaxIterPerCycle,
		MaxTokens:       8192,
		CostTracker:     tracker,
		ForceRestart:    true,
		CircuitBreaker:  cb,
		CostGovernor:    cg,
		ExitGate:        ralph.ExitGate{RequireAllTasksDone: true},
		HistoryWindow:   5,
		AutoVerifyLevel: ralph.AutoVerifyFull,
		ProjectRoot:     r.projectRoot,
		StuckThreshold:  3,
		PhaseMaxTokens: map[string]int{
			"pick_task":    2048,
			"read_context": 4096,
			"refactor":     16384,
			"verify":       4096,
			"mark_done":    2048,
		},
		Hooks: ralph.Hooks{
			OnIterationStart: func(iter int) {
				log.Printf("  [cycle %d] iteration %d", cycleNum, iter)
			},
			OnIterationEnd: func(entry ralph.IterationLog) {
				taskInfo := ""
				if entry.TaskID != "" {
					taskInfo = fmt.Sprintf(" task=%s", entry.TaskID)
				}
				log.Printf("  [cycle %d] iter %d done%s: %s",
					cycleNum, entry.Iteration, taskInfo, truncate(entry.Result, 100))
			},
			OnTaskComplete: func(taskID string) {
				log.Printf("  [cycle %d] task %q completed", cycleNum, taskID)
			},
			OnError: func(iter int, err error) {
				log.Printf("  [cycle %d] iter %d error: %v", cycleNum, iter, err)
			},
			OnCostUpdate: func(iter int, us finops.UsageSummary) {
				cost := estimateCost(us)
				totalCost = cost
			},
		},
	}

	loop, err := ralph.NewLoop(loopCfg)
	if err != nil {
		return 0, fmt.Errorf("create loop: %w", err)
	}

	err = loop.Run(ctx)

	// Clean up progress file.
	os.Remove(filepath.Join(r.projectRoot, progressFile))

	return totalCost, err
}

// hasPendingTasks checks if dry_tasks.md has any unchecked items.
func (r *Runner) hasPendingTasks() bool {
	path := filepath.Join(r.projectRoot, ".ralph/dry_tasks.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "- [ ]")
}

// Git helpers.

func (r *Runner) gitHead() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.projectRoot
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (r *Runner) hasUncommittedChanges() bool {
	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = r.projectRoot
	out, _ := cmd.Output()
	return len(strings.TrimSpace(string(out))) > 0
}

func (r *Runner) gitCommitChanges(cycleNum int) error {
	// Stage only Go source files and dry_tasks.md.
	addCmd := exec.Command("git", "add", "-u", "--", "*.go", ".ralph/dry_tasks.md")
	addCmd.Dir = r.projectRoot
	if err := addCmd.Run(); err != nil {
		// Fallback: add all tracked changes.
		addCmd2 := exec.Command("git", "add", "-u")
		addCmd2.Dir = r.projectRoot
		addCmd2.Run()
	}

	msg := fmt.Sprintf("refactor(dry): ralph cycle %d — automated DRY refactoring", cycleNum)
	commitCmd := exec.Command("git", "commit", "-m", msg)
	commitCmd.Dir = r.projectRoot
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		log.Printf("git commit failed: %s", strings.TrimSpace(string(out)))
		return err
	}
	log.Printf("committed: %s", msg)
	return nil
}

func (r *Runner) gitPush() {
	log.Printf("pushing %d commits...", r.commitsSincePush)
	cmd := exec.Command("git", "push")
	cmd.Dir = r.projectRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("git push failed: %s", strings.TrimSpace(string(out)))
		return
	}
	log.Printf("pushed successfully")
	r.commitsSincePush = 0
}

// State persistence.

func loadState(path string) (RunnerState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return RunnerState{}, nil
		}
		return RunnerState{}, err
	}
	var s RunnerState
	if err := json.Unmarshal(data, &s); err != nil {
		return RunnerState{}, err
	}
	return s, nil
}

func saveState(path string, s RunnerState) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		os.MkdirAll(dir, 0755)
	}
	tmp, err := os.CreateTemp(dir, ".ralph-state-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	tmp.Close()
	return os.Rename(tmpName, path)
}

// Cost estimation (sonnet pricing as conservative baseline).
func estimateCost(us finops.UsageSummary) float64 {
	return float64(us.TotalInputTokens)/1000*0.003 + float64(us.TotalOutputTokens)/1000*0.015
}

func buildFinOpsStack(remainingBudget float64) (*finops.Tracker, interface{}, interface{}) {
	tracker := finops.NewTracker()
	return tracker, nil, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func findProjectRoot() (string, error) {
	// Try current directory first.
	if _, err := os.Stat("go.mod"); err == nil {
		return ".", nil
	}
	// Try home directory path.
	if home := os.Getenv("HOME"); home != "" {
		p := filepath.Join(home, "hairglasses-studio", "hg-mcp")
		if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("cannot find hg-mcp project root (no go.mod found)")
}

// Env helpers.

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("warning: invalid %s=%q, using default %s", key, v, fallback)
		return fallback
	}
	return d
}

func envFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
