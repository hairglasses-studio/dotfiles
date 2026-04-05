# wled

> WLED LED controller discovery and control

**12 tools**

## Tools

- [`aftrs_wled_artnet_config`](#aftrs-wled-artnet-config)
- [`aftrs_wled_brightness`](#aftrs-wled-brightness)
- [`aftrs_wled_color`](#aftrs-wled-color)
- [`aftrs_wled_discover`](#aftrs-wled-discover)
- [`aftrs_wled_effect`](#aftrs-wled-effect)
- [`aftrs_wled_effects_list`](#aftrs-wled-effects-list)
- [`aftrs_wled_palette`](#aftrs-wled-palette)
- [`aftrs_wled_palettes_list`](#aftrs-wled-palettes-list)
- [`aftrs_wled_power`](#aftrs-wled-power)
- [`aftrs_wled_preset_load`](#aftrs-wled-preset-load)
- [`aftrs_wled_preset_save`](#aftrs-wled-preset-save)
- [`aftrs_wled_status`](#aftrs-wled-status)

---

## aftrs_wled_artnet_config

Configure Art-Net settings for WLED device.

**Complexity:** moderate

**Tags:** `wled`, `artnet`, `dmx`, `config`

**Use Cases:**
- Configure Art-Net input
- Setup DMX control

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `enabled` | boolean | Yes | Enable Art-Net input |
| `ip` | string | Yes | WLED device IP address |
| `start_address` | number |  | DMX start address (default: 1) |
| `universe` | number |  | Art-Net universe (default: 0) |

### Example

```json
{
  "enabled": false,
  "ip": "example",
  "start_address": 0,
  "universe": 0
}
```

---

## aftrs_wled_brightness

Set WLED device brightness.

**Complexity:** simple

**Tags:** `wled`, `brightness`, `dim`, `led`

**Use Cases:**
- Adjust LED brightness
- Dim lights

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `brightness` | number | Yes | Brightness level 0-255 |
| `ip` | string | Yes | WLED device IP address |

### Example

```json
{
  "brightness": 0,
  "ip": "example"
}
```

---

## aftrs_wled_color

Set WLED primary color.

**Complexity:** simple

**Tags:** `wled`, `color`, `rgb`, `led`

**Use Cases:**
- Set LED color
- Change RGB values

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `blue` | number | Yes | Blue value 0-255 |
| `green` | number | Yes | Green value 0-255 |
| `ip` | string | Yes | WLED device IP address |
| `red` | number | Yes | Red value 0-255 |

### Example

```json
{
  "blue": 0,
  "green": 0,
  "ip": "example",
  "red": 0
}
```

---

## aftrs_wled_discover

Discover WLED controllers on the network. Scans subnet for WLED devices.

**Complexity:** moderate

**Tags:** `wled`, `discover`, `scan`, `led`, `network`

**Use Cases:**
- Find WLED devices
- Network LED discovery

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `subnet` | string |  | Subnet to scan (e.g., '192.168.1.0/24'). Default: 192.168.1.0/24 |

### Example

```json
{
  "subnet": "example"
}
```

---

## aftrs_wled_effect

Set WLED effect by ID or name.

**Complexity:** simple

**Tags:** `wled`, `effect`, `animation`, `led`

**Use Cases:**
- Change LED effect
- Set animation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `effect_id` | number |  | Effect ID (0-based) |
| `effect_name` | string |  | Effect name (partial match supported) |
| `ip` | string | Yes | WLED device IP address |

### Example

```json
{
  "effect_id": 0,
  "effect_name": "example",
  "ip": "example"
}
```

---

## aftrs_wled_effects_list

List all available effects on a WLED device.

**Complexity:** simple

**Tags:** `wled`, `effects`, `list`, `animations`

**Use Cases:**
- Browse available effects
- Find effect IDs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ip` | string | Yes | WLED device IP address |

### Example

```json
{
  "ip": "example"
}
```

---

## aftrs_wled_palette

Set WLED color palette.

**Complexity:** simple

**Tags:** `wled`, `palette`, `colors`, `led`

**Use Cases:**
- Change color palette
- Set color scheme

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ip` | string | Yes | WLED device IP address |
| `palette_id` | number | Yes | Palette ID |

### Example

```json
{
  "ip": "example",
  "palette_id": 0
}
```

---

## aftrs_wled_palettes_list

List all available color palettes on a WLED device.

**Complexity:** simple

**Tags:** `wled`, `palettes`, `list`, `colors`

**Use Cases:**
- Browse color palettes
- Find palette IDs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ip` | string | Yes | WLED device IP address |

### Example

```json
{
  "ip": "example"
}
```

---

## aftrs_wled_power

Turn WLED device on or off.

**Complexity:** simple

**Tags:** `wled`, `power`, `on`, `off`, `led`

**Use Cases:**
- Turn LEDs on/off
- Power control

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ip` | string | Yes | WLED device IP address |
| `on` | boolean | Yes | true to turn on, false to turn off |

### Example

```json
{
  "ip": "example",
  "on": false
}
```

---

## aftrs_wled_preset_load

Load a saved WLED preset.

**Complexity:** simple

**Tags:** `wled`, `preset`, `load`, `recall`

**Use Cases:**
- Recall LED preset
- Load configuration

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ip` | string | Yes | WLED device IP address |
| `preset_id` | number | Yes | Preset slot ID to load |

### Example

```json
{
  "ip": "example",
  "preset_id": 0
}
```

---

## aftrs_wled_preset_save

Save current WLED state as a preset.

**Complexity:** simple

**Tags:** `wled`, `preset`, `save`, `store`

**Use Cases:**
- Save LED configuration
- Create preset

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ip` | string | Yes | WLED device IP address |
| `name` | string | Yes | Preset name |
| `preset_id` | number | Yes | Preset slot ID (1-250) |

### Example

```json
{
  "ip": "example",
  "name": "example",
  "preset_id": 0
}
```

---

## aftrs_wled_status

Get status of a WLED device including brightness, effect, and segments.

**Complexity:** simple

**Tags:** `wled`, `status`, `led`, `info`

**Use Cases:**
- Check LED status
- View current effect

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ip` | string | Yes | WLED device IP address |

### Example

```json
{
  "ip": "example"
}
```

---

