#!/bin/bash

# Extended Arch Linux Application Installer Script
# Includes additional Electron/native web apps for daily usage

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
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

print_info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

print_category() {
    echo -e "\n${MAGENTA}════ $1 ════${NC}"
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

# Function to enable multilib repository
enable_multilib() {
    print_status "Checking multilib repository..."
    if ! grep -q "^\[multilib\]" /etc/pacman.conf; then
        print_status "Enabling multilib repository..."
        sudo bash -c 'echo -e "\n[multilib]\nInclude = /etc/pacman.d/mirrorlist" >> /etc/pacman.conf'
        sudo pacman -Sy
        print_success "Multilib repository enabled"
    else
        print_success "Multilib repository already enabled"
    fi
}

# Function to show selection menu
show_menu() {
    clear
    echo "╔════════════════════════════════════════════════════╗"
    echo "║   Extended Arch Linux Application Installer       ║"
    echo "╚════════════════════════════════════════════════════╝"
    echo ""
    echo "Select installation option:"
    echo ""
    echo "  1) Essential Apps Only (Your original list)"
    echo "  2) Recommended Additions (Additional useful apps)"
    echo "  3) Full Installation (Everything)"
    echo "  4) Custom Selection"
    echo "  5) Exit"
    echo ""
    read -p "Enter your choice (1-5): " choice
    return $choice
}

# Function to install essential apps (original list)
install_essential() {
    print_category "Installing Essential Applications"
    
    # Official repository packages
    OFFICIAL_PACKAGES=(
        "discord"
        "steam"
        "tailscale"
        "nodejs"
        "npm"
        "curl"
        "wget"
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
    
    # AUR packages
    AUR_PACKAGES=(
        "google-chrome"
        "visual-studio-code-bin"
        "slack-desktop"
        "chrome-remote-desktop"
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
    install_cursor_ide
    
    # Install Claude Code
    install_claude_code
}

# Function to install recommended apps
install_recommended() {
    print_category "Installing Recommended Applications"
    
    # Development & Code Editors
    print_info "Development Tools:"
    install_if_wanted "sublime-text-4" "AUR" "Sublime Text - Fast native text editor"
    install_if_wanted "github-desktop-bin" "AUR" "GitHub Desktop - Git GUI client"
    install_if_wanted "postman-bin" "AUR" "Postman - API development platform"
    install_if_wanted "insomnia-bin" "AUR" "Insomnia - REST/GraphQL client"
    install_if_wanted "dbeaver" "OFFICIAL" "DBeaver - Universal database tool"
    
    # Communication & Collaboration
    print_info "Communication Tools:"
    install_if_wanted "telegram-desktop" "OFFICIAL" "Telegram Desktop"
    install_if_wanted "signal-desktop" "OFFICIAL" "Signal - Secure messenger"
    install_if_wanted "zoom" "AUR" "Zoom - Video conferencing"
    install_if_wanted "teams-for-linux" "AUR" "Microsoft Teams (unofficial)"
    install_if_wanted "element-desktop" "OFFICIAL" "Element - Matrix client"
    
    # Productivity & Notes
    print_info "Productivity Applications:"
    install_if_wanted "obsidian" "OFFICIAL" "Obsidian - Knowledge base"
    install_if_wanted "notion-app-electron" "AUR" "Notion - All-in-one workspace"
    install_if_wanted "logseq-desktop-bin" "AUR" "Logseq - Privacy-first knowledge base"
    install_if_wanted "joplin-desktop" "AUR" "Joplin - Note-taking app"
    install_if_wanted "todoist-electron" "AUR" "Todoist - Task manager"
    install_if_wanted "bitwarden" "OFFICIAL" "Bitwarden - Password manager"
    install_if_wanted "keepassxc" "OFFICIAL" "KeePassXC - Offline password manager"
    
    # Media & Entertainment
    print_info "Media Applications:"
    install_if_wanted "spotify" "AUR" "Spotify - Music streaming"
    install_if_wanted "youtube-music-bin" "AUR" "YouTube Music Desktop"
    install_if_wanted "freetube-bin" "AUR" "FreeTube - Private YouTube client"
    install_if_wanted "vlc" "OFFICIAL" "VLC - Media player"
    install_if_wanted "obs-studio" "OFFICIAL" "OBS Studio - Broadcasting/recording"
    
    # System & Utilities
    print_info "System Utilities:"
    install_if_wanted "balena-etcher" "AUR" "Etcher - USB/SD card writer"
    install_if_wanted "timeshift" "OFFICIAL" "Timeshift - System backup tool"
    install_if_wanted "stacer" "AUR" "Stacer - System optimizer & monitor"
    install_if_wanted "ulauncher" "AUR" "Ulauncher - Application launcher"
    install_if_wanted "flameshot" "OFFICIAL" "Flameshot - Screenshot tool"
    
    # Design & Creative
    print_info "Design Tools:"
    install_if_wanted "figma-linux" "AUR" "Figma - Design tool"
    install_if_wanted "drawio-desktop-bin" "AUR" "draw.io - Diagram editor"
    
    # Terminal Enhancements
    print_info "Terminal Tools:"
    install_if_wanted "kitty" "OFFICIAL" "Kitty - GPU-accelerated terminal"
    install_if_wanted "alacritty" "OFFICIAL" "Alacritty - Fast terminal emulator"
    install_if_wanted "hyper-bin" "AUR" "Hyper - Electron-based terminal"
    install_if_wanted "tabby-bin" "AUR" "Tabby - Modern terminal"
}

# Helper function to install with user confirmation
install_if_wanted() {
    local package=$1
    local source=$2
    local description=$3
    
    if is_installed "$package"; then
        print_success "$package is already installed"
        return
    fi
    
    echo -n "  Install $description? (y/N): "
    read -n 1 -r install_choice
    echo
    
    if [[ $install_choice =~ ^[Yy]$ ]]; then
        if [ "$source" == "OFFICIAL" ]; then
            sudo pacman -S --needed --noconfirm "$package"
        else
            yay -S --noconfirm "$package"
        fi
        print_success "$package installed"
    else
        print_info "Skipping $package"
    fi
}

# Install Cursor IDE function
install_cursor_ide() {
    print_status "Installing Cursor IDE..."
    if [ ! -f /opt/cursor/cursor.AppImage ]; then
        CURSOR_URL="https://downloader.cursor.sh/linux/appImage/x64"
        print_status "Downloading Cursor IDE..."
        wget -q -O /tmp/cursor.AppImage "$CURSOR_URL"
        chmod +x /tmp/cursor.AppImage
        
        sudo mkdir -p /opt/cursor
        sudo mv /tmp/cursor.AppImage /opt/cursor/cursor.AppImage
        
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
        
        sudo ln -sf /opt/cursor/cursor.AppImage /usr/local/bin/cursor
        print_success "Cursor IDE installed"
    else
        print_success "Cursor IDE is already installed"
    fi
}

# Install Claude Code function
install_claude_code() {
    print_status "Installing Claude Code..."
    if ! command -v claude &> /dev/null; then
        print_status "Installing Claude Code via npm..."
        sudo npm install -g @anthropic-ai/claude-cli
        print_success "Claude Code installed"
        print_warning "Remember to authenticate Claude Code with: claude auth"
    else
        print_success "Claude Code is already installed"
    fi
}

# Function for custom selection
custom_selection() {
    print_category "Custom Application Selection"
    
    CATEGORIES=(
        "Development Tools"
        "Communication"
        "Productivity"
        "Media"
        "System Utilities"
        "Design"
        "Terminal"
    )
    
    echo "Select categories to install:"
    echo ""
    for i in "${!CATEGORIES[@]}"; do
        echo "  $((i+1))) ${CATEGORIES[$i]}"
    done
    echo "  8) Back to main menu"
    echo ""
    read -p "Enter your choices (space-separated, e.g., 1 3 5): " -a choices
    
    for choice in "${choices[@]}"; do
        case $choice in
            1) install_dev_tools ;;
            2) install_communication ;;
            3) install_productivity ;;
            4) install_media ;;
            5) install_system_utils ;;
            6) install_design ;;
            7) install_terminal ;;
            8) return ;;
        esac
    done
}

# Category installation functions
install_dev_tools() {
    print_category "Development Tools"
    install_if_wanted "sublime-text-4" "AUR" "Sublime Text"
    install_if_wanted "github-desktop-bin" "AUR" "GitHub Desktop"
    install_if_wanted "postman-bin" "AUR" "Postman"
    install_if_wanted "insomnia-bin" "AUR" "Insomnia"
    install_if_wanted "dbeaver" "OFFICIAL" "DBeaver"
}

install_communication() {
    print_category "Communication Tools"
    install_if_wanted "telegram-desktop" "OFFICIAL" "Telegram"
    install_if_wanted "signal-desktop" "OFFICIAL" "Signal"
    install_if_wanted "zoom" "AUR" "Zoom"
    install_if_wanted "teams-for-linux" "AUR" "Teams"
    install_if_wanted "element-desktop" "OFFICIAL" "Element"
}

install_productivity() {
    print_category "Productivity Tools"
    install_if_wanted "obsidian" "OFFICIAL" "Obsidian"
    install_if_wanted "notion-app-electron" "AUR" "Notion"
    install_if_wanted "logseq-desktop-bin" "AUR" "Logseq"
    install_if_wanted "joplin-desktop" "AUR" "Joplin"
    install_if_wanted "bitwarden" "OFFICIAL" "Bitwarden"
}

install_media() {
    print_category "Media Applications"
    install_if_wanted "spotify" "AUR" "Spotify"
    install_if_wanted "youtube-music-bin" "AUR" "YouTube Music"
    install_if_wanted "freetube-bin" "AUR" "FreeTube"
    install_if_wanted "vlc" "OFFICIAL" "VLC"
    install_if_wanted "obs-studio" "OFFICIAL" "OBS Studio"
}

install_system_utils() {
    print_category "System Utilities"
    install_if_wanted "balena-etcher" "AUR" "Balena Etcher"
    install_if_wanted "timeshift" "OFFICIAL" "Timeshift"
    install_if_wanted "stacer" "AUR" "Stacer"
    install_if_wanted "ulauncher" "AUR" "Ulauncher"
    install_if_wanted "flameshot" "OFFICIAL" "Flameshot"
}

install_design() {
    print_category "Design Tools"
    install_if_wanted "figma-linux" "AUR" "Figma"
    install_if_wanted "drawio-desktop-bin" "AUR" "draw.io"
}

install_terminal() {
    print_category "Terminal Emulators"
    install_if_wanted "kitty" "OFFICIAL" "Kitty"
    install_if_wanted "alacritty" "OFFICIAL" "Alacritty"
    install_if_wanted "hyper-bin" "AUR" "Hyper"
    install_if_wanted "tabby-bin" "AUR" "Tabby"
}

# Post-installation configuration
post_install_config() {
    print_category "Post-Installation Configuration"
    
    # Tailscale
    if command -v tailscale &> /dev/null; then
        if ! systemctl is-enabled tailscaled &> /dev/null; then
            sudo systemctl enable --now tailscaled
            print_success "Tailscale service enabled"
            print_info "Run 'sudo tailscale up' to connect"
        fi
    fi
    
    # Chrome Remote Desktop
    if is_installed "chrome-remote-desktop"; then
        print_info "Chrome Remote Desktop: Visit https://remotedesktop.google.com/access"
    fi
    
    # Claude Code
    if command -v claude &> /dev/null; then
        print_info "Claude Code: Run 'claude auth' to authenticate"
    fi
    
    # Timeshift
    if command -v timeshift &> /dev/null; then
        print_info "Timeshift: Configure backups with 'sudo timeshift-gtk'"
    fi
}

# Main function
main() {
    check_root
    
    while true; do
        show_menu
        choice=$?
        
        case $choice in
            1)
                update_system
                install_aur_helper
                enable_multilib
                install_essential
                post_install_config
                break
                ;;
            2)
                update_system
                install_aur_helper
                install_recommended
                post_install_config
                break
                ;;
            3)
                update_system
                install_aur_helper
                enable_multilib
                install_essential
                install_recommended
                post_install_config
                break
                ;;
            4)
                update_system
                install_aur_helper
                enable_multilib
                custom_selection
                post_install_config
                break
                ;;
            5)
                print_info "Installation cancelled"
                exit 0
                ;;
            *)
                print_error "Invalid choice"
                sleep 2
                ;;
        esac
    done
    
    # Final summary
    echo ""
    echo "╔════════════════════════════════════════════════════╗"
    echo "║        Installation Complete!                     ║"
    echo "╚════════════════════════════════════════════════════╝"
    echo ""
    print_success "All selected applications have been installed!"
    echo ""
    print_warning "Some applications may require logout/login or reboot"
    
    # Optional reboot
    read -p "Would you like to reboot now? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo reboot
    fi
}

# Run main function
main
