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

// HueClient controls Philips Hue lights via the bridge REST API.
type HueClient struct {
	bridgeIP   string
	username   string
	httpClient *http.Client
}

// HueLight represents a single Hue light.
type HueLight struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	ModelID    string `json:"modelid"`
	On         bool   `json:"on"`
	Brightness int    `json:"bri"`        // 0-254
	Hue        int    `json:"hue"`        // 0-65535
	Saturation int    `json:"sat"`        // 0-254
	ColorTemp  int    `json:"ct"`         // 153-500 (mireds)
	Reachable  bool   `json:"reachable"`
	ColorMode  string `json:"colormode"`
}

// HueRoom represents a room/group.
type HueRoom struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Lights []string `json:"lights"`
	AllOn  bool     `json:"all_on"`
	AnyOn  bool     `json:"any_on"`
}

// HueScene represents a saved scene.
type HueScene struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Group   string `json:"group"`
	Lights  []string `json:"lights"`
	Recycle bool   `json:"recycle"`
}

// HueBridgeStatus represents bridge info.
type HueBridgeStatus struct {
	Name       string `json:"name"`
	Model      string `json:"modelid"`
	BridgeID   string `json:"bridgeid"`
	APIVersion string `json:"apiversion"`
	SWVersion  string `json:"swversion"`
	MAC        string `json:"mac"`
	LightCount int    `json:"light_count"`
	GroupCount int    `json:"group_count"`
	SceneCount int    `json:"scene_count"`
}

// HueEntertainmentGroup represents an entertainment zone.
type HueEntertainmentGroup struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Lights  []string `json:"lights"`
	Type    string   `json:"type"`
	Stream  bool     `json:"stream_active"`
}

// HueHealth holds health check results.
type HueHealth struct {
	Connected       bool     `json:"connected"`
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// TestOverrideHueClient allows injection of a mock client in tests.
var TestOverrideHueClient *HueClient

var (
	hueOnce   sync.Once
	hueClient *HueClient
)

// GetHueClient returns the singleton Hue client.
func GetHueClient() (*HueClient, error) {
	if TestOverrideHueClient != nil {
		return TestOverrideHueClient, nil
	}

	var initErr error
	hueOnce.Do(func() {
		bridgeIP := os.Getenv("HUE_BRIDGE_IP")
		username := os.Getenv("HUE_USERNAME")
		if bridgeIP == "" {
			initErr = fmt.Errorf("HUE_BRIDGE_IP not set")
			return
		}
		if username == "" {
			initErr = fmt.Errorf("HUE_USERNAME not set")
			return
		}
		hueClient = &HueClient{
			bridgeIP:   bridgeIP,
			username:   username,
			httpClient: httpclient.Fast(),
		}
	})
	if initErr != nil {
		return nil, initErr
	}
	if hueClient == nil {
		return nil, fmt.Errorf("hue client not initialized")
	}
	return hueClient, nil
}

// NewTestHueClient creates a client for testing.
func NewTestHueClient(bridgeIP, username string) *HueClient {
	return &HueClient{
		bridgeIP:   bridgeIP,
		username:   username,
		httpClient: httpclient.Fast(),
	}
}

func (c *HueClient) baseURL() string {
	return fmt.Sprintf("http://%s/api/%s", c.bridgeIP, c.username)
}

func (c *HueClient) get(ctx context.Context, path string) ([]byte, error) {
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
		return nil, fmt.Errorf("hue API returned %d", resp.StatusCode)
	}
	var buf [32768]byte
	n, _ := resp.Body.Read(buf[:])
	return buf[:n], nil
}

func (c *HueClient) put(ctx context.Context, path string, body interface{}) error {
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hue API returned %d", resp.StatusCode)
	}
	return nil
}

// GetBridgeStatus returns bridge info.
func (c *HueClient) GetBridgeStatus(ctx context.Context) (*HueBridgeStatus, error) {
	data, err := c.get(ctx, "/config")
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	status := &HueBridgeStatus{}
	if v, ok := raw["name"].(string); ok {
		status.Name = v
	}
	if v, ok := raw["modelid"].(string); ok {
		status.Model = v
	}
	if v, ok := raw["bridgeid"].(string); ok {
		status.BridgeID = v
	}
	if v, ok := raw["apiversion"].(string); ok {
		status.APIVersion = v
	}
	if v, ok := raw["swversion"].(string); ok {
		status.SWVersion = v
	}
	if v, ok := raw["mac"].(string); ok {
		status.MAC = v
	}
	return status, nil
}

// GetLights returns all lights with their current state.
func (c *HueClient) GetLights(ctx context.Context) ([]HueLight, error) {
	data, err := c.get(ctx, "/lights")
	if err != nil {
		return nil, err
	}
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	var lights []HueLight
	for id, lightData := range raw {
		light := HueLight{ID: id}
		if v, ok := lightData["name"].(string); ok {
			light.Name = v
		}
		if v, ok := lightData["type"].(string); ok {
			light.Type = v
		}
		if v, ok := lightData["modelid"].(string); ok {
			light.ModelID = v
		}
		if state, ok := lightData["state"].(map[string]interface{}); ok {
			light.On, _ = state["on"].(bool)
			if v, ok := state["bri"].(float64); ok {
				light.Brightness = int(v)
			}
			if v, ok := state["hue"].(float64); ok {
				light.Hue = int(v)
			}
			if v, ok := state["sat"].(float64); ok {
				light.Saturation = int(v)
			}
			if v, ok := state["ct"].(float64); ok {
				light.ColorTemp = int(v)
			}
			light.Reachable, _ = state["reachable"].(bool)
			if v, ok := state["colormode"].(string); ok {
				light.ColorMode = v
			}
		}
		lights = append(lights, light)
	}
	return lights, nil
}

// SetLightState controls a single light.
func (c *HueClient) SetLightState(ctx context.Context, lightID string, state map[string]interface{}) error {
	return c.put(ctx, "/lights/"+lightID+"/state", state)
}

// GetRooms returns all groups/rooms.
func (c *HueClient) GetRooms(ctx context.Context) ([]HueRoom, error) {
	data, err := c.get(ctx, "/groups")
	if err != nil {
		return nil, err
	}
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	var rooms []HueRoom
	for id, groupData := range raw {
		room := HueRoom{ID: id}
		if v, ok := groupData["name"].(string); ok {
			room.Name = v
		}
		if v, ok := groupData["type"].(string); ok {
			room.Type = v
		}
		if lights, ok := groupData["lights"].([]interface{}); ok {
			for _, l := range lights {
				if s, ok := l.(string); ok {
					room.Lights = append(room.Lights, s)
				}
			}
		}
		if state, ok := groupData["state"].(map[string]interface{}); ok {
			room.AllOn, _ = state["all_on"].(bool)
			room.AnyOn, _ = state["any_on"].(bool)
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// SetRoomState controls all lights in a group.
func (c *HueClient) SetRoomState(ctx context.Context, groupID string, state map[string]interface{}) error {
	return c.put(ctx, "/groups/"+groupID+"/action", state)
}

// GetScenes returns all saved scenes.
func (c *HueClient) GetScenes(ctx context.Context) ([]HueScene, error) {
	data, err := c.get(ctx, "/scenes")
	if err != nil {
		return nil, err
	}
	var raw map[string]map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	var scenes []HueScene
	for id, sceneData := range raw {
		scene := HueScene{ID: id}
		if v, ok := sceneData["name"].(string); ok {
			scene.Name = v
		}
		if v, ok := sceneData["type"].(string); ok {
			scene.Type = v
		}
		if v, ok := sceneData["group"].(string); ok {
			scene.Group = v
		}
		if v, ok := sceneData["recycle"].(bool); ok {
			scene.Recycle = v
		}
		if lights, ok := sceneData["lights"].([]interface{}); ok {
			for _, l := range lights {
				if s, ok := l.(string); ok {
					scene.Lights = append(scene.Lights, s)
				}
			}
		}
		scenes = append(scenes, scene)
	}
	return scenes, nil
}

// ActivateScene activates a scene on a group.
func (c *HueClient) ActivateScene(ctx context.Context, sceneID string) error {
	// Find the scene's group first
	scenes, err := c.GetScenes(ctx)
	if err != nil {
		return err
	}
	for _, s := range scenes {
		if s.ID == sceneID {
			if s.Group != "" {
				return c.put(ctx, "/groups/"+s.Group+"/action", map[string]interface{}{
					"scene": sceneID,
				})
			}
			break
		}
	}
	// Fallback: try group 0
	return c.put(ctx, "/groups/0/action", map[string]interface{}{
		"scene": sceneID,
	})
}

// GetEntertainmentGroups returns entertainment zones.
func (c *HueClient) GetEntertainmentGroups(ctx context.Context) ([]HueEntertainmentGroup, error) {
	rooms, err := c.GetRooms(ctx)
	if err != nil {
		return nil, err
	}
	var groups []HueEntertainmentGroup
	for _, r := range rooms {
		if r.Type == "Entertainment" {
			groups = append(groups, HueEntertainmentGroup{
				ID:     r.ID,
				Name:   r.Name,
				Lights: r.Lights,
				Type:   r.Type,
			})
		}
	}
	return groups, nil
}

// Discover finds Hue bridges using the NUPNP endpoint.
func (c *HueClient) Discover(ctx context.Context) ([]map[string]interface{}, error) {
	var devices []map[string]interface{}

	// Try NUPNP (Philips hosted discovery)
	req, err := http.NewRequestWithContext(ctx, "GET", "https://discovery.meethue.com", nil)
	if err == nil {
		resp, err := c.httpClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			var bridges []map[string]interface{}
			var buf [4096]byte
			n, _ := resp.Body.Read(buf[:])
			if json.Unmarshal(buf[:n], &bridges) == nil {
				devices = append(devices, bridges...)
			}
		}
	}

	// Also try configured bridge
	if c.bridgeIP != "" {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", c.bridgeIP), 2*time.Second)
		if err == nil {
			conn.Close()
			devices = append(devices, map[string]interface{}{
				"internalipaddress": c.bridgeIP,
				"source":           "configured",
			})
		}
	}

	return devices, nil
}

// GetHealth performs a health check.
func (c *HueClient) GetHealth(ctx context.Context) (*HueHealth, error) {
	health := &HueHealth{Score: 0, Status: "unknown"}

	status, err := c.GetBridgeStatus(ctx)
	if err != nil {
		health.Status = "offline"
		health.Recommendations = append(health.Recommendations, fmt.Sprintf("Cannot reach Hue bridge at %s: %v", c.bridgeIP, err))
		return health, nil
	}

	health.Connected = true
	health.Score = 100
	health.Status = "healthy"

	if status.Name == "" {
		health.Score -= 10
		health.Recommendations = append(health.Recommendations, "Bridge name not set")
	}

	// Check lights
	lights, err := c.GetLights(ctx)
	if err == nil {
		status.LightCount = len(lights)
		unreachable := 0
		for _, l := range lights {
			if !l.Reachable {
				unreachable++
			}
		}
		if unreachable > 0 {
			health.Score -= unreachable * 5
			health.Recommendations = append(health.Recommendations, fmt.Sprintf("%d light(s) unreachable", unreachable))
		}
	}

	if health.Score < 50 {
		health.Status = "degraded"
	}

	return health, nil
}
