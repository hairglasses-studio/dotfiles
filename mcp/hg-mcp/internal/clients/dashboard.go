// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/httpclient"
)

// DashboardClient aggregates health status from all connected systems.
type DashboardClient struct {
	mu sync.RWMutex

	// System registry
	systems map[string]*SystemConfig

	// Alert state
	alerts    []Alert
	maxAlerts int

	// Monitoring state
	monitoring    bool
	monitorCtx    context.Context
	monitorCancel context.CancelFunc
	pollInterval  time.Duration

	// HTTP client for health checks
	httpClient *http.Client
}

// SystemConfig defines a system to monitor.
type SystemConfig struct {
	Name        string `json:"name"`
	Category    string `json:"category"`    // audio, video, lighting, network, etc.
	Host        string `json:"host"`        // hostname or IP
	Port        int    `json:"port"`        // main port
	Protocol    string `json:"protocol"`    // http, https, osc, etc.
	HealthPath  string `json:"health_path"` // e.g., /health, /api/status
	Description string `json:"description"`
	Critical    bool   `json:"critical"` // is this system critical for show?
}

// SystemStatus represents the current status of a system.
type SystemStatus struct {
	Name        string        `json:"name"`
	Category    string        `json:"category"`
	Status      string        `json:"status"` // online, offline, degraded, unknown
	Latency     time.Duration `json:"latency"`
	LastCheck   time.Time     `json:"last_check"`
	Message     string        `json:"message,omitempty"`
	Critical    bool          `json:"critical"`
	HealthScore int           `json:"health_score"` // 0-100
}

// Alert represents a system alert.
type Alert struct {
	ID        string    `json:"id"`
	System    string    `json:"system"`
	Severity  string    `json:"severity"` // info, warning, error, critical
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// DashboardSummary provides an overview of all systems.
type DashboardSummary struct {
	TotalSystems   int            `json:"total_systems"`
	OnlineCount    int            `json:"online_count"`
	OfflineCount   int            `json:"offline_count"`
	DegradedCount  int            `json:"degraded_count"`
	UnknownCount   int            `json:"unknown_count"`
	OverallHealth  int            `json:"overall_health"` // 0-100
	CriticalIssues int            `json:"critical_issues"`
	ActiveAlerts   int            `json:"active_alerts"`
	LastUpdated    time.Time      `json:"last_updated"`
	Systems        []SystemStatus `json:"systems"`
}

var (
	dashboardClientInstance *DashboardClient
	dashboardOnce           sync.Once
)

// GetDashboardClient returns the singleton dashboard client.
func GetDashboardClient() *DashboardClient {
	dashboardOnce.Do(func() {
		dashboardClientInstance = &DashboardClient{
			systems:      initDefaultSystems(),
			maxAlerts:    100,
			pollInterval: 30 * time.Second,
			httpClient:   httpclient.Fast(),
		}
	})
	return dashboardClientInstance
}

// initDefaultSystems configures default systems to monitor.
func initDefaultSystems() map[string]*SystemConfig {
	return map[string]*SystemConfig{
		"resolume": {
			Name:        "resolume",
			Category:    "video",
			Host:        "localhost",
			Port:        8080,
			Protocol:    "http",
			HealthPath:  "/api/v1/composition",
			Description: "Resolume Arena VJ software",
			Critical:    true,
		},
		"touchdesigner": {
			Name:        "touchdesigner",
			Category:    "video",
			Host:        "localhost",
			Port:        9980,
			Protocol:    "http",
			HealthPath:  "/",
			Description: "TouchDesigner real-time graphics",
			Critical:    false,
		},
		"obs": {
			Name:        "obs",
			Category:    "video",
			Host:        "localhost",
			Port:        4455,
			Protocol:    "ws",
			Description: "OBS Studio streaming/recording",
			Critical:    false,
		},
		"grandma3": {
			Name:        "grandma3",
			Category:    "lighting",
			Host:        "localhost",
			Port:        9000,
			Protocol:    "http",
			Description: "grandMA3 lighting console",
			Critical:    true,
		},
		"atem": {
			Name:        "atem",
			Category:    "video",
			Host:        "localhost",
			Port:        9910,
			Protocol:    "tcp",
			Description: "Blackmagic ATEM switcher",
			Critical:    true,
		},
		"ableton": {
			Name:        "ableton",
			Category:    "audio",
			Host:        "localhost",
			Port:        11000,
			Protocol:    "osc",
			Description: "Ableton Live (via LiveOSC)",
			Critical:    true,
		},
		"dante": {
			Name:        "dante",
			Category:    "audio",
			Host:        "localhost",
			Port:        8800,
			Protocol:    "http",
			Description: "Dante audio network",
			Critical:    true,
		},
		"mqtt": {
			Name:        "mqtt",
			Category:    "network",
			Host:        "localhost",
			Port:        1883,
			Protocol:    "tcp",
			Description: "MQTT message broker",
			Critical:    false,
		},
		"homeassistant": {
			Name:        "homeassistant",
			Category:    "automation",
			Host:        "localhost",
			Port:        8123,
			Protocol:    "http",
			HealthPath:  "/api/",
			Description: "Home Assistant automation",
			Critical:    false,
		},
		"wled": {
			Name:        "wled",
			Category:    "lighting",
			Host:        "localhost",
			Port:        80,
			Protocol:    "http",
			HealthPath:  "/json/state",
			Description: "WLED LED controller",
			Critical:    false,
		},
		"nanoleaf": {
			Name:        "nanoleaf",
			Category:    "lighting",
			Host:        "localhost",
			Port:        16021,
			Protocol:    "http",
			HealthPath:  "/api/v1",
			Description: "Nanoleaf light panels",
			Critical:    false,
		},
		"hue": {
			Name:        "hue",
			Category:    "lighting",
			Host:        "localhost",
			Port:        80,
			Protocol:    "http",
			HealthPath:  "/api/config",
			Description: "Philips Hue bridge",
			Critical:    false,
		},
		"ollama": {
			Name:        "ollama",
			Category:    "ai",
			Host:        "localhost",
			Port:        11434,
			Protocol:    "http",
			HealthPath:  "/api/tags",
			Description: "Ollama local LLM",
			Critical:    false,
		},
	}
}

// GetQuickStatus returns a one-line status of all systems.
func (c *DashboardClient) GetQuickStatus(ctx context.Context) string {
	c.mu.RLock()
	systems := make([]*SystemConfig, 0, len(c.systems))
	for _, sys := range c.systems {
		systems = append(systems, sys)
	}
	c.mu.RUnlock()

	// Sort by category and name
	sort.Slice(systems, func(i, j int) bool {
		if systems[i].Category != systems[j].Category {
			return systems[i].Category < systems[j].Category
		}
		return systems[i].Name < systems[j].Name
	})

	// Check all systems concurrently
	var wg sync.WaitGroup
	results := make(map[string]bool)
	var resultsMu sync.Mutex

	for _, sys := range systems {
		wg.Add(1)
		go func(s *SystemConfig) {
			defer wg.Done()
			online := c.checkSystem(ctx, s)
			resultsMu.Lock()
			results[s.Name] = online
			resultsMu.Unlock()
		}(sys)
	}
	wg.Wait()

	// Build summary
	online := 0
	offline := 0
	var icons []string

	for _, sys := range systems {
		if results[sys.Name] {
			online++
			icons = append(icons, fmt.Sprintf("✓%s", sys.Name))
		} else {
			offline++
			if sys.Critical {
				icons = append(icons, fmt.Sprintf("✗%s!", sys.Name))
			} else {
				icons = append(icons, fmt.Sprintf("✗%s", sys.Name))
			}
		}
	}

	return fmt.Sprintf("%d/%d online | %s", online, len(systems), formatIcons(icons))
}

// formatIcons limits and formats the icon string
func formatIcons(icons []string) string {
	if len(icons) > 8 {
		return fmt.Sprintf("%s... (+%d)", join(icons[:8], " "), len(icons)-8)
	}
	return join(icons, " ")
}

func join(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for _, item := range items[1:] {
		result += sep + item
	}
	return result
}

// GetFullStatus returns detailed status of all systems.
func (c *DashboardClient) GetFullStatus(ctx context.Context) *DashboardSummary {
	c.mu.RLock()
	systems := make([]*SystemConfig, 0, len(c.systems))
	for _, sys := range c.systems {
		systems = append(systems, sys)
	}
	alertCount := 0
	for _, a := range c.alerts {
		if !a.Resolved {
			alertCount++
		}
	}
	c.mu.RUnlock()

	// Sort by category and name
	sort.Slice(systems, func(i, j int) bool {
		if systems[i].Category != systems[j].Category {
			return systems[i].Category < systems[j].Category
		}
		return systems[i].Name < systems[j].Name
	})

	// Check all systems concurrently
	var wg sync.WaitGroup
	statuses := make([]SystemStatus, len(systems))

	for i, sys := range systems {
		wg.Add(1)
		go func(idx int, s *SystemConfig) {
			defer wg.Done()
			statuses[idx] = c.checkSystemDetailed(ctx, s)
		}(i, sys)
	}
	wg.Wait()

	// Build summary
	summary := &DashboardSummary{
		TotalSystems: len(systems),
		LastUpdated:  time.Now(),
		Systems:      statuses,
		ActiveAlerts: alertCount,
	}

	totalHealth := 0
	for _, s := range statuses {
		switch s.Status {
		case "online":
			summary.OnlineCount++
		case "offline":
			summary.OfflineCount++
			if s.Critical {
				summary.CriticalIssues++
			}
		case "degraded":
			summary.DegradedCount++
		default:
			summary.UnknownCount++
		}
		totalHealth += s.HealthScore
	}

	if len(statuses) > 0 {
		summary.OverallHealth = totalHealth / len(statuses)
	}

	return summary
}

// checkSystem performs a quick connectivity check.
func (c *DashboardClient) checkSystem(ctx context.Context, sys *SystemConfig) bool {
	addr := net.JoinHostPort(sys.Host, strconv.Itoa(sys.Port))

	switch sys.Protocol {
	case "http", "https":
		url := fmt.Sprintf("%s://%s", sys.Protocol, addr)
		if sys.HealthPath != "" {
			url += sys.HealthPath
		}
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return false
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode >= 200 && resp.StatusCode < 500

	case "tcp", "ws", "osc":
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			return false
		}
		conn.Close()
		return true

	default:
		// For unknown protocols, try TCP
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	}
}

// checkSystemDetailed performs a detailed connectivity check with latency.
func (c *DashboardClient) checkSystemDetailed(ctx context.Context, sys *SystemConfig) SystemStatus {
	status := SystemStatus{
		Name:      sys.Name,
		Category:  sys.Category,
		LastCheck: time.Now(),
		Critical:  sys.Critical,
	}

	addr := net.JoinHostPort(sys.Host, strconv.Itoa(sys.Port))
	start := time.Now()

	switch sys.Protocol {
	case "http", "https":
		url := fmt.Sprintf("%s://%s", sys.Protocol, addr)
		if sys.HealthPath != "" {
			url += sys.HealthPath
		}
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			status.Status = "offline"
			status.Message = err.Error()
			status.HealthScore = 0
			return status
		}
		resp, err := c.httpClient.Do(req)
		status.Latency = time.Since(start)
		if err != nil {
			status.Status = "offline"
			status.Message = err.Error()
			status.HealthScore = 0
			return status
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			status.Status = "online"
			status.HealthScore = calculateHealthScore(status.Latency)
		} else if resp.StatusCode < 500 {
			status.Status = "degraded"
			status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
			status.HealthScore = 50
		} else {
			status.Status = "offline"
			status.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
			status.HealthScore = 0
		}

	case "tcp", "ws", "osc":
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		status.Latency = time.Since(start)
		if err != nil {
			status.Status = "offline"
			status.Message = err.Error()
			status.HealthScore = 0
			return status
		}
		conn.Close()
		status.Status = "online"
		status.HealthScore = calculateHealthScore(status.Latency)

	default:
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		status.Latency = time.Since(start)
		if err != nil {
			status.Status = "offline"
			status.Message = err.Error()
			status.HealthScore = 0
			return status
		}
		conn.Close()
		status.Status = "online"
		status.HealthScore = calculateHealthScore(status.Latency)
	}

	return status
}

// calculateHealthScore converts latency to a health score.
func calculateHealthScore(latency time.Duration) int {
	ms := latency.Milliseconds()
	if ms < 10 {
		return 100
	}
	if ms < 50 {
		return 95
	}
	if ms < 100 {
		return 90
	}
	if ms < 200 {
		return 80
	}
	if ms < 500 {
		return 70
	}
	if ms < 1000 {
		return 60
	}
	return 50
}

// StartMonitoring begins background health monitoring.
func (c *DashboardClient) StartMonitoring(interval time.Duration) {
	c.mu.Lock()
	if c.monitoring {
		c.mu.Unlock()
		return
	}
	c.monitoring = true
	if interval > 0 {
		c.pollInterval = interval
	}
	c.monitorCtx, c.monitorCancel = context.WithCancel(context.Background())
	c.mu.Unlock()

	go c.monitorLoop()
}

// StopMonitoring stops background health monitoring.
func (c *DashboardClient) StopMonitoring() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.monitoring {
		return
	}
	c.monitoring = false
	if c.monitorCancel != nil {
		c.monitorCancel()
	}
}

// IsMonitoring returns whether background monitoring is active.
func (c *DashboardClient) IsMonitoring() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.monitoring
}

// monitorLoop runs the background monitoring loop.
func (c *DashboardClient) monitorLoop() {
	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.monitorCtx.Done():
			return
		case <-ticker.C:
			c.runHealthCheck()
		}
	}
}

// runHealthCheck performs a single health check cycle.
func (c *DashboardClient) runHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	summary := c.GetFullStatus(ctx)

	// Generate alerts for offline systems
	c.mu.Lock()
	for _, sys := range summary.Systems {
		if sys.Status == "offline" && sys.Critical {
			c.addAlertLocked(Alert{
				ID:        fmt.Sprintf("%s-%d", sys.Name, time.Now().Unix()),
				System:    sys.Name,
				Severity:  "critical",
				Message:   fmt.Sprintf("%s is offline: %s", sys.Name, sys.Message),
				Timestamp: time.Now(),
			})
		} else if sys.Status == "degraded" {
			c.addAlertLocked(Alert{
				ID:        fmt.Sprintf("%s-%d", sys.Name, time.Now().Unix()),
				System:    sys.Name,
				Severity:  "warning",
				Message:   fmt.Sprintf("%s is degraded: %s", sys.Name, sys.Message),
				Timestamp: time.Now(),
			})
		}
	}
	c.mu.Unlock()
}

// addAlertLocked adds an alert (must hold lock).
func (c *DashboardClient) addAlertLocked(alert Alert) {
	c.alerts = append([]Alert{alert}, c.alerts...)
	if len(c.alerts) > c.maxAlerts {
		c.alerts = c.alerts[:c.maxAlerts]
	}
}

// GetAlerts returns active alerts.
func (c *DashboardClient) GetAlerts(includeResolved bool) []Alert {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []Alert
	for _, a := range c.alerts {
		if includeResolved || !a.Resolved {
			result = append(result, a)
		}
	}
	return result
}

// ResolveAlert marks an alert as resolved.
func (c *DashboardClient) ResolveAlert(alertID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.alerts {
		if c.alerts[i].ID == alertID {
			c.alerts[i].Resolved = true
			return true
		}
	}
	return false
}

// AddSystem adds a system to monitor.
func (c *DashboardClient) AddSystem(sys *SystemConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.systems[sys.Name] = sys
}

// RemoveSystem removes a system from monitoring.
func (c *DashboardClient) RemoveSystem(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.systems, name)
}

// GetSystems returns all configured systems.
func (c *DashboardClient) GetSystems() []*SystemConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*SystemConfig, 0, len(c.systems))
	for _, sys := range c.systems {
		result = append(result, sys)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
