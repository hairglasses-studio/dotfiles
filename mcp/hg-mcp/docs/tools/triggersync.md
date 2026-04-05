# triggersync

> Cross-system scene, clip, and cue triggering for Ableton, Resolume, OBS, and grandMA3

**6 tools**

## Tools

- [`aftrs_trigger_health`](#aftrs-trigger-health)
- [`aftrs_trigger_link`](#aftrs-trigger-link)
- [`aftrs_trigger_map`](#aftrs-trigger-map)
- [`aftrs_trigger_scene`](#aftrs-trigger-scene)
- [`aftrs_trigger_status`](#aftrs-trigger-status)
- [`aftrs_trigger_test`](#aftrs-trigger-test)

---

## aftrs_trigger_health

Check trigger sync health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `trigger`, `health`, `diagnostics`

**Use Cases:**
- Check trigger health
- Diagnose issues
- Verify connectivity

---

## aftrs_trigger_link

Create a trigger mapping that links actions across systems

**Complexity:** moderate

**Tags:** `trigger`, `link`, `mapping`, `create`

**Use Cases:**
- Link Ableton scene to Resolume column
- Create cue sync mapping
- Define trigger chain

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `description` | string |  | Description of what this mapping does |
| `name` | string | Yes | Name for the mapping |
| `triggers` | array | Yes | List of trigger targets |

### Example

```json
{
  "description": "example",
  "name": "example",
  "triggers": []
}
```

---

## aftrs_trigger_map

View or configure cross-system trigger mappings

**Complexity:** simple

**Tags:** `trigger`, `mapping`, `config`

**Use Cases:**
- List all mappings
- View mapping details
- Delete a mapping

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string |  | Action: list, get, delete |
| `mapping_id` | string |  | Mapping ID (for get/delete actions) |

### Example

```json
{
  "action": "example",
  "mapping_id": "example"
}
```

---

## aftrs_trigger_scene

Fire a scene across all connected systems (Ableton scene, Resolume column, OBS scene, grandMA3 cue)

**Complexity:** moderate

**Tags:** `trigger`, `scene`, `sync`, `cross-system`

**Use Cases:**
- Fire scene 3 across all systems
- Sync scene change
- Unified cue triggering

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dry_run` | boolean |  | If true, test without actually triggering |
| `scene_index` | integer | Yes | Scene index (0-based) |

### Example

```json
{
  "dry_run": false,
  "scene_index": 0
}
```

---

## aftrs_trigger_status

Get trigger sync status and connected systems

**Complexity:** simple

**Tags:** `trigger`, `status`, `systems`

**Use Cases:**
- Check connected systems
- View trigger capabilities
- Verify setup

---

## aftrs_trigger_test

Test a trigger chain without actually executing it

**Complexity:** simple

**Tags:** `trigger`, `test`, `dry-run`

**Use Cases:**
- Test trigger chain
- Validate mapping
- Check connectivity

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `mapping_id` | string |  | Mapping ID to test |
| `scene_index` | integer |  | Alternatively, test unified scene trigger |

### Example

```json
{
  "mapping_id": "example",
  "scene_index": 0
}
```

---

