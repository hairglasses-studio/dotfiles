# hw-resale Execution Roadmap

> Phase-by-phase sales plan for liquidating hardware inventory via the 39-tool inventory module.

## Phase 0: Data Cleanup (Days 1-2)

**Goal:** Ensure data quality before listing anything.

### Tasks

1. **Populate Specs column** for all high-value items (GPUs, NVMe, DDR5)
   ```
   aftrs_inventory_update sku=HW-001 specs={"vram":"24GB GDDR6X","tdp":"450W","interface":"PCIe 4.0 x16"}
   ```

2. **Fix categories** — recategorize items stuck in "Other"
   ```
   aftrs_inventory_recategorize category=Other dry_run=true
   aftrs_inventory_recategorize category=Other dry_run=false
   ```

3. **Run dedup** — find and merge any duplicates from import
   ```
   aftrs_inventory_duplicates match_type=all
   ```

4. **Collect SMART data** for all storage devices
   ```
   aftrs_inventory_smart_collect sku=HW-002 device_path=/dev/nvme0n1
   ```

5. **Take product photos** — minimum 3 per item, upload via `upload_image`

6. **Set asking prices** using `price_check` and `ebay_price_research`

### Exit Criteria
- Zero items in "Other" category
- Specs populated for all items with AskingPrice > $100
- SMART data on all storage devices
- At least 1 image per item planned for listing

## Phase 1: FB Marketplace Local Sales (Week 1-2)

**Goal:** $3,000-5,000 from zero-fee local sales.

### Target Items
- Complete systems (heavy, not worth shipping)
- GPUs (high-value, local buyers exist)
- Peripherals and cables (not worth shipping individually)
- Bundle deals (3+ items at discount)

### Workflow
```
# Generate listing for each item
aftrs_inventory_listing_generate sku=HW-001 platform=fb_marketplace

# Or use batch generation
aftrs_inventory_batch_listing status=not_listed platform=fb_marketplace

# Also generate FB-specific formatted content
aftrs_inventory_fb_content sku=HW-001

# After sale, record it
aftrs_inventory_sale_record sku=HW-001 sold_price=1200 platform=fb_marketplace
```

### Pricing Strategy
- Price at 80-90% of eBay sold median (no fees = higher net)
- "Price is firm" — avoid lowballers
- Bundle discount: 10% off for 2+ items, 15% for 3+
- Accept cash or Zelle only (avoid Venmo for chargebacks)

### Expected Revenue: $3,000-5,000

## Phase 2: eBay Listings (Week 2-4)

**Goal:** $6,000-8,000 from eBay (accepting 13.25% fee for wide exposure).

### Target Items
- GPUs not sold locally
- Networking gear (niche items need eBay search traffic)
- Thunderbolt docks (brand-specific demand, eBay is strongest)
- Any high-value item not moving on FB

### Workflow
```
# Research pricing
aftrs_inventory_ebay_price_research query="RTX 4090 Founders Edition"

# Find the right category
aftrs_inventory_ebay_categories query="RTX 4090"

# Create listing
aftrs_inventory_ebay_create_listing sku=HW-001 price=1350 category_id=27386

# Monitor active listings
aftrs_inventory_ebay_active_listings query="RTX 4090"

# After sale, record with fees
aftrs_inventory_sale_record sku=HW-001 sold_price=1350 platform=ebay \
  shipping_cost=18 platform_fees=178.88
```

### Pricing Strategy
- Start at eBay sold median (from `ebay_price_research`)
- Free shipping (build $15-20 into price)
- 30-day fixed price listings
- Promoted listings at 2-3% for items >$200
- Calculate net: `sold_price × 0.8675 − shipping`

### Expected Revenue: $6,000-8,000

## Phase 3: r/hardwareswap + Remaining (Week 3-6)

**Goal:** $2,000-3,000 from Reddit sales (low fees, enthusiast community).

### Target Items
- NVMe SSDs (SMART data is a strong differentiator)
- RAM kits (commodity, price-sensitive buyers)
- GPUs (lower than eBay but lower fees)
- Components that enthusiasts want

### Workflow
```
# Generate hardwareswap-formatted listing
aftrs_inventory_listing_generate sku=HW-002 platform=hardwareswap

# Post title format: [USA-XX] [H] Samsung 990 Pro 2TB [W] PayPal, Local Cash
# Include SMART data in post body
# Include timestamps photo link
```

### Pricing Strategy
- Price at eBay median × 0.90 (lower fees offset lower price)
- "Shipped" price includes USPS Priority
- "Local" price 10% less
- Bundle discounts for multi-item purchases

### Expected Revenue: $2,000-3,000

## Phase 4: Automation Milestones (Ongoing)

### Repricing (Weekly)
```
# Check all listed items older than 7 days
aftrs_inventory_stale days=7

# Reprice using market data
aftrs_inventory_reprice sku=HW-001 strategy=fast_sale
aftrs_inventory_reprice sku=HW-001 strategy=median
```

### Morning Dashboard
```
aftrs_inventory_dashboard
```
Review: items needing attention, revenue progress, stale items.

### Stale Item Alerts
Items listed >14 days → reduce price 5%
Items listed >21 days → reduce price 10% or switch platforms
Items listed >30 days → consider lot sale

## Phase 5: Post-Liquidation (Week 6+)

### P&L Export
```
aftrs_inventory_sales_log
aftrs_inventory_export format=csv status=sold
```

### Archive
- Export full inventory to JSON backup
- Document lessons learned
- Keep Config tab target for future reference

### Metrics Review
- Total revenue vs $16,225 target
- Net profit vs $10,000 target
- Average margin by platform
- Fee efficiency (which platform had best net/item)

## Pricing Strategy Matrix

| Category | FB Marketplace | eBay | r/hardwareswap |
|----------|---------------|------|----------------|
| GPU >$500 | Retail × 0.75 | eBay median | eBay median × 0.90 |
| GPU <$500 | Retail × 0.65 | eBay median | eBay median × 0.88 |
| NVMe SSD | Retail × 0.70 | eBay median | eBay median × 0.90 |
| RAM DDR5 | Retail × 0.60 | eBay median | eBay median × 0.85 |
| Networking | — | eBay median | eBay median × 0.90 |
| Thunderbolt | — | eBay median | — |
| Peripherals | Retail × 0.50 | — | — |
| Cables | Retail × 0.30 | Lot sale | — |

## Platform Priority Order

1. **FB Marketplace** — zero fees, immediate cash, test prices
2. **eBay** — widest reach, best for niche and high-value
3. **r/hardwareswap** — low fees, enthusiast community
4. **Lot sale** (eBay/local) — last resort for remaining items

## Related Documentation

- [system.md](system.md) — Architecture: 41 tools, Sheets schema (A-V, 22 columns)
- [objectives.md](objectives.md) — $16,225 revenue target, KPIs, success criteria
- [research.md](research.md) — Platform fees, pricing benchmarks
- [current_bugs.md](current_bugs.md) — Known issues affecting execution
- [testing.md](testing.md) — Test coverage for the tools we're using
- [suggestions.md](suggestions.md) — Enhancements to improve the sales workflow
- [INVENTORY-OPS.md](INVENTORY-OPS.md) — Complete tool reference
