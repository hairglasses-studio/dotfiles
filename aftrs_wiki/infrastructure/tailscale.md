# Tailscale Configuration - AFTRS Network

## Overview
Tailscale provides secure, zero-config VPN connectivity for the AFTRS network, enabling seamless remote access, secure communication, and network segmentation across all devices and services.

## 🏗️ Network Architecture

### Tailscale Network: aftrs.void
- **Network Name**: aftrs.void
- **Domain**: aftrs.void
- **Exit Node**: aftrs-main (OPNSense Router)
- **Subnet Routes**: 192.168.1.0/24, 192.168.2.0/24
- **ACL**: Role-based access control

### Network Topology
```
aftrs.void Network
├── aftrs-main (Exit Node)
│   ├── IP: 100.64.0.1
│   ├── Role: Router/Gateway
│   └── Routes: 192.168.1.0/24
├── aftrs-unraid (File Server)
│   ├── IP: 100.64.0.10
│   ├── Role: Storage/Media
│   └── Services: SMB, NFS, Docker
├── aftrs-dev (Development)
│   ├── IP: 100.64.0.20
│   ├── Role: Development
│   └── Services: SSH, Git, Development
└── Mobile Devices
    ├── IP: 100.64.0.30+
    ├── Role: Remote Access
    └── Services: SSH, Web Access
```

## 🔧 Configuration

### Tailscale ACL (Access Control List)
```json
{
  "tagOwners": {
    "tag:admin": ["hairglasses@gmail.com"],
    "tag:server": ["hairglasses@gmail.com"],
    "tag:dev": ["hairglasses@gmail.com"],
    "tag:mobile": ["hairglasses@gmail.com"]
  },
  "groups": {
    "group:admins": ["hairglasses@gmail.com"],
    "group:developers": ["hairglasses@gmail.com"],
    "group:users": ["hairglasses@gmail.com"]
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
  ],
  "ssh": [
    {
      "action": "accept",
      "src": ["tag:admin"],
      "dst": ["*:22"],
      "users": ["root", "hairglasses"]
    },
    {
      "action": "accept",
      "src": ["tag:dev"],
      "dst": ["aftrs-dev:22"],
      "users": ["hairglasses"]
    }
  ]
}
```

### Host Configuration

#### aftrs-main (OPNSense Router)
```bash
# Tailscale Configuration
tailscale up \
  --advertise-routes=192.168.1.0/24 \
  --advertise-exit-node \
  --hostname=aftrs-main \
  --authkey=tskey-auth-xxxxx

# Exit Node Configuration
echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding = 1' >> /etc/sysctl.conf
sysctl -p

# Firewall Rules (OPNSense)
# Allow Tailscale traffic
# Route traffic through exit node
```

#### aftrs-unraid (UNRAID Server)
```bash
# Tailscale Configuration
tailscale up \
  --hostname=aftrs-unraid \
  --authkey=tskey-auth-xxxxx

# Service Configuration
# SMB: 445/tcp
# NFS: 2049/tcp
# Docker: 2375/tcp
# Web UI: 80/tcp, 443/tcp
```

#### aftrs-dev (Development PC)
```bash
# Tailscale Configuration
tailscale up \
  --hostname=aftrs-dev \
  --authkey=tskey-auth-xxxxx

# SSH Configuration
# Allow Tailscale SSH access
# Development tools accessible
```

## 🚀 Installation and Setup

### Prerequisites
- **Tailscale Account**: hairglasses@gmail.com
- **Admin Access**: Root/sudo access on all hosts
- **Network Connectivity**: Internet access on all hosts
- **Firewall Configuration**: Allow Tailscale traffic

### Installation Process

#### 1. Install Tailscale
```bash
# Ubuntu/Debian
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/jammy.gpg | sudo tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/jammy.tailscale-keyring.list | sudo tee /etc/apt/sources.list.d/tailscale.list
sudo apt update
sudo apt install tailscale

# OPNSense
pkg install tailscale
```

#### 2. Authenticate Hosts
```bash
# Generate auth key
tailscale up --authkey=tskey-auth-xxxxx

# Verify connection
tailscale status
```

#### 3. Configure Exit Node (aftrs-main)
```bash
# Enable exit node
tailscale up --advertise-exit-node

# Configure routing
echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
sysctl -p
```

#### 4. Configure Subnet Routes
```bash
# Advertise local subnets
tailscale up --advertise-routes=192.168.1.0/24,192.168.2.0/24

# Verify routes
tailscale status
```

## 📊 Network Services

### File Sharing (aftrs-unraid)
```bash
# SMB Configuration
# /etc/samba/smb.conf
[global]
   interfaces = lo 100.64.0.10
   bind interfaces only = yes

[share]
   path = /mnt/user/share
   browseable = yes
   read only = no
   guest ok = no
   valid users = hairglasses
```

### SSH Access
```bash
# SSH Configuration
# /etc/ssh/sshd_config
AllowUsers hairglasses
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
```

### Web Services
```bash
# Nginx Configuration
# /etc/nginx/sites-available/tailscale
server {
    listen 100.64.0.10:80;
    server_name aftrs-unraid.aftrs.void;
    
    location / {
        proxy_pass http://localhost:8080;
    }
}
```

## 🔒 Security Configuration

### Access Control
- **Role-Based Access**: Different permissions per user role
- **Tag-Based Security**: Hosts tagged by function
- **SSH Key Authentication**: No password authentication
- **Firewall Rules**: Restrict access to necessary ports

### Network Segmentation
- **Admin Access**: Full access to all hosts
- **Server Access**: Limited to server management
- **Development Access**: Access to dev environment
- **Mobile Access**: Web-only access

### Monitoring and Logging
```bash
# Tailscale Logs
tailscale status --json

# Connection Monitoring
tailscale netcheck

# ACL Verification
tailscale acl
```

## 🔧 TSCTL Integration

### TSCTL Commands
```bash
# Network Status
tsctl status

# Health Check
tsctl doctor

# Tag Management
tsctl check

# Configuration
tsctl setup
```

### Automated Management
```bash
# Bootstrap new host
tsctl bootstrap

# Apply configuration
tsctl apply

# Verify setup
tsctl verify
```

## 🚨 Troubleshooting

### Common Issues

#### Connection Problems
```bash
# Check Tailscale status
tailscale status

# Test connectivity
tailscale ping aftrs-main

# Check routes
tailscale netcheck

# Restart Tailscale
sudo systemctl restart tailscaled
```

#### Exit Node Issues
```bash
# Check exit node status
tailscale status

# Verify routing
ip route show

# Test exit node
curl -s https://checkip.amazonaws.com

# Check firewall rules
iptables -L -n
```

#### DNS Issues
```bash
# Check DNS resolution
nslookup aftrs-main.aftrs.void

# Test local DNS
dig @100.64.0.1 aftrs-unraid.aftrs.void

# Verify DNS configuration
tailscale status --json
```

### Debug Commands
```bash
# Enable debug logging
export TS_DEBUG=1
tailscale up

# Verbose status
tailscale status --verbose

# Network diagnostics
tailscale netcheck --verbose

# ACL debugging
tailscale acl --verbose
```

## 📈 Performance Optimization

### Network Optimization
- **Direct Connection**: Peer-to-peer when possible
- **Relay Fallback**: DERP servers for NAT traversal
- **Route Optimization**: Shortest path routing
- **Bandwidth Management**: QoS for critical traffic

### Security Optimization
- **Key Rotation**: Regular key updates
- **ACL Updates**: Dynamic access control
- **Monitoring**: Real-time security monitoring
- **Backup**: Configuration backup and recovery

## 🔮 Future Enhancements

### Planned Features
- **Multi-Site**: Geographic redundancy
- **Load Balancing**: Multiple exit nodes
- **Advanced ACL**: Fine-grained access control
- **Integration**: Deeper integration with AFTRS CLI
- **Monitoring**: Advanced network analytics

### Scalability Considerations
- **Host Limits**: Current limit of 20 devices
- **Bandwidth**: Monitor usage and optimize
- **Security**: Regular security audits
- **Backup**: Automated configuration backup

## 📚 Related Documentation

- [Network Architecture](./network.md)
- [TSCTL Management](../projects/tsctl.md)
- [AFTRS CLI Integration](../projects/aftrs_cli.md)
- [OPNSense Configuration](./opnsense.md)

## 🔗 Integration Points

### With Other Systems
- **AFTRS CLI**: Network management automation
- **AgentCTL**: Configuration tracking
- **UNRAID Scripts**: Server connectivity
- **OPNSense**: Router integration

### External Services
- **Tailscale API**: Network management
- **SSH**: Secure remote access
- **SMB/NFS**: File sharing
- **Web Services**: HTTP/HTTPS access

---

*Last Updated: 2025-01-14*
*Status: Production*
*Maintainer: AFTRS Development Team* 