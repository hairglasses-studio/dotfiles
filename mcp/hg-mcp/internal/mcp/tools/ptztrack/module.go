// Package ptztrack provides PTZ camera auto-tracking via NDI computer vision.
package ptztrack

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for PTZ Auto-Tracking
type Module struct{}

func (m *Module) Name() string {
	return "ptztrack"
}

func (m *Module) Description() string {
	return "PTZ camera auto-tracking using NDI computer vision"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ptz_track_face",
				mcp.WithDescription("Start face tracking mode - camera follows detected faces."),
				mcp.WithString("ndi_source", mcp.Required(), mcp.Description("NDI source to analyze for faces")),
				mcp.WithString("camera_id", mcp.Description("PTZ camera to control (default if only one)")),
				mcp.WithNumber("speed", mcp.Description("Tracking speed 0.1-1.0 (default: 0.3)")),
				mcp.WithNumber("smoothing", mcp.Description("Movement smoothing 0.1-1.0 (default: 0.5)")),
				mcp.WithBoolean("zoom_to_face", mcp.Description("Auto-zoom to keep face at consistent size")),
			),
			Handler:             handleTrackFace,
			Category:            "ptztrack",
			Subcategory:         "tracking",
			Tags:                []string{"ptz", "ndi", "face", "tracking", "cv", "autofollow"},
			UseCases:            []string{"Auto-follow presenter", "Face tracking camera", "Subject tracking"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_track_motion",
				mcp.WithDescription("Start motion tracking mode - camera follows motion in frame."),
				mcp.WithString("ndi_source", mcp.Required(), mcp.Description("NDI source to analyze for motion")),
				mcp.WithString("camera_id", mcp.Description("PTZ camera to control (default if only one)")),
				mcp.WithNumber("speed", mcp.Description("Tracking speed 0.1-1.0 (default: 0.3)")),
				mcp.WithNumber("threshold", mcp.Description("Motion detection threshold % (default: 5)")),
			),
			Handler:             handleTrackMotion,
			Category:            "ptztrack",
			Subcategory:         "tracking",
			Tags:                []string{"ptz", "ndi", "motion", "tracking", "cv", "autofollow"},
			UseCases:            []string{"Motion-based tracking", "Activity following", "Event tracking"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_track_stop",
				mcp.WithDescription("Stop all auto-tracking on a camera."),
				mcp.WithString("camera_id", mcp.Description("PTZ camera to stop tracking on")),
			),
			Handler:             handleTrackStop,
			Category:            "ptztrack",
			Subcategory:         "tracking",
			Tags:                []string{"ptz", "tracking", "stop"},
			UseCases:            []string{"Stop tracking", "Manual control", "Disable autofollow"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_track_status",
				mcp.WithDescription("Get status of auto-tracking systems."),
				mcp.WithString("camera_id", mcp.Description("PTZ camera to check (all if not specified)")),
			),
			Handler:             handleTrackStatus,
			Category:            "ptztrack",
			Subcategory:         "status",
			Tags:                []string{"ptz", "tracking", "status"},
			UseCases:            []string{"Check tracking status", "Monitor autofollow", "View tracking state"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
		},
		{
			Tool: mcp.NewTool("aftrs_ptz_track_snapshot",
				mcp.WithDescription("Capture frame from NDI source and analyze for tracking targets."),
				mcp.WithString("ndi_source", mcp.Required(), mcp.Description("NDI source to analyze")),
				mcp.WithBoolean("detect_faces", mcp.Description("Detect faces in frame (default: true)")),
				mcp.WithBoolean("detect_motion", mcp.Description("Compare with previous frame for motion")),
			),
			Handler:             handleTrackSnapshot,
			Category:            "ptztrack",
			Subcategory:         "analysis",
			Tags:                []string{"ptz", "ndi", "cv", "analysis", "snapshot"},
			UseCases:            []string{"Analyze frame for targets", "Preview tracking", "Test detection"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ptz",
		},
	}
}

// TrackingSession represents an active tracking session
type TrackingSession struct {
	CameraID   string    `json:"camera_id"`
	NDISource  string    `json:"ndi_source"`
	Mode       string    `json:"mode"` // face, motion
	Speed      float64   `json:"speed"`
	Active     bool      `json:"active"`
	StartTime  time.Time `json:"start_time"`
	FrameCount int       `json:"frame_count"`
	LastTarget string    `json:"last_target"`
	stopCh     chan struct{}
}

var (
	trackingSessions = make(map[string]*TrackingSession) // cameraID -> session
	sessionsMu       sync.RWMutex
)

// getNDIClient creates or returns the NDI client
func getNDIClient() (*clients.NDIClient, error) {
	return clients.NewNDIClient()
}

// getPTZMultiClient creates or returns the PTZ multi-client
func getPTZMultiClient() (*clients.PTZMultiClient, error) {
	return clients.NewPTZMultiClient()
}

// handleTrackFace starts face tracking
func handleTrackFace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ndiSource := tools.GetStringParam(req, "ndi_source")
	cameraID := tools.GetStringParam(req, "camera_id")
	speed := tools.GetFloatParam(req, "speed", 0.3)
	smoothing := tools.GetFloatParam(req, "smoothing", 0.5)
	zoomToFace := tools.GetBoolParam(req, "zoom_to_face", false)

	if speed < 0.1 {
		speed = 0.1
	} else if speed > 1.0 {
		speed = 1.0
	}

	ptzClient, err := getPTZMultiClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Default camera if not specified
	if cameraID == "" {
		cameras := ptzClient.ListCameras()
		if len(cameras) == 1 {
			cameraID = cameras[0].ID
		} else if len(cameras) == 0 {
			return tools.ErrorResult(fmt.Errorf("no PTZ cameras configured")), nil
		} else {
			return tools.ErrorResult(fmt.Errorf("camera_id required when multiple cameras configured")), nil
		}
	}

	ndiClient, err := getNDIClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	camClient, err := ptzClient.GetCamera(cameraID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Stop existing tracking on this camera
	sessionsMu.Lock()
	if existing, exists := trackingSessions[cameraID]; exists {
		close(existing.stopCh)
		delete(trackingSessions, cameraID)
	}

	// Create new session
	session := &TrackingSession{
		CameraID:  cameraID,
		NDISource: ndiSource,
		Mode:      "face",
		Speed:     speed,
		Active:    true,
		StartTime: time.Now(),
		stopCh:    make(chan struct{}),
	}
	trackingSessions[cameraID] = session
	sessionsMu.Unlock()

	// Start tracking goroutine
	go func() {
		var prevPan, prevTilt float64
		frameInterval := time.Millisecond * 200 // 5 FPS for tracking

		for {
			select {
			case <-session.stopCh:
				return
			case <-time.After(frameInterval):
				// Capture frame
				frame, err := ndiClient.CaptureFrame(ctx, ndiSource, "")
				if err != nil {
					continue
				}

				// Detect faces
				result, err := ndiClient.DetectFaces(ctx, frame.FilePath)
				if err != nil || result.FaceCount == 0 {
					continue
				}

				session.FrameCount++

				// Get largest/primary face
				var primaryFace clients.FaceRect
				maxArea := 0
				for _, face := range result.Faces {
					area := face.Width * face.Height
					if area > maxArea {
						maxArea = area
						primaryFace = face
					}
				}

				session.LastTarget = fmt.Sprintf("Face at (%d,%d)", primaryFace.X, primaryFace.Y)

				// Calculate center offset from frame center
				frameWidth := frame.Width
				frameHeight := frame.Height
				if frameWidth == 0 {
					frameWidth = 1920
				}
				if frameHeight == 0 {
					frameHeight = 1080
				}

				faceCenterX := primaryFace.X + primaryFace.Width/2
				faceCenterY := primaryFace.Y + primaryFace.Height/2
				frameCenterX := frameWidth / 2
				frameCenterY := frameHeight / 2

				// Calculate pan/tilt adjustment (-1 to 1)
				offsetX := float64(faceCenterX-frameCenterX) / float64(frameCenterX)
				offsetY := float64(faceCenterY-frameCenterY) / float64(frameCenterY)

				// Apply speed and smoothing
				targetPan := offsetX * speed
				targetTilt := -offsetY * speed // Invert for camera coords

				// Smooth movement
				pan := prevPan + (targetPan-prevPan)*smoothing
				tilt := prevTilt + (targetTilt-prevTilt)*smoothing
				prevPan = pan
				prevTilt = tilt

				// Apply dead zone
				if abs(pan) < 0.05 {
					pan = 0
				}
				if abs(tilt) < 0.05 {
					tilt = 0
				}

				// Calculate zoom if enabled
				zoom := 0.0
				if zoomToFace {
					// Target: face should be ~20% of frame height
					targetFaceRatio := 0.2
					actualFaceRatio := float64(primaryFace.Height) / float64(frameHeight)
					zoomOffset := (targetFaceRatio - actualFaceRatio) * 2
					if abs(zoomOffset) > 0.1 {
						zoom = zoomOffset * speed
					}
				}

				// Move camera
				if pan != 0 || tilt != 0 || zoom != 0 {
					_ = camClient.ContinuousMove(ctx, pan, tilt, zoom)
				} else {
					_ = camClient.Stop(ctx)
				}
			}
		}
	}()

	var sb strings.Builder
	sb.WriteString("# Face Tracking Started\n\n")
	sb.WriteString(fmt.Sprintf("**Camera:** %s\n", cameraID))
	sb.WriteString(fmt.Sprintf("**NDI Source:** %s\n", ndiSource))
	sb.WriteString(fmt.Sprintf("**Speed:** %.2f\n", speed))
	sb.WriteString(fmt.Sprintf("**Smoothing:** %.2f\n", smoothing))
	sb.WriteString(fmt.Sprintf("**Zoom to Face:** %v\n", zoomToFace))
	sb.WriteString("\nCamera will now follow detected faces.\n")
	sb.WriteString("Use `aftrs_ptz_track_stop` to stop tracking.\n")

	return tools.TextResult(sb.String()), nil
}

// handleTrackMotion starts motion tracking
func handleTrackMotion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ndiSource := tools.GetStringParam(req, "ndi_source")
	cameraID := tools.GetStringParam(req, "camera_id")
	speed := tools.GetFloatParam(req, "speed", 0.3)
	threshold := tools.GetFloatParam(req, "threshold", 5.0)

	if speed < 0.1 {
		speed = 0.1
	} else if speed > 1.0 {
		speed = 1.0
	}

	ptzClient, err := getPTZMultiClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Default camera if not specified
	if cameraID == "" {
		cameras := ptzClient.ListCameras()
		if len(cameras) == 1 {
			cameraID = cameras[0].ID
		} else if len(cameras) == 0 {
			return tools.ErrorResult(fmt.Errorf("no PTZ cameras configured")), nil
		} else {
			return tools.ErrorResult(fmt.Errorf("camera_id required when multiple cameras configured")), nil
		}
	}

	ndiClient, err := getNDIClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	camClient, err := ptzClient.GetCamera(cameraID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Stop existing tracking on this camera
	sessionsMu.Lock()
	if existing, exists := trackingSessions[cameraID]; exists {
		close(existing.stopCh)
		delete(trackingSessions, cameraID)
	}

	// Create new session
	session := &TrackingSession{
		CameraID:  cameraID,
		NDISource: ndiSource,
		Mode:      "motion",
		Speed:     speed,
		Active:    true,
		StartTime: time.Now(),
		stopCh:    make(chan struct{}),
	}
	trackingSessions[cameraID] = session
	sessionsMu.Unlock()

	// Start tracking goroutine
	go func() {
		var prevFramePath string
		frameInterval := time.Millisecond * 500 // 2 FPS for motion

		for {
			select {
			case <-session.stopCh:
				return
			case <-time.After(frameInterval):
				// Capture frame
				frame, err := ndiClient.CaptureFrame(ctx, ndiSource, "")
				if err != nil {
					continue
				}

				if prevFramePath == "" {
					prevFramePath = frame.FilePath
					continue
				}

				// Detect motion
				result, err := ndiClient.DetectMotion(ctx, prevFramePath, frame.FilePath, threshold)
				prevFramePath = frame.FilePath

				if err != nil || !result.MotionDetected {
					_ = camClient.Stop(ctx)
					continue
				}

				session.FrameCount++

				// If we have motion regions, pan toward the largest one
				if len(result.Regions) > 0 {
					var maxRegion clients.MotionRegion
					maxScore := 0.0
					for _, r := range result.Regions {
						if r.Score > maxScore {
							maxScore = r.Score
							maxRegion = r
						}
					}

					session.LastTarget = fmt.Sprintf("Motion at (%d,%d)", maxRegion.X, maxRegion.Y)

					// Calculate offset from center
					frameWidth := 1920 // Default
					frameHeight := 1080
					regionCenterX := maxRegion.X + maxRegion.Width/2
					regionCenterY := maxRegion.Y + maxRegion.Height/2
					offsetX := float64(regionCenterX-frameWidth/2) / float64(frameWidth/2)
					offsetY := float64(regionCenterY-frameHeight/2) / float64(frameHeight/2)

					pan := offsetX * speed
					tilt := -offsetY * speed

					if abs(pan) > 0.05 || abs(tilt) > 0.05 {
						_ = camClient.ContinuousMove(ctx, pan, tilt, 0)
					}
				}
			}
		}
	}()

	var sb strings.Builder
	sb.WriteString("# Motion Tracking Started\n\n")
	sb.WriteString(fmt.Sprintf("**Camera:** %s\n", cameraID))
	sb.WriteString(fmt.Sprintf("**NDI Source:** %s\n", ndiSource))
	sb.WriteString(fmt.Sprintf("**Speed:** %.2f\n", speed))
	sb.WriteString(fmt.Sprintf("**Threshold:** %.1f%%\n", threshold))
	sb.WriteString("\nCamera will now follow detected motion.\n")
	sb.WriteString("Use `aftrs_ptz_track_stop` to stop tracking.\n")

	return tools.TextResult(sb.String()), nil
}

// handleTrackStop stops tracking
func handleTrackStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cameraID := tools.GetStringParam(req, "camera_id")

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if cameraID == "" {
		// Stop all tracking
		for id, session := range trackingSessions {
			close(session.stopCh)
			delete(trackingSessions, id)
		}
		return tools.TextResult("# Tracking Stopped\n\nAll tracking sessions have been stopped."), nil
	}

	session, exists := trackingSessions[cameraID]
	if !exists {
		return tools.ErrorResult(fmt.Errorf("no active tracking on camera: %s", cameraID)), nil
	}

	close(session.stopCh)
	delete(trackingSessions, cameraID)

	// Stop camera movement
	ptzClient, err := getPTZMultiClient()
	if err == nil {
		if camClient, err := ptzClient.GetCamera(cameraID); err == nil {
			_ = camClient.Stop(ctx)
		}
	}

	return tools.TextResult(fmt.Sprintf("# Tracking Stopped\n\nTracking on camera '%s' has been stopped.", cameraID)), nil
}

// handleTrackStatus shows tracking status
func handleTrackStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cameraID := tools.GetStringParam(req, "camera_id")

	sessionsMu.RLock()
	defer sessionsMu.RUnlock()

	var sb strings.Builder
	sb.WriteString("# PTZ Tracking Status\n\n")

	if cameraID != "" {
		session, exists := trackingSessions[cameraID]
		if !exists {
			sb.WriteString(fmt.Sprintf("**Camera %s:** No active tracking\n", cameraID))
		} else {
			writeSessionStatus(&sb, session)
		}
	} else {
		if len(trackingSessions) == 0 {
			sb.WriteString("No active tracking sessions.\n")
		} else {
			for _, session := range trackingSessions {
				writeSessionStatus(&sb, session)
				sb.WriteString("\n")
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleTrackSnapshot captures and analyzes a frame
func handleTrackSnapshot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ndiSource := tools.GetStringParam(req, "ndi_source")
	detectFaces := tools.GetBoolParam(req, "detect_faces", true)
	detectMotion := tools.GetBoolParam(req, "detect_motion", false)

	ndiClient, err := getNDIClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Capture frame
	frame, err := ndiClient.CaptureFrame(ctx, ndiSource, "")
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to capture frame: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tracking Snapshot Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**NDI Source:** %s\n", ndiSource))
	sb.WriteString(fmt.Sprintf("**Frame Size:** %dx%d\n", frame.Width, frame.Height))
	sb.WriteString(fmt.Sprintf("**Captured:** %s\n\n", frame.Timestamp))

	// Detect faces
	if detectFaces {
		faceResult, err := ndiClient.DetectFaces(ctx, frame.FilePath)
		sb.WriteString("## Face Detection\n\n")
		if err != nil {
			sb.WriteString(fmt.Sprintf("Error: %v\n", err))
		} else if faceResult.FaceCount == 0 {
			sb.WriteString("No faces detected.\n")
		} else {
			sb.WriteString(fmt.Sprintf("**Faces Found:** %d\n\n", faceResult.FaceCount))
			sb.WriteString("| # | Position | Size | Area |\n")
			sb.WriteString("|---|----------|------|------|\n")
			for i, face := range faceResult.Faces {
				area := face.Width * face.Height
				sb.WriteString(fmt.Sprintf("| %d | (%d, %d) | %dx%d | %d px² |\n",
					i+1, face.X, face.Y, face.Width, face.Height, area))
			}
		}
		sb.WriteString("\n")
	}

	// Detect motion (would need previous frame)
	if detectMotion {
		sb.WriteString("## Motion Detection\n\n")
		sb.WriteString("*Motion detection requires two consecutive frames.*\n")
		sb.WriteString("*Start tracking mode for continuous motion detection.*\n")
	}

	return tools.TextResult(sb.String()), nil
}

// writeSessionStatus writes session status to string builder
func writeSessionStatus(sb *strings.Builder, session *TrackingSession) {
	sb.WriteString(fmt.Sprintf("## Camera: %s\n\n", session.CameraID))
	sb.WriteString(fmt.Sprintf("**Mode:** %s tracking\n", session.Mode))
	sb.WriteString(fmt.Sprintf("**NDI Source:** %s\n", session.NDISource))
	sb.WriteString(fmt.Sprintf("**Speed:** %.2f\n", session.Speed))
	sb.WriteString(fmt.Sprintf("**Active:** %v\n", session.Active))
	sb.WriteString(fmt.Sprintf("**Running Since:** %s\n", session.StartTime.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Frames Processed:** %d\n", session.FrameCount))
	if session.LastTarget != "" {
		sb.WriteString(fmt.Sprintf("**Last Target:** %s\n", session.LastTarget))
	}
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
