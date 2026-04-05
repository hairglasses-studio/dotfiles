#!/bin/bash

#############################################
# Arch Linux Dotfiles Installation Manager
# https://github.com/hairglasses/dotfiles-arch
#############################################

set -e  # Exit on error

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPTS_DIR="$SCRIPT_DIR/scripts"
CONFIGS_DIR="$SCRIPT_DIR/configs"
DOCS_DIR="$SCRIPT_DIR/docs"
BACKUP_DIR="$SCRIPT_DIR/backup"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Functions
print_header() {
    echo -e "\n${MAGENTA}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${MAGENTA}║${CYAN}          Arch Linux Dotfiles Installer                ${MAGENTA}║${NC}"
    echo -e "${MAGENTA}║${CYAN}     https://github.com/hairglasses/dotfiles-arch      ${MAGENTA}║${NC}"
    echo -e "${MAGENTA}╚════════════════════════════════════════════════════════╝${NC}\n"
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

print_section() {
    echo -e "\n${CYAN}════ $1 ════${NC}"
}

# Check if running as root
check_root() {
    if [[ $EUID -eq 0 ]]; then
       print_error "This script should not be run as root!"
       exit 1
    fi
}

# Check if we're on Arch Linux
check_arch() {
    if [ ! -f /etc/arch-release ]; then
        print_error "This script is designed for Arch Linux!"
        print_warning "Detected: $(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
}

# Display usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

OPTIONS:
    --full          Install everything (recommended for new systems)
    --full-auto     Install everything without prompts (unattended)
    --essential     Install only essential applications
    --extended      Install extended application set
    --dev           Install development tools only
    --shell         Install shell enhancements (starship, zsh, etc.)
    --config        Apply configurations only
    --interactive   Interactive selection mode
    --restore       Restore from backup
    --backup        Backup current configurations
    --help          Show this help message

EXAMPLES:
    $0 --full       # Complete installation for new system
    $0 --full-auto  # Unattended full installation
    $0 --essential  # Minimal installation
    $0 --dev        # Development environment only
    $0 --shell      # Shell enhancements only
    $0              # Interactive mode (default)

EOF
}

# Interactive menu
show_menu() {
    clear
    print_header
    
    echo "What would you like to install?"
    echo ""
    echo "  1) Full Installation (Everything)"
    echo "  2) Full Installation - Unattended (No prompts)"
    echo "  3) Essential Apps Only"
    echo "  4) Extended Apps (Productivity & Media)"
    echo "  5) Development Environment"
    echo "  6) Shell Enhancements (Starship, Zsh, etc.)"
    echo "  7) Apply Configurations Only"
    echo "  8) Custom Selection"
    echo "  9) Backup Current Configurations"
    echo "  10) Restore from Backup"
    echo "  11) Exit"
    echo ""
    read -p "Enter your choice (1-11): " choice
}

# Run installation scripts
run_script() {
    local script=$1
    local description=$2
    
    if [ -f "$script" ]; then
        print_section "$description"
        bash "$script"
        print_success "$description completed"
    else
        print_error "Script not found: $script"
    fi
}

# Full installation
install_full() {
    print_section "Starting Full Installation"
    
    run_script "$SCRIPTS_DIR/01-base-system.sh" "Base System Setup"
    run_script "$SCRIPTS_DIR/02-essential-apps.sh" "Essential Applications"
    run_script "$SCRIPTS_DIR/03-extended-apps.sh" "Extended Applications"
    run_script "$SCRIPTS_DIR/04-dev-tools.sh" "Development Tools"
    run_script "$SCRIPTS_DIR/05-system-config.sh" "System Configuration"
    run_script "$SCRIPTS_DIR/06-dotfiles.sh" "Dotfiles Setup"
    run_script "$SCRIPTS_DIR/07-shell-enhancements.sh" "Shell Enhancements"
    
    print_success "Full installation completed!"
}

# Full unattended installation
install_full_auto() {
    print_section "Starting Unattended Full Installation"
    export AUTO_INSTALL=true
    
    run_script "$SCRIPTS_DIR/01-base-system.sh" "Base System Setup"
    run_script "$SCRIPTS_DIR/02-essential-apps.sh" "Essential Applications"
    
    # Run extended apps with auto-yes
    print_section "Extended Applications (Auto-Installing All)"
    yes | bash "$SCRIPTS_DIR/03-extended-apps.sh"
    
    run_script "$SCRIPTS_DIR/04-dev-tools.sh" "Development Tools"
    run_script "$SCRIPTS_DIR/05-system-config.sh" "System Configuration"
    run_script "$SCRIPTS_DIR/06-dotfiles.sh" "Dotfiles Setup"
    run_script "$SCRIPTS_DIR/07-shell-enhancements.sh" "Shell Enhancements"
    
    print_success "Unattended full installation completed!"
}

# Shell enhancements installation
install_shell() {
    print_section "Installing Shell Enhancements"
    
    run_script "$SCRIPTS_DIR/01-base-system.sh" "Base System Setup"
    run_script "$SCRIPTS_DIR/07-shell-enhancements.sh" "Shell Enhancements"
    run_script "$SCRIPTS_DIR/06-dotfiles.sh" "Dotfiles Setup"
    
    print_success "Shell enhancements installed!"
}

# Essential installation
install_essential() {
    print_section "Starting Essential Installation"
    
    run_script "$SCRIPTS_DIR/01-base-system.sh" "Base System Setup"
    run_script "$SCRIPTS_DIR/02-essential-apps.sh" "Essential Applications"
    run_script "$SCRIPTS_DIR/05-system-config.sh" "System Configuration"
    run_script "$SCRIPTS_DIR/06-dotfiles.sh" "Dotfiles Setup"
    
    print_success "Essential installation completed!"
}

# Extended apps installation
install_extended() {
    print_section "Starting Extended Apps Installation"
    
    run_script "$SCRIPTS_DIR/01-base-system.sh" "Base System Setup"
    run_script "$SCRIPTS_DIR/03-extended-apps.sh" "Extended Applications"
    
    print_success "Extended apps installation completed!"
}

# Development tools installation
install_dev() {
    print_section "Starting Development Environment Setup"
    
    run_script "$SCRIPTS_DIR/01-base-system.sh" "Base System Setup"
    run_script "$SCRIPTS_DIR/04-dev-tools.sh" "Development Tools"
    
    print_success "Development environment setup completed!"
}

# Apply configurations
apply_configs() {
    print_section "Applying Configurations"
    
    run_script "$SCRIPTS_DIR/06-dotfiles.sh" "Dotfiles Setup"
    
    print_success "Configurations applied!"
}

# Custom selection
custom_selection() {
    local scripts=()
    
    clear
    print_header
    echo "Select components to install:"
    echo ""
    echo "  [ ] 1. Base System Setup"
    echo "  [ ] 2. Essential Applications"
    echo "  [ ] 3. Extended Applications"
    echo "  [ ] 4. Development Tools"
    echo "  [ ] 5. System Configuration"
    echo "  [ ] 6. Dotfiles Setup"
    echo "  [ ] 7. Shell Enhancements"
    echo ""
    echo "Enter numbers separated by spaces (e.g., 1 3 4):"
    read -a selections
    
    for selection in "${selections[@]}"; do
        case $selection in
            1) scripts+=("$SCRIPTS_DIR/01-base-system.sh:Base System Setup") ;;
            2) scripts+=("$SCRIPTS_DIR/02-essential-apps.sh:Essential Applications") ;;
            3) scripts+=("$SCRIPTS_DIR/03-extended-apps.sh:Extended Applications") ;;
            4) scripts+=("$SCRIPTS_DIR/04-dev-tools.sh:Development Tools") ;;
            5) scripts+=("$SCRIPTS_DIR/05-system-config.sh:System Configuration") ;;
            6) scripts+=("$SCRIPTS_DIR/06-dotfiles.sh:Dotfiles Setup") ;;
            7) scripts+=("$SCRIPTS_DIR/07-shell-enhancements.sh:Shell Enhancements") ;;
        esac
    done
    
    for script_info in "${scripts[@]}"; do
        IFS=':' read -r script description <<< "$script_info"
        run_script "$script" "$description"
    done
    
    print_success "Custom installation completed!"
}

# Backup configurations
backup_configs() {
    print_section "Backing up configurations"
    
    if [ -f "$BACKUP_DIR/backup-configs.sh" ]; then
        bash "$BACKUP_DIR/backup-configs.sh"
        print_success "Backup completed!"
    else
        print_error "Backup script not found!"
    fi
}

# Restore from backup
restore_backup() {
    print_section "Restoring from backup"
    
    if [ -f "$BACKUP_DIR/restore-configs.sh" ]; then
        bash "$BACKUP_DIR/restore-configs.sh"
        print_success "Restore completed!"
    else
        print_error "Restore script not found!"
    fi
}

# Post-installation summary
show_summary() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║            Installation Complete!                     ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Reboot your system"
    echo "  2. Configure Timeshift for backups"
    echo "  3. Set up Tailscale: sudo tailscale up"
    echo "  4. Authenticate Claude Code: claude auth"
    echo "  5. Check docs/QUICK_REFERENCE.md for usage tips"
    echo ""
    print_warning "Some applications may require logout/login to appear in menus"
}

# Main execution
main() {
    # Check prerequisites
    check_root
    check_arch
    
    # Parse command line arguments
    case "${1:-}" in
        --full)
            install_full
            show_summary
            ;;
        --full-auto)
            install_full_auto
            show_summary
            ;;
        --essential)
            install_essential
            show_summary
            ;;
        --extended)
            install_extended
            show_summary
            ;;
        --dev)
            install_dev
            show_summary
            ;;
        --shell)
            install_shell
            show_summary
            ;;
        --config)
            apply_configs
            ;;
        --interactive|"")
            while true; do
                show_menu
                case $choice in
                    1) install_full; show_summary; break ;;
                    2) install_full_auto; show_summary; break ;;
                    3) install_essential; show_summary; break ;;
                    4) install_extended; show_summary; break ;;
                    5) install_dev; show_summary; break ;;
                    6) install_shell; show_summary; break ;;
                    7) apply_configs; break ;;
                    8) custom_selection; show_summary; break ;;
                    9) backup_configs; break ;;
                    10) restore_backup; break ;;
                    11) print_status "Exiting..."; exit 0 ;;
                    *) print_error "Invalid choice!"; sleep 2 ;;
                esac
            done
            ;;
        --backup)
            backup_configs
            ;;
        --restore)
            restore_backup
            ;;
        --help)
            show_usage
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
