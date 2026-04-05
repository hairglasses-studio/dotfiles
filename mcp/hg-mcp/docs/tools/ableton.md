# ableton

> Ableton Live control via AbletonOSC for transport, clips, tracks, and devices

**12 tools**

## Tools

- [`aftrs_ableton_clip_fire`](#aftrs-ableton-clip-fire)
- [`aftrs_ableton_clips`](#aftrs-ableton-clips)
- [`aftrs_ableton_cue_points`](#aftrs-ableton-cue-points)
- [`aftrs_ableton_device`](#aftrs-ableton-device)
- [`aftrs_ableton_devices`](#aftrs-ableton-devices)
- [`aftrs_ableton_health`](#aftrs-ableton-health)
- [`aftrs_ableton_scene_fire`](#aftrs-ableton-scene-fire)
- [`aftrs_ableton_status`](#aftrs-ableton-status)
- [`aftrs_ableton_tempo`](#aftrs-ableton-tempo)
- [`aftrs_ableton_track`](#aftrs-ableton-track)
- [`aftrs_ableton_tracks`](#aftrs-ableton-tracks)
- [`aftrs_ableton_transport`](#aftrs-ableton-transport)

---

## aftrs_ableton_clip_fire

Trigger a specific clip

**Complexity:** moderate

**Tags:** `ableton`, `clip`, `fire`, `trigger`

**Use Cases:**
- Trigger clip
- Launch clip
- Fire specific slot

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `slot_index` | integer | Yes | Clip slot index (0-based) |
| `track_index` | integer | Yes | Track index (0-based) |

### Example

```json
{
  "slot_index": 0,
  "track_index": 0
}
```

---

## aftrs_ableton_clips

List clips in a track

**Complexity:** simple

**Tags:** `ableton`, `clips`, `list`

**Use Cases:**
- List track clips
- View clip slots
- Check clip status

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `track_index` | integer | Yes | Track index (0-based) |

### Example

```json
{
  "track_index": 0
}
```

---

## aftrs_ableton_cue_points

Get or jump to arrangement cue points

**Complexity:** simple

**Tags:** `ableton`, `cue`, `arrangement`, `markers`

**Use Cases:**
- List cue points
- Jump to marker
- Navigate arrangement

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `jump_to` | integer |  | Cue point index to jump to (optional) |

### Example

```json
{
  "jump_to": 0
}
```

---

## aftrs_ableton_device

Get or set device parameters

**Complexity:** moderate

**Tags:** `ableton`, `device`, `parameters`, `automation`

**Use Cases:**
- Get device parameters
- Automate effects
- Control instruments

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `device_index` | integer | Yes | Device index (0-based) |
| `param_index` | integer |  | Parameter index to set (optional) |
| `track_index` | integer | Yes | Track index (0-based) |
| `value` | number |  | Parameter value to set (0.0-1.0) |

### Example

```json
{
  "device_index": 0,
  "param_index": 0,
  "track_index": 0,
  "value": 0
}
```

---

## aftrs_ableton_devices

List devices on a track

**Complexity:** simple

**Tags:** `ableton`, `devices`, `instruments`, `effects`

**Use Cases:**
- List track devices
- View effects chain
- Find instruments

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `track_index` | integer | Yes | Track index (0-based) |

### Example

```json
{
  "track_index": 0
}
```

---

## aftrs_ableton_health

Check Ableton Live connection health and AbletonOSC status

**Complexity:** simple

**Tags:** `ableton`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Check connection
- Diagnose issues
- Verify AbletonOSC

---

## aftrs_ableton_scene_fire

Trigger a scene (horizontal row of clips)

**Complexity:** moderate

**Tags:** `ableton`, `scene`, `fire`, `trigger`

**Use Cases:**
- Trigger scene
- Launch scene row
- Fire all clips in row

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `scene_index` | integer | Yes | Scene index (0-based) |

### Example

```json
{
  "scene_index": 0
}
```

---

## aftrs_ableton_status

Get Ableton Live status including tempo, playing state, and track count

**Complexity:** simple

**Tags:** `ableton`, `live`, `daw`, `status`

**Use Cases:**
- Check Live status
- Get current tempo
- View session info

---

## aftrs_ableton_tempo

Get or set Ableton Live tempo

**Complexity:** simple

**Tags:** `ableton`, `tempo`, `bpm`

**Use Cases:**
- Get current tempo
- Set tempo
- Sync BPM

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm` | number |  | Tempo to set (20-999 BPM). Omit to get current tempo. |

### Example

```json
{
  "bpm": 0
}
```

---

## aftrs_ableton_track

Get or modify a specific track (mute, solo, volume, pan)

**Complexity:** moderate

**Tags:** `ableton`, `track`, `mute`, `solo`, `volume`

**Use Cases:**
- Mute/solo tracks
- Adjust volume
- Arm for recording

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `arm` | boolean |  | Set arm state |
| `index` | integer | Yes | Track index (0-based) |
| `mute` | boolean |  | Set mute state |
| `pan` | number |  | Set pan (-1.0 to 1.0) |
| `solo` | boolean |  | Set solo state |
| `volume` | number |  | Set volume (0.0-1.0) |

### Example

```json
{
  "arm": false,
  "index": 0,
  "mute": false,
  "pan": 0,
  "solo": false,
  "volume": 0
}
```

---

## aftrs_ableton_tracks

List all tracks in Ableton Live session

**Complexity:** simple

**Tags:** `ableton`, `tracks`, `list`

**Use Cases:**
- List all tracks
- View track properties
- Get track names

---

## aftrs_ableton_transport

Control Ableton Live transport (play, stop, record)

**Complexity:** simple

**Tags:** `ableton`, `transport`, `play`, `stop`, `record`

**Use Cases:**
- Start playback
- Stop playback
- Toggle recording

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Transport action to perform |

### Example

```json
{
  "action": "example"
}
```

---

