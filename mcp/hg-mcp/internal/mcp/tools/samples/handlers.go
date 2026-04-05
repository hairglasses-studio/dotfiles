package samples

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleProbe returns audio/video file metadata
func handleProbe(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(filePath) {
		return tools.ErrorResult(fmt.Errorf("file not found: %s", filePath)), nil
	}

	probe, err := runFFprobe(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to probe file: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# File Information\n\n")
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n\n", filePath))

	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Format | %s |\n", probe.Format))
	sb.WriteString(fmt.Sprintf("| Duration | %s (%.1f sec) |\n", formatDuration(probe.Duration), probe.Duration))
	sb.WriteString(fmt.Sprintf("| Codec | %s |\n", probe.Codec))
	if probe.SampleRate > 0 {
		sb.WriteString(fmt.Sprintf("| Sample Rate | %d Hz |\n", probe.SampleRate))
	}
	if probe.Channels > 0 {
		sb.WriteString(fmt.Sprintf("| Channels | %d |\n", probe.Channels))
	}
	if probe.BitRate > 0 {
		sb.WriteString(fmt.Sprintf("| Bit Rate | %d kbps |\n", probe.BitRate/1000))
	}
	sb.WriteString(fmt.Sprintf("| File Size | %s |\n", formatFileSize(probe.FileSize)))

	// Video properties
	if probe.Width > 0 {
		sb.WriteString(fmt.Sprintf("| Resolution | %dx%d |\n", probe.Width, probe.Height))
	}
	if probe.FPS > 0 {
		sb.WriteString(fmt.Sprintf("| FPS | %.2f |\n", probe.FPS))
	}

	return tools.TextResult(sb.String()), nil
}

// handleList lists existing sample packs
func handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directory := tools.GetStringParam(req, "directory")
	if directory == "" {
		var err error
		directory, err = defaultSamplesDir()
		if err != nil {
			return tools.ErrorResult(err), nil
		}
	}

	if !dirExists(directory) {
		return tools.ErrorResult(fmt.Errorf("directory not found: %s", directory)), nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot read directory: %w", err)), nil
	}

	type packInfo struct {
		Name      string
		FileCount int
		HasXML    bool
		TotalSize int64
	}

	var packs []packInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		packPath := filepath.Join(directory, entry.Name())
		subEntries, err := os.ReadDir(packPath)
		if err != nil {
			continue
		}

		info := packInfo{Name: entry.Name()}
		for _, sub := range subEntries {
			if sub.IsDir() {
				continue
			}
			name := sub.Name()
			// Skip hidden files
			if strings.HasPrefix(name, ".") {
				continue
			}
			if strings.HasSuffix(strings.ToLower(name), ".xml") {
				info.HasXML = true
				continue
			}
			if isAudioFile(name) {
				info.FileCount++
				if fi, err := sub.Info(); err == nil {
					info.TotalSize += fi.Size()
				}
			}
		}

		// Only include directories that contain audio files
		if info.FileCount > 0 {
			packs = append(packs, info)
		}
	}

	var sb strings.Builder
	sb.WriteString("# Sample Packs\n\n")
	sb.WriteString(fmt.Sprintf("**Directory:** `%s`\n\n", directory))

	if len(packs) == 0 {
		sb.WriteString("No sample packs found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Pack | Files | XML | Size |\n")
	sb.WriteString("|------|-------|-----|------|\n")
	for _, p := range packs {
		xmlStr := "No"
		if p.HasXML {
			xmlStr = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s |\n", p.Name, p.FileCount, xmlStr, formatFileSize(p.TotalSize)))
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d packs\n", len(packs)))

	return tools.TextResult(sb.String()), nil
}

// --- Audio Analysis Tools ---

// handleDetectSilence finds silence gaps in an audio file
func handleDetectSilence(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	thresholdDB := tools.GetIntParam(req, "threshold_db", -40)
	minDuration := tools.GetFloatParam(req, "min_duration", 0.5)

	args := []string{
		"-i", audioPath,
		"-af", fmt.Sprintf("silencedetect=noise=%ddB:d=%.2f", thresholdDB, minDuration),
		"-f", "null", "-",
	}

	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("silence detection failed: %w", err)), nil
	}

	regions := parseSilenceDetectOutput(stderr)

	var sb strings.Builder
	sb.WriteString("# Silence Detection\n\n")
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Threshold:** %d dB\n", thresholdDB))
	sb.WriteString(fmt.Sprintf("**Min Duration:** %.2f sec\n\n", minDuration))

	if len(regions) == 0 {
		sb.WriteString("No silence gaps detected.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| # | Start | End | Duration |\n")
	sb.WriteString("|---|-------|-----|----------|\n")
	for i, r := range regions {
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %.2fs |\n", i+1, formatDuration(r.Start), formatDuration(r.End), r.Duration))
	}
	sb.WriteString(fmt.Sprintf("\n**Silence gaps found:** %d\n", len(regions)))

	return tools.TextResult(sb.String()), nil
}

// handleDetectBPM estimates BPM of an audio file using onset detection
func handleDetectBPM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	// Extract PCM at 22050 Hz mono for analysis
	const sampleRate = 22050
	samples, err := readPCMf32(ctx, audioPath, sampleRate)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to read audio: %w", err)), nil
	}

	if len(samples) < sampleRate { // need at least 1 second
		return tools.ErrorResult(fmt.Errorf("audio file too short for BPM detection")), nil
	}

	// Compute energy in windows
	windowSize := sampleRate / 100 // 10ms windows
	numWindows := len(samples) / windowSize
	energy := make([]float64, numWindows)
	for i := 0; i < numWindows; i++ {
		sum := 0.0
		for j := 0; j < windowSize; j++ {
			v := float64(samples[i*windowSize+j])
			sum += v * v
		}
		energy[i] = sum / float64(windowSize)
	}

	// Compute onset detection function (first-order difference, half-wave rectified)
	onset := make([]float64, len(energy))
	for i := 1; i < len(energy); i++ {
		diff := energy[i] - energy[i-1]
		if diff > 0 {
			onset[i] = diff
		}
	}

	// Adaptive threshold for peak picking
	meanOnset := 0.0
	for _, v := range onset {
		meanOnset += v
	}
	if len(onset) > 0 {
		meanOnset /= float64(len(onset))
	}
	threshold := meanOnset * 1.5

	// Find onset peaks
	var peakPositions []int
	for i := 1; i < len(onset)-1; i++ {
		if onset[i] > threshold && onset[i] > onset[i-1] && onset[i] >= onset[i+1] {
			peakPositions = append(peakPositions, i)
		}
	}

	if len(peakPositions) < 2 {
		return tools.TextResult("# BPM Detection\n\nNot enough onsets detected to estimate BPM.\n"), nil
	}

	// Compute inter-onset intervals (IOI) in seconds
	var iois []float64
	windowDuration := float64(windowSize) / float64(sampleRate)
	for i := 1; i < len(peakPositions); i++ {
		ioi := float64(peakPositions[i]-peakPositions[i-1]) * windowDuration
		if ioi > 0.2 && ioi < 2.0 { // 30–300 BPM range
			iois = append(iois, ioi)
		}
	}

	if len(iois) < 2 {
		return tools.TextResult("# BPM Detection\n\nNot enough valid onset intervals to estimate BPM.\n"), nil
	}

	// Histogram-based BPM estimation
	// Quantize IOIs to BPM and find the most common
	bpmCounts := make(map[int]int)
	for _, ioi := range iois {
		bpm := int(math.Round(60.0 / ioi))
		// Normalize to common range 60-200
		for bpm > 200 {
			bpm /= 2
		}
		for bpm < 60 && bpm > 0 {
			bpm *= 2
		}
		if bpm >= 60 && bpm <= 200 {
			bpmCounts[bpm]++
		}
	}

	// Find dominant BPM
	bestBPM := 0
	bestCount := 0
	totalVotes := 0
	for bpm, count := range bpmCounts {
		totalVotes += count
		if count > bestCount {
			bestCount = count
			bestBPM = bpm
		}
	}

	// Also count nearby BPMs (±2) for confidence
	nearbyCount := 0
	for bpm, count := range bpmCounts {
		if abs(bpm-bestBPM) <= 2 {
			nearbyCount += count
		}
	}

	confidence := "low"
	if totalVotes > 0 {
		ratio := float64(nearbyCount) / float64(totalVotes)
		if ratio > 0.6 {
			confidence = "high"
		} else if ratio > 0.3 {
			confidence = "medium"
		}
	}

	var sb strings.Builder
	sb.WriteString("# BPM Detection\n\n")
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Estimated BPM:** %d\n", bestBPM))
	sb.WriteString(fmt.Sprintf("**Confidence:** %s\n", confidence))
	sb.WriteString(fmt.Sprintf("**Onsets detected:** %d\n", len(peakPositions)))

	return tools.TextResult(sb.String()), nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// handleWaveform generates a text-based waveform visualization
func handleWaveform(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	resolution := tools.GetIntParam(req, "resolution", 80)
	if resolution < 10 {
		resolution = 10
	}
	if resolution > 200 {
		resolution = 200
	}
	format := tools.OptionalStringParam(req, "format", "text")

	// Extract PCM at 22050 Hz mono
	const sampleRate = 22050
	samples, err := readPCMf32(ctx, audioPath, sampleRate)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to read audio: %w", err)), nil
	}

	if len(samples) == 0 {
		return tools.ErrorResult(fmt.Errorf("audio file is empty")), nil
	}

	// Compute peak amplitudes per window
	windowSize := len(samples) / resolution
	if windowSize < 1 {
		windowSize = 1
	}
	peaks := make([]float64, resolution)
	for i := 0; i < resolution; i++ {
		start := i * windowSize
		end := start + windowSize
		if end > len(samples) {
			end = len(samples)
		}
		maxAmp := 0.0
		for j := start; j < end; j++ {
			v := math.Abs(float64(samples[j]))
			if v > maxAmp {
				maxAmp = v
			}
		}
		peaks[i] = maxAmp
	}

	// Normalize peaks to 0-1
	maxPeak := 0.0
	for _, p := range peaks {
		if p > maxPeak {
			maxPeak = p
		}
	}
	if maxPeak > 0 {
		for i := range peaks {
			peaks[i] /= maxPeak
		}
	}

	duration := float64(len(samples)) / float64(sampleRate)

	if format == "json" {
		type jsonResult struct {
			Peaks      []float64 `json:"peaks"`
			Duration   float64   `json:"duration"`
			SampleRate int       `json:"sample_rate"`
			Resolution int       `json:"resolution"`
		}
		return tools.JSONResult(jsonResult{
			Peaks:      peaks,
			Duration:   duration,
			SampleRate: sampleRate,
			Resolution: resolution,
		}), nil
	}

	// ASCII art waveform
	var sb strings.Builder
	sb.WriteString("# Waveform\n\n")
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n\n", formatDuration(duration)))

	barChars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	sb.WriteString("```\n")
	for _, p := range peaks {
		idx := int(p * float64(len(barChars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(barChars) {
			idx = len(barChars) - 1
		}
		sb.WriteString(barChars[idx])
	}
	sb.WriteString("\n")

	// Timeline markers
	markerInterval := resolution / 8
	if markerInterval < 1 {
		markerInterval = 1
	}
	for i := 0; i < resolution; i++ {
		if i%markerInterval == 0 {
			sb.WriteString("|")
		} else {
			sb.WriteString(" ")
		}
	}
	sb.WriteString("\n")
	for i := 0; i < resolution; i++ {
		if i%markerInterval == 0 {
			ts := formatDuration(duration * float64(i) / float64(resolution))
			sb.WriteString(ts)
			i += len(ts) - 1
		} else {
			sb.WriteString(" ")
		}
	}
	sb.WriteString("\n```\n")

	return tools.TextResult(sb.String()), nil
}

// --- Organization Tools ---

// handleConvert converts audio between formats
func handleConvert(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	targetFormatRaw, errResult := tools.RequireStringParam(req, "format")
	if errResult != nil {
		return errResult, nil
	}
	targetFormat := strings.ToLower(targetFormatRaw)
	validFormats := map[string]bool{"aiff": true, "wav": true, "mp3": true, "flac": true}
	if !validFormats[targetFormat] {
		return tools.ErrorResult(fmt.Errorf("invalid format: %s (supported: aiff, wav, mp3, flac)", targetFormat)), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
		ext := "." + targetFormat
		if targetFormat == "aiff" {
			ext = ".aiff"
		}
		outputPath = filepath.Join(filepath.Dir(audioPath), base+ext)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	sampleRate := tools.GetIntParam(req, "sample_rate", 0) // 0 = preserve
	bitDepth := tools.GetIntParam(req, "bit_depth", 16)
	mp3Bitrate := tools.GetIntParam(req, "mp3_bitrate", 320)

	args := []string{"-y", "-i", audioPath}

	// Sample rate
	if sampleRate > 0 {
		args = append(args, "-ar", strconv.Itoa(sampleRate))
	}

	// Codec settings based on format
	switch targetFormat {
	case "aiff":
		codec := "pcm_s16be"
		if bitDepth == 24 {
			codec = "pcm_s24be"
		}
		args = append(args, "-c:a", codec)
	case "wav":
		codec := "pcm_s16le"
		if bitDepth == 24 {
			codec = "pcm_s24le"
		}
		args = append(args, "-c:a", codec)
	case "mp3":
		args = append(args, "-c:a", "libmp3lame", "-b:a", fmt.Sprintf("%dk", mp3Bitrate))
	case "flac":
		args = append(args, "-c:a", "flac")
		if bitDepth == 24 {
			args = append(args, "-sample_fmt", "s32")
		}
	}

	args = append(args, outputPath)

	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("conversion failed: %w — %s", err, stderr)), nil
	}

	probe, _ := runFFprobe(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Converted\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Format:** %s\n", strings.ToUpper(targetFormat)))
	if probe != nil {
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(probe.Duration)))
		sb.WriteString(fmt.Sprintf("**File Size:** %s\n", formatFileSize(probe.FileSize)))
	}

	return tools.TextResult(sb.String()), nil
}

// handleRename batch renames samples in a pack using a naming pattern
func handleRename(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directory, errResult := tools.RequireStringParam(req, "directory")
	if errResult != nil {
		return errResult, nil
	}
	if !dirExists(directory) {
		return tools.ErrorResult(fmt.Errorf("directory not found: %s", directory)), nil
	}

	pattern, errResult := tools.RequireStringParam(req, "pattern")
	if errResult != nil {
		return errResult, nil
	}

	dryRun := tools.GetBoolParam(req, "dry_run", true)
	packName := filepath.Base(directory)

	// Find audio files, sorted by name
	entries, err := os.ReadDir(directory)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot read directory: %w", err)), nil
	}

	type fileEntry struct {
		Name string
		Path string
	}
	var audioFiles []fileEntry
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if isAudioFile(e.Name()) {
			audioFiles = append(audioFiles, fileEntry{Name: e.Name(), Path: filepath.Join(directory, e.Name())})
		}
	}

	sort.Slice(audioFiles, func(i, j int) bool { return audioFiles[i].Name < audioFiles[j].Name })

	if len(audioFiles) == 0 {
		return tools.ErrorResult(fmt.Errorf("no audio files found in %s", directory)), nil
	}

	type renameOp struct {
		From string
		To   string
	}
	var ops []renameOp

	for i, f := range audioFiles {
		ext := filepath.Ext(f.Name)
		nameNoExt := strings.TrimSuffix(f.Name, ext)

		newName := pattern
		newName = strings.ReplaceAll(newName, "{pack}", packName)
		newName = strings.ReplaceAll(newName, "{name}", nameNoExt)
		newName = strings.ReplaceAll(newName, "{index}", fmt.Sprintf("%03d", i+1))
		newName = strings.ReplaceAll(newName, "{ext}", strings.TrimPrefix(ext, "."))

		// Ensure extension
		if filepath.Ext(newName) == "" {
			newName += ext
		}

		if newName != f.Name {
			ops = append(ops, renameOp{From: f.Name, To: newName})
		}
	}

	var sb strings.Builder
	if dryRun {
		sb.WriteString("# Rename Preview (Dry Run)\n\n")
	} else {
		sb.WriteString("# Renamed\n\n")
	}
	sb.WriteString(fmt.Sprintf("**Directory:** `%s`\n", directory))
	sb.WriteString(fmt.Sprintf("**Pattern:** `%s`\n\n", pattern))

	if len(ops) == 0 {
		sb.WriteString("No files would be renamed (all names already match pattern).\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| From | To |\n")
	sb.WriteString("|------|----|\n")

	errorCount := 0
	for _, op := range ops {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", op.From, op.To))
		if !dryRun {
			fromPath := filepath.Join(directory, op.From)
			toPath := filepath.Join(directory, op.To)
			if err := os.Rename(fromPath, toPath); err != nil {
				sb.WriteString(fmt.Sprintf("| | **ERROR:** %v |\n", err))
				errorCount++
			}
		}
	}

	sb.WriteString(fmt.Sprintf("\n**Files:** %d renamed", len(ops)))
	if errorCount > 0 {
		sb.WriteString(fmt.Sprintf(", %d errors", errorCount))
	}
	if dryRun {
		sb.WriteString(" (dry run — set dry_run=false to apply)")
	}
	sb.WriteString("\n")

	return tools.TextResult(sb.String()), nil
}
