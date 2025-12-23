# Unraid Configuration Backup & Analysis System

A comprehensive backup and analysis toolkit for Unraid servers, designed to capture system configurations and provide optimization recommendations.

## 🚀 Features

- **Comprehensive Backup**: Captures system configs, Docker containers, VMs, network settings, and performance data
- **Git-Friendly**: Creates version-controlled backups with sensitive data filtering
- **Performance Analysis**: Analyzes system metrics and provides optimization recommendations
- **Structured Output**: Organizes data for easy analysis and comparison
- **Automated Reports**: Generates detailed system analysis reports

## 📁 Project Structure

```
unraid_config_backup/
├── backup_unraid.sh              # Basic backup script
├── backup_unraid_enhanced.sh     # Enhanced backup script (recommended)
├── analyze_unraid.py             # Basic analysis tool
├── analyze_unraid_enhanced.py    # Enhanced analysis tool with HTML reports
├── monitor_unraid.py             # Continuous system monitoring
├── config.yaml                   # Configuration file
├── setup.sh                      # Setup script
├── README.md                     # This file
├── requirements.txt              # Python dependencies
├── backups/                      # Generated backup directories
│   ├── unraid_backup_YYYYMMDD_HHMMSS/
│   └── git_unraid_backup_YYYYMMDD_HHMMSS/
└── test_backup.sh               # Test script
```

## 🛠️ Quick Start

### 1. Run the setup script
```bash
./setup.sh
```

### 2. Run the enhanced backup (recommended as root for full access)
```bash
sudo ./backup_unraid_enhanced.sh
```

### 3. Analyze the backup data with HTML report
```bash
python3 analyze_unraid_enhanced.py backups/unraid_backup_YYYYMMDD_HHMMSS/ --html report.html
```

### 4. Test system monitoring
```bash
python3 monitor_unraid.py --once
```

### 5. Commit to git for version control
```bash
git add backups/git_unraid_backup_YYYYMMDD_HHMMSS/
git commit -m "Unraid backup $(date)"
```

## 📊 What Gets Backed Up

### System Configuration
- `/boot` directory (Unraid flash drive configs)
- System configuration files (`/etc/fstab`, `/etc/hosts`, etc.)
- Unraid-specific configuration files
- Kernel and module information

### Docker Environment
- Container list and status
- Docker images and networks
- Volume configurations
- Docker Compose files (if found)

### Virtual Machines
- VM list and status
- VM XML configurations
- Libvirt settings

### Network Configuration
- Network interfaces and IP addresses
- Routing tables
- Firewall rules
- DNS configuration

### Performance Data
- CPU usage and load averages
- Memory usage statistics
- Storage I/O metrics
- Network I/O statistics

### System Logs
- Recent system logs (last 1000 lines)
- Docker container logs
- VM-related logs

## 🔍 Analysis Features

The enhanced analysis tools provide:

- **System Overview**: CPU, memory, and detailed system information
- **Storage Analysis**: Filesystem usage, SMART data, and storage recommendations
- **Docker Analysis**: Container status, resource usage, and cleanup suggestions
- **Network Analysis**: Interface status, bonding, and network configuration
- **Performance Metrics**: Load averages, process analysis, and system performance
- **Enhanced Recommendations**: Severity-based optimization suggestions
- **HTML Reports**: Beautiful, interactive HTML reports with charts
- **Trend Analysis**: Historical data analysis and performance trends
- **Continuous Monitoring**: Real-time system health monitoring with alerts

## 📈 Sample Analysis Output

```
🚀 UNRAID SYSTEM ANALYSIS REPORT
============================================================

📋 SYSTEM OVERVIEW:
   Hostname: unraid-server
   CPU: Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz
   Cores: 12
   Memory: 32G

💾 STORAGE OVERVIEW:
   /dev/md1: 85% used (2.1T/2.5T)
   /dev/sda1: 45% used (45G/100G)

🐳 DOCKER OVERVIEW:
   Total Containers: 15
   Running: 12
   Stopped: 3

🌐 NETWORK OVERVIEW:
   eth0: UP (192.168.1.100)
   br0: UP (192.168.1.101)

⚡ PERFORMANCE OVERVIEW:
   Load Average: 2.34 (1m), 1.89 (5m), 1.45 (15m)
   Uptime: 15 days, 3 hours

💡 OPTIMIZATION RECOMMENDATIONS:
   1. ⚠️  Storage warning: /dev/md1 is 85% full
   2. 🐳 Found 3 stopped Docker containers - consider cleanup
   3. 📈 Moderate system load - monitor resource usage
```

## 🔧 Advanced Usage

### Configuration Management
```bash
# Edit configuration file
nano config.yaml

# Use custom config for backup
CONFIG_FILE="custom_config.yaml" ./backup_unraid_enhanced.sh

# Use custom config for analysis
python3 analyze_unraid_enhanced.py backup_path/ --config custom_config.yaml
```

### Enhanced Analysis with HTML Reports
```bash
# Generate HTML report
python3 analyze_unraid_enhanced.py backups/unraid_backup_YYYYMMDD_HHMMSS/ --html report.html

# Generate JSON report
python3 analyze_unraid_enhanced.py backups/unraid_backup_YYYYMMDD_HHMMSS/ --output analysis_report.json
```

### Continuous Monitoring
```bash
# Start continuous monitoring
python3 monitor_unraid.py

# Run monitoring once
python3 monitor_unraid.py --once

# Custom monitoring interval (30 seconds)
python3 monitor_unraid.py --interval 30
```

### Automated Scheduling
Add to crontab for regular backups and monitoring:
```bash
# Daily enhanced backup at 2 AM
0 2 * * * /path/to/backup_unraid_enhanced.sh

# Weekly analysis with HTML report
0 3 * * 0 python3 /path/to/analyze_unraid_enhanced.py /path/to/latest/backup/ --html weekly_report.html

# Continuous monitoring (every 5 minutes)
*/5 * * * * python3 /path/to/monitor_unraid.py --once
```

## 🛡️ Security Considerations

- Sensitive data (user passwords, API keys) is automatically filtered
- Git-friendly backups exclude logs and performance data
- Root access recommended for complete system access
- Backup directories include `.gitignore` files for sensitive data

## 🔄 Version Control Integration

The system creates two backup types:

1. **Full Backup**: Complete system snapshot with all data
2. **Git Backup**: Clean version suitable for version control

Git backups include:
- Configuration files
- Analysis summaries
- System information
- Optimization recommendations

## 🎯 Optimization Opportunities

The analysis tool identifies:

- **Storage Issues**: High usage, fragmentation, performance bottlenecks
- **Resource Utilization**: CPU, memory, and network optimization
- **Docker Optimization**: Container cleanup, resource limits, image management
- **Network Configuration**: Interface optimization, redundancy planning
- **System Performance**: Load balancing, process optimization

## 🤝 Contributing

This system is designed for high-performance Unraid setups with:
- NVMe storage arrays
- GPU passthrough configurations
- Advanced Docker deployments
- Multi-VM environments

## 📝 Requirements

- **Unraid 6.8+** (tested on latest versions)
- **Bash 4.0+** for backup script
- **Python 3.6+** for analysis tool
- **Root access** (recommended for full backup)

## 🚨 Troubleshooting

### Common Issues

1. **Permission Denied**: Run with `sudo` for full access
2. **Docker Not Available**: Script handles missing Docker gracefully
3. **Large Backup Size**: Use git-friendly version for version control
4. **Analysis Errors**: Check Python dependencies and file permissions

### Debug Mode
```bash
# Enable verbose output
bash -x backup_unraid.sh
```

## 📞 Support

This system is designed for advanced Unraid users who want:
- Comprehensive system monitoring
- Performance optimization insights
- Version-controlled configuration management
- Automated backup and analysis workflows

Perfect for your high-performance Unraid setup with NVMe storage and GPU passthrough! 🚀