package output

import (
	"github.com/hairglasses-studio/mapping"
)

// VariableTarget handles set_var and toggle_var output types.
type VariableTarget struct {
	state *mapping.EngineState
}

// NewVariableTarget creates a variable output target.
func NewVariableTarget(state *mapping.EngineState) *VariableTarget {
	return &VariableTarget{state: state}
}

func (t *VariableTarget) Type() mapping.OutputType { return mapping.OutputSetVar }

func (t *VariableTarget) Execute(action mapping.OutputAction, value float64) error {
	t.state.SetVariable(action.Variable, value)
	return nil
}

func (t *VariableTarget) Close() error { return nil }

// ToggleVarTarget handles toggle_var by flipping a boolean variable.
type ToggleVarTarget struct {
	state *mapping.EngineState
}

// NewToggleVarTarget creates a toggle variable output target.
func NewToggleVarTarget(state *mapping.EngineState) *ToggleVarTarget {
	return &ToggleVarTarget{state: state}
}

func (t *ToggleVarTarget) Type() mapping.OutputType { return mapping.OutputToggleVar }

func (t *ToggleVarTarget) Execute(action mapping.OutputAction, _ float64) error {
	v, _ := t.state.GetVariable(action.Variable)
	if b, ok := v.(bool); ok {
		t.state.SetVariable(action.Variable, !b)
	} else {
		t.state.SetVariable(action.Variable, true)
	}
	return nil
}

func (t *ToggleVarTarget) Close() error { return nil }
