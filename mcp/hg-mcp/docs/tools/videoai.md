# videoai

> AI-powered video processing and enhancement

**16 tools**

## Tools

- [`aftrs_video_batch`](#aftrs-video-batch)
- [`aftrs_video_capabilities`](#aftrs-video-capabilities)
- [`aftrs_video_colorize`](#aftrs-video-colorize)
- [`aftrs_video_denoise`](#aftrs-video-denoise)
- [`aftrs_video_depth`](#aftrs-video-depth)
- [`aftrs_video_enhance`](#aftrs-video-enhance)
- [`aftrs_video_face_restore`](#aftrs-video-face-restore)
- [`aftrs_video_flow`](#aftrs-video-flow)
- [`aftrs_video_inpaint`](#aftrs-video-inpaint)
- [`aftrs_video_interpolate`](#aftrs-video-interpolate)
- [`aftrs_video_matte`](#aftrs-video-matte)
- [`aftrs_video_pipeline_run`](#aftrs-video-pipeline-run)
- [`aftrs_video_segment`](#aftrs-video-segment)
- [`aftrs_video_stabilize`](#aftrs-video-stabilize)
- [`aftrs_video_style_transfer`](#aftrs-video-style-transfer)
- [`aftrs_video_upscale`](#aftrs-video-upscale)

---

## aftrs_video_batch

Batch process multiple videos with parallel workers.

**Complexity:** moderate

**Tags:** `video`, `batch`, `parallel`, `bulk`

**Use Cases:**
- Process multiple videos
- Overnight batch jobs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `inputs` | string | Yes | Comma-separated input file paths or glob pattern |
| `operation` | string | Yes | Operation: upscale, denoise, interpolate, etc. |
| `workers` | number |  | Parallel workers (default: 4) |

### Example

```json
{
  "inputs": "example",
  "operation": "example",
  "workers": 0
}
```

---

## aftrs_video_capabilities

List available video AI capabilities and check GPU status.

**Complexity:** simple

**Tags:** `video`, `capabilities`, `status`, `gpu`

**Use Cases:**
- Check available features
- Verify GPU setup

---

## aftrs_video_colorize

Colorize black and white video using DeOldify.

**Complexity:** moderate

**Tags:** `video`, `colorize`, `bw`, `restore`, `ai`

**Use Cases:**
- Colorize old footage
- Add color to B&W video

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: deoldify (default: deoldify) |

### Example

```json
{
  "input": "example",
  "model": "example"
}
```

---

## aftrs_video_denoise

Remove noise from video using AI denoising models.

**Complexity:** moderate

**Tags:** `video`, `denoise`, `ai`, `noise`, `quality`

**Use Cases:**
- Clean up noisy footage
- Reduce grain

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: fastdvdnet, videnn (default: fastdvdnet) |
| `strength` | number |  | Denoising strength 0.0-1.0 (default: auto) |

### Example

```json
{
  "input": "example",
  "model": "example",
  "strength": 0
}
```

---

## aftrs_video_depth

Generate depth map from video using Depth Anything or Video Depth Anything.

**Complexity:** moderate

**Tags:** `video`, `depth`, `3d`, `map`, `ai`

**Use Cases:**
- Create depth maps for VFX
- 3D video effects

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: depth_anything, video_depth (default: depth_anything) |

### Example

```json
{
  "input": "example",
  "model": "example"
}
```

---

## aftrs_video_enhance

Combined enhancement pipeline: denoise + upscale + optional face restore.

**Complexity:** moderate

**Tags:** `video`, `enhance`, `ai`, `quality`, `pipeline`

**Use Cases:**
- One-click video enhancement
- Batch quality improvement

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `denoise` | boolean |  | Apply denoising (default: true) |
| `face_restore` | boolean |  | Restore faces (default: false) |
| `input` | string | Yes | Input video file path |
| `stabilize` | boolean |  | Stabilize video (default: false) |
| `upscale` | number |  | Upscale factor: 1, 2, 4 (default: 2) |

### Example

```json
{
  "denoise": false,
  "face_restore": false,
  "input": "example",
  "stabilize": false,
  "upscale": 0
}
```

---

## aftrs_video_face_restore

Restore and enhance faces in video using GFPGAN or CodeFormer.

**Complexity:** moderate

**Tags:** `video`, `face`, `restore`, `ai`, `enhancement`

**Use Cases:**
- Enhance faces in video
- Restore old footage

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: gfpgan, codeformer (default: gfpgan) |

### Example

```json
{
  "input": "example",
  "model": "example"
}
```

---

## aftrs_video_flow

Generate optical flow visualization using RAFT or UniMatch.

**Complexity:** moderate

**Tags:** `video`, `flow`, `optical`, `motion`, `visualization`

**Use Cases:**
- Visualize motion
- Create flow effects

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: raft, unimatch (default: raft) |

### Example

```json
{
  "input": "example",
  "model": "example"
}
```

---

## aftrs_video_inpaint

Remove objects from video using ProPainter or E2FGVI inpainting.

**Complexity:** complex

**Tags:** `video`, `inpaint`, `remove`, `object`, `ai`

**Use Cases:**
- Remove unwanted objects
- Clean up footage

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `mask` | string | Yes | Mask video file path (white = remove) |
| `model` | string |  | Model: propainter, e2fgvi (default: propainter) |

### Example

```json
{
  "input": "example",
  "mask": "example",
  "model": "example"
}
```

---

## aftrs_video_interpolate

Increase frame rate using AI frame interpolation (RIFE, FILM).

**Complexity:** moderate

**Tags:** `video`, `interpolate`, `framerate`, `slowmo`, `ai`

**Use Cases:**
- Create slow motion
- Increase FPS

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: rife, film (default: rife) |
| `multiplier` | number |  | Frame rate multiplier: 2, 4, 8 (default: 2) |

### Example

```json
{
  "input": "example",
  "model": "example",
  "multiplier": 0
}
```

---

## aftrs_video_matte

Remove background from video using RobustVideoMatting or MODNet.

**Complexity:** moderate

**Tags:** `video`, `matte`, `background`, `greenscreen`, `alpha`

**Use Cases:**
- Remove background
- Create alpha matte

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `background` | string |  | Background color (e.g., '0,255,0' for green) or 'transparent' |
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: rvm, modnet (default: rvm) |

### Example

```json
{
  "background": "example",
  "input": "example",
  "model": "example"
}
```

---

## aftrs_video_pipeline_run

Run a multi-step video processing pipeline.

**Complexity:** moderate

**Tags:** `video`, `pipeline`, `batch`, `workflow`

**Use Cases:**
- Run custom processing chain
- Automate video workflows

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `steps` | string | Yes | Comma-separated steps (e.g., 'denoise,upscale:scale=4,face') |

### Example

```json
{
  "input": "example",
  "steps": "example"
}
```

---

## aftrs_video_segment

Extract objects from video using SAM2 or Grounded SAM2 with text prompts.

**Complexity:** complex

**Tags:** `video`, `segment`, `sam`, `object`, `extract`

**Use Cases:**
- Extract objects for compositing
- Create masks

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `concepts` | string |  | Comma-separated objects to extract (e.g., 'person,car') |
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: sam2, grounded_sam2 (default: sam2) |

### Example

```json
{
  "concepts": "example",
  "input": "example",
  "model": "example"
}
```

---

## aftrs_video_stabilize

Stabilize shaky video footage.

**Complexity:** moderate

**Tags:** `video`, `stabilize`, `shake`, `smooth`

**Use Cases:**
- Fix shaky footage
- Smooth handheld video

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `smoothing` | number |  | Smoothing factor 0.0-1.0 (default: 0.5) |

### Example

```json
{
  "input": "example",
  "smoothing": 0
}
```

---

## aftrs_video_style_transfer

Apply artistic style transfer to video.

**Complexity:** moderate

**Tags:** `video`, `style`, `transfer`, `artistic`, `ai`

**Use Cases:**
- Apply artistic styles
- Create stylized video

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `strength` | number |  | Style strength 0.0-1.0 (default: 0.8) |
| `style` | string | Yes | Style image path or preset name |

### Example

```json
{
  "input": "example",
  "strength": 0,
  "style": "example"
}
```

---

## aftrs_video_upscale

Upscale video resolution using AI models (RealESRGAN, Video2X).

**Complexity:** moderate

**Tags:** `video`, `upscale`, `ai`, `resolution`, `quality`

**Use Cases:**
- Upscale low-res footage
- Enhance video quality

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `input` | string | Yes | Input video file path |
| `model` | string |  | Model: realesrgan, video2x (default: realesrgan) |
| `scale` | number |  | Scale factor: 2 or 4 (default: 2) |

### Example

```json
{
  "input": "example",
  "model": "example",
  "scale": 0
}
```

---

