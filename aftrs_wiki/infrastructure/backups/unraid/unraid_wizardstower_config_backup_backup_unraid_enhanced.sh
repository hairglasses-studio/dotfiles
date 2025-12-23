#!/bin/bash

# Enhanced Unraid Configuration Backup Script
# Features: Configuration support, parallel processing, compression, monitoring

set -euo pipefail

# Load configuration
CONFIG_FILE="config.yaml"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default configuration
DEFAULT_CONFIG='
backup:
  directory: "backups"
  include_patterns: ["*.conf", "*.ini", "*.xml", "*.yml", "*.yaml", "docker-compose.*"]
  exclude_patterns: ["*.log", "*.tmp", "*.cache", "*.pid", "__pycache__", "node_modules"]
  sensitive_files: ["users.ini", "state.ini", "var.ini", "passwd", "shadow", "*.key", "*.pem", "*.crt"]
analysis:
  thresholds:
    storage_warning: 80
    storage_critical: 90
    memory_warning: 80
    memory_critical: 90
    load_warning: 1.0
    load_critical: 2.0
performance:
  collection:
    cpu_samples: 3
    memory_samples: 3
    disk_samples: 3
    network_samples: 3
  logs:
    system_lines: 1000
    docker_lines: 100
    vm_lines: 50
advanced:
  parallel_backup: true
  max_workers: 4
  compress_backups: true
  compression_level: 6
  retention:
    days: 30
    max_backups: 10
'

# Load configuration from file or use defaults
load_config() {
    if [ -f "$CONFIG_FILE" ]; then
        # Simple YAML parsing (basic implementation)
        BACKUP_DIR=$(grep -A1 "directory:" "$CONFIG_FILE" | tail -1 | sed 's/^[[:space:]]*//')
        PARALLEL_BACKUP=$(grep "parallel_backup:" "$CONFIG_FILE" | grep -o "true\|false" || echo "true")
        MAX_WORKERS=$(grep "max_workers:" "$CONFIG_FILE" | grep -o "[0-9]*" || echo "4")
        COMPRESS_BACKUPS=$(grep "compress_backups:" "$CONFIG_FILE" | grep -o "true\|false" || echo "true")
        COMPRESSION_LEVEL=$(grep "compression_level:" "$CONFIG_FILE" | grep -o "[0-9]*" || echo "6")
        RETENTION_DAYS=$(grep -A1 "retention:" "$CONFIG_FILE" | grep "days:" | grep -o "[0-9]*" || echo "30")
        MAX_BACKUPS=$(grep -A1 "retention:" "$CONFIG_FILE" | grep "max_backups:" | grep -o "[0-9]*" || echo "10")
    else
        # Use defaults
        BACKUP_DIR="backups"
        PARALLEL_BACKUP="true"
        MAX_WORKERS="4"
        COMPRESS_BACKUPS="true"
        COMPRESSION_LEVEL="6"
        RETENTION_DAYS="30"
        MAX_BACKUPS="10"
    fi
    
    # Ensure BACKUP_DIR is relative to script directory
    if [[ "$BACKUP_DIR" != /* ]]; then
        BACKUP_DIR="$SCRIPT_DIR/$BACKUP_DIR"
    fi
}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Enhanced logging with timestamps and levels
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] [INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] [WARN]${NC} $1"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] [ERROR]${NC} $1"
}

debug() {
    if [ "${DEBUG:-false}" = "true" ]; then
        echo -e "${CYAN}[$(date +'%Y-%m-%d %H:%M:%S')] [DEBUG]${NC} $1"
    fi
}

# Performance monitoring
start_timer() {
    START_TIME=$(date +%s)
}

end_timer() {
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    log "Operation completed in ${DURATION} seconds"
}

# Cleanup old backups
cleanup_old_backups() {
    log "Cleaning up old backups..."
    
    # Remove backups older than retention days
    find "$BACKUP_DIR" -name "unraid_backup_*" -type d -mtime +$RETENTION_DAYS -exec rm -rf {} \; 2>/dev/null || true
    
    # Keep only the most recent backups
    local backup_count=$(find "$BACKUP_DIR" -name "unraid_backup_*" -type d | wc -l)
    if [ "$backup_count" -gt "$MAX_BACKUPS" ]; then
        find "$BACKUP_DIR" -name "unraid_backup_*" -type d -printf '%T@ %p\n' | sort -n | head -n $((backup_count - MAX_BACKUPS)) | cut -d' ' -f2- | xargs rm -rf
    fi
    
    log "Cleanup completed"
}

# Enhanced backup structure creation
create_backup_structure() {
    log "Creating enhanced backup directory structure..."
    
    # Create main directories
    mkdir -p "${FULL_BACKUP_PATH}"/{configs,performance,logs,analysis,metadata}
    mkdir -p "${FULL_BACKUP_PATH}/configs"/{system,docker,vm,network,storage,plugins}
    mkdir -p "${FULL_BACKUP_PATH}/performance"/{cpu,memory,storage,network,processes}
    mkdir -p "${FULL_BACKUP_PATH}/logs"/{system,docker,vm,plugins}
    mkdir -p "${FULL_BACKUP_PATH}/analysis"/{reports,recommendations,trends}
    mkdir -p "${FULL_BACKUP_PATH}/metadata"/{checksums,indexes,statistics}
    
    # Create metadata file
    cat > "${FULL_BACKUP_PATH}/metadata/backup_info.json" << EOF
{
  "backup_date": "$(date -Iseconds)",
  "hostname": "$(hostname)",
  "unraid_version": "$(cat /proc/version 2>/dev/null || echo 'Unknown')",
  "kernel_version": "$(uname -r)",
  "script_version": "2.0.0",
  "config_file": "$CONFIG_FILE",
  "parallel_backup": $PARALLEL_BACKUP,
  "max_workers": $MAX_WORKERS,
  "compression_enabled": $COMPRESS_BACKUPS,
  "compression_level": $COMPRESSION_LEVEL
}
EOF
}

# Enhanced system information collection
collect_system_info() {
    log "Collecting comprehensive system information..."
    
    # Basic system info with more details
    cat > "${FULL_BACKUP_PATH}/system_info.txt" << EOF
=== ENHANCED UNRAID SYSTEM INFORMATION ===
Backup Date: $(date -Iseconds)
Hostname: $(hostname)
Kernel: $(uname -r)
Architecture: $(uname -m)
Uname: $(uname -a)
Boot Time: $(uptime -s)
System Load: $(uptime | awk -F'load average:' '{print $2}')
EOF

    # Enhanced CPU information
    log "Collecting detailed CPU information..."
    cat > "${FULL_BACKUP_PATH}/performance/cpu/cpu_info.txt" << EOF
=== ENHANCED CPU INFORMATION ===
CPU Model: $(grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
CPU Cores: $(nproc)
CPU Threads: $(grep -c processor /proc/cpuinfo)
CPU Architecture: $(uname -m)
CPU Frequency: $(grep "cpu MHz" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs) MHz
CPU Cache: $(grep "cache size" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
CPU Flags: $(grep "flags" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
EOF

    # Enhanced memory information
    log "Collecting detailed memory information..."
    cat > "${FULL_BACKUP_PATH}/performance/memory/memory_info.txt" << EOF
=== ENHANCED MEMORY INFORMATION ===
$(free -h)
$(cat /proc/meminfo | grep -E "(MemTotal|MemFree|MemAvailable|SwapTotal|SwapFree|Buffers|Cached|Active|Inactive)")
Memory NUMA Info:
$(numactl --hardware 2>/dev/null || echo "NUMA not available")
EOF

    # Enhanced storage information
    log "Collecting detailed storage information..."
    cat > "${FULL_BACKUP_PATH}/performance/storage/storage_info.txt" << EOF
=== ENHANCED STORAGE INFORMATION ===
Block Devices:
$(lsblk -f -o NAME,SIZE,FSTYPE,MOUNTPOINT,LABEL,UUID)

Filesystem Usage:
$(df -h)

RAID Status:
$(cat /proc/mdstat 2>/dev/null || echo "No RAID arrays found")

Disk I/O Statistics:
$(iostat -x 1 1 2>/dev/null || echo "iostat not available")

Smart Drive Information:
$(for disk in /dev/sd*; do if [ -b "$disk" ]; then echo "=== $disk ==="; smartctl -a "$disk" 2>/dev/null | head -20; fi; done)
EOF
}

# Parallel backup function
parallel_backup() {
    local function_name="$1"
    local max_jobs="$2"
    shift 2
    
    if [ "$PARALLEL_BACKUP" = "true" ]; then
        # Run function in background with job control
        "$function_name" "$@" &
        local job_pid=$!
        
        # Limit concurrent jobs
        while [ $(jobs -r | wc -l) -ge "$max_jobs" ]; do
            sleep 0.1
        done
    else
        # Run synchronously
        "$function_name" "$@"
    fi
}

# Enhanced Unraid configuration backup
backup_unraid_configs() {
    log "Backing up enhanced Unraid configuration files..."
    
    # Flash drive configs with better error handling
    if [ -d "/boot" ]; then
        log "Backing up /boot configuration..."
        rsync -av --exclude="*.log" --exclude="*.tmp" /boot/ "${FULL_BACKUP_PATH}/configs/system/boot/" 2>/dev/null || warn "Could not backup /boot directory"
    fi
    
    # Extended system configuration files
    local config_files=(
        "/etc/fstab"
        "/etc/hosts"
        "/etc/resolv.conf"
        "/etc/network/interfaces"
        "/etc/sysctl.conf"
        "/etc/modules"
        "/etc/rc.local"
        "/etc/crontab"
        "/etc/ssh/sshd_config"
        "/etc/ssh/ssh_config"
    )
    
    for file in "${config_files[@]}"; do
        if [ -f "$file" ]; then
            cp "$file" "${FULL_BACKUP_PATH}/configs/system/"
        fi
    done
    
    # Enhanced Unraid specific files
    local unraid_files=(
        "/var/local/emhttp/var.ini"
        "/var/local/emhttp/state.ini"
        "/var/local/emhttp/disk.ini"
        "/var/local/emhttp/array.ini"
        "/var/local/emhttp/shares.ini"
        "/var/local/emhttp/users.ini"
        "/var/local/emhttp/settings.ini"
        "/var/local/emhttp/plugins.ini"
        "/var/local/emhttp/ident.cfg"
    )
    
    for file in "${unraid_files[@]}"; do
        if [ -f "$file" ]; then
            cp "$file" "${FULL_BACKUP_PATH}/configs/system/"
        fi
    done
    
    # Backup plugin configurations
    if [ -d "/var/local/emhttp/plugins" ]; then
        log "Backing up plugin configurations..."
        rsync -av --exclude="*.log" /var/local/emhttp/plugins/ "${FULL_BACKUP_PATH}/configs/plugins/" 2>/dev/null || warn "Could not backup plugin configs"
    fi
}

# Enhanced Docker configuration backup
backup_docker_configs() {
    log "Backing up enhanced Docker configurations..."
    
    # Docker system info with more details
    docker system info > "${FULL_BACKUP_PATH}/configs/docker/docker_info.txt" 2>/dev/null || warn "Docker not available or not running"
    
    # Enhanced container information
    docker ps -a --format "table {{.ID}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.Size}}\t{{.Names}}" > "${FULL_BACKUP_PATH}/configs/docker/containers.txt" 2>/dev/null || warn "Could not get container list"
    
    # Docker images with size information
    docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.Size}}\t{{.CreatedAt}}" > "${FULL_BACKUP_PATH}/configs/docker/images.txt" 2>/dev/null || warn "Could not get image list"
    
    # Docker networks
    docker network ls --format "table {{.ID}}\t{{.Name}}\t{{.Driver}}\t{{.Scope}}" > "${FULL_BACKUP_PATH}/configs/docker/networks.txt" 2>/dev/null || warn "Could not get network list"
    
    # Docker volumes with details
    docker volume ls --format "table {{.Name}}\t{{.Driver}}" > "${FULL_BACKUP_PATH}/configs/docker/volumes.txt" 2>/dev/null || warn "Could not get volume list"
    
    # Docker system df (disk usage)
    docker system df > "${FULL_BACKUP_PATH}/configs/docker/disk_usage.txt" 2>/dev/null || warn "Could not get disk usage"
    
    # Docker compose files with better search
    find /mnt -name "docker-compose.yml" -o -name "docker-compose.yaml" -o -name "docker-compose.override.yml" 2>/dev/null | while read file; do
        cp "$file" "${FULL_BACKUP_PATH}/configs/docker/"
    done
    
    # Docker daemon configuration
    if [ -f "/etc/docker/daemon.json" ]; then
        cp /etc/docker/daemon.json "${FULL_BACKUP_PATH}/configs/docker/"
    fi
}

# Enhanced VM configuration backup
backup_vm_configs() {
    log "Backing up enhanced VM configurations..."
    
    # VM list and status with more details
    virsh list --all --name > "${FULL_BACKUP_PATH}/configs/vm/vm_list.txt" 2>/dev/null || warn "Libvirt not available"
    
    # Enhanced VM configurations
    if command -v virsh >/dev/null 2>&1; then
        virsh list --name | while read vm; do
            if [ -n "$vm" ]; then
                # VM XML configuration
                virsh dumpxml "$vm" > "${FULL_BACKUP_PATH}/configs/vm/${vm}.xml" 2>/dev/null || warn "Could not backup VM config for $vm"
                
                # VM status and info
                virsh dominfo "$vm" > "${FULL_BACKUP_PATH}/configs/vm/${vm}_info.txt" 2>/dev/null || warn "Could not get VM info for $vm"
                
                # VM block devices
                virsh domblklist "$vm" > "${FULL_BACKUP_PATH}/configs/vm/${vm}_disks.txt" 2>/dev/null || warn "Could not get VM disks for $vm"
            fi
        done
    fi
    
    # Libvirt configuration
    if [ -d "/etc/libvirt" ]; then
        cp -r /etc/libvirt "${FULL_BACKUP_PATH}/configs/vm/" 2>/dev/null || warn "Could not backup libvirt config"
    fi
}

# Enhanced network configuration backup
backup_network_configs() {
    log "Backing up enhanced network configurations..."
    
    # Network interfaces with more details
    ip addr show > "${FULL_BACKUP_PATH}/configs/network/interfaces.txt"
    ip route show > "${FULL_BACKUP_PATH}/configs/network/routes.txt"
    ip link show > "${FULL_BACKUP_PATH}/configs/network/links.txt"
    
    # Network statistics with more detail
    cat /proc/net/dev > "${FULL_BACKUP_PATH}/performance/network/network_stats.txt"
    ss -tuln > "${FULL_BACKUP_PATH}/configs/network/listening_ports.txt"
    
    # Enhanced firewall rules
    iptables -L -n -v > "${FULL_BACKUP_PATH}/configs/network/iptables.txt" 2>/dev/null || warn "iptables not available"
    iptables -t nat -L -n -v > "${FULL_BACKUP_PATH}/configs/network/iptables_nat.txt" 2>/dev/null || warn "iptables nat not available"
    
    # DNS and network configuration
    cat /etc/resolv.conf > "${FULL_BACKUP_PATH}/configs/network/resolv.conf"
    cat /etc/hosts > "${FULL_BACKUP_PATH}/configs/network/hosts"
    
    # Network bonding information
    if [ -d "/proc/net/bonding" ]; then
        cp -r /proc/net/bonding "${FULL_BACKUP_PATH}/configs/network/" 2>/dev/null || warn "Could not backup bonding info"
    fi
}

# Enhanced performance data collection
collect_performance_data() {
    log "Collecting enhanced performance data..."
    
    # Multiple CPU usage samples
    for i in $(seq 1 3); do
        top -bn1 > "${FULL_BACKUP_PATH}/performance/cpu/cpu_usage_${i}.txt"
        sleep 2
    done
    
    # Enhanced memory usage
    cat /proc/meminfo > "${FULL_BACKUP_PATH}/performance/memory/memory_usage.txt"
    free -h > "${FULL_BACKUP_PATH}/performance/memory/memory_summary.txt"
    
    # Enhanced disk I/O with multiple samples
    for i in $(seq 1 3); do
        iostat -x 1 1 > "${FULL_BACKUP_PATH}/performance/storage/disk_io_${i}.txt" 2>/dev/null || warn "iostat not available"
        sleep 2
    done
    
    # Enhanced network I/O
    cat /proc/net/dev > "${FULL_BACKUP_PATH}/performance/network/network_io.txt"
    ss -i > "${FULL_BACKUP_PATH}/performance/network/network_connections.txt"
    
    # Process information
    ps aux --sort=-%cpu | head -50 > "${FULL_BACKUP_PATH}/performance/processes/top_cpu_processes.txt"
    ps aux --sort=-%mem | head -50 > "${FULL_BACKUP_PATH}/performance/processes/top_memory_processes.txt"
    
    # Load average and uptime
    uptime > "${FULL_BACKUP_PATH}/performance/system_load.txt"
    cat /proc/loadavg > "${FULL_BACKUP_PATH}/performance/system_loadavg.txt"
}

# Enhanced log collection
collect_logs() {
    log "Collecting enhanced system logs..."
    
    # System logs with more lines
    journalctl --no-pager -n 2000 > "${FULL_BACKUP_PATH}/logs/system/system.log" 2>/dev/null || warn "Could not collect system logs"
    journalctl --no-pager -n 500 --priority=err > "${FULL_BACKUP_PATH}/logs/system/errors.log" 2>/dev/null || warn "Could not collect error logs"
    
    # Docker logs with better error handling
    if command -v docker >/dev/null 2>&1; then
        docker ps --format "{{.Names}}" | while read container; do
            docker logs --tail 200 "$container" > "${FULL_BACKUP_PATH}/logs/docker/${container}.log" 2>/dev/null || warn "Could not get logs for container $container"
        done
    fi
    
    # VM logs with more detail
    if command -v virsh >/dev/null 2>&1; then
        virsh list --name | while read vm; do
            if [ -n "$vm" ]; then
                virsh domdisplay "$vm" > "${FULL_BACKUP_PATH}/logs/vm/${vm}_display.log" 2>/dev/null || warn "Could not get display info for VM $vm"
                virsh domstate "$vm" > "${FULL_BACKUP_PATH}/logs/vm/${vm}_state.log" 2>/dev/null || warn "Could not get state info for VM $vm"
            fi
        done
    fi
    
    # Plugin logs
    if [ -d "/var/log/plugins" ]; then
        cp -r /var/log/plugins "${FULL_BACKUP_PATH}/logs/plugins/" 2>/dev/null || warn "Could not backup plugin logs"
    fi
}

# Create enhanced analysis summary
create_analysis_summary() {
    log "Creating enhanced analysis summary..."
    
    cat > "${FULL_BACKUP_PATH}/analysis/system_summary.txt" << EOF
=== ENHANCED UNRAID SYSTEM ANALYSIS SUMMARY ===
Backup Date: $(date -Iseconds)
Hostname: $(hostname)

SYSTEM OVERVIEW:
- CPU: $(grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
- Cores: $(nproc)
- Memory: $(free -h | grep Mem | awk '{print $2}')
- Uptime: $(uptime -p)
- Load Average: $(uptime | awk -F'load average:' '{print $2}')

STORAGE OVERVIEW:
$(df -h | grep -E "(Filesystem|/dev/)")

DOCKER STATUS:
$(docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Size}}" 2>/dev/null || echo "Docker not available")

VM STATUS:
$(virsh list --all 2>/dev/null || echo "Libvirt not available")

NETWORK INTERFACES:
$(ip addr show | grep -E "(inet|UP|DOWN)")

PERFORMANCE METRICS:
- Load Average: $(uptime | awk -F'load average:' '{print $2}')
- Memory Usage: $(free | grep Mem | awk '{printf "%.1f%%", $3/$2 * 100.0}')
- Disk Usage: $(df / | tail -1 | awk '{print $5}')
- Active Processes: $(ps aux | wc -l)

ENHANCED RECOMMENDATIONS:
- Check storage array health and SMART status
- Monitor memory usage patterns and swap usage
- Review Docker container resource usage and cleanup stopped containers
- Verify network configuration and bonding setup
- Consider performance tuning opportunities based on load averages
- Review plugin configurations and update if necessary
EOF
}

# Create enhanced git-friendly archive
create_git_archive() {
    log "Creating enhanced git-friendly archive..."
    
    # Create a clean version for git
    local git_backup_dir="${BACKUP_DIR}/git_${BACKUP_NAME}"
    mkdir -p "$git_backup_dir"
    
    # Copy essential configs with better organization
    cp -r "${FULL_BACKUP_PATH}/configs" "$git_backup_dir/"
    cp -r "${FULL_BACKUP_PATH}/analysis" "$git_backup_dir/"
    cp -r "${FULL_BACKUP_PATH}/metadata" "$git_backup_dir/"
    
    # Create enhanced summary files
    cat > "$git_backup_dir/README.md" << EOF
# Enhanced Unraid Configuration Backup - $(date +"%Y-%m-%d %H:%M:%S")

## System Information
- Hostname: $(hostname)
- CPU: $(grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
- Memory: $(free -h | grep Mem | awk '{print $2}')
- Uptime: $(uptime -p)
- Load Average: $(uptime | awk -F'load average:' '{print $2}')

## Backup Contents
- \`configs/\`: System configuration files (enhanced)
- \`analysis/\`: System analysis and recommendations (enhanced)
- \`metadata/\`: Backup metadata and statistics

## Quick Analysis
$(cat "${FULL_BACKUP_PATH}/analysis/system_summary.txt")

## Enhanced Features
- Parallel processing for faster backups
- Compression support for smaller archives
- Retention policy for automatic cleanup
- Enhanced error handling and logging
- Comprehensive system monitoring

## Next Steps
1. Review system_summary.txt for immediate insights
2. Check configs/ for specific configuration details
3. Use this data for performance optimization analysis
4. Consider setting up automated backups with cron
EOF
    
    # Create enhanced .gitignore
    cat > "$git_backup_dir/.gitignore" << EOF
# Ignore sensitive configuration files
configs/system/users.ini
configs/system/state.ini
configs/system/var.ini
configs/system/passwd
configs/system/shadow
configs/system/*.key
configs/system/*.pem
configs/system/*.crt

# Ignore logs that might contain sensitive data
logs/
performance/

# Ignore temporary files
*.tmp
*.log
*.cache

# Ignore large files
*.tar.gz
*.zip
EOF
}

# Compress backup if enabled
compress_backup() {
    if [ "$COMPRESS_BACKUPS" = "true" ]; then
        log "Compressing backup archive..."
        cd "$BACKUP_DIR"
        tar -czf "${BACKUP_NAME}.tar.gz" -C . "$BACKUP_NAME" --remove-files
        log "Backup compressed: ${BACKUP_NAME}.tar.gz"
    fi
}

# Main execution
main() {
    start_timer
    
    # Load configuration
    load_config
    
    # Set backup paths
    TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    BACKUP_NAME="unraid_backup_${TIMESTAMP}"
    FULL_BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"
    
    log "Starting enhanced Unraid configuration backup..."
    log "Configuration: $CONFIG_FILE"
    log "Backup directory: $BACKUP_DIR"
    log "Parallel backup: $PARALLEL_BACKUP"
    log "Max workers: $MAX_WORKERS"
    log "Compression: $COMPRESS_BACKUPS"
    
    # Check if running as root (recommended for full access)
    if [ "$EUID" -ne 0 ]; then
        warn "Not running as root. Some configurations may not be accessible."
    fi
    
    # Cleanup old backups
    cleanup_old_backups
    
    # Create backup structure
    create_backup_structure
    
    # Collect system information
    collect_system_info
    
    # Backup configurations (can run in parallel)
    if [ "$PARALLEL_BACKUP" = "true" ]; then
        backup_unraid_configs &
        backup_docker_configs &
        backup_vm_configs &
        backup_network_configs &
        wait
    else
        backup_unraid_configs
        backup_docker_configs
        backup_vm_configs
        backup_network_configs
    fi
    
    # Collect performance data and logs
    collect_performance_data
    collect_logs
    
    # Create analysis
    create_analysis_summary
    create_git_archive
    
    # Compress if enabled
    compress_backup
    
    end_timer
    
    log "Enhanced backup completed successfully!"
    log "Full backup location: ${FULL_BACKUP_PATH}"
    log "Git-friendly backup location: ${BACKUP_DIR}/git_${BACKUP_NAME}"
    log "Total backup size: $(du -sh "${FULL_BACKUP_PATH}" 2>/dev/null | cut -f1 || echo 'Unknown')"
    
    echo -e "${BLUE}=== ENHANCED BACKUP SUMMARY ===${NC}"
    echo "📁 Full backup: ${FULL_BACKUP_PATH}"
    echo "🔧 Git backup: ${BACKUP_DIR}/git_${BACKUP_NAME}"
    echo "📊 Analysis: ${FULL_BACKUP_PATH}/analysis/system_summary.txt"
    echo "⚙️  Configuration: $CONFIG_FILE"
    echo "🚀 Parallel processing: $PARALLEL_BACKUP"
    echo "🗜️  Compression: $COMPRESS_BACKUPS"
    echo ""
    echo "💡 Next steps:"
    echo "1. Review the enhanced analysis summary"
    echo "2. Commit the git-friendly backup to version control"
    echo "3. Share the analysis data for optimization recommendations"
    echo "4. Consider setting up automated backups"
}

# Run main function
main "$@" 