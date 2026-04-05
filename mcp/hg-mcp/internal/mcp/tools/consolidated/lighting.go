package consolidated

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&lightingStatusModule{})
}

type lightingStatusModule struct{}

func (m *lightingStatusModule) Name() string        { return "lighting_status" }
func (m *lightingStatusModule) Description() string { return "Consolidated lighting status views" }

func (m *lightingStatusModule) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_lighting_status",
				mcp.WithDescription("Aggregated lighting status: WLED + sACN + LedFX + QLC+ + grandMA3 + Nanoleaf + Hue. Health scores, per-system status, and issues."),
			),
			Handler:             handleLightingStatus,
			Category:            "consolidated",
			Subcategory:         "lighting",
			Tags:                []string{"lighting", "status", "consolidated", "wled", "nanoleaf", "hue", "grandma3"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "consolidated",
		},
		{
			Tool: mcp.NewTool("aftrs_dj_status",
				mcp.WithDescription("Aggregated DJ/music status: Rekordbox library + Serato + track counts + sync status. Overview of DJ setup."),
			),
			Handler:             handleDJStatus,
			Category:            "consolidated",
			Subcategory:         "dj",
			Tags:                []string{"dj", "rekordbox", "serato", "status", "library"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "consolidated",
		},
		{
			Tool: mcp.NewTool("aftrs_audio_status",
				mcp.WithDescription("Aggregated audio production status: Ableton + Dante + MIDI devices. Transport state, BPM, connected devices."),
			),
			Handler:             handleAudioStatus,
			Category:            "consolidated",
			Subcategory:         "audio",
			Tags:                []string{"audio", "ableton", "dante", "midi", "status"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "consolidated",
		},
	}
}

func handleLightingStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var components []HealthComponent
	totalScore := 0
	componentCount := 0
	var allIssues []string

	// Check WLED
	if wledClient, err := clients.NewWLEDClient(); err == nil {
		comp := HealthComponent{Name: "WLED", Score: 0, Status: "unknown"}
		devices := wledClient.ListDevices(ctx) // returns []*WLEDDevice (single return)
		comp.Score = 100
		comp.Status = "online"
		comp.Details = fmt.Sprintf("%d device(s)", len(devices))
		if len(devices) == 0 {
			comp.Score = 50
			comp.Issues = append(comp.Issues, "No WLED devices found")
		}
		components = append(components, comp)
		totalScore += comp.Score
		componentCount++
	}

	// Check grandMA3
	if gmaClient, err := clients.NewGrandMA3Client(); err == nil {
		comp := HealthComponent{Name: "grandMA3", Score: 0, Status: "unknown"}
		health, err := gmaClient.GetHealth(ctx)
		if err == nil && health != nil {
			comp.Score = health.Score
			comp.Status = health.Status
			if len(health.Issues) > 0 {
				comp.Issues = health.Issues
			}
		} else {
			comp.Status = "unavailable"
			comp.Issues = append(comp.Issues, "grandMA3 not responding")
		}
		components = append(components, comp)
		totalScore += comp.Score
		componentCount++
	}

	// Check Nanoleaf
	if nlClient, err := clients.GetNanoleafClient(); err == nil {
		comp := HealthComponent{Name: "Nanoleaf", Score: 0, Status: "unknown"}
		health, err := nlClient.GetHealth(ctx)
		if err == nil && health != nil {
			comp.Score = health.Score
			comp.Status = health.Status
			comp.Issues = health.Recommendations
		} else {
			comp.Status = "unavailable"
			comp.Issues = append(comp.Issues, "Nanoleaf not responding")
		}
		components = append(components, comp)
		totalScore += comp.Score
		componentCount++
	}

	// Check Hue
	if hueClient, err := clients.GetHueClient(); err == nil {
		comp := HealthComponent{Name: "Philips Hue", Score: 0, Status: "unknown"}
		health, err := hueClient.GetHealth(ctx)
		if err == nil && health != nil {
			comp.Score = health.Score
			comp.Status = health.Status
			comp.Issues = health.Recommendations
		} else {
			comp.Status = "unavailable"
			comp.Issues = append(comp.Issues, "Hue bridge not responding")
		}
		components = append(components, comp)
		totalScore += comp.Score
		componentCount++
	}

	// Build summary
	overallScore := 0
	if componentCount > 0 {
		overallScore = totalScore / componentCount
	}

	for _, comp := range components {
		for _, issue := range comp.Issues {
			allIssues = append(allIssues, fmt.Sprintf("[%s] %s", comp.Name, issue))
		}
	}

	overallStatus := "healthy"
	if overallScore < 50 {
		overallStatus = "degraded"
	} else if overallScore < 80 {
		overallStatus = "warning"
	}
	if componentCount == 0 {
		overallStatus = "no_systems_configured"
		overallScore = 0
	}

	var table []string
	table = append(table, "System          | Status      | Score | Issues")
	table = append(table, "----------------|-------------|-------|-------")
	for _, comp := range components {
		issues := "-"
		if len(comp.Issues) > 0 {
			issues = strings.Join(comp.Issues, "; ")
		}
		table = append(table, fmt.Sprintf("%-15s | %-11s | %3d   | %s",
			comp.Name, comp.Status, comp.Score, issues))
	}

	return tools.JSONResult(map[string]interface{}{
		"overall_health": overallScore,
		"overall_status": overallStatus,
		"systems":        componentCount,
		"components":     components,
		"issues":         allIssues,
		"status_table":   strings.Join(table, "\n"),
	}), nil
}

func handleDJStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result := map[string]interface{}{
		"systems": map[string]interface{}{},
	}
	systems := result["systems"].(map[string]interface{})
	totalScore := 0
	systemCount := 0

	// Check Rekordbox
	if rbClient, err := clients.GetRekordboxClient(); err == nil {
		rbInfo := map[string]interface{}{"status": "unknown"}
		status, err := rbClient.GetStatus(ctx)
		if err == nil && status != nil {
			rbInfo["status"] = "connected"
			rbInfo["details"] = status
			totalScore += 100
		} else {
			rbInfo["status"] = "unavailable"
		}
		systems["rekordbox"] = rbInfo
		systemCount++
	}

	// Check Serato
	if seratoClient, err := clients.NewSeratoClient(); err == nil {
		seratoInfo := map[string]interface{}{"status": "unknown"}
		status, err := seratoClient.GetStatus(ctx)
		if err == nil && status != nil {
			seratoInfo["status"] = "connected"
			seratoInfo["details"] = status
			totalScore += 100
		} else {
			seratoInfo["status"] = "unavailable"
		}
		systems["serato"] = seratoInfo
		systemCount++
	}

	overallScore := 0
	if systemCount > 0 {
		overallScore = totalScore / systemCount
	}

	overallStatus := "healthy"
	if overallScore < 50 {
		overallStatus = "degraded"
	}
	if systemCount == 0 {
		overallStatus = "no_dj_systems_configured"
	}

	result["overall_health"] = overallScore
	result["overall_status"] = overallStatus
	result["system_count"] = systemCount

	return tools.JSONResult(result), nil
}

func handleAudioStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result := map[string]interface{}{
		"systems": map[string]interface{}{},
	}
	systems := result["systems"].(map[string]interface{})
	totalScore := 0
	systemCount := 0

	// Check Ableton
	if ablClient, err := clients.NewAbletonClient(); err == nil {
		ablInfo := map[string]interface{}{"status": "unknown"}
		status, err := ablClient.GetStatus(ctx)
		if err == nil && status != nil {
			ablInfo["status"] = "connected"
			ablInfo["details"] = status
			totalScore += 100
		} else {
			ablInfo["status"] = "unavailable"
		}
		systems["ableton"] = ablInfo
		systemCount++
	}

	// Check Dante
	if danteClient, err := clients.NewDanteClient(); err == nil {
		danteInfo := map[string]interface{}{"status": "unknown"}
		status, err := danteClient.GetStatus(ctx)
		if err == nil && status != nil {
			danteInfo["status"] = "connected"
			danteInfo["details"] = status
			totalScore += 100
		} else {
			danteInfo["status"] = "unavailable"
		}
		systems["dante"] = danteInfo
		systemCount++
	}

	// Check MIDI
	if midiClient, err := clients.NewMIDIClient(); err == nil {
		midiInfo := map[string]interface{}{"status": "unknown"}
		devices, err := midiClient.GetDevices(ctx)
		if err == nil {
			midiInfo["status"] = "connected"
			midiInfo["devices"] = devices
			totalScore += 100
		} else {
			midiInfo["status"] = "unavailable"
		}
		systems["midi"] = midiInfo
		systemCount++
	}

	overallScore := 0
	if systemCount > 0 {
		overallScore = totalScore / systemCount
	}

	overallStatus := "healthy"
	if overallScore < 50 {
		overallStatus = "degraded"
	}
	if systemCount == 0 {
		overallStatus = "no_audio_systems_configured"
	}

	result["overall_health"] = overallScore
	result["overall_status"] = overallStatus
	result["system_count"] = systemCount

	return tools.JSONResult(result), nil
}
