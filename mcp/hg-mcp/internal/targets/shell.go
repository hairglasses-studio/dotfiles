package targets

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ShellTarget executes shell commands as output actions.
// It wraps the existing makima [commands] pattern with typed parameters.
type ShellTarget struct {
	commands []ShellCommand
}

// ShellCommand defines a named shell command that can be triggered.
type ShellCommand struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Command     string            `json:"command"`                // Command template with {{param}} placeholders
	Params      map[string]string `json:"params,omitempty"`       // param name -> type ("string", "number", "boolean")
	Tags        []string          `json:"tags,omitempty"`
}

// NewShellTarget creates a shell target with the given command definitions.
func NewShellTarget(commands []ShellCommand) *ShellTarget {
	return &ShellTarget{commands: commands}
}

func (t *ShellTarget) ID() string       { return "shell" }
func (t *ShellTarget) Name() string     { return "Shell Commands" }
func (t *ShellTarget) Protocol() string { return "shell" }

func (t *ShellTarget) Connect(_ context.Context) error    { return nil }
func (t *ShellTarget) Disconnect(_ context.Context) error { return nil }

func (t *ShellTarget) Health(_ context.Context) TargetHealth {
	return TargetHealth{Connected: true, Status: "healthy"}
}

func (t *ShellTarget) Actions(_ context.Context) []ActionDescriptor {
	actions := make([]ActionDescriptor, 0, len(t.commands))
	for _, cmd := range t.commands {
		params := make([]ParamDescriptor, 0, len(cmd.Params))
		for name, typ := range cmd.Params {
			params = append(params, ParamDescriptor{
				Name: name,
				Type: typ,
			})
		}
		actions = append(actions, ActionDescriptor{
			ID:          cmd.ID,
			Name:        cmd.Name,
			Description: cmd.Description,
			Category:    "shell",
			Type:        ActionTrigger,
			Parameters:  params,
			Tags:        cmd.Tags,
		})
	}
	return actions
}

func (t *ShellTarget) Execute(ctx context.Context, actionID string, params map[string]any) (*ActionResult, error) {
	var cmd *ShellCommand
	for i := range t.commands {
		if t.commands[i].ID == actionID {
			cmd = &t.commands[i]
			break
		}
	}
	if cmd == nil {
		return &ActionResult{Success: false, Error: fmt.Sprintf("unknown action: %s", actionID)}, nil
	}

	// Substitute template parameters.
	command := cmd.Command
	for key, val := range params {
		placeholder := "{{" + key + "}}"
		command = strings.ReplaceAll(command, placeholder, fmt.Sprintf("%v", val))
	}

	// Execute via sh -c.
	execCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var stdout, stderr bytes.Buffer
	c := exec.CommandContext(execCtx, "sh", "-c", command)
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	if err != nil {
		return &ActionResult{
			Success: false,
			Error:   fmt.Sprintf("%v: %s", err, stderr.String()),
			Data:    map[string]any{"stdout": stdout.String(), "stderr": stderr.String()},
		}, nil
	}

	return &ActionResult{
		Success: true,
		Data:    map[string]any{"stdout": strings.TrimSpace(stdout.String())},
	}, nil
}

func (t *ShellTarget) State(_ context.Context, _ string) (*StateValue, error) {
	return nil, fmt.Errorf("shell target does not support state queries")
}
