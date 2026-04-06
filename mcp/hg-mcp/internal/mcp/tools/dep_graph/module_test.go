package dep_graph

import (
	"context"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

// --- Edge parsing tests ---

func TestParseModGraphLine(t *testing.T) {
	tests := []struct {
		line    string
		from    string
		fromVer string
		to      string
		toVer   string
		ok      bool
	}{
		{
			line:    "github.com/hairglasses-studio/hg-mcp@v0.0.0 github.com/hairglasses-studio/mcpkit@v0.2.0",
			from:    "github.com/hairglasses-studio/hg-mcp",
			fromVer: "v0.0.0",
			to:      "github.com/hairglasses-studio/mcpkit",
			toVer:   "v0.2.0",
			ok:      true,
		},
		{
			line:    "github.com/hairglasses-studio/mcpkit@v0.2.0 github.com/mark3labs/mcp-go@v0.46.0",
			from:    "github.com/hairglasses-studio/mcpkit",
			fromVer: "v0.2.0",
			to:      "github.com/mark3labs/mcp-go",
			toVer:   "v0.46.0",
			ok:      true,
		},
		{
			line: "",
			ok:   false,
		},
		{
			line: "single-field",
			ok:   false,
		},
		{
			line:    "github.com/foo github.com/bar",
			from:    "github.com/foo",
			fromVer: "",
			to:      "github.com/bar",
			toVer:   "",
			ok:      true,
		},
	}

	for _, tt := range tests {
		from, fromVer, to, toVer, ok := parseModGraphLine(tt.line)
		if ok != tt.ok {
			t.Errorf("parseModGraphLine(%q): ok = %v, want %v", tt.line, ok, tt.ok)
			continue
		}
		if !ok {
			continue
		}
		if from != tt.from || fromVer != tt.fromVer || to != tt.to || toVer != tt.toVer {
			t.Errorf("parseModGraphLine(%q) = (%q, %q, %q, %q), want (%q, %q, %q, %q)",
				tt.line, from, fromVer, to, toVer, tt.from, tt.fromVer, tt.to, tt.toVer)
		}
	}
}

func TestSplitModVersion(t *testing.T) {
	tests := []struct {
		input   string
		mod     string
		version string
	}{
		{"github.com/foo/bar@v1.2.3", "github.com/foo/bar", "v1.2.3"},
		{"github.com/foo/bar", "github.com/foo/bar", ""},
		{"foo@v0.0.0-20240101", "foo", "v0.0.0-20240101"},
	}

	for _, tt := range tests {
		mod, ver := splitModVersion(tt.input)
		if mod != tt.mod || ver != tt.version {
			t.Errorf("splitModVersion(%q) = (%q, %q), want (%q, %q)",
				tt.input, mod, ver, tt.mod, tt.version)
		}
	}
}

func TestIsOrgModule(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"github.com/hairglasses-studio/mcpkit", true},
		{"github.com/glasshairs/something", true},
		{"github.com/aftrs-studio/hg-mcp", true},
		{"github.com/mark3labs/mcp-go", false},
		{"golang.org/x/net", false},
	}

	for _, tt := range tests {
		got := isOrgModule(tt.path)
		if got != tt.want {
			t.Errorf("isOrgModule(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestFilterEdgesInternal(t *testing.T) {
	edges := []Edge{
		{From: "a", To: "github.com/hairglasses-studio/mcpkit", ToVersion: "v0.2.0"},
		{From: "a", To: "github.com/mark3labs/mcp-go", ToVersion: "v0.46.0"},
		{From: "b", To: "github.com/glasshairs/x", ToVersion: "v1.0.0"},
	}

	filtered := FilterEdgesInternal(edges)
	if len(filtered) != 2 {
		t.Fatalf("FilterEdgesInternal: got %d edges, want 2", len(filtered))
	}
	if filtered[0].To != "github.com/hairglasses-studio/mcpkit" {
		t.Errorf("expected mcpkit edge first, got %s", filtered[0].To)
	}
	if filtered[1].To != "github.com/glasshairs/x" {
		t.Errorf("expected glasshairs edge second, got %s", filtered[1].To)
	}
}

func TestFilterEdgesByRepo(t *testing.T) {
	edges := []Edge{
		{From: "a", To: "b", SourceRepo: "mcpkit"},
		{From: "c", To: "d", SourceRepo: "dotfiles"},
		{From: "e", To: "f", SourceRepo: "mcpkit"},
	}

	filtered := FilterEdgesByRepo(edges, "mcpkit")
	if len(filtered) != 2 {
		t.Fatalf("FilterEdgesByRepo: got %d edges, want 2", len(filtered))
	}
}

func TestCountUniqueRepos(t *testing.T) {
	edges := []Edge{
		{SourceRepo: "a"},
		{SourceRepo: "b"},
		{SourceRepo: "a"},
		{SourceRepo: "c"},
	}
	if got := CountUniqueRepos(edges); got != 3 {
		t.Errorf("CountUniqueRepos: got %d, want 3", got)
	}
}

// --- Formatter tests ---

func TestFormatMermaid(t *testing.T) {
	edges := []Edge{
		{
			From:        "github.com/hairglasses-studio/dotfiles-mcp",
			FromVersion: "v0.0.0",
			To:          "github.com/hairglasses-studio/mcpkit",
			ToVersion:   "v0.2.0",
		},
		{
			From:        "github.com/hairglasses-studio/ralphglasses",
			FromVersion: "v0.0.0",
			To:          "github.com/hairglasses-studio/mcpkit",
			ToVersion:   "v0.4.1",
		},
	}
	replaces := []Replace{
		{Repo: "dotfiles-mcp", Old: "github.com/hairglasses-studio/mcpkit", New: "../mcpkit", IsLocal: true},
	}

	out := FormatMermaid(edges, replaces)

	if !containsAll(out, "graph TD", "mcpkit", "dotfiles_mcp", "ralphglasses", "v0.2.0", "v0.4.1", "replace (local)") {
		t.Errorf("FormatMermaid output missing expected content:\n%s", out)
	}
}

func TestFormatDOT(t *testing.T) {
	edges := []Edge{
		{
			From:      "github.com/hairglasses-studio/hg-mcp",
			To:        "github.com/hairglasses-studio/mcpkit",
			ToVersion: "v0.2.0",
		},
	}

	out := FormatDOT(edges, nil)

	if !containsAll(out, "digraph deps", "rankdir=TB", "mcpkit", "hg-mcp", "v0.2.0") {
		t.Errorf("FormatDOT output missing expected content:\n%s", out)
	}
}

func TestShortName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"github.com/hairglasses-studio/mcpkit", "mcpkit"},
		{"github.com/glasshairs/foo", "foo"},
		{"github.com/aftrs-studio/hg-mcp", "hg-mcp"},
		{"github.com/mark3labs/mcp-go", "mark3labs/mcp-go"},
		{"golang.org/x/net", "x/net"},
	}

	for _, tt := range tests {
		got := shortName(tt.input)
		if got != tt.want {
			t.Errorf("shortName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Drift detection tests ---

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int // >0, 0, <0
	}{
		{"v0.4.1", "v0.2.0", 1},
		{"v0.2.0", "v0.4.1", -1},
		{"v1.0.0", "v1.0.0", 0},
		{"v2.0.0", "v1.9.9", 1},
		{"v0.0.1", "v0.0.2", -1},
	}

	for _, tt := range tests {
		got := compareVersions(tt.a, tt.b)
		switch {
		case tt.want > 0 && got <= 0:
			t.Errorf("compareVersions(%q, %q) = %d, want > 0", tt.a, tt.b, got)
		case tt.want < 0 && got >= 0:
			t.Errorf("compareVersions(%q, %q) = %d, want < 0", tt.a, tt.b, got)
		case tt.want == 0 && got != 0:
			t.Errorf("compareVersions(%q, %q) = %d, want 0", tt.a, tt.b, got)
		}
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
	}{
		{"v1.2.3", [3]int{1, 2, 3}},
		{"v0.4.1", [3]int{0, 4, 1}},
		{"1.0.0", [3]int{1, 0, 0}},
		{"v0.0.0-20240101120000-abcdef123456", [3]int{0, 0, 0}},
		{"v2.1", [3]int{2, 1, 0}},
	}

	for _, tt := range tests {
		got := parseVersion(tt.input)
		if got != tt.want {
			t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestClassifyDrift(t *testing.T) {
	tests := []struct {
		name     string
		versions map[string]string
		want     string
	}{
		{
			name:     "major drift",
			versions: map[string]string{"a": "v0.2.0", "b": "v1.0.0"},
			want:     "major",
		},
		{
			name:     "minor drift",
			versions: map[string]string{"a": "v0.2.0", "b": "v0.4.1"},
			want:     "minor",
		},
		{
			name:     "patch drift",
			versions: map[string]string{"a": "v0.2.0", "b": "v0.2.1"},
			want:     "patch",
		},
		{
			name:     "single version",
			versions: map[string]string{"a": "v1.0.0"},
			want:     "patch",
		},
	}

	for _, tt := range tests {
		got := classifyDrift(tt.versions)
		if got != tt.want {
			t.Errorf("%s: classifyDrift = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// --- Handler tests ---

func TestDepGraphInvalidFormat(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"format": "xml",
	}

	result, err := handleDepGraph(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid format")
	}
}

func TestDepDriftInvalidSeverity(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"severity": "critical",
	}

	result, err := handleDepDrift(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid severity")
	}
}

func TestDepGraphDefaultParams(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleDepGraph(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %+v", result)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Dep graph output: %.500s", content.Text)
}

func TestDepDriftDefaultParams(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleDepDrift(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("unexpected error result: %+v", result)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Dep drift output: %.500s", content.Text)
}

// --- Helpers ---

func containsAll(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
