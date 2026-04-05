// Package clients provides API clients for external services.
package clients

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// DiscoveryClient manages tool discovery state including recent tools, favorites, and aliases.
type DiscoveryClient struct {
	mu sync.RWMutex

	// Recent tool usage tracking
	recentTools []RecentToolEntry
	maxRecent   int

	// Favorite tools
	favorites map[string]bool

	// Tool aliases
	aliases map[string]string // alias -> full tool name

	// Workflow templates
	workflows map[string][]string // workflow name -> tool sequence

	// System mappings (resolume, ableton, etc.)
	systemMappings map[string][]string // system -> tool names
}

// RecentToolEntry tracks a tool invocation.
type RecentToolEntry struct {
	ToolName  string    `json:"tool_name"`
	Timestamp time.Time `json:"timestamp"`
	Category  string    `json:"category"`
}

// ToolRelation describes a relationship between tools.
type ToolRelation struct {
	ToolName    string   `json:"tool_name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	SharedTags  []string `json:"shared_tags"`
	Relevance   float64  `json:"relevance"` // 0-1 score
}

// WorkflowTemplate defines a sequence of tools for a goal.
type WorkflowTemplate struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Goal        string   `json:"goal"`
	Tools       []string `json:"tools"`
	Tags        []string `json:"tags"`
}

var (
	discoveryClientInstance *DiscoveryClient
	discoveryOnce           sync.Once
)

// GetDiscoveryClient returns the singleton discovery client.
func GetDiscoveryClient() *DiscoveryClient {
	discoveryOnce.Do(func() {
		discoveryClientInstance = &DiscoveryClient{
			maxRecent:      50,
			favorites:      make(map[string]bool),
			aliases:        make(map[string]string),
			workflows:      make(map[string][]string),
			systemMappings: initSystemMappings(),
		}
		initDefaultWorkflows(discoveryClientInstance)
	})
	return discoveryClientInstance
}

// initSystemMappings sets up the system -> tools mappings.
func initSystemMappings() map[string][]string {
	return map[string][]string{
		"resolume": {
			"aftrs_resolume_status", "aftrs_resolume_composition", "aftrs_resolume_layers",
			"aftrs_resolume_clips", "aftrs_resolume_clip_trigger", "aftrs_resolume_layer_control",
			"aftrs_resolume_bpm", "aftrs_resolume_bpm_set", "aftrs_resolume_params",
			"aftrs_resolume_effect_add", "aftrs_resolume_effect_remove", "aftrs_resolume_transition",
		},
		"ableton": {
			"aftrs_ableton_status", "aftrs_ableton_transport", "aftrs_ableton_tempo",
			"aftrs_ableton_tracks", "aftrs_ableton_scenes", "aftrs_ableton_clips",
			"aftrs_ableton_clip_fire", "aftrs_ableton_devices", "aftrs_ableton_params",
		},
		"touchdesigner": {
			"aftrs_td_status", "aftrs_td_nodes", "aftrs_td_node_create",
			"aftrs_td_node_params", "aftrs_td_operators", "aftrs_td_network",
			"aftrs_td_execute", "aftrs_td_project_save", "aftrs_td_project_load",
		},
		"obs": {
			"aftrs_obs_status", "aftrs_obs_scenes", "aftrs_obs_scene_switch",
			"aftrs_obs_sources", "aftrs_obs_streaming", "aftrs_obs_recording",
			"aftrs_obs_virtual_cam", "aftrs_obs_screenshot",
		},
		"grandma3": {
			"aftrs_gma3_status", "aftrs_gma3_playback", "aftrs_gma3_cues",
			"aftrs_gma3_executors", "aftrs_gma3_presets", "aftrs_gma3_command",
		},
		"discord": {
			"aftrs_discord_status", "aftrs_discord_send", "aftrs_discord_channels",
			"aftrs_discord_guilds", "aftrs_discord_users", "aftrs_discord_roles",
		},
		"midi": {
			"aftrs_midi_devices", "aftrs_midi_ports", "aftrs_midi_send_cc",
			"aftrs_midi_send_note", "aftrs_midi_monitor", "aftrs_midi_learn",
		},
		"ndi": {
			"aftrs_ndi_sources", "aftrs_ndi_status", "aftrs_ndi_preview",
			"aftrs_ndi_capture", "aftrs_ndi_cv_detect_faces", "aftrs_ndi_cv_detect_motion",
		},
		"atem": {
			"aftrs_atem_status", "aftrs_atem_inputs", "aftrs_atem_program",
			"aftrs_atem_preview", "aftrs_atem_cut", "aftrs_atem_transition",
		},
		"ptz": {
			"aftrs_ptz_status", "aftrs_ptz_move", "aftrs_ptz_preset_go",
			"aftrs_ptz_presets", "aftrs_ptz_home", "aftrs_ptz_cameras",
		},
		"lighting": {
			"aftrs_dmx_status", "aftrs_dmx_set", "aftrs_dmx_scene",
			"aftrs_dmx_blackout", "aftrs_wled_status", "aftrs_wled_preset",
		},
		"audio": {
			"aftrs_dante_status", "aftrs_dante_devices", "aftrs_dante_routes",
			"aftrs_whisper_transcribe", "aftrs_stems_separate",
		},
		"dj": {
			"aftrs_rekordbox_status", "aftrs_rekordbox_library", "aftrs_rekordbox_playlists",
			"aftrs_serato_status", "aftrs_serato_library", "aftrs_traktor_status",
		},
	}
}

// initDefaultWorkflows sets up common workflow templates.
func initDefaultWorkflows(c *DiscoveryClient) {
	c.workflows = map[string][]string{
		"show_startup": {
			"aftrs_dashboard_quick",
			"aftrs_resolume_status",
			"aftrs_gma3_status",
			"aftrs_obs_status",
			"aftrs_bpm_sync_status",
		},
		"bpm_sync": {
			"aftrs_bpm_sync_status",
			"aftrs_bpm_sync_master",
			"aftrs_bpm_sync_link",
			"aftrs_bpm_sync_push",
		},
		"video_routing": {
			"aftrs_ndi_sources",
			"aftrs_atem_inputs",
			"aftrs_obs_sources",
			"aftrs_resolume_layers",
		},
		"troubleshoot_audio": {
			"aftrs_dante_status",
			"aftrs_dante_devices",
			"aftrs_midi_devices",
			"aftrs_ableton_status",
		},
		"troubleshoot_video": {
			"aftrs_ndi_sources",
			"aftrs_atem_status",
			"aftrs_resolume_status",
			"aftrs_obs_status",
		},
		"emergency_stop": {
			"aftrs_dmx_blackout",
			"aftrs_resolume_layer_control",
			"aftrs_obs_streaming",
		},
	}
}

// RecordToolUse records a tool invocation for recent history.
func (c *DiscoveryClient) RecordToolUse(toolName, category string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := RecentToolEntry{
		ToolName:  toolName,
		Timestamp: time.Now(),
		Category:  category,
	}

	// Add to front
	c.recentTools = append([]RecentToolEntry{entry}, c.recentTools...)

	// Trim to max
	if len(c.recentTools) > c.maxRecent {
		c.recentTools = c.recentTools[:c.maxRecent]
	}
}

// GetRecentTools returns the most recent tool invocations.
func (c *DiscoveryClient) GetRecentTools(limit int) []RecentToolEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 || limit > len(c.recentTools) {
		limit = len(c.recentTools)
	}

	result := make([]RecentToolEntry, limit)
	copy(result, c.recentTools[:limit])
	return result
}

// AddFavorite marks a tool as favorite.
func (c *DiscoveryClient) AddFavorite(toolName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.favorites[toolName] = true
}

// RemoveFavorite removes a tool from favorites.
func (c *DiscoveryClient) RemoveFavorite(toolName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.favorites, toolName)
}

// IsFavorite checks if a tool is favorited.
func (c *DiscoveryClient) IsFavorite(toolName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.favorites[toolName]
}

// GetFavorites returns all favorited tools.
func (c *DiscoveryClient) GetFavorites() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, 0, len(c.favorites))
	for name := range c.favorites {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// SetAlias creates a short alias for a tool.
func (c *DiscoveryClient) SetAlias(alias, toolName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aliases[alias] = toolName
}

// RemoveAlias removes an alias.
func (c *DiscoveryClient) RemoveAlias(alias string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.aliases, alias)
}

// ResolveAlias resolves an alias to full tool name.
func (c *DiscoveryClient) ResolveAlias(alias string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	name, ok := c.aliases[alias]
	return name, ok
}

// GetAliases returns all aliases.
func (c *DiscoveryClient) GetAliases() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string, len(c.aliases))
	for k, v := range c.aliases {
		result[k] = v
	}
	return result
}

// SetFavorites replaces the entire favorites list.
func (c *DiscoveryClient) SetFavorites(favorites []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.favorites = make(map[string]bool)
	for _, f := range favorites {
		c.favorites[f] = true
	}
}

// SetAliases replaces the entire aliases map.
func (c *DiscoveryClient) SetAliases(aliases map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.aliases = make(map[string]string)
	for k, v := range aliases {
		c.aliases[k] = v
	}
}

// GetToolsForSystem returns tools associated with a system.
func (c *DiscoveryClient) GetToolsForSystem(system string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	system = strings.ToLower(system)
	if tools, ok := c.systemMappings[system]; ok {
		result := make([]string, len(tools))
		copy(result, tools)
		return result
	}
	return nil
}

// GetSystems returns all known systems.
func (c *DiscoveryClient) GetSystems() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	systems := make([]string, 0, len(c.systemMappings))
	for sys := range c.systemMappings {
		systems = append(systems, sys)
	}
	sort.Strings(systems)
	return systems
}

// GetWorkflow returns a workflow template by name.
func (c *DiscoveryClient) GetWorkflow(name string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tools, ok := c.workflows[name]
	if !ok {
		return nil, false
	}
	result := make([]string, len(tools))
	copy(result, tools)
	return result, true
}

// GetWorkflows returns all workflow names.
func (c *DiscoveryClient) GetWorkflows() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.workflows))
	for name := range c.workflows {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// FindRelatedTools finds tools related to a given tool based on category and tags.
func (c *DiscoveryClient) FindRelatedTools(toolName, category string, tags []string, allTools map[string]ToolInfo, limit int) []ToolRelation {
	var relations []ToolRelation

	tagSet := make(map[string]bool)
	for _, t := range tags {
		tagSet[strings.ToLower(t)] = true
	}

	for name, info := range allTools {
		if name == toolName {
			continue
		}

		var sharedTags []string
		relevance := 0.0

		// Same category is highly relevant
		if info.Category == category {
			relevance += 0.5
		}

		// Check shared tags
		for _, t := range info.Tags {
			if tagSet[strings.ToLower(t)] {
				sharedTags = append(sharedTags, t)
				relevance += 0.1
			}
		}

		if relevance > 0 {
			if relevance > 1.0 {
				relevance = 1.0
			}
			relations = append(relations, ToolRelation{
				ToolName:    name,
				Description: info.Description,
				Category:    info.Category,
				SharedTags:  sharedTags,
				Relevance:   relevance,
			})
		}
	}

	// Sort by relevance descending
	sort.Slice(relations, func(i, j int) bool {
		return relations[i].Relevance > relations[j].Relevance
	})

	if limit > 0 && len(relations) > limit {
		relations = relations[:limit]
	}

	return relations
}

// ToolInfo holds minimal tool information for relation finding.
type ToolInfo struct {
	Category    string
	Description string
	Tags        []string
}

// GetWorkflowsForGoal finds workflows matching a goal keyword.
func (c *DiscoveryClient) GetWorkflowsForGoal(goal string) []WorkflowTemplate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	goal = strings.ToLower(goal)
	var matches []WorkflowTemplate

	workflowDescriptions := map[string]string{
		"show_startup":       "Start up all systems and verify connectivity",
		"bpm_sync":           "Synchronize BPM across audio and visual systems",
		"video_routing":      "Configure video sources and routing",
		"troubleshoot_audio": "Diagnose audio and MIDI issues",
		"troubleshoot_video": "Diagnose video and NDI issues",
		"emergency_stop":     "Emergency blackout and stop all outputs",
	}

	workflowTags := map[string][]string{
		"show_startup":       {"startup", "health", "status", "check"},
		"bpm_sync":           {"bpm", "tempo", "sync", "music"},
		"video_routing":      {"video", "ndi", "routing", "sources"},
		"troubleshoot_audio": {"audio", "midi", "dante", "troubleshoot", "debug"},
		"troubleshoot_video": {"video", "ndi", "atem", "troubleshoot", "debug"},
		"emergency_stop":     {"emergency", "stop", "blackout", "panic"},
	}

	for name, tools := range c.workflows {
		desc := workflowDescriptions[name]
		tags := workflowTags[name]

		// Check if goal matches name, description, or tags
		matches_goal := strings.Contains(strings.ToLower(name), goal) ||
			strings.Contains(strings.ToLower(desc), goal)

		if !matches_goal {
			for _, tag := range tags {
				if strings.Contains(tag, goal) {
					matches_goal = true
					break
				}
			}
		}

		if matches_goal {
			matches = append(matches, WorkflowTemplate{
				Name:        name,
				Description: desc,
				Goal:        goal,
				Tools:       tools,
				Tags:        tags,
			})
		}
	}

	return matches
}
