# lighting

> DMX/ArtNet lighting control and fixture management

**24 tools**

## Tools

- [`aftrs_artnet_nodes`](#aftrs-artnet-nodes)
- [`aftrs_artnet_status`](#aftrs-artnet-status)
- [`aftrs_chase_control`](#aftrs-chase-control)
- [`aftrs_chase_list`](#aftrs-chase-list)
- [`aftrs_dmx_blackout`](#aftrs-dmx-blackout)
- [`aftrs_dmx_channels`](#aftrs-dmx-channels)
- [`aftrs_dmx_full`](#aftrs-dmx-full)
- [`aftrs_dmx_status`](#aftrs-dmx-status)
- [`aftrs_fixture_color`](#aftrs-fixture-color)
- [`aftrs_fixture_control`](#aftrs-fixture-control)
- [`aftrs_fixture_dimmer`](#aftrs-fixture-dimmer)
- [`aftrs_fixture_list`](#aftrs-fixture-list)
- [`aftrs_group_control`](#aftrs-group-control)
- [`aftrs_group_list`](#aftrs-group-list)
- [`aftrs_lighting_health`](#aftrs-lighting-health)
- [`aftrs_patch_export`](#aftrs-patch-export)
- [`aftrs_patch_fixture`](#aftrs-patch-fixture)
- [`aftrs_patch_list`](#aftrs-patch-list)
- [`aftrs_patch_unpatch`](#aftrs-patch-unpatch)
- [`aftrs_scene_fade`](#aftrs-scene-fade)
- [`aftrs_scene_list`](#aftrs-scene-list)
- [`aftrs_scene_recall`](#aftrs-scene-recall)
- [`aftrs_scene_save`](#aftrs-scene-save)
- [`aftrs_universe_select`](#aftrs-universe-select)

---

## aftrs_artnet_nodes

Discover ArtNet nodes on the network.

**Complexity:** moderate

**Tags:** `artnet`, `nodes`, `discovery`, `network`

**Use Cases:**
- Find ArtNet devices
- Network diagnostics

---

## aftrs_artnet_status

Get detailed ArtNet protocol status.

**Complexity:** simple

**Tags:** `artnet`, `status`, `network`, `protocol`

**Use Cases:**
- Check ArtNet connection
- Network diagnostics

---

## aftrs_chase_control

Control chase playback.

**Complexity:** moderate

**Tags:** `chase`, `sequence`, `control`, `playback`

**Use Cases:**
- Start/stop chases
- Adjust tempo

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start, stop, or tap |
| `bpm` | number |  | BPM for chase speed |
| `chase` | string | Yes | Chase name |

### Example

```json
{
  "action": "example",
  "bpm": 0,
  "chase": "example"
}
```

---

## aftrs_chase_list

List available chases/sequences.

**Complexity:** simple

**Tags:** `chase`, `sequence`, `list`

**Use Cases:**
- View chases
- Check sequences

---

## aftrs_dmx_blackout

Blackout all lights (all channels to zero).

**Complexity:** simple

**Tags:** `dmx`, `blackout`, `all`, `off`

**Use Cases:**
- Emergency blackout
- End of show

---

## aftrs_dmx_channels

Get or set DMX channel values.

**Complexity:** simple

**Tags:** `dmx`, `channels`, `control`

**Use Cases:**
- Read channel values
- Set channel values

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `count` | number |  | Number of channels to read (default: 16) |
| `start_channel` | number |  | Starting channel (1-512, default: 1) |
| `values` | string |  | Comma-separated values to set (e.g., '255,128,0') |

### Example

```json
{
  "count": 0,
  "start_channel": 0,
  "values": "example"
}
```

---

## aftrs_dmx_full

Full up all lights (all channels to max).

**Complexity:** simple

**Tags:** `dmx`, `full`, `all`, `on`

**Use Cases:**
- Work lights
- Testing

---

## aftrs_dmx_status

Get DMX universe status and connection info.

**Complexity:** simple

**Tags:** `dmx`, `artnet`, `status`, `universe`

**Use Cases:**
- Check DMX connection
- Verify universe status

---

## aftrs_fixture_color

Quick color control for a fixture.

**Complexity:** simple

**Tags:** `fixture`, `color`, `rgb`

**Use Cases:**
- Set fixture color
- Quick color changes

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `color` | string | Yes | Color (hex like 'FF0000' or name like 'red') |
| `fixture` | string | Yes | Fixture name |

### Example

```json
{
  "color": "example",
  "fixture": "example"
}
```

---

## aftrs_fixture_control

Control a lighting fixture by name.

**Complexity:** moderate

**Tags:** `fixtures`, `control`, `color`, `dimmer`

**Use Cases:**
- Control fixtures
- Set colors

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `color` | string |  | Color as hex (e.g., 'FF0000' for red) or name ('red', 'blue', etc.) |
| `dimmer` | number |  | Dimmer value (0-255) |
| `fixture` | string | Yes | Fixture name |

### Example

```json
{
  "color": "example",
  "dimmer": 0,
  "fixture": "example"
}
```

---

## aftrs_fixture_dimmer

Quick dimmer control for a fixture.

**Complexity:** simple

**Tags:** `fixture`, `dimmer`, `level`, `brightness`

**Use Cases:**
- Adjust brightness
- Fade fixtures

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `fixture` | string | Yes | Fixture name |
| `level` | number | Yes | Dimmer level (0-100%) |

### Example

```json
{
  "fixture": "example",
  "level": 0
}
```

---

## aftrs_fixture_list

List configured lighting fixtures.

**Complexity:** simple

**Tags:** `fixtures`, `lights`, `list`

**Use Cases:**
- Browse fixtures
- Check fixture config

---

## aftrs_group_control

Control all fixtures in a group.

**Complexity:** moderate

**Tags:** `groups`, `control`, `fixtures`

**Use Cases:**
- Control multiple fixtures
- Group dimming

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `color` | string |  | Color (hex or name) |
| `dimmer` | number |  | Dimmer level (0-100%) |
| `group` | string | Yes | Group name |

### Example

```json
{
  "color": "example",
  "dimmer": 0,
  "group": "example"
}
```

---

## aftrs_group_list

List fixture groups.

**Complexity:** simple

**Tags:** `groups`, `fixtures`, `list`

**Use Cases:**
- View fixture groups
- Check groupings

---

## aftrs_lighting_health

Get lighting system health score.

**Complexity:** moderate

**Tags:** `health`, `status`, `diagnostics`

**Use Cases:**
- Check system health
- Pre-show verification

---

## aftrs_patch_export

Export patch to JSON or CSV format.

**Complexity:** simple

**Tags:** `patch`, `export`, `backup`, `json`, `csv`

**Use Cases:**
- Backup patch sheet
- Export for documentation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `format` | string |  | Export format: json (default) or csv |

### Example

```json
{
  "format": "example"
}
```

---

## aftrs_patch_fixture

Patch a fixture to a DMX address.

**Complexity:** moderate

**Tags:** `patch`, `fixture`, `assign`, `address`

**Use Cases:**
- Add fixture to patch
- Assign DMX address

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `address` | number | Yes | Start DMX address (1-512) |
| `fixture` | string | Yes | Fixture type or name |
| `name` | string |  | Custom fixture name |
| `universe` | number | Yes | DMX universe (0-32767) |

### Example

```json
{
  "address": 0,
  "fixture": "example",
  "name": "example",
  "universe": 0
}
```

---

## aftrs_patch_list

List fixture patch assignments (fixture-to-channel mappings).

**Complexity:** simple

**Tags:** `patch`, `fixtures`, `channels`, `mapping`

**Use Cases:**
- View patch assignments
- Check channel allocation

---

## aftrs_patch_unpatch

Remove a fixture from the patch.

**Complexity:** simple

**Tags:** `patch`, `unpatch`, `remove`, `delete`

**Use Cases:**
- Remove fixture from patch
- Clear address

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `fixture` | string | Yes | Fixture name to unpatch |

### Example

```json
{
  "fixture": "example"
}
```

---

## aftrs_scene_fade

Crossfade to a scene over time.

**Complexity:** moderate

**Tags:** `scene`, `fade`, `crossfade`, `transition`

**Use Cases:**
- Smooth transitions
- Timed scene changes

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `duration` | number |  | Fade duration in seconds (default 2) |
| `scene` | string | Yes | Scene name |

### Example

```json
{
  "duration": 0,
  "scene": "example"
}
```

---

## aftrs_scene_list

List saved lighting scenes.

**Complexity:** simple

**Tags:** `scenes`, `presets`, `list`

**Use Cases:**
- Browse scenes
- Find saved looks

---

## aftrs_scene_recall

Recall a saved lighting scene.

**Complexity:** simple

**Tags:** `scenes`, `recall`, `presets`

**Use Cases:**
- Load saved scene
- Quick lighting change

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `scene` | string | Yes | Scene name to recall |

### Example

```json
{
  "scene": "example"
}
```

---

## aftrs_scene_save

Save current lighting state as a scene.

**Complexity:** simple

**Tags:** `scene`, `save`, `preset`, `store`

**Use Cases:**
- Save current look
- Create presets

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `description` | string |  | Scene description |
| `name` | string | Yes | Scene name |

### Example

```json
{
  "description": "example",
  "name": "example"
}
```

---

## aftrs_universe_select

Select active DMX universe.

**Complexity:** simple

**Tags:** `universe`, `dmx`, `select`

**Use Cases:**
- Switch universes
- Multi-universe control

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `universe` | number | Yes | Universe number (0-32767) |

### Example

```json
{
  "universe": 0
}
```

---

