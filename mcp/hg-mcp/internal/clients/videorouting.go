// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/httpclient"
)

// VideoRoutingClient manages video routing between systems via NDI
type VideoRoutingClient struct {
	mu      sync.RWMutex
	routes  map[string]*VideoRoute
	sources map[string]*VideoSource

	// Clients for video systems
	resolumeClient *ResolumeClient
	obsClient      *OBSClient
	atemClient     *ATEMClient
}

// VideoSource represents an NDI or video source
type VideoSource struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`   // ndi, resolume, obs, atem, touchdesigner
	System     string                 `json:"system"` // Source system name
	URL        string                 `json:"url"`    // NDI URL or identifier
	Resolution string                 `json:"resolution,omitempty"`
	FrameRate  float64                `json:"frame_rate,omitempty"`
	Connected  bool                   `json:"connected"`
	LastSeen   time.Time              `json:"last_seen"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// VideoRoute represents a video routing configuration
type VideoRoute struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Source      *VideoSource `json:"source"`
	Destination string       `json:"destination"` // System receiving the video
	DestInput   string       `json:"dest_input"`  // Input channel/layer on destination
	Active      bool         `json:"active"`
	CreatedAt   time.Time    `json:"created_at"`
}

// VideoMatrix represents the complete routing matrix
type VideoMatrix struct {
	Sources      []*VideoSource `json:"sources"`
	Destinations []string       `json:"destinations"`
	Routes       []*VideoRoute  `json:"routes"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// NewVideoRoutingClient creates a new video routing client
func NewVideoRoutingClient() (*VideoRoutingClient, error) {
	return &VideoRoutingClient{
		routes:  make(map[string]*VideoRoute),
		sources: make(map[string]*VideoSource),
	}, nil
}

// Lazy client initialization
func (c *VideoRoutingClient) getResolumeClient() (*ResolumeClient, error) {
	if c.resolumeClient == nil {
		client, err := NewResolumeClient()
		if err != nil {
			return nil, err
		}
		c.resolumeClient = client
	}
	return c.resolumeClient, nil
}

func (c *VideoRoutingClient) getOBSClient() (*OBSClient, error) {
	if c.obsClient == nil {
		client, err := NewOBSClient()
		if err != nil {
			return nil, err
		}
		c.obsClient = client
	}
	return c.obsClient, nil
}

func (c *VideoRoutingClient) getATEMClient() (*ATEMClient, error) {
	if c.atemClient == nil {
		client, err := NewATEMClient()
		if err != nil {
			return nil, err
		}
		c.atemClient = client
	}
	return c.atemClient, nil
}

// DiscoverSources discovers available video sources from all systems
func (c *VideoRoutingClient) DiscoverSources(ctx context.Context) ([]*VideoSource, error) {
	var sources []*VideoSource
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Discover from Resolume (outputs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if client, err := c.getResolumeClient(); err == nil {
			if outputs, err := client.GetOutputs(ctx); err == nil {
				mu.Lock()
				for _, out := range outputs {
					source := &VideoSource{
						ID:        fmt.Sprintf("resolume_out_%d", out.Index),
						Name:      out.Name,
						Type:      "resolume",
						System:    "resolume",
						Connected: out.Enabled,
						LastSeen:  time.Now(),
						Metadata: map[string]interface{}{
							"output_index": out.Index,
							"width":        out.Width,
							"height":       out.Height,
						},
					}
					if out.Width > 0 && out.Height > 0 {
						source.Resolution = fmt.Sprintf("%dx%d", out.Width, out.Height)
					}
					sources = append(sources, source)
				}
				mu.Unlock()
			}
		}
	}()

	// Discover from OBS (scenes as sources)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if client, err := c.getOBSClient(); err == nil {
			if scenes, err := client.GetScenes(ctx); err == nil {
				mu.Lock()
				for _, scene := range scenes {
					source := &VideoSource{
						ID:        fmt.Sprintf("obs_scene_%s", scene.Name),
						Name:      scene.Name,
						Type:      "obs_scene",
						System:    "obs",
						Connected: true,
						LastSeen:  time.Now(),
					}
					sources = append(sources, source)
				}
				mu.Unlock()
			}
			// Also get OBS sources
			if obsSources, err := client.GetSources(ctx); err == nil {
				mu.Lock()
				for _, src := range obsSources {
					source := &VideoSource{
						ID:        fmt.Sprintf("obs_source_%s", src.Name),
						Name:      src.Name,
						Type:      "obs_source",
						System:    "obs",
						Connected: true,
						LastSeen:  time.Now(),
						Metadata: map[string]interface{}{
							"type": src.Type,
						},
					}
					sources = append(sources, source)
				}
				mu.Unlock()
			}
		}
	}()

	// Discover from ATEM (inputs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if client, err := c.getATEMClient(); err == nil {
			if inputs, err := client.GetInputs(ctx); err == nil {
				mu.Lock()
				for _, input := range inputs {
					source := &VideoSource{
						ID:        fmt.Sprintf("atem_input_%d", input.ID),
						Name:      input.Name,
						Type:      "atem_input",
						System:    "atem",
						Connected: true,
						LastSeen:  time.Now(),
						Metadata: map[string]interface{}{
							"input_id":   input.ID,
							"short_name": input.ShortName,
							"input_type": input.Type,
						},
					}
					sources = append(sources, source)
				}
				mu.Unlock()
			}
		}
	}()

	// Discover NDI sources (simulated - would need NDI SDK)
	wg.Add(1)
	go func() {
		defer wg.Done()
		ndiSources := c.discoverNDISources(ctx)
		mu.Lock()
		sources = append(sources, ndiSources...)
		mu.Unlock()
	}()

	wg.Wait()

	// Update cache
	c.mu.Lock()
	for _, src := range sources {
		c.sources[src.ID] = src
	}
	c.mu.Unlock()

	return sources, nil
}

// discoverNDISources discovers NDI sources on the network
func (c *VideoRoutingClient) discoverNDISources(ctx context.Context) []*VideoSource {
	var sources []*VideoSource

	// Use the NDI client to discover sources
	ndiClient, err := NewNDIClient()
	if err != nil {
		return sources
	}

	ndiSources, err := ndiClient.DiscoverSources(ctx)
	if err == nil {
		for _, ndi := range ndiSources {
			source := &VideoSource{
				ID:        fmt.Sprintf("ndi_%s", ndi.Name),
				Name:      ndi.Name,
				Type:      "ndi",
				System:    "ndi",
				URL:       ndi.URL,
				Connected: ndi.Connected,
				LastSeen:  time.Now(),
			}
			if ndi.Width > 0 && ndi.Height > 0 {
				source.Resolution = fmt.Sprintf("%dx%d", ndi.Width, ndi.Height)
				source.FrameRate = ndi.FPS
			}
			sources = append(sources, source)
		}
	}

	return sources
}

// GetMatrix returns the current video routing matrix
func (c *VideoRoutingClient) GetMatrix(ctx context.Context) (*VideoMatrix, error) {
	// Refresh sources
	sources, err := c.DiscoverSources(ctx)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	routes := make([]*VideoRoute, 0, len(c.routes))
	for _, r := range c.routes {
		routes = append(routes, r)
	}
	c.mu.RUnlock()

	// Available destinations
	destinations := []string{
		"resolume",
		"obs",
		"atem",
		"touchdesigner",
	}

	return &VideoMatrix{
		Sources:      sources,
		Destinations: destinations,
		Routes:       routes,
		UpdatedAt:    time.Now(),
	}, nil
}

// CreateRoute creates a video route from source to destination
func (c *VideoRoutingClient) CreateRoute(ctx context.Context, name, sourceID, destination, destInput string) (*VideoRoute, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find source
	source, exists := c.sources[sourceID]
	if !exists {
		return nil, fmt.Errorf("source not found: %s", sourceID)
	}

	// Validate destination
	validDests := map[string]bool{
		"resolume":      true,
		"obs":           true,
		"atem":          true,
		"touchdesigner": true,
	}
	if !validDests[destination] {
		return nil, fmt.Errorf("invalid destination: %s", destination)
	}

	id := fmt.Sprintf("route_%d", time.Now().UnixNano())
	route := &VideoRoute{
		ID:          id,
		Name:        name,
		Source:      source,
		Destination: destination,
		DestInput:   destInput,
		Active:      true,
		CreatedAt:   time.Now(),
	}

	c.routes[id] = route

	// Apply route
	if err := c.applyRoute(ctx, route); err != nil {
		return route, fmt.Errorf("route created but failed to apply: %w", err)
	}

	return route, nil
}

// applyRoute applies a video route to the destination system
func (c *VideoRoutingClient) applyRoute(ctx context.Context, route *VideoRoute) error {
	switch route.Destination {
	case "resolume":
		// Resolume can receive NDI via clip sources
		client, err := c.getResolumeClient()
		if err != nil {
			return err
		}
		// Load NDI source to a layer/clip
		if route.DestInput != "" {
			var layer, column int
			if _, err := fmt.Sscanf(route.DestInput, "layer/%d/column/%d", &layer, &column); err == nil {
				return client.LoadSource(ctx, layer, column, route.Source.Name)
			}
		}
		return nil

	case "obs":
		// OBS sources can be configured via WebSocket
		// This would require creating/updating an NDI source in OBS
		return nil

	case "atem":
		// ATEM routing is done via input selection
		client, err := c.getATEMClient()
		if err != nil {
			return err
		}
		// Parse destination input as input number
		var input int
		if _, err := fmt.Sscanf(route.DestInput, "input/%d", &input); err == nil {
			return client.SetProgram(ctx, input)
		}
		// Also try me/N/input/M format for flexibility
		var me int
		if _, err := fmt.Sscanf(route.DestInput, "me/%d/input/%d", &me, &input); err == nil {
			return client.SetProgram(ctx, input)
		}
		return nil

	case "touchdesigner":
		// TouchDesigner would need HTTP/OSC API
		return nil

	default:
		return fmt.Errorf("unknown destination: %s", route.Destination)
	}
}

// DeleteRoute removes a video route
func (c *VideoRoutingClient) DeleteRoute(ctx context.Context, routeID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.routes[routeID]; !exists {
		return fmt.Errorf("route not found: %s", routeID)
	}

	delete(c.routes, routeID)
	return nil
}

// SetRouteActive enables or disables a route
func (c *VideoRoutingClient) SetRouteActive(ctx context.Context, routeID string, active bool) error {
	c.mu.Lock()
	route, exists := c.routes[routeID]
	if !exists {
		c.mu.Unlock()
		return fmt.Errorf("route not found: %s", routeID)
	}
	route.Active = active
	c.mu.Unlock()

	if active {
		return c.applyRoute(ctx, route)
	}
	return nil
}

// Note: ATEMInput, GetInputs, SetProgramInput are defined in atem.go
// Note: NDISource and DiscoverNDI are defined in ndi.go

// Helper for HTTP requests (if needed for NDI Tools)
func (c *VideoRoutingClient) httpGet(ctx context.Context, url string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := httpclient.Fast().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
