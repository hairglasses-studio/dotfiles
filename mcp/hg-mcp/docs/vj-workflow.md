# VJ Workflow - Clip Sync & Show Prep

Quick workflow for syncing VJ clips from Google Drive to Resolume Arena.

## Prerequisites

```bash
# Install rclone if needed
brew install rclone

# Configure rclone for Google Drive (one-time)
rclone config
# Select: n (new remote) -> gdrive -> drive -> ... follow prompts

# Verify connection
rclone lsd gdrive:
```

## Quick Start

### Using vj-sync.sh Script

```bash
# Add hg-mcp/bin to PATH or run directly
./bin/vj-sync.sh status          # Check connection
./bin/vj-sync.sh list            # List packs with sizes
./bin/vj-sync.sh quick           # Sync essentials (hackerglasses, hairglasses, masks)
./bin/vj-sync.sh sync fetz       # Sync specific pack
./bin/vj-sync.sh sync-all        # Sync everything (~3.5TB)
```

### Manual rclone Commands

```bash
# Full sync to Resolume Media folder
rclone sync "gdrive:Video/VJ Tools (Laptop Salvage)/DXV3 1080p HQ Clips" ~/Documents/Resolume\ Arena/Media/VJ\ Clips --progress --transfers 8

# Specific pack
rclone sync "gdrive:Video/VJ Tools (Laptop Salvage)/DXV3 1080p HQ Clips/HACKERGLASSES" ~/Documents/Resolume\ Arena/Media/HACKERGLASSES --progress --transfers 8

# Multiple packs at once
for pack in HACKERGLASSES "Hairglasses Visuals" Masks; do rclone sync "gdrive:Video/VJ Tools (Laptop Salvage)/DXV3 1080p HQ Clips/$pack" ~/Documents/Resolume\ Arena/Media/"$pack" --progress --transfers 8; done
```

## Available Packs

| Key | Folder | Description |
|-----|--------|-------------|
| `hackerglasses` | HACKERGLASSES | 9 custom DXV clips |
| `hairglasses` | Hairglasses Visuals | AIRGLASSES, Disco Triscuits, Hackers, Morph |
| `fetz` | Fetz VJ Storage | 3D Beeple, Glitches, Ink, Video Games, Feedback |
| `masks` | Masks | DXV3 masks for compositing |
| `algorave` | Algorave | Animatrix-style clips |
| `relic` | Relic VJ Clips | Collaboration clips |
| `mantissa` | Mantissa | Abstract visuals |
| `tricky` | TrickyFM Visuals | FM-style content |
| `church` | Hairglasses at Church | Church venue clips |
| `footage` | Hairglasses Footage | Location footage (SF, Hawaii, NYC, etc.) |

## Pre-Show Checklist

1. **Check connection**
   ```bash
   ./bin/vj-sync.sh status
   ```

2. **Sync essential packs**
   ```bash
   ./bin/vj-sync.sh quick
   ```

3. **Sync show-specific packs** (optional)
   ```bash
   ./bin/vj-sync.sh sync church algorave
   ```

4. **Open Resolume Arena** and load/create composition

5. **Import synced clips** from `~/Documents/Resolume Arena/Media/`

## Resolume Composition Tips

### Loading Clips
1. Open Resolume Arena
2. Go to Files browser (F3)
3. Navigate to `~/Documents/Resolume Arena/Media/`
4. Drag clips to layers

### Existing Compositions
Your saved compositions from Google Drive:
- `Burning Man Decomp Set 2024.avc`
- `Cabin.avc`
- `New Main.avc`
- `NoTW Show 01-19-24.avc`

To restore a composition:
```bash
rclone copy "gdrive:Video/Resolume Files/Resolume Compositions/Burning Man Decomp Set 2024.avc" ~/Documents/Resolume\ Arena/Compositions/
```

## Optimized rclone Settings

For large video files:
```bash
--transfers 8           # Parallel transfers
--checkers 16           # Parallel hash checks
--buffer-size 256M      # Memory buffer per transfer
--drive-chunk-size 128M # Google Drive upload chunk size
```

For slow connections:
```bash
--transfers 4           # Fewer parallel transfers
--bwlimit 10M          # Limit bandwidth to 10MB/s
```

## Troubleshooting

### rclone not configured
```bash
rclone config
# Follow prompts to set up 'gdrive' remote
```

### VJ Tools folder not found
Check the path exists:
```bash
rclone lsd "gdrive:Video/VJ Tools (Laptop Salvage)"
```

### Resolume not finding clips
Ensure clips are in Resolume's media path:
```bash
ls ~/Documents/Resolume\ Arena/Media/
```

### Transfer too slow
Reduce transfers or add bandwidth limit:
```bash
rclone sync ... --transfers 4 --bwlimit 20M
```

## MCP Tools

If using hg-mcp, these tools are available:

| Tool | Description |
|------|-------------|
| `aftrs_gdrive_list` | List files in a folder |
| `aftrs_gdrive_search` | Search for clips |
| `aftrs_gdrive_download_videos` | Download video files |
| `aftrs_gdrive_vj_sync` | Sync to Resolume folder |
