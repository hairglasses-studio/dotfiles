# paramsync

> Cross-system parameter mapping and modulation

**4 tools**

## Tools

- [`aftrs_param_map`](#aftrs-param-map)
- [`aftrs_param_maps`](#aftrs-param-maps)
- [`aftrs_param_push`](#aftrs-param-push)
- [`aftrs_param_sync`](#aftrs-param-sync)

---

## aftrs_param_map

Create a parameter mapping between two systems. Maps source parameter to target with optional value transform.

**Complexity:** moderate

**Tags:** `parameter`, `mapping`, `modulation`, `sync`

**Use Cases:**
- Map Ableton device to Resolume effect
- Control lighting from audio

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `curve` | string |  | Transform curve: linear, exponential, logarithmic |
| `input_max` | number |  | Input maximum value (default: 1) |
| `input_min` | number |  | Input minimum value (default: 0) |
| `invert` | boolean |  | Invert the output value |
| `name` | string | Yes | Name for this mapping |
| `output_max` | number |  | Output maximum value (default: 1) |
| `output_min` | number |  | Output minimum value (default: 0) |
| `source_path` | string | Yes | Source param path (e.g., 'track/0/device/1/param/3', 'layer/1/opacity', 'master/level') |
| `source_system` | string | Yes | Source system: ableton, resolume, grandma3, touchdesigner |
| `target_path` | string | Yes | Target param path |
| `target_system` | string | Yes | Target system: ableton, resolume, grandma3, touchdesigner |

### Example

```json
{
  "curve": "example",
  "input_max": 0,
  "input_min": 0,
  "invert": false,
  "name": "example",
  "output_max": 0,
  "output_min": 0,
  "source_path": "example",
  "source_system": "example",
  "target_path": "example",
  "target_system": "example"
}
```

---

## aftrs_param_maps

List all parameter mappings.

**Complexity:** simple

**Tags:** `parameter`, `mapping`, `list`

**Use Cases:**
- View active mappings
- Check mapping configuration

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `mapping_id` | string |  | Get details of a specific mapping |

### Example

```json
{
  "mapping_id": "example"
}
```

---

## aftrs_param_push

Manually push a value through a parameter mapping.

**Complexity:** simple

**Tags:** `parameter`, `push`, `value`

**Use Cases:**
- Test mapping
- Manual parameter control

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `mapping_id` | string | Yes | Mapping ID to push through |
| `value` | number | Yes | Value to push (will be transformed) |

### Example

```json
{
  "mapping_id": "example",
  "value": 0
}
```

---

## aftrs_param_sync

Sync a mapping by reading source and pushing to target.

**Complexity:** simple

**Tags:** `parameter`, `sync`, `read`, `write`

**Use Cases:**
- Sync parameter from source to target
- One-shot sync

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `mapping_id` | string | Yes | Mapping ID to sync |

### Example

```json
{
  "mapping_id": "example"
}
```

---

