package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const tailscaleAPIBase = "https://api.tailscale.com"

// TailscaleAPIClient provides Tailscale management via the REST API.
type TailscaleAPIClient struct {
	apiKey  string
	tailnet string
	http    *http.Client
}

// AuthKeyOpts configures a new auth key.
type AuthKeyOpts struct {
	Reusable      bool
	Ephemeral     bool
	Preauthorized bool
	Tags          []string
	Description   string
	ExpirySeconds int
}

// AuthKeyResult is returned after creating an auth key.
type AuthKeyResult struct {
	ID      string    `json:"id"`
	Key     string    `json:"key"`
	Expires time.Time `json:"expires"`
}

// APIDevice represents a device from the Tailscale API.
type APIDevice struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Hostname   string   `json:"hostname"`
	OS         string   `json:"os"`
	Addresses  []string `json:"addresses"`
	Authorized bool     `json:"authorized"`
	Tags       []string `json:"tags"`
	LastSeen   string   `json:"lastSeen"`
	ClientVersion string `json:"clientVersion"`
}

// NewTailscaleAPIClient creates a client using TAILSCALE_API_KEY from env,
// falling back to 1Password.
func NewTailscaleAPIClient() (*TailscaleAPIClient, error) {
	apiKey := os.Getenv("TAILSCALE_API_KEY")
	if apiKey == "" {
		// Try 1Password
		out, err := exec.Command("op", "read", "op://Personal/Tailscale API Key/credential").Output()
		if err == nil {
			apiKey = strings.TrimSpace(string(out))
		}
	}
	if apiKey == "" {
		return nil, fmt.Errorf("TAILSCALE_API_KEY not set and not found in 1Password")
	}

	return &TailscaleAPIClient{
		apiKey:  apiKey,
		tailnet: "-", // default tailnet
		http:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// CreateAuthKey creates a new auth key for device registration.
func (c *TailscaleAPIClient) CreateAuthKey(ctx context.Context, opts AuthKeyOpts) (*AuthKeyResult, error) {
	expiry := opts.ExpirySeconds
	if expiry <= 0 {
		expiry = 86400 // 24 hours
	}

	body := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"devices": map[string]interface{}{
				"create": map[string]interface{}{
					"reusable":      opts.Reusable,
					"ephemeral":     opts.Ephemeral,
					"preauthorized": opts.Preauthorized,
					"tags":          opts.Tags,
				},
			},
		},
		"expirySeconds": expiry,
		"description":   opts.Description,
	}

	var result AuthKeyResult
	if err := c.doJSON(ctx, "POST", fmt.Sprintf("/api/v2/tailnet/%s/keys", c.tailnet), body, &result); err != nil {
		return nil, fmt.Errorf("create auth key: %w", err)
	}
	return &result, nil
}

// ListAPIDevices returns all devices via the API.
func (c *TailscaleAPIClient) ListAPIDevices(ctx context.Context) ([]APIDevice, error) {
	var resp struct {
		Devices []APIDevice `json:"devices"`
	}
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/api/v2/tailnet/%s/devices", c.tailnet), nil, &resp); err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}
	return resp.Devices, nil
}

// GetDevice returns a single device.
func (c *TailscaleAPIClient) GetDevice(ctx context.Context, deviceID string) (*APIDevice, error) {
	var device APIDevice
	if err := c.doJSON(ctx, "GET", fmt.Sprintf("/api/v2/device/%s", deviceID), nil, &device); err != nil {
		return nil, fmt.Errorf("get device %s: %w", deviceID, err)
	}
	return &device, nil
}

// DeleteDevice removes a device from the tailnet.
func (c *TailscaleAPIClient) DeleteDevice(ctx context.Context, deviceID string) error {
	if err := c.doJSON(ctx, "DELETE", fmt.Sprintf("/api/v2/device/%s", deviceID), nil, nil); err != nil {
		return fmt.Errorf("delete device %s: %w", deviceID, err)
	}
	return nil
}

// ApproveDevice authorizes a device on the tailnet.
func (c *TailscaleAPIClient) ApproveDevice(ctx context.Context, deviceID string) error {
	body := map[string]interface{}{"authorized": true}
	if err := c.doJSON(ctx, "POST", fmt.Sprintf("/api/v2/device/%s/authorized", deviceID), body, nil); err != nil {
		return fmt.Errorf("approve device %s: %w", deviceID, err)
	}
	return nil
}

// SetDeviceTags sets ACL tags on a device.
func (c *TailscaleAPIClient) SetDeviceTags(ctx context.Context, deviceID string, tags []string) error {
	body := map[string]interface{}{"tags": tags}
	if err := c.doJSON(ctx, "POST", fmt.Sprintf("/api/v2/device/%s/tags", deviceID), body, nil); err != nil {
		return fmt.Errorf("set tags on %s: %w", deviceID, err)
	}
	return nil
}

func (c *TailscaleAPIClient) doJSON(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := tailscaleAPIBase + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API %s %s returned %d: %s", method, path, resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}
