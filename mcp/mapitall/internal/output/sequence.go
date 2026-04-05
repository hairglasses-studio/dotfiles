package output

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hairglasses-studio/mapping"
)

// SequenceTarget chains multiple output actions with optional delays.
type SequenceTarget struct {
	registry *Registry
}

// NewSequenceTarget creates a sequence output target.
func NewSequenceTarget(registry *Registry) *SequenceTarget {
	return &SequenceTarget{registry: registry}
}

func (t *SequenceTarget) Type() mapping.OutputType { return mapping.OutputSequence }

func (t *SequenceTarget) Execute(action mapping.OutputAction, value float64) error {
	for i, step := range action.Steps {
		target := t.registry.Get(step.Type)
		if target == nil {
			slog.Warn("sequence step: no target", "step", i, "type", step.Type)
			continue
		}

		// Convert SequenceStep to OutputAction for dispatch.
		stepAction := mapping.OutputAction{
			Type:    step.Type,
			Keys:    step.Keys,
			Exec:    step.Exec,
			Target:  step.Target,
			Address: step.Address,
			Port:    step.Port,
			Host:    step.Host,
		}

		if err := target.Execute(stepAction, value); err != nil {
			return fmt.Errorf("sequence step %d (%s): %w", i, step.Type, err)
		}

		if step.DelayMs > 0 {
			time.Sleep(time.Duration(step.DelayMs) * time.Millisecond)
		}
	}
	return nil
}

func (t *SequenceTarget) Close() error { return nil }
