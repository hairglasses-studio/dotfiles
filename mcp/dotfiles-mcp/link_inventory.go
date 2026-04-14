package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type linkSpec struct {
	src string
	dst string
}

func installerScriptPath() string {
	return filepath.Join(dotfilesDir(), "install.sh")
}

func parseLinkSpecsOutput(raw string) ([]linkSpec, error) {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	var links []linkSpec
	lineNo := 0

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid link spec line %d: %q", lineNo, line)
		}
		links = append(links, linkSpec{src: parts[0], dst: parts[1]})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan install inventory: %w", err)
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("installer returned no link specs")
	}
	return links, nil
}

func loadManagedLinkSpecs() ([]linkSpec, error) {
	script := installerScriptPath()
	if _, err := os.Stat(script); err != nil {
		return nil, fmt.Errorf("install script unavailable at %s: %w", script, err)
	}

	cmd := exec.Command("bash", script, "--print-link-specs")
	cmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("install.sh --print-link-specs failed: %w: %s", err, msg)
		}
		return nil, fmt.Errorf("install.sh --print-link-specs failed: %w", err)
	}

	links, err := parseLinkSpecsOutput(stdout.String())
	if err != nil {
		return nil, fmt.Errorf("parse install inventory: %w", err)
	}
	return links, nil
}
