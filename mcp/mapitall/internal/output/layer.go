package output

import (
	"strconv"

	"github.com/hairglasses-studio/mapping"
)

// LayerTarget handles layer switching by mutating EngineState.
type LayerTarget struct {
	state *mapping.EngineState
}

// NewLayerTarget creates a layer switch output target.
func NewLayerTarget(state *mapping.EngineState) *LayerTarget {
	return &LayerTarget{state: state}
}

func (t *LayerTarget) Type() mapping.OutputType { return mapping.OutputLayerSwitch }

func (t *LayerTarget) Execute(action mapping.OutputAction, _ float64) error {
	layer, _ := strconv.Atoi(action.Layer)
	// Target is the device context; empty means global.
	deviceID := action.Target
	if deviceID == "" {
		deviceID = "__global"
	}
	t.state.SetActiveLayer(deviceID, layer)
	return nil
}

func (t *LayerTarget) Close() error { return nil }
