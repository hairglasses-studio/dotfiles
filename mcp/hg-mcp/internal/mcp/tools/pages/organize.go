package pages

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleTag adds, removes, or replaces tags on a page
func handleTag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}

	addStr := tools.GetStringParam(req, "add")
	removeStr := tools.GetStringParam(req, "remove")
	setStr := tools.GetStringParam(req, "set")

	if addStr == "" && removeStr == "" && setStr == "" {
		return tools.ErrorResult(fmt.Errorf("at least one of 'add', 'remove', or 'set' is required")), nil
	}

	var finalTags []string

	if setStr != "" {
		// Replace all tags
		finalTags = splitTags(setStr)
	} else {
		// Get current tags
		page, err := client.GetPage(ctx, pageID)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to get page: %w", err)), nil
		}

		existing := extractTags(page.Properties)
		tagSet := make(map[string]bool)
		for _, t := range existing {
			tagSet[t] = true
		}

		// Add tags
		if addStr != "" {
			for _, t := range splitTags(addStr) {
				tagSet[t] = true
			}
		}

		// Remove tags
		if removeStr != "" {
			for _, t := range splitTags(removeStr) {
				delete(tagSet, t)
			}
		}

		for t := range tagSet {
			finalTags = append(finalTags, t)
		}
	}

	properties := map[string]interface{}{
		"Tags": buildMultiSelectProperty(finalTags),
	}

	page, err := client.UpdatePageProperties(ctx, pageID, properties)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to update tags: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message": "Tags updated successfully",
		"tags":    extractTags(page.Properties),
		"page_id": pageID,
	}), nil
}

// handleMove changes the category of a page
func handleMove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}
	category, errResult := tools.RequireStringParam(req, "category")
	if errResult != nil {
		return errResult, nil
	}

	properties := map[string]interface{}{
		"Category": buildSelectProperty(category),
	}

	_, err = client.UpdatePageProperties(ctx, pageID, properties)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to update category: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("Page %s moved to category '%s'", pageID, category)), nil
}

// handleSetStatus changes the status of a page (Draft, Active, Archived)
func handleSetStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}
	status, errResult := tools.RequireStringParam(req, "status")
	if errResult != nil {
		return errResult, nil
	}

	// Validate status value
	switch status {
	case "Draft", "Active", "Archived":
		// valid
	default:
		return tools.ErrorResult(fmt.Errorf("invalid status '%s': must be Draft, Active, or Archived", status)), nil
	}

	properties := map[string]interface{}{
		"Status": buildSelectProperty(status),
	}

	_, err = client.UpdatePageProperties(ctx, pageID, properties)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to update status: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("Page %s status changed to '%s'", pageID, status)), nil
}

// handleRename changes the title of a page
func handleRename(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pageID, errResult := tools.RequireStringParam(req, "page_id")
	if errResult != nil {
		return errResult, nil
	}
	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	properties := map[string]interface{}{
		"Title": buildTitleProperty(title),
	}

	_, err = client.UpdatePageProperties(ctx, pageID, properties)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to rename page: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("Page %s renamed to '%s'", pageID, title)), nil
}
