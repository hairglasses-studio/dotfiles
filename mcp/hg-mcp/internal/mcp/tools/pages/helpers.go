package pages

import (
	"strings"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewNotionClient)

// getDatabaseID reads the HG_PAGES_DATABASE_ID env var
func getDatabaseID() (string, error) {
	return config.GetEnvRequired("HG_PAGES_DATABASE_ID")
}

// --- Property builders ---

func buildTitleProperty(title string) map[string]interface{} {
	return map[string]interface{}{
		"title": []map[string]interface{}{
			{"text": map[string]string{"content": title}},
		},
	}
}

func buildMultiSelectProperty(values []string) map[string]interface{} {
	options := make([]map[string]string, len(values))
	for i, v := range values {
		options[i] = map[string]string{"name": v}
	}
	return map[string]interface{}{
		"multi_select": options,
	}
}

func buildSelectProperty(value string) map[string]interface{} {
	return map[string]interface{}{
		"select": map[string]string{"name": value},
	}
}

func buildCheckboxProperty(checked bool) map[string]interface{} {
	return map[string]interface{}{
		"checkbox": checked,
	}
}

// --- Property extractors ---

func extractTags(properties map[string]interface{}) []string {
	prop, ok := properties["Tags"]
	if !ok {
		return nil
	}
	propMap, ok := prop.(map[string]interface{})
	if !ok {
		return nil
	}
	arr, ok := propMap["multi_select"].([]interface{})
	if !ok {
		return nil
	}
	tags := make([]string, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			if name, ok := m["name"].(string); ok {
				tags = append(tags, name)
			}
		}
	}
	return tags
}

func extractSelect(properties map[string]interface{}, key string) string {
	prop, ok := properties[key]
	if !ok {
		return ""
	}
	propMap, ok := prop.(map[string]interface{})
	if !ok {
		return ""
	}
	sel, ok := propMap["select"].(map[string]interface{})
	if !ok {
		return ""
	}
	name, _ := sel["name"].(string)
	return name
}

func extractCheckbox(properties map[string]interface{}, key string) bool {
	prop, ok := properties[key]
	if !ok {
		return false
	}
	propMap, ok := prop.(map[string]interface{})
	if !ok {
		return false
	}
	checked, _ := propMap["checkbox"].(bool)
	return checked
}

// --- Filter builder ---

func buildCompoundFilter(filters []map[string]interface{}) map[string]interface{} {
	if len(filters) == 0 {
		return nil
	}
	if len(filters) == 1 {
		return filters[0]
	}
	return map[string]interface{}{
		"and": filters,
	}
}

// --- Output formatting ---

type pageSummary struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Category       string   `json:"category,omitempty"`
	Status         string   `json:"status,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	URL            string   `json:"url"`
	CreatedTime    string   `json:"created_time"`
	LastEditedTime string   `json:"last_edited_time"`
	Archived       bool     `json:"archived"`
}

func formatPageSummary(page clients.NotionPage) pageSummary {
	return pageSummary{
		ID:             page.ID,
		Title:          page.Title,
		Category:       extractSelect(page.Properties, "Category"),
		Status:         extractSelect(page.Properties, "Status"),
		Tags:           extractTags(page.Properties),
		URL:            page.URL,
		CreatedTime:    page.CreatedTime.Format("2006-01-02T15:04:05Z"),
		LastEditedTime: page.LastEditedTime.Format("2006-01-02T15:04:05Z"),
		Archived:       page.Archived,
	}
}

// --- Markdown conversion ---

func blocksToMarkdown(blocks []clients.NotionBlock) string {
	var sb strings.Builder
	for _, block := range blocks {
		text := extractBlockText(block)
		switch block.Type {
		case "heading_1":
			sb.WriteString("# " + text + "\n\n")
		case "heading_2":
			sb.WriteString("## " + text + "\n\n")
		case "heading_3":
			sb.WriteString("### " + text + "\n\n")
		case "paragraph":
			sb.WriteString(text + "\n\n")
		case "bulleted_list_item":
			sb.WriteString("- " + text + "\n")
		case "numbered_list_item":
			sb.WriteString("1. " + text + "\n")
		case "to_do":
			checked := false
			if block.Content != nil {
				if c, ok := block.Content["checked"].(bool); ok {
					checked = c
				}
			}
			if checked {
				sb.WriteString("- [x] " + text + "\n")
			} else {
				sb.WriteString("- [ ] " + text + "\n")
			}
		case "quote":
			sb.WriteString("> " + text + "\n\n")
		case "code":
			lang := ""
			if block.Content != nil {
				if l, ok := block.Content["language"].(string); ok {
					lang = l
				}
			}
			sb.WriteString("```" + lang + "\n" + text + "\n```\n\n")
		case "divider":
			sb.WriteString("---\n\n")
		case "callout":
			icon := extractCalloutIcon(block)
			if icon != "" {
				sb.WriteString("> " + icon + " " + text + "\n\n")
			} else {
				sb.WriteString("> " + text + "\n\n")
			}
		case "toggle":
			sb.WriteString("**" + text + "**\n\n")
		case "image":
			url := extractImageURL(block)
			if url != "" {
				sb.WriteString("![image](" + url + ")\n\n")
			}
		case "bookmark":
			url := extractBookmarkURL(block)
			if url != "" {
				sb.WriteString("[Bookmark](" + url + ")\n\n")
			}
		case "equation":
			expr := extractEquation(block)
			if expr != "" {
				sb.WriteString("$$" + expr + "$$\n\n")
			}
		default:
			if text != "" {
				sb.WriteString(text + "\n\n")
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

func extractImageURL(block clients.NotionBlock) string {
	if block.Content == nil {
		return ""
	}
	// Try external URL first
	if ext, ok := block.Content["external"].(map[string]interface{}); ok {
		if url, ok := ext["url"].(string); ok {
			return url
		}
	}
	// Try Notion-hosted file
	if file, ok := block.Content["file"].(map[string]interface{}); ok {
		if url, ok := file["url"].(string); ok {
			return url
		}
	}
	return ""
}

func extractBookmarkURL(block clients.NotionBlock) string {
	if block.Content == nil {
		return ""
	}
	if url, ok := block.Content["url"].(string); ok {
		return url
	}
	return ""
}

func extractCalloutIcon(block clients.NotionBlock) string {
	if block.Content == nil {
		return ""
	}
	if icon, ok := block.Content["icon"].(map[string]interface{}); ok {
		if emoji, ok := icon["emoji"].(string); ok {
			return emoji
		}
	}
	return ""
}

func extractEquation(block clients.NotionBlock) string {
	if block.Content == nil {
		return ""
	}
	if expr, ok := block.Content["expression"].(string); ok {
		return expr
	}
	return ""
}

func extractBlockText(block clients.NotionBlock) string {
	if block.Content == nil {
		return ""
	}
	richText, ok := block.Content["rich_text"].([]interface{})
	if !ok {
		return ""
	}
	var sb strings.Builder
	for _, rt := range richText {
		m, ok := rt.(map[string]interface{})
		if !ok {
			continue
		}
		text, _ := m["plain_text"].(string)
		if text == "" {
			continue
		}

		// Check for link (href on the rich text object itself)
		href, _ := m["href"].(string)

		// Check annotations
		if ann, ok := m["annotations"].(map[string]interface{}); ok {
			if code, _ := ann["code"].(bool); code {
				text = "`" + text + "`"
			}
			if bold, _ := ann["bold"].(bool); bold {
				text = "**" + text + "**"
			}
			if italic, _ := ann["italic"].(bool); italic {
				text = "*" + text + "*"
			}
			if strike, _ := ann["strikethrough"].(bool); strike {
				text = "~~" + text + "~~"
			}
		}

		if href != "" {
			text = "[" + text + "](" + href + ")"
		}

		sb.WriteString(text)
	}
	return sb.String()
}
