// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	ebayProductionAPIBase = "https://api.ebay.com"
	ebayProductionAuthURL = "https://api.ebay.com/identity/v1/oauth2/token"
	ebaySandboxAPIBase    = "https://api.sandbox.ebay.com"
	ebaySandboxAuthURL    = "https://api.sandbox.ebay.com/identity/v1/oauth2/token"
	ebaySecretsName = "aftrs/inventory/ebay-credentials"

	// OAuth scopes
	ebayInventoryScope = "https://api.ebay.com/oauth/api_scope https://api.ebay.com/oauth/api_scope/sell.inventory https://api.ebay.com/oauth/api_scope/sell.inventory.readonly https://api.ebay.com/oauth/api_scope/sell.marketing.readonly"
)

// EbayCredentials holds eBay API credentials from Secrets Manager
type EbayCredentials struct {
	AppID             string `json:"app_id"`
	DevID             string `json:"dev_id"`
	CertID            string `json:"cert_id"`
	SandboxAppID      string `json:"sandbox_app_id"`
	OAuthRefreshToken string `json:"oauth_refresh_token"`
	Environment       string `json:"environment"` // "production" or "sandbox"
}

// EbayClient provides access to eBay APIs
type EbayClient struct {
	httpClient    *http.Client
	credentials   *EbayCredentials
	accessToken   string
	tokenExpiry   time.Time
	secretsClient *secretsmanager.Client
	useSandbox    bool
	mu            sync.RWMutex
}

// EbayTokenResponse represents OAuth token response
type EbayTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// EbaySearchResult represents a search result from Browse API
type EbaySearchResult struct {
	ItemID         string    `json:"item_id"`
	Title          string    `json:"title"`
	Price          float64   `json:"price"`
	Currency       string    `json:"currency"`
	ListingType    string    `json:"listing_type"`
	Condition      string    `json:"condition"`
	Location       string    `json:"location"`
	ShippingCost   float64   `json:"shipping_cost,omitempty"`
	ImageURL       string    `json:"image_url,omitempty"`
	ItemURL        string    `json:"item_url"`
	EndTime        time.Time `json:"end_time,omitempty"`
	BidCount       int       `json:"bid_count,omitempty"`
	SellerUsername string    `json:"seller_username,omitempty"`
	SellerFeedback int       `json:"seller_feedback,omitempty"`
}

// EbayPriceAnalysis represents price analysis from search results
type EbayPriceAnalysis struct {
	Query         string             `json:"query"`
	ResultCount   int                `json:"result_count"`
	AveragePrice  float64            `json:"average_price"`
	MedianPrice   float64            `json:"median_price"`
	MinPrice      float64            `json:"min_price"`
	MaxPrice      float64            `json:"max_price"`
	SuggestedLow  float64            `json:"suggested_low"`
	SuggestedHigh float64            `json:"suggested_high"`
	Results       []EbaySearchResult `json:"results,omitempty"`
}

// EbayInventoryItem represents an item in eBay inventory
type EbayInventoryItem struct {
	SKU                  string            `json:"sku"`
	Availability         *EbayAvailability `json:"availability,omitempty"`
	Condition            string            `json:"condition"`
	ConditionDescription string            `json:"conditionDescription,omitempty"`
	Product              *EbayProduct      `json:"product"`
}

// EbayAvailability represents item availability
type EbayAvailability struct {
	ShipToLocationAvailability *EbayShipToLocation `json:"shipToLocationAvailability,omitempty"`
}

// EbayShipToLocation represents shipping availability
type EbayShipToLocation struct {
	Quantity int `json:"quantity"`
}

// EbayProduct represents product details for listing
type EbayProduct struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	ImageUrls   []string            `json:"imageUrls,omitempty"`
	Aspects     map[string][]string `json:"aspects,omitempty"`
}

// EbayOffer represents a listing offer
type EbayOffer struct {
	OfferID             string        `json:"offerId,omitempty"`
	SKU                 string        `json:"sku"`
	MarketplaceID       string        `json:"marketplaceId"`
	Format              string        `json:"format"` // FIXED_PRICE, AUCTION
	ListingDuration     string        `json:"listingDuration,omitempty"`
	PricingSummary      *EbayPricing  `json:"pricingSummary"`
	ListingPolicies     *EbayPolicies `json:"listingPolicies,omitempty"`
	CategoryID          string        `json:"categoryId"`
	MerchantLocationKey string        `json:"merchantLocationKey,omitempty"`
	AvailableQuantity   int           `json:"availableQuantity,omitempty"`
}

// EbayPricing represents pricing for an offer
type EbayPricing struct {
	Price *EbayAmount `json:"price"`
}

// EbayAmount represents a monetary amount
type EbayAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// EbayPolicies represents listing policies
type EbayPolicies struct {
	FulfillmentPolicyID string `json:"fulfillmentPolicyId,omitempty"`
	PaymentPolicyID     string `json:"paymentPolicyId,omitempty"`
	ReturnPolicyID      string `json:"returnPolicyId,omitempty"`
}

// EbayCategory represents a category suggestion
type EbayCategory struct {
	CategoryID       string `json:"categoryId"`
	CategoryName     string `json:"categoryName"`
	CategoryTreePath string `json:"categoryTreeNodePath,omitempty"`
}

var (
	ebayClient     *EbayClient
	ebayClientOnce sync.Once
	ebayClientErr  error
)

// GetEbayClient returns the singleton eBay client
func GetEbayClient() (*EbayClient, error) {
	ebayClientOnce.Do(func() {
		ebayClient, ebayClientErr = NewEbayClient()
	})
	return ebayClient, ebayClientErr
}

// NewEbayClient creates a new eBay client
func NewEbayClient() (*EbayClient, error) {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := &EbayClient{
		httpClient: httpclient.Standard(),
		secretsClient: secretsmanager.NewFromConfig(cfg),
		useSandbox:    os.Getenv("EBAY_USE_SANDBOX") == "true",
	}

	// Load credentials from Secrets Manager
	if err := client.loadCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to load eBay credentials: %w", err)
	}

	return client, nil
}

// loadCredentials loads credentials from AWS Secrets Manager
func (c *EbayClient) loadCredentials(ctx context.Context) error {
	// First check environment variables for development
	if appID := os.Getenv("EBAY_APP_ID"); appID != "" {
		c.credentials = &EbayCredentials{
			AppID:             appID,
			DevID:             os.Getenv("EBAY_DEV_ID"),
			CertID:            os.Getenv("EBAY_CERT_ID"),
			SandboxAppID:      os.Getenv("EBAY_SANDBOX_APP_ID"),
			OAuthRefreshToken: os.Getenv("EBAY_OAUTH_REFRESH_TOKEN"),
			Environment:       os.Getenv("EBAY_ENVIRONMENT"),
		}
		return nil
	}

	// Load from Secrets Manager
	result, err := c.secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(ebaySecretsName),
	})
	if err != nil {
		return fmt.Errorf("failed to get secret %s: %w", ebaySecretsName, err)
	}

	var creds EbayCredentials
	if err := json.Unmarshal([]byte(*result.SecretString), &creds); err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	c.credentials = &creds
	c.useSandbox = creds.Environment == "sandbox"
	return nil
}

// IsConfigured returns true if the client has valid credentials
func (c *EbayClient) IsConfigured() bool {
	return c != nil && c.credentials != nil && c.credentials.AppID != ""
}

// getAPIBase returns the appropriate API base URL
func (c *EbayClient) getAPIBase() string {
	if c.useSandbox {
		return ebaySandboxAPIBase
	}
	return ebayProductionAPIBase
}

// getAuthURL returns the appropriate auth URL
func (c *EbayClient) getAuthURL() string {
	if c.useSandbox {
		return ebaySandboxAuthURL
	}
	return ebayProductionAuthURL
}

// getAppID returns the appropriate app ID
func (c *EbayClient) getAppID() string {
	if c.useSandbox && c.credentials.SandboxAppID != "" {
		return c.credentials.SandboxAppID
	}
	return c.credentials.AppID
}

// ensureAccessToken ensures a valid access token is available
func (c *EbayClient) ensureAccessToken(ctx context.Context) error {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-5*time.Minute)) {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	return c.refreshAccessToken(ctx)
}

// refreshAccessToken refreshes the OAuth access token
func (c *EbayClient) refreshAccessToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring lock
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-5*time.Minute)) {
		return nil
	}

	if c.credentials.OAuthRefreshToken == "" {
		// Use client credentials flow (limited to Browse/Finding APIs)
		return c.getClientCredentialsToken(ctx)
	}

	// Use refresh token flow
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.credentials.OAuthRefreshToken)
	data.Set("scope", ebayInventoryScope)

	req, err := http.NewRequestWithContext(ctx, "POST", c.getAuthURL(), strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	// Basic auth with app credentials
	auth := base64.StdEncoding.EncodeToString([]byte(c.getAppID() + ":" + c.credentials.CertID))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp EbayTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// getClientCredentialsToken gets an application token (limited scope)
func (c *EbayClient) getClientCredentialsToken(ctx context.Context) error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "https://api.ebay.com/oauth/api_scope")

	req, err := http.NewRequestWithContext(ctx, "POST", c.getAuthURL(), strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.getAppID() + ":" + c.credentials.CertID))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("client credentials token failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp EbayTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// SearchCompletedItems searches for items to use for price research.
// This uses the Browse API as a replacement for the deprecated Finding API's
// findCompletedItems operation. The Browse API searches active listings rather
// than completed/sold items, but active listing prices still provide useful
// market value estimates for pricing guidance.
func (c *EbayClient) SearchCompletedItems(ctx context.Context, query string, limit int) (*EbayPriceAnalysis, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	// Browse API uses OAuth (not App ID like the deprecated Finding API)
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("sort", "price")
	params.Set("filter", "conditions:{Used|For parts or not working|New|Unspecified}")

	reqURL := fmt.Sprintf("%s/buy/browse/v1/item_summary/search?%s", c.getAPIBase(), params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.RUnlock()
	req.Header.Set("X-EBAY-C-MARKETPLACE-ID", "EBAY_US")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s - %s", resp.Status, string(body))
	}

	// Parse Browse API response
	var browseResp struct {
		Total         int `json:"total"`
		ItemSummaries []struct {
			ItemID string `json:"itemId"`
			Title  string `json:"title"`
			Price  struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"price"`
			Condition    string `json:"condition"`
			Image        struct {
				ImageURL string `json:"imageUrl"`
			} `json:"image"`
			ItemWebURL   string `json:"itemWebUrl"`
			ItemLocation struct {
				City    string `json:"city"`
				Country string `json:"country"`
			} `json:"itemLocation"`
			ShippingOptions []struct {
				ShippingCost struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"shippingCost"`
			} `json:"shippingOptions"`
			Seller struct {
				Username      string `json:"username"`
				FeedbackScore int    `json:"feedbackScore"`
			} `json:"seller"`
			BuyingOptions   []string `json:"buyingOptions"`
			CurrentBidPrice *struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"currentBidPrice"`
		} `json:"itemSummaries"`
	}

	if err := json.Unmarshal(body, &browseResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	analysis := &EbayPriceAnalysis{
		Query:   query,
		Results: []EbaySearchResult{},
	}

	if len(browseResp.ItemSummaries) == 0 {
		return analysis, nil
	}

	var prices []float64

	for _, item := range browseResp.ItemSummaries {
		result := EbaySearchResult{
			ItemID:         item.ItemID,
			Title:          item.Title,
			Condition:      item.Condition,
			ItemURL:        item.ItemWebURL,
			ImageURL:       item.Image.ImageURL,
			SellerUsername:  item.Seller.Username,
			SellerFeedback: item.Seller.FeedbackScore,
		}

		// Map location
		if item.ItemLocation.City != "" && item.ItemLocation.Country != "" {
			result.Location = item.ItemLocation.City + ", " + item.ItemLocation.Country
		} else if item.ItemLocation.Country != "" {
			result.Location = item.ItemLocation.Country
		}

		// Map buying options to listing type
		if len(item.BuyingOptions) > 0 {
			result.ListingType = item.BuyingOptions[0]
		}

		// Parse price
		if price, err := parsePrice(item.Price.Value); err == nil {
			result.Price = price
			result.Currency = item.Price.Currency
			prices = append(prices, price)
		}

		// Parse shipping cost
		if len(item.ShippingOptions) > 0 {
			if sc, err := parsePrice(item.ShippingOptions[0].ShippingCost.Value); err == nil {
				result.ShippingCost = sc
			}
		}

		analysis.Results = append(analysis.Results, result)
	}

	analysis.ResultCount = len(analysis.Results)

	if len(prices) > 0 {
		analysis.MinPrice = minFloat(prices)
		analysis.MaxPrice = maxFloat(prices)
		analysis.AveragePrice = avgFloat(prices)
		analysis.MedianPrice = medianFloat(prices)
		// Suggest prices between 25th and 75th percentile
		analysis.SuggestedLow = percentileFloat(prices, 25)
		analysis.SuggestedHigh = percentileFloat(prices, 75)
	}

	return analysis, nil
}

// SearchActiveListings searches for currently active listings using the Browse API
func (c *EbayClient) SearchActiveListings(ctx context.Context, query string, limit int) ([]EbaySearchResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	// Browse API uses OAuth
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("sort", "newlyListed")

	reqURL := fmt.Sprintf("%s/buy/browse/v1/item_summary/search?%s", c.getAPIBase(), params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.RUnlock()
	req.Header.Set("X-EBAY-C-MARKETPLACE-ID", "EBAY_US")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s - %s", resp.Status, string(body))
	}

	// Parse Browse API response
	var browseResp struct {
		Total         int `json:"total"`
		ItemSummaries []struct {
			ItemID string `json:"itemId"`
			Title  string `json:"title"`
			Price  struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"price"`
			Condition    string `json:"condition"`
			Image        struct {
				ImageURL string `json:"imageUrl"`
			} `json:"image"`
			ItemWebURL   string `json:"itemWebUrl"`
			ItemLocation struct {
				City    string `json:"city"`
				Country string `json:"country"`
			} `json:"itemLocation"`
			ShippingOptions []struct {
				ShippingCost struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"shippingCost"`
			} `json:"shippingOptions"`
			Seller struct {
				Username      string `json:"username"`
				FeedbackScore int    `json:"feedbackScore"`
			} `json:"seller"`
			BuyingOptions   []string `json:"buyingOptions"`
			CurrentBidPrice *struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"currentBidPrice"`
		} `json:"itemSummaries"`
	}

	if err := json.Unmarshal(body, &browseResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(browseResp.ItemSummaries) == 0 {
		return nil, nil
	}

	var results []EbaySearchResult

	for _, item := range browseResp.ItemSummaries {
		result := EbaySearchResult{
			ItemID:         item.ItemID,
			Title:          item.Title,
			Condition:      item.Condition,
			ItemURL:        item.ItemWebURL,
			ImageURL:       item.Image.ImageURL,
			SellerUsername:  item.Seller.Username,
			SellerFeedback: item.Seller.FeedbackScore,
		}

		// Map location
		if item.ItemLocation.City != "" && item.ItemLocation.Country != "" {
			result.Location = item.ItemLocation.City + ", " + item.ItemLocation.Country
		} else if item.ItemLocation.Country != "" {
			result.Location = item.ItemLocation.Country
		}

		// Map buying options to listing type
		if len(item.BuyingOptions) > 0 {
			result.ListingType = item.BuyingOptions[0]
		}

		// Parse price
		if price, err := parsePrice(item.Price.Value); err == nil {
			result.Price = price
			result.Currency = item.Price.Currency
		}

		// Parse shipping cost
		if len(item.ShippingOptions) > 0 {
			if sc, err := parsePrice(item.ShippingOptions[0].ShippingCost.Value); err == nil {
				result.ShippingCost = sc
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// CreateInventoryItem creates or updates an item in eBay inventory
func (c *EbayClient) CreateInventoryItem(ctx context.Context, item *EbayInventoryItem) error {
	if err := c.ensureAccessToken(ctx); err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	body, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	reqURL := fmt.Sprintf("%s/sell/inventory/v1/inventory_item/%s", c.getAPIBase(), item.SKU)
	req, err := http.NewRequestWithContext(ctx, "PUT", reqURL, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.RUnlock()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Language", "en-US")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create inventory item failed: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

// GetInventoryItem retrieves an inventory item by SKU
func (c *EbayClient) GetInventoryItem(ctx context.Context, sku string) (*EbayInventoryItem, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	reqURL := fmt.Sprintf("%s/sell/inventory/v1/inventory_item/%s", c.getAPIBase(), sku)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.RUnlock()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get inventory item failed: %s - %s", resp.Status, string(body))
	}

	var item EbayInventoryItem
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &item, nil
}

// CreateOffer creates a listing offer for an inventory item
func (c *EbayClient) CreateOffer(ctx context.Context, offer *EbayOffer) (string, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	body, err := json.Marshal(offer)
	if err != nil {
		return "", fmt.Errorf("failed to marshal offer: %w", err)
	}

	reqURL := fmt.Sprintf("%s/sell/inventory/v1/offer", c.getAPIBase())
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.RUnlock()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Language", "en-US")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create offer failed: %s - %s", resp.Status, string(respBody))
	}

	var result struct {
		OfferID string `json:"offerId"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.OfferID, nil
}

// PublishOffer publishes an offer to make it live on eBay
func (c *EbayClient) PublishOffer(ctx context.Context, offerID string) (string, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	reqURL := fmt.Sprintf("%s/sell/inventory/v1/offer/%s/publish", c.getAPIBase(), offerID)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.RUnlock()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("publish offer failed: %s - %s", resp.Status, string(respBody))
	}

	var result struct {
		ListingID string `json:"listingId"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.ListingID, nil
}

// GetCategorySuggestions gets category suggestions for a query
func (c *EbayClient) GetCategorySuggestions(ctx context.Context, query string) ([]EbayCategory, error) {
	if err := c.ensureAccessToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	params := url.Values{}
	params.Set("q", query)

	reqURL := fmt.Sprintf("%s/commerce/taxonomy/v1/category_tree/0/get_category_suggestions?%s",
		c.getAPIBase(), params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.RUnlock()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get categories failed: %s - %s", resp.Status, string(body))
	}

	var result struct {
		CategorySuggestions []struct {
			Category EbayCategory `json:"category"`
		} `json:"categorySuggestions"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var categories []EbayCategory
	for _, s := range result.CategorySuggestions {
		categories = append(categories, s.Category)
	}

	return categories, nil
}

// Helper functions for price analysis
func parsePrice(s string) (float64, error) {
	var price float64
	_, err := fmt.Sscanf(s, "%f", &price)
	return price, err
}

func minFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	min := vals[0]
	for _, v := range vals[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func maxFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	max := vals[0]
	for _, v := range vals[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func avgFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func medianFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	// Simple median (not sorting for performance)
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	// Bubble sort for simplicity (small arrays)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func percentileFloat(vals []float64, p int) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	idx := int(float64(len(sorted)-1) * float64(p) / 100)
	return sorted[idx]
}
