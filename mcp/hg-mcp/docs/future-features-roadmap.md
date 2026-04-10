# Future Features Roadmap - hg-mcp

*Last Updated: January 2026*

## Current State (v2.17)

**998 tools across 97 modules** ✅ Exceeded 520+ target by 92%

### Recent Completions (Jan 2026)
- ✅ Priority 6 (Phase 3): Adaptive Self-Healing (5 tools) - healing_learn, healing_suggest, healing_pattern, healing_auto_enable, healing_rollback
- ✅ Priority 5 (Phase 3): Archive Intelligence (8 tools) - dedup, cold_detect, glacier_optimize, cost_forecast, usage_report, cleanup_preview, restore_queue, sync_status
- ✅ Priority 2 (Phase 2): Tidal Hi-Fi Integration (16 tools) - search, track, album, artist, playlist, genres, new_releases, bestsellers, quality_info, similar_artists, mixes
- ✅ Priority 2 (Phase 2): YouTube Music Integration (16 tools) - search, track, album, artist, playlist, charts, new_releases, moods, radio, extract_id
- ✅ Priority 4: Consolidated Workflow Tools (6 tools) - full_ecosystem_health, dj_crate_sync, music_ingest, troubleshoot, backup_verify, creative_suite_status
- ✅ Priority 3: Music-to-Visual AV Sync (10 tools) - BPM sync to Resolume/TD, key-to-color mapping, genre presets, energy intensity, track cues, setlist visuals, live mode
- ✅ Priority 1: Cross-Platform Music Discovery (8 tools) - unified search, track matching, price comparison, metadata enrichment, new releases, library status
- ✅ Priority 11.7: Boomkat Integration (15 tools) - search, release, artist, label, genres, new_releases, bestsellers, recommended, essential, download
- ✅ Priority 11.6: Juno Download Integration (16 tools) - search, track, release, artist, label, genres, new_releases, bestsellers, staff_picks, dj_charts, download
- ✅ Priority 11.5: Traxsource Integration (14 tools) - search, track, release, artist, label, genres, charts, download
- ✅ Priority 11.4: Discogs Integration (12 tools) - search, release, master, artist, label, collection, wantlist, marketplace
- ✅ Priority 11.3: Mixcloud Integration (10 tools) - search, user, shows, favorites, mix, tags, discover, download
- ✅ Phase 15: BPM Sync Hub (6), Trigger Sync (6), MCP Resources (4)
- ✅ Phase 16: MCP Prompts (5), MIDI Mapping (4), Enhanced Workflows (4)
- ✅ Phase 17: Timecode Sync (3), Snapshots (4), Parameter Sync (4)
- ✅ Phase 18: Video Routing (5), Show Control (4)
- ✅ Phase 19: Performance Analytics (6), Setlist Management (8)
- ✅ Phase 4a: Security & RBAC (6 tools - exceeded 3 planned)
- ✅ Phase 4b: Hardware Monitoring (4), Tailscale (3), Backup (4), Stream Deck (10)
- ✅ Phase 20: Video AI (16 tools) - upscale, denoise, interpolate, segment, depth, inpaint
- ✅ Phase 21: Lighting/LED (26 tools) - DMX, ArtNet, fixtures, scenes, groups, chases, patch
- ✅ Phase 22: UNRAID (14 tools) - Docker, VM, storage, monitoring
- ✅ Phase 23: FFGL Dev (7 tools) - scaffold, build, test, validate, package
- ✅ Architecture P0-P3: Registry, Discovery (12), Gateway (5), Consolidated (8), Chains (8), Graph (6), Healing (4), Memory (4), Federation (9)
- ✅ Priority 10.2: OPNsense Firewall (15 tools) - status, rules, NAT, interfaces, services, logs, diagnostics
- ✅ Priority 10.4: Archive Sync (6 tools) - DJ/VJ archive management, Glacier restore, cost estimation
- ✅ Priority 10.5: Video AI Pipeline (6 tools) - DXV3 conversion, queue management, job control
- ✅ Priority 10.6: System Tools (8 tools) - disk, memory, processes, thermal, battery, cache, docker, overview
- ✅ Priority 10.7: Consolidated Dashboard (4 tools) - ecosystem_health, network_overview, storage_dashboard, creative_status
- ✅ Priority 11.2: Bandcamp Integration (8 tools) - search, artist, album, download, tags, tag_releases

### Ecosystem Analysis Complete
Reviewed all hairglasses-studio repos and old-repos for integration opportunities:
- **video-ai-toolkit**: 30+ AI video processing commands
- **sam3-video-segmenter**: Specialized SAM3 object extraction
- **luke-toolkit**: DJ/music automation patterns
- **unraid-monolith**: 23 infrastructure tools
- **old-repos**: Lighting, LED, FFGL plugin patterns

---

## Priority 1: AV Integrations

### 1.1 Blackmagic ATEM Switchers ✅ COMPLETE (8 tools)

**Purpose:** Multi-camera video switching for professional productions

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_atem_status` | Get current switcher state | ✅ |
| `aftrs_atem_inputs` | List available video inputs | ✅ |
| `aftrs_atem_program` | Set program output | ✅ |
| `aftrs_atem_preview` | Set preview input | ✅ |
| `aftrs_atem_cut` | Execute cut transition | ✅ |
| `aftrs_atem_auto` | Execute auto transition | ✅ |
| `aftrs_atem_transition` | Configure transition type/rate | ✅ |
| `aftrs_atem_health` | Health check with troubleshooting | ✅ |

**Implementation:**
- Client: `internal/clients/atem.go` - UDP protocol implementation
- MCP: `internal/mcp/tools/atem/module.go`

### 1.2 PTZ Camera Control ✅ COMPLETE (15 tools)

**Purpose:** Remote pan/tilt/zoom camera operation

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_ptz_discover` | Discover ONVIF cameras | ✅ |
| `aftrs_ptz_status` | Get camera status | ✅ |
| `aftrs_ptz_move` | Pan/tilt/zoom control | ✅ |
| `aftrs_ptz_preset_recall` | Go to stored preset | ✅ |
| `aftrs_ptz_preset_save` | Store current position | ✅ |
| `aftrs_ptz_preset_list` | List all presets | ✅ |
| `aftrs_ptz_home` | Return to home position | ✅ |
| `aftrs_ptz_stop` | Stop movement | ✅ |
| `aftrs_ptz_zoom` | Zoom control | ✅ |
| `aftrs_ptz_focus` | Focus control | ✅ |
| `aftrs_ptz_iris` | Iris adjustment | ✅ |
| `aftrs_ptz_snapshot` | Capture still image | ✅ |
| `aftrs_ptz_stream_url` | Get RTSP stream URL | ✅ |
| `aftrs_ptz_reboot` | Reboot camera | ✅ |
| `aftrs_ptz_health` | Health check | ✅ |

**Implementation:**
- Client: `internal/clients/ptz.go` - ONVIF protocol
- MCP: `internal/mcp/tools/ptz/module.go`

### 1.3 NDI Expansion

**Purpose:** Enhanced video-over-IP routing (already partial implementation)

**Tools to Add:**
| Tool | Description |
|------|-------------|
| `aftrs_ndi_connect` | Connect to specific source |
| `aftrs_ndi_route` | Route to destination |
| `aftrs_ndi_bandwidth` | Monitor bandwidth usage |
| `aftrs_ndi_groups` | Manage NDI groups |

### 1.4 Dante Audio

**Purpose:** Professional audio-over-IP routing

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_dante_discover` | Find Dante devices on network |
| `aftrs_dante_flows` | List audio flows |
| `aftrs_dante_route` | Create/modify routing |
| `aftrs_dante_levels` | Monitor metered levels |
| `aftrs_dante_health` | Clock sync status |

**Integration Details:**
- Protocol: Dante Controller API (port 4439)
- Client file: `internal/clients/dante.go`
- Note: May require Dante Controller license

### 1.5 Video Capture Cards

**Purpose:** Hardware video input/output (DeckLink, AJA)

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_capture_devices` | List connected capture cards |
| `aftrs_capture_signal` | Detect input signal |
| `aftrs_capture_start` | Begin capture |
| `aftrs_playout_start` | Begin playback |

**Integration Details:**
- Client file: `internal/clients/decklink.go`
- Requires device-specific SDKs

### 1.6 Timecode/Sync

**Purpose:** Frame-accurate synchronization across devices

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_timecode_current` | Get current TC |
| `aftrs_timecode_status` | All device sync status |
| `aftrs_genlock_status` | Hardware genlock |
| `aftrs_sync_devices` | Force resync |

**Integration Details:**
- Client file: `internal/clients/timecode.go`
- Complex: Requires real-time monitoring

---

## Priority 2: Automation & Infrastructure ✅ COMPLETE

### 2.1 Hardware Monitoring ✅ IMPLEMENTED

**Purpose:** Real-time temperature, power, fan monitoring

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_cpu_temperature` | CPU core temps | ✅ |
| `aftrs_gpu_temperature` | GPU temp (NVIDIA/AMD) | ✅ |
| `aftrs_power_consumption` | System power draw | ✅ |
| `aftrs_thermal_alert` | Set thresholds | ✅ |

**Implementation:**
- Client: `internal/clients/hwmonitor.go`
- MCP: `internal/mcp/tools/hwmonitor/module.go`

**Recommended Thresholds:**
| Component | Warning | Critical |
|-----------|---------|----------|
| CPU Core | 85°C | 95°C |
| GPU | 75°C | 85°C |
| System Power | 80% | 95% |

### 2.2 Tailscale VPN ✅ IMPLEMENTED

**Purpose:** Secure remote access verification

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_tailscale_status` | Connection status | ✅ |
| `aftrs_tailscale_devices` | List devices on tailnet | ✅ |
| `aftrs_tailscale_ping` | Ping device connectivity | ✅ |

**Implementation:**
- Client: `internal/clients/tailscale.go`
- MCP: `internal/mcp/tools/tailscale/module.go`

### 2.3 Backup Automation ✅ IMPLEMENTED

**Purpose:** Automated project backups to S3/Backblaze

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_backup_projects` | Backup TD/Resolume projects | ✅ |
| `aftrs_backup_status` | Last backup timestamp | ✅ |
| `aftrs_backup_restore` | Restore from backup | ✅ |
| `aftrs_backup_list` | Browse backups | ✅ |

**Implementation:**
- Client: `internal/clients/backup.go`
- MCP: `internal/mcp/tools/backup/module.go`
- **Note:** Use `cr8` or `aftrs` AWS profiles only

**Cost Estimate:**
- Backblaze B2: ~$6/TB/month (vs AWS S3: ~$23/TB/month)
- 500GB projects = ~$3/month

### 2.4 MQTT IoT Control ✅ IMPLEMENTED

**Purpose:** Smart device control (outlets, sensors)

**Tools Implemented (5):**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_mqtt_status` | MQTT broker status | ✅ |
| `aftrs_mqtt_publish` | Send MQTT message | ✅ |
| `aftrs_mqtt_subscribe` | Monitor topic | ✅ |
| `aftrs_mqtt_topics` | List topics | ✅ |
| `aftrs_mqtt_health` | Health check | ✅ |

**Implementation:**
- Client: `internal/clients/mqtt.go`
- MCP: `internal/mcp/tools/mqtt/module.go`

**Studio Use Cases:**
- Smart power strips for equipment control
- Temperature/humidity sensors
- Air quality monitors
- Network switch power control

### 2.5 Home Assistant ✅ IMPLEMENTED

**Purpose:** Studio environment automation

**Tools Implemented (5):**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_ha_status` | HA system status | ✅ |
| `aftrs_ha_devices` | List devices | ✅ |
| `aftrs_ha_services` | Call services | ✅ |
| `aftrs_ha_states` | Get entity states | ✅ |
| `aftrs_ha_health` | Health check | ✅ |

**Implementation:**
- Client: `internal/clients/homeassistant.go`
- MCP: `internal/mcp/tools/homeassistant/module.go`

**Studio Automations:**
- Pre-show: Activate lighting preset, adjust HVAC
- During show: Monitor temperature, equipment status
- Post-show: Shutdown sequence

### 2.6 Calendar/Scheduling

**Purpose:** Show scheduling and automated triggers

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_show_schedule` | List upcoming shows |
| `aftrs_schedule_show` | Create booking |
| `aftrs_show_time_until` | Time to next show |
| `aftrs_notify_show_start` | Alert when show begins |

**Integration Details:**
- API: Google Calendar API
- Client file: `internal/clients/calendar.go`
- Library: `google.golang.org/apis/calendar/v3`

---

## Implementation Roadmap

### Phase 4a: Hardware & Network (v1.6-v1.8)

| Version | Feature | Tools | Effort |
|---------|---------|-------|--------|
| v1.6 | Hardware Monitoring | 4 | Low |
| v1.7 | Tailscale Integration | 3 | Low |
| v1.8 | Backup Automation | 5 | Medium |

### Phase 4b: AV Expansion (v1.9-v2.1)

| Version | Feature | Tools | Effort |
|---------|---------|-------|--------|
| v1.9 | NDI Expansion | 4 | Low |
| v2.0 | ATEM Switchers | 6 | High |
| v2.1 | PTZ Cameras | 6 | Medium |

### Phase 4c: Advanced (v2.2-v2.5)

| Version | Feature | Tools | Effort |
|---------|---------|-------|--------|
| v2.2 | MQTT/IoT | 5 | Medium |
| v2.3 | Home Assistant | 5 | Medium |
| v2.4 | Dante Audio | 6 | High |
| v2.5 | Calendar/Scheduling | 4 | Medium |

### Phase 5: Pro Features (v3.0+)

- Timecode/Sync system
- Video capture cards
- Multi-LLM support
- Community preset sharing

---

## File Structure for New Modules

```
internal/clients/
├── atem.go           # Blackmagic ATEM
├── ptz.go            # PTZ cameras (ONVIF)
├── dante.go          # Dante audio
├── decklink.go       # Video capture
├── timecode.go       # Sync systems
├── hwmonitor.go      # Hardware temps
├── tailscale.go      # VPN status
├── backup.go         # S3/Backblaze
├── mqtt.go           # IoT messaging
├── homeassistant.go  # HA integration
└── calendar.go       # Scheduling

internal/mcp/tools/
├── broadcast/module.go    # ATEM, PTZ
├── audio/module.go        # Dante (new)
├── capture/module.go      # Capture cards
├── sync/module.go         # Timecode
├── infrastructure/module.go # Hardware, backup
├── iot/module.go          # MQTT, HA
└── scheduling/module.go   # Calendar
```

---

## Quick Reference: Libraries

| Integration | Go Library |
|-------------|------------|
| MQTT | `github.com/eclipse/paho.mqtt.golang` |
| S3/Backblaze | `github.com/aws/aws-sdk-go-v2` |
| Tailscale | `github.com/tailscale/tailscale-client-go` |
| Calendar | `google.golang.org/apis/calendar/v3` |
| ONVIF/PTZ | `github.com/use-go/onvif` |

---

---

## Priority 3: Security & Governance ✅ COMPLETE

### 3.1 RBAC System ✅ IMPLEMENTED

**Purpose:** Multi-user access control with role-based permissions

**Roles:**
| Role | Description | Tool Access |
|------|-------------|-------------|
| admin | Full access | All tools |
| operator | Operational access | Read + write tools |
| readonly | View only | Read-only tools |

**Tools Implemented (6 - exceeded 3 planned):**
| Tool | Description | Status |
|------|-------------|--------|
| `security_whoami` | Get current user identity | ✅ |
| `security_audit_log` | View recent tool invocations | ✅ |
| `security_audit_stats` | Audit statistics | ✅ |
| `security_roles` | List available roles | ✅ |
| `security_access_check` | Check tool permissions | ✅ |
| `security_user_access` | Get user access level | ✅ |

**Implementation:**
- `pkg/security/rbac.go` - Role definitions, permissions, user management
- `pkg/security/audit.go` - Event logging, file persistence, statistics
- `internal/mcp/tools/security/module.go` - MCP tools

### 3.2 API Key Management

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `secrets_list` | List configured secrets |
| `secrets_rotate` | Rotate a secret |
| `secrets_audit` | Audit secret usage |

---

## Priority 4: Stream Deck Integration ✅ COMPLETE

### 4.1 Physical Control Surface ✅ IMPLEMENTED

**Purpose:** Trigger MCP tools from Stream Deck buttons

**Tools Implemented (10 - exceeded 4 planned):**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_streamdeck_status` | Connection status | ✅ |
| `aftrs_streamdeck_devices` | List connected devices | ✅ |
| `aftrs_streamdeck_buttons` | Get button states | ✅ |
| `aftrs_streamdeck_set_image` | Set button image | ✅ |
| `aftrs_streamdeck_set_color` | Set button color | ✅ |
| `aftrs_streamdeck_clear` | Clear button | ✅ |
| `aftrs_streamdeck_brightness` | Set brightness | ✅ |
| `aftrs_streamdeck_reset` | Reset device | ✅ |
| `aftrs_streamdeck_refresh` | Refresh buttons | ✅ |
| `aftrs_streamdeck_health` | Health check | ✅ |

**Implementation:**
- Client: `internal/clients/streamdeck.go`
- MCP: `internal/mcp/tools/streamdeck/module.go`

**Use Cases:**
- One-touch scene transitions (TD + Resolume)
- Quick sync triggers
- Alert acknowledgment buttons
- Lighting preset activation

---

## Priority 5: AI/ML Integrations ✅ PARTIAL

### 5.1 Local LLM (Ollama) ✅ IMPLEMENTED

**Purpose:** Offline AI capability using local models

**Tools Implemented (10):**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_ollama_status` | Service status | ✅ |
| `aftrs_ollama_models` | List models | ✅ |
| `aftrs_ollama_loaded` | Loaded model residency | ✅ |
| `aftrs_ollama_show` | Model metadata | ✅ |
| `aftrs_ollama_generate` | Text generation | ✅ |
| `aftrs_ollama_structured` | Schema-constrained JSON | ✅ |
| `aftrs_ollama_chat` | Chat completion | ✅ |
| `aftrs_ollama_tool_chat` | Native tool-call inspection | ✅ |
| `aftrs_ollama_health` | Service health | ✅ |
| `aftrs_ollama_readiness` | Alias/model readiness | ✅ |

**Implementation:**
- Client: `internal/clients/ollama.go`
- MCP: `internal/mcp/tools/ollama/module.go`
- Recommended models: code-primary, code-fast, code-long, nomic-embed-text:v1.5

### 5.2 Audio Transcription (Whisper) ✅ IMPLEMENTED

**Tools Implemented (4):**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_whisper_status` | Service status | ✅ |
| `aftrs_whisper_transcribe` | Transcribe audio | ✅ |
| `aftrs_whisper_translate` | Translate to English | ✅ |
| `aftrs_whisper_health` | Health check | ✅ |

**Implementation:**
- Client: `internal/clients/whisper.go`
- MCP: `internal/mcp/tools/whisper/module.go`
- Supports: Local whisper.cpp or OpenAI Whisper API

### 5.3 Visual AI

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `ai_generate_image` | Generate image from prompt |
| `ai_style_transfer` | Apply artistic styles |
| `ai_describe_image` | Describe image contents |

**Integration Details:**
- Options: Stable Diffusion (local), DALL-E API
- Client file: `internal/clients/imageai.go`

---

## Implementation Phases (Updated)

### Phase 4a: Security & Reliability (v1.7) ✅ COMPLETE
| Feature | Tools | Status |
|---------|-------|--------|
| RBAC System | 6 | ✅ Complete (exceeded 3 planned) |
| Audit Logging | ✓ | ✅ Included in security module |
| Complete TD Client | - | ✅ Complete |
| Complete Resolume Client | - | ✅ Complete |

### Phase 4b: Hardware & Network (v1.8-v1.9) ✅ COMPLETE
| Feature | Tools | Status |
|---------|-------|--------|
| Hardware Monitoring | 4 | ✅ Complete |
| Tailscale Integration | 3 | ✅ Complete |
| Backup Automation | 4 | ✅ Complete |
| Stream Deck | 10 | ✅ Complete (exceeded 4 planned) |

### Phase 4c: AV Expansion (v2.0-v2.1) ✅ COMPLETE
| Feature | Tools | Status |
|---------|-------|--------|
| NDI Expansion | 19 | ✅ Complete (ndicv module) |
| ATEM Switchers | 8 | ✅ Complete |
| PTZ Cameras | 15 | ✅ Complete (exceeded 6 planned) |
| Dante Audio | 7 | ✅ Complete |

### Phase 4d: AI/ML (v2.2-v2.4) ✅ COMPLETE
| Feature | Tools | Status |
|---------|-------|--------|
| Ollama Integration | 5 | ✅ Complete |
| MQTT IoT | 5 | ✅ Complete |
| Home Assistant | 5 | ✅ Complete |
| Whisper Transcription | 4 | ✅ Complete (exceeded 3 planned) |

---

---

## Priority 6: Video AI Integration ✅ COMPLETE

### 6.1 Core Video Enhancement

**Source:** video-ai-toolkit

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_video_upscale` | AI upscaling with RealESRGAN/Video2X |
| `aftrs_video_denoise` | Noise reduction with FastDVDNet |
| `aftrs_video_interpolate` | Frame interpolation with RIFE/FILM |
| `aftrs_video_stabilize` | Video stabilization |
| `aftrs_video_face_restore` | Face restoration with GFPGAN |
| `aftrs_video_enhance` | Combined enhancement pipeline |

### 6.2 Video Segmentation & Depth

| Tool | Description |
|------|-------------|
| `aftrs_video_segment` | Object segmentation with SAM2/Grounded SAM2 |
| `aftrs_video_matte` | Background removal with RobustVideoMatting |
| `aftrs_video_depth` | Depth estimation with DepthAnything |
| `aftrs_video_inpaint` | Object removal with ProPainter |

### 6.3 Creative Video Tools

| Tool | Description |
|------|-------------|
| `aftrs_video_colorize` | B&W colorization with DeOldify |
| `aftrs_video_style_transfer` | Artistic style transfer |
| `aftrs_video_flow` | Optical flow estimation |
| `aftrs_video_generate` | AI video generation |

### 6.4 Video Pipeline System

| Tool | Description |
|------|-------------|
| `aftrs_video_pipeline_create` | Build custom processing pipelines |
| `aftrs_video_pipeline_run` | Execute saved pipelines |
| `aftrs_video_pipeline_list` | List available pipelines |
| `aftrs_video_batch` | Batch process with parallel workers |

**Files:** `internal/clients/videoai.go`, `internal/mcp/tools/videoai/module.go`

---

## Priority 7: Advanced Lighting & LED ✅ COMPLETE (26 tools)

### 7.1 Patch Management

**Source:** aftrs-lighting

| Tool | Description |
|------|-------------|
| `aftrs_lighting_patch_validate` | Check for channel overlaps and bounds |
| `aftrs_lighting_patch_import` | Import YAML patch configuration |
| `aftrs_lighting_patch_export` | Export current patch to YAML |
| `aftrs_lighting_patch_visualize` | Generate patch visualization |

### 7.2 WLED Integration

| Tool | Description |
|------|-------------|
| `aftrs_wled_discover` | Auto-discover WLED controllers |
| `aftrs_wled_config` | Configure WLED device settings |
| `aftrs_wled_sync` | Sync WLED to Art-Net universes |
| `aftrs_wled_effects` | Manage WLED effects and presets |

### 7.3 LEDfx Enhancement

**Source:** ledfx-companion

| Tool | Description |
|------|-------------|
| `aftrs_ledfx_midi_bind` | Create MIDI → effect bindings |
| `aftrs_ledfx_osc_bind` | Create OSC → effect bindings |
| `aftrs_ledfx_scene_dj` | MIDI Clock-synced scene cycling |
| `aftrs_ledfx_curve_config` | Configure value curve transforms |

---

## Priority 8: UNRAID Expansion ✅ COMPLETE (14 tools)

### 8.1 Docker Management

**Source:** llmagent-unraid (23 tools)

| Tool | Description |
|------|-------------|
| `aftrs_docker_containers` | List all containers with status |
| `aftrs_docker_start` | Start container(s) |
| `aftrs_docker_stop` | Stop container(s) |
| `aftrs_docker_logs` | View container logs |

### 8.2 VM & Plugin Management

| Tool | Description |
|------|-------------|
| `aftrs_vm_list` | List VMs with status |
| `aftrs_vm_control` | Start/stop/pause VMs |
| `aftrs_unraid_plugins` | List installed plugins |
| `aftrs_unraid_diagnostics` | Generate diagnostics bundle |

### 8.3 Monitoring Integration

| Tool | Description |
|------|-------------|
| `aftrs_grafana_dashboard` | Launch/query Grafana dashboards |
| `aftrs_prometheus_query` | Query Prometheus metrics |
| `aftrs_loki_logs` | Search aggregated logs |

---

## Priority 9: Resolume Plugin Development ✅ COMPLETE (7 tools)

### 9.1 FFGL Plugin Scaffolding

**Source:** resolume-ffgl-llm-starter-v6

| Tool | Description |
|------|-------------|
| `aftrs_ffgl_create_effect` | Scaffold new FFGL effect plugin |
| `aftrs_ffgl_create_source` | Scaffold new FFGL source plugin |
| `aftrs_ffgl_build` | Build plugin for current platform |
| `aftrs_ffgl_test` | Run golden image tests |
| `aftrs_ffgl_validate_shader` | Validate GLSL shader code |
| `aftrs_ffgl_package` | Package for distribution |

---

## Implementation Priority Matrix

| Phase | Priority | Planned | Actual | Status |
|-------|----------|---------|--------|--------|
| 20: Video AI | P0 | 18 | **16** | ✅ Complete |
| 21: Lighting/LED | P0 | 15 | **26** | ✅ Exceeded |
| 22: UNRAID | P1 | 14 | **14** | ✅ Complete |
| 23: FFGL Dev | P2 | 10 | **7** | ✅ Core tools |
| 24: Automation | P2 | 10 | TBD | Backlog |

---

## Target: 520+ Tools ✅ ACHIEVED

| Metric | Count |
|--------|-------|
| Current tools | **676** |
| Previous count | 545 |
| Added since last update | +131 |
| **Target total** | **520+** ✅ (exceeded by 156) |
| Current modules | **68** ✅ |

---

## Old-Repos Cleanup Status

Analyzed and integrated into roadmap:
- ✅ aftrs-lighting → Phase 21 (Patch Management)
- ✅ ledfx-companion → Phase 21 (LED Enhancement)
- ✅ resolume-ffgl-v5/v6 → Phase 23 (FFGL Development)
- ⏭️ ps2-ndi-streaming → Skipped (too niche)

**Recommendation:** Safe to remove old-repos directory after this update.

---

## Priority 10: Ecosystem Interoperability (NEW - January 2026)

**Analysis Complete:** 15 AFTRS Studio repositories analyzed for integration opportunities

### 10.1 MCP Federation - Remote Tool Aggregation

**Problem:** Three separate MCP servers (hg-mcp, unraid-monolith, opnsense-monolith) run independently with no unified access.

**Solution:** Create MCP Federation client that aggregates remote MCP servers.

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_federation_status` | List all federated servers and their status |
| `aftrs_federation_discover` | Discover tools from remote servers |
| `aftrs_federation_call` | Proxy call to remote MCP tool |

**Implementation Files:**
- `internal/clients/mcp_federation.go` - Remote MCP client
- `internal/mcp/tools/federation/remote.go` - Remote tool proxy

**Environment Config:**
```bash
AFTRS_FEDERATION_UNRAID=http://192.168.50.10:8823
AFTRS_FEDERATION_OPNSENSE=http://192.168.50.1:8822
AFTRS_FEDERATION_CR8=stdio:cr8-mcp
```

### 10.2 OPNsense Firewall Integration

**Problem:** No firewall/network management tools in hg-mcp despite opnsense-monolith having 15+ tools.

**Tools to Create (15 total):**
| Tool | Category | Description |
|------|----------|-------------|
| `aftrs_opnsense_status` | system | Overall firewall status |
| `aftrs_opnsense_firewall_rules` | firewall | List active firewall rules |
| `aftrs_opnsense_firewall_states` | firewall | View connection state table |
| `aftrs_opnsense_nat_rules` | nat | List NAT/port forward rules |
| `aftrs_opnsense_interfaces` | network | List interfaces and VLANs |
| `aftrs_opnsense_routes` | network | View routing table |
| `aftrs_opnsense_services` | services | List/control services |
| `aftrs_opnsense_service_restart` | services | Restart whitelisted service |
| `aftrs_opnsense_logs` | logs | Tail firewall logs |
| `aftrs_opnsense_ping` | diag | Network connectivity test |
| `aftrs_opnsense_traceroute` | diag | Route tracing |
| `aftrs_opnsense_config_backup` | backup | Backup configuration |
| `aftrs_opnsense_config_restore` | backup | Restore from backup |
| `aftrs_opnsense_propose_rule` | firewall | Propose firewall rule change |
| `aftrs_opnsense_apply_proposal` | firewall | Apply pending rule changes |

**Implementation Files:**
- `internal/clients/opnsense.go` - OPNsense API client
- `internal/mcp/tools/opnsense/module.go` - MCP tool definitions

### 10.3 CR8-CLI Deep Integration (300+ Tools)

**Problem:** cr8-cli has 300+ Python tools but hg-mcp only wraps ~20. Lazy loading pattern not adopted.

**Solution:** Implement cr8-cli's lazy loading pattern in Go and expand tool coverage.

**Discovery Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_cr8_discover` | Browse tools by category |
| `aftrs_cr8_schema` | Load specific tool schemas on-demand |
| `aftrs_cr8_catalog` | Full category overview |

**Tool Categories to Expose (via subprocess to cr8-mcp):**
| Category | Tools | Priority |
|----------|-------|----------|
| Consolidated | 15+ (5-in-1 efficiency) | High |
| Queue | 18+ (download management) | High |
| Audio | 11+ (BPM, key, analysis) | High |
| Platforms | 8+ (Spotify, SoundCloud) | Medium |
| DJ | 7+ (Rekordbox, USB export) | Medium |
| Storage | 8+ (S3, backup) | Medium |

**Implementation Files:**
- `internal/mcp/tools/cr8/module.go` - Expand with lazy loading
- `internal/clients/cr8.go` - Add subprocess wrapper

### 10.4 Archive Sync Automation

**Problem:** dj-archive (2TB) and vj-archive (40TB) require manual sync. No unified archive management.

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_archive_status` | Combined status of all archives (DJ, VJ, cloud) |
| `aftrs_archive_sync_dj` | Sync DJ archive to/from S3 |
| `aftrs_archive_sync_vj` | Sync VJ archive to/from S3 |
| `aftrs_archive_glacier_restore` | Initiate Glacier restore for archived content |
| `aftrs_archive_search` | Search across all archives by metadata |
| `aftrs_archive_cost_estimate` | Estimate storage costs across tiers |

**Integration Points:**
- Uses existing `rclone` client for sync operations
- Leverages `data_migration` hash indexing for dedup
- Connects to Terraform state for infrastructure status

**Implementation Files:**
- `internal/mcp/tools/archive/module.go` - Archive MCP tools

### 10.5 Video AI Pipeline Orchestration

**Problem:** video-ai-toolkit and sam3-video-segmenter exist but aren't orchestrated through MCP.

**Tools to Create:**
| Tool | Description |
|------|-------------|
| `aftrs_videoai_pipeline` | Run complete AI pipeline on video |
| `aftrs_videoai_segment` | SAM3 object segmentation |
| `aftrs_videoai_depth` | Depth estimation (Depth Anything) |
| `aftrs_videoai_background` | Background removal (RVM) |
| `aftrs_videoai_inpaint` | Object removal/inpainting |
| `aftrs_videoai_to_dxv3` | Convert output to Resolume DXV3 |
| `aftrs_videoai_queue` | List/manage processing queue |

**Implementation Files:**
- `internal/mcp/tools/videoai/module.go` - Add pipeline tools

### 10.6 System Tools Consolidation (mac-mcp Port)

**Problem:** mac-mcp runs as separate Node.js server instead of integrated into hg-mcp.

**Tools to Port:**
| Tool | Description |
|------|-------------|
| `aftrs_system_disk_usage` | Disk space analysis |
| `aftrs_system_cache_clean` | Clean developer caches (Go, npm, pip, Homebrew) |
| `aftrs_system_docker_prune` | Docker resource cleanup |
| `aftrs_system_memory` | Memory usage analysis |
| `aftrs_system_thermal` | Thermal/throttling status |
| `aftrs_system_battery` | Battery health (macOS) |
| `aftrs_system_processes` | Top resource consumers |
| `aftrs_system_startup` | Startup item management |

**Implementation Files:**
- `internal/clients/macos.go` - macOS system client
- `internal/mcp/tools/system/module.go` - Cross-platform system tools

### 10.7 Consolidated Dashboard Tools

**Problem:** Status checks require calling multiple tools. No unified ecosystem view.

**Solution:** Create consolidated tools following cr8-cli's 5-in-1 pattern.

**Tools to Create:**
| Tool | Combines | Token Savings |
|------|----------|---------------|
| `aftrs_ecosystem_health` | UNRAID + OPNsense + Archives + System | ~70% |
| `hairglasses_studio_status` | Resolume + TouchDesigner + OBS + Audio | ~60% |
| `aftrs_network_overview` | OPNsense + Tailscale + Router | ~65% |
| `aftrs_storage_dashboard` | UNRAID + Archives + Cloud quotas | ~70% |
| `aftrs_creative_status` | DJ library + VJ clips + Projects | ~60% |

**Implementation Files:**
- `internal/mcp/tools/consolidated/module.go` - Efficiency tools

### Ecosystem Interoperability Summary

**Total New Tools: 45+**
| Category | Count |
|----------|-------|
| Federation | 3 |
| OPNsense | 15 |
| Archive | 6 |
| Video AI Pipeline | 7 |
| System | 8 |
| Consolidated | 5 |
| CR8 Discovery | 3 |

**Implementation Priority:**

| Phase | Features | Priority |
|-------|----------|----------|
| Phase 1 | MCP Federation + OPNsense + CR8 lazy loading | High |
| Phase 2 | Archive management + Video AI pipeline | Medium |
| Phase 3 | System tools port + Consolidated dashboards | Lower |

**Architecture Diagram:**
```
┌─────────────────────────────────────────────────────────────┐
│                    Claude / AI Assistant                     │
└─────────────────────────────┬───────────────────────────────┘
                              │
┌─────────────────────────────▼───────────────────────────────┐
│                      AFTRS-MCP (Go)                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐   │
│  │Federation│ │ OPNsense │ │ Archives │ │ Consolidated │   │
│  │  Module  │ │  Module  │ │  Module  │ │   Dashboard  │   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └──────────────┘   │
│       │            │            │                            │
│  ┌────▼────────────▼────────────▼────────────────────────┐  │
│  │              Existing 84 Tool Modules                  │  │
│  │  (UNRAID, Resolume, TouchDesigner, Rclone, etc.)      │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────┬───────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        ▼                     ▼                     ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│ unraid-monolith│    │opnsense-monolith│   │   cr8-cli    │
│  (Python MCP) │    │  (Python MCP) │    │ (Python MCP) │
│  :8823        │    │  :8822        │    │  (stdio)     │
└───────┬───────┘    └───────┬───────┘    └───────┬───────┘
        │                    │                    │
        ▼                    ▼                    ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│  WizardsTower │    │   OPNsense    │    │  DJ Library   │
│  UNRAID Server│    │   Firewall    │    │  + Archives   │
│ 192.168.50.10 │    │ 192.168.50.1  │    │  (S3/Local)   │
└───────────────┘    └───────────────┘    └───────────────┘
```

---

## Priority 11: Music Platform Integrations ✅ PARTIAL

### 10.1 SoundCloud Integration ✅ COMPLETE (16 tools)

**Pattern:** API client + download stack + S3 sync (reusable template for future platforms)

**Tools Implemented:**
| Tool | Category | Description |
|------|----------|-------------|
| `aftrs_soundcloud_status` | status | API connection status |
| `aftrs_soundcloud_health` | status | Health check with recommendations |
| `aftrs_soundcloud_resolve` | utility | Resolve any SoundCloud URL to resource |
| `aftrs_soundcloud_search` | search | Search tracks, users, playlists |
| `aftrs_soundcloud_track` | tracks | Get track details (BPM, key, waveform) |
| `aftrs_soundcloud_user` | users | Get user profile by ID |
| `aftrs_soundcloud_user_tracks` | users | Get user's uploaded tracks |
| `aftrs_soundcloud_likes` | likes | Get user's liked tracks |
| `aftrs_soundcloud_playlist` | playlists | Get playlist details and tracks |
| `aftrs_soundcloud_playlists` | playlists | List user's playlists |
| `aftrs_soundcloud_followers` | users | List followers/following |
| `aftrs_soundcloud_comments` | tracks | Get track comments |
| `aftrs_soundcloud_download` | download | Download track/playlist (scdl + yt-dlp fallback) |
| `aftrs_soundcloud_download_likes` | download | Batch download user's likes |
| `aftrs_soundcloud_sync` | sync | Trigger S3 sync for library |
| `aftrs_soundcloud_check_tools` | status | Verify download stack (scdl, yt-dlp, ffmpeg, aws) |

**Implementation Files:**
- Client: `internal/clients/soundcloud.go` - OAuth 2.0, v1+v2 API endpoints, rate limiting
- MCP: `internal/mcp/tools/soundcloud/module.go` - Tool handlers with lazy init
- Tests: `internal/mcp/tools/soundcloud/module_test.go`

**Reusable Pattern for Future Platforms:**
```
1. API Client (`internal/clients/{platform}.go`)
   - OAuth 2.0 / API key authentication
   - Rate limit handling with 429 retry
   - Circuit breaker for reliability
   - Both official + unofficial API endpoints

2. MCP Module (`internal/mcp/tools/{platform}/module.go`)
   - Lazy client initialization with sync.Once
   - Tool definitions with category/subcategory/tags
   - Handlers returning structured JSON

3. Download Integration
   - Primary tool (platform-specific CLI: scdl, spotdl, etc.)
   - Fallback to yt-dlp for universal support
   - S3 sync via existing musicsync infrastructure

4. Health Check Pattern
   - Credential validation
   - API connectivity test
   - Download tool availability check
   - Degraded status with recommendations
```

### 10.2 Bandcamp Integration ✅ COMPLETE (8 tools)

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_bandcamp_status` | Check connection and download tools | ✅ |
| `aftrs_bandcamp_health` | Health check with recommendations | ✅ |
| `aftrs_bandcamp_search` | Search artists, albums, tracks | ✅ |
| `aftrs_bandcamp_artist` | Get artist profile | ✅ |
| `aftrs_bandcamp_album` | Get album details with tracks | ✅ |
| `aftrs_bandcamp_download` | Download via bandcamp-dl or yt-dlp | ✅ |
| `aftrs_bandcamp_tags` | Get popular genre tags | ✅ |
| `aftrs_bandcamp_tag_releases` | Browse releases by genre | ✅ |

**Implementation:**
- Client: `internal/clients/bandcamp.go` - Web scraping, TralbumData JSON parsing
- MCP: `internal/mcp/tools/bandcamp/module.go`
- Download: bandcamp-dl primary, yt-dlp fallback

### 10.3 Mixcloud Integration ✅ COMPLETE (10 tools)

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_mixcloud_status` | Check API connection and yt-dlp | ✅ |
| `aftrs_mixcloud_health` | Health check with recommendations | ✅ |
| `aftrs_mixcloud_search` | Search mixes, DJs, or tags | ✅ |
| `aftrs_mixcloud_user` | Get DJ/user profile details | ✅ |
| `aftrs_mixcloud_shows` | Get user's uploaded mixes | ✅ |
| `aftrs_mixcloud_favorites` | Get user's favorite mixes | ✅ |
| `aftrs_mixcloud_mix` | Get mix details with tracklist | ✅ |
| `aftrs_mixcloud_tags` | Get popular genre tags | ✅ |
| `aftrs_mixcloud_discover` | Browse mixes by genre tag | ✅ |
| `aftrs_mixcloud_download` | Download via yt-dlp | ✅ |

**Implementation:**
- Client: `internal/clients/mixcloud.go` - Public API v1
- MCP: `internal/mcp/tools/mixcloud/module.go`
- Download: yt-dlp for audio extraction

### 11.4 Discogs Integration ✅ COMPLETE (12 tools)

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_discogs_status` | Check API connection and rate limits | ✅ |
| `aftrs_discogs_health` | Health check with recommendations | ✅ |
| `aftrs_discogs_search` | Search releases, artists, labels | ✅ |
| `aftrs_discogs_release` | Get release details with tracklist | ✅ |
| `aftrs_discogs_master` | Get master release versions | ✅ |
| `aftrs_discogs_artist` | Get artist information | ✅ |
| `aftrs_discogs_artist_releases` | List artist discography | ✅ |
| `aftrs_discogs_label` | Get label information | ✅ |
| `aftrs_discogs_collection` | View user's vinyl/CD collection | ✅ |
| `aftrs_discogs_collection_folders` | List collection folders | ✅ |
| `aftrs_discogs_wantlist` | View user's wantlist | ✅ |
| `aftrs_discogs_marketplace` | Search marketplace listings | ✅ |

**Implementation:**
- Client: `internal/clients/discogs.go` - OAuth token auth, rate limiting
- MCP: `internal/mcp/tools/discogs/module.go`
- No downloads (catalog/collection management only)

### 11.5 Tidal Hi-Fi Integration ✅ COMPLETE (16 tools)

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_tidal_status` | Check Tidal API connection | ✅ |
| `aftrs_tidal_health` | Health check with recommendations | ✅ |
| `aftrs_tidal_search` | Search tracks, albums, artists, playlists | ✅ |
| `aftrs_tidal_track` | Get track details (MQA, Hi-Res info) | ✅ |
| `aftrs_tidal_album` | Get album details with tracks | ✅ |
| `aftrs_tidal_artist` | Get artist profile | ✅ |
| `aftrs_tidal_playlist` | Get playlist details with tracks | ✅ |
| `aftrs_tidal_genres` | Browse available genres | ✅ |
| `aftrs_tidal_new_releases` | Get new album releases | ✅ |
| `aftrs_tidal_bestsellers` | Get bestselling tracks | ✅ |
| `aftrs_tidal_quality_info` | Get audio quality details (MQA, Dolby Atmos, Hi-Res) | ✅ |
| `aftrs_tidal_similar_artists` | Find similar artists | ✅ |
| `aftrs_tidal_artist_top_tracks` | Get artist's top tracks | ✅ |
| `aftrs_tidal_artist_albums` | Get artist's albums | ✅ |
| `aftrs_tidal_mixes` | Get curated mixes/playlists | ✅ |
| `aftrs_tidal_album_tracks` | Get album tracklist | ✅ |

**Implementation:**
- Client: `internal/clients/tidal.go` - OAuth 2.0 client credentials, API v1
- MCP: `internal/mcp/tools/tidal/module.go`
- Features: MQA, Dolby Atmos, Hi-Res, Sony 360 Reality Audio quality detection

### 11.6 YouTube Music Integration ✅ COMPLETE (16 tools)

**Tools Implemented:**
| Tool | Description | Status |
|------|-------------|--------|
| `aftrs_ytmusic_status` | Check yt-dlp and auth status | ✅ |
| `aftrs_ytmusic_health` | Health check with recommendations | ✅ |
| `aftrs_ytmusic_search` | Search songs, artists, albums, playlists | ✅ |
| `aftrs_ytmusic_track` | Get track/video details | ✅ |
| `aftrs_ytmusic_album` | Get album details with tracks | ✅ |
| `aftrs_ytmusic_artist` | Get artist/channel info | ✅ |
| `aftrs_ytmusic_playlist` | Get playlist details with tracks | ✅ |
| `aftrs_ytmusic_charts` | Get YouTube Music charts by country | ✅ |
| `aftrs_ytmusic_new_releases` | Browse new music releases | ✅ |
| `aftrs_ytmusic_moods` | Browse moods and genres | ✅ |
| `aftrs_ytmusic_radio` | Generate radio mix from track | ✅ |
| `aftrs_ytmusic_extract_id` | Extract video ID from URL | ✅ |
| `aftrs_ytmusic_search_songs` | Search specifically for songs | ✅ |
| `aftrs_ytmusic_search_albums` | Search specifically for albums | ✅ |
| `aftrs_ytmusic_search_artists` | Search specifically for artists | ✅ |
| `aftrs_ytmusic_search_playlists` | Search specifically for playlists | ✅ |

**Implementation:**
- Client: `internal/clients/youtube_music.go` - yt-dlp subprocess, optional browser cookies
- MCP: `internal/mcp/tools/ytmusic/module.go`
- Features: 100M+ track catalog, video-to-audio extraction, radio generation

---

## Updated Tool Counts

| Metric | Count |
|--------|-------|
| Previous count | 993 |
| Adaptive Self-Healing tools | +5 |
| **Current tools** | **998** |
| Current modules | **97** |

---
