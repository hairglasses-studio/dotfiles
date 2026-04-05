package inventory

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// InventoryItemsTotal tracks total inventory item operations
	InventoryItemsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_items_total",
			Help: "Total number of inventory item operations",
		},
		[]string{"operation", "category", "status"},
	)

	// InventoryItemsGauge tracks current inventory count by category/status
	InventoryItemsGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "inventory_items_current",
			Help: "Current number of inventory items",
		},
		[]string{"category", "status", "location"},
	)

	// InventoryImportsTotal tracks import operations
	InventoryImportsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_imports_total",
			Help: "Total number of import operations",
		},
		[]string{"source", "status"},
	)

	// InventoryImportItemsTotal tracks items imported per source
	InventoryImportItemsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_import_items_total",
			Help: "Total number of items imported",
		},
		[]string{"source"},
	)

	// InventoryImportDuration tracks import operation duration
	InventoryImportDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "inventory_import_duration_seconds",
			Help:    "Duration of import operations in seconds",
			Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120},
		},
		[]string{"source"},
	)

	// InventoryListingsTotal tracks listing operations
	InventoryListingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_listings_total",
			Help: "Total number of listing operations",
		},
		[]string{"platform", "status"},
	)

	// InventoryEbayAPITotal tracks eBay API calls
	InventoryEbayAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_ebay_api_total",
			Help: "Total number of eBay API calls",
		},
		[]string{"endpoint", "status"},
	)

	// InventoryEbayAPIDuration tracks eBay API call duration
	InventoryEbayAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "inventory_ebay_api_duration_seconds",
			Help:    "Duration of eBay API calls in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
		[]string{"endpoint"},
	)

	// InventorySheetsAPITotal tracks Google Sheets API operations
	InventorySheetsAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_sheets_api_total",
			Help: "Total number of Google Sheets API operations",
		},
		[]string{"operation", "status"},
	)

	// InventorySheetsAPIDuration tracks Google Sheets API operation duration
	InventorySheetsAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "inventory_sheets_api_duration_seconds",
			Help:    "Duration of Google Sheets API operations in seconds",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"operation"},
	)

	// InventoryLocalImageTotal tracks local image filesystem operations
	InventoryLocalImageTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_local_image_total",
			Help: "Total number of local image operations",
		},
		[]string{"operation", "status"},
	)

	// InventorySalesTotal tracks sales recorded
	InventorySalesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_sales_total",
			Help: "Total number of sales recorded",
		},
		[]string{"platform"},
	)

	// InventoryRevenueGauge tracks cumulative revenue
	InventoryRevenueGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "inventory_revenue_dollars",
			Help: "Cumulative revenue from sales in dollars",
		},
	)

	// InventoryValueTotal tracks total inventory value
	InventoryValueTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "inventory_value_dollars",
			Help: "Total inventory value in dollars",
		},
		[]string{"type"},
	)

	// InventoryGmailTotal tracks Gmail API operations
	InventoryGmailTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_gmail_total",
			Help: "Total number of Gmail API operations",
		},
		[]string{"operation", "status"},
	)

	// InventoryGmailOrdersFound tracks orders found via Gmail
	InventoryGmailOrdersFound = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_gmail_orders_found",
			Help: "Total number of orders found via Gmail parsing",
		},
		[]string{"source"},
	)

	// InventoryErrorsTotal tracks errors by type
	InventoryErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_errors_total",
			Help: "Total number of inventory errors",
		},
		[]string{"operation", "error_type"},
	)

	// InventoryLastOperation tracks timestamp of last operation
	InventoryLastOperation = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "inventory_last_operation_timestamp",
			Help: "Unix timestamp of last operation",
		},
		[]string{"operation"},
	)
)

// RecordItemOperation records an inventory item operation metric
func RecordItemOperation(operation, category, status string) {
	InventoryItemsTotal.WithLabelValues(operation, category, status).Inc()
	InventoryLastOperation.WithLabelValues(operation).SetToCurrentTime()
}

// RecordImportResult records import operation metrics
func RecordImportResult(source string, itemsImported int, duration float64, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	InventoryImportsTotal.WithLabelValues(source, status).Inc()
	InventoryImportItemsTotal.WithLabelValues(source).Add(float64(itemsImported))
	InventoryImportDuration.WithLabelValues(source).Observe(duration)
	InventoryLastOperation.WithLabelValues("import_" + source).SetToCurrentTime()
}

// RecordEbayAPICall records an eBay API call metric
func RecordEbayAPICall(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	InventoryEbayAPITotal.WithLabelValues(endpoint, status).Inc()
	InventoryEbayAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordSheetsOperation records a Google Sheets API operation metric
func RecordSheetsOperation(operation string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	InventorySheetsAPITotal.WithLabelValues(operation, status).Inc()
	InventorySheetsAPIDuration.WithLabelValues(operation).Observe(duration)
}

// RecordLocalImageOperation records a local image filesystem operation
func RecordLocalImageOperation(operation string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	InventoryLocalImageTotal.WithLabelValues(operation, status).Inc()
}

// RecordGmailOperation records a Gmail API operation metric
func RecordGmailOperation(operation string, success bool) {
	status := "success"
	if !success {
		status = "failure"
	}
	InventoryGmailTotal.WithLabelValues(operation, status).Inc()
}

// RecordError records an inventory error
func RecordError(operation, errorType string) {
	InventoryErrorsTotal.WithLabelValues(operation, errorType).Inc()
}

// UpdateInventoryGauge updates the current inventory count gauge
func UpdateInventoryGauge(category, status, location string, count float64) {
	InventoryItemsGauge.WithLabelValues(category, status, location).Set(count)
}

// UpdateInventoryValue updates the inventory value gauge
func UpdateInventoryValue(valueType string, value float64) {
	InventoryValueTotal.WithLabelValues(valueType).Set(value)
}
