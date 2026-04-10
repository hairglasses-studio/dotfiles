// Package config provides startup configuration auditing.
package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// ServiceGroup represents a group of related environment variables for a service.
type ServiceGroup struct {
	Name              string
	EnvVars           []string     // at least one must be set to consider the group configured
	Optional          bool         // if true, missing config is expected (not warned about)
	ConnectivityCheck func() error // optional: check if the service is reachable (2s timeout)
}

// ConfigReport contains the results of a configuration audit.
type ConfigReport struct {
	Configured []string
	Missing    []string
}

// knownServices defines the env var groups checked at startup.
// Covers all service integrations in hg-mcp (expanded from 26 to 53).
var knownServices = []ServiceGroup{
	// ── Core ──
	{Name: "Core", EnvVars: []string{"MCP_MODE"}, Optional: true},
	{Name: "Security", EnvVars: []string{"RBAC_USERS"}, Optional: true},
	{Name: "Observability", EnvVars: []string{"DISABLE_TRACING"}, Optional: true},

	// ── Communication ──
	{Name: "Discord", EnvVars: []string{"DISCORD_BOT_TOKEN"}},
	{Name: "Twitch", EnvVars: []string{"TWITCH_CLIENT_ID"}},

	// ── Google Services ──
	{Name: "Google", EnvVars: []string{"GOOGLE_APPLICATION_CREDENTIALS", "GOOGLE_API_KEY", "GOOGLE_CLIENT_ID"}},
	{Name: "YouTube", EnvVars: []string{"YOUTUBE_API_KEY"}},
	{Name: "Inventory", EnvVars: []string{"INVENTORY_SPREADSHEET_ID"}},

	// ── Music Platforms ──
	{Name: "Spotify", EnvVars: []string{"SPOTIFY_CLIENT_ID"}},
	{Name: "SoundCloud", EnvVars: []string{"SOUNDCLOUD_CLIENT_ID"}},
	{Name: "Beatport", EnvVars: []string{"BEATPORT_USERNAME", "BEATPORT_ACCESS_TOKEN"}},
	{Name: "Tidal", EnvVars: []string{"TIDAL_CLIENT_ID"}},
	{Name: "Discogs", EnvVars: []string{"DISCOGS_TOKEN"}},

	// ── DJ Software ──
	{Name: "Rekordbox", EnvVars: []string{"REKORDBOX_DB_PATH"}},
	{Name: "Traktor", EnvVars: []string{"TRAKTOR_COLLECTION_PATH"}},
	{Name: "Serato", EnvVars: []string{"SERATO_LIBRARY_PATH"}},

	// ── Productivity ──
	{Name: "Notion", EnvVars: []string{"NOTION_API_KEY", "NOTION_TOKEN"}},

	// ── Home Automation ──
	{Name: "HomeAssistant", EnvVars: []string{"HASS_URL"}},
	{Name: "MQTT", EnvVars: []string{"MQTT_BROKER"}},
	{Name: "Companion", EnvVars: []string{"COMPANION_HOST"}},
	{Name: "Chataigne", EnvVars: []string{"CHATAIGNE_HOST"}},

	// ── Creative / VJ ──
	{Name: "Resolume", EnvVars: []string{"RESOLUME_OSC_HOST", "RESOLUME_API_PORT"}},
	{Name: "TouchDesigner", EnvVars: []string{"TD_HOST"}},
	{Name: "VJ Clips", EnvVars: []string{"VJ_CLIPS_BUCKET", "VJ_CLIPS_PATH"}},
	{Name: "FFGL", EnvVars: []string{"FFGL_TEMPLATES_DIR"}},

	// ── Music Production ──
	{Name: "Ableton", EnvVars: []string{"ABLETON_OSC_HOST"}},
	{Name: "Max4Live", EnvVars: []string{"M4L_OSC_HOST"}},
	{Name: "Stems/Demucs", EnvVars: []string{"DEMUCS_PATH"}},

	// ── Lighting ──
	{Name: "grandMA3", EnvVars: []string{"GRANDMA3_HOST"}},
	{Name: "ArtNet", EnvVars: []string{"ARTNET_HOST"}},
	{Name: "LedFX", EnvVars: []string{"LEDFX_HOST"}},
	{Name: "QLC+", EnvVars: []string{"QLCPLUS_HOST"}},
	{Name: "sACN/E1.31", EnvVars: []string{"SACN_BIND_ADDR"}},
	{Name: "Nanoleaf", EnvVars: []string{"NANOLEAF_HOST"}},
	{Name: "PhilipsHue", EnvVars: []string{"HUE_BRIDGE_IP"}},

	// ── Video / Streaming ──
	{Name: "OBS", EnvVars: []string{"OBS_HOST"}},
	{Name: "ATEM", EnvVars: []string{"ATEM_HOST"}},
	{Name: "PTZ", EnvVars: []string{"PTZ_HOST"}},
	{Name: "ShowKontrol", EnvVars: []string{"SHOWKONTROL_HOST"}},
	{Name: "VideoAI", EnvVars: []string{"VIDTOOL_PATH"}},

	// ── Audio ──
	{Name: "MIDI", EnvVars: []string{"MIDI_OUTPUT", "MIDI_INPUT"}},
	{Name: "Dante", EnvVars: []string{"DANTE_INTERFACE"}},
	{Name: "Fingerprint", EnvVars: []string{"ACOUSTID_API_KEY", "FPCALC_PATH"}},
	{Name: "Whisper", EnvVars: []string{"WHISPER_PATH", "OPENAI_API_KEY"}},

	// ── AI ──
	{Name: "Ollama", EnvVars: []string{"OLLAMA_BASE_URL", "OLLAMA_HOST"}},
	{Name: "Anthropic", EnvVars: []string{"ANTHROPIC_API_KEY"}},

	// ── Infrastructure ──
	{Name: "Unraid", EnvVars: []string{"UNRAID_HOST"}},
	{Name: "OPNsense", EnvVars: []string{"OPNSENSE_HOST"}},
	{Name: "AWS/CR8", EnvVars: []string{"AWS_PROFILE", "CR8_S3_BUCKET"}},
	{Name: "Tailscale", EnvVars: []string{}, Optional: true},

	// ── Storage & Sync ──
	{Name: "Backup", EnvVars: []string{"BACKUP_CONFIG_PATH", "BACKUP_ROOT"}},
	{Name: "Rclone", EnvVars: []string{"RCLONE_CONFIG"}},

	// ── Marketplace ──
	{Name: "eBay", EnvVars: []string{"EBAY_APP_ID"}},

	// ── Gaming ──
	{Name: "Retrogaming", EnvVars: []string{"PCSX2_PATH", "RETROARCH_PATH"}},

	// ── Secrets & Plugins ──
	{Name: "1Password", EnvVars: []string{"OP_SERVICE_ACCOUNT_TOKEN", "OP_ACCOUNT"}},
	{Name: "Plugins", EnvVars: []string{"GITHUB_TOKEN"}},
}

// CategoryToService maps tool categories to config service names.
// Used to determine if a tool's backing service is configured.
var CategoryToService = map[string]string{
	"discord": "Discord", "discord_admin": "Discord",
	"spotify": "Spotify", "soundcloud": "SoundCloud", "beatport": "Beatport",
	"tidal": "Tidal", "discogs": "Discogs",
	"rekordbox": "Rekordbox", "traktor": "Traktor", "serato": "Serato",
	"notion":        "Notion",
	"homeassistant": "HomeAssistant", "mqtt": "MQTT",
	"companion": "Companion", "chataigne": "Chataigne",
	"resolume": "Resolume", "touchdesigner": "TouchDesigner",
	"vj_clips": "VJ Clips", "ffgl": "FFGL",
	"ableton": "Ableton", "maxforlive": "Max4Live", "stems": "Stems/Demucs",
	"grandma3": "grandMA3", "ledfx": "LedFX", "qlcplus": "QLC+",
	"sacn": "sACN/E1.31", "nanoleaf": "Nanoleaf", "hue": "PhilipsHue",
	"obs": "OBS", "atem": "ATEM", "ptz": "PTZ",
	"showkontrol": "ShowKontrol", "videoai": "VideoAI",
	"midi": "MIDI", "dante": "Dante", "fingerprint": "Fingerprint",
	"whisper": "Whisper", "ollama": "Ollama",
	"unraid": "Unraid", "opnsense": "OPNsense",
	"rclone": "Rclone", "backup": "Backup",
	"inventory":   "Inventory",
	"twitch":      "Twitch",
	"retrogaming": "Retrogaming",
}

// IsServiceConfigured returns true if the named service has its env vars set.
func IsServiceConfigured(serviceName string) bool {
	for _, svc := range knownServices {
		if svc.Name == serviceName {
			return hasAnyEnv(svc.EnvVars)
		}
	}
	return true // unknown services assumed configured
}

// IsCategoryConfigured returns true if the service backing a tool category is configured.
func IsCategoryConfigured(category string) bool {
	svcName, ok := CategoryToService[category]
	if !ok {
		return true // categories without a known service are assumed configured
	}
	return IsServiceConfigured(svcName)
}

// AuditConfig scans known env var groups and returns which services are configured.
func AuditConfig() ConfigReport {
	var report ConfigReport

	for _, svc := range knownServices {
		if svc.Optional || len(svc.EnvVars) == 0 {
			continue
		}
		if hasAnyEnv(svc.EnvVars) {
			report.Configured = append(report.Configured, svc.Name)
		} else {
			report.Missing = append(report.Missing, svc.Name)
		}
	}

	return report
}

// Log prints the config report to the standard logger.
func (r ConfigReport) Log() {
	var parts []string

	for _, name := range r.Configured {
		parts = append(parts, fmt.Sprintf("✓ %s", name))
	}
	for _, name := range r.Missing {
		parts = append(parts, fmt.Sprintf("✗ %s", name))
	}

	log.Printf("Config: %s", strings.Join(parts, "  "))
	log.Printf("  %d services configured, %d unconfigured", len(r.Configured), len(r.Missing))
}

// ConnectivityResult represents a single connectivity check result.
type ConnectivityResult struct {
	Name      string        `json:"name"`
	Reachable bool          `json:"reachable"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration_ms"`
}

// ConnectivityReport contains the results of connectivity checks.
type ConnectivityReport struct {
	Results []ConnectivityResult
	Skipped bool
}

// AuditConnectivity runs connectivity checks for all configured services
// that have a ConnectivityCheck function. Each check has a 2-second timeout.
// Set SKIP_CONNECTIVITY_CHECKS=1 to skip all checks.
func AuditConnectivity() ConnectivityReport {
	report := ConnectivityReport{}

	if os.Getenv("SKIP_CONNECTIVITY_CHECKS") == "1" {
		report.Skipped = true
		return report
	}

	for _, svc := range knownServices {
		if svc.ConnectivityCheck == nil {
			continue
		}
		// Only check configured services
		if !svc.Optional && len(svc.EnvVars) > 0 && !hasAnyEnv(svc.EnvVars) {
			continue
		}

		result := ConnectivityResult{Name: svc.Name}
		start := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		done := make(chan error, 1)
		go func() {
			done <- svc.ConnectivityCheck()
		}()

		select {
		case err := <-done:
			result.Duration = time.Since(start)
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Reachable = true
			}
		case <-ctx.Done():
			result.Duration = time.Since(start)
			result.Error = "timeout (2s)"
		}
		cancel()

		report.Results = append(report.Results, result)
	}

	return report
}

// Log prints the connectivity report to the standard logger.
func (r ConnectivityReport) Log() {
	if r.Skipped {
		log.Printf("Connectivity: skipped (SKIP_CONNECTIVITY_CHECKS=1)")
		return
	}
	if len(r.Results) == 0 {
		return
	}

	reachable := 0
	for _, res := range r.Results {
		if res.Reachable {
			reachable++
		}
	}

	var parts []string
	for _, res := range r.Results {
		if res.Reachable {
			parts = append(parts, fmt.Sprintf("OK %s", res.Name))
		} else {
			parts = append(parts, fmt.Sprintf("FAIL %s (%s)", res.Name, res.Error))
		}
	}
	log.Printf("Connectivity: %s", strings.Join(parts, "  "))
	log.Printf("  %d/%d services reachable", reachable, len(r.Results))
}

func hasAnyEnv(vars []string) bool {
	for _, v := range vars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}
