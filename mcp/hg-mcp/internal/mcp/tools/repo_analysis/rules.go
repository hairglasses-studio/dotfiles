package repo_analysis

import "strings"

// TagMatch represents a matched import rule with its metadata.
type TagMatch struct {
	Name       string `json:"name"`
	Import     string `json:"import"`
	Confidence string `json:"confidence"` // "high" or "medium"
}

// applyRules matches a list of Go module dependency paths against the import
// rules and returns categorized tag matches.
func applyRules(depPaths []string) map[string][]TagMatch {
	result := map[string][]TagMatch{
		"frameworks":    {},
		"protocols":     {},
		"datastores":    {},
		"cloud":         {},
		"ai":            {},
		"observability": {},
	}

	// Build a set of already-seen (category, tag) pairs to avoid duplicates.
	seen := make(map[string]bool)

	for _, rule := range importRules {
		for _, dep := range depPaths {
			if strings.HasPrefix(dep, rule.Pattern) {
				key := rule.Category + ":" + rule.Tag
				if seen[key] {
					continue
				}
				seen[key] = true

				match := TagMatch{
					Name:       rule.Tag,
					Import:     dep,
					Confidence: confidence(dep, rule.Pattern),
				}

				bucket := categoryBucket(rule.Category)
				result[bucket] = append(result[bucket], match)
			}
		}
	}

	return result
}

// collectTags returns a deduplicated slice of all tag names from the matches.
func collectTags(matches map[string][]TagMatch) []string {
	seen := make(map[string]bool)
	var tags []string
	for _, bucket := range matches {
		for _, m := range bucket {
			if !seen[m.Name] {
				seen[m.Name] = true
				tags = append(tags, m.Name)
			}
		}
	}
	return tags
}

// confidence returns "high" if the dep exactly matches the pattern, or "medium"
// if it is a sub-path match (e.g., "github.com/aws/aws-sdk-go-v2/service/s3"
// matching "github.com/aws/aws-sdk-go").
func confidence(dep, pattern string) string {
	if dep == pattern {
		return "high"
	}
	return "medium"
}

// categoryBucket maps a rule category to its output bucket name.
func categoryBucket(category string) string {
	switch category {
	case "framework":
		return "frameworks"
	case "protocol":
		return "protocols"
	case "datastore":
		return "datastores"
	case "cloud":
		return "cloud"
	case "ai":
		return "ai"
	case "observability":
		return "observability"
	default:
		return category
	}
}
