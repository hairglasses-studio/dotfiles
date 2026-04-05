#!/bin/bash

# Arch Linux Application Installer Script
# This script installs various development and productivity tools

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
       print_error "This script should not be run as root!"
       exit 1
    fi
}

# Function to check if a package is installed
is_installed() {
    pacman -Qi "$1" &> /dev/null
}

# Function to install AUR helper (yay)
install_aur_helper() {
    if ! command -v yay &> /dev/null; then
        print_status "Installing yay AUR helper..."
        sudo pacman -S --needed --noconfirm base-devel git
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

# Function to update system
update_system() {
    print_status "Updating system packages..."
    sudo pacman -Syu --noconfirm
    print_success "System updated"
}

# Function to enable multilib repository (needed for Steam)
enable_multilib() {
    print_status "Checking multilib repository..."
    if ! grep -q "^\[multilib\]" /etc/pacman.conf; then
        print_status "Enabling multilib repository for Steam..."
        sudo bash -c 'echo -e "\n[multilib]\nInclude = /etc/pacman.d/mirrorlist" >> /etc/pacman.conf'
        sudo pacman -Sy
        print_success "Multilib repository enabled"
    else
        print_success "Multilib repository already enabled"
    fi
}

# Main installation function
main() {
    clear
    echo "╔════════════════════════════════════════╗"
    echo "║   Arch Linux Application Installer    ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    check_root
    
    # Update system first
    update_system
    
    # Install yay for AUR packages
    install_aur_helper
    
    # Enable multilib for Steam
    enable_multilib
    
    # Install packages from official repositories
    print_status "Installing packages from official repositories..."
    
    OFFICIAL_PACKAGES=(
        "discord"        # Discord
        "steam"          # Steam (requires multilib)
        "tailscale"      # Tailscale VPN
        "nodejs"         # Node.js (needed for Claude Code)
        "npm"            # NPM (needed for Claude Code)
        "curl"           # Needed for various downloads
        "wget"           # Needed for various downloads
    )
    
    for package in "${OFFICIAL_PACKAGES[@]}"; do
        if is_installed "$package"; then
            print_success "$package is already installed"
        else
            print_status "Installing $package..."
            sudo pacman -S --needed --noconfirm "$package"
            print_success "$package installed"
        fi
    done
    
    # Install AUR packages
    print_status "Installing packages from AUR..."
    
    AUR_PACKAGES=(
        "google-chrome"              # Google Chrome
        "visual-studio-code-bin"     # VS Code
        "slack-desktop"              # Slack
        "chrome-remote-desktop"      # Chrome Remote Desktop
    )
    
    for package in "${AUR_PACKAGES[@]}"; do
        if is_installed "$package"; then
            print_success "$package is already installed"
        else
            print_status "Installing $package from AUR..."
            yay -S --noconfirm "$package"
            print_success "$package installed"
        fi
    done
    
    # Install Cursor IDE
    print_status "Installing Cursor IDE..."
    if [ ! -f /opt/cursor/cursor ]; then
        CURSOR_URL="https://downloader.cursor.sh/linux/appImage/x64"
        print_status "Downloading Cursor IDE..."
        wget -O /tmp/cursor.AppImage "$CURSOR_URL"
        chmod +x /tmp/cursor.AppImage
        
        # Extract and install
        sudo mkdir -p /opt/cursor
        sudo mv /tmp/cursor.AppImage /opt/cursor/cursor.AppImage
        
        # Create desktop entry
        sudo tee /usr/share/applications/cursor.desktop > /dev/null <<EOF
[Desktop Entry]
Name=Cursor
Comment=AI-powered code editor
Exec=/opt/cursor/cursor.AppImage
Icon=cursor
Terminal=false
Type=Application
Categories=Development;IDE;
EOF
        
        # Create symlink for command line
        sudo ln -sf /opt/cursor/cursor.AppImage /usr/local/bin/cursor
        
        print_success "Cursor IDE installed"
    else
        print_success "Cursor IDE is already installed"
    fi
    
    # Install Claude Code
    print_status "Installing Claude Code..."
    if ! command -v claude &> /dev/null; then
        print_status "Installing Claude Code via npm..."
        sudo npm install -g @anthropic-ai/claude-cli
        print_success "Claude Code installed"
        print_warning "Remember to authenticate Claude Code with: claude auth"
    else
        print_success "Claude Code is already installed"
    fi
    
    # Check for Cline Tools
    print_warning "Note: 'Cline Tools' was not found in standard repositories."
    print_status "Possible alternatives:"
    echo "  - If you meant 'Cline' (VS Code extension), install it from VS Code marketplace"
    echo "  - If you meant 'CLI tools', many are already installed (git, curl, wget, etc.)"
    echo "  - If it's a specific tool, please provide more details"
    
    # Enable and start services
    print_status "Configuring services..."
    
    # Tailscale
    if systemctl is-enabled tailscaled &> /dev/null; then
        print_success "Tailscale service already enabled"
    else
        sudo systemctl enable --now tailscaled
        print_success "Tailscale service enabled and started"
        print_warning "Run 'sudo tailscale up' to connect to your Tailscale network"
    fi
    
    # Chrome Remote Desktop setup reminder
    print_warning "Chrome Remote Desktop requires additional setup:"
    echo "  1. Open Google Chrome"
    echo "  2. Visit: https://remotedesktop.google.com/access"
    echo "  3. Follow the setup instructions"
    
    # Final summary
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║        Installation Complete!          ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    print_success "The following applications have been installed:"
    echo "  • Google Chrome"
    echo "  • Visual Studio Code"
    echo "  • Cursor IDE"
    echo "  • Claude Code (authenticate with: claude auth)"
    echo "  • Discord"
    echo "  • Slack"
    echo "  • Chrome Remote Desktop"
    echo "  • Steam"
    echo "  • Tailscale (connect with: sudo tailscale up)"
    echo ""
    print_warning "Some applications may require logout/login or reboot to appear in menus"
    
    # Optional reboot prompt
    read -p "Would you like to reboot now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo reboot
    fi
}

# Run main function
main
