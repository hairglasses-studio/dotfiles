package mapping

import (
	"fmt"
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// EngineState — basic operations
// ---------------------------------------------------------------------------

func TestEngineState_NewEngineState(t *testing.T) {
	state := NewEngineState()
	if state == nil {
		t.Fatal("NewEngineState returned nil")
	}
	if state.ActiveModifiers == nil {
		t.Error("ActiveModifiers map is nil")
	}
	if state.ActiveLayer == nil {
		t.Error("ActiveLayer map is nil")
	}
	if state.Variables == nil {
		t.Error("Variables map is nil")
	}
	if state.PickupState == nil {
		t.Error("PickupState map is nil")
	}
	if state.FaderCrossed == nil {
		t.Error("FaderCrossed map is nil")
	}
}

func TestEngineState_SetGetActiveLayer(t *testing.T) {
	state := NewEngineState()

	// Default layer is 0.
	if got := state.GetActiveLayer("dev1"); got != 0 {
		t.Errorf("GetActiveLayer(dev1) = %d, want 0 (default)", got)
	}

	state.SetActiveLayer("dev1", 3)
	if got := state.GetActiveLayer("dev1"); got != 3 {
		t.Errorf("GetActiveLayer(dev1) = %d, want 3", got)
	}

	// Different device.
	state.SetActiveLayer("dev2", 7)
	if got := state.GetActiveLayer("dev2"); got != 7 {
		t.Errorf("GetActiveLayer(dev2) = %d, want 7", got)
	}
	// dev1 unchanged.
	if got := state.GetActiveLayer("dev1"); got != 3 {
		t.Errorf("GetActiveLayer(dev1) = %d, want 3 (unchanged)", got)
	}
}

func TestEngineState_SetGetVariable(t *testing.T) {
	state := NewEngineState()

	// Missing variable.
	_, ok := state.GetVariable("missing")
	if ok {
		t.Error("expected ok=false for missing variable")
	}

	state.SetVariable("mode", "combat")
	val, ok := state.GetVariable("mode")
	if !ok {
		t.Fatal("expected ok=true for set variable")
	}
	if val != "combat" {
		t.Errorf("GetVariable(mode) = %v, want %q", val, "combat")
	}

	// Overwrite.
	state.SetVariable("mode", "explore")
	val, _ = state.GetVariable("mode")
	if val != "explore" {
		t.Errorf("GetVariable(mode) = %v, want %q", val, "explore")
	}
}

func TestEngineState_SetGetActiveApp(t *testing.T) {
	state := NewEngineState()

	if got := state.GetActiveApp(); got != "" {
		t.Errorf("GetActiveApp() = %q, want empty", got)
	}

	state.SetActiveApp("firefox")
	if got := state.GetActiveApp(); got != "firefox" {
		t.Errorf("GetActiveApp() = %q, want %q", got, "firefox")
	}
}

func TestEngineState_SetModifier(t *testing.T) {
	state := NewEngineState()

	state.SetModifier("BTN_TL", true)
	if !state.ActiveModifiers["BTN_TL"] {
		t.Error("expected BTN_TL to be active")
	}

	state.SetModifier("BTN_TL", false)
	if state.ActiveModifiers["BTN_TL"] {
		t.Error("expected BTN_TL to be inactive")
	}
	// Key should be deleted, not just set to false.
	if _, exists := state.ActiveModifiers["BTN_TL"]; exists {
		t.Error("expected BTN_TL to be removed from map")
	}
}

func TestEngineState_RangeVariables(t *testing.T) {
	state := NewEngineState()
	state.SetVariable("a", 1)
	state.SetVariable("b", 2)
	state.SetVariable("c", 3)

	count := 0
	state.RangeVariables(func(key string, value any) bool {
		count++
		return true
	})
	if count != 3 {
		t.Errorf("RangeVariables visited %d entries, want 3", count)
	}
}

func TestEngineState_RangeVariables_EarlyStop(t *testing.T) {
	state := NewEngineState()
	state.SetVariable("a", 1)
	state.SetVariable("b", 2)
	state.SetVariable("c", 3)

	count := 0
	state.RangeVariables(func(key string, value any) bool {
		count++
		return false // stop after first
	})
	if count != 1 {
		t.Errorf("RangeVariables with early stop visited %d entries, want 1", count)
	}
}

// ---------------------------------------------------------------------------
// EngineState — concurrent SetActiveLayer/GetActiveLayer
// ---------------------------------------------------------------------------

func TestEngineState_ConcurrentLayerAccess(t *testing.T) {
	state := NewEngineState()
	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			state.SetActiveLayer("dev1", n)
		}(i)
		go func() {
			defer wg.Done()
			_ = state.GetActiveLayer("dev1")
		}()
	}

	wg.Wait()

	// If no data race, the test passes. Verify we can still read.
	layer := state.GetActiveLayer("dev1")
	if layer < 0 || layer >= goroutines {
		t.Errorf("GetActiveLayer after concurrent access = %d, unexpected", layer)
	}
}

// ---------------------------------------------------------------------------
// EngineState — concurrent SetVariable/GetVariable
// ---------------------------------------------------------------------------

func TestEngineState_ConcurrentVariableAccess(t *testing.T) {
	state := NewEngineState()
	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			state.SetVariable("counter", n)
		}(i)
		go func() {
			defer wg.Done()
			_, _ = state.GetVariable("counter")
		}()
	}

	wg.Wait()

	val, ok := state.GetVariable("counter")
	if !ok {
		t.Fatal("expected counter to exist after concurrent writes")
	}
	n, ok := ToFloat64(val)
	if !ok {
		t.Fatalf("counter value %v is not numeric", val)
	}
	if n < 0 || n >= goroutines {
		t.Errorf("counter = %v, unexpected", n)
	}
}

// ---------------------------------------------------------------------------
// EngineState — concurrent modifier operations
// ---------------------------------------------------------------------------

func TestEngineState_ConcurrentModifierAccess(t *testing.T) {
	state := NewEngineState()
	var wg sync.WaitGroup
	const goroutines = 50

	mods := []string{"BTN_TL", "BTN_TR", "BTN_TL2", "BTN_TR2"}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			mod := mods[n%len(mods)]
			state.SetModifier(mod, n%2 == 0)
		}(i)
	}

	wg.Wait()

	// Should not panic or corrupt state.
	state.SetModifier("BTN_TL", true)
	if !state.ActiveModifiers["BTN_TL"] {
		t.Error("expected BTN_TL to be active after concurrent operations")
	}
}

// ---------------------------------------------------------------------------
// EngineState — concurrent mixed operations
// ---------------------------------------------------------------------------

func TestEngineState_ConcurrentMixedAccess(t *testing.T) {
	state := NewEngineState()
	var wg sync.WaitGroup
	const goroutines = 20

	for i := 0; i < goroutines; i++ {
		wg.Add(5)
		go func(n int) {
			defer wg.Done()
			state.SetActiveLayer(fmt.Sprintf("dev%d", n%3), n%5)
		}(i)
		go func(n int) {
			defer wg.Done()
			state.SetVariable(fmt.Sprintf("var%d", n%5), n)
		}(i)
		go func(n int) {
			defer wg.Done()
			state.SetModifier(fmt.Sprintf("MOD_%d", n%4), n%2 == 0)
		}(i)
		go func(n int) {
			defer wg.Done()
			state.SetActiveApp(fmt.Sprintf("app%d", n%3))
		}(i)
		go func() {
			defer wg.Done()
			_ = state.GetActiveApp()
			_ = state.GetActiveLayer("dev0")
			_, _ = state.GetVariable("var0")
		}()
	}

	wg.Wait()

	// Verify state is readable after all concurrent operations.
	_ = state.GetActiveApp()
	_ = state.GetActiveLayer("dev0")
	_, _ = state.GetVariable("var0")
}

// ---------------------------------------------------------------------------
// EngineState — concurrent Resolve with state mutations
// ---------------------------------------------------------------------------

func TestEngineState_ConcurrentResolveWithMutations(t *testing.T) {
	state := NewEngineState()
	idx := BuildRuleIndex(&MappingProfile{
		Mappings: []MappingRule{
			{Input: "BTN_SOUTH", Output: OutputAction{Type: OutputKey, Keys: []string{"KEY_A"}}},
			{
				Input:     "BTN_SOUTH",
				Modifiers: []string{"BTN_TL"},
				Priority:  10,
				Output:    OutputAction{Type: OutputCommand, Exec: []string{"mod-action"}},
			},
			{
				Input:     "BTN_SOUTH",
				Layer:     2,
				Priority:  5,
				Output:    OutputAction{Type: OutputOSC, Address: "/layer2"},
			},
			{
				Input:     "BTN_SOUTH",
				Condition: &Condition{Variable: "mode", Equals: "combat"},
				Priority:  3,
				Output:    OutputAction{Type: OutputMovement, Target: "cursor_up"},
			},
		},
	})

	var wg sync.WaitGroup
	const goroutines = 30

	for i := 0; i < goroutines; i++ {
		wg.Add(3)
		// Writers.
		go func(n int) {
			defer wg.Done()
			state.SetModifier("BTN_TL", n%2 == 0)
			state.SetActiveLayer("dev1", n%3)
			if n%3 == 0 {
				state.SetVariable("mode", "combat")
			} else {
				state.SetVariable("mode", "explore")
			}
		}(i)
		// Readers via Resolve.
		go func() {
			defer wg.Done()
			_ = idx.Resolve("BTN_SOUTH", state, "dev1")
		}()
		go func() {
			defer wg.Done()
			_ = idx.Resolve("BTN_SOUTH", state, "")
		}()
	}

	wg.Wait()

	// If we get here without -race failures, the test passes.
	r := idx.Resolve("BTN_SOUTH", state, "dev1")
	if r == nil {
		t.Error("expected at least one rule to match after concurrent operations")
	}
}
