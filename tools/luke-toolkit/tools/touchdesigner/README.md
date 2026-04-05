# TouchDesigner Tools

Components, templates, and utilities for TouchDesigner projects.

## Using with Claude

Claude can control TouchDesigner via MCP:

```
"Check the FPS in TouchDesigner"
"List all operators in /project1"
"Set the opacity parameter to 0.8"
"Check for errors in the network"
"What parameters does /project1/geo1 have?"
```

## Project Templates

| Template | Description |
|----------|-------------|
| `basic-video` | Video input/output with effects |
| `live-visuals` | Audio-reactive visuals |
| `projection-map` | Multi-surface projection |
| `ndi-bridge` | NDI input/output |

## Common Tasks

### Monitor Performance

```
"What's the current FPS?"
"Check cook times for expensive operators"
"Find performance bottlenecks"
```

### Control Parameters

```
"Set /project1/level1/opacity to 0.5"
"Toggle /project1/switch1"
"Reset all parameters to default"
```

### Network Management

```
"List all operators in /project1"
"Find all errors"
"What's connected to /project1/out1?"
```

## Tips

1. **Save before experimenting** - Ask Claude to note current settings
2. **Use meaningful names** - Makes it easier to reference operators
3. **Check FPS regularly** - "What's the FPS?" should be quick
4. **Log changes** - "Save this change to the vault"
