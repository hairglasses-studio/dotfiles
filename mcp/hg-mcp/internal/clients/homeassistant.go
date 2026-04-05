// Package clients provides API clients for external services.
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

	"github.com/hairglasses-studio/mcpkit/resilience"
	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// HomeAssistant caches — reduce API calls to local HA instance
var (
	haEntitiesCache = resilience.NewCache[[]HAEntity](15 * time.Second) // Entity list (polled by dashboards)
	haScenesCache   = resilience.NewCache[[]HAScene](30 * time.Second)  // Scene list (rarely changes)
	haStatusCache   = resilience.NewCache[*HAStatus](10 * time.Second)  // Status (polled frequently)
)

// HomeAssistantClient provides Home Assistant integration
type HomeAssistantClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// HAEntity represents a Home Assistant entity
type HAEntity struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
	LastUpdated string                 `json:"last_updated"`
}

// HAService represents a Home Assistant service
type HAService struct {
	Domain      string                 `json:"domain"`
	Service     string                 `json:"service"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// HAEvent represents a Home Assistant event
type HAEvent struct {
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Origin    string                 `json:"origin"`
	TimeFired string                 `json:"time_fired"`
}

// HAAutomation represents a Home Assistant automation
type HAAutomation struct {
	ID            string `json:"id"`
	Alias         string `json:"alias"`
	Description   string `json:"description,omitempty"`
	State         string `json:"state"` // on, off
	Mode          string `json:"mode,omitempty"`
	LastTriggered string `json:"last_triggered,omitempty"`
}

// HAScene represents a Home Assistant scene
type HAScene struct {
	EntityID string `json:"entity_id"`
	Name     string `json:"name"`
	State    string `json:"state"`
}

// HAStatus represents Home Assistant status
type HAStatus struct {
	Connected    bool   `json:"connected"`
	Version      string `json:"version"`
	BaseURL      string `json:"base_url"`
	EntityCount  int    `json:"entity_count"`
	LocationName string `json:"location_name,omitempty"`
}

// HAServiceCall represents a service call request
type HAServiceCall struct {
	EntityID string                 `json:"entity_id,omitempty"`
	Data     map[string]interface{} `json:"-"`
}

// NewHomeAssistantClient creates a new Home Assistant client
var (
	hassClientSingleton *HomeAssistantClient
	hassClientOnce      sync.Once
	hassClientErr       error

	// TestOverrideHomeAssistantClient, when non-nil, is returned by GetHomeAssistantClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideHomeAssistantClient *HomeAssistantClient
)

// GetHomeAssistantClient returns the singleton Home Assistant client.
func GetHomeAssistantClient() (*HomeAssistantClient, error) {
	if TestOverrideHomeAssistantClient != nil {
		return TestOverrideHomeAssistantClient, nil
	}
	hassClientOnce.Do(func() {
		hassClientSingleton, hassClientErr = NewHomeAssistantClient()
	})
	return hassClientSingleton, hassClientErr
}

// NewTestHomeAssistantClient creates an in-memory test client.
func NewTestHomeAssistantClient() *HomeAssistantClient {
	return &HomeAssistantClient{
		baseURL:    "http://localhost:8123",
		token:      "test-token",
		httpClient: httpclient.Fast(),
	}
}

func NewHomeAssistantClient() (*HomeAssistantClient, error) {
	baseURL := os.Getenv("HASS_URL")
	if baseURL == "" {
		baseURL = "http://homeassistant.local:8123"
	}

	token := os.Getenv("HASS_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("HASS_TOKEN environment variable is required")
	}

	return &HomeAssistantClient{
		baseURL:    baseURL,
		token:      token,
		httpClient: httpclient.Standard(),
	}, nil
}

// doRequest performs an HTTP request to Home Assistant
func (c *HomeAssistantClient) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	url := c.baseURL + "/api" + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
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
		return nil, fmt.Errorf("Home Assistant API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetStatus returns Home Assistant status (cached 10s)
func (c *HomeAssistantClient) GetStatus(ctx context.Context) (*HAStatus, error) {
	return haStatusCache.GetOrFetch(ctx, func(ctx context.Context) (*HAStatus, error) {
		status := &HAStatus{
			BaseURL: c.baseURL,
		}

		// Check connection
		data, err := c.doRequest(ctx, "GET", "/", nil)
		if err != nil {
			status.Connected = false
			return status, nil
		}

		status.Connected = true

		// Parse response
		var resp map[string]interface{}
		if err := json.Unmarshal(data, &resp); err == nil {
			if version, ok := resp["version"].(string); ok {
				status.Version = version
			}
			if location, ok := resp["location_name"].(string); ok {
				status.LocationName = location
			}
		}

		// Get entity count
		entities, err := c.GetEntities(ctx, "")
		if err == nil {
			status.EntityCount = len(entities)
		}

		return status, nil
	})
}

// GetEntities returns all entities or filtered by domain (full list cached 15s)
func (c *HomeAssistantClient) GetEntities(ctx context.Context, domain string) ([]HAEntity, error) {
	// Cache the full entity list, then filter in-memory
	entities, err := haEntitiesCache.GetOrFetch(ctx, func(ctx context.Context) ([]HAEntity, error) {
		data, err := c.doRequest(ctx, "GET", "/states", nil)
		if err != nil {
			return nil, err
		}

		var entities []HAEntity
		if err := json.Unmarshal(data, &entities); err != nil {
			return nil, fmt.Errorf("failed to parse entities: %w", err)
		}
		return entities, nil
	})
	if err != nil {
		return nil, err
	}

	// Filter by domain if specified
	if domain != "" {
		filtered := []HAEntity{}
		for _, e := range entities {
			if len(e.EntityID) > len(domain)+1 && e.EntityID[:len(domain)+1] == domain+"." {
				filtered = append(filtered, e)
			}
		}
		return filtered, nil
	}

	return entities, nil
}

// GetEntity returns a specific entity
func (c *HomeAssistantClient) GetEntity(ctx context.Context, entityID string) (*HAEntity, error) {
	data, err := c.doRequest(ctx, "GET", "/states/"+entityID, nil)
	if err != nil {
		return nil, err
	}

	var entity HAEntity
	if err := json.Unmarshal(data, &entity); err != nil {
		return nil, fmt.Errorf("failed to parse entity: %w", err)
	}

	return &entity, nil
}

// CallService calls a Home Assistant service
func (c *HomeAssistantClient) CallService(ctx context.Context, domain, service string, data map[string]interface{}) error {
	endpoint := fmt.Sprintf("/services/%s/%s", domain, service)

	_, err := c.doRequest(ctx, "POST", endpoint, data)
	return err
}

// TurnOn turns on an entity
func (c *HomeAssistantClient) TurnOn(ctx context.Context, entityID string, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["entity_id"] = entityID

	// Determine domain
	domain := "homeassistant"
	if len(entityID) > 0 {
		for i, c := range entityID {
			if c == '.' {
				domain = entityID[:i]
				break
			}
		}
	}

	return c.CallService(ctx, domain, "turn_on", data)
}

// TurnOff turns off an entity
func (c *HomeAssistantClient) TurnOff(ctx context.Context, entityID string) error {
	data := map[string]interface{}{
		"entity_id": entityID,
	}

	// Determine domain
	domain := "homeassistant"
	if len(entityID) > 0 {
		for i, c := range entityID {
			if c == '.' {
				domain = entityID[:i]
				break
			}
		}
	}

	return c.CallService(ctx, domain, "turn_off", data)
}

// Toggle toggles an entity
func (c *HomeAssistantClient) Toggle(ctx context.Context, entityID string) error {
	data := map[string]interface{}{
		"entity_id": entityID,
	}

	return c.CallService(ctx, "homeassistant", "toggle", data)
}

// GetScenes returns all scenes (cached 30s)
func (c *HomeAssistantClient) GetScenes(ctx context.Context) ([]HAScene, error) {
	return haScenesCache.GetOrFetch(ctx, func(ctx context.Context) ([]HAScene, error) {
		entities, err := c.GetEntities(ctx, "scene")
		if err != nil {
			return nil, err
		}

		scenes := make([]HAScene, len(entities))
		for i, e := range entities {
			name := e.EntityID
			if friendlyName, ok := e.Attributes["friendly_name"].(string); ok {
				name = friendlyName
			}
			scenes[i] = HAScene{
				EntityID: e.EntityID,
				Name:     name,
				State:    e.State,
			}
		}

		return scenes, nil
	})
}

// ActivateScene activates a scene
func (c *HomeAssistantClient) ActivateScene(ctx context.Context, sceneID string) error {
	data := map[string]interface{}{
		"entity_id": sceneID,
	}
	return c.CallService(ctx, "scene", "turn_on", data)
}

// GetAutomations returns all automations
func (c *HomeAssistantClient) GetAutomations(ctx context.Context) ([]HAAutomation, error) {
	entities, err := c.GetEntities(ctx, "automation")
	if err != nil {
		return nil, err
	}

	automations := make([]HAAutomation, len(entities))
	for i, e := range entities {
		alias := e.EntityID
		if friendlyName, ok := e.Attributes["friendly_name"].(string); ok {
			alias = friendlyName
		}

		auto := HAAutomation{
			ID:    e.EntityID,
			Alias: alias,
			State: e.State,
		}

		if mode, ok := e.Attributes["mode"].(string); ok {
			auto.Mode = mode
		}
		if lastTriggered, ok := e.Attributes["last_triggered"].(string); ok {
			auto.LastTriggered = lastTriggered
		}

		automations[i] = auto
	}

	return automations, nil
}

// TriggerAutomation triggers an automation
func (c *HomeAssistantClient) TriggerAutomation(ctx context.Context, automationID string) error {
	data := map[string]interface{}{
		"entity_id": automationID,
	}
	return c.CallService(ctx, "automation", "trigger", data)
}

// FireEvent fires a custom event
func (c *HomeAssistantClient) FireEvent(ctx context.Context, eventType string, eventData map[string]interface{}) error {
	endpoint := "/events/" + eventType
	_, err := c.doRequest(ctx, "POST", endpoint, eventData)
	return err
}

// IsConnected checks if connected to Home Assistant
func (c *HomeAssistantClient) IsConnected(ctx context.Context) bool {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return false
	}
	return status.Connected
}
