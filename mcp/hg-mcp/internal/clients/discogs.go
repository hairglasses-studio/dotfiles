// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

const (
	discogsAPIBase = "https://api.discogs.com"
)

// DiscogsClient provides access to Discogs API
type DiscogsClient struct {
	token      string
	httpClient *http.Client
	mu         sync.RWMutex
	rateLimit  *discogsRateLimit
}

type discogsRateLimit struct {
	remaining int
	total     int
	resetAt   time.Time
}

// DiscogsArtist represents a Discogs artist
type DiscogsArtist struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	RealName    string          `json:"realname,omitempty"`
	Profile     string          `json:"profile,omitempty"`
	URLs        []string        `json:"urls,omitempty"`
	Images      []DiscogsImage  `json:"images,omitempty"`
	Members     []DiscogsMember `json:"members,omitempty"`
	Aliases     []DiscogsAlias  `json:"aliases,omitempty"`
	DataQuality string          `json:"data_quality,omitempty"`
	ResourceURL string          `json:"resource_url"`
	URI         string          `json:"uri"`
}

// DiscogsImage represents an image
type DiscogsImage struct {
	Type   string `json:"type"`
	URI    string `json:"uri"`
	URI150 string `json:"uri150"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// DiscogsMember represents a group member
type DiscogsMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Active      bool   `json:"active"`
	ResourceURL string `json:"resource_url"`
}

// DiscogsAlias represents an artist alias
type DiscogsAlias struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ResourceURL string `json:"resource_url"`
}

// DiscogsRelease represents a release (album, single, etc.)
type DiscogsRelease struct {
	ID          int                `json:"id"`
	Title       string             `json:"title"`
	Year        int                `json:"year,omitempty"`
	Country     string             `json:"country,omitempty"`
	Released    string             `json:"released,omitempty"`
	Notes       string             `json:"notes,omitempty"`
	Genres      []string           `json:"genres,omitempty"`
	Styles      []string           `json:"styles,omitempty"`
	Artists     []DiscogsArtistRef `json:"artists,omitempty"`
	Labels      []DiscogsLabelRef  `json:"labels,omitempty"`
	Formats     []DiscogsFormat    `json:"formats,omitempty"`
	Tracklist   []DiscogsTrack     `json:"tracklist,omitempty"`
	Images      []DiscogsImage     `json:"images,omitempty"`
	Thumb       string             `json:"thumb,omitempty"`
	CoverImage  string             `json:"cover_image,omitempty"`
	MasterID    int                `json:"master_id,omitempty"`
	MasterURL   string             `json:"master_url,omitempty"`
	ResourceURL string             `json:"resource_url"`
	URI         string             `json:"uri"`
	DataQuality string             `json:"data_quality,omitempty"`
	Community   *DiscogsCommunity  `json:"community,omitempty"`
	LowestPrice float64            `json:"lowest_price,omitempty"`
	NumForSale  int                `json:"num_for_sale,omitempty"`
}

// DiscogsArtistRef is a reference to an artist in a release
type DiscogsArtistRef struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ANV         string `json:"anv,omitempty"` // Artist name variation
	Join        string `json:"join,omitempty"`
	Role        string `json:"role,omitempty"`
	ResourceURL string `json:"resource_url"`
}

// DiscogsLabelRef is a reference to a label
type DiscogsLabelRef struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CatNo       string `json:"catno,omitempty"` // Catalog number
	ResourceURL string `json:"resource_url"`
}

// DiscogsFormat represents a release format
type DiscogsFormat struct {
	Name         string   `json:"name"`
	Qty          string   `json:"qty"`
	Descriptions []string `json:"descriptions,omitempty"`
	Text         string   `json:"text,omitempty"`
}

// DiscogsTrack represents a track in a release
type DiscogsTrack struct {
	Position string             `json:"position"`
	Title    string             `json:"title"`
	Duration string             `json:"duration,omitempty"`
	Artists  []DiscogsArtistRef `json:"artists,omitempty"`
}

// DiscogsCommunity represents community data
type DiscogsCommunity struct {
	Have   int `json:"have"`
	Want   int `json:"want"`
	Rating struct {
		Count   int     `json:"count"`
		Average float64 `json:"average"`
	} `json:"rating,omitempty"`
}

// DiscogsLabel represents a record label
type DiscogsLabel struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Profile     string            `json:"profile,omitempty"`
	ContactInfo string            `json:"contact_info,omitempty"`
	URLs        []string          `json:"urls,omitempty"`
	Images      []DiscogsImage    `json:"images,omitempty"`
	ParentLabel *DiscogsLabelRef  `json:"parent_label,omitempty"`
	SubLabels   []DiscogsLabelRef `json:"sublabels,omitempty"`
	DataQuality string            `json:"data_quality,omitempty"`
	ResourceURL string            `json:"resource_url"`
	URI         string            `json:"uri"`
}

// DiscogsMaster represents a master release
type DiscogsMaster struct {
	ID          int                `json:"id"`
	Title       string             `json:"title"`
	Year        int                `json:"year,omitempty"`
	Artists     []DiscogsArtistRef `json:"artists,omitempty"`
	Genres      []string           `json:"genres,omitempty"`
	Styles      []string           `json:"styles,omitempty"`
	Images      []DiscogsImage     `json:"images,omitempty"`
	Tracklist   []DiscogsTrack     `json:"tracklist,omitempty"`
	MainRelease int                `json:"main_release"`
	ResourceURL string             `json:"resource_url"`
	URI         string             `json:"uri"`
	NumForSale  int                `json:"num_for_sale,omitempty"`
	LowestPrice float64            `json:"lowest_price,omitempty"`
}

// DiscogsSearchResult represents search results
type DiscogsSearchResult struct {
	Results    []DiscogsSearchItem `json:"results"`
	Pagination DiscogsPagination   `json:"pagination"`
}

// DiscogsSearchItem represents a search result item
type DiscogsSearchItem struct {
	ID          int      `json:"id"`
	Type        string   `json:"type"` // release, master, artist, label
	Title       string   `json:"title"`
	Thumb       string   `json:"thumb,omitempty"`
	CoverImage  string   `json:"cover_image,omitempty"`
	Year        string   `json:"year,omitempty"`
	Country     string   `json:"country,omitempty"`
	Format      []string `json:"format,omitempty"`
	Label       []string `json:"label,omitempty"`
	Genre       []string `json:"genre,omitempty"`
	Style       []string `json:"style,omitempty"`
	Barcode     []string `json:"barcode,omitempty"`
	CatNo       string   `json:"catno,omitempty"`
	ResourceURL string   `json:"resource_url"`
	URI         string   `json:"uri"`
	Community   *struct {
		Have int `json:"have"`
		Want int `json:"want"`
	} `json:"community,omitempty"`
}

// DiscogsPagination represents API pagination
type DiscogsPagination struct {
	Page    int `json:"page"`
	Pages   int `json:"pages"`
	PerPage int `json:"per_page"`
	Items   int `json:"items"`
	URLs    struct {
		First string `json:"first,omitempty"`
		Last  string `json:"last,omitempty"`
		Prev  string `json:"prev,omitempty"`
		Next  string `json:"next,omitempty"`
	} `json:"urls"`
}

// DiscogsCollectionFolder represents a collection folder
type DiscogsCollectionFolder struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Count       int    `json:"count"`
	ResourceURL string `json:"resource_url"`
}

// DiscogsCollectionItem represents an item in collection
type DiscogsCollectionItem struct {
	ID         int            `json:"id"`
	InstanceID int            `json:"instance_id"`
	FolderID   int            `json:"folder_id"`
	Rating     int            `json:"rating"`
	DateAdded  string         `json:"date_added"`
	BasicInfo  DiscogsRelease `json:"basic_information"`
	Notes      []DiscogsNote  `json:"notes,omitempty"`
}

// DiscogsNote represents a note on a collection item
type DiscogsNote struct {
	FieldID int    `json:"field_id"`
	Value   string `json:"value"`
}

// DiscogsWantlistItem represents an item in wantlist
type DiscogsWantlistItem struct {
	ID          int            `json:"id"`
	Rating      int            `json:"rating"`
	DateAdded   string         `json:"date_added"`
	BasicInfo   DiscogsRelease `json:"basic_information"`
	Notes       string         `json:"notes,omitempty"`
	ResourceURL string         `json:"resource_url"`
}

// DiscogsMarketplaceListing represents a marketplace listing
type DiscogsMarketplaceListing struct {
	ID              int            `json:"id"`
	Status          string         `json:"status"`
	Condition       string         `json:"condition"`
	SleeveCondition string         `json:"sleeve_condition,omitempty"`
	Price           DiscogsPrice   `json:"price"`
	OriginalPrice   DiscogsPrice   `json:"original_price,omitempty"`
	ShippingPrice   DiscogsPrice   `json:"shipping_price,omitempty"`
	AllowOffers     bool           `json:"allow_offers"`
	Comments        string         `json:"comments,omitempty"`
	Seller          DiscogsSeller  `json:"seller"`
	Release         DiscogsRelease `json:"release"`
	ResourceURL     string         `json:"resource_url"`
	URI             string         `json:"uri"`
	Posted          string         `json:"posted,omitempty"`
}

// DiscogsPrice represents a price
type DiscogsPrice struct {
	Value    float64 `json:"value"`
	Currency string  `json:"currency"`
}

// DiscogsSeller represents a seller
type DiscogsSeller struct {
	ID          int     `json:"id"`
	Username    string  `json:"username"`
	Rating      float64 `json:"rating"`
	NumRatings  int     `json:"num_ratings"`
	ResourceURL string  `json:"resource_url"`
}

// DiscogsStatus represents connection status
type DiscogsStatus struct {
	Connected     bool   `json:"connected"`
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username,omitempty"`
	RateLimit     struct {
		Remaining int    `json:"remaining"`
		Total     int    `json:"total"`
		ResetAt   string `json:"reset_at,omitempty"`
	} `json:"rate_limit"`
}

// DiscogsHealth represents health status
type DiscogsHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	Authenticated   bool     `json:"authenticated"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewDiscogsClient creates a new Discogs client
func NewDiscogsClient() (*DiscogsClient, error) {
	token := os.Getenv("DISCOGS_TOKEN")

	return &DiscogsClient{
		token: token,
		httpClient: httpclient.Standard(),
		rateLimit: &discogsRateLimit{
			remaining: 60,
			total:     60,
		},
	}, nil
}

// GetStatus returns the client status
func (c *DiscogsClient) GetStatus(ctx context.Context) (*DiscogsStatus, error) {
	status := &DiscogsStatus{
		Authenticated: c.token != "",
	}

	// Test API connectivity with identity endpoint
	if c.token != "" {
		identity, err := c.getIdentity(ctx)
		if err == nil && identity != nil {
			status.Connected = true
			status.Username = identity.Username
		}
	} else {
		// Test without auth
		_, err := c.doRequest(ctx, discogsAPIBase+"/database/search?q=test&per_page=1")
		status.Connected = err == nil
	}

	c.mu.RLock()
	status.RateLimit.Remaining = c.rateLimit.remaining
	status.RateLimit.Total = c.rateLimit.total
	if !c.rateLimit.resetAt.IsZero() {
		status.RateLimit.ResetAt = c.rateLimit.resetAt.Format(time.RFC3339)
	}
	c.mu.RUnlock()

	return status, nil
}

type discogsIdentity struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	ResourceURL  string `json:"resource_url"`
	ConsumerName string `json:"consumer_name"`
}

func (c *DiscogsClient) getIdentity(ctx context.Context) (*discogsIdentity, error) {
	body, err := c.doRequest(ctx, discogsAPIBase+"/oauth/identity")
	if err != nil {
		return nil, err
	}

	var identity discogsIdentity
	if err := json.Unmarshal(body, &identity); err != nil {
		return nil, err
	}

	return &identity, nil
}

// GetHealth returns health status with recommendations
func (c *DiscogsClient) GetHealth(ctx context.Context) (*DiscogsHealth, error) {
	status, _ := c.GetStatus(ctx)

	health := &DiscogsHealth{
		Score:         100,
		Status:        "healthy",
		Connected:     status.Connected,
		Authenticated: status.Authenticated,
	}

	if !status.Connected {
		health.Score -= 50
		health.Issues = append(health.Issues, "Cannot connect to Discogs API")
		health.Recommendations = append(health.Recommendations, "Check internet connection")
	}

	if !status.Authenticated {
		health.Score -= 20
		health.Issues = append(health.Issues, "Not authenticated - some features limited")
		health.Recommendations = append(health.Recommendations, "Set DISCOGS_TOKEN for full access")
		health.Recommendations = append(health.Recommendations, "Get token at: https://www.discogs.com/settings/developers")
	}

	if status.RateLimit.Remaining < 10 {
		health.Score -= 10
		health.Issues = append(health.Issues, fmt.Sprintf("Low rate limit: %d remaining", status.RateLimit.Remaining))
		health.Recommendations = append(health.Recommendations, "Wait for rate limit reset")
	}

	if health.Score >= 80 {
		health.Status = "healthy"
	} else if health.Score >= 50 {
		health.Status = "degraded"
	} else {
		health.Status = "critical"
	}

	return health, nil
}

// Search searches Discogs database
func (c *DiscogsClient) Search(ctx context.Context, query string, searchType string, page int) (*DiscogsSearchResult, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("per_page", "20")
	params.Set("page", strconv.Itoa(page))

	if searchType != "" && searchType != "all" {
		params.Set("type", searchType)
	}

	searchURL := fmt.Sprintf("%s/database/search?%s", discogsAPIBase, params.Encode())

	body, err := c.doRequest(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	var result DiscogsSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	return &result, nil
}

// GetRelease gets release details by ID
func (c *DiscogsClient) GetRelease(ctx context.Context, releaseID int) (*DiscogsRelease, error) {
	releaseURL := fmt.Sprintf("%s/releases/%d", discogsAPIBase, releaseID)

	body, err := c.doRequest(ctx, releaseURL)
	if err != nil {
		return nil, err
	}

	var release DiscogsRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

// GetMaster gets master release details by ID
func (c *DiscogsClient) GetMaster(ctx context.Context, masterID int) (*DiscogsMaster, error) {
	masterURL := fmt.Sprintf("%s/masters/%d", discogsAPIBase, masterID)

	body, err := c.doRequest(ctx, masterURL)
	if err != nil {
		return nil, err
	}

	var master DiscogsMaster
	if err := json.Unmarshal(body, &master); err != nil {
		return nil, fmt.Errorf("failed to parse master: %w", err)
	}

	return &master, nil
}

// GetArtist gets artist details by ID
func (c *DiscogsClient) GetArtist(ctx context.Context, artistID int) (*DiscogsArtist, error) {
	artistURL := fmt.Sprintf("%s/artists/%d", discogsAPIBase, artistID)

	body, err := c.doRequest(ctx, artistURL)
	if err != nil {
		return nil, err
	}

	var artist DiscogsArtist
	if err := json.Unmarshal(body, &artist); err != nil {
		return nil, fmt.Errorf("failed to parse artist: %w", err)
	}

	return &artist, nil
}

// GetArtistReleases gets an artist's releases
func (c *DiscogsClient) GetArtistReleases(ctx context.Context, artistID int, page int) (*DiscogsSearchResult, error) {
	releasesURL := fmt.Sprintf("%s/artists/%d/releases?per_page=20&page=%d", discogsAPIBase, artistID, page)

	body, err := c.doRequest(ctx, releasesURL)
	if err != nil {
		return nil, err
	}

	var result DiscogsSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	return &result, nil
}

// GetLabel gets label details by ID
func (c *DiscogsClient) GetLabel(ctx context.Context, labelID int) (*DiscogsLabel, error) {
	labelURL := fmt.Sprintf("%s/labels/%d", discogsAPIBase, labelID)

	body, err := c.doRequest(ctx, labelURL)
	if err != nil {
		return nil, err
	}

	var label DiscogsLabel
	if err := json.Unmarshal(body, &label); err != nil {
		return nil, fmt.Errorf("failed to parse label: %w", err)
	}

	return &label, nil
}

// GetLabelReleases gets a label's releases
func (c *DiscogsClient) GetLabelReleases(ctx context.Context, labelID int, page int) (*DiscogsSearchResult, error) {
	releasesURL := fmt.Sprintf("%s/labels/%d/releases?per_page=20&page=%d", discogsAPIBase, labelID, page)

	body, err := c.doRequest(ctx, releasesURL)
	if err != nil {
		return nil, err
	}

	var result DiscogsSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}

	return &result, nil
}

// GetCollection gets user's collection (requires auth)
func (c *DiscogsClient) GetCollection(ctx context.Context, username string, folderID int, page int) ([]DiscogsCollectionItem, *DiscogsPagination, error) {
	if username == "" {
		identity, err := c.getIdentity(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("username required or authenticate with token")
		}
		username = identity.Username
	}

	collectionURL := fmt.Sprintf("%s/users/%s/collection/folders/%d/releases?per_page=50&page=%d",
		discogsAPIBase, username, folderID, page)

	body, err := c.doRequest(ctx, collectionURL)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Releases   []DiscogsCollectionItem `json:"releases"`
		Pagination DiscogsPagination       `json:"pagination"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse collection: %w", err)
	}

	return result.Releases, &result.Pagination, nil
}

// GetCollectionFolders gets user's collection folders
func (c *DiscogsClient) GetCollectionFolders(ctx context.Context, username string) ([]DiscogsCollectionFolder, error) {
	if username == "" {
		identity, err := c.getIdentity(ctx)
		if err != nil {
			return nil, fmt.Errorf("username required or authenticate with token")
		}
		username = identity.Username
	}

	foldersURL := fmt.Sprintf("%s/users/%s/collection/folders", discogsAPIBase, username)

	body, err := c.doRequest(ctx, foldersURL)
	if err != nil {
		return nil, err
	}

	var result struct {
		Folders []DiscogsCollectionFolder `json:"folders"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse folders: %w", err)
	}

	return result.Folders, nil
}

// GetWantlist gets user's wantlist
func (c *DiscogsClient) GetWantlist(ctx context.Context, username string, page int) ([]DiscogsWantlistItem, *DiscogsPagination, error) {
	if username == "" {
		identity, err := c.getIdentity(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("username required or authenticate with token")
		}
		username = identity.Username
	}

	wantlistURL := fmt.Sprintf("%s/users/%s/wants?per_page=50&page=%d", discogsAPIBase, username, page)

	body, err := c.doRequest(ctx, wantlistURL)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Wants      []DiscogsWantlistItem `json:"wants"`
		Pagination DiscogsPagination     `json:"pagination"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse wantlist: %w", err)
	}

	return result.Wants, &result.Pagination, nil
}

// GetMarketplaceListing gets a marketplace listing
func (c *DiscogsClient) GetMarketplaceListing(ctx context.Context, listingID int) (*DiscogsMarketplaceListing, error) {
	listingURL := fmt.Sprintf("%s/marketplace/listings/%d", discogsAPIBase, listingID)

	body, err := c.doRequest(ctx, listingURL)
	if err != nil {
		return nil, err
	}

	var listing DiscogsMarketplaceListing
	if err := json.Unmarshal(body, &listing); err != nil {
		return nil, fmt.Errorf("failed to parse listing: %w", err)
	}

	return &listing, nil
}

// SearchMarketplace searches the marketplace for a release
func (c *DiscogsClient) SearchMarketplace(ctx context.Context, releaseID int, currency string) ([]DiscogsMarketplaceListing, error) {
	if currency == "" {
		currency = "USD"
	}

	marketURL := fmt.Sprintf("%s/marketplace/listings?release_id=%d&currency=%s",
		discogsAPIBase, releaseID, currency)

	body, err := c.doRequest(ctx, marketURL)
	if err != nil {
		return nil, err
	}

	var result struct {
		Listings []DiscogsMarketplaceListing `json:"listings"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse marketplace listings: %w", err)
	}

	return result.Listings, nil
}

// doRequest performs an HTTP request with auth and rate limiting
func (c *DiscogsClient) doRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "AFTRS-MCP/1.0 +https://github.com/hairglasses-studio/hg-mcp")
	req.Header.Set("Accept", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Discogs token="+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit info from headers
	c.mu.Lock()
	if remaining := resp.Header.Get("X-Discogs-Ratelimit-Remaining"); remaining != "" {
		c.rateLimit.remaining, _ = strconv.Atoi(remaining)
	}
	if total := resp.Header.Get("X-Discogs-Ratelimit"); total != "" {
		c.rateLimit.total, _ = strconv.Atoi(total)
	}
	c.mu.Unlock()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded - wait and retry")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// ParseReleaseURL extracts release ID from a Discogs URL
func ParseReleaseURL(discogsURL string) (int, error) {
	// Format: https://www.discogs.com/release/123456-Artist-Title
	parts := strings.Split(discogsURL, "/release/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid Discogs release URL")
	}

	idPart := strings.Split(parts[1], "-")[0]
	return strconv.Atoi(idPart)
}
