// Package executil provides a shared command-execution helper used by
// the small dotfiles MCP servers (systemd-mcp, tmux-mcp, process-mcp).
package executil

import (
	"bytes"
	"os/exec"
	"strings"
)

// RunCmd executes a command and returns trimmed stdout, stderr, and any error.
func RunCmd(name string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.Command(name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}
