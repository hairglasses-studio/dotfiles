// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RetroGamingClient provides access to emulators and retro gaming systems
type RetroGamingClient struct {
	pcsx2Path     string
	retroarchPath string
	gamesPath     string
}

// PS2Status represents PCSX2 emulator status
type PS2Status struct {
	Running    bool    `json:"running"`
	GameLoaded string  `json:"game_loaded,omitempty"`
	FPS        float64 `json:"fps,omitempty"`
	Speed      string  `json:"speed,omitempty"`
	Resolution string  `json:"resolution,omitempty"`
	Renderer   string  `json:"renderer,omitempty"`
}

// PS2Game represents a PS2 game
type PS2Game struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   string `json:"size"`
	Format string `json:"format"` // ISO, BIN, etc.
}

// SaveState represents a save state
type SaveState struct {
	Slot      int       `json:"slot"`
	Game      string    `json:"game"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// CaptureDevice represents a video capture device
type CaptureDevice struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	Resolution string `json:"resolution,omitempty"`
	FPS        int    `json:"fps,omitempty"`
	Connected  bool   `json:"connected"`
}

// NewRetroGamingClient creates a new retro gaming client
func NewRetroGamingClient() (*RetroGamingClient, error) {
	pcsx2Path := os.Getenv("PCSX2_PATH")
	if pcsx2Path == "" {
		// Try common locations
		commonPaths := []string{
			"/Applications/PCSX2.app/Contents/MacOS/PCSX2",
			"/usr/bin/pcsx2",
			"C:\\Program Files\\PCSX2\\pcsx2.exe",
		}
		for _, p := range commonPaths {
			if _, err := os.Stat(p); err == nil {
				pcsx2Path = p
				break
			}
		}
	}

	retroarchPath := os.Getenv("RETROARCH_PATH")
	if retroarchPath == "" {
		retroarchPath, _ = exec.LookPath("retroarch")
	}

	gamesPath := os.Getenv("PS2_GAMES_PATH")
	if gamesPath == "" {
		gamesPath = filepath.Join(os.Getenv("HOME"), "Games", "PS2")
	}

	return &RetroGamingClient{
		pcsx2Path:     pcsx2Path,
		retroarchPath: retroarchPath,
		gamesPath:     gamesPath,
	}, nil
}

// GetPS2Status returns PCSX2 emulator status
func (c *RetroGamingClient) GetPS2Status(ctx context.Context) (*PS2Status, error) {
	status := &PS2Status{
		Running: false,
	}

	// Check if PCSX2 is running
	cmd := exec.CommandContext(ctx, "pgrep", "-x", "pcsx2")
	if err := cmd.Run(); err == nil {
		status.Running = true
		// In a real implementation, we'd query PCSX2's status
		// via its IPC interface or log parsing
	}

	// Also check for PCSX2-QT variant
	if !status.Running {
		cmd = exec.CommandContext(ctx, "pgrep", "-f", "PCSX2")
		if err := cmd.Run(); err == nil {
			status.Running = true
		}
	}

	return status, nil
}

// ListPS2Games lists available PS2 games
func (c *RetroGamingClient) ListPS2Games(ctx context.Context) ([]PS2Game, error) {
	games := []PS2Game{}

	if c.gamesPath == "" {
		return games, nil
	}

	// Check if games directory exists
	if _, err := os.Stat(c.gamesPath); os.IsNotExist(err) {
		return games, nil
	}

	// Scan for ISO/BIN files
	extensions := []string{".iso", ".bin", ".img", ".mdf", ".nrg"}

	err := filepath.Walk(c.gamesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, validExt := range extensions {
			if ext == validExt {
				// Get file size
				size := formatFileSize(info.Size())

				games = append(games, PS2Game{
					Name:   strings.TrimSuffix(info.Name(), ext),
					Path:   path,
					Size:   size,
					Format: strings.ToUpper(strings.TrimPrefix(ext, ".")),
				})
				break
			}
		}
		return nil
	})

	if err != nil {
		return games, err
	}

	return games, nil
}

// formatFileSize formats bytes to human readable
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ListSaveStates lists save states for a game
func (c *RetroGamingClient) ListSaveStates(ctx context.Context, gameName string) ([]SaveState, error) {
	states := []SaveState{}

	// PCSX2 save states are typically in ~/.config/PCSX2/sstates/
	sstatesPath := filepath.Join(os.Getenv("HOME"), ".config", "PCSX2", "sstates")

	if _, err := os.Stat(sstatesPath); os.IsNotExist(err) {
		return states, nil
	}

	files, err := os.ReadDir(sstatesPath)
	if err != nil {
		return states, nil
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		// PCSX2 save states are named like: GAME_ID.XX.p2s
		if strings.HasSuffix(name, ".p2s") {
			info, err := file.Info()
			if err != nil {
				continue
			}

			// Extract slot number
			slot := 0
			parts := strings.Split(name, ".")
			if len(parts) >= 2 {
				fmt.Sscanf(parts[len(parts)-2], "%d", &slot)
			}

			states = append(states, SaveState{
				Slot:      slot,
				Game:      strings.Split(name, ".")[0],
				Timestamp: info.ModTime(),
				Path:      filepath.Join(sstatesPath, name),
			})
		}
	}

	return states, nil
}

// GetCaptureDevices lists video capture devices
func (c *RetroGamingClient) GetCaptureDevices(ctx context.Context) ([]CaptureDevice, error) {
	devices := []CaptureDevice{}

	// On Linux, check /dev/video*
	videoDevices, _ := filepath.Glob("/dev/video*")
	for _, dev := range videoDevices {
		devices = append(devices, CaptureDevice{
			Name:      filepath.Base(dev),
			Path:      dev,
			Connected: true,
		})
	}

	// If no devices found, return placeholder
	if len(devices) == 0 {
		devices = append(devices, CaptureDevice{
			Name:      "No capture devices detected",
			Connected: false,
		})
	}

	return devices, nil
}

// PCSX2Path returns the configured PCSX2 path
func (c *RetroGamingClient) PCSX2Path() string {
	return c.pcsx2Path
}

// GamesPath returns the configured games path
func (c *RetroGamingClient) GamesPath() string {
	return c.gamesPath
}

// RetroArchPath returns the configured RetroArch path
func (c *RetroGamingClient) RetroArchPath() string {
	return c.retroarchPath
}
