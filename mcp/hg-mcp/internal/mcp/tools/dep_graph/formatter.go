package dep_graph

import (
	"fmt"
	"sort"
	"strings"
)

// shortName extracts a short display name from a full module path.
// E.g., "github.com/hairglasses-studio/mcpkit" -> "mcpkit"
//
//	"github.com/hairglasses-studio/dotfiles/mcp/dotfiles-mcp" -> "dotfiles/mcp/dotfiles-mcp"
func shortName(modPath string) string {
	for _, prefix := range orgPatterns {
		if strings.HasPrefix(modPath, prefix) {
			return strings.TrimPrefix(modPath, prefix)
		}
	}
	// For external modules, return the last two segments.
	parts := strings.Split(modPath, "/")
	if len(parts) > 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return modPath
}

// sanitizeMermaidID makes a string safe for use as a Mermaid node ID.
func sanitizeMermaidID(s string) string {
	r := strings.NewReplacer("/", "_", "-", "_", ".", "_", "@", "_")
	return r.Replace(s)
}

// FormatMermaid renders edges and replaces as a Mermaid flowchart.
func FormatMermaid(edges []Edge, replaces []Replace) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Collect unique nodes.
	nodes := make(map[string]bool)
	for _, e := range edges {
		nodes[e.From] = true
		nodes[e.To] = true
	}

	// Declare nodes.
	sortedNodes := make([]string, 0, len(nodes))
	for n := range nodes {
		sortedNodes = append(sortedNodes, n)
	}
	sort.Strings(sortedNodes)
	for _, n := range sortedNodes {
		id := sanitizeMermaidID(shortName(n))
		label := shortName(n)
		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", id, label))
	}

	// Draw edges.
	for _, e := range edges {
		fromID := sanitizeMermaidID(shortName(e.From))
		toID := sanitizeMermaidID(shortName(e.To))
		label := e.ToVersion
		if label == "" {
			label = "local"
		}
		sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", fromID, label, toID))
	}

	// Draw replace directives as dashed edges.
	for _, r := range replaces {
		fromID := sanitizeMermaidID(r.Repo)
		toID := sanitizeMermaidID(shortName(r.Old))
		label := "replace"
		if r.IsLocal {
			label = "replace (local)"
		}
		sb.WriteString(fmt.Sprintf("    %s -.->|%s| %s\n", fromID, label, toID))
	}

	return sb.String()
}

// FormatDOT renders edges and replaces as a Graphviz DOT digraph.
func FormatDOT(edges []Edge, replaces []Replace) string {
	var sb strings.Builder
	sb.WriteString("digraph deps {\n")
	sb.WriteString("    rankdir=TB;\n")
	sb.WriteString("    node [shape=box style=\"rounded,filled\" fillcolor=\"#0c5525\" fontcolor=white];\n")

	// Collect unique nodes.
	nodes := make(map[string]bool)
	for _, e := range edges {
		nodes[e.From] = true
		nodes[e.To] = true
	}
	sortedNodes := make([]string, 0, len(nodes))
	for n := range nodes {
		sortedNodes = append(sortedNodes, n)
	}
	sort.Strings(sortedNodes)
	for _, n := range sortedNodes {
		sb.WriteString(fmt.Sprintf("    \"%s\";\n", shortName(n)))
	}

	// Draw edges.
	for _, e := range edges {
		label := e.ToVersion
		if label == "" {
			label = "local"
		}
		sb.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\" [label=\"%s\"];\n",
			shortName(e.From), shortName(e.To), label))
	}

	// Draw replace directives as dashed edges.
	for _, r := range replaces {
		label := "replace"
		if r.IsLocal {
			label = "replace (local)"
		}
		sb.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\" [label=\"%s\" style=dashed];\n",
			r.Repo, shortName(r.Old), label))
	}

	sb.WriteString("}\n")
	return sb.String()
}
