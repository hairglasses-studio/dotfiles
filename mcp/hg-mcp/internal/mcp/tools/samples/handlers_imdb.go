package samples

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleIMDBScenePreview performs a dry-run of scene matching without extracting audio
func handleIMDBScenePreview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imdbID, errResult := tools.RequireStringParam(req, "imdb_id")
	if errResult != nil {
		return errResult, nil
	}

	srtPath, errResult := tools.RequireStringParam(req, "srt_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(srtPath) {
		return tools.ErrorResult(fmt.Errorf("SRT file not found: %s", srtPath)), nil
	}

	minSceneDuration := tools.GetFloatParam(req, "min_scene_duration", 15.0)
	maxSceneDuration := tools.GetFloatParam(req, "max_scene_duration", 60.0)
	matchThreshold := tools.GetFloatParam(req, "match_threshold", 0.3)
	maxScenes := tools.GetIntParam(req, "max_scenes", 30)
	characterFilter := tools.GetStringParam(req, "character")

	// Step 1: Fetch IMDB quotes
	quotes, err := fetchIMDBQuotes(imdbID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to fetch IMDB quotes: %w", err)), nil
	}
	if len(quotes) == 0 {
		return tools.TextResult(fmt.Sprintf("# IMDB Scene Preview\n\nNo quotes found for %s on IMDB.\n", imdbID)), nil
	}

	// Optional character filter
	if characterFilter != "" {
		filterLower := strings.ToLower(characterFilter)
		var filtered []IMDBQuote
		for _, q := range quotes {
			keep := false
			for _, l := range q.Lines {
				if strings.Contains(strings.ToLower(l.Character), filterLower) {
					keep = true
					break
				}
			}
			if keep {
				filtered = append(filtered, q)
			}
		}
		quotes = filtered
		if len(quotes) == 0 {
			return tools.TextResult(fmt.Sprintf("# IMDB Scene Preview\n\nNo quotes matching character \"%s\" found for %s.\n", characterFilter, imdbID)), nil
		}
	}

	// Step 2: Parse subtitles
	entries, err := ParseSubtitle(srtPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse subtitle: %w", err)), nil
	}
	if len(entries) == 0 {
		return tools.ErrorResult(fmt.Errorf("no subtitle entries found in %s", srtPath)), nil
	}

	// Step 3: Match quotes against SRT
	matches := matchQuotesAgainstSRT(quotes, entries, matchThreshold)
	if len(matches) == 0 {
		return tools.TextResult(fmt.Sprintf("# IMDB Scene Preview: %s\n\nNo quotes matched against subtitles (threshold: %.0f%%).\nTry lowering match_threshold.\n", imdbID, matchThreshold*100)), nil
	}

	// Step 4: Expand scene windows
	for i := range matches {
		expandSceneWindow(&matches[i], entries, minSceneDuration, maxSceneDuration)
	}

	// Step 5: Merge overlapping scenes
	matches = mergeOverlappingScenes(matches)

	// Step 6: Limit
	if len(matches) > maxScenes {
		matches = matches[:maxScenes]
	}

	// Build rich preview output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# IMDB Scene Preview: %s\n\n", imdbID))
	sb.WriteString(fmt.Sprintf("**Quotes fetched:** %d\n", len(quotes)))
	sb.WriteString(fmt.Sprintf("**Scenes matched:** %d\n", len(matches)))
	sb.WriteString(fmt.Sprintf("**Threshold:** %.0f%%\n", matchThreshold*100))
	if characterFilter != "" {
		sb.WriteString(fmt.Sprintf("**Character filter:** %s\n", characterFilter))
	}

	totalDur := 0.0
	for _, m := range matches {
		totalDur += m.EndTime - m.StartTime
	}
	sb.WriteString(fmt.Sprintf("**Total duration:** %s\n\n", formatDuration(totalDur)))

	// Confidence table
	sb.WriteString("## Scene Summary\n\n")
	sb.WriteString("| # | Start | End | Duration | Confidence | Scene |\n")
	sb.WriteString("|---|-------|-----|----------|------------|-------|\n")

	for i, m := range matches {
		durSec := m.EndTime - m.StartTime
		sceneName := ""
		if len(m.Quote.Lines) > 0 {
			first := m.Quote.Lines[0]
			if first.Character != "" {
				sceneName = first.Character + ": " + first.Text
			} else {
				sceneName = first.Text
			}
			if len(sceneName) > 50 {
				sceneName = sceneName[:50] + "..."
			}
		}
		sceneName = strings.ReplaceAll(sceneName, "|", "\\|")

		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %.1fs | %.0f%% | %s |\n",
			i+1, formatDuration(m.StartTime), formatDuration(m.EndTime),
			durSec, m.Confidence*100, sceneName))
	}

	// Detailed dialogue for each scene
	sb.WriteString("\n## Scene Detail\n")
	for i, m := range matches {
		sb.WriteString(fmt.Sprintf("\n### Scene %d — %s → %s (%.1fs, %.0f%% confidence)\n\n",
			i+1, formatDuration(m.StartTime), formatDuration(m.EndTime),
			m.EndTime-m.StartTime, m.Confidence*100))

		// IMDB dialogue
		sb.WriteString("**IMDB Dialogue:**\n")
		for _, l := range m.Quote.Lines {
			if l.Character != "" {
				sb.WriteString(fmt.Sprintf("> **%s:** %s\n", l.Character, l.Text))
			} else {
				sb.WriteString(fmt.Sprintf("> %s\n", l.Text))
			}
		}
		for _, mq := range m.MergedQuotes {
			for _, l := range mq.Lines {
				if l.Character != "" {
					sb.WriteString(fmt.Sprintf("> **%s:** %s\n", l.Character, l.Text))
				} else {
					sb.WriteString(fmt.Sprintf("> %s\n", l.Text))
				}
			}
		}

		// SRT context
		sb.WriteString("\n**SRT Context:**\n")
		for _, e := range m.SRTEntries {
			startDisp := secondsToSRTTime(e.StartTime)[:8]
			endDisp := secondsToSRTTime(e.EndTime)[:8]
			sb.WriteString(fmt.Sprintf("  [%s → %s] %s\n", startDisp, endDisp, e.Text))
		}
	}

	// Suggested extraction command
	sb.WriteString("\n## Suggested Extraction\n\n")
	sb.WriteString("```\n")
	sb.WriteString(fmt.Sprintf("imdb_scene_extract imdb_id=%s srt_path=%s source_path=<VIDEO> pack_name=<PACK>\n", imdbID, srtPath))
	sb.WriteString("```\n")

	return tools.TextResult(sb.String()), nil
}

// handleIMDBQuotes fetches and displays IMDB quotes for a title
func handleIMDBQuotes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imdbID, errResult := tools.RequireStringParam(req, "imdb_id")
	if errResult != nil {
		return errResult, nil
	}

	maxQuotes := tools.GetIntParam(req, "max_quotes", 50)

	quotes, err := fetchIMDBQuotes(imdbID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to fetch quotes: %w", err)), nil
	}

	if len(quotes) == 0 {
		return tools.TextResult(fmt.Sprintf("# IMDB Quotes: %s\n\nNo quotes found for this title.\n", imdbID)), nil
	}

	if len(quotes) > maxQuotes {
		quotes = quotes[:maxQuotes]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# IMDB Quotes: %s\n\n", imdbID))
	sb.WriteString(fmt.Sprintf("**Quotes found:** %d\n\n", len(quotes)))

	sb.WriteString("| # | Character | Dialogue |\n")
	sb.WriteString("|---|-----------|----------|\n")

	for i, q := range quotes {
		for j, line := range q.Lines {
			num := ""
			if j == 0 {
				num = fmt.Sprintf("%d", i+1)
			}
			char := line.Character
			if char == "" {
				char = "—"
			}
			// Escape pipes in dialogue for markdown table
			text := strings.ReplaceAll(line.Text, "|", "\\|")
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", num, char, text))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleIMDBSceneExtract fetches IMDB quotes, matches against SRT, and extracts scene audio clips
func handleIMDBSceneExtract(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	imdbID, errResult := tools.RequireStringParam(req, "imdb_id")
	if errResult != nil {
		return errResult, nil
	}

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

	artist := tools.GetStringParam(req, "artist")
	if artist == "" {
		artist = packName
	}
	minSceneDuration := tools.GetFloatParam(req, "min_scene_duration", 15.0)
	maxSceneDuration := tools.GetFloatParam(req, "max_scene_duration", 60.0)
	matchThreshold := tools.GetFloatParam(req, "match_threshold", 0.3)
	maxScenes := tools.GetIntParam(req, "max_scenes", 30)
	process := tools.GetBoolParam(req, "process", true)
	generateXML := tools.GetBoolParam(req, "generate_xml", true)
	targetSampleRate := tools.GetIntParam(req, "target_sample_rate", 0)
	targetFormat := tools.GetStringParam(req, "target_format")

	// Step 1: Fetch IMDB quotes
	quotes, err := fetchIMDBQuotes(imdbID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to fetch IMDB quotes: %w", err)), nil
	}
	if len(quotes) == 0 {
		return tools.TextResult(fmt.Sprintf("# IMDB Scene Extract\n\nNo quotes found for %s on IMDB.\n", imdbID)), nil
	}

	// Step 2: Parse subtitles
	entries, err := ParseSubtitle(srtPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse subtitle: %w", err)), nil
	}
	if len(entries) == 0 {
		return tools.ErrorResult(fmt.Errorf("no subtitle entries found in %s", srtPath)), nil
	}

	// Step 3: Match quotes against SRT
	matches := matchQuotesAgainstSRT(quotes, entries, matchThreshold)
	if len(matches) == 0 {
		return tools.TextResult(fmt.Sprintf("# IMDB Scene Extract: %s\n\nNo quotes matched against subtitles (threshold: %.0f%%).\nTry lowering match_threshold.\n", packName, matchThreshold*100)), nil
	}

	// Step 4: Expand scene windows
	for i := range matches {
		expandSceneWindow(&matches[i], entries, minSceneDuration, maxSceneDuration)
	}

	// Step 5: Merge overlapping scenes
	matches = mergeOverlappingScenes(matches)

	// Step 6: Limit to max scenes
	if len(matches) > maxScenes {
		matches = matches[:maxScenes]
	}

	// Step 7: Create pack directory
	samplesBaseDir, err := defaultSamplesDir()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	packDir := filepath.Join(samplesBaseDir, packName)
	if err := os.MkdirAll(packDir, 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot create pack directory: %w", err)), nil
	}

	// Step 8: Extract video audio to temp AIFF if needed
	audioSource := sourcePath
	var tempAudioFile string
	if isVideoFile(ctx, sourcePath) {
		tempAudioFile = filepath.Join(packDir, ".full-audio.aiff")
		args := []string{"-y", "-i", sourcePath, "-vn", "-acodec", "pcm_s16be", "-ar", "44100", "-ac", "2", tempAudioFile}
		_, stderr, err := runFFmpeg(ctx, args...)
		if err != nil {
			return tools.TextResult(fmt.Sprintf("# IMDB Scene Extract — Error\n\nFailed to extract audio from video.\n\n**Error:** %v\n```\n%s\n```\n", err, stderr)), nil
		}
		audioSource = tempAudioFile
	}

	// Step 9: Convert QuoteMatches to SampleDefinitions
	var defs []SampleDefinition
	for i, m := range matches {
		// Build slug from Character + Text of first quote line
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

		// Build human-readable display name from all primary quote lines
		var displayParts []string
		for _, l := range m.Quote.Lines {
			if l.Character != "" {
				displayParts = append(displayParts, fmt.Sprintf("%s: %s", l.Character, l.Text))
			} else {
				displayParts = append(displayParts, l.Text)
			}
		}
		displayName := strings.Join(displayParts, " / ")

		// Build description including merged quotes
		var descParts []string
		descParts = append(descParts, displayParts...)
		for _, mq := range m.MergedQuotes {
			for _, l := range mq.Lines {
				if l.Character != "" {
					descParts = append(descParts, fmt.Sprintf("%s: %s", l.Character, l.Text))
				} else {
					descParts = append(descParts, l.Text)
				}
			}
		}

		defs = append(defs, SampleDefinition{
			Name:        name,
			DisplayName: displayName,
			Start:       fmt.Sprintf("%.3f", m.StartTime),
			End:         fmt.Sprintf("%.3f", m.EndTime),
			Description: strings.Join(descParts, " / "),
		})
	}

	// Step 10: Run batch pipeline
	results := runBatchPipeline(ctx, audioSource, packDir, defs, process)

	// Step 11: CDJ conversion if target_sample_rate > 0
	cdjDir := filepath.Join(packDir, "cdj")
	var converted int
	if targetSampleRate > 0 {
		if err := os.MkdirAll(cdjDir, 0755); err != nil {
			return tools.ErrorResult(fmt.Errorf("cannot create cdj directory: %w", err)), nil
		}
		for i, res := range results {
			if res.Error != "" {
				continue
			}
			probe, err := runFFprobe(ctx, res.Path)
			if err != nil {
				continue
			}
			if probe.SampleRate <= targetSampleRate {
				continue
			}
			ext := filepath.Ext(res.Path)
			baseName := strings.TrimSuffix(filepath.Base(res.Path), ext)
			outExt := ext
			if targetFormat != "" {
				outExt = "." + targetFormat
			}
			outPath := filepath.Join(cdjDir, baseName+outExt)

			args := []string{"-y", "-i", res.Path, "-ar", strconv.Itoa(targetSampleRate)}
			switch strings.TrimPrefix(strings.ToLower(outExt), ".") {
			case "wav":
				args = append(args, "-c:a", "pcm_s16le")
			case "aiff", "aif":
				args = append(args, "-c:a", "pcm_s16be")
			case "flac":
				args = append(args, "-c:a", "flac")
			default:
				args = append(args, "-c:a", "pcm_s16le")
			}
			args = append(args, outPath)

			if _, _, err := runFFmpeg(ctx, args...); err == nil {
				results[i].Path = outPath
				converted++
			}
		}
	}

	// Step 12: Generate Rekordbox XML if requested
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
	sb.WriteString(fmt.Sprintf("# IMDB Scene Extract: %s\n\n", packName))
	sb.WriteString(fmt.Sprintf("**IMDB ID:** %s\n", imdbID))
	sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", sourcePath))
	sb.WriteString(fmt.Sprintf("**Subtitles:** `%s`\n", srtPath))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", packDir))
	sb.WriteString(fmt.Sprintf("**Quotes fetched:** %d\n", len(quotes)))
	sb.WriteString(fmt.Sprintf("**Scenes matched:** %d\n\n", len(matches)))

	// Match confidence table
	sb.WriteString("## Scene Matches\n\n")
	sb.WriteString("| # | Scene | Start | End | Duration | Confidence | Status |\n")
	sb.WriteString("|---|-------|-------|-----|----------|------------|--------|\n")

	successCount := 0
	for i, m := range matches {
		status := "OK"
		if i < len(results) && results[i].Error != "" {
			status = "ERROR"
		} else {
			successCount++
		}

		durSec := m.EndTime - m.StartTime
		// First line of the quote for the scene name
		sceneName := ""
		if len(m.Quote.Lines) > 0 {
			first := m.Quote.Lines[0]
			if first.Character != "" {
				sceneName = first.Character + ": " + first.Text
			} else {
				sceneName = first.Text
			}
			if len(sceneName) > 50 {
				sceneName = sceneName[:50] + "..."
			}
		}
		sceneName = strings.ReplaceAll(sceneName, "|", "\\|")

		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %.1fs | %.0f%% | %s |\n",
			i+1, sceneName,
			formatDuration(m.StartTime), formatDuration(m.EndTime),
			durSec, m.Confidence*100, status))
	}

	sb.WriteString(fmt.Sprintf("\n**Extracted:** %d/%d scenes\n", successCount, len(matches)))
	if targetSampleRate > 0 {
		sb.WriteString(fmt.Sprintf("**CDJ converted:** %d files → `%s` (%d Hz)\n", converted, cdjDir, targetSampleRate))
	}
	if xmlPath != "" {
		sb.WriteString(fmt.Sprintf("**Rekordbox XML:** `%s`\n", xmlPath))
	}

	// Show errors
	for _, res := range results {
		if res.Error != "" {
			sb.WriteString(fmt.Sprintf("\n**Error** (%s): %s\n", res.Name, res.Error))
		}
	}

	return tools.TextResult(sb.String()), nil
}
