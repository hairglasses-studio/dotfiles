#!/bin/bash

#############################################
# 02 - Essential Applications
# Core applications for daily use
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

# Check if package is installed
is_installed() {
    pacman -Qi "$1" &> /dev/null
}

# Install browsers
install_browsers() {
    print_status "Installing web browsers..."
    
    # Firefox
    if ! is_installed firefox; then
        sudo pacman -S --needed --noconfirm firefox
        print_success "Firefox installed"
    else
        print_success "Firefox already installed"
    fi
    
    # Google Chrome
    if ! is_installed google-chrome; then
        yay -S --needed --noconfirm google-chrome
        print_success "Google Chrome installed"
    else
        print_success "Google Chrome already installed"
    fi
}

# Install code editors
install_editors() {
    print_status "Installing code editors..."
    
    # VS Code
    if ! is_installed visual-studio-code-bin; then
        yay -S --needed --noconfirm visual-studio-code-bin
        print_success "Visual Studio Code installed"
    else
        print_success "Visual Studio Code already installed"
    fi
    
    # Sublime Text
    if ! is_installed sublime-text-4; then
        yay -S --needed --noconfirm sublime-text-4
        print_success "Sublime Text installed"
    else
        print_success "Sublime Text already installed"
    fi
}

# Install Cursor IDE
install_cursor() {
    print_status "Installing Cursor IDE..."
    
    if [ ! -f /opt/cursor/cursor.AppImage ]; then
        # Download Cursor AppImage
        CURSOR_URL="https://downloader.cursor.sh/linux/appImage/x64"
        wget -q -O /tmp/cursor.AppImage "$CURSOR_URL"
        chmod +x /tmp/cursor.AppImage
        
        # Install to /opt
        sudo mkdir -p /opt/cursor
        sudo mv /tmp/cursor.AppImage /opt/cursor/cursor.AppImage
        
        # Create desktop entry
        sudo tee /usr/share/applications/cursor.desktop > /dev/null <<'EOF'
[Desktop Entry]
Name=Cursor
Comment=AI-powered code editor
Exec=/opt/cursor/cursor.AppImage
Icon=cursor
Terminal=false
Type=Application
Categories=Development;IDE;TextEditor;
EOF
        
        # Create command line launcher
        sudo ln -sf /opt/cursor/cursor.AppImage /usr/local/bin/cursor
        
        print_success "Cursor IDE installed"
    else
        print_success "Cursor IDE already installed"
    fi
}

# Install Claude Code
install_claude_code() {
    print_status "Installing Claude Code..."
    
    # Ensure npm is installed
    if ! command -v npm &> /dev/null; then
        sudo pacman -S --needed --noconfirm nodejs npm
    fi
    
    if ! command -v claude &> /dev/null; then
        sudo npm install -g @anthropic-ai/claude-cli
        print_success "Claude Code installed"
        print_warning "Remember to authenticate: claude auth"
    else
        print_success "Claude Code already installed"
    fi
}

# Install communication tools
install_communication() {
    print_status "Installing communication tools..."
    
    # Discord
    if ! is_installed discord; then
        sudo pacman -S --needed --noconfirm discord
        print_success "Discord installed"
    else
        print_success "Discord already installed"
    fi
    
    # Slack
    if ! is_installed slack-desktop; then
        yay -S --needed --noconfirm slack-desktop
        print_success "Slack installed"
    else
        print_success "Slack already installed"
    fi
    
    # Telegram
    if ! is_installed telegram-desktop; then
        sudo pacman -S --needed --noconfirm telegram-desktop
        print_success "Telegram installed"
    else
        print_success "Telegram already installed"
    fi
    
    # Signal
    if ! is_installed signal-desktop; then
        sudo pacman -S --needed --noconfirm signal-desktop
        print_success "Signal installed"
    else
        print_success "Signal already installed"
    fi
}

# Install remote access tools
install_remote_tools() {
    print_status "Installing remote access tools..."
    
    # Chrome Remote Desktop
    if ! is_installed chrome-remote-desktop; then
        yay -S --needed --noconfirm chrome-remote-desktop
        print_success "Chrome Remote Desktop installed"
        print_warning "Setup at: https://remotedesktop.google.com/access"
    else
        print_success "Chrome Remote Desktop already installed"
    fi
    
    # Tailscale
    if ! is_installed tailscale; then
        sudo pacman -S --needed --noconfirm tailscale
        sudo systemctl enable --now tailscaled
        print_success "Tailscale installed"
        print_warning "Connect with: sudo tailscale up"
    else
        print_success "Tailscale already installed"
    fi
}

# Install gaming
install_gaming() {
    print_status "Installing gaming tools..."
    
    # Steam
    if ! is_installed steam; then
        sudo pacman -S --needed --noconfirm steam
        # Install 32-bit graphics drivers
        if lspci | grep -i nvidia > /dev/null; then
            sudo pacman -S --needed --noconfirm lib32-nvidia-utils
        else
            sudo pacman -S --needed --noconfirm lib32-mesa lib32-vulkan-radeon
        fi
        print_success "Steam installed"
    else
        print_success "Steam already installed"
    fi
}

# Install productivity tools
install_productivity() {
    print_status "Installing productivity tools..."
    
    # Obsidian
    if ! is_installed obsidian; then
        sudo pacman -S --needed --noconfirm obsidian
        print_success "Obsidian installed"
    else
        print_success "Obsidian already installed"
    fi
    
    # Bitwarden
    if ! is_installed bitwarden; then
        sudo pacman -S --needed --noconfirm bitwarden
        print_success "Bitwarden installed"
    else
        print_success "Bitwarden already installed"
    fi
    
    # Timeshift
    if ! is_installed timeshift; then
        sudo pacman -S --needed --noconfirm timeshift
        print_success "Timeshift installed"
        print_warning "Configure backups: sudo timeshift-gtk"
    else
        print_success "Timeshift already installed"
    fi
}

# Install system utilities
install_utilities() {
    print_status "Installing system utilities..."
    
    # Flameshot
    if ! is_installed flameshot; then
        sudo pacman -S --needed --noconfirm flameshot
        print_success "Flameshot installed"
    else
        print_success "Flameshot already installed"
    fi
    
    # Ulauncher
    if ! is_installed ulauncher; then
        yay -S --needed --noconfirm ulauncher
        print_success "Ulauncher installed"
    else
        print_success "Ulauncher already installed"
    fi
    
    # VLC
    if ! is_installed vlc; then
        sudo pacman -S --needed --noconfirm vlc
        print_success "VLC installed"
    else
        print_success "VLC already installed"
    fi
}

# Install terminal emulators
install_terminals() {
    print_status "Installing terminal emulators..."
    
    # Kitty
    if ! is_installed kitty; then
        sudo pacman -S --needed --noconfirm kitty
        print_success "Kitty installed"
    else
        print_success "Kitty already installed"
    fi
    
    # Alacritty
    if ! is_installed alacritty; then
        sudo pacman -S --needed --noconfirm alacritty
        print_success "Alacritty installed"
    else
        print_success "Alacritty already installed"
    fi
}

# Main execution
main() {
    print_status "Installing essential applications..."
    
    install_browsers
    install_editors
    install_cursor
    install_claude_code
    install_communication
    install_remote_tools
    install_gaming
    install_productivity
    install_utilities
    install_terminals
    
    print_success "Essential applications installed!"
}

main
