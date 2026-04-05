#!/bin/bash

#############################################
# 03 - Extended Applications
# Additional productivity and media apps
#############################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_status() { echo -e "${BLUE}==>${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_info() { echo -e "${CYAN}ℹ${NC} $1"; }

# Check if package is installed
is_installed() {
    pacman -Qi "$1" &> /dev/null
}

# Helper function for optional installation
install_optional() {
    local package=$1
    local source=$2
    local description=$3
    
    if is_installed "$package"; then
        print_success "$package already installed"
        return
    fi
    
    # Check if we're in auto mode
    if [ "${AUTO_INSTALL:-false}" = "true" ]; then
        # Auto-install in unattended mode
        if [ "$source" == "OFFICIAL" ]; then
            sudo pacman -S --needed --noconfirm "$package"
        else
            yay -S --needed --noconfirm "$package"
        fi
        print_success "$package installed (auto)"
    else
        # Interactive mode
        echo -n "Install $description? (y/N): "
        read -n 1 -r
        echo
        
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            if [ "$source" == "OFFICIAL" ]; then
                sudo pacman -S --needed --noconfirm "$package"
            else
                yay -S --needed --noconfirm "$package"
            fi
            print_success "$package installed"
        else
            print_info "Skipping $package"
        fi
    fi
}

# Development tools
install_dev_extras() {
    print_status "Additional development tools..."
    
    install_optional "github-desktop-bin" "AUR" "GitHub Desktop - Git GUI"
    install_optional "postman-bin" "AUR" "Postman - API testing"
    install_optional "insomnia-bin" "AUR" "Insomnia - REST client"
    install_optional "dbeaver" "OFFICIAL" "DBeaver - Database manager"
    install_optional "beekeeper-studio-bin" "AUR" "Beekeeper Studio - SQL editor"
}

# Note-taking and productivity
install_productivity_extras() {
    print_status "Productivity applications..."
    
    install_optional "notion-app-electron" "AUR" "Notion - All-in-one workspace"
    install_optional "logseq-desktop-bin" "AUR" "Logseq - Privacy-first notes"
    install_optional "joplin-desktop" "AUR" "Joplin - Open source notes"
    install_optional "todoist-electron" "AUR" "Todoist - Task management"
    install_optional "keepassxc" "OFFICIAL" "KeePassXC - Offline passwords"
    install_optional "anki" "OFFICIAL" "Anki - Spaced repetition learning"
}

# Communication extras
install_communication_extras() {
    print_status "Additional communication tools..."
    
    install_optional "zoom" "AUR" "Zoom - Video conferencing"
    install_optional "teams-for-linux" "AUR" "Microsoft Teams (unofficial)"
    install_optional "element-desktop" "OFFICIAL" "Element - Matrix client"
    install_optional "whatsapp-for-linux" "AUR" "WhatsApp Desktop"
    install_optional "skypeforlinux-bin" "AUR" "Skype"
}

# Media applications
install_media() {
    print_status "Media applications..."
    
    install_optional "spotify" "AUR" "Spotify - Music streaming"
    install_optional "youtube-music-bin" "AUR" "YouTube Music Desktop"
    install_optional "freetube-bin" "AUR" "FreeTube - Private YouTube"
    install_optional "obs-studio" "OFFICIAL" "OBS Studio - Streaming/recording"
    install_optional "mpv" "OFFICIAL" "MPV - Minimal media player"
    install_optional "celluloid" "OFFICIAL" "Celluloid - MPV GUI"
    install_optional "kodi" "OFFICIAL" "Kodi - Media center"
}

# Design and creative tools
install_creative() {
    print_status "Design and creative tools..."
    
    install_optional "figma-linux" "AUR" "Figma - Design tool"
    install_optional "drawio-desktop-bin" "AUR" "draw.io - Diagrams"
    install_optional "gimp" "OFFICIAL" "GIMP - Image editor"
    install_optional "inkscape" "OFFICIAL" "Inkscape - Vector graphics"
    install_optional "krita" "OFFICIAL" "Krita - Digital painting"
    install_optional "blender" "OFFICIAL" "Blender - 3D creation"
}

# System utilities extras
install_system_extras() {
    print_status "Additional system utilities..."
    
    install_optional "balena-etcher" "AUR" "Etcher - USB flasher"
    install_optional "ventoy-bin" "AUR" "Ventoy - Multi-boot USB"
    install_optional "stacer" "AUR" "Stacer - System optimizer"
    install_optional "bleachbit" "OFFICIAL" "BleachBit - System cleaner"
    install_optional "gparted" "OFFICIAL" "GParted - Partition editor"
    install_optional "baobab" "OFFICIAL" "Disk Usage Analyzer"
}

# Terminal tools extras
install_terminal_extras() {
    print_status "Additional terminal tools..."
    
    install_optional "hyper-bin" "AUR" "Hyper - Electron terminal"
    install_optional "tabby-bin" "AUR" "Tabby - Modern terminal"
    install_optional "terminator" "OFFICIAL" "Terminator - Multi-window terminal"
    install_optional "tmux" "OFFICIAL" "Tmux - Terminal multiplexer"
    install_optional "zellij" "OFFICIAL" "Zellij - Modern tmux alternative"
}

# Cloud storage clients
install_cloud_storage() {
    print_status "Cloud storage clients..."
    
    install_optional "dropbox" "AUR" "Dropbox"
    install_optional "onedrive-abraunegg" "AUR" "OneDrive client"
    install_optional "megasync-bin" "AUR" "MEGA sync client"
    install_optional "insync" "AUR" "Google Drive sync"
}

# Electron optimization
optimize_electron_apps() {
    print_status "Optimizing Electron applications..."
    
    # Create Electron flags configuration
    mkdir -p ~/.config
    cat > ~/.config/electron-flags.conf << 'EOF'
# Hardware acceleration
--enable-features=UseOzonePlatform
--ozone-platform=wayland
--enable-accelerated-video-decode
--enable-gpu-rasterization
--enable-zero-copy
--ignore-gpu-blocklist

# Performance
--max-old-space-size=4096
--disable-background-timer-throttling
EOF
    
    print_success "Electron apps optimized"
}

# Main execution with menu
main() {
    clear
    print_status "Extended Applications Installer"
    echo ""
    echo "This will install additional recommended applications."
    echo "You'll be prompted for each category."
    echo ""
    
    echo -n "Install development extras? (y/N): "
    read -n 1 -r dev_choice
    echo
    
    echo -n "Install productivity apps? (y/N): "
    read -n 1 -r prod_choice
    echo
    
    echo -n "Install communication extras? (y/N): "
    read -n 1 -r comm_choice
    echo
    
    echo -n "Install media applications? (y/N): "
    read -n 1 -r media_choice
    echo
    
    echo -n "Install creative tools? (y/N): "
    read -n 1 -r creative_choice
    echo
    
    echo -n "Install system utilities? (y/N): "
    read -n 1 -r system_choice
    echo
    
    echo -n "Install terminal extras? (y/N): "
    read -n 1 -r term_choice
    echo
    
    echo -n "Install cloud storage? (y/N): "
    read -n 1 -r cloud_choice
    echo
    
    # Install based on choices
    [[ $dev_choice =~ ^[Yy]$ ]] && install_dev_extras
    [[ $prod_choice =~ ^[Yy]$ ]] && install_productivity_extras
    [[ $comm_choice =~ ^[Yy]$ ]] && install_communication_extras
    [[ $media_choice =~ ^[Yy]$ ]] && install_media
    [[ $creative_choice =~ ^[Yy]$ ]] && install_creative
    [[ $system_choice =~ ^[Yy]$ ]] && install_system_extras
    [[ $term_choice =~ ^[Yy]$ ]] && install_terminal_extras
    [[ $cloud_choice =~ ^[Yy]$ ]] && install_cloud_storage
    
    # Always optimize Electron apps
    optimize_electron_apps
    
    print_success "Extended applications installation completed!"
}

main
