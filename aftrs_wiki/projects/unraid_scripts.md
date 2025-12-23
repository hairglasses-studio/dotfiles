# UNRAID Scripts - Server Automation and Maintenance

## Overview
UNRAID Scripts is a comprehensive automation and maintenance system for UNRAID server management. It provides modular tools for system maintenance, backup automation, media processing, and server health monitoring within the AFTRS infrastructure.

## 🏗️ Architecture

### Core Components
- **Maintenance Scripts**: Automated system maintenance and health checks
- **CLI Tools**: Command-line utilities for server management
- **Cron Scripts**: Scheduled automation and backup tasks
- **Media Processing**: Automated media handling and organization
- **Backup Automation**: System backup and recovery procedures
- **Health Monitoring**: Server performance and status tracking

### Key Features
- **Automated Maintenance**: Scheduled system cleanup and optimization
- **Backup Management**: Automated backup to UNRAID storage
- **Media Processing**: FFmpeg-based media tagging and organization
- **Sora Integration**: OpenAI Sora video backup and archiving
- **Health Monitoring**: System performance and status tracking
- **CLI Tools**: Unified command-line interface for server management

## 📁 Project Structure

```
unraid_scripts/
├── 📖 README.md                    # Project documentation
├── 📁 maintenance_scripts/         # System maintenance
│   ├── run_sora_backup.sh         # Sora video backup
│   ├── backup_to_unraid.sh        # UNRAID backup automation
│   ├── auth_tokens.py             # Authentication management
│   ├── tag_with_ffmpeg.py         # Media tagging with FFmpeg
│   ├── sora_downloader.py         # Sora video downloader
│   └── sora_archive_schema.sql    # Database schema
├── 📁 cli_tools/                   # Command-line utilities
├── 📁 cron_scripts/                # Scheduled automation
└── 📁 profiles/                    # Configuration profiles
```

## 🚀 Installation

### Prerequisites
- **UNRAID Server**: 6.9+ with Docker support
- **SSH Access**: Root access to UNRAID server
- **Docker**: Container support for automation
- **Python 3.9+**: For Python-based scripts
- **FFmpeg**: Media processing capabilities

### Installation Process
```bash
# Clone repository
git clone https://github.com/hairglasses/unraid_scripts.git
cd unraid_scripts

# Install dependencies
./install.sh

# Configure environment
cp .env.example .env
nano .env

# Setup cron jobs
./setup_cron.sh
```

### Docker Installation
```bash
# Using Docker Compose
docker-compose up -d

# Or build locally
docker build -t unraid_scripts .
docker run -d --name unraid_scripts unraid_scripts
```

## 🔧 Configuration

### Environment Setup
```bash
# UNRAID Configuration
UNRAID_HOST=192.168.1.100
UNRAID_USER=root
UNRAID_SHARE=/mnt/user/backups

# Backup Configuration
BACKUP_SOURCE=/mnt/cache/appdata
BACKUP_DESTINATION=/mnt/user/backups
BACKUP_RETENTION=30

# Media Processing
MEDIA_SOURCE=/mnt/user/media
MEDIA_DESTINATION=/mnt/user/processed
FFMPEG_PATH=/usr/bin/ffmpeg

# Sora Configuration
SORA_API_KEY=your-sora-api-key
SORA_DOWNLOAD_PATH=/mnt/user/sora
SORA_DATABASE=postgresql://user:pass@localhost/sora
```

### Cron Configuration
```bash
# Daily maintenance
0 2 * * * /usr/local/bin/unraid_maintenance.sh

# Weekly backup
0 3 * * 0 /usr/local/bin/unraid_backup.sh

# Hourly health check
0 * * * * /usr/local/bin/unraid_health.sh

# Daily Sora backup
0 4 * * * /usr/local/bin/sora_backup.sh
```

## 📊 Features

### Core Commands

#### System Maintenance
```bash
# Run maintenance tasks
unraid_maintenance.sh

# System health check
unraid_health.sh

# Performance monitoring
unraid_performance.sh

# Cleanup old files
unraid_cleanup.sh
```

#### Backup Management
```bash
# Create backup
unraid_backup.sh

# Restore from backup
unraid_restore.sh <backup_name>

# List backups
unraid_backup.sh --list

# Backup verification
unraid_backup.sh --verify
```

#### Media Processing
```bash
# Tag media files
tag_with_ffmpeg.py --input /path/to/media

# Process media batch
process_media.sh --batch

# Organize media library
organize_media.sh

# Extract metadata
extract_metadata.sh
```

#### Sora Integration
```bash
# Backup Sora videos
run_sora_backup.sh

# Download Sora content
sora_downloader.py --url <sora_url>

# Archive Sora metadata
sora_archive.py

# Sora health check
sora_health.sh
```

#### CLI Tools
```bash
# System status
unraid status

# Performance metrics
unraid performance

# Storage usage
unraid storage

# Docker status
unraid docker
```

### Advanced Features

#### Automated Workflows
```bash
# Daily maintenance workflow
unraid_maintenance.sh --full

# Weekly backup workflow
unraid_backup.sh --weekly

# Monthly cleanup workflow
unraid_cleanup.sh --monthly

# Health monitoring workflow
unraid_health.sh --continuous
```

#### Media Management
```bash
# Batch media processing
process_media_batch.sh

# Media deduplication
deduplicate_media.sh

# Media organization
organize_media_library.sh

# Media backup
backup_media.sh
```

## 🔄 Workflow

### Initial Setup
1. **Install Scripts**: `./install.sh`
2. **Configure Environment**: Edit `.env` file
3. **Setup Cron Jobs**: `./setup_cron.sh`
4. **Test Installation**: `unraid_health.sh`

### Daily Operations
1. **Health Check**: `unraid_health.sh`
2. **Maintenance**: `unraid_maintenance.sh`
3. **Backup Check**: `unraid_backup.sh --status`
4. **Media Processing**: `process_media.sh`

### Weekly Operations
1. **Full Backup**: `unraid_backup.sh --weekly`
2. **Deep Cleanup**: `unraid_cleanup.sh --deep`
3. **Performance Review**: `unraid_performance.sh --report`
4. **Sora Backup**: `run_sora_backup.sh`

### Monthly Operations
1. **System Audit**: `unraid_audit.sh`
2. **Storage Optimization**: `unraid_optimize.sh`
3. **Backup Verification**: `unraid_backup.sh --verify-all`
4. **Media Archive**: `archive_media.sh`

## 📈 Current Status

### ✅ Completed Features
- [x] Automated maintenance scripts
- [x] Backup automation system
- [x] Media processing with FFmpeg
- [x] Sora video backup integration
- [x] Health monitoring and reporting
- [x] Cron job automation
- [x] CLI tools for server management
- [x] Docker containerization
- [x] Performance monitoring
- [x] Storage optimization

### 🔄 In Progress
- [ ] Advanced analytics dashboard
- [ ] Real-time monitoring
- [ ] Automated failover
- [ ] Multi-server management

### 📋 Planned Features
- [ ] Web dashboard interface
- [ ] Advanced backup strategies
- [ ] Media transcoding pipeline
- [ ] Automated troubleshooting
- [ ] Performance optimization
- [ ] Security auditing tools

## 🛠️ Technical Decisions

### Architecture Design
- **Script-Based**: Modular bash and Python scripts
- **Cron-Driven**: Scheduled automation
- **Docker-Ready**: Container deployment support
- **CLI-Centric**: Command-line automation

### Backup Strategy
- **Incremental Backups**: Efficient storage usage
- **Retention Policies**: Automated cleanup
- **Verification**: Backup integrity checking
- **Recovery Testing**: Automated restore testing

### Media Processing
- **FFmpeg Integration**: Professional media handling
- **Batch Processing**: Efficient large-scale operations
- **Metadata Preservation**: Maintain file information
- **Organization**: Automated library management

## 🔍 Troubleshooting

### Common Issues

#### Backup Problems
```bash
# Check backup status
unraid_backup.sh --status

# Verify backup integrity
unraid_backup.sh --verify

# Check storage space
df -h /mnt/user/backups

# Test backup restore
unraid_backup.sh --test-restore
```

#### Media Processing Issues
```bash
# Check FFmpeg installation
ffmpeg -version

# Test media processing
tag_with_ffmpeg.py --test

# Check media permissions
ls -la /mnt/user/media/

# Verify media paths
unraid_media.sh --verify-paths
```

#### Sora Integration Issues
```bash
# Check Sora API key
echo $SORA_API_KEY

# Test Sora connection
sora_downloader.py --test

# Check Sora database
psql -d sora -c "SELECT COUNT(*) FROM videos;"

# Verify Sora backup
run_sora_backup.sh --verify
```

#### Performance Issues
```bash
# Check system resources
unraid_performance.sh

# Monitor disk usage
df -h

# Check CPU usage
htop

# Monitor memory usage
free -h
```

### Debug Mode
```bash
# Enable debug logging
export UNRAID_DEBUG=1
unraid_maintenance.sh

# Verbose output
unraid_health.sh --verbose
```

## 📚 Related Documentation

- [UNRAID Server Configuration](../infrastructure/unraid.md)
- [Backup Strategies](../patterns/backup_strategies.md)
- [Media Processing Workflows](../patterns/media_processing.md)
- [Automation Patterns](../patterns/automation.md)

## 🔗 Integration Points

### With Other Systems
- **AgentCTL**: Server status integration
- **CR8 CLI**: Media processing workflows
- **OAI CLI**: Content backup integration
- **AFTRS CLI**: Network connectivity

### External Services
- **UNRAID API**: Server management
- **PostgreSQL**: Database storage
- **FFmpeg**: Media processing
- **Cron**: Scheduled automation

---

*Last Updated: 2025-01-14*
*Status: Production Ready*
*Maintainer: AFTRS Development Team* 