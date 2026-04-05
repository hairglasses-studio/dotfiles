# hw-resale Research Findings

> Consolidated platform research, pricing benchmarks, and selling strategy.

## Platform Fee Comparison

| Platform | Seller Fee | Payment Fee | Total Fee | Payout Speed |
|----------|-----------|-------------|-----------|-------------|
| eBay | 13.25% (most categories) | Included in seller fee | ~13.25% | 2-3 business days |
| FB Marketplace (local) | 0% | 0% (cash/Zelle) | 0% | Immediate |
| FB Marketplace (shipped) | 5% or $0.40 min | Included | ~5% | 5 business days |
| r/hardwareswap | 0% | PayPal G&S 3.49% + $0.49 | ~3.9% | 1-3 business days |
| Mercari | 10% | Included | 10% | 3-5 business days |
| Swappa | Flat fee $5-$25 | PayPal 3.49% + $0.49 | ~5-8% | Immediate via PayPal |

## Platform Pros/Cons

### eBay
**Pros:**
- Largest buyer pool — fastest sell-through for niche items
- Built-in buyer protection attracts higher-spend buyers
- Price research via completed listings (our `ebay_price_research` tool)
- Structured listings (item specifics, condition, returns) build trust
- Best for items >$100 where fees are worth the exposure

**Cons:**
- 13.25% fee is highest of all platforms
- Returns policy required (14-day minimum recommended)
- Buyer disputes can be costly (eBay tends to favor buyers)
- Requires packaging and shipping infrastructure
- OAuth token management overhead

### FB Marketplace (Local)
**Pros:**
- 0% fees on local cash/Zelle transactions
- No shipping required — saves time and cost
- Instant cash flow
- Good for bulky/heavy items (cases, monitors, UPS)
- Negotiation common but prices are more transparent now

**Cons:**
- Smaller buyer pool than eBay
- No-shows are common (30-40% flake rate)
- No buyer protection — cash only
- Geography-limited (metro area)
- Safety concerns with high-value items

### r/hardwareswap
**Pros:**
- Low fees (~3.9% via PayPal G&S)
- Enthusiast community — buyers know what they're buying
- Reputation system (confirmed trades) builds trust
- Good for GPUs, NVMe, RAM, and niche components
- Timestamps add authenticity (SMART data is a strong differentiator)

**Cons:**
- Smaller audience than eBay
- Requires Reddit account with some karma
- Title format is strict: `[USA-XX] [H] Item [W] PayPal`
- No buyer protection beyond PayPal
- Posts can get buried if not well-timed (weekday mornings best)

## Recommended Platform by Category

| Category | 1st Choice | 2nd Choice | Rationale |
|----------|-----------|------------|-----------|
| GPU (>$500) | eBay | r/hardwareswap | eBay's buyer pool justifies fees for high-value GPUs |
| GPU (<$500) | r/hardwareswap | FB Marketplace | Lower fees matter more at lower price points |
| NVMe/SSD | r/hardwareswap | eBay | Enthusiasts know the value; SMART data helps |
| RAM (DDR5) | r/hardwareswap | eBay | Standard commodity — price-sensitive buyers |
| Networking | eBay | FB Marketplace | Niche items need eBay's search traffic |
| Thunderbolt | eBay | r/hardwareswap | Niche — needs wide exposure |
| Peripherals | FB Marketplace | eBay | Local pickup avoids shipping fragile items |
| Complete Systems | FB Marketplace | eBay | Heavy items benefit from local pickup |
| Cables/Accessories | FB Marketplace | Bundle lot on eBay | Not worth shipping individually |
| KVM | eBay | r/hardwareswap | Niche buyers search eBay specifically |

## Condition Grading Guide

| Condition | Description | Price Impact vs. New |
|-----------|-------------|---------------------|
| New / Sealed | Factory sealed, untouched | 85-95% of retail |
| New (Open Box) | Opened but unused, all accessories | 75-85% of retail |
| Like New | Minimal use (<1 month), no marks | 70-80% of retail |
| Used — Excellent | Light use, no cosmetic issues | 60-75% of retail |
| Used — Good | Normal wear, fully functional | 50-65% of retail |
| Used — Fair | Visible wear, works correctly | 35-50% of retail |
| For Parts | May not work, sold as-is | 15-30% of retail |

### Condition Mapping to eBay

| Our Condition | eBay Condition | eBay Condition ID |
|---------------|---------------|-------------------|
| new | NEW | 1000 |
| like_new | LIKE_NEW | 3000 |
| good | VERY_GOOD | 4000 |
| fair | GOOD | 5000 |
| poor | ACCEPTABLE | 6000 |
| for_parts | FOR_PARTS_OR_NOT_WORKING | 7000 |

## Shipping Cost Reference

| Item Type | Weight (est.) | USPS Priority | UPS Ground | Recommendation |
|-----------|-------------|---------------|------------|----------------|
| GPU | 3-5 lbs | $15-22 | $12-18 | UPS Ground (insurance) |
| NVMe SSD | <1 lb | $8-10 | $8-12 | USPS Priority (fastest) |
| RAM kit | <1 lb | $8-10 | $8-12 | USPS Priority |
| Motherboard | 3-4 lbs | $15-20 | $12-16 | UPS Ground |
| PSU | 5-8 lbs | $18-28 | $14-22 | UPS Ground |
| Router/Switch | 2-4 lbs | $12-18 | $10-15 | USPS Priority |
| Thunderbolt dock | 1-2 lbs | $10-14 | $10-14 | USPS Priority |
| Complete system | 15-40 lbs | N/A | $25-60 | UPS Ground only |
| Cable bundles | 1-2 lbs | $8-12 | $10-14 | USPS Priority |

**Tips:**
- Always include insurance for items >$100
- USPS Priority includes $50 insurance free
- Ship within 2 business days (eBay penalizes late shipping)
- Use original packaging when available (mention in listing)
- Signature confirmation for items >$750 (eBay requirement)

## Market Pricing Benchmarks (as of early 2026)

### GPUs

| Model | New Retail | Used (Good) | r/hardwareswap | eBay Sold Median |
|-------|-----------|-------------|----------------|-----------------|
| RTX 4090 | $1,599 | $1,200-1,500 | $1,100-1,350 | $1,250-1,400 |
| RTX 4080 Super | $999 | $750-850 | $700-800 | $750-850 |
| RTX 4070 Ti Super | $799 | $550-650 | $500-600 | $550-650 |
| RTX 3080 | — (EOL) | $300-400 | $280-350 | $320-380 |

### Storage (NVMe Gen4/5)

| Model | New Retail | Used (Good) | Typical Sold |
|-------|-----------|-------------|-------------|
| Samsung 990 Pro 2TB | $180 | $130-160 | $140-155 |
| Samsung 990 Pro 1TB | $100 | $70-85 | $75-85 |
| WD SN850X 2TB | $160 | $120-145 | $130-145 |
| Crucial T700 2TB (Gen5) | $250 | $180-220 | $190-210 |

### Memory (DDR5)

| Configuration | New Retail | Used | Typical Sold |
|--------------|-----------|------|-------------|
| 32GB (2×16) DDR5-5600 | $80-100 | $55-75 | $60-75 |
| 64GB (2×32) DDR5-5600 | $160-200 | $120-160 | $130-160 |
| 32GB (2×16) DDR5-6000+ | $100-150 | $70-110 | $80-110 |

### Networking

| Type | New Retail | Used | Notes |
|------|-----------|------|-------|
| WiFi 6E Router | $200-350 | $120-220 | Brand matters (Netgear/Asus) |
| 10GbE NIC | $50-80 | $30-50 | Commodity, price-sensitive |
| Managed Switch (8-port) | $100-200 | $60-130 | Enterprise brands hold value |
| Thunderbolt 4 Dock | $200-350 | $130-250 | CalDigit/OWC hold value best |

## eBay Listing Best Practices

1. **Timing:** List Sunday evening or Monday morning (highest search traffic)
2. **Duration:** 30-day fixed price (not auction for used electronics)
3. **Title:** Use all 80 characters — include brand, model, specs, condition
4. **Category:** Use `ebay_categories` tool to find the right category ID
5. **Photos:** Minimum 3, show all sides + any defects + accessories
6. **Item specifics:** Fill in every field eBay offers (improves search ranking)
7. **Shipping:** Free shipping (build cost into price) or calculated shipping
8. **Returns:** 30-day returns (eBay rewards this with higher search placement)
9. **Promoted listings:** 2-3% ad rate for slow-moving items

## r/hardwareswap Best Practices

1. **Title format:** `[USA-XX] [H] Item Name [W] PayPal, Local Cash`
2. **Timestamps:** Required — photo with username + date on paper next to item
3. **SMART data:** Include for storage devices (our `smart_collect` tool generates this)
4. **Pricing:** Research recent posts for same item. Price slightly below to sell fast.
5. **Posting time:** Weekday mornings EST (10am-12pm) for best visibility
6. **Comment before PM:** Always include this — subreddit etiquette
7. **Reputation:** Note confirmed trades count if >5
8. **Bundle discounts:** Offer 5-10% off for multi-item purchases
9. **Shipping:** "Shipped" price includes USPS Priority; "Local" price is lower

## Related Documentation

- [system.md](system.md) — Architecture and tool inventory (41 tools)
- [objectives.md](objectives.md) — Revenue targets using these benchmarks
- [roadmap.md](roadmap.md) — Sales execution timeline
- [current_bugs.md](current_bugs.md) — Known issues affecting platform integrations
- [INVENTORY-OPS.md](INVENTORY-OPS.md) — Tool reference for daily workflow
