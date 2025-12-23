#!/usr/bin/env bash
# AFTRS Shell Wrapper Setup Script
# Creates symlinks to dotfiles and sets up the shell environment with submodules

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AFTRS_SHELL_HOME="${AFTRS_SHELL_HOME:-$HOME/.aftrs-shell}"

# Functions
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running from the correct directory
check_installation_directory() {
    if [[ ! -f "$SCRIPT_DIR/.gitmodules" ]] || [[ ! -d "$SCRIPT_DIR/submodules" ]]; then
        print_error "This script must be run from the aftrs-shell repository root directory."
        print_error "Make sure you have cloned the repository with submodules:"
        echo "  git clone --recursive <repository-url>"
        exit 1
    fi
}

# Initialize git submodules
init_submodules() {
    print_status "Initializing git submodules..."
    cd "$SCRIPT_DIR"
    
    if git submodule status | grep -q '^-'; then
        print_status "Updating submodules..."
        git submodule update --init --recursive
    else
        print_success "Submodules already initialized"
    fi
}

# Create AFTRS_SHELL_HOME directory if it doesn't exist
setup_shell_home() {
    if [[ "$SCRIPT_DIR" != "$AFTRS_SHELL_HOME" ]]; then
        print_status "Setting up shell home directory at $AFTRS_SHELL_HOME"
        
        if [[ -e "$AFTRS_SHELL_HOME" && ! -L "$AFTRS_SHELL_HOME" ]]; then
            print_warning "Directory $AFTRS_SHELL_HOME already exists"
            read -p "Remove existing directory? [y/N]: " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                rm -rf "$AFTRS_SHELL_HOME"
            else
                print_error "Installation cancelled"
                exit 1
            fi
        fi
        
        # Create symlink to the repository
        ln -sfn "$SCRIPT_DIR" "$AFTRS_SHELL_HOME"
        print_success "Created symlink: $AFTRS_SHELL_HOME -> $SCRIPT_DIR"
    else
        print_success "Already running from shell home directory"
    fi
}

# Create backup of existing files
backup_existing_files() {
    local files=(".zshrc" ".gitconfig" ".vimrc")
    local backup_dir="$HOME/.aftrs-shell-backup-$(date +%Y%m%d-%H%M%S)"
    local backup_created=false
    
    for file in "${files[@]}"; do
        if [[ -f "$HOME/$file" && ! -L "$HOME/$file" ]]; then
            if [[ "$backup_created" == "false" ]]; then
                mkdir -p "$backup_dir"
                print_status "Creating backup directory: $backup_dir"
                backup_created=true
            fi
            
            mv "$HOME/$file" "$backup_dir/"
            print_status "Backed up $file to $backup_dir/"
        fi
    done
    
    if [[ "$backup_created" == "true" ]]; then
        print_success "Existing files backed up to $backup_dir"
    fi
}

# Create symlinks to dotfiles
create_symlinks() {
    print_status "Creating symlinks to dotfiles..."
    
    # Define files to symlink
    declare -A dotfiles=(
        ["$AFTRS_SHELL_HOME/dotfiles/.zshrc"]="$HOME/.zshrc"
    )
    
    # Add additional dotfiles if they exist
    if [[ -f "$AFTRS_SHELL_HOME/dotfiles/.gitconfig" ]]; then
        dotfiles["$AFTRS_SHELL_HOME/dotfiles/.gitconfig"]="$HOME/.gitconfig"
    fi
    
    if [[ -f "$AFTRS_SHELL_HOME/dotfiles/.vimrc" ]]; then
        dotfiles["$AFTRS_SHELL_HOME/dotfiles/.vimrc"]="$HOME/.vimrc"
    fi
    
    # Create symlinks
    for source in "${!dotfiles[@]}"; do
        target="${dotfiles[$source]}"
        
        if [[ -L "$target" ]]; then
            rm "$target"
        fi
        
        ln -s "$source" "$target"
        print_success "Created symlink: $target -> $source"
    done
}

# Set zsh as default shell if not already
set_default_shell() {
    if [[ "$SHELL" != "$(which zsh)" ]]; then
        print_status "Setting zsh as default shell..."
        
        # Add zsh to /etc/shells if not present
        if ! grep -q "$(which zsh)" /etc/shells 2>/dev/null; then
            print_status "Adding zsh to /etc/shells (requires sudo)"
            echo "$(which zsh)" | sudo tee -a /etc/shells
        fi
        
        # Change default shell
        print_status "Changing default shell to zsh (requires sudo)"
        sudo chsh -s "$(which zsh)" "$USER"
        print_success "Default shell changed to zsh"
        print_warning "Please log out and log back in for the change to take effect"
    else
        print_success "Zsh is already the default shell"
    fi
}

# Generate initial antidote cache
generate_antidote_cache() {
    if [[ -f "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins.txt" ]]; then
        print_status "Generating antidote plugin cache..."
        # Source the new zshrc to generate cache
        AFTRS_SHELL_HOME="$AFTRS_SHELL_HOME" zsh -c "source $HOME/.zshrc && echo 'Cache generated successfully'" 2>/dev/null || true
        print_success "Plugin cache generated"
    fi
}

# Main installation function
main() {
    print_status "Starting AFTRS Shell Wrapper setup..."
    
    # Check prerequisites
    if ! command -v zsh >/dev/null 2>&1; then
        print_error "zsh is not installed. Please install zsh first."
        exit 1
    fi
    
    if ! command -v git >/dev/null 2>&1; then
        print_error "git is not installed. Please install git first."
        exit 1
    fi
    
    # Run installation steps
    check_installation_directory
    init_submodules
    setup_shell_home
    backup_existing_files
    create_symlinks
    set_default_shell
    generate_antidote_cache
    
    print_success "AFTRS Shell Wrapper setup completed successfully!"
    print_status "To start using the new shell configuration:"
    echo "  1. Start a new terminal session, or"
    echo "  2. Run: source ~/.zshrc"
    echo ""
    print_status "The first run may take a moment to download and install plugins."
    echo ""
    print_status "Plugin configuration can be customized by editing:"
    echo "  $AFTRS_SHELL_HOME/dotfiles/zsh_plugins.txt"
}

# Run main function
main "$@"
