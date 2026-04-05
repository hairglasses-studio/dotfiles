// Package notion provides MCP tools for Notion workspace integration.
package notion

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Notion tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "notion"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Notion workspace integration for pages, databases, and content management"
}

// Tools returns the Notion tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_notion_search",
				mcp.WithDescription("Search across all Notion pages and databases"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query text")),
				mcp.WithString("filter", mcp.Description("Filter results: 'page' or 'database' (default: all)"), mcp.Enum("page", "database")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 25, max: 100)")),
			),
			Handler:             handleNotionSearch,
			Category:            "notion",
			Subcategory:         "search",
			Tags:                []string{"notion", "search", "pages", "databases"},
			UseCases:            []string{"Find pages", "Search content", "Locate databases"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_page",
				mcp.WithDescription("Get a Notion page by ID with its properties"),
				mcp.WithString("page_id", mcp.Required(), mcp.Description("The Notion page ID (UUID format)")),
			),
			Handler:             handleNotionPage,
			Category:            "notion",
			Subcategory:         "pages",
			Tags:                []string{"notion", "page", "get", "properties"},
			UseCases:            []string{"Get page details", "View properties", "Check metadata"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_page_content",
				mcp.WithDescription("Get the content blocks of a Notion page"),
				mcp.WithString("page_id", mcp.Required(), mcp.Description("The Notion page ID")),
				mcp.WithNumber("limit", mcp.Description("Maximum blocks to return (default: 100)")),
			),
			Handler:             handleNotionPageContent,
			Category:            "notion",
			Subcategory:         "pages",
			Tags:                []string{"notion", "page", "content", "blocks"},
			UseCases:            []string{"Read page content", "Get blocks", "Extract text"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_database",
				mcp.WithDescription("Get a Notion database schema and metadata"),
				mcp.WithString("database_id", mcp.Required(), mcp.Description("The Notion database ID")),
			),
			Handler:             handleNotionDatabase,
			Category:            "notion",
			Subcategory:         "databases",
			Tags:                []string{"notion", "database", "schema", "properties"},
			UseCases:            []string{"Get database schema", "View properties", "Check structure"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_database_query",
				mcp.WithDescription("Query a Notion database with optional filters and sorting"),
				mcp.WithString("database_id", mcp.Required(), mcp.Description("The Notion database ID")),
				mcp.WithString("filter_property", mcp.Description("Property name to filter on")),
				mcp.WithString("filter_value", mcp.Description("Value to filter for. For checkbox use 'true'/'false'. For date use YYYY-MM-DD.")),
				mcp.WithString("filter_type", mcp.Description("Property type for filter: 'rich_text' (default), 'select', 'checkbox', 'date', 'number', 'multi_select'"), mcp.Enum("rich_text", "select", "checkbox", "date", "number", "multi_select")),
				mcp.WithString("filter_operator", mcp.Description("Filter operator. rich_text: 'contains'/'equals'. select: 'equals'/'is_empty'/'is_not_empty'. date: 'equals'/'before'/'after'/'on_or_before'/'on_or_after'/'past_week'/'next_week'. number: 'equals'/'greater_than'/'less_than'. Default varies by type.")),
				mcp.WithString("sort_property", mcp.Description("Property name to sort by")),
				mcp.WithString("sort_direction", mcp.Description("Sort direction: 'ascending' or 'descending'"), mcp.Enum("ascending", "descending")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 100)")),
			),
			Handler:             handleNotionDatabaseQuery,
			Category:            "notion",
			Subcategory:         "databases",
			Tags:                []string{"notion", "database", "query", "filter", "sort"},
			UseCases:            []string{"Query database", "Filter records", "List entries"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_create_page",
				mcp.WithDescription("Create a new Notion page in a parent page or database"),
				mcp.WithString("parent_id", mcp.Required(), mcp.Description("Parent page or database ID")),
				mcp.WithString("title", mcp.Required(), mcp.Description("Page title")),
				mcp.WithString("content", mcp.Description("Initial page content (plain text, will be added as paragraph)")),
				mcp.WithBoolean("is_database", mcp.Description("Whether parent is a database (default: false)")),
			),
			Handler:             handleNotionCreatePage,
			Category:            "notion",
			Subcategory:         "pages",
			Tags:                []string{"notion", "create", "page", "new"},
			UseCases:            []string{"Create page", "Add database entry", "New document"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_append",
				mcp.WithDescription("Append content to an existing Notion page"),
				mcp.WithString("page_id", mcp.Required(), mcp.Description("The Notion page ID to append to")),
				mcp.WithString("content", mcp.Required(), mcp.Description("Text content to append")),
				mcp.WithString("block_type", mcp.Description("Block type: 'paragraph', 'heading', 'bullet', 'todo' (default: paragraph)"), mcp.Enum("paragraph", "heading", "bullet", "todo")),
			),
			Handler:             handleNotionAppend,
			Category:            "notion",
			Subcategory:         "pages",
			Tags:                []string{"notion", "append", "content", "update"},
			UseCases:            []string{"Add content", "Update page", "Append text"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_create_db_entry",
				mcp.WithDescription("Create a new row in a Notion database with typed properties (select, date, checkbox, number, multi_select, url, rich_text)"),
				mcp.WithString("database_id", mcp.Required(), mcp.Description("The Notion database ID to add the entry to")),
				mcp.WithString("title", mcp.Required(), mcp.Description("Value for the title property (first column)")),
				mcp.WithString("properties", mcp.Description("JSON object of properties. Each key is a property name, value is {\"type\": \"<type>\", \"value\": <val>}. Types: 'rich_text', 'select', 'multi_select', 'date', 'checkbox', 'number', 'url'. Example: {\"Status\":{\"type\":\"select\",\"value\":\"Reading\"}, \"Rating\":{\"type\":\"number\",\"value\":4}}")),
			),
			Handler:             handleNotionCreateDBEntry,
			Category:            "notion",
			Subcategory:         "databases",
			Tags:                []string{"notion", "database", "create", "entry", "row", "properties"},
			UseCases:            []string{"Add habit entry", "Log contact", "Track job application", "Add reading list item"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_update_page",
				mcp.WithDescription("Update properties on an existing Notion page or database row"),
				mcp.WithString("page_id", mcp.Required(), mcp.Description("The Notion page ID to update")),
				mcp.WithString("properties", mcp.Required(), mcp.Description("JSON object of properties to update. Each key is a property name, value is {\"type\": \"<type>\", \"value\": <val>}. Types: 'title', 'rich_text', 'select', 'multi_select', 'date', 'checkbox', 'number', 'url'. Example: {\"Status\":{\"type\":\"select\",\"value\":\"Finished\"}, \"Rating\":{\"type\":\"number\",\"value\":5}}")),
			),
			Handler:             handleNotionUpdatePage,
			Category:            "notion",
			Subcategory:         "pages",
			Tags:                []string{"notion", "update", "page", "properties", "edit"},
			UseCases:            []string{"Update habit entry", "Change job status", "Mark book as finished", "Update contact info"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "notion",
		},
		{
			Tool: mcp.NewTool("aftrs_notion_status",
				mcp.WithDescription("Get Notion workspace connection status and current user info"),
			),
			Handler:             handleNotionStatus,
			Category:            "notion",
			Subcategory:         "status",
			Tags:                []string{"notion", "status", "user", "workspace"},
			UseCases:            []string{"Check connection", "View user info", "Verify access"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "notion",
		},
	}
}

var getNotionClient = tools.LazyClient(clients.GetNotionClient)

func handleNotionSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	filter := tools.GetStringParam(req, "filter")
	limit := tools.GetIntParam(req, "limit", 25)

	results, err := client.Search(ctx, query, filter, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to search: %w", err)), nil
	}

	// Format results
	formatted := map[string]interface{}{
		"query":       query,
		"count":       len(results.Results),
		"has_more":    results.HasMore,
		"next_cursor": results.NextCursor,
		"results":     results.Results,
	}

	return tools.JSONResult(formatted), nil
}

func handleNotionPage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}

	page, err := client.GetPage(ctx, pageID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get page: %w", err)), nil
	}

	return tools.JSONResult(page), nil
}

func handleNotionPageContent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 100)

	blocks, err := client.GetBlockChildren(ctx, pageID, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get blocks: %w", err)), nil
	}

	result := map[string]interface{}{
		"page_id": pageID,
		"count":   len(blocks),
		"blocks":  blocks,
	}

	return tools.JSONResult(result), nil
}

func handleNotionDatabase(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	databaseID, errResult := tools.RequireStringParam(req, "database_id")
	if errResult != nil {
		return errResult, nil
	}

	db, err := client.GetDatabase(ctx, databaseID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get database: %w", err)), nil
	}

	return tools.JSONResult(db), nil
}

func handleNotionDatabaseQuery(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	databaseID, errResult := tools.RequireStringParam(req, "database_id")
	if errResult != nil {
		return errResult, nil
	}

	query := &clients.NotionDatabaseQuery{
		PageSize: tools.GetIntParam(req, "limit", 100),
	}

	// Add filter if specified
	filterProp := tools.GetStringParam(req, "filter_property")
	filterVal := tools.GetStringParam(req, "filter_value")
	filterType := tools.GetStringParam(req, "filter_type")
	filterOp := tools.GetStringParam(req, "filter_operator")
	if filterProp != "" && filterVal != "" {
		query.Filter = buildDatabaseFilter(filterProp, filterVal, filterType, filterOp)
	}

	// Add sort if specified
	sortProp := tools.GetStringParam(req, "sort_property")
	sortDir := tools.GetStringParam(req, "sort_direction")
	if sortProp != "" {
		if sortDir == "" {
			sortDir = "descending"
		}
		query.Sorts = []clients.NotionSort{
			{Property: sortProp, Direction: sortDir},
		}
	}

	pages, hasMore, nextCursor, err := client.QueryDatabase(ctx, databaseID, query)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to query database: %w", err)), nil
	}

	result := map[string]interface{}{
		"database_id": databaseID,
		"count":       len(pages),
		"has_more":    hasMore,
		"next_cursor": nextCursor,
		"pages":       pages,
	}

	return tools.JSONResult(result), nil
}

func handleNotionCreatePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	parentID, errResult := tools.RequireStringParam(req, "parent_id")
	if errResult != nil {
		return errResult, nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	content := tools.GetStringParam(req, "content")
	isDatabase := tools.GetBoolParam(req, "is_database", false)

	// Build content blocks
	var blocks []map[string]interface{}
	if content != "" {
		blocks = append(blocks, clients.CreateTextBlock(content))
	}

	page, err := client.CreatePage(ctx, parentID, isDatabase, title, blocks)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create page: %w", err)), nil
	}

	result := map[string]interface{}{
		"page":    page,
		"message": fmt.Sprintf("Successfully created page: %s", page.Title),
	}

	return tools.JSONResult(result), nil
}

func handleNotionAppend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
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

	var block map[string]interface{}
	switch blockType {
	case "heading":
		block = clients.CreateHeadingBlock(content, 2)
	case "bullet":
		block = clients.CreateBulletBlock(content)
	case "todo":
		block = clients.CreateTodoBlock(content, false)
	default:
		block = clients.CreateTextBlock(content)
	}

	err = client.AppendBlocks(ctx, pageID, []map[string]interface{}{block})
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to append content: %w", err)), nil
	}

	result := map[string]interface{}{
		"page_id":    pageID,
		"block_type": blockType,
		"message":    "Successfully appended content to page",
	}

	return tools.JSONResult(result), nil
}

func handleNotionCreateDBEntry(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	databaseID, errResult := tools.RequireStringParam(req, "database_id")
	if errResult != nil {
		return errResult, nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	// Build properties starting with title
	notionProps := map[string]interface{}{
		"title": map[string]interface{}{
			"title": []map[string]interface{}{
				{"text": map[string]string{"content": title}},
			},
		},
	}

	// Parse additional typed properties if provided
	propsJSON := tools.GetStringParam(req, "properties")
	if propsJSON != "" {
		var typedProps map[string]map[string]interface{}
		if err := json.Unmarshal([]byte(propsJSON), &typedProps); err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid properties JSON: %w", err)), nil
		}
		for name, prop := range typedProps {
			notionProp, err := buildNotionProperty(prop)
			if err != nil {
				return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("property %q: %w", name, err)), nil
			}
			notionProps[name] = notionProp
		}
	}

	page, err := client.CreatePageWithProperties(ctx, databaseID, notionProps, nil)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create database entry: %w", err)), nil
	}

	result := map[string]interface{}{
		"page_id": page.ID,
		"url":     page.URL,
		"title":   title,
		"message": fmt.Sprintf("Successfully created entry: %s", title),
	}

	return tools.JSONResult(result), nil
}

func handleNotionUpdatePage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}

	propsJSON, errResult := tools.RequireStringParam(req, "properties")
	if errResult != nil {
		return errResult, nil
	}

	var typedProps map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(propsJSON), &typedProps); err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid properties JSON: %w", err)), nil
	}

	notionProps := make(map[string]interface{}, len(typedProps))
	for name, prop := range typedProps {
		notionProp, err := buildNotionProperty(prop)
		if err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("property %q: %w", name, err)), nil
		}
		notionProps[name] = notionProp
	}

	page, err := client.UpdatePageProperties(ctx, pageID, notionProps)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to update page: %w", err)), nil
	}

	result := map[string]interface{}{
		"page_id": page.ID,
		"url":     page.URL,
		"title":   page.Title,
		"message": "Successfully updated page properties",
	}

	return tools.JSONResult(result), nil
}

// buildNotionProperty converts a typed property definition to Notion API format.
// Input: {"type": "select", "value": "Good"}
// Output: {"select": {"name": "Good"}}
func buildNotionProperty(prop map[string]interface{}) (map[string]interface{}, error) {
	propType, ok := prop["type"].(string)
	if !ok || propType == "" {
		return nil, fmt.Errorf("missing or invalid 'type' field")
	}

	value := prop["value"]

	switch propType {
	case "title":
		strVal := fmt.Sprintf("%v", value)
		return map[string]interface{}{
			"title": []map[string]interface{}{
				{"text": map[string]string{"content": strVal}},
			},
		}, nil

	case "rich_text":
		strVal := fmt.Sprintf("%v", value)
		return map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"text": map[string]string{"content": strVal}},
			},
		}, nil

	case "select":
		strVal := fmt.Sprintf("%v", value)
		return map[string]interface{}{
			"select": map[string]string{"name": strVal},
		}, nil

	case "multi_select":
		var names []map[string]string
		switch v := value.(type) {
		case []interface{}:
			for _, item := range v {
				names = append(names, map[string]string{"name": fmt.Sprintf("%v", item)})
			}
		case string:
			names = append(names, map[string]string{"name": v})
		default:
			return nil, fmt.Errorf("multi_select value must be an array of strings")
		}
		return map[string]interface{}{
			"multi_select": names,
		}, nil

	case "date":
		strVal := fmt.Sprintf("%v", value)
		return map[string]interface{}{
			"date": map[string]string{"start": strVal},
		}, nil

	case "checkbox":
		boolVal := false
		switch v := value.(type) {
		case bool:
			boolVal = v
		case string:
			boolVal = v == "true"
		case float64:
			boolVal = v != 0
		}
		return map[string]interface{}{
			"checkbox": boolVal,
		}, nil

	case "number":
		var numVal float64
		switch v := value.(type) {
		case float64:
			numVal = v
		case int:
			numVal = float64(v)
		case string:
			if _, err := fmt.Sscanf(v, "%f", &numVal); err != nil {
				return nil, fmt.Errorf("invalid number value: %s", v)
			}
		default:
			return nil, fmt.Errorf("number value must be numeric")
		}
		return map[string]interface{}{
			"number": numVal,
		}, nil

	case "url":
		strVal := fmt.Sprintf("%v", value)
		return map[string]interface{}{
			"url": strVal,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported property type: %s", propType)
	}
}

// buildDatabaseFilter constructs a Notion API filter based on property type and operator.
func buildDatabaseFilter(property, value, filterType, operator string) map[string]interface{} {
	if filterType == "" {
		filterType = "rich_text"
	}

	filter := map[string]interface{}{
		"property": property,
	}

	switch filterType {
	case "rich_text":
		if operator == "" {
			operator = "contains"
		}
		filter["rich_text"] = map[string]interface{}{
			operator: value,
		}

	case "select":
		if operator == "" {
			operator = "equals"
		}
		if operator == "is_empty" || operator == "is_not_empty" {
			filter["select"] = map[string]interface{}{
				operator: true,
			}
		} else {
			filter["select"] = map[string]interface{}{
				operator: value,
			}
		}

	case "multi_select":
		if operator == "" {
			operator = "contains"
		}
		if operator == "is_empty" || operator == "is_not_empty" {
			filter["multi_select"] = map[string]interface{}{
				operator: true,
			}
		} else {
			filter["multi_select"] = map[string]interface{}{
				operator: value,
			}
		}

	case "checkbox":
		boolVal := value == "true"
		filter["checkbox"] = map[string]interface{}{
			"equals": boolVal,
		}

	case "date":
		if operator == "" {
			operator = "equals"
		}
		if operator == "past_week" || operator == "next_week" {
			filter["date"] = map[string]interface{}{
				operator: map[string]interface{}{},
			}
		} else {
			filter["date"] = map[string]interface{}{
				operator: value,
			}
		}

	case "number":
		if operator == "" {
			operator = "equals"
		}
		var numVal float64
		fmt.Sscanf(value, "%f", &numVal)
		filter["number"] = map[string]interface{}{
			operator: numVal,
		}
	}

	return filter
}

func handleNotionStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getNotionClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Notion client: %w", err)), nil
	}

	user, err := client.GetUser(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get user: %w", err)), nil
	}

	result := map[string]interface{}{
		"status":      "connected",
		"user":        user,
		"api_version": "2022-06-28",
	}

	return tools.JSONResult(result), nil
}
