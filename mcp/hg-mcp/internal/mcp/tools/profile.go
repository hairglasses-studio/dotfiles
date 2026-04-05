package tools

import (
	"os"
	"sort"
	"strings"
)

// Profile controls which tool domains load eagerly vs deferred.
type Profile string

const (
	ProfileDefault Profile = "default" // Core platform + show control essentials
	ProfileOps     Profile = "ops"     // + infrastructure + messaging
	ProfileFull    Profile = "full"    // Everything (current behavior)
)

// HgToolProfileFromEnv reads the HG_MCP_PROFILE environment variable and
// returns the normalized profile name (default, ops, or full).
func HgToolProfileFromEnv() string {
	return hgToolProfile()
}

func hgToolProfile() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("HG_MCP_PROFILE"))) {
	case "", "default":
		return "default"
	case "ops":
		return "ops"
	case "full":
		return "full"
	default:
		return "default"
	}
}

func shouldDeferTool(profile string, td ToolDefinition) bool {
	switch profile {
	case "full":
		return false
	case "ops":
		return td.RuntimeGroup != RuntimeGroupPlatform &&
			td.RuntimeGroup != RuntimeGroupInfrastructure &&
			td.RuntimeGroup != RuntimeGroupShowControl
	default:
		if td.RuntimeGroup == RuntimeGroupPlatform {
			return false
		}
		switch td.Category {
		case "consolidated", "workflows", "workflow_automation", "studio", "dashboard", "gateway":
			return false
		default:
			return true
		}
	}
}

// runtimeGroupToLabel provides human-readable names for runtime groups.
var runtimeGroupToLabel = map[string]string{
	RuntimeGroupDJMusic:         "DJ & Music",
	RuntimeGroupVJVideo:         "VJ & Video",
	RuntimeGroupLighting:        "Lighting",
	RuntimeGroupAudioProduction: "Audio Production",
	RuntimeGroupShowControl:     "Show Control",
	RuntimeGroupInfrastructure:  "Infrastructure",
	RuntimeGroupMessaging:       "Messaging",
	RuntimeGroupInventory:       "Inventory",
	RuntimeGroupStreaming:       "Streaming",
	RuntimeGroupPlatform:        "Platform",
}

// RuntimeGroupLabel returns the human-readable label for a runtime group.
func RuntimeGroupLabel(group string) string {
	if label, ok := runtimeGroupToLabel[group]; ok {
		return label
	}
	return group
}

// AllRuntimeGroups returns all known runtime groups sorted alphabetically.
func AllRuntimeGroups() []string {
	groups := make([]string, 0, len(runtimeGroupToLabel))
	for g := range runtimeGroupToLabel {
		groups = append(groups, g)
	}
	sort.Strings(groups)
	return groups
}

// EagerGroups returns runtime groups that load immediately for a profile.
func EagerGroups(profile string) []string {
	switch profile {
	case "ops":
		return []string{RuntimeGroupPlatform, RuntimeGroupInfrastructure, RuntimeGroupShowControl}
	case "full":
		return AllRuntimeGroups()
	default: // "default"
		return []string{RuntimeGroupPlatform}
	}
}

// DeferredGroups returns runtime groups that are deferred for a profile.
func DeferredGroups(profile string) []string {
	if profile == "full" {
		return nil
	}
	eager := make(map[string]bool)
	for _, g := range EagerGroups(profile) {
		eager[g] = true
	}
	var deferred []string
	for _, g := range AllRuntimeGroups() {
		if !eager[g] {
			deferred = append(deferred, g)
		}
	}
	return deferred
}
