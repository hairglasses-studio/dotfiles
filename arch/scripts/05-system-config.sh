#!/bin/bash

#############################################
# 05 - System Configuration
# Optimize and configure Arch Linux system
#############################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() { echo -e "${BLUE}==>${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }

# Enable systemd services
enable_services() {
    print_status "Enabling essential services..."
    
    # Network
    sudo systemctl enable --now NetworkManager
    
    # Bluetooth
    sudo systemctl enable --now bluetooth
    
    # Printing
    sudo systemctl enable --now cups
    
    # Firewall
    sudo systemctl enable --now ufw
    
    # SSD optimization
    sudo systemctl enable --now fstrim.timer
    
    # Power management
    if [ -f /proc/acpi/battery/BAT0/state ] || [ -d /sys/class/power_supply/BAT0 ]; then
        sudo systemctl enable --now tlp
        sudo systemctl enable --now thermald
    fi
    
    # Early OOM killer
    sudo systemctl enable --now earlyoom
    
    print_success "Services enabled"
}

# Configure swappiness
configure_swap() {
    print_status "Configuring swap settings..."
    
    # Reduce swappiness for desktop use
    echo "vm.swappiness=10" | sudo tee /etc/sysctl.d/99-swappiness.conf
    
    # Configure zram if available
    if ! systemctl is-enabled zram-generator &> /dev/null; then
        sudo pacman -S --needed --noconfirm zram-generator
        
        sudo tee /etc/systemd/zram-generator.conf > /dev/null << 'EOF'
[zram0]
zram-size = ram / 2
compression-algorithm = zstd
EOF
        
        sudo systemctl daemon-reload
        sudo systemctl start systemd-zram-setup@zram0.service
    fi
    
    print_success "Swap configured"
}

# Configure CPU governor
configure_cpu() {
    print_status "Configuring CPU governor..."
    
    # Install cpupower
    sudo pacman -S --needed --noconfirm cpupower
    
    # Set performance governor for desktop
    echo -n "Use performance CPU governor? (better for desktop) (Y/n): "
    read -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        sudo sed -i "s/^#governor=.*/governor='performance'/" /etc/default/cpupower
        sudo systemctl enable --now cpupower
        print_success "CPU governor set to performance"
    else
        sudo sed -i "s/^#governor=.*/governor='powersave'/" /etc/default/cpupower
        sudo systemctl enable --now cpupower
        print_success "CPU governor set to powersave"
    fi
}

# Configure GRUB
configure_grub() {
    print_status "Configuring GRUB..."
    
    # Backup GRUB config
    sudo cp /etc/default/grub /etc/default/grub.backup
    
    # Reduce GRUB timeout
    sudo sed -i 's/^GRUB_TIMEOUT=.*/GRUB_TIMEOUT=3/' /etc/default/grub
    
    # Enable GRUB menu saving
    sudo sed -i 's/^#GRUB_SAVEDEFAULT=.*/GRUB_SAVEDEFAULT=true/' /etc/default/grub
    
    # Quiet boot (optional)
    echo -n "Enable quiet boot? (hide boot messages) (y/N): "
    read -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo sed -i 's/^GRUB_CMDLINE_LINUX_DEFAULT=.*/GRUB_CMDLINE_LINUX_DEFAULT="quiet loglevel=3 nowatchdog"/' /etc/default/grub
    fi
    
    # Update GRUB
    sudo grub-mkconfig -o /boot/grub/grub.cfg
    
    print_success "GRUB configured"
}

# Configure makepkg for faster compilation
configure_makepkg() {
    print_status "Configuring makepkg for faster compilation..."
    
    # Backup makepkg.conf
    sudo cp /etc/makepkg.conf /etc/makepkg.conf.backup
    
    # Set makeflags for parallel compilation
    CORES=$(nproc)
    sudo sed -i "s/^#MAKEFLAGS=.*/MAKEFLAGS=\"-j$CORES\"/" /etc/makepkg.conf
    
    # Enable compression with multiple cores
    sudo sed -i "s/^COMPRESSZST=.*/COMPRESSZST=(zstd -c -T0 --ultra -20 -)/" /etc/makepkg.conf
    
    # Use all cores for compression
    sudo sed -i "s/^#COMPRESSXZ=.*/COMPRESSXZ=(xz -c -z -T0 -)/" /etc/makepkg.conf
    
    print_success "Makepkg configured for $CORES cores"
}

# Configure systemd
configure_systemd() {
    print_status "Configuring systemd..."
    
    # Faster boot
    sudo systemctl disable NetworkManager-wait-online.service
    
    # Configure journald
    sudo sed -i 's/^#SystemMaxUse=.*/SystemMaxUse=100M/' /etc/systemd/journald.conf
    sudo systemctl restart systemd-journald
    
    # Set default target
    if systemctl get-default | grep -q graphical; then
        print_success "Graphical target already set"
    else
        sudo systemctl set-default graphical.target
        print_success "Default target set to graphical"
    fi
}

# Configure limits
configure_limits() {
    print_status "Configuring system limits..."
    
    # Increase file watchers for development
    echo "fs.inotify.max_user_watches=524288" | sudo tee /etc/sysctl.d/40-inotify.conf
    
    # Increase file descriptors
    sudo tee /etc/security/limits.d/20-nofile.conf > /dev/null << 'EOF'
* soft nofile 65535
* hard nofile 65535
EOF
    
    # Apply sysctl settings
    sudo sysctl --system
    
    print_success "System limits configured"
}

# Install and configure Timeshift
configure_timeshift() {
    print_status "Configuring Timeshift..."
    
    if ! command -v timeshift &> /dev/null; then
        sudo pacman -S --needed --noconfirm timeshift
    fi
    
    print_warning "Please configure Timeshift manually:"
    print_warning "  Run: sudo timeshift-gtk"
    print_warning "  Set up automatic snapshots before system updates"
}

# Configure firewall
configure_firewall() {
    print_status "Configuring firewall..."
    
    # Install ufw if not present
    sudo pacman -S --needed --noconfirm ufw
    
    # Basic firewall rules
    sudo ufw default deny incoming
    sudo ufw default allow outgoing
    
    # Allow common services
    echo -n "Allow SSH connections? (y/N): "
    read -n 1 -r
    echo
    [[ $REPLY =~ ^[Yy]$ ]] && sudo ufw allow ssh
    
    echo -n "Allow HTTP/HTTPS? (y/N): "
    read -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo ufw allow http
        sudo ufw allow https
    fi
    
    # Enable firewall
    sudo ufw --force enable
    
    print_success "Firewall configured"
}

# Hardware acceleration
configure_graphics() {
    print_status "Configuring graphics and hardware acceleration..."
    
    # Detect GPU
    if lspci | grep -i nvidia > /dev/null; then
        print_status "NVIDIA GPU detected"
        sudo pacman -S --needed --noconfirm \
            nvidia \
            nvidia-utils \
            nvidia-settings \
            lib32-nvidia-utils \
            nvtop
    elif lspci | grep -i amd > /dev/null; then
        print_status "AMD GPU detected"
        sudo pacman -S --needed --noconfirm \
            mesa \
            lib32-mesa \
            vulkan-radeon \
            lib32-vulkan-radeon \
            libva-mesa-driver \
            lib32-libva-mesa-driver \
            mesa-vdpau \
            lib32-mesa-vdpau \
            radeontop
    else
        print_status "Intel GPU detected"
        sudo pacman -S --needed --noconfirm \
            mesa \
            lib32-mesa \
            vulkan-intel \
            lib32-vulkan-intel \
            libva-intel-driver \
            lib32-libva-intel-driver \
            intel-gpu-tools
    fi
    
    # Common packages
    sudo pacman -S --needed --noconfirm \
        vulkan-tools \
        libva-utils \
        vdpauinfo
    
    print_success "Graphics configured"
}

# Main execution
main() {
    print_status "Configuring system..."
    
    enable_services
    configure_swap
    configure_cpu
    configure_grub
    configure_makepkg
    configure_systemd
    configure_limits
    configure_timeshift
    configure_firewall
    configure_graphics
    
    print_success "System configuration completed!"
    print_warning "A reboot is recommended to apply all changes"
}

main
