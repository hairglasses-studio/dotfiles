// Package video provides AI video processing tools for hg-mcp.
package video

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// DefaultTimeout for video processing operations (30 minutes)
const DefaultTimeout = 30 * time.Minute

// VideoInfo holds parsed video metadata
type VideoInfo struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FPS        float64 `json:"fps"`
	Duration   float64 `json:"duration"`
	FrameCount int     `json:"frame_count"`
	Codec      string  `json:"codec"`
	Format     string  `json:"format"`
}

// getVidtoolPath returns the path to the vidtool CLI
func getVidtoolPath() string {
	return config.GetEnv("VIDTOOL_PATH", "vidtool")
}

// getOutputDir returns the default output directory
func getOutputDir() string {
	return config.GetEnv("VIDTOOL_OUTPUT_DIR", "")
}

// getDevice returns the compute device (cuda/cpu)
func getDevice() string {
	return config.GetEnv("VIDTOOL_DEVICE", "cuda")
}

// runVidtool executes a vidtool command and returns stdout, stderr, and any error
func runVidtool(ctx context.Context, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, getVidtoolPath(), args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// runFFprobe executes ffprobe to get video metadata
func runFFprobe(ctx context.Context, videoPath string) (*VideoInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	// Parse JSON output
	var result struct {
		Streams []struct {
			CodecType    string `json:"codec_type"`
			CodecName    string `json:"codec_name"`
			Width        int    `json:"width"`
			Height       int    `json:"height"`
			RFrameRate   string `json:"r_frame_rate"`
			NbFrames     string `json:"nb_frames"`
			Duration     string `json:"duration"`
			AvgFrameRate string `json:"avg_frame_rate"`
		} `json:"streams"`
		Format struct {
			Filename   string `json:"filename"`
			FormatName string `json:"format_name"`
			Duration   string `json:"duration"`
		} `json:"format"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	info := &VideoInfo{
		Format: result.Format.FormatName,
	}

	// Parse duration from format
	if result.Format.Duration != "" {
		fmt.Sscanf(result.Format.Duration, "%f", &info.Duration)
	}

	// Find video stream
	for _, stream := range result.Streams {
		if stream.CodecType == "video" {
			info.Width = stream.Width
			info.Height = stream.Height
			info.Codec = stream.CodecName

			// Parse frame rate (format: "30/1" or "30000/1001")
			if stream.AvgFrameRate != "" {
				var num, den float64
				if n, _ := fmt.Sscanf(stream.AvgFrameRate, "%f/%f", &num, &den); n == 2 && den != 0 {
					info.FPS = num / den
				}
			} else if stream.RFrameRate != "" {
				var num, den float64
				if n, _ := fmt.Sscanf(stream.RFrameRate, "%f/%f", &num, &den); n == 2 && den != 0 {
					info.FPS = num / den
				}
			}

			// Parse frame count
			if stream.NbFrames != "" {
				fmt.Sscanf(stream.NbFrames, "%d", &info.FrameCount)
			}

			// If frame count not available, calculate from duration and fps
			if info.FrameCount == 0 && info.Duration > 0 && info.FPS > 0 {
				info.FrameCount = int(info.Duration * info.FPS)
			}

			break
		}
	}

	return info, nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// formatDuration formats a duration in seconds to human-readable string
func formatDuration(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
