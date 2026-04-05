# ndicv

> NDI computer vision and frame analysis tools

**6 tools**

## Tools

- [`aftrs_ndi_capture_frame`](#aftrs-ndi-capture-frame)
- [`aftrs_ndi_detect_faces`](#aftrs-ndi-detect-faces)
- [`aftrs_ndi_detect_motion`](#aftrs-ndi-detect-motion)
- [`aftrs_ndi_detect_qr`](#aftrs-ndi-detect-qr)
- [`aftrs_ndi_ocr`](#aftrs-ndi-ocr)
- [`aftrs_ndi_scene_change`](#aftrs-ndi-scene-change)

---

## aftrs_ndi_capture_frame

Capture a single frame from an NDI source.

**Complexity:** simple

**Tags:** `ndi`, `video`, `capture`, `frame`, `snapshot`

**Use Cases:**
- Capture video frame
- Take NDI snapshot
- Get video still

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `base64` | boolean |  | Return frame as base64 instead of file path |
| `output_path` | string |  | Output file path (optional, auto-generated if not provided) |
| `source` | string | Yes | NDI source name to capture from |

### Example

```json
{
  "base64": false,
  "output_path": "example",
  "source": "example"
}
```

---

## aftrs_ndi_detect_faces

Detect faces in an NDI video frame using OpenCV.

**Complexity:** simple

**Tags:** `ndi`, `video`, `face`, `detection`, `opencv`, `cv`

**Use Cases:**
- Detect faces in video
- Count people on camera
- Face tracking

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `frame_path` | string |  | Path to existing frame image (used if source not provided) |
| `source` | string |  | NDI source name (captures new frame if provided) |

### Example

```json
{
  "frame_path": "example",
  "source": "example"
}
```

---

## aftrs_ndi_detect_motion

Detect motion between two video frames.

**Complexity:** simple

**Tags:** `ndi`, `video`, `motion`, `detection`, `security`, `cv`

**Use Cases:**
- Detect motion in video
- Security monitoring
- Activity detection

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `delay_ms` | number |  | Delay between frame captures in ms (default: 500) |
| `frame1_path` | string |  | Path to first frame (used if source not provided) |
| `frame2_path` | string |  | Path to second frame (used if source not provided) |
| `source` | string |  | NDI source name (captures two frames with delay) |
| `threshold` | number |  | Motion detection threshold percentage (default: 5%) |

### Example

```json
{
  "delay_ms": 0,
  "frame1_path": "example",
  "frame2_path": "example",
  "source": "example",
  "threshold": 0
}
```

---

## aftrs_ndi_detect_qr

Detect and decode QR codes and barcodes in an NDI frame.

**Complexity:** simple

**Tags:** `ndi`, `video`, `qr`, `barcode`, `detection`, `scan`

**Use Cases:**
- Scan QR codes from video
- Read barcodes
- Decode visual data

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `frame_path` | string |  | Path to existing frame image (used if source not provided) |
| `source` | string |  | NDI source name (captures new frame if provided) |

### Example

```json
{
  "frame_path": "example",
  "source": "example"
}
```

---

## aftrs_ndi_ocr

Extract text from an NDI video frame using OCR.

**Complexity:** simple

**Tags:** `ndi`, `video`, `ocr`, `text`, `extraction`, `tesseract`

**Use Cases:**
- Read text from video
- Extract on-screen text
- Video OCR

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `frame_path` | string |  | Path to existing frame image (used if source not provided) |
| `language` | string |  | OCR language (default: eng) |
| `source` | string |  | NDI source name (captures new frame if provided) |

### Example

```json
{
  "frame_path": "example",
  "language": "example",
  "source": "example"
}
```

---

## aftrs_ndi_scene_change

Detect scene changes in NDI video stream.

**Complexity:** simple

**Tags:** `ndi`, `video`, `scene`, `change`, `detection`, `cv`

**Use Cases:**
- Detect scene changes
- Video segmentation
- Shot detection

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `delay_ms` | number |  | Delay between frame captures in ms (default: 1000) |
| `frame1_path` | string |  | Path to first frame (used if source not provided) |
| `frame2_path` | string |  | Path to second frame (used if source not provided) |
| `source` | string |  | NDI source name (captures two frames with delay) |
| `threshold` | number |  | Scene change threshold percentage (default: 30%) |

### Example

```json
{
  "delay_ms": 0,
  "frame1_path": "example",
  "frame2_path": "example",
  "source": "example",
  "threshold": 0
}
```

---

