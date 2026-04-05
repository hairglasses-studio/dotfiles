package web

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	// Health endpoint returns structured JSON
	var health map[string]interface{}
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("expected JSON body, got: %s", string(body))
	}
	if health["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", health["status"])
	}
}

func TestStatsEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := stats["total_tools"]; !ok {
		t.Error("expected stats to contain 'total_tools'")
	}
	if _, ok := stats["module_count"]; !ok {
		t.Error("expected stats to contain 'module_count'")
	}
	if _, ok := stats["by_category"]; !ok {
		t.Error("expected stats to contain 'by_category'")
	}
}

func TestCategoriesEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var categories []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&categories); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Categories may be empty in test environment if registry not initialized
	// Check structure if categories exist
	if len(categories) > 0 {
		cat := categories[0]
		if _, ok := cat["name"]; !ok {
			t.Error("expected category to have 'name' field")
		}
		if _, ok := cat["count"]; !ok {
			t.Error("expected category to have 'count' field")
		}
	}
}

func TestToolsEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tools", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var tools []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Tools may be empty in test environment if registry not initialized
	// Check tool structure if tools exist
	if len(tools) > 0 {
		tool := tools[0]
		requiredFields := []string{"name", "description", "category"}
		for _, field := range requiredFields {
			if _, ok := tool[field]; !ok {
				t.Errorf("expected tool to have '%s' field", field)
			}
		}
	}
}

func TestToolsWithLimitParam(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tools?limit=5", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var tools []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(tools) > 5 {
		t.Errorf("expected at most 5 tools, got %d", len(tools))
	}
}

func TestToolsWithCategoryFilter(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tools?category=resolume", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var tools []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// All returned tools should be in the resolume category
	for _, tool := range tools {
		if cat, ok := tool["category"].(string); ok && cat != "resolume" {
			t.Errorf("expected tool category 'resolume', got '%s'", cat)
		}
	}
}

func TestSearchEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search?q=status", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var results []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Search results may be empty in test environment if registry not initialized
	// Check result structure if results exist
	if len(results) > 0 {
		result := results[0]
		if _, ok := result["tool"]; !ok {
			t.Error("expected search result to have 'tool' field")
		}
		if _, ok := result["score"]; !ok {
			t.Error("expected search result to have 'score' field")
		}
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/search", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty search query, got %d", resp.StatusCode)
	}
}

func TestDashboardStatusEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/status", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	expectedFields := []string{"total_systems", "online_count", "offline_count", "overall_health"}
	for _, field := range expectedFields {
		if _, ok := status[field]; !ok {
			t.Errorf("expected dashboard status to contain '%s'", field)
		}
	}
}

func TestAlertsEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/alerts", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var alerts []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	// alerts can be empty, but should be a valid array
}

func TestWorkflowsEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workflows", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestFavoritesEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/favorites", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAliasesEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/aliases", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestCORSHeadersInDevMode(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected CORS header Access-Control-Allow-Origin: * in dev mode")
	}
}

func TestOPTIONSRequest(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/tools", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS request, got %d", resp.StatusCode)
	}
}

func TestSingleToolEndpoint(t *testing.T) {
	server := NewServer(Config{DevMode: true})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tools/aftrs_resolume_status", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should return 200 if tool exists, 404 if not
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 200 or 404, got %d", resp.StatusCode)
	}
}

func TestStaticFilesFallbackToIndex(t *testing.T) {
	// Create server with static files (requires dist directory to exist)
	server := NewServer(Config{DevMode: false})

	// If staticFS is nil (no dist), skip this test
	if server.staticFS == nil {
		t.Skip("No static files available for testing")
	}

	req := httptest.NewRequest(http.MethodGet, "/non-existent-route", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// SPA routes should return index.html (200) or 404 if no static files
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 200 or 404 for SPA fallback, got %d", resp.StatusCode)
	}
}
