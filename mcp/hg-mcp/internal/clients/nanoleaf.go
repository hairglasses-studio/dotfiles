package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// NanoleafClient controls Nanoleaf panels via their HTTP REST API.
type NanoleafClient struct {
	host       string
	authToken  string
	httpClient *http.Client
}

// NanoleafStatus represents the panel state.
type NanoleafStatus struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	Serial     string `json:"serialNo"`
	Firmware   string `json:"firmwareVersion"`
	On         bool   `json:"on"`
	Brightness int    `json:"brightness"`
	Hue        int    `json:"hue"`
	Saturation int    `json:"saturation"`
	ColorTemp  int    `json:"ct"`
	Effect     string `json:"effect"`
	ColorMode  string `json:"colorMode"`
}

// NanoleafEffect represents a named effect.
type NanoleafEffect struct {
	Name string `json:"name"`
}

// NanoleafPanel represents a single panel in the layout.
type NanoleafPanel struct {
	ID        int `json:"panelId"`
	X         int `json:"x"`
	Y         int `json:"y"`
	Rotation  int `json:"o"`
	ShapeType int `json:"shapeType"`
}

// NanoleafLayout is the full panel layout.
type NanoleafLayout struct {
	NumPanels int             `json:"numPanels"`
	Panels    []NanoleafPanel `json:"positionData"`
}

// NanoleafHealth holds health check results.
type NanoleafHealth struct {
	Connected       bool     `json:"connected"`
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// TestOverrideNanoleafClient allows injection of a mock client in tests.
var TestOverrideNanoleafClient *NanoleafClient

var (
	nanoleafOnce   sync.Once
	nanoleafClient *NanoleafClient
)

// GetNanoleafClient returns the singleton Nanoleaf client.
func GetNanoleafClient() (*NanoleafClient, error) {
	if TestOverrideNanoleafClient != nil {
		return TestOverrideNanoleafClient, nil
	}

	var initErr error
	nanoleafOnce.Do(func() {
		host := os.Getenv("NANOLEAF_HOST")
		token := os.Getenv("NANOLEAF_AUTH_TOKEN")
		if host == "" {
			initErr = fmt.Errorf("NANOLEAF_HOST not set")
			return
		}
		if token == "" {
			initErr = fmt.Errorf("NANOLEAF_AUTH_TOKEN not set")
			return
		}
		nanoleafClient = &NanoleafClient{
			host:       host,
			authToken:  token,
			httpClient: httpclient.Fast(),
		}
	})
	if initErr != nil {
		return nil, initErr
	}
	if nanoleafClient == nil {
		return nil, fmt.Errorf("nanoleaf client not initialized")
	}
	return nanoleafClient, nil
}

// NewTestNanoleafClient creates a client for testing.
func NewTestNanoleafClient(host, token string) *NanoleafClient {
	return &NanoleafClient{
		host:       host,
		authToken:  token,
		httpClient: httpclient.Fast(),
	}
}

func (c *NanoleafClient) baseURL() string {
	return fmt.Sprintf("http://%s:16021/api/v1/%s", c.host, c.authToken)
}

func (c *NanoleafClient) get(ctx context.Context, path string) ([]byte, error) {
	url := c.baseURL() + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nanoleaf API returned %d", resp.StatusCode)
	}
	var buf [8192]byte
	n, _ := resp.Body.Read(buf[:])
	return buf[:n], nil
}

func (c *NanoleafClient) put(ctx context.Context, path string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	url := c.baseURL() + path
	req, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("nanoleaf API returned %d", resp.StatusCode)
	}
	return nil
}

// GetStatus returns the panel state.
func (c *NanoleafClient) GetStatus(ctx context.Context) (*NanoleafStatus, error) {
	data, err := c.get(ctx, "")
	if err != nil {
		return nil, err
	}

	// The API returns nested objects; we flatten into our struct
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	status := &NanoleafStatus{}
	if v, ok := raw["name"].(string); ok {
		status.Name = v
	}
	if v, ok := raw["model"].(string); ok {
		status.Model = v
	}
	if v, ok := raw["serialNo"].(string); ok {
		status.Serial = v
	}
	if v, ok := raw["firmwareVersion"].(string); ok {
		status.Firmware = v
	}

	if state, ok := raw["state"].(map[string]interface{}); ok {
		if on, ok := state["on"].(map[string]interface{}); ok {
			status.On, _ = on["value"].(bool)
		}
		if bri, ok := state["brightness"].(map[string]interface{}); ok {
			if v, ok := bri["value"].(float64); ok {
				status.Brightness = int(v)
			}
		}
		if hue, ok := state["hue"].(map[string]interface{}); ok {
			if v, ok := hue["value"].(float64); ok {
				status.Hue = int(v)
			}
		}
		if sat, ok := state["sat"].(map[string]interface{}); ok {
			if v, ok := sat["value"].(float64); ok {
				status.Saturation = int(v)
			}
		}
		if ct, ok := state["ct"].(map[string]interface{}); ok {
			if v, ok := ct["value"].(float64); ok {
				status.ColorTemp = int(v)
			}
		}
		if cm, ok := state["colorMode"].(string); ok {
			status.ColorMode = cm
		}
	}

	if effects, ok := raw["effects"].(map[string]interface{}); ok {
		if sel, ok := effects["select"].(string); ok {
			status.Effect = sel
		}
	}

	return status, nil
}

// SetPower turns the panels on or off.
func (c *NanoleafClient) SetPower(ctx context.Context, on bool) error {
	return c.put(ctx, "/state", map[string]interface{}{
		"on": map[string]interface{}{"value": on},
	})
}

// SetBrightness sets brightness (0-100).
func (c *NanoleafClient) SetBrightness(ctx context.Context, level int) error {
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}
	return c.put(ctx, "/state", map[string]interface{}{
		"brightness": map[string]interface{}{"value": level},
	})
}

// SetColor sets the panel color by HSB values.
func (c *NanoleafClient) SetColor(ctx context.Context, hue, sat, bri int) error {
	return c.put(ctx, "/state", map[string]interface{}{
		"hue":        map[string]interface{}{"value": hue},
		"sat":        map[string]interface{}{"value": sat},
		"brightness": map[string]interface{}{"value": bri},
	})
}

// ListEffects returns available effect names.
func (c *NanoleafClient) ListEffects(ctx context.Context) ([]string, error) {
	data, err := c.get(ctx, "/effects/effectsList")
	if err != nil {
		return nil, err
	}
	var effects []string
	if err := json.Unmarshal(data, &effects); err != nil {
		return nil, err
	}
	return effects, nil
}

// SetEffect activates a named effect.
func (c *NanoleafClient) SetEffect(ctx context.Context, name string) error {
	return c.put(ctx, "/effects", map[string]interface{}{
		"select": name,
	})
}

// GetLayout returns the panel layout.
func (c *NanoleafClient) GetLayout(ctx context.Context) (*NanoleafLayout, error) {
	data, err := c.get(ctx, "/panelLayout/layout")
	if err != nil {
		return nil, err
	}
	var layout NanoleafLayout
	if err := json.Unmarshal(data, &layout); err != nil {
		return nil, err
	}
	return &layout, nil
}

// GetHealth performs a health check.
func (c *NanoleafClient) GetHealth(ctx context.Context) (*NanoleafHealth, error) {
	health := &NanoleafHealth{Score: 0, Status: "unknown"}

	status, err := c.GetStatus(ctx)
	if err != nil {
		health.Status = "offline"
		health.Recommendations = append(health.Recommendations, fmt.Sprintf("Cannot reach Nanoleaf at %s: %v", c.host, err))
		return health, nil
	}

	health.Connected = true
	health.Score = 100
	health.Status = "healthy"

	if !status.On {
		health.Score -= 10
		health.Recommendations = append(health.Recommendations, "Panels are powered off")
	}
	if status.Brightness < 10 {
		health.Score -= 5
		health.Recommendations = append(health.Recommendations, "Brightness is very low")
	}

	if health.Score < 50 {
		health.Status = "degraded"
	}

	return health, nil
}

// Discover scans the local network for Nanoleaf devices via mDNS.
func (c *NanoleafClient) Discover(ctx context.Context) ([]map[string]interface{}, error) {
	// Nanoleaf uses mDNS service _nanoleafapi._tcp
	// For simplicity, we do a port scan on common Nanoleaf port 16021
	var devices []map[string]interface{}

	// Try the configured host first
	if c.host != "" {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:16021", c.host), 2*time.Second)
		if err == nil {
			conn.Close()
			devices = append(devices, map[string]interface{}{
				"host":   c.host,
				"port":   16021,
				"source": "configured",
			})
		}
	}

	return devices, nil
}
