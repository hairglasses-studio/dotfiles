# hairglasses-studio GitHub Organization Analysis

*Research conducted: December 23, 2025*

## Organization Overview

| Attribute | Value |
|-----------|-------|
| Name | Aftrs Studio Audio & Visual Projects |
| Created | December 23, 2025 |
| Total Repos | 79 (public + private) |
| Location | United States |

---

## Recently Migrated Repositories (Dec 23, 2025)

21 project-category repositories were migrated into the organization today:

### Core Infrastructure
| Repo | Created | Description |
|------|---------|-------------|
| **hg-mcp** | 8:59 AM | Centralized MCP server for studio operations |
| **aftrs-terraform** | 9:00 AM | Terraform infrastructure code |

### AV/Live Performance (7:14-7:28 AM)
| Repo | Focus |
|------|-------|
| **touchdesigner-projects** | TouchDesigner visual programming |
| **resolume-projects** | Resolume VJ/live performance |
| **ffgl-shader-projects** | FFGL shader effects library |
| **lighting-projects** | Lighting control and automation |
| **projection-mapping-projects** | Projection mapping technology |
| **midi-mapping-projects** | MIDI controller mapping |
| **dj-projects** | DJ tools and automation |

### Streaming & Video (7:17-7:28 AM)
| Repo | Focus |
|------|-------|
| **aftrs-live-streaming-tools** | Live streaming capabilities |
| **ndi-projects** | NDI (Network Device Interface) |
| **video-processing-projects** | General video processing |
| **video-woodchipper** | Video processing/manipulation |
| **ai-art-processing-projects** | AI-powered art processing |

### Retro Gaming/Emulation (7:19-7:25 AM)
| Repo | Focus |
|------|-------|
| **emulator-projects** | Emulator development/config |
| **ps2-projects** | PlayStation 2 projects |
| **ps4-projects** | PlayStation 4 projects |
| **dreamcast-projects** | Dreamcast console |
| **xbox-projects** | Xbox-related projects |
| **studio-automation-projects** | General automation |

---

## Pre-Existing Repositories (58 repos)

### Key Active Projects
| Repo | Created | Description |
|------|---------|-------------|
| **ebay-tool** | Nov 19, 2025 | eBay price tracker (Flask, PostgreSQL, Redis, AWS ECS) |
| **disk-price-tracker** | Oct 18, 2025 | SSD/disk price monitoring |
| **steam-deck-main** | Oct 8, 2025 | Steam Deck projects |
| **t2-mac-linux** | Oct 8, 2025 | Mac Linux tools |

### UNRAID Ecosystem (8+ repos)
| Repo | Purpose |
|------|---------|
| **unraid-monolith** | Comprehensive UNRAID config |
| **unraid-observability-stack** | Monitoring/observability |
| **unraid_llm_agent** | AI agent for UNRAID |
| **llmagent-unraid** | LLM agent tools |
| **unraid_scripts** | Automation scripts |
| **unraid_snapshot_exporter** | Backup/export tools |
| **tailscale_unraid_plugin** | Tailscale integration |
| **ledfx-companion** | LED effects control |

### PS2 Ecosystem (20+ repos)
The organization has extensive PlayStation 2 infrastructure:

| Repo | Purpose |
|------|---------|
| **ps2_cli** | PS2 CLI build tool |
| **ps2-base-image** | Base images |
| **ps2-test-image** | Test images |
| **ps2-all-images** | All PS2 images |
| **ps2-linux-modern-starter** | PS2 Linux starter |
| **ps2-homebrew-modern** | Modern homebrew |
| **ps2-homebrew-starter** | Homebrew starter |
| **pcsx2_scaffold** | PCSX2 emulator scaffolding |
| **ps2_leaked_source_code_repos** | Source code archive |
| **ps2_sdks_internal_collection** | SDK collection |
| **mcp2_firmware_and_memcard_image_archive** | Firmware/memory cards |
| **ps2_all** | PS2 utilities |
| **ps2-ndi** | PS2 NDI streaming |
| **ps2mc** | Memory card tools |
| **hg_ps2_bootstrap** | PS2 dotfiles config |
| **hairglasses_ps2_visualizer_classic** | Classic visualizer |
| **hairglasses_ps2_visualizer_modern** | Modern visualizer |

### Infrastructure & Automation
| Repo | Purpose |
|------|---------|
| **agentctl** | Agent control tool |
| **aftrs-cli** | Aftrs Studio CLI |
| **tailscale-acl** | Tailscale access control |
| **opnsense-monolith** | OPNsense networking |
| **aftrs-nordvpn-qbittorrent** | VPN + torrent integration |

### Obsidian & Documentation
| Repo | Purpose |
|------|---------|
| **aftrs_wiki** | Wiki/knowledge base |
| **aftrs-obsidianmd** | Obsidian integration |
| **aftrs-code-llm-plan** | LLM planning |

### Visualization & Graphics
| Repo | Purpose |
|------|---------|
| **aura** | Visualization/graphics |
| **hg_aura** | Aura tools |
| **helios-glitch-lab** | Glitch art/effects |

### Development & Deployment
| Repo | Purpose |
|------|---------|
| **langgraph-deployment-plan** | LangGraph deployment |
| **resolume-plugin-test-1** | Resolume plugin dev |
| **annex-network-agent** | Network agent |
| **aftrs-cline** | Cline AI integration |

### Utilities
| Repo | Purpose |
|------|---------|
| **open-webui** | Web UI framework |
| **graphplz** | Graph visualization |
| **cbonsai** | ASCII tree visualization |
| **dotfiles** | System configuration |
| **aftrs-shell** | Shell utilities |
| **retro-session-orchestrator** | Retro gaming orchestration |
| **mcp2-toolbox** | MCP tools |
| **ps2debug** | PS2 debugging |
| **org-admin** | Organization admin |
| **.github** | GitHub org config |

---

## Domain Analysis

### Domain 1: AV/Live Performance
**Repos:** 9 | **Key Tech:** TouchDesigner, Resolume, FFGL, MIDI, DMX

Key integration points:
- TouchDesigner MCP server already exists (external)
- Resolume uses OSC protocol
- FFGL shaders can be loaded in both TD and Resolume
- MIDI maps to all AV software

### Domain 2: Streaming & Video
**Repos:** 5 | **Key Tech:** NDI, video processing, AI art

Key integration points:
- NDI for inter-application video routing
- Video-woodchipper for batch processing
- AI art processing for generative content

### Domain 3: Studio Automation
**Repos:** 15+ | **Key Tech:** UNRAID, Tailscale, automation scripts

Key integration points:
- UNRAID API for storage/VM management
- Tailscale for secure remote access
- Automation scripts for scheduled tasks

### Domain 4: Retro Gaming/PS2
**Repos:** 20+ | **Key Tech:** PCSX2, OPL, memory cards, visualizers

Key integration points:
- PCSX2 for PS2 emulation
- Memory card management tools
- Audio visualizers for streaming
- PS2 NDI streaming capability

---

## Activity Patterns

| Period | Activity |
|--------|----------|
| Today (Dec 23) | 21 repos migrated - major organization |
| Nov 2025 | ebay-tool active development |
| Oct 2025 | disk-price-tracker, Steam Deck focus |
| Jul-Sep 2025 | Bulk of repo creation |
| Aug 2025 | Organization admin setup |

---

## Key Insights

1. **Creative Studio Focus** - Clear AV/live performance orientation
2. **Heavy PS2 Investment** - 20+ repos indicate significant retro gaming focus
3. **Strong UNRAID Foundation** - 8+ repos for server infrastructure
4. **LLM Integration Interest** - Multiple agent/LLM repos
5. **Recent Organization** - Today's migration indicates consolidation effort
6. **MCP Foundation** - hg-mcp positioned as central control plane

---

## Recommendations for hg-mcp Integration

### Priority 1: Existing Infrastructure
- UNRAID (8+ repos already exist)
- PS2 ecosystem (20+ repos with tooling)
- Tailscale (ACL + plugin repos)

### Priority 2: AV Software
- TouchDesigner (wrap existing MCP)
- Resolume (implement OSC client)
- FFGL (shader loading/management)

### Priority 3: Streaming
- NDI (discovery and routing)
- Video processing (woodchipper integration)

### Priority 4: Cross-Cutting
- Obsidian vault (knowledge graph)
- Automation scripts (workflow execution)
