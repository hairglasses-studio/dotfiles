#!/bin/bash

# Arch Linux Dotfiles Setup Script
# Quick setup for new Arch installations

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_header() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║          Arch Linux Dotfiles Setup Script             ║${NC}"
    echo -e "${BLUE}║           https://github.com/hairglasses/dotfiles-arch ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

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

# Check if running on Arch
check_arch() {
    if [ ! -f /etc/arch-release ]; then
        print_error "This script is designed for Arch Linux"
        exit 1
    fi
}

# Make all scripts executable
make_executable() {
    print_status "Making scripts executable..."
    chmod +x scripts/*.sh
    chmod +x scripts/helpers/*.sh
    print_success "Scripts are now executable"
}

# Install yay if not present
install_yay() {
    if ! command -v yay &> /dev/null; then
        print_status "Installing yay AUR helper..."
        sudo pacman -S --needed --noconfirm base-devel git
        cd /tmp
        git clone https://aur.archlinux.org/yay.git
        cd yay
        makepkg -si --noconfirm
        cd -
        rm -rf /tmp/yay
        print_success "yay installed successfully"
    else
        print_success "yay is already installed"
    fi
}

# Copy configurations
copy_configs() {
    print_status "Copying configuration files..."
    
    # Electron flags
    if [ ! -f ~/.config/electron-flags.conf ]; then
        cp configs/electron-flags.conf ~/.config/
        print_success "Electron flags configured"
    else
        print_warning "Electron flags already exist, skipping"
    fi
    
    # ZRAM configuration (needs root)
    if [ ! -f /etc/systemd/zram-generator.conf ]; then
        print_status "Installing ZRAM configuration (requires sudo)..."
        sudo cp configs/zram-generator.conf /etc/systemd/
        sudo pacman -S --needed --noconfirm zram-generator
        print_success "ZRAM configured"
    else
        print_warning "ZRAM already configured, skipping"
    fi
}

# Quick system update
system_update() {
    print_status "Updating system packages..."
    sudo pacman -Syu --noconfirm
    print_success "System updated"
}

# Show menu
show_menu() {
    echo ""
    echo "What would you like to do?"
    echo ""
    echo "  1) Quick Setup (yay + configs only)"
    echo "  2) Install Essential Apps"
    echo "  3) Run Extended Installer (interactive)"
    echo "  4) Configure System Services (requires sudo)"
    echo "  5) Optimize Electron Apps"
    echo "  6) Full Setup (everything)"
    echo "  7) Exit"
    echo ""
    read -p "Enter your choice (1-7): " choice
}

# Main function
main() {
    print_header
    check_arch
    
    # Always make scripts executable first
    make_executable
    
    while true; do
        show_menu
        
        case $choice in
            1)
                print_status "Running Quick Setup..."
                install_yay
                copy_configs
                print_success "Quick setup complete!"
                ;;
            2)
                print_status "Installing Essential Apps..."
                system_update
                install_yay
                ./scripts/arch-installer.sh
                ;;
            3)
                print_status "Running Extended Installer..."
                system_update
                install_yay
                ./scripts/arch-extended-installer.sh
                ;;
            4)
                print_status "Configuring System Services..."
                sudo ./scripts/helpers/configure-services.sh
                ;;
            5)
                print_status "Optimizing Electron Apps..."
                ./scripts/helpers/optimize-electron.sh
                ;;
            6)
                print_status "Running Full Setup..."
                system_update
                install_yay
                copy_configs
                ./scripts/arch-extended-installer.sh
                sudo ./scripts/helpers/configure-services.sh
                ./scripts/helpers/optimize-electron.sh
                print_success "Full setup complete!"
                break
                ;;
            7)
                print_status "Exiting..."
                exit 0
                ;;
            *)
                print_error "Invalid choice"
                ;;
        esac
        
        if [ "$choice" != "7" ]; then
            echo ""
            read -p "Press Enter to continue..."
        fi
    done
    
    # Final message
    echo ""
    print_success "Setup complete!"
    echo ""
    echo "Next steps:"
    echo "  • Reboot your system for all changes to take effect"
    echo "  • Configure Timeshift for system backups"
    echo "  • Authenticate services (Claude Code, Tailscale, etc.)"
    echo "  • Customize your desktop environment"
    echo ""
    echo "For more information, see:"
    echo "  • README.md - General documentation"
    echo "  • docs/MIGRATION_GUIDE.md - Ubuntu to Arch migration"
    echo "  • docs/APP_RECOMMENDATIONS.md - Application details"
    echo "  • docs/QUICK_REFERENCE.md - Command reference"
    echo ""
}

# Run main
main
