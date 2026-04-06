package repo_analysis

import (
	"os"
	"path/filepath"
	"strings"
)

// classifyRepo determines the type of a repository based on its structure and
// dependencies. Returns one of: "mcp-server", "tui", "cli", "server",
// "library", "container", "python-project", "node-project", "standalone",
// "unknown".
func classifyRepo(repoPath string, depPaths []string, langs []LanguageInfo) string {
	hasMCPKit := false
	hasBubbletea := false
	hasNetHTTP := false

	for _, dep := range depPaths {
		switch {
		case strings.HasPrefix(dep, "github.com/hairglasses-studio/mcpkit"):
			hasMCPKit = true
		case strings.HasPrefix(dep, "github.com/charmbracelet/bubbletea"):
			hasBubbletea = true
		case dep == "net/http":
			hasNetHTTP = true
		}
	}

	// 1. MCP server: imports mcpkit and has tool module patterns
	if hasMCPKit && hasToolModule(repoPath) {
		return "mcp-server"
	}

	// 2. Has cmd/ directory with main.go
	if hasCmdMain(repoPath) {
		// Check if it's a server (has net/http listener or common server frameworks)
		if hasNetHTTP || hasServerFramework(depPaths) {
			return "server"
		}
		// TUI if bubbletea
		if hasBubbletea {
			return "tui"
		}
		return "cli"
	}

	// 3. Bubbletea without cmd/ — still a TUI
	if hasBubbletea {
		return "tui"
	}

	// 4. No main package, has exported packages — library
	if !hasMainGo(repoPath) && hasExportedPackages(repoPath) {
		return "library"
	}

	// 5. Has Dockerfile
	if fileExists(filepath.Join(repoPath, "Dockerfile")) {
		return "container"
	}

	// 6. Primarily Python
	if primaryLanguage(langs) == "Python" {
		return "python-project"
	}

	// 7. Primarily TypeScript/JavaScript
	primary := primaryLanguage(langs)
	if primary == "TypeScript" || primary == "JavaScript" {
		return "node-project"
	}

	// 8. Has main.go in root
	if fileExists(filepath.Join(repoPath, "main.go")) {
		return "standalone"
	}

	return "unknown"
}

// hasToolModule checks if the repo has files matching the ToolModule pattern
// (a function named Tools() returning []ToolDefinition or similar).
func hasToolModule(repoPath string) bool {
	// Look for module.go files under internal/mcp/tools/ or tools/
	patterns := []string{
		filepath.Join(repoPath, "internal", "mcp", "tools", "*", "module.go"),
		filepath.Join(repoPath, "tools", "*", "module.go"),
	}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

// hasCmdMain checks if there's a cmd/ directory containing at least one main.go.
func hasCmdMain(repoPath string) bool {
	cmdDir := filepath.Join(repoPath, "cmd")
	info, err := os.Stat(cmdDir)
	if err != nil || !info.IsDir() {
		return false
	}
	matches, err := filepath.Glob(filepath.Join(cmdDir, "*", "main.go"))
	if err != nil {
		return false
	}
	return len(matches) > 0
}

// hasMainGo checks if there is a main.go in the root directory.
func hasMainGo(repoPath string) bool {
	return fileExists(filepath.Join(repoPath, "main.go"))
}

// hasExportedPackages checks if there are .go files with exported functions
// outside of cmd/ and internal/ directories. A simplified heuristic that checks
// for .go files in the root or top-level packages.
func hasExportedPackages(repoPath string) bool {
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			if e.Name() != "main.go" {
				return true
			}
		}
	}
	// Check top-level subdirectories (excluding cmd/, internal/, vendor/)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "cmd" || name == "internal" || name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") {
			continue
		}
		goFiles, _ := filepath.Glob(filepath.Join(repoPath, name, "*.go"))
		if len(goFiles) > 0 {
			return true
		}
	}
	return false
}

// hasServerFramework checks if any dependencies are common HTTP server frameworks.
func hasServerFramework(depPaths []string) bool {
	serverPrefixes := []string{
		"github.com/gin-gonic/gin",
		"github.com/labstack/echo",
		"github.com/gofiber/fiber",
		"github.com/gorilla/mux",
		"github.com/go-chi/chi",
	}
	for _, dep := range depPaths {
		for _, prefix := range serverPrefixes {
			if strings.HasPrefix(dep, prefix) {
				return true
			}
		}
	}
	return false
}

// primaryLanguage returns the language with the highest file count.
func primaryLanguage(langs []LanguageInfo) string {
	if len(langs) == 0 {
		return ""
	}
	best := langs[0]
	for _, l := range langs[1:] {
		if l.Files > best.Files {
			best = l
		}
	}
	return best.Name
}

// fileExists returns true if the path exists and is a regular file or directory.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
