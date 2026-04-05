package inventory

import (
	"strings"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

func TestRenderListing_FBMarketplace(t *testing.T) {
	item := &clients.InventoryItem{
		Name:          "ASUS ROG Strix RTX 4090",
		AskingPrice:   2200,
		Condition:     "Used",
		Model:         "ROG-STRIX-RTX4090-O24G",
		Category:      "GPU",
		CurrentRetail: 2755,
		Notes:         "Excellent condition, barely used.",
	}

	listing, err := RenderListing(item, "fb_marketplace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(listing, "ASUS ROG Strix RTX 4090") {
		t.Error("listing should contain item name")
	}
	if !strings.Contains(listing, "$2200") {
		t.Error("listing should contain asking price")
	}
	if !strings.Contains(listing, "Cash or Zelle") {
		t.Error("listing should contain payment info")
	}
}

func TestRenderListing_Ebay(t *testing.T) {
	item := &clients.InventoryItem{
		Name:        "Samsung 990 Pro 2TB",
		AskingPrice: 180,
		Condition:   "new",
		Model:       "MZ-V9P2T0B/AM",
		Category:    "Storage",
		ASIN:        "B0BHJJ9Y77",
		Notes:       "Factory sealed.",
	}

	listing, err := RenderListing(item, "ebay")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(listing, "Samsung 990 Pro 2TB") {
		t.Error("listing should contain item name")
	}
	if !strings.Contains(listing, "SHIPPING:") {
		t.Error("listing should contain shipping section")
	}
}

func TestRenderListing_Hardwareswap(t *testing.T) {
	item := &clients.InventoryItem{
		Name:        "Corsair DDR5 32GB Kit",
		AskingPrice: 120,
		Condition:   "like_new",
		Category:    "RAM",
		SmartData:   "Health: PASSED, Temp: 35C",
	}

	listing, err := RenderListing(item, "hardwareswap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(listing, "[H] Corsair DDR5 32GB Kit [W] PayPal") {
		t.Error("listing should contain hardwareswap format")
	}
	if !strings.Contains(listing, "SMART Data") {
		t.Error("listing should contain SMART data block")
	}
}

func TestRenderListing_InvalidPlatform(t *testing.T) {
	item := &clients.InventoryItem{Name: "Test"}
	_, err := RenderListing(item, "craigslist")
	if err == nil {
		t.Error("expected error for invalid platform")
	}
}

func TestFormatConditionDisplay(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"new", "New / Sealed"},
		{"like_new", "Like New"},
		{"Used", "Used — Excellent"},
		{"Renewed", "Renewed / Refurbished"},
		{"for_parts", "For Parts / Not Working"},
		{"Custom Condition", "Custom Condition"},
	}

	for _, tt := range tests {
		got := formatConditionDisplay(tt.input)
		if got != tt.expected {
			t.Errorf("formatConditionDisplay(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
