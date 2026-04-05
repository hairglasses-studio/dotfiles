package inventory

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

// Listing templates embedded as Go constants.
// Ported from tools/hw-resale/templates/

const fbMarketplaceTemplate = `{name} — ${asking_price}

Condition: {condition_display}
{specs_block}

{notes}

Price is firm. Local pickup preferred (can meet in public spot).
Cash or Zelle. No trades.

Also selling other hardware — ask about bundle deals!`

const ebayTemplate = `{name}

CONDITION: {condition_display}

SPECIFICATIONS:
{specs_block}

DESCRIPTION:
{notes}

WHAT'S INCLUDED:
- {name}
- Original packaging (if noted)

SHIPPING:
- Ships within 2 business days via USPS Priority / UPS Ground
- Insurance included
- Continental US only

RETURNS:
- 14-day return policy for items not as described
- Buyer pays return shipping

PAYMENT:
- PayPal / eBay Managed Payments

Check our other listings for bundle deals on GPUs, SSDs, and networking gear!`

const hardwareswapTemplate = `[{category_tag}] [H] {name} [W] PayPal, Local Cash

**Price:** ${asking_price} shipped / ${local_price} local

**Condition:** {condition_display}

**Specs:**
{specs_block}

**Notes:** {notes}

**Timestamps:** [link to timestamps]

{smart_data_block}

Comment before PM. No chat please.`

// validPlatforms lists supported listing platforms.
var validPlatforms = map[string]string{
	"fb_marketplace": fbMarketplaceTemplate,
	"ebay":           ebayTemplate,
	"hardwareswap":   hardwareswapTemplate,
}

// RenderListing generates a marketplace listing from an inventory item.
func RenderListing(item *clients.InventoryItem, platform string) (string, error) {
	tmpl, ok := validPlatforms[platform]
	if !ok {
		return "", fmt.Errorf("invalid platform %q (valid: fb_marketplace, ebay, hardwareswap)", platform)
	}

	askingPrice := item.AskingPrice
	if askingPrice == 0 {
		askingPrice = item.CurrentRetail * 0.8
	}

	conditionDisplay := formatConditionDisplay(item.Condition)
	specsBlock := buildSpecsBlock(item)
	localPrice := askingPrice * 0.9
	smartBlock := ""
	if item.SmartData != "" {
		smartBlock = "**SMART Data:**\n" + item.SmartData
	}

	notes := item.Notes
	if notes == "" {
		notes = "No additional notes."
	}

	r := strings.NewReplacer(
		"{name}", item.Name,
		"{asking_price}", fmt.Sprintf("%.0f", askingPrice),
		"{condition_display}", conditionDisplay,
		"{specs_block}", specsBlock,
		"{notes}", notes,
		"{category_tag}", "USA-XX",
		"{local_price}", fmt.Sprintf("%.0f", localPrice),
		"{smart_data_block}", smartBlock,
	)

	return strings.TrimSpace(r.Replace(tmpl)), nil
}

// formatConditionDisplay normalizes internal condition strings to human-readable display text.
func formatConditionDisplay(condition string) string {
	switch strings.ToLower(condition) {
	case "new", "new / sealed":
		return "New / Sealed"
	case "new_open_box", "new (open box)":
		return "New (Open Box)"
	case "like_new":
		return "Like New"
	case "used", "used_excellent", "used — excellent":
		return "Used — Excellent"
	case "used_good", "used — good", "good":
		return "Used — Good"
	case "used_fair", "used — fair", "fair":
		return "Used — Fair"
	case "renewed":
		return "Renewed / Refurbished"
	case "for_parts":
		return "For Parts / Not Working"
	default:
		return condition
	}
}

// buildSpecsBlock generates the specs section for marketplace listings.
// Prefers the Specs map (col V) when populated; falls back to ASIN/MSRP metadata.
func buildSpecsBlock(item *clients.InventoryItem) string {
	var lines []string

	// Prefer Specs map when populated (actual hardware specs)
	if len(item.Specs) > 0 {
		if item.Model != "" {
			lines = append(lines, fmt.Sprintf("  - Model: %s", item.Model))
		}
		// Render specs in a stable order: known keys first, then alphabetical
		knownKeys := []string{"vram", "memory", "capacity", "interface", "speed", "tdp", "form_factor", "chipset"}
		rendered := make(map[string]bool)
		for _, k := range knownKeys {
			if v, ok := item.Specs[k]; ok {
				lines = append(lines, fmt.Sprintf("  - %s: %s", formatSpecKey(k), v))
				rendered[k] = true
			}
		}
		// Remaining keys alphabetically
		var remaining []string
		for k := range item.Specs {
			if !rendered[k] {
				remaining = append(remaining, k)
			}
		}
		sort.Strings(remaining)
		for _, k := range remaining {
			lines = append(lines, fmt.Sprintf("  - %s: %s", formatSpecKey(k), item.Specs[k]))
		}
		if item.Quantity > 1 {
			lines = append(lines, fmt.Sprintf("  - Quantity Available: %d", item.Quantity))
		}
		return strings.Join(lines, "\n")
	}

	// Fallback: generic specs from ASIN/MSRP metadata
	if item.Model != "" {
		lines = append(lines, fmt.Sprintf("  - Model: %s", item.Model))
	}
	if item.Category != "" {
		lines = append(lines, fmt.Sprintf("  - Category: %s", item.Category))
	}
	if item.ASIN != "" {
		lines = append(lines, fmt.Sprintf("  - Amazon ASIN: %s", item.ASIN))
	}
	if item.MSRP > 0 {
		lines = append(lines, fmt.Sprintf("  - MSRP: $%.0f", item.MSRP))
	}
	if item.CurrentRetail > 0 {
		lines = append(lines, fmt.Sprintf("  - Current Retail: $%.0f", item.CurrentRetail))
	}
	if item.Quantity > 1 {
		lines = append(lines, fmt.Sprintf("  - Quantity Available: %d", item.Quantity))
	}
	if len(lines) == 0 {
		return "  N/A"
	}
	return strings.Join(lines, "\n")
}

// formatSpecKey converts snake_case spec keys to Title Case for display.
func formatSpecKey(key string) string {
	parts := strings.Split(key, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
