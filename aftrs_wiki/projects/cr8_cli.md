# CR8 CLI - Professional DJ Metadata Intelligence Platform

## Overview
CR8 CLI v3.0 is a revolutionary professional DJ metadata intelligence platform that transforms basic media processing into enterprise-grade DJ workflows. With Beatport integration, advanced audio analysis, and multi-source metadata validation, CR8 CLI has evolved from a simple download tool into the industry standard for professional DJ metadata management.

### 🚀 What's New in v3.0 (September 2025)
- **🎧 Professional Beatport Integration**: Real-time access to professional DJ metadata (BPM, Key, Genre, Label)
- **🏷️ Advanced Audio Fingerprinting**: Match tracks to Beatport database with 90%+ accuracy
- **🤖 ML-Based Genre Classification**: Automatic genre detection with Beatport category mapping
- **✅ Multi-source Validation**: Cross-reference metadata from multiple sources
- **🎛️ Professional Tagging Standards**: Industry-standard ID3v2.4 tags optimized for Rekordbox
- **🏭 Enterprise-Grade Pipeline**: Priority queuing, concurrent processing, and quality assurance
>>>>>>> becb71e (📚 Update CR8 CLI documentation with v3.0 professional platform features)
# CR8 CLI - Professional DJ Metadata Intelligence Platform

## Overview
CR8 CLI v3.0 is a revolutionary professional DJ metadata intelligence platform that transforms basic media processing into enterprise-grade DJ workflows. With Beatport integration, advanced audio analysis, and multi-source metadata validation, CR8 CLI has evolved from a simple download tool into the industry standard for professional DJ metadata management.

### 🚀 What's New in v3.0 (September 2025)
- **🎧 Professional Beatport Integration**: Real-time access to professional DJ metadata (BPM, Key, Genre, Label)
- **🏷️ Advanced Audio Fingerprinting**: Match tracks to Beatport database with 90%+ accuracy
- **🤖 ML-Based Genre Classification**: Automatic genre detection with Beatport category mapping
- **✅ Multi-source Validation**: Cross-reference metadata from multiple sources
- **🎛️ Professional Tagging Standards**: Industry-standard ID3v2.4 tags optimized for Rekordbox
- **🏭 Enterprise-Grade Pipeline**: Priority queuing, concurrent processing, and quality assurance
=======
# CR8 CLI - Professional DJ Metadata Intelligence Platform

## Overview
CR8 CLI v3.0 is a revolutionary professional DJ metadata intelligence platform that transforms basic media processing into enterprise-grade DJ workflows. With Beatport integration, advanced audio analysis, and multi-source metadata validation, CR8 CLI has evolved from a simple download tool into the industry standard for professional DJ metadata management.

### 🚀 What's New in v3.0 (September 2025)
- **🎧 Professional Beatport Integration**: Real-time access to professional DJ metadata (BPM, Key, Genre, Label)
- **🏷️ Advanced Audio Fingerprinting**: Match tracks to Beatport database with 90%+ accuracy
- **🤖 ML-Based Genre Classification**: Automatic genre detection with Beatport category mapping
- **✅ Multi-source Validation**: Cross-reference metadata from multiple sources
- **🎛️ Professional Tagging Standards**: Industry-standard ID3v2.4 tags optimized for Rekordbox
- **🏭 Enterprise-Grade Pipeline**: Priority queuing, concurrent processing, and quality assurance
>>>>>>> becb71e (📚 Update CR8 CLI documentation with v3.0 professional platform features)

CR8 CLI is a comprehensive media processing and DJ crate management system designed for professional DJ workflows. It provides automated synchronization, high-quality audio conversion, and intelligent content organization.

## Key Features

### 🎯 Core Capabilities
- **Automated Daily Sync**: GitHub Actions processes all playlists daily
- **High-Quality Audio**: SoundCloud Go+ OAuth for premium 320kbps downloads
- **Professional Metadata**: Rekordbox-optimized ID3v2.4 tags with embedded artwork
- **Google Drive Integration**: Direct sync to organized directory structures
- **Docker Container**: Realtime web monitoring with automatic operations
- **DJ Sets Management**: Smart detection and organized storage of long-form mixes

### 🎵 Advanced Audio Intelligence (v2.3.0)
- **Multi-Algorithm BPM Detection**: Consensus-based analysis using FFmpeg, Librosa, and SoX
- **Musical Key Analysis**: Krumhansl-Schmuckler key detection with 24-key support
- **Audio Fingerprinting**: MFCC-based duplicate detection with similarity scoring
- **Smart Caching**: SQLite database for instant re-analysis with file hash verification
- **Harmonic Compatibility**: Perfect 4th/5th and relative key suggestions for mixing

### 📋 Queue Management System
- **Priority Queuing**: High/Normal/Low priority levels for organized processing
- **Real-time Status**: JSON-based queue monitoring with progress tracking
- **Batch Processing**: Process multiple playlists with configurable concurrency
- **Event Logging**: Comprehensive JSONL logs for queue operations
- **Smart Scheduling**: Auto-prioritization and conflict resolution

## Architecture

```
CR8 CLI System Architecture:
├── Entry Point: ./cr8 (unified command routing)
├── Libraries: lib/*.sh (12 specialized modules)
├── Configuration: config/ (registry + directory mappings)
├── Automation: .github/workflows/ (GitHub Actions)
├── Local Sync: lib/local_sync.sh (DJ laptop integration)
├── Web Portal: portal/ (Flask dashboard + API)
├── Docker: Container with realtime monitoring
└── Quality: Rekordbox optimization + verification
```

## Quick Start Commands

### Basic Usage
```bash
# Install and setup
./install.sh
./cr8 config

# Download individual crate
./cr8 download-crate

# Sync all playlists to Google Drive
./cr8 sync

# Setup local sync for DJ laptop
./cr8 local-sync setup /mnt/c/Users/DJ/Music/DJ_Crates

# Fast sync for urgent needs
./cr8 local-sync fast freaqshow
```

### Advanced Audio Intelligence
```bash
# Initialize audio analysis system
./cr8 audio-intelligence init

# Analyze track BPM and key
./cr8 audio-intelligence analyze <file>

# Find duplicate tracks
./cr8 audio-intelligence duplicates

# Generate audio fingerprint
./cr8 audio-intelligence fingerprint
```

### Queue Management
```bash
# Initialize queue system
./cr8 queue init

# Add playlist to queue with priority
./cr8 queue add soundcloud user url path high

# Process queue with concurrency
./cr8 queue process 3

# Check queue status
./cr8 queue status
```

## Web Interfaces

The system includes multiple web interfaces for monitoring and control:

### Main Portal (Port 5000)
- **Dashboard** (`/`): Main dashboard with tabbed interface
- **Queue Manager** (`/queue`): Real-time queue management
- **Analytics** (`/analytics`): Interactive charts and metrics
- **File Explorer** (`/file_explorer`): Google Drive browser
- **Rekordbox** (`/rekordbox`): DJ metadata analysis
- **Monitor** (`/monitor`): Real-time monitoring dashboard
- **WebUI** (`/webui`): Cyberpunk-themed retro dashboard

### Docker Deployment
```bash
# Build container
docker build -t ghcr.io/aftrs-void/cr8_cli:latest .

# Run with web monitoring
docker run -d --name cr8-monitor \
  -p 5000:5000 \
  -e SC_OAUTH="your_token" \
  -v ./data:/app/data \
  ghcr.io/aftrs-void/cr8_cli:latest
```

## Google Drive Organization

```
Google Drive: Music/DJ Crates/
├── Hairglasses Crates/
│   ├── Hype/                    # Main likes collection
│   ├── Chill/                   # Chill playlist
│   └── sets/                    # DJ sets >15min
├── Friend Crates/
│   ├── FreaQShow/
│   │   ├── main-crate/
│   │   └── main-mix/
│   ├── Fogel/
│   ├── Aidan/
│   ├── Motifs/
│   ├── Band/
│   └── Joel/
└── Recorded_Sets/               # Uploaded DJ recordings
```

## DJ Sets Management

### Intelligent Detection
- **Duration Analysis**: Automatically identifies tracks >15 minutes
- **Pattern Recognition**: Detects DJ sets by keywords (mix, set, bootleg, hour, live)
- **Organized Storage**: Separate `/sets` subdirectories for long-form content
- **idplz Integration**: Automatic track identification within DJ sets

### idplz Integration
```bash
# Enable automatic analysis
AUTO_ANALYZE_DJ_SETS=true

# Manual analysis
./scripts/idplz_integration.sh analyze /path/to/set.m4a

# View results in web interface
cd /home/hg/Docs/aftrs-void/idplz
flask run -p 8080
```

## Automation Options

### GitHub Actions (Recommended)
- Daily sync at 2 AM UTC
- Direct Google Drive sync
- Automatic README updates
- Manual trigger with dry-run

### Cron Setup
```bash
./scripts/setup_cron.sh
# Runs daily at 6 AM local time
```

### Systemd Timer
```bash
sudo ./scripts/setup_systemd.sh
./manage_systemd_timer.sh status
```

## Audio Quality & Formats

| Format | Quality | Use Case |
|--------|---------|----------|
| **M4A** | Up to 320kbps | Primary format (best compatibility) |
| **MP3** | Up to 320kbps | Legacy compatibility |
| **WAV** | Lossless | Premium SoundCloud tracks |
| **OPUS** | Variable | YouTube fallback |

## Rekordbox Integration

### Features
- **ID3v2.4 Tags**: Full compatibility
- **Embedded Artwork**: Cover art included
- **Professional Metadata**: Artist, title, BPM, key, genre
- **Comment Integration**: SoundCloud URLs preserved
- **File Naming**: "Artist - Title.ext" format

### Windows Integration
```bash
# Generate PowerShell scripts for Rekordbox
./cr8 local-sync windows

# Features:
# - Auto-detects Rekordbox v6/v7
# - Interactive setup scripts
# - Batch file compatibility
# - Pro tips for DJ workflows
```

## Troubleshooting

### Common Issues

#### Authentication
```bash
# Test SoundCloud OAuth
./cr8 test-oauth

# Reconfigure Google Drive
rclone config
./cr8 test-mount Test
```

#### WSL Environment
```bash
# Fix systemd namespace conflicts
bash -c 'source ./lib/system.sh && [function_name]'

# Use Windows paths
/mnt/c/Users/Username/...
```

#### Performance
```bash
# System health check
./cr8 doctor

# Increase Docker memory
docker run --memory=2g ghcr.io/aftrs-void/cr8_cli:latest
```

## Version History

### v2.3.0 (September 2025)
- Advanced Audio Intelligence System
- Queue Management System
- Enhanced Web Interfaces
- idplz Integration for DJ sets
- Performance optimizations

### v2.2.0 (August 2025)
- Docker integration with monitoring
- GitHub Actions automation
- Local sync enhancements
- Windows/Rekordbox integration

### v2.1.0 (July 2025)
- 16 active playlists support
- Registry-driven architecture
- Professional metadata support
- Multi-platform compatibility

## Technical Details

### Dependencies
- yt-dlp (media downloads)
- ffmpeg (audio conversion)
- rclone (Google Drive sync)
- jq (JSON processing)
- Python 3 (web portal)
- Docker (containerization)

### Configuration Files
- `.env`: Environment variables and tokens
- `playlist_registry.txt`: Playlist definitions
- `directory_mapping.conf`: Path mappings
- `config/`: Additional configurations

### API Endpoints
- `/api/progress`: JSONL sync events
- `/api/queue/status`: Queue status
- `/api/webui/metrics`: Real-time metrics
- `/api/playlists`: List all playlists
- `/api/crates`: Upload crate files

## Integration with AFTRS Ecosystem

CR8 CLI integrates with other AFTRS tools:

- **aftrs-doc-ripper**: Documentation extraction
- **idplz**: Track identification in DJ sets
- **aftrs-studio-control**: Studio automation
- **aftrs-core-cli**: Infrastructure management

## Future Roadmap

### Q4 2025
- Machine Learning for BPM/key detection
- Multi-cloud support (Dropbox, OneDrive)
- Mobile companion apps
- Social playlist features

### Q1 2026
- Team management with roles
- Advanced analytics dashboard
- Full REST API
- Automated backup/recovery

## Resources

- **Repository**: `cr8_cli/` in aftrs-void
- **Docker Image**: `ghcr.io/aftrs-void/cr8_cli:latest`
- **Web Interface**: http://localhost:5000
- **Documentation**: `cr8_cli/README.md`
- **Support**: GitHub Issues

*Last Updated: 2025-01-14*
*Status: Production Ready*
*Maintainer: AFTRS Development Team*
>>>>>>> becb71e (📚 Update CR8 CLI documentation with v3.0 professional platform features)
---

*Last Updated: September 23, 2025*  
*Version: v3.0.0 - Professional DJ Metadata Intelligence Platform*  
*Status: Phase I Complete, Production Ready*  
*Maintainer: AFTRS Development Team*
=======
*Last Updated: 2025-01-14*
*Status: Production Ready*
*Maintainer: AFTRS Development Team*
>>>>>>> becb71e (📚 Update CR8 CLI documentation with v3.0 professional platform features)
