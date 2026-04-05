package inventory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var validReturnStatuses = map[string]bool{
	"none":      true,
	"requested": true,
	"approved":  true,
	"completed": true,
	"denied":    true,
}

// handleReturnStart initiates a return/dispute for an inventory item.
func handleReturnStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	reason, errResult := tools.RequireStringParam(req, "reason")
	if errResult != nil {
		return errResult, nil
	}

	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	item, err := client.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("item not found: %w", err)), nil
	}

	// Build dispute notes
	note := fmt.Sprintf("[%s] Return requested: %s", time.Now().Format("2006-01-02"), reason)
	disputeNotes := note
	if item.DisputeNotes != "" {
		disputeNotes = item.DisputeNotes + " | " + note
	}

	updates := map[string]interface{}{
		"return_status": "requested",
		"dispute_notes": disputeNotes,
	}

	updated, err := client.UpdateItem(ctx, sku, updates)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to update item: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"sku":           sku,
		"name":          updated.Name,
		"return_status": "requested",
		"dispute_notes": disputeNotes,
		"message":       fmt.Sprintf("Return initiated for %s", updated.Name),
	}), nil
}

// handleReturnResolve resolves a return/dispute (approve, complete, or deny).
func handleReturnResolve(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	resolutionRaw, errResult := tools.RequireStringParam(req, "resolution")
	if errResult != nil {
		return errResult, nil
	}
	resolution := strings.ToLower(resolutionRaw)
	if !validReturnStatuses[resolution] || resolution == "none" || resolution == "requested" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid resolution: %s (use approved, completed, or denied)", resolution)), nil
	}

	notes := tools.GetStringParam(req, "notes")

	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	item, err := client.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("item not found: %w", err)), nil
	}

	note := fmt.Sprintf("[%s] Return %s", time.Now().Format("2006-01-02"), resolution)
	if notes != "" {
		note += ": " + notes
	}
	disputeNotes := note
	if item.DisputeNotes != "" {
		disputeNotes = item.DisputeNotes + " | " + note
	}

	updates := map[string]interface{}{
		"return_status": resolution,
		"dispute_notes": disputeNotes,
	}

	updated, err := client.UpdateItem(ctx, sku, updates)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to update item: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"sku":           sku,
		"name":          updated.Name,
		"return_status": resolution,
		"dispute_notes": disputeNotes,
		"message":       fmt.Sprintf("Return %s for %s", resolution, updated.Name),
	}), nil
}
