// Package clients provides API clients for external services.
package clients

import (
	"context"
	"io"
)

// InventoryBackend defines the interface for inventory storage backends.
// Implementations: SheetsInventoryClient (Google Sheets), InventoryClient (DynamoDB, legacy).
type InventoryBackend interface {
	// CRUD
	ListItems(ctx context.Context, filter *InventoryFilter) ([]InventoryItem, error)
	GetItem(ctx context.Context, sku string) (*InventoryItem, error)
	AddItem(ctx context.Context, item *InventoryItem) (*InventoryItem, error)
	UpdateItem(ctx context.Context, sku string, updates map[string]interface{}) (*InventoryItem, error)
	DeleteItem(ctx context.Context, sku string) error
	SearchItems(ctx context.Context, query string, limit int) ([]InventoryItem, error)

	// Analytics
	GetSummary(ctx context.Context) (*InventorySummary, error)
	GetCategories(ctx context.Context) ([]InventoryCategory, error)
	GetLocations(ctx context.Context) (map[string]int, error)
	GetStaleItems(ctx context.Context, days int) ([]InventoryItem, error)

	// Listing
	MarkAsSold(ctx context.Context, sku string, soldPrice float64, platform string) (*InventoryItem, error)

	// Import
	ImportAmazonCSV(ctx context.Context, csvData io.Reader) (*ImportResult, error)
	ImportNeweggCSV(ctx context.Context, csvData io.Reader) (*ImportResult, error)

	// Data quality
	RecategorizeItem(item *InventoryItem) *CategorizationResult
	RecategorizeItems(ctx context.Context, filter *InventoryFilter, applyChanges bool) ([]CategorizationResult, error)
	FindDuplicates(ctx context.Context, matchType string) ([]InventoryDuplicateGroup, error)

	// Images (local filesystem)
	UploadImage(ctx context.Context, sku string, imageData []byte, filename string) (string, error)
	ListImages(ctx context.Context, sku string) ([]string, error)
	DeleteImage(ctx context.Context, sku, filename string) error

	// Sales (from hw-resale)
	RecordSale(ctx context.Context, sale *SaleRecord) (*SaleRecord, error)
	GetSalesLog(ctx context.Context) ([]SaleRecord, error)
	GetSalesSummary(ctx context.Context) (*SalesSummary, error)

	// Config
	GetConfig(ctx context.Context) (*InventoryConfig, error)
	UpdateConfig(ctx context.Context, key, value string) error

	// State
	IsConfigured() bool
}
