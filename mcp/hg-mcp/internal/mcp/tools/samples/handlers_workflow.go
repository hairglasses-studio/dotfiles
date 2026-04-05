package samples

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// SampleDefinition represents a single sample in a batch operation
type SampleDefinition struct {
	Name        string `json:"name"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Description string `json:"description"`
	DisplayName string `json:"display_name,omitempty"`
}

// handleBatch orchestrates extracting + processing multiple samples from a source
func handleBatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, errResult := tools.RequireStringParam(req, "source_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(sourcePath) {
		return tools.ErrorResult(fmt.Errorf("source file not found: %s", sourcePath)), nil
	}

	packName, errResult := tools.RequireStringParam(req, "pack_name")
	if errResult != nil {
		return errResult, nil
	}

	samplesJSON, errResult := tools.RequireStringParam(req, "samples")
	if errResult != nil {
		return errResult, nil
	}

	var sampleDefs []SampleDefinition
	if err := json.Unmarshal([]byte(samplesJSON), &sampleDefs); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid samples JSON: %w", err)), nil
	}
	if len(sampleDefs) == 0 {
		return tools.ErrorResult(fmt.Errorf("samples array is empty")), nil
	}

	artist := tools.GetStringParam(req, "artist")
	if artist == "" {
		artist = packName
	}
	generateXML := tools.GetBoolParam(req, "generate_xml", true)

	// Create pack directory
	samplesDir, err := defaultSamplesDir()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	packDir := filepath.Join(samplesDir, packName)
	if err := os.MkdirAll(packDir, 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create pack directory: %w", err)), nil
	}

	// Step 1: If source is video, extract audio first
	audioSource := sourcePath
	var tempAudioFile string
	if isVideoFile(ctx, sourcePath) {
		tempAudioFile = filepath.Join(packDir, ".full-audio.aiff")
		args := []string{
			"-y", "-i", sourcePath,
			"-vn", "-acodec", "pcm_s16be",
			"-ar", "44100", "-ac", "2",
			tempAudioFile,
		}
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			var sb strings.Builder
			sb.WriteString("# Batch — Error\n\n")
			sb.WriteString("Failed to extract audio from video source.\n\n")
			sb.WriteString(fmt.Sprintf("**Error:** %v\n", err))
			if stderr != "" {
				sb.WriteString("\n**Details:**\n```\n")
				sb.WriteString(stderr)
				sb.WriteString("\n```\n")
			}
			return tools.TextResult(sb.String()), nil
		}
		audioSource = tempAudioFile
	}

	// Step 2: Clip and process each sample
	results := runBatchPipeline(ctx, audioSource, packDir, sampleDefs, true)

	// Step 3: Generate Rekordbox XML if requested
	var xmlPath string
	if generateXML {
		xmlPath = filepath.Join(packDir, packName+".xml")
		var tracks []RekordboxTrack
		for i, res := range results {
			if res.Error != "" {
				continue
			}
			probe, _ := runFFprobe(ctx, res.Path)
			trackName := res.Name
			if res.DisplayName != "" {
				trackName = res.DisplayName
			}
			t := RekordboxTrack{
				ID:       i + 1,
				Name:     trackName,
				Artist:   artist,
				Album:    packName,
				Genre:    "DJ Sample",
				FilePath: res.Path,
				Duration: res.Duration,
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

	// Step 4: Clean up temp audio
	if tempAudioFile != "" {
		os.Remove(tempAudioFile)
	}

	// Build result summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Sample Pack: %s\n\n", packName))
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", sourcePath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n\n", packDir))

	// Summary table
	sb.WriteString("| # | Sample | Duration | Status |\n")
	sb.WriteString("|---|--------|----------|--------|\n")

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

	sb.WriteString(fmt.Sprintf("\n**Extracted:** %d/%d samples\n", successCount, len(sampleDefs)))

	if xmlPath != "" {
		sb.WriteString(fmt.Sprintf("**Rekordbox XML:** `%s`\n", xmlPath))
	}

	// Show any errors
	for _, res := range results {
		if res.Error != "" {
			sb.WriteString(fmt.Sprintf("\n**Error** (%s): %s\n", res.Name, res.Error))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// batchResult holds the result of processing a single sample in a batch
type batchResult struct {
	Name        string
	DisplayName string
	Path        string
	Duration    float64
	Description string
	Error       string
}

// runBatchPipeline clips and processes samples from an audio source into a pack directory.
// If process is true, applies trim/normalize/fade to each clip.
func runBatchPipeline(ctx context.Context, audioSource, packDir string, defs []SampleDefinition, process bool) []batchResult {
	results := make([]batchResult, 0, len(defs))

	for _, sd := range defs {
		outputFile := filepath.Join(packDir, sd.Name+".aiff")
		res := batchResult{Name: sd.Name, DisplayName: sd.DisplayName, Path: outputFile, Description: sd.Description}

		// Clip
		clipArgs := []string{
			"-y", "-i", audioSource,
			"-ss", parseTimestamp(sd.Start),
			"-to", parseTimestamp(sd.End),
			"-c:a", "pcm_s16be",
			outputFile,
		}
		_, stderr, err := runFFmpeg(ctx, clipArgs...)
		if err != nil {
			res.Error = fmt.Sprintf("clip failed: %v — %s", err, stderr)
			results = append(results, res)
			continue
		}

		if !process {
			res.Duration, _ = getAudioDuration(ctx, outputFile)
			results = append(results, res)
			continue
		}

		// Process: trim + normalize + fade (two-pass)
		beforeDur, _ := getAudioDuration(ctx, outputFile)

		// Pass 1: silence trim + normalize + fade-in
		tmpProcessed := outputFile + ".proc.aiff"
		filterChain := "silenceremove=start_periods=1:start_silence=0.02:start_threshold=-40dB," +
			"areverse,silenceremove=start_periods=1:start_silence=0.02:start_threshold=-40dB,areverse," +
			"loudnorm=I=-14:TP=-1:LRA=11," +
			"afade=t=in:d=0.01"

		procArgs := []string{
			"-y", "-i", outputFile,
			"-af", filterChain,
			"-c:a", "pcm_s16be",
			tmpProcessed,
		}
		_, _, procErr := runFFmpeg(ctx, procArgs...)
		if procErr != nil {
			os.Remove(tmpProcessed)
			res.Duration = beforeDur
			results = append(results, res)
			continue
		}

		// Pass 2: fade-out using probed duration
		pass1Dur, _ := getAudioDuration(ctx, tmpProcessed)
		if pass1Dur > 0.01 {
			fadeOutStart := pass1Dur - 0.01
			pass2Output := outputFile + ".pass2.aiff"
			fade2Args := []string{
				"-y", "-i", tmpProcessed,
				"-af", fmt.Sprintf("afade=t=out:st=%.3f:d=0.01", fadeOutStart),
				"-c:a", "pcm_s16be",
				pass2Output,
			}
			_, _, fade2Err := runFFmpeg(ctx, fade2Args...)
			if fade2Err == nil {
				os.Remove(tmpProcessed)
				tmpProcessed = pass2Output
			} else {
				os.Remove(pass2Output)
			}
		}

		// Replace original clip with processed version
		os.Remove(outputFile)
		os.Rename(tmpProcessed, outputFile)
		os.Remove(outputFile + ".proc.aiff")
		os.Remove(outputFile + ".pass2.aiff")

		res.Duration, _ = getAudioDuration(ctx, outputFile)
		results = append(results, res)
	}

	return results
}

// handlePackNormalize normalizes loudness across all samples in a pack
func handlePackNormalize(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directory, errResult := tools.RequireStringParam(req, "directory")
	if errResult != nil {
		return errResult, nil
	}
	if !dirExists(directory) {
		return tools.ErrorResult(fmt.Errorf("directory not found: %s", directory)), nil
	}

	targetLUFS := tools.GetFloatParam(req, "target_lufs", -14.0)
	tolerance := tools.GetFloatParam(req, "tolerance", 1.0)
	dryRun := tools.GetBoolParam(req, "dry_run", true)

	entries, err := os.ReadDir(directory)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot read directory: %w", err)), nil
	}

	type fileResult struct {
		Name       string
		BeforeLUFS float64
		AfterLUFS  float64
		Status     string
	}
	var results []fileResult

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !isAudioFile(entry.Name()) {
			continue
		}

		filePath := filepath.Join(directory, entry.Name())
		measurement, err := measureLoudness(ctx, filePath)
		if err != nil {
			results = append(results, fileResult{Name: entry.Name(), Status: fmt.Sprintf("ERROR: %v", err)})
			continue
		}

		diff := measurement.IntegratedLUFS - targetLUFS
		if diff >= -tolerance && diff <= tolerance {
			results = append(results, fileResult{
				Name: entry.Name(), BeforeLUFS: measurement.IntegratedLUFS,
				AfterLUFS: measurement.IntegratedLUFS, Status: "OK (within tolerance)",
			})
			continue
		}

		if dryRun {
			results = append(results, fileResult{
				Name: entry.Name(), BeforeLUFS: measurement.IntegratedLUFS,
				AfterLUFS: targetLUFS, Status: "WOULD NORMALIZE",
			})
			continue
		}

		// Apply loudnorm
		tmpFile := filePath + ".norm.tmp"
		args := []string{"-y", "-i", filePath, "-af", fmt.Sprintf("loudnorm=I=%.1f:TP=-1:LRA=11", targetLUFS), "-c:a", "pcm_s16be", tmpFile}
		_, _, normErr := runFFmpeg(ctx, args...)
		if normErr != nil {
			os.Remove(tmpFile)
			results = append(results, fileResult{Name: entry.Name(), BeforeLUFS: measurement.IntegratedLUFS, Status: fmt.Sprintf("ERROR: %v", normErr)})
			continue
		}

		os.Remove(filePath)
		os.Rename(tmpFile, filePath)

		// Re-measure
		afterMeasurement, _ := measureLoudness(ctx, filePath)
		afterLUFS := targetLUFS
		if afterMeasurement != nil {
			afterLUFS = afterMeasurement.IntegratedLUFS
		}
		results = append(results, fileResult{Name: entry.Name(), BeforeLUFS: measurement.IntegratedLUFS, AfterLUFS: afterLUFS, Status: "NORMALIZED"})
	}

	var sb strings.Builder
	if dryRun {
		sb.WriteString("# Pack Normalize (Dry Run)\n\n")
	} else {
		sb.WriteString("# Pack Normalized\n\n")
	}
	sb.WriteString(fmt.Sprintf("**Directory:** `%s`\n", directory))
	sb.WriteString(fmt.Sprintf("**Target:** %.1f LUFS (±%.1f dB)\n\n", targetLUFS, tolerance))

	sb.WriteString("| File | Before | After | Status |\n")
	sb.WriteString("|------|--------|-------|--------|\n")
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("| %s | %.1f LUFS | %.1f LUFS | %s |\n", r.Name, r.BeforeLUFS, r.AfterLUFS, r.Status))
	}

	if dryRun {
		sb.WriteString("\n*Set dry_run=false to apply normalization.*\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handlePackExport exports a pack in multiple formats
func handlePackExport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directory, errResult := tools.RequireStringParam(req, "directory")
	if errResult != nil {
		return errResult, nil
	}
	if !dirExists(directory) {
		return tools.ErrorResult(fmt.Errorf("directory not found: %s", directory)), nil
	}

	formatsJSON, errResult := tools.RequireStringParam(req, "formats")
	if errResult != nil {
		return errResult, nil
	}
	var formats []string
	if err := json.Unmarshal([]byte(formatsJSON), &formats); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid formats JSON: %w", err)), nil
	}

	outputDir := tools.GetStringParam(req, "output_dir")
	if outputDir == "" {
		outputDir = filepath.Join(filepath.Dir(directory), filepath.Base(directory)+"-export")
	}

	mp3Bitrate := tools.GetIntParam(req, "mp3_bitrate", 320)
	generateXML := tools.GetBoolParam(req, "generate_xml", true)

	// Find audio files in source
	entries, err := os.ReadDir(directory)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot read directory: %w", err)), nil
	}

	var audioFiles []string
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if isAudioFile(e.Name()) {
			audioFiles = append(audioFiles, filepath.Join(directory, e.Name()))
		}
	}

	if len(audioFiles) == 0 {
		return tools.ErrorResult(fmt.Errorf("no audio files found in %s", directory)), nil
	}

	packName := filepath.Base(directory)

	type exportResult struct {
		Format  string
		Count   int
		Errors  int
		XMLPath string
	}
	var results []exportResult

	for _, targetFormat := range formats {
		targetFormat = strings.ToLower(strings.TrimSpace(targetFormat))
		formatDir := filepath.Join(outputDir, targetFormat)
		if err := os.MkdirAll(formatDir, 0755); err != nil {
			results = append(results, exportResult{Format: targetFormat, Errors: len(audioFiles)})
			continue
		}

		successCount, errorCount := 0, 0
		var tracks []RekordboxTrack

		for i, srcPath := range audioFiles {
			base := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
			ext := "." + targetFormat
			outPath := filepath.Join(formatDir, base+ext)

			var args []string
			switch targetFormat {
			case "wav":
				args = []string{"-y", "-i", srcPath, "-c:a", "pcm_s16le", outPath}
			case "mp3":
				args = []string{"-y", "-i", srcPath, "-c:a", "libmp3lame", "-b:a", fmt.Sprintf("%dk", mp3Bitrate), outPath}
			case "flac":
				args = []string{"-y", "-i", srcPath, "-c:a", "flac", outPath}
			case "aiff":
				args = []string{"-y", "-i", srcPath, "-c:a", "pcm_s16be", outPath}
			default:
				errorCount++
				continue
			}

			_, _, err := runFFmpeg(ctx, args...)
			if err != nil {
				errorCount++
				continue
			}
			successCount++

			if generateXML {
				probe, _ := runFFprobe(ctx, outPath)
				t := RekordboxTrack{
					ID: i + 1, Name: base, Artist: packName, Album: packName,
					Genre: "DJ Sample", FilePath: outPath,
				}
				if probe != nil {
					t.Duration = probe.Duration
					t.FileSize = probe.FileSize
					t.SampleRate = probe.SampleRate
					t.BitRate = probe.BitRate
				}
				tracks = append(tracks, t)
			}
		}

		var xmlPath string
		if generateXML && len(tracks) > 0 {
			xmlPath = filepath.Join(formatDir, packName+".xml")
			if err := GenerateRekordboxXML(tracks, packName, xmlPath); err != nil {
				xmlPath = ""
			}
		}

		results = append(results, exportResult{Format: targetFormat, Count: successCount, Errors: errorCount, XMLPath: xmlPath})
	}

	var sb strings.Builder
	sb.WriteString("# Pack Export\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", directory))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n\n", outputDir))

	sb.WriteString("| Format | Converted | Errors | XML |\n")
	sb.WriteString("|--------|-----------|--------|-----|\n")
	for _, r := range results {
		xmlStr := "—"
		if r.XMLPath != "" {
			xmlStr = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %s |\n", strings.ToUpper(r.Format), r.Count, r.Errors, xmlStr))
	}

	return tools.TextResult(sb.String()), nil
}

// movieSource represents one movie entry in a multi-movie pack
type movieSource struct {
	IMDBID     string `json:"imdb_id"`
	SourcePath string `json:"source_path"`
	SRTPath    string `json:"srt_path"`
	Label      string `json:"label"`
}

// handleMultiMoviePack builds a themed sample pack from multiple movies in one call
func handleMultiMoviePack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	packName, errResult := tools.RequireStringParam(req, "pack_name")
	if errResult != nil {
		return errResult, nil
	}

	sourcesJSON, errResult := tools.RequireStringParam(req, "sources")
	if errResult != nil {
		return errResult, nil
	}

	var sources []movieSource
	if err := json.Unmarshal([]byte(sourcesJSON), &sources); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid sources JSON: %w", err)), nil
	}
	if len(sources) == 0 {
		return tools.ErrorResult(fmt.Errorf("sources array is empty")), nil
	}

	// Validate all sources exist
	for i, src := range sources {
		if src.IMDBID == "" {
			return tools.ErrorResult(fmt.Errorf("source[%d]: imdb_id is required", i)), nil
		}
		if src.SourcePath == "" {
			return tools.ErrorResult(fmt.Errorf("source[%d]: source_path is required", i)), nil
		}
		if !fileExists(src.SourcePath) {
			return tools.ErrorResult(fmt.Errorf("source[%d]: file not found: %s", i, src.SourcePath)), nil
		}
		if src.SRTPath == "" {
			return tools.ErrorResult(fmt.Errorf("source[%d]: srt_path is required", i)), nil
		}
		if !fileExists(src.SRTPath) {
			return tools.ErrorResult(fmt.Errorf("source[%d]: SRT not found: %s", i, src.SRTPath)), nil
		}
		if src.Label == "" {
			sources[i].Label = src.IMDBID
		}
	}

	matchThreshold := tools.GetFloatParam(req, "match_threshold", 0.3)
	maxScenesPerMovie := tools.GetIntParam(req, "max_scenes_per_movie", 30)
	minSceneDuration := tools.GetFloatParam(req, "min_scene_duration", 15.0)
	maxSceneDuration := tools.GetFloatParam(req, "max_scene_duration", 60.0)
	process := tools.GetBoolParam(req, "process", true)
	generateXML := tools.GetBoolParam(req, "generate_xml", true)
	targetSampleRate := tools.GetIntParam(req, "target_sample_rate", 0)
	artist := tools.GetStringParam(req, "artist")
	if artist == "" {
		artist = packName
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

	// Process each movie
	type movieResult struct {
		Label       string
		ScenesFound int
		Extracted   int
		Errors      int
		TotalDur    float64
	}
	var movieResults []movieResult
	var allTracks []RekordboxTrack
	trackID := 1

	for _, src := range sources {
		labelSlug := slugify(src.Label)
		movieDir := filepath.Join(packDir, labelSlug)
		if err := os.MkdirAll(movieDir, 0755); err != nil {
			movieResults = append(movieResults, movieResult{Label: src.Label, Errors: 1})
			continue
		}

		// Fetch quotes
		quotes, err := fetchIMDBQuotes(src.IMDBID)
		if err != nil || len(quotes) == 0 {
			movieResults = append(movieResults, movieResult{Label: src.Label})
			continue
		}

		// Parse SRT
		entries, err := ParseSubtitle(src.SRTPath)
		if err != nil || len(entries) == 0 {
			movieResults = append(movieResults, movieResult{Label: src.Label})
			continue
		}

		// Match
		matches := matchQuotesAgainstSRT(quotes, entries, matchThreshold)
		for i := range matches {
			expandSceneWindow(&matches[i], entries, minSceneDuration, maxSceneDuration)
		}
		matches = mergeOverlappingScenes(matches)
		if len(matches) > maxScenesPerMovie {
			matches = matches[:maxScenesPerMovie]
		}

		if len(matches) == 0 {
			movieResults = append(movieResults, movieResult{Label: src.Label})
			continue
		}

		// Extract audio from video if needed
		audioSource := src.SourcePath
		var tempAudio string
		if isVideoFile(ctx, src.SourcePath) {
			tempAudio = filepath.Join(movieDir, ".full-audio.aiff")
			args := []string{"-y", "-i", src.SourcePath, "-vn", "-acodec", "pcm_s16be", "-ar", "44100", "-ac", "2", tempAudio}
			if _, _, err := runFFmpeg(ctx, args...); err != nil {
				movieResults = append(movieResults, movieResult{Label: src.Label, ScenesFound: len(matches), Errors: len(matches)})
				continue
			}
			audioSource = tempAudio
		}

		// Build sample definitions
		var defs []SampleDefinition
		for i, m := range matches {
			slugText := ""
			if len(m.Quote.Lines) > 0 {
				first := m.Quote.Lines[0]
				if first.Character != "" {
					slugText = first.Character + " " + first.Text
				} else {
					slugText = first.Text
				}
			}
			if slugText == "" {
				slugText = m.MatchedText
			}
			name := fmt.Sprintf("%03d-%s", i+1, slugify(slugText))

			var displayParts []string
			for _, l := range m.Quote.Lines {
				if l.Character != "" {
					displayParts = append(displayParts, fmt.Sprintf("%s: %s", l.Character, l.Text))
				} else {
					displayParts = append(displayParts, l.Text)
				}
			}

			defs = append(defs, SampleDefinition{
				Name:        name,
				DisplayName: strings.Join(displayParts, " / "),
				Start:       fmt.Sprintf("%.3f", m.StartTime),
				End:         fmt.Sprintf("%.3f", m.EndTime),
			})
		}

		results := runBatchPipeline(ctx, audioSource, movieDir, defs, process)

		// CDJ conversion
		if targetSampleRate > 0 {
			cdjDir := filepath.Join(movieDir, "cdj")
			os.MkdirAll(cdjDir, 0755)
			for i, res := range results {
				if res.Error != "" {
					continue
				}
				probe, err := runFFprobe(ctx, res.Path)
				if err != nil || probe.SampleRate <= targetSampleRate {
					continue
				}
				outPath := filepath.Join(cdjDir, filepath.Base(res.Path))
				args := []string{"-y", "-i", res.Path, "-ar", strconv.Itoa(targetSampleRate), "-c:a", "pcm_s16be", outPath}
				if _, _, err := runFFmpeg(ctx, args...); err == nil {
					results[i].Path = outPath
				}
			}
		}

		// Collect tracks for combined XML
		mr := movieResult{Label: src.Label, ScenesFound: len(matches)}
		for _, res := range results {
			if res.Error != "" {
				mr.Errors++
				continue
			}
			mr.Extracted++
			mr.TotalDur += res.Duration

			if generateXML {
				probe, _ := runFFprobe(ctx, res.Path)
				trackName := res.Name
				if res.DisplayName != "" {
					trackName = res.DisplayName
				}
				t := RekordboxTrack{
					ID: trackID, Name: trackName, Artist: src.Label, Album: packName,
					Genre: "DJ Sample", FilePath: res.Path, Duration: res.Duration,
				}
				if probe != nil {
					t.FileSize = probe.FileSize
					t.SampleRate = probe.SampleRate
					t.BitRate = probe.BitRate
				}
				allTracks = append(allTracks, t)
				trackID++
			}
		}
		movieResults = append(movieResults, mr)

		if tempAudio != "" {
			os.Remove(tempAudio)
		}
	}

	// Generate combined Rekordbox XML
	var xmlPath string
	if generateXML && len(allTracks) > 0 {
		xmlPath = filepath.Join(packDir, packName+".xml")
		if err := GenerateRekordboxXML(allTracks, packName, xmlPath); err != nil {
			xmlPath = fmt.Sprintf("(XML generation failed: %v)", err)
		}
	}

	// Build summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Multi-Movie Pack: %s\n\n", packName))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", packDir))
	sb.WriteString(fmt.Sprintf("**Movies:** %d\n\n", len(sources)))

	sb.WriteString("| Movie | Scenes Found | Extracted | Errors | Duration |\n")
	sb.WriteString("|-------|-------------|-----------|--------|----------|\n")
	totalExtracted := 0
	totalErrors := 0
	totalDuration := 0.0
	for _, mr := range movieResults {
		sb.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %s |\n",
			mr.Label, mr.ScenesFound, mr.Extracted, mr.Errors, formatDuration(mr.TotalDur)))
		totalExtracted += mr.Extracted
		totalErrors += mr.Errors
		totalDuration += mr.TotalDur
	}

	sb.WriteString(fmt.Sprintf("\n**Total extracted:** %d scenes\n", totalExtracted))
	sb.WriteString(fmt.Sprintf("**Total duration:** %s\n", formatDuration(totalDuration)))
	if totalErrors > 0 {
		sb.WriteString(fmt.Sprintf("**Errors:** %d\n", totalErrors))
	}
	if xmlPath != "" {
		sb.WriteString(fmt.Sprintf("**Rekordbox XML:** `%s`\n", xmlPath))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSceneAssemble assembles extracted clips into a continuous mix with crossfades and cue points
func handleSceneAssemble(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directory, errResult := tools.RequireStringParam(req, "directory")
	if errResult != nil {
		return errResult, nil
	}
	if !dirExists(directory) {
		return tools.ErrorResult(fmt.Errorf("directory not found: %s", directory)), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath == "" {
		outputPath = filepath.Join(directory, filepath.Base(directory)+"-assembled.aiff")
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	order := tools.OptionalStringParam(req, "order", "chronological")
	customOrderJSON := tools.GetStringParam(req, "custom_order")
	crossfadeMs := tools.GetIntParam(req, "crossfade_ms", 2000)
	gapMs := tools.GetIntParam(req, "gap_ms", 0)
	normalize := tools.GetBoolParam(req, "normalize", false)
	generateXML := tools.GetBoolParam(req, "generate_xml", true)

	// Scan directory for audio files
	dirEntries, err := os.ReadDir(directory)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot read directory: %w", err)), nil
	}

	var audioFiles []string
	for _, e := range dirEntries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if isAudioFile(e.Name()) {
			audioFiles = append(audioFiles, filepath.Join(directory, e.Name()))
		}
	}

	if len(audioFiles) < 2 {
		return tools.ErrorResult(fmt.Errorf("need at least 2 audio files, found %d", len(audioFiles))), nil
	}

	// Order files
	switch order {
	case "chronological":
		sort.Strings(audioFiles) // alphabetical = chronological for numbered files
	case "random":
		rand.Shuffle(len(audioFiles), func(i, j int) { audioFiles[i], audioFiles[j] = audioFiles[j], audioFiles[i] })
	case "custom":
		if customOrderJSON == "" {
			return tools.ErrorResult(fmt.Errorf("custom_order JSON array required when order=custom")), nil
		}
		var customOrder []string
		if err := json.Unmarshal([]byte(customOrderJSON), &customOrder); err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid custom_order JSON: %w", err)), nil
		}
		// Resolve custom order to full paths
		fileMap := make(map[string]string)
		for _, f := range audioFiles {
			fileMap[filepath.Base(f)] = f
			fileMap[strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))] = f
		}
		var ordered []string
		for _, name := range customOrder {
			if path, ok := fileMap[name]; ok {
				ordered = append(ordered, path)
			} else if fileExists(name) {
				ordered = append(ordered, name)
			}
		}
		if len(ordered) < 2 {
			return tools.ErrorResult(fmt.Errorf("custom_order resolved to %d files, need at least 2", len(ordered))), nil
		}
		audioFiles = ordered
	default:
		return tools.ErrorResult(fmt.Errorf("invalid order: %s (use: chronological, random, custom)", order)), nil
	}

	// Get durations for all clips (needed for cue points)
	type clipInfo struct {
		Path     string
		Duration float64
	}
	var clips []clipInfo
	for _, f := range audioFiles {
		dur, _ := getAudioDuration(ctx, f)
		clips = append(clips, clipInfo{Path: f, Duration: dur})
	}

	// Build the assembled output
	if crossfadeMs > 0 && gapMs == 0 {
		// Use acrossfade filter chain
		crossfadeSec := float64(crossfadeMs) / 1000.0
		var inputs []string
		for _, f := range audioFiles {
			inputs = append(inputs, "-i", f)
		}

		var filterParts []string
		prevLabel := "[0:a]"
		for i := 1; i < len(audioFiles); i++ {
			outLabel := fmt.Sprintf("[xf%d]", i)
			filterParts = append(filterParts, fmt.Sprintf("%s[%d:a]acrossfade=d=%.3f%s", prevLabel, i, crossfadeSec, outLabel))
			prevLabel = outLabel
		}
		// Remove brackets from last output
		if len(filterParts) > 0 {
			lastPart := filterParts[len(filterParts)-1]
			lastLabel := fmt.Sprintf("[xf%d]", len(audioFiles)-1)
			filterParts[len(filterParts)-1] = strings.TrimSuffix(lastPart, lastLabel)
		}

		filter := strings.Join(filterParts, ";")
		if normalize {
			filter += ",loudnorm=I=-14:TP=-1:LRA=11"
		}

		args := append([]string{"-y"}, inputs...)
		args = append(args, "-filter_complex", filter, "-c:a", "pcm_s16be", outputPath)
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("assembly failed: %w — %s", err, stderr)), nil
		}
	} else {
		// Use concat demuxer with optional silence gaps
		listFile := outputPath + ".concat.txt"
		var listContent strings.Builder
		gapSec := float64(gapMs) / 1000.0

		for i, f := range audioFiles {
			absPath, _ := filepath.Abs(f)
			listContent.WriteString(fmt.Sprintf("file '%s'\n", absPath))
			if gapMs > 0 && i < len(audioFiles)-1 {
				// Generate silence file
				silenceFile := outputPath + fmt.Sprintf(".silence-%d.aiff", i)
				silArgs := []string{"-y", "-f", "lavfi", "-i", fmt.Sprintf("anullsrc=r=44100:cl=stereo:d=%.3f", gapSec), "-c:a", "pcm_s16be", silenceFile}
				runFFmpeg(ctx, silArgs...)
				listContent.WriteString(fmt.Sprintf("file '%s'\n", silenceFile))
				defer os.Remove(silenceFile)
			}
		}

		if err := os.WriteFile(listFile, []byte(listContent.String()), 0644); err != nil {
			return tools.ErrorResult(fmt.Errorf("cannot create concat list: %w", err)), nil
		}
		defer os.Remove(listFile)

		args := []string{"-y", "-f", "concat", "-safe", "0", "-i", listFile}
		if normalize {
			args = append(args, "-af", "loudnorm=I=-14:TP=-1:LRA=11")
		}
		args = append(args, "-c:a", "pcm_s16be", outputPath)
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("assembly failed: %w — %s", err, stderr)), nil
		}
	}

	// Calculate scene boundary timestamps
	crossfadeSec := float64(crossfadeMs) / 1000.0
	gapSec := float64(gapMs) / 1000.0
	var boundaries []float64
	position := 0.0
	boundaries = append(boundaries, 0)
	for i, clip := range clips {
		if i < len(clips)-1 {
			if crossfadeMs > 0 && gapMs == 0 {
				position += clip.Duration - crossfadeSec
			} else {
				position += clip.Duration + gapSec
			}
			boundaries = append(boundaries, position)
		}
	}

	// Generate Rekordbox XML with cue points
	var xmlPath string
	if generateXML {
		totalDur, _ := getAudioDuration(ctx, outputPath)
		probe, _ := runFFprobe(ctx, outputPath)

		var cues []CuePoint
		colors := [][3]int{{40, 226, 20}, {230, 20, 20}, {0, 100, 255}, {255, 127, 0}, {200, 0, 200}}
		for i, pos := range boundaries {
			color := colors[i%len(colors)]
			cues = append(cues, CuePoint{
				Num:   i,
				Name:  fmt.Sprintf("Scene %d", i+1),
				Start: pos,
				Type:  0, // hot cue
				Red:   color[0],
				Green: color[1],
				Blue:  color[2],
			})
		}

		t := RekordboxTrack{
			ID: 1, Name: filepath.Base(outputPath), Artist: filepath.Base(directory),
			Album: filepath.Base(directory), Genre: "DJ Sample Mix",
			FilePath: outputPath, Duration: totalDur, CuePoints: cues,
		}
		if probe != nil {
			t.FileSize = probe.FileSize
			t.SampleRate = probe.SampleRate
			t.BitRate = probe.BitRate
		}

		xmlPath = filepath.Join(filepath.Dir(outputPath), strings.TrimSuffix(filepath.Base(outputPath), filepath.Ext(outputPath))+".xml")
		if err := GenerateRekordboxXML([]RekordboxTrack{t}, filepath.Base(directory)+"-assembled", xmlPath); err != nil {
			xmlPath = fmt.Sprintf("(XML generation failed: %v)", err)
		}
	}

	totalDur, _ := getAudioDuration(ctx, outputPath)
	probe, _ := runFFprobe(ctx, outputPath)

	var sb strings.Builder
	sb.WriteString("# Scene Assembly\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", directory))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", outputPath))
	sb.WriteString(fmt.Sprintf("**Clips:** %d\n", len(audioFiles)))
	sb.WriteString(fmt.Sprintf("**Order:** %s\n", order))
	if crossfadeMs > 0 {
		sb.WriteString(fmt.Sprintf("**Crossfade:** %d ms\n", crossfadeMs))
	}
	if gapMs > 0 {
		sb.WriteString(fmt.Sprintf("**Gap:** %d ms\n", gapMs))
	}
	sb.WriteString(fmt.Sprintf("**Total duration:** %s\n", formatDuration(totalDur)))
	if probe != nil {
		sb.WriteString(fmt.Sprintf("**File Size:** %s\n", formatFileSize(probe.FileSize)))
	}

	// Scene boundary table
	sb.WriteString("\n## Scene Boundaries\n\n")
	sb.WriteString("| # | Clip | Cue Point |\n")
	sb.WriteString("|---|------|-----------|\n")
	for i, f := range audioFiles {
		cueStr := formatDuration(boundaries[i])
		sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", i+1, filepath.Base(f), cueStr))
	}

	if xmlPath != "" {
		sb.WriteString(fmt.Sprintf("\n**Rekordbox XML:** `%s`\n", xmlPath))
	}

	return tools.TextResult(sb.String()), nil
}
