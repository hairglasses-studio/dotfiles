package samples

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/mcpkit/sanitize"
)

// --- SRT Subtitle Tools ---

// handleSRTSearch searches SRT subtitles for dialogue matching a query
func handleSRTSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	srtPath, errResult := tools.RequireStringParam(req, "srt_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(srtPath) {
		return tools.ErrorResult(fmt.Errorf("SRT file not found: %s", srtPath)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	useRegex := tools.GetBoolParam(req, "regex", false)
	contextLines := tools.GetIntParam(req, "context_lines", 1)

	entries, err := ParseSubtitle(srtPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse subtitle: %w", err)), nil
	}

	// Build matcher
	var matcher func(string) bool
	if useRegex {
		re, err := regexp.Compile("(?i)" + query)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid regex: %w", err)), nil
		}
		matcher = func(s string) bool { return re.MatchString(s) }
	} else {
		queryLower := strings.ToLower(query)
		matcher = func(s string) bool { return strings.Contains(strings.ToLower(s), queryLower) }
	}

	// Find matches with context
	type match struct {
		entry   SRTEntry
		context []SRTEntry
	}

	matchIndices := make(map[int]bool)
	for i, e := range entries {
		if matcher(e.Text) {
			matchIndices[i] = true
		}
	}

	var matches []match
	seen := make(map[int]bool)
	for i := range entries {
		if !matchIndices[i] {
			continue
		}
		if seen[i] {
			continue
		}
		seen[i] = true
		m := match{entry: entries[i]}
		// Gather context
		start := i - contextLines
		if start < 0 {
			start = 0
		}
		end := i + contextLines
		if end >= len(entries) {
			end = len(entries) - 1
		}
		for j := start; j <= end; j++ {
			if j != i {
				m.context = append(m.context, entries[j])
			}
		}
		matches = append(matches, m)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# SRT Search: \"%s\"\n\n", query))
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", srtPath))
	sb.WriteString(fmt.Sprintf("**Matches:** %d\n\n", len(matches)))

	if len(matches) == 0 {
		sb.WriteString("No matches found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| # | Start | End | Text |\n")
	sb.WriteString("|---|-------|-----|------|\n")
	for _, m := range matches {
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | **%s** |\n", m.entry.Index, m.entry.StartTS, m.entry.EndTS, m.entry.Text))
		for _, c := range m.context {
			sb.WriteString(fmt.Sprintf("| | %s | %s | %s |\n", c.StartTS, c.EndTS, c.Text))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleSRTExtract searches subtitles and auto-extracts matching dialogue as audio samples
func handleSRTExtract(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	packName, errResult := tools.RequireStringParam(req, "pack_name")
	if errResult != nil {
		return errResult, nil
	}

	useRegex := tools.GetBoolParam(req, "regex", false)
	paddingMs := tools.GetIntParam(req, "padding_ms", 200)
	process := tools.GetBoolParam(req, "process", true)
	generateXML := tools.GetBoolParam(req, "generate_xml", true)
	artist := tools.GetStringParam(req, "artist")
	if artist == "" {
		artist = packName
	}
	maxResults := tools.GetIntParam(req, "max_results", 50)

	// Parse subtitle file
	entries, err := ParseSubtitle(srtPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse subtitle: %w", err)), nil
	}

	// Build matcher
	var matcher func(string) bool
	if useRegex {
		re, err := regexp.Compile("(?i)" + query)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid regex: %w", err)), nil
		}
		matcher = func(s string) bool { return re.MatchString(s) }
	} else {
		queryLower := strings.ToLower(query)
		matcher = func(s string) bool { return strings.Contains(strings.ToLower(s), queryLower) }
	}

	// Filter entries
	var matched []SRTEntry
	for _, e := range entries {
		if matcher(e.Text) {
			matched = append(matched, e)
			if len(matched) >= maxResults {
				break
			}
		}
	}

	if len(matched) == 0 {
		return tools.TextResult(fmt.Sprintf("# SRT Extract\n\nNo subtitles matching \"%s\" found in `%s`.\n", query, srtPath)), nil
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

	// If source is video, extract audio first
	audioSource := sourcePath
	var tempAudioFile string
	if isVideoFile(ctx, sourcePath) {
		tempAudioFile = filepath.Join(packDir, ".full-audio.aiff")
		args := []string{"-y", "-i", sourcePath, "-vn", "-acodec", "pcm_s16be", "-ar", "44100", "-ac", "2", tempAudioFile}
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.TextResult(fmt.Sprintf("# SRT Extract — Error\n\nFailed to extract audio from video.\n\n**Error:** %v\n```\n%s\n```\n", err, stderr)), nil
		}
		audioSource = tempAudioFile
	}

	// Convert matched SRT entries to SampleDefinitions with padding
	paddingSec := float64(paddingMs) / 1000.0
	var defs []SampleDefinition
	for i, e := range matched {
		startSec := e.StartTime - paddingSec
		if startSec < 0 {
			startSec = 0
		}
		endSec := e.EndTime + paddingSec

		name := fmt.Sprintf("%03d-%s", i+1, slugify(e.Text))
		if len(name) > 80 {
			name = name[:80]
		}

		defs = append(defs, SampleDefinition{
			Name:        name,
			Start:       fmt.Sprintf("%.3f", startSec),
			End:         fmt.Sprintf("%.3f", endSec),
			Description: e.Text,
		})
	}

	// Run batch pipeline
	results := runBatchPipeline(ctx, audioSource, packDir, defs, process)

	// Generate Rekordbox XML if requested
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

	// Clean up temp audio
	if tempAudioFile != "" {
		os.Remove(tempAudioFile)
	}

	// Build result summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# SRT Extract: %s\n\n", packName))
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", sourcePath))
	sb.WriteString(fmt.Sprintf("**Subtitles:** `%s`\n", srtPath))
	sb.WriteString(fmt.Sprintf("**Query:** \"%s\"\n", query))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n\n", packDir))

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

	sb.WriteString(fmt.Sprintf("\n**Extracted:** %d/%d samples\n", successCount, len(defs)))
	if xmlPath != "" {
		sb.WriteString(fmt.Sprintf("**Rekordbox XML:** `%s`\n", xmlPath))
	}

	for _, res := range results {
		if res.Error != "" {
			sb.WriteString(fmt.Sprintf("\n**Error** (%s): %s\n", res.Name, res.Error))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleSRTList finds subtitle files alongside a video or in a directory
func handleSRTList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("path not found: %s", path)), nil
	}

	recursive := tools.GetBoolParam(req, "recursive", false)

	var files []string
	if info.IsDir() && recursive {
		files = findSubtitleFilesRecursive(path)
	} else {
		files = findSubtitleFiles(path)
	}

	var sb strings.Builder
	sb.WriteString("# Subtitle Files\n\n")
	sb.WriteString(fmt.Sprintf("**Path:** `%s`\n\n", path))

	if len(files) == 0 {
		sb.WriteString("No subtitle files found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| File | Format | Lines | Duration |\n")
	sb.WriteString("|------|--------|-------|----------|\n")

	for _, f := range files {
		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(f), "."))
		lines := 0
		durStr := "—"
		entries, err := ParseSubtitle(f)
		if err == nil {
			lines = len(entries)
			if len(entries) > 0 {
				last := entries[len(entries)-1]
				durStr = formatDuration(last.EndTime)
			}
		}
		relPath := f
		if info.IsDir() {
			if rel, err := filepath.Rel(path, f); err == nil {
				relPath = rel
			}
		} else {
			relPath = filepath.Base(f)
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n", relPath, ext, lines, durStr))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSRTDump displays SRT contents as a readable timeline
func handleSRTDump(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	srtPath, errResult := tools.RequireStringParam(req, "srt_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(srtPath) {
		return tools.ErrorResult(fmt.Errorf("SRT file not found: %s", srtPath)), nil
	}

	entries, err := ParseSubtitle(srtPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse subtitle: %w", err)), nil
	}

	// Parse optional time range filters
	var startFilter, endFilter float64
	hasStartFilter := false
	hasEndFilter := false
	if s := tools.GetStringParam(req, "start"); s != "" {
		if v, err := parseTimestampToSeconds(s); err == nil {
			startFilter = v
			hasStartFilter = true
		}
	}
	if s := tools.GetStringParam(req, "end"); s != "" {
		if v, err := parseTimestampToSeconds(s); err == nil {
			endFilter = v
			hasEndFilter = true
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# SRT Timeline: %s\n\n", filepath.Base(srtPath)))
	sb.WriteString(fmt.Sprintf("**Total entries:** %d\n\n", len(entries)))

	count := 0
	for _, e := range entries {
		if hasStartFilter && e.EndTime < startFilter {
			continue
		}
		if hasEndFilter && e.StartTime > endFilter {
			continue
		}
		// Format as HH:MM:SS for display
		startDisp := secondsToSRTTime(e.StartTime)
		endDisp := secondsToSRTTime(e.EndTime)
		// Trim to HH:MM:SS (drop milliseconds for readability)
		startDisp = startDisp[:8]
		endDisp = endDisp[:8]
		sb.WriteString(fmt.Sprintf("[%s → %s] %s\n", startDisp, endDisp, e.Text))
		count++
	}

	if count == 0 {
		sb.WriteString("No entries in the specified time range.\n")
	} else {
		sb.WriteString(fmt.Sprintf("\n**Displayed:** %d entries\n", count))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSRTGenerate generates SRT subtitles from a video/audio file using Whisper
func handleSRTGenerate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, errResult := tools.RequireStringParam(req, "source_path")
	if errResult != nil {
		return errResult, nil
	}
	if err := sanitize.MediaPath(sourcePath); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid source_path: %w", err)), nil
	}
	if !fileExists(sourcePath) {
		return tools.ErrorResult(fmt.Errorf("source file not found: %s", sourcePath)), nil
	}

	// Check whisper is installed
	if _, err := exec.LookPath("whisper"); err != nil {
		return tools.ErrorResult(fmt.Errorf("whisper not found in PATH — install with: pip install openai-whisper")), nil
	}

	outputPath := tools.GetStringParam(req, "output_path")
	if outputPath != "" {
		if err := sanitize.MediaPath(outputPath); err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid output_path: %w", err)), nil
		}
	}
	language := tools.GetStringParam(req, "language")
	model := tools.OptionalStringParam(req, "model", "base")

	// Determine output directory
	outputDir := filepath.Dir(sourcePath)
	if outputPath != "" {
		outputDir = filepath.Dir(outputPath)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create output directory: %w", err)), nil
	}

	// If source is video, extract audio to temp AIFF first
	audioSource := sourcePath
	var tempFile string
	if isVideoFile(ctx, sourcePath) {
		tempFile = filepath.Join(outputDir, ".whisper-temp.wav")
		args := []string{"-y", "-i", sourcePath, "-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1", tempFile}
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to extract audio for whisper: %w — %s", err, stderr)), nil
		}
		audioSource = tempFile
	}

	// Run whisper CLI
	args := []string{audioSource, "--model", model, "--output_dir", outputDir, "--output_format", "srt"}
	if language != "" {
		args = append(args, "--language", language)
	}

	cmd := exec.CommandContext(ctx, "whisper", args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if tempFile != "" {
			os.Remove(tempFile)
		}
		return tools.ErrorResult(fmt.Errorf("whisper failed: %w — %s", err, stderr.String())), nil
	}

	// Clean up temp audio
	if tempFile != "" {
		os.Remove(tempFile)
	}

	// Find the generated SRT file (whisper names it <basename>.srt)
	baseName := strings.TrimSuffix(filepath.Base(audioSource), filepath.Ext(audioSource))
	generatedSRT := filepath.Join(outputDir, baseName+".srt")

	// If user specified output_path, rename
	if outputPath != "" && outputPath != generatedSRT {
		if err := os.Rename(generatedSRT, outputPath); err != nil {
			// Try copy if rename fails (cross-device)
			data, readErr := os.ReadFile(generatedSRT)
			if readErr == nil {
				os.WriteFile(outputPath, data, 0644)
				os.Remove(generatedSRT)
			}
		}
		generatedSRT = outputPath
	}

	// Parse generated SRT for summary
	entries, err := ParseSRT(generatedSRT)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("# SRT Generate\n\nWhisper completed but could not parse output.\n**File:** `%s`\n**Error:** %v\n", generatedSRT, err)), nil
	}

	totalDur := 0.0
	if len(entries) > 0 {
		totalDur = entries[len(entries)-1].EndTime
	}

	var sb strings.Builder
	sb.WriteString("# SRT Generated\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", sourcePath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", generatedSRT))
	sb.WriteString(fmt.Sprintf("**Model:** %s\n", model))
	if language != "" {
		sb.WriteString(fmt.Sprintf("**Language:** %s\n", language))
	}
	sb.WriteString(fmt.Sprintf("**Entries:** %d\n", len(entries)))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(totalDur)))

	return tools.TextResult(sb.String()), nil
}

// characterLineRegex detects character attribution: "CHARACTER NAME: dialogue text"
var characterLineRegex = regexp.MustCompile(`^([A-Z][A-Z\s.']+):\s*(.+)`)

// handleSRTCharacterList lists speaking characters with line counts from SRT and optionally IMDB
func handleSRTCharacterList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	srtPath, errResult := tools.RequireStringParam(req, "srt_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(srtPath) {
		return tools.ErrorResult(fmt.Errorf("SRT file not found: %s", srtPath)), nil
	}

	imdbID := tools.GetStringParam(req, "imdb_id")

	entries, err := ParseSubtitle(srtPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse subtitle: %w", err)), nil
	}

	// Detect characters from SRT via regex
	type charInfo struct {
		Lines     int
		FirstTime float64
		LastTime  float64
		IMDBMatch bool
	}
	characters := make(map[string]*charInfo)

	for _, e := range entries {
		m := characterLineRegex.FindStringSubmatch(e.Text)
		if m == nil {
			continue
		}
		name := strings.TrimSpace(m[1])
		info, ok := characters[name]
		if !ok {
			info = &charInfo{FirstTime: e.StartTime}
			characters[name] = info
		}
		info.Lines++
		info.LastTime = e.EndTime
	}

	// Optionally cross-reference with IMDB
	imdbCharacters := make(map[string]bool)
	if imdbID != "" {
		quotes, err := fetchIMDBQuotes(imdbID)
		if err == nil {
			for _, q := range quotes {
				for _, l := range q.Lines {
					if l.Character != "" {
						imdbCharacters[strings.ToUpper(l.Character)] = true
					}
				}
			}
			// Add IMDB-only characters not found in SRT
			for name := range imdbCharacters {
				if _, ok := characters[name]; ok {
					characters[name].IMDBMatch = true
				}
			}
		}
	}

	// Sort by line count descending
	type charRow struct {
		Name string
		Info *charInfo
	}
	var rows []charRow
	for name, info := range characters {
		// Check case-insensitive IMDB match
		if imdbCharacters[strings.ToUpper(name)] {
			info.IMDBMatch = true
		}
		rows = append(rows, charRow{Name: name, Info: info})
	}
	// Sort by lines descending
	for i := 0; i < len(rows); i++ {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].Info.Lines > rows[i].Info.Lines {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Character List: %s\n\n", filepath.Base(srtPath)))
	sb.WriteString(fmt.Sprintf("**Subtitle entries:** %d\n", len(entries)))
	sb.WriteString(fmt.Sprintf("**Characters found:** %d\n", len(rows)))
	if imdbID != "" {
		sb.WriteString(fmt.Sprintf("**IMDB cross-ref:** %s (%d characters)\n", imdbID, len(imdbCharacters)))
	}
	sb.WriteString("\n")

	if len(rows) == 0 {
		sb.WriteString("No character attributions detected in subtitles.\n")
		sb.WriteString("(Character detection requires format: `CHARACTER NAME: dialogue text`)\n")

		// If IMDB characters found, list them
		if len(imdbCharacters) > 0 {
			sb.WriteString("\n## IMDB Characters\n\n")
			for name := range imdbCharacters {
				sb.WriteString(fmt.Sprintf("- %s\n", name))
			}
		}
		return tools.TextResult(sb.String()), nil
	}

	imdbCol := ""
	imdbHeader := ""
	if imdbID != "" {
		imdbCol = " IMDB |"
		imdbHeader = "------|"
	}
	sb.WriteString(fmt.Sprintf("| Character | Lines | First | Last |%s\n", imdbCol))
	sb.WriteString(fmt.Sprintf("|-----------|-------|-------|------|%s\n", imdbHeader))
	for _, r := range rows {
		imdbVal := ""
		if imdbID != "" {
			if r.Info.IMDBMatch {
				imdbVal = " Yes |"
			} else {
				imdbVal = " — |"
			}
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s |%s\n",
			r.Name, r.Info.Lines,
			formatDuration(r.Info.FirstTime), formatDuration(r.Info.LastTime),
			imdbVal))
	}

	return tools.TextResult(sb.String()), nil
}
