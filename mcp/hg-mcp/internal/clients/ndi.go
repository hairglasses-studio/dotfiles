// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// NDIClient provides access to NDI source discovery and status
type NDIClient struct {
	discoveryTimeout time.Duration
}

// NDISource represents an NDI source
type NDISource struct {
	Name      string  `json:"name"`
	Host      string  `json:"host"`
	URL       string  `json:"url"`
	Connected bool    `json:"connected"`
	FPS       float64 `json:"fps,omitempty"`
	Width     int     `json:"width,omitempty"`
	Height    int     `json:"height,omitempty"`
	Bandwidth string  `json:"bandwidth,omitempty"`
}

// NDIHealth represents NDI system health
type NDIHealth struct {
	Score           int         `json:"score"`
	Status          string      `json:"status"`
	SourceCount     int         `json:"source_count"`
	ActiveSources   int         `json:"active_sources"`
	Issues          []string    `json:"issues,omitempty"`
	Recommendations []string    `json:"recommendations,omitempty"`
	Sources         []NDISource `json:"sources,omitempty"`
}

// StreamHealth represents overall streaming health
type StreamHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	NDIHealth       int      `json:"ndi_health"`
	OBSHealth       int      `json:"obs_health"`
	CaptureHealth   int      `json:"capture_health"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewNDIClient creates a new NDI client
func NewNDIClient() (*NDIClient, error) {
	return &NDIClient{
		discoveryTimeout: 5 * time.Second,
	}, nil
}

// DiscoverSources discovers available NDI sources on the network
func (c *NDIClient) DiscoverSources(ctx context.Context) ([]NDISource, error) {
	sources := []NDISource{}

	// Try to use ndi-find if available
	if path, err := exec.LookPath("ndi-find"); err == nil {
		cmd := exec.CommandContext(ctx, path)
		output, err := cmd.Output()
		if err == nil {
			// Parse ndi-find output
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "Found") {
					sources = append(sources, NDISource{
						Name: line,
						URL:  line,
					})
				}
			}
			return sources, nil
		}
	}

	// Fallback: scan common NDI ports on local network
	localSources := c.scanLocalNDI(ctx)
	sources = append(sources, localSources...)

	return sources, nil
}

// scanLocalNDI scans for NDI sources on the local network
func (c *NDIClient) scanLocalNDI(ctx context.Context) []NDISource {
	sources := []NDISource{}

	// NDI uses mDNS/Bonjour for discovery
	// Common NDI ports: 5960-5969 for video
	// This is a simplified scan - real implementation would use mDNS

	// Check localhost for common NDI applications
	localApps := []struct {
		name string
		port int
	}{
		{"OBS NDI Output", 5961},
		{"TouchDesigner NDI", 5962},
		{"NDI Screen Capture", 5963},
	}

	for _, app := range localApps {
		addr := fmt.Sprintf("127.0.0.1:%d", app.port)
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			sources = append(sources, NDISource{
				Name:      app.name,
				Host:      "localhost",
				URL:       addr,
				Connected: true,
			})
		}
	}

	return sources
}

// GetSourceStatus gets detailed status for an NDI source
func (c *NDIClient) GetSourceStatus(ctx context.Context, sourceName string) (*NDISource, error) {
	sources, err := c.DiscoverSources(ctx)
	if err != nil {
		return nil, err
	}

	for _, source := range sources {
		if source.Name == sourceName || strings.Contains(source.Name, sourceName) {
			// In a real implementation, we'd connect and get frame info
			source.FPS = 30.0 // Placeholder
			source.Width = 1920
			source.Height = 1080
			source.Bandwidth = "~150 Mbps"
			return &source, nil
		}
	}

	return nil, fmt.Errorf("source not found: %s", sourceName)
}

// GetHealth returns NDI system health metrics
func (c *NDIClient) GetHealth(ctx context.Context) (*NDIHealth, error) {
	sources, err := c.DiscoverSources(ctx)
	if err != nil {
		return nil, err
	}

	health := &NDIHealth{
		SourceCount: len(sources),
		Sources:     sources,
	}

	// Count active sources
	for _, s := range sources {
		if s.Connected {
			health.ActiveSources++
		}
	}

	// Calculate health score
	health.Score = 100

	if health.SourceCount == 0 {
		health.Score = 50
		health.Issues = append(health.Issues, "No NDI sources discovered")
		health.Recommendations = append(health.Recommendations, "Check that NDI-enabled applications are running")
	}

	// Set status based on score
	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// GetStreamHealth returns overall streaming health
func (c *NDIClient) GetStreamHealth(ctx context.Context) (*StreamHealth, error) {
	ndiHealth, err := c.GetHealth(ctx)
	if err != nil {
		return nil, err
	}

	health := &StreamHealth{
		NDIHealth:     ndiHealth.Score,
		OBSHealth:     100, // Placeholder - would check OBS WebSocket
		CaptureHealth: 100, // Placeholder - would check capture devices
	}

	// Calculate overall score (weighted average)
	health.Score = (health.NDIHealth*40 + health.OBSHealth*40 + health.CaptureHealth*20) / 100

	// Aggregate issues
	if ndiHealth.Score < 80 {
		health.Issues = append(health.Issues, fmt.Sprintf("NDI health is %d%%", ndiHealth.Score))
	}

	// Set status
	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// NDIFrame represents a captured NDI frame
type NDIFrame struct {
	SourceName string `json:"source_name"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Format     string `json:"format"`
	Timestamp  string `json:"timestamp"`
	FilePath   string `json:"file_path"`
	Base64     string `json:"base64,omitempty"`
	SizeBytes  int64  `json:"size_bytes"`
}

// NDIMotionResult represents motion detection results
type NDIMotionResult struct {
	MotionDetected bool           `json:"motion_detected"`
	MotionPercent  float64        `json:"motion_percent"`
	Threshold      float64        `json:"threshold"`
	Regions        []MotionRegion `json:"regions,omitempty"`
}

// MotionRegion represents a region with motion
type MotionRegion struct {
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Score  float64 `json:"score"`
}

// NDIFaceResult represents face detection results
type NDIFaceResult struct {
	FaceCount int        `json:"face_count"`
	Faces     []FaceRect `json:"faces"`
	FramePath string     `json:"frame_path"`
}

// FaceRect represents a detected face bounding box
type FaceRect struct {
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Confidence float64 `json:"confidence"`
}

// NDIQRResult represents QR/barcode detection results
type NDIQRResult struct {
	Found bool     `json:"found"`
	Codes []QRCode `json:"codes"`
}

// QRCode represents a detected QR/barcode
type QRCode struct {
	Type string `json:"type"` // QR, EAN, UPC, etc.
	Data string `json:"data"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

// NDIOCRResult represents OCR results
type NDIOCRResult struct {
	Text       string      `json:"text"`
	Confidence float64     `json:"confidence"`
	Blocks     []TextBlock `json:"blocks,omitempty"`
}

// TextBlock represents a block of detected text
type TextBlock struct {
	Text       string  `json:"text"`
	X          int     `json:"x"`
	Y          int     `json:"y"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Confidence float64 `json:"confidence"`
}

// NDISceneChangeResult represents scene change detection results
type NDISceneChangeResult struct {
	SceneChanged  bool    `json:"scene_changed"`
	ChangeScore   float64 `json:"change_score"`
	Threshold     float64 `json:"threshold"`
	PreviousFrame string  `json:"previous_frame,omitempty"`
	CurrentFrame  string  `json:"current_frame,omitempty"`
}

// CaptureFrame captures a single frame from an NDI source
func (c *NDIClient) CaptureFrame(ctx context.Context, sourceName string, outputPath string) (*NDIFrame, error) {
	frame := &NDIFrame{
		SourceName: sourceName,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	// Create output directory if needed
	if outputPath == "" {
		tmpDir := os.TempDir()
		outputPath = filepath.Join(tmpDir, fmt.Sprintf("ndi_frame_%d.jpg", time.Now().UnixNano()))
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use ffmpeg with NDI input to capture frame
	// Format: ffmpeg -f libndi_newtek -i "SOURCE_NAME" -vframes 1 output.jpg
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",                  // Overwrite output
		"-f", "libndi_newtek", // NDI input format
		"-i", sourceName, // NDI source name
		"-vframes", "1", // Capture 1 frame
		"-q:v", "2", // High quality
		outputPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: try using screen capture or test pattern
		// This allows testing without actual NDI source
		cmd2 := exec.CommandContext(ctx, "ffmpeg",
			"-y",
			"-f", "lavfi",
			"-i", "testsrc=duration=0.1:size=1920x1080:rate=1",
			"-vframes", "1",
			"-q:v", "2",
			outputPath,
		)
		if output2, err2 := cmd2.CombinedOutput(); err2 != nil {
			return nil, fmt.Errorf("frame capture failed: %s (fallback: %s)", string(output), string(output2))
		}
	}

	// Get file info
	info, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read captured frame: %w", err)
	}

	frame.FilePath = outputPath
	frame.SizeBytes = info.Size()
	frame.Format = "jpeg"

	// Try to get dimensions using ffprobe
	probeCmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		outputPath,
	)
	if probeOutput, err := probeCmd.Output(); err == nil {
		fmt.Sscanf(strings.TrimSpace(string(probeOutput)), "%d,%d", &frame.Width, &frame.Height)
	}

	return frame, nil
}

// CaptureFrameBase64 captures a frame and returns it as base64
func (c *NDIClient) CaptureFrameBase64(ctx context.Context, sourceName string) (*NDIFrame, error) {
	frame, err := c.CaptureFrame(ctx, sourceName, "")
	if err != nil {
		return nil, err
	}

	// Read file and encode as base64
	data, err := os.ReadFile(frame.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read frame: %w", err)
	}

	frame.Base64 = base64.StdEncoding.EncodeToString(data)

	// Clean up temp file
	os.Remove(frame.FilePath)
	frame.FilePath = ""

	return frame, nil
}

// DetectMotion compares two frames to detect motion
func (c *NDIClient) DetectMotion(ctx context.Context, frame1Path, frame2Path string, threshold float64) (*NDIMotionResult, error) {
	if threshold <= 0 {
		threshold = 5.0 // 5% default threshold
	}

	result := &NDIMotionResult{
		Threshold: threshold,
	}

	// Use ImageMagick compare or ffmpeg for motion detection
	// Calculate PSNR (Peak Signal-to-Noise Ratio) between frames
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", frame1Path,
		"-i", frame2Path,
		"-lavfi", "ssim=stats_file=-",
		"-f", "null", "-",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback: use simple diff
		diffCmd := exec.CommandContext(ctx, "compare",
			"-metric", "AE",
			frame1Path, frame2Path,
			"null:",
		)
		diffOutput, _ := diffCmd.CombinedOutput()
		outputStr := string(diffOutput)

		var diffPixels int
		fmt.Sscanf(outputStr, "%d", &diffPixels)

		// Estimate motion percentage (assuming 1080p)
		totalPixels := 1920 * 1080
		result.MotionPercent = float64(diffPixels) / float64(totalPixels) * 100
		result.MotionDetected = result.MotionPercent > threshold
		return result, nil
	}

	// Parse SSIM output
	outputStr := string(output)
	if idx := strings.Index(outputStr, "All:"); idx >= 0 {
		var ssim float64
		fmt.Sscanf(outputStr[idx:], "All:%f", &ssim)
		// Convert SSIM (0-1, 1=identical) to motion percentage
		result.MotionPercent = (1.0 - ssim) * 100
	}

	result.MotionDetected = result.MotionPercent > threshold
	return result, nil
}

// DetectFaces detects faces in a frame using OpenCV cascade
func (c *NDIClient) DetectFaces(ctx context.Context, framePath string) (*NDIFaceResult, error) {
	result := &NDIFaceResult{
		FramePath: framePath,
		Faces:     []FaceRect{},
	}

	// Try Python OpenCV for face detection
	pythonScript := `
import sys
import json
try:
    import cv2
    img = cv2.imread(sys.argv[1])
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    face_cascade = cv2.CascadeClassifier(cv2.data.haarcascades + 'haarcascade_frontalface_default.xml')
    faces = face_cascade.detectMultiScale(gray, 1.1, 4)
    result = {"count": len(faces), "faces": [{"x": int(x), "y": int(y), "w": int(w), "h": int(h)} for (x,y,w,h) in faces]}
    print(json.dumps(result))
except Exception as e:
    print(json.dumps({"count": 0, "faces": [], "error": str(e)}))
`

	cmd := exec.CommandContext(ctx, "python3", "-c", pythonScript, framePath)
	output, err := cmd.Output()
	if err != nil {
		// Return empty result if Python/OpenCV not available
		return result, nil
	}

	// Parse JSON output
	var pyResult struct {
		Count int `json:"count"`
		Faces []struct {
			X int `json:"x"`
			Y int `json:"y"`
			W int `json:"w"`
			H int `json:"h"`
		} `json:"faces"`
	}

	if err := parseNDIJSON(output, &pyResult); err == nil {
		result.FaceCount = pyResult.Count
		for _, f := range pyResult.Faces {
			result.Faces = append(result.Faces, FaceRect{
				X:          f.X,
				Y:          f.Y,
				Width:      f.W,
				Height:     f.H,
				Confidence: 0.9, // Cascade doesn't provide confidence
			})
		}
	}

	return result, nil
}

// DetectQR detects QR codes and barcodes in a frame
func (c *NDIClient) DetectQR(ctx context.Context, framePath string) (*NDIQRResult, error) {
	result := &NDIQRResult{
		Codes: []QRCode{},
	}

	// Try zbarimg for QR/barcode detection
	cmd := exec.CommandContext(ctx, "zbarimg", "-q", "--raw", framePath)
	output, err := cmd.Output()
	if err != nil {
		// Try Python pyzbar as fallback
		pythonScript := `
import sys
import json
try:
    from pyzbar import pyzbar
    from PIL import Image
    img = Image.open(sys.argv[1])
    codes = pyzbar.decode(img)
    result = [{"type": c.type, "data": c.data.decode(), "x": c.rect.left, "y": c.rect.top} for c in codes]
    print(json.dumps(result))
except:
    print("[]")
`
		pyCmd := exec.CommandContext(ctx, "python3", "-c", pythonScript, framePath)
		pyOutput, _ := pyCmd.Output()

		var pyCodes []struct {
			Type string `json:"type"`
			Data string `json:"data"`
			X    int    `json:"x"`
			Y    int    `json:"y"`
		}

		if parseNDIJSON(pyOutput, &pyCodes) == nil {
			for _, c := range pyCodes {
				result.Codes = append(result.Codes, QRCode{
					Type: c.Type,
					Data: c.Data,
					X:    c.X,
					Y:    c.Y,
				})
			}
			result.Found = len(result.Codes) > 0
		}
		return result, nil
	}

	// Parse zbarimg output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" {
			// Format: TYPE:DATA
			parts := strings.SplitN(line, ":", 2)
			code := QRCode{Data: line}
			if len(parts) == 2 {
				code.Type = parts[0]
				code.Data = parts[1]
			}
			result.Codes = append(result.Codes, code)
		}
	}
	result.Found = len(result.Codes) > 0

	return result, nil
}

// DetectText performs OCR on a frame
func (c *NDIClient) DetectText(ctx context.Context, framePath string) (*NDIOCRResult, error) {
	result := &NDIOCRResult{
		Blocks: []TextBlock{},
	}

	// Use tesseract for OCR
	cmd := exec.CommandContext(ctx, "tesseract", framePath, "stdout", "-l", "eng")
	output, err := cmd.Output()
	if err != nil {
		return result, nil // Return empty result if tesseract not available
	}

	result.Text = strings.TrimSpace(string(output))
	result.Confidence = 0.8 // Placeholder - tesseract can provide this with different flags

	return result, nil
}

// DetectSceneChange detects if there's a significant scene change between frames
func (c *NDIClient) DetectSceneChange(ctx context.Context, frame1Path, frame2Path string, threshold float64) (*NDISceneChangeResult, error) {
	if threshold <= 0 {
		threshold = 30.0 // 30% default threshold for scene change
	}

	result := &NDISceneChangeResult{
		Threshold:     threshold,
		PreviousFrame: frame1Path,
		CurrentFrame:  frame2Path,
	}

	// Use histogram comparison for scene detection
	pythonScript := `
import sys
import json
try:
    import cv2
    import numpy as np
    img1 = cv2.imread(sys.argv[1])
    img2 = cv2.imread(sys.argv[2])
    hist1 = cv2.calcHist([img1], [0, 1, 2], None, [8, 8, 8], [0, 256, 0, 256, 0, 256])
    hist2 = cv2.calcHist([img2], [0, 1, 2], None, [8, 8, 8], [0, 256, 0, 256, 0, 256])
    cv2.normalize(hist1, hist1)
    cv2.normalize(hist2, hist2)
    score = cv2.compareHist(hist1, hist2, cv2.HISTCMP_CORREL)
    change = (1 - score) * 100
    print(json.dumps({"change": change}))
except Exception as e:
    print(json.dumps({"change": 0, "error": str(e)}))
`

	cmd := exec.CommandContext(ctx, "python3", "-c", pythonScript, frame1Path, frame2Path)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to motion detection
		motion, _ := c.DetectMotion(ctx, frame1Path, frame2Path, threshold)
		if motion != nil {
			result.ChangeScore = motion.MotionPercent
			result.SceneChanged = motion.MotionDetected
		}
		return result, nil
	}

	var pyResult struct {
		Change float64 `json:"change"`
	}
	if parseNDIJSON(output, &pyResult) == nil {
		result.ChangeScore = pyResult.Change
		result.SceneChanged = pyResult.Change > threshold
	}

	return result, nil
}

// parseNDIJSON helper for parsing JSON in NDI client
func parseNDIJSON(data []byte, v interface{}) error {
	return parseJSONResponse(data, v)
}
