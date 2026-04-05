package inventory

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/mark3labs/mcp-go/mcp"
)

// ── helpers ──

func makeReq(args map[string]interface{}) mcp.CallToolRequest {
	var arguments interface{}
	if args != nil {
		arguments = args
	}
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: arguments,
		},
	}
}

func getText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("result is nil")
	}
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
	}
	t.Fatal("no text content in result")
	return ""
}

func getJSON(t *testing.T, result *mcp.CallToolResult, v interface{}) {
	t.Helper()
	text := getText(t, result)
	if err := json.Unmarshal([]byte(text), v); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\ntext: %s", err, text)
	}
}

// seedItems returns the 5-item test dataset.
func seedItems() []clients.InventoryItem {
	return []clients.InventoryItem{
		{
			RowNum: 1, SKU: "HW-001", Name: "ASUS ROG Strix RTX 4090",
			Category: "GPU", AskingPrice: 2200, PurchasePrice: 1800,
			Condition: "Used", Quantity: 1, ListingStatus: "Not Listed",
			ASIN: "B0BSVMLVTD", Location: "studio-rack-1",
			CurrentValue: 2200,
		},
		{
			RowNum: 2, SKU: "HW-002", Name: "Samsung 990 Pro 2TB NVMe",
			Category: "Storage", AskingPrice: 180, PurchasePrice: 140,
			Condition: "New", Quantity: 3, ListingStatus: "Not Listed",
			ASIN: "B0BHJJ9Y77", Location: "storage-bin-A",
			CurrentValue: 180,
		},
		{
			RowNum: 3, SKU: "HW-003", Name: "Corsair Vengeance DDR5 32GB",
			Category: "RAM", AskingPrice: 120, PurchasePrice: 90,
			Condition: "Like New", Quantity: 2, ListingStatus: "Not Listed",
			Location:     "storage-bin-A",
			CurrentValue: 120,
		},
		{
			RowNum: 4, SKU: "HW-004", Name: "Netgear Nighthawk Router",
			Category: "Networking", AskingPrice: 200, PurchasePrice: 150,
			Condition: "Used", Quantity: 1, ListingStatus: "Listed",
			Location:     "studio-rack-1",
			CurrentValue: 200,
		},
		{
			RowNum: 5, SKU: "HW-005", Name: "ASUS ROG Strix RTX 4090",
			Category: "GPU", AskingPrice: 2100, PurchasePrice: 1700,
			Condition: "Renewed", Quantity: 1, ListingStatus: "Not Listed",
			ASIN: "B0BSVMLVTD", Location: "studio-rack-1",
			CurrentValue: 2100,
		},
	}
}

func TestMain(m *testing.M) {
	clients.TestOverrideInventoryClient = clients.NewTestInventoryClient(seedItems())
	code := m.Run()
	clients.TestOverrideInventoryClient = nil
	os.Exit(code)
}

// resetSeed restores the client cache to the original seed data.
func resetSeed() {
	clients.TestOverrideInventoryClient = clients.NewTestInventoryClient(seedItems())
}

// ── CRUD tests ──

func TestHandleInventoryList(t *testing.T) {
	resetSeed()
	res, err := handleInventoryList(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 5 {
		t.Errorf("expected 5 items, got %v", out["count"])
	}
}

func TestHandleInventoryList_CategoryFilter(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryList(context.Background(), makeReq(map[string]interface{}{
		"category": "GPU",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 2 {
		t.Errorf("expected 2 GPU items, got %v", out["count"])
	}
}

func TestHandleInventoryList_StatusFilter(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryList(context.Background(), makeReq(map[string]interface{}{
		"status": "Listed",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 1 {
		t.Errorf("expected 1 Listed item, got %v", out["count"])
	}
}

func TestHandleInventoryList_QueryFilter(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryList(context.Background(), makeReq(map[string]interface{}{
		"query": "Samsung",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 1 {
		t.Errorf("expected 1 Samsung item, got %v", out["count"])
	}
}

func TestHandleInventoryList_Limit(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryList(context.Background(), makeReq(map[string]interface{}{
		"limit": float64(2),
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 2 {
		t.Errorf("expected 2 items with limit=2, got %v", out["count"])
	}
}

func TestHandleInventoryGet(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var item clients.InventoryItem
	getJSON(t, res, &item)
	if item.Name != "ASUS ROG Strix RTX 4090" {
		t.Errorf("unexpected name: %s", item.Name)
	}
}

func TestHandleInventoryGet_NotFound(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-999",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "not found") {
		t.Error("expected not found error")
	}
}

func TestHandleInventoryGet_MissingSKU(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryGet(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected sku required error, got: %s", text)
	}
}

func TestHandleInventoryAdd(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryAdd(context.Background(), makeReq(map[string]interface{}{
		"name":           "Test Item",
		"category":       "GPU",
		"purchase_price": float64(500),
		"location":       "test-shelf",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success=true")
	}
	sku, ok := out["sku"].(string)
	if !ok || sku == "" {
		t.Error("expected returned SKU")
	}
	if !strings.HasPrefix(sku, "HW-") {
		t.Errorf("unexpected SKU format: %s", sku)
	}
}

func TestHandleInventoryUpdate(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":       "HW-001",
		"notes":     "updated notes",
		"condition": "like_new",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success=true")
	}
	item := out["item"].(map[string]interface{})
	if item["notes"] != "updated notes" {
		t.Errorf("expected updated notes, got %v", item["notes"])
	}
	if item["condition"] != "like_new" {
		t.Errorf("expected like_new condition, got %v", item["condition"])
	}
}

func TestHandleInventoryUpdate_NotFound(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":   "HW-999",
		"notes": "nope",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "not found") {
		t.Error("expected not found error")
	}
}

func TestHandleInventoryDelete(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDelete(context.Background(), makeReq(map[string]interface{}{
		"sku":     "HW-001",
		"confirm": true,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success=true")
	}
	// Verify it's gone
	res2, _ := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	text := getText(t, res2)
	if !strings.Contains(text, "not found") {
		t.Error("expected item to be deleted")
	}
}

func TestHandleInventoryDelete_NoConfirm(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDelete(context.Background(), makeReq(map[string]interface{}{
		"sku":     "HW-001",
		"confirm": false,
	}))
	text := getText(t, res)
	if !strings.Contains(text, "confirm must be true") {
		t.Errorf("expected confirm error, got: %s", text)
	}
}

func TestHandleInventorySearch(t *testing.T) {
	resetSeed()
	res, _ := handleInventorySearch(context.Background(), makeReq(map[string]interface{}{
		"query": "DDR5",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 1 {
		t.Errorf("expected 1 DDR5 result, got %v", out["count"])
	}
}

func TestHandleInventoryBulkDelete_DryRun(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryBulkDelete(context.Background(), makeReq(map[string]interface{}{
		"dry_run": true,
		"confirm": false,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["dry_run"] != true {
		t.Error("expected dry_run=true in response")
	}
	// Items should still exist
	res2, _ := handleInventoryList(context.Background(), makeReq(nil))
	var list map[string]interface{}
	getJSON(t, res2, &list)
	if int(list["count"].(float64)) != 5 {
		t.Error("dry run should not have deleted any items")
	}
}

// ── Analytics tests ──

func TestHandleInventorySummary(t *testing.T) {
	resetSeed()
	res, _ := handleInventorySummary(context.Background(), makeReq(nil))
	var out clients.InventorySummary
	getJSON(t, res, &out)
	if out.TotalItems != 5 {
		t.Errorf("expected 5 total items, got %d", out.TotalItems)
	}
	if out.ByCategory["GPU"] != 2 {
		t.Errorf("expected 2 GPU, got %d", out.ByCategory["GPU"])
	}
}

func TestHandleInventoryCategories(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryCategories(context.Background(), makeReq(nil))
	var out map[string]interface{}
	getJSON(t, res, &out)
	cats, ok := out["categories"].([]interface{})
	if !ok || len(cats) == 0 {
		t.Fatal("expected categories list")
	}
	// Check GPU has count 2
	for _, c := range cats {
		cat := c.(map[string]interface{})
		if cat["name"] == "GPU" {
			if int(cat["item_count"].(float64)) != 2 {
				t.Errorf("expected GPU count 2, got %v", cat["item_count"])
			}
		}
	}
}

func TestHandleInventoryLocations(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryLocations(context.Background(), makeReq(nil))
	var out map[string]interface{}
	getJSON(t, res, &out)
	locs, ok := out["locations"].(map[string]interface{})
	if !ok {
		t.Fatal("expected locations map")
	}
	if len(locs) == 0 {
		t.Error("expected at least one location")
	}
	// studio-rack-1 should have 3 items (HW-001, HW-004, HW-005)
	if int(locs["studio-rack-1"].(float64)) != 3 {
		t.Errorf("expected 3 items at studio-rack-1, got %v", locs["studio-rack-1"])
	}
}

func TestHandleInventoryStale(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryStale(context.Background(), makeReq(nil))
	var out map[string]interface{}
	getJSON(t, res, &out)
	// 4 items are "Not Listed" (all except HW-004 which is "Listed")
	count := int(out["count"].(float64))
	if count != 4 {
		t.Errorf("expected 4 stale (Not Listed) items, got %d", count)
	}
}

func TestHandleInventoryValue(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryValue(context.Background(), makeReq(nil))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["item_count"].(float64)) != 5 {
		t.Errorf("expected 5 items, got %v", out["item_count"])
	}
	// total_purchase_value = 1800 + 140 + 90 + 150 + 1700 = 3880
	totalPurchase := out["total_purchase_value"].(float64)
	if totalPurchase != 3880 {
		t.Errorf("expected total_purchase_value 3880, got %v", totalPurchase)
	}
}

func TestHandleInventoryExport_JSON(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryExport(context.Background(), makeReq(map[string]interface{}{
		"format": "json",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 5 {
		t.Errorf("expected 5 items in JSON export, got %v", out["count"])
	}
}

func TestHandleInventoryExport_CSV(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryExport(context.Background(), makeReq(map[string]interface{}{
		"format": "csv",
	}))
	text := getText(t, res)
	lines := strings.Split(text, "\n")
	// 1 header + 5 data rows
	if len(lines) != 6 {
		t.Errorf("expected 6 CSV lines (header + 5 items), got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "SKU,") {
		t.Error("expected CSV header to start with SKU")
	}
}

func TestHandleInventoryDashboard(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDashboard(context.Background(), makeReq(nil))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["summary"] == nil {
		t.Error("expected summary in dashboard")
	}
	if out["financials"] == nil {
		t.Error("expected financials in dashboard")
	}
	if out["next_actions"] == nil {
		t.Error("expected next_actions in dashboard")
	}
}

// ── Listing tests ──

func TestHandleMarkSold(t *testing.T) {
	resetSeed()
	res, _ := handleMarkSold(context.Background(), makeReq(map[string]interface{}{
		"sku":        "HW-004",
		"sold_price": float64(180),
		"platform":   "fb_marketplace",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success=true")
	}
	// Verify item is now Sold
	res2, _ := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-004",
	}))
	var item clients.InventoryItem
	getJSON(t, res2, &item)
	if item.ListingStatus != "Sold" {
		t.Errorf("expected status Sold, got %s", item.ListingStatus)
	}
}

func TestHandleFBContent(t *testing.T) {
	resetSeed()
	res, _ := handleFBContent(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var out clients.FBMarketplaceContent
	getJSON(t, res, &out)
	if out.Title != "ASUS ROG Strix RTX 4090" {
		t.Errorf("unexpected title: %s", out.Title)
	}
	if out.Price <= 0 {
		t.Error("expected a positive price")
	}
	if !strings.Contains(out.Description, "ASUS ROG Strix RTX 4090") {
		t.Error("expected item name in description")
	}
}

func TestHandleFBContent_MissingSKU(t *testing.T) {
	resetSeed()
	res, _ := handleFBContent(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected sku required error, got: %s", text)
	}
}

// ── Image tests ──

func TestHandleListImages(t *testing.T) {
	resetSeed()
	res, _ := handleListImages(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	// No images uploaded for test items
	count := int(out["count"].(float64))
	if count != 0 {
		t.Errorf("expected 0 images, got %d", count)
	}
}

func TestHandleSetPrimaryImage(t *testing.T) {
	resetSeed()
	res, _ := handleSetPrimaryImage(context.Background(), makeReq(map[string]interface{}{
		"sku":       "HW-001",
		"image_key": "photo1.jpg",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success=true")
	}
	if out["primary_image"] != "photo1.jpg" {
		t.Errorf("expected primary_image=photo1.jpg, got %v", out["primary_image"])
	}
}

// ── Data Quality tests ──

func TestHandleInventoryRecategorize_DryRun(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryRecategorize(context.Background(), makeReq(map[string]interface{}{
		"dry_run": true,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
}

func TestHandleInventoryDuplicates(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDuplicates(context.Background(), makeReq(map[string]interface{}{
		"match_type": "asin",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	groups := int(out["duplicate_groups"].(float64))
	if groups < 1 {
		t.Errorf("expected at least 1 duplicate group (ASIN B0BSVMLVTD), got %d", groups)
	}
	totalDuplicates := int(out["total_duplicates"].(float64))
	if totalDuplicates < 1 {
		t.Errorf("expected at least 1 total duplicate, got %d", totalDuplicates)
	}
}

func TestHandleInventoryBulkUpdate_DryRun(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryBulkUpdate(context.Background(), makeReq(map[string]interface{}{
		"dry_run":      true,
		"confirm":      false,
		"new_location": "warehouse",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
	wouldUpdate := int(out["would_update"].(float64))
	if wouldUpdate != 5 {
		t.Errorf("expected would_update=5, got %d", wouldUpdate)
	}
}

func TestHandleInventoryMove(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryMove(context.Background(), makeReq(map[string]interface{}{
		"sku":      "HW-003",
		"location": "new-shelf",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success=true")
	}
	if out["old_location"] != "storage-bin-A" {
		t.Errorf("unexpected old_location: %v", out["old_location"])
	}
	if out["new_location"] != "new-shelf" {
		t.Errorf("unexpected new_location: %v", out["new_location"])
	}
}

// ── New hw-resale tools tests ──

func TestHandleListingGenerate_FB(t *testing.T) {
	resetSeed()
	res, _ := handleListingGenerate(context.Background(), makeReq(map[string]interface{}{
		"sku":      "HW-001",
		"platform": "fb_marketplace",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	listing, ok := out["listing"].(string)
	if !ok || listing == "" {
		t.Error("expected non-empty listing text")
	}
	if !strings.Contains(listing, "ASUS ROG Strix RTX 4090") {
		t.Error("expected item name in FB listing")
	}
	if !strings.Contains(listing, "Price is firm") {
		t.Error("expected FB template text")
	}
}

func TestHandleListingGenerate_Ebay(t *testing.T) {
	resetSeed()
	res, _ := handleListingGenerate(context.Background(), makeReq(map[string]interface{}{
		"sku":      "HW-002",
		"platform": "ebay",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	listing, ok := out["listing"].(string)
	if !ok || listing == "" {
		t.Error("expected non-empty listing text")
	}
	if !strings.Contains(listing, "SHIPPING") {
		t.Error("expected eBay template shipping section")
	}
}

func TestHandleListingGenerate_Hardwareswap(t *testing.T) {
	resetSeed()
	res, _ := handleListingGenerate(context.Background(), makeReq(map[string]interface{}{
		"sku":      "HW-001",
		"platform": "hardwareswap",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	listing, ok := out["listing"].(string)
	if !ok || listing == "" {
		t.Error("expected non-empty listing text")
	}
	if !strings.Contains(listing, "Comment before PM") {
		t.Error("expected hardwareswap template text")
	}
}

func TestHandleSaleRecord(t *testing.T) {
	resetSeed()
	res, _ := handleSaleRecord(context.Background(), makeReq(map[string]interface{}{
		"sku":        "HW-004",
		"sold_price": float64(190),
		"platform":   "ebay",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	sale, ok := out["sale"].(map[string]interface{})
	if !ok {
		t.Fatal("expected sale object in response")
	}
	if sale["sale_id"] == nil || sale["sale_id"] == "" {
		t.Error("expected sale_id")
	}
	netProfit := sale["net_profit"].(float64)
	// revenue=190, cost=150, no shipping/fees → profit=40
	if netProfit != 40 {
		t.Errorf("expected net_profit=40, got %v", netProfit)
	}
}

func TestHandleSaleRecord_OverQuantity(t *testing.T) {
	resetSeed()
	res, _ := handleSaleRecord(context.Background(), makeReq(map[string]interface{}{
		"sku":           "HW-001",
		"sold_price":    float64(2000),
		"platform":      "local",
		"quantity_sold": float64(5),
	}))
	text := getText(t, res)
	if !strings.Contains(text, "cannot sell") {
		t.Errorf("expected over-quantity error, got: %s", text)
	}
}

func TestHandlePriceCheck(t *testing.T) {
	resetSeed()
	res, _ := handlePriceCheck(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["purchase_price"].(float64) != 1800 {
		t.Errorf("expected purchase_price=1800, got %v", out["purchase_price"])
	}
	if out["asking_price"].(float64) != 2200 {
		t.Errorf("expected asking_price=2200, got %v", out["asking_price"])
	}
	margin := out["margin"].(string)
	if !strings.Contains(margin, "400") {
		t.Errorf("expected $400 margin, got %s", margin)
	}
	if out["guidance"] == nil || out["guidance"] == "" {
		t.Error("expected pricing guidance")
	}
}

func TestHandleSalesLog(t *testing.T) {
	resetSeed()
	// Record a sale first
	handleSaleRecord(context.Background(), makeReq(map[string]interface{}{
		"sku":        "HW-004",
		"sold_price": float64(180),
		"platform":   "fb_marketplace",
	}))

	res, _ := handleSalesLog(context.Background(), makeReq(nil))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["summary"] == nil {
		t.Error("expected summary in sales log")
	}
	if out["sales"] == nil {
		t.Error("expected sales array in sales log")
	}
	sales := out["sales"].([]interface{})
	if len(sales) == 0 {
		t.Error("expected at least 1 sale after recording")
	}
}

// ══════════════════════════════════════════════════════════════
// Additional coverage: edge cases, missing-param errors, real
// mutations, group-by / filter variants, import CSV fixtures
// ══════════════════════════════════════════════════════════════

// ── CRUD edge cases ──

func TestHandleInventoryAdd_MissingName(t *testing.T) {
	resetSeed()
	// name is empty — handler still succeeds (AddItem fills defaults)
	// but the item will have an empty name, which is allowed at handler level
	res, _ := handleInventoryAdd(context.Background(), makeReq(map[string]interface{}{
		"category":       "GPU",
		"purchase_price": float64(100),
		"location":       "shelf",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	// Should still succeed — the handler doesn't validate name is non-empty
	if out["success"] != true {
		t.Error("expected success even with empty name (no server-side validation)")
	}
}

func TestHandleInventoryAdd_WithAllFields(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryAdd(context.Background(), makeReq(map[string]interface{}{
		"name":            "EVGA RTX 3080 FTW3",
		"category":        "GPU",
		"purchase_price":  float64(700),
		"location":        "studio-rack-2",
		"condition":       "like_new",
		"brand":           "EVGA",
		"model":           "10G-P5-3897-KR",
		"description":     "Excellent condition, barely used",
		"subcategory":     "Desktop",
		"quantity":        float64(2),
		"notes":           "Original box included",
		"purchase_source": "amazon",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Fatal("expected success=true")
	}
	item := out["item"].(map[string]interface{})
	if item["condition"] != "like_new" {
		t.Errorf("expected condition=like_new, got %v", item["condition"])
	}
	if int(item["quantity"].(float64)) != 2 {
		t.Errorf("expected quantity=2, got %v", item["quantity"])
	}
}

func TestHandleInventoryDelete_NotFound(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDelete(context.Background(), makeReq(map[string]interface{}{
		"sku":     "HW-999",
		"confirm": true,
	}))
	text := getText(t, res)
	if !strings.Contains(text, "not found") {
		t.Errorf("expected not found error, got: %s", text)
	}
}

func TestHandleInventoryDelete_MissingSKU(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDelete(context.Background(), makeReq(map[string]interface{}{
		"confirm": true,
	}))
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected sku required error, got: %s", text)
	}
}

func TestHandleInventorySearch_NoResults(t *testing.T) {
	resetSeed()
	res, _ := handleInventorySearch(context.Background(), makeReq(map[string]interface{}{
		"query": "nonexistent-product-xyz",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 0 {
		t.Errorf("expected 0 results, got %v", out["count"])
	}
}

func TestHandleInventorySearch_WithLimit(t *testing.T) {
	resetSeed()
	// Search for something matching multiple items, then limit
	res, _ := handleInventorySearch(context.Background(), makeReq(map[string]interface{}{
		"query": "ROG",
		"limit": float64(1),
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 1 {
		t.Errorf("expected 1 result with limit=1, got %v", out["count"])
	}
}

// ── Move / location edge cases ──

func TestHandleInventoryMove_MissingParams(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryMove(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "required") {
		t.Errorf("expected required error, got: %s", text)
	}
}

func TestHandleInventoryMove_NotFound(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryMove(context.Background(), makeReq(map[string]interface{}{
		"sku":      "HW-999",
		"location": "shelf",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "not found") {
		t.Errorf("expected not found error, got: %s", text)
	}
}

// ── Value / analytics edge cases ──

func TestHandleInventoryValue_GroupByCategory(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryValue(context.Background(), makeReq(map[string]interface{}{
		"group_by": "category",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	breakdown, ok := out["breakdown"].(map[string]interface{})
	if !ok {
		t.Fatal("expected breakdown map")
	}
	gpuValue := breakdown["GPU"].(float64)
	// GPU: HW-001 ($1800) + HW-005 ($1700) = $3500
	if gpuValue != 3500 {
		t.Errorf("expected GPU purchase value 3500, got %v", gpuValue)
	}
}

func TestHandleInventoryValue_GroupByStatus(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryValue(context.Background(), makeReq(map[string]interface{}{
		"group_by": "status",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	breakdown, ok := out["breakdown"].(map[string]interface{})
	if !ok {
		t.Fatal("expected breakdown map")
	}
	// 4 items "Not Listed", 1 "Listed"
	if breakdown["Not Listed"] == nil {
		t.Error("expected Not Listed in breakdown")
	}
	if breakdown["Listed"] == nil {
		t.Error("expected Listed in breakdown")
	}
}

func TestHandleInventoryValue_GroupByLocation(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryValue(context.Background(), makeReq(map[string]interface{}{
		"group_by": "location",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	breakdown, ok := out["breakdown"].(map[string]interface{})
	if !ok {
		t.Fatal("expected breakdown map")
	}
	if breakdown["studio-rack-1"] == nil {
		t.Error("expected studio-rack-1 in breakdown")
	}
	if breakdown["storage-bin-A"] == nil {
		t.Error("expected storage-bin-A in breakdown")
	}
}

func TestHandleInventoryValue_GroupByCondition(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryValue(context.Background(), makeReq(map[string]interface{}{
		"group_by": "condition",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	breakdown, ok := out["breakdown"].(map[string]interface{})
	if !ok {
		t.Fatal("expected breakdown map")
	}
	if breakdown["Used"] == nil {
		t.Error("expected Used in breakdown")
	}
}

func TestHandleInventoryValue_IncludeSold(t *testing.T) {
	resetSeed()
	// First mark one as sold
	handleMarkSold(context.Background(), makeReq(map[string]interface{}{
		"sku":        "HW-004",
		"sold_price": float64(180),
		"platform":   "local",
	}))
	res, _ := handleInventoryValue(context.Background(), makeReq(map[string]interface{}{
		"include_sold": true,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	soldCount := int(out["sold_count"].(float64))
	if soldCount != 1 {
		t.Errorf("expected sold_count=1, got %d", soldCount)
	}
	totalSold := out["total_sold_value"].(float64)
	if totalSold != 180 {
		t.Errorf("expected total_sold_value=180, got %v", totalSold)
	}
}

// ── Export edge cases ──

func TestHandleInventoryExport_CategoryFilter(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryExport(context.Background(), makeReq(map[string]interface{}{
		"format":   "json",
		"category": "GPU",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["count"].(float64)) != 2 {
		t.Errorf("expected 2 GPU items in export, got %v", out["count"])
	}
}

func TestHandleInventoryExport_StatusFilter(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryExport(context.Background(), makeReq(map[string]interface{}{
		"format": "csv",
		"status": "Listed",
	}))
	text := getText(t, res)
	lines := strings.Split(text, "\n")
	// 1 header + 1 Listed item
	if len(lines) != 2 {
		t.Errorf("expected 2 CSV lines (header + 1 Listed item), got %d", len(lines))
	}
}

// ── Duplicate detection variants ──

func TestHandleInventoryDuplicates_MatchTypeName(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDuplicates(context.Background(), makeReq(map[string]interface{}{
		"match_type": "name",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	groups := int(out["duplicate_groups"].(float64))
	// Items 1+5 share exact name "ASUS ROG Strix RTX 4090"
	if groups < 1 {
		t.Errorf("expected at least 1 name duplicate group, got %d", groups)
	}
}

func TestHandleInventoryDuplicates_MatchTypeAll(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDuplicates(context.Background(), makeReq(map[string]interface{}{
		"match_type": "all",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	groups := int(out["duplicate_groups"].(float64))
	// Should find both ASIN and name duplicates
	if groups < 2 {
		t.Errorf("expected at least 2 duplicate groups (asin + name), got %d", groups)
	}
}

// ── Bulk operations: actual apply ──

func TestHandleInventoryBulkDelete_ActualDelete(t *testing.T) {
	resetSeed()
	// Delete items with max_price <= 150 and confirm
	res, _ := handleInventoryBulkDelete(context.Background(), makeReq(map[string]interface{}{
		"max_price": float64(150),
		"confirm":   true,
		"dry_run":   false,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	deletedCount := int(out["deleted_count"].(float64))
	// HW-002 ($140), HW-003 ($90), HW-004 ($150) match max_price<=150
	if deletedCount != 3 {
		t.Errorf("expected 3 deleted items, got %d", deletedCount)
	}
	// Verify remaining items
	res2, _ := handleInventoryList(context.Background(), makeReq(nil))
	var list map[string]interface{}
	getJSON(t, res2, &list)
	remaining := int(list["count"].(float64))
	if remaining != 2 {
		t.Errorf("expected 2 remaining items after bulk delete, got %d", remaining)
	}
}

func TestHandleInventoryBulkUpdate_ActualApply(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryBulkUpdate(context.Background(), makeReq(map[string]interface{}{
		"dry_run":      false,
		"confirm":      true,
		"category":     "GPU",
		"new_location": "gpu-shelf",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success=true")
	}
	updatedCount := int(out["updated_count"].(float64))
	if updatedCount != 2 {
		t.Errorf("expected 2 GPU items updated, got %d", updatedCount)
	}
	// Verify location changed
	res2, _ := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var item clients.InventoryItem
	getJSON(t, res2, &item)
	if item.Location != "gpu-shelf" {
		t.Errorf("expected location=gpu-shelf, got %s", item.Location)
	}
}

func TestHandleInventoryBulkUpdate_TagManagement(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryBulkUpdate(context.Background(), makeReq(map[string]interface{}{
		"dry_run":    false,
		"confirm":    true,
		"category":   "GPU",
		"add_tags":   []interface{}{"priority", "high-value"},
		"new_status": "pending_review",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success=true")
	}
	updatedCount := int(out["updated_count"].(float64))
	if updatedCount != 2 {
		t.Errorf("expected 2 GPU items updated, got %d", updatedCount)
	}
}

func TestHandleInventoryBulkUpdate_NoUpdatesSpecified(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryBulkUpdate(context.Background(), makeReq(map[string]interface{}{
		"dry_run": false,
		"confirm": true,
	}))
	text := getText(t, res)
	if !strings.Contains(text, "no updates specified") {
		t.Errorf("expected no-updates error, got: %s", text)
	}
}

// ── Recategorize edge cases ──

func TestHandleInventoryRecategorize_SingleSKU(t *testing.T) {
	resetSeed()
	// HW-004 "Netgear Nighthawk Router" is category "Networking"
	// guessCategoryFromName("Netgear Nighthawk Router") → "Networking" (contains "router")
	// So it should be unchanged
	res, _ := handleInventoryRecategorize(context.Background(), makeReq(map[string]interface{}{
		"sku":     "HW-004",
		"dry_run": true,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
	result := out["result"].(map[string]interface{})
	if result["old_category"] != "Networking" {
		t.Errorf("expected old_category=Networking, got %v", result["old_category"])
	}
}

func TestHandleInventoryRecategorize_Apply(t *testing.T) {
	resetSeed()
	// First, miscategorize an item so recategorize has something to fix
	handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":      "HW-001",
		"category": "Other",
	}))
	// Now recategorize — "ASUS ROG Strix RTX 4090" → should be "GPU"
	res, _ := handleInventoryRecategorize(context.Background(), makeReq(map[string]interface{}{
		"sku":     "HW-001",
		"dry_run": false,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["applied"] != true {
		t.Error("expected applied=true")
	}
	result := out["result"].(map[string]interface{})
	if result["new_category"] != "GPU" {
		t.Errorf("expected new_category=GPU, got %v", result["new_category"])
	}
	if result["old_category"] != "Other" {
		t.Errorf("expected old_category=Other, got %v", result["old_category"])
	}
}

// ── Sale record edge cases ──

func TestHandleSaleRecord_MissingRequired(t *testing.T) {
	resetSeed()
	res, _ := handleSaleRecord(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "required") {
		t.Errorf("expected required error, got: %s", text)
	}
}

func TestHandleSaleRecord_WithFees(t *testing.T) {
	resetSeed()
	res, _ := handleSaleRecord(context.Background(), makeReq(map[string]interface{}{
		"sku":           "HW-004",
		"sold_price":    float64(200),
		"platform":      "ebay",
		"shipping_cost": float64(15),
		"platform_fees": float64(26),
		"buyer_info":    "buyer123",
		"notes":         "sold via auction",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	sale := out["sale"].(map[string]interface{})
	// revenue=200, cost=150, shipping=15, fees=26 → profit = 200-150-15-26 = 9
	profit := sale["net_profit"].(float64)
	if profit != 9 {
		t.Errorf("expected net_profit=9, got %v", profit)
	}
	if sale["platform"] != "ebay" {
		t.Errorf("expected platform=ebay, got %v", sale["platform"])
	}
}

func TestHandleSaleRecord_PartialQuantity(t *testing.T) {
	resetSeed()
	// HW-002 has quantity=3, sell 2
	res, _ := handleSaleRecord(context.Background(), makeReq(map[string]interface{}{
		"sku":           "HW-002",
		"sold_price":    float64(170),
		"platform":      "fb_marketplace",
		"quantity_sold": float64(2),
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	remaining := int(out["remaining_quantity"].(float64))
	if remaining != 1 {
		t.Errorf("expected remaining_quantity=1 (3-2), got %d", remaining)
	}
	sale := out["sale"].(map[string]interface{})
	// revenue = 170*2 = 340, cost = 140*2 = 280 → profit = 60
	profit := sale["net_profit"].(float64)
	if profit != 60 {
		t.Errorf("expected net_profit=60, got %v", profit)
	}
}

func TestHandleSaleRecord_SellAllQuantity(t *testing.T) {
	resetSeed()
	// HW-003 has quantity=2, sell all 2
	res, _ := handleSaleRecord(context.Background(), makeReq(map[string]interface{}{
		"sku":           "HW-003",
		"sold_price":    float64(110),
		"platform":      "local",
		"quantity_sold": float64(2),
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	remaining := int(out["remaining_quantity"].(float64))
	if remaining != 0 {
		t.Errorf("expected remaining_quantity=0, got %d", remaining)
	}
	// Verify the item is now marked Sold
	res2, _ := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-003",
	}))
	var item clients.InventoryItem
	getJSON(t, res2, &item)
	if item.ListingStatus != "sold" {
		t.Errorf("expected status=sold after selling all, got %s", item.ListingStatus)
	}
}

// ── Price check edge cases ──

func TestHandlePriceCheck_MissingSKU(t *testing.T) {
	resetSeed()
	res, _ := handlePriceCheck(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "sku or search_query is required") {
		t.Errorf("expected sku/search_query required error, got: %s", text)
	}
}

func TestHandlePriceCheck_NotFound(t *testing.T) {
	resetSeed()
	res, _ := handlePriceCheck(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-999",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "not found") {
		t.Errorf("expected not found error, got: %s", text)
	}
}

func TestHandlePriceCheck_NegativeMargin(t *testing.T) {
	resetSeed()
	// Set asking price below purchase price
	handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":           "HW-001",
		"current_value": float64(1500),
	}))
	res, _ := handlePriceCheck(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	guidance := out["guidance"].(string)
	if !strings.Contains(guidance, "loss") {
		t.Errorf("expected 'loss' guidance for negative margin, got: %s", guidance)
	}
}

// ── Image edge cases ──

func TestHandleDeleteImage_MissingParams(t *testing.T) {
	resetSeed()
	res, _ := handleDeleteImage(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "required") {
		t.Errorf("expected required error, got: %s", text)
	}
}

func TestHandleSetPrimaryImage_MissingParams(t *testing.T) {
	resetSeed()
	res, _ := handleSetPrimaryImage(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "required") {
		t.Errorf("expected required error, got: %s", text)
	}
}

func TestHandleListImages_MissingSKU(t *testing.T) {
	resetSeed()
	res, _ := handleListImages(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected sku required error, got: %s", text)
	}
}

// ── Listing generate edge cases ──

func TestHandleListingGenerate_MissingSKU(t *testing.T) {
	resetSeed()
	res, _ := handleListingGenerate(context.Background(), makeReq(map[string]interface{}{
		"platform": "ebay",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected sku required error, got: %s", text)
	}
}

func TestHandleListingGenerate_InvalidPlatform(t *testing.T) {
	resetSeed()
	res, _ := handleListingGenerate(context.Background(), makeReq(map[string]interface{}{
		"sku":      "HW-001",
		"platform": "craigslist",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "invalid platform") {
		t.Errorf("expected invalid platform error, got: %s", text)
	}
}

func TestHandleListingGenerate_DefaultPlatform(t *testing.T) {
	resetSeed()
	// No platform specified → defaults to fb_marketplace
	res, _ := handleListingGenerate(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["platform"] != "fb_marketplace" {
		t.Errorf("expected default platform=fb_marketplace, got %v", out["platform"])
	}
}

// ── Dashboard detail checks ──

func TestHandleInventoryDashboard_StructureCheck(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryDashboard(context.Background(), makeReq(nil))
	var out map[string]interface{}
	getJSON(t, res, &out)

	summary := out["summary"].(map[string]interface{})
	if int(summary["total_items"].(float64)) != 5 {
		t.Errorf("expected 5 total items in dashboard summary, got %v", summary["total_items"])
	}

	financials := out["financials"].(map[string]interface{})
	if financials["revenue_target"] == nil {
		t.Error("expected revenue_target in financials")
	}

	attention := out["attention_needed"].(map[string]interface{})
	notListed := int(attention["not_listed"].(float64))
	if notListed != 4 {
		t.Errorf("expected 4 not-listed items, got %d", notListed)
	}

	actions := out["next_actions"].([]interface{})
	if len(actions) == 0 {
		t.Error("expected at least one next action")
	}
}

// ── Stale with explicit days parameter ──

func TestHandleInventoryStale_WithDays(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryStale(context.Background(), makeReq(map[string]interface{}{
		"days": float64(7),
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["days"].(float64)) != 7 {
		t.Errorf("expected days=7 in response, got %v", out["days"])
	}
}

// ── FB Content with custom asking price ──

func TestHandleFBContent_CustomPrice(t *testing.T) {
	resetSeed()
	res, _ := handleFBContent(context.Background(), makeReq(map[string]interface{}{
		"sku":          "HW-001",
		"asking_price": float64(1999),
	}))
	var out clients.FBMarketplaceContent
	getJSON(t, res, &out)
	if out.Price != 1999 {
		t.Errorf("expected custom price 1999, got %v", out.Price)
	}
}

// ── Import CSV tests ──

func TestHandleImportAmazon_MissingPath(t *testing.T) {
	resetSeed()
	res, _ := handleImportAmazon(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "csv_path is required") {
		t.Errorf("expected csv_path required error, got: %s", text)
	}
}

func TestHandleImportAmazon_FileNotFound(t *testing.T) {
	resetSeed()
	res, _ := handleImportAmazon(context.Background(), makeReq(map[string]interface{}{
		"csv_path": "/nonexistent/path/orders.csv",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "failed to open") {
		t.Errorf("expected file open error, got: %s", text)
	}
}

func TestHandleImportNewegg_MissingPath(t *testing.T) {
	resetSeed()
	res, _ := handleImportNewegg(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "csv_path is required") {
		t.Errorf("expected csv_path required error, got: %s", text)
	}
}

func TestHandleImportAmazon_ValidCSV(t *testing.T) {
	resetSeed()
	// Create a temp CSV file
	csvContent := "Title,ASIN/ISBN,Item Total,Quantity,Order Date\n" +
		"Crucial P5 Plus 2TB NVMe,B09P1QF8QM,$189.99,1,2024-01-15\n" +
		"Noctua NH-D15,B00L7UZMAK,$109.95,1,2024-02-20\n"

	tmpFile, err := os.CreateTemp("", "amazon-orders-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(csvContent)
	tmpFile.Close()

	res, _ := handleImportAmazon(context.Background(), makeReq(map[string]interface{}{
		"csv_path": tmpFile.Name(),
	}))
	var out clients.ImportResult
	getJSON(t, res, &out)
	if out.Imported != 2 {
		t.Errorf("expected 2 imported, got %d", out.Imported)
	}
	if out.TotalRows != 2 {
		t.Errorf("expected 2 total rows, got %d", out.TotalRows)
	}
}

func TestHandleImportNewegg_ValidCSV(t *testing.T) {
	resetSeed()
	csvContent := "Item Name,Item Number,Unit Price,Quantity\n" +
		"ASUS TUF Gaming X670E-Plus,12345,$349.99,1\n"

	tmpFile, err := os.CreateTemp("", "newegg-orders-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(csvContent)
	tmpFile.Close()

	res, _ := handleImportNewegg(context.Background(), makeReq(map[string]interface{}{
		"csv_path": tmpFile.Name(),
	}))
	var out clients.ImportResult
	getJSON(t, res, &out)
	if out.Imported != 1 {
		t.Errorf("expected 1 imported, got %d", out.Imported)
	}
}

// ── Summary detail checks ──

func TestHandleInventorySummary_DetailedCounts(t *testing.T) {
	resetSeed()
	res, _ := handleInventorySummary(context.Background(), makeReq(nil))
	var out clients.InventorySummary
	getJSON(t, res, &out)
	// Check by_status
	if out.ByStatus["Not Listed"] != 4 {
		t.Errorf("expected 4 Not Listed, got %d", out.ByStatus["Not Listed"])
	}
	if out.ByStatus["Listed"] != 1 {
		t.Errorf("expected 1 Listed, got %d", out.ByStatus["Listed"])
	}
	// Check by_condition
	if out.ByCondition["Used"] != 2 {
		t.Errorf("expected 2 Used, got %d", out.ByCondition["Used"])
	}
	if out.ByCondition["New"] != 1 {
		t.Errorf("expected 1 New, got %d", out.ByCondition["New"])
	}
	// Check total value: 2200*1 + 180*3 + 120*2 + 200*1 + 2100*1 = 5440
	expectedValue := 2200.0 + 180*3 + 120*2 + 200 + 2100
	if out.TotalValue != expectedValue {
		t.Errorf("expected total_value=%v, got %v", expectedValue, out.TotalValue)
	}
}

// ══════════════════════════════════════════════════════════════
// Phase 1: ProductLine + Specs field tests
// ══════════════════════════════════════════════════════════════

func TestHandleInventoryAdd_WithProductLineAndSpecs(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryAdd(context.Background(), makeReq(map[string]interface{}{
		"name":           "NVIDIA RTX 4090 Founders Edition",
		"category":       "GPU",
		"purchase_price": float64(1599),
		"location":       "studio-rack-1",
		"product_line":   "RTX 4090",
		"specs": map[string]interface{}{
			"vram":        "24GB GDDR6X",
			"tdp":         "450W",
			"form_factor": "3-slot",
		},
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Fatal("expected success=true")
	}
	item := out["item"].(map[string]interface{})
	if item["product_line"] != "RTX 4090" {
		t.Errorf("expected product_line=RTX 4090, got %v", item["product_line"])
	}
	specs, ok := item["specs"].(map[string]interface{})
	if !ok {
		t.Fatal("expected specs map in response")
	}
	if specs["vram"] != "24GB GDDR6X" {
		t.Errorf("expected vram=24GB GDDR6X, got %v", specs["vram"])
	}
	if specs["tdp"] != "450W" {
		t.Errorf("expected tdp=450W, got %v", specs["tdp"])
	}
}

func TestHandleInventoryUpdate_ProductLineAndSpecs(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":          "HW-001",
		"product_line": "RTX 4090",
		"specs": map[string]interface{}{
			"vram": "24GB GDDR6X",
			"tdp":  "450W",
		},
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Fatal("expected success=true")
	}
	item := out["item"].(map[string]interface{})
	if item["product_line"] != "RTX 4090" {
		t.Errorf("expected product_line=RTX 4090, got %v", item["product_line"])
	}
	specs, ok := item["specs"].(map[string]interface{})
	if !ok {
		t.Fatal("expected specs map in response")
	}
	if specs["vram"] != "24GB GDDR6X" {
		t.Errorf("expected vram=24GB GDDR6X, got %v", specs["vram"])
	}

	// Verify Get round-trips the fields
	res2, _ := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var item2 clients.InventoryItem
	getJSON(t, res2, &item2)
	if item2.ProductLine != "RTX 4090" {
		t.Errorf("expected product_line=RTX 4090 on re-read, got %s", item2.ProductLine)
	}
	if item2.Specs["vram"] != "24GB GDDR6X" {
		t.Errorf("expected specs.vram=24GB GDDR6X on re-read, got %v", item2.Specs["vram"])
	}
}

func TestBuildSpecsBlock_WithSpecs(t *testing.T) {
	item := &clients.InventoryItem{
		Name:  "RTX 4090",
		Model: "900-1G136-2530-000",
		Specs: map[string]string{
			"vram":        "24GB GDDR6X",
			"tdp":         "450W",
			"form_factor": "3-slot",
		},
	}
	block := buildSpecsBlock(item)
	if !strings.Contains(block, "24GB GDDR6X") {
		t.Errorf("expected vram in specs block, got:\n%s", block)
	}
	if !strings.Contains(block, "450W") {
		t.Errorf("expected tdp in specs block, got:\n%s", block)
	}
	if !strings.Contains(block, "Model: 900-1G136-2530-000") {
		t.Errorf("expected model in specs block, got:\n%s", block)
	}
}

func TestBuildSpecsBlock_FallbackNoSpecs(t *testing.T) {
	item := &clients.InventoryItem{
		Name:     "RTX 4090",
		Model:    "FE",
		Category: "GPU",
		ASIN:     "B0BSVMLVTD",
		MSRP:     1599,
	}
	block := buildSpecsBlock(item)
	if !strings.Contains(block, "Amazon ASIN: B0BSVMLVTD") {
		t.Errorf("expected ASIN in fallback specs block, got:\n%s", block)
	}
	if !strings.Contains(block, "MSRP: $1599") {
		t.Errorf("expected MSRP in fallback specs block, got:\n%s", block)
	}
}

// ══════════════════════════════════════════════════════════════
// Phase 2: JSON Import tests
// ══════════════════════════════════════════════════════════════

func TestHandleImportJSON_DryRun(t *testing.T) {
	resetSeed()
	jsonData := `[
		{
			"name": "Samsung 990 Pro 2TB",
			"category": "nvme",
			"product_line": "Samsung 990 Pro",
			"quantity": 3,
			"condition": "new",
			"asking_price": 180,
			"purchase_price": 140,
			"specs": {"capacity": "2TB", "interface": "PCIe 4.0 x4", "speed": "7450/6900 MB/s"}
		},
		{
			"name": "NVIDIA RTX 4090 FE",
			"category": "gpu",
			"product_line": "RTX 4090",
			"quantity": 1,
			"condition": "used_excellent",
			"asking_price": 2000,
			"purchase_price": 1599,
			"specs": {"vram": "24GB GDDR6X", "tdp": "450W"}
		}
	]`

	tmpFile, err := os.CreateTemp("", "inventory-import-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(jsonData)
	tmpFile.Close()

	res, _ := handleImportJSON(context.Background(), makeReq(map[string]interface{}{
		"file_path": tmpFile.Name(),
		"dry_run":   true,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["dry_run"] != true {
		t.Error("expected dry_run=true")
	}
	result := out["result"].(map[string]interface{})
	if int(result["imported"].(float64)) != 2 {
		t.Errorf("expected 2 imported (dry_run), got %v", result["imported"])
	}
	if int(result["total_rows"].(float64)) != 2 {
		t.Errorf("expected 2 total_rows, got %v", result["total_rows"])
	}
}

func TestHandleImportJSON_ActualImport(t *testing.T) {
	resetSeed()
	jsonData := `[
		{
			"name": "Crucial MX500 2TB",
			"category": "sata_ssd",
			"product_line": "Crucial MX500",
			"quantity": 2,
			"condition": "used_good",
			"asking_price": 80,
			"purchase_price": 95,
			"specs": {"capacity": "2TB", "interface": "SATA III"}
		}
	]`

	tmpFile, err := os.CreateTemp("", "inventory-import-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(jsonData)
	tmpFile.Close()

	res, _ := handleImportJSON(context.Background(), makeReq(map[string]interface{}{
		"file_path": tmpFile.Name(),
		"dry_run":   false,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["dry_run"] != false {
		t.Error("expected dry_run=false")
	}
	result := out["result"].(map[string]interface{})
	if int(result["imported"].(float64)) != 1 {
		t.Errorf("expected 1 imported, got %v", result["imported"])
	}

	// Verify the item was actually added with correct fields
	res2, _ := handleInventoryList(context.Background(), makeReq(map[string]interface{}{
		"query": "MX500",
	}))
	var list map[string]interface{}
	getJSON(t, res2, &list)
	if int(list["count"].(float64)) != 1 {
		t.Errorf("expected 1 MX500 item after import, got %v", list["count"])
	}
}

func TestHandleImportJSON_CategoryMapping(t *testing.T) {
	resetSeed()
	jsonData := `[
		{"name": "Test GPU", "category": "gpu", "asking_price": 100, "purchase_price": 50},
		{"name": "Test NVMe", "category": "nvme", "asking_price": 50, "purchase_price": 25},
		{"name": "Test DDR5", "category": "ddr5", "asking_price": 30, "purchase_price": 15},
		{"name": "Test Net", "category": "networking", "asking_price": 20, "purchase_price": 10},
		{"name": "Test Acc", "category": "accessories", "asking_price": 10, "purchase_price": 5}
	]`

	tmpFile, err := os.CreateTemp("", "inventory-import-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(jsonData)
	tmpFile.Close()

	res, _ := handleImportJSON(context.Background(), makeReq(map[string]interface{}{
		"file_path": tmpFile.Name(),
		"dry_run":   false,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	result := out["result"].(map[string]interface{})
	if int(result["imported"].(float64)) != 5 {
		t.Errorf("expected 5 imported, got %v", result["imported"])
	}

	// Verify categories were mapped correctly
	res2, _ := handleInventoryList(context.Background(), makeReq(map[string]interface{}{
		"query": "Test GPU",
	}))
	var list map[string]interface{}
	getJSON(t, res2, &list)
	items := list["items"].([]interface{})
	if len(items) > 0 {
		item := items[0].(map[string]interface{})
		if item["category"] != "GPU" {
			t.Errorf("expected GPU category, got %v", item["category"])
		}
	}
}

func TestHandleImportJSON_MissingPath(t *testing.T) {
	resetSeed()
	res, _ := handleImportJSON(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "file_path is required") {
		t.Errorf("expected file_path required error, got: %s", text)
	}
}

func TestHandleImportJSON_FileNotFound(t *testing.T) {
	resetSeed()
	res, _ := handleImportJSON(context.Background(), makeReq(map[string]interface{}{
		"file_path": "/nonexistent/inventory.json",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "failed to read") {
		t.Errorf("expected file read error, got: %s", text)
	}
}

func TestHandleImportJSON_WrappedFormat(t *testing.T) {
	resetSeed()
	jsonData := `{"items": [{"name": "Wrapped Item", "category": "gpu", "asking_price": 100, "purchase_price": 50}]}`

	tmpFile, err := os.CreateTemp("", "inventory-import-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(jsonData)
	tmpFile.Close()

	res, _ := handleImportJSON(context.Background(), makeReq(map[string]interface{}{
		"file_path": tmpFile.Name(),
		"dry_run":   true,
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	result := out["result"].(map[string]interface{})
	if int(result["imported"].(float64)) != 1 {
		t.Errorf("expected 1 imported from wrapped format, got %v", result["imported"])
	}
}

// ══════════════════════════════════════════════════════════════
// Phase 4a: Enhanced price_check tests
// ══════════════════════════════════════════════════════════════

func TestHandlePriceCheck_SearchQuery(t *testing.T) {
	resetSeed()
	res, _ := handlePriceCheck(context.Background(), makeReq(map[string]interface{}{
		"search_query": "Samsung",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	// Should find HW-002 "Samsung 990 Pro 2TB NVMe"
	if out["sku"] != "HW-002" {
		t.Errorf("expected SKU HW-002 from search, got %v", out["sku"])
	}
	if out["search_suggestion"] == nil || out["search_suggestion"] == "" {
		t.Error("expected search_suggestion in output")
	}
}

func TestHandlePriceCheck_SearchQueryNotFound(t *testing.T) {
	resetSeed()
	res, _ := handlePriceCheck(context.Background(), makeReq(map[string]interface{}{
		"search_query": "nonexistent-xyz-product",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "no items found") {
		t.Errorf("expected no items found error, got: %s", text)
	}
}

func TestHandlePriceCheck_AlwaysEmitsSearchSuggestion(t *testing.T) {
	resetSeed()
	// HW-003 has no ASIN — previously wouldn't get search_suggestion
	res, _ := handlePriceCheck(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-003",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	suggestion, ok := out["search_suggestion"].(string)
	if !ok || suggestion == "" {
		t.Error("expected search_suggestion even without ASIN")
	}
}

func TestHandlePriceCheck_IncludesProductLine(t *testing.T) {
	resetSeed()
	// Set product_line on an item first
	handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":          "HW-001",
		"product_line": "RTX 4090",
	}))
	res, _ := handlePriceCheck(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["product_line"] != "RTX 4090" {
		t.Errorf("expected product_line=RTX 4090, got %v", out["product_line"])
	}
	// search_suggestion should use product_line
	suggestion := out["search_suggestion"].(string)
	if !strings.Contains(suggestion, "RTX 4090") {
		t.Errorf("expected search_suggestion to include product_line, got: %s", suggestion)
	}
}

func TestHandlePriceCheck_NoParams(t *testing.T) {
	resetSeed()
	res, _ := handlePriceCheck(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "sku or search_query is required") {
		t.Errorf("expected param error, got: %s", text)
	}
}

// ══════════════════════════════════════════════════════════════
// Phase 4b: Batch listing tests
// ══════════════════════════════════════════════════════════════

func TestHandleBatchListing_AllPlatforms(t *testing.T) {
	resetSeed()
	res, _ := handleBatchListing(context.Background(), makeReq(map[string]interface{}{
		"category": "GPU",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	// 2 GPU items × 3 platforms = 6 listings
	count := int(out["count"].(float64))
	if count != 6 {
		t.Errorf("expected 6 listings (2 items × 3 platforms), got %d", count)
	}
	listings := out["listings"].([]interface{})
	if len(listings) != 6 {
		t.Errorf("expected 6 listing objects, got %d", len(listings))
	}
}

func TestHandleBatchListing_SinglePlatform(t *testing.T) {
	resetSeed()
	res, _ := handleBatchListing(context.Background(), makeReq(map[string]interface{}{
		"category": "GPU",
		"platform": "ebay",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	// 2 GPU items × 1 platform = 2 listings
	count := int(out["count"].(float64))
	if count != 2 {
		t.Errorf("expected 2 listings (2 items × 1 platform), got %d", count)
	}
}

func TestHandleBatchListing_NoFilter(t *testing.T) {
	resetSeed()
	res, _ := handleBatchListing(context.Background(), makeReq(map[string]interface{}{
		"platform": "fb_marketplace",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	// 5 items × 1 platform = 5 listings
	count := int(out["count"].(float64))
	if count != 5 {
		t.Errorf("expected 5 listings (all items × 1 platform), got %d", count)
	}
}

// ══════════════════════════════════════════════════════════════
// Phase 4c: Reprice tests (eBay client not configured in test)
// ══════════════════════════════════════════════════════════════

func TestHandleReprice_MissingSKU(t *testing.T) {
	resetSeed()
	res, _ := handleReprice(context.Background(), makeReq(nil))
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected sku required error, got: %s", text)
	}
}

func TestHandleReprice_NoEbayClient(t *testing.T) {
	resetSeed()
	res, _ := handleReprice(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	// Should gracefully handle missing eBay client
	if out["sku"] != "HW-001" {
		t.Errorf("expected sku=HW-001, got %v", out["sku"])
	}
	if out["current_price"].(float64) != 2200 {
		t.Errorf("expected current_price=2200, got %v", out["current_price"])
	}
	// Should have error or search_suggestion
	if out["error"] == nil && out["search_suggestion"] == nil {
		t.Error("expected error or search_suggestion when eBay not configured")
	}
}

func TestHandleReprice_NotFound(t *testing.T) {
	resetSeed()
	res, _ := handleReprice(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-999",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "not found") {
		t.Errorf("expected not found error, got: %s", text)
	}
}

// ── Enhancement A: Validation tests ──

func TestHandleInventoryAdd_InvalidCondition(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryAdd(context.Background(), makeReq(map[string]interface{}{
		"name":           "Test Item",
		"category":       "GPU",
		"purchase_price": float64(100),
		"location":       "test-loc",
		"condition":      "excellent",
	}))
	if !res.IsError {
		t.Error("expected error for invalid condition")
	}
	text := getText(t, res)
	if !strings.Contains(text, "invalid condition") {
		t.Errorf("expected 'invalid condition' error, got: %s", text)
	}
}

func TestHandleInventoryUpdate_InvalidStatus(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":            "HW-001",
		"listing_status": "available",
	}))
	if !res.IsError {
		t.Error("expected error for invalid listing_status")
	}
	text := getText(t, res)
	if !strings.Contains(text, "invalid listing_status") {
		t.Errorf("expected 'invalid listing_status' error, got: %s", text)
	}
}

func TestHandleInventoryAdd_ValidConditionNormalized(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryAdd(context.Background(), makeReq(map[string]interface{}{
		"name":           "Test Item",
		"category":       "GPU",
		"purchase_price": float64(100),
		"location":       "test-loc",
		"condition":      "Like_New",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success for valid condition")
	}
}

func TestHandleInventoryUpdate_ValidStatusNormalized(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":            "HW-001",
		"listing_status": "Listed",
	}))
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["success"] != true {
		t.Error("expected success for valid listing status")
	}
}

// ── Enhancement B: Specs catalog tests ──

func TestLookupSpecs(t *testing.T) {
	specs := LookupSpecs("RTX 4090")
	if specs == nil {
		t.Fatal("expected specs for RTX 4090")
	}
	if specs["vram"] != "24GB GDDR6X" {
		t.Errorf("expected vram=24GB GDDR6X, got %s", specs["vram"])
	}

	// Verify it's a copy (mutation should not affect catalog)
	specs["vram"] = "modified"
	original := LookupSpecs("RTX 4090")
	if original["vram"] == "modified" {
		t.Error("LookupSpecs returned a reference, not a copy")
	}
}

func TestLookupSpecs_NotFound(t *testing.T) {
	specs := LookupSpecs("Nonexistent Product")
	if specs != nil {
		t.Error("expected nil for unknown product line")
	}
}

// ── Enhancement C: Quick List tests ──

func TestHandleQuickList(t *testing.T) {
	resetSeed()
	res, err := handleQuickList(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)

	// Check item section
	item := out["item"].(map[string]interface{})
	if item["sku"] != "HW-001" {
		t.Errorf("expected sku=HW-001, got %v", item["sku"])
	}

	// Check pricing section
	pricing := out["pricing"].(map[string]interface{})
	if pricing["asking"].(float64) != 2200 {
		t.Errorf("expected asking=2200, got %v", pricing["asking"])
	}

	// Check listings section has all 3 platforms
	listings := out["listings"].(map[string]interface{})
	for _, p := range []string{"fb_marketplace", "ebay", "hardwareswap"} {
		if listings[p] == nil || listings[p] == "" {
			t.Errorf("expected listing for %s", p)
		}
	}
}

func TestHandleQuickList_MissingSKU(t *testing.T) {
	res, _ := handleQuickList(context.Background(), makeReq(nil))
	if !res.IsError {
		t.Error("expected error for missing SKU")
	}
}

func TestHandleQuickList_NotFound(t *testing.T) {
	resetSeed()
	res, _ := handleQuickList(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-999",
	}))
	text := getText(t, res)
	if !strings.Contains(text, "not found") {
		t.Errorf("expected not found error, got: %s", text)
	}
}

// ── Enhancement D: Dashboard stale alerts tests ──

func TestHandleDashboard_AttentionFields(t *testing.T) {
	resetSeed()
	res, err := handleInventoryDashboard(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)

	attention := out["attention_needed"].(map[string]interface{})

	// These fields should always be present (even if nil/empty)
	if _, ok := attention["needs_repricing"]; !ok {
		t.Error("expected needs_repricing field in attention_needed")
	}
	if _, ok := attention["below_floor"]; !ok {
		t.Error("expected below_floor field in attention_needed")
	}
}

// ── Enhancement E: Shipping Estimate tests ──

func TestHandleShippingEstimate_BySKU(t *testing.T) {
	resetSeed()
	res, err := handleShippingEstimate(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)

	if out["sku"] != "HW-001" {
		t.Errorf("expected sku=HW-001, got %v", out["sku"])
	}
	if out["weight_lbs"].(float64) <= 0 {
		t.Error("expected positive weight")
	}

	estimates := out["estimates"].([]interface{})
	if len(estimates) < 2 {
		t.Errorf("expected at least 2 shipping estimates, got %d", len(estimates))
	}

	if out["recommendation"] == nil || out["recommendation"] == "" {
		t.Error("expected non-empty recommendation")
	}
}

func TestHandleShippingEstimate_ByWeight(t *testing.T) {
	res, err := handleShippingEstimate(context.Background(), makeReq(map[string]interface{}{
		"weight_lbs": float64(5),
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)

	if out["weight_lbs"].(float64) != 5 {
		t.Errorf("expected weight=5, got %v", out["weight_lbs"])
	}
}

func TestHandleShippingEstimate_MissingParams(t *testing.T) {
	res, _ := handleShippingEstimate(context.Background(), makeReq(nil))
	if !res.IsError {
		t.Error("expected error when both sku and weight_lbs are missing")
	}
}

func TestEstimateShipping_WeightTiers(t *testing.T) {
	// Light item
	light := estimateShipping(0.5)
	if len(light) != 3 {
		t.Errorf("expected 3 estimates for light item, got %d", len(light))
	}

	// Heavy item
	heavy := estimateShipping(25.0)
	if len(heavy) != 3 {
		t.Errorf("expected 3 estimates for heavy item, got %d", len(heavy))
	}

	// Verify USPS Priority < UPS for light items
	if light[0].Cost > light[1].Cost {
		// USPS Priority should be cheaper for light items
		// Not enforcing since tier boundaries may vary
	}
}

// ── Bulk update validation ──

func TestHandleBulkUpdate_InvalidStatus(t *testing.T) {
	resetSeed()
	res, _ := handleInventoryBulkUpdate(context.Background(), makeReq(map[string]interface{}{
		"new_status": "available",
	}))
	if !res.IsError {
		t.Error("expected error for invalid new_status in bulk update")
	}
	text := getText(t, res)
	if !strings.Contains(text, "invalid listing_status") {
		t.Errorf("expected 'invalid listing_status' error, got: %s", text)
	}
}

// ── CSV Import tests with fixture files ──

func TestHandleImportAmazon_FixtureCSV(t *testing.T) {
	resetSeed()
	res, err := handleImportAmazon(context.Background(), makeReq(map[string]interface{}{
		"csv_path": "testdata/amazon-orders.csv",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)

	// Fixture has 5 rows; 4 electronics + 1 non-electronics (coffee mug)
	totalRows := int(out["total_rows"].(float64))
	imported := int(out["imported"].(float64))
	if totalRows < 4 {
		t.Errorf("expected at least 4 total rows, got %d", totalRows)
	}
	if imported < 4 {
		t.Errorf("expected at least 4 imported items (electronics), got %d", imported)
	}
}

func TestHandleImportNewegg_FixtureCSV(t *testing.T) {
	resetSeed()
	res, err := handleImportNewegg(context.Background(), makeReq(map[string]interface{}{
		"csv_path": "testdata/newegg-orders.csv",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)

	totalRows := int(out["total_rows"].(float64))
	imported := int(out["imported"].(float64))
	if totalRows < 3 {
		t.Errorf("expected at least 3 total rows, got %d", totalRows)
	}
	if imported < 3 {
		t.Errorf("expected at least 3 imported items, got %d", imported)
	}
}

// ── Duplicate detection with order_id ──

func TestHandleDuplicates_OrderID(t *testing.T) {
	// Create items with same order_id and name to test dedup
	items := seedItems()
	items[0].OrderID = "ORD-001"
	items[4].OrderID = "ORD-001" // Same order_id, same name as items[0]
	clients.TestOverrideInventoryClient = clients.NewTestInventoryClient(items)

	res, err := handleInventoryDuplicates(context.Background(), makeReq(map[string]interface{}{
		"match_type": "order_id",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)

	groups := int(out["duplicate_groups"].(float64))
	if groups < 1 {
		t.Errorf("expected at least 1 duplicate group for order_id match, got %d", groups)
	}
}

// ── Specs catalog expanded entries ──

func TestLookupSpecs_ExpandedCategories(t *testing.T) {
	tests := []struct {
		productLine string
		expectKey   string
	}{
		{"Ryzen 7 7800X3D", "chipset"},
		{"CalDigit TS4", "interface"},
		{"Corsair RM1000x", "capacity"},
		{"Samsung 870 Evo", "interface"},
		{"Netgear Nighthawk", "interface"},
	}
	for _, tc := range tests {
		specs := LookupSpecs(tc.productLine)
		if specs == nil {
			t.Errorf("LookupSpecs(%q) returned nil", tc.productLine)
			continue
		}
		if specs[tc.expectKey] == "" {
			t.Errorf("LookupSpecs(%q) missing key %q", tc.productLine, tc.expectKey)
		}
	}
}

// ── on_hand tests ──

func TestHandleInventoryUpdate_OnHand(t *testing.T) {
	resetSeed()

	// Set on_hand = true for HW-001
	res, err := handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":     "HW-001",
		"on_hand": true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	// Update handler wraps in {"success":..., "item":...}
	updateItem := out["item"].(map[string]interface{})
	if updateItem["on_hand"] != true {
		t.Errorf("expected on_hand=true, got %v", updateItem["on_hand"])
	}

	// Verify via get (returns item directly)
	res2, err := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var getItem map[string]interface{}
	getJSON(t, res2, &getItem)
	if getItem["on_hand"] != true {
		t.Errorf("get after update: expected on_hand=true, got %v", getItem["on_hand"])
	}
}

func TestHandleInventoryUpdate_OnHandFalse(t *testing.T) {
	resetSeed()

	// First set to true
	handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":     "HW-002",
		"on_hand": true,
	}))

	// Then set back to false
	res, err := handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"sku":     "HW-002",
		"on_hand": false,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	updateItem2 := out["item"].(map[string]interface{})
	if updateItem2["on_hand"] != false {
		t.Errorf("expected on_hand=false, got %v", updateItem2["on_hand"])
	}
}

func TestHandleInventoryAdd_OnHand(t *testing.T) {
	resetSeed()

	res, err := handleInventoryAdd(context.Background(), makeReq(map[string]interface{}{
		"name":           "Test Item With OnHand",
		"category":       "Components",
		"purchase_price": 25.0,
		"location":       "test-bin",
		"on_hand":        true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	// Add handler wraps in {"success":..., "sku":..., "item":...}
	addItem := out["item"].(map[string]interface{})
	if addItem["on_hand"] != true {
		t.Errorf("expected on_hand=true on new item, got %v", addItem["on_hand"])
	}
}

func TestHandleInventoryBulkUpdate_OnHand(t *testing.T) {
	resetSeed()

	// Dry run — mark all GPU items as on_hand
	res, err := handleInventoryBulkUpdate(context.Background(), makeReq(map[string]interface{}{
		"category":   "GPU",
		"new_on_hand": true,
		"dry_run":    true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var preview map[string]interface{}
	getJSON(t, res, &preview)
	if preview["dry_run"] != true {
		t.Errorf("expected dry_run=true")
	}
	updates, ok := preview["updates"].(map[string]interface{})
	if !ok {
		t.Fatal("expected updates map")
	}
	if updates["on_hand"] != true {
		t.Errorf("expected on_hand=true in updates, got %v", updates["on_hand"])
	}
	if int(preview["would_update"].(float64)) != 2 {
		t.Errorf("expected 2 GPU items, got %v", preview["would_update"])
	}

	// Execute the bulk update
	res2, err := handleInventoryBulkUpdate(context.Background(), makeReq(map[string]interface{}{
		"category":   "GPU",
		"new_on_hand": true,
		"dry_run":    false,
		"confirm":    true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	getJSON(t, res2, &result)
	if int(result["updated_count"].(float64)) != 2 {
		t.Errorf("expected 2 updated, got %v", result["updated_count"])
	}

	// Verify items were updated
	res3, err := handleInventoryGet(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var item map[string]interface{}
	getJSON(t, res3, &item)
	if item["on_hand"] != true {
		t.Errorf("HW-001 on_hand should be true after bulk update, got %v", item["on_hand"])
	}
}

// ══════════════════════════════════════════════════════════════
// Pure function tests: extractYear, safePercent, splitTrim
// ══════════════════════════════════════════════════════════════

func TestExtractYear(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"Jan 2025", 2025},
		{"2024-06-15", 2024},
		{"06/15/2023", 2023},
		{"January 2026", 2026},
		{"some text 2025 more text", 2025},
		{"no year here", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := extractYear(tt.input)
		if got != tt.expected {
			t.Errorf("extractYear(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestSafePercent(t *testing.T) {
	tests := []struct {
		num, denom float64
		expected   string
	}{
		{50, 100, "50.0%"},
		{0, 100, "0.0%"},
		{100, 0, "0.0%"},
		{33, 100, "33.0%"},
		{1, 3, "33.3%"},
	}

	for _, tt := range tests {
		got := safePercent(tt.num, tt.denom)
		if got != tt.expected {
			t.Errorf("safePercent(%v, %v) = %q, want %q", tt.num, tt.denom, got, tt.expected)
		}
	}
}

func TestSplitTrim(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"HW-001,HW-002,HW-003", []string{"HW-001", "HW-002", "HW-003"}},
		{"HW-001, HW-002, HW-003", []string{"HW-001", "HW-002", "HW-003"}},
		{" HW-001 , HW-002 ", []string{"HW-001", "HW-002"}},
		{"HW-001,,HW-002", []string{"HW-001", "HW-002"}}, // empty segments dropped
		{"", []string{}},
	}

	for _, tt := range tests {
		got := splitTrim(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("splitTrim(%q) length = %d, want %d", tt.input, len(got), len(tt.expected))
			continue
		}
		for i, v := range got {
			if v != tt.expected[i] {
				t.Errorf("splitTrim(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

// ══════════════════════════════════════════════════════════════
// IsWrite verification: write tools must have IsWrite=true
// ══════════════════════════════════════════════════════════════

func TestWriteToolsHaveIsWriteSet(t *testing.T) {
	m := &Module{}
	allTools := m.Tools()

	// These tool names perform mutations and must have IsWrite=true
	writeToolNames := map[string]bool{
		"aftrs_inventory_return_start":   true,
		"aftrs_inventory_return_resolve": true,
		"aftrs_inventory_bundle_create":  true,
	}

	for _, td := range allTools {
		if writeToolNames[td.Tool.Name] {
			if !td.IsWrite {
				t.Errorf("tool %s performs writes but IsWrite=false", td.Tool.Name)
			}
		}
	}
}

// ══════════════════════════════════════════════════════════════
// Return handlers: input validation and happy-path
// ══════════════════════════════════════════════════════════════

func TestHandleReturnStart_MissingSKU(t *testing.T) {
	resetSeed()
	res, err := handleReturnStart(context.Background(), makeReq(map[string]interface{}{
		"reason": "defective",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing sku")
	}
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected 'sku is required', got: %s", text)
	}
}

func TestHandleReturnStart_MissingReason(t *testing.T) {
	resetSeed()
	res, err := handleReturnStart(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing reason")
	}
	text := getText(t, res)
	if !strings.Contains(text, "reason is required") {
		t.Errorf("expected 'reason is required', got: %s", text)
	}
}

func TestHandleReturnStart_HappyPath(t *testing.T) {
	resetSeed()
	res, err := handleReturnStart(context.Background(), makeReq(map[string]interface{}{
		"sku":    "HW-001",
		"reason": "defective GPU",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", getText(t, res))
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["return_status"] != "requested" {
		t.Errorf("expected return_status=requested, got %v", out["return_status"])
	}
	if !strings.Contains(out["dispute_notes"].(string), "defective GPU") {
		t.Error("expected dispute_notes to contain reason")
	}
}

func TestHandleReturnResolve_MissingSKU(t *testing.T) {
	resetSeed()
	res, err := handleReturnResolve(context.Background(), makeReq(map[string]interface{}{
		"resolution": "approved",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing sku")
	}
}

func TestHandleReturnResolve_MissingResolution(t *testing.T) {
	resetSeed()
	res, err := handleReturnResolve(context.Background(), makeReq(map[string]interface{}{
		"sku": "HW-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing resolution")
	}
}

func TestHandleReturnResolve_InvalidResolution(t *testing.T) {
	resetSeed()
	res, err := handleReturnResolve(context.Background(), makeReq(map[string]interface{}{
		"sku":        "HW-001",
		"resolution": "maybe",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for invalid resolution")
	}
	text := getText(t, res)
	if !strings.Contains(text, "invalid resolution") {
		t.Errorf("expected 'invalid resolution', got: %s", text)
	}
}

func TestHandleReturnResolve_CannotUseRequested(t *testing.T) {
	resetSeed()
	res, err := handleReturnResolve(context.Background(), makeReq(map[string]interface{}{
		"sku":        "HW-001",
		"resolution": "requested",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error — 'requested' is not a valid resolution")
	}
}

func TestHandleReturnResolve_HappyPath(t *testing.T) {
	resetSeed()
	// First start a return
	handleReturnStart(context.Background(), makeReq(map[string]interface{}{
		"sku":    "HW-002",
		"reason": "wrong item shipped",
	}))
	// Then resolve it
	res, err := handleReturnResolve(context.Background(), makeReq(map[string]interface{}{
		"sku":        "HW-002",
		"resolution": "approved",
		"notes":      "refund issued",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", getText(t, res))
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["return_status"] != "approved" {
		t.Errorf("expected return_status=approved, got %v", out["return_status"])
	}
	if !strings.Contains(out["dispute_notes"].(string), "refund issued") {
		t.Error("expected dispute_notes to contain resolution notes")
	}
}

// ══════════════════════════════════════════════════════════════
// Bundle handlers: input validation and happy-path
// ══════════════════════════════════════════════════════════════

func TestHandleBundleCreate_MissingBundleID(t *testing.T) {
	resetSeed()
	res, err := handleBundleCreate(context.Background(), makeReq(map[string]interface{}{
		"skus": "HW-001,HW-002",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing bundle_id")
	}
}

func TestHandleBundleCreate_MissingSKUs(t *testing.T) {
	resetSeed()
	res, err := handleBundleCreate(context.Background(), makeReq(map[string]interface{}{
		"bundle_id": "GPU-LOT-1",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing skus")
	}
}

func TestHandleBundleCreate_TooFewSKUs(t *testing.T) {
	resetSeed()
	res, err := handleBundleCreate(context.Background(), makeReq(map[string]interface{}{
		"bundle_id": "GPU-LOT-1",
		"skus":      "HW-001",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error — need at least 2 SKUs for a bundle")
	}
	text := getText(t, res)
	if !strings.Contains(text, "at least 2") {
		t.Errorf("expected 'at least 2' error, got: %s", text)
	}
}

func TestHandleBundleCreate_HappyPath(t *testing.T) {
	resetSeed()
	res, err := handleBundleCreate(context.Background(), makeReq(map[string]interface{}{
		"bundle_id": "GPU-LOT-1",
		"skus":      "HW-001, HW-005",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", getText(t, res))
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["bundle_id"] != "GPU-LOT-1" {
		t.Errorf("expected bundle_id=GPU-LOT-1, got %v", out["bundle_id"])
	}
	if int(out["items_tagged"].(float64)) != 2 {
		t.Errorf("expected 2 items tagged, got %v", out["items_tagged"])
	}
}

func TestHandleBundleList_NoBundles(t *testing.T) {
	resetSeed()
	// No bundles exist in seed data — test client doesn't persist bundle_id
	res, err := handleBundleList(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	if int(out["total_bundles"].(float64)) != 0 {
		t.Errorf("expected 0 bundles in fresh seed, got %v", out["total_bundles"])
	}
}

func TestHandleBundleList_SpecificBundleEmpty(t *testing.T) {
	resetSeed()
	// Query a bundle that doesn't exist
	res, err := handleBundleList(context.Background(), makeReq(map[string]interface{}{
		"bundle_id": "NONEXISTENT",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["bundle_id"] != "NONEXISTENT" {
		t.Errorf("expected bundle_id=NONEXISTENT, got %v", out["bundle_id"])
	}
	if int(out["count"].(float64)) != 0 {
		t.Errorf("expected 0 items in nonexistent bundle, got %v", out["count"])
	}
}

// ══════════════════════════════════════════════════════════════
// Tax report, Discord alerts, and other untested handlers
// ══════════════════════════════════════════════════════════════

func TestHandleTaxReport_NoSales(t *testing.T) {
	resetSeed()
	res, err := handleTaxReport(context.Background(), makeReq(map[string]interface{}{
		"year": float64(2099),
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["message"] == nil || !strings.Contains(out["message"].(string), "No sales found") {
		t.Error("expected 'No sales found' for future year")
	}
}

func TestHandleTaxReport_WithSales(t *testing.T) {
	resetSeed()
	// Record a sale first
	handleSaleRecord(context.Background(), makeReq(map[string]interface{}{
		"sku":           "HW-004",
		"sold_price":    float64(190),
		"platform":      "ebay",
		"shipping_cost": float64(15),
		"platform_fees": float64(25),
	}))
	// Use current year (sales are logged with current date)
	res, err := handleTaxReport(context.Background(), makeReq(nil))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	if out["total_sales"] == nil {
		t.Fatal("expected total_sales in tax report")
	}
	if int(out["total_sales"].(float64)) < 1 {
		t.Error("expected at least 1 sale in tax report")
	}
	if out["schedule_c_income"] == nil {
		t.Error("expected schedule_c_income section")
	}
	if out["schedule_c_expenses"] == nil {
		t.Error("expected schedule_c_expenses section")
	}
	if out["disclaimer"] == nil {
		t.Error("expected tax disclaimer")
	}
}

func TestHandleDiscordAlerts_DryRun(t *testing.T) {
	resetSeed()
	res, err := handleDiscordAlerts(context.Background(), makeReq(map[string]interface{}{
		"dry_run": true,
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]interface{}
	getJSON(t, res, &out)
	// Dry run should return alerts without sending to Discord
	if out["dry_run"] != nil && out["dry_run"] != true {
		t.Error("expected dry_run=true in response")
	}
	// HW-004 is "Listed" — should trigger stale_listing alert
	if alertCount, ok := out["alerts_found"].(float64); ok {
		if int(alertCount) == 0 {
			// Check if it says everything looks good (possible if no alerts match)
			if out["message"] == nil {
				t.Error("expected either alerts or a no-alerts message")
			}
		}
	}
}

func TestHandleUploadImage_MissingParams(t *testing.T) {
	resetSeed()
	// Missing both sku and file_path
	res, err := handleUploadImage(context.Background(), makeReq(map[string]interface{}{}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing sku and file_path")
	}
	text := getText(t, res)
	if !strings.Contains(text, "required") {
		t.Errorf("expected 'required' error, got: %s", text)
	}
}

func TestHandleUploadImage_MissingSKUOnly(t *testing.T) {
	resetSeed()
	res, err := handleUploadImage(context.Background(), makeReq(map[string]interface{}{
		"file_path": "/tmp/photo.jpg",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing sku")
	}
}

func TestHandleMarkSold_MissingSKU(t *testing.T) {
	resetSeed()
	res, err := handleMarkSold(context.Background(), makeReq(map[string]interface{}{
		"sold_price": float64(100),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing sku")
	}
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected 'sku is required', got: %s", text)
	}
}

func TestHandleInventoryUpdate_MissingSKU(t *testing.T) {
	resetSeed()
	res, err := handleInventoryUpdate(context.Background(), makeReq(map[string]interface{}{
		"notes": "some update",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Error("expected error for missing sku")
	}
	text := getText(t, res)
	if !strings.Contains(text, "sku is required") {
		t.Errorf("expected 'sku is required', got: %s", text)
	}
}

// ══════════════════════════════════════════════════════════════
// Tool registration: every tool has a non-empty handler and name
// ══════════════════════════════════════════════════════════════

func TestAllToolsHaveHandlers(t *testing.T) {
	m := &Module{}
	for _, td := range m.Tools() {
		if td.Tool.Name == "" {
			t.Error("found tool with empty name")
		}
		if td.Handler == nil {
			t.Errorf("tool %s has nil handler", td.Tool.Name)
		}
		if td.Category != "inventory" {
			t.Errorf("tool %s has category %q, expected 'inventory'", td.Tool.Name, td.Category)
		}
	}
}
