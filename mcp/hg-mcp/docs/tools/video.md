# video

> AI video processing tools using video-ai-toolkit

**8 tools**

## Tools

- [`aftrs_video_info`](#aftrs-video-info)
- [`aftrs_video_matrix`](#aftrs-video-matrix)
- [`aftrs_video_pipeline`](#aftrs-video-pipeline)
- [`aftrs_video_process`](#aftrs-video-process)
- [`aftrs_video_processors`](#aftrs-video-processors)
- [`aftrs_video_random`](#aftrs-video-random)
- [`aftrs_video_route`](#aftrs-video-route)
- [`aftrs_video_sources`](#aftrs-video-sources)

---

## aftrs_video_info

Get video file information (resolution, fps, duration, codec).

**Complexity:** simple

**Tags:** `video`, `info`, `metadata`, `probe`

**Use Cases:**
- Check video properties
- Verify resolution and fps
- Inspect video metadata

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `video_path` | string | Yes | Path to video file |

### Example

```json
{
  "video_path": "example"
}
```

---

## aftrs_video_matrix

View the complete video routing matrix showing all sources, destinations, and active routes.

**Complexity:** simple

**Tags:** `video`, `matrix`, `routing`, `overview`

**Use Cases:**
- View video routing overview
- Check active routes
- Audit video flow

---

## aftrs_video_pipeline

Run a multi-step video processing pipeline. Chain multiple processors in sequence.

**Complexity:** moderate

**Tags:** `video`, `pipeline`, `chain`, `workflow`

**Use Cases:**
- Chain multiple processors
- Create processing workflows
- Batch enhance videos

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `output_dir` | string |  | Output directory for processed video |
| `steps` | string | Yes | Pipeline steps (format: processor1,processor2:param=value). Example: denoise,upscale:scale=4,style:style=/path/art.jpg |
| `video_path` | string | Yes | Path to input video file |

### Example

```json
{
  "output_dir": "example",
  "steps": "example",
  "video_path": "example"
}
```

---

## aftrs_video_process

Process video with an AI model (denoise, upscale, depth, style, etc.)

**Complexity:** moderate

**Tags:** `video`, `ai`, `process`, `denoise`, `upscale`, `depth`, `style`

**Use Cases:**
- Upscale low-resolution video
- Remove noise
- Generate depth maps
- Apply artistic styles

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `output_dir` | string |  | Output directory for processed video (uses VIDTOOL_OUTPUT_DIR if not specified) |
| `params` | string |  | Processor-specific parameters (format: key=value,key=value). Examples: scale=4,preset=anime or sigma=25 or style=/path/to/art.jpg,alpha=0.8 |
| `processor` | string | Yes | Processor ID: denoise, upscale, depth, style, colorize, stabilize, face, flow, interpolate, matte, segment, inpaint, generate |
| `video_path` | string | Yes | Path to input video file |

### Example

```json
{
  "output_dir": "example",
  "params": "example",
  "processor": "example",
  "video_path": "example"
}
```

---

## aftrs_video_processors

List available video processors and their capabilities.

**Complexity:** simple

**Tags:** `video`, `list`, `processors`, `help`

**Use Cases:**
- Discover available processors
- Learn about capabilities
- Find the right tool

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | string |  | Filter by category: enhancement, analysis, creative, composition, generation |

### Example

```json
{
  "category": "example"
}
```

---

## aftrs_video_random

Generate and run a random processing pipeline for experimentation. Creates unexpected combinations of processors.

**Complexity:** moderate

**Tags:** `video`, `random`, `experiment`, `creative`

**Use Cases:**
- Experiment with random effects
- Discover new processing combinations
- Creative exploration

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `categories` | string |  | Filter by categories (comma-separated): enhancement, analysis, creative, composition, generation |
| `exclude` | string |  | Exclude processors (comma-separated): e.g., generate,inpaint |
| `max_steps` | number |  | Maximum number of processing steps (default: 4) |
| `min_steps` | number |  | Minimum number of processing steps (default: 2) |
| `preview` | boolean |  | Preview the pipeline without running it |
| `seed` | number |  | Random seed for reproducibility |
| `video_path` | string | Yes | Path to input video file |

### Example

```json
{
  "categories": "example",
  "exclude": "example",
  "max_steps": 0,
  "min_steps": 0,
  "preview": false,
  "seed": 0,
  "video_path": "example"
}
```

---

## aftrs_video_route

Create or manage video routes between systems. Routes NDI/video sources to destinations.

**Complexity:** moderate

**Tags:** `video`, `ndi`, `route`, `routing`

**Use Cases:**
- Route NDI to Resolume
- Send video to ATEM
- Configure video matrix

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `action` | string | Yes | Action: create, delete, enable, disable, list |
| `dest_input` | string |  | Destination input (e.g., 'layer/1/column/1' for Resolume, 'me/1/input/1' for ATEM) |
| `destination` | string |  | Destination system: resolume, obs, atem, touchdesigner |
| `name` | string |  | Route name (for create) |
| `route_id` | string |  | Route ID (for delete/enable/disable) |
| `source_id` | string |  | Source ID (for create) |

### Example

```json
{
  "action": "example",
  "dest_input": "example",
  "destination": "example",
  "name": "example",
  "route_id": "example",
  "source_id": "example"
}
```

---

## aftrs_video_sources

List all available video sources from Resolume, OBS, ATEM, and NDI. Discovers sources across systems.

**Complexity:** simple

**Tags:** `video`, `ndi`, `sources`, `discover`

**Use Cases:**
- Find available video sources
- Check NDI streams
- View video inputs

---

