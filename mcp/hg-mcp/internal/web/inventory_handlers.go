package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

// handleInventoryItems handles GET /api/v1/inventory (list) and POST /api/v1/inventory (create)
func (s *Server) handleInventoryItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleInventoryList(w, r)
	case http.MethodPost:
		s.handleInventoryCreate(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleInventoryList handles GET /api/v1/inventory
func (s *Server) handleInventoryList(w http.ResponseWriter, r *http.Request) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get inventory client: "+err.Error())
		return
	}

	filter := parseInventoryFilter(r)

	items, err := client.ListItems(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list items: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

// handleInventoryCreate handles POST /api/v1/inventory
func (s *Server) handleInventoryCreate(w http.ResponseWriter, r *http.Request) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get inventory client: "+err.Error())
		return
	}

	var item clients.InventoryItem
	if err := readJSON(r, &item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if item.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if item.Category == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}
	if item.Location == "" {
		writeError(w, http.StatusBadRequest, "location is required")
		return
	}

	added, err := client.AddItem(r.Context(), &item)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add item: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, added)
}

// handleInventoryItem handles /api/v1/inventory/{sku} for GET, PUT, DELETE
func (s *Server) handleInventoryItem(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/inventory/")
	parts := strings.Split(path, "/")
	sku := parts[0]

	if sku == "" {
		writeError(w, http.StatusBadRequest, "SKU is required")
		return
	}

	// Handle special endpoints
	if sku == "summary" {
		s.handleInventorySummary(w, r)
		return
	}
	if sku == "categories" {
		s.handleInventoryCategories(w, r)
		return
	}
	if sku == "locations" {
		s.handleInventoryLocations(w, r)
		return
	}
	if sku == "bulk-delete" {
		s.handleInventoryBulkDelete(w, r)
		return
	}

	// Handle images sub-path
	if len(parts) > 1 && parts[1] == "images" {
		s.handleInventoryImages(w, r, sku)
		return
	}

	client, err := clients.GetInventoryClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get inventory client: "+err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		item, err := client.GetItem(r.Context(), sku)
		if err != nil {
			writeError(w, http.StatusNotFound, "item not found: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, item)

	case http.MethodPut:
		var updates map[string]interface{}
		if err := readJSON(r, &updates); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		item, err := client.UpdateItem(r.Context(), sku, updates)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update item: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, item)

	case http.MethodDelete:
		if err := client.DeleteItem(r.Context(), sku); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete item: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "sku": sku})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleInventorySummary handles GET /api/v1/inventory/summary
func (s *Server) handleInventorySummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client, err := clients.GetInventoryClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get inventory client: "+err.Error())
		return
	}

	summary, err := client.GetSummary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get summary: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// handleInventoryCategories handles GET /api/v1/inventory/categories
func (s *Server) handleInventoryCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client, err := clients.GetInventoryClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get inventory client: "+err.Error())
		return
	}

	categories, err := client.GetCategories(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get categories: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

// handleInventoryLocations handles GET /api/v1/inventory/locations
func (s *Server) handleInventoryLocations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	client, err := clients.GetInventoryClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get inventory client: "+err.Error())
		return
	}

	locations, err := client.GetLocations(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get locations: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, locations)
}

// handleInventoryBulkDelete handles POST /api/v1/inventory/bulk-delete
func (s *Server) handleInventoryBulkDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		SKUs    []string `json:"skus"`
		Confirm bool     `json:"confirm"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if !req.Confirm {
		writeError(w, http.StatusBadRequest, "confirm must be true to delete items")
		return
	}

	if len(req.SKUs) == 0 {
		writeError(w, http.StatusBadRequest, "skus array is required")
		return
	}

	client, err := clients.GetInventoryClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get inventory client: "+err.Error())
		return
	}

	var deleted []string
	var errors []string

	for _, sku := range req.SKUs {
		if err := client.DeleteItem(r.Context(), sku); err != nil {
			errors = append(errors, sku+": "+err.Error())
		} else {
			deleted = append(deleted, sku)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"deleted":       len(deleted),
		"errors":        len(errors),
		"deleted_skus":  deleted,
		"error_details": errors,
	})
}

// handleInventoryImages handles image operations for an item
func (s *Server) handleInventoryImages(w http.ResponseWriter, r *http.Request, sku string) {
	client, err := clients.GetInventoryClient()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get inventory client: "+err.Error())
		return
	}

	switch r.Method {
	case http.MethodGet:
		images, err := client.ListImages(r.Context(), sku)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list images: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"sku":    sku,
			"images": images,
			"count":  len(images),
		})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// parseInventoryFilter parses query parameters into an InventoryFilter
func parseInventoryFilter(r *http.Request) *clients.InventoryFilter {
	q := r.URL.Query()

	filter := &clients.InventoryFilter{
		Category:    q.Get("category"),
		Subcategory: q.Get("subcategory"),
		Location:    q.Get("location"),
		Brand:       q.Get("brand"),
		Query:       q.Get("query"),
	}

	if status := q.Get("status"); status != "" {
		filter.Status = clients.ListingStatus(status)
	}

	if condition := q.Get("condition"); condition != "" {
		filter.Condition = clients.ItemCondition(condition)
	}

	if source := q.Get("source"); source != "" {
		filter.Source = clients.PurchaseSource(source)
	}

	if minPrice := q.Get("min_price"); minPrice != "" {
		if v, err := strconv.ParseFloat(minPrice, 64); err == nil {
			filter.MinPrice = v
		}
	}
	if maxPrice := q.Get("max_price"); maxPrice != "" {
		if v, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			filter.MaxPrice = v
		}
	}

	if limit := q.Get("limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			filter.Limit = v
		}
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	if offset := q.Get("offset"); offset != "" {
		if v, err := strconv.Atoi(offset); err == nil {
			filter.Offset = v
		}
	}

	if tags := q.Get("tags"); tags != "" {
		filter.Tags = strings.Split(tags, ",")
	}

	return filter
}
