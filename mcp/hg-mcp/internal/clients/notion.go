// Package clients provides client implementations for external services.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

const (
	notionAPIVersion = "2022-06-28"
	notionBaseURL    = "https://api.notion.com/v1"
)

// NotionClient provides Notion API operations
type NotionClient struct {
	apiKey     string
	httpClient *http.Client
	mu         sync.RWMutex
}

// NotionPage represents a Notion page
type NotionPage struct {
	ID             string                 `json:"id"`
	CreatedTime    time.Time              `json:"created_time"`
	LastEditedTime time.Time              `json:"last_edited_time"`
	CreatedBy      NotionUser             `json:"created_by"`
	LastEditedBy   NotionUser             `json:"last_edited_by"`
	Archived       bool                   `json:"archived"`
	Icon           *NotionIcon            `json:"icon,omitempty"`
	Cover          *NotionFile            `json:"cover,omitempty"`
	URL            string                 `json:"url"`
	Title          string                 `json:"title,omitempty"`
	Properties     map[string]interface{} `json:"properties,omitempty"`
	Parent         NotionParent           `json:"parent"`
}

// NotionDatabase represents a Notion database
type NotionDatabase struct {
	ID             string                 `json:"id"`
	CreatedTime    time.Time              `json:"created_time"`
	LastEditedTime time.Time              `json:"last_edited_time"`
	Title          []NotionRichText       `json:"title"`
	Description    []NotionRichText       `json:"description,omitempty"`
	Icon           *NotionIcon            `json:"icon,omitempty"`
	Cover          *NotionFile            `json:"cover,omitempty"`
	Properties     map[string]interface{} `json:"properties,omitempty"`
	URL            string                 `json:"url"`
	Archived       bool                   `json:"archived"`
}

// NotionUser represents a Notion user
type NotionUser struct {
	ID        string `json:"id"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Type      string `json:"type,omitempty"`
	Email     string `json:"email,omitempty"`
}

// NotionIcon represents page/database icon
type NotionIcon struct {
	Type     string `json:"type"`
	Emoji    string `json:"emoji,omitempty"`
	External *struct {
		URL string `json:"url"`
	} `json:"external,omitempty"`
}

// NotionFile represents a file/image
type NotionFile struct {
	Type     string `json:"type"`
	External *struct {
		URL string `json:"url"`
	} `json:"external,omitempty"`
}

// NotionParent represents page parent
type NotionParent struct {
	Type       string `json:"type"`
	PageID     string `json:"page_id,omitempty"`
	DatabaseID string `json:"database_id,omitempty"`
	Workspace  bool   `json:"workspace,omitempty"`
}

// NotionRichText represents rich text content
type NotionRichText struct {
	Type        string             `json:"type"`
	Text        *NotionTextContent `json:"text,omitempty"`
	PlainText   string             `json:"plain_text"`
	Annotations *NotionAnnotations `json:"annotations,omitempty"`
}

// NotionTextContent represents text within rich text
type NotionTextContent struct {
	Content string      `json:"content"`
	Link    *NotionLink `json:"link,omitempty"`
}

// NotionLink represents a link
type NotionLink struct {
	URL string `json:"url"`
}

// NotionAnnotations represents text formatting
type NotionAnnotations struct {
	Bold          bool   `json:"bold"`
	Italic        bool   `json:"italic"`
	Strikethrough bool   `json:"strikethrough"`
	Underline     bool   `json:"underline"`
	Code          bool   `json:"code"`
	Color         string `json:"color"`
}

// NotionBlock represents a block in a page
type NotionBlock struct {
	ID             string                 `json:"id"`
	Type           string                 `json:"type"`
	CreatedTime    time.Time              `json:"created_time"`
	LastEditedTime time.Time              `json:"last_edited_time"`
	HasChildren    bool                   `json:"has_children"`
	Archived       bool                   `json:"archived"`
	Content        map[string]interface{} `json:"content,omitempty"`
}

// NotionSearchResult represents a search result
type NotionSearchResult struct {
	Object     string        `json:"object"`
	Results    []interface{} `json:"results"`
	HasMore    bool          `json:"has_more"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// NotionDatabaseQuery represents query parameters
type NotionDatabaseQuery struct {
	Filter      map[string]interface{} `json:"filter,omitempty"`
	Sorts       []NotionSort           `json:"sorts,omitempty"`
	StartCursor string                 `json:"start_cursor,omitempty"`
	PageSize    int                    `json:"page_size,omitempty"`
}

// NotionSort represents sort configuration
type NotionSort struct {
	Property  string `json:"property,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Direction string `json:"direction"`
}

var (
	notionClient     *NotionClient
	notionClientOnce sync.Once
	notionClientErr  error

	// TestOverrideNotionClient, when non-nil, is returned by GetNotionClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideNotionClient *NotionClient
)

// GetNotionClient returns the singleton Notion client
func GetNotionClient() (*NotionClient, error) {
	if TestOverrideNotionClient != nil {
		return TestOverrideNotionClient, nil
	}
	notionClientOnce.Do(func() {
		notionClient, notionClientErr = NewNotionClient()
	})
	return notionClient, notionClientErr
}

// NewTestNotionClient creates an in-memory test client.
// All API calls will fail unless methods are overridden via embedding.
func NewTestNotionClient() *NotionClient {
	return &NotionClient{
		apiKey:     "test-api-key",
		httpClient: httpclient.Fast(),
	}
}

// NewNotionClient creates a new Notion client
func NewNotionClient() (*NotionClient, error) {
	apiKey := os.Getenv("NOTION_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("NOTION_TOKEN")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no Notion API key configured (set NOTION_API_KEY or NOTION_TOKEN)")
	}

	return &NotionClient{
		apiKey:     apiKey,
		httpClient: httpclient.Standard(),
	}, nil
}

// IsConfigured returns true if the client is properly configured
func (c *NotionClient) IsConfigured() bool {
	return c != nil && c.apiKey != ""
}

func (c *NotionClient) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, notionBaseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Notion-Version", notionAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Search searches across all pages and databases
func (c *NotionClient) Search(ctx context.Context, query string, filter string, pageSize int) (*NotionSearchResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if pageSize <= 0 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}

	body := map[string]interface{}{
		"query":     query,
		"page_size": pageSize,
	}

	if filter == "page" || filter == "database" {
		body["filter"] = map[string]string{"value": filter, "property": "object"}
	}

	respBody, err := c.doRequest(ctx, "POST", "/search", body)
	if err != nil {
		return nil, err
	}

	var result NotionSearchResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetPage retrieves a page by ID
func (c *NotionClient) GetPage(ctx context.Context, pageID string) (*NotionPage, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	respBody, err := c.doRequest(ctx, "GET", "/pages/"+pageID, nil)
	if err != nil {
		return nil, err
	}

	var page NotionPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}

	// Extract title from properties if available
	page.Title = extractPageTitle(page.Properties)

	return &page, nil
}

// GetDatabase retrieves a database by ID
func (c *NotionClient) GetDatabase(ctx context.Context, databaseID string) (*NotionDatabase, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	respBody, err := c.doRequest(ctx, "GET", "/databases/"+databaseID, nil)
	if err != nil {
		return nil, err
	}

	var db NotionDatabase
	if err := json.Unmarshal(respBody, &db); err != nil {
		return nil, fmt.Errorf("failed to parse database: %w", err)
	}

	return &db, nil
}

// QueryDatabase queries a database with filters and sorts
func (c *NotionClient) QueryDatabase(ctx context.Context, databaseID string, query *NotionDatabaseQuery) ([]NotionPage, bool, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if query == nil {
		query = &NotionDatabaseQuery{}
	}
	if query.PageSize <= 0 {
		query.PageSize = 100
	}

	respBody, err := c.doRequest(ctx, "POST", "/databases/"+databaseID+"/query", query)
	if err != nil {
		return nil, false, "", err
	}

	var result struct {
		Results    []json.RawMessage `json:"results"`
		HasMore    bool              `json:"has_more"`
		NextCursor string            `json:"next_cursor"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, false, "", fmt.Errorf("failed to parse response: %w", err)
	}

	pages := make([]NotionPage, 0, len(result.Results))
	for _, raw := range result.Results {
		var page NotionPage
		if err := json.Unmarshal(raw, &page); err == nil {
			page.Title = extractPageTitle(page.Properties)
			pages = append(pages, page)
		}
	}

	return pages, result.HasMore, result.NextCursor, nil
}

// GetBlockChildren gets child blocks of a block/page with automatic pagination.
// Fetches all pages of results up to a maximum of 500 blocks.
func (c *NotionClient) GetBlockChildren(ctx context.Context, blockID string, pageSize int) ([]NotionBlock, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if pageSize <= 0 {
		pageSize = 100
	}

	const maxBlocks = 500
	var allBlocks []NotionBlock
	cursor := ""

	for {
		path := fmt.Sprintf("/blocks/%s/children?page_size=%d", blockID, pageSize)
		if cursor != "" {
			path += "&start_cursor=" + cursor
		}

		respBody, err := c.doRequest(ctx, "GET", path, nil)
		if err != nil {
			return nil, err
		}

		var result struct {
			Results    []json.RawMessage `json:"results"`
			HasMore    bool              `json:"has_more"`
			NextCursor string            `json:"next_cursor"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, raw := range result.Results {
			var block NotionBlock
			if err := json.Unmarshal(raw, &block); err == nil {
				// Parse block-specific content
				var fullBlock map[string]interface{}
				json.Unmarshal(raw, &fullBlock)
				if content, ok := fullBlock[block.Type]; ok {
					if contentMap, ok := content.(map[string]interface{}); ok {
						block.Content = contentMap
					}
				}
				allBlocks = append(allBlocks, block)
			}
		}

		if !result.HasMore || result.NextCursor == "" || len(allBlocks) >= maxBlocks {
			break
		}
		cursor = result.NextCursor
	}

	return allBlocks, nil
}

// CreatePage creates a new page in a parent page or database
func (c *NotionClient) CreatePage(ctx context.Context, parentID string, isDatabase bool, title string, content []map[string]interface{}) (*NotionPage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]interface{}{}

	// Set parent
	if isDatabase {
		body["parent"] = map[string]string{"database_id": parentID}
		body["properties"] = map[string]interface{}{
			"title": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]string{"content": title}},
				},
			},
		}
	} else {
		body["parent"] = map[string]string{"page_id": parentID}
		body["properties"] = map[string]interface{}{
			"title": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]string{"content": title}},
				},
			},
		}
	}

	// Add content blocks if provided
	if len(content) > 0 {
		body["children"] = content
	}

	respBody, err := c.doRequest(ctx, "POST", "/pages", body)
	if err != nil {
		return nil, err
	}

	var page NotionPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}

	page.Title = title
	return &page, nil
}

// AppendBlocks appends blocks to a page
func (c *NotionClient) AppendBlocks(ctx context.Context, pageID string, blocks []map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]interface{}{
		"children": blocks,
	}

	_, err := c.doRequest(ctx, "PATCH", "/blocks/"+pageID+"/children", body)
	return err
}

// GetUser gets the current bot user info
func (c *NotionClient) GetUser(ctx context.Context) (*NotionUser, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	respBody, err := c.doRequest(ctx, "GET", "/users/me", nil)
	if err != nil {
		return nil, err
	}

	var user NotionUser
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}

	return &user, nil
}

// ListUsers lists all users in the workspace
func (c *NotionClient) ListUsers(ctx context.Context) ([]NotionUser, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	respBody, err := c.doRequest(ctx, "GET", "/users", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Results []NotionUser `json:"results"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse users: %w", err)
	}

	return result.Results, nil
}

// Helper to extract page title from properties
func extractPageTitle(properties map[string]interface{}) string {
	if properties == nil {
		return ""
	}

	// Try common title property names
	for _, key := range []string{"title", "Title", "Name", "name"} {
		if prop, ok := properties[key]; ok {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if titleArr, ok := propMap["title"].([]interface{}); ok && len(titleArr) > 0 {
					if first, ok := titleArr[0].(map[string]interface{}); ok {
						if plainText, ok := first["plain_text"].(string); ok {
							return plainText
						}
					}
				}
			}
		}
	}
	return ""
}

// UpdatePageProperties updates properties on an existing page
func (c *NotionClient) UpdatePageProperties(ctx context.Context, pageID string, properties map[string]interface{}) (*NotionPage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]interface{}{
		"properties": properties,
	}

	respBody, err := c.doRequest(ctx, "PATCH", "/pages/"+pageID, body)
	if err != nil {
		return nil, err
	}

	var page NotionPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}

	page.Title = extractPageTitle(page.Properties)
	return &page, nil
}

// ArchivePage soft-deletes a page by setting archived=true
func (c *NotionClient) ArchivePage(ctx context.Context, pageID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]interface{}{
		"archived": true,
	}

	_, err := c.doRequest(ctx, "PATCH", "/pages/"+pageID, body)
	return err
}

// CreatePageWithProperties creates a new page with full property control
func (c *NotionClient) CreatePageWithProperties(ctx context.Context, databaseID string, properties map[string]interface{}, children []map[string]interface{}) (*NotionPage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	body := map[string]interface{}{
		"parent":     map[string]string{"database_id": databaseID},
		"properties": properties,
	}

	if len(children) > 0 {
		body["children"] = children
	}

	respBody, err := c.doRequest(ctx, "POST", "/pages", body)
	if err != nil {
		return nil, err
	}

	var page NotionPage
	if err := json.Unmarshal(respBody, &page); err != nil {
		return nil, fmt.Errorf("failed to parse page: %w", err)
	}

	page.Title = extractPageTitle(page.Properties)
	return &page, nil
}

// Helper to create a text block
func CreateTextBlock(text string) map[string]interface{} {
	return map[string]interface{}{
		"object": "block",
		"type":   "paragraph",
		"paragraph": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"type": "text", "text": map[string]string{"content": text}},
			},
		},
	}
}

// Helper to create a heading block
func CreateHeadingBlock(text string, level int) map[string]interface{} {
	if level < 1 {
		level = 1
	}
	if level > 3 {
		level = 3
	}

	headingType := fmt.Sprintf("heading_%d", level)
	return map[string]interface{}{
		"object": "block",
		"type":   headingType,
		headingType: map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"type": "text", "text": map[string]string{"content": text}},
			},
		},
	}
}

// Helper to create a bullet list item
func CreateBulletBlock(text string) map[string]interface{} {
	return map[string]interface{}{
		"object": "block",
		"type":   "bulleted_list_item",
		"bulleted_list_item": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"type": "text", "text": map[string]string{"content": text}},
			},
		},
	}
}

// Helper to create a todo block
func CreateTodoBlock(text string, checked bool) map[string]interface{} {
	return map[string]interface{}{
		"object": "block",
		"type":   "to_do",
		"to_do": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{"type": "text", "text": map[string]string{"content": text}},
			},
			"checked": checked,
		},
	}
}
