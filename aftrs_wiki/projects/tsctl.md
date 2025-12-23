# TSCTL - Tailscale Network Management

## Overview
TSCTL is a comprehensive Tailscale network management and automation tool that provides simplified configuration, status monitoring, and network administration for the AFTRS Tailscale network. It streamlines the process of managing Tailscale hosts, tags, and routes across the distributed infrastructure.

## 🏗️ Architecture

### Core Components
- **CLI Interface**: Command-line tool for network management
- **Configuration Management**: YAML-based configuration storage
- **Tailscale Integration**: Direct API integration with Tailscale
- **Tag Management**: Automated tag application and verification
- **Health Monitoring**: Network status and connectivity checks
- **Bootstrap Scripts**: Automated setup and configuration

### Key Features
- **Interactive Setup**: Guided configuration process
- **Status Monitoring**: Real-time network status
- **Tag Verification**: Automatic tag compliance checking
- **Health Diagnostics**: Comprehensive network diagnostics
- **Automated Hooks**: Tailscale login hook integration

## 📁 Project Structure

```
tsctl/
├── 📖 README.md                    # Project documentation
├── 🚀 install.sh                   # Installation script
├── 🔧 Makefile                     # Build and development tasks
├── 📁 bin/                         # Executable scripts
├── 📁 lib/                         # Core library functions
├── 📁 config/                      # Configuration templates
├── 📁 completions/                 # Shell completion scripts
│   ├── .bash_completion.d/        # Bash completion
│   └── .zsh/                      # Zsh completion
├── 📁 install/                     # Installation utilities
├── 🔗 tsctl-bootstrap.sh          # Bootstrap script
└── 🏷️  tsctl-autotag.sh          # Tag automation script
```

## 🚀 Installation

### Prerequisites
- GitHub CLI (`gh`)
- Tailscale client installed
- Access to private repository: `hairglasses/tsctl`

### Installation Process
```bash
# Clone the repository
gh repo clone hairglasses/tsctl
cd tsctl

# Install locally
./install.sh

# Verify installation
tsctl doctor
```

### Installation Locations
- **Binary**: `~/.local/bin/tsctl`
- **Configuration**: `~/.config/tsctl/`
- **Completions**: `~/.bash_completion.d/` and `~/.zsh/`

## 🔧 Configuration

### Setup Process
```bash
# Interactive setup
tsctl setup

# This will guide you through:
# 1. Site selection
# 2. Role assignment
# 3. Optional tag configuration
# 4. Tailscale login hook setup
```

### Configuration Files
```yaml
# ~/.config/tsctl/config.yaml
site: "aftrs-main"
role: "development"
tags:
  - "env:dev"
  - "team:aftrs"
  - "location:home"
```

## 📊 Features

### Core Commands

#### Status Monitoring
```bash
# View current Tailscale status
tsctl status

# Output includes:
# - Active hostname
# - Applied tags
# - Advertised routes
# - Connection status
```

#### Health Diagnostics
```bash
# Comprehensive network audit
tsctl doctor

# Checks for:
# - Missing dependencies
# - Configuration issues
# - Network connectivity
# - Tag compliance
```

#### Tag Management
```bash
# Verify tag compliance
tsctl check

# Compare applied tags to saved configuration
# Reports any mismatches or missing tags
```

#### Network Administration
```bash
# Remove configuration and hooks
tsctl unhook

# Cleans up:
# - Tailscale login hooks
# - Configuration files
# - Completions
```

### Advanced Features

#### Bootstrap Automation
```bash
# Automated setup script
./tsctl-bootstrap.sh

# Performs:
# - Repository cloning
# - Installation
# - Initial configuration
# - Hook setup
```

#### Tag Automation
```bash
# Automated tag application
./tsctl-autotag.sh

# Applies tags based on:
# - Host environment
# - User role
# - Network location
```

## 🔄 Workflow

### Initial Setup
1. **Clone Repository**: `gh repo clone hairglasses/tsctl`
2. **Install**: `./install.sh`
3. **Configure**: `tsctl setup`
4. **Verify**: `tsctl doctor`

### Daily Operations
1. **Check Status**: `tsctl status`
2. **Verify Tags**: `tsctl check`
3. **Monitor Health**: `tsctl doctor`

### Troubleshooting
1. **Diagnose Issues**: `tsctl doctor`
2. **Reconfigure**: `tsctl setup`
3. **Clean Reset**: `tsctl unhook`

## 📈 Current Status

### ✅ Completed Features
- [x] Interactive setup and configuration
- [x] Status monitoring and reporting
- [x] Tag management and verification
- [x] Health diagnostics and auditing
- [x] Shell completions (Bash/Zsh)
- [x] Bootstrap automation scripts
- [x] Tailscale login hook integration
- [x] Configuration management

### 🔄 In Progress
- [ ] Advanced network analytics
- [ ] Multi-site management
- [ ] Automated failover
- [ ] Performance monitoring

### 📋 Planned Features
- [ ] Web dashboard integration
- [ ] API for external tools
- [ ] Advanced routing management
- [ ] Security auditing tools
- [ ] Network topology visualization

## 🛠️ Technical Decisions

### Architecture Design
- **CLI-First**: Command-line interface for automation
- **Configuration-Driven**: YAML-based configuration
- **Hook Integration**: Tailscale login hook automation
- **Shell Integration**: Completions for better UX

### Security Considerations
- **Private Repository**: Secure access control
- **Tag-Based Security**: Role-based access control
- **Configuration Validation**: Input sanitization
- **Audit Trail**: Comprehensive logging

### Network Management
- **Automated Tagging**: Consistent network policies
- **Health Monitoring**: Proactive issue detection
- **Status Verification**: Configuration compliance
- **Bootstrap Automation**: Simplified deployment

## 🔍 Troubleshooting

### Common Issues

#### Installation Problems
```bash
# Check GitHub CLI access
gh auth status

# Verify repository access
gh repo view hairglasses/tsctl

# Reinstall if needed
./install.sh
```

#### Configuration Issues
```bash
# Check configuration
cat ~/.config/tsctl/config.yaml

# Reset configuration
tsctl unhook
tsctl setup
```

#### Network Connectivity
```bash
# Check Tailscale status
tailscale status

# Verify tags
tsctl check

# Run diagnostics
tsctl doctor
```

### Debug Mode
```bash
# Enable debug logging
export TSCTL_DEBUG=1
tsctl status

# View detailed output
tsctl doctor --verbose
```

## 📚 Related Documentation

- [Tailscale Configuration](../infrastructure/tailscale.md)
- [Network Architecture](../infrastructure/network.md)
- [CLI Design Patterns](../patterns/cli_design.md)
- [AFTRS CLI Integration](./aftrs_cli.md)

## 🔗 Integration Points

### With Other Systems
- **AgentCTL**: Network status integration
- **AFTRS CLI**: Network management automation
- **UNRAID Scripts**: Server connectivity monitoring
- **OPNSense Router**: Network routing coordination

### External Services
- **Tailscale API**: Direct network management
- **GitHub**: Repository access and updates
- **Shell Environment**: Completions and hooks

---

*Last Updated: 2025-01-14*
*Status: Production Ready*
*Maintainer: AFTRS Development Team* 