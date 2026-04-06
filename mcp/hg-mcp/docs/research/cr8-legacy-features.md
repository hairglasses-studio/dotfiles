# CR8-CLI Legacy Features Documentation

Extracted from legacy codebase at `~/.cursor/worktrees/cr8-cli__Workspace_/ky8nu/` (frozen Nov 12, 2025).

## Features to Preserve

### 1. YAML Playlist Registry

**Source**: `config/playlists.yaml`

A declarative playlist configuration system with:
- User-based organization with display names and base paths
- Service-specific playlists (SoundCloud, YouTube)
- Priority levels and sync intervals
- Enable/disable flags per playlist
- Shared and archived sections

**Key Pattern**:
```yaml
users:
  username:
    display_name: "Display Name"
    base_path: "DJ Crates/Username"
    enabled: true
    playlists:
      soundcloud:
        - name: "Playlist Name"
          url: "https://soundcloud.com/..."
          subdirectory: "SubDir"
          priority: high
          notes: "Description"
```

**Integration Ideas**:
- MCP tool discovery pattern (declarative tool config)
- YAML-based workflow definitions
- User preference management

---

### 2. Notification System

**Source**: `cr8_cli/notifications.py`

Multi-channel notification manager supporting:
- SMTP email notifications
- Slack webhook integration
- Configurable alert types (failures, stalls, successes)
- Daily summary reports with HTML formatting
- Worker health monitoring alerts

**Key Classes**:
- `NotificationManager` - Main notification orchestrator
- `send_failure_alert()` - Queue processing failures
- `send_worker_stall_alert()` - Worker timeout detection
- `send_daily_summary()` - Scheduled summary emails

**Integration Ideas**:
- Alert infrastructure for hg-mcp operations
- Incident notification pipeline
- Scheduled reporting system

---

### 3. Supabase Integration

**Source**: `cr8_cli/supabase_client.py`

Cloud database client supporting:
- Track metadata and audio analysis storage
- Playlist sync status tracking
- Cross-device queue management
- Real-time collaboration features
- Service role key support for backend operations

**Key Features**:
- Automatic table initialization
- Connection testing with graceful fallback
- Row-level security (RLS) bypass option

**Integration Ideas**:
- Optional cloud backend for hg-mcp
- Multi-device state synchronization
- Real-time collaboration patterns

---

### 4. Bash Utility Library (46 Scripts)

**Source**: `lib/*.sh`

Comprehensive shell script library:

| Script | Size | Purpose |
|--------|------|---------|
| `audio_intelligence.sh` | 27KB | Audio analysis and ML inference |
| `genre_classifier.sh` | 21KB | Genre detection algorithms |
| `gdrive_direct.sh` | 21KB | Google Drive API integration |
| `auto_recovery.sh` | 20KB | Self-healing and error recovery |
| `beatport_commands.sh` | 19KB | Beatport scraping and API |
| `dynamodb_queue.sh` | 19KB | DynamoDB queue management |
| `health_monitor.sh` | 19KB | System health monitoring |
| `audio_analysis.sh` | 18KB | FFprobe/FFmpeg analysis |
| `enhanced_config.sh` | 17KB | Configuration management |
| `aws_integrations.sh` | 16KB | AWS service integrations |
| `event_bus.sh` | 15KB | Event-driven architecture |
| `gdrive_copy.sh` | 14KB | GDrive file operations |
| `individual_download.sh` | 14KB | Single track download |
| `auth.sh` | 10KB | Authentication helpers |
| `aws_logger.sh` | 9KB | CloudWatch logging |
| `health_checker.sh` | 7KB | Health check utilities |
| `core.sh` | 6KB | Core functions |
| `crate.sh` | 6KB | DJ crate management |
| `database.sh` | 4KB | Local SQLite operations |
| `correlation_id.sh` | 2KB | Request tracing |

**Integration Ideas**:
- Reference library for shell-based MCP tools
- Health monitoring patterns
- Event-driven architecture reference

---

### 5. Observability Stack

**Source**: `docker/observability/`

Full monitoring stack configuration:
- `prometheus/` - Metrics collection
- `grafana/` - Dashboards and visualization
- `alertmanager/` - Alert routing
- `loki/` - Log aggregation
- `promtail/` - Log collection agent
- `alerts/` - Alert rule definitions

**Integration Ideas**:
- Observability patterns for MCP servers
- Dashboard templates
- Alert configuration examples

---

### 6. Event Bus Architecture

**Source**: `lib/event_bus.sh`

Event-driven communication pattern:
- Pub/sub message passing
- Event queuing and replay
- Handler registration
- Dead letter queue handling

**Integration Ideas**:
- Async MCP tool coordination
- Workflow orchestration
- Decoupled component communication

---

## Archive Instructions

After extracting features above:

```bash
# Create archive directory
mkdir -p ~/Archives/cr8-cli-legacy-2025-11

# Move legacy worktree
mv ~/.cursor/worktrees/cr8-cli__Workspace_/ky8nu/ ~/Archives/cr8-cli-legacy-2025-11/

# Create README in archive
cat > ~/Archives/cr8-cli-legacy-2025-11/README.md << 'EOF'
# CR8-CLI Legacy Codebase Archive

Frozen: November 12, 2025
Primary codebase: ~/Docs/cr8-cli/cr8-cli/

Features extracted to hg-mcp research docs:
- YAML playlist registry pattern
- Notification system
- Supabase integration
- 46 bash utility scripts
- Observability stack configs
- Event bus architecture
EOF
```

---

## Migration Status

| Feature | Status | Target |
|---------|--------|--------|
| YAML Playlist Registry | Documented | MCP tool patterns |
| Notification System | Documented | Alert infrastructure |
| Supabase Integration | Documented | Optional backend |
| Bash Utilities | Cataloged | Reference library |
| Observability Stack | Documented | Monitoring patterns |
| Event Bus | Documented | Async patterns |

---

*Last updated: 2025-12-30*
