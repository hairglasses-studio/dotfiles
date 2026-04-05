# Video Processing Tools

Tools for video conversion, processing, and AI-powered effects.

## Quick Commands

### Convert Video Format

```bash
./convert.sh input.mov output.mp4
```

### Batch Convert All MOVs

```bash
./batch-convert.sh ~/Videos/raw mp4
```

### Get Video Info

```bash
./video-info.sh video.mp4
```

## AI Processing (via aftrs-mcp)

Ask Claude:
```
"Remove background from interview.mp4"
"Upscale old_video.mp4 to 4K"
"Extract the person from the video"
"Add depth estimation to scene.mp4"
```

## Export Presets

| Preset | Resolution | Codec | Use Case |
|--------|------------|-------|----------|
| web | 1080p | H.264 | YouTube/streaming |
| archive | Original | ProRes | Master copy |
| social | 720p | H.264 | Instagram/TikTok |
| projector | 4K | H.264 | Live projection |

```bash
./export.sh video.mov --preset web
```

## Available Models

For AI processing, these models are available:

- **RVM** - Real-time background removal
- **SAM 2** - Object segmentation
- **Depth Anything** - Depth estimation
- **Real-ESRGAN** - Upscaling
- **ProPainter** - Video inpainting
