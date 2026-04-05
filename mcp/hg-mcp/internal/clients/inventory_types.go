// Package clients provides API clients for external services.
package clients

import "time"

// ItemCondition represents the condition of an inventory item
type ItemCondition string

const (
	ConditionNew      ItemCondition = "new"
	ConditionLikeNew  ItemCondition = "like_new"
	ConditionGood     ItemCondition = "good"
	ConditionFair     ItemCondition = "fair"
	ConditionPoor     ItemCondition = "poor"
	ConditionForParts ItemCondition = "for_parts"
)

// ListingStatus represents the sale status of an item
type ListingStatus string

const (
	StatusNotListed     ListingStatus = "not_listed"
	StatusPendingReview ListingStatus = "pending_review"
	StatusListed        ListingStatus = "listed"
	StatusSold          ListingStatus = "sold"
	StatusKeeping       ListingStatus = "keeping"
)

// PurchaseSource represents where an item was purchased
type PurchaseSource string

const (
	SourceAmazon      PurchaseSource = "amazon"
	SourceNewegg      PurchaseSource = "newegg"
	SourceEbay        PurchaseSource = "ebay"
	SourceBestBuy     PurchaseSource = "bestbuy"
	SourceMicroCenter PurchaseSource = "microcenter"
	SourceBHPhoto     PurchaseSource = "bhphoto"
	SourceOther       PurchaseSource = "other"
)

// InventoryItem represents an electronics inventory item.
// Maps to Google Sheet "Inventory" tab columns A-V.
type InventoryItem struct {
	// Sheet columns A-N (existing)
	RowNum        int     `json:"row_num"`                     // col A: #
	Category      string  `json:"category"`                    // col B
	Name          string  `json:"name"`                        // col C: Product
	Model         string  `json:"model,omitempty"`             // col D: Model / SKU
	Quantity      int     `json:"quantity"`                    // col E: Qty
	Condition     string  `json:"condition"`                   // col F
	PurchaseDate  string  `json:"purchase_date,omitempty"`     // col G: e.g. "Apr 2025"
	ASIN          string  `json:"asin,omitempty"`              // col H: Amazon ASIN
	MSRP          float64 `json:"msrp,omitempty"`              // col I
	CurrentRetail float64 `json:"current_retail,omitempty"`    // col J
	AskingPrice   float64 `json:"asking_price,omitempty"`      // col K: Recommended FB Price
	TotalRevenue  float64 `json:"total_est_revenue,omitempty"` // col L: = K * E (computed)
	ListingStatus string  `json:"listing_status"`              // col M: Status
	Notes         string  `json:"notes,omitempty"`             // col N

	// Sheet columns O-T (appended)
	SoldPrice     float64 `json:"sold_price,omitempty"`     // col O
	SoldDate      string  `json:"sold_date,omitempty"`      // col P
	SoldPlatform  string  `json:"sold_platform,omitempty"`  // col Q
	Location      string  `json:"location,omitempty"`       // col R
	SmartData     string  `json:"smart_data,omitempty"`     // col S: JSON string
	PurchasePrice float64 `json:"purchase_price,omitempty"` // col T

	// Sheet columns U-W (product line + specs + on-hand)
	ProductLine string            `json:"product_line,omitempty"` // col U: groups identical products (e.g. "Samsung 990 Pro")
	Specs       map[string]string `json:"specs,omitempty"`        // col V: JSON-encoded tech specs (e.g. {"vram":"24GB","tdp":"450W"})
	OnHand      bool              `json:"on_hand"`                // col W: physically located and ready to list

	// Sheet columns X-Z (return/dispute tracking + bundles)
	ReturnStatus string `json:"return_status,omitempty"` // col X: none, requested, approved, completed, denied
	DisputeNotes string `json:"dispute_notes,omitempty"` // col Y: free-text notes about returns/disputes
	BundleID     string `json:"bundle_id,omitempty"`     // col Z: groups items sold as a bundle/lot

	// Computed / in-memory only (not stored in sheet)
	SKU            string         `json:"sku"` // Generated: HW-001, HW-002, etc.
	Description    string         `json:"description,omitempty"`
	Subcategory    string         `json:"subcategory,omitempty"`
	Brand          string         `json:"brand,omitempty"`
	SerialNumber   string         `json:"serial_number,omitempty"`
	PurchaseSource PurchaseSource `json:"purchase_source,omitempty"`
	VendorName     string         `json:"vendor_name,omitempty"` // Rich vendor display name (e.g. "Amazon.com", "Micro Center - Dallas")
	VendorURL      string         `json:"vendor_url,omitempty"`  // Vendor website or order page URL
	OrderID        string         `json:"order_id,omitempty"`
	ProductURL     string         `json:"product_url,omitempty"`
	CurrentValue   float64        `json:"current_value,omitempty"`
	EbayListingID  string         `json:"ebay_listing_id,omitempty"`
	EbayURL        string         `json:"ebay_url,omitempty"`
	ListedPrice    float64        `json:"listed_price,omitempty"`
	PrimaryImage   string         `json:"primary_image,omitempty"`
	Images         []string       `json:"images,omitempty"`
	WarrantyExpiry *time.Time     `json:"warranty_expiry,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// SaleRecord represents a completed sale in the Sales tab.
type SaleRecord struct {
	SaleID        string  `json:"sale_id"`        // col A
	ItemNum       int     `json:"item_num"`       // col B: row number from inventory
	ItemName      string  `json:"item_name"`      // col C
	QuantitySold  int     `json:"quantity_sold"`  // col D
	SoldPrice     float64 `json:"sold_price"`     // col E: per unit
	PurchasePrice float64 `json:"purchase_price"` // col F: per unit
	Revenue       float64 `json:"revenue"`        // col G: E * D
	Cost          float64 `json:"cost"`           // col H: F * D
	ShippingCost  float64 `json:"shipping_cost"`  // col I
	PlatformFees  float64 `json:"platform_fees"`  // col J
	NetProfit     float64 `json:"net_profit"`     // col K: G - H - I - J
	Platform      string  `json:"platform"`       // col L
	BuyerInfo     string  `json:"buyer_info"`     // col M
	Notes         string  `json:"notes"`          // col N
	Date          string  `json:"date"`           // col O
}

// SalesSummary represents aggregated sales data.
type SalesSummary struct {
	TotalSales   int     `json:"total_sales"`
	TotalRevenue float64 `json:"total_revenue"`
	TotalCost    float64 `json:"total_cost"`
	TotalFees    float64 `json:"total_fees"`
	TotalProfit  float64 `json:"total_profit"`
	UnitsSold    int     `json:"units_sold"`
}

// InventoryConfig represents key-value config from the Config tab.
type InventoryConfig struct {
	TargetRevenue float64 `json:"target_revenue"`
	Created       string  `json:"created"`
	LastUpdated   string  `json:"last_updated"`
}

// InventoryOrder represents a purchase order from a source
type InventoryOrder struct {
	OrderID    string         `json:"order_id"`
	Source     PurchaseSource `json:"source"`
	OrderDate  time.Time      `json:"order_date"`
	Total      float64        `json:"total"`
	ItemCount  int            `json:"item_count"`
	ImportedAt time.Time      `json:"imported_at"`
}

// InventoryListing represents a marketplace listing
type InventoryListing struct {
	ListingID   string     `json:"listing_id"`
	SKU         string     `json:"sku"`
	Platform    string     `json:"platform"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	URL         string     `json:"url,omitempty"`
	Status      string     `json:"status"`
	Views       int        `json:"views,omitempty"`
	Watchers    int        `json:"watchers,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty"`
}

// InventoryCategory represents a category with subcategories
type InventoryCategory struct {
	Name          string   `json:"name"`
	Subcategories []string `json:"subcategories,omitempty"`
	ItemCount     int      `json:"item_count"`
}

// InventorySummary represents an overview of the inventory
type InventorySummary struct {
	TotalItems    int             `json:"total_items"`
	TotalValue    float64         `json:"total_value"`
	TotalCost     float64         `json:"total_cost"`
	ByCategory    map[string]int  `json:"by_category"`
	ByStatus      map[string]int  `json:"by_status"`
	ByCondition   map[string]int  `json:"by_condition"`
	ByLocation    map[string]int  `json:"by_location"`
	RecentlyAdded []InventoryItem `json:"recently_added,omitempty"`
	TopValueItems []InventoryItem `json:"top_value_items,omitempty"`
}

// InventoryFilter represents filters for listing items
type InventoryFilter struct {
	Category    string         `json:"category,omitempty"`
	Subcategory string         `json:"subcategory,omitempty"`
	Status      ListingStatus  `json:"status,omitempty"`
	Condition   ItemCondition  `json:"condition,omitempty"`
	Location    string         `json:"location,omitempty"`
	Source      PurchaseSource `json:"source,omitempty"`
	MinPrice    float64        `json:"min_price,omitempty"`
	MaxPrice    float64        `json:"max_price,omitempty"`
	Brand       string         `json:"brand,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Query       string         `json:"query,omitempty"`
	Limit       int            `json:"limit,omitempty"`
	Offset      int            `json:"offset,omitempty"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	Source        PurchaseSource `json:"source"`
	TotalRows     int            `json:"total_rows"`
	Imported      int            `json:"imported"`
	Skipped       int            `json:"skipped"`
	Errors        int            `json:"errors"`
	NewItems      []string       `json:"new_items,omitempty"`
	SkippedItems  []string       `json:"skipped_items,omitempty"`
	ErrorMessages []string       `json:"error_messages,omitempty"`
}

// CategorizationResult represents the result of recategorizing an item
type CategorizationResult struct {
	SKU            string `json:"sku"`
	Name           string `json:"name"`
	OldCategory    string `json:"old_category"`
	NewCategory    string `json:"new_category"`
	OldSubcategory string `json:"old_subcategory,omitempty"`
	NewSubcategory string `json:"new_subcategory,omitempty"`
	OldBrand       string `json:"old_brand,omitempty"`
	NewBrand       string `json:"new_brand,omitempty"`
	Changed        bool   `json:"changed"`
}

// InventoryDuplicateGroup represents a group of potential duplicate inventory items
type InventoryDuplicateGroup struct {
	MatchType  string          `json:"match_type"`
	MatchValue string          `json:"match_value"`
	Items      []InventoryItem `json:"items"`
	Count      int             `json:"count"`
}

// AmazonOrderRow represents a row from Amazon order history CSV
type AmazonOrderRow struct {
	OrderDate    string `csv:"Order Date"`
	OrderID      string `csv:"Order ID"`
	Title        string `csv:"Title"`
	Category     string `csv:"Category"`
	ASIN         string `csv:"ASIN/ISBN"`
	Seller       string `csv:"Seller"`
	Quantity     string `csv:"Quantity"`
	ItemSubtotal string `csv:"Item Subtotal"`
	ItemTotal    string `csv:"Item Total"`
	ShipmentDate string `csv:"Shipment Date"`
	ReleaseDate  string `csv:"Release Date"`
	Condition    string `csv:"Condition"`
	ListPrice    string `csv:"List Price Per Unit"`
}

// NeweggOrderRow represents a row from Newegg order history
type NeweggOrderRow struct {
	OrderNumber string `csv:"Order Number"`
	OrderDate   string `csv:"Order Date"`
	OrderStatus string `csv:"Order Status"`
	ItemName    string `csv:"Item Name"`
	ItemNumber  string `csv:"Item Number"`
	Quantity    string `csv:"Quantity"`
	UnitPrice   string `csv:"Unit Price"`
	Subtotal    string `csv:"Subtotal"`
}

// FBMarketplaceContent represents generated content for FB Marketplace
type FBMarketplaceContent struct {
	Title       string   `json:"title"`
	Price       float64  `json:"price"`
	Category    string   `json:"category"`
	Condition   string   `json:"condition"`
	Description string   `json:"description"`
	ImageURLs   []string `json:"image_urls,omitempty"`
	Ready       bool     `json:"ready"`
	Message     string   `json:"message,omitempty"`
}

// PriceSuggestion represents AI-suggested pricing
type PriceSuggestion struct {
	SuggestedPrice float64  `json:"suggested_price"`
	MinPrice       float64  `json:"min_price"`
	MaxPrice       float64  `json:"max_price"`
	Confidence     float64  `json:"confidence"`
	Reasoning      string   `json:"reasoning"`
	ComparableURLs []string `json:"comparable_urls,omitempty"`
}

// GmailOrderEmail represents a parsed order confirmation email
type GmailOrderEmail struct {
	EmailID string            `json:"email_id"`
	Subject string            `json:"subject"`
	From    string            `json:"from"`
	Date    time.Time         `json:"date"`
	Source  PurchaseSource    `json:"source"`
	OrderID string            `json:"order_id,omitempty"`
	Total   float64           `json:"total,omitempty"`
	Items   []ParsedOrderItem `json:"items,omitempty"`
	RawBody string            `json:"-"`
}

// ParsedOrderItem represents an item parsed from an order email
type ParsedOrderItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price,omitempty"`
	ASIN     string  `json:"asin,omitempty"`
	ItemID   string  `json:"item_id,omitempty"`
	URL      string  `json:"url,omitempty"`
}

// GmailImportConfig configures Gmail order import
type GmailImportConfig struct {
	SinceDays  int      `json:"since_days,omitempty"`
	Sources    []string `json:"sources,omitempty"`
	MaxResults int      `json:"max_results,omitempty"`
	DryRun     bool     `json:"dry_run,omitempty"`
	IncludeAll bool     `json:"include_all,omitempty"`
}

// DefaultCategories returns the predefined inventory categories
func DefaultCategories() []InventoryCategory {
	return []InventoryCategory{
		{Name: "GPU", Subcategories: []string{"Desktop", "Mobile", "Workstation"}},
		{Name: "CPU", Subcategories: []string{"Intel", "AMD", "ARM"}},
		{Name: "RAM", Subcategories: []string{"DDR4", "DDR5", "Laptop", "Server"}},
		{Name: "Storage", Subcategories: []string{"SSD", "HDD", "NVMe", "External"}},
		{Name: "NVMe Gen5", Subcategories: []string{}},
		{Name: "Motherboard", Subcategories: []string{"ATX", "mATX", "ITX", "Server"}},
		{Name: "PSU", Subcategories: []string{"ATX", "SFX", "Modular"}},
		{Name: "Case", Subcategories: []string{"Full Tower", "Mid Tower", "SFF", "Server"}},
		{Name: "Cooling", Subcategories: []string{"Air", "AIO", "Custom Loop"}},
		{Name: "Peripherals", Subcategories: []string{"Keyboard", "Mouse", "Monitor", "Audio"}},
		{Name: "Networking", Subcategories: []string{"Router", "Switch", "NIC", "Cable"}},
		{Name: "Thunderbolt", Subcategories: []string{"Dock", "Hub", "Cable", "Enclosure"}},
		{Name: "KVM", Subcategories: []string{"Switch", "Cables", "USB"}},
		{Name: "Cables", Subcategories: []string{"Display", "USB", "Power", "Network"}},
		{Name: "Components", Subcategories: []string{"Fans", "Brackets", "Adapters"}},
		{Name: "Complete Systems", Subcategories: []string{"Desktop", "Laptop", "Server"}},
		{Name: "Other", Subcategories: []string{}},
	}
}
