# AFTRS CLI - Asset-Driven Network Management System

## Overview
AFTRS CLI is a comprehensive network automation and monitoring toolkit for the aftrs.void infrastructure, featuring asset-driven testing, real-time monitoring, web dashboard, and automated management capabilities. It provides centralized control for configuring, managing, and automating infrastructure within the aftrs.void LAN.

## 🏗️ Architecture

### Core Components
- **Asset-Driven Management**: YAML-based asset inventory with relationships and metadata
- **Real-time Monitoring**: Tailscale session tracking and health diagnostics
- **Advanced Testing**: Comcast DMZ analysis and NAT chain visualization
- **Web Dashboard**: Real-time monitoring with custom widgets and alerting
- **Performance Optimization**: Parallel execution, intelligent caching, and adaptive timeouts
- **Automation Pipeline**: Scheduled tasks and automated remediation

### Key Features
- **Centralized Asset Inventory**: Comprehensive network asset tracking
- **Dynamic Test Generation**: Asset-driven test creation and execution
- **Real-time Monitoring**: Live network health and performance monitoring
- **Web Dashboard**: Interactive monitoring interface with custom widgets
- **Tailscale Integration**: Comprehensive Tailscale management with backward compatibility
- **Automated Documentation**: Auto-generated network diagrams and documentation

## 📁 Project Structure

```
aftrs_cli/
├── 📖 README.md                    # Comprehensive project documentation
├── 🚀 aftrs.sh                     # Main CLI dispatcher
├── 📁 network_assets/              # Asset management
│   ├── assets.yaml                # Asset inventory
│   ├── asset_loader.py            # Asset query engine
│   ├── asset_merger.py            # Import/merge utilities
│   └── generate_docs_from_assets.py
├── 📁 scripts/                     # Advanced modules
│   ├── tailscale_manager.sh       # Tailscale management
│   ├── tailscale_monitor.sh       # Real-time monitoring
│   ├── comcast_dmz_checker.sh     # DMZ analysis
│   ├── nat_chain_visualizer.sh    # NAT visualization
│   ├── router_tester.sh           # Router testing
│   ├── remote_site_tester.sh      # Remote site testing
│   ├── speed_tester.sh            # Performance testing
│   ├── performance_optimizer.sh   # Optimization engine
│   ├── dynamic_test_generator.sh  # Asset-driven testing
│   ├── asset_import_scheduler.sh  # Scheduled imports
│   └── asset_monitoring_pipeline.sh # Monitoring pipeline
├── 📁 web_dashboard/               # Web interface
│   ├── dashboard.py               # Flask application
│   └── templates/                 # HTML templates
├── 📁 logs/                        # Log files
├── 📁 docs/                        # Documentation
├── 📁 dhcp/                        # DHCP management
├── 📁 dns/                         # DNS management
└── 📁 install.sh                   # Installation script
```

## 🚀 Installation

### Prerequisites
- **OPNSense Router**: Fresh or existing installation
- **Root Access**: SSH or direct access to router
- **Network Connectivity**: Access to aftrs.void LAN
- **Tailscale Client**: Network connectivity setup
- **Python 3.7+**: For web dashboard and asset management

### Bootstrap Installation
```bash
# Automated bootstrap
curl -sSL https://your-repo/bootstrap_opnsense_agent.sh | sudo bash

# Manual installation
git clone https://github.com/hairglasses/aftrs_cli.git
cd aftrs_cli
./install.sh
```

### Installation Locations
- **CLI Binary**: `/usr/local/bin/aftrs_cli`
- **Configuration**: `/usr/local/aftrs_cli/`
- **Backups**: `/usr/local/aftrs_cli/backups`
- **Scripts**: `/usr/local/aftrs_cli/scripts`
- **Web Dashboard**: `/usr/local/aftrs_cli/web_dashboard`

## 🔧 Configuration

### Asset Management
```bash
# List all assets
./aftrs.sh assets list

# Query by type
./aftrs.sh assets list --type router

# Query by site
./aftrs.sh assets list --site annex

# Export as JSON
./aftrs.sh assets list --format json

# Advanced filtering
./aftrs.sh assets list --service smb --status online
```

### Network Configuration
```bash
# DNS Management
./aftrs.sh dns list
./aftrs.sh dns add <hostname> <ip>
./aftrs.sh dns remove <hostname>

# DHCP Management
./aftrs.sh dhcp list
./aftrs.sh dhcp add <mac> <ip> <hostname>
./aftrs.sh dhcp remove <mac>

# Firewall Management
./aftrs.sh firewall list
./aftrs.sh firewall add <rule>
./aftrs.sh firewall remove <rule>
```

## 📊 Features

### Core Commands

#### Asset Management
```bash
# List assets
./aftrs.sh assets list

# Query assets
./aftrs.sh assets query --type router --site annex

# Export assets
./aftrs.sh assets export --format json

# Import assets
./aftrs.sh import notion network_assets/notion_export.csv
```

#### Network Testing
```bash
# Generate tests from assets
./aftrs.sh tests generate

# Run all tests
./aftrs.sh tests run

# Run specific test categories
./aftrs.sh tests run --category connectivity
./aftrs.sh tests run --category services

# Parallel execution
./aftrs.sh tests run --parallel 4
```

#### Advanced Diagnostics
```bash
# Complete network audit
./aftrs.sh diag

# Tailscale monitoring
./aftrs.sh monitor status
./aftrs.sh monitor topology
./aftrs.sh monitor health

# DMZ analysis
./aftrs.sh dmz reachability
./aftrs.sh dmz bridge
./aftrs.sh dmz full

# NAT visualization
./aftrs.sh nat topology
./aftrs.sh nat live
./aftrs.sh nat report
```

#### Web Dashboard
```bash
# Start dashboard
./aftrs.sh dashboard start

# Stop dashboard
./aftrs.sh dashboard stop

# Check status
./aftrs.sh dashboard status
```

#### Tailscale Management
```bash
# Tailscale status
./aftrs.sh tailscale status

# Network topology
./aftrs.sh tailscale topology

# Health check
./aftrs.sh tailscale doctor

# Service exposure
./aftrs.sh tailscale expose start 8080 myapp

# Tag management
./aftrs.sh tailscale tags list
./aftrs.sh tailscale tags add my-tag
```

### Advanced Features

#### Performance Optimization
```bash
# Enable performance optimizations
./aftrs.sh optimize enable

# Run optimization tests
./aftrs.sh optimize test

# Check optimization status
./aftrs.sh optimize status
```

#### Remote Site Management
```bash
# Test remote site connectivity
./aftrs.sh remote test annex

# Check remote services
./aftrs.sh remote services annex

# Monitor remote topology
./aftrs.sh remote topology annex
```

#### UNRAID Integration
```bash
# Test UNRAID connectivity
./aftrs.sh unraid test

# Check SMB shares
./aftrs.sh unraid smb

# Monitor active streams
./aftrs.sh unraid streams
```

#### Documentation Generation
```bash
# Generate asset tables
./aftrs.sh docs table

# Generate network diagrams
./aftrs.sh docs diagram

# Generate comprehensive documentation
./aftrs.sh docs generate
```

#### Automation Pipeline
```bash
# Setup asset import scheduler
./aftrs.sh import schedule --interval 3600

# Setup monitoring pipeline
./aftrs.sh monitoring start

# Check automation status
./aftrs.sh monitoring status
```

## 🧠 Model Management

### Agentic Coding Models
- **Setup Models**: Pull, alias, and prepare agentic coding models for use

```bash
# Setup models with default settings
./aftrs.sh models

# Setup models with specific version and context size
./aftrs.sh models --model-version 2.0 --context-size 131072
```

This feature allows for flexible model management, enabling users to specify model versions and context sizes as needed.

## 🔄 Workflow

### Initial Setup
1. **Bootstrap Router**: `./aftrs.sh bootstrap`
2. **Configure Assets**: Edit `network_assets/assets.yaml`
3. **Import Assets**: `./aftrs.sh import notion <csv_file>`
4. **Generate Tests**: `./aftrs.sh tests generate`
5. **Start Dashboard**: `./aftrs.sh dashboard start`
6. **Verify Setup**: `./aftrs.sh diag`

### Daily Operations
1. **Check Status**: `./aftrs.sh assets list`
2. **Monitor Health**: `./aftrs.sh monitor health`
3. **Run Tests**: `./aftrs.sh tests run`
4. **View Dashboard**: Access web interface
5. **Check Alerts**: `./aftrs.sh monitoring status`

### Troubleshooting
1. **Diagnose Issues**: `./aftrs.sh diag`
2. **Check Performance**: `./aftrs.sh optimize test`
3. **Test Connectivity**: `./aftrs.sh tests run --category connectivity`
4. **View Logs**: `tail -f logs/*.log`

## Recent Updates

### Subnet Connectivity Testing
- Added connectivity tests between subnets `10.0.0.0/24`, `10.1.10.0/24`, and `11.1.11.0/24`.
- Integrated these tests into the `ts_netcheck.sh` script for automated diagnostics.

### Auto-Tagging Enhancement
- Improved auto-tagging based on Tailscale ACL.
- Devices are now automatically tagged according to ACL rules, enhancing network management and security.

### Configuration Updates
- Updated `cursor_config.json` to include auto-tagging configuration.
- Ensured seamless integration with existing Tailscale management features.

### Documentation
- Updated documentation to reflect these changes and provide guidance on new features.

---

## 📈 Current Status

### ✅ Completed Features
- [x] Subnet connectivity testing
- [x] Auto-tagging based on Tailscale ACL
- [x] Asset-driven management system
- [x] Real-time network monitoring
- [x] Advanced diagnostic modules
- [x] Web dashboard with custom widgets
- [x] Tailscale integration with backward compatibility
- [x] Performance optimization engine
- [x] Dynamic test generation
- [x] Automated documentation generation
- [x] Asset import/export capabilities
- [x] Monitoring pipeline with alerting
- [x] Remote site management
- [x] UNRAID integration
- [x] OPNsense integration and bootstrap
- [x] DNS management with Unbound overrides
- [x] DHCP static lease management
- [x] Firewall rule management
- [x] Configuration backup system

---

*Last Updated: 2025-01-15*
*Status: Production Ready*
*Maintainer: AFTRS Development Team* 

## 🚀 Bulk GitHub Repo Management (ghorg + gum)

AFTRS CLI now supports fully automated, prettified bulk cloning, syncing, status checking, autopush, and reset of all GitHub org/user repos using [ghorg](https://github.com/gabrie30/ghorg) and [gum](https://github.com/charmbracelet/gum).

- **All repos are always cloned and synced to `~/Docs` (your home Docs directory).** The script will create and populate this directory if it doesn't exist or is empty.
- Interactive (gum) or headless (YAML config) workflows
- All install scripts ensure `ghorg` and `gum` are present
- See [Bulk Repo Management Guide](../aftrs_cli/docs/bulk_repo_management.md) for full details, troubleshooting, and advanced automation tips.

# Archived Repo Cleanup

## Overview
A new function has been added to the AFTRS CLI toolkit to automatically clean up locally cloned repositories that have been marked as archived. This helps maintain a tidy workspace and ensures that obsolete or deprecated codebases do not persist on local storage.

## How It Works
- The function reads the list of archived repositories from the consolidation log (`consolidation_log.json`).
- It checks for any matching directories in the local `~/Docs/aftrs-void` folder.
- If a locally cloned repo is found and marked as archived, it is deleted from disk.
- The function can be run manually or integrated into automation workflows.

## Usage
The cleanup script is located at:
```
aftrs_cli/core_framework/modules/aftrs-cli/scripts/cleanup_archived_repos.py
```
To run the cleanup:
```bash
python cleanup_archived_repos.py
```

## Implementation Details
- Reads `consolidation_log.json` for archived repo names.
- Removes matching directories from `~/Docs/aftrs-void`.
- Logs actions to the console for auditability.

## Maintenance
- Update the consolidation log as repos are archived.
- Run the cleanup function periodically or as part of consolidation workflows.

---
_Last updated: 2025-09-23_