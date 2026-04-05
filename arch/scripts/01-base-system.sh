#!/bin/bash

#############################################
# 01 - Base System Setup
# Prepares Arch Linux with essential tools
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

# Update system
update_system() {
    print_status "Updating system packages..."
    sudo pacman -Syu --noconfirm
    print_success "System updated"
}

# Install base development tools
install_base_devel() {
    print_status "Installing base development tools..."
    sudo pacman -S --needed --noconfirm \
        base-devel \
        git \
        wget \
        curl \
        vim \
        nano \
        htop \
        neofetch \
        tree \
        unzip \
        p7zip \
        rsync \
        openssh \
        man-db \
        man-pages
    print_success "Base development tools installed"
}

# Install yay AUR helper
install_yay() {
    if ! command -v yay &> /dev/null; then
        print_status "Installing yay AUR helper..."
        cd /tmp
        git clone https://aur.archlinux.org/yay.git
        cd yay
        makepkg -si --noconfirm
        cd ..
        rm -rf yay
        print_success "yay installed successfully"
    else
        print_success "yay is already installed"
    fi
}

# Enable multilib repository (for Steam, Wine, etc.)
enable_multilib() {
    print_status "Checking multilib repository..."
    if ! grep -q "^\[multilib\]" /etc/pacman.conf; then
        print_status "Enabling multilib repository..."
        sudo bash -c 'cat >> /etc/pacman.conf << EOF

[multilib]
Include = /etc/pacman.d/mirrorlist
EOF'
        sudo pacman -Sy
        print_success "Multilib repository enabled"
    else
        print_success "Multilib repository already enabled"
    fi
}

# Optimize pacman
optimize_pacman() {
    print_status "Optimizing pacman configuration..."
    
    # Backup original config
    sudo cp /etc/pacman.conf /etc/pacman.conf.backup
    
    # Enable parallel downloads
    sudo sed -i 's/^#ParallelDownloads/ParallelDownloads/' /etc/pacman.conf
    
    # Enable color output
    sudo sed -i 's/^#Color/Color/' /etc/pacman.conf
    
    # Enable VerbosePkgLists
    sudo sed -i 's/^#VerbosePkgLists/VerbosePkgLists/' /etc/pacman.conf
    
    print_success "Pacman optimized"
}

# Update mirrorlist for faster downloads
update_mirrorlist() {
    print_status "Updating mirrorlist for optimal speed..."
    
    if ! command -v reflector &> /dev/null; then
        sudo pacman -S --needed --noconfirm reflector
    fi
    
    # Backup current mirrorlist
    sudo cp /etc/pacman.d/mirrorlist /etc/pacman.d/mirrorlist.backup
    
    # Update mirrorlist with fastest mirrors
    sudo reflector \
        --latest 20 \
        --protocol https \
        --sort rate \
        --save /etc/pacman.d/mirrorlist
    
    print_success "Mirrorlist updated with fastest mirrors"
}

# Install essential system utilities
install_system_utils() {
    print_status "Installing essential system utilities..."
    sudo pacman -S --needed --noconfirm \
        networkmanager \
        network-manager-applet \
        bluez \
        bluez-utils \
        cups \
        system-config-printer \
        ufw \
        tlp \
        powertop \
        thermald \
        earlyoom
    
    # Enable services
    sudo systemctl enable --now NetworkManager
    sudo systemctl enable --now bluetooth
    sudo systemctl enable --now cups
    sudo systemctl enable --now ufw
    sudo systemctl enable --now earlyoom
    
    # Configure firewall
    sudo ufw default deny incoming
    sudo ufw default allow outgoing
    sudo ufw enable
    
    print_success "System utilities installed and configured"
}

# Install filesystem tools
install_filesystem_tools() {
    print_status "Installing filesystem tools..."
    sudo pacman -S --needed --noconfirm \
        ntfs-3g \
        exfat-utils \
        btrfs-progs \
        dosfstools \
        e2fsprogs \
        xfsprogs \
        f2fs-tools \
        gvfs \
        gvfs-mtp \
        gvfs-gphoto2 \
        gvfs-smb \
        sshfs
    print_success "Filesystem tools installed"
}

# Install multimedia codecs
install_codecs() {
    print_status "Installing multimedia codecs..."
    sudo pacman -S --needed --noconfirm \
        ffmpeg \
        gstreamer \
        gst-plugins-base \
        gst-plugins-good \
        gst-plugins-bad \
        gst-plugins-ugly \
        gst-libav \
        libdvdcss
    print_success "Multimedia codecs installed"
}

# Install fonts
install_fonts() {
    print_status "Installing fonts..."
    sudo pacman -S --needed --noconfirm \
        ttf-dejavu \
        ttf-liberation \
        ttf-roboto \
        ttf-roboto-mono \
        ttf-ubuntu-font-family \
        noto-fonts \
        noto-fonts-cjk \
        noto-fonts-emoji \
        ttf-font-awesome \
        ttf-fira-code \
        ttf-fira-mono \
        ttf-jetbrains-mono \
        adobe-source-code-pro-fonts
    
    # Install Microsoft fonts from AUR
    yay -S --needed --noconfirm ttf-ms-fonts
    
    print_success "Fonts installed"
}

# Main execution
main() {
    print_status "Starting base system setup..."
    
    update_system
    install_base_devel
    install_yay
    enable_multilib
    optimize_pacman
    update_mirrorlist
    install_system_utils
    install_filesystem_tools
    install_codecs
    install_fonts
    
    print_success "Base system setup completed!"
}

main
