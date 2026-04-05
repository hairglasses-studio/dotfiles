package inventory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Model-number regex patterns for OCR text matching.
var modelPatterns = []struct {
	Name    string
	Pattern *regexp.Regexp
}{
	{"Samsung SSD", regexp.MustCompile(`MZ-[A-Z0-9]+(?:/[A-Z0-9]+)?`)},
	{"Sabrent", regexp.MustCompile(`SB-RKT[A-Z0-9]+-[0-9]+[A-Z]*`)},
	{"Seagate", regexp.MustCompile(`ST[0-9]+[A-Z]+[0-9]*`)},
	{"G.Skill", regexp.MustCompile(`F[0-9]+-[0-9]+[A-Z0-9]+`)},
	{"WD", regexp.MustCompile(`WDS[0-9]+[A-Z0-9]+`)},
	{"Crucial", regexp.MustCompile(`CT[0-9]+[A-Z0-9]+`)},
	{"Intel", regexp.MustCompile(`SSDPE[A-Z0-9]+`)},
	{"Corsair", regexp.MustCompile(`CM[A-Z0-9]+-[0-9]+[A-Z0-9]+`)},
	{"Generic model", regexp.MustCompile(`[A-Z]{2,}[-][A-Z0-9]{3,}[-][A-Z0-9]+`)},
}

// detectedModel represents a model number found via OCR.
type detectedModel struct {
	PatternName string `json:"pattern_name"`
	ModelNumber string `json:"model_number"`
}

// photoMatch holds matches against existing inventory/specs catalog.
type photoMatch struct {
	Source      string            `json:"source"`       // "inventory" or "specs_catalog"
	SKU        string            `json:"sku,omitempty"` // inventory SKU if matched
	Name       string            `json:"name"`
	Model      string            `json:"model,omitempty"`
	Specs      map[string]string `json:"specs,omitempty"`
	Confidence string            `json:"confidence"` // "exact", "partial"
}

// getPhotosFolderID returns the configured Drive folder ID for inventory photos.
func getPhotosFolderID(override string) string {
	if override != "" {
		return override
	}
	return config.GetEnv("INVENTORY_PHOTOS_DRIVE_FOLDER_ID", "")
}

// extractModels runs regex patterns against OCR text and returns detected models.
func extractModels(text string) []detectedModel {
	var results []detectedModel
	seen := map[string]bool{}
	for _, p := range modelPatterns {
		matches := p.Pattern.FindAllString(text, -1)
		for _, m := range matches {
			if !seen[m] {
				seen[m] = true
				results = append(results, detectedModel{
					PatternName: p.Name,
					ModelNumber: m,
				})
			}
		}
	}
	return results
}

// matchModelsToInventory checks detected models against existing inventory items and specs catalog.
func matchModelsToInventory(ctx context.Context, models []detectedModel) ([]photoMatch, error) {
	var matches []photoMatch

	// Get inventory client for item matching
	invClient, err := clients.GetInventoryClient()
	if err == nil {
		items, listErr := invClient.ListItems(ctx, &clients.InventoryFilter{Limit: 500})
		if listErr == nil {
			for _, m := range models {
				upper := strings.ToUpper(m.ModelNumber)
				for _, item := range items {
					if strings.ToUpper(item.Model) == upper ||
						strings.Contains(strings.ToUpper(item.Name), upper) {
						matches = append(matches, photoMatch{
							Source:     "inventory",
							SKU:       item.SKU,
							Name:      item.Name,
							Model:     item.Model,
							Confidence: "exact",
						})
					}
				}
			}
		}
	}

	// Check against specs catalog
	for _, m := range models {
		for productLine, specs := range specsCatalog {
			if strings.Contains(strings.ToUpper(productLine), strings.ToUpper(m.ModelNumber)) ||
				strings.Contains(strings.ToUpper(m.ModelNumber), strings.ToUpper(strings.ReplaceAll(productLine, " ", ""))) {
				matches = append(matches, photoMatch{
					Source:     "specs_catalog",
					Name:       productLine,
					Model:      m.ModelNumber,
					Specs:      specs,
					Confidence: "partial",
				})
			}
		}
	}

	return matches, nil
}

// ocrImage is defined in photo_preprocess.go — preprocesses + tries 4 rotations + tesseract.

// ── Handlers ──

// handlePhotoSync downloads photos from a Drive folder into local images/{sku}/ dirs.
// Filename convention: HW-001-front.jpg → images/HW-001/front.jpg
func handlePhotoSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	folderID := getPhotosFolderID(tools.GetStringParam(req, "folder_id"))
	if folderID == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("folder_id is required (or set INVENTORY_PHOTOS_DRIVE_FOLDER_ID)")), nil
	}

	driveClient, err := clients.GetGDriveClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get Drive client: %w", err)), nil
	}

	invClient, err := clients.GetInventoryClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get inventory client: %w", err)), nil
	}

	// List image files in the Drive folder
	files, err := driveClient.SearchFiles(ctx, "", "image", folderID, 100)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list photos: %w", err)), nil
	}

	if len(files) == 0 {
		return tools.JSONResult(map[string]interface{}{
			"synced": 0,
			"message": "no image files found in folder",
		}), nil
	}

	type syncResult struct {
		FileName string `json:"file_name"`
		SKU      string `json:"sku"`
		Label    string `json:"label"`
		Status   string `json:"status"`
	}

	var results []syncResult
	synced := 0
	skipped := 0

	for _, f := range files {
		// Parse filename: expected format "HW-001-front.jpg" or "HW-001_sticker.png"
		name := strings.TrimSuffix(f.Name, filepath.Ext(f.Name))
		parts := strings.SplitN(name, "-", 3)

		var sku, label string
		if len(parts) >= 3 && strings.EqualFold(parts[0], "HW") {
			sku = strings.ToUpper(parts[0] + "-" + parts[1])
			label = parts[2]
		} else {
			// Try underscore separator: "HW-001_sticker.jpg"
			underParts := strings.SplitN(name, "_", 2)
			if len(underParts) == 2 && strings.HasPrefix(strings.ToUpper(underParts[0]), "HW-") {
				sku = strings.ToUpper(underParts[0])
				label = underParts[1]
			} else {
				results = append(results, syncResult{
					FileName: f.Name, Status: "skipped: filename does not match HW-NNN-label pattern",
				})
				skipped++
				continue
			}
		}

		destFilename := label + filepath.Ext(f.Name)

		// Check if image already exists locally
		existing, _ := invClient.ListImages(ctx, sku)
		alreadyExists := false
		for _, img := range existing {
			if filepath.Base(img) == destFilename {
				alreadyExists = true
				break
			}
		}
		if alreadyExists {
			results = append(results, syncResult{
				FileName: f.Name, SKU: sku, Label: label, Status: "skipped: already exists",
			})
			skipped++
			continue
		}

		// Download to temp then upload via inventory client
		tmpDir := os.TempDir()
		tmpPath := filepath.Join(tmpDir, f.Name)
		_, err := driveClient.DownloadFile(ctx, f.ID, tmpPath)
		if err != nil {
			results = append(results, syncResult{
				FileName: f.Name, SKU: sku, Label: label, Status: fmt.Sprintf("error: %v", err),
			})
			continue
		}

		imgData, err := os.ReadFile(tmpPath)
		os.Remove(tmpPath)
		if err != nil {
			results = append(results, syncResult{
				FileName: f.Name, SKU: sku, Label: label, Status: fmt.Sprintf("error reading: %v", err),
			})
			continue
		}

		_, err = invClient.UploadImage(ctx, sku, imgData, destFilename)
		if err != nil {
			results = append(results, syncResult{
				FileName: f.Name, SKU: sku, Label: label, Status: fmt.Sprintf("error uploading: %v", err),
			})
			continue
		}

		results = append(results, syncResult{
			FileName: f.Name, SKU: sku, Label: label, Status: "synced",
		})
		synced++
	}

	RecordItemOperation("photo_sync", "", "success")

	return tools.JSONResult(map[string]interface{}{
		"total":   len(files),
		"synced":  synced,
		"skipped": skipped,
		"results": results,
	}), nil
}

// handlePhotoIngest downloads photos from Drive → runs OCR → extracts model numbers → matches inventory.
func handlePhotoIngest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	folderID := getPhotosFolderID(tools.GetStringParam(req, "folder_id"))
	if folderID == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("folder_id is required (or set INVENTORY_PHOTOS_DRIVE_FOLDER_ID)")), nil
	}

	limit := tools.GetIntParam(req, "limit", 10)

	driveClient, err := clients.GetGDriveClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get Drive client: %w", err)), nil
	}

	// List image files in the Drive folder
	files, err := driveClient.SearchFiles(ctx, "", "image", folderID, int64(limit))
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list photos: %w", err)), nil
	}

	if len(files) == 0 {
		return tools.JSONResult(map[string]interface{}{
			"processed": 0,
			"message":   "no image files found in folder",
		}), nil
	}

	type ingestResult struct {
		FileName       string         `json:"file_name"`
		DetectedText   string         `json:"detected_text,omitempty"`
		DetectedModels []detectedModel `json:"detected_models,omitempty"`
		Matches        []photoMatch   `json:"matches,omitempty"`
		Error          string         `json:"error,omitempty"`
	}

	var results []ingestResult
	tmpDir := os.TempDir()

	for _, f := range files {
		// Only process images
		if !strings.HasPrefix(f.MimeType, "image/") {
			continue
		}

		r := ingestResult{FileName: f.Name}

		// Download to temp file
		tmpPath := filepath.Join(tmpDir, "photo_ingest_"+f.Name)
		_, err := driveClient.DownloadFile(ctx, f.ID, tmpPath)
		if err != nil {
			r.Error = fmt.Sprintf("download failed: %v", err)
			results = append(results, r)
			continue
		}

		// Run OCR with preprocessing + multi-rotation
		text, _, err := ocrImage(ctx, tmpPath)
		os.Remove(tmpPath)
		if err != nil {
			r.Error = fmt.Sprintf("OCR failed: %v", err)
			results = append(results, r)
			continue
		}

		r.DetectedText = text

		// Extract model numbers
		r.DetectedModels = extractModels(text)

		// Match against inventory and specs catalog
		if len(r.DetectedModels) > 0 {
			r.Matches, _ = matchModelsToInventory(ctx, r.DetectedModels)
		}

		results = append(results, r)
	}

	RecordItemOperation("photo_ingest", "", "success")

	return tools.JSONResult(map[string]interface{}{
		"processed": len(results),
		"results":   results,
	}), nil
}

// handlePhotoScan runs OCR on a single local image file and matches against inventory.
func handlePhotoScan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return tools.CodedErrorResult(tools.ErrNotFound, fmt.Errorf("file not found: %s", filePath)), nil
	}

	// Run OCR with preprocessing + multi-rotation
	text, confidence, err := ocrImage(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("OCR failed: %w", err)), nil
	}

	// Extract model numbers
	models := extractModels(text)

	// Match against inventory and specs catalog
	var matches []photoMatch
	if len(models) > 0 {
		matches, _ = matchModelsToInventory(ctx, models)
	}

	RecordItemOperation("photo_scan", "", "success")

	return tools.JSONResult(map[string]interface{}{
		"file":            filePath,
		"detected_text":   text,
		"confidence":      confidence,
		"detected_models": models,
		"matches":         matches,
	}), nil
}
