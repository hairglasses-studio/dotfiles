# mapping

Shared Go package for the mapitall controller-mapping ecosystem. Provides the unified type system, rule-resolution engine, TOML profile parsing, and legacy-format conversion used by MCP tool servers and the mapping daemon.

## Build & Test

```bash
go build ./...
go vet ./...
go test ./... -count=1 -race
```

## File Map

| File | Purpose |
|------|---------|
| `types.go` | Core data model: `MappingRule`, `MappingProfile`, `OutputAction`, `Condition`, `ValueTransform`, `ProfileMeta`, `DeviceConfig`, `MappingSettings`, `AppOverride`. All TOML/JSON serializable. |
| `engine.go` | Rule-resolution engine: `RuleIndex` (O(1) source lookup), `BuildRuleIndex()`, `Resolve()`, `matchBest()`. `EngineState` holds runtime mutable state (active modifiers, app, layer, variables) with RWMutex concurrency safety. |
| `profile.go` | TOML profile loading: `LoadMappingProfile()` reads from disk, `ParseMappingProfile()` parses a string. Auto-detects unified vs legacy makima format. |
| `convert.go` | Legacy-to-unified migration: `ConvertLegacyToUnified()` transforms makima v1 flat profiles into the schema-versioned unified format. `ParseLegacyInput()` handles the `MOD-MOD-EVENT` string convention. |
| `lut.go` | Pre-compiled 256-entry lookup tables (`CompiledTransform`) for fast value transformation on the hot path. Avoids per-event log/exp/scurve computation. |
| `validate.go` | Profile validation: `ValidateProfile()` checks for missing fields, bad curve types, and format issues. Returns typed `ValidationIssue` slices. |
| `discovery.go` | Profile directory scanning: `ListMappingProfiles()` enumerates TOML files and produces `MappingProfileSummary` listings. |

## Key Types

- **`MappingRule`** -- Atomic input-to-output binding. Fields: Input, Modifiers, Output, Value (transform), Condition, Priority, Layer.
- **`MappingProfile`** -- Complete device config: profile metadata, device config, settings, mapping rules, app overrides. Supports both unified (schema_version >= 1) and legacy makima formats.
- **`RuleIndex`** -- Compiled lookup structure built from a `MappingProfile`. Call `Resolve(source, state, deviceID)` for O(1) best-match resolution with modifier, layer, condition, priority, and app-override support.
- **`EngineState`** -- Mutable runtime state: active modifiers, active app (window class), per-device active layer, variables, pickup/fader state. All methods are concurrency-safe (RWMutex).
- **`OutputAction`** -- What a rule produces: key press, command exec, mouse movement, OSC message, WebSocket message, MIDI out, D-Bus call, layer switch, variable toggle/set, or sequence of steps.
- **`ValueTransform`** -- Continuous value scaling with configurable curve (linear, log, exp, S-curve), input/output ranges, relative mode, threshold, and pickup behavior.
- **`Condition`** -- Gates a rule on engine variable state (equals, not_equal, greater_than, less_than).
