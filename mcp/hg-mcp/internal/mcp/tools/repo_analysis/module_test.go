package repo_analysis

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// TestModuleRegistration validates module metadata and tool definitions.
func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

// TestApplyRules verifies that the rules engine correctly maps import paths to tags.
func TestApplyRules(t *testing.T) {
	deps := []string{
		"github.com/hairglasses-studio/mcpkit",
		"github.com/mark3labs/mcp-go",
		"github.com/charmbracelet/bubbletea",
		"go.opentelemetry.io/otel",
		"github.com/mattn/go-sqlite3",
		"github.com/aws/aws-sdk-go-v2",
		"github.com/anthropics/anthropic-sdk-go",
	}

	matches := applyRules(deps)

	// Check frameworks.
	frameworkTags := tagNames(matches["frameworks"])
	assertContains(t, frameworkTags, "mcpkit", "frameworks should contain mcpkit")
	assertContains(t, frameworkTags, "mcp-go", "frameworks should contain mcp-go")
	assertContains(t, frameworkTags, "bubbletea", "frameworks should contain bubbletea")

	// Check protocols (mcp-go should also appear here).
	protocolTags := tagNames(matches["protocols"])
	assertContains(t, protocolTags, "mcp", "protocols should contain mcp")

	// Check datastores.
	datastoreTags := tagNames(matches["datastores"])
	assertContains(t, datastoreTags, "sqlite", "datastores should contain sqlite")

	// Check cloud.
	cloudTags := tagNames(matches["cloud"])
	assertContains(t, cloudTags, "aws", "cloud should contain aws")

	// Check AI.
	aiTags := tagNames(matches["ai"])
	assertContains(t, aiTags, "anthropic", "ai should contain anthropic")

	// Check observability.
	obsTags := tagNames(matches["observability"])
	assertContains(t, obsTags, "opentelemetry", "observability should contain opentelemetry")
}

// TestApplyRulesEmpty verifies that empty input returns empty results.
func TestApplyRulesEmpty(t *testing.T) {
	matches := applyRules(nil)
	for category, bucket := range matches {
		if len(bucket) != 0 {
			t.Errorf("expected empty %s bucket, got %d entries", category, len(bucket))
		}
	}
}

// TestApplyRulesNoDuplicates verifies that the same tag is not added twice
// within the same category.
func TestApplyRulesNoDuplicates(t *testing.T) {
	// mcp-go has entries in both "framework" and "protocol" categories,
	// but within each category it should only appear once.
	deps := []string{
		"github.com/mark3labs/mcp-go",
		"github.com/mark3labs/mcp-go/mcp",
	}
	matches := applyRules(deps)

	for category, bucket := range matches {
		seen := make(map[string]bool)
		for _, m := range bucket {
			if seen[m.Name] {
				t.Errorf("duplicate tag %q in category %s", m.Name, category)
			}
			seen[m.Name] = true
		}
	}
}

// TestCollectTags verifies tag collection from matches.
func TestCollectTags(t *testing.T) {
	matches := applyRules([]string{
		"github.com/hairglasses-studio/mcpkit",
		"github.com/mattn/go-sqlite3",
	})
	tags := collectTags(matches)
	if len(tags) == 0 {
		t.Fatal("expected non-empty tags")
	}
	assertContains(t, tags, "mcpkit", "tags should contain mcpkit")
	assertContains(t, tags, "sqlite", "tags should contain sqlite")
}

// TestClassifyRepo verifies repo type classification heuristics.
func TestClassifyRepo(t *testing.T) {
	// Create a temp directory structure to test classification.
	tmpDir := t.TempDir()

	// Test: library (has .go files, no main.go, no cmd/)
	libDir := filepath.Join(tmpDir, "mylib")
	os.MkdirAll(filepath.Join(libDir, "pkg"), 0o755)
	os.WriteFile(filepath.Join(libDir, "lib.go"), []byte("package mylib\n"), 0o644)
	os.WriteFile(filepath.Join(libDir, "pkg", "helper.go"), []byte("package pkg\n"), 0o644)

	result := classifyRepo(libDir, nil, nil)
	if result != "library" {
		t.Errorf("classifyRepo(library) = %q, want %q", result, "library")
	}

	// Test: cli (has cmd/ with main.go)
	cliDir := filepath.Join(tmpDir, "mycli")
	os.MkdirAll(filepath.Join(cliDir, "cmd", "app"), 0o755)
	os.WriteFile(filepath.Join(cliDir, "cmd", "app", "main.go"), []byte("package main\n"), 0o644)

	result = classifyRepo(cliDir, nil, nil)
	if result != "cli" {
		t.Errorf("classifyRepo(cli) = %q, want %q", result, "cli")
	}

	// Test: tui (has bubbletea dep)
	tuiDir := filepath.Join(tmpDir, "mytui")
	os.MkdirAll(filepath.Join(tuiDir, "cmd", "tui"), 0o755)
	os.WriteFile(filepath.Join(tuiDir, "cmd", "tui", "main.go"), []byte("package main\n"), 0o644)

	result = classifyRepo(tuiDir, []string{"github.com/charmbracelet/bubbletea"}, nil)
	if result != "tui" {
		t.Errorf("classifyRepo(tui) = %q, want %q", result, "tui")
	}

	// Test: server (has cmd/ and gin framework)
	srvDir := filepath.Join(tmpDir, "myserver")
	os.MkdirAll(filepath.Join(srvDir, "cmd", "server"), 0o755)
	os.WriteFile(filepath.Join(srvDir, "cmd", "server", "main.go"), []byte("package main\n"), 0o644)

	result = classifyRepo(srvDir, []string{"github.com/gin-gonic/gin"}, nil)
	if result != "server" {
		t.Errorf("classifyRepo(server) = %q, want %q", result, "server")
	}

	// Test: python-project
	pyDir := filepath.Join(tmpDir, "mypy")
	os.MkdirAll(pyDir, 0o755)
	os.WriteFile(filepath.Join(pyDir, "main.py"), []byte("print('hello')\n"), 0o644)

	result = classifyRepo(pyDir, nil, []LanguageInfo{{Name: "Python", Files: 10, Percentage: 80}})
	if result != "python-project" {
		t.Errorf("classifyRepo(python) = %q, want %q", result, "python-project")
	}

	// Test: container (has Dockerfile, no other signals)
	dockerDir := filepath.Join(tmpDir, "mydocker")
	os.MkdirAll(dockerDir, 0o755)
	os.WriteFile(filepath.Join(dockerDir, "Dockerfile"), []byte("FROM alpine\n"), 0o644)

	result = classifyRepo(dockerDir, nil, nil)
	if result != "container" {
		t.Errorf("classifyRepo(container) = %q, want %q", result, "container")
	}

	// Test: standalone (has main.go in root)
	standaloneDir := filepath.Join(tmpDir, "mystandalone")
	os.MkdirAll(standaloneDir, 0o755)
	os.WriteFile(filepath.Join(standaloneDir, "main.go"), []byte("package main\n"), 0o644)

	result = classifyRepo(standaloneDir, nil, nil)
	if result != "standalone" {
		t.Errorf("classifyRepo(standalone) = %q, want %q", result, "standalone")
	}
}

// TestParseGoMod verifies go.mod parsing.
func TestParseGoMod(t *testing.T) {
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")

	content := `module github.com/example/test

go 1.26.1

require (
	github.com/hairglasses-studio/mcpkit v0.2.0
	github.com/mark3labs/mcp-go v0.46.0
)

require (
	github.com/indirect/dep v1.0.0 // indirect
)

replace github.com/hairglasses-studio/mcpkit => ../mcpkit
`
	os.WriteFile(goModPath, []byte(content), 0o644)

	info, err := parseGoMod(goModPath)
	if err != nil {
		t.Fatalf("parseGoMod failed: %v", err)
	}

	if info.ModulePath != "github.com/example/test" {
		t.Errorf("ModulePath = %q, want %q", info.ModulePath, "github.com/example/test")
	}
	if info.GoVersion != "1.26.1" {
		t.Errorf("GoVersion = %q, want %q", info.GoVersion, "1.26.1")
	}
	if len(info.DirectDeps) < 2 {
		t.Errorf("DirectDeps count = %d, want at least 2", len(info.DirectDeps))
	}
	if info.ReplaceCount != 1 {
		t.Errorf("ReplaceCount = %d, want 1", info.ReplaceCount)
	}
}

// TestScanRepoReal scans the hg-mcp repo itself (integration test).
func TestScanRepoReal(t *testing.T) {
	// Find the hg-mcp repo path by going up from the test directory.
	// This test file is at internal/mcp/tools/repo_analysis/module_test.go
	// so the repo root is 4 levels up.
	wd, err := os.Getwd()
	if err != nil {
		t.Skip("cannot determine working directory")
	}
	repoRoot := filepath.Join(wd, "..", "..", "..", "..")
	repoRoot, err = filepath.Abs(repoRoot)
	if err != nil {
		t.Skip("cannot resolve repo root")
	}

	// Verify it looks like the hg-mcp repo.
	if !fileExists(filepath.Join(repoRoot, "go.mod")) {
		t.Skip("not running from hg-mcp repo tree")
	}

	result, err := scanRepo(repoRoot, false)
	if err != nil {
		t.Fatalf("scanRepo failed: %v", err)
	}

	if result.Name == "" {
		t.Error("result.Name should not be empty")
	}
	if result.Path == "" {
		t.Error("result.Path should not be empty")
	}
	if result.ModulePath == "" {
		t.Error("result.ModulePath should not be empty for a Go repo")
	}
	if len(result.Languages) == 0 {
		t.Error("result.Languages should not be empty")
	}
	if result.ScannedAt == "" {
		t.Error("result.ScannedAt should not be empty")
	}

	// hg-mcp should be classified as something reasonable.
	validTypes := map[string]bool{
		"mcp-server": true, "cli": true, "server": true, "library": true,
	}
	if !validTypes[result.Type] {
		t.Errorf("result.Type = %q, expected one of mcp-server/cli/server/library", result.Type)
	}

	// Should have Go as a primary language.
	found := false
	for _, lang := range result.Languages {
		if lang.Name == "Go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("hg-mcp should have Go as a detected language")
	}
}

// TestDetectLanguages verifies language detection on a synthetic directory.
func TestDetectLanguages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files of various types.
	files := map[string]string{
		"main.go":       "package main\n",
		"lib.go":        "package lib\n",
		"util_test.go":  "package lib\n",
		"app.ts":        "console.log('hi');\n",
		"style.css":     "body { color: red; }\n",
		"README.md":     "# Hello\n",
		"config.yaml":   "key: value\n",
		"data.json":     "{}",
		".hidden/f.go":  "package hidden\n",
		"vendor/v.go":   "package vendor\n",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		os.MkdirAll(filepath.Dir(path), 0o755)
		os.WriteFile(path, []byte(content), 0o644)
	}

	counts, total := detectLanguages(tmpDir)

	// .hidden/ and vendor/ should be skipped.
	if counts["Go"] != 3 {
		t.Errorf("Go file count = %d, want 3", counts["Go"])
	}
	if counts["TypeScript"] != 1 {
		t.Errorf("TypeScript file count = %d, want 1", counts["TypeScript"])
	}
	// Go:3, TypeScript:1, CSS:1, Markdown:1, YAML:1, JSON:1 = 8
	if total != 8 {
		t.Errorf("total = %d, want 8 (go:3, ts:1, css:1, md:1, yaml:1, json:1)", total)
	}

	// Verify hidden and vendor were excluded.
	langs := buildLanguageInfo(counts, total)
	if len(langs) == 0 {
		t.Fatal("expected non-empty language list")
	}
	// First language should be Go (most files).
	if langs[0].Name != "Go" {
		t.Errorf("primary language = %q, want Go", langs[0].Name)
	}
}

// TestHandleRepoScanMissingPath verifies error handling for missing repo path.
func TestHandleRepoScanMissingPath(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleRepoScan(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected error result for missing repo_path")
	}
}

// TestHandleRepoScanNonexistentPath verifies error handling for a nonexistent path.
func TestHandleRepoScanNonexistentPath(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"repo_path": "/tmp/definitely-does-not-exist-" + t.Name(),
	}

	result, err := handleRepoScan(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.IsError {
		t.Error("expected error result for nonexistent path")
	}
}

// TestHandleRepoScanInvalidMode verifies error handling for invalid mode.
func TestHandleRepoScanInvalidMode(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"repo_path": "/tmp",
		"mode":      "invalid",
	}

	result, err := handleRepoScan(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for invalid mode")
	}
}

// TestHandleRepoCompareMissingRepos verifies error handling.
func TestHandleRepoCompareMissingRepos(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleRepoCompare(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for missing repos param")
	}
}

// TestHandleRepoCompareSingleRepo verifies minimum repo count requirement.
func TestHandleRepoCompareSingleRepo(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"repos": "single-repo",
	}

	result, err := handleRepoCompare(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error result for single repo comparison")
	}
}

// TestResolveRepoPath verifies path resolution logic.
func TestResolveRepoPath(t *testing.T) {
	// Name without slash should be resolved to ~/hairglasses-studio/<name>.
	resolved := resolveRepoPath("mcpkit")
	if !filepath.IsAbs(resolved) {
		t.Errorf("resolved path should be absolute, got %q", resolved)
	}
	if filepath.Base(resolved) != "mcpkit" {
		t.Errorf("resolved path base = %q, want %q", filepath.Base(resolved), "mcpkit")
	}

	// Absolute path should be returned as-is.
	abs := "/home/user/repos/test"
	resolved = resolveRepoPath(abs)
	if resolved != abs {
		t.Errorf("resolveRepoPath(%q) = %q, want %q", abs, resolved, abs)
	}
}

// TestDetectBuildSystems verifies build system detection.
func TestDetectBuildSystems(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some build system files.
	os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte("all:\n"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM alpine\n"), 0o644)
	os.MkdirAll(filepath.Join(tmpDir, ".github", "workflows"), 0o755)

	systems := detectBuildSystems(tmpDir)

	found := make(map[string]bool)
	for _, s := range systems {
		found[s] = true
	}

	if !found["make"] {
		t.Error("expected 'make' in build systems")
	}
	if !found["docker"] {
		t.Error("expected 'docker' in build systems")
	}
	if !found["github-actions"] {
		t.Error("expected 'github-actions' in build systems")
	}
}

// TestScoreRelevance verifies relevance scoring logic.
func TestScoreRelevance(t *testing.T) {
	// mcpkit import -> high
	r := scoreRelevance([]string{"github.com/hairglasses-studio/mcpkit"}, nil)
	if r != "high" {
		t.Errorf("scoreRelevance(mcpkit) = %q, want %q", r, "high")
	}

	// Go files but no mcpkit -> medium
	r = scoreRelevance(nil, []LanguageInfo{{Name: "Go", Files: 10}})
	if r != "medium" {
		t.Errorf("scoreRelevance(go) = %q, want %q", r, "medium")
	}

	// Neither -> low
	r = scoreRelevance(nil, []LanguageInfo{{Name: "Python", Files: 10}})
	if r != "low" {
		t.Errorf("scoreRelevance(python) = %q, want %q", r, "low")
	}
}

// TestScanTempRepo scans a synthetic repo to verify end-to-end output.
func TestScanTempRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal Go repo.
	goMod := `module github.com/example/test

go 1.26.1

require (
	github.com/hairglasses-studio/mcpkit v0.2.0
	github.com/mark3labs/mcp-go v0.46.0
)
`
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte("build:\n\tgo build\n"), 0o644)

	result, err := scanRepo(tmpDir, false)
	if err != nil {
		t.Fatalf("scanRepo failed: %v", err)
	}

	if result.ModulePath != "github.com/example/test" {
		t.Errorf("ModulePath = %q, want %q", result.ModulePath, "github.com/example/test")
	}
	if result.GoVersion != "1.26.1" {
		t.Errorf("GoVersion = %q, want %q", result.GoVersion, "1.26.1")
	}
	if result.DirectDeps != 2 {
		t.Errorf("DirectDeps = %d, want 2", result.DirectDeps)
	}
	if result.Relevance != "high" {
		t.Errorf("Relevance = %q, want %q (mcpkit dep)", result.Relevance, "high")
	}
	if len(result.BuildSystems) == 0 {
		t.Error("expected at least one build system (make)")
	}
	if len(result.Tags) == 0 {
		t.Error("expected non-empty tags")
	}
}

// TestCountLOC verifies line counting.
func TestCountLOC(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a 5-line Go file.
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n"), 0o644)

	langs := map[string]int{"Go": 1}
	loc := countLOC(tmpDir, langs)

	if loc == nil {
		t.Fatal("expected non-nil LOC info")
	}
	if loc.Total == 0 {
		t.Error("expected non-zero total LOC")
	}
	if loc.ByLanguage["Go"] == 0 {
		t.Error("expected non-zero Go LOC")
	}
}

// --- helpers ---

func tagNames(matches []TagMatch) []string {
	names := make([]string, len(matches))
	for i, m := range matches {
		names[i] = m.Name
	}
	return names
}

func assertContains(t *testing.T, slice []string, item string, msg string) {
	t.Helper()
	for _, s := range slice {
		if s == item {
			return
		}
	}
	t.Errorf("%s: %v does not contain %q", msg, slice, item)
}

// Ensure tools package is used (satisfies import).
var _ = tools.TextResult
