# AFTRS MCP Architecture

## Overview

Centralized Model Context Protocol (MCP) server aggregating tools from across the AFTRS ecosystem, providing a unified interface for AI assistants to interact with infrastructure and creative projects.

## Planned Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      AFTRS MCP Server                           │
├─────────────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    MCP Protocol Layer                     │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │  │
│  │  │  stdio   │ │   SSE    │ │   HTTP   │ │  WebSocket   │  │  │
│  │  │Transport │ │Transport │ │Transport │ │  Transport   │  │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────────┘  │  │
│  └───────────────────────────────────────────────────────────┘  │
│                          │                                      │
│  ┌───────────────────────▼───────────────────────────────────┐  │
│  │                   Tool Registry                           │  │
│  │  ┌──────────────────────────────────────────────────────┐ │  │
│  │  │  Tool Registration │ Discovery │ Routing │ Auth      │ │  │
│  │  └──────────────────────────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────┘  │
│                          │                                      │
│  ┌───────────────────────▼───────────────────────────────────┐  │
│  │                   Tool Categories                         │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐  │  │
│  │  │ Infra    │ │ Creative │ │ Network  │ │   DevOps     │  │  │
│  │  │ Tools    │ │  Tools   │ │  Tools   │ │   Tools      │  │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────────┘  │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  UNRAID Tools   │ │  OPNsense Tools │ │   CR8 Tools     │
│  (from monolith)│ │ (from monolith) │ │ (from cr8-cli)  │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

## Planned Components

### Transport Layer
- **stdio** - Primary transport for local tools
- **SSE** - Server-Sent Events for streaming
- **HTTP** - REST API for remote access
- **WebSocket** - Real-time bidirectional communication

### Tool Registry
- Dynamic tool registration
- Capability discovery
- Smart routing based on tool type
- Authentication and authorization

### Tool Categories

| Category | Source | Tools |
|----------|--------|-------|
| Infrastructure | unraid-monolith | Docker, VM, disk, backup management |
| Network | opnsense-monolith | Firewall, NAT, VLAN, diagnostics |
| Creative | cr8-cli | Music sync, audio processing, playlists |
| Visual | visual-projects | TouchDesigner, Resolume, shaders |
| Gaming | gaming-projects | Console homebrew, emulation |
| Studio | studio-projects | MIDI, lighting, NDI |

## Technology Stack

| Layer | Technology |
|-------|------------|
| Language | Python 3.11+ |
| MCP Framework | mcp-python |
| API | FastAPI (optional) |
| Config | YAML, TOML |
| Testing | pytest |
| CI/CD | GitHub Actions |

## Integration Points

### Existing MCP Servers to Consolidate
- cr8-cli MCP tools (300+ tools)
- unraid-monolith LLM agent
- opnsense-monolith MCP plugin

### AI Assistants
- Claude Code
- Cline
- Continue
- Custom agents
