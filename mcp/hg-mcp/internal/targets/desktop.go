package targets

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// DesktopTarget simulates keyboard/mouse events and manages windows
// using platform-specific tools (ydotool, wtype, xdotool, hyprctl, swaymsg).
type DesktopTarget struct {
	platform string // "linux-wayland", "linux-x11", "darwin"
	wm       string // "hyprland", "sway", "none"
}

// NewDesktopTarget creates a desktop target with auto-detected platform.
func NewDesktopTarget() *DesktopTarget {
	t := &DesktopTarget{}
	t.detectPlatform()
	return t
}

func (t *DesktopTarget) detectPlatform() {
	switch runtime.GOOS {
	case "darwin":
		t.platform = "darwin"
		t.wm = "none"
	case "linux":
		if _, err := exec.LookPath("hyprctl"); err == nil {
			t.platform = "linux-wayland"
			t.wm = "hyprland"
		} else if _, err := exec.LookPath("swaymsg"); err == nil {
			t.platform = "linux-wayland"
			t.wm = "sway"
		} else {
			t.platform = "linux-x11"
			t.wm = "none"
		}
	default:
		t.platform = runtime.GOOS
		t.wm = "none"
	}
}

func (t *DesktopTarget) ID() string       { return "desktop" }
func (t *DesktopTarget) Name() string     { return "Desktop Control" }
func (t *DesktopTarget) Protocol() string { return "platform" }

func (t *DesktopTarget) Connect(_ context.Context) error    { return nil }
func (t *DesktopTarget) Disconnect(_ context.Context) error { return nil }

func (t *DesktopTarget) Health(_ context.Context) TargetHealth {
	return TargetHealth{
		Connected: true,
		Status:    "healthy",
		Issues:    nil,
	}
}

func (t *DesktopTarget) Actions(_ context.Context) []ActionDescriptor {
	actions := []ActionDescriptor{
		{
			ID: "key_press", Name: "Press Key", Category: "input", Type: ActionTrigger,
			Description: "Simulate a key press (uses ydotool on Wayland, xdotool on X11)",
			Parameters: []ParamDescriptor{
				{Name: "key", Type: "string", Required: true, Description: "Key name or combo (e.g. 'Return', 'ctrl+shift+f', 'KEY_A')"},
			},
			Tags: []string{"keyboard", "input"},
		},
		{
			ID: "type_text", Name: "Type Text", Category: "input", Type: ActionTrigger,
			Description: "Type a string of text at the cursor position",
			Parameters: []ParamDescriptor{
				{Name: "text", Type: "string", Required: true, Description: "Text to type"},
			},
			Tags: []string{"keyboard", "input"},
		},
		{
			ID: "mouse_click", Name: "Mouse Click", Category: "input", Type: ActionTrigger,
			Description: "Click at coordinates or current position",
			Parameters: []ParamDescriptor{
				{Name: "button", Type: "select", Options: []string{"left", "right", "middle"}, Default: "left"},
				{Name: "x", Type: "number", Description: "X coordinate (omit for current position)"},
				{Name: "y", Type: "number", Description: "Y coordinate (omit for current position)"},
			},
			Tags: []string{"mouse", "input"},
		},
		{
			ID: "run_command", Name: "Run Command", Category: "system", Type: ActionTrigger,
			Description: "Execute a shell command",
			Parameters: []ParamDescriptor{
				{Name: "command", Type: "string", Required: true, Description: "Shell command to execute"},
			},
			Tags: []string{"shell", "system"},
		},
	}

	// Add window manager-specific actions.
	if t.wm == "hyprland" || t.wm == "sway" {
		actions = append(actions,
			ActionDescriptor{
				ID: "workspace_switch", Name: "Switch Workspace", Category: "window", Type: ActionTrigger,
				Description: "Switch to a workspace by number",
				Parameters: []ParamDescriptor{
					{Name: "workspace", Type: "number", Required: true, Description: "Workspace number"},
				},
				Tags: []string{"workspace", "window"},
			},
			ActionDescriptor{
				ID: "window_focus", Name: "Focus Window", Category: "window", Type: ActionTrigger,
				Description: "Focus a window by class name",
				Parameters: []ParamDescriptor{
					{Name: "class", Type: "string", Required: true, Description: "Window class name"},
				},
				Tags: []string{"window", "focus"},
			},
		)
	}

	return actions
}

func (t *DesktopTarget) Execute(ctx context.Context, actionID string, params map[string]any) (*ActionResult, error) {
	execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch actionID {
	case "key_press":
		key, _ := params["key"].(string)
		if key == "" {
			return &ActionResult{Success: false, Error: "key is required"}, nil
		}
		return t.execKeyPress(execCtx, key)

	case "type_text":
		text, _ := params["text"].(string)
		if text == "" {
			return &ActionResult{Success: false, Error: "text is required"}, nil
		}
		return t.execTypeText(execCtx, text)

	case "mouse_click":
		button, _ := params["button"].(string)
		if button == "" {
			button = "left"
		}
		return t.execMouseClick(execCtx, button, params["x"], params["y"])

	case "run_command":
		cmd, _ := params["command"].(string)
		if cmd == "" {
			return &ActionResult{Success: false, Error: "command is required"}, nil
		}
		return t.execCommand(execCtx, cmd)

	case "workspace_switch":
		ws := fmt.Sprintf("%v", params["workspace"])
		return t.execWorkspaceSwitch(execCtx, ws)

	case "window_focus":
		class, _ := params["class"].(string)
		return t.execWindowFocus(execCtx, class)

	default:
		return &ActionResult{Success: false, Error: fmt.Sprintf("unknown action: %s", actionID)}, nil
	}
}

func (t *DesktopTarget) State(_ context.Context, _ string) (*StateValue, error) {
	return nil, fmt.Errorf("desktop target does not support state queries")
}

// ---------------------------------------------------------------------------
// Platform-specific implementations
// ---------------------------------------------------------------------------

func (t *DesktopTarget) execKeyPress(ctx context.Context, key string) (*ActionResult, error) {
	var cmd *exec.Cmd
	switch t.platform {
	case "linux-wayland":
		if _, err := exec.LookPath("ydotool"); err == nil {
			cmd = exec.CommandContext(ctx, "ydotool", "key", key)
		} else if _, err := exec.LookPath("wtype"); err == nil {
			cmd = exec.CommandContext(ctx, "wtype", "-k", key)
		}
	case "linux-x11":
		cmd = exec.CommandContext(ctx, "xdotool", "key", key)
	case "darwin":
		// Use osascript for key simulation on macOS
		script := fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, key)
		cmd = exec.CommandContext(ctx, "osascript", "-e", script)
	}

	if cmd == nil {
		return &ActionResult{Success: false, Error: "no key simulation tool available"}, nil
	}

	return runCmd(cmd)
}

func (t *DesktopTarget) execTypeText(ctx context.Context, text string) (*ActionResult, error) {
	var cmd *exec.Cmd
	switch t.platform {
	case "linux-wayland":
		if _, err := exec.LookPath("wtype"); err == nil {
			cmd = exec.CommandContext(ctx, "wtype", text)
		} else if _, err := exec.LookPath("ydotool"); err == nil {
			cmd = exec.CommandContext(ctx, "ydotool", "type", text)
		}
	case "linux-x11":
		cmd = exec.CommandContext(ctx, "xdotool", "type", "--", text)
	case "darwin":
		script := fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, text)
		cmd = exec.CommandContext(ctx, "osascript", "-e", script)
	}

	if cmd == nil {
		return &ActionResult{Success: false, Error: "no text typing tool available"}, nil
	}

	return runCmd(cmd)
}

func (t *DesktopTarget) execMouseClick(ctx context.Context, button string, x, y any) (*ActionResult, error) {
	buttonMap := map[string]string{"left": "1", "right": "3", "middle": "2"}
	btn := buttonMap[button]
	if btn == "" {
		btn = "1"
	}

	var cmd *exec.Cmd
	switch t.platform {
	case "linux-wayland", "linux-x11":
		if _, err := exec.LookPath("ydotool"); err == nil {
			args := []string{"click", btn}
			cmd = exec.CommandContext(ctx, "ydotool", args...)
		}
	case "darwin":
		script := fmt.Sprintf(`tell application "System Events" to click`)
		cmd = exec.CommandContext(ctx, "osascript", "-e", script)
	}

	if cmd == nil {
		return &ActionResult{Success: false, Error: "no mouse click tool available"}, nil
	}

	return runCmd(cmd)
}

func (t *DesktopTarget) execCommand(ctx context.Context, command string) (*ActionResult, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	return runCmd(cmd)
}

func (t *DesktopTarget) execWorkspaceSwitch(ctx context.Context, workspace string) (*ActionResult, error) {
	var cmd *exec.Cmd
	switch t.wm {
	case "hyprland":
		cmd = exec.CommandContext(ctx, "hyprctl", "dispatch", "workspace", workspace)
	case "sway":
		cmd = exec.CommandContext(ctx, "swaymsg", "workspace", "number", workspace)
	default:
		return &ActionResult{Success: false, Error: "no window manager detected"}, nil
	}
	return runCmd(cmd)
}

func (t *DesktopTarget) execWindowFocus(ctx context.Context, class string) (*ActionResult, error) {
	var cmd *exec.Cmd
	switch t.wm {
	case "hyprland":
		cmd = exec.CommandContext(ctx, "hyprctl", "dispatch", "focuswindow", class)
	case "sway":
		cmd = exec.CommandContext(ctx, "swaymsg", fmt.Sprintf(`[app_id="%s"] focus`, class))
	default:
		return &ActionResult{Success: false, Error: "no window manager detected"}, nil
	}
	return runCmd(cmd)
}

func runCmd(cmd *exec.Cmd) (*ActionResult, error) {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return &ActionResult{
			Success: false,
			Error:   fmt.Sprintf("%v: %s", err, strings.TrimSpace(stderr.String())),
		}, nil
	}

	return &ActionResult{
		Success: true,
		Data:    map[string]any{"output": strings.TrimSpace(stdout.String())},
	}, nil
}

var _ OutputTarget = (*DesktopTarget)(nil)
