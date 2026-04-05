# AFTRS MCP Tools

MCP (Model Context Protocol) lets Claude Code control studio equipment and run automation. The aftrs-mcp server provides 170+ tools for video, audio, TouchDesigner, streaming, and more.

## What's Available

| Category | Examples |
|----------|----------|
| **Video** | AI processing, format conversion, segmentation |
| **TouchDesigner** | Control operators, monitor FPS, manage projects |
| **Streaming** | NDI sources, OBS control, stream health |
| **Discord** | Send notifications, manage sessions |
| **Vault** | Save notes, search documentation, log sessions |

## Using MCP Tools

Just ask Claude in natural language:

```
"Check the TouchDesigner FPS"
"List available NDI sources"
"Save this session to the vault"
"Process this video with background removal"
```

Claude will use the appropriate MCP tools automatically.

## Video Processing

### Quick Examples

```
"Convert video.mov to mp4"
"Remove the background from interview.mp4"
"Upscale old_footage.mp4 to 4K"
"Extract the person from greenscreen.mp4"
```

### Available Models

| Task | Model |
|------|-------|
| Background removal | RVM, MODNet |
| Object segmentation | SAM 2, Grounded SAM |
| Upscaling | Real-ESRGAN |
| Depth estimation | Depth Anything |
| Inpainting | ProPainter |

## TouchDesigner

```
"What's the current FPS in TouchDesigner?"
"List all operators in /project1"
"Set the opacity parameter to 0.5"
"Check for errors in the network"
```

## Streaming

```
"What NDI sources are available?"
"Check stream health"
"Is OBS recording?"
```

## Vault (Notes & Documentation)

```
"Save today's session notes"
"Search the vault for NDI setup"
"Create a runbook for the projector setup"
"Log this as a show event"
```

## Discord

```
"Send a notification that the stream is live"
"Post the session summary to Discord"
```

## Finding Tools

Ask Claude:
```
"What MCP tools are available for video?"
"Show me TouchDesigner tools"
"What can I do with the vault?"
```

Or use the skill:
```
/tools
```

## Example Workflow

```
You: I have a video called raw_footage.mp4. Remove the background and save it.

Claude: I'll process that video for you using background removal.
[Runs video processing tool]
Output saved to raw_footage_nobg.mp4

You: Save a note about this to the vault

Claude: I'll save a session note.
[Creates vault entry]
Saved to vault: video-processing/2024-12-24-background-removal.md

You: Send a Discord notification that it's done

Claude: Sending notification...
[Uses Discord tool]
Notification sent!
```

## Tips

1. **Be specific about files** - Give the full filename
2. **Ask what's possible** - "What can I do with this video?"
3. **Chain tasks** - "Process the video, then save notes about it"
4. **Check status** - "Is the processing done?"

## Troubleshooting

### Tool not working?

```
"Check if aftrs-mcp is running"
"What's the status of TouchDesigner?"
```

### Need more options?

```
"What parameters does the video upscale tool have?"
"Show me all options for background removal"
```
