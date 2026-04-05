# hw-resale Architecture Improvement Recommendations

> Prioritized suggestions for improving the inventory module's reliability, DX, and automation.

## Priority 1: Data Model Fixes

### Enhancement A: Condition/ListingStatus Validation

**Problem:** `Condition` and `ListingStatus` are `string` fields in `InventoryItem`. The schema defines enum constraints via `"enum"` in tool InputSchema, but handlers don't validate at write time. Typos like "good condition" or "Listed" (capitalized) create orphan values.

**Fix:** Add validation in `handleInventoryAdd` and `handleInventoryUpdate`:
```go
var validConditions = map[string]bool{
    "new": true, "like_new": true, "good": true,
    "fair": true, "poor": true, "for_parts": true,
}

// In handler, before writing:
condition := strings.ToLower(tools.GetStringParam(req, "condition"))
if condition != "" && !validConditions[condition] {
    return tools.CodedErrorResult(tools.ErrInvalidParam,
        fmt.Errorf("invalid condition %q (valid: new, like_new, good, fair, poor, for_parts)", condition)), nil
}
```

**Impact:** Prevents data drift in columns F (Condition) and M (ListingStatus).
**Files:** `internal/mcp/tools/inventory/module.go`
**See also:** [current_bugs.md](current_bugs.md) DATA-001

### Enhancement B: Specs Catalog

**Problem:** Column V (Specs) is empty for most items, causing `buildSpecsBlock()` to fall back to generic ASIN/MSRP metadata in listings. Listings without real specs (VRAM, capacity, interface) look generic and sell slower.

**Fix:** Create a `specsCatalog` map keyed by ProductLine, with default specs. Use in:
- `import_json`: auto-populate Specs if ProductLine matches catalog
- `price_check`: include suggested specs in output
- New `quick_list` tool: auto-enrich from catalog

```go
var specsCatalog = map[string]map[string]string{
    "RTX 4090": {"vram": "24GB GDDR6X", "tdp": "450W", "interface": "PCIe 4.0 x16"},
    "Samsung 990 Pro": {"interface": "PCIe 4.0 x4 NVMe", "form_factor": "M.2 2280"},
    // ...
}
```

**Impact:** Richer listing content without manual data entry per item.
**Files:** New `specs_catalog.go` in inventory package

### Date Normalization (Future)

**Problem:** `PurchaseDate` (col G) is free-text — "Apr 2025", "2024-12-15", "last month".
**Fix:** Parse on write using `time.Parse` with multiple format attempts, store as ISO 8601.
**Impact:** Enables date-based sorting and stale detection.

### SKU Generation (Future)

**Problem:** SKUs are `HW-{RowNum}` — risk of collision on delete+re-add.
**Fix:** Use monotonically increasing counter stored in Config tab, or UUID prefix.

## Priority 2: Tool Enhancements

### Enhancement C: Quick List Workflow Tool

**Problem:** Listing an item requires 3+ tool calls: `price_check` → `listing_generate` (×3 platforms) → manual price decision. This is the most common workflow.

**Fix:** New `aftrs_inventory_quick_list` tool that chains:
1. Fetch item by SKU
2. Run margin analysis (like `price_check`)
3. Generate listing text for all 3 platforms
4. Return everything in one response

```
Input:  {"sku": "HW-001"}
Output: {
    "item": {...},
    "pricing": {"asking": 1200, "margin_pct": 33, "guidance": "..."},
    "listings": {
        "fb_marketplace": "ASUS ROG Strix RTX 4090 — $1200\n...",
        "ebay": "ASUS ROG Strix RTX 4090\n\nCONDITION: Used — Good\n...",
        "hardwareswap": "[USA-XX] [H] ASUS ROG Strix RTX 4090 [W] PayPal..."
    }
}
```

**Impact:** Reduces listing workflow from 4+ tool calls to 1.
**Files:** `internal/mcp/tools/inventory/module.go`

### Enhancement D: Stale Alerts in Dashboard

**Problem:** Dashboard shows "stale_30_days" count but doesn't flag items needing repricing (listed >7 days) or items below floor price.

**Fix:** Enhance `handleInventoryDashboard` to include:
- `needs_repricing`: items listed >7 days (configurable)
- `below_floor_price`: items where AskingPrice < PurchasePrice × 0.5

**Impact:** Dashboard becomes actionable — shows exactly what needs attention.
**Files:** `internal/mcp/tools/inventory/module.go` (dashboard handler)

### Enhancement E: Shipping Estimate Tool

**Problem:** No tool to estimate shipping costs. Users must look up rates manually. This affects P&L planning for eBay/hardwareswap sales.

**Fix:** New `aftrs_inventory_shipping_estimate` tool with weight-based tier estimation:
```
Input:  {"sku": "HW-001"} or {"weight_lbs": 4, "length": 12, "width": 10, "height": 6}
Output: {
    "estimates": {
        "usps_priority": {"cost": 18.50, "days": "2-3"},
        "ups_ground":    {"cost": 14.20, "days": "3-5"},
        "usps_ground":   {"cost": 11.80, "days": "5-7"}
    },
    "recommendation": "UPS Ground — best value for 4lb GPU shipment",
    "insurance_note": "Include insurance ($1,350 value) — add ~$5"
}
```

When given a SKU, estimate weight from category defaults. When given explicit dimensions, use those.

**Impact:** More accurate P&L planning; helps decide local-only vs shipped pricing.
**Files:** `internal/mcp/tools/inventory/module.go`

### Price Tracking (Future)

**Problem:** No automated price monitoring. The `reprice` tool must be called manually per SKU.
**Fix:** Cron-triggered batch reprice that flags items where market has shifted >10%.

### eBay Sold Webhook (Future)

**Problem:** Must manually check if eBay listings sold.
**Fix:** Poll eBay Selling API for status changes, prompt `sale_record` creation.

## Priority 3: Workflow Improvements

### Template Customization (Future)

**Problem:** Templates are Go constants in `templates.go`. Users can't customize without code changes.
**Fix:** Load templates from Config tab or local file. Fall back to hardcoded defaults.

### Morning Dashboard Enhancement

The current dashboard is good. Could be enhanced with:
- Revenue progress bar (ASCII)
- Top 3 recommended actions with specific tool commands
- Market price changes since last check

## Priority 4: Prompt Engineering Tips

For best results when using inventory tools with Claude Code:

1. **"Check my inventory dashboard"** → triggers `aftrs_inventory_dashboard`
2. **"List my GPUs"** → triggers `aftrs_inventory_list category=GPU`
3. **"How much can I get for HW-001?"** → triggers `aftrs_inventory_price_check sku=HW-001` + `aftrs_inventory_ebay_price_research`
4. **"Create listings for all unlisted GPUs"** → triggers `aftrs_inventory_batch_listing category=GPU status=not_listed`
5. **"I sold HW-001 for $1200 on eBay, shipping was $18"** → triggers `aftrs_inventory_sale_record`

## Priority 5: Documentation Debt

### Tool Count Sync

The README says "730+ tools across 78 modules" but actual count is 863+ across 98 modules. Inventory has 41 tools.

**Files to update:**
- `README.md` — module/tool counts
- `CLAUDE.md` — tool count reference

### Outdated Root Docs

- `docs/ARCHITECTURE.md` — may not reflect recent changes
- `docs/Roadmap.md` — generic project roadmap, not hw-resale-specific
- These should link to the new hw-resale doc suite

## Implementation Status

| Enhancement | Status | Priority |
|-------------|--------|----------|
| A: Condition/Status validation | **Implemented** | P1 |
| B: Specs catalog | **Implemented** | P1 |
| C: Quick list tool | **Implemented** | P2 |
| D: Dashboard stale alerts | **Implemented** | P2 |
| E: Shipping estimate | **Implemented** | P2 |
| Date normalization | Future | P3 |
| SKU generation | Future | P3 |
| Price tracking | Future | P3 |
| eBay webhook | Future | P4 |
| Template customization | Future | P4 |

## Related Documentation

- [system.md](system.md) — Architecture: 41 tools, Sheets schema (A-V, 22 columns)
- [current_bugs.md](current_bugs.md) — Bugs these suggestions address
- [testing.md](testing.md) — Coverage gaps for new tools
- [roadmap.md](roadmap.md) — Sales timeline that benefits from these enhancements
- [objectives.md](objectives.md) — $16,225 revenue target
- [research.md](research.md) — Market data driving pricing tools
