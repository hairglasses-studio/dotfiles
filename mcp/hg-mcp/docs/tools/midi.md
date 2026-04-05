# midi

> MIDI control for AV equipment and software

**17 tools**

## Tools

- [`aftrs_midi_cc`](#aftrs-midi-cc)
- [`aftrs_midi_devices`](#aftrs-midi-devices)
- [`aftrs_midi_health`](#aftrs-midi-health)
- [`aftrs_midi_learn`](#aftrs-midi-learn)
- [`aftrs_midi_map_create`](#aftrs-midi-map-create)
- [`aftrs_midi_map_delete`](#aftrs-midi-map-delete)
- [`aftrs_midi_map_list`](#aftrs-midi-map-list)
- [`aftrs_midi_mappings`](#aftrs-midi-mappings)
- [`aftrs_midi_note`](#aftrs-midi-note)
- [`aftrs_midi_panic`](#aftrs-midi-panic)
- [`aftrs_midi_pitch`](#aftrs-midi-pitch)
- [`aftrs_midi_profile_load`](#aftrs-midi-profile-load)
- [`aftrs_midi_profile_save`](#aftrs-midi-profile-save)
- [`aftrs_midi_profiles`](#aftrs-midi-profiles)
- [`aftrs_midi_program`](#aftrs-midi-program)
- [`aftrs_midi_status`](#aftrs-midi-status)
- [`aftrs_midi_transport`](#aftrs-midi-transport)

---

## aftrs_midi_cc

Send a MIDI Control Change message.

**Complexity:** simple

**Tags:** `midi`, `cc`, `control`, `fader`

**Use Cases:**
- Control parameters
- Send fader values

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel` | number | Yes | MIDI channel (1-16) |
| `controller` | number | Yes | Controller number (0-127) |
| `device` | string |  | Output device name (uses default if omitted) |
| `value` | number | Yes | Value (0-127) |

### Example

```json
{
  "channel": 0,
  "controller": 0,
  "device": "example",
  "value": 0
}
```

---

## aftrs_midi_devices

List all MIDI input and output devices.

**Complexity:** simple

**Tags:** `midi`, `devices`, `list`

**Use Cases:**
- View connected MIDI devices
- Find device names

---

## aftrs_midi_health

Get MIDI system health and recommendations.

**Complexity:** simple

**Tags:** `midi`, `health`, `monitoring`

**Use Cases:**
- Check system health
- Diagnose issues

---

## aftrs_midi_learn

Enter MIDI learn mode to capture incoming messages.

**Complexity:** moderate

**Tags:** `midi`, `learn`, `capture`

**Use Cases:**
- Capture MIDI input
- Create mappings

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `timeout` | number |  | Timeout in seconds (default 10) |

### Example

```json
{
  "timeout": 0
}
```

---

## aftrs_midi_map_create

Create a MIDI→tool mapping. Maps MIDI CC/note/program to invoke a tool with optional value transformation.

**Complexity:** moderate

**Tags:** `midi`, `mapping`, `tool`, `automation`

**Use Cases:**
- Map faders to tool parameters
- Create MIDI automation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel` | number | Yes | MIDI channel (1-16) |
| `invert` | boolean |  | Invert the value mapping |
| `name` | string | Yes | Human-readable name for this mapping |
| `number` | number | Yes | CC number, note number, or program number (0-127) |
| `output_max` | number |  | Output range maximum (default 127) |
| `output_min` | number |  | Output range minimum (default 0) |
| `parameters` | object |  | Static parameters to pass to the tool |
| `target_param` | string |  | Parameter name to receive the mapped MIDI value |
| `tool_name` | string | Yes | Target tool to invoke (e.g., aftrs_resolume_bpm) |
| `type` | string | Yes | Message type: cc, note, or program |

### Example

```json
{
  "channel": 0,
  "invert": false,
  "name": "example",
  "number": 0,
  "output_max": 0,
  "output_min": 0,
  "parameters": {},
  "target_param": "example",
  "tool_name": "example",
  "type": "example"
}
```

---

## aftrs_midi_map_delete

Delete a MIDI→tool mapping by ID.

**Complexity:** simple

**Tags:** `midi`, `mapping`, `delete`

**Use Cases:**
- Remove mappings
- Clean up configuration

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `id` | string | Yes | Mapping ID to delete |

### Example

```json
{
  "id": "example"
}
```

---

## aftrs_midi_map_list

List all MIDI→tool mappings with their configurations.

**Complexity:** simple

**Tags:** `midi`, `mapping`, `list`

**Use Cases:**
- View current mappings
- Check configuration

---

## aftrs_midi_mappings

List configured MIDI control mappings.

**Complexity:** simple

**Tags:** `midi`, `mapping`, `config`

**Use Cases:**
- View control mappings
- Check configuration

---

## aftrs_midi_note

Send a MIDI note on/off message.

**Complexity:** simple

**Tags:** `midi`, `note`, `trigger`

**Use Cases:**
- Trigger MIDI notes
- Control instruments

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel` | number | Yes | MIDI channel (1-16) |
| `device` | string |  | Output device name (uses default if omitted) |
| `note` | number | Yes | Note number (0-127, C4=60) |
| `off` | boolean |  | Send note off instead of note on |
| `velocity` | number |  | Velocity (0-127, default 100) |

### Example

```json
{
  "channel": 0,
  "device": "example",
  "note": 0,
  "off": false,
  "velocity": 0
}
```

---

## aftrs_midi_panic

Send MIDI panic (all notes off) to stop stuck notes.

**Complexity:** simple

**Tags:** `midi`, `panic`, `stop`, `reset`

**Use Cases:**
- Stop stuck notes
- Emergency reset

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `device` | string |  | Output device name (uses default if omitted) |

### Example

```json
{
  "device": "example"
}
```

---

## aftrs_midi_pitch

Send a MIDI Pitch Bend message.

**Complexity:** simple

**Tags:** `midi`, `pitch`, `bend`

**Use Cases:**
- Pitch bend control
- Modulation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel` | number | Yes | MIDI channel (1-16) |
| `device` | string |  | Output device name (uses default if omitted) |
| `value` | number | Yes | Pitch bend value (-8192 to 8191, 0=center) |

### Example

```json
{
  "channel": 0,
  "device": "example",
  "value": 0
}
```

---

## aftrs_midi_profile_load

Load MIDI mappings from a saved profile, replacing current mappings.

**Complexity:** simple

**Tags:** `midi`, `profile`, `load`, `recall`

**Use Cases:**
- Switch between configurations
- Recall presets

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Profile name to load |

### Example

```json
{
  "name": "example"
}
```

---

## aftrs_midi_profile_save

Save current MIDI mappings to a named profile for later recall.

**Complexity:** simple

**Tags:** `midi`, `profile`, `save`, `backup`

**Use Cases:**
- Save mapping configurations
- Create presets

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `description` | string |  | Profile description |
| `name` | string | Yes | Profile name |

### Example

```json
{
  "description": "example",
  "name": "example"
}
```

---

## aftrs_midi_profiles

List all saved MIDI mapping profiles.

**Complexity:** simple

**Tags:** `midi`, `profile`, `list`

**Use Cases:**
- View saved profiles
- Find presets

---

## aftrs_midi_program

Send a MIDI Program Change message.

**Complexity:** simple

**Tags:** `midi`, `program`, `preset`, `patch`

**Use Cases:**
- Change presets
- Select patches

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `channel` | number | Yes | MIDI channel (1-16) |
| `device` | string |  | Output device name (uses default if omitted) |
| `program` | number | Yes | Program number (0-127) |

### Example

```json
{
  "channel": 0,
  "device": "example",
  "program": 0
}
```

---

## aftrs_midi_status

Get MIDI system status including connected devices.

**Complexity:** simple

**Tags:** `midi`, `status`, `devices`

**Use Cases:**
- Check MIDI devices
- View MIDI configuration

---

## aftrs_midi_transport

Send MIDI transport control (start, stop, continue).

**Complexity:** simple

**Tags:** `midi`, `transport`, `start`, `stop`

**Use Cases:**
- Control playback
- Sync devices

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start, stop, or continue |
| `device` | string |  | Output device name (uses default if omitted) |

### Example

```json
{
  "action": "example",
  "device": "example"
}
```

---

