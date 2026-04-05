# ptz

> ONVIF PTZ camera control for pan, tilt, zoom, and preset management

**15 tools**

## Tools

- [`aftrs_ptz_camera_add`](#aftrs-ptz-camera-add)
- [`aftrs_ptz_camera_control`](#aftrs-ptz-camera-control)
- [`aftrs_ptz_cameras`](#aftrs-ptz-cameras)
- [`aftrs_ptz_goto_preset`](#aftrs-ptz-goto-preset)
- [`aftrs_ptz_health`](#aftrs-ptz-health)
- [`aftrs_ptz_home`](#aftrs-ptz-home)
- [`aftrs_ptz_move`](#aftrs-ptz-move)
- [`aftrs_ptz_presets`](#aftrs-ptz-presets)
- [`aftrs_ptz_status`](#aftrs-ptz-status)
- [`aftrs_ptz_stop`](#aftrs-ptz-stop)
- [`aftrs_ptz_tour_create`](#aftrs-ptz-tour-create)
- [`aftrs_ptz_tour_start`](#aftrs-ptz-tour-start)
- [`aftrs_ptz_tour_status`](#aftrs-ptz-tour-status)
- [`aftrs_ptz_tour_stop`](#aftrs-ptz-tour-stop)
- [`aftrs_ptz_tours`](#aftrs-ptz-tours)

---

## aftrs_ptz_camera_add

Add a new PTZ camera to the system

**Complexity:** moderate

**Tags:** `ptz`, `camera`, `add`, `configure`

**Use Cases:**
- Add new camera
- Configure camera
- Expand camera fleet

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `host` | string | Yes | Camera IP address or hostname |
| `id` | string | Yes | Unique camera identifier |
| `name` | string |  | Human-readable camera name |
| `password` | string |  | Camera password |
| `port` | string |  | ONVIF port (default: 80) |
| `username` | string |  | Camera username (default: admin) |

### Example

```json
{
  "host": "example",
  "id": "example",
  "name": "example",
  "password": "example",
  "port": "example",
  "username": "example"
}
```

---

## aftrs_ptz_camera_control

Control a specific camera by ID (pan, tilt, zoom, preset)

**Complexity:** moderate

**Tags:** `ptz`, `camera`, `control`, `multicam`

**Use Cases:**
- Control specific camera
- Multi-camera operation
- Remote camera control

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: move, stop, preset, home |
| `camera_id` | string | Yes | Camera identifier to control |
| `pan` | number |  | Pan value (-1.0 to 1.0) for move action |
| `preset` | string |  | Preset token for preset action |
| `tilt` | number |  | Tilt value (-1.0 to 1.0) for move action |
| `zoom` | number |  | Zoom value (-1.0 to 1.0) for move action |

### Example

```json
{
  "action": "example",
  "camera_id": "example",
  "pan": 0,
  "preset": "example",
  "tilt": 0,
  "zoom": 0
}
```

---

## aftrs_ptz_cameras

List all configured PTZ cameras with connection status

**Complexity:** simple

**Tags:** `ptz`, `cameras`, `list`, `multicam`

**Use Cases:**
- List all cameras
- Check camera status
- View camera fleet

---

## aftrs_ptz_goto_preset

Move camera to a saved preset position

**Complexity:** moderate

**Tags:** `ptz`, `preset`, `goto`, `recall`

**Use Cases:**
- Recall saved position
- Move to preset shot
- Quick camera positioning

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `preset` | string | Yes | Preset token or name to move to |

### Example

```json
{
  "preset": "example"
}
```

---

## aftrs_ptz_health

Check PTZ camera connection health and get troubleshooting recommendations

**Complexity:** simple

**Tags:** `ptz`, `health`, `diagnostics`, `troubleshooting`

**Use Cases:**
- Diagnose connection issues
- Check camera health
- Get troubleshooting tips

---

## aftrs_ptz_home

Move camera to home position

**Complexity:** simple

**Tags:** `ptz`, `home`, `reset`

**Use Cases:**
- Return to home
- Reset camera position
- Center camera

---

## aftrs_ptz_move

Move camera with pan, tilt, and zoom controls. Values range from -1.0 to 1.0

**Complexity:** moderate

**Tags:** `ptz`, `move`, `pan`, `tilt`, `zoom`

**Use Cases:**
- Pan camera left/right
- Tilt camera up/down
- Zoom in/out

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `mode` | string |  | Movement mode: continuous (keep moving), relative (move by amount), absolute (move to position) |
| `pan` | number |  | Pan speed/position (-1.0=left, 0=stop, 1.0=right) |
| `tilt` | number |  | Tilt speed/position (-1.0=down, 0=stop, 1.0=up) |
| `zoom` | number |  | Zoom speed/position (-1.0=out, 0=stop, 1.0=in) |

### Example

```json
{
  "mode": "example",
  "pan": 0,
  "tilt": 0,
  "zoom": 0
}
```

---

## aftrs_ptz_presets

List all saved PTZ presets

**Complexity:** simple

**Tags:** `ptz`, `presets`, `list`

**Use Cases:**
- List saved positions
- Find preset tokens
- View available shots

---

## aftrs_ptz_status

Get PTZ camera status including connection state and current position

**Complexity:** simple

**Tags:** `ptz`, `camera`, `onvif`, `status`

**Use Cases:**
- Check camera connection
- Get current position
- Verify PTZ availability

---

## aftrs_ptz_stop

Stop all PTZ movement immediately

**Complexity:** simple

**Tags:** `ptz`, `stop`, `halt`

**Use Cases:**
- Stop camera movement
- Emergency stop
- Halt all axes

---

## aftrs_ptz_tour_create

Create an automated preset tour for a camera

**Complexity:** moderate

**Tags:** `ptz`, `tour`, `patrol`, `automation`

**Use Cases:**
- Create camera tour
- Automate camera patrol
- Setup preset sequence

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `camera_id` | string |  | Camera to run tour on (default if only one camera) |
| `dwell_time` | number |  | Seconds to stay at each preset (default: 5) |
| `from_all_presets` | boolean |  | Create tour from all camera presets |
| `loop` | boolean |  | Loop tour continuously (default: false) |
| `name` | string | Yes | Tour name |
| `presets` | array |  | List of preset tokens to visit |

### Example

```json
{
  "camera_id": "example",
  "dwell_time": 0,
  "from_all_presets": false,
  "loop": false,
  "name": "example",
  "presets": []
}
```

---

## aftrs_ptz_tour_start

Start a camera tour

**Complexity:** simple

**Tags:** `ptz`, `tour`, `start`, `patrol`

**Use Cases:**
- Start tour
- Begin patrol
- Activate automation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tour_id` | string | Yes | Tour ID to start |

### Example

```json
{
  "tour_id": "example"
}
```

---

## aftrs_ptz_tour_status

Get status of a camera tour

**Complexity:** simple

**Tags:** `ptz`, `tour`, `status`

**Use Cases:**
- Check tour progress
- Monitor patrol
- View current preset

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tour_id` | string | Yes | Tour ID to check |

### Example

```json
{
  "tour_id": "example"
}
```

---

## aftrs_ptz_tour_stop

Stop a running camera tour

**Complexity:** simple

**Tags:** `ptz`, `tour`, `stop`, `halt`

**Use Cases:**
- Stop tour
- Halt patrol
- Pause automation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tour_id` | string | Yes | Tour ID to stop |

### Example

```json
{
  "tour_id": "example"
}
```

---

## aftrs_ptz_tours

List all configured camera tours

**Complexity:** simple

**Tags:** `ptz`, `tour`, `list`

**Use Cases:**
- List tours
- View configured patrols
- Check tour status

---

