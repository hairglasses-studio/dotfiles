# unraid_snapshot_exporter

*Small utility consolidated from standalone repository*

---

# Unraid Snapshot Exporter

A comprehensive system diagnostics and configuration exporter for Unraid optimization and troubleshooting. This plugin provides detailed system analysis, performance monitoring, and diagnostic tools designed to provide rich context for large-context LLMs.

## 🌟 Features

### Core Functionality
- **Full System Snapshots**: Complete system diagnostics in a single ZIP file
- **Advanced System Analysis**: Structured JSON output for LLM consumption
- **Performance Monitoring**: Real-time metrics collection over time
- **Network Diagnostics**: Comprehensive network analysis and connectivity testing
- **Docker Analysis**: Detailed container and image analysis
- **VM Analysis**: Virtual machine resource and configuration analysis

### Specialized Tools
- **Hardware Passthrough Detection**: IOMMU and VFIO analysis
- **Security Analysis**: Container and VM security assessment
- **Resource Usage Tracking**: CPU, memory, disk, and network monitoring
- **Configuration Backup**: Unraid-specific configuration collection

## 🚀 Installation

### Manual Installation
1. Download the plugin files to your Unraid server
2. Copy the plugin directory to `/usr/local/emhttp/plugins/unraid_snapshot_exporter/`
3. Make scripts executable:
   ```bash
   chmod +x /usr/local/emhttp/plugins/unraid_snapshot_exporter/scripts/*.sh
   chmod +x /usr/local/emhttp/plugins/unraid_snapshot_exporter/scripts/*.py
   ```
4. Install Python dependencies:
   ```bash
   pip3 install psutil
   ```

### Plugin Installation
1. Download the `.plg` file
2. Install via Unraid's Plugin Manager
3. Access via the Unraid web interface

## 📊 Diagnostic Tools

### 1. Full System Snapshot
Generates a comprehensive snapshot including:
- System architecture and hardware information
- Performance metrics and resource usage
- Network configuration and connectivity
- Docker containers and images
- Virtual machines and configurations
- System logs and error messages
- Unraid-specific configurations

**Usage:**
```bash
bash /usr/local/emhttp/plugins/unraid_snapshot_exporter/scripts/generate_snapshot.sh
```

### 2. System Analysis
Advanced Python-based analysis providing structured data:
- JSON-formatted system information
- Performance metrics with timestamps
- Hardware capabilities assessment
- Storage system analysis
- Network interface details
- Docker environment analysis
- VM resource allocation

**Usage:**
```bash
python3 /usr/local/emhttp/plugins/unraid_snapshot_exporter/scripts/system_analyzer.py /tmp/analysis_output
```

### 3. Performance Collection
Time-series performance data collection:
- CPU usage per core over time
- Memory utilization patterns
- Disk I/O statistics
- Network interface performance
- Docker container metrics
- VM resource consumption

**Usage:**
```bash
bash /usr/local/emhttp/plugins/unraid_snapshot_exporter/scripts/performance_collector.sh [duration] [interval]
```

### 4. Network Diagnostics
Comprehensive network analysis:
- Interface configuration and status
- Routing table analysis
- Connectivity testing (ping, traceroute)
- Firewall rule inspection
- Docker network analysis
- VM network configuration

**Usage:**
```bash
bash /usr/local/emhttp/plugins/unraid_snapshot_exporter/scripts/network_diagnostics.sh
```

### 5. Docker Analysis
Detailed Docker environment analysis:
- Container status and resource usage
- Image analysis and history
- Network configuration
- Volume management
- Security assessment
- Performance metrics

**Usage:**
```bash
bash /usr/local/emhttp/plugins/unraid_snapshot_exporter/scripts/docker_analyzer.sh
```

### 6. VM Analysis
Virtual machine comprehensive analysis:
- VM status and resource allocation
- Network interface configuration
- Storage pool analysis
- Hardware passthrough detection
- Security assessment
- Performance monitoring

**Usage:**
```bash
bash /usr/local/emhttp/plugins/unraid_snapshot_exporter/scripts/vm_analyzer.sh
```

## 🎨 Web Interface

The plugin includes a modern web GUI with:
- **Psychedelic/Tribal Aesthetics**: Custom color scheme and animations
- **Real-time System Overview**: Live system statistics
- **Interactive Diagnostic Tools**: One-click analysis execution
- **Recent Snapshots**: Download previous diagnostic files
- **Responsive Design**: Works on desktop and mobile devices

### Features
- Animated cosmic particle effects
- Real-time system stat updates
- Form validation and loading states
- Keyboard shortcuts (Ctrl+S for snapshot, Ctrl+A for analysis)
- Smooth animations and transitions

## 📁 Output Structure

### Snapshot Contents
```
unraid_snapshot_YYYYMMDD_HHMMSS/
├── system/
│   ├── cpuinfo.txt
│   ├── meminfo.txt
│   ├── pci_devices.txt
│   ├── block_devices.txt
│   └── unraid_config.txt
├── performance/
│   ├── top_*.txt
│   ├── memory_*.txt
│   ├── disk_usage_*.txt
│   └── iostat.txt
├── network/
│   ├── interfaces.txt
│   ├── routing.txt
│   ├── connectivity_tests.txt
│   └── iptables.txt
├── docker/
│   ├── containers.txt
│   ├── images.txt
│   ├── logs_*.txt
│   └── stats.txt
├── vm/
│   ├── vm_list.txt
│   ├── vm_*_info.txt
│   └── vm_*_config.xml
├── logs/
│   ├── journal.txt
│   ├── dmesg.txt
│   └── syslog.txt
├── configs/
│   ├── fstab
│   ├── hosts
│   └── docker-compose files
└── summary.txt
```

### Analysis Outputs
- **system_analysis.json**: Structured system data
- **performance_data_*/**: Time-series performance metrics
- **network_diagnostics_*/**: Network analysis results
- **docker_analysis_*/**: Docker environment details
- **vm_analysis_*/**: Virtual machine information

## 🔧 Requirements

### System Requirements
- Unraid 6.8+ (tested on 6.12+)
- Python 3.6+ with psutil module
- Standard Linux tools (ip, ss, docker, virsh)
- Sufficient disk space for snapshots

### Dependencies
```bash
# Python dependencies
pip3 install psutil

# Optional but recommended
apt-get install sysstat iotop htop
```

## 🎯 Use Cases

### For LLM Analysis
The structured output is specifically designed for large-context LLMs:
- **JSON-formatted data** for easy parsing
- **Comprehensive system context** for informed recommendations
- **Performance baselines** for optimization suggestions
- **Security assessments** for vulnerability analysis
- **Configuration analysis** for best practices

### For System Optimization
- **Performance bottlenecks** identification
- **Resource utilization** analysis
- **Network optimization** recommendations
- **Docker container** optimization
- **VM resource allocation** tuning

### For Troubleshooting
- **System diagnostics** for issue identification
- **Log analysis** for error tracking
- **Configuration validation** for setup issues
- **Hardware compatibility** verification
- **Security assessment** for vulnerabilities

## 🛠️ Advanced Usage

### Custom Analysis Scripts
You can extend the functionality by creating custom analysis scripts:

```bash
#!/bin/bash
# Custom analysis script
OUTPUT_DIR="/tmp/custom_analysis_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$OUTPUT_DIR"

# Your custom analysis here
echo "Custom analysis complete: $OUTPUT_DIR"
```

### Integration with Monitoring Systems
The JSON outputs can be integrated with monitoring systems:
- **Prometheus** metrics export
- **Grafana** dashboard data
- **ELK Stack** log analysis
- **Custom monitoring** solutions

## 🔒 Security Considerations

- **No sensitive data collection** by default
- **Local processing only** - no external data transmission
- **Configurable scope** for data collection
- **Secure file permissions** on output files
- **Optional encryption** for sensitive snapshots

## 📈 Performance Impact

- **Minimal resource usage** during collection
- **Configurable collection intervals** to reduce impact
- **Background processing** for long-running analyses
- **Automatic cleanup** of temporary files
- **Compressed output** to minimize storage requirements

## 🤝 Contributing

Contributions are welcome! Areas for improvement:
- Additional diagnostic tools
- Enhanced web interface features
- Integration with external monitoring systems
- Performance optimizations
- Additional output formats

## 📄 License

MIT License - see LICENSE file for details.

## 🙏 Acknowledgments

- Unraid community for the excellent platform
- Python psutil library for system metrics
- Font Awesome for the beautiful icons
- The open-source community for inspiration

---

**Created with ❤️ for the Unraid community by Mitch**

*"In the vast digital cosmos, every system tells a story. This tool helps you read those stories."*

## Files from Original Repository

- `LICENSE`
- `README.md` - Documentation
- `plugin/unraid_snapshot_exporter_minimal.plg`
- `plugin/unraid_snapshot_exporter.plg`
- `plugin/unraid_snapshot_exporter_simple.plg`
- `scripts/docker_analyzer.sh` - Script/executable
- `scripts/system_analyzer.py` - Script/executable
- `scripts/network_diagnostics.sh` - Script/executable
- `scripts/generate_snapshot.sh` - Script/executable
- `scripts/performance_collector.sh` - Script/executable
- `scripts/vm_analyzer.sh` - Script/executable
- `.git/config`
- `.git/HEAD`
- `.git/description`
- `.git/index`
- `.git/packed-refs`
- `.git/COMMIT_EDITMSG`
- `.git/FETCH_HEAD`
- `webgui/download.php`
- `webgui/index.php`
- `webgui/script.js` - Script/executable
- `webgui/style.css`
- `.git/info/exclude`
- `.git/logs/HEAD`
- `.git/hooks/commit-msg.sample`
- `.git/hooks/pre-rebase.sample`
- `.git/hooks/pre-commit.sample`
- `.git/hooks/applypatch-msg.sample`
- `.git/hooks/fsmonitor-watchman.sample`
- `.git/hooks/pre-receive.sample`
- `.git/hooks/prepare-commit-msg.sample`
- `.git/hooks/post-update.sample`
- `.git/hooks/pre-merge-commit.sample`
- `.git/hooks/pre-applypatch.sample`
- `.git/hooks/pre-push.sample`
- `.git/hooks/update.sample`
- `.git/hooks/push-to-checkout.sample`
- `.git/objects/ca/3f67c85ae489fc7e3c9b7adb04ac7d52431835`
- `.git/objects/pack/pack-fc99c98a2783359411239d5208ba4e48c86ae796.idx`
- `.git/objects/pack/pack-fc99c98a2783359411239d5208ba4e48c86ae796.pack`
- `.git/objects/29/da555f65d14871c2e97c0e67e6a246ab9788a3`
- `.git/objects/77/9dd2dd6c6267fb98f594470c48de9c037010fd`
- `.git/logs/refs/heads/main`
- `.git/logs/refs/remotes/origin/HEAD`
- `.git/logs/refs/remotes/origin/main`
- `.git/refs/heads/main`
- `.git/refs/remotes/origin/HEAD`
- `.git/refs/remotes/origin/main`

---

*Repository consolidated on 2025-09-23 - originally located at `unraid_snapshot_exporter/`*

**Note:** This utility was small enough to be documented here rather than maintained as a separate repository. 
If active development resumes, consider recreating as a standalone repository.
