# AFTRS MCP Roadmap

## Current Status: Active Development

**Last Updated:** December 2025
**Tool Count:** 182+ tools across 22 modules

This repository serves as the centralized MCP server for the AFTRS ecosystem, aggregating tools from various projects.

---

## Phase 1: Foundation ✅ COMPLETE

### 1.1 Core Infrastructure
- [x] Set up Go MCP server skeleton (based on cobb patterns)
- [x] Implement modular tool registration system
- [x] Create transport layer (stdio for local, SSE for remote)
- [x] Add configuration management via environment variables

### 1.2 Tool Migration
- [x] Import cr8-cli MCP tools (182+ tools)
- [x] Import unraid-monolith agent tools
- [x] Import opnsense-monolith MCP tools
- [x] Create unified tool interface with lazy loading

### 1.3 Documentation
- [x] Tool API documentation
- [x] Integration guides
- [x] Example configurations
- [ ] Testing guides (in progress)

---

## Phase 2: Integration 🔄 IN PROGRESS

### 2.1 Enhanced Routing
- [x] Smart tool discovery with lazy loading (65% token savings)
- [x] Context-aware routing
- [ ] Load balancing
- [ ] Fallback mechanisms

### 2.2 Security (Priority: HIGH)
- [ ] GitHub token authentication
- [ ] RBAC tool-level permissions (admin, operator, readonly)
- [ ] Audit logging for all write operations
- [ ] Rate limiting per user/API key
- [ ] API key management and rotation

### 2.3 Monitoring
- [x] OpenTelemetry integration
- [x] Prometheus metrics export
- [x] Structured logging with context
- [x] Health checks with service status
- [x] Circuit breaker pattern with state tracking
- [x] Grafana dashboard with music sync metrics

---

## Phase 3: Advanced Features (Q3 2025)

### 3.1 Tool Composition
- [ ] Consolidated tools (replace 5-10 calls with 1)
- [ ] Multi-tool workflows
- [ ] Tool chaining
- [ ] Error handling pipelines

### 3.2 Context Management
- [ ] RequestContext per-request isolation
- [ ] Session persistence
- [ ] Result caching with TTL categories
- [ ] Memory optimization

### 3.3 Remote Access
- [ ] HTTP API endpoint with API Gateway
- [ ] SSE WebSocket support
- [ ] AWS Lambda deployment option
- [ ] Discord bot interface

---

## Phase 4: Ecosystem Expansion 🔄 IN PROGRESS

### 4.1 New Tool Categories
- [x] Visual arts tools (TouchDesigner, Resolume) - stub clients need completion
- [x] Gaming tools (retrogaming module)
- [x] Studio tools (MIDI, lighting, NDI)
- [ ] Automation tools (home automation, IoT)
- [ ] Broadcast tools (ATEM, PTZ cameras)
- [ ] Audio networking (Dante)

### 4.2 Third-Party Integration
- [ ] Plugin marketplace
- [ ] Community tools
- [ ] Integration templates
- [ ] SDK for custom tools

---

## Phase 5: Music Intelligence & DJ Tools (NEW)

### 5.1 Spotify Integration (Priority: HIGH)
- [ ] `sync_spotify_playlist` - Sync Spotify playlists to local storage
- [ ] `cr8_spotify_import` - Import track metadata from Spotify
- [ ] `cr8_spotify_recommendations` - Get AI-powered recommendations
- [ ] OAuth integration for Spotify API

### 5.2 Audio Analysis Tools (Priority: HIGH)
- [ ] `cr8_fingerprint_track` - Generate audio fingerprint (AcoustID/Chromaprint)
- [ ] `cr8_find_duplicates` - Detect duplicate tracks across services
- [ ] `cr8_analyze_audio` - BPM, key, energy analysis (existing Lambda, need MCP wrapper)
- [ ] `cr8_key_detect` - Harmonic key detection with Camelot notation

### 5.3 Stems & Remixing (Priority: MEDIUM)
- [ ] `cr8_separate_stems` - Split tracks using Demucs/Spleeter
- [ ] `cr8_extract_vocals` - Isolate vocal track
- [ ] `cr8_extract_drums` - Isolate drum track
- [ ] GPU acceleration for real-time processing

### 5.4 DJ Set Preparation (Priority: MEDIUM)
- [ ] `cr8_harmonic_graph` - Build harmonic mixing graph
- [ ] `cr8_mix_suggestions` - Suggest compatible tracks
- [ ] `cr8_generate_setlist` - AI-assisted setlist generation
- [ ] `cr8_energy_curve` - Plan energy flow for sets

---

## Phase 6: Streaming & Broadcasting (NEW)

### 6.1 Stream Deck Integration (Priority: HIGH)
- [ ] `streamdeck_buttons` - List and configure buttons
- [ ] `streamdeck_trigger` - Trigger button actions
- [ ] `streamdeck_profile_switch` - Switch profiles
- [ ] Physical control surface for MCP tools

### 6.2 Twitch Integration (Priority: MEDIUM)
- [ ] `streaming_twitch_status` - Stream status and metrics
- [ ] `streaming_twitch_chat` - Chat integration
- [ ] `streaming_twitch_clips` - Create/manage clips
- [ ] `streaming_twitch_predictions` - Channel predictions

### 6.3 YouTube Live (Priority: MEDIUM)
- [ ] `streaming_youtube_status` - Stream status
- [ ] `streaming_youtube_schedule` - Schedule streams
- [ ] `streaming_youtube_chat` - Live chat integration

### 6.4 Multi-Platform (Priority: LOW)
- [ ] `streaming_multicast_start` - Simulcast to multiple platforms
- [ ] `streaming_health_all` - Unified stream health check

---

## Phase 7: AI/ML Integrations (NEW)

### 7.1 Local LLM (Ollama) (Priority: MEDIUM)
- [ ] `ai_ollama_query` - Query local LLM
- [ ] `ai_ollama_models` - List/manage local models
- [ ] `ai_ollama_generate` - Text generation
- [ ] Offline AI capability for network-restricted environments

### 7.2 Audio Transcription (Priority: MEDIUM)
- [ ] `ai_transcribe_audio` - Whisper-based audio transcription
- [ ] `ai_transcribe_video` - Video transcription
- [ ] `ai_translate_audio` - Translation with transcription

### 7.3 Visual AI (Priority: LOW)
- [ ] `ai_generate_image` - Image generation (Stable Diffusion)
- [ ] `ai_style_transfer` - Apply artistic styles
- [ ] Integration with TouchDesigner for real-time visuals

### 7.4 Music Recommendations (Priority: MEDIUM)
- [ ] `cr8_recommend_tracks` - AI-powered track recommendations
- [ ] `cr8_similar_artists` - Find similar artists
- [ ] `cr8_playlist_generator` - Auto-generate playlists

---

## Cobb Project Insights (Reference Implementation)

Analysis of ~/example-corp/cobb reveals production-grade patterns for Go MCP servers:

### Architecture Patterns to Adopt

1. **Modular Tool System**
   ```go
   type ToolModule interface {
       Name() string
       Description() string
       Tools() []ToolDefinition
   }
   ```
   - Each module = independent package
   - Blank imports trigger registration via `init()`
   - Enables 380+ tools with clean organization

2. **Lazy Loading Strategy**
   - Problem: 380+ tools = ~65K tokens on startup
   - Solution: Three-tier disclosure
     - `tool_discover()` - names only (~500 tokens)
     - `tool_schema()` - on-demand full schemas
     - `tool_search()` - category/keyword filtering
   - **Token Savings: 65%**

3. **Consolidated Tools Pattern**
   - Replace 8-12 individual calls with 1 aggregated tool
   - Example: `cluster_health_full()` replaces pods + deployments + events + queues + alerts
   - Use `OutputSchema` for structured responses

4. **Request Context Isolation**
   ```go
   type RequestContext struct {
       UserID      string
       DisplayName string
       Roles       []Role
       SessionID   string
       RequestID   string
       Cluster     string
   }
   ```
   - Avoid global state
   - Pass via `context.Context`
   - Enables multi-user SSE mode

5. **Health Check Pattern**
   ```go
   type HealthResponse struct {
       Status     string  // healthy, degraded, critical
       Score      int     // 0-100
       Issues     []Issue
       Components map[string]ComponentHealth
   }
   ```

6. **Async Tasks (MCP SEP-1686)**
   - Long operations return task ID immediately
   - REST endpoints for polling: `GET /tasks/{id}`
   - SSE broadcasting for real-time updates

### Security Patterns

1. **GitHub Token Validation** - Simple, low-friction auth
2. **RBAC Roles** - admin, platform, support, readonly
3. **Secrets via Environment** - external-secrets operator
4. **Sensitive Key Sanitization** - Remove passwords/tokens from traces

### Observability Patterns

1. **OpenTelemetry Spans** - Tool call tracing
2. **Prometheus Metrics** - `/metrics` endpoint
3. **Structured Logging** - LogToolCall, LogAuth, LogCache
4. **Cache Metrics** - Hit/miss tracking with TTL categories

### Performance Characteristics

| Metric | Value |
|--------|-------|
| Simple read (K8s pods) | 200-500ms |
| Aggregated tool | 1-3s (parallel goroutines) |
| Complex diagnosis | 5-10s |
| Base memory | 50-100MB |
| Tool catalog cache hit | 95%+ |

---

## Discord Bot Integration (Planned)

### Overview

Create a Discord bot to allow Luke and Mitch to interact with Claude Code and hg-mcp tools via the AFTRS Discord server.

### Architecture: AWS Lambda + Interactions Endpoint

**Why Lambda (Not EC2/ECS):**
- AWS Lambda free tier: 1M requests/month, 400K GB-seconds
- Pay only for compute time used
- No server maintenance
- Perfect for 2-user Discord server

**Key Pattern: Discord HTTP Interactions Endpoint**
- Discord bots typically need persistent WebSocket connections
- Lambda functions are stateless and short-lived
- Solution: Use Discord's HTTP Interactions Endpoint instead
- Discord sends HTTP POST to Lambda via API Gateway
- Lambda processes command, returns response

### Implementation Plan

#### Phase D1: Basic Bot Infrastructure
- [ ] Create Discord application on Developer Portal
- [ ] Set up AWS Lambda function (Python or Go)
- [ ] Configure API Gateway endpoint
- [ ] Register Interactions Endpoint URL with Discord
- [ ] Implement signature verification (required by Discord)

#### Phase D2: Claude Integration
- [ ] Integrate Anthropic Claude API
- [ ] Create `/ask` slash command for general questions
- [ ] Create `/code` slash command for code-related queries
- [ ] Implement conversation context (DynamoDB state)
- [ ] Add rate limiting per user

#### Phase D3: MCP Tool Access
- [ ] Connect bot to hg-mcp server
- [ ] Create `/tools` command to list available tools
- [ ] Create `/run <tool>` command to execute MCP tools
- [ ] Add tool output formatting for Discord
- [ ] Implement permission checks

#### Phase D4: Advanced Features
- [ ] Thread-based conversations for context
- [ ] File upload/download support
- [ ] Code block formatting with syntax highlighting
- [ ] Reaction-based feedback
- [ ] Usage analytics

### AWS Infrastructure (cr8 profile)

```
Lambda Function (discord-bot)
├── API Gateway (HTTPS endpoint)
├── DynamoDB (conversation state)
├── CloudWatch (logging/monitoring)
└── IAM Role (permissions)
```

### Cost Estimate (2 users, moderate usage)

| Service | Estimated Monthly |
|---------|-------------------|
| Lambda | $0 (free tier) |
| API Gateway | $0-1 |
| DynamoDB | $0 (free tier) |
| Claude API | $5-20 (usage-based) |
| **Total** | **~$5-25/month** |

### Discord Slash Commands

| Command | Description |
|---------|-------------|
| `/ask <question>` | Ask Claude a question |
| `/code <request>` | Code generation/explanation |
| `/tools` | List available MCP tools |
| `/run <tool> [args]` | Execute an MCP tool |
| `/status` | Check bot and system status |
| `/help` | Show available commands |

### References

- [pixegami/discord-bot-lambda](https://github.com/pixegami/discord-bot-lambda) - Python example
- [Discord Interactions Endpoint docs](https://discord.com/developers/docs/interactions/receiving-and-responding)
- [AWS Lambda Discord Bot Guide](https://betterprogramming.pub/build-a-discord-bot-with-aws-lambda-api-gateway-cc1cff750292)

---

## Phase D5: Discord Admin Integration (NEW - December 2025)

Leveraging full Discord admin permissions to add 35 new MCP tools for server management.

### D5.1 Server Management Tools (Priority: HIGH)

| Tool | Description | Use Case |
|------|-------------|----------|
| `discord_admin_create_channel` | Create text/voice channels | Project-specific channels |
| `discord_admin_delete_channel` | Delete channels | Cleanup archived projects |
| `discord_admin_create_category` | Create channel categories | Organize by project/team |
| `discord_admin_edit_channel` | Edit channel name/topic/permissions | Update project status |
| `discord_admin_set_permissions` | Set channel permission overwrites | Private project channels |

### D5.2 Role Management Tools (Priority: HIGH)

| Tool | Description | Use Case |
|------|-------------|----------|
| `discord_admin_create_role` | Create roles with color/permissions | Team department roles |
| `discord_admin_delete_role` | Delete roles | Remove obsolete roles |
| `discord_admin_assign_role` | Assign role to member | Grant project access |
| `discord_admin_revoke_role` | Remove role from member | Revoke access |
| `discord_admin_list_roles` | List all roles with members | Audit role membership |

### D5.3 Member Management Tools (Priority: MEDIUM)

| Tool | Description | Use Case |
|------|-------------|----------|
| `discord_admin_kick` | Kick member from server | Remove inactive users |
| `discord_admin_ban` | Ban member (with optional reason) | Remove bad actors |
| `discord_admin_unban` | Unban member | Restore access |
| `discord_admin_timeout` | Timeout member (temporary mute) | Moderation |
| `discord_admin_set_nickname` | Set member nickname | Enforce naming conventions |
| `discord_admin_member_info` | Get detailed member info | Audit user activity |

### D5.4 Voice Channel Tools (Priority: HIGH for Studio)

| Tool | Description | Use Case |
|------|-------------|----------|
| `discord_admin_voice_move` | Move member to voice channel | Direct to recording room |
| `discord_admin_voice_disconnect` | Disconnect member from voice | Clear idle users |
| `discord_admin_voice_mute` | Server mute member | Enforce silence during recording |
| `discord_admin_voice_deafen` | Server deafen member | Full audio isolation |
| `discord_admin_voice_status` | Get voice channel status | See who's in each channel |

### D5.5 Moderation & Audit Tools (Priority: MEDIUM)

| Tool | Description | Use Case |
|------|-------------|----------|
| `discord_admin_purge` | Bulk delete messages | Cleanup spam/off-topic |
| `discord_admin_pin` | Pin/unpin messages | Mark important announcements |
| `discord_admin_audit_log` | Query Discord audit log | Track server changes |
| `discord_admin_slowmode` | Set channel slowmode | Rate limit discussions |
| `discord_admin_lock_channel` | Lock/unlock channel | Prevent messages temporarily |

### D5.6 Scheduled Events (Priority: MEDIUM)

| Tool | Description | Use Case |
|------|-------------|----------|
| `discord_admin_create_event` | Create scheduled event | Weekly jam sessions |
| `discord_admin_update_event` | Update event details | Change time/location |
| `discord_admin_delete_event` | Delete scheduled event | Cancel sessions |
| `discord_admin_list_events` | List upcoming events | View schedule |

### D5.7 Interactive Components (Priority: HIGH)

| Tool | Description | Use Case |
|------|-------------|----------|
| `discord_send_buttons` | Send message with action buttons | Approval workflows |
| `discord_send_select` | Send message with dropdown menu | Multi-option choices |
| `discord_send_modal` | Trigger modal form | Complex data input |
| `discord_update_message` | Update message with new components | Dynamic UIs |

### D5.8 Guild Settings (Priority: LOW)

| Tool | Description | Use Case |
|------|-------------|----------|
| `discord_admin_server_info` | Get full server info | Audit server settings |

---

## Phase D6: Discord Bot Access Control (NEW)

### D6.1 Role-Based Slash Commands

| Permission Level | Commands |
|------------------|----------|
| **Admin-Only** | `/td restart`, `/admin *`, `/purge` |
| **Moderator** | `/kick`, `/timeout`, `/mute` |
| **Operator** | `/sync`, `/health`, `/td status` |
| **Public** | `/help`, `/ping`, `/whoami`, `/tools` |

### D6.2 Implementation Notes

- Map Discord roles to RBAC roles (admin, operator, readonly)
- All admin operations logged to security audit
- Rate limiting: 10 admin operations per minute
- Destructive operations (kick/ban/delete) require confirmation

### D6.3 Gateway Intents Required

```go
session.Identify.Intents = discordgo.IntentsGuilds |
    discordgo.IntentsGuildMembers |      // PRIVILEGED - member tracking
    discordgo.IntentsGuildModeration |   // ban/kick events
    discordgo.IntentsGuildVoiceStates |  // voice channel tracking
    discordgo.IntentsGuildMessages |
    discordgo.IntentsDirectMessages |
    discordgo.IntentsMessageContent      // PRIVILEGED - message parsing
```

---

## Go Secrets Manager (Planned)

### Overview

Implement a Go-based secrets manager for hg-mcp, inspired by cobb's secretsoperator and `webb_whoami()` patterns.

### Design Principles

1. **12-Factor App Compliance** - Secrets via environment variables
2. **External Secrets Integration** - AWS Secrets Manager + external-secrets operator
3. **Sanitization** - Remove sensitive values from logs/traces
4. **Rotation Support** - Hot-reload secrets without restart

### Implementation

#### Core Components

```go
// secrets/manager.go
type SecretsManager struct {
    providers []SecretProvider
    cache     map[string]*Secret
    ttl       time.Duration
}

type SecretProvider interface {
    Name() string
    Get(key string) (*Secret, error)
    List() ([]string, error)
}

type Secret struct {
    Key       string
    Value     string
    ExpiresAt time.Time
    Source    string  // "env", "aws", "1password"
}
```

#### Providers

1. **Environment Provider** - Read from `os.Getenv()`
2. **AWS Secrets Manager Provider** - Use cr8 profile
3. **1Password CLI Provider** - Use `op read` command
4. **File Provider** - Read from `.env` or secrets files

#### whoami Pattern

```go
// Inspired by webb_whoami()
func Whoami() (*Identity, error) {
    return &Identity{
        User:        os.Getenv("USER"),
        AWSProfile:  os.Getenv("AWS_PROFILE"),
        AWSAccount:  getAWSAccountID(),
        GitHubUser:  getGitHubUser(),
        Cluster:     os.Getenv("KUBECTL_CONTEXT"),
        Permissions: getRoles(),
    }, nil
}
```

#### Sensitive Key Patterns

```go
var SensitiveKeys = []string{
    "password", "secret", "token", "key", "credential",
    "auth", "bearer", "api_key", "private", "oauth",
}

func Sanitize(params map[string]any) map[string]any {
    // Remove values matching sensitive patterns
}
```

### AWS Secrets Manager Integration

```yaml
# Helm values for external-secrets
externalSecrets:
  enabled: true
  secretStore: aftrs-aws-secretsmanager
  secrets:
    - envVar: DISCORD_BOT_TOKEN
      remoteKey: hg-mcp/discord-bot-token
    - envVar: ANTHROPIC_API_KEY
      remoteKey: hg-mcp/anthropic-api-key
    - envVar: GITHUB_TOKEN
      remoteKey: hg-mcp/github-token
```

---

## Tool Migration Priority

| Source | Tools | Priority |
|--------|-------|----------|
| cr8-cli | 300+ MCP tools | High |
| unraid-monolith | 20+ agent tools | High |
| opnsense-monolith | 15+ network tools | Medium |
| visual-projects | TBD | Medium |
| studio-projects | TBD | Medium |
| gaming-projects | TBD | Low |

---

## Performance Targets

| Metric | Target |
|--------|--------|
| Tool discovery time | <100ms |
| Average tool execution | <500ms |
| Concurrent connections | 100+ |
| Uptime | 99.9% |
| Token savings (lazy loading) | 65%+ |
| Consolidated tool reduction | 5-10x fewer calls |

---

## Go Migration Path (from Python)

### Rationale

- Deeper MCP tooling integration (native Go MCP libraries)
- Single binary deployment
- Better performance for queue processing
- Native concurrency support
- Lower memory footprint

### Scope: Go Components

- MCP server core
- Queue workers and processors
- Storage operations
- API endpoints
- Secrets manager
- Discord bot

### Scope: Keep in Python

- Audio analysis (librosa)
- ML models (TensorFlow/PyTorch)
- Vector similarity search
- Complex data science workflows

---

## Development Log

### 2025-12-29: AWS Migration Complete (cr8-cli)

- Migrated Supabase PostgreSQL → DynamoDB (7 tables)
- Migrated storage backend → S3 (cr8 profile)
- Created migration MCP tools
- 999+ tracks, 27 playlists, 21 queue items migrated
- Local downloads migration in progress (28GB → S3)

### 2025-12-29: Cobb Analysis Complete

- Analyzed ~/example-corp/cobb for Go MCP patterns
- Documented modular tool system, lazy loading, consolidated tools
- Identified authentication, observability, caching patterns
- Created implementation roadmap based on cobb architecture

### 2025-12-30: AWS Infrastructure & Observability

**Completed:**
- Analysis worker processing CR8 queue (~13 tracks/min)
- Observability stack: Prometheus, OTEL Collector, Grafana (Docker Compose)
- OpenTelemetry instrumentation for all MCP tools
- New tools: `cr8_status`, `cr8_track_search`, `cr8_queue_status`
- Terraform modules for Lambda-based analysis worker

**In Progress:**
- Lambda container build and deployment (cr8-analysis-worker)

---

## CR8 AWS Infrastructure Assessment

### Currently Managed by Terraform (hg-mcp/terraform/)

| Resource | Status | Notes |
|----------|--------|-------|
| **DynamoDB Tables (10)** | | |
| cr8_tracks | ✅ Imported | Main track metadata |
| cr8_playlists | ✅ Imported | Playlist definitions |
| cr8_playlist_tracks | ✅ Imported | Playlist-track mappings |
| cr8_sync_queue | ✅ Imported | Analysis queue |
| cr8_audio_analysis | ✅ Imported | Analysis results cache |
| cr8_sync_history | ✅ Imported | Sync operation logs |
| cr8_track_history | ✅ Imported | Track modification history |
| cr8_user_preferences | ✅ Imported | User settings |
| cr8_playlist_state | ✅ Imported | Playlist sync state |
| cr8-terraform-locks | ✅ Created | State locking |
| **S3 Buckets (3)** | | |
| cr8-music-storage | ✅ Imported | Audio file storage (us-east-1) |
| aftrs-dj-archive | ✅ Imported | DJ set archives (ap-southeast-2) |
| aftrs-vj-archive | ✅ Imported | VJ content archives (ap-southeast-2) |
| **ECR Repositories (2)** | | |
| cr8-analysis-lambda | ✅ Created | Lambda container repo |
| cr8-cli/mcp | ✅ Imported | MCP server container |
| **Lambda & EventBridge** | | |
| cr8-analysis-worker | ✅ Created | Audio analysis Lambda |
| cr8-analysis-schedule | ✅ Created | Scheduled trigger (5 min) |
| **IAM Roles (4)** | | |
| cr8-analysis-lambda-role | ✅ Created | Lambda execution role |
| cr8-mcp-execution | ✅ Imported | MCP ECS task execution |
| cr8-mcp-task | ✅ Imported | MCP ECS task role |
| github-actions-hg-mcp | ✅ Created | OIDC CI/CD role |

### Legacy AWS Resources (NOT Managed)

These resources exist but are intentionally NOT imported. They may be deprecated or used by other systems:

#### S3 Buckets
| Bucket | Purpose | Status |
|--------|---------|--------|
| cr8-audio-files-REDACTED | Legacy audio bucket | Deprecated - use cr8-music-storage |
| cr8-tf-state-usw1 | us-west-1 state bucket | Unknown - investigate |

#### IAM Roles (ECS - may still be in use)
| Role | Purpose | Status |
|------|---------|--------|
| cr8-ecs-task-dev | ECS task execution | Active if using ECS |
| cr8-ecs-task-execution-dev | ECS service role | Active if using ECS |
| cr8-github-actions-dev | Old CI/CD role | May be replaced by github-actions-hg-mcp |

#### IAM Roles (Low Priority / Deprecated)
| Role | Purpose | Status |
|------|---------|--------|
| cr8-spot-launcher-dev | Spot instance management | Low priority |
| cr8-spot-processor-dev | Spot processing | Low priority |
| cr8-lambda-enrichment-dev | Lambda enrichment | Low priority |
| cr8-geo-fallback-role | Geo failover | Low priority |

#### ECR Repositories
| Repository | Purpose | Status |
|------------|---------|--------|
| aftrs-void/cr8-cli | CLI container | Legacy - other project |

### Infrastructure Import Progress

- [x] **Phase I1: Critical Infrastructure** (2025-12-30)
  - ✅ Import `cr8-mcp-execution` and `cr8-mcp-task` IAM roles
  - ✅ Import `cr8-cli/mcp` ECR repository
  - ✅ Lambda `cr8-analysis-worker` already managed

- [x] **Phase I2: Data & Archives** (2025-12-30)
  - ✅ All 10 DynamoDB tables verified in Terraform state
  - ✅ Import `aftrs-dj-archive` S3 bucket (ap-southeast-2)
  - ✅ Import `aftrs-vj-archive` S3 bucket (ap-southeast-2)

- [x] **Phase I3: Documentation** (2025-12-30)
  - ✅ Document legacy resources above
  - ⏸️ ECS roles left unmanaged (may still be in use)
  - ⏸️ Spot/Lambda enrichment roles left unmanaged (low priority)

---

## Legacy CR8-CLI Features (Extracted 2025-12-30)

Features extracted from legacy codebase at `~/.cursor/worktrees/cr8-cli__Workspace_/ky8nu/` before archiving.
Full documentation: `docs/research/cr8-legacy-features.md`

### High Priority Features

| Feature | Description | Target Module |
|---------|-------------|---------------|
| **YAML Playlist Registry** | Declarative playlist config with user/service hierarchy | MCP tool discovery patterns |
| **Queue Management** | DynamoDB queue with retry logic, dead letter handling | Task scheduling module |
| **Event Bus Architecture** | Pub/sub event system with replay and dead letter queue | Async MCP coordination |

### Medium Priority Features

| Feature | Description | Target Module |
|---------|-------------|---------------|
| **Notification System** | Email + Slack alerts, daily summaries, worker health monitoring | Alert infrastructure |
| **Supabase Integration** | Cloud database client with RLS bypass, real-time sync | Optional cloud backend |
| **Observability Stack** | Prometheus, Grafana, Loki, AlertManager configs | Monitoring templates |

### Low Priority / Reference

| Feature | Description | Notes |
|---------|-------------|-------|
| **46 Bash Utilities** | Audio analysis, AWS integrations, health monitoring | Reference library |
| **Auto Recovery** | Self-healing patterns for queue workers | Reliability patterns |
| **Genre Classifier** | ML-based genre detection in bash | Audio intelligence |

### Implementation Ideas

1. **YAML Tool Registry** - Apply playlist registry pattern to MCP tool configuration
2. **Consolidated Health Tool** - Port health_monitor.sh patterns to Go
3. **Alert Pipeline** - Notifications.py → Go notification service
4. **Event-Driven Tools** - Event bus patterns for async tool coordination

---

### 2025-12-30: CR8-CLI Path Fix & Legacy Archive

**S3 Path Generation Fixed:**
- Fixed `normalize_path()` in `sync_playlist_likes.py`
- New tracks now use `downloads/dj_crates/{user}/{playlist}/` hierarchy
- Previous flat paths: `downloads/{user}/`

**Legacy Codebase Archived:**
- Documented 6 major feature categories
- Created `docs/research/cr8-legacy-features.md`
- Legacy worktree: `~/.cursor/worktrees/cr8-cli__Workspace_/ky8nu/`
- Archive destination: `~/Archives/cr8-cli-legacy-2025-11/`

---

### 2025-12-30: Enhanced Sync & Deduplication System

**DynamoDB Tables Created:**
- `cr8_sync_state` - Tracks sync progress per playlist (resume capability)
- `cr8_sync_logs` - Historical sync session logs with duration, counts, errors

**Title-Based Deduplication:**
- In-memory title cache loads 8,400+ tracks at sync start
- Normalized title matching (ignores remixes, special chars, brackets)
- **90% hit rate** in testing - 18/20 tracks reused without re-download
- Falls back to URL-based matching for new-format tracks

**Resume Capability:**
- Saves progress every 10 tracks to `cr8_sync_state`
- Resumes from last position if sync is interrupted
- Tracks downloaded/skipped/failed counts across sessions

**New MCP Tools:**
- `cr8_sync_status` - View sync state, progress, history, reset state
- `cr8_playlist_list` - Lightweight playlist listing (~150 tokens)

**Type Fixes:**
- Fixed playlist_id type mismatch (string vs number) in DynamoDB queries
- Consistent string storage for playlist_id in playlist_tracks table

### 2025-12-30: DynamoDB Performance Optimization

**Problem:** Database scan phase took 8+ minutes due to `scan()` with FilterExpression

**Solution:**
- Changed `scan()` → `query()` using `playlist_id` partition key (already HASH key)
- Changed N+1 `get_item()` → `batch_get_item()` (100 items per batch)

**Results:**
| Phase | Before | After |
|-------|--------|-------|
| Database lookup | 8+ minutes | 6 seconds |
| Total sync time | ~10 minutes | ~23 seconds |

**Files modified:**
- `sync_playlist_likes.py` - `get_existing_track_urls()`
- `sync_to_rekordbox.py` - `get_playlist_track_ids()`

### 2025-12-30: Parallel Downloads & Bug Fixes

**Parallel Downloads:**
- Converted `download_track()` to true async using `asyncio.create_subprocess_exec`
- Added `process_single_track()` worker function with semaphore-based concurrency
- Batch processing with `asyncio.gather()` for parallel execution
- Configurable via `CR8_PARALLEL_DOWNLOADS` env var (default: 5)
- Resume capability preserved with state saved after each batch

**Bug Fixes:**
- Fixed ValidationException on playlist metadata update (str vs int key type)

### 2025-12-30: Beatport Metadata Enrichment

**Lambda Migration (Supabase → DynamoDB):**
- Replaced `update_supabase()` with `update_dynamodb()`
- Uses ExpressionAttributeNames to handle reserved words
- Added `enrichment_source` tracking field

**New Metadata Fields from Beatport API:**

| Field | Description |
|-------|-------------|
| label | Record label name |
| sub_genre | Specific sub-genre (e.g., "Tech House") |
| mix_name | "Original Mix", "Extended Mix", etc. |
| remixer | Remixer artist names |
| isrc | International Standard Recording Code |
| release_date | Track release date |
| artwork_url | Cover art image URL |
| camelot_key | Camelot wheel notation (8A, 1B) |
| beatport_id | Beatport track ID |
| beatport_url | Link to Beatport page |
| catalog_number | Label catalog number (from Discogs) |

**DynamoDB Schema Updates:**
- Added `label` attribute to cr8_tracks
- Added `label-index` GSI for label-based queries
- All 4 GSIs now active: url-index, bpm_bucket-index, artist-index, label-index

**Files Modified:**
- `terraform/dynamodb.tf` - Added label attribute and GSI
- `lambda/metadata_enrichment/handler.py` - Full DynamoDB migration + extended fields

**Setup Required - API Tokens:**
```bash
# Create secrets in AWS Secrets Manager
aws secretsmanager create-secret \
  --name cr8-api-tokens-dev \
  --secret-string '{"BEATPORT_ACCESS_TOKEN":"...","DISCOGS_TOKEN":"..."}'
```

To obtain tokens:
- **Beatport**: Apply at https://api.beatport.com/v4/docs/ (requires label/partner account)
- **Discogs**: Create app at https://www.discogs.com/settings/developers

---

## Future CR8 Enhancements (Not Yet Implemented)

### High Priority

| Enhancement | Description | Status |
|-------------|-------------|--------|
| **S3 Duplicate Detection** | Check S3 for existing files before download | Planned |
| **Parallel Downloads** | asyncio.gather for concurrent track downloads | ✅ Done |
| **Smart Timeout** | Dynamic timeout based on file size hints | Planned |
| **Rekordbox Auto-Sync** | Trigger after playlist sync completion | ✅ Done |

### Medium Priority

| Enhancement | Description | Status |
|-------------|-------------|--------|
| **Global Source URL Index** | GSI on cr8_tracks for source_url lookups | Planned |
| **Title Normalization Index** | Pre-compute normalized titles in DynamoDB | Planned |
| **CloudWatch Metrics** | Publish sync metrics for dashboards | Planned |
| **Failure Retry Queue** | Re-attempt failed downloads with backoff | Planned |
| **DynamoDB scan→query** | Use partition key queries instead of scans | ✅ Done |
| **Beatport Metadata Enrichment** | Extended track metadata from Beatport API | ✅ Done |
| **Label GSI** | Query tracks by record label | ✅ Done |

### Low Priority

| Enhancement | Description | Status |
|-------------|-------------|--------|
| **Geo-Proxy Support** | VPN/proxy for geo-restricted tracks | Planned |
| **Audio Fingerprinting** | Dedupe by audio hash, not just title | Planned |
| **Playlist Diff Alerts** | Notify when source playlist changes | Planned |
