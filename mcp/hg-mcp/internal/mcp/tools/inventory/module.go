// Package inventory provides MCP tools for electronics inventory management.
package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// fbConditionMap maps internal condition values to Facebook Marketplace condition labels.
// Note: "poor" maps to "Fair" as FB Marketplace has no "Poor" option.
var fbConditionMap = map[string]string{
	"new": "New", "New": "New",
	"like_new": "Like New", "Like New": "Like New",
	"good": "Good", "Good": "Good", "Used": "Good",
	"fair": "Fair", "Fair": "Fair",
	"poor":      "Fair",
	"for_parts": "For parts",
	"Renewed":   "Like New",
}

// Module implements the inventory tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "inventory"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Electronics inventory management for tracking, listing, and selling items via eBay and Facebook Marketplace"
}

// Tools returns the inventory tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// CRUD Operations
		{
			Tool: mcp.NewTool("aftrs_inventory_list",
				mcp.WithDescription("List inventory items with optional filters (category, status, location, price range)"),
				mcp.WithString("category", mcp.Description("Filter by category (GPU, CPU, RAM, Storage, etc.)")),
				mcp.WithString("status", mcp.Description("Filter by listing status (not_listed, pending_review, listed, sold, keeping)"), mcp.Enum("not_listed", "pending_review", "listed", "sold", "keeping")),
				mcp.WithString("location", mcp.Description("Filter by physical location")),
				mcp.WithString("condition", mcp.Description("Filter by condition (new, like_new, good, fair, poor, for_parts)"), mcp.Enum("new", "like_new", "good", "fair", "poor", "for_parts")),
				mcp.WithNumber("min_price", mcp.Description("Minimum purchase price")),
				mcp.WithNumber("max_price", mcp.Description("Maximum purchase price")),
				mcp.WithString("query", mcp.Description("Text search across name, description, brand, model")),
				mcp.WithString("source", mcp.Description("Filter by purchase source (amazon, newegg, ebay, bestbuy, microcenter, bhphoto, other)"), mcp.Enum("amazon", "newegg", "ebay", "bestbuy", "microcenter", "bhphoto", "other")),
				mcp.WithNumber("limit", mcp.Description("Maximum items to return (default: 50)")),
			),
			Handler:             handleInventoryList,
			Category:            "inventory",
			Subcategory:         "crud",
			Tags:                []string{"inventory", "list", "search", "filter"},
			UseCases:            []string{"Browse inventory", "Find items to sell", "Search by category"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
			OutputSchema: tools.ObjectOutputSchema(map[string]interface{}{
				"count": map[string]interface{}{
					"type":        "integer",
					"description": "Number of items returned",
				},
				"items": map[string]interface{}{
					"type":        "array",
					"description": "List of inventory items",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"sku":            map[string]interface{}{"type": "string", "description": "Item SKU (e.g., HW-001)"},
							"name":           map[string]interface{}{"type": "string", "description": "Item name/title"},
							"category":       map[string]interface{}{"type": "string", "description": "Item category (GPU, CPU, RAM, etc.)"},
							"condition":      map[string]interface{}{"type": "string", "description": "Item condition"},
							"quantity":       map[string]interface{}{"type": "integer", "description": "Quantity available"},
							"purchase_price": map[string]interface{}{"type": "number", "description": "Original purchase price"},
							"asking_price":   map[string]interface{}{"type": "number", "description": "Recommended selling price"},
							"listing_status": map[string]interface{}{"type": "string", "description": "Listing status (not_listed, listed, sold, keeping)"},
							"location":       map[string]interface{}{"type": "string", "description": "Physical storage location"},
							"on_hand":        map[string]interface{}{"type": "boolean", "description": "Whether item is physically located and ready"},
						},
					},
				},
			}, []string{"count", "items"}),
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_get",
				mcp.WithDescription("Get detailed information about a specific inventory item by SKU"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU (e.g., GPU-ABC12345)")),
			),
			Handler:             handleInventoryGet,
			Category:            "inventory",
			Subcategory:         "crud",
			Tags:                []string{"inventory", "get", "details", "sku"},
			UseCases:            []string{"View item details", "Check item status", "Get listing info"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
			OutputSchema: tools.ObjectOutputSchema(map[string]interface{}{
				"sku":            map[string]interface{}{"type": "string", "description": "Item SKU (e.g., HW-001)"},
				"name":           map[string]interface{}{"type": "string", "description": "Item name/title"},
				"category":       map[string]interface{}{"type": "string", "description": "Item category"},
				"model":          map[string]interface{}{"type": "string", "description": "Model number"},
				"brand":          map[string]interface{}{"type": "string", "description": "Brand/manufacturer"},
				"condition":      map[string]interface{}{"type": "string", "description": "Item condition"},
				"quantity":       map[string]interface{}{"type": "integer", "description": "Quantity"},
				"purchase_price": map[string]interface{}{"type": "number", "description": "Original purchase price"},
				"purchase_date":  map[string]interface{}{"type": "string", "description": "Purchase date (e.g., Apr 2025)"},
				"asking_price":   map[string]interface{}{"type": "number", "description": "Recommended selling price"},
				"listing_status": map[string]interface{}{"type": "string", "description": "Listing status"},
				"location":       map[string]interface{}{"type": "string", "description": "Physical storage location"},
				"notes":          map[string]interface{}{"type": "string", "description": "Additional notes"},
				"on_hand":        map[string]interface{}{"type": "boolean", "description": "Whether item is physically located and ready"},
				"product_line":   map[string]interface{}{"type": "string", "description": "Product line grouping"},
				"specs":          map[string]interface{}{"type": "object", "description": "Technical specifications key-value pairs"},
				"vendor_name":    map[string]interface{}{"type": "string", "description": "Vendor display name"},
				"serial_number":  map[string]interface{}{"type": "string", "description": "Serial number"},
			}, []string{"sku", "name", "category"}),
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_add",
				mcp.WithDescription("Add a new item to the inventory"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Item name/title")),
				mcp.WithString("category", mcp.Required(), mcp.Description("Category (GPU, CPU, RAM, Storage, Motherboard, PSU, Case, Cooling, Peripherals, Networking, Thunderbolt, KVM, Cables, Components, Complete Systems, Other)")),
				mcp.WithNumber("purchase_price", mcp.Required(), mcp.Description("Original purchase price")),
				mcp.WithString("condition", mcp.Description("Condition (new, like_new, good, fair, poor, for_parts)"), mcp.Enum("new", "like_new", "good", "fair", "poor", "for_parts")),
				mcp.WithString("location", mcp.Required(), mcp.Description("Physical storage location (e.g., studio-rack-1, storage-bin-A)")),
				mcp.WithString("description", mcp.Description("Item description")),
				mcp.WithString("brand", mcp.Description("Brand/manufacturer")),
				mcp.WithString("model", mcp.Description("Model number")),
				mcp.WithString("serial_number", mcp.Description("Serial number if available")),
				mcp.WithString("subcategory", mcp.Description("Subcategory within main category")),
				mcp.WithNumber("quantity", mcp.Description("Quantity (default: 1)")),
				mcp.WithString("purchase_source", mcp.Description("Purchase source (amazon, newegg, ebay, bestbuy, microcenter, bhphoto, other)"), mcp.Enum("amazon", "newegg", "ebay", "bestbuy", "microcenter", "bhphoto", "other")),
				mcp.WithString("vendor_name", mcp.Description("Vendor display name (e.g. 'Micro Center - Dallas', 'eBay seller: techdeals99')")),
				mcp.WithString("vendor_url", mcp.Description("Vendor website or order page URL")),
				mcp.WithString("order_id", mcp.Description("Original order ID")),
				mcp.WithArray("tags", mcp.Description("Tags for organization (sell, keep, gift, parts)"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithString("notes", mcp.Description("Additional notes")),
				mcp.WithString("product_line", mcp.Description("Product line grouping (e.g., 'Samsung 990 Pro', 'RTX 4090')")),
				mcp.WithObject("specs", mcp.Description("Technical specs as key-value pairs (e.g., {\"vram\": \"24GB GDDR6X\", \"tdp\": \"450W\"})"), mcp.AdditionalProperties(map[string]any{"type": "string"})),
			),
			Handler:             handleInventoryAdd,
			Category:            "inventory",
			Subcategory:         "crud",
			Tags:                []string{"inventory", "add", "create", "new"},
			UseCases:            []string{"Add new item", "Manual entry", "Record purchase"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_update",
				mcp.WithDescription("Update an existing inventory item"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU to update")),
				mcp.WithString("name", mcp.Description("New item name")),
				mcp.WithString("description", mcp.Description("New description")),
				mcp.WithString("condition", mcp.Description("Updated condition"), mcp.Enum("new", "like_new", "good", "fair", "poor", "for_parts")),
				mcp.WithString("location", mcp.Description("New location")),
				mcp.WithNumber("current_value", mcp.Description("Current estimated value")),
				mcp.WithString("listing_status", mcp.Description("Updated listing status"), mcp.Enum("not_listed", "pending_review", "listed", "sold", "keeping")),
				mcp.WithNumber("listed_price", mcp.Description("Listing price")),
				mcp.WithString("notes", mcp.Description("Updated notes")),
				mcp.WithArray("tags", mcp.Description("Updated tags"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithString("product_line", mcp.Description("Product line grouping (e.g., 'Samsung 990 Pro', 'RTX 4090')")),
				mcp.WithObject("specs", mcp.Description("Technical specs as key-value pairs (e.g., {\"vram\": \"24GB GDDR6X\", \"tdp\": \"450W\"})"), mcp.AdditionalProperties(map[string]any{"type": "string"})),
				mcp.WithBoolean("on_hand", mcp.Description("Whether the item is physically located and on-hand ready to list")),
			),
			Handler:             handleInventoryUpdate,
			Category:            "inventory",
			Subcategory:         "crud",
			Tags:                []string{"inventory", "update", "edit", "modify"},
			UseCases:            []string{"Update item info", "Change location", "Adjust pricing"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_delete",
				mcp.WithDescription("Remove an item from the inventory"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU to delete")),
				mcp.WithBoolean("confirm", mcp.Required(), mcp.Description("Confirm deletion (must be true)")),
			),
			Handler:             handleInventoryDelete,
			Category:            "inventory",
			Subcategory:         "crud",
			Tags:                []string{"inventory", "delete", "remove"},
			UseCases:            []string{"Remove sold item", "Clean up duplicates", "Delete mistakes"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_bulk_delete",
				mcp.WithDescription("Delete multiple inventory items matching a filter (use with caution)"),
				mcp.WithString("source", mcp.Description("Delete items from this purchase source (amazon, newegg, ebay, etc.)"), mcp.Enum("amazon", "newegg", "ebay", "bestbuy", "microcenter", "bhphoto", "other")),
				mcp.WithNumber("max_price", mcp.Description("Only delete items with purchase price at or below this value")),
				mcp.WithBoolean("dry_run", mcp.Description("If true, show items that would be deleted without actually deleting")),
				mcp.WithBoolean("confirm", mcp.Required(), mcp.Description("Confirm bulk deletion (must be true to actually delete)")),
			),
			Handler:             handleInventoryBulkDelete,
			Category:            "inventory",
			Subcategory:         "crud",
			Tags:                []string{"inventory", "delete", "bulk", "cleanup"},
			UseCases:            []string{"Delete failed imports", "Clean up test data", "Remove items by source"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_search",
				mcp.WithDescription("Full-text search across all inventory items"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query (searches name, description, brand, model)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 25)")),
			),
			Handler:             handleInventorySearch,
			Category:            "inventory",
			Subcategory:         "crud",
			Tags:                []string{"inventory", "search", "find", "query"},
			UseCases:            []string{"Find specific items", "Search by brand", "Locate products"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},

		// Import Operations
		{
			Tool: mcp.NewTool("aftrs_inventory_import_amazon",
				mcp.WithDescription("Import items from an Amazon order history CSV file"),
				mcp.WithString("csv_path", mcp.Required(), mcp.Description("Path to the Amazon order history CSV file")),
			),
			Handler:             handleImportAmazon,
			Category:            "inventory",
			Subcategory:         "import",
			Tags:                []string{"inventory", "import", "amazon", "csv"},
			UseCases:            []string{"Bulk import from Amazon", "Import order history"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_import_newegg",
				mcp.WithDescription("Import items from a Newegg order history CSV file"),
				mcp.WithString("csv_path", mcp.Required(), mcp.Description("Path to the Newegg order history CSV file")),
			),
			Handler:             handleImportNewegg,
			Category:            "inventory",
			Subcategory:         "import",
			Tags:                []string{"inventory", "import", "newegg", "csv"},
			UseCases:            []string{"Bulk import from Newegg", "Import order history"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_import_gmail",
				mcp.WithDescription("Import orders from Gmail order confirmation emails (Amazon, Newegg, eBay, Best Buy). Parses historical emails to extract purchase history."),
				mcp.WithNumber("since_days", mcp.Description("How many days back to search for orders (default: 365)")),
				mcp.WithArray("sources", mcp.Description("Specific sources to import: amazon, newegg, ebay, bestbuy (default: all)"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithNumber("max_results", mcp.Description("Maximum emails to process per source (default: 100)")),
				mcp.WithBoolean("dry_run", mcp.Description("If true, only parse and report what would be imported without actually importing")),
				mcp.WithBoolean("include_all", mcp.Description("If true, include all orders (not just electronics). Default: false (electronics only)")),
			),
			Handler:             handleImportGmail,
			Category:            "inventory",
			Subcategory:         "import",
			Tags:                []string{"inventory", "import", "gmail", "email", "amazon", "newegg", "ebay"},
			UseCases:            []string{"Bulk import from email history", "Import order confirmations", "Automated order tracking"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_find_receipt",
				mcp.WithDescription("Search Gmail for the original purchase confirmation email for an inventory item. Looks up the item by SKU, then searches Gmail for matching order emails by product name and order ID."),
				mcp.WithString("sku", mcp.Description("Item SKU (e.g., HW-001). If provided, searches Gmail using the item's name and notes.")),
				mcp.WithString("query", mcp.Description("Free-text search query. Use instead of SKU for manual searches (e.g., product name, order number).")),
			),
			Handler:             handleFindReceipt,
			Category:            "inventory",
			Subcategory:         "lookup",
			Tags:                []string{"inventory", "gmail", "receipt", "order", "search"},
			UseCases:            []string{"Find original purchase email", "Verify purchase price", "Get order number for returns"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},

		// Analytics & Summary
		{
			Tool: mcp.NewTool("aftrs_inventory_summary",
				mcp.WithDescription("Get inventory overview: counts by category, status, total value, recently added, and top value items"),
			),
			Handler:             handleInventorySummary,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "summary", "overview", "dashboard"},
			UseCases:            []string{"Inventory overview", "Value assessment", "Status check"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_categories",
				mcp.WithDescription("List all inventory categories with item counts"),
			),
			Handler:             handleInventoryCategories,
			Category:            "inventory",
			Subcategory:         "organization",
			Tags:                []string{"inventory", "categories", "list"},
			UseCases:            []string{"View categories", "Category counts"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_locations",
				mcp.WithDescription("List all storage locations with item counts"),
			),
			Handler:             handleInventoryLocations,
			Category:            "inventory",
			Subcategory:         "organization",
			Tags:                []string{"inventory", "locations", "storage"},
			UseCases:            []string{"View locations", "Find items by location"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_stale",
				mcp.WithDescription("Find items not listed or sold within specified days"),
				mcp.WithNumber("days", mcp.Description("Days since last update (default: 30)")),
			),
			Handler:             handleInventoryStale,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "stale", "aging", "unlisted"},
			UseCases:            []string{"Find forgotten items", "Prioritize listings", "Cleanup check"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},

		// Listing Operations
		{
			Tool: mcp.NewTool("aftrs_inventory_mark_sold",
				mcp.WithDescription("Mark an item as sold with sale details"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU")),
				mcp.WithNumber("sold_price", mcp.Required(), mcp.Description("Final sale price")),
				mcp.WithString("platform", mcp.Required(), mcp.Description("Sale platform (ebay, fb_marketplace, local, other)")),
			),
			Handler:             handleMarkSold,
			Category:            "inventory",
			Subcategory:         "listing",
			Tags:                []string{"inventory", "sold", "sale", "complete"},
			UseCases:            []string{"Record sale", "Mark as sold", "Track revenue"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_fb_content",
				mcp.WithDescription("Generate copy-paste ready content for Facebook Marketplace listing"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU to generate content for")),
				mcp.WithNumber("asking_price", mcp.Description("Asking price (optional, will suggest if not provided)")),
			),
			Handler:             handleFBContent,
			Category:            "inventory",
			Subcategory:         "listing",
			Tags:                []string{"inventory", "facebook", "marketplace", "listing"},
			UseCases:            []string{"Create FB listing", "Generate description", "Prepare for sale"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},

		// Image Operations
		{
			Tool: mcp.NewTool("aftrs_inventory_upload_image",
				mcp.WithDescription("Upload an image for an inventory item from a local file path"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU")),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Local path to image file")),
				mcp.WithBoolean("set_primary", mcp.Description("Set as primary image (default: false)")),
			),
			Handler:             handleUploadImage,
			Category:            "inventory",
			Subcategory:         "images",
			Tags:                []string{"inventory", "image", "upload", "photo"},
			UseCases:            []string{"Add product photos", "Upload images", "Set primary photo"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_list_images",
				mcp.WithDescription("List all images for an inventory item"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU")),
			),
			Handler:             handleListImages,
			Category:            "inventory",
			Subcategory:         "images",
			Tags:                []string{"inventory", "images", "list", "photos"},
			UseCases:            []string{"View item photos", "Check image count"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},

		// Consolidated Dashboard
		{
			Tool: mcp.NewTool("aftrs_inventory_dashboard",
				mcp.WithDescription("Full inventory dashboard: summary, recent activity, items needing attention, and quick stats"),
			),
			Handler:             handleInventoryDashboard,
			Category:            "inventory",
			Subcategory:         "consolidated",
			Tags:                []string{"inventory", "dashboard", "overview", "consolidated"},
			UseCases:            []string{"Morning check", "Full status", "Quick overview"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},

		// eBay Integration
		{
			Tool: mcp.NewTool("aftrs_inventory_ebay_price_research",
				mcp.WithDescription("Research prices on eBay by searching completed/sold listings. Returns price analysis with min, max, average, median, and suggested price range."),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query (e.g., 'RTX 4090 Founders Edition')")),
				mcp.WithNumber("limit", mcp.Description("Maximum results to analyze (default: 25, max: 100)")),
			),
			Handler:             handleEbayPriceResearch,
			Category:            "inventory",
			Subcategory:         "ebay",
			Tags:                []string{"inventory", "ebay", "pricing", "research"},
			UseCases:            []string{"Research selling prices", "Set competitive prices", "Market analysis"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_ebay_active_listings",
				mcp.WithDescription("Search currently active eBay listings for market comparison"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 25, max: 100)")),
			),
			Handler:             handleEbayActiveListings,
			Category:            "inventory",
			Subcategory:         "ebay",
			Tags:                []string{"inventory", "ebay", "search", "listings"},
			UseCases:            []string{"Check competition", "See active listings", "Market comparison"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_ebay_categories",
				mcp.WithDescription("Get eBay category suggestions for a product"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Product name or description to find categories for")),
			),
			Handler:             handleEbayCategories,
			Category:            "inventory",
			Subcategory:         "ebay",
			Tags:                []string{"inventory", "ebay", "categories"},
			UseCases:            []string{"Find listing category", "Category lookup"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_ebay_create_listing",
				mcp.WithDescription("Create an eBay listing from an inventory item (requires OAuth refresh token)"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Inventory item SKU to list")),
				mcp.WithNumber("price", mcp.Required(), mcp.Description("Listing price in USD")),
				mcp.WithString("category_id", mcp.Required(), mcp.Description("eBay category ID (use ebay_categories to find)")),
				mcp.WithNumber("quantity", mcp.Description("Quantity to list (default: 1)")),
				mcp.WithString("condition", mcp.Description("eBay condition ID (NEW, LIKE_NEW, VERY_GOOD, GOOD, ACCEPTABLE)"), mcp.Enum("NEW", "LIKE_NEW", "VERY_GOOD", "GOOD", "ACCEPTABLE")),
				mcp.WithBoolean("sandbox", mcp.Description("Use sandbox environment for testing (default: false)")),
			),
			Handler:             handleEbayCreateListing,
			Category:            "inventory",
			Subcategory:         "ebay",
			Tags:                []string{"inventory", "ebay", "listing", "sell"},
			UseCases:            []string{"List item on eBay", "Create auction", "Sell item"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		// Analytics Tools
		{
			Tool: mcp.NewTool("aftrs_inventory_value",
				mcp.WithDescription("Calculate total inventory value with breakdown by category, status, and location"),
				mcp.WithBoolean("include_sold", mcp.Description("Include sold items in calculations (default: false)")),
				mcp.WithString("group_by", mcp.Description("Group value by field (category, status, location, condition)"), mcp.Enum("category", "status", "location", "condition")),
			),
			Handler:             handleInventoryValue,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "value", "analytics", "report"},
			UseCases:            []string{"Calculate inventory value", "Financial reporting", "Value breakdown"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_export",
				mcp.WithDescription("Export inventory to CSV or JSON format"),
				mcp.WithString("format", mcp.Description("Export format (csv, json)"), mcp.Enum("csv", "json")),
				mcp.WithString("status", mcp.Description("Filter by listing status"), mcp.Enum("not_listed", "pending_review", "listed", "sold", "keeping")),
				mcp.WithString("category", mcp.Description("Filter by category")),
			),
			Handler:             handleInventoryExport,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "export", "csv", "json", "backup"},
			UseCases:            []string{"Export inventory", "Backup data", "Spreadsheet import"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		// Data Quality Tools
		{
			Tool: mcp.NewTool("aftrs_inventory_recategorize",
				mcp.WithDescription("Re-analyze and update categories for inventory items based on product names. Uses expanded keyword matching to improve categorization, extract brands, and set subcategories."),
				mcp.WithString("sku", mcp.Description("Single item SKU to recategorize (optional - if not provided, uses filter)")),
				mcp.WithString("category", mcp.Description("Only recategorize items in this category (e.g., 'Other' to fix uncategorized items)")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview changes without applying (default: true)")),
				mcp.WithNumber("limit", mcp.Description("Maximum items to process (default: 100)")),
			),
			Handler:             handleInventoryRecategorize,
			Category:            "inventory",
			Subcategory:         "data_quality",
			Tags:                []string{"inventory", "recategorize", "category", "brand", "cleanup"},
			UseCases:            []string{"Fix Other category items", "Extract brands from names", "Improve categorization"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_duplicates",
				mcp.WithDescription("Find potential duplicate items by ASIN, name similarity, or order ID"),
				mcp.WithString("match_type", mcp.Description("Type of duplicate matching: asin (exact ASIN match), name (exact name match), order_id (same order + item name), all (check all three)"), mcp.Enum("asin", "name", "order_id", "all")),
			),
			Handler:             handleInventoryDuplicates,
			Category:            "inventory",
			Subcategory:         "data_quality",
			Tags:                []string{"inventory", "duplicates", "cleanup", "data_quality"},
			UseCases:            []string{"Find duplicate imports", "Clean up data", "Merge duplicates"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_bulk_update",
				mcp.WithDescription("Update multiple inventory items at once based on filter criteria"),
				mcp.WithString("category", mcp.Description("Filter: only update items in this category")),
				mcp.WithString("status", mcp.Description("Filter: only update items with this listing status"), mcp.Enum("not_listed", "pending_review", "listed", "sold", "keeping")),
				mcp.WithString("source", mcp.Description("Filter: only update items from this purchase source"), mcp.Enum("amazon", "newegg", "ebay", "bestbuy", "microcenter", "bhphoto", "other")),
				mcp.WithString("new_status", mcp.Description("Update: set listing status to this value"), mcp.Enum("not_listed", "pending_review", "listed", "sold", "keeping")),
				mcp.WithString("new_location", mcp.Description("Update: move items to this location")),
				mcp.WithArray("add_tags", mcp.Description("Update: add these tags to items"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithArray("remove_tags", mcp.Description("Update: remove these tags from items"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithBoolean("new_on_hand", mcp.Description("Update: set on_hand status (true = physically located and ready to list)")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview changes without applying (default: true)")),
				mcp.WithBoolean("confirm", mcp.Description("Confirm bulk update (must be true to apply changes when dry_run is false)")),
				mcp.WithNumber("limit", mcp.Description("Maximum items to update (default: 100, max: 500)")),
			),
			Handler:             handleInventoryBulkUpdate,
			Category:            "inventory",
			Subcategory:         "data_quality",
			Tags:                []string{"inventory", "bulk", "update", "batch"},
			UseCases:            []string{"Update status for multiple items", "Move items in bulk", "Batch tag updates"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},

		// Organization Tools
		{
			Tool: mcp.NewTool("aftrs_inventory_move",
				mcp.WithDescription("Move an item to a different storage location"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU to move")),
				mcp.WithString("location", mcp.Required(), mcp.Description("New storage location (e.g., studio-rack-1, storage-bin-A)")),
			),
			Handler:             handleInventoryMove,
			Category:            "inventory",
			Subcategory:         "organization",
			Tags:                []string{"inventory", "move", "location", "organize"},
			UseCases:            []string{"Move item to new location", "Reorganize storage"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		// Image Management Tools
		{
			Tool: mcp.NewTool("aftrs_inventory_delete_image",
				mcp.WithDescription("Delete an image from an inventory item"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU")),
				mcp.WithString("image_key", mcp.Required(), mcp.Description("Image key/filename to delete (from list_images)")),
			),
			Handler:             handleDeleteImage,
			Category:            "inventory",
			Subcategory:         "images",
			Tags:                []string{"inventory", "image", "delete", "photo"},
			UseCases:            []string{"Remove unwanted image", "Clean up photos"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_set_primary_image",
				mcp.WithDescription("Set the primary image for an inventory item (used as thumbnail and first listing image)"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU")),
				mcp.WithString("image_key", mcp.Required(), mcp.Description("Image key/filename to set as primary")),
			),
			Handler:             handleSetPrimaryImage,
			Category:            "inventory",
			Subcategory:         "images",
			Tags:                []string{"inventory", "image", "primary", "thumbnail"},
			UseCases:            []string{"Set main product photo", "Change primary image"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},

		// ── New tools (ported from hw-resale) ──

		{
			Tool: mcp.NewTool("aftrs_inventory_listing_generate",
				mcp.WithDescription("Generate marketplace listing text for fb_marketplace, ebay, or hardwareswap using templates"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU (e.g., HW-001)")),
				mcp.WithString("platform", mcp.Description("Target platform: fb_marketplace, ebay, hardwareswap"), mcp.Enum("fb_marketplace", "ebay", "hardwareswap")),
			),
			Handler:             handleListingGenerate,
			Category:            "inventory",
			Subcategory:         "listing",
			Tags:                []string{"inventory", "listing", "template", "generate"},
			UseCases:            []string{"Create marketplace post", "Generate eBay listing", "Reddit hardwareswap post"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_smart_collect",
				mcp.WithDescription("Run smartctl on a storage device and store parsed SMART data on the inventory item"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU of the storage device")),
				mcp.WithString("device_path", mcp.Description("Linux device path (e.g., /dev/sda, /dev/nvme0n1). Default: /dev/sda")),
			),
			Handler:             handleSmartCollect,
			Category:            "inventory",
			Subcategory:         "data_quality",
			Tags:                []string{"inventory", "smart", "storage", "health"},
			UseCases:            []string{"Check drive health", "Add SMART data to listing", "Verify storage device"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_sale_record",
				mcp.WithDescription("Record a completed sale with full P&L (revenue, cost, fees, profit) to Sales tab"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU that was sold")),
				mcp.WithNumber("sold_price", mcp.Required(), mcp.Description("Actual sale price per unit")),
				mcp.WithString("platform", mcp.Required(), mcp.Description("Where it sold (ebay, fb_marketplace, hardwareswap, local)")),
				mcp.WithNumber("quantity_sold", mcp.Description("Units sold (default: 1)")),
				mcp.WithNumber("shipping_cost", mcp.Description("Shipping cost in USD (default: 0)")),
				mcp.WithNumber("platform_fees", mcp.Description("Platform/payment fees in USD (default: 0)")),
				mcp.WithString("buyer_info", mcp.Description("Optional buyer identifier/notes")),
				mcp.WithString("notes", mcp.Description("Sale notes")),
			),
			Handler:             handleSaleRecord,
			Category:            "inventory",
			Subcategory:         "listing",
			Tags:                []string{"inventory", "sale", "record", "pnl"},
			UseCases:            []string{"Record a sale", "Track P&L", "Log revenue"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_price_check",
				mcp.WithDescription("Margin analysis: asking vs purchase price, margin %, pricing guidance, and eBay search suggestion"),
				mcp.WithString("sku", mcp.Description("Item SKU to analyze (required if search_query not provided)")),
				mcp.WithString("search_query", mcp.Description("Free-text search to find the item (alternative to SKU)")),
			),
			Handler:             handlePriceCheck,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "price", "margin", "analysis"},
			UseCases:            []string{"Check profit margin", "Pricing guidance", "Revenue analysis"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_sales_log",
				mcp.WithDescription("View sales history with P&L summary from the Sales tab"),
			),
			Handler:             handleSalesLog,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "sales", "log", "history"},
			UseCases:            []string{"View sales history", "Check P&L", "Revenue tracking"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},

		{
			Tool: mcp.NewTool("aftrs_inventory_tax_report",
				mcp.WithDescription("Generate tax preparation data from sales log. Produces Schedule C (Profit or Loss from Business) line items: gross receipts, COGS, expenses by category (shipping, platform fees), and net profit. Filterable by tax year."),
				mcp.WithNumber("year", mcp.Description("Tax year to report on (default: current year)")),
			),
			Handler:             handleTaxReport,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "tax", "schedule-c", "report", "irs"},
			UseCases:            []string{"Tax preparation", "Schedule C data", "Annual P&L summary"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},

		{
			Tool: mcp.NewTool("aftrs_inventory_discord_alerts",
				mcp.WithDescription("Scan inventory for alert conditions and send Discord notifications. Checks for: items below floor price (asking < 50% of purchase), stale listings (listed >30 days without sale), and low-stock items. Sends formatted alerts to the configured Discord channel."),
				mcp.WithString("channel_id", mcp.Description("Discord channel ID to send alerts to (uses default if not specified)")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview alerts without sending to Discord (default: false)")),
			),
			Handler:             handleDiscordAlerts,
			Category:            "inventory",
			Subcategory:         "alerts",
			Tags:                []string{"inventory", "discord", "alerts", "notifications"},
			UseCases:            []string{"Daily inventory health check", "Price floor monitoring", "Stale listing detection"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},

		// ── Phase 2+4 tools ──

		{
			Tool: mcp.NewTool("aftrs_inventory_import_json",
				mcp.WithDescription("Import inventory from a JSON file (supports Python hw-resale format or generic [{name, category, ...}] arrays)"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the JSON inventory file")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview import without writing (default: true)")),
				mcp.WithString("source", mcp.Description("Import source label (default: json_import)")),
			),
			Handler:             handleImportJSON,
			Category:            "inventory",
			Subcategory:         "import",
			Tags:                []string{"inventory", "import", "json", "migrate"},
			UseCases:            []string{"Import from Python prototype", "Bulk JSON import", "Migrate inventory"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_batch_listing",
				mcp.WithDescription("Generate marketplace listings for all items matching a filter. Returns listing text for each item × platform."),
				mcp.WithString("category", mcp.Description("Filter by category (e.g., GPU, Storage)")),
				mcp.WithString("status", mcp.Description("Filter by listing status (default: not_listed)"), mcp.Enum("not_listed", "pending_review", "listed", "sold", "keeping")),
				mcp.WithString("platform", mcp.Description("Target platform (default: all three)"), mcp.Enum("fb_marketplace", "ebay", "hardwareswap", "all")),
				mcp.WithNumber("limit", mcp.Description("Maximum items to process (default: 50)")),
			),
			Handler:             handleBatchListing,
			Category:            "inventory",
			Subcategory:         "listing",
			Tags:                []string{"inventory", "listing", "batch", "bulk"},
			UseCases:            []string{"Generate all GPU listings", "Bulk listing creation", "Prepare listings for sale"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_quick_list",
				mcp.WithDescription("One-command listing workflow: fetch item, compute margin analysis, and generate listing text for all 3 platforms"),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU")),
			),
			Handler:             handleQuickList,
			Category:            "inventory",
			Subcategory:         "listing",
			Tags:                []string{"inventory", "listing", "quick", "workflow"},
			UseCases:            []string{"One-command listing", "Price check + listing in one call"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_shipping_estimate",
				mcp.WithDescription("Estimate shipping costs by carrier (USPS Priority, UPS Ground, USPS Ground) based on item category weight or manual weight/dimensions"),
				mcp.WithString("sku", mcp.Description("Item SKU (uses category for weight estimate)")),
				mcp.WithNumber("weight_lbs", mcp.Description("Manual weight in pounds (overrides category estimate)")),
			),
			Handler:             handleShippingEstimate,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "shipping", "estimate", "cost"},
			UseCases:            []string{"Estimate shipping cost", "P&L planning", "Set shipped price"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_reprice",
				mcp.WithDescription("Suggest repricing based on eBay completed listings. Computes fast-sale (×0.95 median) and max-profit (×1.05 median) suggestions."),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU to reprice")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview price change without applying (default: true)")),
				mcp.WithString("strategy", mcp.Description("Pricing strategy: fast_sale (×0.95 median), max_profit (×1.05 median), median (exact median)"), mcp.Enum("fast_sale", "max_profit", "median")),
			),
			Handler:             handleReprice,
			Category:            "inventory",
			Subcategory:         "analytics",
			Tags:                []string{"inventory", "reprice", "ebay", "pricing"},
			UseCases:            []string{"Auto-reprice from market data", "Competitive pricing", "Price optimization"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
		},
		// Photo Pipeline
		{
			Tool: mcp.NewTool("aftrs_inventory_photo_sync",
				mcp.WithDescription("Sync product photos from a Google Drive folder into local images/{sku}/ directories. Files must follow naming convention: HW-001-front.jpg or HW-001_sticker.png"),
				mcp.WithString("folder_id", mcp.Description("Google Drive folder ID (defaults to INVENTORY_PHOTOS_DRIVE_FOLDER_ID env var)")),
			),
			Handler:             handlePhotoSync,
			Category:            "inventory",
			Subcategory:         "photos",
			Tags:                []string{"inventory", "photos", "sync", "drive"},
			UseCases:            []string{"Sync product photos from phone", "Download photos from Drive"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
			Timeout:             300000000000, // 5 minutes
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_photo_ingest",
				mcp.WithDescription("Download photos from a Google Drive folder, run OCR via tesseract, extract model numbers, and match against inventory and specs catalog"),
				mcp.WithString("folder_id", mcp.Description("Google Drive folder ID (defaults to INVENTORY_PHOTOS_DRIVE_FOLDER_ID env var)")),
				mcp.WithNumber("limit", mcp.Description("Maximum photos to process (default: 10)")),
			),
			Handler:             handlePhotoIngest,
			Category:            "inventory",
			Subcategory:         "photos",
			Tags:                []string{"inventory", "photos", "ocr", "ingest", "drive"},
			UseCases:            []string{"OCR product stickers from photos", "Auto-identify hardware from photos"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "inventory",
			Timeout:             600000000000, // 10 minutes
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_photo_scan",
				mcp.WithDescription("Run OCR on a single local image file, extract model numbers, and match against inventory and specs catalog. Useful for testing OCR on individual photos without Drive."),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the image file to scan")),
			),
			Handler:             handlePhotoScan,
			Category:            "inventory",
			Subcategory:         "photos",
			Tags:                []string{"inventory", "photos", "ocr", "scan"},
			UseCases:            []string{"Test OCR on a single photo", "Identify hardware from a sticker image"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_return_start",
				mcp.WithDescription("Initiate a return or dispute for an inventory item. Sets status to 'requested' and logs the reason."),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU")),
				mcp.WithString("reason", mcp.Required(), mcp.Description("Reason for return/dispute")),
			),
			Handler:             handleReturnStart,
			Category:            "inventory",
			Subcategory:         "returns",
			Tags:                []string{"inventory", "return", "dispute", "start"},
			UseCases:            []string{"Initiate a return", "Start a dispute"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_return_resolve",
				mcp.WithDescription("Resolve a return/dispute (approve, complete, or deny). Updates status and adds resolution notes."),
				mcp.WithString("sku", mcp.Required(), mcp.Description("Item SKU")),
				mcp.WithString("resolution", mcp.Required(), mcp.Description("Resolution: approved, completed, or denied")),
				mcp.WithString("notes", mcp.Description("Resolution notes (optional)")),
			),
			Handler:             handleReturnResolve,
			Category:            "inventory",
			Subcategory:         "returns",
			Tags:                []string{"inventory", "return", "dispute", "resolve"},
			UseCases:            []string{"Approve a return", "Deny a dispute", "Complete a return"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_bundle_create",
				mcp.WithDescription("Create a bundle by assigning a bundle ID to multiple items. Groups items for sale as a lot."),
				mcp.WithString("bundle_id", mcp.Required(), mcp.Description("Bundle identifier (e.g., 'GPU-LOT-1')")),
				mcp.WithString("skus", mcp.Required(), mcp.Description("Comma-separated SKUs to include in the bundle")),
			),
			Handler:             handleBundleCreate,
			Category:            "inventory",
			Subcategory:         "bundles",
			Tags:                []string{"inventory", "bundle", "lot", "group"},
			UseCases:            []string{"Group items for bundle sale", "Create a lot listing"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "inventory",
		},
		{
			Tool: mcp.NewTool("aftrs_inventory_bundle_list",
				mcp.WithDescription("List all bundles with item counts, or list items in a specific bundle."),
				mcp.WithString("bundle_id", mcp.Description("Bundle ID to list items for (omit to list all bundles)")),
			),
			Handler:             handleBundleList,
			Category:            "inventory",
			Subcategory:         "bundles",
			Tags:                []string{"inventory", "bundle", "lot", "list"},
			UseCases:            []string{"View bundle contents", "List all bundles"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "inventory",
		},
	}
}

// ── Validation ──

var validConditions = map[string]bool{
	"new": true, "like_new": true, "good": true,
	"fair": true, "poor": true, "for_parts": true,
}

var validStatuses = map[string]bool{
	"not_listed": true, "pending_review": true, "listed": true,
	"sold": true, "keeping": true,
}

func validateCondition(v string) error {
	if v == "" {
		return nil
	}
	normalized := strings.ToLower(v)
	if !validConditions[normalized] {
		return fmt.Errorf("invalid condition %q (valid: new, like_new, good, fair, poor, for_parts)", v)
	}
	return nil
}

func validateListingStatus(v string) error {
	if v == "" {
		return nil
	}
	normalized := strings.ToLower(v)
	if !validStatuses[normalized] {
		return fmt.Errorf("invalid listing_status %q (valid: not_listed, pending_review, listed, sold, keeping)", v)
	}
	return nil
}

// ── Shipping estimation ──

// shippingWeights maps categories to estimated weight in pounds.
var shippingWeights = map[string]float64{
	"GPU":              4.0,
	"CPU":              1.0,
	"RAM":              0.5,
	"Storage":          0.5,
	"NVMe Gen5":        0.5,
	"Motherboard":      3.5,
	"PSU":              6.0,
	"Case":             20.0,
	"Cooling":          3.0,
	"Peripherals":      2.0,
	"Networking":       3.0,
	"Thunderbolt":      1.5,
	"KVM":              2.0,
	"Cables":           1.0,
	"Components":       1.5,
	"Complete Systems": 25.0,
}

type shippingEstimate struct {
	Carrier string  `json:"carrier"`
	Service string  `json:"service"`
	Cost    float64 `json:"cost"`
	Days    string  `json:"days"`
}

func estimateShipping(weightLbs float64) []shippingEstimate {
	estimates := []shippingEstimate{}

	// USPS Priority Mail (simplified tiers)
	var uspsCost float64
	switch {
	case weightLbs <= 1:
		uspsCost = 9.50
	case weightLbs <= 3:
		uspsCost = 13.00
	case weightLbs <= 5:
		uspsCost = 17.00
	case weightLbs <= 10:
		uspsCost = 22.00
	case weightLbs <= 20:
		uspsCost = 30.00
	default:
		uspsCost = 45.00
	}

	// UPS Ground (simplified tiers)
	var upsCost float64
	switch {
	case weightLbs <= 1:
		upsCost = 10.00
	case weightLbs <= 3:
		upsCost = 12.00
	case weightLbs <= 5:
		upsCost = 15.00
	case weightLbs <= 10:
		upsCost = 18.00
	case weightLbs <= 20:
		upsCost = 24.00
	default:
		upsCost = 40.00
	}

	// USPS Ground Advantage
	var uspsGroundCost float64
	switch {
	case weightLbs <= 1:
		uspsGroundCost = 7.00
	case weightLbs <= 3:
		uspsGroundCost = 9.50
	case weightLbs <= 5:
		uspsGroundCost = 12.00
	case weightLbs <= 10:
		uspsGroundCost = 15.00
	case weightLbs <= 20:
		uspsGroundCost = 20.00
	default:
		uspsGroundCost = 35.00
	}

	if weightLbs <= 70 {
		estimates = append(estimates, shippingEstimate{
			Carrier: "USPS", Service: "Priority Mail",
			Cost: uspsCost, Days: "2-3",
		})
	}
	estimates = append(estimates, shippingEstimate{
		Carrier: "UPS", Service: "Ground",
		Cost: upsCost, Days: "3-5",
	})
	if weightLbs <= 70 {
		estimates = append(estimates, shippingEstimate{
			Carrier: "USPS", Service: "Ground Advantage",
			Cost: uspsGroundCost, Days: "5-7",
		})
	}

	return estimates
}

// Handler functions

func handleInventoryList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	filter := &clients.InventoryFilter{
		Category: tools.GetStringParam(req, "category"),
		Location: tools.GetStringParam(req, "location"),
		Query:    tools.GetStringParam(req, "query"),
		MinPrice: tools.GetFloatParam(req, "min_price", 0),
		MaxPrice: tools.GetFloatParam(req, "max_price", 0),
		Limit:    tools.GetIntParam(req, "limit", 50),
	}

	if v := tools.GetStringParam(req, "status"); v != "" {
		filter.Status = clients.ListingStatus(v)
	}
	if v := tools.GetStringParam(req, "condition"); v != "" {
		filter.Condition = clients.ItemCondition(v)
	}
	if v := tools.GetStringParam(req, "source"); v != "" {
		filter.Source = clients.PurchaseSource(v)
	}

	items, err := client.ListItems(ctx, filter)
	if err != nil {
		RecordError("list", "client_error")
		return tools.ErrorResult(err), nil
	}

	RecordItemOperation("list", filter.Category, "success")

	result := map[string]interface{}{
		"count": len(items),
		"items": items,
	}

	return tools.JSONResult(result), nil
}

func handleInventoryGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	item, err := client.GetItem(ctx, sku)
	if err != nil {
		RecordError("get", "client_error")
		return tools.ErrorResult(err), nil
	}

	RecordItemOperation("get", item.Category, "success")

	return tools.JSONResult(item), nil
}

func handleInventoryAdd(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	item := &clients.InventoryItem{
		Name:          tools.GetStringParam(req, "name"),
		Category:      tools.GetStringParam(req, "category"),
		PurchasePrice: tools.GetFloatParam(req, "purchase_price", 0),
		Location:      tools.GetStringParam(req, "location"),
		Description:   tools.GetStringParam(req, "description"),
		Brand:         tools.GetStringParam(req, "brand"),
		Model:         tools.GetStringParam(req, "model"),
		SerialNumber:  tools.GetStringParam(req, "serial_number"),
		Subcategory:   tools.GetStringParam(req, "subcategory"),
		Quantity:      tools.GetIntParam(req, "quantity", 1),
		OrderID:       tools.GetStringParam(req, "order_id"),
		Notes:         tools.GetStringParam(req, "notes"),
		Tags:          tools.GetStringArrayParam(req, "tags"),
		PurchaseDate:  time.Now().Format("Jan 2006"),
	}

	if v := tools.GetStringParam(req, "condition"); v != "" {
		if err := validateCondition(v); err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, err), nil
		}
		item.Condition = strings.ToLower(v)
	}
	if v := tools.GetStringParam(req, "purchase_source"); v != "" {
		item.PurchaseSource = clients.PurchaseSource(v)
	}
	if v := tools.GetStringParam(req, "vendor_name"); v != "" {
		item.VendorName = v
	}
	if v := tools.GetStringParam(req, "vendor_url"); v != "" {
		item.VendorURL = v
	}
	if v := tools.GetStringParam(req, "product_line"); v != "" {
		item.ProductLine = v
	}
	if specs := getSpecsParam(req); specs != nil {
		item.Specs = specs
	}
	if tools.HasParam(req, "on_hand") {
		item.OnHand = tools.GetBoolParam(req, "on_hand", false)
	}
	// Auto-populate specs from catalog if ProductLine is set and Specs is empty
	if item.ProductLine != "" && len(item.Specs) == 0 {
		if catalogSpecs := LookupSpecs(item.ProductLine); catalogSpecs != nil {
			item.Specs = catalogSpecs
		}
	}

	added, err := client.AddItem(ctx, item)
	if err != nil {
		RecordError("add", "client_error")
		return tools.ErrorResult(err), nil
	}

	RecordItemOperation("add", added.Category, "success")

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"sku":     added.SKU,
		"item":    added,
	}), nil
}

func handleInventoryUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	updates := make(map[string]interface{})

	// Check each updatable field
	if v := tools.GetStringParam(req, "name"); v != "" {
		updates["name"] = v
	}
	if v := tools.GetStringParam(req, "category"); v != "" {
		updates["category"] = v
	}
	if v := tools.GetStringParam(req, "subcategory"); v != "" {
		updates["subcategory"] = v
	}
	if v := tools.GetStringParam(req, "condition"); v != "" {
		if err := validateCondition(v); err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, err), nil
		}
		updates["condition"] = strings.ToLower(v)
	}
	if v := tools.GetStringParam(req, "location"); v != "" {
		updates["location"] = v
	}
	if v := tools.GetStringParam(req, "description"); v != "" {
		updates["description"] = v
	}
	if v := tools.GetStringParam(req, "brand"); v != "" {
		updates["brand"] = v
	}
	if v := tools.GetStringParam(req, "model"); v != "" {
		updates["model"] = v
	}
	if v := tools.GetStringParam(req, "notes"); v != "" {
		updates["notes"] = v
	}
	if v := tools.GetStringParam(req, "listing_status"); v != "" {
		if err := validateListingStatus(v); err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, err), nil
		}
		updates["listing_status"] = strings.ToLower(v)
	}
	if v := tools.GetFloatParam(req, "current_value", -1); v >= 0 {
		updates["current_value"] = v
	}
	if tags := tools.GetStringArrayParam(req, "tags"); len(tags) > 0 {
		updates["tags"] = tags
	}
	if v := tools.GetStringParam(req, "product_line"); v != "" {
		updates["product_line"] = v
	}
	if specs := getSpecsParam(req); specs != nil {
		updates["specs"] = specs
	}
	if tools.HasParam(req, "on_hand") {
		updates["on_hand"] = tools.GetBoolParam(req, "on_hand", false)
	}

	updated, err := client.UpdateItem(ctx, sku, updates)
	if err != nil {
		RecordError("update", "client_error")
		return tools.ErrorResult(err), nil
	}

	RecordItemOperation("update", updated.Category, "success")

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"item":    updated,
	}), nil
}

func handleInventoryDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}
	confirm := tools.GetBoolParam(req, "confirm", false)

	if !confirm {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("confirm must be true to delete item")), nil
	}

	if err := client.DeleteItem(ctx, sku); err != nil {
		RecordError("delete", "client_error")
		return tools.ErrorResult(err), nil
	}

	RecordItemOperation("delete", "", "success")

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Item %s deleted", sku),
	}), nil
}

func handleInventoryBulkDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	confirm := tools.GetBoolParam(req, "confirm", false)
	dryRun := tools.GetBoolParam(req, "dry_run", false)
	maxPrice := tools.GetFloatParam(req, "max_price", -1)

	// Build filter
	filter := &clients.InventoryFilter{
		Limit: 1000, // High limit for bulk operations
	}

	if v := tools.GetStringParam(req, "source"); v != "" {
		filter.Source = clients.PurchaseSource(v)
	}

	// Get items matching filter
	items, err := client.ListItems(ctx, filter)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Apply max_price filter if specified
	var toDelete []clients.InventoryItem
	for _, item := range items {
		if maxPrice >= 0 && item.PurchasePrice > maxPrice {
			continue
		}
		toDelete = append(toDelete, item)
	}

	if len(toDelete) == 0 {
		return tools.JSONResult(map[string]interface{}{
			"success": true,
			"message": "No items match the deletion criteria",
			"count":   0,
		}), nil
	}

	// Dry run - just show what would be deleted
	if dryRun || !confirm {
		itemSummaries := make([]map[string]interface{}, 0, len(toDelete))
		for _, item := range toDelete {
			itemSummaries = append(itemSummaries, map[string]interface{}{
				"sku":            item.SKU,
				"name":           item.Name,
				"purchase_price": item.PurchasePrice,
				"source":         item.PurchaseSource,
			})
		}
		return tools.JSONResult(map[string]interface{}{
			"dry_run":      true,
			"would_delete": len(toDelete),
			"items":        itemSummaries,
			"message":      "Set confirm=true and dry_run=false to actually delete these items",
		}), nil
	}

	// Actually delete items
	var deleted []string
	var errors []string
	for _, item := range toDelete {
		if err := client.DeleteItem(ctx, item.SKU); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", item.SKU, err))
		} else {
			deleted = append(deleted, item.SKU)
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"success":       len(errors) == 0,
		"deleted_count": len(deleted),
		"error_count":   len(errors),
		"deleted_skus":  deleted,
		"errors":        errors,
	}), nil
}

func handleInventorySearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	query := tools.GetStringParam(req, "query")
	limit := tools.GetIntParam(req, "limit", 25)

	items, err := client.SearchItems(ctx, query, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"query": query,
		"count": len(items),
		"items": items,
	}), nil
}

func handleImportAmazon(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	csvPath, errResult := tools.RequireStringParam(req, "csv_path")
	if errResult != nil {
		return errResult, nil
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("failed to open file: %w", err)), nil
	}
	defer file.Close()

	result, err := client.ImportAmazonCSV(ctx, file)
	if err != nil {
		RecordError("import_amazon", "client_error")
		RecordImportResult("amazon", 0, 0, false)
		return tools.ErrorResult(err), nil
	}

	RecordImportResult("amazon", result.Imported, 0, result.Errors == 0)

	return tools.JSONResult(result), nil
}

func handleImportNewegg(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	csvPath, errResult := tools.RequireStringParam(req, "csv_path")
	if errResult != nil {
		return errResult, nil
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("failed to open file: %w", err)), nil
	}
	defer file.Close()

	result, err := client.ImportNeweggCSV(ctx, file)
	if err != nil {
		RecordError("import_newegg", "client_error")
		RecordImportResult("newegg", 0, 0, false)
		return tools.ErrorResult(err), nil
	}

	RecordImportResult("newegg", result.Imported, 0, result.Errors == 0)

	return tools.JSONResult(result), nil
}

func handleImportGmail(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Use extended timeout for Gmail import (10 minutes)
	// Each email fetch can take up to a few seconds, and we may have hundreds of emails
	importCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	invClient, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	gmailClient, err := clients.GetGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get gmail client: %w", err)), nil
	}

	// Build config from request parameters
	config := &clients.GmailImportConfig{
		SinceDays:  tools.GetIntParam(req, "since_days", 365),
		MaxResults: tools.GetIntParam(req, "max_results", 100),
		DryRun:     tools.GetBoolParam(req, "dry_run", false),
		IncludeAll: tools.GetBoolParam(req, "include_all", false),
	}

	// Parse sources array
	if sources := tools.GetStringArrayParam(req, "sources"); len(sources) > 0 {
		config.Sources = sources
	}

	result, err := invClient.ImportGmailOrders(importCtx, gmailClient, config)
	if err != nil {
		RecordError("import_gmail", err.Error())
		return tools.ErrorResult(err), nil
	}

	// Record metrics
	RecordImportResult("gmail", result.Imported, 0, result.Errors == 0)

	return tools.JSONResult(result), nil
}

func handleFindReceipt(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sku := tools.GetStringParam(req, "sku")
	query := tools.GetStringParam(req, "query")

	if sku == "" && query == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("either sku or query is required")), nil
	}

	gmailClient, err := clients.GetGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get gmail client: %w", err)), nil
	}

	// If SKU provided, look up item to build search query
	var searchQueries []string
	var itemName string

	if sku != "" {
		invClient, err := clients.GetInventoryClient()
		if err != nil {
			return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
		}

		item, err := invClient.GetItem(ctx, sku)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("item %s not found: %w", sku, err)), nil
		}
		itemName = item.Name

		// Search by product name in order confirmation contexts
		searchQueries = append(searchQueries,
			fmt.Sprintf("(%s) (order OR confirmation OR receipt OR shipped)", item.Name),
		)

		// Also search by ASIN if available
		if item.ASIN != "" {
			searchQueries = append(searchQueries, item.ASIN)
		}

		// Search notes for order ID patterns (e.g., "Order: 112-1234567-8901234")
		if item.Notes != "" {
			// Extract order IDs from notes
			orderIDRegex := regexp.MustCompile(`(?i)order[:\s#]*([\w-]{8,})`)
			if matches := orderIDRegex.FindStringSubmatch(item.Notes); len(matches) > 1 {
				searchQueries = append(searchQueries, matches[1])
			}
		}
	} else {
		searchQueries = append(searchQueries, query)
		itemName = query
	}

	// Search Gmail with each query and deduplicate results
	seen := make(map[string]bool)
	var results []map[string]interface{}

	for _, q := range searchQueries {
		emails, err := gmailClient.SearchMessages(ctx, q, 10)
		if err != nil {
			continue
		}

		for _, email := range emails {
			if seen[email.ID] {
				continue
			}
			seen[email.ID] = true

			results = append(results, map[string]interface{}{
				"id":         email.ID,
				"subject":    email.Subject,
				"from":       email.From,
				"date":       email.Date.Format("2006-01-02"),
				"snippet":    email.Snippet,
				"thread_id":  email.ThreadID,
				"has_attach": email.HasAttach,
			})
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"item":            itemName,
		"sku":             sku,
		"queries_used":    searchQueries,
		"results_found":   len(results),
		"matching_emails": results,
	}), nil
}

func handleInventorySummary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	summary, err := client.GetSummary(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(summary), nil
}

func handleInventoryCategories(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	categories, err := client.GetCategories(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"categories": categories,
	}), nil
}

func handleInventoryLocations(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	locations, err := client.GetLocations(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"locations": locations,
	}), nil
}

func handleInventoryStale(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	days := tools.GetIntParam(req, "days", 30)

	items, err := client.GetStaleItems(ctx, days)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"days":  days,
		"count": len(items),
		"items": items,
	}), nil
}

func handleMarkSold(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}
	soldPrice := tools.GetFloatParam(req, "sold_price", 0)
	platform := tools.GetStringParam(req, "platform")

	item, err := client.MarkAsSold(ctx, sku, soldPrice, platform)
	if err != nil {
		RecordError("mark_sold", "client_error")
		return tools.ErrorResult(err), nil
	}

	RecordItemOperation("sold", item.Category, "success")
	InventoryListingsTotal.WithLabelValues(platform, "sold").Inc()

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Item %s marked as sold for $%.2f on %s", sku, soldPrice, platform),
		"item":    item,
	}), nil
}

func handleFBContent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	item, err := client.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Determine asking price
	askingPrice := tools.GetFloatParam(req, "asking_price", 0)
	if askingPrice == 0 {
		askingPrice = item.CurrentValue
	}
	if askingPrice == 0 {
		// Suggest 70% of purchase price as starting point
		askingPrice = item.PurchasePrice * 0.7
	}

	// Map condition to Facebook Marketplace label
	fbCondition := fbConditionMap[item.Condition]
	if fbCondition == "" {
		fbCondition = "Good"
	}

	// Build description
	var desc strings.Builder
	desc.WriteString(item.Name)
	if item.Brand != "" {
		desc.WriteString(fmt.Sprintf("\n\nBrand: %s", item.Brand))
	}
	if item.Model != "" {
		desc.WriteString(fmt.Sprintf("\nModel: %s", item.Model))
	}
	if item.Description != "" {
		desc.WriteString(fmt.Sprintf("\n\n%s", item.Description))
	}
	desc.WriteString(fmt.Sprintf("\n\nCondition: %s", fbCondition))
	if item.Notes != "" {
		desc.WriteString(fmt.Sprintf("\n\nNotes: %s", item.Notes))
	}
	desc.WriteString("\n\n---")
	desc.WriteString("\nLocal pickup preferred. Cash or Venmo accepted.")
	desc.WriteString("\nPrice is firm unless bundled with other items.")

	content := &clients.FBMarketplaceContent{
		Title:       item.Name,
		Price:       askingPrice,
		Category:    "Electronics",
		Condition:   fbCondition,
		Description: desc.String(),
		ImageURLs:   item.Images,
		Ready:       len(item.Images) > 0,
	}

	if !content.Ready {
		content.Message = "Add at least one photo before listing"
	}

	return tools.JSONResult(content), nil
}

func handleUploadImage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}
	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}
	setPrimary := tools.GetBoolParam(req, "set_primary", false)

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("failed to read file: %w", err)), nil
	}

	// Get filename from path
	parts := strings.Split(filePath, "/")
	filename := parts[len(parts)-1]

	// Upload to local filesystem
	path, err := client.UploadImage(ctx, sku, data, filename)
	if err != nil {
		RecordLocalImageOperation("upload", false)
		return tools.ErrorResult(err), nil
	}

	RecordLocalImageOperation("upload", true)

	// Set as primary if requested
	if setPrimary {
		_, err = client.UpdateItem(ctx, sku, map[string]interface{}{
			"primary_image": path,
		})
		if err != nil {
			return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("uploaded but failed to set as primary: %w", err)), nil
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"success":     true,
		"path":        path,
		"set_primary": setPrimary,
	}), nil
}

func handleListImages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	images, err := client.ListImages(ctx, sku)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"sku":    sku,
		"count":  len(images),
		"images": images,
	}), nil
}

func handleInventoryDashboard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	// Get summary (nil-safe: default to empty if missing)
	summary, err := client.GetSummary(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	if summary == nil {
		summary = &clients.InventorySummary{
			ByCategory: make(map[string]int),
			ByStatus:   make(map[string]int),
		}
	}

	// Get stale items
	staleItems, _ := client.GetStaleItems(ctx, 30)

	// Get sales summary from Sales tab
	salesSummary, _ := client.GetSalesSummary(ctx)
	if salesSummary == nil {
		salesSummary = &clients.SalesSummary{}
	}

	// Get config for revenue target
	config, _ := client.GetConfig(ctx)
	if config == nil {
		config = &clients.InventoryConfig{TargetRevenue: 16225}
	}

	// Revenue target progress
	progressPct := float64(0)
	remaining := config.TargetRevenue
	if config.TargetRevenue > 0 {
		progressPct = (salesSummary.TotalRevenue / config.TargetRevenue) * 100
		remaining = config.TargetRevenue - salesSummary.TotalRevenue
		if remaining < 0 {
			remaining = 0
		}
	}

	// Available items for "next actions"
	allItems, _ := client.ListItems(ctx, &clients.InventoryFilter{})
	var availableItems []clients.InventoryItem
	var notListedCount int
	for _, item := range allItems {
		status := strings.ToLower(strings.ReplaceAll(item.ListingStatus, " ", "_"))
		if status != "sold" {
			availableItems = append(availableItems, item)
		}
		if status == "not_listed" {
			notListedCount++
		}
	}

	// On-hand tracking
	var onHandCount int
	var onHandValue float64
	var notOnHandItems []map[string]interface{}
	for _, item := range availableItems {
		if item.OnHand {
			onHandCount++
			onHandValue += item.AskingPrice * float64(item.Quantity)
		} else {
			notOnHandItems = append(notOnHandItems, map[string]interface{}{
				"sku":  item.SKU,
				"name": item.Name,
			})
		}
	}

	// Build next actions
	var nextActions []string
	if notListedCount > 0 {
		nextActions = append(nextActions, fmt.Sprintf("List %d unlisted items on marketplaces", notListedCount))
	}
	for _, item := range availableItems {
		cat := strings.ToLower(item.Category)
		if (strings.Contains(cat, "storage") || strings.Contains(cat, "nvme") || strings.Contains(cat, "ssd") || strings.Contains(cat, "hdd")) && item.SmartData == "" {
			nextActions = append(nextActions, "Run smart_collect on storage devices to add health data")
			break
		}
	}
	if len(availableItems) > 0 {
		highest := availableItems[0]
		for _, item := range availableItems[1:] {
			if item.AskingPrice > highest.AskingPrice {
				highest = item
			}
		}
		nextActions = append(nextActions, fmt.Sprintf("Priority: sell %s ($%.0f)", highest.Name, highest.AskingPrice))
	}
	if remaining > 0 {
		nextActions = append(nextActions, fmt.Sprintf("$%.0f remaining to hit $%.0f revenue target", remaining, config.TargetRevenue))
	}

	// Enhancement D: detect items needing repricing (listed >7 days)
	// and items below floor price (asking < purchase × 0.5)
	staleListedItems, _ := client.GetStaleItems(ctx, 7)
	var needsRepricing []map[string]interface{}
	for _, item := range staleListedItems {
		status := strings.ToLower(item.ListingStatus)
		if status == "listed" || status == string(clients.StatusListed) {
			needsRepricing = append(needsRepricing, map[string]interface{}{
				"sku":          item.SKU,
				"name":         item.Name,
				"asking_price": item.AskingPrice,
			})
		}
	}

	var belowFloor []map[string]interface{}
	for _, item := range availableItems {
		if item.PurchasePrice > 0 && item.AskingPrice > 0 && item.AskingPrice < item.PurchasePrice*0.5 {
			belowFloor = append(belowFloor, map[string]interface{}{
				"sku":            item.SKU,
				"name":           item.Name,
				"asking_price":   item.AskingPrice,
				"purchase_price": item.PurchasePrice,
			})
		}
	}

	dashboard := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_items":       summary.TotalItems,
			"total_cost":        summary.TotalCost,
			"total_value":       summary.TotalValue,
			"items_by_status":   summary.ByStatus,
			"items_by_category": summary.ByCategory,
		},
		"financials": map[string]interface{}{
			"total_revenue":       salesSummary.TotalRevenue,
			"total_cost":          salesSummary.TotalCost,
			"total_fees":          salesSummary.TotalFees,
			"net_profit":          salesSummary.TotalProfit,
			"units_sold":          salesSummary.UnitsSold,
			"revenue_target":      config.TargetRevenue,
			"progress_pct":        fmt.Sprintf("%.1f%%", progressPct),
			"remaining_to_target": remaining,
		},
		"attention_needed": map[string]interface{}{
			"not_listed":      notListedCount,
			"stale_30_days":   len(staleItems),
			"needs_repricing": needsRepricing,
			"below_floor":     belowFloor,
		},
		"on_hand": map[string]interface{}{
			"count":            onHandCount,
			"total_items":      len(availableItems),
			"est_resale_value": onHandValue,
			"not_located":      notOnHandItems,
		},
		"top_value_items": summary.TopValueItems,
		"next_actions":    nextActions,
	}

	return tools.JSONResult(dashboard), nil
}

// eBay handler functions

func handleEbayPriceResearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ebayClient, err := clients.GetEbayClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get eBay client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	analysis, err := ebayClient.SearchCompletedItems(ctx, query, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(analysis), nil
}

func handleEbayActiveListings(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ebayClient, err := clients.GetEbayClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get eBay client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	results, err := ebayClient.SearchActiveListings(ctx, query, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	}), nil
}

func handleEbayCategories(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ebayClient, err := clients.GetEbayClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get eBay client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	categories, err := ebayClient.GetCategorySuggestions(ctx, query)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"query":      query,
		"count":      len(categories),
		"categories": categories,
	}), nil
}

func handleEbayCreateListing(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Get inventory client for item data
	invClient, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	// Get eBay client
	ebayClient, err := clients.GetEbayClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get eBay client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	price := tools.GetFloatParam(req, "price", 0)
	if price <= 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("price must be greater than 0")), nil
	}

	categoryID, errResult := tools.RequireStringParam(req, "category_id")
	if errResult != nil {
		return errResult, nil
	}

	quantity := tools.GetIntParam(req, "quantity", 1)
	condition := tools.OptionalStringParam(req, "condition", "GOOD")

	// Get inventory item
	item, err := invClient.GetItem(ctx, sku)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get inventory item: %w", err)), nil
	}
	if item == nil {
		return tools.CodedErrorResult(tools.ErrNotFound, fmt.Errorf("inventory item not found: %s", sku)), nil
	}

	// Create eBay inventory item
	ebayItem := &clients.EbayInventoryItem{
		SKU:       item.SKU,
		Condition: condition,
		Product: &clients.EbayProduct{
			Title:       item.Name,
			Description: item.Description,
			ImageUrls:   item.Images,
			Aspects: map[string][]string{
				"Brand": {item.Brand},
				"Model": {item.Model},
			},
		},
		Availability: &clients.EbayAvailability{
			ShipToLocationAvailability: &clients.EbayShipToLocation{
				Quantity: quantity,
			},
		},
	}

	// Add condition description based on inventory condition
	condLower := strings.ToLower(item.Condition)
	switch {
	case condLower == "new" || condLower == "new / sealed":
		ebayItem.ConditionDescription = "Brand new, unopened"
	case condLower == "like_new" || condLower == "renewed":
		ebayItem.ConditionDescription = "Like new, minimal use"
	case condLower == "good" || condLower == "used":
		ebayItem.ConditionDescription = "Good condition, normal wear"
	case condLower == "fair":
		ebayItem.ConditionDescription = "Fair condition, shows wear but fully functional"
	default:
		ebayItem.ConditionDescription = item.Notes
	}

	// Create inventory item in eBay
	if err := ebayClient.CreateInventoryItem(ctx, ebayItem); err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create eBay inventory item: %w", err)), nil
	}

	// Create offer
	offer := &clients.EbayOffer{
		SKU:           item.SKU,
		MarketplaceID: "EBAY_US",
		Format:        "FIXED_PRICE",
		PricingSummary: &clients.EbayPricing{
			Price: &clients.EbayAmount{
				Value:    fmt.Sprintf("%.2f", price),
				Currency: "USD",
			},
		},
		CategoryID:        categoryID,
		AvailableQuantity: quantity,
	}

	offerID, err := ebayClient.CreateOffer(ctx, offer)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create eBay offer: %w", err)), nil
	}

	// Publish the offer
	listingID, err := ebayClient.PublishOffer(ctx, offerID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to publish eBay listing: %w", err)), nil
	}

	// Update inventory item with listing info
	updates := map[string]interface{}{
		"listing_status":  string(clients.StatusListed),
		"ebay_listing_id": listingID,
		"ebay_url":        fmt.Sprintf("https://www.ebay.com/itm/%s", listingID),
		"listed_price":    price,
	}

	if _, updateErr := invClient.UpdateItem(ctx, sku, updates); updateErr != nil {
		// Listing was created on eBay — log warning but don't fail the response
		_ = updateErr
	}

	InventoryListingsTotal.WithLabelValues("ebay", "listed").Inc()

	return tools.JSONResult(map[string]interface{}{
		"success":    true,
		"listing_id": listingID,
		"offer_id":   offerID,
		"ebay_url":   fmt.Sprintf("https://www.ebay.com/itm/%s", listingID),
		"price":      price,
		"sku":        sku,
	}), nil
}

func handleInventoryValue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	includeSold := tools.GetBoolParam(req, "include_sold", false)
	groupBy := tools.GetStringParam(req, "group_by")

	filter := &clients.InventoryFilter{Limit: 1000}
	if !includeSold {
		// Exclude sold items by not setting status filter (get all non-sold)
	}

	items, err := client.ListItems(ctx, filter)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var totalPurchaseValue, totalCurrentValue, totalListedValue, totalSoldValue float64
	itemCount := 0
	soldCount := 0
	breakdown := make(map[string]float64)

	for _, item := range items {
		if strings.EqualFold(item.ListingStatus, "sold") {
			if includeSold {
				totalSoldValue += item.SoldPrice
				soldCount++
			}
			continue
		}
		itemCount++
		totalPurchaseValue += item.PurchasePrice
		if item.CurrentValue > 0 {
			totalCurrentValue += item.CurrentValue
		} else {
			totalCurrentValue += item.PurchasePrice
		}
		if item.ListedPrice > 0 {
			totalListedValue += item.ListedPrice
		}

		// Group by requested field
		var key string
		switch groupBy {
		case "category":
			key = item.Category
		case "status":
			key = string(item.ListingStatus)
		case "location":
			key = item.Location
		case "condition":
			key = string(item.Condition)
		}
		if key != "" {
			breakdown[key] += item.PurchasePrice
		}
	}

	result := map[string]interface{}{
		"item_count":           itemCount,
		"total_purchase_value": totalPurchaseValue,
		"total_current_value":  totalCurrentValue,
		"total_listed_value":   totalListedValue,
	}
	if includeSold {
		result["sold_count"] = soldCount
		result["total_sold_value"] = totalSoldValue
	}
	if len(breakdown) > 0 {
		result["breakdown"] = breakdown
	}

	return tools.JSONResult(result), nil
}

func handleInventoryExport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	format := tools.OptionalStringParam(req, "format", "json")

	filter := &clients.InventoryFilter{
		Category: tools.GetStringParam(req, "category"),
		Limit:    10000,
	}
	if v := tools.GetStringParam(req, "status"); v != "" {
		filter.Status = clients.ListingStatus(v)
	}

	items, err := client.ListItems(ctx, filter)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if format == "csv" {
		var lines []string
		lines = append(lines, "SKU,Name,Category,Brand,Model,Condition,PurchasePrice,CurrentValue,ListingStatus,Location")
		for _, item := range items {
			line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%.2f,%.2f,%s,%s",
				item.SKU, escapeCsvField(item.Name), item.Category, item.Brand, item.Model,
				item.Condition, item.PurchasePrice, item.CurrentValue, item.ListingStatus, item.Location)
			lines = append(lines, line)
		}
		return tools.TextResult(strings.Join(lines, "\n")), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"count": len(items),
		"items": items,
	}), nil
}

func escapeCsvField(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

func handleInventoryMove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}
	location, errResult := tools.RequireStringParam(req, "location")
	if errResult != nil {
		return errResult, nil
	}

	// Get current item to verify it exists
	item, err := client.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	oldLocation := item.Location

	updates := map[string]interface{}{
		"location": location,
	}

	if _, err := client.UpdateItem(ctx, sku, updates); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":      true,
		"sku":          sku,
		"old_location": oldLocation,
		"new_location": location,
	}), nil
}

func handleDeleteImage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}
	imageKey, errResult := tools.RequireStringParam(req, "image_key")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.DeleteImage(ctx, sku, imageKey); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":   true,
		"sku":       sku,
		"image_key": imageKey,
		"message":   "Image deleted successfully",
	}), nil
}

func handleSetPrimaryImage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}
	imageKey, errResult := tools.RequireStringParam(req, "image_key")
	if errResult != nil {
		return errResult, nil
	}

	// Update item with new primary image
	updates := map[string]interface{}{
		"primary_image": imageKey,
	}

	if _, err := client.UpdateItem(ctx, sku, updates); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":       true,
		"sku":           sku,
		"primary_image": imageKey,
		"message":       "Primary image updated successfully",
	}), nil
}

// Data Quality Handlers

func handleInventoryRecategorize(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku := tools.GetStringParam(req, "sku")
	dryRun := tools.GetBoolParam(req, "dry_run", true)
	limit := tools.GetIntParam(req, "limit", 100)

	// If single SKU provided, recategorize just that item
	if sku != "" {
		item, err := client.GetItem(ctx, sku)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		if item == nil {
			return tools.CodedErrorResult(tools.ErrNotFound, fmt.Errorf("item not found: %s", sku)), nil
		}

		result := client.RecategorizeItem(item)

		if !dryRun && result.Changed {
			updates := map[string]interface{}{
				"category": result.NewCategory,
			}
			if result.NewSubcategory != "" {
				updates["subcategory"] = result.NewSubcategory
			}
			if result.NewBrand != "" {
				updates["brand"] = result.NewBrand
			}
			if _, err := client.UpdateItem(ctx, sku, updates); err != nil {
				return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to apply changes: %w", err)), nil
			}
		}

		return tools.JSONResult(map[string]interface{}{
			"dry_run": dryRun,
			"result":  result,
			"applied": !dryRun && result.Changed,
		}), nil
	}

	// Batch recategorization with filter
	filter := &clients.InventoryFilter{
		Category: tools.GetStringParam(req, "category"),
		Limit:    limit,
	}

	results, err := client.RecategorizeItems(ctx, filter, !dryRun)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Count changes
	changedCount := 0
	for _, r := range results {
		if r.Changed {
			changedCount++
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"dry_run":   dryRun,
		"total":     len(results),
		"changed":   changedCount,
		"unchanged": len(results) - changedCount,
		"applied":   !dryRun,
		"results":   results,
	}), nil
}

func handleInventoryDuplicates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	matchType := tools.OptionalStringParam(req, "match_type", "all")

	groups, err := client.FindDuplicates(ctx, matchType)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Calculate totals
	totalDuplicates := 0
	for _, g := range groups {
		totalDuplicates += g.Count - 1 // -1 because one is the "original"
	}

	return tools.JSONResult(map[string]interface{}{
		"match_type":       matchType,
		"duplicate_groups": len(groups),
		"total_duplicates": totalDuplicates,
		"groups":           groups,
	}), nil
}

func handleInventoryBulkUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	dryRun := tools.GetBoolParam(req, "dry_run", true)
	confirm := tools.GetBoolParam(req, "confirm", false)
	limit := tools.GetIntParam(req, "limit", 100)
	if limit > 500 {
		limit = 500
	}

	// Build filter
	filter := &clients.InventoryFilter{
		Category: tools.GetStringParam(req, "category"),
		Limit:    limit,
	}
	if v := tools.GetStringParam(req, "status"); v != "" {
		filter.Status = clients.ListingStatus(v)
	}
	if v := tools.GetStringParam(req, "source"); v != "" {
		filter.Source = clients.PurchaseSource(v)
	}

	// Get matching items
	items, err := client.ListItems(ctx, filter)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(items) == 0 {
		return tools.JSONResult(map[string]interface{}{
			"message": "No items match the filter criteria",
			"count":   0,
		}), nil
	}

	// Build updates
	updates := make(map[string]interface{})
	if v := tools.GetStringParam(req, "new_status"); v != "" {
		if err := validateListingStatus(v); err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, err), nil
		}
		updates["listing_status"] = strings.ToLower(v)
	}
	if v := tools.GetStringParam(req, "new_location"); v != "" {
		updates["location"] = v
	}

	// on_hand is a bool param — check if explicitly provided
	if tools.HasParam(req, "new_on_hand") {
		updates["on_hand"] = tools.GetBoolParam(req, "new_on_hand", false)
	}

	addTags := tools.GetStringArrayParam(req, "add_tags")
	removeTags := tools.GetStringArrayParam(req, "remove_tags")

	if len(updates) == 0 && len(addTags) == 0 && len(removeTags) == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("no updates specified (use new_status, new_location, new_on_hand, add_tags, or remove_tags)")), nil
	}

	// Preview mode
	if dryRun || !confirm {
		itemSummaries := make([]map[string]interface{}, 0, len(items))
		for _, item := range items {
			itemSummaries = append(itemSummaries, map[string]interface{}{
				"sku":      item.SKU,
				"name":     item.Name,
				"category": item.Category,
				"status":   item.ListingStatus,
				"location": item.Location,
				"on_hand":  item.OnHand,
			})
		}
		return tools.JSONResult(map[string]interface{}{
			"dry_run":      true,
			"would_update": len(items),
			"updates":      updates,
			"add_tags":     addTags,
			"remove_tags":  removeTags,
			"items":        itemSummaries,
			"message":      "Set dry_run=false and confirm=true to apply these changes",
		}), nil
	}

	// Apply updates
	var updated []string
	var errors []string

	for _, item := range items {
		itemUpdates := make(map[string]interface{})
		for k, v := range updates {
			itemUpdates[k] = v
		}

		// Handle tag updates
		if len(addTags) > 0 || len(removeTags) > 0 {
			newTags := make([]string, 0, len(item.Tags))
			// Keep existing tags that aren't being removed
			for _, t := range item.Tags {
				shouldRemove := false
				for _, rt := range removeTags {
					if t == rt {
						shouldRemove = true
						break
					}
				}
				if !shouldRemove {
					newTags = append(newTags, t)
				}
			}
			// Add new tags (avoid duplicates)
			for _, at := range addTags {
				exists := false
				for _, t := range newTags {
					if t == at {
						exists = true
						break
					}
				}
				if !exists {
					newTags = append(newTags, at)
				}
			}
			itemUpdates["tags"] = newTags
		}

		if _, err := client.UpdateItem(ctx, item.SKU, itemUpdates); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", item.SKU, err))
		} else {
			updated = append(updated, item.SKU)
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"success":       len(errors) == 0,
		"updated_count": len(updated),
		"error_count":   len(errors),
		"updated_skus":  updated,
		"errors":        errors,
	}), nil
}

// ── New handlers (ported from hw-resale) ──

func handleListingGenerate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	platform := tools.OptionalStringParam(req, "platform", "fb_marketplace")

	item, err := client.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	listing, err := RenderListing(item, platform)
	if err != nil {
		RecordError("listing_generate", "render_error")
		return tools.ErrorResult(err), nil
	}

	InventoryListingsTotal.WithLabelValues(platform, "generated").Inc()

	return tools.JSONResult(map[string]interface{}{
		"sku":             sku,
		"platform":        platform,
		"listing":         listing,
		"suggested_title": item.Name,
		"suggested_price": item.AskingPrice,
	}), nil
}

func handleSmartCollect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	devicePath := tools.OptionalStringParam(req, "device_path", "/dev/sda")

	// Verify item exists and is a storage device
	item, err := client.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	storageCategories := map[string]bool{
		"Storage": true, "NVMe Gen5": true, "storage": true,
		"nvme": true, "sata_ssd": true, "hdd": true,
	}
	if !storageCategories[item.Category] {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("item %s is category %q, not a storage device", sku, item.Category)), nil
	}

	smartData, err := CollectSMARTData(ctx, devicePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Serialize and store on item
	smartJSON, _ := json.Marshal(smartData)
	_, err = client.UpdateItem(ctx, sku, map[string]interface{}{
		"smart_data": string(smartJSON),
	})
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("SMART data collected but failed to save: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"sku":         sku,
		"device_path": devicePath,
		"smart_data":  smartData,
		"message":     "SMART data collected and saved to inventory.",
	}), nil
}

func handleSaleRecord(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	soldPrice := tools.GetFloatParam(req, "sold_price", 0)
	if soldPrice <= 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("sold_price is required and must be > 0")), nil
	}
	platform, errResult := tools.RequireStringParam(req, "platform")
	if errResult != nil {
		return errResult, nil
	}

	// Get item for details
	item, err := client.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	qtySold := tools.GetIntParam(req, "quantity_sold", 1)
	if qtySold > item.Quantity {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("cannot sell %d units — only %d available", qtySold, item.Quantity)), nil
	}

	sale := &clients.SaleRecord{
		ItemNum:       item.RowNum,
		ItemName:      item.Name,
		QuantitySold:  qtySold,
		SoldPrice:     soldPrice,
		PurchasePrice: item.PurchasePrice,
		ShippingCost:  tools.GetFloatParam(req, "shipping_cost", 0),
		PlatformFees:  tools.GetFloatParam(req, "platform_fees", 0),
		Platform:      platform,
		BuyerInfo:     tools.GetStringParam(req, "buyer_info"),
		Notes:         tools.GetStringParam(req, "notes"),
		Date:          time.Now().Format(time.RFC3339),
	}

	recorded, err := client.RecordSale(ctx, sale)
	if err != nil {
		RecordError("sale_record", "client_error")
		return tools.ErrorResult(err), nil
	}

	InventorySalesTotal.WithLabelValues(platform).Inc()
	InventoryRevenueGauge.Add(soldPrice)

	// Update inventory item
	newQty := item.Quantity - qtySold
	updates := map[string]interface{}{
		"sold_price":    soldPrice,
		"sold_platform": platform,
		"sold_date":     time.Now().Format("Jan 2006"),
	}
	if newQty <= 0 {
		updates["listing_status"] = "sold"
		updates["quantity"] = 0
	} else {
		updates["quantity"] = newQty
	}
	_, _ = client.UpdateItem(ctx, sku, updates)

	return tools.JSONResult(map[string]interface{}{
		"sale":               recorded,
		"remaining_quantity": newQty,
	}), nil
}

func handlePriceCheck(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku := tools.GetStringParam(req, "sku")
	searchQuery := tools.GetStringParam(req, "search_query")

	var item *clients.InventoryItem

	if sku != "" {
		item, err = client.GetItem(ctx, sku)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
	} else if searchQuery != "" {
		items, err := client.SearchItems(ctx, searchQuery, 1)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		if len(items) == 0 {
			return tools.ErrorResult(fmt.Errorf("no items found matching %q", searchQuery)), nil
		}
		item = &items[0]
		sku = item.SKU
	} else {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("sku or search_query is required")), nil
	}

	margin := item.AskingPrice - item.PurchasePrice
	marginPct := float64(0)
	if item.PurchasePrice > 0 {
		marginPct = (margin / item.PurchasePrice) * 100
	}

	result := map[string]interface{}{
		"sku":            sku,
		"name":           item.Name,
		"purchase_price": item.PurchasePrice,
		"asking_price":   item.AskingPrice,
		"msrp":           item.MSRP,
		"current_retail": item.CurrentRetail,
		"margin":         fmt.Sprintf("$%.2f", margin),
		"margin_pct":     fmt.Sprintf("%.1f%%", marginPct),
		"condition":      item.Condition,
	}

	if item.ProductLine != "" {
		result["product_line"] = item.ProductLine
	}
	if len(item.Specs) > 0 {
		result["specs"] = item.Specs
	}

	// Pricing guidance
	if marginPct < 0 {
		result["guidance"] = "Selling at a loss. Consider raising price or bundling."
	} else if marginPct < 10 {
		result["guidance"] = "Thin margin. Price is competitive but leaves little room."
	} else if marginPct < 30 {
		result["guidance"] = "Healthy margin. Good balance of competitiveness and profit."
	} else {
		result["guidance"] = "Strong margin. May sell slower at this price point."
	}

	// Always emit search_suggestion
	searchName := item.Name
	if item.ProductLine != "" {
		searchName = item.ProductLine
	}
	result["search_suggestion"] = fmt.Sprintf("Search eBay sold listings for: %s %s", searchName, item.Condition)

	return tools.JSONResult(result), nil
}

func handleSalesLog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sales, err := client.GetSalesLog(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	summary, err := client.GetSalesSummary(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"summary": summary,
		"sales":   sales,
	}), nil
}

// ── Tax Report ──

func handleTaxReport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	year := tools.GetIntParam(req, "year", time.Now().Year())

	sales, err := client.GetSalesLog(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Filter to the requested tax year
	var filtered []clients.SaleRecord
	for _, s := range sales {
		saleYear := extractYear(s.Date)
		if saleYear == year {
			filtered = append(filtered, s)
		}
	}

	if len(filtered) == 0 {
		return tools.JSONResult(map[string]interface{}{
			"year":    year,
			"message": fmt.Sprintf("No sales found for tax year %d", year),
		}), nil
	}

	// Aggregate Schedule C line items
	var grossReceipts, cogs, shippingExpenses, platformFees, totalExpenses, netProfit float64
	platformBreakdown := make(map[string]float64)
	var unitsSold int

	for _, s := range filtered {
		grossReceipts += s.Revenue
		cogs += s.Cost
		shippingExpenses += s.ShippingCost
		platformFees += s.PlatformFees
		netProfit += s.NetProfit
		unitsSold += s.QuantitySold
		platformBreakdown[s.Platform] += s.Revenue
	}
	totalExpenses = shippingExpenses + platformFees

	// Build Schedule C report
	report := map[string]interface{}{
		"tax_year":    year,
		"total_sales": len(filtered),
		"units_sold":  unitsSold,

		// Schedule C Part I — Income
		"schedule_c_income": map[string]interface{}{
			"line_1_gross_receipts": grossReceipts,
			"line_4_cogs":           cogs,
			"line_7_gross_income":   grossReceipts - cogs,
		},

		// Schedule C Part II — Expenses
		"schedule_c_expenses": map[string]interface{}{
			"line_27a_other_expenses": map[string]interface{}{
				"shipping_costs": shippingExpenses,
				"platform_fees":  platformFees,
			},
			"total_expenses": totalExpenses,
		},

		// Schedule C Part I — Net
		"schedule_c_net": map[string]interface{}{
			"line_31_net_profit_or_loss": grossReceipts - cogs - totalExpenses,
		},

		// Summary breakdown
		"breakdown": map[string]interface{}{
			"gross_receipts":    grossReceipts,
			"cost_of_goods":     cogs,
			"shipping_expenses": shippingExpenses,
			"platform_fees":     platformFees,
			"total_expenses":    totalExpenses,
			"net_profit":        netProfit,
			"gross_margin":      safePercent(grossReceipts-cogs, grossReceipts),
			"net_margin":        safePercent(netProfit, grossReceipts),
		},

		// By platform
		"revenue_by_platform": platformBreakdown,

		"disclaimer": "This is a summary for tax preparation purposes only. Consult a tax professional for actual filing.",
	}

	return tools.JSONResult(report), nil
}

func extractYear(dateStr string) int {
	// Try common date formats: "Jan 2006", "2006-01-02", "01/02/2006", "2006"
	for _, layout := range []string{"Jan 2006", "2006-01-02", "01/02/2006", "January 2006"} {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.Year()
		}
	}
	// Try extracting 4-digit year from anywhere in the string
	for i := 0; i <= len(dateStr)-4; i++ {
		if dateStr[i] >= '1' && dateStr[i] <= '2' {
			var y int
			if _, err := fmt.Sscanf(dateStr[i:], "%4d", &y); err == nil && y >= 2000 && y <= 2100 {
				return y
			}
		}
	}
	return 0
}

func safePercent(num, denom float64) string {
	if denom == 0 {
		return "0.0%"
	}
	return fmt.Sprintf("%.1f%%", (num/denom)*100)
}

// ── Discord Inventory Alerts ──

func handleDiscordAlerts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	invClient, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	channelID := tools.GetStringParam(req, "channel_id")
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	items, err := invClient.ListItems(ctx, &clients.InventoryFilter{Limit: 1000})
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var alerts []map[string]interface{}

	for _, item := range items {
		// Skip sold items
		if strings.EqualFold(item.ListingStatus, "sold") {
			continue
		}

		// Alert: below floor price (asking < 50% of purchase price)
		if item.PurchasePrice > 0 && item.AskingPrice > 0 && item.AskingPrice < item.PurchasePrice*0.5 {
			alerts = append(alerts, map[string]interface{}{
				"type":    "below_floor",
				"level":   "warning",
				"sku":     item.SKU,
				"name":    item.Name,
				"message": fmt.Sprintf("Asking $%.0f is below 50%% floor of purchase price $%.0f", item.AskingPrice, item.PurchasePrice),
			})
		}

		// Alert: stale listing (listed status but no sale for >30 days)
		if strings.EqualFold(item.ListingStatus, "listed") && item.PurchaseDate != "" {
			// Approximate: if listed and purchase was >60 days ago, probably stale
			alerts = append(alerts, map[string]interface{}{
				"type":    "stale_listing",
				"level":   "info",
				"sku":     item.SKU,
				"name":    item.Name,
				"message": fmt.Sprintf("Listed but not yet sold (price: $%.0f)", item.AskingPrice),
			})
		}

		// Alert: no asking price set
		if item.AskingPrice <= 0 && !strings.EqualFold(item.ListingStatus, "keeping") {
			alerts = append(alerts, map[string]interface{}{
				"type":    "no_price",
				"level":   "warning",
				"sku":     item.SKU,
				"name":    item.Name,
				"message": "No asking price set — needs pricing",
			})
		}
	}

	if len(alerts) == 0 {
		return tools.JSONResult(map[string]interface{}{
			"alerts_found": 0,
			"message":      "No inventory alerts — everything looks good",
		}), nil
	}

	// Build summary message for Discord
	var belowFloor, stale, noPrice int
	for _, a := range alerts {
		switch a["type"] {
		case "below_floor":
			belowFloor++
		case "stale_listing":
			stale++
		case "no_price":
			noPrice++
		}
	}

	summary := fmt.Sprintf("**Inventory Alert Summary** (%d issues)\n", len(alerts))
	if belowFloor > 0 {
		summary += fmt.Sprintf("- %d items below floor price\n", belowFloor)
	}
	if stale > 0 {
		summary += fmt.Sprintf("- %d stale listings\n", stale)
	}
	if noPrice > 0 {
		summary += fmt.Sprintf("- %d items missing prices\n", noPrice)
	}

	// Add details for each alert
	summary += "\n**Details:**\n"
	for _, a := range alerts {
		icon := "ℹ️"
		if a["level"] == "warning" {
			icon = "⚠️"
		}
		summary += fmt.Sprintf("%s **%s** — %s\n", icon, a["sku"], a["message"])
	}

	if dryRun {
		return tools.JSONResult(map[string]interface{}{
			"dry_run":      true,
			"alerts_found": len(alerts),
			"below_floor":  belowFloor,
			"stale":        stale,
			"no_price":     noPrice,
			"alerts":       alerts,
			"message":      summary,
		}), nil
	}

	// Send to Discord
	discordClient, err := clients.NewDiscordClient()
	if err != nil {
		return tools.JSONResult(map[string]interface{}{
			"alerts_found":  len(alerts),
			"alerts":        alerts,
			"discord_error": fmt.Sprintf("Could not send to Discord: %v", err),
			"message":       summary,
		}), nil
	}

	level := "info"
	if belowFloor > 0 {
		level = "warning"
	}
	_, sendErr := discordClient.SendNotification(ctx, channelID, "Inventory Alerts", summary, level)

	result := map[string]interface{}{
		"alerts_found": len(alerts),
		"below_floor":  belowFloor,
		"stale":        stale,
		"no_price":     noPrice,
		"alerts":       alerts,
		"discord_sent": sendErr == nil,
	}
	if sendErr != nil {
		result["discord_error"] = sendErr.Error()
	}

	return tools.JSONResult(result), nil
}

// ── Quick List (Enhancement C) ──

func handleQuickList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	item, err := client.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Auto-enrich specs from catalog if empty
	if len(item.Specs) == 0 && item.ProductLine != "" {
		if catalogSpecs := LookupSpecs(item.ProductLine); catalogSpecs != nil {
			item.Specs = catalogSpecs
		}
	}

	// Margin analysis
	margin := item.AskingPrice - item.PurchasePrice
	marginPct := float64(0)
	if item.PurchasePrice > 0 {
		marginPct = (margin / item.PurchasePrice) * 100
	}

	var guidance string
	if marginPct < 0 {
		guidance = "Selling at a loss. Consider raising price or bundling."
	} else if marginPct < 10 {
		guidance = "Thin margin. Price is competitive but leaves little room."
	} else if marginPct < 30 {
		guidance = "Healthy margin. Good balance of competitiveness and profit."
	} else {
		guidance = "Strong margin. May sell slower at this price point."
	}

	// Generate listings for all 3 platforms
	listings := make(map[string]string)
	for _, platform := range []string{"fb_marketplace", "ebay", "hardwareswap"} {
		text, err := RenderListing(item, platform)
		if err != nil {
			listings[platform] = fmt.Sprintf("(error: %v)", err)
		} else {
			listings[platform] = text
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"item": map[string]interface{}{
			"sku":            sku,
			"name":           item.Name,
			"category":       item.Category,
			"condition":      item.Condition,
			"product_line":   item.ProductLine,
			"specs":          item.Specs,
			"asking_price":   item.AskingPrice,
			"purchase_price": item.PurchasePrice,
		},
		"pricing": map[string]interface{}{
			"asking":     item.AskingPrice,
			"margin":     fmt.Sprintf("$%.2f", margin),
			"margin_pct": fmt.Sprintf("%.1f%%", marginPct),
			"guidance":   guidance,
		},
		"listings": listings,
	}), nil
}

// ── Shipping Estimate (Enhancement E) ──

func handleShippingEstimate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sku := tools.GetStringParam(req, "sku")
	weightLbs := tools.GetFloatParam(req, "weight_lbs", 0)

	var itemName, category string

	if sku != "" {
		client, err := clients.GetInventoryClient()
		if err != nil {
			return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
		}
		item, err := client.GetItem(ctx, sku)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		itemName = item.Name
		category = item.Category
		if weightLbs == 0 {
			if w, ok := shippingWeights[category]; ok {
				weightLbs = w
			} else {
				weightLbs = 3.0 // default
			}
		}
	} else if weightLbs == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("sku or weight_lbs is required")), nil
	}

	estimates := estimateShipping(weightLbs)

	// Find best value
	var recommendation string
	if len(estimates) > 0 {
		best := estimates[0]
		for _, e := range estimates[1:] {
			if e.Cost < best.Cost {
				best = e
			}
		}
		recommendation = fmt.Sprintf("%s %s — best value for %.1f lb shipment", best.Carrier, best.Service, weightLbs)
	}

	result := map[string]interface{}{
		"weight_lbs":     weightLbs,
		"estimates":      estimates,
		"recommendation": recommendation,
	}
	if sku != "" {
		result["sku"] = sku
		result["name"] = itemName
		result["category"] = category
	}
	if weightLbs > 0 {
		// Insurance note for items that may need it
		result["insurance_note"] = "Include insurance for items valued over $100. USPS Priority includes $50 free."
	}

	return tools.JSONResult(result), nil
}

// ── Helpers ──

// getSpecsParam extracts a map[string]string from a "specs" parameter.
func getSpecsParam(req mcp.CallToolRequest) map[string]string {
	args, ok := req.Params.Arguments.(map[string]interface{})
	if !ok {
		return nil
	}
	raw, ok := args["specs"]
	if !ok || raw == nil {
		return nil
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// ── Phase 2: JSON Import ──

// jsonImportItem is the structure expected from Python hw-resale JSON exports.
type jsonImportItem struct {
	Name         string            `json:"name"`
	Category     string            `json:"category"`
	ProductLine  string            `json:"product_line"`
	Quantity     int               `json:"quantity"`
	Condition    string            `json:"condition"`
	PurchaseDate string            `json:"purchase_date"`
	AskingPrice  float64           `json:"asking_price"`
	MSRP         float64           `json:"msrp"`
	Notes        string            `json:"notes"`
	Specs        map[string]string `json:"specs"`
	Location     string            `json:"location"`
	Model        string            `json:"model"`
	ASIN         string            `json:"asin"`
	// Python prototype fields
	PurchasePrice float64 `json:"purchase_price"`
	CurrentValue  float64 `json:"current_value"`
}

// mapCategory normalizes Python prototype categories to Go inventory categories.
func mapCategory(cat string) string {
	switch strings.ToLower(cat) {
	case "gpu":
		return "GPU"
	case "nvme", "sata_ssd", "hdd", "storage":
		return "Storage"
	case "ddr5", "ddr4", "ram":
		return "RAM"
	case "networking":
		return "Networking"
	case "accessories", "components":
		return "Components"
	case "cpu":
		return "CPU"
	case "motherboard":
		return "Motherboard"
	case "psu":
		return "PSU"
	case "cooling":
		return "Cooling"
	case "peripherals":
		return "Peripherals"
	case "thunderbolt":
		return "Thunderbolt"
	default:
		return "Other"
	}
}

// mapCondition normalizes Python prototype conditions to Go inventory conditions.
func mapCondition(cond string) string {
	switch strings.ToLower(cond) {
	case "new", "new_sealed":
		return "new"
	case "new_open_box", "used_excellent":
		return "like_new"
	case "used_good":
		return "good"
	case "used_fair":
		return "fair"
	case "for_parts":
		return "for_parts"
	default:
		return cond
	}
}

func handleImportJSON(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}
	dryRun := tools.GetBoolParam(req, "dry_run", true)
	source := tools.OptionalStringParam(req, "source", "json_import")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to read file: %w", err)), nil
	}

	var items []jsonImportItem
	if err := json.Unmarshal(data, &items); err != nil {
		// Try unwrapping from {"items": [...]} or {"inventory": [...]}
		var wrapper map[string]json.RawMessage
		if err2 := json.Unmarshal(data, &wrapper); err2 != nil {
			return tools.ErrorResult(fmt.Errorf("failed to parse JSON: %w", err)), nil
		}
		for _, key := range []string{"items", "inventory", "products"} {
			if raw, ok := wrapper[key]; ok {
				if err3 := json.Unmarshal(raw, &items); err3 == nil {
					break
				}
			}
		}
		if len(items) == 0 {
			return tools.ErrorResult(fmt.Errorf("no items found in JSON (expected array or {items:[...]})")), nil
		}
	}

	result := &clients.ImportResult{
		Source:    clients.PurchaseSource(source),
		TotalRows: len(items),
	}

	for _, src := range items {
		if src.Name == "" {
			result.Skipped++
			result.SkippedItems = append(result.SkippedItems, "(empty name)")
			continue
		}

		item := &clients.InventoryItem{
			Name:          src.Name,
			Category:      mapCategory(src.Category),
			ProductLine:   src.ProductLine,
			Condition:     mapCondition(src.Condition),
			AskingPrice:   src.AskingPrice,
			PurchasePrice: src.PurchasePrice,
			MSRP:          src.MSRP,
			Notes:         src.Notes,
			Specs:         src.Specs,
			Location:      src.Location,
			Model:         src.Model,
			ASIN:          src.ASIN,
			PurchaseDate:  src.PurchaseDate,
			Quantity:      src.Quantity,
		}
		if item.Quantity == 0 {
			item.Quantity = 1
		}
		if item.AskingPrice == 0 && src.CurrentValue > 0 {
			item.AskingPrice = src.CurrentValue
		}
		if item.PurchaseDate == "" {
			item.PurchaseDate = "imported"
		}
		// Auto-populate specs from catalog if ProductLine is set and Specs is empty
		if item.ProductLine != "" && len(item.Specs) == 0 {
			if catalogSpecs := LookupSpecs(item.ProductLine); catalogSpecs != nil {
				item.Specs = catalogSpecs
			}
		}

		if dryRun {
			result.Imported++
			result.NewItems = append(result.NewItems, fmt.Sprintf("[dry_run] %s (%s) $%.0f", item.Name, item.Category, item.AskingPrice))
			continue
		}

		added, err := client.AddItem(ctx, item)
		if err != nil {
			result.Errors++
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("Failed to add %q: %v", src.Name, err))
			continue
		}
		result.Imported++
		result.NewItems = append(result.NewItems, added.SKU)
	}

	return tools.JSONResult(map[string]interface{}{
		"dry_run": dryRun,
		"result":  result,
	}), nil
}

// ── Phase 4b: Batch Listing ──

func handleBatchListing(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	filter := &clients.InventoryFilter{
		Category: tools.GetStringParam(req, "category"),
		Limit:    tools.GetIntParam(req, "limit", 50),
	}
	if v := tools.GetStringParam(req, "status"); v != "" {
		filter.Status = clients.ListingStatus(v)
	}

	items, err := client.ListItems(ctx, filter)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	platformParam := tools.GetStringParam(req, "platform")
	platforms := []string{"fb_marketplace", "ebay", "hardwareswap"}
	if platformParam != "" && platformParam != "all" {
		platforms = []string{platformParam}
	}

	type listingOutput struct {
		SKU      string `json:"sku"`
		Name     string `json:"name"`
		Platform string `json:"platform"`
		Listing  string `json:"listing"`
	}

	var listings []listingOutput
	var errors []string

	for _, item := range items {
		for _, p := range platforms {
			text, err := RenderListing(&item, p)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s/%s: %v", item.SKU, p, err))
				continue
			}
			listings = append(listings, listingOutput{
				SKU:      item.SKU,
				Name:     item.Name,
				Platform: p,
				Listing:  text,
			})
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"count":    len(listings),
		"listings": listings,
		"errors":   errors,
	}), nil
}

// ── Phase 4c: Reprice ──

func handleReprice(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	invClient, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	sku, errResult := tools.RequireStringParam(req, "sku")
	if errResult != nil {
		return errResult, nil
	}

	dryRun := tools.GetBoolParam(req, "dry_run", true)
	strategy := tools.OptionalStringParam(req, "strategy", "fast_sale")

	item, err := invClient.GetItem(ctx, sku)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Build search query from product line or name
	searchQuery := item.Name
	if item.ProductLine != "" {
		searchQuery = item.ProductLine
	}

	// Try eBay price research
	ebayClient, err := clients.GetEbayClient()
	if err != nil {
		// eBay not configured — return suggestion without market data
		return tools.JSONResult(map[string]interface{}{
			"sku":               sku,
			"name":              item.Name,
			"current_price":     item.AskingPrice,
			"search_query":      searchQuery,
			"error":             "eBay client not configured — manual research needed",
			"search_suggestion": fmt.Sprintf("Search eBay sold listings for: %s", searchQuery),
		}), nil
	}

	analysis, err := ebayClient.SearchCompletedItems(ctx, searchQuery, 25)
	if err != nil {
		return tools.JSONResult(map[string]interface{}{
			"sku":               sku,
			"name":              item.Name,
			"current_price":     item.AskingPrice,
			"search_query":      searchQuery,
			"error":             fmt.Sprintf("eBay search failed: %v", err),
			"search_suggestion": fmt.Sprintf("Search eBay sold listings for: %s", searchQuery),
		}), nil
	}

	if analysis.ResultCount == 0 {
		return tools.JSONResult(map[string]interface{}{
			"sku":           sku,
			"name":          item.Name,
			"current_price": item.AskingPrice,
			"search_query":  searchQuery,
			"message":       "No completed listings found — cannot compute market price",
		}), nil
	}

	median := analysis.MedianPrice

	var suggestedPrice float64
	switch strategy {
	case "fast_sale":
		suggestedPrice = median * 0.95
	case "max_profit":
		suggestedPrice = median * 1.05
	default:
		suggestedPrice = median
	}

	result := map[string]interface{}{
		"sku":             sku,
		"name":            item.Name,
		"current_price":   item.AskingPrice,
		"search_query":    searchQuery,
		"sample_size":     analysis.ResultCount,
		"median_sold":     fmt.Sprintf("$%.2f", median),
		"min_sold":        fmt.Sprintf("$%.2f", analysis.MinPrice),
		"max_sold":        fmt.Sprintf("$%.2f", analysis.MaxPrice),
		"suggested_price": fmt.Sprintf("$%.2f", suggestedPrice),
		"strategy":        strategy,
		"price_change":    fmt.Sprintf("$%.2f", suggestedPrice-item.AskingPrice),
		"dry_run":         dryRun,
	}

	if !dryRun {
		_, err := invClient.UpdateItem(ctx, sku, map[string]interface{}{
			"asking_price": suggestedPrice,
		})
		if err != nil {
			result["update_error"] = err.Error()
		} else {
			result["updated"] = true
		}
	}

	return tools.JSONResult(result), nil
}
