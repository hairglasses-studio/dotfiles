package samples

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handleRekordboxXML generates a Rekordbox XML playlist from files in a directory
func handleRekordboxXML(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	directory, errResult := tools.RequireStringParam(req, "directory")
	if errResult != nil {
		return errResult, nil
	}
	if !dirExists(directory) {
		return tools.ErrorResult(fmt.Errorf("directory not found: %s", directory)), nil
	}

	playlistName := tools.GetStringParam(req, "playlist_name")
	if playlistName == "" {
		playlistName = filepath.Base(directory)
	}

	artist := tools.GetStringParam(req, "artist")
	genre := tools.OptionalStringParam(req, "genre", "DJ Sample")

	targetSampleRate := tools.GetIntParam(req, "target_sample_rate", 0)
	targetFormat := tools.GetStringParam(req, "target_format")

	// Scan directory for audio files
	entries, err := os.ReadDir(directory)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot read directory: %w", err)), nil
	}

	// Create cdj/ subdirectory if conversion is needed
	cdjDir := filepath.Join(directory, "cdj")
	if targetSampleRate > 0 {
		if err := os.MkdirAll(cdjDir, 0755); err != nil {
			return tools.ErrorResult(fmt.Errorf("cannot create cdj directory: %w", err)), nil
		}
	}

	var tracks []RekordboxTrack
	var converted, skipped int
	id := 1
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(directory, entry.Name())
		if !isAudioFile(filePath) {
			continue
		}

		probe, err := runFFprobe(ctx, filePath)
		if err != nil {
			continue
		}

		trackPath := filePath
		trackSampleRate := probe.SampleRate
		trackBitRate := probe.BitRate
		trackFileSize := probe.FileSize

		// Convert if sample rate exceeds target
		if targetSampleRate > 0 && probe.SampleRate > targetSampleRate {
			ext := filepath.Ext(entry.Name())
			baseName := strings.TrimSuffix(entry.Name(), ext)

			// Determine output format/extension
			outExt := ext
			if targetFormat != "" {
				outExt = "." + targetFormat
			}
			outPath := filepath.Join(cdjDir, baseName+outExt)

			// Build ffmpeg args
			args := []string{"-y", "-i", filePath, "-ar", strconv.Itoa(targetSampleRate)}
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

			if _, stderr, err := runFFmpeg(ctx, args...); err != nil {
				return tools.ErrorResult(fmt.Errorf("conversion failed for %s: %w — %s", entry.Name(), err, stderr)), nil
			}

			// Re-probe converted file for accurate metadata
			if convProbe, err := runFFprobe(ctx, outPath); err == nil {
				trackSampleRate = convProbe.SampleRate
				trackBitRate = convProbe.BitRate
				trackFileSize = convProbe.FileSize
			}
			trackPath = outPath
			converted++
		} else if targetSampleRate > 0 {
			skipped++
		}

		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		tracks = append(tracks, RekordboxTrack{
			ID:         id,
			Name:       name,
			Artist:     artist,
			Album:      playlistName,
			Genre:      genre,
			FilePath:   trackPath,
			Duration:   probe.Duration,
			FileSize:   trackFileSize,
			SampleRate: trackSampleRate,
			BitRate:    trackBitRate,
		})
		id++
	}

	if len(tracks) == 0 {
		return tools.ErrorResult(fmt.Errorf("no audio files found in %s", directory)), nil
	}

	outputPath := filepath.Join(directory, playlistName+".xml")
	if err := GenerateRekordboxXML(tracks, playlistName, outputPath); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to generate XML: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Rekordbox XML Generated\n\n")
	sb.WriteString(fmt.Sprintf("**Playlist:** %s\n", playlistName))
	sb.WriteString(fmt.Sprintf("**Tracks:** %d\n", len(tracks)))
	sb.WriteString(fmt.Sprintf("**Output:** `%s`\n\n", outputPath))

	if targetSampleRate > 0 {
		sb.WriteString("## CDJ Compatibility Conversion\n\n")
		sb.WriteString(fmt.Sprintf("**Target sample rate:** %d Hz\n", targetSampleRate))
		if targetFormat != "" {
			sb.WriteString(fmt.Sprintf("**Target format:** %s\n", strings.ToUpper(targetFormat)))
		}
		sb.WriteString(fmt.Sprintf("**Converted:** %d files → `%s`\n", converted, cdjDir))
		sb.WriteString(fmt.Sprintf("**Already compatible:** %d files (used as-is)\n\n", skipped))
	}

	sb.WriteString("| # | Track | Duration |\n")
	sb.WriteString("|---|-------|----------|\n")
	for i, t := range tracks {
		sb.WriteString(fmt.Sprintf("| %d | %s | %s |\n", i+1, t.Name, formatDuration(t.Duration)))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSetCuePoints injects cue points into a Rekordbox XML file
func handleSetCuePoints(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	xmlPath, errResult := tools.RequireStringParam(req, "xml_path")
	if errResult != nil {
		return errResult, nil
	}
	if !fileExists(xmlPath) {
		return tools.ErrorResult(fmt.Errorf("XML file not found: %s", xmlPath)), nil
	}

	trackName, errResult := tools.RequireStringParam(req, "track_name")
	if errResult != nil {
		return errResult, nil
	}

	cuesJSON, errResult := tools.RequireStringParam(req, "cue_points")
	if errResult != nil {
		return errResult, nil
	}

	type cueInput struct {
		Position float64 `json:"position"`
		Name     string  `json:"name"`
		Type     int     `json:"type"` // 0=hot cue, 1=memory cue
		Red      int     `json:"red"`
		Green    int     `json:"green"`
		Blue     int     `json:"blue"`
	}
	var cues []cueInput
	if err := json.Unmarshal([]byte(cuesJSON), &cues); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid cue_points JSON: %w", err)), nil
	}

	// Read and parse XML
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot read XML: %w", err)), nil
	}

	// Simple approach: reload, find track, add cues, regenerate
	// For simplicity, we parse the high-level structure
	var doc djPlaylists
	if err := xml.Unmarshal(data, &doc); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid Rekordbox XML: %w", err)), nil
	}

	found := false
	for i := range doc.Collection.Tracks {
		if strings.EqualFold(doc.Collection.Tracks[i].Name, trackName) {
			found = true
			for j, c := range cues {
				name := c.Name
				if name == "" {
					name = fmt.Sprintf("Cue %d", j+1)
				}
				doc.Collection.Tracks[i].PositionMarks = append(doc.Collection.Tracks[i].PositionMarks, xmlPositionMark{
					Name:  name,
					Type:  c.Type,
					Start: c.Position,
					Num:   j,
					Red:   c.Red,
					Green: c.Green,
					Blue:  c.Blue,
				})
			}
			break
		}
	}

	if !found {
		return tools.ErrorResult(fmt.Errorf("track %q not found in XML", trackName)), nil
	}

	// Write back
	f, err := os.Create(xmlPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot write XML: %w", err)), nil
	}
	defer f.Close()
	f.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to encode XML: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Cue Points Set\n\n")
	sb.WriteString(fmt.Sprintf("**XML:** `%s`\n", xmlPath))
	sb.WriteString(fmt.Sprintf("**Track:** %s\n", trackName))
	sb.WriteString(fmt.Sprintf("**Cues added:** %d\n\n", len(cues)))

	sb.WriteString("| # | Position | Name | Type |\n")
	sb.WriteString("|---|----------|------|------|\n")
	for i, c := range cues {
		typeStr := "Hot Cue"
		if c.Type == 1 {
			typeStr = "Memory Cue"
		}
		name := c.Name
		if name == "" {
			name = fmt.Sprintf("Cue %d", i+1)
		}
		sb.WriteString(fmt.Sprintf("| %d | %.3fs | %s | %s |\n", i+1, c.Position, name, typeStr))
	}

	return tools.TextResult(sb.String()), nil
}
