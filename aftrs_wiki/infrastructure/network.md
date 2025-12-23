# Network Architecture - AFTRS Asset-Driven Infrastructure

## Overview
The AFTRS network infrastructure is a comprehensive, asset-driven system designed for high performance, security, and reliability. It integrates physical hardware, cloud services, VPN connectivity, and software-defined networking with real-time monitoring, automated testing, and web-based management to create a unified network ecosystem.

## 🏗️ Physical Network Topology

### Core Infrastructure
```
Internet (Comcast)
    ↓
Comcast Modem (Router Mode)
    ↓
OPNSense Router (Primary Gateway)
    ↓
Core Switch (Managed)
    ↓
┌─────────────────┬─────────────────┬─────────────────┐
│   UNRAID Server │  Development PC │  Mobile Devices │
│   (11.1.11.10)  │   (11.1.11.20) │   (11.1.11.30) │
└─────────────────┴─────────────────┴─────────────────┘
```

### Network Segments
- **WAN**: Comcast Internet (1Gbps/35Mbps)
- **LAN**: 11.1.11.0/24 (Primary Network)
- **DMZ**: 11.1.12.0/24 (Isolated Services)
- **IoT**: 11.1.13.0/24 (IoT Devices)
- **Guest**: 11.1.14.0/24 (Guest Network)

## 🎯 Asset-Driven Management

### Asset Inventory System
The network is managed through a centralized asset inventory (`network_assets/assets.yaml`) that includes:

#### Network Assets
```yaml
assets:
  - id: "aftrs-main"
    name: "AFTRS Main Router"
    type: "router"
    ip: "11.1.11.1"
    site: "annex"
    status: "online"
    services:
      - name: "web-interface"
        port: 443
        protocol: "https"
      - name: "ssh"
        port: 22
        protocol: "ssh"

  - id: "comcast-modem"
    name: "Comcast Business Modem"
    type: "modem"
    ip: "10.1.10.1"
    site: "annex"
    status: "online"
    services:
      - name: "web-interface"
        port: 80
        protocol: "http"

  - id: "aftrs-unraid"
    name: "AFTRS UNRAID Server"
    type: "server"
    ip: "11.1.11.10"
    site: "annex"
    status: "online"
    services:
      - name: "smb"
        port: 445
        protocol: "smb"
      - name: "web-interface"
        port: 80
        protocol: "http"
```

#### Asset Relationships
```yaml
relationships:
  - from: "comcast-modem"
    to: "aftrs-main"
    type: "wan-connection"
    description: "WAN connection from modem to router"
  
  - from: "aftrs-main"
    to: "aftrs-unraid"
    type: "lan-connection"
    description: "LAN connection to UNRAID server"
```

### Asset-Driven Testing
- **Dynamic Test Generation**: Tests created automatically from asset inventory
- **Parallel Execution**: Multiple tests run simultaneously for efficiency
- **Intelligent Caching**: Test results cached to improve performance
- **Adaptive Timeouts**: Timeouts adjusted based on network conditions

## 🌐 Internet Connectivity

### Comcast Modem Configuration
- **Model**: Arris SB8200 (DOCSIS 3.1)
- **Mode**: Router Mode (NAT enabled)
- **IP**: 10.1.10.1 (WAN Gateway)
- **Port**: 1Gbps Ethernet to OPNSense
- **DMZ**: Configured for OPNSense router

### OPNSense Router Configuration
- **Model**: Custom Build (Intel i3-6100, 8GB RAM)
- **OS**: OPNSense 23.1+
- **WAN**: 10.1.10.2 (DHCP from Comcast)
- **LAN**: 11.1.11.1/24
- **Features**: Firewall, NAT, DHCP, DNS, Monitoring

## 🔒 Security Architecture

### Firewall Zones
```
Internet (Untrusted)
    ↓
WAN Interface (OPNSense)
    ↓
DMZ Zone (Semi-Trusted)
    ↓
LAN Zone (Trusted)
    ↓
IoT Zone (Restricted)
```

### Security Policies
- **Default Deny**: All traffic blocked by default
- **Explicit Allow**: Only authorized traffic permitted
- **Zone Isolation**: Strict separation between networks
- **Intrusion Detection**: Suricata IDS/IPS
- **VPN Access**: Tailscale for remote connectivity
- **Asset-Based Security**: Security policies tied to asset inventory

## 🐛 Tailscale Network Integration

### Tailscale Configuration
- **Network**: aftrs.void
- **Subnet Routes**: 11.1.11.0/24, 11.1.12.0/24
- **Exit Node**: OPNSense Router
- **ACL**: Role-based access control
- **Real-time Monitoring**: Live session tracking and health diagnostics

### Tailscale Hosts
```
aftrs-main (OPNSense Router)
├── IP: 100.64.0.1
├── Role: Exit Node
├── Routes: 11.1.11.0/24
└── Status: Monitored via AFTRS CLI

aftrs-unraid (UNRAID Server)
├── IP: 100.64.0.10
├── Role: File Server
├── Services: SMB, NFS, Docker
└── Status: Monitored via AFTRS CLI

aftrs-dev (Development PC)
├── IP: 100.64.0.20
├── Role: Development
├── Services: SSH, Git, Development Tools
└── Status: Monitored via AFTRS CLI
```

## 📡 DNS and DHCP

### DNS Configuration
- **Primary DNS**: OPNSense (11.1.11.1)
- **Secondary DNS**: 1.1.1.1, 8.8.8.8
- **Local Domain**: aftrs.void
- **DNS Override**: Unbound for local resolution
- **Asset-Driven DNS**: DNS entries managed through asset inventory

### DHCP Configuration
- **Server**: OPNSense DHCP
- **Range**: 11.1.11.100-11.1.11.200
- **Lease Time**: 24 hours
- **Static Leases**: Critical devices from asset inventory
- **Asset Integration**: DHCP leases tied to asset management

### DNS Overrides
```
aftrs-main.aftrs.void → 11.1.11.1
aftrs-unraid.aftrs.void → 11.1.11.10
aftrs-dev.aftrs.void → 11.1.11.20
*.aftrs.void → 11.1.11.1 (Catch-all)
```

## 🖥️ Server Infrastructure

### UNRAID Server (aftrs-unraid)
- **Hardware**: Custom Build (AMD Ryzen 7, 32GB RAM)
- **Storage**: 4x 8TB HDD + 2x 1TB NVMe
- **Network**: 2.5Gbps Ethernet
- **Services**: Docker, VMs, File Sharing, Dynamix Active Streams
- **IP**: 11.1.11.10 (LAN), 100.64.0.10 (Tailscale)
- **Monitoring**: SMB share monitoring, active streams tracking

### Development PC (aftrs-dev)
- **Hardware**: Custom Build (Intel i7, 32GB RAM)
- **OS**: Ubuntu 22.04 LTS
- **Network**: 1Gbps Ethernet + WiFi
- **Services**: Development Tools, Git, Docker
- **IP**: 11.1.11.20 (LAN), 100.64.0.20 (Tailscale)

## 🔧 Network Services

### Core Services (OPNSense)
- **Firewall**: Stateful packet inspection
- **NAT**: Port forwarding and address translation
- **DHCP**: Dynamic host configuration
- **DNS**: Local name resolution
- **VPN**: Tailscale integration
- **Monitoring**: Traffic analysis and logging
- **Asset Management**: Integration with asset inventory

### Application Services (UNRAID)
- **Docker**: Container orchestration
- **VMs**: Virtual machine hosting
- **File Sharing**: SMB, NFS, WebDAV
- **Backup**: Automated backup services
- **Media**: Plex, Jellyfin, Sonarr
- **Monitoring**: Dynamix Active Streams plugin

### Development Services (Development PC)
- **Git**: Version control
- **Docker**: Development containers
- **SSH**: Remote access
- **Development Tools**: IDEs, compilers, debuggers

## 📊 Network Monitoring

### Real-time Monitoring
- **Asset Status**: Live monitoring of all network assets
- **Performance Metrics**: Bandwidth, latency, throughput tracking
- **Service Health**: Continuous service availability monitoring
- **Alert System**: Configurable alerts with multiple notification channels
- **Web Dashboard**: Real-time monitoring interface with custom widgets

### Advanced Diagnostics
- **Comcast DMZ Analysis**: Modem configuration detection and port forwarding verification
- **NAT Chain Visualization**: Live packet flow analysis and topology mapping
- **Tailscale Monitoring**: Real-time session tracking and health diagnostics
- **Performance Optimization**: Parallel execution, intelligent caching, adaptive timeouts

### Traffic Analysis
- **Bandwidth Monitoring**: Real-time traffic analysis
- **Connection Tracking**: Active connection monitoring
- **Protocol Analysis**: Application-level traffic analysis
- **Performance Metrics**: Latency, throughput, packet loss

### Security Monitoring
- **Intrusion Detection**: Suricata IDS/IPS
- **Log Analysis**: Centralized log collection
- **Alert System**: Automated security notifications
- **Vulnerability Scanning**: Regular security assessments
- **Asset-Based Security**: Security monitoring tied to asset inventory

## 🔄 Network Automation

### Asset-Driven Automation
- **Dynamic Test Generation**: Tests created automatically from asset inventory
- **Automated Documentation**: Network diagrams and documentation generated from assets
- **Scheduled Tasks**: Asset import, monitoring, and maintenance automation
- **Performance Optimization**: Parallel testing and intelligent caching
- **Web Dashboard**: Real-time monitoring and management interface

### Automated Tasks
- **DHCP Management**: Static lease automation from asset inventory
- **DNS Updates**: Dynamic DNS registration from asset inventory
- **Firewall Rules**: Automated rule management based on asset relationships
- **Backup Scheduling**: Network configuration backup
- **Health Monitoring**: Automated health checks for all assets

### Integration Points
- **AFTRS CLI**: Comprehensive asset-driven network management
- **TSCTL**: Full Tailscale management integration
- **AgentCTL**: Configuration tracking and automation
- **UNRAID Scripts**: Server automation and monitoring
- **Web Dashboard**: Real-time monitoring and management interface

## 🚨 Troubleshooting

### Asset-Driven Diagnostics
```bash
# Complete network audit
./aftrs.sh diag

# Asset status check
./aftrs.sh assets list

# Dynamic test generation and execution
./aftrs.sh tests generate
./aftrs.sh tests run

# Real-time monitoring
./aftrs.sh monitor status
./aftrs.sh monitor topology
./aftrs.sh monitor health

# Advanced diagnostics
./aftrs.sh dmz reachability
./aftrs.sh dmz bridge
./aftrs.sh nat topology
./aftrs.sh nat live
```

### Common Network Issues

#### Connectivity Problems
```bash
# Check physical connections
ping 11.1.11.1

# Test DNS resolution
nslookup google.com

# Verify DHCP
ip addr show

# Check firewall rules
iptables -L

# Asset-based connectivity test
./aftrs.sh tests run --category connectivity
```

#### Performance Issues
```bash
# Test bandwidth
speedtest-cli

# Check latency
ping -c 10 8.8.8.8

# Monitor traffic
iftop

# Analyze connections
netstat -tuln

# Performance optimization
./aftrs.sh optimize test
./aftrs.sh optimize status
```

#### Security Issues
```bash
# Check firewall status
systemctl status firewall

# Review security logs
tail -f /var/log/firewall.log

# Test port accessibility
nmap -sT 11.1.11.1

# Verify VPN connectivity
tailscale status

# Asset-based security audit
./aftrs.sh tests run --category security
```

### Network Diagnostics

#### OPNSense Diagnostics
```bash
# System status
systemctl status opnsense

# Interface status
ifconfig

# Routing table
netstat -rn

# Firewall rules
pfctl -sr
```

#### Tailscale Diagnostics
```bash
# Network status
tailscale status

# Connectivity test
tailscale ping aftrs-unraid

# Route verification
tailscale netcheck

# ACL verification
tailscale acl

# Asset-based Tailscale monitoring
./aftrs.sh tailscale status
./aftrs.sh tailscale topology
```

## 📈 Performance Optimization

### Asset-Driven Optimization
- **Parallel Testing**: Multiple tests run simultaneously based on asset relationships
- **Intelligent Caching**: Test results cached to improve performance
- **Adaptive Timeouts**: Timeouts adjusted based on network conditions and asset types
- **Performance Monitoring**: Real-time performance tracking for all assets

### Bandwidth Management
- **QoS**: Quality of Service for critical traffic
- **Traffic Shaping**: Bandwidth allocation
- **Load Balancing**: Multiple WAN connections
- **Caching**: DNS and web content caching

### Network Optimization
- **Jumbo Frames**: 9000 byte MTU for local traffic
- **Link Aggregation**: Bonded network interfaces
- **VLANs**: Network segmentation
- **Wireless Optimization**: 5GHz channel optimization

## 🌐 Web Dashboard

### Dashboard Features
- **Real-time Asset Monitoring**: Live status updates and health checks
- **Network Topology Visualization**: Interactive network diagrams
- **Performance Metrics**: Bandwidth, latency, and service health tracking
- **Alert Management**: Configurable alerts with multiple notification channels
- **Test Execution**: Web-based test running and result viewing
- **Custom Widgets**: Performance charts, maintenance schedules, relationship maps

### Dashboard Access
```bash
# Start dashboard
./aftrs.sh dashboard start

# Stop dashboard
./aftrs.sh dashboard stop

# Check status
./aftrs.sh dashboard status

# Access URL: http://localhost:8080
```

## 🔮 Future Enhancements

### Planned Improvements
- **Machine Learning**: Predictive network analysis and anomaly detection
- **Advanced Analytics**: Historical data analysis and trend prediction
- **Mobile App**: Mobile monitoring interface
- **API Gateway**: REST API for external integrations
- **Multi-site Management**: Geographic network management
- **Zero-touch Provisioning**: Automated network device provisioning
- **Security Analysis**: Network security auditing and threat detection
- **Capacity Planning**: Network capacity analysis and planning
- **Disaster Recovery**: Automated backup and recovery procedures

### Scalability Considerations
- **Multi-Site**: Geographic redundancy with asset-driven management
- **Cloud Integration**: Hybrid cloud architecture
- **IoT Expansion**: Large-scale IoT deployment with asset tracking
- **Edge Computing**: Distributed computing resources
- **Asset-Driven Scaling**: Scaling based on asset inventory and relationships

## 📚 Related Documentation

- [OPNSense Configuration](./opnsense.md)
- [Tailscale Configuration](./tailscale.md)
- [UNRAID Server Configuration](./unraid.md)
- [AFTRS CLI Integration](../projects/aftrs_cli.md)
- [Asset Management Guide](../projects/asset_management.md)
- [Performance Optimization Guide](../projects/optimization_guide.md)

---

*Last Updated: 2025-01-15*
*Status: Production with Asset-Driven Management*
*Maintainer: AFTRS Development Team* 