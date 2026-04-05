# hw-resale System Architecture

> Inventory module of the hg-mcp Go MCP server — 47 tools for hardware liquidation.

## Architecture Overview

```
┌──────────────┐     ┌──────────────┐     ┌───────────────────────────┐
│  Claude Code │────▶│  MCP Server  │────▶│  inventory/module.go      │
│  (stdio/SSE) │     │  (hg-mcp)    │     │  39 tool handlers         │
└──────────────┘     └──────────────┘     └─────────┬─────────────────┘
                                                    │
                     ┌──────────────────────────────┼──────────────────────┐
                     │                              │                      │
              ┌──────▼──────┐              ┌────────▼───────┐    ┌────────▼───────┐
              │ Google      │              │ eBay API       │    │ Local FS       │
              │ Sheets API  │              │ (Browse/Sell)  │    │ (images)       │
              │             │              │                │    │                │
              │ Inventory   │              │ Price research │    │ ~/.local/share │
              │ Sales       │              │ Create listing │    │ /hg-mcp/       │
              │ Config tabs │              │ Category lookup│    │ inventory/     │
              └─────────────┘              └────────────────┘    └────────────────┘
                     │
              ┌──────▼──────┐
              │ Gmail API   │
              │ (import)    │
              └─────────────┘
```

## Google Sheets Schema

### Inventory Tab (Columns A-V, 22 columns)

| Col | Header             | Field           | Type    | Notes                          |
|-----|--------------------|-----------------|---------|--------------------------------|
| A   | #                  | RowNum          | int     | Auto-incremented row number    |
| B   | Category           | Category        | string  | GPU, CPU, RAM, Storage, etc.   |
| C   | Product            | Name            | string  | Item name/title                |
| D   | Model / SKU        | Model           | string  | Manufacturer model number      |
| E   | Qty                | Quantity        | int     | Units in stock                 |
| F   | Condition          | Condition       | string  | new, like_new, good, fair, poor, for_parts |
| G   | Purchase Date      | PurchaseDate    | string  | Free-text (e.g., "Apr 2025")  |
| H   | Amazon ASIN        | ASIN            | string  | Amazon Standard ID             |
| I   | MSRP               | MSRP            | float64 | Manufacturer suggested price   |
| J   | Current Retail     | CurrentRetail   | float64 | Current market retail price     |
| K   | Recommended FB Price| AskingPrice    | float64 | Target asking price            |
| L   | Total Est. Revenue | TotalRevenue    | float64 | = K × E (computed)             |
| M   | Status             | ListingStatus   | string  | not_listed, pending_review, listed, sold, keeping |
| N   | Notes              | Notes           | string  | Free-text notes                |
| O   | SoldPrice          | SoldPrice       | float64 | Actual sale price              |
| P   | SoldDate           | SoldDate        | string  | Date of sale                   |
| Q   | SoldPlatform       | SoldPlatform    | string  | Where it sold                  |
| R   | Location           | Location        | string  | Physical storage location      |
| S   | SmartData          | SmartData       | string  | JSON: smartctl health data     |
| T   | PurchasePrice      | PurchasePrice   | float64 | Original purchase cost         |
| U   | ProductLine        | ProductLine     | string  | Groups identical products (e.g., "Samsung 990 Pro") |
| V   | Specs              | Specs           | string  | JSON-encoded tech specs (e.g., `{"vram":"24GB"}`) |

### Sales Tab (Columns A-O, 15 columns)

| Col | Header        | Field         | Type    | Notes                      |
|-----|---------------|---------------|---------|----------------------------|
| A   | SaleID        | SaleID        | string  | Unique sale identifier     |
| B   | ItemNum       | ItemNum       | int     | Row number from Inventory  |
| C   | ItemName      | ItemName      | string  | Product name               |
| D   | QuantitySold  | QuantitySold  | int     | Units sold                 |
| E   | SoldPrice     | SoldPrice     | float64 | Per-unit sale price        |
| F   | PurchasePrice | PurchasePrice | float64 | Per-unit cost basis        |
| G   | Revenue       | Revenue       | float64 | E × D                     |
| H   | Cost          | Cost          | float64 | F × D                     |
| I   | ShippingCost  | ShippingCost  | float64 | Shipping expense           |
| J   | PlatformFees  | PlatformFees  | float64 | eBay/PayPal fees           |
| K   | NetProfit     | NetProfit     | float64 | G − H − I − J             |
| L   | Platform      | Platform      | string  | ebay, fb_marketplace, etc. |
| M   | BuyerInfo     | BuyerInfo     | string  | Buyer identifier/notes     |
| N   | Notes         | Notes         | string  | Sale notes                 |
| O   | Date          | Date          | string  | Sale date                  |

### Config Tab

Key-value pairs: `TargetRevenue`, `Created`, `LastUpdated`.

## Client Architecture

### SheetsInventoryClient

- **File:** `internal/clients/sheets_inventory.go`
- **Pattern:** Singleton via `GetInventoryClient()` + `sync.Once`
- **Auth:** Application Default Credentials (ADC) — `gcloud auth application-default login`
- **Cache:** 30-second TTL, full inventory loaded on first access, invalidated on writes
- **Test override:** `TestOverrideInventoryClient` var + `NewTestInventoryClient()` for in-memory testing
- **Env var:** `INVENTORY_SPREADSHEET_ID` (defaults to hardcoded ID)

### eBay Client

- **Browse API:** Read-only searches (price research, active listings, category suggestions)
- **Sell API:** OAuth2 with refresh token — `CreateInventoryItem`, `CreateOffer`, `PublishOffer`
- **Env vars:** `EBAY_APP_ID`, `EBAY_CERT_ID`, `EBAY_OAUTH_REFRESH_TOKEN`

### Gmail Client

- **Used by:** `import_gmail` tool
- **Auth:** OAuth2 credentials via `GMAIL_APPLICATION_CREDENTIALS`
- **Known issue:** eBay HTML email parsing broken (see [current_bugs.md](current_bugs.md) BUG-001)

## Listing Templates

Three platform templates in `internal/mcp/tools/inventory/templates.go`:

| Platform       | Template Variable       | Key Features                           |
|----------------|-------------------------|----------------------------------------|
| fb_marketplace | `fbMarketplaceTemplate` | Name + price, condition, specs, "price is firm" |
| ebay           | `ebayTemplate`          | Structured sections: specs, included, shipping, returns |
| hardwareswap   | `hardwareswapTemplate`  | Reddit format: [H]/[W] title, timestamps, SMART data |

**Rendering pipeline:** `RenderListing()` → `formatConditionDisplay()` + `buildSpecsBlock()` → string replacement.

`buildSpecsBlock()` prefers the `Specs` map (col V) when populated. Falls back to ASIN/MSRP metadata fields. Known key ordering: vram, memory, capacity, interface, speed, tdp, form_factor, chipset.

## SMART Data Collection

- **File:** `internal/mcp/tools/inventory/smart.go`
- Runs `smartctl -a -j <device>` and parses JSON output
- Supports NVMe (percentage_used, data_units_written/read) and SATA (reallocated_sectors, total_lbas_written)
- Device path validated to start with `/dev/`
- 30-second timeout per device
- Stored as JSON string in column S

## Image Storage

```
~/.local/share/hg-mcp/inventory/images/{sku}/
```

- Local filesystem storage (no cloud)
- Upload via `upload_image` with local file path
- `set_primary` flag sets the thumbnail/first listing image
- Referenced by listing generators for `{image_urls}` in templates

## Observability

15 Prometheus metric families in `internal/mcp/tools/inventory/metrics.go`:

| Metric | Type | Labels | Purpose |
|--------|------|--------|---------|
| `inventory_items_total` | Counter | operation, category, status | Item CRUD operations |
| `inventory_items_current` | Gauge | category, status, location | Current counts |
| `inventory_imports_total` | Counter | source, status | Import operations |
| `inventory_import_items_total` | Counter | source | Items imported |
| `inventory_import_duration_seconds` | Histogram | source | Import latency |
| `inventory_listings_total` | Counter | platform, status | Listing operations |
| `inventory_ebay_api_total` | Counter | endpoint, status | eBay API calls |
| `inventory_ebay_api_duration_seconds` | Histogram | endpoint | eBay API latency |
| `inventory_sheets_api_total` | Counter | operation, status | Sheets API calls |
| `inventory_sheets_api_duration_seconds` | Histogram | operation | Sheets API latency |
| `inventory_local_image_total` | Counter | operation, status | Image FS operations |
| `inventory_sales_total` | Counter | platform | Sales recorded |
| `inventory_revenue_dollars` | Gauge | — | Cumulative revenue |
| `inventory_value_dollars` | Gauge | type | Inventory value |
| `inventory_gmail_total` | Counter | operation, status | Gmail API calls |
| `inventory_gmail_orders_found` | Counter | source | Orders parsed |
| `inventory_errors_total` | Counter | operation, error_type | Error tracking |
| `inventory_last_operation_timestamp` | Gauge | operation | Last activity |

All tool handlers are also instrumented by the global `wrapHandler()` in `registry.go` (tracing, duration, panic recovery, 30s default timeout).

## Complete Tool Inventory (41 tools)

### CRUD (7)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_list` | `handleInventoryList` | List/filter items |
| `aftrs_inventory_get` | `handleInventoryGet` | Get item by SKU |
| `aftrs_inventory_add` | `handleInventoryAdd` | Add new item |
| `aftrs_inventory_update` | `handleInventoryUpdate` | Update item fields |
| `aftrs_inventory_delete` | `handleInventoryDelete` | Delete item |
| `aftrs_inventory_bulk_delete` | `handleInventoryBulkDelete` | Bulk delete with filters |
| `aftrs_inventory_search` | `handleInventorySearch` | Full-text search |

### Import (4)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_import_amazon` | `handleImportAmazon` | Amazon CSV import |
| `aftrs_inventory_import_newegg` | `handleImportNewegg` | Newegg CSV import |
| `aftrs_inventory_import_gmail` | `handleImportGmail` | Gmail order parsing |
| `aftrs_inventory_import_json` | `handleImportJSON` | JSON file import (hw-resale compat) |

### Analytics (8)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_summary` | `handleInventorySummary` | Inventory overview |
| `aftrs_inventory_stale` | `handleInventoryStale` | Find stale items |
| `aftrs_inventory_value` | `handleInventoryValue` | Valuation breakdown |
| `aftrs_inventory_export` | `handleInventoryExport` | CSV/JSON export |
| `aftrs_inventory_price_check` | `handlePriceCheck` | Margin analysis |
| `aftrs_inventory_sales_log` | `handleSalesLog` | P&L history |
| `aftrs_inventory_reprice` | `handleReprice` | eBay-based repricing |
| `aftrs_inventory_shipping_estimate` | `handleShippingEstimate` | Carrier cost estimation |

### Listing (6)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_mark_sold` | `handleMarkSold` | Mark item sold |
| `aftrs_inventory_fb_content` | `handleFBContent` | FB Marketplace content |
| `aftrs_inventory_listing_generate` | `handleListingGenerate` | Multi-platform listing text |
| `aftrs_inventory_sale_record` | `handleSaleRecord` | Record sale with P&L |
| `aftrs_inventory_batch_listing` | `handleBatchListing` | Bulk listing generation |
| `aftrs_inventory_quick_list` | `handleQuickList` | One-command: pricing + all 3 listings |

### Images (4)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_upload_image` | `handleUploadImage` | Upload image file |
| `aftrs_inventory_list_images` | `handleListImages` | List item images |
| `aftrs_inventory_delete_image` | `handleDeleteImage` | Delete image |
| `aftrs_inventory_set_primary_image` | `handleSetPrimaryImage` | Set primary image |

### eBay Integration (4)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_ebay_price_research` | `handleEbayPriceResearch` | Sold listing comps |
| `aftrs_inventory_ebay_active_listings` | `handleEbayActiveListings` | Active listings search |
| `aftrs_inventory_ebay_categories` | `handleEbayCategories` | Category suggestions |
| `aftrs_inventory_ebay_create_listing` | `handleEbayCreateListing` | Create + publish listing |

### Data Quality (4)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_recategorize` | `handleInventoryRecategorize` | Auto-categorize items |
| `aftrs_inventory_duplicates` | `handleInventoryDuplicates` | Find duplicates |
| `aftrs_inventory_bulk_update` | `handleInventoryBulkUpdate` | Batch field updates |
| `aftrs_inventory_smart_collect` | `handleSmartCollect` | SMART data from drives |

### Organization (3)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_categories` | `handleInventoryCategories` | List categories |
| `aftrs_inventory_locations` | `handleInventoryLocations` | List locations |
| `aftrs_inventory_move` | `handleInventoryMove` | Move item location |

### Dashboard (1)
| Tool | Handler | Description |
|------|---------|-------------|
| `aftrs_inventory_dashboard` | `handleInventoryDashboard` | Full dashboard with financials |

## Key Files

| File | Purpose |
|------|---------|
| `internal/clients/inventory_types.go` | Data model: InventoryItem (A-V), SaleRecord, filters |
| `internal/clients/sheets_inventory.go` | Sheets backend: cache, CRUD, TestOverride |
| `internal/clients/sheets_inventory_sales.go` | Sales tab: P&L tracking |
| `internal/mcp/tools/inventory/module.go` | 39 tool definitions + all handlers |
| `internal/mcp/tools/inventory/templates.go` | 3 listing templates + RenderListing() |
| `internal/mcp/tools/inventory/metrics.go` | 15 Prometheus metric families |
| `internal/mcp/tools/inventory/smart.go` | SMART data collection (smartctl parser) |
| `internal/mcp/tools/inventory/handlers_test.go` | Handler test suite |

## Related Documentation

- [INVENTORY-OPS.md](INVENTORY-OPS.md) — Daily workflow quickstart
- [current_bugs.md](current_bugs.md) — Known bugs and limitations
- [testing.md](testing.md) — Test architecture and coverage
- [research.md](research.md) — Platform research and pricing benchmarks
- [objectives.md](objectives.md) — Revenue targets and KPIs
- [roadmap.md](roadmap.md) — Sales execution plan
- [suggestions.md](suggestions.md) — Architecture improvement recommendations
