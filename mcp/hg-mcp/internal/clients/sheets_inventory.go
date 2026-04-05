// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	defaultSpreadsheetID = "[REDACTED-SHEETS-ID]"
	defaultInventoryTab  = "Sell Inventory"
	defaultHeaderRow     = 4 // 1-indexed row where column headers live (rows above are title/subtitle)
	salesTab             = "Sales"
	configTab            = "Config"
	cacheTTL             = 30 * time.Second
	imageBaseDir         = ".local/share/hg-mcp/inventory/images"
)

// SheetsInventoryClient implements InventoryBackend using Google Sheets.
type SheetsInventoryClient struct {
	service       *sheets.Service
	spreadsheetID string
	inventoryTab  string // tab name (default: "Sell Inventory")
	dataHeaderRow int    // 1-indexed row number for column headers (default: 4)
	testMode      bool   // when true, all writes go to cache only (no Sheets API calls)

	mu        sync.RWMutex
	cache     []InventoryItem
	cacheTime time.Time
	headerMap map[string]int // column name → index
	headerRow []string

	salesCache     []SaleRecord
	salesCacheTime time.Time
}

var (
	sheetsClient     *SheetsInventoryClient
	sheetsClientOnce sync.Once
	sheetsClientErr  error

	// TestOverrideInventoryClient, when non-nil, is returned by GetInventoryClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideInventoryClient *SheetsInventoryClient
)

// GetInventoryClient returns the singleton Sheets inventory client.
func GetInventoryClient() (*SheetsInventoryClient, error) {
	if TestOverrideInventoryClient != nil {
		return TestOverrideInventoryClient, nil
	}
	sheetsClientOnce.Do(func() {
		sheetsClient, sheetsClientErr = NewSheetsInventoryClient()
	})
	return sheetsClient, sheetsClientErr
}

// NewTestInventoryClient creates an in-memory test client pre-populated with items.
// All reads come from cache; all writes update cache only (no Sheets API calls).
func NewTestInventoryClient(items []InventoryItem) *SheetsInventoryClient {
	hm := map[string]int{
		"#": 0, "category": 1, "product": 2, "model / sku": 3,
		"qty": 4, "condition": 5, "purchase date": 6, "amazon asin": 7,
		"msrp": 8, "current retail": 9, "recommended fb price": 10,
		"total est. revenue": 11, "status": 12, "notes": 13,
		"soldprice": 14, "solddate": 15, "soldplatform": 16,
		"location": 17, "smartdata": 18, "purchaseprice": 19,
		"productline": 20, "specs": 21, "onhand": 22,
		"listedprice": 23, "ebaylistingid": 24, "ebayurl": 25,
	}
	return &SheetsInventoryClient{
		testMode:      true,
		spreadsheetID: "test-spreadsheet",
		inventoryTab:  defaultInventoryTab,
		dataHeaderRow: defaultHeaderRow,
		cache:         append([]InventoryItem{}, items...),
		cacheTime:     time.Now(),
		headerMap:     hm,
	}
}

// NewSheetsInventoryClient creates a new Sheets-backed inventory client.
// Auth via Application Default Credentials (ADC):
//  1. GOOGLE_APPLICATION_CREDENTIALS env var
//  2. gcloud auth application-default login
func NewSheetsInventoryClient() (*SheetsInventoryClient, error) {
	ctx := context.Background()

	srv, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create Sheets service (run: gcloud auth application-default login --scopes=https://www.googleapis.com/auth/spreadsheets): %w", err)
	}

	sid := os.Getenv("INVENTORY_SPREADSHEET_ID")
	if sid == "" {
		sid = defaultSpreadsheetID
	}

	tab := os.Getenv("INVENTORY_TAB_NAME")
	if tab == "" {
		tab = defaultInventoryTab
	}
	hdrRow := defaultHeaderRow
	if v := os.Getenv("INVENTORY_HEADER_ROW"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			hdrRow = n
		}
	}

	return &SheetsInventoryClient{
		service:       srv,
		spreadsheetID: sid,
		inventoryTab:  tab,
		dataHeaderRow: hdrRow,
		headerMap:     make(map[string]int),
	}, nil
}

// IsConfigured returns true if the client is ready.
func (c *SheetsInventoryClient) IsConfigured() bool {
	return c != nil && (c.service != nil || c.testMode)
}

// ---------- cache helpers ----------

func (c *SheetsInventoryClient) cacheValid() bool {
	return !c.cacheTime.IsZero() && time.Since(c.cacheTime) < cacheTTL && len(c.cache) > 0
}

func (c *SheetsInventoryClient) invalidateCache() {
	c.cacheTime = time.Time{}
	c.salesCacheTime = time.Time{}
}

func (c *SheetsInventoryClient) refreshCache(ctx context.Context) error {
	if c.testMode {
		// In test mode, cache is pre-populated; just refresh the timestamp.
		c.cacheTime = time.Now()
		return nil
	}

	// Read from the header row onward, skipping title/subtitle rows above it
	readRange := fmt.Sprintf("'%s'!A%d:Z", c.inventoryTab, c.dataHeaderRow)
	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, readRange).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to read Inventory sheet: %w", err)
	}

	if len(resp.Values) < 2 {
		c.cache = nil
		c.cacheTime = time.Now()
		return nil
	}

	// Parse header row
	header := resp.Values[0]
	c.headerRow = make([]string, len(header))
	c.headerMap = make(map[string]int, len(header))
	for i, h := range header {
		name := strings.TrimSpace(fmt.Sprintf("%v", h))
		c.headerRow[i] = name
		c.headerMap[strings.ToLower(name)] = i
	}

	// Parse data rows
	items := make([]InventoryItem, 0, len(resp.Values)-1)
	for _, row := range resp.Values[1:] {
		item := c.rowToItem(row)
		if item.Name == "" {
			continue
		}
		items = append(items, item)
	}

	c.cache = items
	c.cacheTime = time.Now()
	return nil
}

func (c *SheetsInventoryClient) ensureCache(ctx context.Context) error {
	if c.cacheValid() {
		return nil
	}
	return c.refreshCache(ctx)
}

// ---------- row ↔ item marshaling ----------

func (c *SheetsInventoryClient) cellStr(row []interface{}, col string) string {
	idx, ok := c.headerMap[strings.ToLower(col)]
	if !ok || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", row[idx]))
}

func (c *SheetsInventoryClient) cellFloat(row []interface{}, col string) float64 {
	s := c.cellStr(row, col)
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, ",", "")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func (c *SheetsInventoryClient) cellInt(row []interface{}, col string) int {
	s := c.cellStr(row, col)
	v, _ := strconv.Atoi(s)
	return v
}

func (c *SheetsInventoryClient) rowToItem(row []interface{}) InventoryItem {
	rowNum := c.cellInt(row, "#")
	item := InventoryItem{
		RowNum:        rowNum,
		Category:      c.cellStr(row, "Category"),
		Name:          c.cellStr(row, "Product"),
		Model:         c.cellStr(row, "Model / SKU"),
		Quantity:      c.cellInt(row, "Qty"),
		Condition:     c.cellStr(row, "Condition"),
		PurchaseDate:  c.cellStr(row, "Purchase Date"),
		ASIN:          c.cellStr(row, "Amazon ASIN"),
		MSRP:          c.cellFloat(row, "MSRP"),
		CurrentRetail: c.cellFloat(row, "Current Retail"),
		AskingPrice:   c.cellFloat(row, "Recommended FB Price"),
		TotalRevenue:  c.cellFloat(row, "Total Est. Revenue"),
		ListingStatus: c.cellStr(row, "Status"),
		Notes:         c.cellStr(row, "Notes"),

		// Extended columns
		SoldPrice:     c.cellFloat(row, "SoldPrice"),
		SoldDate:      c.cellStr(row, "SoldDate"),
		SoldPlatform:  c.cellStr(row, "SoldPlatform"),
		Location:      c.cellStr(row, "Location"),
		SmartData:     c.cellStr(row, "SmartData"),
		PurchasePrice: c.cellFloat(row, "PurchasePrice"),
		ProductLine:   c.cellStr(row, "ProductLine"),

		// eBay listing tracking columns (X-Z)
		ListedPrice:   c.cellFloat(row, "ListedPrice"),
		EbayListingID: c.cellStr(row, "EbayListingID"),
		EbayURL:       c.cellStr(row, "EbayURL"),
	}

	// Parse OnHand boolean
	if oh := c.cellStr(row, "OnHand"); strings.EqualFold(oh, "true") {
		item.OnHand = true
	}

	// Parse Specs from JSON string
	if specsJSON := c.cellStr(row, "Specs"); specsJSON != "" {
		var specs map[string]string
		if json.Unmarshal([]byte(specsJSON), &specs) == nil {
			item.Specs = specs
		}
	}

	if item.Quantity == 0 {
		item.Quantity = 1
	}
	if item.ListingStatus == "" {
		item.ListingStatus = "Not Listed"
	}

	// Generate SKU from row number
	if rowNum > 0 {
		item.SKU = fmt.Sprintf("HW-%03d", rowNum)
	}

	// Populate CurrentValue from asking price for compatibility
	item.CurrentValue = item.AskingPrice

	return item
}

func (c *SheetsInventoryClient) itemToRow(item InventoryItem) []interface{} {
	// Build a row matching all known columns
	totalRev := item.AskingPrice * float64(item.Quantity)
	if item.TotalRevenue > 0 {
		totalRev = item.TotalRevenue
	}

	row := make([]interface{}, 26) // A through Z
	row[0] = item.RowNum
	row[1] = item.Category
	row[2] = item.Name
	row[3] = item.Model
	row[4] = item.Quantity
	row[5] = item.Condition
	row[6] = item.PurchaseDate
	row[7] = item.ASIN
	row[8] = formatPrice(item.MSRP)
	row[9] = formatPrice(item.CurrentRetail)
	row[10] = formatPrice(item.AskingPrice)
	row[11] = formatPrice(totalRev)
	row[12] = item.ListingStatus
	row[13] = item.Notes
	row[14] = formatPriceOrEmpty(item.SoldPrice)
	row[15] = item.SoldDate
	row[16] = item.SoldPlatform
	row[17] = item.Location
	row[18] = item.SmartData
	row[19] = formatPriceOrEmpty(item.PurchasePrice)
	row[20] = item.ProductLine
	specsJSON := ""
	if len(item.Specs) > 0 {
		if b, err := json.Marshal(item.Specs); err == nil {
			specsJSON = string(b)
		}
	}
	row[21] = specsJSON
	if item.OnHand {
		row[22] = "TRUE"
	} else {
		row[22] = "FALSE"
	}
	row[23] = formatPriceOrEmpty(item.ListedPrice)
	row[24] = item.EbayListingID
	row[25] = item.EbayURL

	return row
}

func formatPrice(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("$%.0f", v)
}

func formatPriceOrEmpty(v float64) interface{} {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("$%.2f", v)
}

// ---------- CRUD ----------

func (c *SheetsInventoryClient) ListItems(ctx context.Context, filter *InventoryFilter) ([]InventoryItem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCache(ctx); err != nil {
		return nil, err
	}

	return applyFilters(c.cache, filter), nil
}

func (c *SheetsInventoryClient) GetItem(ctx context.Context, sku string) (*InventoryItem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCache(ctx); err != nil {
		return nil, err
	}

	for _, item := range c.cache {
		if item.SKU == sku {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("item not found: %s", sku)
}

func (c *SheetsInventoryClient) AddItem(ctx context.Context, item *InventoryItem) (*InventoryItem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCache(ctx); err != nil {
		return nil, err
	}

	// Determine next row number
	maxRow := 0
	for _, existing := range c.cache {
		if existing.RowNum > maxRow {
			maxRow = existing.RowNum
		}
	}
	item.RowNum = maxRow + 1
	item.SKU = fmt.Sprintf("HW-%03d", item.RowNum)

	if item.Quantity == 0 {
		item.Quantity = 1
	}
	if item.ListingStatus == "" {
		item.ListingStatus = "Not Listed"
	}
	if item.Condition == "" {
		item.Condition = "Used"
	}

	if !c.testMode {
		// Ensure new columns exist in header
		if err := c.ensureExtendedHeaders(ctx); err != nil {
			return nil, fmt.Errorf("failed to extend headers: %w", err)
		}

		row := c.itemToRow(*item)
		vr := &sheets.ValueRange{Values: [][]interface{}{row}}

		appendRange := fmt.Sprintf("'%s'!A%d", c.inventoryTab, c.dataHeaderRow)
		_, err := c.service.Spreadsheets.Values.Append(c.spreadsheetID, appendRange, vr).
			ValueInputOption("USER_ENTERED").Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to append row: %w", err)
		}
	}

	c.cache = append(c.cache, *item)
	// Mark cache as expired so the next operation refreshes from Sheets,
	// ensuring row numbers and data stay consistent after appends.
	c.invalidateCache()

	return item, nil
}

func (c *SheetsInventoryClient) UpdateItem(ctx context.Context, sku string, updates map[string]interface{}) (*InventoryItem, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCache(ctx); err != nil {
		return nil, err
	}

	idx := -1
	for i, item := range c.cache {
		if item.SKU == sku {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("item not found: %s", sku)
	}

	item := &c.cache[idx]

	// Apply updates
	for key, value := range updates {
		switch key {
		case "name":
			item.Name = value.(string)
		case "category":
			item.Category = value.(string)
		case "model":
			item.Model = value.(string)
		case "condition":
			item.Condition = fmt.Sprintf("%v", value)
		case "notes":
			item.Notes = value.(string)
		case "location":
			item.Location = value.(string)
		case "quantity":
			item.Quantity = toInt(value)
		case "listing_status":
			item.ListingStatus = value.(string)
		case "asking_price":
			item.AskingPrice = toFloat(value)
		case "purchase_price":
			item.PurchasePrice = toFloat(value)
		case "current_value":
			item.CurrentValue = toFloat(value)
			item.AskingPrice = toFloat(value)
		case "sold_price":
			item.SoldPrice = toFloat(value)
		case "sold_platform":
			item.SoldPlatform = value.(string)
		case "sold_date":
			item.SoldDate = value.(string)
		case "smart_data":
			item.SmartData = fmt.Sprintf("%v", value)
		case "asin":
			item.ASIN = value.(string)
		case "primary_image":
			item.PrimaryImage = value.(string)
		case "listed_price":
			item.ListedPrice = toFloat(value)
		case "ebay_listing_id":
			item.EbayListingID = value.(string)
		case "ebay_url":
			item.EbayURL = value.(string)
		case "description":
			item.Description = value.(string)
		case "subcategory":
			item.Subcategory = value.(string)
		case "brand":
			item.Brand = value.(string)
		case "tags":
			if tags, ok := value.([]string); ok {
				item.Tags = tags
			}
		case "product_line":
			item.ProductLine = value.(string)
		case "specs":
			switch v := value.(type) {
			case map[string]string:
				item.Specs = v
			case map[string]interface{}:
				m := make(map[string]string, len(v))
				for k, val := range v {
					m[k] = fmt.Sprintf("%v", val)
				}
				item.Specs = m
			}
		case "on_hand":
			switch v := value.(type) {
			case bool:
				item.OnHand = v
			default:
				item.OnHand = fmt.Sprintf("%v", v) == "true"
			}
		}
	}

	if !c.testMode {
		// Write back to sheet: RowNum is 1-based item index, header is at dataHeaderRow
		sheetRow := item.RowNum + c.dataHeaderRow
		rangeStr := fmt.Sprintf("'%s'!A%d:Z%d", c.inventoryTab, sheetRow, sheetRow)
		row := c.itemToRow(*item)
		vr := &sheets.ValueRange{Values: [][]interface{}{row}}

		_, err := c.service.Spreadsheets.Values.Update(c.spreadsheetID, rangeStr, vr).
			ValueInputOption("USER_ENTERED").Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to update row %d: %w", sheetRow, err)
		}
	}

	return item, nil
}

func (c *SheetsInventoryClient) DeleteItem(ctx context.Context, sku string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureCache(ctx); err != nil {
		return err
	}

	idx := -1
	for i, item := range c.cache {
		if item.SKU == sku {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("item not found: %s", sku)
	}

	item := c.cache[idx]

	if !c.testMode {
		sheetRow := item.RowNum + c.dataHeaderRow

		// Clear the row (don't delete to preserve row numbering)
		rangeStr := fmt.Sprintf("'%s'!A%d:Z%d", c.inventoryTab, sheetRow, sheetRow)
		emptyRow := make([]interface{}, 23)
		vr := &sheets.ValueRange{Values: [][]interface{}{emptyRow}}

		_, err := c.service.Spreadsheets.Values.Update(c.spreadsheetID, rangeStr, vr).
			ValueInputOption("USER_ENTERED").Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to clear row %d: %w", sheetRow, err)
		}
	}

	c.cache = append(c.cache[:idx], c.cache[idx+1:]...)
	return nil
}

func (c *SheetsInventoryClient) SearchItems(ctx context.Context, query string, limit int) ([]InventoryItem, error) {
	return c.ListItems(ctx, &InventoryFilter{
		Query: query,
		Limit: limit,
	})
}

// ---------- Analytics ----------

func (c *SheetsInventoryClient) GetSummary(ctx context.Context) (*InventorySummary, error) {
	items, err := c.ListItems(ctx, &InventoryFilter{})
	if err != nil {
		return nil, err
	}

	summary := &InventorySummary{
		TotalItems:  len(items),
		ByCategory:  make(map[string]int),
		ByStatus:    make(map[string]int),
		ByCondition: make(map[string]int),
		ByLocation:  make(map[string]int),
	}

	for _, item := range items {
		summary.TotalValue += item.AskingPrice * float64(item.Quantity)
		summary.TotalCost += item.PurchasePrice * float64(item.Quantity)
		summary.ByCategory[item.Category]++
		summary.ByStatus[item.ListingStatus]++
		summary.ByCondition[item.Condition]++
		if item.Location != "" {
			summary.ByLocation[item.Location]++
		}
	}

	// Recently added (last 10)
	if len(items) > 10 {
		summary.RecentlyAdded = items[len(items)-10:]
	} else {
		summary.RecentlyAdded = items
	}

	// Top value (top 10 by asking price)
	topValue := make([]InventoryItem, len(items))
	copy(topValue, items)
	for i := 0; i < len(topValue)-1 && i < 10; i++ {
		for j := i + 1; j < len(topValue); j++ {
			if topValue[j].AskingPrice > topValue[i].AskingPrice {
				topValue[i], topValue[j] = topValue[j], topValue[i]
			}
		}
	}
	if len(topValue) > 10 {
		topValue = topValue[:10]
	}
	summary.TopValueItems = topValue

	return summary, nil
}

func (c *SheetsInventoryClient) GetCategories(ctx context.Context) ([]InventoryCategory, error) {
	items, err := c.ListItems(ctx, &InventoryFilter{})
	if err != nil {
		return nil, err
	}

	catMap := make(map[string]int)
	for _, item := range items {
		catMap[item.Category]++
	}

	defaults := DefaultCategories()
	result := make([]InventoryCategory, 0, len(defaults))
	for _, def := range defaults {
		cat := InventoryCategory{
			Name:          def.Name,
			Subcategories: def.Subcategories,
			ItemCount:     catMap[def.Name],
		}
		result = append(result, cat)
	}

	// Add any categories from the sheet not in defaults
	for cat, count := range catMap {
		found := false
		for _, def := range defaults {
			if def.Name == cat {
				found = true
				break
			}
		}
		if !found {
			result = append(result, InventoryCategory{Name: cat, ItemCount: count})
		}
	}

	return result, nil
}

func (c *SheetsInventoryClient) GetLocations(ctx context.Context) (map[string]int, error) {
	items, err := c.ListItems(ctx, &InventoryFilter{})
	if err != nil {
		return nil, err
	}

	locations := make(map[string]int)
	for _, item := range items {
		if item.Location != "" {
			locations[item.Location]++
		}
	}
	return locations, nil
}

func (c *SheetsInventoryClient) GetStaleItems(ctx context.Context, days int) ([]InventoryItem, error) {
	items, err := c.ListItems(ctx, &InventoryFilter{})
	if err != nil {
		return nil, err
	}

	var stale []InventoryItem
	for _, item := range items {
		if strings.EqualFold(item.ListingStatus, "Not Listed") || item.ListingStatus == string(StatusNotListed) {
			stale = append(stale, item)
		}
	}
	return stale, nil
}

func (c *SheetsInventoryClient) MarkAsSold(ctx context.Context, sku string, soldPrice float64, platform string) (*InventoryItem, error) {
	return c.UpdateItem(ctx, sku, map[string]interface{}{
		"listing_status": "Sold",
		"sold_price":     soldPrice,
		"sold_platform":  platform,
		"sold_date":      time.Now().Format("Jan 2006"),
	})
}

// ---------- Import (CSV) ----------

func (c *SheetsInventoryClient) ImportAmazonCSV(ctx context.Context, csvData io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(csvData)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
		colIndex[strings.ToLower(col)] = i
	}

	result := &ImportResult{Source: SourceAmazon}
	priceRegex := regexp.MustCompile(`[\d,]+\.?\d*`)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("CSV read error: %v", err))
			continue
		}
		result.TotalRows++

		title := getCSVColumn(record, colIndex, "Title", "items", "Product Name", "product", "item", "description")
		if title == "" {
			result.Skipped++
			continue
		}

		priceStr := getCSVColumn(record, colIndex, "Item Total", "total", "Total Owed", "Unit Price", "price")
		var purchasePrice float64
		if matches := priceRegex.FindString(priceStr); matches != "" {
			purchasePrice, _ = strconv.ParseFloat(strings.ReplaceAll(matches, ",", ""), 64)
		}
		if purchasePrice <= 0 {
			result.Skipped++
			continue
		}

		asin := getCSVColumn(record, colIndex, "ASIN/ISBN", "ASIN", "asin")
		qtyStr := getCSVColumn(record, colIndex, "Quantity", "quantity", "qty")
		qty := 1
		if q, err := strconv.Atoi(qtyStr); err == nil && q > 0 {
			qty = q
		}

		item := &InventoryItem{
			Name:          title,
			Category:      guessCategoryFromName(title),
			PurchasePrice: purchasePrice,
			PurchaseDate:  getCSVColumn(record, colIndex, "Order Date", "date", "order date"),
			ASIN:          asin,
			Quantity:      qty,
			Condition:     "New",
			ListingStatus: "Not Listed",
			AskingPrice:   purchasePrice * 0.7,
		}

		added, err := c.AddItem(ctx, item)
		if err != nil {
			result.Errors++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("Failed to add '%s': %v", title, err))
			continue
		}
		result.Imported++
		result.NewItems = append(result.NewItems, added.SKU)
	}

	return result, nil
}

func (c *SheetsInventoryClient) ImportNeweggCSV(ctx context.Context, csvData io.Reader) (*ImportResult, error) {
	reader := csv.NewReader(csvData)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	result := &ImportResult{Source: SourceNewegg}
	priceRegex := regexp.MustCompile(`[\d.]+`)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors++
			continue
		}
		result.TotalRows++

		itemName := ""
		if idx, ok := colIndex["Item Name"]; ok && idx < len(record) {
			itemName = record[idx]
		}
		if itemName == "" {
			result.Skipped++
			continue
		}

		var purchasePrice float64
		if idx, ok := colIndex["Unit Price"]; ok && idx < len(record) {
			if m := priceRegex.FindString(record[idx]); m != "" {
				purchasePrice, _ = strconv.ParseFloat(m, 64)
			}
		}

		item := &InventoryItem{
			Name:          itemName,
			Category:      guessCategoryFromName(itemName),
			PurchasePrice: purchasePrice,
			Quantity:      1,
			Condition:     "New",
			ListingStatus: "Not Listed",
			AskingPrice:   purchasePrice * 0.7,
		}

		added, err := c.AddItem(ctx, item)
		if err != nil {
			result.Errors++
			continue
		}
		result.Imported++
		result.NewItems = append(result.NewItems, added.SKU)
	}

	return result, nil
}

// ---------- Data Quality ----------

func (c *SheetsInventoryClient) RecategorizeItem(item *InventoryItem) *CategorizationResult {
	newCat := guessCategoryFromName(item.Name)
	return &CategorizationResult{
		SKU:         item.SKU,
		Name:        item.Name,
		OldCategory: item.Category,
		NewCategory: newCat,
		Changed:     newCat != item.Category,
	}
}

func (c *SheetsInventoryClient) RecategorizeItems(ctx context.Context, filter *InventoryFilter, applyChanges bool) ([]CategorizationResult, error) {
	items, err := c.ListItems(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []CategorizationResult
	for _, item := range items {
		r := c.RecategorizeItem(&item)
		if r.Changed {
			results = append(results, *r)
			if applyChanges {
				if _, err := c.UpdateItem(ctx, item.SKU, map[string]interface{}{
					"category": r.NewCategory,
				}); err != nil {
					return nil, fmt.Errorf("failed to update %s: %w", item.SKU, err)
				}
			}
		}
	}
	return results, nil
}

func (c *SheetsInventoryClient) FindDuplicates(ctx context.Context, matchType string) ([]InventoryDuplicateGroup, error) {
	items, err := c.ListItems(ctx, &InventoryFilter{})
	if err != nil {
		return nil, err
	}

	var groups []InventoryDuplicateGroup

	if matchType == "asin" || matchType == "all" {
		asinMap := make(map[string][]InventoryItem)
		for _, item := range items {
			if item.ASIN != "" {
				asinMap[item.ASIN] = append(asinMap[item.ASIN], item)
			}
		}
		for asin, list := range asinMap {
			if len(list) > 1 {
				groups = append(groups, InventoryDuplicateGroup{
					MatchType: "exact_asin", MatchValue: asin,
					Items: list, Count: len(list),
				})
			}
		}
	}

	if matchType == "name" || matchType == "all" {
		nameMap := make(map[string][]InventoryItem)
		for _, item := range items {
			key := strings.ToLower(strings.TrimSpace(item.Name))
			nameMap[key] = append(nameMap[key], item)
		}
		for name, list := range nameMap {
			if len(list) > 1 {
				groups = append(groups, InventoryDuplicateGroup{
					MatchType: "exact_name", MatchValue: name,
					Items: list, Count: len(list),
				})
			}
		}
	}

	if matchType == "order_id" || matchType == "all" {
		// Group by order_id + item_name combination to catch re-imports
		type orderKey struct{ orderID, name string }
		orderMap := make(map[orderKey][]InventoryItem)
		for _, item := range items {
			if item.OrderID != "" {
				k := orderKey{item.OrderID, strings.ToLower(strings.TrimSpace(item.Name))}
				orderMap[k] = append(orderMap[k], item)
			}
		}
		for k, list := range orderMap {
			if len(list) > 1 {
				groups = append(groups, InventoryDuplicateGroup{
					MatchType:  "order_id_name",
					MatchValue: k.orderID + " / " + k.name,
					Items:      list,
					Count:      len(list),
				})
			}
		}
	}

	return groups, nil
}

// ---------- Images (local filesystem) ----------

func (c *SheetsInventoryClient) imageDir(sku string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, imageBaseDir, sku)
}

func (c *SheetsInventoryClient) UploadImage(ctx context.Context, sku string, imageData []byte, filename string) (string, error) {
	dir := c.imageDir(sku)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create image directory: %w", err)
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, imageData, 0o644); err != nil {
		return "", fmt.Errorf("failed to write image: %w", err)
	}

	return path, nil
}

func (c *SheetsInventoryClient) ListImages(ctx context.Context, sku string) ([]string, error) {
	dir := c.imageDir(sku)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	var images []string
	for _, e := range entries {
		if !e.IsDir() {
			images = append(images, e.Name())
		}
	}
	return images, nil
}

func (c *SheetsInventoryClient) DeleteImage(ctx context.Context, sku, filename string) error {
	path := filepath.Join(c.imageDir(sku), filename)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}

// ---------- Config ----------

func (c *SheetsInventoryClient) GetConfig(ctx context.Context) (*InventoryConfig, error) {
	if c.testMode {
		return &InventoryConfig{TargetRevenue: 15555}, nil
	}

	if err := c.ensureConfigTab(ctx); err != nil {
		return nil, err
	}

	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, configTab).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read Config tab: %w", err)
	}

	cfg := &InventoryConfig{TargetRevenue: 15555}
	for _, row := range resp.Values {
		if len(row) < 2 {
			continue
		}
		key := strings.TrimSpace(fmt.Sprintf("%v", row[0]))
		val := strings.TrimSpace(fmt.Sprintf("%v", row[1]))
		switch key {
		case "target_revenue":
			cfg.TargetRevenue, _ = strconv.ParseFloat(val, 64)
		case "created":
			cfg.Created = val
		case "last_updated":
			cfg.LastUpdated = val
		}
	}
	return cfg, nil
}

func (c *SheetsInventoryClient) UpdateConfig(ctx context.Context, key, value string) error {
	if err := c.ensureConfigTab(ctx); err != nil {
		return err
	}

	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, configTab).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to read Config tab: %w", err)
	}

	// Find existing key or append
	rowIdx := -1
	for i, row := range resp.Values {
		if len(row) > 0 && strings.TrimSpace(fmt.Sprintf("%v", row[0])) == key {
			rowIdx = i + 1
			break
		}
	}

	if rowIdx > 0 {
		rangeStr := fmt.Sprintf("%s!A%d:B%d", configTab, rowIdx, rowIdx)
		vr := &sheets.ValueRange{Values: [][]interface{}{{key, value}}}
		_, err = c.service.Spreadsheets.Values.Update(c.spreadsheetID, rangeStr, vr).
			ValueInputOption("USER_ENTERED").Context(ctx).Do()
	} else {
		vr := &sheets.ValueRange{Values: [][]interface{}{{key, value}}}
		_, err = c.service.Spreadsheets.Values.Append(c.spreadsheetID, configTab, vr).
			ValueInputOption("USER_ENTERED").Context(ctx).Do()
	}

	if err != nil {
		return fmt.Errorf("failed to update config key %q: %w", key, err)
	}

	// Update last_updated
	if key != "last_updated" {
		_ = c.UpdateConfig(ctx, "last_updated", time.Now().Format(time.RFC3339))
	}

	return nil
}

// ---------- Tab helpers ----------

func (c *SheetsInventoryClient) ensureExtendedHeaders(ctx context.Context) error {
	// Check if we already have the extended headers
	if _, ok := c.headerMap["soldprice"]; ok {
		return nil
	}

	// Read current header from the actual header row
	hdrRange := fmt.Sprintf("'%s'!%d:%d", c.inventoryTab, c.dataHeaderRow, c.dataHeaderRow)
	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, hdrRange).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return err
	}

	currentCols := 0
	if len(resp.Values) > 0 {
		currentCols = len(resp.Values[0])
	}

	// Need at least 26 columns (A-Z: original 14 + 9 extended + 3 eBay listing)
	if currentCols >= 26 {
		return nil
	}

	// Append new headers starting at the next column
	newHeaders := []string{"SoldPrice", "SoldDate", "SoldPlatform", "Location", "SmartData", "PurchasePrice", "ProductLine", "Specs", "OnHand", "ListedPrice", "EbayListingID", "EbayURL"}
	startCol := currentCols
	endCol := startCol + len(newHeaders) - 1

	rangeStr := fmt.Sprintf("'%s'!%s%d:%s%d", c.inventoryTab, colLetter(startCol), c.dataHeaderRow, colLetter(endCol), c.dataHeaderRow)
	row := make([]interface{}, len(newHeaders))
	for i, h := range newHeaders {
		row[i] = h
	}
	vr := &sheets.ValueRange{Values: [][]interface{}{row}}

	_, err = c.service.Spreadsheets.Values.Update(c.spreadsheetID, rangeStr, vr).
		ValueInputOption("USER_ENTERED").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to add extended headers: %w", err)
	}

	// Update header map
	for i, h := range newHeaders {
		c.headerMap[strings.ToLower(h)] = startCol + i
	}

	return nil
}

func (c *SheetsInventoryClient) ensureConfigTab(ctx context.Context) error {
	// Try reading — if it fails, create the tab
	_, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, configTab+"!A1").
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err == nil {
		return nil
	}

	// Create Config tab
	addReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{Title: configTab},
			}},
		},
	}
	if _, err := c.service.Spreadsheets.BatchUpdate(c.spreadsheetID, addReq).Context(ctx).Do(); err != nil {
		// Tab might already exist
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create Config tab: %w", err)
		}
	}

	// Seed with defaults
	vr := &sheets.ValueRange{Values: [][]interface{}{
		{"target_revenue", "15555"},
		{"created", time.Now().Format("2006-01-02")},
		{"last_updated", time.Now().Format(time.RFC3339)},
	}}
	_, err = c.service.Spreadsheets.Values.Append(c.spreadsheetID, configTab, vr).
		ValueInputOption("USER_ENTERED").Context(ctx).Do()
	return err
}

// ---------- Filtering ----------

func applyFilters(items []InventoryItem, filter *InventoryFilter) []InventoryItem {
	if filter == nil {
		return items
	}

	var filtered []InventoryItem
	for _, item := range items {
		if filter.Category != "" && !strings.EqualFold(item.Category, filter.Category) {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(item.ListingStatus, string(filter.Status)) {
			continue
		}
		if filter.Condition != "" && !strings.EqualFold(item.Condition, string(filter.Condition)) {
			continue
		}
		if filter.Location != "" && !strings.EqualFold(item.Location, filter.Location) {
			continue
		}
		if filter.Brand != "" && !strings.EqualFold(item.Brand, filter.Brand) {
			continue
		}
		if filter.MinPrice > 0 && item.AskingPrice < filter.MinPrice {
			continue
		}
		if filter.MaxPrice > 0 && item.AskingPrice > filter.MaxPrice {
			continue
		}
		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			if !strings.Contains(strings.ToLower(item.Name), q) &&
				!strings.Contains(strings.ToLower(item.Category), q) &&
				!strings.Contains(strings.ToLower(item.Model), q) &&
				!strings.Contains(strings.ToLower(item.Notes), q) &&
				!strings.Contains(strings.ToLower(item.ProductLine), q) {
				continue
			}
		}
		filtered = append(filtered, item)
	}

	if filter.Offset > 0 && filter.Offset < len(filtered) {
		filtered = filtered[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(filtered) {
		filtered = filtered[:filter.Limit]
	}

	return filtered
}

// ---------- Utility ----------

func colLetter(idx int) string {
	if idx < 26 {
		return string(rune('A' + idx))
	}
	return string(rune('A'+idx/26-1)) + string(rune('A'+idx%26))
}

// getCSVColumn retrieves a column value from a CSV record using multiple possible column names.
func getCSVColumn(record []string, colIndex map[string]int, names ...string) string {
	for _, name := range names {
		if idx, ok := colIndex[name]; ok && idx < len(record) {
			return record[idx]
		}
		if idx, ok := colIndex[strings.ToLower(name)]; ok && idx < len(record) {
			return record[idx]
		}
	}
	return ""
}

func guessCategoryFromName(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "rtx") || strings.Contains(n, "gtx") || strings.Contains(n, "radeon") || strings.Contains(n, "gpu") || strings.Contains(n, "graphics"):
		return "GPU"
	case strings.Contains(n, "ryzen") || strings.Contains(n, "core i") || strings.Contains(n, "cpu") || strings.Contains(n, "processor"):
		return "CPU"
	case strings.Contains(n, "ddr5") || strings.Contains(n, "ddr4") || strings.Contains(n, "ram") || strings.Contains(n, "memory"):
		return "RAM"
	case strings.Contains(n, "nvme") || strings.Contains(n, "ssd") || strings.Contains(n, "hdd") || strings.Contains(n, "drive"):
		return "Storage"
	case strings.Contains(n, "motherboard") || strings.Contains(n, "mainboard"):
		return "Motherboard"
	case strings.Contains(n, "power supply") || strings.Contains(n, "psu"):
		return "PSU"
	case strings.Contains(n, "case") || strings.Contains(n, "chassis"):
		return "Case"
	case strings.Contains(n, "cooler") || strings.Contains(n, "fan") || strings.Contains(n, "thermal"):
		return "Cooling"
	case strings.Contains(n, "keyboard") || strings.Contains(n, "mouse") || strings.Contains(n, "monitor"):
		return "Peripherals"
	case strings.Contains(n, "router") || strings.Contains(n, "switch") || strings.Contains(n, "ethernet"):
		return "Networking"
	case strings.Contains(n, "thunderbolt"):
		return "Thunderbolt"
	case strings.Contains(n, "cable") || strings.Contains(n, "adapter"):
		return "Cables"
	default:
		return "Other"
	}
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(val, "$", ""), ",", ""), 64)
		return f
	default:
		return 0
	}
}

// ImportGmailOrders is implemented in sheets_inventory_gmail.go

// ExportItems exports all items matching the filter.
func (c *SheetsInventoryClient) ExportItems(ctx context.Context, filter *InventoryFilter) ([]InventoryItem, error) {
	return c.ListItems(ctx, filter)
}

// GetPresignedUploadURL is a no-op for local filesystem storage.
func (c *SheetsInventoryClient) GetPresignedUploadURL(ctx context.Context, sku, filename string, expiresIn time.Duration) (string, error) {
	return "", fmt.Errorf("presigned URLs not supported — use local file upload instead")
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}
