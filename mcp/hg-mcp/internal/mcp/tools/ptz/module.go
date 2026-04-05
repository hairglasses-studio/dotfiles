// Package ptz provides MCP tools for ONVIF PTZ camera control.
package ptz

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the PTZ tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "ptz"
}

// Description returns the module description
func (m *Module) Description() string {
	return "ONVIF PTZ camera control for pan, tilt, zoom, and preset management"
}

// Tools returns the PTZ tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ptz_status",
				mcp.WithDescription("Get PTZ camera status including connection state and current position"),
			),
			Handler:             handlePTZStatus,
			Category:            "ptz",
			Subcategory:         "status",
			Tags:                []string{"ptz", "camera", "onvif", "status"},
			UseCases:            []string{"Check camera connection", "Get current position", "Verify PTZ availability"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_move",
				mcp.WithDescription("Move camera with pan, tilt, and zoom controls. Values range from -1.0 to 1.0"),
				mcp.WithNumber("pan", mcp.Description("Pan speed/position (-1.0=left, 0=stop, 1.0=right)"), mcp.Min(-1.0), mcp.Max(1.0)),
				mcp.WithNumber("tilt", mcp.Description("Tilt speed/position (-1.0=down, 0=stop, 1.0=up)"), mcp.Min(-1.0), mcp.Max(1.0)),
				mcp.WithNumber("zoom", mcp.Description("Zoom speed/position (-1.0=out, 0=stop, 1.0=in)"), mcp.Min(-1.0), mcp.Max(1.0)),
				mcp.WithString("mode", mcp.Description("Movement mode: continuous (keep moving), relative (move by amount), absolute (move to position)"), mcp.Enum("continuous", "relative", "absolute"), mcp.DefaultString("continuous")),
			),
			Handler:             handlePTZMove,
			Category:            "ptz",
			Subcategory:         "control",
			Tags:                []string{"ptz", "move", "pan", "tilt", "zoom"},
			UseCases:            []string{"Pan camera left/right", "Tilt camera up/down", "Zoom in/out"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_stop",
				mcp.WithDescription("Stop all PTZ movement immediately"),
			),
			Handler:             handlePTZStop,
			Category:            "ptz",
			Subcategory:         "control",
			Tags:                []string{"ptz", "stop", "halt"},
			UseCases:            []string{"Stop camera movement", "Emergency stop", "Halt all axes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_presets",
				mcp.WithDescription("List all saved PTZ presets"),
			),
			Handler:             handlePTZPresets,
			Category:            "ptz",
			Subcategory:         "presets",
			Tags:                []string{"ptz", "presets", "list"},
			UseCases:            []string{"List saved positions", "Find preset tokens", "View available shots"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_goto_preset",
				mcp.WithDescription("Move camera to a saved preset position"),
				mcp.WithString("preset", mcp.Required(), mcp.Description("Preset token or name to move to")),
			),
			Handler:             handlePTZGotoPreset,
			Category:            "ptz",
			Subcategory:         "presets",
			Tags:                []string{"ptz", "preset", "goto", "recall"},
			UseCases:            []string{"Recall saved position", "Move to preset shot", "Quick camera positioning"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_home",
				mcp.WithDescription("Move camera to home position"),
			),
			Handler:             handlePTZHome,
			Category:            "ptz",
			Subcategory:         "presets",
			Tags:                []string{"ptz", "home", "reset"},
			UseCases:            []string{"Return to home", "Reset camera position", "Center camera"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_health",
				mcp.WithDescription("Check PTZ camera connection health and get troubleshooting recommendations"),
			),
			Handler:             handlePTZHealth,
			Category:            "ptz",
			Subcategory:         "status",
			Tags:                []string{"ptz", "health", "diagnostics", "troubleshooting"},
			UseCases:            []string{"Diagnose connection issues", "Check camera health", "Get troubleshooting tips"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
		},
		// Multi-camera management tools
		{
			Tool: mcp.NewTool("aftrs_ptz_cameras",
				mcp.WithDescription("List all configured PTZ cameras with connection status"),
			),
			Handler:             handlePTZCameras,
			Category:            "ptz",
			Subcategory:         "multicam",
			Tags:                []string{"ptz", "cameras", "list", "multicam"},
			UseCases:            []string{"List all cameras", "Check camera status", "View camera fleet"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_camera_add",
				mcp.WithDescription("Add a new PTZ camera to the system"),
				mcp.WithString("id", mcp.Required(), mcp.Description("Unique camera identifier")),
				mcp.WithString("name", mcp.Description("Human-readable camera name")),
				mcp.WithString("host", mcp.Required(), mcp.Description("Camera IP address or hostname")),
				mcp.WithString("port", mcp.Description("ONVIF port (default: 80)")),
				mcp.WithString("username", mcp.Description("Camera username (default: admin)")),
				mcp.WithString("password", mcp.Description("Camera password")),
			),
			Handler:             handlePTZCameraAdd,
			Category:            "ptz",
			Subcategory:         "multicam",
			Tags:                []string{"ptz", "camera", "add", "configure"},
			UseCases:            []string{"Add new camera", "Configure camera", "Expand camera fleet"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_camera_control",
				mcp.WithDescription("Control a specific camera by ID (pan, tilt, zoom, preset)"),
				mcp.WithString("camera_id", mcp.Required(), mcp.Description("Camera identifier to control")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: move, stop, preset, home"), mcp.Enum("move", "stop", "preset", "home")),
				mcp.WithNumber("pan", mcp.Description("Pan value (-1.0 to 1.0) for move action")),
				mcp.WithNumber("tilt", mcp.Description("Tilt value (-1.0 to 1.0) for move action")),
				mcp.WithNumber("zoom", mcp.Description("Zoom value (-1.0 to 1.0) for move action")),
				mcp.WithString("preset", mcp.Description("Preset token for preset action")),
			),
			Handler:             handlePTZCameraControl,
			Category:            "ptz",
			Subcategory:         "multicam",
			Tags:                []string{"ptz", "camera", "control", "multicam"},
			UseCases:            []string{"Control specific camera", "Multi-camera operation", "Remote camera control"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		// Tour/patrol tools
		{
			Tool: mcp.NewTool("aftrs_ptz_tour_create",
				mcp.WithDescription("Create an automated preset tour for a camera"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Tour name")),
				mcp.WithString("camera_id", mcp.Description("Camera to run tour on (default if only one camera)")),
				mcp.WithArray("presets", mcp.Description("List of preset tokens to visit"), mcp.WithStringItems()),
				mcp.WithNumber("dwell_time", mcp.Description("Seconds to stay at each preset (default: 5)")),
				mcp.WithBoolean("loop", mcp.Description("Loop tour continuously (default: false)")),
				mcp.WithBoolean("from_all_presets", mcp.Description("Create tour from all camera presets")),
			),
			Handler:             handlePTZTourCreate,
			Category:            "ptz",
			Subcategory:         "tour",
			Tags:                []string{"ptz", "tour", "patrol", "automation"},
			UseCases:            []string{"Create camera tour", "Automate camera patrol", "Setup preset sequence"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_tours",
				mcp.WithDescription("List all configured camera tours"),
			),
			Handler:             handlePTZTours,
			Category:            "ptz",
			Subcategory:         "tour",
			Tags:                []string{"ptz", "tour", "list"},
			UseCases:            []string{"List tours", "View configured patrols", "Check tour status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_tour_start",
				mcp.WithDescription("Start a camera tour"),
				mcp.WithString("tour_id", mcp.Required(), mcp.Description("Tour ID to start")),
			),
			Handler:             handlePTZTourStart,
			Category:            "ptz",
			Subcategory:         "tour",
			Tags:                []string{"ptz", "tour", "start", "patrol"},
			UseCases:            []string{"Start tour", "Begin patrol", "Activate automation"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_tour_stop",
				mcp.WithDescription("Stop a running camera tour"),
				mcp.WithString("tour_id", mcp.Required(), mcp.Description("Tour ID to stop")),
			),
			Handler:             handlePTZTourStop,
			Category:            "ptz",
			Subcategory:         "tour",
			Tags:                []string{"ptz", "tour", "stop", "halt"},
			UseCases:            []string{"Stop tour", "Halt patrol", "Pause automation"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_tour_status",
				mcp.WithDescription("Get status of a camera tour"),
				mcp.WithString("tour_id", mcp.Required(), mcp.Description("Tour ID to check")),
			),
			Handler:             handlePTZTourStatus,
			Category:            "ptz",
			Subcategory:         "tour",
			Tags:                []string{"ptz", "tour", "status"},
			UseCases:            []string{"Check tour progress", "Monitor patrol", "View current preset"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
		},
	}
}

var getPTZClient = tools.LazyClient(clients.NewPTZClient)
var getPTZMultiClient = tools.LazyClient(clients.NewPTZMultiClient)

func handlePTZStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handlePTZMove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	pan := tools.GetFloatParam(req, "pan", 0)
	tilt := tools.GetFloatParam(req, "tilt", 0)
	zoom := tools.GetFloatParam(req, "zoom", 0)
	mode := tools.OptionalStringParam(req, "mode", "continuous")

	var moveErr error
	switch mode {
	case "continuous":
		moveErr = client.ContinuousMove(ctx, pan, tilt, zoom)
	case "relative":
		moveErr = client.RelativeMove(ctx, pan, tilt, zoom)
	case "absolute":
		moveErr = client.AbsoluteMove(ctx, pan, tilt, zoom)
	default:
		return tools.ErrorResult(fmt.Errorf("invalid mode: %s", mode)), nil
	}

	if moveErr != nil {
		return tools.ErrorResult(fmt.Errorf("failed to move: %w", moveErr)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"mode":    mode,
		"pan":     pan,
		"tilt":    tilt,
		"zoom":    zoom,
		"message": fmt.Sprintf("%s move command sent (pan=%.2f, tilt=%.2f, zoom=%.2f)", mode, pan, tilt, zoom),
	}), nil
}

func handlePTZStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	if err := client.Stop(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to stop: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "PTZ movement stopped",
	}), nil
}

func handlePTZPresets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	presets, err := client.GetPresets(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get presets: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"presets": presets,
		"count":   len(presets),
	}), nil
}

func handlePTZGotoPreset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	preset, errResult := tools.RequireStringParam(req, "preset")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.GotoPreset(ctx, preset); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to goto preset: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"preset":  preset,
		"message": fmt.Sprintf("Moving to preset: %s", preset),
	}), nil
}

func handlePTZHome(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	if err := client.GotoHome(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to goto home: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "Moving to home position",
	}), nil
}

func handlePTZHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}

// Multi-camera handlers

func handlePTZCameras(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZMultiClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	cameras := client.ListCameras()
	statuses := client.GetAllCameraStatus(ctx)

	return tools.JSONResult(map[string]interface{}{
		"cameras": cameras,
		"status":  statuses,
		"count":   len(cameras),
	}), nil
}

func handlePTZCameraAdd(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZMultiClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	config := &clients.PTZCameraConfig{
		ID:       tools.GetStringParam(req, "id"),
		Name:     tools.GetStringParam(req, "name"),
		Host:     tools.GetStringParam(req, "host"),
		Port:     tools.GetStringParam(req, "port"),
		Username: tools.GetStringParam(req, "username"),
		Password: tools.GetStringParam(req, "password"),
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	if err := client.AddCamera(config); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"camera":  config,
		"message": fmt.Sprintf("Camera '%s' added successfully", config.ID),
	}), nil
}

func handlePTZCameraControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	multiClient, err := getPTZMultiClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	cameraID := tools.GetStringParam(req, "camera_id")
	action := tools.GetStringParam(req, "action")

	client, err := multiClient.GetCamera(cameraID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var actionErr error
	var message string

	switch action {
	case "move":
		pan := tools.GetFloatParam(req, "pan", 0)
		tilt := tools.GetFloatParam(req, "tilt", 0)
		zoom := tools.GetFloatParam(req, "zoom", 0)
		actionErr = client.ContinuousMove(ctx, pan, tilt, zoom)
		message = fmt.Sprintf("Moving camera %s (pan=%.2f, tilt=%.2f, zoom=%.2f)", cameraID, pan, tilt, zoom)
	case "stop":
		actionErr = client.Stop(ctx)
		message = fmt.Sprintf("Stopped camera %s", cameraID)
	case "preset":
		preset, errResult := tools.RequireStringParam(req, "preset")
		if errResult != nil {
			return errResult, nil
		}
		actionErr = client.GotoPreset(ctx, preset)
		message = fmt.Sprintf("Moving camera %s to preset %s", cameraID, preset)
	case "home":
		actionErr = client.GotoHome(ctx)
		message = fmt.Sprintf("Moving camera %s to home position", cameraID)
	default:
		return tools.ErrorResult(fmt.Errorf("unknown action: %s", action)), nil
	}

	if actionErr != nil {
		return tools.ErrorResult(actionErr), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":   true,
		"camera_id": cameraID,
		"action":    action,
		"message":   message,
	}), nil
}

// Tour handlers

func handlePTZTourCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZMultiClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	name := tools.GetStringParam(req, "name")
	cameraID := tools.GetStringParam(req, "camera_id")
	dwellTime := tools.GetIntParam(req, "dwell_time", 5)
	loop := tools.GetBoolParam(req, "loop", false)
	fromAllPresets := tools.GetBoolParam(req, "from_all_presets", false)

	var tour *clients.PTZTour

	if fromAllPresets {
		// Create from all camera presets
		if cameraID == "" {
			cameras := client.ListCameras()
			if len(cameras) == 1 {
				cameraID = cameras[0].ID
			} else {
				return tools.ErrorResult(fmt.Errorf("camera_id required when multiple cameras configured")), nil
			}
		}

		var createErr error
		tour, createErr = client.CreateTourFromPresets(ctx, cameraID, name, dwellTime, loop)
		if createErr != nil {
			return tools.ErrorResult(createErr), nil
		}
	} else {
		// Create from specified presets
		presets := tools.GetStringArrayParam(req, "presets")

		if len(presets) == 0 {
			return tools.ErrorResult(fmt.Errorf("presets or from_all_presets is required")), nil
		}

		steps := make([]clients.PTZTourStep, len(presets))
		for i, p := range presets {
			steps[i] = clients.PTZTourStep{
				PresetToken: p,
				DwellTime:   dwellTime,
				MoveSpeed:   0.5,
			}
		}

		tour = &clients.PTZTour{
			Name:     name,
			CameraID: cameraID,
			Steps:    steps,
			Loop:     loop,
		}

		if err := client.CreateTour(tour); err != nil {
			return tools.ErrorResult(err), nil
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"tour":    tour,
		"message": fmt.Sprintf("Tour '%s' created with %d steps", tour.Name, len(tour.Steps)),
	}), nil
}

func handlePTZTours(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZMultiClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	tours := client.ListTours()

	return tools.JSONResult(map[string]interface{}{
		"tours": tours,
		"count": len(tours),
	}), nil
}

func handlePTZTourStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZMultiClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	tourID, errResult := tools.RequireStringParam(req, "tour_id")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.StartTour(ctx, tourID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"tour_id": tourID,
		"message": fmt.Sprintf("Tour '%s' started", tourID),
	}), nil
}

func handlePTZTourStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZMultiClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	tourID, errResult := tools.RequireStringParam(req, "tour_id")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.StopTour(tourID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"tour_id": tourID,
		"message": fmt.Sprintf("Tour '%s' stopped", tourID),
	}), nil
}

func handlePTZTourStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getPTZMultiClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create PTZ client: %w", err)), nil
	}

	tourID, errResult := tools.RequireStringParam(req, "tour_id")
	if errResult != nil {
		return errResult, nil
	}

	status, err := client.GetTourStatus(tourID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(status), nil
}
