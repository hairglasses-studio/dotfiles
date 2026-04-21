// Package executil provides a shared command-execution helper used by
// dotfiles-mcp and its consolidated handler set (mod_systemd, mod_tmux,
// mod_process, etc.) plus mapitall.
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
