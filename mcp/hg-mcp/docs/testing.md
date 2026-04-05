# hw-resale Testing Guide

> Test architecture, coverage analysis, and how to add new tests.

## Test Overview

The inventory module has tests across 5 files:

| File | Tests | Coverage Area |
|------|-------|---------------|
| `handlers_test.go` | 106+ | Handler functions for all CRUD, analytics, listing, data quality tools |
| `templates_test.go` | 5 | RenderListing(), buildSpecsBlock(), formatConditionDisplay() |
| `smart_test.go` | 2 | CollectSMARTData() JSON parsing |
| `module_test.go` | 1 | Module registration and tool count |
| `sheets_inventory_test.go` (clients/) | 10 | Client CRUD, cache, filter, TestOverride |

## Running Tests

```bash
# All inventory tests
go test ./internal/mcp/tools/inventory/...

# All client tests (includes sheets_inventory)
go test ./internal/clients/...

# Specific test by name
go test ./internal/mcp/tools/inventory/... -run TestHandleInventoryList

# Verbose output
go test -v ./internal/mcp/tools/inventory/...

# Skip slow/integration tests
go test -short ./internal/mcp/tools/inventory/...

# All project tests
go test ./...
go test -short ./...
```

## TestOverride Architecture

Tests use an in-memory mock instead of hitting Google Sheets:

```go
// In TestMain (handlers_test.go):
func TestMain(m *testing.M) {
    clients.TestOverrideInventoryClient = clients.NewTestInventoryClient(seedItems())
    code := m.Run()
    clients.TestOverrideInventoryClient = nil
    os.Exit(code)
}
```

`NewTestInventoryClient()` creates a `SheetsInventoryClient` with `testMode: true`. All reads come from the in-memory cache; all writes update cache only (no Sheets API calls).

### Seed Dataset (5 items)

| SKU    | Name                         | Category    | Price | Condition | Status     |
|--------|------------------------------|-------------|-------|-----------|------------|
| HW-001 | ASUS ROG Strix RTX 4090      | GPU         | $1800 | Used      | Not Listed |
| HW-002 | Samsung 990 Pro 2TB NVMe     | Storage     | $140  | New       | Not Listed |
| HW-003 | Corsair Vengeance DDR5 32GB  | RAM         | $90   | Like New  | Not Listed |
| HW-004 | Netgear Nighthawk Router     | Networking  | $150  | Used      | Listed     |
| HW-005 | ASUS ROG Strix RTX 4090      | GPU         | $1700 | Renewed   | Not Listed |

Items HW-001 and HW-005 are intentional duplicates (same name, different SKU/condition) for duplicate detection tests.

### Reset Between Tests

```go
func resetSeed() {
    clients.TestOverrideInventoryClient = clients.NewTestInventoryClient(seedItems())
}
```

Call `resetSeed()` at the start of any test that modifies inventory state.

## Coverage by Tool Category

### CRUD (7 tools) — Well Covered

| Tool | Tests | Status |
|------|-------|--------|
| `inventory_list` | filter by category, status, condition, query, source, limit | Covered |
| `inventory_get` | valid SKU, missing SKU, empty SKU | Covered |
| `inventory_add` | full fields, minimal fields, missing required | Covered |
| `inventory_update` | single field, multiple fields, missing SKU | Covered |
| `inventory_delete` | valid, missing confirm, missing SKU | Covered |
| `inventory_bulk_delete` | dry_run, with source filter | Covered |
| `inventory_search` | query match, no results | Covered |

### Import (4 tools) — Partial

| Tool | Tests | Status |
|------|-------|--------|
| `import_amazon` | missing csv_path, invalid path | **Partial** (no valid CSV fixture) |
| `import_newegg` | missing csv_path, invalid path | **Partial** (no valid CSV fixture) |
| `import_gmail` | — | **Not covered** (needs Gmail mock) |
| `import_json` | valid JSON, dry_run, invalid path, empty file | Covered |

### Analytics (8 tools) — Well Covered

| Tool | Tests | Status |
|------|-------|--------|
| `inventory_summary` | basic call, category counts | Covered |
| `inventory_stale` | default days, custom days | Covered |
| `inventory_value` | basic, include_sold, group_by | Covered |
| `inventory_export` | JSON, CSV format | Covered |
| `price_check` | by SKU, by search_query, missing both | Covered |
| `sales_log` | basic call | Covered |
| `reprice` | dry_run, strategy options | Covered |
| `shipping_estimate` | by SKU, by weight, missing params | Covered |

### Listing (6 tools) — Well Covered

| Tool | Tests | Status |
|------|-------|--------|
| `mark_sold` | valid, missing SKU | Covered |
| `fb_content` | valid SKU, with asking_price | Covered |
| `listing_generate` | each platform, default platform | Covered |
| `sale_record` | valid, missing fields | Covered |
| `batch_listing` | all platforms, filtered | Covered |
| `quick_list` | valid SKU, missing SKU, not found | Covered |

### Images (4 tools) — Partial

| Tool | Tests | Status |
|------|-------|--------|
| `upload_image` | missing params | **Partial** (no valid file fixture) |
| `list_images` | valid SKU | Covered |
| `delete_image` | missing params | **Partial** |
| `set_primary_image` | missing params | **Partial** |

### eBay Integration (4 tools) — Not Covered

| Tool | Tests | Status |
|------|-------|--------|
| `ebay_price_research` | — | **Not covered** (needs eBay client mock) |
| `ebay_active_listings` | — | **Not covered** |
| `ebay_categories` | — | **Not covered** |
| `ebay_create_listing` | — | **Not covered** |

### Data Quality (4 tools) — Covered

| Tool | Tests | Status |
|------|-------|--------|
| `recategorize` | single SKU, dry_run, apply | Covered |
| `duplicates` | all, asin, order_id match types | Covered |
| `bulk_update` | dry_run, with filters | Covered |
| `smart_collect` | — | **Partial** (smart.go unit tested, handler not) |

### Organization (3 tools) — Covered

| Tool | Tests | Status |
|------|-------|--------|
| `categories` | basic call | Covered |
| `locations` | basic call | Covered |
| `move` | valid, missing params | Covered |

### Dashboard (1 tool) — Covered

| Tool | Tests | Status |
|------|-------|--------|
| `dashboard` | basic call, financials present | Covered |

## Coverage Gaps Summary

1. **eBay tools (4):** Zero tests. Need `TestOverrideEbayClient` similar to inventory pattern.
2. **Gmail import:** Needs mock Gmail client. Complex — HTML parsing, multi-source support.
3. **Amazon/Newegg CSV import:** Need fixture CSV files in `testdata/` directory.
4. **Image upload/delete:** Need temp directory fixtures for file operations.
5. **SMART collect handler:** `smart.go` has unit tests for parsing, but the handler itself (which calls `exec.Command`) is untested.

## How to Add Tests

### Pattern: Simple Handler Test

```go
func TestHandleMyTool(t *testing.T) {
    resetSeed()

    req := makeReq(map[string]interface{}{
        "sku": "HW-001",
    })

    result, err := handleMyTool(context.Background(), req)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    var resp map[string]interface{}
    getJSON(t, result, &resp)

    if resp["success"] != true {
        t.Errorf("expected success=true, got %v", resp["success"])
    }
}
```

### Pattern: Error Case Test

```go
func TestHandleMyTool_MissingSKU(t *testing.T) {
    req := makeReq(map[string]interface{}{})

    result, err := handleMyTool(context.Background(), req)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if !result.IsError {
        t.Error("expected error result for missing SKU")
    }
}
```

### Adding a Mock Client for Untested Integrations

Follow the `TestOverrideInventoryClient` pattern:

1. In the client package, add a `TestOverride*Client` var
2. In `Get*Client()`, check the override first
3. Create `NewTest*Client()` with in-memory data
4. In `TestMain`, set the override before `m.Run()`

## Related Documentation

- [system.md](system.md) — Architecture and tool inventory
- [current_bugs.md](current_bugs.md) — Known issues (some affect test coverage decisions)
- [suggestions.md](suggestions.md) — Recommended testing improvements
- [INVENTORY-OPS.md](INVENTORY-OPS.md) — Tool reference for understanding expected behavior
