#!/usr/bin/env bash
set -euo pipefail

# ── Dotfiles Installer ─────────────────────────
# Symlinks configs from this repo to their expected locations.
# Idempotent — safe to run multiple times.

DOTFILES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$HOME/.dotfiles-backup-$(date +%Y%m%d-%H%M%S)"
BACKUP_CREATED=false

# ── Logging ─────────────────────────────────────
log_info()    { printf "\033[0;34m[INFO]\033[0m  %s\n" "$1"; }
log_success() { printf "\033[0;32m[OK]\033[0m    %s\n" "$1"; }
log_warn()    { printf "\033[0;33m[WARN]\033[0m  %s\n" "$1"; }
log_error()   { printf "\033[0;31m[ERR]\033[0m   %s\n" "$1"; }

# ── Backup ──────────────────────────────────────
backup_file() {
    local target="$1"
    if [[ ! -d "$BACKUP_DIR" ]]; then
        mkdir -p "$BACKUP_DIR"
        BACKUP_CREATED=true
        log_info "Backup directory: $BACKUP_DIR"
    fi
    local rel="${target#$HOME/}"
    mkdir -p "$BACKUP_DIR/$(dirname "$rel")"
    mv "$target" "$BACKUP_DIR/$rel"
    log_warn "Backed up: $target"
}

# ── Core symlink function ───────────────────────
link_file() {
    local src="$1" dst="$2"

    # Already correctly linked
    if [[ -L "$dst" ]] && [[ "$(readlink "$dst")" == "$src" ]]; then
        log_success "Already linked: $dst"
        return 0
    fi

    # Backup existing file/dir/symlink
    if [[ -e "$dst" ]] || [[ -L "$dst" ]]; then
        backup_file "$dst"
    fi

    # Create parent directory
    mkdir -p "$(dirname "$dst")"

    # Create symlink
    ln -sf "$src" "$dst"
    log_success "Linked: $dst -> $src"
}

# ── Homebrew ────────────────────────────────────
install_homebrew() {
    if ! command -v brew &>/dev/null; then
        log_info "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    else
        log_success "Homebrew already installed"
    fi
}

install_brew_packages() {
    if command -v brew &>/dev/null && [[ -f "$DOTFILES_DIR/Brewfile" ]]; then
        log_info "Installing Homebrew packages from Brewfile..."
        brew bundle --file="$DOTFILES_DIR/Brewfile" --no-lock 2>&1 | while read -r line; do
            log_info "  $line"
        done
    fi
}

# ── Oh My Zsh ──────────────────────────────────
install_omz() {
    if [[ ! -d "$HOME/.oh-my-zsh" ]]; then
        log_info "Installing Oh My Zsh..."
        sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended --keep-zshrc
    else
        log_success "Oh My Zsh already installed"
    fi
}

install_omz_plugins() {
    local ZSH_CUSTOM="${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}"

    declare -A plugins=(
        ["fast-syntax-highlighting"]="https://github.com/zdharma-continuum/fast-syntax-highlighting.git"
        ["zsh-autosuggestions"]="https://github.com/zsh-users/zsh-autosuggestions.git"
        ["zsh-completions"]="https://github.com/zsh-users/zsh-completions.git"
        ["you-should-use"]="https://github.com/MichaelAquilina/zsh-you-should-use.git"
        ["fzf-tab"]="https://github.com/Aloxaf/fzf-tab.git"
    )

    for name in "${!plugins[@]}"; do
        local dir="$ZSH_CUSTOM/plugins/$name"
        if [[ -d "$dir" ]]; then
            log_success "Plugin already installed: $name"
        else
            log_info "Installing plugin: $name"
            git clone --depth=1 "${plugins[$name]}" "$dir" 2>/dev/null
        fi
    done

    # Powerlevel10k theme
    local p10k_dir="$ZSH_CUSTOM/themes/powerlevel10k"
    if [[ -d "$p10k_dir" ]]; then
        log_success "Theme already installed: powerlevel10k"
    else
        log_info "Installing theme: powerlevel10k"
        git clone --depth=1 https://github.com/romkatv/powerlevel10k.git "$p10k_dir" 2>/dev/null
    fi
}

# ── Neovim ──────────────────────────────────────
install_vim_plug() {
    local plug_path="${XDG_DATA_HOME:-$HOME/.local/share}/nvim/site/autoload/plug.vim"
    if [[ -f "$plug_path" ]]; then
        log_success "vim-plug already installed"
    else
        log_info "Installing vim-plug..."
        curl -fLo "$plug_path" --create-dirs \
            https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim 2>/dev/null
    fi

    # Create nvim persistence directories
    mkdir -p "$HOME/.local/share/nvim/backup"
    mkdir -p "$HOME/.local/share/nvim/undo"
    mkdir -p "$HOME/.local/share/nvim/swap"
}

# ── Create symlinks ────────────────────────────
create_symlinks() {
    log_info "Creating symlinks..."

    # Individual files
    link_file "$DOTFILES_DIR/zsh/zshrc"              "$HOME/.zshrc"
    link_file "$DOTFILES_DIR/zsh/p10k.zsh"           "$HOME/.p10k.zsh"
    link_file "$DOTFILES_DIR/zsh/zshenv"             "$HOME/.zshenv"
    link_file "$DOTFILES_DIR/git/gitconfig"          "$HOME/.gitconfig"
    link_file "$DOTFILES_DIR/ssh/config"             "$HOME/.ssh/config"
    link_file "$DOTFILES_DIR/starship/starship.toml" "$HOME/.config/starship.toml"

    # Directory symlinks
    link_file "$DOTFILES_DIR/ghostty"    "$HOME/.config/ghostty"
    link_file "$DOTFILES_DIR/nvim"       "$HOME/.config/nvim"
    link_file "$DOTFILES_DIR/bat"        "$HOME/.config/bat"
    link_file "$DOTFILES_DIR/fastfetch"  "$HOME/.config/fastfetch"
    link_file "$DOTFILES_DIR/git/delta"  "$HOME/.config/delta"
    link_file "$DOTFILES_DIR/git/ignore" "$HOME/.config/git/ignore"
    link_file "$DOTFILES_DIR/gh"         "$HOME/.config/gh"
    link_file "$DOTFILES_DIR/k9s"        "$HOME/.config/k9s"
}

# ── Main ────────────────────────────────────────
main() {
    echo ""
    echo "  dotfiles installer"
    echo "  ──────────────────"
    echo "  repo: $DOTFILES_DIR"
    echo ""

    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_homebrew
        install_brew_packages
    fi

    install_omz
    install_omz_plugins
    install_vim_plug
    create_symlinks

    echo ""
    if $BACKUP_CREATED; then
        log_warn "Backups saved to: $BACKUP_DIR"
    fi
    log_success "Done! Restart your shell or run: source ~/.zshrc"
    echo ""
}

main "$@"
