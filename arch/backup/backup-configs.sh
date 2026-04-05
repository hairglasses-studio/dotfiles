#!/bin/bash

#############################################
# Backup Configurations Script
# Backs up system and user configurations
#############################################

set -e

# Configuration
BACKUP_DIR="$HOME/arch-backup-$(date +%Y%m%d-%H%M%S)"
REPO_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

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

# Create backup directory
create_backup_dir() {
    mkdir -p "$BACKUP_DIR"
    print_success "Created backup directory: $BACKUP_DIR"
}

# Backup system files
backup_system() {
    print_status "Backing up system configurations..."
    
    mkdir -p "$BACKUP_DIR/system"
    
    # Pacman
    if [ -f /etc/pacman.conf ]; then
        cp /etc/pacman.conf "$BACKUP_DIR/system/"
        print_success "Backed up pacman.conf"
    fi
    
    # Mirrorlist
    if [ -f /etc/pacman.d/mirrorlist ]; then
        cp /etc/pacman.d/mirrorlist "$BACKUP_DIR/system/"
        print_success "Backed up mirrorlist"
    fi
    
    # fstab
    if [ -f /etc/fstab ]; then
        cp /etc/fstab "$BACKUP_DIR/system/"
        print_success "Backed up fstab"
    fi
    
    # GRUB
    if [ -f /etc/default/grub ]; then
        cp /etc/default/grub "$BACKUP_DIR/system/"
        print_success "Backed up GRUB config"
    fi
    
    # Hostname
    if [ -f /etc/hostname ]; then
        cp /etc/hostname "$BACKUP_DIR/system/"
        print_success "Backed up hostname"
    fi
    
    # Hosts
    if [ -f /etc/hosts ]; then
        cp /etc/hosts "$BACKUP_DIR/system/"
        print_success "Backed up hosts"
    fi
}

# Backup package lists
backup_packages() {
    print_status "Backing up package lists..."
    
    mkdir -p "$BACKUP_DIR/packages"
    
    # Official packages
    pacman -Qqen > "$BACKUP_DIR/packages/official-packages.txt"
    print_success "Backed up official package list"
    
    # AUR packages
    pacman -Qqem > "$BACKUP_DIR/packages/aur-packages.txt"
    print_success "Backed up AUR package list"
    
    # All packages with versions
    pacman -Q > "$BACKUP_DIR/packages/all-packages-versions.txt"
    print_success "Backed up all packages with versions"
    
    # Explicitly installed
    pacman -Qqe > "$BACKUP_DIR/packages/explicitly-installed.txt"
    print_success "Backed up explicitly installed packages"
}

# Backup user configurations
backup_user_configs() {
    print_status "Backing up user configurations..."
    
    mkdir -p "$BACKUP_DIR/configs"
    
    # Shell configs
    [ -f ~/.bashrc ] && cp ~/.bashrc "$BACKUP_DIR/configs/"
    [ -f ~/.zshrc ] && cp ~/.zshrc "$BACKUP_DIR/configs/"
    [ -f ~/.profile ] && cp ~/.profile "$BACKUP_DIR/configs/"
    
    # Git
    [ -f ~/.gitconfig ] && cp ~/.gitconfig "$BACKUP_DIR/configs/"
    [ -f ~/.gitignore_global ] && cp ~/.gitignore_global "$BACKUP_DIR/configs/"
    
    # SSH (without private keys)
    if [ -d ~/.ssh ]; then
        mkdir -p "$BACKUP_DIR/configs/ssh"
        [ -f ~/.ssh/config ] && cp ~/.ssh/config "$BACKUP_DIR/configs/ssh/"
        [ -f ~/.ssh/known_hosts ] && cp ~/.ssh/known_hosts "$BACKUP_DIR/configs/ssh/"
        [ -f ~/.ssh/authorized_keys ] && cp ~/.ssh/authorized_keys "$BACKUP_DIR/configs/ssh/"
        cp ~/.ssh/*.pub "$BACKUP_DIR/configs/ssh/" 2>/dev/null || true
    fi
    
    print_success "Backed up user configurations"
}

# Backup application configs
backup_app_configs() {
    print_status "Backing up application configurations..."
    
    mkdir -p "$BACKUP_DIR/app-configs"
    
    # VS Code
    if [ -d ~/.config/Code ]; then
        mkdir -p "$BACKUP_DIR/app-configs/vscode"
        [ -f ~/.config/Code/User/settings.json ] && \
            cp ~/.config/Code/User/settings.json "$BACKUP_DIR/app-configs/vscode/"
        [ -f ~/.config/Code/User/keybindings.json ] && \
            cp ~/.config/Code/User/keybindings.json "$BACKUP_DIR/app-configs/vscode/"
        code --list-extensions > "$BACKUP_DIR/app-configs/vscode/extensions.txt" 2>/dev/null || true
    fi
    
    # Terminal configs
    [ -d ~/.config/kitty ] && cp -r ~/.config/kitty "$BACKUP_DIR/app-configs/"
    [ -d ~/.config/alacritty ] && cp -r ~/.config/alacritty "$BACKUP_DIR/app-configs/"
    [ -d ~/.config/terminator ] && cp -r ~/.config/terminator "$BACKUP_DIR/app-configs/"
    
    print_success "Backed up application configurations"
}

# Backup desktop environment
backup_desktop() {
    print_status "Backing up desktop environment settings..."
    
    mkdir -p "$BACKUP_DIR/desktop"
    
    # Detect and backup DE settings
    if [ "$XDG_CURRENT_DESKTOP" = "GNOME" ] || [ "$DESKTOP_SESSION" = "gnome" ]; then
        dconf dump / > "$BACKUP_DIR/desktop/gnome-settings.dconf"
        print_success "Backed up GNOME settings"
    fi
    
    if [ "$XDG_CURRENT_DESKTOP" = "KDE" ] || [ "$DESKTOP_SESSION" = "plasma" ]; then
        [ -d ~/.config/kde* ] && cp -r ~/.config/kde* "$BACKUP_DIR/desktop/"
        [ -d ~/.config/plasma* ] && cp -r ~/.config/plasma* "$BACKUP_DIR/desktop/"
        print_success "Backed up KDE settings"
    fi
    
    if [ "$XDG_CURRENT_DESKTOP" = "XFCE" ]; then
        [ -d ~/.config/xfce4 ] && cp -r ~/.config/xfce4 "$BACKUP_DIR/desktop/"
        print_success "Backed up XFCE settings"
    fi
}

# Create restore script
create_restore_script() {
    print_status "Creating restore script..."
    
    cat > "$BACKUP_DIR/restore.sh" << 'RESTORE_SCRIPT'
#!/bin/bash

# Arch Linux Configuration Restore Script

set -e

BACKUP_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Arch Linux Configuration Restore"
echo "================================="
echo ""
echo "This will restore configurations from: $BACKUP_DIR"
echo ""
read -p "Continue? (y/N): " -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

# Restore packages
if [ -f "$BACKUP_DIR/packages/official-packages.txt" ]; then
    echo "Installing official packages..."
    sudo pacman -S --needed $(cat "$BACKUP_DIR/packages/official-packages.txt")
fi

if [ -f "$BACKUP_DIR/packages/aur-packages.txt" ] && command -v yay &> /dev/null; then
    echo "Installing AUR packages..."
    yay -S --needed $(cat "$BACKUP_DIR/packages/aur-packages.txt")
fi

# Restore user configs
echo "Restoring user configurations..."
[ -f "$BACKUP_DIR/configs/bashrc" ] && cp "$BACKUP_DIR/configs/bashrc" ~/.bashrc
[ -f "$BACKUP_DIR/configs/zshrc" ] && cp "$BACKUP_DIR/configs/zshrc" ~/.zshrc
[ -f "$BACKUP_DIR/configs/gitconfig" ] && cp "$BACKUP_DIR/configs/gitconfig" ~/.gitconfig

# Restore VS Code extensions
if [ -f "$BACKUP_DIR/app-configs/vscode/extensions.txt" ] && command -v code &> /dev/null; then
    echo "Installing VS Code extensions..."
    while read extension; do
        code --install-extension "$extension"
    done < "$BACKUP_DIR/app-configs/vscode/extensions.txt"
fi

# Restore GNOME settings
if [ -f "$BACKUP_DIR/desktop/gnome-settings.dconf" ] && command -v dconf &> /dev/null; then
    echo "Restoring GNOME settings..."
    dconf load / < "$BACKUP_DIR/desktop/gnome-settings.dconf"
fi

echo "Restore completed!"
echo "Please review system configurations in $BACKUP_DIR/system/ manually"
RESTORE_SCRIPT
    
    chmod +x "$BACKUP_DIR/restore.sh"
    print_success "Created restore script"
}

# Create archive
create_archive() {
    print_status "Creating backup archive..."
    
    cd "$(dirname "$BACKUP_DIR")"
    tar -czf "$(basename "$BACKUP_DIR").tar.gz" "$(basename "$BACKUP_DIR")"
    
    print_success "Created archive: $(basename "$BACKUP_DIR").tar.gz"
    print_info "Archive location: $(dirname "$BACKUP_DIR")/$(basename "$BACKUP_DIR").tar.gz"
}

# Main execution
main() {
    clear
    echo "Arch Linux Configuration Backup"
    echo "================================"
    echo ""
    
    create_backup_dir
    backup_system
    backup_packages
    backup_user_configs
    backup_app_configs
    backup_desktop
    create_restore_script
    
    echo ""
    echo -n "Create compressed archive? (Y/n): "
    read -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        create_archive
    fi
    
    echo ""
    print_success "Backup completed successfully!"
    print_info "Backup location: $BACKUP_DIR"
    print_info "To restore, run: $BACKUP_DIR/restore.sh"
}

# Run main function
main
