# AFTRS-Void Compatible MCP Servers

This document catalogs Model Context Protocol (MCP) servers that are compatible with and valuable for the AFTRS-Void ecosystem, organized by category and priority level.

## 🎯 High Priority MCP Servers

### Infrastructure & DevOps
| Server                    | Use Case                                     | AFTRS Projects                           | Install Command                               |
| ------------------------- | -------------------------------------------- | ---------------------------------------- | --------------------------------------------- |
| **GitHub MCP Server**     | Repository management, CI/CD, issue tracking | All projects                             | `npx @modelcontextprotocol/server-github`     |
| **Docker MCP Server**     | Container management, image building         | aftrs-code-llm-plan, cr8_cli, open-webui | `npx @modelcontextprotocol/server-docker`     |
| **AWS MCP Server**        | Cloud infrastructure management              | All cloud deployments                    | `npx @modelcontextprotocol/server-aws`        |
| **Prometheus MCP Server** | Metrics collection and monitoring            | aftrs-code-llm-plan, monitoring stacks   | `npx @modelcontextprotocol/server-prometheus` |
| **Tailscale MCP Server**  | VPN management, network topology             | aftrs_cli, network projects              | `npx @modelcontextprotocol/server-tailscale`  |

### Database & Storage
| Server                    | Use Case                             | AFTRS Projects                | Install Command                               |
| ------------------------- | ------------------------------------ | ----------------------------- | --------------------------------------------- |
| **PostgreSQL MCP Server** | Database operations, user management | cr8_cli, aftrs-code-llm-plan  | `npx @modelcontextprotocol/server-postgresql` |
| **SQLite MCP Server**     | Local database management            | cr8_cli, aftrs_cli assets     | `npx @modelcontextprotocol/server-sqlite`     |
| **S3 MCP Server**         | Object storage, file management      | Backup systems, media storage | `npx @modelcontextprotocol/server-s3`         |

### Development Tools
| Server                         | Use Case                            | AFTRS Projects                          | Install Command                                    |
| ------------------------------ | ----------------------------------- | --------------------------------------- | -------------------------------------------------- |
| **Notion MCP Server**          | Knowledge management, documentation | Project planning, team collaboration    | `npx @modelcontextprotocol/server-notion`          |
| **Google Calendar MCP Server** | Scheduling, event management        | Project planning, maintenance schedules | `npx @modelcontextprotocol/server-google-calendar` |
| **Slack MCP Server**           | Team communication, alerts          | Monitoring alerts, team coordination    | `npx @modelcontextprotocol/server-slack`           |

## 🚀 Medium Priority MCP Servers

### Media & Content Processing
| Server                      | Use Case                            | AFTRS Projects             | Install Command                               |
| --------------------------- | ----------------------------------- | -------------------------- | --------------------------------------------- |
| **YouTube MCP Server**      | Video content management            | cr8_cli, media processing  | `npx @modelcontextprotocol/server-youtube`    |
| **SoundCloud MCP Server**   | Audio content, metadata extraction  | cr8_cli DJ workflows       | `npx @modelcontextprotocol/server-soundcloud` |
| **Spotify MCP Server**      | Music metadata, playlist management | cr8_cli music intelligence | `npx @modelcontextprotocol/server-spotify`    |
| **Google Drive MCP Server** | File storage, sync management       | cr8_cli sync operations    | `npx @modelcontextprotocol/server-gdrive`     |

### Network & Security
| Server                       | Use Case                     | AFTRS Projects            | Install Command                                |
| ---------------------------- | ---------------------------- | ------------------------- | ---------------------------------------------- |
| **Cloudflare MCP Server**    | DNS management, security     | Domain management, CDN    | `npx @modelcontextprotocol/server-cloudflare`  |
| **Let's Encrypt MCP Server** | SSL certificate management   | HTTPS automation          | `npx @modelcontextprotocol/server-letsencrypt` |
| **OpenVPN MCP Server**       | VPN configuration management | Network security projects | `npx @modelcontextprotocol/server-openvpn`     |

### System Administration
| Server                 | Use Case                           | AFTRS Projects            | Install Command                            |
| ---------------------- | ---------------------------------- | ------------------------- | ------------------------------------------ |
| **systemd MCP Server** | Service management, system control | Linux service automation  | `npx @modelcontextprotocol/server-systemd` |
| **Nginx MCP Server**   | Web server configuration           | Reverse proxy management  | `npx @modelcontextprotocol/server-nginx`   |
| **Caddy MCP Server**   | Modern web server management       | aftrs-code-llm-plan proxy | `npx @modelcontextprotocol/server-caddy`   |

## 💡 Specialized MCP Servers

### AFTRS-Specific Use Cases
| Server                    | Use Case                              | AFTRS Projects                 | Install Command                               |
| ------------------------- | ------------------------------------- | ------------------------------ | --------------------------------------------- |
| **Grafana MCP Server**    | Dashboard management, visualization   | aftrs-code-llm-plan monitoring | `npx @modelcontextprotocol/server-grafana`    |
| **Discord MCP Server**    | Community management, bot integration | Community projects             | `npx @modelcontextprotocol/server-discord`    |
| **Jira MCP Server**       | Project management, issue tracking    | Enterprise project management  | `npx @modelcontextprotocol/server-jira`       |
| **Confluence MCP Server** | Documentation management              | Enterprise documentation       | `npx @modelcontextprotocol/server-confluence` |

### Hardware & IoT Integration
| Server                        | Use Case                             | AFTRS Projects                         | Install Command                                  |
| ----------------------------- | ------------------------------------ | -------------------------------------- | ------------------------------------------------ |
| **MQTT MCP Server**           | IoT device communication             | aftrs-lighting, hallway-light-project  | `npx @modelcontextprotocol/server-mqtt`          |
| **Home Assistant MCP Server** | Smart home automation                | Lighting and automation projects       | `npx @modelcontextprotocol/server-homeassistant` |
| **Unraid MCP Server**         | Server management, container control | unraid-monolith, server administration | `npx @modelcontextprotocol/server-unraid`        |

### Development & CI/CD
| Server                    | Use Case                               | AFTRS Projects        | Install Command                               |
| ------------------------- | -------------------------------------- | --------------------- | --------------------------------------------- |
| **GitLab MCP Server**     | CI/CD pipelines, repository management | Alternative to GitHub | `npx @modelcontextprotocol/server-gitlab`     |
| **Jenkins MCP Server**    | Build automation, deployment           | CI/CD pipelines       | `npx @modelcontextprotocol/server-jenkins`    |
| **Kubernetes MCP Server** | Container orchestration                | Cloud deployments     | `npx @modelcontextprotocol/server-kubernetes` |

## 🔧 Custom MCP Server Opportunities

### AFTRS-Void Specific Servers
| Potential Server        | Use Case                                 | Implementation Priority             |
| ----------------------- | ---------------------------------------- | ----------------------------------- |
| **OPNsense MCP Server** | Router/firewall configuration management | High - Custom implementation needed |
| **Beatport MCP Server** | Professional DJ metadata lookup          | High - cr8_cli integration          |
| **yt-dlp MCP Server**   | Advanced media downloading               | Medium - cr8_cli enhancement        |
| **rclone MCP Server**   | Cloud storage synchronization            | Medium - Multiple project use       |
| **FFmpeg MCP Server**   | Audio/video processing automation        | Medium - Media projects             |

### Network Management Servers
| Potential Server           | Use Case                       | Implementation Priority         |
| -------------------------- | ------------------------------ | ------------------------------- |
| **DHCP Management Server** | Dynamic IP allocation          | Medium - aftrs_cli integration  |
| **DNS Management Server**  | DNS record management          | Medium - Network administration |
| **Network Scanner Server** | Asset discovery and monitoring | Low - aftrs_cli enhancement     |

## 📊 Server Priority Matrix

### Immediate Implementation (Next 30 days)
1. **GitHub MCP Server** - Repository management for all projects
2. **Docker MCP Server** - Container management for deployed services
3. **PostgreSQL MCP Server** - Database operations for cr8_cli and monitoring
4. **Slack MCP Server** - Alert notifications and team communication
5. **Notion MCP Server** - Knowledge management and documentation

### Short-term Implementation (30-90 days)
1. **AWS/S3 MCP Server** - Cloud storage and backup management
2. **Prometheus MCP Server** - Enhanced monitoring integration
3. **YouTube/SoundCloud Servers** - Media processing enhancement
4. **Tailscale MCP Server** - Network management automation
5. **Google Calendar MCP Server** - Scheduling and maintenance planning

### Long-term Implementation (3-6 months)
1. **Custom OPNsense Server** - Router management automation
2. **Custom Beatport Server** - Professional DJ metadata
3. **Kubernetes MCP Server** - Container orchestration
4. **Home Assistant MCP Server** - IoT and lighting automation
5. **Custom yt-dlp/rclone Servers** - Enhanced media workflows

## 🎯 Project-Specific MCP Server Mapping

### AFTRS CLI Projects
**Primary Servers Needed:**
- GitHub MCP Server (repository management)
- Tailscale MCP Server (network topology management)
- SQLite MCP Server (asset database management)
- Prometheus MCP Server (monitoring integration)
- Custom OPNsense Server (router configuration)

**Use Cases:**
- Automated network asset discovery and documentation
- Real-time monitoring and alerting for network changes
- Router configuration management and backup
- Tailscale network topology visualization

### CR8 CLI Projects
**Primary Servers Needed:**
- YouTube/SoundCloud MCP Servers (media API integration)
- Spotify MCP Server (metadata enhancement)
- Google Drive MCP Server (sync management)
- PostgreSQL MCP Server (database operations)
- Custom Beatport Server (professional DJ metadata)

**Use Cases:**
- Professional DJ metadata enhancement workflows
- Automated playlist sync and quality assurance
- Advanced audio analysis and BPM detection
- Multi-source metadata validation

### AFTRS Code LLM Plan
**Primary Servers Needed:**
- Docker MCP Server (container management)
- Prometheus MCP Server (metrics collection)
- Grafana MCP Server (dashboard management)
- Slack MCP Server (alert notifications)
- AWS/S3 MCP Server (log backup and storage)

**Use Cases:**
- Real-time LLM cost monitoring and alerting
- Container orchestration and health monitoring
- Automated dashboard provisioning
- Budget management and usage analytics

### Unraid/Server Management Projects
**Primary Servers Needed:**
- Unraid MCP Server (server management)
- Docker MCP Server (container orchestration)
- systemd MCP Server (service management)
- Nginx/Caddy MCP Servers (reverse proxy management)
- Home Assistant MCP Server (automation)

**Use Cases:**
- Automated server provisioning and maintenance
- Container lifecycle management
- Service monitoring and restart automation
- Home automation integration

### Network/OpnSense Projects
**Primary Servers Needed:**
- Custom OPNsense Server (firewall management)
- DHCP Management Server (IP allocation)
- DNS Management Server (record management)
- Let's Encrypt Server (SSL automation)
- Network Scanner Server (asset discovery)

**Use Cases:**
- Automated firewall rule management
- Dynamic DNS and DHCP configuration
- SSL certificate lifecycle management
- Network security monitoring

## 🚀 Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)
**Goal:** Establish core development and infrastructure capabilities

**Priority Servers:**
1. **GitHub MCP Server**
   - Install: `npx @modelcontextprotocol/server-github`
   - Configure: Add to Cline MCP settings
   - Use Cases: Repository management across all projects

2. **Docker MCP Server**
   - Install: `npx @modelcontextprotocol/server-docker`
   - Configure: Connect to local Docker daemon
   - Use Cases: Container management for all deployed services

3. **Slack MCP Server**
   - Install: `npx @modelcontextprotocol/server-slack`
   - Configure: Bot token and channels
   - Use Cases: Notification system for all monitoring alerts

4. **Notion MCP Server**
   - Install: `npx @modelcontextprotocol/server-notion`
   - Configure: API keys and database access
   - Use Cases: Knowledge management and documentation

**Expected Outcomes:**
- Streamlined repository management across all AFTRS projects
- Centralized container management for deployed services
- Unified notification system for monitoring and alerts
- Enhanced documentation and knowledge management

### Phase 2: Data & Monitoring (Weeks 5-8)
**Goal:** Enhance data management and monitoring capabilities

**Priority Servers:**
1. **PostgreSQL MCP Server**
   - Install: `npx @modelcontextprotocol/server-postgresql`
   - Configure: Database connections for cr8_cli and monitoring
   - Use Cases: Enhanced database operations and user management

2. **Prometheus MCP Server**
   - Install: `npx @modelcontextprotocol/server-prometheus`
   - Configure: Metrics collection endpoints
   - Use Cases: Advanced monitoring integration

3. **SQLite MCP Server**
   - Install: `npx @modelcontextprotocol/server-sqlite`
   - Configure: Local databases for aftrs_cli assets
   - Use Cases: Asset management and caching

4. **S3 MCP Server**
   - Install: `npx @modelcontextprotocol/server-s3`
   - Configure: AWS credentials and buckets
   - Use Cases: Backup automation and log storage

**Expected Outcomes:**
- Enhanced database management for DJ and monitoring applications
- Advanced metrics collection and analysis capabilities
- Improved asset management for network infrastructure
- Automated backup and storage management

### Phase 3: Media & Automation (Weeks 9-12)
**Goal:** Implement media processing and network automation

**Priority Servers:**
1. **YouTube MCP Server**
   - Install: `npx @modelcontextprotocol/server-youtube`
   - Configure: API keys and quotas
   - Use Cases: Enhanced cr8_cli media processing

2. **Google Calendar MCP Server**
   - Install: `npx @modelcontextprotocol/server-google-calendar`
   - Configure: Service accounts and permissions
   - Use Cases: Maintenance scheduling and project planning

3. **Tailscale MCP Server**
   - Install: `npx @modelcontextprotocol/server-tailscale`
   - Configure: API keys and network access
   - Use Cases: Network topology management

4. **Grafana MCP Server**
   - Install: `npx @modelcontextprotocol/server-grafana`
   - Configure: Dashboard API access
   - Use Cases: Advanced visualization management

**Expected Outcomes:**
- Streamlined media processing workflows for DJ applications
- Automated maintenance scheduling and project coordination
- Enhanced network management and topology visualization
- Professional dashboard management and customization

### Phase 4: Custom Solutions (Weeks 13-16)
**Goal:** Implement custom MCP servers for AFTRS-specific needs

**Custom Development:**
1. **OPNsense MCP Server**
   - Develop: Custom server for router management
   - Features: Configuration backup, rule management, monitoring
   - Integration: AFTRS CLI network management

2. **Beatport MCP Server**
   - Develop: Professional DJ metadata lookup
   - Features: Track identification, BPM/key analysis, genre classification
   - Integration: CR8 CLI metadata enhancement

3. **yt-dlp MCP Server**
   - Develop: Advanced media downloading server
   - Features: Quality optimization, batch processing, progress tracking
   - Integration: CR8 CLI download workflows

**Expected Outcomes:**
- Professional router management capabilities
- Industry-standard DJ metadata enhancement
- Advanced media processing automation

## 🔧 Technical Implementation Notes

### MCP Server Configuration
All MCP servers are configured in the Cline settings file (`~/.config/cline/settings.json`):

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "your-token-here"
      }
    },
    "docker": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-docker"]
    },
    "postgresql": {
      "command": "npx",
      "args": ["@modelcontextprotocol/server-postgresql"],
      "env": {
        "POSTGRES_CONNECTION_STRING": "postgresql://user:pass@host:port/db"
      }
    }
  }
}
```

### Security Considerations
- All API keys and sensitive credentials stored in environment variables
- MCP servers run with minimal required permissions
- Network access restricted to necessary endpoints only
- Regular security audits and credential rotation

### Performance Optimization
- MCP servers configured with appropriate timeouts and retry logic
- Resource usage monitoring for all deployed servers
- Caching strategies implemented where applicable
- Load balancing for high-usage scenarios

### Monitoring & Observability
- All MCP server interactions logged for debugging
- Performance metrics collected for optimization
- Error tracking and alerting configured
- Regular health checks and status monitoring

## 📚 Additional Resources

### Documentation Links
- [MCP Official Documentation](https://modelcontextprotocol.io/)
- [Cline MCP Integration Guide](https://docs.cline.bot/mcp)
- [AFTRS Development Principles](../PRINCIPLES.md)

### Community Resources
- [MCP Server Registry](https://github.com/modelcontextprotocol/servers)
- [Cline Community Discord](https://discord.gg/cline)
- [AFTRS Void GitHub Organization](https://github.com/aftrs-void)

---

**Document Version:** 1.0
**Last Updated:** September 23, 2025
**Next Review:** October 23, 2025
**Maintained By:** AFTRS Development Team
