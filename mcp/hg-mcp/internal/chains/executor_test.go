package chains

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// newTestExecutor creates an executor with a registry containing fake tools.
func newTestExecutor(t *testing.T, fakeTools map[string]tools.ToolHandlerFunc) *Executor {
	t.Helper()
	reg := tools.NewToolRegistry()
	for name, handler := range fakeTools {
		reg.RegisterModule(&fakeModule{name: name, handler: handler})
	}
	return NewExecutor(t.TempDir(), reg)
}

// fakeModule is a minimal module that registers a single tool.
type fakeModule struct {
	name    string
	handler tools.ToolHandlerFunc
}

func (m *fakeModule) Name() string        { return m.name }
func (m *fakeModule) Description() string { return "test" }
func (m *fakeModule) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool:    mcp.Tool{Name: m.name},
			Handler: m.handler,
		},
	}
}

// okHandler returns a successful tool result with the given text output.
func okHandler(text string) tools.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: text}},
		}, nil
	}
}

// failHandler returns an error result.
func failHandler(msg string) tools.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{mcp.TextContent{Type: "text", Text: msg}},
		}, nil
	}
}

// waitForExecution polls until the execution reaches a terminal or expected status.
func waitForExecution(t *testing.T, e *Executor, execID string, timeout time.Duration, statuses ...ChainStatus) *ChainExecution {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			exec, _ := e.GetExecution(execID)
			t.Fatalf("timeout waiting for execution %s; current status: %s", execID, exec.Status)
			return nil
		default:
			exec, ok := e.GetExecution(execID)
			if !ok {
				t.Fatalf("execution %s not found", execID)
			}
			for _, s := range statuses {
				if exec.Status == s {
					return exec
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestSequentialExecution(t *testing.T) {
	var order []string
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"step_a": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			order = append(order, "a")
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "ok_a"}}}, nil
		},
		"step_b": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			order = append(order, "b")
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "ok_b"}}}, nil
		},
	})

	chain := &Chain{
		ID:   "seq_test",
		Name: "Sequential Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Step A", Type: StepTypeTool, Tool: "step_a", OnError: OnErrorStop},
			{ID: "2", Name: "Step B", Type: StepTypeTool, Tool: "step_b", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "seq_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed, got %s (error: %s)", final.Status, final.Error)
	}
	if len(order) != 2 || order[0] != "a" || order[1] != "b" {
		t.Fatalf("expected [a b], got %v", order)
	}
	if len(final.StepResults) != 2 {
		t.Fatalf("expected 2 step results, got %d", len(final.StepResults))
	}
}

func TestParallelSteps(t *testing.T) {
	var count atomic.Int32
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"par_a": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			count.Add(1)
			time.Sleep(50 * time.Millisecond)
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "a"}}}, nil
		},
		"par_b": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			count.Add(1)
			time.Sleep(50 * time.Millisecond)
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "b"}}}, nil
		},
	})

	chain := &Chain{
		ID:   "par_test",
		Name: "Parallel Test",
		Steps: []ChainStep{
			{
				ID: "1", Name: "Parallel", Type: StepTypeParallel, OnError: OnErrorStop,
				ParallelSteps: []ChainStep{
					{ID: "1a", Name: "Par A", Type: StepTypeTool, Tool: "par_a", OnError: OnErrorStop},
					{ID: "1b", Name: "Par B", Type: StepTypeTool, Tool: "par_b", OnError: OnErrorStop},
				},
			},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "par_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed, got %s (error: %s)", final.Status, final.Error)
	}
	if count.Load() != 2 {
		t.Fatalf("expected 2 parallel tasks to run, got %d", count.Load())
	}
}

func TestGateApprove(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"post_gate": okHandler("after_gate"),
	})

	chain := &Chain{
		ID:   "gate_test",
		Name: "Gate Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Gate", Type: StepTypeGate, GateMessage: "Approve?", OnError: OnErrorStop},
			{ID: "2", Name: "After Gate", Type: StepTypeTool, Tool: "post_gate", OnError: OnErrorStop},
		},
		Timeout: 10 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "gate_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Wait for pause
	paused := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusPaused)
	if paused.Status != ChainStatusPaused {
		t.Fatalf("expected paused, got %s", paused.Status)
	}

	// Check pending gates
	gates := exec.ListPendingGates()
	if len(gates) != 1 {
		t.Fatalf("expected 1 pending gate, got %d", len(gates))
	}

	// Approve
	if err := exec.ApproveGate(context.Background(), result.ID, true, "tester", "approved"); err != nil {
		t.Fatalf("ApproveGate: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed after approval, got %s (error: %s)", final.Status, final.Error)
	}
	if len(final.StepResults) != 2 {
		t.Fatalf("expected 2 step results, got %d", len(final.StepResults))
	}
}

func TestGateReject(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"should_not_run": okHandler("nope"),
	})

	chain := &Chain{
		ID:   "gate_reject",
		Name: "Gate Reject",
		Steps: []ChainStep{
			{ID: "1", Name: "Gate", Type: StepTypeGate, GateMessage: "Approve?", OnError: OnErrorStop},
			{ID: "2", Name: "After", Type: StepTypeTool, Tool: "should_not_run", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "gate_reject", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusPaused)

	// Reject
	if err := exec.ApproveGate(context.Background(), result.ID, false, "tester", "no"); err != nil {
		t.Fatalf("ApproveGate: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCancelled)
	if final.Status != ChainStatusCancelled {
		t.Fatalf("expected cancelled, got %s", final.Status)
	}
}

func TestRetryOnError(t *testing.T) {
	var attempts atomic.Int32
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"flaky": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			n := attempts.Add(1)
			if n < 3 {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{mcp.TextContent{Type: "text", Text: fmt.Sprintf("fail %d", n)}},
				}, nil
			}
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "ok"}}}, nil
		},
	})

	chain := &Chain{
		ID:   "retry_test",
		Name: "Retry Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Flaky Step", Type: StepTypeTool, Tool: "flaky", OnError: OnErrorRetry, MaxRetries: 5},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "retry_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed after retry, got %s (error: %s)", final.Status, final.Error)
	}
	// Initial attempt + 2 retries = 3 total
	if attempts.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestParamSubstitution(t *testing.T) {
	var capturedArgs map[string]interface{}
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"echo_tool": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if args, ok := req.Params.Arguments.(map[string]interface{}); ok {
				capturedArgs = args
			}
			msg, _ := capturedArgs["message"].(string)
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: msg}}}, nil
		},
	})

	chain := &Chain{
		ID:   "param_test",
		Name: "Param Test",
		Parameters: []ChainParameter{
			{Name: "greeting", Type: "string", Required: true},
		},
		Steps: []ChainStep{
			{
				ID: "1", Name: "Echo", Type: StepTypeTool, Tool: "echo_tool",
				Inputs:  map[string]interface{}{"message": "Hello, {{greeting}}!"},
				OnError: OnErrorStop,
			},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "param_test", map[string]interface{}{"greeting": "World"}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed, got %s", final.Status)
	}
	if capturedArgs["message"] != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got %q", capturedArgs["message"])
	}
}

func TestChainTimeout(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"slow_tool": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(5 * time.Second):
				return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "done"}}}, nil
			}
		},
	})

	chain := &Chain{
		ID:   "timeout_test",
		Name: "Timeout Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Slow", Type: StepTypeTool, Tool: "slow_tool", OnError: OnErrorStop},
		},
		Timeout: 200 * time.Millisecond,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "timeout_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusFailed)
	if final.Status != ChainStatusFailed {
		t.Fatalf("expected failed from timeout, got %s", final.Status)
	}
}

func TestCancelExecution(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"blocker": okHandler("ok"),
	})

	chain := &Chain{
		ID:   "cancel_test",
		Name: "Cancel Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Gate", Type: StepTypeGate, GateMessage: "Wait...", OnError: OnErrorStop},
			{ID: "2", Name: "After", Type: StepTypeTool, Tool: "blocker", OnError: OnErrorStop},
		},
		Timeout: 10 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "cancel_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusPaused)

	if err := exec.CancelExecution(result.ID); err != nil {
		t.Fatalf("CancelExecution: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 1*time.Second, ChainStatusCancelled)
	if final.Status != ChainStatusCancelled {
		t.Fatalf("expected cancelled, got %s", final.Status)
	}
}

func TestOnErrorContinue(t *testing.T) {
	var ranSecond bool
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"fail_tool": failHandler("intentional failure"),
		"success_tool": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ranSecond = true
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "ok"}}}, nil
		},
	})

	chain := &Chain{
		ID:   "continue_test",
		Name: "Continue Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Failing", Type: StepTypeTool, Tool: "fail_tool", OnError: OnErrorContinue},
			{ID: "2", Name: "Success", Type: StepTypeTool, Tool: "success_tool", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "continue_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed despite first step failure, got %s", final.Status)
	}
	if !ranSecond {
		t.Fatal("second step should have run after first step failure with OnErrorContinue")
	}
}

func TestDelayStep(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"after_delay": okHandler("done"),
	})

	chain := &Chain{
		ID:   "delay_test",
		Name: "Delay Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Wait", Type: StepTypeDelay, Inputs: map[string]interface{}{"duration": 50 * time.Millisecond}, OnError: OnErrorStop},
			{ID: "2", Name: "After", Type: StepTypeTool, Tool: "after_delay", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	start := time.Now()
	result, err := exec.Execute(context.Background(), "delay_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed, got %s", final.Status)
	}
	elapsed := time.Since(start)
	if elapsed < 40*time.Millisecond {
		t.Fatalf("delay step should have waited at least 40ms, elapsed %v", elapsed)
	}
}

func TestBranchStepMatchFirst(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"check_format": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "mp4"}},
			}, nil
		},
		"process_mp4":   okHandler("processed_mp4"),
		"process_mkv":   okHandler("processed_mkv"),
		"process_other": okHandler("processed_other"),
	})

	chain := &Chain{
		ID:   "branch_test",
		Name: "Branch Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Check Format", Type: StepTypeTool, Tool: "check_format", OnError: OnErrorStop},
			{
				ID: "2", Name: "Format Branch", Type: StepTypeBranch, OnError: OnErrorStop,
				Branches: []BranchCase{
					{Condition: "steps.1.text == mp4", Steps: []ChainStep{
						{ID: "2a", Name: "Process MP4", Type: StepTypeTool, Tool: "process_mp4", OnError: OnErrorStop},
					}},
					{Condition: "steps.1.text == mkv", Steps: []ChainStep{
						{ID: "2b", Name: "Process MKV", Type: StepTypeTool, Tool: "process_mkv", OnError: OnErrorStop},
					}},
					{Condition: "true", Steps: []ChainStep{
						{ID: "2c", Name: "Process Other", Type: StepTypeTool, Tool: "process_other", OnError: OnErrorStop},
					}},
				},
			},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "branch_test", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed, got %s (error: %s)", final.Status, final.Error)
	}

	// Should have 3 results: check_format, process_mp4 (from branch), and the branch step itself
	// The branch step appends sub-step results + its own result
	found := false
	for _, sr := range final.StepResults {
		if sr.StepName == "Process MP4" && sr.Status == ChainStatusCompleted {
			found = true
		}
		if sr.StepName == "Process MKV" {
			t.Fatal("MKV branch should not have executed")
		}
	}
	if !found {
		t.Fatal("expected Process MP4 step to have run")
	}
}

func TestBranchStepNoMatch(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"check": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "unknown"}}}, nil
		},
		"branch_a": okHandler("a"),
	})

	chain := &Chain{
		ID:   "no_match",
		Name: "No Match",
		Steps: []ChainStep{
			{ID: "1", Name: "Check", Type: StepTypeTool, Tool: "check", OnError: OnErrorStop},
			{
				ID: "2", Name: "Branch", Type: StepTypeBranch, OnError: OnErrorStop,
				Branches: []BranchCase{
					{Condition: "steps.1.text == specific_value", Steps: []ChainStep{
						{ID: "2a", Name: "A", Type: StepTypeTool, Tool: "branch_a", OnError: OnErrorStop},
					}},
				},
			},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "no_match", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed (no match should still complete), got %s", final.Status)
	}
	// Branch step result should indicate no match
	branchResult := final.StepResults[len(final.StepResults)-1]
	if branchResult.Output["branch"] != "none" {
		t.Fatalf("expected branch=none, got %v", branchResult.Output["branch"])
	}
}

func TestStepOutputChaining(t *testing.T) {
	var capturedMessage string
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"producer": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "hello_from_step1"}}}, nil
		},
		"consumer": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if args, ok := req.Params.Arguments.(map[string]interface{}); ok {
				capturedMessage, _ = args["msg"].(string)
			}
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "done"}}}, nil
		},
	})

	chain := &Chain{
		ID:   "output_chain",
		Name: "Output Chain",
		Steps: []ChainStep{
			{ID: "producer", Name: "Producer", Type: StepTypeTool, Tool: "producer", OnError: OnErrorStop},
			{
				ID: "consumer", Name: "Consumer", Type: StepTypeTool, Tool: "consumer",
				Inputs:  map[string]interface{}{"msg": "Got: {{steps.producer.text}}"},
				OnError: OnErrorStop,
			},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "output_chain", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed, got %s (error: %s)", final.Status, final.Error)
	}
	if capturedMessage != "Got: hello_from_step1" {
		t.Fatalf("expected 'Got: hello_from_step1', got %q", capturedMessage)
	}
}

func TestChainStepExecution(t *testing.T) {
	var ranInner bool
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"outer_step": okHandler("outer_ok"),
		"inner_step": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ranInner = true
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "inner_ok"}}}, nil
		},
	})

	// Save the inner chain first
	innerChain := &Chain{
		ID:   "inner_chain",
		Name: "Inner Chain",
		Steps: []ChainStep{
			{ID: "i1", Name: "Inner Step", Type: StepTypeTool, Tool: "inner_step", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(innerChain)

	// Outer chain references inner chain via StepTypeChain
	outerChain := &Chain{
		ID:   "outer_chain",
		Name: "Outer Chain",
		Steps: []ChainStep{
			{ID: "1", Name: "Outer Step", Type: StepTypeTool, Tool: "outer_step", OnError: OnErrorStop},
			{ID: "2", Name: "Run Inner", Type: StepTypeChain, ChainID: "inner_chain", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(outerChain)

	result, err := exec.Execute(context.Background(), "outer_chain", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted, ChainStatusFailed)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed, got %s (error: %s)", final.Status, final.Error)
	}
	if !ranInner {
		t.Fatal("inner chain step should have been executed")
	}
}

func TestGateTimeoutReject(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"after_gate": okHandler("should_not_run"),
	})

	chain := &Chain{
		ID:   "gate_timeout_reject",
		Name: "Gate Timeout Reject",
		Steps: []ChainStep{
			{
				ID: "1", Name: "Gate", Type: StepTypeGate,
				GateMessage:       "Approve?",
				GateTimeout:       100 * time.Millisecond,
				GateDefaultAction: "reject",
				OnError:           OnErrorStop,
			},
			{ID: "2", Name: "After", Type: StepTypeTool, Tool: "after_gate", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "gate_timeout_reject", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Should auto-reject after 100ms
	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCancelled)
	if final.Status != ChainStatusCancelled {
		t.Fatalf("expected cancelled from gate timeout reject, got %s", final.Status)
	}
}

func TestGateTimeoutApprove(t *testing.T) {
	var ranAfter bool
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"post_gate": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			ranAfter = true
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "ok"}}}, nil
		},
	})

	chain := &Chain{
		ID:   "gate_timeout_approve",
		Name: "Gate Timeout Approve",
		Steps: []ChainStep{
			{
				ID: "1", Name: "Gate", Type: StepTypeGate,
				GateMessage:       "Approve?",
				GateTimeout:       100 * time.Millisecond,
				GateDefaultAction: "approve",
				OnError:           OnErrorStop,
			},
			{ID: "2", Name: "After", Type: StepTypeTool, Tool: "post_gate", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "gate_timeout_approve", map[string]interface{}{}, "test")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Should auto-approve and continue to next step
	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed from gate timeout approve, got %s (error: %s)", final.Status, final.Error)
	}
	if !ranAfter {
		t.Fatal("step after gate should have run after auto-approve")
	}
}

func TestMissingRequiredParam(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"noop": okHandler("ok"),
	})

	chain := &Chain{
		ID:   "param_required",
		Name: "Param Required",
		Parameters: []ChainParameter{
			{Name: "required_param", Type: "string", Required: true},
		},
		Steps: []ChainStep{
			{ID: "1", Name: "Step", Type: StepTypeTool, Tool: "noop", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	_, err := exec.Execute(context.Background(), "param_required", map[string]interface{}{}, "test")
	if err == nil {
		t.Fatal("expected error for missing required parameter")
	}
}

func TestDryRunToolStep(t *testing.T) {
	var toolCalled bool
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{
		"noop": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			toolCalled = true
			return &mcp.CallToolResult{Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "executed"}}}, nil
		},
	})

	chain := &Chain{
		ID:   "dryrun_test",
		Name: "Dry Run Test",
		Steps: []ChainStep{
			{ID: "1", Name: "Step 1", Type: StepTypeTool, Tool: "noop", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	// Execute with dry-run = true
	result, err := exec.Execute(context.Background(), "dryrun_test", map[string]interface{}{}, "test", true)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed, got %s", final.Status)
	}
	if !final.DryRun {
		t.Fatal("expected DryRun=true on execution")
	}
	if toolCalled {
		t.Fatal("tool should NOT have been called in dry-run mode")
	}
	if len(final.StepResults) != 1 {
		t.Fatalf("expected 1 step result, got %d", len(final.StepResults))
	}
	if final.StepResults[0].Output["dry_run"] != true {
		t.Fatal("expected dry_run=true in step output")
	}
}

func TestDryRunGateStep(t *testing.T) {
	exec := newTestExecutor(t, map[string]tools.ToolHandlerFunc{})

	chain := &Chain{
		ID:   "dryrun_gate",
		Name: "Dry Run Gate",
		Steps: []ChainStep{
			{ID: "1", Name: "Approval", Type: StepTypeGate, GateMessage: "Approve?", OnError: OnErrorStop},
		},
		Timeout: 5 * time.Second,
	}
	_ = exec.SaveChain(chain)

	result, err := exec.Execute(context.Background(), "dryrun_gate", map[string]interface{}{}, "test", true)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	final := waitForExecution(t, exec, result.ID, 3*time.Second, ChainStatusCompleted)
	if final.Status != ChainStatusCompleted {
		t.Fatalf("expected completed (auto-approved), got %s", final.Status)
	}
	if len(final.StepResults) != 1 {
		t.Fatalf("expected 1 step result, got %d", len(final.StepResults))
	}
	if final.StepResults[0].Output["auto_approved"] != true {
		t.Fatal("expected auto_approved=true in gate output")
	}
}
