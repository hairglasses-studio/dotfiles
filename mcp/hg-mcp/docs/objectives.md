# hw-resale Objectives and KPIs

> Revenue targets, key performance indicators, and success criteria.

## Revenue Target

| Metric | Value |
|--------|-------|
| **Total Target Revenue** | $16,225 |
| **Target Net Profit** | $10,000+ |
| **Expected Fee Loss** | ~$1,200 (weighted avg ~8% across platforms) |
| **Expected Shipping Cost** | ~$500 |
| **Timeline** | 6 weeks |

## KPI Table

| KPI | Target | How to Measure |
|-----|--------|----------------|
| Items listed | 100% of sellable inventory | `aftrs_inventory_list status=not_listed` → 0 remaining |
| Items sold | 80%+ of listed items | `aftrs_inventory_dashboard` → sold count / total |
| Avg days to sale | <14 days (FB), <21 days (eBay), <7 days (hardwareswap) | Sale date − list date |
| Gross margin | >50% average | `aftrs_inventory_sales_log` → total_revenue / total_cost |
| Fee ratio | <10% of revenue | `aftrs_inventory_sales_log` → total_fees / total_revenue |
| Net profit per item | >$50 average | total_profit / units_sold |
| Listing quality | 3+ photos per item, specs populated | `aftrs_inventory_list_images` per SKU |
| Price accuracy | Within 10% of eBay median | `aftrs_inventory_reprice` suggestions vs asking |

## Category Priorities (Ranked by Expected Revenue)

| Priority | Category | Est. Items | Est. Revenue | Target Platform | Notes |
|----------|----------|-----------|-------------|-----------------|-------|
| 1 | GPU | 3-5 | $4,000-6,000 | eBay + r/hardwareswap | Highest value per item |
| 2 | Storage (NVMe/SSD) | 5-10 | $1,500-2,500 | r/hardwareswap | Include SMART data |
| 3 | Networking/Thunderbolt | 3-5 | $1,000-2,000 | eBay | Niche, needs wide exposure |
| 4 | RAM (DDR5) | 3-5 | $500-1,000 | r/hardwareswap | Commodity, price-sensitive |
| 5 | Complete Systems | 1-2 | $1,000-2,000 | FB Marketplace | Local pickup preferred |
| 6 | Peripherals/Cables | 5-10 | $500-1,000 | FB Marketplace | Bundle deals |
| 7 | Components (misc) | 5-10 | $500-1,000 | FB/eBay | Case-by-case |

## 6-Week Sell-Through Timeline

### Week 1: Setup & FB Marketplace

- Populate Specs column for all high-value items
- Run `recategorize` on all items in "Other" category
- Generate FB Marketplace listings for 5-8 highest-value local-friendly items
- Target: $2,000-3,000 in local sales

### Week 2: eBay Listings Launch

- List top 10 items on eBay (GPUs, networking, Thunderbolt)
- Use `ebay_price_research` for competitive pricing
- Run `smart_collect` on all storage devices
- Target: 10 eBay listings live

### Week 3: r/hardwareswap + Repricing

- Post storage devices and RAM on r/hardwareswap with SMART data
- Run `reprice` on items listed >7 days without interest
- Target: $3,000-4,000 cumulative revenue

### Week 4: Acceleration

- Relist any stale items with price drops (5-10%)
- Bundle slow-moving items
- Cross-post popular items across platforms
- Target: $8,000-10,000 cumulative revenue

### Week 5-6: Close Out

- Aggressive repricing on remaining inventory
- Lot sales for low-value remaining items
- Final P&L export
- Target: $14,000-16,225 cumulative revenue

## Success Criteria

| Criteria | Threshold | Measurement |
|----------|-----------|-------------|
| Revenue target hit | ≥$16,225 gross | `aftrs_inventory_sales_log` total_revenue |
| Net profit target | ≥$10,000 | `aftrs_inventory_sales_log` total_profit |
| Inventory cleared | ≥80% of items sold or keeping | Status breakdown in dashboard |
| Zero data loss | No items lost or corrupted | Inventory count matches audit |
| All sales recorded | Every sale has P&L entry | Sales tab count matches sold items |

## Prometheus Metrics for Tracking

These metrics from `internal/mcp/tools/inventory/metrics.go` directly measure progress:

| Metric | What it Tracks |
|--------|---------------|
| `inventory_revenue_dollars` | Cumulative revenue toward $16,225 target |
| `inventory_sales_total{platform}` | Sales by platform (validates platform strategy) |
| `inventory_items_current{status="sold"}` | Sold item count |
| `inventory_value_dollars{type="asking"}` | Remaining unsold inventory value |
| `inventory_listings_total{platform,status}` | Listing activity across platforms |
| `inventory_last_operation_timestamp` | Tracks when tools were last used |

## Revenue Target Source

The $16,225 target comes from the Google Sheets Config tab (`TargetRevenue` field). The dashboard tool reads this and displays progress:

```
"revenue_target": 16225,
"progress_pct": "12.3%",
"remaining_to_target": 14225
```

Adjust the target in the Config tab of the spreadsheet if needed.

## Related Documentation

- [system.md](system.md) — Architecture: 41 tools, Google Sheets schema (A-V, 22 columns)
- [research.md](research.md) — Platform fees, pricing benchmarks, best practices
- [roadmap.md](roadmap.md) — Detailed execution plan by phase
- [current_bugs.md](current_bugs.md) — Issues that may affect sell-through
- [INVENTORY-OPS.md](INVENTORY-OPS.md) — Daily workflow with tool commands
