# Resolume Arena/Avenue Setup Guide

This guide explains how to configure Resolume for integration with hg-mcp tools.

## Prerequisites

- Resolume Arena 7+ or Avenue 7+ (REST API requires v7+)
- Network access on ports 7000 (OSC) and dynamic REST API port

## Configuration

### 1. Enable OSC in Resolume

1. Open Resolume
2. Go to **Preferences** (Arena > Preferences on Mac, Edit > Preferences on Windows)
3. Navigate to **OSC** tab
4. Enable **OSC Input**
5. Set **Input Port** to `7000` (or your preferred port)
6. Optionally enable **OSC Output** for bidirectional communication

### 2. REST API Configuration

Resolume Arena 7+ automatically starts a REST API webserver on a dynamic port. The hg-mcp client auto-discovers this port, but you can also:

1. Check the port in **Preferences > Webserver**
2. Set the `RESOLUME_API_PORT` environment variable to override auto-discovery

### 3. Environment Variables

Configure the Resolume client via environment variables:

```bash
# OSC Configuration
export RESOLUME_OSC_HOST="127.0.0.1"  # Default: 127.0.0.1
export RESOLUME_OSC_PORT="7000"        # Default: 7000

# REST API (optional - auto-discovered if not set)
export RESOLUME_API_PORT="8080"        # Will scan ports if not set
```

## Verify Connection

### Test OSC Connection

```bash
# Using hg-mcp
aftrs resolume status

# Or via MCP tool
aftrs_resolume_status
```

### Test REST API

```bash
# Direct curl test (port may vary)
curl http://localhost:8080/api/v1/product
```

Expected response:
```json
{
  "name": "Arena",
  "major": 7,
  "minor": 16,
  "micro": 0,
  "revision": 28505
}
```

## Available Endpoints

### OSC Addresses (Input)

| Address | Type | Description |
|---------|------|-------------|
| `/composition/tempocontroller/tempo` | float | Set BPM |
| `/composition/tempocontroller/tempotap` | int | Tap tempo |
| `/composition/layers/{n}/clips/{m}/connect` | int | Trigger clip |
| `/composition/layers/{n}/video/opacity/values` | float | Layer opacity |
| `/composition/layers/{n}/bypassed` | int | Bypass layer |
| `/composition/layers/{n}/solo` | int | Solo layer |
| `/composition/columns/{n}/connect` | int | Trigger column |
| `/composition/video/opacity/values` | float | Master opacity |
| `/composition/crossfader` | float | Crossfader position |
| `/composition/dashboard/link{n}` | string | Dashboard text (1-8) |

### REST API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/product` | Product info |
| GET | `/api/v1/composition` | Full composition state |
| GET | `/api/v1/composition/layers/{n}` | Layer details |
| GET | `/api/v1/composition/layers/{n}/clips/{m}` | Clip details |
| GET | `/api/v1/composition/layers/{n}/clips/{m}/thumbnail` | Clip thumbnail |
| POST | `/api/v1/composition/layers/{n}/clips/{m}/connect` | Trigger clip |
| POST | `/api/v1/composition/layers/{n}/clips/{m}/open` | Load media |
| GET | `/api/v1/output` | Output configuration |
| GET | `/api/v1/effects/video` | Available effects |
| GET | `/api/v1/sources/video` | Available sources |

## Integration with hg-mcp

Once configured, these tools will work:

### Status & Health
- `aftrs_resolume_status` - Get connection status
- `aftrs_resolume_health` - Health check with metrics

### Layer Control
- `aftrs_resolume_layers_list` - List all layers
- `aftrs_resolume_layer_opacity` - Set layer opacity
- `aftrs_resolume_layer_bypass` - Bypass/enable layer
- `aftrs_resolume_layer_solo` - Solo layer

### Clip Control
- `aftrs_resolume_clips_list` - List clips in layer
- `aftrs_resolume_clip_trigger` - Trigger a clip
- `aftrs_resolume_clip_load` - Load media into clip slot
- `aftrs_resolume_clip_details` - Get clip metadata

### Tempo & Sync
- `aftrs_resolume_bpm_get` - Get current BPM
- `aftrs_resolume_bpm_set` - Set BPM
- `aftrs_resolume_tap_tempo` - Tap tempo

### Effects
- `aftrs_resolume_effects_list` - List layer effects
- `aftrs_resolume_effect_toggle` - Enable/disable effect
- `aftrs_resolume_effect_add` - Add effect to layer

### Dashboard Text
- `aftrs_resolume_set_artist` - Set artist display
- `aftrs_resolume_set_title` - Set title display
- `aftrs_resolume_set_now_playing` - Set artist + title

## Bidirectional Communication

### OSC Listener

For receiving feedback from Resolume (BPM changes, clip triggers, etc.), start the OSC listener:

```go
// In Go code
listener := clients.NewResolumeOSCListener(7001) // Different port from input
listener.OnBPMChange(func(bpm float64) {
    fmt.Printf("BPM changed to: %.1f\n", bpm)
})
listener.Start(ctx)
```

Configure Resolume to send OSC output:
1. Preferences > OSC > Enable OSC Output
2. Set Output Host to your machine's IP
3. Set Output Port to 7001 (or your listener port)

## Troubleshooting

### Connection Refused

1. Ensure Resolume is running
2. Check OSC is enabled in Preferences
3. Verify firewall allows connections on ports 7000 and the REST API port
4. Try restarting Resolume

### REST API Not Found

1. Ensure you're using Resolume Arena/Avenue 7+
2. Check Preferences > Webserver for the actual port
3. Set `RESOLUME_API_PORT` environment variable

### OSC Messages Not Received

1. Verify the OSC input port matches your configuration
2. Check for port conflicts with other applications
3. Test with a dedicated OSC tool like TouchOSC or Protokol

### Clips Not Triggering

1. Ensure clips are loaded in the target slots
2. Check layer is not bypassed
3. Verify column/layer indices (1-based in hg-mcp)

## Performance Notes

- OSC is UDP-based and very fast for real-time control
- REST API provides more detailed information but has higher latency
- Use OSC for time-sensitive operations (BPM sync, clip triggers)
- Use REST API for setup/configuration operations

## Dashboard Text Display

To display track info in your visuals:

1. Create a Text source in Resolume
2. In the source settings, link the text to Dashboard String 1 (or 2-8)
3. Use `aftrs_resolume_set_now_playing` to update the display

Dashboard String mapping:
- String 1: Artist
- String 2: Title
- String 3: Key
- String 4: BPM
- String 5: Genre
- Strings 6-8: Custom use
