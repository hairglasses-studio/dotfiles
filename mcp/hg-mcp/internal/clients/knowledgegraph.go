// Package clients provides API clients for external services.
package clients

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// KnowledgeGraphClient provides semantic graph operations over vault documents
type KnowledgeGraphClient struct {
	vaultPath string
	nodes     map[string]*GraphNode
	edges     []GraphEdge
}

// GraphNode represents a node in the knowledge graph
type GraphNode struct {
	ID       string            `json:"id"`
	Path     string            `json:"path"`
	Title    string            `json:"title"`
	Type     string            `json:"type"` // document, project, show, equipment, runbook
	Tags     []string          `json:"tags"`
	Links    []string          `json:"links"` // wiki-style links
	Created  time.Time         `json:"created"`
	Modified time.Time         `json:"modified"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GraphEdge represents a connection between nodes
type GraphEdge struct {
	Source   string            `json:"source"`
	Target   string            `json:"target"`
	Type     string            `json:"type"`
	Weight   float64           `json:"weight"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GraphSearchResult represents a graph-enhanced search result
type GraphSearchResult struct {
	Node      *GraphNode `json:"node"`
	Score     float64    `json:"score"`
	Hops      int        `json:"hops"`
	Path      []string   `json:"path,omitempty"`
	Relevance string     `json:"relevance"`
}

// GraphInsights represents analytics about the knowledge graph
type GraphInsights struct {
	TotalNodes      int            `json:"total_nodes"`
	TotalEdges      int            `json:"total_edges"`
	NodesByType     map[string]int `json:"nodes_by_type"`
	EdgesByType     map[string]int `json:"edges_by_type"`
	MostConnected   []GraphNode    `json:"most_connected"`
	Orphans         []GraphNode    `json:"orphans"`
	RecentlyUpdated []GraphNode    `json:"recently_updated"`
	Clusters        [][]string     `json:"clusters,omitempty"`
}

// SimilarShow represents a similar past show
type SimilarShow struct {
	Show            *GraphNode `json:"show"`
	Similarity      float64    `json:"similarity"`
	SharedTags      []string   `json:"shared_tags"`
	SharedEquipment []string   `json:"shared_equipment,omitempty"`
}

// ResolutionPath represents how an issue was resolved before
type ResolutionPath struct {
	Problem    string   `json:"problem"`
	Resolution string   `json:"resolution"`
	Source     string   `json:"source"`
	Confidence float64  `json:"confidence"`
	Steps      []string `json:"steps,omitempty"`
}

// NewKnowledgeGraphClient creates a new knowledge graph client
func NewKnowledgeGraphClient() (*KnowledgeGraphClient, error) {
	return &KnowledgeGraphClient{
		vaultPath: config.Get().AftrsVaultPath,
		nodes:     make(map[string]*GraphNode),
		edges:     []GraphEdge{},
	}, nil
}

// generateNodeID creates a unique ID for a node
func generateNodeID(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:8])
}

// RebuildGraph rebuilds the knowledge graph from vault documents
func (c *KnowledgeGraphClient) RebuildGraph(ctx context.Context) error {
	c.nodes = make(map[string]*GraphNode)
	c.edges = []GraphEdge{}

	if _, err := os.Stat(c.vaultPath); os.IsNotExist(err) {
		return nil // Empty graph if vault doesn't exist
	}

	// Walk vault and build nodes
	err := filepath.Walk(c.vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(c.vaultPath, path)
		node := c.parseDocument(path, relPath, info)
		if node != nil {
			c.nodes[node.ID] = node
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Build edges based on relationships
	c.buildEdges()

	return nil
}

// parseDocument parses a markdown document into a graph node
func (c *KnowledgeGraphClient) parseDocument(path, relPath string, info os.FileInfo) *GraphNode {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	node := &GraphNode{
		ID:       generateNodeID(relPath),
		Path:     relPath,
		Title:    strings.TrimSuffix(info.Name(), ".md"),
		Modified: info.ModTime(),
		Metadata: make(map[string]string),
	}

	// Determine type based on path
	switch {
	case strings.HasPrefix(relPath, "projects/"):
		node.Type = "project"
	case strings.HasPrefix(relPath, "shows/"):
		node.Type = "show"
	case strings.HasPrefix(relPath, "equipment/"):
		node.Type = "equipment"
	case strings.HasPrefix(relPath, "runbooks/"):
		node.Type = "runbook"
	case strings.HasPrefix(relPath, "learnings/"):
		node.Type = "learning"
	case strings.HasPrefix(relPath, "sessions/"):
		node.Type = "session"
	default:
		node.Type = "document"
	}

	// Extract wiki-style links [[link]]
	linkRegex := regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	matches := linkRegex.FindAllStringSubmatch(string(content), -1)
	for _, match := range matches {
		if len(match) > 1 {
			node.Links = append(node.Links, match[1])
		}
	}

	// Extract tags #tag
	tagRegex := regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)
	tagMatches := tagRegex.FindAllStringSubmatch(string(content), -1)
	for _, match := range tagMatches {
		if len(match) > 1 {
			node.Tags = append(node.Tags, match[1])
		}
	}

	return node
}

// buildEdges creates edges between nodes
func (c *KnowledgeGraphClient) buildEdges() {
	// Edge weights from roadmap
	edgeWeights := map[string]float64{
		"link":       1.0,
		"project":    1.0,
		"resolution": 1.0,
		"venue":      0.8,
		"equipment":  0.7,
		"symptom":    0.6,
		"tag":        0.5,
		"temporal":   0.3,
	}

	// Build link edges
	for _, node := range c.nodes {
		for _, link := range node.Links {
			// Find target node by title
			for _, target := range c.nodes {
				if strings.EqualFold(target.Title, link) || strings.Contains(strings.ToLower(target.Path), strings.ToLower(link)) {
					c.edges = append(c.edges, GraphEdge{
						Source: node.ID,
						Target: target.ID,
						Type:   "link",
						Weight: edgeWeights["link"],
					})
					break
				}
			}
		}

		// Build tag edges (nodes sharing tags)
		for _, otherNode := range c.nodes {
			if node.ID == otherNode.ID {
				continue
			}
			sharedTags := intersect(node.Tags, otherNode.Tags)
			if len(sharedTags) > 0 {
				c.edges = append(c.edges, GraphEdge{
					Source: node.ID,
					Target: otherNode.ID,
					Type:   "tag",
					Weight: edgeWeights["tag"] * float64(len(sharedTags)),
					Metadata: map[string]string{
						"shared_tags": strings.Join(sharedTags, ","),
					},
				})
			}
		}

		// Build temporal edges (same day)
		for _, otherNode := range c.nodes {
			if node.ID == otherNode.ID {
				continue
			}
			if sameDay(node.Modified, otherNode.Modified) {
				c.edges = append(c.edges, GraphEdge{
					Source: node.ID,
					Target: otherNode.ID,
					Type:   "temporal",
					Weight: edgeWeights["temporal"],
				})
			}
		}
	}
}

// intersect returns common elements between two slices
func intersect(a, b []string) []string {
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}
	var result []string
	for _, item := range b {
		if set[item] {
			result = append(result, item)
		}
	}
	return result
}

// sameDay checks if two times are on the same day
func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// Search performs graph-enhanced search
func (c *KnowledgeGraphClient) Search(ctx context.Context, query string, maxHops int) ([]GraphSearchResult, error) {
	if len(c.nodes) == 0 {
		c.RebuildGraph(ctx)
	}

	query = strings.ToLower(query)
	var results []GraphSearchResult

	// Direct matches (hop 0)
	for _, node := range c.nodes {
		score := c.calculateMatchScore(node, query)
		if score > 0 {
			results = append(results, GraphSearchResult{
				Node:      node,
				Score:     score,
				Hops:      0,
				Relevance: "direct match",
			})
		}
	}

	// Graph traversal for related nodes
	if maxHops > 0 {
		visited := make(map[string]bool)
		for _, result := range results {
			visited[result.Node.ID] = true
		}

		for hop := 1; hop <= maxHops; hop++ {
			for _, result := range results {
				if result.Hops != hop-1 {
					continue
				}
				// Find connected nodes
				for _, edge := range c.edges {
					var targetID string
					if edge.Source == result.Node.ID {
						targetID = edge.Target
					} else if edge.Target == result.Node.ID {
						targetID = edge.Source
					} else {
						continue
					}

					if visited[targetID] {
						continue
					}
					visited[targetID] = true

					if targetNode, ok := c.nodes[targetID]; ok {
						results = append(results, GraphSearchResult{
							Node:      targetNode,
							Score:     result.Score * edge.Weight * 0.5,
							Hops:      hop,
							Path:      append(result.Path, result.Node.ID),
							Relevance: fmt.Sprintf("connected via %s", edge.Type),
						})
					}
				}
			}
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// calculateMatchScore calculates how well a node matches a query
func (c *KnowledgeGraphClient) calculateMatchScore(node *GraphNode, query string) float64 {
	score := 0.0

	// Title match
	if strings.Contains(strings.ToLower(node.Title), query) {
		score += 1.0
	}

	// Tag match
	for _, tag := range node.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			score += 0.5
		}
	}

	// Path match
	if strings.Contains(strings.ToLower(node.Path), query) {
		score += 0.3
	}

	return score
}

// GetInsights returns analytics about the knowledge graph
func (c *KnowledgeGraphClient) GetInsights(ctx context.Context) (*GraphInsights, error) {
	if len(c.nodes) == 0 {
		c.RebuildGraph(ctx)
	}

	insights := &GraphInsights{
		TotalNodes:  len(c.nodes),
		TotalEdges:  len(c.edges),
		NodesByType: make(map[string]int),
		EdgesByType: make(map[string]int),
	}

	// Count nodes by type
	for _, node := range c.nodes {
		insights.NodesByType[node.Type]++
	}

	// Count edges by type
	for _, edge := range c.edges {
		insights.EdgesByType[edge.Type]++
	}

	// Find most connected nodes
	connectionCount := make(map[string]int)
	for _, edge := range c.edges {
		connectionCount[edge.Source]++
		connectionCount[edge.Target]++
	}

	type nodeConnections struct {
		node  *GraphNode
		count int
	}
	var connections []nodeConnections
	for id, count := range connectionCount {
		if node, ok := c.nodes[id]; ok {
			connections = append(connections, nodeConnections{node, count})
		}
	}
	sort.Slice(connections, func(i, j int) bool {
		return connections[i].count > connections[j].count
	})

	for i := 0; i < len(connections) && i < 5; i++ {
		insights.MostConnected = append(insights.MostConnected, *connections[i].node)
	}

	// Find orphan nodes (no connections)
	for id, node := range c.nodes {
		if connectionCount[id] == 0 {
			insights.Orphans = append(insights.Orphans, *node)
		}
	}

	// Recently updated
	var nodes []*GraphNode
	for _, node := range c.nodes {
		nodes = append(nodes, node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Modified.After(nodes[j].Modified)
	})
	for i := 0; i < len(nodes) && i < 5; i++ {
		insights.RecentlyUpdated = append(insights.RecentlyUpdated, *nodes[i])
	}

	return insights, nil
}

// GetContext returns related context for a document
func (c *KnowledgeGraphClient) GetContext(ctx context.Context, documentPath string, maxResults int) ([]GraphSearchResult, error) {
	if len(c.nodes) == 0 {
		c.RebuildGraph(ctx)
	}

	// Find the document node
	var sourceNode *GraphNode
	for _, node := range c.nodes {
		if node.Path == documentPath || strings.Contains(node.Path, documentPath) {
			sourceNode = node
			break
		}
	}

	if sourceNode == nil {
		return nil, fmt.Errorf("document not found: %s", documentPath)
	}

	// Find connected nodes
	var results []GraphSearchResult
	visited := map[string]bool{sourceNode.ID: true}

	for _, edge := range c.edges {
		var targetID string
		if edge.Source == sourceNode.ID {
			targetID = edge.Target
		} else if edge.Target == sourceNode.ID {
			targetID = edge.Source
		} else {
			continue
		}

		if visited[targetID] {
			continue
		}
		visited[targetID] = true

		if targetNode, ok := c.nodes[targetID]; ok {
			results = append(results, GraphSearchResult{
				Node:      targetNode,
				Score:     edge.Weight,
				Hops:      1,
				Relevance: edge.Type,
			})
		}
	}

	// Sort by score and limit
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

// FindSimilarShows finds shows similar to a given show or criteria
func (c *KnowledgeGraphClient) FindSimilarShows(ctx context.Context, criteria string) ([]SimilarShow, error) {
	if len(c.nodes) == 0 {
		c.RebuildGraph(ctx)
	}

	var shows []SimilarShow
	criteriaLower := strings.ToLower(criteria)

	for _, node := range c.nodes {
		if node.Type != "show" {
			continue
		}

		similarity := 0.0
		var sharedTags []string

		// Match by tags
		for _, tag := range node.Tags {
			if strings.Contains(criteriaLower, strings.ToLower(tag)) {
				similarity += 0.2
				sharedTags = append(sharedTags, tag)
			}
		}

		// Match by title
		if strings.Contains(strings.ToLower(node.Title), criteriaLower) {
			similarity += 0.5
		}

		if similarity > 0 {
			shows = append(shows, SimilarShow{
				Show:       node,
				Similarity: similarity,
				SharedTags: sharedTags,
			})
		}
	}

	// Sort by similarity
	sort.Slice(shows, func(i, j int) bool {
		return shows[i].Similarity > shows[j].Similarity
	})

	return shows, nil
}

// FindResolutionPath finds how similar issues were resolved before
func (c *KnowledgeGraphClient) FindResolutionPath(ctx context.Context, issue string) ([]ResolutionPath, error) {
	if len(c.nodes) == 0 {
		c.RebuildGraph(ctx)
	}

	var paths []ResolutionPath
	issueLower := strings.ToLower(issue)

	// Search for resolution patterns in learnings and runbooks
	for _, node := range c.nodes {
		if node.Type != "learning" && node.Type != "runbook" {
			continue
		}

		// Check if this document is relevant to the issue
		relevance := 0.0
		for _, tag := range node.Tags {
			if strings.Contains(issueLower, strings.ToLower(tag)) {
				relevance += 0.3
			}
		}
		if strings.Contains(strings.ToLower(node.Title), issueLower) {
			relevance += 0.5
		}

		if relevance > 0 {
			paths = append(paths, ResolutionPath{
				Problem:    issue,
				Resolution: fmt.Sprintf("See: %s", node.Title),
				Source:     node.Path,
				Confidence: relevance,
			})
		}
	}

	// Sort by confidence
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Confidence > paths[j].Confidence
	})

	return paths, nil
}

// VaultPath returns the configured vault path
func (c *KnowledgeGraphClient) VaultPath() string {
	return c.vaultPath
}

// NodeCount returns the number of nodes in the graph
func (c *KnowledgeGraphClient) NodeCount() int {
	return len(c.nodes)
}

// EdgeCount returns the number of edges in the graph
func (c *KnowledgeGraphClient) EdgeCount() int {
	return len(c.edges)
}
