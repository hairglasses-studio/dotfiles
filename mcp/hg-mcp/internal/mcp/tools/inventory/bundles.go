package inventory

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleBundleCreate assigns a bundle ID to a set of items, grouping them as a lot.
func handleBundleCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	bundleID, errResult := tools.RequireStringParam(req, "bundle_id")
	if errResult != nil {
		return errResult, nil
	}

	skusStr, errResult := tools.RequireStringParam(req, "skus")
	if errResult != nil {
		return errResult, nil
	}

	skus := splitTrim(skusStr)
	if len(skus) < 2 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("at least 2 SKUs required for a bundle")), nil
	}

	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var updated []string
	var errors []string

	for _, sku := range skus {
		updates := map[string]interface{}{
			"bundle_id": bundleID,
		}
		_, err := client.UpdateItem(ctx, sku, updates)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", sku, err))
		} else {
			updated = append(updated, sku)
		}
	}

	result := map[string]interface{}{
		"bundle_id":    bundleID,
		"items_tagged": len(updated),
		"skus":         updated,
	}
	if len(errors) > 0 {
		result["errors"] = errors
	}

	return tools.JSONResult(result), nil
}

// handleBundleList lists all bundles or items in a specific bundle.
func handleBundleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	bundleID := tools.GetStringParam(req, "bundle_id")

	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	items, err := client.ListItems(ctx, nil)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list items: %w", err)), nil
	}

	if bundleID != "" {
		// List items in a specific bundle
		var bundleItems []map[string]interface{}
		for _, item := range items {
			if item.BundleID == bundleID {
				bundleItems = append(bundleItems, map[string]interface{}{
					"sku":      item.SKU,
					"name":     item.Name,
					"category": item.Category,
					"price":    item.AskingPrice,
				})
			}
		}
		return tools.JSONResult(map[string]interface{}{
			"bundle_id": bundleID,
			"count":     len(bundleItems),
			"items":     bundleItems,
		}), nil
	}

	// List all bundles with item counts
	bundleCounts := map[string]int{}
	for _, item := range items {
		if item.BundleID != "" {
			bundleCounts[item.BundleID]++
		}
	}

	var bundles []map[string]interface{}
	for id, count := range bundleCounts {
		bundles = append(bundles, map[string]interface{}{
			"bundle_id":  id,
			"item_count": count,
		})
	}

	return tools.JSONResult(map[string]interface{}{
		"total_bundles": len(bundles),
		"bundles":       bundles,
	}), nil
}

func splitTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
