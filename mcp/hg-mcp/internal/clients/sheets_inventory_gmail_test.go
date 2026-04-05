package clients

import (
	"testing"
	"time"
)

func TestParseAmazonEmail(t *testing.T) {
	email := &Email{
		Subject: "Your Amazon.com order of Samsung 990 Pro 2TB NVMe SSD has shipped!",
		Body: `Your Amazon.com order #112-1234567-8901234

1 of: Samsung 990 Pro 2TB NVMe SSD

Order Total: $159.99

https://www.amazon.com/dp/B0BHJJ9Y77`,
		Date: time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC),
	}

	orders := parseAmazonEmail(email)
	if len(orders) == 0 {
		t.Fatal("expected at least 1 order, got 0")
	}

	order := orders[0]
	if order.Name != "Samsung 990 Pro 2TB NVMe SSD" {
		t.Errorf("name = %q, want %q", order.Name, "Samsung 990 Pro 2TB NVMe SSD")
	}
	if order.PurchasePrice != 159.99 {
		t.Errorf("price = %v, want 159.99", order.PurchasePrice)
	}
	if order.OrderID != "112-1234567-8901234" {
		t.Errorf("orderID = %q, want %q", order.OrderID, "112-1234567-8901234")
	}
	if order.ASIN != "B0BHJJ9Y77" {
		t.Errorf("ASIN = %q, want %q", order.ASIN, "B0BHJJ9Y77")
	}
	if order.PurchaseDate != "Jun 2024" {
		t.Errorf("date = %q, want %q", order.PurchaseDate, "Jun 2024")
	}
}

func TestParseAmazonEmailFallbackSubject(t *testing.T) {
	email := &Email{
		Subject: "Your Amazon.com order of Corsair RM850x PSU has shipped!",
		Body:    `Thank you for your order. Total: $119.99`,
		Date:    time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	orders := parseAmazonEmail(email)
	if len(orders) == 0 {
		t.Fatal("expected at least 1 order from subject fallback")
	}
	if orders[0].Name != "Corsair RM850x PSU" {
		t.Errorf("name = %q, want %q", orders[0].Name, "Corsair RM850x PSU")
	}
	if orders[0].PurchasePrice != 119.99 {
		t.Errorf("price = %v, want 119.99", orders[0].PurchasePrice)
	}
}

func TestParseEbayEmailHTMLTable(t *testing.T) {
	email := &Email{
		Subject: "Order confirmed: EVGA RTX 3080 FTW3 Ultra",
		Body: `<html>
<body>
<p>Order number: 12-34567-89012</p>
<table>
<tr><td>Item</td><td>Qty</td><td>Price</td></tr>
<tr><td>EVGA RTX 3080 FTW3 Ultra Gaming 10GB</td><td>1</td><td>$449.99</td></tr>
</table>
</body>
</html>`,
		Date: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC),
	}

	orders := parseEbayEmail(email)
	if len(orders) == 0 {
		t.Fatal("expected at least 1 order from eBay HTML table")
	}

	order := orders[0]
	if order.Name != "EVGA RTX 3080 FTW3 Ultra Gaming 10GB" {
		t.Errorf("name = %q, want %q", order.Name, "EVGA RTX 3080 FTW3 Ultra Gaming 10GB")
	}
	if order.PurchasePrice != 449.99 {
		t.Errorf("price = %v, want 449.99", order.PurchasePrice)
	}
	if order.OrderID != "12-34567-89012" {
		t.Errorf("orderID = %q, want %q", order.OrderID, "12-34567-89012")
	}
}

func TestParseEbayEmailRegexFallback(t *testing.T) {
	email := &Email{
		Subject: "Order confirmed: G.Skill Trident Z5 RGB 32GB DDR5",
		Body:    `Congratulations! You won: G.Skill Trident Z5 RGB 32GB DDR5-6000\nTotal price: $89.99\nOrder #: 98-76543-21098`,
		Date:    time.Date(2024, 5, 10, 0, 0, 0, 0, time.UTC),
	}

	orders := parseEbayEmail(email)
	if len(orders) == 0 {
		t.Fatal("expected at least 1 order from eBay regex fallback")
	}
	if orders[0].PurchasePrice != 89.99 {
		t.Errorf("price = %v, want 89.99", orders[0].PurchasePrice)
	}
}

func TestParseNeweggEmail(t *testing.T) {
	email := &Email{
		Subject: "Order Confirmation - Newegg.com",
		Body: `<html>
<body>
<p>Order Number: 8234567890</p>
<table>
<tr><td>Product</td><td>Qty</td><td>Price</td></tr>
<tr><td>Sabrent Rocket 4 Plus 2TB NVMe SSD</td><td>1</td><td>$179.99</td></tr>
</table>
</body>
</html>`,
		Date: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	orders := parseNeweggEmail(email)
	if len(orders) == 0 {
		t.Fatal("expected at least 1 order from Newegg")
	}
	if orders[0].Name != "Sabrent Rocket 4 Plus 2TB NVMe SSD" {
		t.Errorf("name = %q, want %q", orders[0].Name, "Sabrent Rocket 4 Plus 2TB NVMe SSD")
	}
	if orders[0].PurchasePrice != 179.99 {
		t.Errorf("price = %v, want 179.99", orders[0].PurchasePrice)
	}
}

func TestCleanHTMLText(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{`<b>Bold</b> text`, "Bold text"},
		{`Price &amp; Tax`, "Price & Tax"},
		{`  multiple   spaces  `, "multiple spaces"},
		{`<a href="x">Link</a>`, "Link"},
		{`plain text`, "plain text"},
	}

	for _, tt := range tests {
		got := cleanHTMLText(tt.input)
		if got != tt.want {
			t.Errorf("cleanHTMLText(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsLikelyElectronics(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"Samsung 990 Pro 2TB NVMe SSD", true},
		{"EVGA RTX 3080 FTW3 Ultra", true},
		{"Corsair Vengeance DDR5-6000 32GB", true},
		{"USB-C Hub Adapter", true},
		{"Organic Dog Food 30lb Bag", false},
		{"Kitchen Paper Towels 12-pack", false},
	}

	for _, tt := range tests {
		got := isLikelyElectronics(tt.name)
		if got != tt.want {
			t.Errorf("isLikelyElectronics(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParsePriceStr(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"159.99", 159.99},
		{"1,299.99", 1299.99},
		{"0.99", 0.99},
		{"abc", 0},
	}

	for _, tt := range tests {
		got := parsePriceStr(tt.input)
		if got != tt.want {
			t.Errorf("parsePriceStr(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFormatEmailDate(t *testing.T) {
	d := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	if got := formatEmailDate(d); got != "Jun 2024" {
		t.Errorf("formatEmailDate = %q, want %q", got, "Jun 2024")
	}
	if got := formatEmailDate(time.Time{}); got != "" {
		t.Errorf("formatEmailDate(zero) = %q, want empty", got)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 10); got != "short" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("this is a very long string", 10); got != "this is..." {
		t.Errorf("truncate long = %q, want %q", got, "this is...")
	}
}
