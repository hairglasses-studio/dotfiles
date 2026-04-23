// mod_events.go — surface structured events from the dotfiles-event-bus
// daemon via MCP tools. The daemon writes to ~/.local/state/dotfiles/
// events.jsonl (see scripts/event-bus.py); this module reads that stream
// and returns filtered, bounded slices of it.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hairglasses-studio/mcpkit/handler"
	"github.com/hairglasses-studio/mcpkit/registry"
)

// EventsTailInput selects which events to return.
type EventsTailInput struct {
	// SinceMinutes is the lookback window; 0 returns all events up to Limit.
	SinceMinutes int `json:"since_minutes,omitempty" jsonschema:"description=Return only events within this many minutes of now (0 = no time filter)"`
	// Type filters by the event type field (matches error_code in most rules).
	Type string `json:"type,omitempty" jsonschema:"description=Filter by event type (e.g. hypr_reload_induced_drm, audio_sink_lost). Omit for all types."`
	// Severity filters to low | medium | high when set.
	Severity string `json:"severity,omitempty" jsonschema:"enum=low,enum=medium,enum=high,description=Filter to events at or above this severity"`
	// Limit caps the number of results; default 50.
	Limit int `json:"limit,omitempty" jsonschema:"description=Max events to return (default: 50, max: 500)"`
}

// EventRecord mirrors the JSONL schema written by event-bus.py.
type EventRecord struct {
	Type        string          `json:"type"`
	At          string          `json:"at"`
	Fingerprint string          `json:"fingerprint,omitempty"`
	ErrorCode   string          `json:"error_code,omitempty"`
	Severity    string          `json:"severity,omitempty"`
	Rule        string          `json:"rule,omitempty"`
	Source      string          `json:"source,omitempty"`
	Correlation json.RawMessage `json:"correlation,omitempty"`
}

type EventsTailOutput struct {
	Events []EventRecord `json:"events"`
	Count  int           `json:"count"`
	Path   string        `json:"path"`
	Exists bool          `json:"exists"`
}

// EventsModule exposes the event-bus stream as a read-only MCP surface.
type EventsModule struct{}

func (m *EventsModule) Name() string { return "events" }
func (m *EventsModule) Description() string {
	return "Read structured events from the dotfiles-event-bus daemon. Each event carries a remediation error_code that consumers can dispatch via remediation_lookup."
}

func (m *EventsModule) Tools() []registry.ToolDefinition {
	return []registry.ToolDefinition{
		handler.TypedHandler[EventsTailInput, EventsTailOutput](
			"events_tail",
			"Tail structured events from ~/.local/state/dotfiles/events.jsonl. Filters by time window, type, and severity. Use this to drive /heal (pending fixes) and /canary (post-deploy liveness).",
			func(_ context.Context, input EventsTailInput) (EventsTailOutput, error) {
				path := eventsLogPath()
				out := EventsTailOutput{Path: path, Events: []EventRecord{}}

				f, err := os.Open(path)
				if err != nil {
					if os.IsNotExist(err) {
						// Not an error — the event bus may simply not have
						// started yet. Return an empty tail with exists=false.
						out.Exists = false
						return out, nil
					}
					return out, err
				}
				defer f.Close()
				out.Exists = true

				limit := input.Limit
				if limit <= 0 {
					limit = 50
				}
				if limit > 500 {
					limit = 500
				}

				var cutoff time.Time
				if input.SinceMinutes > 0 {
					cutoff = time.Now().Add(-time.Duration(input.SinceMinutes) * time.Minute)
				}

				minSeverity := severityRank(input.Severity)
				wantType := strings.TrimSpace(input.Type)

				// Read all lines, filter, then take the tail — events.jsonl
				// is small (append-only, rotated externally), so a single
				// pass is fine.
				var all []EventRecord
				sc := bufio.NewScanner(f)
				sc.Buffer(make([]byte, 64*1024), 1024*1024)
				for sc.Scan() {
					var rec EventRecord
					if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
						continue // skip corrupt line
					}
					if wantType != "" && rec.Type != wantType && rec.ErrorCode != wantType {
						continue
					}
					if minSeverity > 0 && severityRank(rec.Severity) < minSeverity {
						continue
					}
					if !cutoff.IsZero() {
						if at, err := time.Parse(time.RFC3339, rec.At); err == nil && at.Before(cutoff) {
							continue
						}
					}
					all = append(all, rec)
				}
				if err := sc.Err(); err != nil {
					return out, fmt.Errorf("scan events.jsonl: %w", err)
				}

				if len(all) > limit {
					all = all[len(all)-limit:]
				}
				out.Events = all
				out.Count = len(all)
				return out, nil
			},
		),
	}
}

func eventsLogPath() string {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		base = filepath.Join(homeDir(), ".local", "state")
	}
	return filepath.Join(base, "dotfiles", "events.jsonl")
}

// severityRank returns an ordering so "low" < "medium" < "high". Unknown
// values rank 0 so they pass any min-severity filter.
func severityRank(s string) int {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "low":
		return 1
	case "medium":
		return 2
	case "high":
		return 3
	default:
		return 0
	}
}
