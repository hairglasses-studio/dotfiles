package targets

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Registry tests
// ---------------------------------------------------------------------------

func TestRegistry_RegisterAndList(t *testing.T) {
	reg := NewRegistry()

	shell := NewShellTarget([]ShellCommand{
		{ID: "echo", Name: "Echo", Command: "echo hello"},
	})
	if err := reg.Register(shell); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Duplicate registration should fail.
	if err := reg.Register(shell); err == nil {
		t.Error("expected error for duplicate registration")
	}

	health := reg.List(context.Background())
	if _, ok := health["shell"]; !ok {
		t.Error("expected shell target in list")
	}
}

func TestRegistry_ConnectAndExecute(t *testing.T) {
	reg := NewRegistry()

	shell := NewShellTarget([]ShellCommand{
		{ID: "echo", Name: "Echo", Command: "echo hello"},
	})
	reg.Register(shell)

	if err := reg.Connect(context.Background(), "shell"); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	result, err := reg.Execute(context.Background(), "shell", "echo", nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
}

func TestRegistry_AllActions(t *testing.T) {
	reg := NewRegistry()

	shell := NewShellTarget([]ShellCommand{
		{ID: "cmd1", Name: "Command 1", Command: "echo 1"},
		{ID: "cmd2", Name: "Command 2", Command: "echo 2"},
	})
	reg.Register(shell)
	reg.Connect(context.Background(), "shell")

	actions := reg.AllActions()
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}

	// Check qualified IDs.
	ids := map[string]bool{}
	for _, a := range actions {
		ids[a.FullID] = true
	}
	if !ids["shell.cmd1"] || !ids["shell.cmd2"] {
		t.Errorf("expected shell.cmd1 and shell.cmd2, got %v", ids)
	}
}

func TestRegistry_SearchActions(t *testing.T) {
	reg := NewRegistry()

	shell := NewShellTarget([]ShellCommand{
		{ID: "volume_up", Name: "Volume Up", Command: "amixer set Master 5%+"},
		{ID: "brightness", Name: "Set Brightness", Command: "brightnessctl set 50%"},
	})
	reg.Register(shell)
	reg.Connect(context.Background(), "shell")

	results := reg.SearchActions("volume")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for 'volume', got %d", len(results))
	}
	if results[0].Action.ID != "volume_up" {
		t.Errorf("expected volume_up, got %s", results[0].Action.ID)
	}
}

func TestRegistry_Unregister(t *testing.T) {
	reg := NewRegistry()
	shell := NewShellTarget(nil)
	reg.Register(shell)

	if err := reg.Unregister(context.Background(), "shell"); err != nil {
		t.Fatalf("Unregister: %v", err)
	}

	if _, ok := reg.Get("shell"); ok {
		t.Error("target should be removed after unregister")
	}
}

// ---------------------------------------------------------------------------
// Shell target tests
// ---------------------------------------------------------------------------

func TestShellTarget_Execute(t *testing.T) {
	shell := NewShellTarget([]ShellCommand{
		{ID: "echo", Name: "Echo", Command: "echo hello world"},
	})

	result, err := shell.Execute(context.Background(), "echo", nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success, got: %s", result.Error)
	}
	if stdout, ok := result.Data["stdout"].(string); !ok || stdout != "hello world" {
		t.Errorf("stdout = %q, want 'hello world'", result.Data["stdout"])
	}
}

func TestShellTarget_ExecuteWithParams(t *testing.T) {
	shell := NewShellTarget([]ShellCommand{
		{
			ID:      "volume",
			Name:    "Set Volume",
			Command: "echo volume={{level}}",
			Params:  map[string]string{"level": "number"},
		},
	})

	result, err := shell.Execute(context.Background(), "volume", map[string]any{"level": 75})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}
	if stdout := result.Data["stdout"].(string); stdout != "volume=75" {
		t.Errorf("stdout = %q, want 'volume=75'", stdout)
	}
}

func TestShellTarget_ExecuteUnknownAction(t *testing.T) {
	shell := NewShellTarget(nil)
	result, err := shell.Execute(context.Background(), "nonexistent", nil)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if result.Success {
		t.Error("expected failure for unknown action")
	}
}

func TestShellTarget_Actions(t *testing.T) {
	shell := NewShellTarget([]ShellCommand{
		{ID: "a", Name: "Action A", Command: "echo a", Tags: []string{"audio"}},
		{ID: "b", Name: "Action B", Command: "echo b"},
	})

	actions := shell.Actions(context.Background())
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	if actions[0].Category != "shell" {
		t.Errorf("category = %q, want 'shell'", actions[0].Category)
	}
}

func TestShellTarget_Health(t *testing.T) {
	shell := NewShellTarget(nil)
	h := shell.Health(context.Background())
	if !h.Connected {
		t.Error("shell should always be connected")
	}
	if h.Status != "healthy" {
		t.Errorf("status = %q, want 'healthy'", h.Status)
	}
}

// ---------------------------------------------------------------------------
// Feedback bus tests
// ---------------------------------------------------------------------------

func TestFeedbackBus_PublishAndSubscribe(t *testing.T) {
	bus := NewFeedbackBus()

	var received atomic.Int32
	bus.Subscribe(FeedbackFilter{}, func(event FeedbackEvent) {
		received.Add(1)
	})

	bus.Publish(FeedbackEvent{
		TargetID:  "obs",
		Path:      "/scene/current",
		NewValue:  "Gaming",
		Timestamp: time.Now(),
	})

	// Wait briefly for async dispatch.
	time.Sleep(50 * time.Millisecond)

	if got := received.Load(); got != 1 {
		t.Errorf("received %d events, want 1", got)
	}
}

func TestFeedbackBus_Filter(t *testing.T) {
	bus := NewFeedbackBus()

	var obsCount, allCount atomic.Int32

	bus.Subscribe(FeedbackFilter{TargetIDs: []string{"obs"}}, func(event FeedbackEvent) {
		obsCount.Add(1)
	})
	bus.Subscribe(FeedbackFilter{}, func(event FeedbackEvent) {
		allCount.Add(1)
	})

	bus.Publish(FeedbackEvent{TargetID: "obs", Path: "/scene"})
	bus.Publish(FeedbackEvent{TargetID: "resolume", Path: "/bpm"})

	time.Sleep(50 * time.Millisecond)

	if got := obsCount.Load(); got != 1 {
		t.Errorf("obs subscriber got %d events, want 1", got)
	}
	if got := allCount.Load(); got != 2 {
		t.Errorf("all subscriber got %d events, want 2", got)
	}
}

func TestFeedbackBus_PathFilter(t *testing.T) {
	bus := NewFeedbackBus()

	var count atomic.Int32
	bus.Subscribe(FeedbackFilter{PathPrefix: "/layers"}, func(event FeedbackEvent) {
		count.Add(1)
	})

	bus.Publish(FeedbackEvent{Path: "/layers/1/opacity"})
	bus.Publish(FeedbackEvent{Path: "/scene/current"}) // Should not match
	bus.Publish(FeedbackEvent{Path: "/layers/2/bypass"})

	time.Sleep(50 * time.Millisecond)

	if got := count.Load(); got != 2 {
		t.Errorf("path subscriber got %d events, want 2", got)
	}
}

func TestFeedbackBus_Unsubscribe(t *testing.T) {
	bus := NewFeedbackBus()

	var count atomic.Int32
	id := bus.Subscribe(FeedbackFilter{}, func(event FeedbackEvent) {
		count.Add(1)
	})

	bus.Publish(FeedbackEvent{Path: "/test"})
	time.Sleep(50 * time.Millisecond)

	bus.Unsubscribe(id)
	bus.Publish(FeedbackEvent{Path: "/test2"})
	time.Sleep(50 * time.Millisecond)

	if got := count.Load(); got != 1 {
		t.Errorf("got %d events after unsubscribe, want 1", got)
	}
}

// ---------------------------------------------------------------------------
// Action registry tests
// ---------------------------------------------------------------------------

func TestActionRegistry_IndexAndSearch(t *testing.T) {
	reg := NewActionRegistry()

	reg.IndexTarget("obs", []ActionDescriptor{
		{ID: "scene_switch", Name: "Switch Scene", Tags: []string{"scene"}},
		{ID: "source_volume", Name: "Source Volume", Tags: []string{"audio"}},
	})
	reg.IndexTarget("resolume", []ActionDescriptor{
		{ID: "layer_opacity", Name: "Layer Opacity", Tags: []string{"video"}},
	})

	all := reg.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(all))
	}

	results := reg.Search("volume")
	if len(results) != 1 || results[0].FullID != "obs.source_volume" {
		t.Errorf("search 'volume' = %v, want [obs.source_volume]", results)
	}

	audio := reg.ByTag("audio")
	if len(audio) != 1 || audio[0].FullID != "obs.source_volume" {
		t.Errorf("ByTag 'audio' = %v", audio)
	}
}

func TestActionRegistry_RemoveTarget(t *testing.T) {
	reg := NewActionRegistry()
	reg.IndexTarget("obs", []ActionDescriptor{
		{ID: "a", Name: "A", Tags: []string{"tag1"}},
	})
	reg.IndexTarget("resolume", []ActionDescriptor{
		{ID: "b", Name: "B", Tags: []string{"tag1"}},
	})

	reg.RemoveTarget("obs")

	all := reg.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 action after remove, got %d", len(all))
	}
	if all[0].FullID != "resolume.b" {
		t.Errorf("expected resolume.b, got %s", all[0].FullID)
	}

	tagged := reg.ByTag("tag1")
	if len(tagged) != 1 {
		t.Errorf("expected 1 tagged action after remove, got %d", len(tagged))
	}
}
