# hw-resale Known Issues and Limitations

> Tracked bugs, data quality issues, feature gaps, and documentation debt.

## Bugs

### BUG-001: Gmail eBay Import Broken — FIXED

**Severity:** High — blocks eBay order history import
**File:** `internal/clients/sheets_inventory_gmail.go`
**Symptom:** eBay sends HTML emails. The text parser extracted `<table>` tags as item names and misparsed prices (e.g., $6548378.00 from item numbers).
**Fix:** Replaced regex-based HTML parsing with `golang.org/x/net/html` tokenizer in `parseEbayHTMLTable`. Added price sanity check ($0.01–$50,000). Handles nested tables and tags correctly.
**Status:** Fixed — regex fallback in `parseEbayEmail` also now has price bounds check.

### BUG-002: Duplicate Detection Insufficient — FIXED

**Severity:** Medium — creates duplicate items on Gmail re-import
**Status:** Fixed — `FindDuplicates()` now supports `order_id` match type using `order_id + item_name` combination.
**Test:** `TestHandleDuplicates_OrderID`

## Limitations

### LIMIT-001: Google Sheets API Rate Limits

**Impact:** Low (well-mitigated)
**Limit:** ~100 requests per 100 seconds per user per project.
**Mitigation:** 30-second cache TTL means most reads hit cache. Writes invalidate cache but are infrequent in normal use. Bulk operations (batch_listing, bulk_update) use single sheet reads.
**Risk:** Running multiple Claude Code sessions simultaneously against the same spreadsheet could hit limits.

### LIMIT-002: eBay Browse API Daily Limits

**Impact:** Medium — affects price_research and active_listings tools
**Limit:** 5,000 calls/day on Browse API (varies by app tier).
**Mitigation:** Not currently rate-limited at the application level (the global `pkg/ratelimit` module handles this if the eBay module is tagged with circuit breaker groups).

### LIMIT-003: eBay OAuth Refresh Token Expiration

**Impact:** High when it happens — blocks all eBay Sell API operations
**Symptom:** `ebay_create_listing` returns 401 after token expiry.
**Fix:** Manual re-auth flow required. No automatic token refresh implemented.

## Data Quality Issues

### DATA-001: Condition/ListingStatus Not Validated — FIXED

**Impact:** Medium — typos create orphan values in analytics
**Status:** Fixed — `validateCondition()` and `validateListingStatus()` now enforce enums and normalize to lowercase in `handleInventoryAdd`, `handleInventoryUpdate`, and `handleInventoryBulkUpdate`.
**Tests:** `TestHandleInventoryAdd_InvalidCondition`, `TestHandleInventoryUpdate_InvalidStatus`, `TestHandleBulkUpdate_InvalidStatus`

### DATA-002: PurchaseDate Free-Text

**Impact:** Low — prevents date-based queries/sorting
**Field:** `PurchaseDate` (col G) stored as free-text (e.g., "Apr 2025", "2024-12-15", "last month").
**Effect:** `stale` tool and dashboard age calculations can't reliably parse dates.
**Fix:** Normalize to ISO 8601 on write, with display formatting on read.

### DATA-003: SKU Collision Risk

**Impact:** Low — theoretical
**Mechanism:** SKUs are `HW-{RowNum}` format. Deleting row 5 then adding a new item could reuse `HW-005` if the row number is recycled.
**Current behavior:** Row numbers are assigned sequentially from max existing RowNum, so collisions only occur if items are deleted and re-added in specific sequences.
**Fix:** Use UUID-based or monotonically increasing SKU generation.

### DATA-004: Specs Column Sparsely Populated — MITIGATED

**Impact:** Medium — listing templates fall back to generic ASIN/MSRP metadata
**Status:** Mitigated — `specs_catalog.go` provides `LookupSpecs()` with defaults for 30+ product lines. Auto-populates on `handleInventoryAdd` and `handleImportJSON` when ProductLine matches.
**Remaining:** Items without a ProductLine still need manual Specs population via `handleInventoryUpdate`.

## Feature Gaps

### FEAT-001: No Automated Price Tracking

**Impact:** Medium — manual repricing only
**Description:** No scheduled job to re-check eBay prices. The `reprice` tool must be called manually per SKU.
**Desired:** Periodic check that flags items where market price has shifted >10%.

### FEAT-002: No eBay Sold Notifications

**Impact:** Medium — must manually check eBay for sold items
**Description:** No webhook or polling to detect when an eBay listing sells.
**Desired:** Auto-detect eBay sales and prompt `sale_record` creation.

### FEAT-003: No Shipping Cost Estimation

**Impact:** Low — mental math for shipping costs
**Description:** No tool to estimate shipping costs by carrier (USPS, UPS, FedEx) based on weight/dimensions.
**Desired:** `shipping_estimate` tool for accurate P&L planning.

### FEAT-004: No Template Customization

**Impact:** Low — templates are Go constants
**Description:** Listing templates are hardcoded in `templates.go`. Users can't customize wording, add/remove sections, or adjust formatting without code changes.
**Desired:** User-editable templates (config file or sheet tab).

## Documentation Debt

### DOC-001: Tool Count Inconsistencies — FIXED

**Files:** `README.md`, `CLAUDE.md`, various docs
**Issue:** README said "880+ tools across 120 modules" but the server now has 1,190+ tools across 119 modules. Inventory module has 51 tools.
**Status:** Fixed — README and CLAUDE.md updated to reflect actual counts.

### DOC-002: Outdated Architecture.md

**File:** `docs/ARCHITECTURE.md`
**Issue:** May not reflect recent inventory system changes, enhancement sweep, or new packages (pkg/ratelimit, pkg/cache, pkg/httpclient).

### DOC-003: Outdated Roadmap.md

**File:** `docs/Roadmap.md`
**Issue:** Generic project roadmap doesn't include hw-resale-specific milestones.
**See also:** [roadmap.md](roadmap.md) for the hw-resale-specific execution plan.

## Recently Completed

### SEC-001: Shell Injection in data_migration.go — FIXED

**Severity:** Critical
**Files:** `internal/clients/data_migration.go`, `pkg/sanitize/validators.go`
**Issue:** 6 `cmd /c` + `fmt.Sprintf` shell injection vectors in WSL mount/unmount/repair functions. User-controlled device paths, mount points, filesystem types, and decryption keys interpolated into shell strings.
**Fix:** Replaced all `cmd /c` patterns with `wsl -u root` argument arrays. Added DevicePath, FileSystemType, MountPoint validators. LUKS keys now passed via stdin.
**Status:** Fixed — 0 `cmd /c` calls remaining in the codebase.

### SEC-002: eBay Finding API Deprecated — FIXED

**Severity:** High
**File:** `internal/clients/ebay.go`
**Issue:** Price research tools used deprecated Finding API (`svcs.ebay.com`), decommissioned Feb 2025.
**Fix:** Migrated to Browse API (`api.ebay.com/buy/browse/v1/item_summary/search`) with OAuth Bearer auth.

### SEC-003: Unvalidated exec Paths — PARTIALLY FIXED

**Severity:** Medium
**Files:** `internal/mcp/tools/rclone/module.go`, `internal/mcp/tools/samples/ffmpeg.go`, `internal/clients/backup.go`
**Issue:** User-provided file/rclone paths passed to exec without validation.
**Fix:** Added RclonePath and MediaPath validators, applied to 8 rclone handlers, ffprobe/ffmpeg/whisper calls, and backup restore. 34 sanitization calls now in place.
**Remaining:** ~185 exec calls with hardcoded or low-risk arguments (integers, enum values).

### FEAT-005: MCP Tool Annotations — DONE

**Status:** All 1,190+ tools now have MCP annotations (readOnlyHint, destructiveHint, idempotentHint, openWorldHint) auto-inferred from tool name patterns.

### FEAT-006: Streamable HTTP Transport — DONE

**Status:** `MCP_MODE=streamable` supported (MCP 2025-03-26 spec). SSE mode deprecated April 2026.

### FEAT-007: 23 Missing Module Registrations — FIXED

**Status:** All 118 tool modules now registered in main.go (was missing 23).

## Pre-existing Issues (Not Ours)

- `go vet` IPv6 warnings in dashboard/grandma3/ledfx/lighting/obs/resolume clients
- `go vet` lock copy warning in `cmd/beatport-sync/main.go:328`
- These predate the inventory module and hw-resale work

## Related Documentation

- [system.md](system.md) — Full system architecture
- [inventory-import-improvements.md](inventory-import-improvements.md) — Detailed import bug analysis
- [testing.md](testing.md) — Test coverage gaps
- [suggestions.md](suggestions.md) — Recommended fixes
