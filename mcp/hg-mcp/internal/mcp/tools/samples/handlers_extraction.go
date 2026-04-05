package samples

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleExtractAudio extracts full audio from a video file to AIFF
func handleExtractAudio(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	videoPath, errResult := tools.RequireStringParam(req, "video_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(videoPath) {
		return tools.ErrorResult(fmt.Errorf("video file not found: %s", videoPath)), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath == "" {
		samplesDir, err := defaultSamplesDir()
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		base := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
		outputPath = filepath.Join(samplesDir, base+"-full.aiff")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	sampleRate := tools.GetIntParam(req, "sample_rate", 44100)
	bitDepth := tools.GetIntParam(req, "bit_depth", 16)

	codec := "pcm_s16be"
	if bitDepth == 24 {
		codec = "pcm_s24be"
	}

	args := []string{
		"-y", "-i", videoPath,
		"-vn",
		"-acodec", codec,
		"-ar", fmt.Sprintf("%d", sampleRate),
		"-ac", "2",
		outputPath,
	}

	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		var sb strings.Builder
		sb.WriteString("# Extract Audio — Error\n\n")
		sb.WriteString(fmt.Sprintf("Failed to extract audio from `%s`\n\n", videoPath))
		sb.WriteString(fmt.Sprintf("**Error:** %v\n", err))
		if stderr != "" {
			sb.WriteString("\n**Details:**\n```\n")
			sb.WriteString(stderr)
			sb.WriteString("\n```\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	// Probe the output for duration/size
	probe, _ := runFFprobe(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Extract Audio\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", videoPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Format:** AIFF %d-bit / %d Hz / stereo\n", bitDepth, sampleRate))
	if probe != nil {
		sb.WriteString(fmt.Sprintf("**Duration:** %s (%.1f sec)\n", formatDuration(probe.Duration), probe.Duration))
		sb.WriteString(fmt.Sprintf("**File Size:** %s\n", formatFileSize(probe.FileSize)))
	}

	return tools.TextResult(sb.String()), nil
}

// handleClip extracts a time-range clip from an audio file
func handleClip(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	start, errResult := tools.RequireStringParam(req, "start")
	if errResult != nil {
		return errResult, nil
	}

	end, errResult := tools.RequireStringParam(req, "end")
	if errResult != nil {
		return errResult, nil
	}

	outputPath, errResult := tools.RequireStringParam(req, "output_path")
	if errResult != nil {
		return errResult, nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	args := []string{
		"-y", "-i", audioPath,
		"-ss", parseTimestamp(start),
		"-to", parseTimestamp(end),
		"-c:a", "pcm_s16be",
		outputPath,
	}

	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		var sb strings.Builder
		sb.WriteString("# Clip — Error\n\n")
		sb.WriteString(fmt.Sprintf("Failed to clip `%s` [%s → %s]\n\n", audioPath, start, end))
		sb.WriteString(fmt.Sprintf("**Error:** %v\n", err))
		if stderr != "" {
			sb.WriteString("\n**Details:**\n```\n")
			sb.WriteString(stderr)
			sb.WriteString("\n```\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	// Probe output for duration
	duration, _ := getAudioDuration(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Clip Extracted\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Range:** %s → %s\n", start, end))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	if duration > 0 {
		sb.WriteString(fmt.Sprintf("**Duration:** %s (%.1f sec)\n", formatDuration(duration), duration))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSplitBySilence splits audio into segments at silence boundaries
func handleSplitBySilence(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	outputDir := tools.GetStringParam(req, "output_dir")
	if outputDir == "" {
		base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
		outputDir = filepath.Join(filepath.Dir(audioPath), base+"-split")
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	thresholdDB := tools.GetIntParam(req, "threshold_db", -40)
	minSilence := tools.GetFloatParam(req, "min_silence", 0.5)
	minSegment := tools.GetFloatParam(req, "min_segment", 0.5)
	process := tools.GetBoolParam(req, "process", true)
	generateXML := tools.GetBoolParam(req, "generate_xml", true)
	naming := tools.OptionalStringParam(req, "naming", "sequential")

	// Detect silence gaps
	args := []string{
		"-i", audioPath,
		"-af", fmt.Sprintf("silencedetect=noise=%ddB:d=%.2f", thresholdDB, minSilence),
		"-f", "null", "-",
	}
	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("silence detection failed: %w", err)), nil
	}

	regions := parseSilenceDetectOutput(stderr)
	totalDuration, _ := getAudioDuration(ctx, audioPath)

	// Compute non-silent segments
	type segment struct {
		Start float64
		End   float64
	}
	var segments []segment

	prevEnd := 0.0
	for _, r := range regions {
		if r.Start-prevEnd >= minSegment {
			segments = append(segments, segment{Start: prevEnd, End: r.Start})
		}
		prevEnd = r.End
	}
	// Final segment after last silence
	if totalDuration-prevEnd >= minSegment {
		segments = append(segments, segment{Start: prevEnd, End: totalDuration})
	}

	if len(segments) == 0 {
		return tools.TextResult("# Split by Silence\n\nNo segments found above minimum duration threshold.\n"), nil
	}

	// Convert segments to SampleDefinitions and run batch pipeline
	var defs []SampleDefinition
	for i, seg := range segments {
		var name string
		if naming == "timestamp" {
			name = fmt.Sprintf("seg-%s-%s", formatDuration(seg.Start), formatDuration(seg.End))
			name = strings.ReplaceAll(name, " ", "")
		} else {
			name = fmt.Sprintf("segment-%03d", i+1)
		}
		defs = append(defs, SampleDefinition{
			Name:  name,
			Start: fmt.Sprintf("%.3f", seg.Start),
			End:   fmt.Sprintf("%.3f", seg.End),
		})
	}

	results := runBatchPipeline(ctx, audioPath, outputDir, defs, process)

	// Generate Rekordbox XML if requested
	var xmlPath string
	if generateXML {
		packName := filepath.Base(outputDir)
		xmlPath = filepath.Join(outputDir, packName+".xml")
		var tracks []RekordboxTrack
		for i, res := range results {
			if res.Error != "" {
				continue
			}
			probe, _ := runFFprobe(ctx, res.Path)
			t := RekordboxTrack{
				ID: i + 1, Name: res.Name, Artist: packName, Album: packName,
				Genre: "DJ Sample", FilePath: res.Path, Duration: res.Duration,
			}
			if probe != nil {
				t.FileSize = probe.FileSize
				t.SampleRate = probe.SampleRate
				t.BitRate = probe.BitRate
			}
			tracks = append(tracks, t)
		}
		if len(tracks) > 0 {
			if err := GenerateRekordboxXML(tracks, packName, xmlPath); err != nil {
				xmlPath = fmt.Sprintf("(XML generation failed: %v)", err)
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("# Split by Silence\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputDir))
	sb.WriteString(fmt.Sprintf("**Silence gaps:** %d\n", len(regions)))
	sb.WriteString(fmt.Sprintf("**Segments:** %d\n\n", len(segments)))

	sb.WriteString("| # | Segment | Duration | Status |\n")
	sb.WriteString("|---|---------|----------|--------|\n")
	successCount := 0
	for i, res := range results {
		status := "OK"
		durStr := ""
		if res.Error != "" {
			status = "ERROR"
		} else {
			successCount++
			durStr = formatDuration(res.Duration)
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n", i+1, res.Name, durStr, status))
	}
	sb.WriteString(fmt.Sprintf("\n**Extracted:** %d/%d segments\n", successCount, len(segments)))
	if xmlPath != "" {
		sb.WriteString(fmt.Sprintf("**Rekordbox XML:** `%s`\n", xmlPath))
	}

	return tools.TextResult(sb.String()), nil
}

// handleScoreExtract extracts non-dialogue sections (film score/ambient music) from gaps in subtitles
func handleScoreExtract(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, errResult := tools.RequireStringParam(req, "source_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(sourcePath) {
		return tools.ErrorResult(fmt.Errorf("source file not found: %s", sourcePath)), nil
	}

	srtPath, errResult := tools.RequireStringParam(req, "srt_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(srtPath) {
		return tools.ErrorResult(fmt.Errorf("SRT file not found: %s", srtPath)), nil
	}

	packName, errResult := tools.RequireStringParam(req, "pack_name")
	if errResult != nil {
		return errResult, nil
	}

	minGap := tools.GetFloatParam(req, "min_gap", 5.0)
	maxDuration := tools.GetFloatParam(req, "max_duration", 60.0)
	padding := tools.GetFloatParam(req, "padding", 0.5)
	process := tools.GetBoolParam(req, "process", true)
	generateXML := tools.GetBoolParam(req, "generate_xml", true)
	artist := tools.GetStringParam(req, "artist")
	if artist == "" {
		artist = packName
	}

	// Parse subtitles
	entries, err := ParseSubtitle(srtPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse subtitle: %w", err)), nil
	}
	if len(entries) == 0 {
		return tools.ErrorResult(fmt.Errorf("no subtitle entries found in %s", srtPath)), nil
	}

	// Get total audio duration
	totalDuration, _ := getAudioDuration(ctx, sourcePath)
	if totalDuration <= 0 {
		// Try to estimate from last subtitle
		totalDuration = entries[len(entries)-1].EndTime + 30
	}

	// Find gaps between subtitle entries
	type gap struct {
		Start float64
		End   float64
	}
	var gaps []gap

	// Gap before first entry (opening score)
	if entries[0].StartTime >= minGap+padding {
		gaps = append(gaps, gap{Start: 0, End: entries[0].StartTime - padding})
	}

	// Gaps between consecutive entries
	for i := 1; i < len(entries); i++ {
		gapStart := entries[i-1].EndTime + padding
		gapEnd := entries[i].StartTime - padding
		if gapEnd-gapStart >= minGap {
			gaps = append(gaps, gap{Start: gapStart, End: gapEnd})
		}
	}

	// Gap after last entry (closing score)
	lastEnd := entries[len(entries)-1].EndTime + padding
	if totalDuration-lastEnd >= minGap {
		gaps = append(gaps, gap{Start: lastEnd, End: totalDuration})
	}

	if len(gaps) == 0 {
		return tools.TextResult(fmt.Sprintf("# Score Extract: %s\n\nNo gaps >= %.1fs found between dialogue entries.\nTry lowering min_gap.\n", packName, minGap)), nil
	}

	// Split long gaps into segments capped at maxDuration
	var segments []gap
	for _, g := range gaps {
		dur := g.End - g.Start
		if dur <= maxDuration {
			segments = append(segments, g)
		} else {
			pos := g.Start
			for pos+minGap <= g.End {
				segEnd := pos + maxDuration
				if segEnd > g.End {
					segEnd = g.End
				}
				if segEnd-pos >= minGap {
					segments = append(segments, gap{Start: pos, End: segEnd})
				}
				pos = segEnd
			}
		}
	}

	// Create pack directory
	samplesDir, err := defaultSamplesDir()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	packDir := filepath.Join(samplesDir, packName)
	if err := os.MkdirAll(packDir, 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create pack directory: %w", err)), nil
	}

	// Extract audio if video
	audioSource := sourcePath
	var tempAudioFile string
	if isVideoFile(ctx, sourcePath) {
		tempAudioFile = filepath.Join(packDir, ".full-audio.aiff")
		args := []string{"-y", "-i", sourcePath, "-vn", "-acodec", "pcm_s16be", "-ar", "44100", "-ac", "2", tempAudioFile}
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to extract audio: %w — %s", err, stderr)), nil
		}
		audioSource = tempAudioFile
	}

	// Convert segments to SampleDefinitions
	var defs []SampleDefinition
	for i, seg := range segments {
		mins := int(seg.Start) / 60
		secs := int(seg.Start) % 60
		name := fmt.Sprintf("score-%03d-%02dm%02ds", i+1, mins, secs)
		defs = append(defs, SampleDefinition{
			Name:        name,
			Start:       fmt.Sprintf("%.3f", seg.Start),
			End:         fmt.Sprintf("%.3f", seg.End),
			Description: fmt.Sprintf("Score segment at %s (%.1fs)", formatDuration(seg.Start), seg.End-seg.Start),
		})
	}

	// Run batch pipeline
	results := runBatchPipeline(ctx, audioSource, packDir, defs, process)

	// Generate Rekordbox XML
	var xmlPath string
	if generateXML {
		xmlPath = filepath.Join(packDir, packName+".xml")
		var tracks []RekordboxTrack
		for i, res := range results {
			if res.Error != "" {
				continue
			}
			probe, _ := runFFprobe(ctx, res.Path)
			t := RekordboxTrack{
				ID: i + 1, Name: res.Name, Artist: artist, Album: packName,
				Genre: "Film Score", FilePath: res.Path, Duration: res.Duration,
			}
			if probe != nil {
				t.FileSize = probe.FileSize
				t.SampleRate = probe.SampleRate
				t.BitRate = probe.BitRate
			}
			tracks = append(tracks, t)
		}
		if len(tracks) > 0 {
			if err := GenerateRekordboxXML(tracks, packName, xmlPath); err != nil {
				xmlPath = fmt.Sprintf("(XML generation failed: %v)", err)
			}
		}
	}

	// Clean up temp
	if tempAudioFile != "" {
		os.Remove(tempAudioFile)
	}

	// Build summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Score Extract: %s\n\n", packName))
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", sourcePath))
	sb.WriteString(fmt.Sprintf("**Subtitles:** `%s`\n", srtPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", packDir))
	sb.WriteString(fmt.Sprintf("**Min gap:** %.1fs | **Max duration:** %.1fs | **Padding:** %.1fs\n", minGap, maxDuration, padding))
	sb.WriteString(fmt.Sprintf("**Gaps found:** %d | **Segments:** %d\n\n", len(gaps), len(segments)))

	sb.WriteString("| # | Segment | Start | End | Duration | Status |\n")
	sb.WriteString("|---|---------|-------|-----|----------|--------|\n")
	successCount := 0
	totalScoreDur := 0.0
	for i, res := range results {
		status := "OK"
		durStr := ""
		if res.Error != "" {
			status = "ERROR"
		} else {
			successCount++
			durStr = formatDuration(res.Duration)
			totalScoreDur += res.Duration
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s |\n",
			i+1, res.Name,
			formatDuration(segments[i].Start), formatDuration(segments[i].End),
			durStr, status))
	}

	dialogueDur := 0.0
	for i := 0; i < len(entries); i++ {
		dialogueDur += entries[i].EndTime - entries[i].StartTime
	}

	sb.WriteString(fmt.Sprintf("\n**Extracted:** %d/%d segments\n", successCount, len(segments)))
	sb.WriteString(fmt.Sprintf("**Total score duration:** %s\n", formatDuration(totalScoreDur)))
	sb.WriteString(fmt.Sprintf("**Dialogue coverage:** %s (%.0f%% of total)\n", formatDuration(dialogueDur), dialogueDur/math.Max(totalDuration, 1)*100))
	if xmlPath != "" {
		sb.WriteString(fmt.Sprintf("**Rekordbox XML:** `%s`\n", xmlPath))
	}

	return tools.TextResult(sb.String()), nil
}
