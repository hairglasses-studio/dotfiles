package samples

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// IMDBQuoteLine represents a single line of dialogue from an IMDB quote
type IMDBQuoteLine struct {
	Character string
	Text      string
}

// IMDBQuote represents a grouped dialogue exchange from IMDB
type IMDBQuote struct {
	Lines []IMDBQuoteLine
}

// FullText returns all dialogue lines concatenated for matching
func (q IMDBQuote) FullText() string {
	var parts []string
	for _, l := range q.Lines {
		parts = append(parts, l.Text)
	}
	return strings.Join(parts, " ")
}

// QuoteMatch represents a matched scene between an IMDB quote and SRT entries
type QuoteMatch struct {
	Quote        IMDBQuote
	SRTEntries   []SRTEntry
	StartTime    float64
	EndTime      float64
	Confidence   float64
	MatchedText  string
	MergedQuotes []IMDBQuote // secondary quotes absorbed during merge
}

// fetchIMDBQuotes fetches and parses quotes from an IMDB title page
func fetchIMDBQuotes(imdbID string) ([]IMDBQuote, error) {
	url := fmt.Sprintf("https://www.imdb.com/title/%s/quotes/", imdbID)

	client := httpclient.Standard()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch IMDB quotes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("IMDB returned status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return parseIMDBQuotesHTML(string(body)), nil
}

// quoteBlockRegex matches data-testid="item-html" divs containing quote blocks
var quoteBlockRegex = regexp.MustCompile(`(?s)data-testid="item-html"[^>]*>(.*?)</div>`)

// quoteLiRegex matches individual <li> elements within a quote block
var quoteLiRegex = regexp.MustCompile(`(?s)<li[^>]*>(.*?)</li>`)

// quoteLineRegex matches character name + dialogue: <a ...>CharName</a>: dialogue
var quoteLineRegex = regexp.MustCompile(`(?s)<a[^>]*>([^<]+)</a>\s*:\s*(.+)`)

// parseIMDBQuotesHTML extracts quotes from IMDB HTML
func parseIMDBQuotesHTML(html string) []IMDBQuote {
	var quotes []IMDBQuote

	blocks := quoteBlockRegex.FindAllStringSubmatch(html, -1)
	for _, block := range blocks {
		if len(block) < 2 {
			continue
		}
		content := block[1]
		lis := quoteLiRegex.FindAllStringSubmatch(content, -1)
		if len(lis) == 0 {
			continue
		}

		var q IMDBQuote
		for _, li := range lis {
			if len(li) < 2 {
				continue
			}
			liContent := li[1]
			match := quoteLineRegex.FindStringSubmatch(liContent)
			if match != nil && len(match) >= 3 {
				character := strings.TrimSpace(decodeHTMLEntities(match[1]))
				text := strings.TrimSpace(decodeHTMLEntities(stripHTMLTags(match[2])))
				if text != "" {
					q.Lines = append(q.Lines, IMDBQuoteLine{Character: character, Text: text})
				}
			} else {
				// Line without character name attribution
				text := strings.TrimSpace(decodeHTMLEntities(stripHTMLTags(liContent)))
				if text != "" {
					q.Lines = append(q.Lines, IMDBQuoteLine{Text: text})
				}
			}
		}
		if len(q.Lines) > 0 {
			quotes = append(quotes, q)
		}
	}

	return quotes
}

// stripHTMLTags removes all HTML tags from a string
func stripHTMLTags(s string) string {
	return htmlTagRegex.ReplaceAllString(s, "")
}

// decodeHTMLEntities converts common HTML entities to their characters
func decodeHTMLEntities(s string) string {
	replacer := strings.NewReplacer(
		"&#39;", "'",
		"&#x27;", "'",
		"&apos;", "'",
		"&quot;", "\"",
		"&#34;", "\"",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&nbsp;", " ",
		"&#x2F;", "/",
		"&#47;", "/",
		"&hellip;", "...",
		"&#8230;", "...",
		"&mdash;", "—",
		"&#8212;", "—",
		"&ndash;", "–",
		"&#8211;", "–",
	)
	return replacer.Replace(s)
}

// normalizeText lowercases, strips punctuation, and collapses whitespace for fuzzy matching
func normalizeText(s string) string {
	s = strings.ToLower(s)
	var result []rune
	prevSpace := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result = append(result, r)
			prevSpace = false
		} else if !prevSpace && len(result) > 0 {
			result = append(result, ' ')
			prevSpace = true
		}
	}
	return strings.TrimSpace(string(result))
}

// fuzzyMatch computes Dice coefficient on character bigrams between two strings (0.0-1.0)
func fuzzyMatch(a, b string) float64 {
	a = normalizeText(a)
	b = normalizeText(b)
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1.0
	}

	bigramsA := charBigrams(a)
	bigramsB := charBigrams(b)

	if len(bigramsA) == 0 || len(bigramsB) == 0 {
		return 0
	}

	// Count intersection
	intersection := 0
	used := make(map[string]int)
	for bg, count := range bigramsB {
		used[bg] = count
	}
	for bg, countA := range bigramsA {
		if countB, ok := used[bg]; ok {
			min := countA
			if countB < min {
				min = countB
			}
			intersection += min
		}
	}

	return 2.0 * float64(intersection) / float64(len(bigramsA)+len(bigramsB))
}

// charBigrams returns a frequency map of character bigrams in a string
func charBigrams(s string) map[string]int {
	runes := []rune(s)
	bigrams := make(map[string]int)
	for i := 0; i < len(runes)-1; i++ {
		bg := string(runes[i : i+2])
		bigrams[bg]++
	}
	return bigrams
}

// matchQuotesAgainstSRT slides a window of SRT entries across all quotes and returns best matches
func matchQuotesAgainstSRT(quotes []IMDBQuote, entries []SRTEntry, threshold float64) []QuoteMatch {
	var matches []QuoteMatch

	for _, quote := range quotes {
		quoteText := quote.FullText()
		if quoteText == "" {
			continue
		}

		var bestMatch *QuoteMatch

		// Slide window of 1-5 SRT entries
		for windowSize := 1; windowSize <= 5 && windowSize <= len(entries); windowSize++ {
			for i := 0; i <= len(entries)-windowSize; i++ {
				window := entries[i : i+windowSize]

				// Combine text from window
				var windowParts []string
				for _, e := range window {
					windowParts = append(windowParts, e.Text)
				}
				windowText := strings.Join(windowParts, " ")

				score := fuzzyMatch(quoteText, windowText)
				if score >= threshold {
					if bestMatch == nil || score > bestMatch.Confidence {
						entriesCopy := make([]SRTEntry, len(window))
						copy(entriesCopy, window)
						bestMatch = &QuoteMatch{
							Quote:       quote,
							SRTEntries:  entriesCopy,
							StartTime:   window[0].StartTime,
							EndTime:     window[len(window)-1].EndTime,
							Confidence:  score,
							MatchedText: windowText,
						}
					}
				}
			}
		}

		if bestMatch != nil {
			matches = append(matches, *bestMatch)
		}
	}

	// Sort by start time
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].StartTime < matches[j].StartTime
	})

	return matches
}

// expandSceneWindow expands a matched SRT range to meet minimum duration, capped at maximum
func expandSceneWindow(match *QuoteMatch, entries []SRTEntry, minDuration, maxDuration float64) {
	currentDuration := match.EndTime - match.StartTime
	if currentDuration >= minDuration {
		return
	}

	// Find the indices of the first and last matched entry
	startIdx := -1
	endIdx := -1
	for i, e := range entries {
		if startIdx == -1 && e.StartTime >= match.StartTime-0.01 && e.StartTime <= match.StartTime+0.01 {
			startIdx = i
		}
		if e.EndTime >= match.EndTime-0.01 && e.EndTime <= match.EndTime+0.01 {
			endIdx = i
		}
	}

	if startIdx == -1 || endIdx == -1 {
		// Fallback: expand time directly
		needed := minDuration - currentDuration
		match.StartTime -= needed / 2
		match.EndTime += needed / 2
		if match.StartTime < 0 {
			match.EndTime += -match.StartTime
			match.StartTime = 0
		}
		if match.EndTime-match.StartTime > maxDuration {
			match.EndTime = match.StartTime + maxDuration
		}
		return
	}

	// Expand outward alternating before/after until we hit minDuration or maxDuration
	for match.EndTime-match.StartTime < minDuration {
		expanded := false

		// Try expanding backward
		if startIdx > 0 {
			startIdx--
			match.StartTime = entries[startIdx].StartTime
			match.SRTEntries = append([]SRTEntry{entries[startIdx]}, match.SRTEntries...)
			expanded = true
			if match.EndTime-match.StartTime >= minDuration {
				break
			}
		}

		// Try expanding forward
		if endIdx < len(entries)-1 {
			endIdx++
			match.EndTime = entries[endIdx].EndTime
			match.SRTEntries = append(match.SRTEntries, entries[endIdx])
			expanded = true
		}

		if !expanded {
			break
		}

		// Cap at maximum
		if match.EndTime-match.StartTime > maxDuration {
			match.EndTime = match.StartTime + maxDuration
			break
		}
	}
}

// mergeOverlappingScenes merges scenes with overlapping time ranges
func mergeOverlappingScenes(matches []QuoteMatch) []QuoteMatch {
	if len(matches) <= 1 {
		return matches
	}

	// Sort by start time (should already be sorted)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].StartTime < matches[j].StartTime
	})

	merged := []QuoteMatch{matches[0]}
	for i := 1; i < len(matches); i++ {
		last := &merged[len(merged)-1]
		cur := matches[i]

		if cur.StartTime <= last.EndTime {
			// Overlapping — merge time ranges
			if cur.EndTime > last.EndTime {
				last.EndTime = cur.EndTime
			}
			// Keep the higher-confidence quote as primary; demote the other
			if cur.Confidence > last.Confidence {
				last.MergedQuotes = append(last.MergedQuotes, last.Quote)
				last.Quote = cur.Quote
				last.Confidence = cur.Confidence
			} else {
				last.MergedQuotes = append(last.MergedQuotes, cur.Quote)
			}
			// Merge SRT entries (deduplicate by index)
			seen := make(map[int]bool)
			for _, e := range last.SRTEntries {
				seen[e.Index] = true
			}
			for _, e := range cur.SRTEntries {
				if !seen[e.Index] {
					last.SRTEntries = append(last.SRTEntries, e)
					seen[e.Index] = true
				}
			}
			last.MatchedText += " | " + cur.MatchedText
		} else {
			merged = append(merged, cur)
		}
	}

	return merged
}
