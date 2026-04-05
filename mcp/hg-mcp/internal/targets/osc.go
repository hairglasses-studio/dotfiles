package targets

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// OSCTarget sends OSC messages to any OSC-enabled software
// (Resolume Arena, TouchDesigner, VDMX, MadMapper, etc.).
type OSCTarget struct {
	id       string
	name     string
	host     string
	port     int
	actions  []ActionDescriptor
	mu       sync.Mutex
	client   *osc.Client
	conn     net.Conn
	feedback *oscFeedbackListener
}

// OSCTargetConfig configures a generic OSC target.
type OSCTargetConfig struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Host         string              `json:"host"`
	Port         int                 `json:"port"`
	Actions      []OSCActionConfig   `json:"actions,omitempty"`
	FeedbackPort int                 `json:"feedback_port,omitempty"` // 0 = no feedback
}

// OSCActionConfig defines an OSC action.
type OSCActionConfig struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Address     string   `json:"address"`      // OSC address pattern, e.g. "/composition/layers/1/video/opacity"
	Type        string   `json:"type"`          // "trigger", "set_value", "toggle"
	ParamType   string   `json:"param_type"`    // "float32", "int32", "string", "bool"
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// NewOSCTarget creates a generic OSC output target.
func NewOSCTarget(config OSCTargetConfig) *OSCTarget {
	if config.ID == "" {
		config.ID = fmt.Sprintf("osc_%s_%d", config.Host, config.Port)
	}
	if config.Name == "" {
		config.Name = fmt.Sprintf("OSC (%s:%d)", config.Host, config.Port)
	}

	t := &OSCTarget{
		id:   config.ID,
		name: config.Name,
		host: config.Host,
		port: config.Port,
	}

	// Build action descriptors from config.
	for _, ac := range config.Actions {
		actionType := ActionTrigger
		switch ac.Type {
		case "set_value":
			actionType = ActionSetValue
		case "toggle":
			actionType = ActionToggle
		}

		params := []ParamDescriptor{{
			Name: "value",
			Type: ac.ParamType,
			Min:  ac.Min,
			Max:  ac.Max,
		}}
		if actionType == ActionTrigger {
			params = nil
		}

		t.actions = append(t.actions, ActionDescriptor{
			ID:          ac.ID,
			Name:        ac.Name,
			Description: ac.Description,
			Category:    "osc",
			Type:        actionType,
			Parameters:  params,
			Tags:        ac.Tags,
		})
	}

	// Default actions if none configured.
	if len(t.actions) == 0 {
		step := 0.01
		one := 1.0
		zero := 0.0
		t.actions = []ActionDescriptor{
			{
				ID: "send_float", Name: "Send Float", Category: "osc", Type: ActionSetValue,
				Description: "Send a float value to an OSC address",
				Parameters: []ParamDescriptor{
					{Name: "address", Type: "string", Required: true, Description: "OSC address (e.g. /layer1/opacity)"},
					{Name: "value", Type: "number", Required: true, Min: &zero, Max: &one, Step: &step},
				},
			},
			{
				ID: "send_int", Name: "Send Integer", Category: "osc", Type: ActionSetValue,
				Description: "Send an integer value to an OSC address",
				Parameters: []ParamDescriptor{
					{Name: "address", Type: "string", Required: true},
					{Name: "value", Type: "number", Required: true},
				},
			},
			{
				ID: "send_trigger", Name: "Send Trigger", Category: "osc", Type: ActionTrigger,
				Description: "Send an OSC message with no arguments (trigger)",
				Parameters: []ParamDescriptor{
					{Name: "address", Type: "string", Required: true},
				},
			},
			{
				ID: "send_string", Name: "Send String", Category: "osc", Type: ActionSetValue,
				Description: "Send a string value to an OSC address",
				Parameters: []ParamDescriptor{
					{Name: "address", Type: "string", Required: true},
					{Name: "value", Type: "string", Required: true},
				},
			},
		}
	}

	return t
}

func (t *OSCTarget) ID() string       { return t.id }
func (t *OSCTarget) Name() string     { return t.name }
func (t *OSCTarget) Protocol() string { return "osc" }

func (t *OSCTarget) Connect(_ context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.client = osc.NewClient(t.host, t.port)
	return nil
}

func (t *OSCTarget) Disconnect(_ context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.client = nil
	if t.feedback != nil {
		t.feedback.close()
		t.feedback = nil
	}
	return nil
}

func (t *OSCTarget) Health(_ context.Context) TargetHealth {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.client == nil {
		return TargetHealth{Connected: false, Status: "disconnected"}
	}
	return TargetHealth{Connected: true, Status: "healthy"}
}

func (t *OSCTarget) Actions(_ context.Context) []ActionDescriptor {
	return t.actions
}

func (t *OSCTarget) Execute(_ context.Context, actionID string, params map[string]any) (*ActionResult, error) {
	t.mu.Lock()
	client := t.client
	t.mu.Unlock()

	if client == nil {
		return &ActionResult{Success: false, Error: "not connected"}, nil
	}

	address, _ := params["address"].(string)

	if address == "" {
		return &ActionResult{Success: false, Error: "address is required"}, nil
	}

	msg := osc.NewMessage(address)

	// Add value argument based on action type.
	if val, ok := params["value"]; ok {
		switch v := val.(type) {
		case float64:
			if actionID == "send_int" {
				msg.Append(int32(v))
			} else {
				msg.Append(float32(v))
			}
		case float32:
			msg.Append(v)
		case int:
			msg.Append(int32(v))
		case string:
			msg.Append(v)
		case bool:
			if v {
				msg.Append(int32(1))
			} else {
				msg.Append(int32(0))
			}
		default:
			msg.Append(fmt.Sprintf("%v", v))
		}
	}

	if err := client.Send(msg); err != nil {
		return &ActionResult{Success: false, Error: err.Error()}, nil
	}

	return &ActionResult{
		Success: true,
		Data:    map[string]any{"address": address, "host": t.host, "port": t.port},
	}, nil
}

func (t *OSCTarget) State(_ context.Context, _ string) (*StateValue, error) {
	return nil, fmt.Errorf("use OSCQuery for state queries (not yet implemented)")
}

// ---------------------------------------------------------------------------
// OSC feedback listener (optional)
// ---------------------------------------------------------------------------

type oscFeedbackListener struct {
	server *osc.Server
	done   chan struct{}
}

func (l *oscFeedbackListener) close() {
	if l.done != nil {
		close(l.done)
	}
}

// ---------------------------------------------------------------------------
// Resolume preset factory
// ---------------------------------------------------------------------------

// NewResolumeTarget creates an OSC target preconfigured for Resolume Arena.
func NewResolumeTarget(host string, port int) *OSCTarget {
	one := 1.0
	zero := 0.0
	_ = 0.01 // step available for future use
	maxBPM := 999.0
	minBPM := 20.0

	return NewOSCTarget(OSCTargetConfig{
		ID:   "resolume",
		Name: "Resolume Arena",
		Host: host,
		Port: port,
		Actions: []OSCActionConfig{
			{ID: "layer_opacity", Name: "Layer Opacity", Address: "/composition/layers/*/video/opacity/values", Type: "set_value", ParamType: "float32", Min: &zero, Max: &one, Tags: []string{"video", "layer"}},
			{ID: "master_opacity", Name: "Master Opacity", Address: "/composition/video/opacity/values", Type: "set_value", ParamType: "float32", Min: &zero, Max: &one, Tags: []string{"video", "master"}},
			{ID: "crossfader", Name: "Crossfader", Address: "/composition/video/crossfader/values", Type: "set_value", ParamType: "float32", Min: &zero, Max: &one, Tags: []string{"video", "mix"}},
			{ID: "trigger_column", Name: "Trigger Column", Address: "/composition/columns/*/connect", Type: "trigger", Tags: []string{"clip", "trigger"}},
			{ID: "trigger_clip", Name: "Trigger Clip", Address: "/composition/layers/*/clips/*/connect", Type: "trigger", Tags: []string{"clip", "trigger"}},
			{ID: "set_bpm", Name: "Set BPM", Address: "/composition/tempocontroller/tempo", Type: "set_value", ParamType: "float32", Min: &minBPM, Max: &maxBPM, Tags: []string{"tempo", "bpm"}},
			{ID: "tap_tempo", Name: "Tap Tempo", Address: "/composition/tempocontroller/tempotap", Type: "trigger", Tags: []string{"tempo", "bpm"}},
			{ID: "layer_bypass", Name: "Layer Bypass", Address: "/composition/layers/*/bypassed", Type: "toggle", ParamType: "int32", Tags: []string{"video", "layer"}},
			{ID: "layer_solo", Name: "Layer Solo", Address: "/composition/layers/*/solo", Type: "toggle", ParamType: "int32", Tags: []string{"video", "layer"}},
		},
	})
}

// NewTouchDesignerTarget creates an OSC target preconfigured for TouchDesigner.
func NewTouchDesignerTarget(host string, port int) *OSCTarget {
	return NewOSCTarget(OSCTargetConfig{
		ID:   "touchdesigner",
		Name: "TouchDesigner",
		Host: host,
		Port: port,
		// TouchDesigner uses custom OSC addresses, so start with generic actions.
		// Users configure specific addresses via the mapping profile.
	})
}

// Ensure OSCTarget implements OutputTarget.
var _ OutputTarget = (*OSCTarget)(nil)

// Ensure time is used (for feedback listener).
var _ = time.Now
