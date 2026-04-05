package samples

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// SRTEntry represents a single subtitle entry from an SRT file
type SRTEntry struct {
	Index     int
	StartTime float64 // seconds
	EndTime   float64 // seconds
	StartTS   string  // original "HH:MM:SS,mmm" string
	EndTS     string  // original "HH:MM:SS,mmm" string
	Text      string  // cleaned text (HTML tags stripped)
}

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// ParseSRT reads an SRT file and returns parsed subtitle entries
func ParseSRT(path string) ([]SRTEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open SRT file: %w", err)
	}
	defer f.Close()

	var entries []SRTEntry
	scanner := bufio.NewScanner(f)

	for {
		// Skip blank lines to find the next index
		var indexLine string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// Strip BOM from first line
			line = strings.TrimPrefix(line, "\xef\xbb\xbf")
			if line != "" {
				indexLine = line
				break
			}
		}
		if indexLine == "" {
			break // EOF
		}

		idx, err := strconv.Atoi(indexLine)
		if err != nil {
			// Not an index line, skip
			continue
		}

		// Read timestamp line
		if !scanner.Scan() {
			break
		}
		tsLine := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(tsLine, "-->", 2)
		if len(parts) != 2 {
			continue
		}

		startTS := strings.TrimSpace(parts[0])
		endTS := strings.TrimSpace(parts[1])

		startSec, err := srtTimeToSeconds(startTS)
		if err != nil {
			continue
		}
		endSec, err := srtTimeToSeconds(endTS)
		if err != nil {
			continue
		}

		// Read text lines until blank line or EOF
		var textLines []string
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				break
			}
			// Strip HTML tags
			cleaned := htmlTagRegex.ReplaceAllString(line, "")
			textLines = append(textLines, cleaned)
		}

		entries = append(entries, SRTEntry{
			Index:     idx,
			StartTime: startSec,
			EndTime:   endSec,
			StartTS:   startTS,
			EndTS:     endTS,
			Text:      strings.Join(textLines, " "),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SRT file: %w", err)
	}

	return entries, nil
}

// srtTimeToSeconds converts "HH:MM:SS,mmm" to seconds
func srtTimeToSeconds(ts string) (float64, error) {
	ts = strings.TrimSpace(ts)
	// Handle both comma and period as ms separator
	ts = strings.Replace(ts, ",", ".", 1)

	// Parse HH:MM:SS.mmm
	parts := strings.SplitN(ts, ":", 3)
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid SRT timestamp: %s", ts)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours in timestamp: %s", ts)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes in timestamp: %s", ts)
	}

	seconds, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return 0, fmt.Errorf("invalid seconds in timestamp: %s", ts)
	}

	return float64(hours)*3600 + float64(minutes)*60 + seconds, nil
}

// secondsToSRTTime converts seconds to "HH:MM:SS,mmm" format
func secondsToSRTTime(sec float64) string {
	if sec < 0 {
		sec = 0
	}
	hours := int(sec) / 3600
	minutes := (int(sec) % 3600) / 60
	secs := int(sec) % 60
	ms := int((sec - float64(int(sec))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, ms)
}

// subtitleExtensions are file extensions treated as subtitle files
var subtitleExtensions = map[string]bool{
	".srt": true,
	".vtt": true,
}

// findSubtitleFiles finds .srt/.vtt files alongside a video or in a directory
func findSubtitleFiles(path string) []string {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	var dir string
	var baseName string
	if info.IsDir() {
		dir = path
	} else {
		dir = filepath.Dir(path)
		baseName = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var results []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if !subtitleExtensions[ext] {
			continue
		}
		// If we have a base name, prefer matching files
		if baseName != "" {
			entryBase := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			// Also strip language suffix like .en, .eng
			entryBase = strings.TrimSuffix(entryBase, filepath.Ext(entryBase))
			if !strings.EqualFold(entryBase, baseName) && !strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(baseName)) {
				continue
			}
		}
		results = append(results, filepath.Join(dir, entry.Name()))
	}

	return results
}

// findSubtitleFilesRecursive finds subtitle files recursively in a directory
func findSubtitleFilesRecursive(dir string) []string {
	var results []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if subtitleExtensions[ext] {
			results = append(results, path)
		}
		return nil
	})
	return results
}

// ParseSubtitle is an alias for ParseSRT for backward compatibility
var ParseSubtitle = ParseSRT

// slugify converts text to a URL/filename-safe slug
func slugify(text string) string {
	text = strings.ToLower(text)
	var result []rune
	prevDash := false
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result = append(result, r)
			prevDash = false
		} else if !prevDash && len(result) > 0 {
			result = append(result, '-')
			prevDash = true
		}
	}
	// Trim trailing dash
	s := string(result)
	s = strings.TrimRight(s, "-")
	if len(s) > 60 {
		s = s[:60]
		s = strings.TrimRight(s, "-")
	}
	return s
}
