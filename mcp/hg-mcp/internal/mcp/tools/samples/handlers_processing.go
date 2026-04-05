package samples

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleProcess processes an audio clip: trim silence, normalize, add fades
func handleProcess(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	inPlace := outputPath == ""
	if inPlace {
		outputPath = audioPath
	}

	trimSilence := tools.GetBoolParam(req, "trim_silence", true)
	silenceThreshold := tools.GetIntParam(req, "silence_threshold", -40)
	normalize := tools.GetBoolParam(req, "normalize", true)
	targetLUFS := tools.GetIntParam(req, "target_lufs", -14)
	fadeMs := tools.GetIntParam(req, "fade_ms", 10)

	// Get before duration
	beforeDuration, _ := getAudioDuration(ctx, audioPath)

	fadeSec := float64(fadeMs) / 1000.0

	// Build filter chain
	var filters []string
	if trimSilence {
		threshDB := fmt.Sprintf("%ddB", silenceThreshold)
		// Trim leading silence
		filters = append(filters, fmt.Sprintf("silenceremove=start_periods=1:start_silence=0.02:start_threshold=%s", threshDB))
		// Trim trailing silence (reverse, trim, reverse back)
		filters = append(filters, fmt.Sprintf("areverse,silenceremove=start_periods=1:start_silence=0.02:start_threshold=%s,areverse", threshDB))
	}
	if normalize {
		filters = append(filters, fmt.Sprintf("loudnorm=I=%d:TP=-1:LRA=11", targetLUFS))
	}
	// Fade-in
	if fadeMs > 0 {
		filters = append(filters, fmt.Sprintf("afade=t=in:d=%.3f", fadeSec))
	}

	// Use a temp file if in-place
	actualOutput := outputPath
	if inPlace {
		actualOutput = audioPath + ".tmp.aiff"
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(actualOutput), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	// Pass 1: apply trim + normalize + fade-in
	args := []string{"-y", "-i", audioPath}
	if len(filters) > 0 {
		args = append(args, "-af", strings.Join(filters, ","))
	}
	args = append(args, "-c:a", "pcm_s16be", actualOutput)

	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		// Clean up temp file on error
		if inPlace {
			os.Remove(actualOutput)
		}
		var sb strings.Builder
		sb.WriteString("# Process — Error\n\n")
		sb.WriteString(fmt.Sprintf("Failed to process `%s`\n\n", audioPath))
		sb.WriteString(fmt.Sprintf("**Error:** %v\n", err))
		if stderr != "" {
			sb.WriteString("\n**Details:**\n```\n")
			sb.WriteString(stderr)
			sb.WriteString("\n```\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	// Pass 2: apply fade-out using the actual duration from pass 1
	if fadeMs > 0 {
		pass1Duration, err := getAudioDuration(ctx, actualOutput)
		if err == nil && pass1Duration > fadeSec {
			fadeOutStart := pass1Duration - fadeSec
			pass2Output := actualOutput + ".pass2.aiff"
			args2 := []string{
				"-y", "-i", actualOutput,
				"-af", fmt.Sprintf("afade=t=out:st=%.3f:d=%.3f", fadeOutStart, fadeSec),
				"-c:a", "pcm_s16be",
				pass2Output,
			}
			_, _, err2 := runFFmpeg(ctx, args2...)
			if err2 == nil {
				os.Remove(actualOutput)
				os.Rename(pass2Output, actualOutput)
			} else {
				// If pass2 fails, keep pass1 output
				os.Remove(pass2Output)
			}
		}
	}

	// Move temp file to final destination if in-place
	if inPlace {
		os.Remove(outputPath)
		if err := os.Rename(actualOutput, outputPath); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to replace original file: %w", err)), nil
		}
	}

	// Get after duration
	afterDuration, _ := getAudioDuration(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Audio Processed\n\n")
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Trim Silence:** %v (threshold: %d dB)\n", trimSilence, silenceThreshold))
	sb.WriteString(fmt.Sprintf("**Normalize:** %v (target: %d LUFS)\n", normalize, targetLUFS))
	sb.WriteString(fmt.Sprintf("**Fades:** %d ms\n", fadeMs))
	if beforeDuration > 0 {
		sb.WriteString(fmt.Sprintf("**Before:** %s (%.1f sec)\n", formatDuration(beforeDuration), beforeDuration))
	}
	if afterDuration > 0 {
		sb.WriteString(fmt.Sprintf("**After:** %s (%.1f sec)\n", formatDuration(afterDuration), afterDuration))
	}

	return tools.TextResult(sb.String()), nil
}

// --- Sample Manipulation Tools ---

// handleReverse reverses an audio sample
func handleReverse(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
		outputPath = filepath.Join(filepath.Dir(audioPath), base+"-reversed.aiff")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	args := []string{"-y", "-i", audioPath, "-af", "areverse", "-c:a", "pcm_s16be", outputPath}
	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("reverse failed: %w — %s", err, stderr)), nil
	}

	duration, _ := getAudioDuration(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Reversed\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))

	return tools.TextResult(sb.String()), nil
}

// handlePitchShift shifts pitch by semitones while preserving duration
func handlePitchShift(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	if !tools.HasParam(req, "semitones") {
		return tools.ErrorResult(fmt.Errorf("semitones is required")), nil
	}
	semitones := tools.GetFloatParam(req, "semitones", 0)
	if semitones < -12 || semitones > 12 {
		return tools.ErrorResult(fmt.Errorf("semitones must be between -12 and +12")), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
		sign := "+"
		if semitones < 0 {
			sign = ""
		}
		outputPath = filepath.Join(filepath.Dir(audioPath), fmt.Sprintf("%s-pitch%s%g.aiff", base, sign, semitones))
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	// asetrate changes pitch, atempo compensates duration, aresample normalizes sample rate
	factor := math.Pow(2, semitones/12.0)
	filter := fmt.Sprintf("asetrate=44100*%.10f,atempo=%.10f,aresample=44100", factor, 1.0/factor)

	args := []string{"-y", "-i", audioPath, "-af", filter, "-c:a", "pcm_s16be", outputPath}
	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("pitch shift failed: %w — %s", err, stderr)), nil
	}

	duration, _ := getAudioDuration(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Pitch Shifted\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sign := "+"
	if semitones < 0 {
		sign = ""
	}
	sb.WriteString(fmt.Sprintf("**Shift:** %s%g semitones\n", sign, semitones))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))

	return tools.TextResult(sb.String()), nil
}

// handleTimeStretch changes speed without changing pitch
func handleTimeStretch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	if !tools.HasParam(req, "speed") {
		return tools.ErrorResult(fmt.Errorf("speed is required")), nil
	}
	speed := tools.GetFloatParam(req, "speed", 1.0)
	if speed < 0.25 || speed > 4.0 {
		return tools.ErrorResult(fmt.Errorf("speed must be between 0.25 and 4.0")), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
		outputPath = filepath.Join(filepath.Dir(audioPath), fmt.Sprintf("%s-%gx.aiff", base, speed))
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	// Chain atempo filters — each can only do 0.5 to 2.0
	filter := buildAtempoChain(speed)

	args := []string{"-y", "-i", audioPath, "-af", filter, "-c:a", "pcm_s16be", outputPath}
	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("time stretch failed: %w — %s", err, stderr)), nil
	}

	duration, _ := getAudioDuration(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Time Stretched\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Speed:** %gx\n", speed))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))

	return tools.TextResult(sb.String()), nil
}

// buildAtempoChain creates a chain of atempo filters for the given speed factor.
// Each atempo filter is limited to the 0.5–2.0 range.
func buildAtempoChain(speed float64) string {
	var parts []string
	remaining := speed
	for remaining > 2.0 {
		parts = append(parts, "atempo=2.0")
		remaining /= 2.0
	}
	for remaining < 0.5 {
		parts = append(parts, "atempo=0.5")
		remaining /= 0.5
	}
	parts = append(parts, fmt.Sprintf("atempo=%.6f", remaining))
	return strings.Join(parts, ",")
}

// handleLoop creates a looping version of a sample
func handleLoop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPath, errResult := tools.RequireStringParam(req, "audio_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPath) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPath)), nil
	}

	repeats := tools.GetIntParam(req, "repeats", 4)
	if repeats < 2 {
		repeats = 2
	}
	if repeats > 100 {
		repeats = 100
	}
	crossfadeMs := tools.GetIntParam(req, "crossfade_ms", 0)

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
		outputPath = filepath.Join(filepath.Dir(audioPath), base+"-loop.aiff")
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	if crossfadeMs <= 0 {
		// Simple loop: use stream_loop
		args := []string{"-y", "-stream_loop", strconv.Itoa(repeats - 1), "-i", audioPath, "-c:a", "pcm_s16be", outputPath}
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("loop failed: %w — %s", err, stderr)), nil
		}
	} else {
		// Crossfade loop: concat with acrossfade between copies
		crossfadeSec := float64(crossfadeMs) / 1000.0

		// Build complex filter: chain acrossfade between copies
		// For N repeats we need N inputs and N-1 crossfades
		var inputs []string
		for i := 0; i < repeats; i++ {
			inputs = append(inputs, "-i", audioPath)
		}

		// Build filter_complex
		var filterParts []string
		prevLabel := "[0:a]"
		for i := 1; i < repeats; i++ {
			outLabel := fmt.Sprintf("[xf%d]", i)
			filterParts = append(filterParts, fmt.Sprintf("%s[%d:a]acrossfade=d=%.3f%s", prevLabel, i, crossfadeSec, outLabel))
			prevLabel = outLabel
		}

		if len(filterParts) > 0 {
			// Last label doesn't get brackets for output
			lastPart := filterParts[len(filterParts)-1]
			lastLabel := fmt.Sprintf("[xf%d]", repeats-1)
			filterParts[len(filterParts)-1] = strings.TrimSuffix(lastPart, lastLabel)
		}

		args := append([]string{"-y"}, inputs...)
		args = append(args, "-filter_complex", strings.Join(filterParts, ";"), "-c:a", "pcm_s16be", outputPath)
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("crossfade loop failed: %w — %s", err, stderr)), nil
		}
	}

	duration, _ := getAudioDuration(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Looped\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** `%s`\n", audioPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Repeats:** %d\n", repeats))
	if crossfadeMs > 0 {
		sb.WriteString(fmt.Sprintf("**Crossfade:** %d ms\n", crossfadeMs))
	}
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))

	return tools.TextResult(sb.String()), nil
}

// handleConcat concatenates multiple audio files
func handleConcat(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pathsJSON, errResult := tools.RequireStringParam(req, "audio_paths")
	if errResult != nil {
		return errResult, nil
	}

	var audioPaths []string
	if err := json.Unmarshal([]byte(pathsJSON), &audioPaths); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid audio_paths JSON: %w", err)), nil
	}
	if len(audioPaths) < 2 {
		return tools.ErrorResult(fmt.Errorf("at least 2 audio files required")), nil
	}

	for _, p := range audioPaths {
		if !fileExists(p) {
			return tools.ErrorResult(fmt.Errorf("audio file not found: %s", p)), nil
		}
	}

	outputPath, errResult := tools.RequireStringParam(req, "output_path")
	if errResult != nil {
		return errResult, nil
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	crossfadeMs := tools.GetIntParam(req, "crossfade_ms", 0)
	normalize := tools.GetBoolParam(req, "normalize", false)

	if crossfadeMs <= 0 {
		// Simple concat using concat demuxer
		listFile := outputPath + ".concat.txt"
		var listContent strings.Builder
		for _, p := range audioPaths {
			absPath, _ := filepath.Abs(p)
			listContent.WriteString(fmt.Sprintf("file '%s'\n", absPath))
		}
		if err := os.WriteFile(listFile, []byte(listContent.String()), 0644); err != nil {
			return tools.ErrorResult(fmt.Errorf("cannot create concat list: %w", err)), nil
		}
		defer os.Remove(listFile)

		args := []string{"-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c:a", "pcm_s16be", outputPath}
		if normalize {
			args = []string{"-y", "-f", "concat", "-safe", "0", "-i", listFile, "-af", "loudnorm=I=-14:TP=-1:LRA=11", "-c:a", "pcm_s16be", outputPath}
		}
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("concat failed: %w — %s", err, stderr)), nil
		}
	} else {
		// Concat with crossfade using acrossfade filter
		crossfadeSec := float64(crossfadeMs) / 1000.0
		var inputs []string
		for _, p := range audioPaths {
			inputs = append(inputs, "-i", p)
		}

		var filterParts []string
		prevLabel := "[0:a]"
		for i := 1; i < len(audioPaths); i++ {
			outLabel := fmt.Sprintf("[xf%d]", i)
			filterParts = append(filterParts, fmt.Sprintf("%s[%d:a]acrossfade=d=%.3f%s", prevLabel, i, crossfadeSec, outLabel))
			prevLabel = outLabel
		}
		// Remove brackets from last output label
		if len(filterParts) > 0 {
			lastPart := filterParts[len(filterParts)-1]
			lastLabel := fmt.Sprintf("[xf%d]", len(audioPaths)-1)
			filterParts[len(filterParts)-1] = strings.TrimSuffix(lastPart, lastLabel)
		}

		args := append([]string{"-y"}, inputs...)
		args = append(args, "-filter_complex", strings.Join(filterParts, ";"), "-c:a", "pcm_s16be", outputPath)
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("crossfade concat failed: %w — %s", err, stderr)), nil
		}
	}

	duration, _ := getAudioDuration(ctx, outputPath)
	probe, _ := runFFprobe(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Concatenated\n\n")
	sb.WriteString(fmt.Sprintf("**Files:** %d inputs\n", len(audioPaths)))
	for i, p := range audioPaths {
		sb.WriteString(fmt.Sprintf("  %d. `%s`\n", i+1, filepath.Base(p)))
	}
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	if crossfadeMs > 0 {
		sb.WriteString(fmt.Sprintf("**Crossfade:** %d ms\n", crossfadeMs))
	}
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))
	if probe != nil {
		sb.WriteString(fmt.Sprintf("**File Size:** %s\n", formatFileSize(probe.FileSize)))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDialogueIsolate extracts dialogue via center-channel or bandpass filtering
func handleDialogueIsolate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, errResult := tools.RequireStringParam(req, "source_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(sourcePath) {
		return tools.ErrorResult(fmt.Errorf("source file not found: %s", sourcePath)), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath == "" {
		base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
		outputPath = filepath.Join(filepath.Dir(sourcePath), base+"-dialogue.aiff")
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	method := tools.OptionalStringParam(req, "method", "center")
	lowFreq := tools.GetIntParam(req, "low_freq", 300)
	highFreq := tools.GetIntParam(req, "high_freq", 3400)

	// Build filter based on method
	var filter string
	switch method {
	case "center":
		filter = "stereotools=mlev=1:slev=0"
	case "highpass":
		filter = fmt.Sprintf("highpass=f=%d,lowpass=f=%d", lowFreq, highFreq)
	case "both":
		filter = fmt.Sprintf("stereotools=mlev=1:slev=0,highpass=f=%d,lowpass=f=%d", lowFreq, highFreq)
	default:
		return tools.ErrorResult(fmt.Errorf("invalid method: %s (use: center, highpass, both)", method)), nil
	}

	// If source is video, extract audio first
	audioSource := sourcePath
	var tempFile string
	if isVideoFile(ctx, sourcePath) {
		tempFile = outputPath + ".temp.aiff"
		args := []string{"-y", "-i", sourcePath, "-vn", "-acodec", "pcm_s16be", "-ar", "44100", "-ac", "2", tempFile}
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to extract audio: %w — %s", err, stderr)), nil
		}
		audioSource = tempFile
	}

	args := []string{"-y", "-i", audioSource, "-af", filter, "-c:a", "pcm_s16be", outputPath}
	_, stderr, err := runFFmpeg(ctx, args...)
	if tempFile != "" {
		os.Remove(tempFile)
	}
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("dialogue isolation failed: %w — %s", err, stderr)), nil
	}

	duration, _ := getAudioDuration(ctx, outputPath)
	probe, _ := runFFprobe(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Dialogue Isolated\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", sourcePath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Method:** %s\n", method))
	switch method {
	case "center":
		sb.WriteString("**Filter:** Mid-channel extraction (removes panned content)\n")
	case "highpass":
		sb.WriteString(fmt.Sprintf("**Filter:** Vocal band %d–%d Hz\n", lowFreq, highFreq))
	case "both":
		sb.WriteString(fmt.Sprintf("**Filter:** Mid-channel + vocal band %d–%d Hz\n", lowFreq, highFreq))
	}
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))
	if probe != nil {
		sb.WriteString(fmt.Sprintf("**File Size:** %s\n", formatFileSize(probe.FileSize)))
	}
	sb.WriteString("\nUse this output as `source_path` for `imdb_scene_extract` or `srt_extract` for cleaner dialogue samples.\n")

	return tools.TextResult(sb.String()), nil
}

// handleMix crossfades two audio files together
func handleMix(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	audioPathA, errResult := tools.RequireStringParam(req, "audio_path_a")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPathA) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPathA)), nil
	}

	audioPathB, errResult := tools.RequireStringParam(req, "audio_path_b")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(audioPathB) {
		return tools.ErrorResult(fmt.Errorf("audio file not found: %s", audioPathB)), nil
	}

	outputPath, errResult := tools.RequireStringParam(req, "output_path")
	if errResult != nil {
		return errResult, nil
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	crossfadeMs := tools.GetIntParam(req, "crossfade_ms", 2000)
	crossfadeSec := float64(crossfadeMs) / 1000.0
	curve := tools.OptionalStringParam(req, "curve", "equal_power")

	// Map friendly curve names to ffmpeg curve names
	curveMap := map[string]string{
		"linear":      "tri",
		"equal_power": "exp",
		"logarithmic": "log",
		"tri":         "tri",
		"exp":         "exp",
		"log":         "log",
	}
	ffmpegCurve, ok := curveMap[curve]
	if !ok {
		ffmpegCurve = "exp"
	}

	filter := fmt.Sprintf("[0:a][1:a]acrossfade=d=%.3f:c1=%s:c2=%s", crossfadeSec, ffmpegCurve, ffmpegCurve)
	args := []string{"-y", "-i", audioPathA, "-i", audioPathB, "-filter_complex", filter, "-c:a", "pcm_s16be", outputPath}

	_, stderr, err := runFFmpeg(ctx, args...)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("mix failed: %w — %s", err, stderr)), nil
	}

	duration, _ := getAudioDuration(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Mixed\n\n")
	sb.WriteString(fmt.Sprintf("**Track A:** `%s`\n", filepath.Base(audioPathA)))
	sb.WriteString(fmt.Sprintf("**Track B:** `%s`\n", filepath.Base(audioPathB)))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Crossfade:** %d ms (%s)\n", crossfadeMs, curve))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))

	return tools.TextResult(sb.String()), nil
}
