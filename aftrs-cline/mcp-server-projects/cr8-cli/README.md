# CR8 CLI MCP Server

Professional DJ Media Processing MCP Server for Cline AI Assistant

## Overview

This MCP server provides AI access to the CR8 CLI system, enabling intelligent management of professional DJ workflows including Beatport integration, audio analysis, playlist management, and media processing.

## Features

### 🎵 Media Download & Processing
- **download_track**: Download single tracks with metadata enhancement
- **download_playlist**: Batch download entire playlists with quality assurance
- **analyze_audio**: Comprehensive audio analysis (BPM, key, genre)
- **beatport_lookup**: Search Beatport database for professional DJ metadata

### 🎛️ Playlist Management
- **list_playlists**: List all registered playlists with filtering
- **sync_playlists**: Sync playlists with latest tracks
- **export_for_rekordbox**: Export playlists for Rekordbox DJ software

### 💾 Database & Cache Management
- **query_database**: Query CR8 database with advanced filtering
- **clear_cache**: Manage download and metadata caches

### 🔧 System Management
- **system_status**: Get comprehensive system health information
- **run_diagnostics**: Perform system diagnostics and API tests
- **enhance_collection**: Batch enhance existing audio collections

## Installation

```bash
cd aftrs-cline/mcp-server-projects/cr8-cli
npm install
npm run build
```

## Configuration

### Environment Variables

```bash
# CR8 CLI installation path
CR8_CLI_PATH=/home/hg/Docs/aftrs-void/cr8_cli

# Optional: Custom cache and database paths
CR8_CACHE_PATH=/path/to/cache
CR8_DB_PATH=/path/to/database
```

### Cline MCP Configuration

Add to your Cline MCP settings (`~/.config/cline/settings.json`):

```json
{
  "mcpServers": {
    "cr8-cli": {
      "command": "node",
      "args": ["/home/hg/Docs/aftrs-void/aftrs-cline/mcp-server-projects/cr8-cli/dist/index.js"],
      "env": {
        "CR8_CLI_PATH": "/home/hg/Docs/aftrs-void/cr8_cli"
      }
    }
  }
}
```

## Usage Examples

### Download a Track with Enhancement
```typescript
// Via Cline AI Assistant
"Download this track with Beatport metadata enhancement: https://soundcloud.com/artist/track"

// Calls: download_track
{
  "url": "https://soundcloud.com/artist/track",
  "quality": "320k",
  "enhance_metadata": true
}
```

### Batch Download Playlist
```typescript
// Via Cline AI Assistant
"Download the FreaQShow playlist and sync to Google Drive"

// Calls: download_playlist
{
  "url": "https://soundcloud.com/freaqshow/sets/main-crate",
  "user": "freaqshow",
  "sync_to_drive": true
}
```

### Analyze Audio File
```typescript
// Via Cline AI Assistant
"Analyze this audio file for BPM and key: /path/to/track.m4a"

// Calls: analyze_audio
{
  "file_path": "/path/to/track.m4a",
  "analysis_type": "all"
}
```

### Search Beatport Database
```typescript
// Via Cline AI Assistant
"Search Beatport for 'Deadmau5 - Ghosts n Stuff' with BPM between 120-130"

// Calls: beatport_lookup
{
  "query": "Deadmau5 - Ghosts n Stuff",
  "filters": {
    "bpm_range": {
      "min": 120,
      "max": 130
    }
  }
}
```

### Query Database with Filters
```typescript
// Via Cline AI Assistant
"Show me all house tracks with BPM between 120-128 in the database"

// Calls: query_database
{
  "query_type": "tracks",
  "filters": {
    "genre": "house",
    "bpm_range": {
      "min": 120,
      "max": 128
    }
  },
  "limit": 50
}
```

### Export for Rekordbox
```typescript
// Via Cline AI Assistant
"Export my Hype playlist for Rekordbox with cue points"

// Calls: export_for_rekordbox
{
  "playlist_name": "Hype",
  "export_format": "xml",
  "include_cue_points": true
}
```

## Integration with CR8 CLI

This MCP server interfaces with the existing CR8 CLI system at `/home/hg/Docs/aftrs-void/cr8_cli`, executing commands and returning results through the MCP protocol.

### Supported CR8 CLI Commands
- `./cr8 download <url>`
- `./cr8 gdrive-sync <user> <url>`
- `./cr8 audio-intelligence analyze <file>`
- `./cr8 beatport search <query>`
- `./cr8 crate list`
- `./cr8 sync`
- `./cr8 db query`
- `./cr8 status`
- `./cr8 doctor`
- `./cr8 batch enhance`
- `./cr8 rekordbox export`

## Error Handling

The server includes comprehensive error handling:
- Command execution errors are captured and returned
- Invalid parameters are validated
- Timeout protection for long-running operations
- Graceful degradation for missing dependencies

## Development

### Building
```bash
npm run build
```

### Development Mode
```bash
npm run dev
```

### Testing
```bash
# Test connection to CR8 CLI
CR8_CLI_PATH=/path/to/cr8_cli npm run dev

# Verify MCP server functionality
echo '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | node dist/index.js
```

## Troubleshooting

### Common Issues

1. **CR8 CLI Path Not Found**
   - Ensure `CR8_CLI_PATH` environment variable is set correctly
   - Verify CR8 CLI is executable: `chmod +x /path/to/cr8_cli/cr8`

2. **Permission Errors**
   - Check file permissions for CR8 CLI directory
   - Ensure MCP server has read/write access to cache directories

3. **API Connection Issues**
   - Verify internet connectivity for Beatport/streaming service APIs
   - Check API keys in CR8 CLI configuration

4. **Database Connection Errors**
   - Ensure SQLite database file is accessible
   - Check database permissions and path configuration

## License

MIT License - see LICENSE file for details

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with comprehensive tests
4. Submit a pull request with detailed description

## Related Projects

- [CR8 CLI](../../../cr8_cli) - Main media processing system
- [AFTRS CLI MCP Server](../aftrs-cli) - Network management MCP server
- [OPNsense MCP Server](../opnsense-monolith) - Router management MCP server

---

**Part of the AFTRS-Void MCP Server Ecosystem**
