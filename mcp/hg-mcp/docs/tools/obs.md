# obs

> OBS Studio streaming and recording control via WebSocket

**15 tools**

## Tools

- [`aftrs_obs_audio`](#aftrs-obs-audio)
- [`aftrs_obs_health`](#aftrs-obs-health)
- [`aftrs_obs_mute`](#aftrs-obs-mute)
- [`aftrs_obs_record`](#aftrs-obs-record)
- [`aftrs_obs_replay`](#aftrs-obs-replay)
- [`aftrs_obs_scene_switch`](#aftrs-obs-scene-switch)
- [`aftrs_obs_scenes`](#aftrs-obs-scenes)
- [`aftrs_obs_settings`](#aftrs-obs-settings)
- [`aftrs_obs_source_visibility`](#aftrs-obs-source-visibility)
- [`aftrs_obs_sources`](#aftrs-obs-sources)
- [`aftrs_obs_status`](#aftrs-obs-status)
- [`aftrs_obs_stream`](#aftrs-obs-stream)
- [`aftrs_obs_studio_mode`](#aftrs-obs-studio-mode)
- [`aftrs_obs_virtualcam`](#aftrs-obs-virtualcam)
- [`aftrs_obs_volume`](#aftrs-obs-volume)

---

## aftrs_obs_audio

List audio sources with volume and mute status.

**Complexity:** simple

**Tags:** `obs`, `audio`, `volume`, `mixer`

**Use Cases:**
- View audio sources
- Check audio levels

---

## aftrs_obs_health

Get OBS system health and recommendations.

**Complexity:** moderate

**Tags:** `obs`, `health`, `monitoring`, `status`

**Use Cases:**
- Check system health
- Monitor performance

---

## aftrs_obs_mute

Mute or unmute an audio source.

**Complexity:** simple

**Tags:** `obs`, `audio`, `mute`, `toggle`

**Use Cases:**
- Mute microphone
- Toggle audio

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `mute` | boolean |  | True to mute, false to unmute (default: toggle) |
| `source` | string | Yes | Audio source name |

### Example

```json
{
  "mute": false,
  "source": "example"
}
```

---

## aftrs_obs_record

Control recording (start, stop, pause, resume).

**Complexity:** moderate

**Tags:** `obs`, `record`, `capture`, `video`

**Use Cases:**
- Start recording
- Stop recording

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start, stop, pause, resume |

### Example

```json
{
  "action": "example"
}
```

---

## aftrs_obs_replay

Control replay buffer (start, stop, save).

**Complexity:** moderate

**Tags:** `obs`, `replay`, `buffer`, `instant`

**Use Cases:**
- Save instant replay
- Capture highlights

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start, stop, or save |

### Example

```json
{
  "action": "example"
}
```

---

## aftrs_obs_scene_switch

Switch to a different scene.

**Complexity:** simple

**Tags:** `obs`, `scene`, `switch`, `transition`

**Use Cases:**
- Change active scene
- Queue preview scene

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `preview` | boolean |  | Switch preview instead of program (studio mode) |
| `scene` | string | Yes | Scene name to switch to |

### Example

```json
{
  "preview": false,
  "scene": "example"
}
```

---

## aftrs_obs_scenes

List all scenes in OBS.

**Complexity:** simple

**Tags:** `obs`, `scenes`, `list`

**Use Cases:**
- View available scenes
- Check scene order

---

## aftrs_obs_settings

View streaming or recording settings.

**Complexity:** simple

**Tags:** `obs`, `settings`, `config`, `output`

**Use Cases:**
- Check stream settings
- View recording format

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `type` | string |  | Settings type: stream or record (default: both) |

### Example

```json
{
  "type": "example"
}
```

---

## aftrs_obs_source_visibility

Show or hide a source in a scene.

**Complexity:** simple

**Tags:** `obs`, `source`, `visibility`, `toggle`

**Use Cases:**
- Toggle source visibility
- Show/hide overlays

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `scene` | string | Yes | Scene containing the source |
| `source` | string | Yes | Source name |
| `visible` | boolean | Yes | True to show, false to hide |

### Example

```json
{
  "scene": "example",
  "source": "example",
  "visible": false
}
```

---

## aftrs_obs_sources

List sources in a scene or all sources.

**Complexity:** simple

**Tags:** `obs`, `sources`, `list`

**Use Cases:**
- View scene sources
- Check source configuration

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `scene` | string |  | Scene name (omit to list all sources) |

### Example

```json
{
  "scene": "example"
}
```

---

## aftrs_obs_status

Get OBS Studio status including streaming, recording, and performance metrics.

**Complexity:** simple

**Tags:** `obs`, `status`, `streaming`, `recording`

**Use Cases:**
- Check OBS status
- Monitor stream health

---

## aftrs_obs_stream

Control streaming (start, stop, toggle).

**Complexity:** moderate

**Tags:** `obs`, `stream`, `live`, `broadcast`

**Use Cases:**
- Start streaming
- Stop streaming

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start, stop, or toggle |

### Example

```json
{
  "action": "example"
}
```

---

## aftrs_obs_studio_mode

Control studio mode (enable, disable, transition).

**Complexity:** simple

**Tags:** `obs`, `studio`, `preview`, `transition`

**Use Cases:**
- Enable studio mode
- Trigger transition

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: enable, disable, transition, or status |

### Example

```json
{
  "action": "example"
}
```

---

## aftrs_obs_virtualcam

Control virtual camera (start, stop).

**Complexity:** simple

**Tags:** `obs`, `virtualcam`, `camera`, `video`

**Use Cases:**
- Start virtual camera
- Use OBS output in other apps

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start or stop |

### Example

```json
{
  "action": "example"
}
```

---

## aftrs_obs_volume

Set volume for an audio source.

**Complexity:** simple

**Tags:** `obs`, `audio`, `volume`, `level`

**Use Cases:**
- Adjust audio levels
- Set microphone volume

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `source` | string | Yes | Audio source name |
| `volume` | number | Yes | Volume in dB (-100 to 26) |

### Example

```json
{
  "source": "example",
  "volume": 0
}
```

---

