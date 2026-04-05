# atem

> Blackmagic ATEM video switcher control for live production

**8 tools**

## Tools

- [`aftrs_atem_auto`](#aftrs-atem-auto)
- [`aftrs_atem_cut`](#aftrs-atem-cut)
- [`aftrs_atem_health`](#aftrs-atem-health)
- [`aftrs_atem_inputs`](#aftrs-atem-inputs)
- [`aftrs_atem_preview`](#aftrs-atem-preview)
- [`aftrs_atem_program`](#aftrs-atem-program)
- [`aftrs_atem_status`](#aftrs-atem-status)
- [`aftrs_atem_transition`](#aftrs-atem-transition)

---

## aftrs_atem_auto

Execute an auto transition from preview to program using current transition settings

**Complexity:** moderate

**Tags:** `atem`, `auto`, `transition`, `mix`, `dissolve`

**Use Cases:**
- Smooth transition
- Auto mix
- Dissolve to preview

---

## aftrs_atem_cut

Execute an instant cut transition from preview to program

**Complexity:** moderate

**Tags:** `atem`, `cut`, `transition`, `instant`

**Use Cases:**
- Instant switch
- Hard cut
- Quick transition

---

## aftrs_atem_health

Check ATEM switcher connection health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `atem`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Diagnose connection issues
- Check switcher health
- Get troubleshooting tips

---

## aftrs_atem_inputs

List all available video inputs on the ATEM switcher with names and types

**Complexity:** simple

**Tags:** `atem`, `inputs`, `video`, `sources`

**Use Cases:**
- List available sources
- Find input IDs
- Check input availability

---

## aftrs_atem_preview

Set the preview output to a specific input source

**Complexity:** moderate

**Tags:** `atem`, `preview`, `prepare`

**Use Cases:**
- Prepare next shot
- Set preview source
- Stage next transition

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input_id` | integer | Yes | Input ID to set as preview output |

### Example

```json
{
  "input_id": 0
}
```

---

## aftrs_atem_program

Set the program (live) output to a specific input source

**Complexity:** moderate

**Tags:** `atem`, `program`, `live`, `output`

**Use Cases:**
- Switch live output
- Change program source
- Direct cut to input

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input_id` | integer | Yes | Input ID to set as program output (1=Input1, 2=Input2, etc.) |

### Example

```json
{
  "input_id": 0
}
```

---

## aftrs_atem_status

Get ATEM switcher status including model, connection state, and current program/preview inputs

**Complexity:** simple

**Tags:** `atem`, `switcher`, `video`, `broadcast`, `status`

**Use Cases:**
- Check ATEM connection
- Monitor switcher state
- Verify program/preview sources

---

## aftrs_atem_transition

Configure transition type and rate for auto transitions

**Complexity:** moderate

**Tags:** `atem`, `transition`, `settings`, `mix`, `wipe`

**Use Cases:**
- Set transition type
- Configure mix duration
- Setup wipe effect

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `rate` | integer |  | Transition duration in frames (e.g., 30 for 1 second at 30fps) |
| `style` | string | Yes | Transition style: mix, dip, wipe, dve, or sting |

### Example

```json
{
  "rate": 0,
  "style": "example"
}
```

---

