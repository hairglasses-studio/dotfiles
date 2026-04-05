# homeassistant

> Home Assistant smart home integration

**5 tools**

## Tools

- [`aftrs_hass_automations`](#aftrs-hass-automations)
- [`aftrs_hass_control`](#aftrs-hass-control)
- [`aftrs_hass_entities`](#aftrs-hass-entities)
- [`aftrs_hass_scenes`](#aftrs-hass-scenes)
- [`aftrs_hass_status`](#aftrs-hass-status)

---

## aftrs_hass_automations

List and trigger Home Assistant automations.

**Complexity:** simple

**Tags:** `homeassistant`, `automations`, `trigger`, `workflows`

**Use Cases:**
- List automations
- Trigger automation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `trigger` | string |  | Automation ID to trigger |

### Example

```json
{
  "trigger": "example"
}
```

---

## aftrs_hass_control

Control a Home Assistant entity (turn on/off/toggle).

**Complexity:** simple

**Tags:** `homeassistant`, `control`, `switch`, `light`

**Use Cases:**
- Control lights
- Toggle switches

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: on, off, toggle |
| `brightness` | number |  | Brightness for lights (0-255) |
| `color` | string |  | Color for lights (hex or name) |
| `entity_id` | string | Yes | Entity ID to control |

### Example

```json
{
  "action": "example",
  "brightness": 0,
  "color": "example",
  "entity_id": "example"
}
```

---

## aftrs_hass_entities

List Home Assistant entities.

**Complexity:** simple

**Tags:** `homeassistant`, `entities`, `devices`, `list`

**Use Cases:**
- List smart devices
- Find entity IDs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `domain` | string |  | Filter by domain (light, switch, sensor, etc.) |
| `search` | string |  | Search entities by name |

### Example

```json
{
  "domain": "example",
  "search": "example"
}
```

---

## aftrs_hass_scenes

List and activate Home Assistant scenes.

**Complexity:** simple

**Tags:** `homeassistant`, `scenes`, `presets`, `automation`

**Use Cases:**
- List scenes
- Activate scene

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `activate` | string |  | Scene ID to activate |

### Example

```json
{
  "activate": "example"
}
```

---

## aftrs_hass_status

Get Home Assistant connection status.

**Complexity:** simple

**Tags:** `homeassistant`, `hass`, `status`, `smarthome`

**Use Cases:**
- Check HA connection
- View HA status

---

