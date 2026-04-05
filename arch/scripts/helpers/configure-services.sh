#!/bin/bash

# System Services Configuration Script
# Configures and enables essential services for a desktop Arch Linux system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Function to enable service if installed
enable_service() {
    local service=$1
    local package=$2
    local description=$3
    
    if systemctl list-unit-files | grep -q "$service"; then
        if systemctl is-enabled "$service" &> /dev/null; then
            print_success "$description already enabled"
        else
            systemctl enable --now "$service"
            print_success "$description enabled and started"
        fi
    else
        print_warning "$description not found. Install with: sudo pacman -S $package"
    fi
}

# Main configuration
main() {
    check_root
    
    echo "╔════════════════════════════════════════════════════════╗"
    echo "║     Arch Linux System Services Configuration          ║"
    echo "╚════════════════════════════════════════════════════════╝"
    echo ""
    
    # Network Services
    print_status "Configuring Network Services..."
    enable_service "NetworkManager.service" "networkmanager" "NetworkManager"
    enable_service "systemd-resolved.service" "systemd" "DNS Resolver"
    
    # Bluetooth
    print_status "Configuring Bluetooth..."
    enable_service "bluetooth.service" "bluez" "Bluetooth"
    
    # Printing
    print_status "Configuring Printing Services..."
    enable_service "cups.service" "cups" "CUPS Printing"
    enable_service "cups-browsed.service" "cups-pk-helper" "Network Printer Discovery"
    
    # Audio
    print_status "Configuring Audio Services..."
    if command -v pipewire &> /dev/null; then
        systemctl --user enable --now pipewire.service
        systemctl --user enable --now pipewire-pulse.service
        systemctl --user enable --now wireplumber.service
        print_success "PipeWire audio system enabled"
    else
        print_warning "PipeWire not installed. Install with: sudo pacman -S pipewire pipewire-pulse wireplumber"
    fi
    
    # SSH
    print_status "Configuring SSH..."
    if systemctl list-unit-files | grep -q "sshd.service"; then
        read -p "Enable SSH server? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            systemctl enable --now sshd.service
            print_success "SSH server enabled"
            print_warning "Remember to configure /etc/ssh/sshd_config for security"
        fi
    fi
    
    # Firewall
    print_status "Configuring Firewall..."
    if command -v ufw &> /dev/null; then
        ufw --force enable
        ufw default deny incoming
        ufw default allow outgoing
        systemctl enable --now ufw.service
        print_success "UFW firewall enabled with default rules"
    else
        print_warning "UFW not installed. Install with: sudo pacman -S ufw"
    fi
    
    # Docker
    print_status "Configuring Docker..."
    if systemctl list-unit-files | grep -q "docker.service"; then
        systemctl enable --now docker.service
        if ! groups $SUDO_USER | grep -q docker; then
            usermod -aG docker $SUDO_USER
            print_success "Docker enabled and $SUDO_USER added to docker group"
            print_warning "Logout and login for docker group changes to take effect"
        else
            print_success "Docker already configured"
        fi
    else
        print_info "Docker not installed. Install with: sudo pacman -S docker docker-compose"
    fi
    
    # Tailscale
    print_status "Configuring Tailscale..."
    enable_service "tailscaled.service" "tailscale" "Tailscale VPN"
    
    # Timeshift
    print_status "Configuring Timeshift..."
    if command -v timeshift &> /dev/null; then
        # Enable cronie for scheduled snapshots
        enable_service "cronie.service" "cronie" "Cron scheduler for Timeshift"
        print_info "Configure Timeshift with: sudo timeshift-gtk"
    else
        print_info "Timeshift not installed. Install with: sudo pacman -S timeshift"
    fi
    
    # SSD Optimization
    print_status "Configuring SSD Optimizations..."
    if [ -f /sys/block/nvme0n1/queue/scheduler ] || [ -f /sys/block/sda/queue/scheduler ]; then
        enable_service "fstrim.timer" "util-linux" "Weekly SSD TRIM"
    fi
    
    # Power Management (for laptops)
    print_status "Checking Power Management..."
    if [ -d /sys/class/power_supply/BAT0 ] || [ -d /sys/class/power_supply/BAT1 ]; then
        print_info "Laptop detected - configuring power management"
        if command -v tlp &> /dev/null; then
            systemctl enable --now tlp.service
            systemctl mask systemd-rfkill.service
            systemctl mask systemd-rfkill.socket
            print_success "TLP power management enabled"
        else
            print_warning "TLP not installed. Install with: sudo pacman -S tlp tlp-rdw"
        fi
        
        if command -v auto-cpufreq &> /dev/null; then
            systemctl enable --now auto-cpufreq.service
            print_success "auto-cpufreq enabled"
        else
            print_info "Consider installing auto-cpufreq: yay -S auto-cpufreq"
        fi
    fi
    
    # Display Manager Check
    print_status "Checking Display Manager..."
    if ! systemctl is-enabled gdm &> /dev/null && \
       ! systemctl is-enabled sddm &> /dev/null && \
       ! systemctl is-enabled lightdm &> /dev/null; then
        print_warning "No display manager enabled. System will boot to console."
        print_info "Install a display manager:"
        echo "  GNOME:  sudo pacman -S gdm && sudo systemctl enable gdm"
        echo "  KDE:    sudo pacman -S sddm && sudo systemctl enable sddm"
        echo "  XFCE:   sudo pacman -S lightdm && sudo systemctl enable lightdm"
    fi
    
    # System Clock Sync
    print_status "Configuring Time Synchronization..."
    timedatectl set-ntp true
    systemctl enable --now systemd-timesyncd.service
    print_success "NTP time synchronization enabled"
    
    # Avahi (Network Discovery)
    print_status "Configuring Network Discovery..."
    enable_service "avahi-daemon.service" "avahi" "Avahi mDNS/DNS-SD"
    
    # Configure systemd-resolved for better DNS
    if systemctl is-active systemd-resolved &> /dev/null; then
        print_status "Optimizing DNS resolution..."
        cat > /etc/systemd/resolved.conf.d/dns.conf << EOF
[Resolve]
DNS=1.1.1.1#cloudflare-dns.com 1.0.0.1#cloudflare-dns.com
FallbackDNS=8.8.8.8#dns.google 8.8.4.4#dns.google
DNSSEC=allow-downgrade
DNSOverTLS=opportunistic
Cache=yes
CacheFromLocalhost=yes
EOF
        systemctl restart systemd-resolved
        print_success "DNS resolver optimized with Cloudflare and Google DNS"
    fi
    
    # Performance Tuning
    print_status "Applying Performance Tuning..."
    
    # Increase file watchers for development
    echo "fs.inotify.max_user_watches=524288" > /etc/sysctl.d/40-max-user-watches.conf
    
    # Reduce swappiness for desktop usage
    echo "vm.swappiness=10" > /etc/sysctl.d/99-swappiness.conf
    
    # Apply sysctl settings
    sysctl --system &> /dev/null
    print_success "Performance tuning applied"
    
    # Summary
    echo ""
    echo "════════════════════════════════════════════════════════"
    print_success "Service configuration complete!"
    echo ""
    echo "Enabled services:"
    systemctl list-unit-files --state=enabled --no-pager | grep -E "(NetworkManager|bluetooth|cups|docker|tailscaled|fstrim|tlp|gdm|sddm|lightdm)" || true
    echo ""
    print_info "Additional Configuration Needed:"
    echo "  • Tailscale: Run 'sudo tailscale up' to connect"
    echo "  • Timeshift: Run 'sudo timeshift-gtk' to configure backups"
    echo "  • Docker: Logout and login for group changes"
    echo "  • Firewall: Configure additional rules as needed"
    echo ""
    print_warning "Reboot recommended for all changes to take effect"
    echo "════════════════════════════════════════════════════════"
    
    read -p "Would you like to reboot now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        reboot
    fi
}

# Run main function
main
