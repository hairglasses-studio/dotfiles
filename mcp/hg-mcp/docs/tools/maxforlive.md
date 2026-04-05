# maxforlive

> Max for Live device control and parameter automation via OSC

**8 tools**

## Tools

- [`aftrs_m4l_devices`](#aftrs-m4l-devices)
- [`aftrs_m4l_health`](#aftrs-m4l-health)
- [`aftrs_m4l_macro`](#aftrs-m4l-macro)
- [`aftrs_m4l_mappings`](#aftrs-m4l-mappings)
- [`aftrs_m4l_parameter_set`](#aftrs-m4l-parameter-set)
- [`aftrs_m4l_parameters`](#aftrs-m4l-parameters)
- [`aftrs_m4l_send`](#aftrs-m4l-send)
- [`aftrs_m4l_status`](#aftrs-m4l-status)

---

## aftrs_m4l_devices

List all Max for Live devices connected via the bridge

**Complexity:** simple

**Tags:** `m4l`, `devices`, `list`

**Use Cases:**
- List M4L devices
- Find device IDs
- View connected devices

---

## aftrs_m4l_health

Check Max for Live bridge health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `m4l`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Check connection
- Diagnose issues
- Verify bridge device

---

## aftrs_m4l_macro

Trigger, store, or recall a macro/preset on an M4L device

**Complexity:** moderate

**Tags:** `m4l`, `macro`, `preset`, `recall`

**Use Cases:**
- Trigger macro
- Store preset
- Recall settings

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action to perform |
| `device_id` | string | Yes | Device ID |
| `macro_id` | string |  | Macro ID or name (for trigger/store/recall) |

### Example

```json
{
  "action": "example",
  "device_id": "example",
  "macro_id": "example"
}
```

---

## aftrs_m4l_mappings

List parameter mappings between M4L devices

**Complexity:** simple

**Tags:** `m4l`, `mappings`, `modulation`

**Use Cases:**
- View mappings
- Check modulation sources
- List connections

---

## aftrs_m4l_parameter_set

Set a parameter value on a Max for Live device

**Complexity:** moderate

**Tags:** `m4l`, `parameter`, `set`, `automation`

**Use Cases:**
- Set parameter value
- Automate device
- Control effect

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `device_id` | string | Yes | Device ID |
| `normalized` | boolean |  | If true, treat value as 0-1 normalized |
| `param_name` | string | Yes | Parameter name |
| `value` | number | Yes | Value to set |

### Example

```json
{
  "device_id": "example",
  "normalized": false,
  "param_name": "example",
  "value": 0
}
```

---

## aftrs_m4l_parameters

Get parameters for a Max for Live device

**Complexity:** simple

**Tags:** `m4l`, `parameters`, `get`

**Use Cases:**
- View device parameters
- Get parameter values
- List controls

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `device_id` | string | Yes | Device ID to get parameters for |

### Example

```json
{
  "device_id": "example"
}
```

---

## aftrs_m4l_send

Send a custom OSC message to a Max for Live device

**Complexity:** moderate

**Tags:** `m4l`, `osc`, `send`, `custom`

**Use Cases:**
- Send custom message
- Trigger custom action
- Direct OSC control

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `address` | string | Yes | OSC address (e.g., /m4l/device/custom) |
| `args` | array |  | OSC message arguments |

### Example

```json
{
  "address": "example",
  "args": []
}
```

---

## aftrs_m4l_status

Get Max for Live bridge connection status and device count

**Complexity:** simple

**Tags:** `m4l`, `maxforlive`, `ableton`, `status`

**Use Cases:**
- Check M4L connection
- View device count
- Verify bridge

---

