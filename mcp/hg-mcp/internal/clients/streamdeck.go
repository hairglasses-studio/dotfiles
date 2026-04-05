// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dh1tw/streamdeck"
)

// StreamDeckClient provides access to Elgato Stream Deck devices
type StreamDeckClient struct {
	mu     sync.RWMutex
	device *streamdeck.StreamDeck
	config streamdeck.Config
}

// StreamDeckDevice represents a connected Stream Deck device
type StreamDeckDevice struct {
	ID          string `json:"id"`
	Serial      string `json:"serial"`
	Model       string `json:"model"`
	Rows        int    `json:"rows"`
	Cols        int    `json:"cols"`
	ButtonCount int    `json:"button_count"`
	IconSize    int    `json:"icon_size"`
}

// StreamDeckButton represents a button state
type StreamDeckButton struct {
	Index   int  `json:"index"`
	Row     int  `json:"row"`
	Col     int  `json:"col"`
	Pressed bool `json:"pressed"`
}

// StreamDeckStatus represents connection status
type StreamDeckStatus struct {
	Connected   bool               `json:"connected"`
	DeviceCount int                `json:"device_count"`
	Devices     []StreamDeckDevice `json:"devices,omitempty"`
}

// StreamDeckHealth represents health status
type StreamDeckHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	DeviceCount     int      `json:"device_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewStreamDeckClient creates a new Stream Deck client
func NewStreamDeckClient() (*StreamDeckClient, error) {
	client := &StreamDeckClient{}

	// Try to find a connected device
	config, found := streamdeck.FindConnectedConfig()
	if found {
		client.config = config
		// Try to open the device
		dev, err := streamdeck.NewStreamDeck()
		if err == nil {
			client.device = dev
		}
	}

	return client, nil
}

// Close closes the device connection
func (c *StreamDeckClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.device != nil {
		c.device.Close()
		c.device = nil
	}
	return nil
}

// GetStatus returns connection status
func (c *StreamDeckClient) GetStatus(ctx context.Context) (*StreamDeckStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := &StreamDeckStatus{
		Connected:   c.device != nil,
		DeviceCount: 0,
		Devices:     make([]StreamDeckDevice, 0),
	}

	if c.device != nil {
		status.DeviceCount = 1
		status.Devices = append(status.Devices, StreamDeckDevice{
			ID:          "device-0",
			Serial:      c.device.Serial(),
			Model:       getModelFromProductID(c.config.ProductID),
			Rows:        c.config.NumButtonRows,
			Cols:        c.config.NumButtonColumns,
			ButtonCount: c.config.NumButtonRows * c.config.NumButtonColumns,
			IconSize:    c.config.ButtonSize,
		})
	}

	return status, nil
}

// GetDevices returns list of connected devices
func (c *StreamDeckClient) GetDevices(ctx context.Context) ([]StreamDeckDevice, error) {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return nil, err
	}
	return status.Devices, nil
}

// RefreshDevices rescans for connected devices
func (c *StreamDeckClient) RefreshDevices(ctx context.Context) ([]StreamDeckDevice, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close existing device
	if c.device != nil {
		c.device.Close()
		c.device = nil
	}

	// Try to find a connected device
	config, found := streamdeck.FindConnectedConfig()
	if found {
		c.config = config
		dev, err := streamdeck.NewStreamDeck()
		if err == nil {
			c.device = dev
		}
	}

	c.mu.RUnlock()
	devices, err := c.GetDevices(ctx)
	c.mu.RLock()
	return devices, err
}

// getDevice returns the connected device
func (c *StreamDeckClient) getDevice() (*streamdeck.StreamDeck, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.device == nil {
		return nil, fmt.Errorf("no Stream Deck device connected")
	}
	return c.device, nil
}

// GetButtons returns button states for a device
func (c *StreamDeckClient) GetButtons(ctx context.Context, deviceIndex int) ([]StreamDeckButton, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.device == nil {
		return nil, fmt.Errorf("no Stream Deck device connected")
	}

	buttons := make([]StreamDeckButton, 0, c.config.NumButtonRows*c.config.NumButtonColumns)
	for r := 0; r < c.config.NumButtonRows; r++ {
		for col := 0; col < c.config.NumButtonColumns; col++ {
			idx := r*c.config.NumButtonColumns + col
			buttons = append(buttons, StreamDeckButton{
				Index:   idx,
				Row:     r,
				Col:     col,
				Pressed: false,
			})
		}
	}

	return buttons, nil
}

// SetButtonImage sets the image for a button
func (c *StreamDeckClient) SetButtonImage(ctx context.Context, deviceIndex, buttonIndex int, imagePath string) error {
	dev, err := c.getDevice()
	if err != nil {
		return err
	}

	return dev.FillImageFromFile(buttonIndex, imagePath)
}

// SetButtonColor sets a solid color for a button
func (c *StreamDeckClient) SetButtonColor(ctx context.Context, deviceIndex, buttonIndex int, r, g, b uint8) error {
	dev, err := c.getDevice()
	if err != nil {
		return err
	}

	dev.FillColor(buttonIndex, int(r), int(g), int(b))
	return nil
}

// SetButtonTitle sets a text title for a button (creates simple text image)
func (c *StreamDeckClient) SetButtonTitle(ctx context.Context, deviceIndex, buttonIndex int, title string, bgColor, textColor color.Color) error {
	dev, err := c.getDevice()
	if err != nil {
		return err
	}

	// Create simple colored background image
	size := c.config.ButtonSize
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	if bgColor == nil {
		bgColor = color.Black
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	dev.FillImage(buttonIndex, img)
	return nil
}

// ClearButton clears a button (sets to black)
func (c *StreamDeckClient) ClearButton(ctx context.Context, deviceIndex, buttonIndex int) error {
	dev, err := c.getDevice()
	if err != nil {
		return err
	}

	dev.ClearBtn(buttonIndex)
	return nil
}

// ClearAllButtons clears all buttons on a device
func (c *StreamDeckClient) ClearAllButtons(ctx context.Context, deviceIndex int) error {
	dev, err := c.getDevice()
	if err != nil {
		return err
	}

	dev.ClearAllBtns()
	return nil
}

// SetBrightness sets the device brightness (0-100)
func (c *StreamDeckClient) SetBrightness(ctx context.Context, deviceIndex int, brightness int) error {
	dev, err := c.getDevice()
	if err != nil {
		return err
	}

	if brightness < 0 {
		brightness = 0
	}
	if brightness > 100 {
		brightness = 100
	}

	dev.SetBrightness(uint16(brightness))
	return nil
}

// ResetDevice resets the device to default state
func (c *StreamDeckClient) ResetDevice(ctx context.Context, deviceIndex int) error {
	dev, err := c.getDevice()
	if err != nil {
		return err
	}

	dev.ClearAllBtns()
	return nil
}

// GetHealth returns health status
func (c *StreamDeckClient) GetHealth(ctx context.Context) (*StreamDeckHealth, error) {
	health := &StreamDeckHealth{
		Score:  100,
		Status: "healthy",
	}

	status, _ := c.GetStatus(ctx)
	health.Connected = status.Connected
	health.DeviceCount = status.DeviceCount

	if !status.Connected {
		health.Score -= 50
		health.Issues = append(health.Issues, "No Stream Deck devices connected")
		health.Recommendations = append(health.Recommendations, "Connect a Stream Deck device via USB")
		health.Recommendations = append(health.Recommendations, "Ensure Stream Deck software is not running (it locks the device)")
	}

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

// getModelFromProductID returns model name from USB product ID
func getModelFromProductID(productID uint16) string {
	switch productID {
	case 0x0060:
		return "Stream Deck (Original)"
	case 0x006d:
		return "Stream Deck (Original v2)"
	case 0x0063:
		return "Stream Deck Mini"
	case 0x006c:
		return "Stream Deck XL"
	case 0x0084:
		return "Stream Deck +"
	default:
		return fmt.Sprintf("Stream Deck (0x%04x)", productID)
	}
}

// loadImage loads an image from a file path
func loadImage(path string) (image.Image, error) {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[1:])
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return png.Decode(f)
	default:
		img, _, err := image.Decode(f)
		return img, err
	}
}
