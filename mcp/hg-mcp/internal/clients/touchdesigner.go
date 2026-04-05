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
	"strings"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// TouchDesignerClient provides access to TouchDesigner's external API
type TouchDesignerClient struct {
	host       string
	port       string
	httpClient *http.Client
}

// TDStatus represents TouchDesigner project status
type TDStatus struct {
	Connected    bool    `json:"connected"`
	ProjectName  string  `json:"project_name"`
	FPS          float64 `json:"fps"`
	RealTimeFPS  float64 `json:"realtime_fps"`
	CookTime     float64 `json:"cook_time_ms"`
	GPUMemory    string  `json:"gpu_memory"`
	CPUUsage     float64 `json:"cpu_usage"`
	ErrorCount   int     `json:"error_count"`
	WarningCount int     `json:"warning_count"`
	Version      string  `json:"version"`
}

// TDOperator represents a TouchDesigner operator
type TDOperator struct {
	Name       string            `json:"name"`
	Path       string            `json:"path"`
	Type       string            `json:"type"`
	Family     string            `json:"family"`
	HasErrors  bool              `json:"has_errors"`
	CookTime   float64           `json:"cook_time_ms"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

// TDNetworkHealth represents health metrics for an operator network
type TDNetworkHealth struct {
	Score           int          `json:"score"`
	Status          string       `json:"status"`
	TotalOperators  int          `json:"total_operators"`
	ErrorCount      int          `json:"error_count"`
	WarningCount    int          `json:"warning_count"`
	SlowOperators   []TDOperator `json:"slow_operators,omitempty"`
	Recommendations []string     `json:"recommendations,omitempty"`
}

// TDExecuteResult represents the result of Python execution
type TDExecuteResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// NewTouchDesignerClient creates a new TouchDesigner client
func NewTouchDesignerClient() (*TouchDesignerClient, error) {
	host := os.Getenv("TD_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TD_PORT")
	if port == "" {
		port = "9980" // Default TD WebServer port
	}

	return &TouchDesignerClient{
		host: host,
		port: port,
		httpClient: httpclient.Fast(),
	}, nil
}

// baseURL returns the base URL for TD API
func (c *TouchDesignerClient) baseURL() string {
	return fmt.Sprintf("http://%s:%s", c.host, c.port)
}

// doRequest performs an HTTP request to TouchDesigner
func (c *TouchDesignerClient) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	url := c.baseURL() + endpoint

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

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
		return nil, fmt.Errorf("TD API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetStatus returns the current TouchDesigner project status
func (c *TouchDesignerClient) GetStatus(ctx context.Context) (*TDStatus, error) {
	respBody, err := c.doRequest(ctx, "GET", "/status", nil)
	if err != nil {
		// Return disconnected status on error
		return &TDStatus{Connected: false}, nil
	}

	var status TDStatus
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status response: %w", err)
	}

	return &status, nil
}

// GetOperators lists operators in a network path
func (c *TouchDesignerClient) GetOperators(ctx context.Context, path string) ([]TDOperator, error) {
	if path == "" {
		path = "/"
	}

	endpoint := fmt.Sprintf("/operators?path=%s", path)
	respBody, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Operators []TDOperator `json:"operators"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse operators response: %w", err)
	}

	return result.Operators, nil
}

// GetParameters gets parameters for a specific operator
func (c *TouchDesignerClient) GetParameters(ctx context.Context, operatorPath string) (map[string]interface{}, error) {
	if operatorPath == "" {
		return nil, fmt.Errorf("operator path is required")
	}

	endpoint := fmt.Sprintf("/parameters?path=%s", operatorPath)
	respBody, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var params map[string]interface{}
	if err := json.Unmarshal(respBody, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters response: %w", err)
	}

	return params, nil
}

// SetParameter sets a parameter value on an operator
func (c *TouchDesignerClient) SetParameter(ctx context.Context, operatorPath, paramName string, value interface{}) error {
	if operatorPath == "" {
		return fmt.Errorf("operator path is required")
	}
	if paramName == "" {
		return fmt.Errorf("parameter name is required")
	}

	body := map[string]interface{}{
		"path":  operatorPath,
		"param": paramName,
		"value": value,
	}

	respBody, err := c.doRequest(ctx, "POST", "/parameters", body)
	if err != nil {
		return err
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to parse set parameter response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("failed to set parameter")
	}

	return nil
}

// ExecutePython executes Python code in TouchDesigner's context
func (c *TouchDesignerClient) ExecutePython(ctx context.Context, script string) (*TDExecuteResult, error) {
	if script == "" {
		return nil, fmt.Errorf("script is required")
	}

	body := map[string]interface{}{
		"code": script,
	}

	respBody, err := c.doRequest(ctx, "POST", "/execute", body)
	if err != nil {
		return &TDExecuteResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	var result TDExecuteResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse execute response: %w", err)
	}

	return &result, nil
}

// GetNetworkHealth analyzes an operator network and returns health metrics
func (c *TouchDesignerClient) GetNetworkHealth(ctx context.Context, path string) (*TDNetworkHealth, error) {
	if path == "" {
		path = "/"
	}

	endpoint := fmt.Sprintf("/network_health?path=%s", path)
	respBody, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var health TDNetworkHealth
	if err := json.Unmarshal(respBody, &health); err != nil {
		return nil, fmt.Errorf("failed to parse network health response: %w", err)
	}

	return &health, nil
}

// IsConnected checks if TouchDesigner is reachable
func (c *TouchDesignerClient) IsConnected(ctx context.Context) bool {
	respBody, err := c.doRequest(ctx, "GET", "/health", nil)
	if err != nil {
		return false
	}

	var health struct {
		Status    string `json:"status"`
		Connected bool   `json:"connected"`
	}
	if err := json.Unmarshal(respBody, &health); err != nil {
		return false
	}

	return health.Connected
}

// Host returns the configured host
func (c *TouchDesignerClient) Host() string {
	return c.host
}

// Port returns the configured port
func (c *TouchDesignerClient) Port() string {
	return c.port
}

// TDTexture represents a texture TOP
type TDTexture struct {
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	Format   string  `json:"format"`
	MemoryMB float64 `json:"memory_mb"`
}

// TDPerformance represents performance metrics
type TDPerformance struct {
	FPS          float64      `json:"fps"`
	CookTime     float64      `json:"cook_time_ms"`
	FrameTime    float64      `json:"frame_time_ms"`
	GPUTime      float64      `json:"gpu_time_ms"`
	CPUUsage     float64      `json:"cpu_usage"`
	TopCookTimes []TDOperator `json:"top_cook_times"`
}

// TDGPUMemory represents GPU memory usage
type TDGPUMemory struct {
	TotalMB     float64 `json:"total_mb"`
	UsedMB      float64 `json:"used_mb"`
	FreeMB      float64 `json:"free_mb"`
	TextureMB   float64 `json:"texture_mb"`
	BufferMB    float64 `json:"buffer_mb"`
	Utilization float64 `json:"utilization_percent"`
}

// TDError represents a TouchDesigner error
type TDError struct {
	Operator  string    `json:"operator"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"` // error, warning
	Timestamp time.Time `json:"timestamp"`
}

// TDTimeline represents a timeline CHOP
type TDTimeline struct {
	Name     string  `json:"name"`
	Path     string  `json:"path"`
	Length   float64 `json:"length_frames"`
	Position float64 `json:"position_frames"`
	Playing  bool    `json:"playing"`
	Loop     bool    `json:"loop"`
	Rate     float64 `json:"rate_fps"`
}

// TDComponent represents a container COMP
type TDComponent struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	Children  int    `json:"children"`
	HasErrors bool   `json:"has_errors"`
}

// TDProjectInfo represents detailed project information
type TDProjectInfo struct {
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	Version       string    `json:"version"`
	BuildVersion  string    `json:"build_version"`
	SaveTime      time.Time `json:"save_time"`
	OperatorCount int       `json:"operator_count"`
	CompCount     int       `json:"comp_count"`
	TOPCount      int       `json:"top_count"`
	CHOPCount     int       `json:"chop_count"`
	DATCount      int       `json:"dat_count"`
	SOPCount      int       `json:"sop_count"`
	MATCount      int       `json:"mat_count"`
}

// GetTextures lists texture TOPs in a network
func (c *TouchDesignerClient) GetTextures(ctx context.Context, path string) ([]TDTexture, error) {
	if path == "" {
		path = "/project1"
	}

	script := fmt.Sprintf(`
import json
textures = []
parent_op = op('%s')
if parent_op:
    for top in parent_op.findChildren(type=TOP, depth=1):
        try:
            textures.append({
                'name': top.name,
                'path': top.path,
                'width': top.width if hasattr(top, 'width') else 0,
                'height': top.height if hasattr(top, 'height') else 0,
                'format': str(top.par.format.eval()) if hasattr(top.par, 'format') else '',
                'memory_mb': 0
            })
        except:
            pass
print(json.dumps({'textures': textures}))
`, path)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get textures: %s", result.Error)
	}

	var data struct {
		Textures []TDTexture `json:"textures"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse textures: %w", err)
	}

	return data.Textures, nil
}

// GetPerformance returns performance profiling data
func (c *TouchDesignerClient) GetPerformance(ctx context.Context) (*TDPerformance, error) {
	respBody, err := c.doRequest(ctx, "GET", "/performance", nil)
	if err != nil {
		return nil, err
	}

	var perf TDPerformance
	if err := json.Unmarshal(respBody, &perf); err != nil {
		return nil, fmt.Errorf("failed to parse performance response: %w", err)
	}

	return &perf, nil
}

// GetGPUMemory returns GPU memory usage
func (c *TouchDesignerClient) GetGPUMemory(ctx context.Context) (*TDGPUMemory, error) {
	script := `
import json
import td
mem = {
    'total_mb': td.gpuMemTotal / 1024 / 1024 if hasattr(td, 'gpuMemTotal') else 0,
    'used_mb': td.gpuMemUsed / 1024 / 1024 if hasattr(td, 'gpuMemUsed') else 0,
    'free_mb': (td.gpuMemTotal - td.gpuMemUsed) / 1024 / 1024 if hasattr(td, 'gpuMemTotal') and hasattr(td, 'gpuMemUsed') else 0,
    'texture_mb': 0,
    'buffer_mb': 0,
    'utilization_percent': (td.gpuMemUsed / td.gpuMemTotal * 100) if hasattr(td, 'gpuMemTotal') and td.gpuMemTotal > 0 else 0
}
print(json.dumps(mem))
`
	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get GPU memory: %s", result.Error)
	}

	var mem TDGPUMemory
	if err := json.Unmarshal([]byte(result.Output), &mem); err != nil {
		return nil, fmt.Errorf("failed to parse GPU memory: %w", err)
	}

	return &mem, nil
}

// GetErrors returns current errors and warnings
func (c *TouchDesignerClient) GetErrors(ctx context.Context) ([]TDError, error) {
	respBody, err := c.doRequest(ctx, "GET", "/errors", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Errors []struct {
			Path  string `json:"path"`
			Error string `json:"error"`
		} `json:"errors"`
		Warnings []struct {
			Path    string `json:"path"`
			Warning string `json:"warning"`
		} `json:"warnings"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse errors response: %w", err)
	}

	var tdErrors []TDError
	for _, e := range result.Errors {
		tdErrors = append(tdErrors, TDError{
			Operator:  e.Path,
			Message:   e.Error,
			Severity:  "error",
			Timestamp: time.Now(),
		})
	}
	for _, w := range result.Warnings {
		tdErrors = append(tdErrors, TDError{
			Operator:  w.Path,
			Message:   w.Warning,
			Severity:  "warning",
			Timestamp: time.Now(),
		})
	}

	return tdErrors, nil
}

// BackupProject creates a backup of the current project
func (c *TouchDesignerClient) BackupProject(ctx context.Context, destination string) (string, error) {
	if destination == "" {
		destination = fmt.Sprintf("backup_%s.toe", time.Now().Format("20060102_150405"))
	}

	escapedDest := strings.ReplaceAll(destination, "\\", "\\\\")
	escapedDest = strings.ReplaceAll(escapedDest, "'", "\\'")

	script := fmt.Sprintf(`
import json
try:
    project.save('%s')
    print(json.dumps({'success': True, 'path': '%s'}))
except Exception as e:
    print(json.dumps({'success': False, 'error': str(e)}))
`, escapedDest, escapedDest)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return "", err
	}
	if !result.Success {
		return "", fmt.Errorf("failed to backup project: %s", result.Error)
	}

	var data struct {
		Success bool   `json:"success"`
		Path    string `json:"path"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return "", fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return "", fmt.Errorf("backup failed: %s", data.Error)
	}

	return data.Path, nil
}

// RecallPreset recalls a parameter preset
func (c *TouchDesignerClient) RecallPreset(ctx context.Context, presetPath string, operatorPath string) error {
	if presetPath == "" {
		return fmt.Errorf("preset path is required")
	}
	if operatorPath == "" {
		return fmt.Errorf("operator path is required")
	}

	escapedPreset := strings.ReplaceAll(presetPath, "\\", "\\\\")
	escapedPreset = strings.ReplaceAll(escapedPreset, "'", "\\'")

	script := fmt.Sprintf(`
import json
op_ref = op('%s')
if op_ref:
    try:
        with open('%s', 'r') as f:
            preset = json.load(f)
        errors = []
        for name, value in preset.items():
            try:
                if hasattr(op_ref.par, name):
                    getattr(op_ref.par, name).val = value
            except Exception as e:
                errors.append(f"{name}: {str(e)}")
        if errors:
            print(json.dumps({'success': False, 'error': '; '.join(errors)}))
        else:
            print(json.dumps({'success': True}))
    except Exception as e:
        print(json.dumps({'success': False, 'error': str(e)}))
else:
    print(json.dumps({'success': False, 'error': 'Operator not found'}))
`, operatorPath, escapedPreset)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to recall preset: %s", result.Error)
	}

	var data struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return fmt.Errorf("%s", data.Error)
	}

	return nil
}

// TriggerCue triggers a timeline cue
func (c *TouchDesignerClient) TriggerCue(ctx context.Context, timelinePath string, cue string) error {
	if timelinePath == "" {
		return fmt.Errorf("timeline path is required")
	}
	if cue == "" {
		return fmt.Errorf("cue value is required")
	}

	// Escape cue value for Python string
	escapedCue := strings.ReplaceAll(cue, "'", "\\'")

	script := fmt.Sprintf(`
import json
timeline = op('%s')
if timeline:
    try:
        timeline.par.cue = '%s'
        timeline.par.cuepulse.pulse()
        print(json.dumps({'success': True}))
    except Exception as e:
        print(json.dumps({'success': False, 'error': str(e)}))
else:
    print(json.dumps({'success': False, 'error': 'Timeline not found'}))
`, timelinePath, escapedCue)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to trigger cue: %s", result.Error)
	}

	var data struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return fmt.Errorf("%s", data.Error)
	}

	return nil
}

// GetTimelines lists timeline CHOPs
func (c *TouchDesignerClient) GetTimelines(ctx context.Context, path string) ([]TDTimeline, error) {
	if path == "" {
		path = "/project1"
	}

	script := fmt.Sprintf(`
import json
timelines = []
parent_op = op('%s')
if parent_op:
    for chop in parent_op.findChildren(type=CHOP, depth=1):
        if 'timeline' in chop.type.lower() or hasattr(chop.par, 'length'):
            try:
                timelines.append({
                    'name': chop.name,
                    'path': chop.path,
                    'length_frames': float(chop.par.length.eval()) if hasattr(chop.par, 'length') else 0,
                    'position_frames': float(chop.par.cue.eval()) if hasattr(chop.par, 'cue') else 0,
                    'playing': bool(chop.par.play.eval()) if hasattr(chop.par, 'play') else False,
                    'loop': bool(chop.par.loop.eval()) if hasattr(chop.par, 'loop') else False,
                    'rate_fps': float(chop.par.rate.eval()) if hasattr(chop.par, 'rate') else 60
                })
            except:
                pass
print(json.dumps({'timelines': timelines}))
`, path)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get timelines: %s", result.Error)
	}

	var data struct {
		Timelines []TDTimeline `json:"timelines"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse timelines: %w", err)
	}

	return data.Timelines, nil
}

// GetComponents lists container COMPs
func (c *TouchDesignerClient) GetComponents(ctx context.Context, path string) ([]TDComponent, error) {
	if path == "" {
		path = "/project1"
	}

	script := fmt.Sprintf(`
import json
components = []
parent_op = op('%s')
if parent_op:
    for comp in parent_op.findChildren(type=COMP, depth=1):
        try:
            components.append({
                'name': comp.name,
                'path': comp.path,
                'type': comp.type,
                'children': len(comp.children),
                'has_errors': len(comp.errors()) > 0 if hasattr(comp, 'errors') else False
            })
        except:
            pass
print(json.dumps({'components': components}))
`, path)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get components: %s", result.Error)
	}

	var data struct {
		Components []TDComponent `json:"components"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse components: %w", err)
	}

	return data.Components, nil
}

// GetProjectInfo returns detailed project information
func (c *TouchDesignerClient) GetProjectInfo(ctx context.Context) (*TDProjectInfo, error) {
	respBody, err := c.doRequest(ctx, "GET", "/project_info", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Name   string `json:"name"`
		Folder string `json:"folder"`
		File   string `json:"save_file"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse project info response: %w", err)
	}

	info := &TDProjectInfo{
		Name: result.Name,
		Path: result.File,
	}
	return info, nil
}

// TDContainer represents a container for creation
type TDContainer struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Type     string `json:"type"` // base, container, geo, etc.
	Template string `json:"template,omitempty"`
}

// TDCHOP represents a CHOP with channel data
type TDCHOP struct {
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	Type       string      `json:"type"`
	NumChans   int         `json:"num_channels"`
	Length     int         `json:"length"`
	SampleRate float64     `json:"sample_rate"`
	Channels   []TDChannel `json:"channels,omitempty"`
}

// TDChannel represents a single CHOP channel
type TDChannel struct {
	Name  string  `json:"name"`
	Index int     `json:"index"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Value float64 `json:"value"`
}

// TDDAT represents a DAT operator
type TDDAT struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Type    string `json:"type"`
	NumRows int    `json:"num_rows"`
	NumCols int    `json:"num_cols"`
	Content string `json:"content,omitempty"`
}

// TDRenderSettings represents render output settings
type TDRenderSettings struct {
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	FPS          float64 `json:"fps"`
	PixelFormat  string  `json:"pixel_format"`
	Antialiasing int     `json:"antialiasing"`
	OutputPath   string  `json:"output_path,omitempty"`
}

// TDCustomPar represents a custom parameter
type TDCustomPar struct {
	Name    string      `json:"name"`
	Label   string      `json:"label"`
	Type    string      `json:"type"` // Float, Int, Menu, Toggle, Str, etc.
	Value   interface{} `json:"value"`
	Default interface{} `json:"default"`
	Min     float64     `json:"min,omitempty"`
	Max     float64     `json:"max,omitempty"`
}

// TDTox represents a TOX file
type TDTox struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
}

// TDVariable represents a project variable
type TDVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"` // str, int, float
}

// TDSnapshot represents a parameter snapshot
type TDSnapshot struct {
	Name      string            `json:"name"`
	Timestamp time.Time         `json:"timestamp"`
	Operator  string            `json:"operator"`
	Params    map[string]string `json:"params"`
}

// CreateContainer creates a new container
func (c *TouchDesignerClient) CreateContainer(ctx context.Context, parentPath, name, containerType string) (*TDContainer, error) {
	if parentPath == "" {
		parentPath = "/project1"
	}
	if name == "" {
		return nil, fmt.Errorf("container name is required")
	}
	if containerType == "" {
		containerType = "container"
	}

	script := fmt.Sprintf(`
import json
parent = op('%s')
if parent:
    new_comp = parent.create(%sCOMP, '%s')
    result = {
        'name': new_comp.name,
        'path': new_comp.path,
        'type': new_comp.type
    }
    print(json.dumps(result))
else:
    print(json.dumps({'error': 'Parent not found'}))
`, parentPath, containerType, name)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to create container: %s", result.Error)
	}

	var data struct {
		Name  string `json:"name"`
		Path  string `json:"path"`
		Type  string `json:"type"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}
	if data.Error != "" {
		return nil, fmt.Errorf("%s", data.Error)
	}

	return &TDContainer{
		Name: data.Name,
		Path: data.Path,
		Type: data.Type,
	}, nil
}

// DeleteOperator deletes an operator
func (c *TouchDesignerClient) DeleteOperator(ctx context.Context, path string) error {
	if path == "" {
		return fmt.Errorf("operator path is required")
	}

	script := fmt.Sprintf(`
import json
target = op('%s')
if target:
    target.destroy()
    print(json.dumps({'success': True}))
else:
    print(json.dumps({'success': False, 'error': 'Operator not found'}))
`, path)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to delete operator: %s", result.Error)
	}

	var data struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return fmt.Errorf("%s", data.Error)
	}

	return nil
}

// GetCHOPs lists CHOP operators in a network
func (c *TouchDesignerClient) GetCHOPs(ctx context.Context, path string) ([]TDCHOP, error) {
	if path == "" {
		path = "/project1"
	}

	script := fmt.Sprintf(`
import json
chops = []
parent_op = op('%s')
if parent_op:
    for chop in parent_op.findChildren(type=CHOP, depth=1):
        try:
            chops.append({
                'name': chop.name,
                'path': chop.path,
                'type': chop.type,
                'num_channels': chop.numChans if hasattr(chop, 'numChans') else 0,
                'length': chop.numSamples if hasattr(chop, 'numSamples') else 0,
                'sample_rate': float(chop.rate) if hasattr(chop, 'rate') else 0
            })
        except:
            pass
print(json.dumps({'chops': chops}))
`, path)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get CHOPs: %s", result.Error)
	}

	var data struct {
		CHOPs []TDCHOP `json:"chops"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse CHOPs: %w", err)
	}

	return data.CHOPs, nil
}

// GetCHOPChannels gets channel values from a CHOP
func (c *TouchDesignerClient) GetCHOPChannels(ctx context.Context, chopPath string) (*TDCHOP, error) {
	if chopPath == "" {
		return nil, fmt.Errorf("CHOP path is required")
	}

	script := fmt.Sprintf(`
import json
chop_op = op('%s')
if chop_op and chop_op.isCHOP:
    channels = []
    for i, chan in enumerate(chop_op.chans()):
        channels.append({
            'name': chan.name,
            'index': i,
            'min': float(chan.min()) if hasattr(chan, 'min') else 0,
            'max': float(chan.max()) if hasattr(chan, 'max') else 0,
            'value': float(chan.eval()) if hasattr(chan, 'eval') else 0
        })
    result = {
        'name': chop_op.name,
        'path': chop_op.path,
        'type': chop_op.type,
        'num_channels': chop_op.numChans,
        'length': chop_op.numSamples,
        'sample_rate': float(chop_op.rate),
        'channels': channels
    }
    print(json.dumps(result))
else:
    print(json.dumps({'error': 'CHOP not found or invalid'}))
`, chopPath)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get CHOP channels: %s", result.Error)
	}

	var data struct {
		Name       string      `json:"name"`
		Path       string      `json:"path"`
		Type       string      `json:"type"`
		NumChans   int         `json:"num_channels"`
		Length     int         `json:"length"`
		SampleRate float64     `json:"sample_rate"`
		Channels   []TDChannel `json:"channels"`
		Error      string      `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse CHOP: %w", err)
	}
	if data.Error != "" {
		return nil, fmt.Errorf("%s", data.Error)
	}

	return &TDCHOP{
		Name:       data.Name,
		Path:       data.Path,
		Type:       data.Type,
		NumChans:   data.NumChans,
		Length:     data.Length,
		SampleRate: data.SampleRate,
		Channels:   data.Channels,
	}, nil
}

// GetDATs lists DAT operators in a network
func (c *TouchDesignerClient) GetDATs(ctx context.Context, path string) ([]TDDAT, error) {
	if path == "" {
		path = "/project1"
	}

	script := fmt.Sprintf(`
import json
dats = []
parent_op = op('%s')
if parent_op:
    for dat in parent_op.findChildren(type=DAT, depth=1):
        try:
            dats.append({
                'name': dat.name,
                'path': dat.path,
                'type': dat.type,
                'num_rows': dat.numRows if hasattr(dat, 'numRows') else 0,
                'num_cols': dat.numCols if hasattr(dat, 'numCols') else 0
            })
        except:
            pass
print(json.dumps({'dats': dats}))
`, path)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get DATs: %s", result.Error)
	}

	var data struct {
		DATs []TDDAT `json:"dats"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse DATs: %w", err)
	}

	return data.DATs, nil
}

// GetDATContent gets the content of a DAT
func (c *TouchDesignerClient) GetDATContent(ctx context.Context, datPath string) (*TDDAT, error) {
	if datPath == "" {
		return nil, fmt.Errorf("DAT path is required")
	}

	script := fmt.Sprintf(`
import json
dat_op = op('%s')
if dat_op and dat_op.isDAT:
    result = {
        'name': dat_op.name,
        'path': dat_op.path,
        'type': dat_op.type,
        'num_rows': dat_op.numRows if hasattr(dat_op, 'numRows') else 0,
        'num_cols': dat_op.numCols if hasattr(dat_op, 'numCols') else 0,
        'content': dat_op.text if hasattr(dat_op, 'text') else ''
    }
    print(json.dumps(result))
else:
    print(json.dumps({'error': 'DAT not found or invalid'}))
`, datPath)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get DAT content: %s", result.Error)
	}

	var data struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Type    string `json:"type"`
		NumRows int    `json:"num_rows"`
		NumCols int    `json:"num_cols"`
		Content string `json:"content"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse DAT: %w", err)
	}
	if data.Error != "" {
		return nil, fmt.Errorf("%s", data.Error)
	}

	return &TDDAT{
		Name:    data.Name,
		Path:    data.Path,
		Type:    data.Type,
		NumRows: data.NumRows,
		NumCols: data.NumCols,
		Content: data.Content,
	}, nil
}

// SetDATContent sets the content of a DAT
func (c *TouchDesignerClient) SetDATContent(ctx context.Context, datPath, content string) error {
	if datPath == "" {
		return fmt.Errorf("DAT path is required")
	}

	// Escape content for Python string
	escapedContent := strings.ReplaceAll(content, "\\", "\\\\")
	escapedContent = strings.ReplaceAll(escapedContent, "'", "\\'")
	escapedContent = strings.ReplaceAll(escapedContent, "\n", "\\n")
	escapedContent = strings.ReplaceAll(escapedContent, "\r", "\\r")

	script := fmt.Sprintf(`
import json
dat_op = op('%s')
if dat_op and dat_op.isDAT:
    dat_op.text = '%s'
    print(json.dumps({'success': True}))
else:
    print(json.dumps({'success': False, 'error': 'DAT not found or invalid'}))
`, datPath, escapedContent)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to set DAT content: %s", result.Error)
	}

	var data struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return fmt.Errorf("%s", data.Error)
	}

	return nil
}

// GetRenderSettings gets render output settings
func (c *TouchDesignerClient) GetRenderSettings(ctx context.Context) (*TDRenderSettings, error) {
	script := `
import json
import td
result = {
    'width': project.cookWidth,
    'height': project.cookHeight,
    'fps': project.cookRate,
    'pixel_format': 'RGBA8',
    'antialiasing': 4
}
print(json.dumps(result))
`
	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get render settings: %s", result.Error)
	}

	var settings TDRenderSettings
	if err := json.Unmarshal([]byte(result.Output), &settings); err != nil {
		return nil, fmt.Errorf("failed to parse render settings: %w", err)
	}

	return &settings, nil
}

// SetRenderSettings sets render output settings
func (c *TouchDesignerClient) SetRenderSettings(ctx context.Context, settings *TDRenderSettings) error {
	if settings == nil {
		return fmt.Errorf("settings are required")
	}

	script := fmt.Sprintf(`
import json
project.cookWidth = %d
project.cookHeight = %d
project.cookRate = %f
print(json.dumps({'success': True}))
`, settings.Width, settings.Height, settings.FPS)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to set render settings: %s", result.Error)
	}

	return nil
}

// GetCustomPars gets custom parameters from an operator
func (c *TouchDesignerClient) GetCustomPars(ctx context.Context, operatorPath string) ([]TDCustomPar, error) {
	if operatorPath == "" {
		return nil, fmt.Errorf("operator path is required")
	}

	script := fmt.Sprintf(`
import json
op_ref = op('%s')
pars = []
if op_ref:
    for page in op_ref.customPages:
        for par in page.pars:
            try:
                pars.append({
                    'name': par.name,
                    'label': par.label,
                    'type': str(par.mode),
                    'value': par.eval() if par.mode != ParMode.PYTHON else str(par.expr),
                    'default': par.default,
                    'min': float(par.min) if par.min is not None else 0,
                    'max': float(par.max) if par.max is not None else 0
                })
            except:
                pass
print(json.dumps({'pars': pars}))
`, operatorPath)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get custom parameters: %s", result.Error)
	}

	var data struct {
		Pars []TDCustomPar `json:"pars"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	return data.Pars, nil
}

// ExportTox exports an operator as a TOX file
func (c *TouchDesignerClient) ExportTox(ctx context.Context, operatorPath, destination string) (*TDTox, error) {
	if operatorPath == "" {
		return nil, fmt.Errorf("operator path is required")
	}
	if destination == "" {
		return nil, fmt.Errorf("destination path is required")
	}

	escapedDest := strings.ReplaceAll(destination, "\\", "\\\\")

	script := fmt.Sprintf(`
import json
op_ref = op('%s')
if op_ref:
    try:
        op_ref.save('%s')
        result = {
            'name': op_ref.name,
            'path': '%s',
            'success': True
        }
        print(json.dumps(result))
    except Exception as e:
        print(json.dumps({'success': False, 'error': str(e)}))
else:
    print(json.dumps({'success': False, 'error': 'Operator not found'}))
`, operatorPath, escapedDest, escapedDest)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to export TOX: %s", result.Error)
	}

	var data struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return nil, fmt.Errorf("%s", data.Error)
	}

	return &TDTox{
		Name: data.Name,
		Path: data.Path,
	}, nil
}

// ImportTox imports a TOX file
func (c *TouchDesignerClient) ImportTox(ctx context.Context, toxPath, parentPath string) (*TDTox, error) {
	if toxPath == "" {
		return nil, fmt.Errorf("TOX path is required")
	}
	if parentPath == "" {
		parentPath = "/project1"
	}

	escapedToxPath := strings.ReplaceAll(toxPath, "\\", "\\\\")

	script := fmt.Sprintf(`
import json
parent = op('%s')
if parent:
    try:
        new_op = parent.loadTox('%s')
        result = {
            'name': new_op.name,
            'path': new_op.path,
            'success': True
        }
        print(json.dumps(result))
    except Exception as e:
        print(json.dumps({'success': False, 'error': str(e)}))
else:
    print(json.dumps({'success': False, 'error': 'Parent not found'}))
`, parentPath, escapedToxPath)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to import TOX: %s", result.Error)
	}

	var data struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return nil, fmt.Errorf("%s", data.Error)
	}

	return &TDTox{
		Name: data.Name,
		Path: data.Path,
	}, nil
}

// GetVariables gets project variables
func (c *TouchDesignerClient) GetVariables(ctx context.Context) ([]TDVariable, error) {
	script := `
import json
vars = []
for name in project.paths.keys():
    try:
        val = project.paths[name]
        vars.append({
            'name': name,
            'value': str(val),
            'type': 'str'
        })
    except:
        pass
print(json.dumps({'vars': vars}))
`
	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to get variables: %s", result.Error)
	}

	var data struct {
		Vars []TDVariable `json:"vars"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	return data.Vars, nil
}

// SetVariable sets a project variable
func (c *TouchDesignerClient) SetVariable(ctx context.Context, name, value string) error {
	if name == "" {
		return fmt.Errorf("variable name is required")
	}

	escapedValue := strings.ReplaceAll(value, "'", "\\'")

	script := fmt.Sprintf(`
import json
project.paths['%s'] = '%s'
print(json.dumps({'success': True}))
`, name, escapedValue)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to set variable: %s", result.Error)
	}

	return nil
}

// CreateSnapshot creates a parameter snapshot
func (c *TouchDesignerClient) CreateSnapshot(ctx context.Context, operatorPath, snapshotName string) (*TDSnapshot, error) {
	if operatorPath == "" {
		return nil, fmt.Errorf("operator path is required")
	}
	if snapshotName == "" {
		snapshotName = fmt.Sprintf("snapshot_%s", time.Now().Format("20060102_150405"))
	}

	script := fmt.Sprintf(`
import json
import datetime
op_ref = op('%s')
params = {}
if op_ref:
    for par in op_ref.pars():
        try:
            params[par.name] = str(par.eval())
        except:
            pass
    result = {
        'name': '%s',
        'timestamp': datetime.datetime.now().isoformat(),
        'operator': '%s',
        'params': params
    }
    print(json.dumps(result))
else:
    print(json.dumps({'error': 'Operator not found'}))
`, operatorPath, snapshotName, operatorPath)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("failed to create snapshot: %s", result.Error)
	}

	var data struct {
		Name      string            `json:"name"`
		Timestamp string            `json:"timestamp"`
		Operator  string            `json:"operator"`
		Params    map[string]string `json:"params"`
		Error     string            `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot: %w", err)
	}
	if data.Error != "" {
		return nil, fmt.Errorf("%s", data.Error)
	}

	timestamp, _ := time.Parse(time.RFC3339, data.Timestamp)

	return &TDSnapshot{
		Name:      data.Name,
		Timestamp: timestamp,
		Operator:  data.Operator,
		Params:    data.Params,
	}, nil
}

// RestoreSnapshot restores a parameter snapshot
func (c *TouchDesignerClient) RestoreSnapshot(ctx context.Context, operatorPath string, snapshot *TDSnapshot) error {
	if operatorPath == "" {
		return fmt.Errorf("operator path is required")
	}
	if snapshot == nil || len(snapshot.Params) == 0 {
		return fmt.Errorf("snapshot with parameters is required")
	}

	// Convert params map to Python dict string
	paramsJSON, err := json.Marshal(snapshot.Params)
	if err != nil {
		return fmt.Errorf("failed to serialize params: %w", err)
	}

	script := fmt.Sprintf(`
import json
op_ref = op('%s')
params = %s
if op_ref:
    errors = []
    for name, value in params.items():
        try:
            if hasattr(op_ref.par, name):
                getattr(op_ref.par, name).val = value
        except Exception as e:
            errors.append(f"{name}: {str(e)}")
    if errors:
        print(json.dumps({'success': False, 'error': '; '.join(errors)}))
    else:
        print(json.dumps({'success': True}))
else:
    print(json.dumps({'success': False, 'error': 'Operator not found'}))
`, operatorPath, string(paramsJSON))

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to restore snapshot: %s", result.Error)
	}

	var data struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return fmt.Errorf("%s", data.Error)
	}

	return nil
}

// PulseParameter pulses a pulse-type parameter
func (c *TouchDesignerClient) PulseParameter(ctx context.Context, operatorPath, paramName string) error {
	if operatorPath == "" {
		return fmt.Errorf("operator path is required")
	}
	if paramName == "" {
		return fmt.Errorf("parameter name is required")
	}

	script := fmt.Sprintf(`
import json
op_ref = op('%s')
if op_ref and hasattr(op_ref.par, '%s'):
    getattr(op_ref.par, '%s').pulse()
    print(json.dumps({'success': True}))
else:
    print(json.dumps({'success': False, 'error': 'Operator or parameter not found'}))
`, operatorPath, paramName, paramName)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to pulse parameter: %s", result.Error)
	}

	var data struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return fmt.Errorf("%s", data.Error)
	}

	return nil
}

// ResetOperator resets an operator to defaults
func (c *TouchDesignerClient) ResetOperator(ctx context.Context, operatorPath string) error {
	if operatorPath == "" {
		return fmt.Errorf("operator path is required")
	}

	script := fmt.Sprintf(`
import json
op_ref = op('%s')
if op_ref:
    for par in op_ref.pars():
        try:
            par.val = par.default
        except:
            pass
    print(json.dumps({'success': True}))
else:
    print(json.dumps({'success': False, 'error': 'Operator not found'}))
`, operatorPath)

	result, err := c.ExecutePython(ctx, script)
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("failed to reset operator: %s", result.Error)
	}

	var data struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.Output), &data); err != nil {
		return fmt.Errorf("failed to parse result: %w", err)
	}
	if !data.Success {
		return fmt.Errorf("%s", data.Error)
	}

	return nil
}
