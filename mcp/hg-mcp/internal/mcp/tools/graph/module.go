// Package graph provides knowledge graph tools for hg-mcp.
package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewKnowledgeGraphClient)

// Module implements the ToolModule interface for knowledge graph tools
type Module struct{}

func (m *Module) Name() string {
	return "graph"
}

func (m *Module) Description() string {
	return "Knowledge graph tools for semantic search and context discovery"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_graph_rebuild",
				mcp.WithDescription("Rebuild the knowledge graph from vault documents. Extracts [[wiki links]] and #tags to build semantic connections."),
			),
			Handler:             handleGraphRebuild,
			Category:            "graph",
			Subcategory:         "maintenance",
			Tags:                []string{"graph", "rebuild", "index", "vault"},
			UseCases:            []string{"Update graph after vault changes", "Initial graph build"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "graph",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_graph_search",
				mcp.WithDescription("Graph-enhanced search: finds documents and related content through semantic connections."),
				mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
				mcp.WithNumber("max_hops", mcp.Description("Maximum graph hops (default: 2)")),
			),
			Handler:             handleGraphSearch,
			Category:            "graph",
			Subcategory:         "search",
			Tags:                []string{"graph", "search", "semantic", "query"},
			UseCases:            []string{"Find related documents", "Discover connections"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "graph",
		},
		{
			Tool: mcp.NewTool("aftrs_graph_insights",
				mcp.WithDescription("Get analytics about the knowledge graph: node counts, clusters, orphans, most connected documents."),
			),
			Handler:             handleGraphInsights,
			Category:            "graph",
			Subcategory:         "analytics",
			Tags:                []string{"graph", "analytics", "insights", "stats"},
			UseCases:            []string{"Understand vault structure", "Find orphan documents"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "graph",
		},
		{
			Tool: mcp.NewTool("aftrs_context_from_graph",
				mcp.WithDescription("Get related context for a document via graph connections."),
				mcp.WithString("document", mcp.Description("Document path or name"), mcp.Required()),
				mcp.WithNumber("max_results", mcp.Description("Maximum results (default: 10)")),
			),
			Handler:             handleContextFromGraph,
			Category:            "graph",
			Subcategory:         "context",
			Tags:                []string{"graph", "context", "related", "connections"},
			UseCases:            []string{"Get context for a show", "Find related runbooks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "graph",
		},
		{
			Tool: mcp.NewTool("aftrs_similar_shows",
				mcp.WithDescription("Find shows similar to given criteria (tags, equipment, venue)."),
				mcp.WithString("criteria", mcp.Description("Search criteria (tags, equipment, venue, etc.)"), mcp.Required()),
			),
			Handler:             handleSimilarShows,
			Category:            "graph",
			Subcategory:         "shows",
			Tags:                []string{"graph", "shows", "similar", "history"},
			UseCases:            []string{"Find past shows like this", "Reference similar events"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "graph",
		},
		{
			Tool: mcp.NewTool("aftrs_resolution_path",
				mcp.WithDescription("Find how similar issues were resolved before based on learnings and runbooks."),
				mcp.WithString("issue", mcp.Description("Issue or problem description"), mcp.Required()),
			),
			Handler:             handleResolutionPath,
			Category:            "graph",
			Subcategory:         "resolution",
			Tags:                []string{"graph", "resolution", "troubleshooting", "history"},
			UseCases:            []string{"Find past solutions", "Troubleshoot issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "graph",
		},
	}
}

// handleGraphRebuild handles the aftrs_graph_rebuild tool
func handleGraphRebuild(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create graph client: %w", err)), nil
	}

	if err := client.RebuildGraph(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to rebuild graph: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Knowledge Graph Rebuilt\n\n")
	sb.WriteString(fmt.Sprintf("**Vault Path:** `%s`\n", client.VaultPath()))
	sb.WriteString(fmt.Sprintf("**Nodes:** %d\n", client.NodeCount()))
	sb.WriteString(fmt.Sprintf("**Edges:** %d\n\n", client.EdgeCount()))
	sb.WriteString("Graph ready for semantic search and context discovery.")

	return tools.TextResult(sb.String()), nil
}

// handleGraphSearch handles the aftrs_graph_search tool
func handleGraphSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}
	maxHops := tools.GetIntParam(req, "max_hops", 2)

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create graph client: %w", err)), nil
	}

	results, err := client.Search(ctx, query, maxHops)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Graph Search: \"%s\"\n\n", query))
	sb.WriteString(fmt.Sprintf("Found **%d** results (max %d hops):\n\n", len(results), maxHops))

	if len(results) == 0 {
		sb.WriteString("No matches found. Try rebuilding the graph with `aftrs_graph_rebuild`.\n")
	} else {
		sb.WriteString("| Document | Type | Score | Hops | Relevance |\n")
		sb.WriteString("|----------|------|-------|------|------------|\n")

		for i, r := range results {
			if i >= 20 {
				sb.WriteString(fmt.Sprintf("\n... and %d more results\n", len(results)-20))
				break
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %.2f | %d | %s |\n",
				r.Node.Title, r.Node.Type, r.Score, r.Hops, r.Relevance))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleGraphInsights handles the aftrs_graph_insights tool
func handleGraphInsights(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create graph client: %w", err)), nil
	}

	insights, err := client.GetInsights(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get insights: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Knowledge Graph Insights\n\n")
	sb.WriteString(fmt.Sprintf("**Total Nodes:** %d\n", insights.TotalNodes))
	sb.WriteString(fmt.Sprintf("**Total Edges:** %d\n\n", insights.TotalEdges))

	// Nodes by type
	if len(insights.NodesByType) > 0 {
		sb.WriteString("## Nodes by Type\n\n")
		sb.WriteString("| Type | Count |\n")
		sb.WriteString("|------|-------|\n")
		for nodeType, count := range insights.NodesByType {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", nodeType, count))
		}
		sb.WriteString("\n")
	}

	// Edges by type
	if len(insights.EdgesByType) > 0 {
		sb.WriteString("## Edges by Type\n\n")
		sb.WriteString("| Type | Count |\n")
		sb.WriteString("|------|-------|\n")
		for edgeType, count := range insights.EdgesByType {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", edgeType, count))
		}
		sb.WriteString("\n")
	}

	// Most connected
	if len(insights.MostConnected) > 0 {
		sb.WriteString("## Most Connected Documents\n\n")
		sb.WriteString("| Document | Type |\n")
		sb.WriteString("|----------|------|\n")
		for _, node := range insights.MostConnected {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", node.Title, node.Type))
		}
		sb.WriteString("\n")
	}

	// Orphans
	if len(insights.Orphans) > 0 {
		sb.WriteString("## Orphan Documents (No Connections)\n\n")
		for i, node := range insights.Orphans {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("... and %d more orphans\n", len(insights.Orphans)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", node.Title, node.Type))
		}
		sb.WriteString("\n")
	}

	// Recently updated
	if len(insights.RecentlyUpdated) > 0 {
		sb.WriteString("## Recently Updated\n\n")
		for _, node := range insights.RecentlyUpdated {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", node.Title, node.Modified.Format("2006-01-02")))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleContextFromGraph handles the aftrs_context_from_graph tool
func handleContextFromGraph(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	document, errResult := tools.RequireStringParam(req, "document")
	if errResult != nil {
		return errResult, nil
	}
	maxResults := tools.GetIntParam(req, "max_results", 10)

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create graph client: %w", err)), nil
	}

	results, err := client.GetContext(ctx, document, maxResults)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get context: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Related Context: %s\n\n", document))
	sb.WriteString(fmt.Sprintf("Found **%d** related documents:\n\n", len(results)))

	if len(results) == 0 {
		sb.WriteString("No related documents found. The document may not be in the graph or has no connections.\n")
	} else {
		sb.WriteString("| Document | Type | Connection | Weight |\n")
		sb.WriteString("|----------|------|------------|--------|\n")

		for _, r := range results {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %.2f |\n",
				r.Node.Title, r.Node.Type, r.Relevance, r.Score))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleSimilarShows handles the aftrs_similar_shows tool
func handleSimilarShows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	criteria, errResult := tools.RequireStringParam(req, "criteria")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create graph client: %w", err)), nil
	}

	shows, err := client.FindSimilarShows(ctx, criteria)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to find similar shows: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Similar Shows: \"%s\"\n\n", criteria))
	sb.WriteString(fmt.Sprintf("Found **%d** similar shows:\n\n", len(shows)))

	if len(shows) == 0 {
		sb.WriteString("No similar shows found. Try different criteria or add show documents to the vault.\n")
	} else {
		sb.WriteString("| Show | Similarity | Shared Tags |\n")
		sb.WriteString("|------|------------|-------------|\n")

		for i, s := range shows {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("\n... and %d more shows\n", len(shows)-10))
				break
			}
			tags := strings.Join(s.SharedTags, ", ")
			if tags == "" {
				tags = "-"
			}
			sb.WriteString(fmt.Sprintf("| %s | %.0f%% | %s |\n",
				s.Show.Title, s.Similarity*100, tags))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleResolutionPath handles the aftrs_resolution_path tool
func handleResolutionPath(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issue, errResult := tools.RequireStringParam(req, "issue")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create graph client: %w", err)), nil
	}

	paths, err := client.FindResolutionPath(ctx, issue)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to find resolution paths: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Resolution Paths: \"%s\"\n\n", issue))
	sb.WriteString(fmt.Sprintf("Found **%d** potential resolutions:\n\n", len(paths)))

	if len(paths) == 0 {
		sb.WriteString("No past resolutions found. This may be a new issue or try different search terms.\n")
	} else {
		for i, p := range paths {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("\n... and %d more resolutions\n", len(paths)-5))
				break
			}

			confidence := "low"
			emoji := "🟡"
			if p.Confidence >= 0.7 {
				confidence = "high"
				emoji = "🟢"
			} else if p.Confidence >= 0.4 {
				confidence = "medium"
				emoji = "🟠"
			}

			sb.WriteString(fmt.Sprintf("### %d. %s %s (%.0f%% confidence)\n\n", i+1, emoji, p.Resolution, p.Confidence*100))
			sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", p.Source))
			sb.WriteString(fmt.Sprintf("**Confidence:** %s\n\n", confidence))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
