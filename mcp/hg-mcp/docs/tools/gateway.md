# Gateway Module

> Unified gateway tools for domain operations - saves tokens by consolidating related tools

**5 tools** in this module

## Overview

Gateway tools provide single entry points for domain operations, consolidating multiple related tools into unified interfaces. This pattern:
- Reduces token usage by ~40% compared to calling individual tools
- Simplifies AI interactions with consistent parameter patterns
- Provides a cleaner abstraction for common workflows

## Tools

### aftrs_dj

Unified DJ operations for Serato, Rekordbox, and Traktor.

**Parameters:**
| Parameter | Required | Description |
|-----------|----------|-------------|
| `software` | Yes | Target software: `serato`, `rekordbox`, `traktor` |
| `action` | Yes | Operation: `status`, `search`, `playlists`, `track_info`, `history`, `crates`, `export`, `sync` |
| `query` | No | Search query or track/playlist ID |
| `limit` | No | Max results to return (default: 20) |

**Examples:**
```
# Check Serato status
aftrs_dj software=serato action=status

# Search Rekordbox library
aftrs_dj software=rekordbox action=search query="daft punk"

# List Traktor playlists
aftrs_dj software=traktor action=playlists
```

---

### aftrs_av

Unified AV control for Resolume, TouchDesigner, and OBS.

**Parameters:**
| Parameter | Required | Description |
|-----------|----------|-------------|
| `software` | Yes | Target software: `resolume`, `touchdesigner`, `obs` |
| `action` | Yes | Operation: `status`, `health`, `layers`, `clips`, `trigger`, `effects`, `output`, `scenes`, `sources` |
| `target` | No | Target layer/clip/scene name or ID |
| `value` | No | Value to set (for trigger, effects) |

**Examples:**
```
# Check Resolume status
aftrs_av software=resolume action=status

# Trigger OBS scene
aftrs_av software=obs action=trigger target="Main Camera"

# Get TouchDesigner health
aftrs_av software=touchdesigner action=health
```

---

### aftrs_lighting

Unified lighting control for grandMA3, DMX/ArtNet, and WLED.

**Parameters:**
| Parameter | Required | Description |
|-----------|----------|-------------|
| `system` | Yes | Target system: `grandma3`, `dmx`, `wled` |
| `action` | Yes | Operation: `status`, `health`, `fixtures`, `scenes`, `presets`, `blackout`, `color`, `effect`, `intensity` |
| `target` | No | Target fixture/scene/preset name or IP |
| `value` | No | Value to set (color hex, intensity 0-100, effect name) |

**Examples:**
```
# grandMA3 status
aftrs_lighting system=grandma3 action=status

# Set WLED color
aftrs_lighting system=wled action=color target="192.168.1.100" value="255,0,128"

# DMX blackout
aftrs_lighting system=dmx action=blackout
```

---

### aftrs_audio

Unified audio control for Ableton Live, Dante, and MIDI.

**Parameters:**
| Parameter | Required | Description |
|-----------|----------|-------------|
| `system` | Yes | Target system: `ableton`, `dante`, `midi` |
| `action` | Yes | Operation: `status`, `transport`, `tracks`, `devices`, `routing`, `bpm`, `play`, `stop`, `record` |
| `target` | No | Target track/device/route name or ID |
| `value` | No | Value to set (bpm, volume, etc) |

**Examples:**
```
# Ableton status
aftrs_audio system=ableton action=status

# Set BPM
aftrs_audio system=ableton action=bpm value="128"

# Dante routing
aftrs_audio system=dante action=routing
```

---

### aftrs_streaming

Unified streaming control for Twitch, YouTube Live, and NDI.

**Parameters:**
| Parameter | Required | Description |
|-----------|----------|-------------|
| `platform` | Yes | Target platform: `twitch`, `youtube`, `ndi` |
| `action` | Yes | Operation: `status`, `go_live`, `end_stream`, `sources`, `chat`, `viewers`, `title`, `game` |
| `value` | No | Value to set (title, message, game) |

**Examples:**
```
# Twitch status
aftrs_streaming platform=twitch action=status

# Update stream title
aftrs_streaming platform=twitch action=title value="Late Night Coding"

# NDI sources
aftrs_streaming platform=ndi action=sources
```

---

## Architecture

Gateway tools follow the "mega-tool" pattern from webb MCP server architecture. Each gateway:

1. **Routes by domain** - First parameter specifies the target system
2. **Routes by action** - Second parameter specifies the operation
3. **Returns structured output** - Consistent markdown-formatted responses
4. **Handles errors gracefully** - Client errors wrapped with helpful hints

```
┌───────────────────────────────────────────────────────┐
│                    aftrs_av Gateway                   │
├─────────────────┬─────────────────┬──────────────────┤
│    resolume     │  touchdesigner  │       obs        │
│   ───────────   │   ───────────   │   ───────────    │
│   status        │   status        │   status         │
│   layers        │   health        │   scenes         │
│   clips         │   network       │   sources        │
│   trigger       │   pulse         │   trigger        │
└─────────────────┴─────────────────┴──────────────────┘
```

## Token Savings

| Scenario | Without Gateway | With Gateway | Savings |
|----------|-----------------|--------------|---------|
| Check 3 DJ softwares | 3 tool calls | 1 tool call | ~66% |
| Query AV status | 3+ separate calls | 1 call | ~50% |
| Multi-system lighting | Multiple tools | 1 gateway | ~40% |

## Tags

`gateway`, `dj`, `av`, `lighting`, `audio`, `streaming`, `consolidated`, `unified`
