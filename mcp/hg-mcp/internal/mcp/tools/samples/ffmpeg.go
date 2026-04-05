// Package samples provides DJ sample extraction tools for hg-mcp.
package samples

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hairglasses-studio/mcpkit/sanitize"
)

const (
	// ProcessTimeout for extract/process operations (long audio files)
	ProcessTimeout = 10 * time.Minute
	// ProbeTimeout for probe operations
	ProbeTimeout = 30 * time.Second
)

// ProbeResult holds parsed ffprobe metadata
type ProbeResult struct {
	Duration   float64 `json:"duration"`
	SampleRate int     `json:"sample_rate"`
	BitRate    int     `json:"bit_rate"`
	Channels   int     `json:"channels"`
	Codec      string  `json:"codec"`
	Format     string  `json:"format"`
	FileSize   int64   `json:"file_size"`
	// Video fields (only populated for video files)
	Width  int     `json:"width,omitempty"`
	Height int     `json:"height,omitempty"`
	FPS    float64 `json:"fps,omitempty"`
}

// runFFmpeg executes an ffmpeg command and returns stdout, stderr, and any error
func runFFmpeg(ctx context.Context, args ...string) (string, string, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", "", fmt.Errorf("ffmpeg not found in PATH - please install ffmpeg")
	}

	ctx, cancel := context.WithTimeout(ctx, ProcessTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// runFFprobe executes ffprobe and returns parsed metadata
func runFFprobe(ctx context.Context, filePath string) (*ProbeResult, error) {
	if err := sanitize.MediaPath(filePath); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, fmt.Errorf("ffprobe not found in PATH - please install ffmpeg")
	}

	ctx, cancel := context.WithTimeout(ctx, ProbeTimeout)
	defer cancel()

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var result struct {
		Streams []struct {
			CodecType    string `json:"codec_type"`
			CodecName    string `json:"codec_name"`
			SampleRate   string `json:"sample_rate"`
			Channels     int    `json:"channels"`
			BitRate      string `json:"bit_rate"`
			Width        int    `json:"width"`
			Height       int    `json:"height"`
			AvgFrameRate string `json:"avg_frame_rate"`
			RFrameRate   string `json:"r_frame_rate"`
		} `json:"streams"`
		Format struct {
			FormatName string `json:"format_name"`
			Duration   string `json:"duration"`
			Size       string `json:"size"`
			BitRate    string `json:"bit_rate"`
		} `json:"format"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	info := &ProbeResult{
		Format: result.Format.FormatName,
	}

	// Parse duration
	if result.Format.Duration != "" {
		fmt.Sscanf(result.Format.Duration, "%f", &info.Duration)
	}

	// Parse file size
	if result.Format.Size != "" {
		fmt.Sscanf(result.Format.Size, "%d", &info.FileSize)
	}

	// Parse format-level bit rate
	if result.Format.BitRate != "" {
		var br int
		fmt.Sscanf(result.Format.BitRate, "%d", &br)
		info.BitRate = br
	}

	// Find audio and video streams
	for _, stream := range result.Streams {
		switch stream.CodecType {
		case "audio":
			info.Codec = stream.CodecName
			info.Channels = stream.Channels
			if stream.SampleRate != "" {
				fmt.Sscanf(stream.SampleRate, "%d", &info.SampleRate)
			}
			if stream.BitRate != "" {
				var br int
				fmt.Sscanf(stream.BitRate, "%d", &br)
				if br > 0 {
					info.BitRate = br
				}
			}
		case "video":
			info.Width = stream.Width
			info.Height = stream.Height
			fr := stream.AvgFrameRate
			if fr == "" {
				fr = stream.RFrameRate
			}
			if fr != "" {
				var num, den float64
				if n, _ := fmt.Sscanf(fr, "%f/%f", &num, &den); n == 2 && den != 0 {
					info.FPS = num / den
				}
			}
		}
	}

	return info, nil
}

// isVideoFile checks if a file has a video stream
func isVideoFile(ctx context.Context, path string) bool {
	probe, err := runFFprobe(ctx, path)
	if err != nil {
		return false
	}
	return probe.Width > 0 && probe.Height > 0
}

// getAudioDuration returns the duration of an audio file in seconds
func getAudioDuration(ctx context.Context, path string) (float64, error) {
	probe, err := runFFprobe(ctx, path)
	if err != nil {
		return 0, err
	}
	return probe.Duration, nil
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return info.IsDir()
}

// formatDuration formats a duration in seconds to human-readable string
func formatDuration(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	ms := int((seconds - float64(int(seconds))) * 1000)

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	if ms > 0 {
		return fmt.Sprintf("%d.%ds", secs, ms/100)
	}
	return fmt.Sprintf("%ds", secs)
}

// formatFileSize formats bytes to human-readable string
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// defaultSamplesDir returns ~/Music/Samples, creating it if needed
func defaultSamplesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, "Music", "Samples")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create samples directory: %w", err)
	}
	return dir, nil
}

// parseTimestamp converts MM:SS, HH:MM:SS, or raw seconds to an ffmpeg-compatible timestamp string.
// Returns the original string if already in a valid format.
func parseTimestamp(ts string) string {
	ts = strings.TrimSpace(ts)
	// If it's a pure number (seconds), return as-is — ffmpeg accepts seconds
	if _, err := strconv.ParseFloat(ts, 64); err == nil {
		return ts
	}
	// If it has colons, it's already MM:SS or HH:MM:SS
	return ts
}

// audioExtensions are file extensions treated as audio files
var audioExtensions = map[string]bool{
	".aiff": true,
	".aif":  true,
	".wav":  true,
	".mp3":  true,
	".flac": true,
	".m4a":  true,
	".ogg":  true,
}

// isAudioFile checks if a file path has an audio extension
func isAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return audioExtensions[ext]
}

// SilenceRegion represents a detected silence gap
type SilenceRegion struct {
	Start    float64
	End      float64
	Duration float64
}

// parseSilenceDetectOutput parses ffmpeg silencedetect filter stderr output
func parseSilenceDetectOutput(stderr string) []SilenceRegion {
	var regions []SilenceRegion
	var currentStart float64
	hasStart := false

	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, "silence_start:"); idx >= 0 {
			valStr := strings.TrimSpace(line[idx+len("silence_start:"):])
			// May have trailing info after space
			if spIdx := strings.IndexByte(valStr, ' '); spIdx >= 0 {
				valStr = valStr[:spIdx]
			}
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				currentStart = v
				hasStart = true
			}
		}
		if idx := strings.Index(line, "silence_end:"); idx >= 0 {
			valStr := strings.TrimSpace(line[idx+len("silence_end:"):])
			if spIdx := strings.IndexByte(valStr, ' '); spIdx >= 0 {
				valStr = valStr[:spIdx]
			}
			if v, err := strconv.ParseFloat(valStr, 64); err == nil && hasStart {
				dur := v - currentStart
				// Also try to parse the explicit duration
				if durIdx := strings.Index(line, "silence_duration:"); durIdx >= 0 {
					durStr := strings.TrimSpace(line[durIdx+len("silence_duration:"):])
					if d, err2 := strconv.ParseFloat(durStr, 64); err2 == nil {
						dur = d
					}
				}
				regions = append(regions, SilenceRegion{
					Start:    currentStart,
					End:      v,
					Duration: dur,
				})
				hasStart = false
			}
		}
	}
	return regions
}

// parseTimestampToSeconds converts various timestamp formats to seconds.
// Supports: "HH:MM:SS,mmm", "HH:MM:SS.mmm", "HH:MM:SS", "MM:SS", "SS.mmm", plain seconds.
func parseTimestampToSeconds(ts string) (float64, error) {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return 0, fmt.Errorf("empty timestamp")
	}
	// Try plain number first
	if v, err := strconv.ParseFloat(ts, 64); err == nil {
		return v, nil
	}
	// Replace comma with dot for SRT format
	ts = strings.Replace(ts, ",", ".", 1)
	parts := strings.Split(ts, ":")
	switch len(parts) {
	case 2: // MM:SS or MM:SS.mmm
		min, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %s", parts[0])
		}
		sec, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %s", parts[1])
		}
		return float64(min)*60 + sec, nil
	case 3: // HH:MM:SS or HH:MM:SS.mmm
		hours, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %s", parts[0])
		}
		min, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %s", parts[1])
		}
		sec, err := strconv.ParseFloat(parts[2], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %s", parts[2])
		}
		return float64(hours)*3600 + float64(min)*60 + sec, nil
	}
	return 0, fmt.Errorf("unrecognized timestamp format: %s", ts)
}

// readPCMf32 extracts raw 32-bit float PCM data from an audio file via ffmpeg.
// Returns mono samples at the given sample rate.
func readPCMf32(ctx context.Context, audioPath string, sampleRate int) ([]float32, error) {
	if err := sanitize.MediaPath(audioPath); err != nil {
		return nil, fmt.Errorf("invalid audio path: %w", err)
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH")
	}

	ctx, cancel := context.WithTimeout(ctx, ProcessTimeout)
	defer cancel()

	args := []string{
		"-i", audioPath,
		"-ac", "1",
		"-ar", fmt.Sprintf("%d", sampleRate),
		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"pipe:1",
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg PCM extraction failed: %w — %s", err, stderr.String())
	}
	_ = stdout

	// Convert byte slice to float32 slice
	if len(out)%4 != 0 {
		out = out[:len(out)-(len(out)%4)]
	}
	samples := make([]float32, len(out)/4)
	for i := range samples {
		bits := uint32(out[i*4]) | uint32(out[i*4+1])<<8 | uint32(out[i*4+2])<<16 | uint32(out[i*4+3])<<24
		samples[i] = math.Float32frombits(bits)
	}

	return samples, nil
}
