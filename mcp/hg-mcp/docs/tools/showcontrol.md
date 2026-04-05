# showcontrol

> High-level show control and multi-system orchestration

**4 tools**

## Tools

- [`aftrs_show_emergency`](#aftrs-show-emergency)
- [`aftrs_show_start`](#aftrs-show-start)
- [`aftrs_show_status`](#aftrs-show-status)
- [`aftrs_show_stop`](#aftrs-show-stop)

---

## aftrs_show_emergency

Emergency stop all systems immediately. Stops all playback, triggers blackout, and mutes audio.

**Complexity:** simple

**Tags:** `emergency`, `stop`, `blackout`, `safety`

**Use Cases:**
- Emergency situation
- Technical failure
- Safety stop

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `reason` | string |  | Reason for emergency stop |

### Example

```json
{
  "reason": "example"
}
```

---

## aftrs_show_start

Start a show by initializing all systems. Runs health checks, syncs BPM, and prepares systems for performance.

**Complexity:** complex

**Tags:** `show`, `start`, `initialize`, `performance`

**Use Cases:**
- Start a DJ set
- Initialize live performance
- Begin show

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm` | number |  | Initial BPM (default: 120) |
| `show_name` | string |  | Name of the show/set |
| `skip_health_check` | boolean |  | Skip initial health checks |

### Example

```json
{
  "bpm": 0,
  "show_name": "example",
  "skip_health_check": false
}
```

---

## aftrs_show_status

Get comprehensive show status across all systems. Shows health, sync status, and performance metrics.

**Complexity:** simple

**Tags:** `show`, `status`, `health`, `overview`

**Use Cases:**
- Check show health
- Monitor systems
- Pre-show checklist

---

## aftrs_show_stop

Stop the show gracefully. Fades out audio/video, stops playback, and saves state.

**Complexity:** complex

**Tags:** `show`, `stop`, `end`, `fade`

**Use Cases:**
- End a performance
- Graceful shutdown
- Emergency stop

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `blackout` | boolean |  | Trigger blackout on lighting (default: true) |
| `fade_time` | number |  | Fade out time in seconds (default: 3) |
| `save_snapshot` | boolean |  | Save current state before stopping (default: true) |

### Example

```json
{
  "blackout": false,
  "fade_time": 0,
  "save_snapshot": false
}
```

---

