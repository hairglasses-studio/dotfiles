# ledfx

> LedFX audio-reactive LED control via REST API

**16 tools**

## Tools

- [`aftrs_ledfx_audio`](#aftrs-ledfx-audio)
- [`aftrs_ledfx_bpm`](#aftrs-ledfx-bpm)
- [`aftrs_ledfx_devices`](#aftrs-ledfx-devices)
- [`aftrs_ledfx_effect`](#aftrs-ledfx-effect)
- [`aftrs_ledfx_effects`](#aftrs-ledfx-effects)
- [`aftrs_ledfx_find_devices`](#aftrs-ledfx-find-devices)
- [`aftrs_ledfx_gradient`](#aftrs-ledfx-gradient)
- [`aftrs_ledfx_health`](#aftrs-ledfx-health)
- [`aftrs_ledfx_presets`](#aftrs-ledfx-presets)
- [`aftrs_ledfx_scene`](#aftrs-ledfx-scene)
- [`aftrs_ledfx_scenes`](#aftrs-ledfx-scenes)
- [`aftrs_ledfx_segments`](#aftrs-ledfx-segments)
- [`aftrs_ledfx_solid_color`](#aftrs-ledfx-solid-color)
- [`aftrs_ledfx_status`](#aftrs-ledfx-status)
- [`aftrs_ledfx_virtual`](#aftrs-ledfx-virtual)
- [`aftrs_ledfx_virtuals`](#aftrs-ledfx-virtuals)

---

## aftrs_ledfx_audio

List or set audio input devices.

**Complexity:** simple

**Tags:** `ledfx`, `audio`, `input`, `microphone`

**Use Cases:**
- View audio inputs
- Change audio source

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `device_index` | number |  | Device index to set as active (omit to list devices) |

### Example

```json
{
  "device_index": 0
}
```

---

## aftrs_ledfx_bpm

Get or set BPM for audio-reactive effects.

**Complexity:** simple

**Tags:** `ledfx`, `bpm`, `tempo`, `sync`

**Use Cases:**
- Sync LEDs to music tempo
- Manual BPM override

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm` | number |  | BPM to set (omit to get current BPM) |
| `mode` | string |  | Sync mode: manual, auto, tap |

### Example

```json
{
  "bpm": 0,
  "mode": "example"
}
```

---

## aftrs_ledfx_devices

List physical LED devices configured in LedFX.

**Complexity:** simple

**Tags:** `ledfx`, `devices`, `led`, `wled`

**Use Cases:**
- View configured LED devices
- Check device status

---

## aftrs_ledfx_effect

Apply or manage an effect on a virtual device.

**Complexity:** moderate

**Tags:** `ledfx`, `effect`, `apply`, `configure`

**Use Cases:**
- Apply audio-reactive effect
- Configure effect parameters

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `config` | string |  | JSON effect configuration |
| `effect` | string | Yes | Effect type to apply |
| `preset` | string |  | Preset name to apply |
| `virtual_id` | string | Yes | Virtual device ID |

### Example

```json
{
  "config": "example",
  "effect": "example",
  "preset": "example",
  "virtual_id": "example"
}
```

---

## aftrs_ledfx_effects

List available effect types in LedFX.

**Complexity:** simple

**Tags:** `ledfx`, `effects`, `list`, `audio-reactive`

**Use Cases:**
- Browse available effects
- Find effect by category

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | string |  | Filter by category (optional) |

### Example

```json
{
  "category": "example"
}
```

---

## aftrs_ledfx_find_devices

Auto-discover WLED devices on the network.

**Complexity:** moderate

**Tags:** `ledfx`, `discovery`, `wled`, `network`

**Use Cases:**
- Find WLED devices
- Setup new LED strips

---

## aftrs_ledfx_gradient

Create or apply a color gradient to effects.

**Complexity:** moderate

**Tags:** `ledfx`, `gradient`, `colors`, `palette`

**Use Cases:**
- Create custom color gradients
- Apply rainbow effects

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `colors` | string | Yes | Comma-separated hex colors (e.g., 'FF0000,00FF00,0000FF') |
| `name` | string |  | Save gradient with this name |
| `virtual_id` | string | Yes | Virtual device ID |

### Example

```json
{
  "colors": "example",
  "name": "example",
  "virtual_id": "example"
}
```

---

## aftrs_ledfx_health

Get LedFX system health and recommendations.

**Complexity:** moderate

**Tags:** `ledfx`, `health`, `monitoring`, `status`

**Use Cases:**
- Check system health
- Diagnose issues

---

## aftrs_ledfx_presets

List presets for an effect type.

**Complexity:** simple

**Tags:** `ledfx`, `presets`, `effects`, `config`

**Use Cases:**
- Browse effect presets
- Find preset configurations

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `effect` | string | Yes | Effect type to get presets for |

### Example

```json
{
  "effect": "example"
}
```

---

## aftrs_ledfx_scene

Manage LedFX scenes (activate, save, delete).

**Complexity:** moderate

**Tags:** `ledfx`, `scene`, `activate`, `save`

**Use Cases:**
- Switch scenes
- Save current state
- Delete old scenes

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: activate, save, delete |
| `id` | string |  | Scene ID (for activate/delete) |
| `name` | string |  | Scene name (for save action) |

### Example

```json
{
  "action": "example",
  "id": "example",
  "name": "example"
}
```

---

## aftrs_ledfx_scenes

List saved scenes in LedFX.

**Complexity:** simple

**Tags:** `ledfx`, `scenes`, `snapshots`, `presets`

**Use Cases:**
- View saved scenes
- Browse show configurations

---

## aftrs_ledfx_segments

List or configure LED segments for a device.

**Complexity:** moderate

**Tags:** `ledfx`, `segments`, `led`, `zones`

**Use Cases:**
- Split LED strip into zones
- Configure segments

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: list, create, delete |
| `device_id` | string |  | Device ID (omit to list all segments) |
| `end` | number |  | Segment end LED index |
| `start` | number |  | Segment start LED index |

### Example

```json
{
  "action": "example",
  "device_id": "example",
  "end": 0,
  "start": 0
}
```

---

## aftrs_ledfx_solid_color

Set a solid color on a virtual device.

**Complexity:** simple

**Tags:** `ledfx`, `color`, `solid`, `rgb`

**Use Cases:**
- Quick solid color
- Static LED color

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `color` | string | Yes | Color (hex like 'FF0000' or name like 'red') |
| `virtual_id` | string | Yes | Virtual device ID |

### Example

```json
{
  "color": "example",
  "virtual_id": "example"
}
```

---

## aftrs_ledfx_status

Get LedFX connection status, version, and device counts.

**Complexity:** simple

**Tags:** `ledfx`, `led`, `status`, `audio`

**Use Cases:**
- Check LedFX connection
- View system overview

---

## aftrs_ledfx_virtual

Control a virtual LED device (activate, set effect, clear).

**Complexity:** moderate

**Tags:** `ledfx`, `virtual`, `control`, `effect`

**Use Cases:**
- Apply effect to virtual
- Activate/deactivate LEDs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: activate, deactivate, set_effect, clear (default: show info) |
| `config` | string |  | JSON effect config (for set_effect action) |
| `effect` | string |  | Effect type (for set_effect action) |
| `id` | string | Yes | Virtual device ID |

### Example

```json
{
  "action": "example",
  "config": "example",
  "effect": "example",
  "id": "example"
}
```

---

## aftrs_ledfx_virtuals

List virtual LED devices with their active effects.

**Complexity:** simple

**Tags:** `ledfx`, `virtuals`, `effects`, `mapping`

**Use Cases:**
- View virtual devices
- Check active effects

---

