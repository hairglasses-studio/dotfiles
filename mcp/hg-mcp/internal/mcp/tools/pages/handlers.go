package pages

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleList queries the pages database with optional filters
func handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	dbID, err := getDatabaseID()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	limit := tools.GetIntParam(req, "limit", 25)
	if limit > 100 {
		limit = 100
	}
	cursor := tools.GetStringParam(req, "cursor")
	category := tools.GetStringParam(req, "category")
	tag := tools.GetStringParam(req, "tag")
	status := tools.GetStringParam(req, "status")

	// Build filters: always exclude templates
	var filters []map[string]interface{}
	filters = append(filters, map[string]interface{}{
		"property": "Template",
		"checkbox": map[string]interface{}{"equals": false},
	})

	if category != "" {
		filters = append(filters, map[string]interface{}{
			"property": "Category",
			"select":   map[string]interface{}{"equals": category},
		})
	}
	if tag != "" {
		filters = append(filters, map[string]interface{}{
			"property":     "Tags",
			"multi_select": map[string]interface{}{"contains": tag},
		})
	}
	if status != "" {
		filters = append(filters, map[string]interface{}{
			"property": "Status",
			"select":   map[string]interface{}{"equals": status},
		})
	}

	query := &clients.NotionDatabaseQuery{
		Filter: buildCompoundFilter(filters),
		Sorts: []clients.NotionSort{
			{Timestamp: "last_edited_time", Direction: "descending"},
		},
		PageSize: limit,
	}
	if cursor != "" {
		query.StartCursor = cursor
	}

	pages, hasMore, nextCursor, err := client.QueryDatabase(ctx, dbID, query)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to query pages: %w", err)), nil
	}

	summaries := make([]pageSummary, len(pages))
	for i, p := range pages {
		summaries[i] = formatPageSummary(p)
	}

	result := map[string]interface{}{
		"pages":    summaries,
		"count":    len(summaries),
		"has_more": hasMore,
	}
	if hasMore && nextCursor != "" {
		result["next_cursor"] = nextCursor
	}

	return tools.JSONResult(result), nil
}

// handleRead gets a page's properties and optionally its content
func handleRead(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}
	includeContent := tools.GetBoolParam(req, "content", false)
	format := tools.OptionalStringParam(req, "format", "json")

	page, err := client.GetPage(ctx, pageID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get page: %w", err)), nil
	}

	result := map[string]interface{}{
		"page": formatPageSummary(*page),
	}

	if includeContent {
		blocks, err := client.GetBlockChildren(ctx, pageID, 100)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to get page content: %w", err)), nil
		}

		if format == "markdown" {
			md := blocksToMarkdown(blocks)
			result["content"] = md
		} else {
			blockData := make([]map[string]interface{}, len(blocks))
			for i, b := range blocks {
				blockData[i] = map[string]interface{}{
					"id":           b.ID,
					"type":         b.Type,
					"has_children": b.HasChildren,
					"content":      b.Content,
				}
			}
			result["content"] = blockData
		}
	}

	return tools.JSONResult(result), nil
}

// handleWrite creates a new page in the database
func handleWrite(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	dbID, err := getDatabaseID()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}
	content := tools.GetStringParam(req, "content")
	tagsStr := tools.GetStringParam(req, "tags")
	category := tools.GetStringParam(req, "category")
	status := tools.OptionalStringParam(req, "status", "Draft")

	// Build properties
	properties := map[string]interface{}{
		"Title":    buildTitleProperty(title),
		"Status":   buildSelectProperty(status),
		"Template": buildCheckboxProperty(false),
	}
	if tagsStr != "" {
		tags := splitTags(tagsStr)
		if len(tags) > 0 {
			properties["Tags"] = buildMultiSelectProperty(tags)
		}
	}
	if category != "" {
		properties["Category"] = buildSelectProperty(category)
	}

	// Build content blocks
	var children []map[string]interface{}
	if content != "" {
		children = append(children, clients.CreateTextBlock(content))
	}

	page, err := client.CreatePageWithProperties(ctx, dbID, properties, children)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create page: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message": "Page created successfully",
		"page":    formatPageSummary(*page),
	}), nil
}

// handleUpdate appends blocks to an existing page
func handleUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}
	content, errResult := tools.RequireStringParam(req, "content")
	if errResult != nil {
		return errResult, nil
	}
	blockType := tools.OptionalStringParam(req, "block_type", "paragraph")

	block := buildBlock(blockType, content)
	if err := client.AppendBlocks(ctx, pageID, []map[string]interface{}{block}); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to append content: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("Appended %s block to page %s", blockType, pageID)), nil
}

// handleSearch searches pages by title within the notes database.
// Uses a database query with rich_text.contains filter for reliable, scoped results.
func handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	dbID, err := getDatabaseID()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}
	limit := tools.GetIntParam(req, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	// Query the database directly with a title contains filter, excluding templates
	dbQuery := &clients.NotionDatabaseQuery{
		Filter: buildCompoundFilter([]map[string]interface{}{
			{
				"property":  "Title",
				"rich_text": map[string]interface{}{"contains": query},
			},
			{
				"property": "Template",
				"checkbox": map[string]interface{}{"equals": false},
			},
		}),
		Sorts: []clients.NotionSort{
			{Timestamp: "last_edited_time", Direction: "descending"},
		},
		PageSize: limit,
	}

	pages, _, _, err := client.QueryDatabase(ctx, dbID, dbQuery)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	matches := make([]pageSummary, len(pages))
	for i, p := range pages {
		matches[i] = formatPageSummary(p)
	}

	return tools.JSONResult(map[string]interface{}{
		"results": matches,
		"count":   len(matches),
		"query":   query,
	}), nil
}

// handleDelete archives (soft-deletes) a page
func handleDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.ArchivePage(ctx, pageID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to archive page: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("Page %s archived successfully", pageID)), nil
}

// --- helpers ---

func buildBlock(blockType, content string) map[string]interface{} {
	switch blockType {
	case "heading_1":
		return clients.CreateHeadingBlock(content, 1)
	case "heading_2":
		return clients.CreateHeadingBlock(content, 2)
	case "heading_3":
		return clients.CreateHeadingBlock(content, 3)
	case "bulleted_list_item":
		return clients.CreateBulletBlock(content)
	case "to_do":
		return clients.CreateTodoBlock(content, false)
	case "numbered_list_item":
		return map[string]interface{}{
			"object": "block",
			"type":   "numbered_list_item",
			"numbered_list_item": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]string{"content": content}},
				},
			},
		}
	case "quote":
		return map[string]interface{}{
			"object": "block",
			"type":   "quote",
			"quote": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]string{"content": content}},
				},
			},
		}
	case "code":
		return map[string]interface{}{
			"object": "block",
			"type":   "code",
			"code": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"type": "text", "text": map[string]string{"content": content}},
				},
				"language": "plain text",
			},
		}
	default:
		return clients.CreateTextBlock(content)
	}
}

func splitTags(s string) []string {
	parts := strings.Split(s, ",")
	var tags []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func parsePageFromMap(m map[string]interface{}) clients.NotionPage {
	page := clients.NotionPage{}
	if id, ok := m["id"].(string); ok {
		page.ID = id
	}
	if url, ok := m["url"].(string); ok {
		page.URL = url
	}
	if archived, ok := m["archived"].(bool); ok {
		page.Archived = archived
	}
	if props, ok := m["properties"].(map[string]interface{}); ok {
		page.Properties = props
		// Extract title
		for _, key := range []string{"title", "Title", "Name", "name"} {
			if prop, ok := props[key]; ok {
				if propMap, ok := prop.(map[string]interface{}); ok {
					if titleArr, ok := propMap["title"].([]interface{}); ok && len(titleArr) > 0 {
						if first, ok := titleArr[0].(map[string]interface{}); ok {
							if plainText, ok := first["plain_text"].(string); ok {
								page.Title = plainText
								break
							}
						}
					}
				}
			}
		}
	}
	if let, ok := m["last_edited_time"].(string); ok {
		if t, err := time.Parse(time.RFC3339, let); err == nil {
			page.LastEditedTime = t
		}
	}
	return page
}
