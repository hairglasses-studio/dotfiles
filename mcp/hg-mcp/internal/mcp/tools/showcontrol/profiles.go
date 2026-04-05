package showcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// ShowProfile defines a reusable show configuration
type ShowProfile struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Systems     []string `json:"systems,omitempty"`    // Systems to check: ableton, resolume, obs, grandma3
	InitialBPM  float64  `json:"initial_bpm,omitempty"` // BPM to set at start (0 = skip)
	SnapshotID  string   `json:"snapshot_id,omitempty"` // Snapshot to recall at start
	Chains      []string `json:"chains,omitempty"`      // Chain IDs to run after start
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func profilesDir() string {
	return filepath.Join(config.Get().AftrsDataDir, "profiles")
}

func loadProfile(name string) (*ShowProfile, error) {
	data, err := os.ReadFile(filepath.Join(profilesDir(), name+".json"))
	if err != nil {
		return nil, fmt.Errorf("profile not found: %s", name)
	}
	var p ShowProfile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("invalid profile: %w", err)
	}
	return &p, nil
}

func saveProfile(p *ShowProfile) error {
	dir := profilesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, p.Name+".json"), data, 0644)
}

func handleShowProfileSave(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	profile := &ShowProfile{
		Name:        name,
		Description: tools.GetStringParam(req, "description"),
		InitialBPM:  tools.GetFloatParam(req, "bpm", 0),
		SnapshotID:  tools.GetStringParam(req, "snapshot_id"),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Parse systems
	if sys := tools.GetStringParam(req, "systems"); sys != "" {
		profile.Systems = splitCSV(sys)
	}

	// Parse chains
	if ch := tools.GetStringParam(req, "chains"); ch != "" {
		profile.Chains = splitCSV(ch)
	}

	// Check if exists to preserve created_at
	existing, err := loadProfile(name)
	if err == nil {
		profile.CreatedAt = existing.CreatedAt
	} else {
		profile.CreatedAt = profile.UpdatedAt
	}

	if err := saveProfile(profile); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Show profile **%s** saved.\n\n- BPM: %.0f\n- Systems: %v\n- Snapshot: %s\n- Chains: %v",
		name, profile.InitialBPM, profile.Systems, profile.SnapshotID, profile.Chains)), nil
}

func handleShowProfileLoad(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	profile, err := loadProfile(name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"name":        profile.Name,
		"description": profile.Description,
		"systems":     profile.Systems,
		"initial_bpm": profile.InitialBPM,
		"snapshot_id": profile.SnapshotID,
		"chains":      profile.Chains,
		"created_at":  profile.CreatedAt,
		"updated_at":  profile.UpdatedAt,
	}), nil
}

func handleShowProfileList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dir := profilesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return tools.TextResult("No show profiles found. Use `aftrs_show_profile_save` to create one."), nil
		}
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Show Profiles\n\n")

	count := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		profile, err := loadProfile(name)
		if err != nil {
			continue
		}
		count++
		bpm := "default"
		if profile.InitialBPM > 0 {
			bpm = fmt.Sprintf("%.0f", profile.InitialBPM)
		}
		sb.WriteString(fmt.Sprintf("- **%s**: %s (BPM: %s, Systems: %d, Chains: %d)\n",
			profile.Name, profile.Description, bpm, len(profile.Systems), len(profile.Chains)))
	}

	if count == 0 {
		sb.WriteString("No profiles found.\n")
	}

	return tools.TextResult(sb.String()), nil
}

func splitCSV(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		p := strings.TrimSpace(part)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
