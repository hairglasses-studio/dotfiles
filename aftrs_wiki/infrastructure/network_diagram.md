# Network Diagram - AFTRS Infrastructure

## Overview
This document provides a comprehensive network diagram of the AFTRS infrastructure, including physical hardware, cloud services, VPN connectivity, and software networks. It serves as a reference for understanding the complete network topology.

## 🏗️ Physical Network Topology

### Core Infrastructure Layout
```
Internet (Comcast 1Gbps/35Mbps)
    ↓
Arris SB8200 Modem (Bridge Mode)
    ↓
OPNSense Router (aftrs-main)
├── Hardware: Intel i3-6100, 8GB RAM
├── OS: OPNSense 23.1+
├── WAN: DHCP from Comcast
├── LAN: 192.168.1.1/24
└── Features: Firewall, NAT, DHCP, DNS
    ↓
Managed Switch (2.5Gbps)
    ↓
┌─────────────────┬─────────────────┬─────────────────┬─────────────────┐
│   UNRAID Server │ Development PC  │ Mobile Devices  │ IoT Devices     │
│   (192.168.1.10)│   (192.168.1.20)│   (192.168.1.30)│   (192.168.1.40)│
│   aftrs-unraid  │   aftrs-dev     │   Mobile        │   IoT           │
└─────────────────┴─────────────────┴─────────────────┴─────────────────┘
```

### Network Segments
```
Network Segments:
├── WAN: Comcast Internet (1Gbps/35Mbps)
├── LAN: 192.168.1.0/24 (Primary Network)
├── DMZ: 192.168.2.0/24 (Isolated Services)
├── IoT: 192.168.3.0/24 (IoT Devices)
└── Guest: 192.168.4.0/24 (Guest Network)
```

## 🌐 Internet Connectivity

### Comcast Modem Configuration
```
Arris SB8200 (DOCSIS 3.1)
├── Mode: Bridge Mode (No NAT)
├── IP: DHCP from Comcast
├── Port: 1Gbps Ethernet to OPNSense
├── Features: DOCSIS 3.1, IPv6 support
└── Status: Operational
```

### OPNSense Router Configuration
```
aftrs-main (OPNSense Router)
├── Hardware: Intel i3-6100, 8GB RAM, 2x NIC
├── OS: OPNSense 23.1+
├── WAN Interface: DHCP from Comcast
├── LAN Interface: 192.168.1.1/24
├── Services:
│   ├── Firewall (pfSense)
│   ├── NAT (Port Forwarding)
│   ├── DHCP Server
│   ├── DNS (Unbound)
│   ├── VPN (Tailscale)
│   └── Monitoring (Netflow)
└── Status: Primary Gateway
```

## 🐛 Tailscale VPN Network

### Tailscale Network: aftrs.void
```
Tailscale Network Topology:
aftrs.void Network
├── aftrs-main (Exit Node)
│   ├── IP: 100.64.0.1
│   ├── Role: Router/Gateway
│   ├── Routes: 192.168.1.0/24
│   ├── Services: Firewall, NAT, DNS
│   └── Status: Exit Node Active
├── aftrs-unraid (File Server)
│   ├── IP: 100.64.0.10
│   ├── Role: Storage/Media Server
│   ├── Services: SMB, NFS, Docker, VMs
│   ├── Storage: 4x 8TB HDD + 2x 1TB NVMe
│   └── Status: File Server Active
├── aftrs-dev (Development)
│   ├── IP: 100.64.0.20
│   ├── Role: Development Workstation
│   ├── Services: SSH, Git, Development Tools
│   ├── OS: Ubuntu 22.04 LTS
│   └── Status: Development Active
└── Mobile Devices
    ├── IP: 100.64.0.30+
    ├── Role: Remote Access
    ├── Services: SSH, Web Access
    └── Status: On-demand
```

### Tailscale ACL Configuration
```json
{
  "tagOwners": {
    "tag:admin": ["hairglasses@gmail.com"],
    "tag:server": ["hairglasses@gmail.com"],
    "tag:dev": ["hairglasses@gmail.com"],
    "tag:mobile": ["hairglasses@gmail.com"]
  },
  "acls": [
    {
      "action": "accept",
      "src": ["tag:admin"],
      "dst": ["*:*"]
    },
    {
      "action": "accept",
      "src": ["tag:server"],
      "dst": ["aftrs-main:22", "aftrs-unraid:22", "aftrs-dev:22"]
    },
    {
      "action": "accept",
      "src": ["tag:dev"],
      "dst": ["aftrs-dev:*", "aftrs-unraid:22"]
    },
    {
      "action": "accept",
      "src": ["tag:mobile"],
      "dst": ["aftrs-main:80", "aftrs-main:443", "aftrs-unraid:80"]
    }
  ]
}
```

## 🖥️ Server Infrastructure

### UNRAID Server (aftrs-unraid)
```
Hardware Configuration:
├── CPU: AMD Ryzen 7 3700X (8-core, 16-thread)
├── RAM: 32GB DDR4-3200
├── Storage:
│   ├── Cache: 2x 1TB NVMe SSD
│   ├── Array: 4x 8TB HDD (Parity)
│   └── Total: 24TB usable + 2TB cache
├── Network: 2.5Gbps Ethernet
├── GPU: NVIDIA RTX 3060 (GPU Passthrough)
└── OS: UNRAID 6.9+

Services Running:
├── Docker Containers:
│   ├── AgentCTL Wiki (Port 8080)
│   ├── CR8 CLI Web Interface (Port 5000)
│   ├── OAI CLI Dashboard (Port 8081)
│   ├── PostgreSQL Database (Port 5432)
│   └── Various Media Services
├── Virtual Machines:
│   ├── Ubuntu Development VM
│   ├── Windows Gaming VM (GPU Passthrough)
│   └── Testing VMs
├── File Sharing:
│   ├── SMB Shares (Windows/Mac)
│   ├── NFS Shares (Linux)
│   └── WebDAV (Web Access)
└── Backup Services:
    ├── Automated Backups
    ├── Cloud Sync (Google Drive)
    └── Version Control
```

### Development PC (aftrs-dev)
```
Hardware Configuration:
├── CPU: Intel i7-10700K (8-core, 16-thread)
├── RAM: 32GB DDR4-3200
├── Storage: 2x 1TB NVMe SSD
├── Network: 1Gbps Ethernet + WiFi 6
├── GPU: NVIDIA RTX 3070
└── OS: Ubuntu 22.04 LTS

Development Environment:
├── Development Tools:
│   ├── Git (Version Control)
│   ├── Docker (Containerization)
│   ├── Python 3.11+ (Programming)
│   ├── Node.js (Web Development)
│   └── Various IDEs and Editors
├── CLI Tools:
│   ├── AFTRS CLI (Network Management)
│   ├── TSCTL (Tailscale Management)
│   ├── CR8 CLI (Media Processing)
│   ├── OAI CLI (Content Archiving)
│   └── AgentCTL (Wiki System)
├── Creative Tools:
│   ├── Blender (3D Modeling)
│   ├── Unreal Engine (Game Development)
│   ├── DaVinci Resolve (Video Editing)
│   ├── Ableton Live (Audio Production)
│   └── Various Audio/Visual Tools
└── Network Access:
    ├── SSH to all servers
    ├── SMB/NFS file access
    ├── Web interface access
    └── Remote development
```

## 🔧 Network Services

### DNS Configuration
```
Primary DNS: OPNSense (192.168.1.1)
Secondary DNS: 1.1.1.1, 8.8.8.8
Local Domain: aftrs.void

DNS Overrides:
├── aftrs-main.aftrs.void → 192.168.1.1
├── aftrs-unraid.aftrs.void → 192.168.1.10
├── aftrs-dev.aftrs.void → 192.168.1.20
├── *.aftrs.void → 192.168.1.1 (Catch-all)
└── Various service-specific overrides
```

### DHCP Configuration
```
DHCP Server: OPNSense
Range: 192.168.1.100-192.168.1.200
Lease Time: 24 hours
Gateway: 192.168.1.1
DNS: 192.168.1.1

Static Leases:
├── aftrs-unraid: 192.168.1.10 (MAC: 00:11:22:33:44:55)
├── aftrs-dev: 192.168.1.20 (MAC: 00:11:22:33:44:66)
├── Mobile Devices: 192.168.1.30+ (DHCP)
└── IoT Devices: 192.168.1.40+ (DHCP)
```

### Firewall Zones
```
Firewall Configuration:
├── WAN Zone (Internet)
│   ├── Default: Block all
│   ├── Allow: Established connections
│   └── NAT: Port forwarding rules
├── LAN Zone (Trusted)
│   ├── Default: Allow all
│   ├── Services: All internal services
│   └── Monitoring: Traffic analysis
├── DMZ Zone (Semi-Trusted)
│   ├── Default: Allow specific
│   ├── Services: Web servers, APIs
│   └── Isolation: Limited LAN access
├── IoT Zone (Restricted)
│   ├── Default: Allow specific
│   ├── Services: IoT device communication
│   └── Isolation: No LAN access
└── Guest Zone (Isolated)
    ├── Default: Allow specific
    ├── Services: Internet access only
    └── Isolation: No internal access
```

## 📡 Wireless Network

### WiFi Configuration
```
Primary WiFi: aftrs-network (5GHz)
├── SSID: aftrs-network
├── Security: WPA3-Personal
├── Band: 5GHz (802.11ac)
├── Channel: Auto-optimized
└── Range: Full property coverage

Guest WiFi: aftrs-guest (2.4GHz)
├── SSID: aftrs-guest
├── Security: WPA2-Personal
├── Band: 2.4GHz (802.11n)
├── Isolation: Guest network only
└── Access: Internet only
```

## 🔒 Security Architecture

### Network Security
```
Security Layers:
├── Physical Security
│   ├── Router location: Secure area
│   ├── Server location: Climate controlled
│   └── Cable management: Organized
├── Network Security
│   ├── Firewall: OPNSense with rules
│   ├── VPN: Tailscale for remote access
│   ├── VLANs: Network segmentation
│   └── Monitoring: Traffic analysis
├── Application Security
│   ├── HTTPS: SSL/TLS encryption
│   ├── Authentication: Multi-factor
│   ├── Updates: Regular security patches
│   └── Backups: Encrypted backups
└── Data Security
    ├── Encryption: At rest and in transit
    ├── Access Control: Role-based
    ├── Audit Logs: Comprehensive logging
    └── Incident Response: Automated alerts
```

### Access Control
```
User Access Levels:
├── Admin (hairglasses@gmail.com)
│   ├── Full system access
│   ├── Configuration management
│   ├── User management
│   └── Security administration
├── Developer
│   ├── Development environment access
│   ├── Code repository access
│   ├── Testing environment access
│   └── Limited production access
├── User
│   ├── File sharing access
│   ├── Media streaming access
│   ├── Web service access
│   └── Basic system access
└── Guest
    ├── Internet access only
    ├── Time-limited access
    ├── Rate-limited bandwidth
    └── No internal access
```

## 📊 Monitoring and Analytics

### Network Monitoring
```
Monitoring Systems:
├── OPNSense Monitoring
│   ├── Traffic analysis
│   ├── Connection tracking
│   ├── Performance metrics
│   └── Security alerts
├── Tailscale Monitoring
│   ├── Network status
│   ├── Connection health
│   ├── Route verification
│   └── ACL compliance
├── Server Monitoring
│   ├── System resources
│   ├── Service health
│   ├── Storage usage
│   └── Performance metrics
└── Application Monitoring
    ├── Web service health
    ├── Database performance
    ├── Container status
    └── Error tracking
```

### Performance Metrics
```
Key Performance Indicators:
├── Network Performance
│   ├── Bandwidth: 1Gbps WAN, 2.5Gbps LAN
│   ├── Latency: <5ms local, <20ms remote
│   ├── Packet Loss: <0.1%
│   └── Uptime: 99.9%
├── Server Performance
│   ├── CPU Usage: <80% average
│   ├── Memory Usage: <90% average
│   ├── Storage Usage: <85% average
│   └── Response Time: <100ms
├── Application Performance
│   ├── Web Response: <500ms
│   ├── Database Queries: <50ms
│   ├── File Transfers: 100MB/s
│   └── Backup Speed: 50MB/s
└── Security Metrics
    ├── Failed Logins: <5/day
    ├── Security Alerts: <1/week
    ├── Vulnerability Scans: Weekly
    └── Patch Compliance: 100%
```

## 🔄 Backup and Recovery

### Backup Strategy
```
Backup Systems:
├── System Backups
│   ├── OPNSense: Configuration backup
│   ├── UNRAID: Array and cache backup
│   ├── Development: Code repository backup
│   └── Database: PostgreSQL backup
├── Data Backups
│   ├── Local: UNRAID array protection
│   ├── Cloud: Google Drive sync
│   ├── Offsite: External drive rotation
│   └── Version: Git version control
├── Recovery Procedures
│   ├── System Recovery: Automated restore
│   ├── Data Recovery: Point-in-time restore
│   ├── Network Recovery: Failover procedures
│   └── Disaster Recovery: Complete rebuild
└── Testing
    ├── Backup Verification: Weekly tests
    ├── Recovery Testing: Monthly drills
    ├── Performance Testing: Quarterly
    └── Security Testing: Continuous
```

## 🔮 Future Enhancements

### Planned Upgrades
```
Network Upgrades:
├── WAN: 2.5Gbps fiber (when available)
├── LAN: 10Gbps infrastructure
├── WiFi: WiFi 6E deployment
└── Security: Advanced threat protection

Server Upgrades:
├── UNRAID: Additional storage expansion
├── Development: GPU upgrade for AI/ML
├── Network: Redundant internet connections
└── Monitoring: Advanced analytics platform

Software Upgrades:
├── OPNSense: Latest stable release
├── Tailscale: Advanced ACL features
├── Docker: Latest container runtime
└── Applications: Latest versions
```

### Scalability Considerations
```
Growth Planning:
├── Multi-Site: Geographic redundancy
├── Cloud Integration: Hybrid cloud architecture
├── IoT Expansion: Large-scale IoT deployment
├── AI Integration: Advanced AI capabilities
└── Edge Computing: Distributed computing
```

## 📚 Related Documentation

- [Network Architecture](./network.md)
- [Tailscale Configuration](./tailscale.md)
- [OPNSense Configuration](./opnsense.md)
- [UNRAID Server Configuration](./unraid.md)

---

*Last Updated: 2025-01-14*
*Status: Production*
*Maintainer: AFTRS Development Team* 