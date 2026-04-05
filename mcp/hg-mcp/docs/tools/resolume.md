# resolume

> Resolume Arena/Avenue VJ software control via OSC

**38 tools**

## Tools

- [`aftrs_resolume_audio_mute`](#aftrs-resolume-audio-mute)
- [`aftrs_resolume_audio_pan`](#aftrs-resolume-audio-pan)
- [`aftrs_resolume_audio_tracks`](#aftrs-resolume-audio-tracks)
- [`aftrs_resolume_audio_volume`](#aftrs-resolume-audio-volume)
- [`aftrs_resolume_autopilot`](#aftrs-resolume-autopilot)
- [`aftrs_resolume_bpm`](#aftrs-resolume-bpm)
- [`aftrs_resolume_bypass`](#aftrs-resolume-bypass)
- [`aftrs_resolume_clear`](#aftrs-resolume-clear)
- [`aftrs_resolume_clip_info`](#aftrs-resolume-clip-info)
- [`aftrs_resolume_clip_load`](#aftrs-resolume-clip-load)
- [`aftrs_resolume_clip_properties`](#aftrs-resolume-clip-properties)
- [`aftrs_resolume_clip_speed`](#aftrs-resolume-clip-speed)
- [`aftrs_resolume_clip_thumbnail`](#aftrs-resolume-clip-thumbnail)
- [`aftrs_resolume_clip_transport`](#aftrs-resolume-clip-transport)
- [`aftrs_resolume_clips`](#aftrs-resolume-clips)
- [`aftrs_resolume_columns`](#aftrs-resolume-columns)
- [`aftrs_resolume_crossfade`](#aftrs-resolume-crossfade)
- [`aftrs_resolume_deck`](#aftrs-resolume-deck)
- [`aftrs_resolume_effect_mix`](#aftrs-resolume-effect-mix)
- [`aftrs_resolume_effect_params`](#aftrs-resolume-effect-params)
- [`aftrs_resolume_effect_set`](#aftrs-resolume-effect-set)
- [`aftrs_resolume_effect_toggle`](#aftrs-resolume-effect-toggle)
- [`aftrs_resolume_effects`](#aftrs-resolume-effects)
- [`aftrs_resolume_group_control`](#aftrs-resolume-group-control)
- [`aftrs_resolume_groups`](#aftrs-resolume-groups)
- [`aftrs_resolume_health`](#aftrs-resolume-health)
- [`aftrs_resolume_layer_opacity`](#aftrs-resolume-layer-opacity)
- [`aftrs_resolume_layers`](#aftrs-resolume-layers)
- [`aftrs_resolume_local_clips`](#aftrs-resolume-local-clips)
- [`aftrs_resolume_master`](#aftrs-resolume-master)
- [`aftrs_resolume_output`](#aftrs-resolume-output)
- [`aftrs_resolume_quick_setup`](#aftrs-resolume-quick-setup)
- [`aftrs_resolume_random_trigger`](#aftrs-resolume-random-trigger)
- [`aftrs_resolume_record`](#aftrs-resolume-record)
- [`aftrs_resolume_show_info`](#aftrs-resolume-show-info)
- [`aftrs_resolume_solo`](#aftrs-resolume-solo)
- [`aftrs_resolume_status`](#aftrs-resolume-status)
- [`aftrs_resolume_trigger`](#aftrs-resolume-trigger)

---

## aftrs_resolume_audio_mute

Mute or unmute audio for a layer or master.

**Complexity:** simple

**Tags:** `resolume`, `audio`, `mute`, `silence`

**Use Cases:**
- Mute audio
- Toggle audio

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number |  | Layer number (omit for master) |
| `mute` | boolean | Yes | True to mute, false to unmute |

### Example

```json
{
  "layer": 0,
  "mute": false
}
```

---

## aftrs_resolume_audio_pan

Set audio pan for a layer.

**Complexity:** simple

**Tags:** `resolume`, `audio`, `pan`, `stereo`

**Use Cases:**
- Pan audio
- Stereo positioning

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number | Yes | Layer number (1-based) |
| `pan` | number | Yes | Pan position (-100=left, 0=center, 100=right) |

### Example

```json
{
  "layer": 0,
  "pan": 0
}
```

---

## aftrs_resolume_audio_tracks

List audio tracks/layers with volume and mute status.

**Complexity:** simple

**Tags:** `resolume`, `audio`, `tracks`, `volume`

**Use Cases:**
- View audio status
- Check volumes

---

## aftrs_resolume_audio_volume

Set audio volume for a layer or master.

**Complexity:** simple

**Tags:** `resolume`, `audio`, `volume`, `level`

**Use Cases:**
- Adjust audio levels
- Control volume

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number |  | Layer number (omit for master) |
| `volume` | number | Yes | Volume (0-100) |

### Example

```json
{
  "layer": 0,
  "volume": 0
}
```

---

## aftrs_resolume_autopilot

Control autopilot/random mode.

**Complexity:** moderate

**Tags:** `resolume`, `autopilot`, `random`, `automation`

**Use Cases:**
- Auto-VJ mode
- Random clip triggering

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `enabled` | boolean |  | Enable or disable autopilot |
| `interval` | number |  | Interval in seconds between changes |
| `mode` | string |  | Mode: random, sequential, bpm |

### Example

```json
{
  "enabled": false,
  "interval": 0,
  "mode": "example"
}
```

---

## aftrs_resolume_bpm

Get or set the BPM (beats per minute).

**Complexity:** simple

**Tags:** `resolume`, `bpm`, `tempo`, `sync`

**Use Cases:**
- Adjust tempo
- Sync to music

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm` | number |  | BPM to set (20-999). Omit to just read current BPM. |
| `tap` | boolean |  | If true, sends a tap tempo signal instead of setting BPM. |

### Example

```json
{
  "bpm": 0,
  "tap": false
}
```

---

## aftrs_resolume_bypass

Bypass or enable a layer.

**Complexity:** simple

**Tags:** `resolume`, `layer`, `bypass`, `mute`

**Use Cases:**
- Mute layer output
- Toggle layer visibility

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bypass` | boolean |  | True to bypass, false to enable (default: toggle) |
| `layer` | number | Yes | Layer number (1-based) |

### Example

```json
{
  "bypass": false,
  "layer": 0
}
```

---

## aftrs_resolume_clear

Clear all layers or a specific layer.

**Complexity:** simple

**Tags:** `resolume`, `clear`, `disconnect`, `reset`

**Use Cases:**
- Clear all clips
- Reset composition

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number |  | Layer to clear (omit to clear all) |

### Example

```json
{
  "layer": 0
}
```

---

## aftrs_resolume_clip_info

Get detailed information about a clip including resolution, duration, and properties.

**Complexity:** simple

**Tags:** `resolume`, `clip`, `info`, `details`

**Use Cases:**
- View clip details
- Check clip properties

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `column` | number | Yes | Column number (1-based) |
| `layer` | number | Yes | Layer number (1-based) |

### Example

```json
{
  "column": 0,
  "layer": 0
}
```

---

## aftrs_resolume_clip_load

Load a video file into a clip slot.

**Complexity:** moderate

**Tags:** `resolume`, `clip`, `load`, `import`

**Use Cases:**
- Load new clips
- Add media

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `column` | number | Yes | Column number (1-based) |
| `file` | string | Yes | Path to video file |
| `layer` | number | Yes | Layer number (1-based) |

### Example

```json
{
  "column": 0,
  "file": "example",
  "layer": 0
}
```

---

## aftrs_resolume_clip_properties

Set clip playback properties: trigger style, beat snap, direction.

**Complexity:** moderate

**Tags:** `resolume`, `clip`, `properties`, `playback`

**Use Cases:**
- Configure clip playback
- Set trigger mode

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `beat_snap` | string |  | Beat snap: off, beat, bar, 4bars |
| `column` | number | Yes | Column number (1-based) |
| `direction` | string |  | Playback direction: forward, backward, pingpong |
| `layer` | number | Yes | Layer number (1-based) |
| `trigger_style` | string |  | Trigger style: toggle, gate, retrigger |

### Example

```json
{
  "beat_snap": "example",
  "column": 0,
  "direction": "example",
  "layer": 0,
  "trigger_style": "example"
}
```

---

## aftrs_resolume_clip_speed

Set playback speed for a clip.

**Complexity:** simple

**Tags:** `resolume`, `clip`, `speed`, `playback`

**Use Cases:**
- Slow motion
- Speed up playback

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `column` | number | Yes | Column number (1-based) |
| `layer` | number | Yes | Layer number (1-based) |
| `speed` | number | Yes | Speed multiplier (e.g., 0.5=half, 2.0=double) |

### Example

```json
{
  "column": 0,
  "layer": 0,
  "speed": 0
}
```

---

## aftrs_resolume_clip_thumbnail

Get clip thumbnail as base64 encoded PNG.

**Complexity:** simple

**Tags:** `resolume`, `clip`, `thumbnail`, `preview`

**Use Cases:**
- Preview clip
- Get clip image

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `column` | number | Yes | Column number (1-based) |
| `layer` | number | Yes | Layer number (1-based) |

### Example

```json
{
  "column": 0,
  "layer": 0
}
```

---

## aftrs_resolume_clip_transport

Control clip transport: seek to position.

**Complexity:** simple

**Tags:** `resolume`, `clip`, `transport`, `seek`

**Use Cases:**
- Seek clip position
- Jump to time

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `column` | number | Yes | Column number (1-based) |
| `layer` | number | Yes | Layer number (1-based) |
| `position` | number |  | Seek position (0-100 as percentage) |

### Example

```json
{
  "column": 0,
  "layer": 0,
  "position": 0
}
```

---

## aftrs_resolume_clips

List clips in a specific layer's clip bank.

**Complexity:** simple

**Tags:** `resolume`, `clips`, `bank`, `media`

**Use Cases:**
- Browse clip bank
- Find specific clips

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number | Yes | Layer number (1-based) |

### Example

```json
{
  "layer": 0
}
```

---

## aftrs_resolume_columns

List columns in the composition.

**Complexity:** simple

**Tags:** `resolume`, `columns`, `scenes`, `composition`

**Use Cases:**
- View column layout
- List scenes

---

## aftrs_resolume_crossfade

Crossfade between decks (A/B mixing).

**Complexity:** simple

**Tags:** `resolume`, `crossfade`, `deck`, `mix`

**Use Cases:**
- Mix between decks
- Transition compositions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `position` | number | Yes | Crossfade position (0=Deck A, 100=Deck B) |

### Example

```json
{
  "position": 0
}
```

---

## aftrs_resolume_deck

Get information about decks in the composition.

**Complexity:** simple

**Tags:** `resolume`, `deck`, `composition`

**Use Cases:**
- View deck configuration
- Check composition structure

---

## aftrs_resolume_effect_mix

Set the mix/intensity of an effect.

**Complexity:** simple

**Tags:** `resolume`, `effect`, `mix`, `intensity`

**Use Cases:**
- Adjust effect intensity
- Fade effects

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `effect` | number | Yes | Effect index (1-based) |
| `layer` | number | Yes | Layer number (1-based) |
| `mix` | number | Yes | Mix value (0-100) |

### Example

```json
{
  "effect": 0,
  "layer": 0,
  "mix": 0
}
```

---

## aftrs_resolume_effect_params

Get all parameters for an effect with their current values and ranges.

**Complexity:** simple

**Tags:** `resolume`, `effect`, `params`, `control`

**Use Cases:**
- Explore effect parameters
- Find parameter IDs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `effect` | number | Yes | Effect index (1-based) |
| `layer` | number | Yes | Layer number (1-based) |

### Example

```json
{
  "effect": 0,
  "layer": 0
}
```

---

## aftrs_resolume_effect_set

Set an effect parameter by ID or name.

**Complexity:** moderate

**Tags:** `resolume`, `effect`, `param`, `set`, `control`

**Use Cases:**
- Adjust effect parameters
- Fine-tune effects

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `effect` | number | Yes | Effect index (1-based) |
| `layer` | number | Yes | Layer number (1-based) |
| `param_id` | number |  | Parameter ID (use aftrs_resolume_effect_params to find) |
| `param_name` | string |  | Parameter name (alternative to ID) |
| `value` | number | Yes | Value to set (0-100 for percentages) |

### Example

```json
{
  "effect": 0,
  "layer": 0,
  "param_id": 0,
  "param_name": "example",
  "value": 0
}
```

---

## aftrs_resolume_effect_toggle

Toggle an effect on or off.

**Complexity:** simple

**Tags:** `resolume`, `effect`, `toggle`, `bypass`

**Use Cases:**
- Enable/disable effects
- Control effect chain

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `effect` | number | Yes | Effect index (1-based) |
| `enabled` | boolean |  | True to enable, false to disable |
| `layer` | number |  | Layer number (omit for master) |

### Example

```json
{
  "effect": 0,
  "enabled": false,
  "layer": 0
}
```

---

## aftrs_resolume_effects

List effects on a layer or the master.

**Complexity:** simple

**Tags:** `resolume`, `effects`, `fx`, `processing`

**Use Cases:**
- View effect chains
- Check effect status

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number |  | Layer number (1-based). Omit for master effects. |

### Example

```json
{
  "layer": 0
}
```

---

## aftrs_resolume_group_control

Control a layer group: opacity, bypass, solo.

**Complexity:** simple

**Tags:** `resolume`, `group`, `control`, `opacity`

**Use Cases:**
- Control layer groups
- Group fading

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bypass` | boolean |  | Bypass the group |
| `group` | number | Yes | Group number (1-based) |
| `opacity` | number |  | Opacity (0-100) |
| `solo` | boolean |  | Solo the group |

### Example

```json
{
  "bypass": false,
  "group": 0,
  "opacity": 0,
  "solo": false
}
```

---

## aftrs_resolume_groups

List all layer groups in the composition.

**Complexity:** simple

**Tags:** `resolume`, `groups`, `layers`, `organization`

**Use Cases:**
- View layer groups
- Check group status

---

## aftrs_resolume_health

Get overall Resolume system health and recommendations.

**Complexity:** moderate

**Tags:** `resolume`, `health`, `monitoring`, `status`

**Use Cases:**
- Check system health
- Troubleshoot issues

---

## aftrs_resolume_layer_opacity

Set opacity for a specific layer.

**Complexity:** simple

**Tags:** `resolume`, `layer`, `opacity`, `control`

**Use Cases:**
- Adjust layer visibility
- Fade layers

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number | Yes | Layer number (1-based) |
| `opacity` | number | Yes | Opacity value (0-100) |

### Example

```json
{
  "layer": 0,
  "opacity": 0
}
```

---

## aftrs_resolume_layers

List all layers with opacity and active clip information.

**Complexity:** simple

**Tags:** `resolume`, `layers`, `opacity`, `clips`

**Use Cases:**
- View layer status
- Check active clips

---

## aftrs_resolume_local_clips

List locally synced VJ clips available for loading into Resolume.

**Complexity:** simple

**Tags:** `resolume`, `clips`, `local`, `library`, `vj`

**Use Cases:**
- Browse local VJ clips
- Find clips to load

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `pack` | string |  | Filter by pack name (e.g., 'hackerglasses', 'hairglasses') |

### Example

```json
{
  "pack": "example"
}
```

---

## aftrs_resolume_master

Get or set the master output level.

**Complexity:** simple

**Tags:** `resolume`, `master`, `level`, `output`

**Use Cases:**
- Adjust master output
- Fade to black

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `level` | number |  | Master level (0-100). Omit to read current. |

### Example

```json
{
  "level": 0
}
```

---

## aftrs_resolume_output

Get output routing and display configuration.

**Complexity:** simple

**Tags:** `resolume`, `output`, `display`, `routing`

**Use Cases:**
- Check output config
- View display setup

---

## aftrs_resolume_quick_setup

Quick VJ setup: clear layers, set BPM, and prepare for show.

**Complexity:** moderate

**Tags:** `resolume`, `setup`, `show`, `quick`, `vj`

**Use Cases:**
- Quick show prep
- Reset for new set

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm` | number |  | BPM to set (default: 128) |
| `clear` | boolean |  | Clear all layers first (default: true) |

### Example

```json
{
  "bpm": 0,
  "clear": false
}
```

---

## aftrs_resolume_random_trigger

Trigger a random clip from a layer or all layers.

**Complexity:** simple

**Tags:** `resolume`, `random`, `trigger`, `vj`, `auto`

**Use Cases:**
- Random VJ mode
- Auto-trigger clips

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number |  | Layer to trigger random clip on (omit for random layer) |

### Example

```json
{
  "layer": 0
}
```

---

## aftrs_resolume_record

Control recording in Resolume.

**Complexity:** moderate

**Tags:** `resolume`, `record`, `capture`, `export`

**Use Cases:**
- Record performance
- Capture output

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: 'start', 'stop', or 'status' |

### Example

```json
{
  "action": "example"
}
```

---

## aftrs_resolume_show_info

Get comprehensive show status: connection, layers, clips, sync status, and recommendations.

**Complexity:** moderate

**Tags:** `resolume`, `show`, `status`, `overview`, `vj`

**Use Cases:**
- Pre-show check
- VJ status overview

---

## aftrs_resolume_solo

Solo a specific layer.

**Complexity:** simple

**Tags:** `resolume`, `layer`, `solo`, `isolate`

**Use Cases:**
- Isolate layer for preview
- Focus on single layer

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `layer` | number | Yes | Layer number (1-based) |
| `solo` | boolean |  | True to solo, false to unsolo (default: toggle) |

### Example

```json
{
  "layer": 0,
  "solo": false
}
```

---

## aftrs_resolume_status

Get Resolume Arena/Avenue status including connection state and BPM.

**Complexity:** simple

**Tags:** `resolume`, `status`, `vj`, `connection`

**Use Cases:**
- Check Resolume connection
- View current BPM

---

## aftrs_resolume_trigger

Trigger a clip or column in Resolume.

**Complexity:** moderate

**Tags:** `resolume`, `trigger`, `clip`, `column`

**Use Cases:**
- Trigger clips
- Fire columns

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `column` | number | Yes | Column number (1-based). |
| `layer` | number |  | Layer number (1-based). Required for clip trigger. |

### Example

```json
{
  "column": 0,
  "layer": 0
}
```

---

