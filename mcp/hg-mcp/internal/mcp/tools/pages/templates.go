package pages

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleTemplates lists pages marked as templates
func handleTemplates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	dbID, err := getDatabaseID()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	category := tools.GetStringParam(req, "category")

	var filters []map[string]interface{}
	filters = append(filters, map[string]interface{}{
		"property": "Template",
		"checkbox": map[string]interface{}{"equals": true},
	})
	if category != "" {
		filters = append(filters, map[string]interface{}{
			"property": "Category",
			"select":   map[string]interface{}{"equals": category},
		})
	}

	query := &clients.NotionDatabaseQuery{
		Filter: buildCompoundFilter(filters),
		Sorts: []clients.NotionSort{
			{Property: "Title", Direction: "ascending"},
		},
		PageSize: 100,
	}

	pages, _, _, err := client.QueryDatabase(ctx, dbID, query)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to query templates: %w", err)), nil
	}

	summaries := make([]pageSummary, len(pages))
	for i, p := range pages {
		summaries[i] = formatPageSummary(p)
	}

	return tools.JSONResult(map[string]interface{}{
		"templates": summaries,
		"count":     len(summaries),
	}), nil
}

// handleFromTemplate creates a new page by copying a template's content and properties
func handleFromTemplate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	dbID, err := getDatabaseID()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	templateID, errResult := tools.RequireStringParam(req, "template_id")
	if errResult != nil {
		return errResult, nil
	}
	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}
	tagsStr := tools.GetStringParam(req, "tags")
	category := tools.GetStringParam(req, "category")

	// Read template page properties
	templatePage, err := client.GetPage(ctx, templateID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get template: %w", err)), nil
	}

	// Read template blocks
	blocks, err := client.GetBlockChildren(ctx, templateID, 100)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get template content: %w", err)), nil
	}

	// Build new page properties from template
	properties := map[string]interface{}{
		"Title":    buildTitleProperty(title),
		"Template": buildCheckboxProperty(false),
		"Status":   buildSelectProperty("Draft"),
	}

	// Use template's category unless overridden
	if category != "" {
		properties["Category"] = buildSelectProperty(category)
	} else {
		templateCategory := extractSelect(templatePage.Properties, "Category")
		if templateCategory != "" {
			properties["Category"] = buildSelectProperty(templateCategory)
		}
	}

	// Use provided tags or copy from template
	if tagsStr != "" {
		tags := splitTags(tagsStr)
		if len(tags) > 0 {
			properties["Tags"] = buildMultiSelectProperty(tags)
		}
	} else {
		templateTags := extractTags(templatePage.Properties)
		if len(templateTags) > 0 {
			properties["Tags"] = buildMultiSelectProperty(templateTags)
		}
	}

	// Convert template blocks for creation (strip API metadata)
	children := convertBlocksForCreate(blocks)

	page, err := client.CreatePageWithProperties(ctx, dbID, properties, children)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create page from template: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message":     "Page created from template",
		"page":        formatPageSummary(*page),
		"template_id": templateID,
	}), nil
}

// convertBlocksForCreate strips API metadata from blocks so they can be used in page creation.
// Only handles top-level blocks (nested/children blocks are not recursively copied in v1).
func convertBlocksForCreate(blocks []clients.NotionBlock) []map[string]interface{} {
	var result []map[string]interface{}
	for _, block := range blocks {
		converted := convertBlock(block)
		if converted != nil {
			result = append(result, converted)
		}
	}
	return result
}

func convertBlock(block clients.NotionBlock) map[string]interface{} {
	if block.Content == nil {
		return nil
	}

	switch block.Type {
	case "paragraph", "heading_1", "heading_2", "heading_3",
		"bulleted_list_item", "numbered_list_item", "to_do",
		"quote", "callout", "toggle":
		return map[string]interface{}{
			"object":   "block",
			"type":     block.Type,
			block.Type: block.Content,
		}
	case "code":
		return map[string]interface{}{
			"object": "block",
			"type":   "code",
			"code":   block.Content,
		}
	case "divider":
		return map[string]interface{}{
			"object":  "block",
			"type":    "divider",
			"divider": map[string]interface{}{},
		}
	default:
		// For unsupported block types, try to pass through content as-is
		if block.Content != nil {
			return map[string]interface{}{
				"object":   "block",
				"type":     block.Type,
				block.Type: block.Content,
			}
		}
		return nil
	}
}
