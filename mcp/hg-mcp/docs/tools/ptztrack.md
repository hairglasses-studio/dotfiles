# ptztrack

> PTZ camera auto-tracking using NDI computer vision

**5 tools**

## Tools

- [`aftrs_ptz_track_face`](#aftrs-ptz-track-face)
- [`aftrs_ptz_track_motion`](#aftrs-ptz-track-motion)
- [`aftrs_ptz_track_snapshot`](#aftrs-ptz-track-snapshot)
- [`aftrs_ptz_track_status`](#aftrs-ptz-track-status)
- [`aftrs_ptz_track_stop`](#aftrs-ptz-track-stop)

---

## aftrs_ptz_track_face

Start face tracking mode - camera follows detected faces.

**Complexity:** moderate

**Tags:** `ptz`, `ndi`, `face`, `tracking`, `cv`, `autofollow`

**Use Cases:**
- Auto-follow presenter
- Face tracking camera
- Subject tracking

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `camera_id` | string |  | PTZ camera to control (default if only one) |
| `ndi_source` | string | Yes | NDI source to analyze for faces |
| `smoothing` | number |  | Movement smoothing 0.1-1.0 (default: 0.5) |
| `speed` | number |  | Tracking speed 0.1-1.0 (default: 0.3) |
| `zoom_to_face` | boolean |  | Auto-zoom to keep face at consistent size |

### Example

```json
{
  "camera_id": "example",
  "ndi_source": "example",
  "smoothing": 0,
  "speed": 0,
  "zoom_to_face": false
}
```

---

## aftrs_ptz_track_motion

Start motion tracking mode - camera follows motion in frame.

**Complexity:** moderate

**Tags:** `ptz`, `ndi`, `motion`, `tracking`, `cv`, `autofollow`

**Use Cases:**
- Motion-based tracking
- Activity following
- Event tracking

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `camera_id` | string |  | PTZ camera to control (default if only one) |
| `ndi_source` | string | Yes | NDI source to analyze for motion |
| `speed` | number |  | Tracking speed 0.1-1.0 (default: 0.3) |
| `threshold` | number |  | Motion detection threshold % (default: 5) |

### Example

```json
{
  "camera_id": "example",
  "ndi_source": "example",
  "speed": 0,
  "threshold": 0
}
```

---

## aftrs_ptz_track_snapshot

Capture frame from NDI source and analyze for tracking targets.

**Complexity:** simple

**Tags:** `ptz`, `ndi`, `cv`, `analysis`, `snapshot`

**Use Cases:**
- Analyze frame for targets
- Preview tracking
- Test detection

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `detect_faces` | boolean |  | Detect faces in frame (default: true) |
| `detect_motion` | boolean |  | Compare with previous frame for motion |
| `ndi_source` | string | Yes | NDI source to analyze |

### Example

```json
{
  "detect_faces": false,
  "detect_motion": false,
  "ndi_source": "example"
}
```

---

## aftrs_ptz_track_status

Get status of auto-tracking systems.

**Complexity:** simple

**Tags:** `ptz`, `tracking`, `status`

**Use Cases:**
- Check tracking status
- Monitor autofollow
- View tracking state

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `camera_id` | string |  | PTZ camera to check (all if not specified) |

### Example

```json
{
  "camera_id": "example"
}
```

---

## aftrs_ptz_track_stop

Stop all auto-tracking on a camera.

**Complexity:** simple

**Tags:** `ptz`, `tracking`, `stop`

**Use Cases:**
- Stop tracking
- Manual control
- Disable autofollow

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `camera_id` | string |  | PTZ camera to stop tracking on |

### Example

```json
{
  "camera_id": "example"
}
```

---

