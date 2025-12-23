#!/bin/bash

# Unraid Configuration Backup Script
# Captures comprehensive system data for optimization analysis

set -euo pipefail

# Configuration
BACKUP_DIR="$(dirname "$0")/backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="unraid_backup_${TIMESTAMP}"
FULL_BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create backup directory structure
create_backup_structure() {
    log "Creating backup directory structure..."
    mkdir -p "${FULL_BACKUP_PATH}"/{configs,performance,logs,analysis}
    mkdir -p "${FULL_BACKUP_PATH}/configs"/{system,docker,vm,network,storage}
    mkdir -p "${FULL_BACKUP_PATH}/performance"/{cpu,memory,storage,network}
    mkdir -p "${FULL_BACKUP_PATH}/logs"/{system,docker,vm}
}

# System information collection
collect_system_info() {
    log "Collecting system information..."
    
    # Basic system info
    cat > "${FULL_BACKUP_PATH}/system_info.txt" << EOF
=== UNRAID SYSTEM INFORMATION ===
Backup Date: $(date)
Hostname: $(hostname)
Kernel: $(uname -r)
Uname: $(uname -a)
EOF

    # CPU information
    log "Collecting CPU information..."
    cat > "${FULL_BACKUP_PATH}/performance/cpu/cpu_info.txt" << EOF
=== CPU INFORMATION ===
CPU Model: $(grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
CPU Cores: $(nproc)
CPU Threads: $(grep -c processor /proc/cpuinfo)
CPU Architecture: $(uname -m)
EOF

    # Memory information
    log "Collecting memory information..."
    cat > "${FULL_BACKUP_PATH}/performance/memory/memory_info.txt" << EOF
=== MEMORY INFORMATION ===
$(free -h)
$(cat /proc/meminfo | grep -E "(MemTotal|MemFree|MemAvailable|SwapTotal|SwapFree)")
EOF

    # Storage information
    log "Collecting storage information..."
    cat > "${FULL_BACKUP_PATH}/performance/storage/storage_info.txt" << EOF
=== STORAGE INFORMATION ===
$(lsblk -f)
$(df -h)
$(cat /proc/mdstat 2>/dev/null || echo "No RAID arrays found")
EOF
}

# Unraid specific configuration backup
backup_unraid_configs() {
    log "Backing up Unraid configuration files..."
    
    # Flash drive configs (if accessible)
    if [ -d "/boot" ]; then
        log "Backing up /boot configuration..."
        cp -r /boot "${FULL_BACKUP_PATH}/configs/system/" 2>/dev/null || warn "Could not backup /boot directory"
    fi
    
    # System configuration files
    local config_files=(
        "/etc/fstab"
        "/etc/hosts"
        "/etc/resolv.conf"
        "/etc/network/interfaces"
        "/etc/sysctl.conf"
        "/etc/modules"
    )
    
    for file in "${config_files[@]}"; do
        if [ -f "$file" ]; then
            cp "$file" "${FULL_BACKUP_PATH}/configs/system/"
        fi
    done
    
    # Unraid specific files
    local unraid_files=(
        "/var/local/emhttp/var.ini"
        "/var/local/emhttp/state.ini"
        "/var/local/emhttp/disk.ini"
        "/var/local/emhttp/array.ini"
        "/var/local/emhttp/shares.ini"
        "/var/local/emhttp/users.ini"
        "/var/local/emhttp/settings.ini"
    )
    
    for file in "${unraid_files[@]}"; do
        if [ -f "$file" ]; then
            cp "$file" "${FULL_BACKUP_PATH}/configs/system/"
        fi
    done
}

# Docker configuration backup
backup_docker_configs() {
    log "Backing up Docker configurations..."
    
    # Docker system info
    docker system info > "${FULL_BACKUP_PATH}/configs/docker/docker_info.txt" 2>/dev/null || warn "Docker not available or not running"
    
    # Docker containers
    docker ps -a > "${FULL_BACKUP_PATH}/configs/docker/containers.txt" 2>/dev/null || warn "Could not get container list"
    
    # Docker images
    docker images > "${FULL_BACKUP_PATH}/configs/docker/images.txt" 2>/dev/null || warn "Could not get image list"
    
    # Docker networks
    docker network ls > "${FULL_BACKUP_PATH}/configs/docker/networks.txt" 2>/dev/null || warn "Could not get network list"
    
    # Docker volumes
    docker volume ls > "${FULL_BACKUP_PATH}/configs/docker/volumes.txt" 2>/dev/null || warn "Could not get volume list"
    
    # Docker compose files (if any)
    find /mnt -name "docker-compose.yml" -o -name "docker-compose.yaml" 2>/dev/null | while read file; do
        cp "$file" "${FULL_BACKUP_PATH}/configs/docker/"
    done
}

# VM configuration backup
backup_vm_configs() {
    log "Backing up VM configurations..."
    
    # VM list and status
    virsh list --all > "${FULL_BACKUP_PATH}/configs/vm/vm_list.txt" 2>/dev/null || warn "Libvirt not available"
    
    # VM configurations
    if command -v virsh >/dev/null 2>&1; then
        virsh list --name | while read vm; do
            if [ -n "$vm" ]; then
                virsh dumpxml "$vm" > "${FULL_BACKUP_PATH}/configs/vm/${vm}.xml" 2>/dev/null || warn "Could not backup VM config for $vm"
            fi
        done
    fi
}

# Network configuration backup
backup_network_configs() {
    log "Backing up network configurations..."
    
    # Network interfaces
    ip addr show > "${FULL_BACKUP_PATH}/configs/network/interfaces.txt"
    ip route show > "${FULL_BACKUP_PATH}/configs/network/routes.txt"
    
    # Network statistics
    cat /proc/net/dev > "${FULL_BACKUP_PATH}/performance/network/network_stats.txt"
    
    # Firewall rules (if iptables is available)
    iptables -L -n -v > "${FULL_BACKUP_PATH}/configs/network/iptables.txt" 2>/dev/null || warn "iptables not available"
    
    # DNS configuration
    cat /etc/resolv.conf > "${FULL_BACKUP_PATH}/configs/network/resolv.conf"
}

# Performance data collection
collect_performance_data() {
    log "Collecting performance data..."
    
    # CPU usage
    top -bn1 > "${FULL_BACKUP_PATH}/performance/cpu/cpu_usage.txt"
    
    # Memory usage
    cat /proc/meminfo > "${FULL_BACKUP_PATH}/performance/memory/memory_usage.txt"
    
    # Disk I/O
    iostat -x 1 1 > "${FULL_BACKUP_PATH}/performance/storage/disk_io.txt" 2>/dev/null || warn "iostat not available"
    
    # Network I/O
    cat /proc/net/dev > "${FULL_BACKUP_PATH}/performance/network/network_io.txt"
    
    # Load average
    uptime > "${FULL_BACKUP_PATH}/performance/system_load.txt"
}

# Log collection
collect_logs() {
    log "Collecting system logs..."
    
    # System logs (last 1000 lines)
    journalctl --no-pager -n 1000 > "${FULL_BACKUP_PATH}/logs/system/system.log" 2>/dev/null || warn "Could not collect system logs"
    
    # Docker logs (if available)
    if command -v docker >/dev/null 2>&1; then
        docker ps --format "{{.Names}}" | while read container; do
            docker logs --tail 100 "$container" > "${FULL_BACKUP_PATH}/logs/docker/${container}.log" 2>/dev/null || warn "Could not get logs for container $container"
        done
    fi
    
    # VM logs (if available)
    if command -v virsh >/dev/null 2>&1; then
        virsh list --name | while read vm; do
            if [ -n "$vm" ]; then
                virsh domdisplay "$vm" > "${FULL_BACKUP_PATH}/logs/vm/${vm}_display.log" 2>/dev/null || warn "Could not get display info for VM $vm"
            fi
        done
    fi
}

# Create analysis summary
create_analysis_summary() {
    log "Creating analysis summary..."
    
    cat > "${FULL_BACKUP_PATH}/analysis/system_summary.txt" << EOF
=== UNRAID SYSTEM ANALYSIS SUMMARY ===
Backup Date: $(date)
Hostname: $(hostname)

SYSTEM OVERVIEW:
- CPU: $(grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
- Cores: $(nproc)
- Memory: $(free -h | grep Mem | awk '{print $2}')
- Uptime: $(uptime -p)

STORAGE OVERVIEW:
$(df -h | grep -E "(Filesystem|/dev/)")

DOCKER STATUS:
$(docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "Docker not available")

VM STATUS:
$(virsh list --all 2>/dev/null || echo "Libvirt not available")

NETWORK INTERFACES:
$(ip addr show | grep -E "(inet|UP|DOWN)")

PERFORMANCE METRICS:
- Load Average: $(uptime | awk -F'load average:' '{print $2}')
- Memory Usage: $(free | grep Mem | awk '{printf "%.1f%%", $3/$2 * 100.0}')
- Disk Usage: $(df / | tail -1 | awk '{print $5}')

RECOMMENDATIONS:
- Check storage array health
- Monitor memory usage patterns
- Review Docker container resource usage
- Verify network configuration
- Consider performance tuning opportunities
EOF
}

# Create git-friendly archive
create_git_archive() {
    log "Creating git-friendly archive..."
    
    # Create a clean version for git
    local git_backup_dir="${BACKUP_DIR}/git_${BACKUP_NAME}"
    mkdir -p "$git_backup_dir"
    
    # Copy essential configs
    cp -r "${FULL_BACKUP_PATH}/configs" "$git_backup_dir/"
    cp -r "${FULL_BACKUP_PATH}/analysis" "$git_backup_dir/"
    
    # Create summary files
    cat > "$git_backup_dir/README.md" << EOF
# Unraid Configuration Backup - $(date +"%Y-%m-%d %H:%M:%S")

## System Information
- Hostname: $(hostname)
- CPU: $(grep "model name" /proc/cpuinfo | head -1 | cut -d: -f2 | xargs)
- Memory: $(free -h | grep Mem | awk '{print $2}')
- Uptime: $(uptime -p)

## Backup Contents
- \`configs/\`: System configuration files
- \`analysis/\`: System analysis and recommendations

## Quick Analysis
$(cat "${FULL_BACKUP_PATH}/analysis/system_summary.txt")

## Next Steps
1. Review system_summary.txt for immediate insights
2. Check configs/ for specific configuration details
3. Use this data for performance optimization analysis
EOF
    
    # Create .gitignore for sensitive data
    cat > "$git_backup_dir/.gitignore" << EOF
# Ignore sensitive configuration files
configs/system/users.ini
configs/system/state.ini
configs/system/var.ini

# Ignore logs that might contain sensitive data
logs/
performance/

# Ignore temporary files
*.tmp
*.log
EOF
}

# Main execution
main() {
    log "Starting Unraid configuration backup..."
    
    # Check if running as root (recommended for full access)
    if [ "$EUID" -ne 0 ]; then
        warn "Not running as root. Some configurations may not be accessible."
    fi
    
    create_backup_structure
    collect_system_info
    backup_unraid_configs
    backup_docker_configs
    backup_vm_configs
    backup_network_configs
    collect_performance_data
    collect_logs
    create_analysis_summary
    create_git_archive
    
    log "Backup completed successfully!"
    log "Full backup location: ${FULL_BACKUP_PATH}"
    log "Git-friendly backup location: ${BACKUP_DIR}/git_${BACKUP_NAME}"
    log "Total backup size: $(du -sh "${FULL_BACKUP_PATH}" | cut -f1)"
    
    echo -e "${BLUE}=== BACKUP SUMMARY ===${NC}"
    echo "📁 Full backup: ${FULL_BACKUP_PATH}"
    echo "🔧 Git backup: ${BACKUP_DIR}/git_${BACKUP_NAME}"
    echo "📊 Analysis: ${FULL_BACKUP_PATH}/analysis/system_summary.txt"
    echo ""
    echo "💡 Next steps:"
    echo "1. Review the analysis summary"
    echo "2. Commit the git-friendly backup to version control"
    echo "3. Share the analysis data for optimization recommendations"
}

# Run main function
main "$@" 