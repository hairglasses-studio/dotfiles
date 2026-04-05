package clients

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/sheets/v4"
)

// RecordSale records a completed sale to the Sales tab and updates inventory.
func (c *SheetsInventoryClient) RecordSale(ctx context.Context, sale *SaleRecord) (*SaleRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.testMode {
		// Ensure Sales tab exists
		if err := c.ensureSalesTab(ctx); err != nil {
			return nil, err
		}
	}

	// Auto-generate sale ID
	sale.SaleID = fmt.Sprintf("SALE-%03d", len(c.salesCache)+1)

	// Calculate financials
	sale.Revenue = sale.SoldPrice * float64(sale.QuantitySold)
	sale.Cost = sale.PurchasePrice * float64(sale.QuantitySold)
	sale.NetProfit = sale.Revenue - sale.Cost - sale.ShippingCost - sale.PlatformFees

	if sale.Date == "" {
		sale.Date = time.Now().Format(time.RFC3339)
	}
	if sale.QuantitySold == 0 {
		sale.QuantitySold = 1
	}

	if !c.testMode {
		// Append to Sales tab
		row := saleToRow(sale)
		vr := &sheets.ValueRange{Values: [][]interface{}{row}}
		_, err := c.service.Spreadsheets.Values.Append(c.spreadsheetID, salesTab, vr).
			ValueInputOption("USER_ENTERED").Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("failed to append sale: %w", err)
		}
	}

	c.salesCache = append(c.salesCache, *sale)
	c.salesCacheTime = time.Now()
	return sale, nil
}

// GetSalesLog returns all sales from the Sales tab.
func (c *SheetsInventoryClient) GetSalesLog(ctx context.Context) ([]SaleRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.testMode {
		return c.salesCache, nil
	}

	if !c.salesCacheTime.IsZero() && time.Since(c.salesCacheTime) < cacheTTL {
		return c.salesCache, nil
	}

	rows, err := c.readSalesRows(ctx)
	if err != nil {
		return nil, err
	}

	c.salesCache = rows
	c.salesCacheTime = time.Now()
	return rows, nil
}

// GetSalesSummary returns aggregated sales data.
func (c *SheetsInventoryClient) GetSalesSummary(ctx context.Context) (*SalesSummary, error) {
	sales, err := c.GetSalesLog(ctx)
	if err != nil {
		return nil, err
	}

	summary := &SalesSummary{TotalSales: len(sales)}
	for _, s := range sales {
		summary.TotalRevenue += s.Revenue
		summary.TotalCost += s.Cost
		summary.TotalFees += s.ShippingCost + s.PlatformFees
		summary.TotalProfit += s.NetProfit
		summary.UnitsSold += s.QuantitySold
	}
	return summary, nil
}

// ---------- internal helpers ----------

func (c *SheetsInventoryClient) ensureSalesTab(ctx context.Context) error {
	_, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, salesTab+"!A1").
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err == nil {
		return nil
	}

	// Create Sales tab
	addReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{Title: salesTab},
			}},
		},
	}
	if _, bErr := c.service.Spreadsheets.BatchUpdate(c.spreadsheetID, addReq).Context(ctx).Do(); bErr != nil {
		if !strings.Contains(bErr.Error(), "already exists") {
			return fmt.Errorf("failed to create Sales tab: %w", bErr)
		}
	}

	// Write header
	header := []interface{}{
		"SaleID", "ItemNum", "ItemName", "QuantitySold", "SoldPrice",
		"PurchasePrice", "Revenue", "Cost", "ShippingCost", "PlatformFees",
		"NetProfit", "Platform", "BuyerInfo", "Notes", "Date",
	}
	vr := &sheets.ValueRange{Values: [][]interface{}{header}}
	_, err = c.service.Spreadsheets.Values.Update(c.spreadsheetID, salesTab+"!A1:O1", vr).
		ValueInputOption("USER_ENTERED").Context(ctx).Do()
	return err
}

func (c *SheetsInventoryClient) readSalesRows(ctx context.Context) ([]SaleRecord, error) {
	if err := c.ensureSalesTab(ctx); err != nil {
		return nil, err
	}

	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, salesTab).
		ValueRenderOption("FORMATTED_VALUE").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read Sales tab: %w", err)
	}

	if len(resp.Values) < 2 {
		return nil, nil
	}

	var sales []SaleRecord
	for _, row := range resp.Values[1:] {
		s := rowToSale(row)
		if s.SaleID != "" {
			sales = append(sales, s)
		}
	}
	return sales, nil
}

func saleToRow(s *SaleRecord) []interface{} {
	return []interface{}{
		s.SaleID,
		s.ItemNum,
		s.ItemName,
		s.QuantitySold,
		fmt.Sprintf("$%.2f", s.SoldPrice),
		fmt.Sprintf("$%.2f", s.PurchasePrice),
		fmt.Sprintf("$%.2f", s.Revenue),
		fmt.Sprintf("$%.2f", s.Cost),
		fmt.Sprintf("$%.2f", s.ShippingCost),
		fmt.Sprintf("$%.2f", s.PlatformFees),
		fmt.Sprintf("$%.2f", s.NetProfit),
		s.Platform,
		s.BuyerInfo,
		s.Notes,
		s.Date,
	}
}

func rowToSale(row []interface{}) SaleRecord {
	get := func(idx int) string {
		if idx >= len(row) {
			return ""
		}
		return strings.TrimSpace(fmt.Sprintf("%v", row[idx]))
	}
	getFloat := func(idx int) float64 {
		s := get(idx)
		s = strings.ReplaceAll(s, "$", "")
		s = strings.ReplaceAll(s, ",", "")
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}
	getInt := func(idx int) int {
		v, _ := strconv.Atoi(get(idx))
		return v
	}

	return SaleRecord{
		SaleID:        get(0),
		ItemNum:       getInt(1),
		ItemName:      get(2),
		QuantitySold:  getInt(3),
		SoldPrice:     getFloat(4),
		PurchasePrice: getFloat(5),
		Revenue:       getFloat(6),
		Cost:          getFloat(7),
		ShippingCost:  getFloat(8),
		PlatformFees:  getFloat(9),
		NetProfit:     getFloat(10),
		Platform:      get(11),
		BuyerInfo:     get(12),
		Notes:         get(13),
		Date:          get(14),
	}
}
