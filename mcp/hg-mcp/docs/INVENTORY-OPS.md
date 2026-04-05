# Inventory Operations Quickstart

Hardware resale inventory management via Google Sheets backend + MCP tools.

## Authentication Setup

### Google Sheets (required)

```bash
# Install gcloud CLI, then authenticate:
gcloud auth application-default login \
  --scopes=https://www.googleapis.com/auth/spreadsheets

# Set spreadsheet ID (or use default in .env.example):
export INVENTORY_SPREADSHEET_ID=YOUR_SPREADSHEET_ID
```

The backend uses Application Default Credentials (ADC). Your Google account must have edit access to the spreadsheet.

### Gmail Import (optional)

Requires OAuth2 credentials for Gmail access. Run `cmd/gmail-auth` to complete the OAuth flow, then set:

```bash
export GMAIL_APPLICATION_CREDENTIALS=/path/to/gmail-credentials.json
```

### eBay API (optional)

For eBay price research and listing creation:

```bash
export EBAY_APP_ID=...
export EBAY_CERT_ID=...
export EBAY_OAUTH_REFRESH_TOKEN=...
```

## Daily Workflow

### 1. Review Inventory

```
aftrs_inventory_list                          # Browse all items
aftrs_inventory_list status=not_listed        # Items ready to list
aftrs_inventory_stale days=30                 # Items sitting too long
aftrs_inventory_dashboard                     # Full overview with stats
```

### 2. Price Check

```
aftrs_inventory_price_check sku=HW-0042      # Margin analysis + guidance
aftrs_inventory_ebay_price_research query="RTX 3070"  # Market comps
aftrs_inventory_value                         # Portfolio valuation
```

### 3. Generate Listings

```
aftrs_inventory_quick_list sku=HW-0042       # One-shot: pricing + all 3 platform listings
aftrs_inventory_listing_generate sku=HW-0042 platform=fb_marketplace
aftrs_inventory_listing_generate sku=HW-0042 platform=ebay
aftrs_inventory_listing_generate sku=HW-0042 platform=hardwareswap
aftrs_inventory_batch_listing category=GPU    # Bulk: all GPU listings at once
aftrs_inventory_fb_content sku=HW-0042       # FB Marketplace formatted post
aftrs_inventory_ebay_create_listing sku=HW-0042 price=199.99 category_id=...
aftrs_inventory_shipping_estimate sku=HW-0042 # Estimate shipping cost for P&L planning
```

### 4. Record Sales

```
aftrs_inventory_sale_record sku=HW-0042 sold_price=185 platform=ebay \
  shipping_cost=12.50 platform_fees=24.42

aftrs_inventory_mark_sold sku=HW-0042 sold_price=185 platform=ebay
```

### 5. Review P&L

```
aftrs_inventory_sales_log                     # All sales + summary
```

## Tool Reference

### CRUD Operations

| Tool | Description |
|------|-------------|
| `aftrs_inventory_list` | List/filter items (category, status, location, price range, text search) |
| `aftrs_inventory_get` | Get item details by SKU |
| `aftrs_inventory_add` | Add new item (auto-generates SKU) |
| `aftrs_inventory_update` | Update item fields |
| `aftrs_inventory_delete` | Delete item (requires confirm=true) |
| `aftrs_inventory_search` | Full-text search across all fields |
| `aftrs_inventory_move` | Move item to new location |
| `aftrs_inventory_recategorize` | Change item category |
| `aftrs_inventory_bulk_update` | Bulk update multiple items |
| `aftrs_inventory_bulk_delete` | Bulk delete with filters |
| `aftrs_inventory_duplicates` | Find duplicate items |

### Import

| Tool | Description |
|------|-------------|
| `aftrs_inventory_import_amazon` | Import from Amazon order CSV |
| `aftrs_inventory_import_newegg` | Import from Newegg order CSV |
| `aftrs_inventory_import_gmail` | Parse purchase emails (Amazon working; eBay broken — see below) |
| `aftrs_inventory_import_json` | Import from JSON file (supports Python hw-resale format) |

### Analytics & Reporting

| Tool | Description |
|------|-------------|
| `aftrs_inventory_summary` | Category counts and status breakdown |
| `aftrs_inventory_categories` | List all categories |
| `aftrs_inventory_locations` | List all locations |
| `aftrs_inventory_stale` | Items not updated in N days |
| `aftrs_inventory_dashboard` | Full dashboard with stats |
| `aftrs_inventory_value` | Total inventory valuation |
| `aftrs_inventory_export` | Export inventory to JSON/CSV |

### Listing & Sales

| Tool | Description |
|------|-------------|
| `aftrs_inventory_listing_generate` | Generate listing text (fb_marketplace, ebay, hardwareswap) |
| `aftrs_inventory_fb_content` | Facebook Marketplace formatted content |
| `aftrs_inventory_mark_sold` | Mark item as sold |
| `aftrs_inventory_sale_record` | Record sale with P&L tracking |
| `aftrs_inventory_price_check` | Margin analysis and pricing guidance |
| `aftrs_inventory_sales_log` | View all sales + P&L summary |
| `aftrs_inventory_batch_listing` | Bulk-generate listings for filtered items across platforms |
| `aftrs_inventory_quick_list` | One-command: pricing analysis + listing text for all 3 platforms |
| `aftrs_inventory_reprice` | Suggest repricing from eBay completed listings (fast_sale/median/max_profit) |
| `aftrs_inventory_shipping_estimate` | Estimate shipping cost by carrier (USPS/UPS) from SKU or weight |

### eBay Integration

| Tool | Description |
|------|-------------|
| `aftrs_inventory_ebay_price_research` | Search eBay sold/active listings for comps |
| `aftrs_inventory_ebay_active_listings` | View your active eBay listings |
| `aftrs_inventory_ebay_categories` | Search eBay categories |
| `aftrs_inventory_ebay_create_listing` | Create and publish eBay listing |

### Images

| Tool | Description |
|------|-------------|
| `aftrs_inventory_upload_image` | Upload image for item |
| `aftrs_inventory_list_images` | List images for item |
| `aftrs_inventory_delete_image` | Delete an image |
| `aftrs_inventory_set_primary_image` | Set primary display image |

### Hardware

| Tool | Description |
|------|-------------|
| `aftrs_inventory_smart_collect` | Collect SMART data from drives |

## Image Management

Images are stored on the local filesystem:

```
~/.local/share/hg-mcp/inventory/images/{sku}/
```

Upload via `aftrs_inventory_upload_image` with `file_path` pointing to the image. Set `set_primary=true` to make it the main display image.

Listing generators (`listing_generate`, `fb_content`) reference images from the item's `images` array.

## Known Limitations

1. **Gmail import — eBay emails broken**: eBay sends HTML emails. The parser extracts `<table>` tags as item names and misparses prices. Amazon parsing works well. See `docs/inventory-import-improvements.md` for details.

2. **Duplicate detection**: Gmail re-import can create duplicates. Current check uses order_id only — needs order_id + item_name combination check.

3. **Newegg/BestBuy import**: Untested (no sample emails in initial import run).

4. **eBay API rate limits**: The eBay Browse API has strict rate limits. Price research may throttle under heavy use.

5. **Google Sheets rate limits**: Sheets API allows ~100 requests/100 seconds. The client uses a 30-second cache TTL to stay well under limits.
