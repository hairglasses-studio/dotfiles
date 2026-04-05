package inventory

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

// preprocessScript is an embedded Python script for image preprocessing.
// Handles: EXIF auto-rotation, label auto-crop, optional rotation,
// grayscale, contrast, sharpen, upscale.
const preprocessScript = `
import sys
from PIL import Image, ImageOps, ImageEnhance, ImageFilter, ImageStat

input_path = sys.argv[1]
output_path = sys.argv[2]
rotation = int(sys.argv[3]) if len(sys.argv) > 3 else 0

img = Image.open(input_path)

# Auto-rotate based on EXIF orientation tag (handles phone camera rotation)
img = ImageOps.exif_transpose(img)

# Auto-crop to brightest region (product label stickers are white/light)
# Divide image into a grid and find the brightest cluster
gray_detect = img.convert('L')
w, h = gray_detect.size
grid = 6  # 6x6 grid
cell_w, cell_h = w // grid, h // grid
best_brightness = 0
best_cells = []

for gy in range(grid):
    for gx in range(grid):
        box = (gx * cell_w, gy * cell_h, (gx + 1) * cell_w, (gy + 1) * cell_h)
        cell = gray_detect.crop(box)
        brightness = ImageStat.Stat(cell).mean[0]
        best_cells.append((brightness, gx, gy))

# Find cells above 70th percentile brightness
best_cells.sort(reverse=True)
threshold = best_cells[len(best_cells) // 3][0]  # top third
bright = [(gx, gy) for br, gx, gy in best_cells if br >= threshold]

if bright:
    min_gx = max(0, min(gx for gx, gy in bright) - 1)
    max_gx = min(grid - 1, max(gx for gx, gy in bright) + 1)
    min_gy = max(0, min(gy for gx, gy in bright) - 1)
    max_gy = min(grid - 1, max(gy for gx, gy in bright) + 1)
    crop_box = (min_gx * cell_w, min_gy * cell_h, (max_gx + 1) * cell_w, (max_gy + 1) * cell_h)
    cropped = img.crop(crop_box)
    # Only use crop if it's meaningfully smaller (at least 20% reduction)
    if cropped.size[0] * cropped.size[1] < w * h * 0.8:
        img = cropped

# Apply rotation for multi-angle OCR attempts
if rotation:
    img = img.rotate(rotation, expand=True)

# Convert to grayscale
gray = img.convert('L')

# Auto-contrast (stretch histogram to full range)
gray = ImageOps.autocontrast(gray)

# Boost contrast for label text
enhancer = ImageEnhance.Contrast(gray)
gray = enhancer.enhance(2.0)

# Sharpen
gray = gray.filter(ImageFilter.SHARPEN)

# Upscale for better OCR on small text
w, h = gray.size
if max(w, h) < 4000:
    scale = 4000.0 / max(w, h)
    gray = gray.resize((int(w * scale), int(h * scale)), Image.LANCZOS)

gray.save(output_path)
`

// preprocessImage applies PIL-based image preprocessing for OCR.
// Returns the path to the preprocessed image (caller must clean up).
// Falls back to the original path if Python/PIL is unavailable.
func preprocessImage(ctx context.Context, inputPath string, rotation int) (string, bool, error) {
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("ocr_%s_r%d.png", base, rotation))

	cmd := exec.CommandContext(ctx, "python3", "-c", preprocessScript,
		inputPath, outputPath, fmt.Sprintf("%d", rotation))
	if _, err := cmd.CombinedOutput(); err != nil {
		// Fallback: return original path
		return inputPath, false, nil
	}

	return outputPath, true, nil
}

// ocrImage runs OCR on an image, trying 4 rotations with preprocessing.
// Returns the best detected text and a confidence score.
func ocrImage(ctx context.Context, imagePath string) (string, float64, error) {
	rotations := []int{0, 90, 180, 270}

	bestText := ""
	bestScore := 0
	bestConf := 0.0

	for _, rot := range rotations {
		processed, created, err := preprocessImage(ctx, imagePath, rot)
		if err != nil {
			continue
		}

		text, err := runTesseract(ctx, processed)
		if created {
			os.Remove(processed)
		}
		if err != nil || text == "" {
			continue
		}

		score := scoreOCRText(text)
		if score > bestScore {
			bestText = text
			bestScore = score
			bestConf = 0.8
		}

		// Early exit if we found model number patterns
		if score >= 20 {
			break
		}
	}

	// Fallback: try raw image (no preprocessing)
	if bestScore == 0 {
		text, err := runTesseract(ctx, imagePath)
		if err == nil && len(strings.TrimSpace(text)) > 0 {
			bestText = text
			bestConf = 0.3
		}
	}

	return bestText, bestConf, nil
}

// runTesseract calls tesseract directly with optimal settings for product labels.
func runTesseract(ctx context.Context, imagePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "tesseract", imagePath, "stdout",
		"-l", "eng",
		"--psm", "6", // Assume uniform block of text
	)
	// Capture stdout only; tesseract writes warnings to stderr (e.g. "Estimating
	// resolution") which cause cmd.Output() to report failure even on success.
	var stdout strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = nil // discard stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("tesseract failed: %w", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// scoreOCRText scores OCR output by how likely it contains useful product info.
// Higher score = more likely to be a valid product label reading.
func scoreOCRText(text string) int {
	if text == "" {
		return 0
	}

	score := 0

	// Count alphanumeric characters vs total (ratio indicates real text vs garbage)
	var alphaNum, total int
	for _, r := range text {
		total++
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			alphaNum++
		}
	}
	if total > 0 {
		ratio := float64(alphaNum) / float64(total)
		if ratio > 0.4 {
			score += alphaNum / 10
		}
	}

	// Bonus for model number regex matches
	for _, p := range modelPatterns {
		if p.Pattern.MatchString(text) {
			score += 20
		}
	}

	// Bonus for common product label keywords
	upper := strings.ToUpper(text)
	keywords := []string{
		"MODEL", "SERIAL", "P/N", "S/N", "REV", "FW",
		"SAMSUNG", "SABRENT", "SEAGATE", "CRUCIAL", "CORSAIR",
		"KINGSTON", "WD", "INTEL", "HYNIX", "MICRON", "TOSHIBA",
		"PRODUCT OF", "MADE IN", "DC", "SSD", "NVMe", "DDR",
	}
	for _, kw := range keywords {
		if strings.Contains(upper, kw) {
			score += 5
		}
	}

	return score
}
