// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/httpclient"
)

// PTZClient provides access to ONVIF-compatible PTZ cameras
type PTZClient struct {
	host         string
	port         string
	username     string
	password     string
	httpClient   *http.Client
	mediaURL     string
	ptzURL       string
	profileToken string
}

// PTZCamera represents a discovered camera
type PTZCamera struct {
	Name      string   `json:"name"`
	Host      string   `json:"host"`
	Model     string   `json:"model,omitempty"`
	Profiles  []string `json:"profiles,omitempty"`
	HasPTZ    bool     `json:"has_ptz"`
	Connected bool     `json:"connected"`
}

// PTZStatus represents camera PTZ status
type PTZStatus struct {
	Connected    bool        `json:"connected"`
	Host         string      `json:"host"`
	ProfileToken string      `json:"profile_token"`
	Position     PTZPosition `json:"position"`
	Moving       bool        `json:"moving"`
}

// PTZPosition represents pan/tilt/zoom position
type PTZPosition struct {
	Pan  float64 `json:"pan"`
	Tilt float64 `json:"tilt"`
	Zoom float64 `json:"zoom"`
}

// PTZPreset represents a saved preset position
type PTZPreset struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

// PTZHealth represents health status
type PTZHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	CamerasFound    int      `json:"cameras_found"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewPTZClient creates a new PTZ client
func NewPTZClient() (*PTZClient, error) {
	host := os.Getenv("PTZ_HOST")
	if host == "" {
		host = "192.168.1.100"
	}

	port := os.Getenv("PTZ_PORT")
	if port == "" {
		port = "80"
	}

	username := os.Getenv("PTZ_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("PTZ_PASSWORD")

	return &PTZClient{
		host:       host,
		port:       port,
		username:   username,
		password:   password,
		httpClient: httpclient.Fast(),
	}, nil
}

// baseURL returns the base URL for ONVIF
func (c *PTZClient) baseURL() string {
	if c.port == "80" {
		return fmt.Sprintf("http://%s", c.host)
	}
	return fmt.Sprintf("http://%s:%s", c.host, c.port)
}

// createSecurityHeader creates WS-Security UsernameToken header
func (c *PTZClient) createSecurityHeader() string {
	nonce := make([]byte, 16)
	for i := range nonce {
		nonce[i] = byte(i + int(time.Now().UnixNano()%256))
	}
	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	created := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	// Password digest = Base64(SHA1(nonce + created + password))
	h := sha1.New()
	h.Write(nonce)
	h.Write([]byte(created))
	h.Write([]byte(c.password))
	digest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf(`<Security xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
		<UsernameToken>
			<Username>%s</Username>
			<Password Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest">%s</Password>
			<Nonce EncodingType="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary">%s</Nonce>
			<Created xmlns="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">%s</Created>
		</UsernameToken>
	</Security>`, c.username, digest, nonceBase64, created)
}

// soapRequest sends a SOAP request
func (c *PTZClient) soapRequest(ctx context.Context, url, action, body string) ([]byte, error) {
	envelope := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Header>%s</s:Header>
	<s:Body>%s</s:Body>
</s:Envelope>`, c.createSecurityHeader(), body)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(envelope))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	if action != "" {
		req.Header.Set("SOAPAction", action)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// GetCapabilities retrieves device capabilities
func (c *PTZClient) GetCapabilities(ctx context.Context) error {
	url := c.baseURL() + "/onvif/device_service"
	body := `<GetCapabilities xmlns="http://www.onvif.org/ver10/device/wsdl">
		<Category>All</Category>
	</GetCapabilities>`

	respData, err := c.soapRequest(ctx, url, "", body)
	if err != nil {
		return err
	}

	// Parse response to extract service URLs
	respStr := string(respData)

	// Extract Media URL
	if idx := strings.Index(respStr, "<tt:Media>"); idx >= 0 {
		if endIdx := strings.Index(respStr[idx:], "</tt:Media>"); endIdx >= 0 {
			mediaSection := respStr[idx : idx+endIdx]
			if urlIdx := strings.Index(mediaSection, "<tt:XAddr>"); urlIdx >= 0 {
				if urlEnd := strings.Index(mediaSection[urlIdx:], "</tt:XAddr>"); urlEnd >= 0 {
					c.mediaURL = mediaSection[urlIdx+10 : urlIdx+urlEnd]
				}
			}
		}
	}

	// Extract PTZ URL
	if idx := strings.Index(respStr, "<tt:PTZ>"); idx >= 0 {
		if endIdx := strings.Index(respStr[idx:], "</tt:PTZ>"); endIdx >= 0 {
			ptzSection := respStr[idx : idx+endIdx]
			if urlIdx := strings.Index(ptzSection, "<tt:XAddr>"); urlIdx >= 0 {
				if urlEnd := strings.Index(ptzSection[urlIdx:], "</tt:XAddr>"); urlEnd >= 0 {
					c.ptzURL = ptzSection[urlIdx+10 : urlIdx+urlEnd]
				}
			}
		}
	}

	// Default URLs if not found
	if c.mediaURL == "" {
		c.mediaURL = c.baseURL() + "/onvif/media_service"
	}
	if c.ptzURL == "" {
		c.ptzURL = c.baseURL() + "/onvif/ptz_service"
	}

	return nil
}

// GetProfiles retrieves media profiles
func (c *PTZClient) GetProfiles(ctx context.Context) ([]string, error) {
	if c.mediaURL == "" {
		if err := c.GetCapabilities(ctx); err != nil {
			return nil, err
		}
	}

	body := `<GetProfiles xmlns="http://www.onvif.org/ver10/media/wsdl"/>`

	respData, err := c.soapRequest(ctx, c.mediaURL, "", body)
	if err != nil {
		return nil, err
	}

	// Parse profiles
	var profiles []string
	respStr := string(respData)

	// Find all profile tokens
	search := respStr
	for {
		if idx := strings.Index(search, "token=\""); idx < 0 {
			break
		} else {
			search = search[idx+7:]
			if endIdx := strings.Index(search, "\""); endIdx >= 0 {
				token := search[:endIdx]
				if !strings.Contains(token, " ") && len(token) < 50 {
					profiles = append(profiles, token)
				}
				search = search[endIdx:]
			}
		}
	}

	if len(profiles) > 0 {
		c.profileToken = profiles[0]
	}

	return profiles, nil
}

// GetStatus returns PTZ status
func (c *PTZClient) GetStatus(ctx context.Context) (*PTZStatus, error) {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return &PTZStatus{Connected: false, Host: c.host}, nil
		}
	}

	if c.ptzURL == "" {
		if err := c.GetCapabilities(ctx); err != nil {
			return &PTZStatus{Connected: false, Host: c.host}, nil
		}
	}

	body := fmt.Sprintf(`<GetStatus xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
	</GetStatus>`, c.profileToken)

	respData, err := c.soapRequest(ctx, c.ptzURL, "", body)
	if err != nil {
		return &PTZStatus{Connected: false, Host: c.host}, nil
	}

	status := &PTZStatus{
		Connected:    true,
		Host:         c.host,
		ProfileToken: c.profileToken,
		Position:     PTZPosition{},
	}

	// Parse position from response
	respStr := string(respData)
	status.Position = parsePosition(respStr)

	// Check if moving
	if strings.Contains(respStr, "<MoveStatus>") {
		status.Moving = strings.Contains(respStr, ">MOVING<") || strings.Contains(respStr, ">Moving<")
	}

	return status, nil
}

// parsePosition extracts PTZ position from SOAP response
func parsePosition(respStr string) PTZPosition {
	pos := PTZPosition{}

	// Parse pan/tilt
	if idx := strings.Index(respStr, "<PanTilt"); idx >= 0 {
		section := respStr[idx:]
		if endIdx := strings.Index(section, "/>"); endIdx >= 0 || endIdx == -1 {
			if endIdx == -1 {
				endIdx = strings.Index(section, ">")
			}
			if endIdx > 0 {
				ptSection := section[:endIdx]
				if xIdx := strings.Index(ptSection, "x=\""); xIdx >= 0 {
					fmt.Sscanf(ptSection[xIdx+3:], "%f", &pos.Pan)
				}
				if yIdx := strings.Index(ptSection, "y=\""); yIdx >= 0 {
					fmt.Sscanf(ptSection[yIdx+3:], "%f", &pos.Tilt)
				}
			}
		}
	}

	// Parse zoom
	if idx := strings.Index(respStr, "<Zoom"); idx >= 0 {
		section := respStr[idx:]
		if endIdx := strings.Index(section, "/>"); endIdx >= 0 || endIdx == -1 {
			if endIdx == -1 {
				endIdx = strings.Index(section, ">")
			}
			if endIdx > 0 {
				zSection := section[:endIdx]
				if xIdx := strings.Index(zSection, "x=\""); xIdx >= 0 {
					fmt.Sscanf(zSection[xIdx+3:], "%f", &pos.Zoom)
				}
			}
		}
	}

	return pos
}

// ContinuousMove starts continuous PTZ movement
func (c *PTZClient) ContinuousMove(ctx context.Context, pan, tilt, zoom float64) error {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return err
		}
	}

	body := fmt.Sprintf(`<ContinuousMove xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
		<Velocity>
			<PanTilt xmlns="http://www.onvif.org/ver10/schema" x="%f" y="%f"/>
			<Zoom xmlns="http://www.onvif.org/ver10/schema" x="%f"/>
		</Velocity>
	</ContinuousMove>`, c.profileToken, pan, tilt, zoom)

	_, err := c.soapRequest(ctx, c.ptzURL, "", body)
	return err
}

// Stop stops all PTZ movement
func (c *PTZClient) Stop(ctx context.Context) error {
	if c.profileToken == "" {
		return fmt.Errorf("no profile token available")
	}

	body := fmt.Sprintf(`<Stop xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
		<PanTilt>true</PanTilt>
		<Zoom>true</Zoom>
	</Stop>`, c.profileToken)

	_, err := c.soapRequest(ctx, c.ptzURL, "", body)
	return err
}

// AbsoluteMove moves to an absolute position
func (c *PTZClient) AbsoluteMove(ctx context.Context, pan, tilt, zoom float64) error {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return err
		}
	}

	body := fmt.Sprintf(`<AbsoluteMove xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
		<Position>
			<PanTilt xmlns="http://www.onvif.org/ver10/schema" x="%f" y="%f"/>
			<Zoom xmlns="http://www.onvif.org/ver10/schema" x="%f"/>
		</Position>
	</AbsoluteMove>`, c.profileToken, pan, tilt, zoom)

	_, err := c.soapRequest(ctx, c.ptzURL, "", body)
	return err
}

// RelativeMove moves relative to current position
func (c *PTZClient) RelativeMove(ctx context.Context, pan, tilt, zoom float64) error {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return err
		}
	}

	body := fmt.Sprintf(`<RelativeMove xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
		<Translation>
			<PanTilt xmlns="http://www.onvif.org/ver10/schema" x="%f" y="%f"/>
			<Zoom xmlns="http://www.onvif.org/ver10/schema" x="%f"/>
		</Translation>
	</RelativeMove>`, c.profileToken, pan, tilt, zoom)

	_, err := c.soapRequest(ctx, c.ptzURL, "", body)
	return err
}

// GetPresets retrieves saved presets
func (c *PTZClient) GetPresets(ctx context.Context) ([]PTZPreset, error) {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return nil, err
		}
	}

	body := fmt.Sprintf(`<GetPresets xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
	</GetPresets>`, c.profileToken)

	respData, err := c.soapRequest(ctx, c.ptzURL, "", body)
	if err != nil {
		return nil, err
	}

	// Parse presets
	var presets []PTZPreset
	respStr := string(respData)

	// Simple parsing for preset tokens and names
	search := respStr
	for {
		presetIdx := strings.Index(search, "<tptz:Preset")
		if presetIdx < 0 {
			break
		}
		search = search[presetIdx:]

		endIdx := strings.Index(search, "</tptz:Preset>")
		if endIdx < 0 {
			break
		}
		presetSection := search[:endIdx]

		preset := PTZPreset{}

		// Get token
		if tokenIdx := strings.Index(presetSection, "token=\""); tokenIdx >= 0 {
			tokenEnd := strings.Index(presetSection[tokenIdx+7:], "\"")
			if tokenEnd >= 0 {
				preset.Token = presetSection[tokenIdx+7 : tokenIdx+7+tokenEnd]
			}
		}

		// Get name
		if nameIdx := strings.Index(presetSection, "<tt:Name>"); nameIdx >= 0 {
			nameEnd := strings.Index(presetSection[nameIdx:], "</tt:Name>")
			if nameEnd >= 0 {
				preset.Name = presetSection[nameIdx+9 : nameIdx+nameEnd]
			}
		}

		if preset.Token != "" {
			presets = append(presets, preset)
		}

		search = search[endIdx:]
	}

	return presets, nil
}

// GotoPreset moves to a saved preset
func (c *PTZClient) GotoPreset(ctx context.Context, presetToken string) error {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return err
		}
	}

	body := fmt.Sprintf(`<GotoPreset xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
		<PresetToken>%s</PresetToken>
	</GotoPreset>`, c.profileToken, presetToken)

	_, err := c.soapRequest(ctx, c.ptzURL, "", body)
	return err
}

// SetPreset saves current position as preset
func (c *PTZClient) SetPreset(ctx context.Context, name string) (*PTZPreset, error) {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return nil, err
		}
	}

	body := fmt.Sprintf(`<SetPreset xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
		<PresetName>%s</PresetName>
	</SetPreset>`, c.profileToken, name)

	respData, err := c.soapRequest(ctx, c.ptzURL, "", body)
	if err != nil {
		return nil, err
	}

	// Parse preset token from response
	respStr := string(respData)
	preset := &PTZPreset{Name: name}

	if tokenIdx := strings.Index(respStr, "<tptz:PresetToken>"); tokenIdx >= 0 {
		tokenEnd := strings.Index(respStr[tokenIdx:], "</tptz:PresetToken>")
		if tokenEnd >= 0 {
			preset.Token = respStr[tokenIdx+18 : tokenIdx+tokenEnd]
		}
	}

	return preset, nil
}

// GotoHome moves to home position
func (c *PTZClient) GotoHome(ctx context.Context) error {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return err
		}
	}

	body := fmt.Sprintf(`<GotoHomePosition xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
	</GotoHomePosition>`, c.profileToken)

	_, err := c.soapRequest(ctx, c.ptzURL, "", body)
	return err
}

// SetHome sets current position as home
func (c *PTZClient) SetHome(ctx context.Context) error {
	if c.profileToken == "" {
		if _, err := c.GetProfiles(ctx); err != nil {
			return err
		}
	}

	body := fmt.Sprintf(`<SetHomePosition xmlns="http://www.onvif.org/ver20/ptz/wsdl">
		<ProfileToken>%s</ProfileToken>
	</SetHomePosition>`, c.profileToken)

	_, err := c.soapRequest(ctx, c.ptzURL, "", body)
	return err
}

// GetHealth returns health status
func (c *PTZClient) GetHealth(ctx context.Context) (*PTZHealth, error) {
	health := &PTZHealth{
		Score:  100,
		Status: "healthy",
	}

	status, err := c.GetStatus(ctx)
	if err != nil || !status.Connected {
		health.Score = 0
		health.Status = "critical"
		health.Connected = false
		health.Issues = append(health.Issues, "Cannot connect to PTZ camera")
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Check network connectivity to %s:%s", c.host, c.port))
		health.Recommendations = append(health.Recommendations,
			"Verify camera is ONVIF-compatible and enabled")
		health.Recommendations = append(health.Recommendations,
			"Check PTZ_HOST, PTZ_USERNAME, PTZ_PASSWORD environment variables")
	} else {
		health.Connected = true
		health.CamerasFound = 1
	}

	return health, nil
}

// Host returns the configured host
func (c *PTZClient) Host() string {
	return c.host
}

// Internal XML types for parsing
type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    Body     `xml:"Body"`
}

type Body struct {
	Content []byte `xml:",innerxml"`
}

// PTZCameraConfig represents configuration for a single camera
type PTZCameraConfig struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// PTZTourStep represents a step in a preset tour
type PTZTourStep struct {
	PresetToken string  `json:"preset_token"`
	PresetName  string  `json:"preset_name,omitempty"`
	DwellTime   int     `json:"dwell_time"` // Seconds to stay at preset
	MoveSpeed   float64 `json:"move_speed"` // Speed to move to preset (0-1)
}

// PTZTour represents an automated preset tour
type PTZTour struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	CameraID string        `json:"camera_id"`
	Steps    []PTZTourStep `json:"steps"`
	Loop     bool          `json:"loop"`
	Active   bool          `json:"active"`
}

// PTZTourStatus represents the current tour execution status
type PTZTourStatus struct {
	TourID        string `json:"tour_id"`
	TourName      string `json:"tour_name"`
	Active        bool   `json:"active"`
	CurrentStep   int    `json:"current_step"`
	TotalSteps    int    `json:"total_steps"`
	CurrentPreset string `json:"current_preset"`
	TimeRemaining int    `json:"time_remaining"` // Seconds until next step
}

// PTZMultiClient manages multiple PTZ cameras
type PTZMultiClient struct {
	cameras     map[string]*PTZClient
	configs     map[string]*PTZCameraConfig
	tours       map[string]*PTZTour
	activeTours map[string]*tourRunner
}

// tourRunner manages active tour execution
type tourRunner struct {
	tour   *PTZTour
	stopCh chan struct{}
	status *PTZTourStatus
}

// NewPTZMultiClient creates a new multi-camera PTZ manager
func NewPTZMultiClient() (*PTZMultiClient, error) {
	mc := &PTZMultiClient{
		cameras:     make(map[string]*PTZClient),
		configs:     make(map[string]*PTZCameraConfig),
		tours:       make(map[string]*PTZTour),
		activeTours: make(map[string]*tourRunner),
	}

	// Load default camera from environment
	defaultHost := os.Getenv("PTZ_HOST")
	if defaultHost != "" {
		config := &PTZCameraConfig{
			ID:       "default",
			Name:     "Default Camera",
			Host:     defaultHost,
			Port:     os.Getenv("PTZ_PORT"),
			Username: os.Getenv("PTZ_USERNAME"),
			Password: os.Getenv("PTZ_PASSWORD"),
		}
		if config.Port == "" {
			config.Port = "80"
		}
		if config.Username == "" {
			config.Username = "admin"
		}
		mc.configs["default"] = config
	}

	// Load additional cameras from PTZ_CAMERAS env (JSON array)
	if camerasJSON := os.Getenv("PTZ_CAMERAS"); camerasJSON != "" {
		var cameras []PTZCameraConfig
		if err := parseJSONResponse([]byte(camerasJSON), &cameras); err == nil {
			for i := range cameras {
				cam := cameras[i]
				if cam.ID == "" {
					cam.ID = fmt.Sprintf("camera_%d", i+1)
				}
				mc.configs[cam.ID] = &cam
			}
		}
	}

	return mc, nil
}

// AddCamera adds a camera to the manager
func (mc *PTZMultiClient) AddCamera(config *PTZCameraConfig) error {
	if config.ID == "" {
		return fmt.Errorf("camera ID is required")
	}
	if config.Host == "" {
		return fmt.Errorf("camera host is required")
	}

	if config.Port == "" {
		config.Port = "80"
	}
	if config.Username == "" {
		config.Username = "admin"
	}

	mc.configs[config.ID] = config
	// Remove any cached client so it gets recreated with new config
	delete(mc.cameras, config.ID)

	return nil
}

// RemoveCamera removes a camera from the manager
func (mc *PTZMultiClient) RemoveCamera(cameraID string) error {
	if _, exists := mc.configs[cameraID]; !exists {
		return fmt.Errorf("camera not found: %s", cameraID)
	}

	// Stop any active tours for this camera
	for tourID, runner := range mc.activeTours {
		if mc.tours[tourID].CameraID == cameraID {
			close(runner.stopCh)
			delete(mc.activeTours, tourID)
		}
	}

	delete(mc.cameras, cameraID)
	delete(mc.configs, cameraID)
	return nil
}

// ListCameras returns all configured cameras
func (mc *PTZMultiClient) ListCameras() []*PTZCameraConfig {
	cameras := make([]*PTZCameraConfig, 0, len(mc.configs))
	for _, config := range mc.configs {
		cameras = append(cameras, config)
	}
	return cameras
}

// GetCamera returns a PTZ client for a specific camera
func (mc *PTZMultiClient) GetCamera(cameraID string) (*PTZClient, error) {
	// Return cached client if exists
	if client, exists := mc.cameras[cameraID]; exists {
		return client, nil
	}

	// Get config
	config, exists := mc.configs[cameraID]
	if !exists {
		return nil, fmt.Errorf("camera not found: %s", cameraID)
	}

	// Create new client
	client := &PTZClient{
		host:       config.Host,
		port:       config.Port,
		username:   config.Username,
		password:   config.Password,
		httpClient: httpclient.Fast(),
	}

	mc.cameras[cameraID] = client
	return client, nil
}

// GetAllCameraStatus returns status for all cameras
func (mc *PTZMultiClient) GetAllCameraStatus(ctx context.Context) map[string]*PTZStatus {
	statuses := make(map[string]*PTZStatus)

	for cameraID := range mc.configs {
		client, err := mc.GetCamera(cameraID)
		if err != nil {
			statuses[cameraID] = &PTZStatus{Connected: false}
			continue
		}

		status, err := client.GetStatus(ctx)
		if err != nil {
			statuses[cameraID] = &PTZStatus{Connected: false}
		} else {
			statuses[cameraID] = status
		}
	}

	return statuses
}

// CreateTour creates a new preset tour
func (mc *PTZMultiClient) CreateTour(tour *PTZTour) error {
	if tour.ID == "" {
		tour.ID = fmt.Sprintf("tour_%d", time.Now().UnixNano())
	}
	if tour.Name == "" {
		tour.Name = tour.ID
	}
	if tour.CameraID == "" {
		// Use default camera if only one configured
		if len(mc.configs) == 1 {
			for cameraID := range mc.configs {
				tour.CameraID = cameraID
				break
			}
		} else {
			return fmt.Errorf("camera_id is required when multiple cameras configured")
		}
	}
	if len(tour.Steps) == 0 {
		return fmt.Errorf("tour must have at least one step")
	}

	// Validate camera exists
	if _, exists := mc.configs[tour.CameraID]; !exists {
		return fmt.Errorf("camera not found: %s", tour.CameraID)
	}

	// Set default dwell times
	for i := range tour.Steps {
		if tour.Steps[i].DwellTime <= 0 {
			tour.Steps[i].DwellTime = 5 // 5 seconds default
		}
		if tour.Steps[i].MoveSpeed <= 0 {
			tour.Steps[i].MoveSpeed = 0.5 // Half speed default
		}
	}

	mc.tours[tour.ID] = tour
	return nil
}

// DeleteTour deletes a tour
func (mc *PTZMultiClient) DeleteTour(tourID string) error {
	// Stop if running
	if runner, exists := mc.activeTours[tourID]; exists {
		close(runner.stopCh)
		delete(mc.activeTours, tourID)
	}

	if _, exists := mc.tours[tourID]; !exists {
		return fmt.Errorf("tour not found: %s", tourID)
	}

	delete(mc.tours, tourID)
	return nil
}

// ListTours returns all configured tours
func (mc *PTZMultiClient) ListTours() []*PTZTour {
	tours := make([]*PTZTour, 0, len(mc.tours))
	for _, tour := range mc.tours {
		// Update active status
		_, tour.Active = mc.activeTours[tour.ID]
		tours = append(tours, tour)
	}
	return tours
}

// StartTour starts a preset tour
func (mc *PTZMultiClient) StartTour(ctx context.Context, tourID string) error {
	tour, exists := mc.tours[tourID]
	if !exists {
		return fmt.Errorf("tour not found: %s", tourID)
	}

	// Stop existing tour if running
	if runner, exists := mc.activeTours[tourID]; exists {
		close(runner.stopCh)
	}

	// Get camera client
	client, err := mc.GetCamera(tour.CameraID)
	if err != nil {
		return err
	}

	// Create tour runner
	runner := &tourRunner{
		tour:   tour,
		stopCh: make(chan struct{}),
		status: &PTZTourStatus{
			TourID:     tourID,
			TourName:   tour.Name,
			Active:     true,
			TotalSteps: len(tour.Steps),
		},
	}

	mc.activeTours[tourID] = runner

	// Start tour in background
	go func() {
		stepIdx := 0
		for {
			select {
			case <-runner.stopCh:
				runner.status.Active = false
				return
			default:
				step := tour.Steps[stepIdx]
				runner.status.CurrentStep = stepIdx + 1
				runner.status.CurrentPreset = step.PresetName
				if runner.status.CurrentPreset == "" {
					runner.status.CurrentPreset = step.PresetToken
				}

				// Move to preset
				_ = client.GotoPreset(ctx, step.PresetToken)

				// Dwell at preset
				runner.status.TimeRemaining = step.DwellTime
				for i := step.DwellTime; i > 0; i-- {
					select {
					case <-runner.stopCh:
						runner.status.Active = false
						return
					case <-time.After(time.Second):
						runner.status.TimeRemaining = i - 1
					}
				}

				// Move to next step
				stepIdx++
				if stepIdx >= len(tour.Steps) {
					if tour.Loop {
						stepIdx = 0
					} else {
						runner.status.Active = false
						delete(mc.activeTours, tourID)
						return
					}
				}
			}
		}
	}()

	return nil
}

// StopTour stops a running tour
func (mc *PTZMultiClient) StopTour(tourID string) error {
	runner, exists := mc.activeTours[tourID]
	if !exists {
		return fmt.Errorf("tour is not running: %s", tourID)
	}

	close(runner.stopCh)
	delete(mc.activeTours, tourID)
	return nil
}

// GetTourStatus returns the status of a tour
func (mc *PTZMultiClient) GetTourStatus(tourID string) (*PTZTourStatus, error) {
	runner, exists := mc.activeTours[tourID]
	if !exists {
		// Check if tour exists but not running
		if tour, exists := mc.tours[tourID]; exists {
			return &PTZTourStatus{
				TourID:     tourID,
				TourName:   tour.Name,
				Active:     false,
				TotalSteps: len(tour.Steps),
			}, nil
		}
		return nil, fmt.Errorf("tour not found: %s", tourID)
	}

	return runner.status, nil
}

// CreateTourFromPresets creates a tour from all presets on a camera
func (mc *PTZMultiClient) CreateTourFromPresets(ctx context.Context, cameraID, tourName string, dwellTime int, loop bool) (*PTZTour, error) {
	client, err := mc.GetCamera(cameraID)
	if err != nil {
		return nil, err
	}

	presets, err := client.GetPresets(ctx)
	if err != nil {
		return nil, err
	}

	if len(presets) == 0 {
		return nil, fmt.Errorf("no presets found on camera")
	}

	if dwellTime <= 0 {
		dwellTime = 5
	}

	steps := make([]PTZTourStep, len(presets))
	for i, preset := range presets {
		steps[i] = PTZTourStep{
			PresetToken: preset.Token,
			PresetName:  preset.Name,
			DwellTime:   dwellTime,
			MoveSpeed:   0.5,
		}
	}

	tour := &PTZTour{
		Name:     tourName,
		CameraID: cameraID,
		Steps:    steps,
		Loop:     loop,
	}

	if err := mc.CreateTour(tour); err != nil {
		return nil, err
	}

	return tour, nil
}
