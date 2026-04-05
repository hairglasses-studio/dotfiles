# showkontrol

> Showkontrol timecode and cue management for synchronized live performances

**10 tools**

## Tools

- [`aftrs_showkontrol_cue_go`](#aftrs-showkontrol-cue-go)
- [`aftrs_showkontrol_cues`](#aftrs-showkontrol-cues)
- [`aftrs_showkontrol_go`](#aftrs-showkontrol-go)
- [`aftrs_showkontrol_health`](#aftrs-showkontrol-health)
- [`aftrs_showkontrol_show`](#aftrs-showkontrol-show)
- [`aftrs_showkontrol_shows`](#aftrs-showkontrol-shows)
- [`aftrs_showkontrol_status`](#aftrs-showkontrol-status)
- [`aftrs_showkontrol_timecode`](#aftrs-showkontrol-timecode)
- [`aftrs_showkontrol_timecode_start`](#aftrs-showkontrol-timecode-start)
- [`aftrs_showkontrol_timecode_stop`](#aftrs-showkontrol-timecode-stop)

---

## aftrs_showkontrol_cue_go

Fire a specific cue by ID

**Complexity:** moderate

**Tags:** `showkontrol`, `cue`, `fire`, `go`

**Use Cases:**
- Fire cue
- Trigger specific cue
- Manual cue execution

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `cue_id` | string | Yes | Cue ID to fire |

### Example

```json
{
  "cue_id": "example"
}
```

---

## aftrs_showkontrol_cues

List cues in the current show

**Complexity:** simple

**Tags:** `showkontrol`, `cues`, `list`

**Use Cases:**
- List cues
- View cue list
- Check cue timings

---

## aftrs_showkontrol_go

Fire the next cue in sequence

**Complexity:** simple

**Tags:** `showkontrol`, `go`, `next`, `cue`

**Use Cases:**
- Fire next cue
- Advance show
- Continue sequence

---

## aftrs_showkontrol_health

Check Showkontrol connection health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `showkontrol`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Check connection
- Diagnose issues
- Verify setup

---

## aftrs_showkontrol_show

Get or load a specific show

**Complexity:** moderate

**Tags:** `showkontrol`, `show`, `load`

**Use Cases:**
- Load show
- Get show details
- Switch shows

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `load` | boolean |  | If true, load the show as current |
| `show_id` | string | Yes | Show ID to get or load |

### Example

```json
{
  "load": false,
  "show_id": "example"
}
```

---

## aftrs_showkontrol_shows

List available shows in Showkontrol

**Complexity:** simple

**Tags:** `showkontrol`, `shows`, `list`

**Use Cases:**
- List shows
- Find show IDs
- Browse available shows

---

## aftrs_showkontrol_status

Get Showkontrol system status including timecode and current show

**Complexity:** simple

**Tags:** `showkontrol`, `timecode`, `status`, `cue`

**Use Cases:**
- Check system status
- View timecode
- Get current show

---

## aftrs_showkontrol_timecode

Get or set timecode position

**Complexity:** simple

**Tags:** `showkontrol`, `timecode`, `position`

**Use Cases:**
- Get timecode
- Jump to position
- Set playhead

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `position` | string |  | Position to jump to (HH:MM:SS:FF or seconds). Omit to get current position. |

### Example

```json
{
  "position": "example"
}
```

---

## aftrs_showkontrol_timecode_start

Start timecode playback

**Complexity:** simple

**Tags:** `showkontrol`, `timecode`, `start`, `play`

**Use Cases:**
- Start timecode
- Begin playback
- Run show

---

## aftrs_showkontrol_timecode_stop

Stop timecode playback

**Complexity:** simple

**Tags:** `showkontrol`, `timecode`, `stop`

**Use Cases:**
- Stop timecode
- Pause show
- Halt playback

---

