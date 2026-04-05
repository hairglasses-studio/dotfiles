# Inventory Import Tool Improvements

## Issues Found During Initial Import (2026-01-07)

### eBay Parser Issues
- eBay sends HTML emails, parser extracts `<table>` tags as item names
- Need to strip HTML or use HTML parser for eBay emails
- Price extraction picking up item numbers (e.g., $6548378.00)

### Parser Improvements Needed
1. **HTML Email Support**: eBay (and possibly others) use HTML emails
   - Strip HTML tags before text parsing
   - Or use `golang.org/x/net/html` parser for structured extraction

2. **Price Sanity Check**: Filter out obviously wrong prices
   - Maximum reasonable price threshold (e.g., $50,000)
   - Minimum price threshold (e.g., $0.01)

3. **Item Name Validation**:
   - Filter out HTML tags in names
   - Minimum name length check
   - Check for common non-product patterns

### Import Statistics (Initial Run)
- Emails scanned: 210
- Successfully parsed: 18
- Skipped (no items/already imported): 203
- Amazon parsing: Working well
- Newegg parsing: Untested (no emails found)
- eBay parsing: Broken (HTML emails)
- BestBuy parsing: Untested (no emails found)

### Duplicate Detection Bug
- Same items imported multiple times with different SKUs
- Current check only looks at order ID, but re-imports occur when:
  - Tool is run multiple times
  - Same email is processed again
- **Fix needed**: Check order_id + item_name combination, not just order_id

### Future Enhancements
1. Add support for more vendors (B&H Photo, Adorama, Micro Center)
2. Extract ASIN/UPC from emails for product matching
3. Link to product URLs from email
4. Extract shipping/tracking info
5. Handle multi-item orders better
6. Add duplicate detection by product name (not just order ID)
