// mod_keybinds.go — Hyprland keybind inventory, search, audit, and ticker management
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/mcpkit/handler"
	"github.com/hairglasses-studio/mcpkit/registry"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type hyprBind struct {
	Modifiers   string `json:"modifiers"`
	Key         string `json:"key"`
	Description string `json:"description"`
	Dispatcher  string `json:"dispatcher"`
	Args        string `json:"args"`
	Locked      bool   `json:"locked"`
	Mouse       bool   `json:"mouse"`
	Submap      string `json:"submap,omitempty"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func hyprctlBinds() ([]map[string]any, error) {
	cmd := exec.Command("hyprctl", "binds", "-j")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("hyprctl binds -j failed: %w", err)
	}
	var raw []map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("parse hyprctl binds: %w", err)
	}
	return raw, nil
}

func decodeModmask(mask int) string {
	var parts []string
	if mask&64 != 0 {
		parts = append(parts, "Super")
	}
	if mask&1 != 0 {
		parts = append(parts, "Shift")
	}
	if mask&4 != 0 {
		parts = append(parts, "Ctrl")
	}
	if mask&8 != 0 {
		parts = append(parts, "Alt")
	}
	return strings.Join(parts, "+")
}

func rawGetString(raw map[string]any, key string) string {
	if v, ok := raw[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func rawGetBool(raw map[string]any, key string) bool {
	if v, ok := raw[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func rawGetInt(raw map[string]any, key string) int {
	if v, ok := raw[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return 0
}

func bindFromRaw(raw map[string]any) hyprBind {
	return hyprBind{
		Modifiers:   decodeModmask(rawGetInt(raw, "modmask")),
		Key:         rawGetString(raw, "key"),
		Description: rawGetString(raw, "description"),
		Dispatcher:  rawGetString(raw, "dispatcher"),
		Args:        rawGetString(raw, "arg"),
		Locked:      rawGetBool(raw, "locked"),
		Mouse:       rawGetBool(raw, "mouse"),
		Submap:      rawGetString(raw, "submap"),
	}
}

func describedBinds() ([]hyprBind, error) {
	raw, err := hyprctlBinds()
	if err != nil {
		return nil, err
	}
	var binds []hyprBind
	for _, r := range raw {
		if !rawGetBool(r, "has_description") {
			continue
		}
		if rawGetBool(r, "mouse") {
			continue
		}
		if rawGetString(r, "submap") != "" {
			continue
		}
		binds = append(binds, bindFromRaw(r))
	}
	return binds, nil
}

// ---------------------------------------------------------------------------
// Module
// ---------------------------------------------------------------------------

// KeybindsModule provides Hyprland keybind inventory and audit tools.
type KeybindsModule struct{}

func (m *KeybindsModule) Name() string        { return "keybinds" }
func (m *KeybindsModule) Description() string { return "Hyprland keybind inventory, search, audit, and ticker management" }

func (m *KeybindsModule) Tools() []registry.ToolDefinition {
	return []registry.ToolDefinition{
		// ── keybinds_list ─────────────────────────────
		{
			Tool: mcp.Tool{
				Name:        "keybinds_list",
				Description: "List all Hyprland keybinds with descriptions, matching the ticker display. Returns structured JSON with modifiers, key, description, dispatcher, and args.",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]any{},
				},
			},
			Handler: func(_ context.Context, _ registry.CallToolRequest) (*registry.CallToolResult, error) {
				binds, err := describedBinds()
				if err != nil {
					return handler.ErrorResult(err), nil
				}
				result := map[string]any{
					"binds": binds,
					"count": len(binds),
				}
				b, _ := json.MarshalIndent(result, "", "  ")
				return &registry.CallToolResult{
					Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(b)}},
				}, nil
			},
		},

		// ── keybinds_search ──────────────────────────
		{
			Tool: mcp.Tool{
				Name:        "keybinds_search",
				Description: "Search keybinds by description or key name. Case-insensitive match against description and key fields.",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Search term to match against keybind descriptions and key names",
						},
					},
					Required: []string{"query"},
				},
			},
			Handler: func(_ context.Context, req registry.CallToolRequest) (*registry.CallToolResult, error) {
				var input struct {
					Query string `json:"query"`
				}
				if req.Params.Arguments != nil {
					b, _ := json.Marshal(req.Params.Arguments)
					json.Unmarshal(b, &input)
				}
				if input.Query == "" {
					return handler.CodedErrorResult(handler.ErrInvalidParam, fmt.Errorf("query must not be empty")), nil
				}

				binds, err := describedBinds()
				if err != nil {
					return handler.ErrorResult(err), nil
				}

				q := strings.ToLower(input.Query)
				var matches []hyprBind
				for _, bind := range binds {
					if strings.Contains(strings.ToLower(bind.Description), q) ||
						strings.Contains(strings.ToLower(bind.Key), q) ||
						strings.Contains(strings.ToLower(bind.Modifiers), q) {
						matches = append(matches, bind)
					}
				}

				result := map[string]any{
					"query":   input.Query,
					"matches": matches,
					"count":   len(matches),
				}
				b, _ := json.MarshalIndent(result, "", "  ")
				return &registry.CallToolResult{
					Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(b)}},
				}, nil
			},
		},

		// ── keybinds_free_slots ──────────────────────
		{
			Tool: mcp.Tool{
				Name:        "keybinds_free_slots",
				Description: "Show unbound key combinations available for new keybinds. Reports free Super+letter, Super+number, and modifier combos.",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]any{},
				},
			},
			Handler: func(_ context.Context, _ registry.CallToolRequest) (*registry.CallToolResult, error) {
				binds, err := describedBinds()
				if err != nil {
					return handler.ErrorResult(err), nil
				}

				// Build occupied set: "Super+K", "Super+Shift+Q", etc.
				occupied := make(map[string]bool)
				superLetters := make(map[string]bool) // track which letters have a plain Super bind
				for _, bind := range binds {
					combo := bind.Key
					if bind.Modifiers != "" {
						combo = bind.Modifiers + "+" + bind.Key
					}
					occupied[combo] = true
					if bind.Modifiers == "Super" && len(bind.Key) == 1 {
						superLetters[bind.Key] = true
					}
				}

				// Check Super+A through Super+Z
				var freeLetters []string
				for c := 'A'; c <= 'Z'; c++ {
					key := string(c)
					if !occupied["Super+"+key] && !occupied["Super+"+strings.ToLower(key)] {
						freeLetters = append(freeLetters, key)
					}
				}

				// Check Super+0 through Super+9
				var freeNumbers []string
				for c := '0'; c <= '9'; c++ {
					key := string(c)
					if !occupied["Super+"+key] {
						freeNumbers = append(freeNumbers, key)
					}
				}

				// Check modifier combos on bound letters
				type freeMod struct {
					Modifier string `json:"modifier"`
					Key      string `json:"key"`
				}
				var freeCombos []freeMod
				modifiers := []string{"Super+Shift", "Super+Ctrl", "Super+Alt"}
				for letter := range superLetters {
					for _, mod := range modifiers {
						combo := mod + "+" + letter
						comboLower := mod + "+" + strings.ToLower(letter)
						if !occupied[combo] && !occupied[comboLower] {
							freeCombos = append(freeCombos, freeMod{Modifier: mod, Key: strings.ToUpper(letter)})
						}
					}
				}

				sort.Strings(freeLetters)
				sort.Strings(freeNumbers)

				result := map[string]any{
					"free_letters": freeLetters,
					"free_numbers": freeNumbers,
					"free_combos":  freeCombos,
					"summary":      fmt.Sprintf("%d of 26 letters free, %d of 10 numbers free, %d modifier combos on bound letters free", len(freeLetters), len(freeNumbers), len(freeCombos)),
				}
				b, _ := json.MarshalIndent(result, "", "  ")
				return &registry.CallToolResult{
					Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(b)}},
				}, nil
			},
		},

		// ── keybinds_conflicts ───────────────────────
		{
			Tool: mcp.Tool{
				Name:        "keybinds_conflicts",
				Description: "Check for conflicting or duplicate keybinds. Reports any key+modifier combinations bound more than once.",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]any{},
				},
			},
			Handler: func(_ context.Context, _ registry.CallToolRequest) (*registry.CallToolResult, error) {
				raw, err := hyprctlBinds()
				if err != nil {
					return handler.ErrorResult(err), nil
				}

				// Group by modmask+key
				type groupKey struct {
					Modmask int
					Key     string
					Submap  string
				}
				groups := make(map[string][]hyprBind)
				for _, r := range raw {
					mask := rawGetInt(r, "modmask")
					key := rawGetString(r, "key")
					sub := rawGetString(r, "submap")
					gk := fmt.Sprintf("%d|%s|%s", mask, key, sub)
					groups[gk] = append(groups[gk], bindFromRaw(r))
				}

				var conflicts []map[string]any
				for _, binds := range groups {
					if len(binds) > 1 {
						conflicts = append(conflicts, map[string]any{
							"key":       binds[0].Modifiers + "+" + binds[0].Key,
							"count":     len(binds),
							"entries":   binds,
						})
					}
				}

				result := map[string]any{
					"conflicts":      conflicts,
					"conflict_count": len(conflicts),
				}
				if len(conflicts) == 0 {
					result["message"] = "No duplicate keybinds detected"
				}
				b, _ := json.MarshalIndent(result, "", "  ")
				return &registry.CallToolResult{
					Content: []mcp.Content{mcp.TextContent{Type: "text", Text: string(b)}},
				}, nil
			},
		},

		// ── keybinds_refresh_ticker ──────────────────
		{
			Tool: mcp.Tool{
				Name:        "keybinds_refresh_ticker",
				Description: "Restart the keybind ticker bar to pick up new keybinds immediately instead of waiting for the 5-minute poll.",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]any{},
				},
			},
			Handler: func(_ context.Context, _ registry.CallToolRequest) (*registry.CallToolResult, error) {
				cmd := exec.Command("systemctl", "--user", "restart", "dotfiles-keybind-ticker.service")
				out, err := cmd.CombinedOutput()
				if err != nil {
					return handler.ErrorResult(fmt.Errorf("ticker restart failed: %w: %s", err, string(out))), nil
				}
				return &registry.CallToolResult{
					Content: []mcp.Content{mcp.TextContent{Type: "text", Text: "Keybind ticker restarted — new binds will appear within seconds"}},
				}, nil
			},
		},
	}
}
