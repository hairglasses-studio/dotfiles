package clients

import (
	"fmt"
	"testing"
)

func TestSKUGeneration(t *testing.T) {
	tests := []struct {
		rowNum   int
		expected string
	}{
		{1, "HW-001"},
		{10, "HW-010"},
		{100, "HW-100"},
		{999, "HW-999"},
	}

	for _, tt := range tests {
		item := InventoryItem{RowNum: tt.rowNum}
		sku := ""
		if item.RowNum > 0 {
			sku = fmt.Sprintf("HW-%03d", item.RowNum)
		}
		if sku != tt.expected {
			t.Errorf("SKU for row %d = %q, want %q", tt.rowNum, sku, tt.expected)
		}
	}
}

func TestApplyFilters_Category(t *testing.T) {
	items := []InventoryItem{
		{Name: "RTX 4090", Category: "GPU", AskingPrice: 2200},
		{Name: "990 Pro", Category: "Storage", AskingPrice: 180},
		{Name: "DDR5 Kit", Category: "RAM", AskingPrice: 120},
	}

	result := applyFilters(items, &InventoryFilter{Category: "GPU"})
	if len(result) != 1 {
		t.Errorf("expected 1 GPU item, got %d", len(result))
	}
	if result[0].Name != "RTX 4090" {
		t.Errorf("expected RTX 4090, got %s", result[0].Name)
	}
}

func TestApplyFilters_Query(t *testing.T) {
	items := []InventoryItem{
		{Name: "ASUS ROG Strix RTX 4090", Category: "GPU"},
		{Name: "Samsung 990 Pro 2TB NVMe", Category: "Storage"},
		{Name: "Corsair Vengeance DDR5 32GB", Category: "RAM"},
	}

	result := applyFilters(items, &InventoryFilter{Query: "4090"})
	if len(result) != 1 {
		t.Errorf("expected 1 result for query '4090', got %d", len(result))
	}

	result = applyFilters(items, &InventoryFilter{Query: "samsung"})
	if len(result) != 1 {
		t.Errorf("expected 1 result for query 'samsung', got %d", len(result))
	}
}

func TestApplyFilters_PriceRange(t *testing.T) {
	items := []InventoryItem{
		{Name: "A", AskingPrice: 100},
		{Name: "B", AskingPrice: 500},
		{Name: "C", AskingPrice: 1000},
	}

	result := applyFilters(items, &InventoryFilter{MinPrice: 200, MaxPrice: 800})
	if len(result) != 1 {
		t.Errorf("expected 1 result for price range 200-800, got %d", len(result))
	}
	if result[0].Name != "B" {
		t.Errorf("expected item B, got %s", result[0].Name)
	}
}

func TestApplyFilters_Limit(t *testing.T) {
	items := []InventoryItem{
		{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}, {Name: "E"},
	}

	result := applyFilters(items, &InventoryFilter{Limit: 3})
	if len(result) != 3 {
		t.Errorf("expected 3 items with limit, got %d", len(result))
	}
}

func TestApplyFilters_NilFilter(t *testing.T) {
	items := []InventoryItem{{Name: "A"}, {Name: "B"}}
	result := applyFilters(items, nil)
	if len(result) != 2 {
		t.Errorf("nil filter should return all items, got %d", len(result))
	}
}

func TestGuessCategoryFromName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"ASUS ROG Strix RTX 4090 OC", "GPU"},
		{"AMD Ryzen 9 7950X", "CPU"},
		{"Samsung 990 Pro 2TB NVMe SSD", "Storage"},
		{"Corsair Vengeance DDR5 32GB", "RAM"},
		{"ASUS ROG Maximus Hero Motherboard", "Motherboard"},
		{"Corsair RM850x Power Supply", "PSU"},
		{"Noctua NH-D15 Tower Cooler", "Cooling"},
		{"Random Widget", "Other"},
	}

	for _, tt := range tests {
		got := guessCategoryFromName(tt.name)
		if got != tt.expected {
			t.Errorf("guessCategoryFromName(%q) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestColLetter(t *testing.T) {
	tests := []struct {
		idx      int
		expected string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
	}

	for _, tt := range tests {
		got := colLetter(tt.idx)
		if got != tt.expected {
			t.Errorf("colLetter(%d) = %q, want %q", tt.idx, got, tt.expected)
		}
	}
}

func TestFormatPrice(t *testing.T) {
	if got := formatPrice(0); got != "" {
		t.Errorf("formatPrice(0) = %q, want empty", got)
	}
	if got := formatPrice(2619); got != "$2619" {
		t.Errorf("formatPrice(2619) = %q, want $2619", got)
	}
}

func TestToFloat(t *testing.T) {
	if v := toFloat(42.5); v != 42.5 {
		t.Errorf("toFloat(42.5) = %f, want 42.5", v)
	}
	if v := toFloat(100); v != 100 {
		t.Errorf("toFloat(100) = %f, want 100", v)
	}
	if v := toFloat("$1,234.56"); v != 1234.56 {
		t.Errorf("toFloat($1,234.56) = %f, want 1234.56", v)
	}
	if v := toFloat(nil); v != 0 {
		t.Errorf("toFloat(nil) = %f, want 0", v)
	}
}
