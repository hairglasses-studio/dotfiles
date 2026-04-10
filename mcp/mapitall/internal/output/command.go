package output

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
)

// CommandTarget executes shell commands.
type CommandTarget struct{}

// NewCommandTarget creates a command output target.
func NewCommandTarget() *CommandTarget { return &CommandTarget{} }

func (t *CommandTarget) Type() mapping.OutputType { return mapping.OutputCommand }

func (t *CommandTarget) Execute(action mapping.OutputAction, value float64) error {
	if len(action.Exec) == 0 {
		return fmt.Errorf("no command specified")
	}

	// Substitute {value} and {scaled} placeholders.
	args := make([]string, len(action.Exec))
	for i, arg := range action.Exec {
		arg = strings.ReplaceAll(arg, "{value}", fmt.Sprintf("%.4f", value))
		arg = strings.ReplaceAll(arg, "{scaled}", fmt.Sprintf("%.0f", value))
		args[i] = arg
	}

	cmd := exec.Command(args[0], args[1:]...)
	return cmd.Run()
}

func (t *CommandTarget) Close() error { return nil }
