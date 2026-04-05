package clients

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Gmail order search queries per source
var gmailOrderQueries = map[PurchaseSource]string{
	SourceAmazon:  `from:auto-confirm@amazon.com subject:"Your Amazon.com order"`,
	SourceNewegg:  `from:info@newegg.com subject:"Order Confirmation"`,
	SourceEbay:    `(from:ebay@ebay.com OR from:noreply@ebay.com) subject:("Order confirmed" OR "order confirmation" OR "You won")`,
	SourceBestBuy: `from:BestBuyInfo@emailinfo.bestbuy.com subject:"Order Confirmation"`,
}

// ImportGmailOrders searches Gmail for order confirmations and imports them as inventory items.
func (c *SheetsInventoryClient) ImportGmailOrders(ctx context.Context, gmailClient *GmailClient, config *GmailImportConfig) (*ImportResult, error) {
	if gmailClient == nil {
		return nil, fmt.Errorf("gmail client is required")
	}
	if config == nil {
		config = &GmailImportConfig{SinceDays: 365, MaxResults: 100}
	}
	if config.SinceDays <= 0 {
		config.SinceDays = 365
	}
	if config.MaxResults <= 0 {
		config.MaxResults = 100
	}

	// Determine which sources to scan
	sources := config.Sources
	if len(sources) == 0 {
		sources = []string{"amazon", "newegg", "ebay", "bestbuy"}
	}

	result := &ImportResult{
		Source: "gmail",
	}

	sinceDate := time.Now().AddDate(0, 0, -config.SinceDays).Format("2006/01/02")

	for _, sourceName := range sources {
		source := PurchaseSource(strings.ToLower(sourceName))
		query, ok := gmailOrderQueries[source]
		if !ok {
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("unknown source: %s", sourceName))
			result.Errors++
			continue
		}

		// Add date filter
		fullQuery := fmt.Sprintf("%s after:%s", query, sinceDate)

		emails, err := gmailClient.SearchMessages(ctx, fullQuery, int64(config.MaxResults))
		if err != nil {
			result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("%s search failed: %v", source, err))
			result.Errors++
			continue
		}

		for _, email := range emails {
			// Fetch full body
			fullEmail, err := gmailClient.GetMessage(ctx, email.ID, true)
			if err != nil {
				result.Errors++
				result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("failed to fetch email %s: %v", email.ID, err))
				continue
			}

			orders := parseOrderEmail(source, fullEmail)
			if len(orders) == 0 {
				result.Skipped++
				result.SkippedItems = append(result.SkippedItems, fmt.Sprintf("[%s] %s", source, truncate(fullEmail.Subject, 60)))
				continue
			}

			for _, order := range orders {
				result.TotalRows++

				// Skip non-electronics unless IncludeAll
				if !config.IncludeAll && !isLikelyElectronics(order.Name) {
					result.Skipped++
					result.SkippedItems = append(result.SkippedItems, fmt.Sprintf("[%s] %s ($%.2f) — not electronics", source, truncate(order.Name, 50), order.PurchasePrice))
					continue
				}

				if config.DryRun {
					result.Imported++
					result.NewItems = append(result.NewItems, fmt.Sprintf("[DRY RUN] [%s] %s — $%.2f — %s", source, truncate(order.Name, 50), order.PurchasePrice, order.OrderID))
					continue
				}

				item := &InventoryItem{
					Name:          order.Name,
					Category:      guessCategoryFromName(order.Name),
					PurchasePrice: order.PurchasePrice,
					PurchaseDate:  order.PurchaseDate,
					ASIN:          order.ASIN,
					Quantity:      max(order.Quantity, 1),
					Condition:     "new",
					ListingStatus: "not_listed",
					Notes:         fmt.Sprintf("Imported from %s email. Order: %s", source, order.OrderID),
				}

				added, err := c.AddItem(ctx, item)
				if err != nil {
					result.Errors++
					result.ErrorMessages = append(result.ErrorMessages, fmt.Sprintf("failed to add '%s': %v", truncate(order.Name, 40), err))
					continue
				}
				result.Imported++
				result.NewItems = append(result.NewItems, fmt.Sprintf("%s — %s ($%.2f)", added.SKU, truncate(order.Name, 40), order.PurchasePrice))
			}
		}
	}

	return result, nil
}

// parsedOrder holds extracted order data from an email
type parsedOrder struct {
	Name          string
	PurchasePrice float64
	PurchaseDate  string
	OrderID       string
	Quantity      int
	ASIN          string
}

// parseOrderEmail dispatches to the right parser based on source
func parseOrderEmail(source PurchaseSource, email *Email) []parsedOrder {
	switch source {
	case SourceAmazon:
		return parseAmazonEmail(email)
	case SourceNewegg:
		return parseNeweggEmail(email)
	case SourceEbay:
		return parseEbayEmail(email)
	case SourceBestBuy:
		return parseBestBuyEmail(email)
	default:
		return nil
	}
}

// --- Amazon parser ---

var (
	amazonOrderIDRegex = regexp.MustCompile(`(?i)order\s*#?\s*([\d-]{10,})`)
	amazonItemRegex    = regexp.MustCompile(`(?i)(\d+)\s+of:\s+(.+?)(?:\n|$)`)
	amazonPriceRegex   = regexp.MustCompile(`\$\s*([\d,]+\.?\d*)`)
	amazonASINRegex    = regexp.MustCompile(`/(?:dp|gp/product)/([A-Z0-9]{10})`)
)

func parseAmazonEmail(email *Email) []parsedOrder {
	body := email.Body
	subject := email.Subject

	orderID := ""
	if m := amazonOrderIDRegex.FindStringSubmatch(subject + " " + body); len(m) > 1 {
		orderID = m[1]
	}

	purchaseDate := formatEmailDate(email.Date)

	// Try structured "X of: Product Name" pattern
	items := amazonItemRegex.FindAllStringSubmatch(body, -1)
	if len(items) > 0 {
		var orders []parsedOrder
		prices := amazonPriceRegex.FindAllStringSubmatch(body, -1)
		asins := amazonASINRegex.FindAllStringSubmatch(body, -1)

		for i, item := range items {
			qty := 1
			if q := parseInt(item[1]); q > 0 {
				qty = q
			}
			name := strings.TrimSpace(item[2])
			name = cleanHTMLText(name)

			var price float64
			if i < len(prices) {
				price = parsePriceStr(prices[i][1])
			}

			asin := ""
			if i < len(asins) {
				asin = asins[i][1]
			}

			if name != "" && price > 0 {
				orders = append(orders, parsedOrder{
					Name:          name,
					PurchasePrice: price,
					PurchaseDate:  purchaseDate,
					OrderID:       orderID,
					Quantity:      qty,
					ASIN:          asin,
				})
			}
		}
		if len(orders) > 0 {
			return orders
		}
	}

	// Fallback: try to extract from subject + total price
	name := extractAmazonItemFromSubject(subject)
	if name != "" {
		prices := amazonPriceRegex.FindAllStringSubmatch(body, -1)
		if len(prices) > 0 {
			price := parsePriceStr(prices[0][1])
			if price > 0 {
				return []parsedOrder{{
					Name:          name,
					PurchasePrice: price,
					PurchaseDate:  purchaseDate,
					OrderID:       orderID,
					Quantity:      1,
				}}
			}
		}
	}

	return nil
}

func extractAmazonItemFromSubject(subject string) string {
	// Amazon subjects: "Your Amazon.com order of PRODUCT NAME has shipped!"
	// or "Your Amazon.com order #XXX-YYY-ZZZ"
	ofIdx := strings.Index(strings.ToLower(subject), " of ")
	if ofIdx > 0 {
		name := subject[ofIdx+4:]
		name = strings.TrimSuffix(name, " has shipped!")
		name = strings.TrimSuffix(name, " has shipped")
		name = strings.TrimSpace(name)
		if name != "" {
			return name
		}
	}
	return ""
}

// --- eBay parser ---

var (
	ebayOrderIDRegex = regexp.MustCompile(`(?i)order\s*(?:number|#|ID)[:\s]*([\d-]+)`)
	ebayItemRegex    = regexp.MustCompile(`(?i)(?:item|you\s+(?:won|bought|purchased))[:\s]+(.+?)(?:\n|<|$)`)
	ebayPriceRegex   = regexp.MustCompile(`(?i)(?:total|price|amount|won for|purchase price)[:\s]*\$\s*([\d,]+\.?\d*)`)
	ebayTDRegex      = regexp.MustCompile(`(?i)<td[^>]*>\s*(.*?)\s*</td>`)
	ebayTRRegex      = regexp.MustCompile(`(?is)<tr[^>]*>(.*?)</tr>`)
)

func parseEbayEmail(email *Email) []parsedOrder {
	body := email.Body
	purchaseDate := formatEmailDate(email.Date)

	orderID := ""
	if m := ebayOrderIDRegex.FindStringSubmatch(body); len(m) > 1 {
		orderID = m[1]
	}

	// Strategy 1: Parse HTML tables (common in eBay confirmation emails)
	orders := parseEbayHTMLTable(body, orderID, purchaseDate)
	if len(orders) > 0 {
		return orders
	}

	// Strategy 2: Regex-based extraction from text/HTML body
	var name string
	if m := ebayItemRegex.FindStringSubmatch(body); len(m) > 1 {
		name = cleanHTMLText(strings.TrimSpace(m[1]))
	}

	// Try subject line: "Order confirmed: ITEM NAME"
	if name == "" {
		subject := email.Subject
		for _, prefix := range []string{"Order confirmed: ", "order confirmation: ", "You won! "} {
			if idx := strings.Index(strings.ToLower(subject), strings.ToLower(prefix)); idx >= 0 {
				name = strings.TrimSpace(subject[idx+len(prefix):])
				break
			}
		}
	}

	var price float64
	if m := ebayPriceRegex.FindStringSubmatch(body); len(m) > 1 {
		price = parsePriceStr(m[1])
	}

	if name != "" && price >= 0.01 && price <= 50000 {
		return []parsedOrder{{
			Name:          name,
			PurchasePrice: price,
			PurchaseDate:  purchaseDate,
			OrderID:       orderID,
			Quantity:      1,
		}}
	}

	return nil
}

// parseEbayHTMLTable extracts items from eBay's HTML table format using a proper
// HTML tokenizer. eBay order emails typically have a table with columns: Item, Qty, Price.
// BUG-001: replaced regex-based HTML parsing with golang.org/x/net/html tokenizer
// to correctly handle nested tags, attributes, and malformed markup.
func parseEbayHTMLTable(body, orderID, purchaseDate string) []parsedOrder {
	tokenizer := html.NewTokenizer(strings.NewReader(body))

	var orders []parsedOrder
	priceRe := regexp.MustCompile(`\$\s*([\d,]+\.?\d*)`)
	qtyRe := regexp.MustCompile(`^\d+$`)

	headerWords := map[string]bool{
		"item": true, "product": true, "description": true,
		"qty": true, "quantity": true,
	}

	// State tracking
	var tableDepth int
	var inRow, inCell bool
	var currentRow []string
	var cellText strings.Builder

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			goto done
		case html.StartTagToken:
			tn, _ := tokenizer.TagName()
			tag := string(tn)
			switch tag {
			case "table":
				tableDepth++
			case "tr":
				if tableDepth > 0 {
					inRow = true
					currentRow = nil
				}
			case "td", "th":
				if inRow {
					inCell = true
					cellText.Reset()
				}
			}
		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			tag := string(tn)
			switch tag {
			case "table":
				if tableDepth > 0 {
					tableDepth--
				}
			case "tr":
				if inRow && len(currentRow) >= 2 {
					// Check if this is a header row
					isHeader := false
					for _, ct := range currentRow {
						if headerWords[strings.ToLower(strings.TrimSpace(ct))] {
							isHeader = true
							break
						}
					}

					if !isHeader {
						// Find item name: first non-empty cell
						var name string
						var nameIdx int
						for i, ct := range currentRow {
							t := strings.TrimSpace(ct)
							if t != "" {
								name = t
								nameIdx = i
								break
							}
						}

						// Search remaining cells for price and quantity
						var price float64
						qty := 1
						for i, ct := range currentRow {
							if i == nameIdx {
								continue
							}
							t := strings.TrimSpace(ct)
							if t == "" {
								continue
							}
							if m := priceRe.FindStringSubmatch(t); len(m) > 1 {
								if p := parsePriceStr(m[1]); p > 0 {
									price = p
								}
							} else if qtyRe.MatchString(t) {
								if q, err := strconv.Atoi(t); err == nil && q >= 1 && q <= 99 {
									qty = q
								}
							}
						}

						// BUG-001: price sanity check
						if name != "" && price >= 0.01 && price <= 50000 {
							orders = append(orders, parsedOrder{
								Name:          name,
								PurchasePrice: price,
								PurchaseDate:  purchaseDate,
								OrderID:       orderID,
								Quantity:      qty,
							})
						}
					}
				}
				inRow = false
			case "td", "th":
				if inCell {
					text := strings.TrimSpace(cellText.String())
					currentRow = append(currentRow, text)
					inCell = false
				}
			}
		case html.TextToken:
			if inCell {
				cellText.WriteString(tokenizer.Token().Data)
			}
		}
	}
done:
	return orders
}

// --- Newegg parser ---

var (
	neweggOrderIDRegex = regexp.MustCompile(`(?i)order\s*(?:number|#)[:\s]*([\d]+)`)
	neweggItemRegex    = regexp.MustCompile(`(?i)(?:item|product)\s*(?:description|name)?[:\s]+(.+?)(?:\n|<br|$)`)
	neweggPriceRegex   = regexp.MustCompile(`\$\s*([\d,]+\.?\d*)`)
)

func parseNeweggEmail(email *Email) []parsedOrder {
	body := email.Body
	purchaseDate := formatEmailDate(email.Date)

	orderID := ""
	if m := neweggOrderIDRegex.FindStringSubmatch(body); len(m) > 1 {
		orderID = m[1]
	}

	// Try HTML table parsing first (Newegg uses tables too)
	orders := parseHTMLTableGeneric(body, orderID, purchaseDate)
	if len(orders) > 0 {
		return orders
	}

	// Regex fallback
	if m := neweggItemRegex.FindStringSubmatch(body); len(m) > 1 {
		name := cleanHTMLText(strings.TrimSpace(m[1]))
		prices := neweggPriceRegex.FindAllStringSubmatch(body, -1)
		if name != "" && len(prices) > 0 {
			price := parsePriceStr(prices[0][1])
			if price > 0 {
				return []parsedOrder{{
					Name:          name,
					PurchasePrice: price,
					PurchaseDate:  purchaseDate,
					OrderID:       orderID,
					Quantity:      1,
				}}
			}
		}
	}

	return nil
}

// --- Best Buy parser ---

var (
	bestBuyOrderIDRegex = regexp.MustCompile(`(?i)order\s*(?:number|#)[:\s]*(BBY[\d-]+|[\d]{10,})`)
	bestBuyItemRegex    = regexp.MustCompile(`(?i)(?:item|product)[:\s]+(.+?)(?:\n|<|$)`)
	bestBuyPriceRegex   = regexp.MustCompile(`\$\s*([\d,]+\.?\d*)`)
)

func parseBestBuyEmail(email *Email) []parsedOrder {
	body := email.Body
	purchaseDate := formatEmailDate(email.Date)

	orderID := ""
	if m := bestBuyOrderIDRegex.FindStringSubmatch(body); len(m) > 1 {
		orderID = m[1]
	}

	// Try HTML table parsing
	orders := parseHTMLTableGeneric(body, orderID, purchaseDate)
	if len(orders) > 0 {
		return orders
	}

	// Regex fallback
	if m := bestBuyItemRegex.FindStringSubmatch(body); len(m) > 1 {
		name := cleanHTMLText(strings.TrimSpace(m[1]))
		prices := bestBuyPriceRegex.FindAllStringSubmatch(body, -1)
		if name != "" && len(prices) > 0 {
			price := parsePriceStr(prices[0][1])
			if price > 0 {
				return []parsedOrder{{
					Name:          name,
					PurchasePrice: price,
					PurchaseDate:  purchaseDate,
					OrderID:       orderID,
					Quantity:      1,
				}}
			}
		}
	}

	return nil
}

// --- Generic HTML table parser ---

func parseHTMLTableGeneric(body, orderID, purchaseDate string) []parsedOrder {
	rows := ebayTRRegex.FindAllStringSubmatch(body, -1)
	if len(rows) < 2 {
		return nil
	}

	var orders []parsedOrder
	priceRe := regexp.MustCompile(`\$\s*([\d,]+\.?\d*)`)

	for _, row := range rows {
		cells := ebayTDRegex.FindAllStringSubmatch(row[1], -1)
		if len(cells) < 2 {
			continue
		}

		var cellTexts []string
		for _, cell := range cells {
			cellTexts = append(cellTexts, cleanHTMLText(cell[1]))
		}

		firstCell := strings.ToLower(cellTexts[0])
		if firstCell == "" || firstCell == "item" || firstCell == "product" || firstCell == "description" || firstCell == "qty" {
			continue
		}

		name := cellTexts[0]
		var price float64
		qty := 1

		for _, cellText := range cellTexts[1:] {
			if m := priceRe.FindStringSubmatch(cellText); len(m) > 1 {
				if p := parsePriceStr(m[1]); p > 0 {
					price = p
				}
			} else if q := parseInt(cellText); q > 0 && q < 100 {
				qty = q
			}
		}

		if name != "" && len(name) > 3 && price > 0 {
			orders = append(orders, parsedOrder{
				Name:          name,
				PurchasePrice: price,
				PurchaseDate:  purchaseDate,
				OrderID:       orderID,
				Quantity:      qty,
			})
		}
	}

	return orders
}

// --- Utility functions ---

// cleanHTMLText strips HTML tags and decodes common entities
func cleanHTMLText(s string) string {
	// Strip HTML tags
	tagRe := regexp.MustCompile(`<[^>]*>`)
	s = tagRe.ReplaceAllString(s, "")

	// Decode common HTML entities
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&nbsp;", " ",
		"&#x27;", "'",
		"&#x2F;", "/",
	)
	s = replacer.Replace(s)

	// Collapse whitespace
	spaceRe := regexp.MustCompile(`\s+`)
	s = spaceRe.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}

func parsePriceStr(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	v, err := parsePrice(s)
	if err != nil {
		return 0
	}
	return v
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	var v int
	fmt.Sscanf(s, "%d", &v)
	return v
}

func formatEmailDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 2006")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// isLikelyElectronics checks if an item name suggests it's an electronics product
func isLikelyElectronics(name string) bool {
	lower := strings.ToLower(name)
	keywords := []string{
		"gpu", "cpu", "ram", "ssd", "hdd", "nvme", "m.2", "pcie",
		"motherboard", "graphics", "processor", "memory", "ddr",
		"power supply", "psu", "case", "cooler", "fan",
		"monitor", "display", "keyboard", "mouse",
		"usb", "hdmi", "thunderbolt", "ethernet", "wifi",
		"router", "switch", "hub", "adapter", "cable",
		"nvidia", "amd", "intel", "samsung", "western digital", "seagate",
		"corsair", "crucial", "g.skill", "evga", "asus", "msi", "gigabyte",
		"storage", "drive", "card", "controller", "dock",
		"raspberry pi", "arduino", "microcontroller",
		"camera", "capture", "audio", "interface", "headphone",
		"speaker", "microphone", "preamp", "dac", "amp",
		"rtx", "gtx", "radeon", "geforce", "ryzen", "core i",
		"tb4", "usb-c", "type-c", "gen4", "gen5",
	}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
