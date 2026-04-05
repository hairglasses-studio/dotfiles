# grandma3

> grandMA3 lighting console control via OSC

**12 tools**

## Tools

- [`aftrs_gma3_blackout`](#aftrs-gma3-blackout)
- [`aftrs_gma3_clear`](#aftrs-gma3-clear)
- [`aftrs_gma3_command`](#aftrs-gma3-command)
- [`aftrs_gma3_executor`](#aftrs-gma3-executor)
- [`aftrs_gma3_fixture`](#aftrs-gma3-fixture)
- [`aftrs_gma3_health`](#aftrs-gma3-health)
- [`aftrs_gma3_macro`](#aftrs-gma3-macro)
- [`aftrs_gma3_master`](#aftrs-gma3-master)
- [`aftrs_gma3_preset`](#aftrs-gma3-preset)
- [`aftrs_gma3_sequence`](#aftrs-gma3-sequence)
- [`aftrs_gma3_status`](#aftrs-gma3-status)
- [`aftrs_gma3_timecode`](#aftrs-gma3-timecode)

---

## aftrs_gma3_blackout

Toggle blackout on/off.

**Complexity:** simple

**Tags:** `grandma3`, `blackout`, `safety`

**Use Cases:**
- Emergency blackout
- Show start/end

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `on` | boolean |  | True for blackout, false to release (default: toggle) |

### Example

```json
{
  "on": false
}
```

---

## aftrs_gma3_clear

Clear programmer or all playback.

**Complexity:** simple

**Tags:** `grandma3`, `clear`, `programmer`

**Use Cases:**
- Clear programmer
- Reset all playback

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `scope` | string |  | Scope: programmer (default), all |

### Example

```json
{
  "scope": "example"
}
```

---

## aftrs_gma3_command

Send a command line instruction to grandMA3.

**Complexity:** moderate

**Tags:** `grandma3`, `command`, `raw`

**Use Cases:**
- Execute any grandMA3 command
- Advanced control

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `command` | string | Yes | Command to execute (e.g., 'Go+ Sequence 1', 'Fixture 1 At 50') |

### Example

```json
{
  "command": "example"
}
```

---

## aftrs_gma3_executor

Control an executor (trigger, stop, set fader).

**Complexity:** simple

**Tags:** `grandma3`, `executor`, `playback`, `fader`

**Use Cases:**
- Trigger cue
- Adjust fader
- Flash effect

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: go, stop, flash_on, flash_off, fader |
| `executor` | number | Yes | Executor number |
| `page` | number |  | Executor page (default: 1) |
| `value` | number |  | Fader value 0-100 (for fader action) |

### Example

```json
{
  "action": "example",
  "executor": 0,
  "page": 0,
  "value": 0
}
```

---

## aftrs_gma3_fixture

Control fixtures (select, set dimmer, set attribute).

**Complexity:** moderate

**Tags:** `grandma3`, `fixture`, `dimmer`, `attribute`

**Use Cases:**
- Select fixtures
- Set dimmer levels
- Adjust colors

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: select, dimmer, attribute, clear |
| `attribute` | string |  | Attribute name (for attribute action) |
| `fixtures` | string |  | Fixture selection (e.g., '1 Thru 10', '1+2+3') |
| `value` | number |  | Value 0-100 (for dimmer/attribute) |

### Example

```json
{
  "action": "example",
  "attribute": "example",
  "fixtures": "example",
  "value": 0
}
```

---

## aftrs_gma3_health

Get grandMA3 system health and recommendations.

**Complexity:** moderate

**Tags:** `grandma3`, `health`, `monitoring`, `status`

**Use Cases:**
- Check system health
- Diagnose issues

---

## aftrs_gma3_macro

Trigger a macro.

**Complexity:** simple

**Tags:** `grandma3`, `macro`, `automation`

**Use Cases:**
- Run automated sequences
- Trigger complex actions

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `macro` | number | Yes | Macro number |

### Example

```json
{
  "macro": 0
}
```

---

## aftrs_gma3_master

Control grand master or speed masters.

**Complexity:** simple

**Tags:** `grandma3`, `master`, `dimmer`, `speed`, `bpm`

**Use Cases:**
- Adjust grand master
- Set BPM
- Tap tempo

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bpm` | number |  | BPM value (for speed master) |
| `number` | number |  | Speed master number (for speed type) |
| `tap` | boolean |  | Tap tempo (for speed master) |
| `type` | string | Yes | Master type: grand, speed |
| `value` | number |  | Value 0-100 (percentage) |

### Example

```json
{
  "bpm": 0,
  "number": 0,
  "tap": false,
  "type": "example",
  "value": 0
}
```

---

## aftrs_gma3_preset

Store or recall presets.

**Complexity:** moderate

**Tags:** `grandma3`, `preset`, `store`, `recall`

**Use Cases:**
- Save looks
- Recall presets

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: store, call |
| `number` | number | Yes | Preset number |
| `type` | string | Yes | Preset type (e.g., Dimmer, Color, Position, Gobo, All) |

### Example

```json
{
  "action": "example",
  "number": 0,
  "type": "example"
}
```

---

## aftrs_gma3_sequence

Control a sequence (go, stop, pause, go to cue).

**Complexity:** simple

**Tags:** `grandma3`, `sequence`, `cue`, `playback`

**Use Cases:**
- Advance cue
- Jump to specific cue
- Stop sequence

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: go, go_next, go_prev, stop, pause, goto_cue |
| `cue` | number |  | Cue number (for goto_cue action) |
| `sequence` | number | Yes | Sequence number |

### Example

```json
{
  "action": "example",
  "cue": 0,
  "sequence": 0
}
```

---

## aftrs_gma3_status

Get grandMA3 console connection status.

**Complexity:** simple

**Tags:** `grandma3`, `lighting`, `status`, `osc`

**Use Cases:**
- Check console connection
- Verify OSC settings

---

## aftrs_gma3_timecode

Control timecode playback.

**Complexity:** moderate

**Tags:** `grandma3`, `timecode`, `sync`, `playback`

**Use Cases:**
- Start timecode show
- Jump to position

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: start, stop, goto |
| `slot` | number |  | Timecode slot number (default: 1) |
| `time` | string |  | Timecode position HH:MM:SS.FF (for goto action) |

### Example

```json
{
  "action": "example",
  "slot": 0,
  "time": "example"
}
```

---

