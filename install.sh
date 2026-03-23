#!/usr/bin/env bash
set -euo pipefail

# ── Dotfiles Installer ─────────────────────────
# Symlinks configs from this repo to their expected locations.
# Idempotent — safe to run multiple times.

DOTFILES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$HOME/.dotfiles-backup-$(date +%Y%m%d-%H%M%S)"
BACKUP_CREATED=false
CHECK_ONLY=false

if [[ "${1:-}" == "--check" ]]; then
    CHECK_ONLY=true
fi

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
        brew bundle --file="$DOTFILES_DIR/Brewfile" 2>&1 | while read -r line; do
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

    local plugin_list=(
        "fast-syntax-highlighting|https://github.com/zdharma-continuum/fast-syntax-highlighting.git"
        "zsh-autosuggestions|https://github.com/zsh-users/zsh-autosuggestions.git"
        "zsh-completions|https://github.com/zsh-users/zsh-completions.git"
        "you-should-use|https://github.com/MichaelAquilina/zsh-you-should-use.git"
        "fzf-tab|https://github.com/Aloxaf/fzf-tab.git"
    )

    for entry in "${plugin_list[@]}"; do
        local name="${entry%%|*}"
        local url="${entry##*|}"
        local dir="$ZSH_CUSTOM/plugins/$name"
        if [[ -d "$dir" ]]; then
            log_success "Plugin already installed: $name"
        else
            log_info "Installing plugin: $name"
            git clone --depth=1 "$url" "$dir" 2>/dev/null
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

# ── RetroVisor (CRT overlay for macOS) ────────
install_retrovisor() {
    local app_path="/Applications/RetroVisor.app"
    if [[ -d "$app_path" ]]; then
        log_success "RetroVisor already installed"
        return 0
    fi
    log_info "Installing RetroVisor (CRT shader overlay for macOS)..."
    local latest_url
    latest_url=$(curl -s https://api.github.com/repos/dirkwhoffmann/RetroVisor/releases/latest \
        | grep "browser_download_url.*dmg" | head -1 | cut -d '"' -f 4)
    if [[ -z "$latest_url" ]]; then
        log_warn "Could not find RetroVisor release — install manually from https://github.com/dirkwhoffmann/RetroVisor/releases"
        return 0
    fi
    local dmg_path="/tmp/RetroVisor.dmg"
    curl -fsSL "$latest_url" -o "$dmg_path"
    local mount_point
    mount_point=$(hdiutil attach "$dmg_path" -nobrowse -quiet | tail -1 | awk '{print $3}')
    cp -R "$mount_point/RetroVisor.app" /Applications/ 2>/dev/null || true
    hdiutil detach "$mount_point" -quiet 2>/dev/null || true
    rm -f "$dmg_path"
    if [[ -d "$app_path" ]]; then
        log_success "RetroVisor installed to /Applications"
    else
        log_warn "RetroVisor install may have failed — install manually from https://github.com/dirkwhoffmann/RetroVisor/releases"
    fi
}

# ── Tmux Plugin Manager ──────────────────────
install_tpm() {
    local tpm_dir="$HOME/.tmux/plugins/tpm"
    if [[ -d "$tpm_dir" ]]; then
        log_success "TPM already installed"
    else
        log_info "Installing TPM..."
        git clone --depth=1 https://github.com/tmux-plugins/tpm "$tpm_dir" 2>/dev/null
    fi
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
    link_file "$DOTFILES_DIR/lazygit"   "$HOME/.config/lazygit"

    # Desktop rice
    link_file "$DOTFILES_DIR/aerospace/aerospace.toml" "$HOME/.aerospace.toml"
    link_file "$DOTFILES_DIR/sketchybar"  "$HOME/.config/sketchybar"
    link_file "$DOTFILES_DIR/borders"     "$HOME/.config/borders"
    link_file "$DOTFILES_DIR/btop"        "$HOME/.config/btop"
    link_file "$DOTFILES_DIR/yazi"        "$HOME/.config/yazi"
    link_file "$DOTFILES_DIR/cava"        "$HOME/.config/cava"
    link_file "$DOTFILES_DIR/glow"        "$HOME/.config/glow"

    # Individual file symlinks (non-XDG)
    link_file "$DOTFILES_DIR/tmux/tmux.conf" "$HOME/.tmux.conf"
}

# ── Check mode ─────────────────────────────────
check_symlinks() {
    local errors=0

    check_link() {
        local src="$1" dst="$2"
        if [[ -L "$dst" ]] && [[ "$(readlink "$dst")" == "$src" ]]; then
            log_success "OK: $dst"
        elif [[ -L "$dst" ]]; then
            log_error "Wrong target: $dst -> $(readlink "$dst") (expected $src)"
            errors=$((errors + 1))
        elif [[ -e "$dst" ]]; then
            log_warn "Exists but not a symlink: $dst"
            errors=$((errors + 1))
        else
            log_error "Missing: $dst"
            errors=$((errors + 1))
        fi
    }

    log_info "Checking symlinks..."
    check_link "$DOTFILES_DIR/zsh/zshrc"              "$HOME/.zshrc"
    check_link "$DOTFILES_DIR/zsh/p10k.zsh"           "$HOME/.p10k.zsh"
    check_link "$DOTFILES_DIR/zsh/zshenv"             "$HOME/.zshenv"
    check_link "$DOTFILES_DIR/git/gitconfig"          "$HOME/.gitconfig"
    check_link "$DOTFILES_DIR/ssh/config"             "$HOME/.ssh/config"
    check_link "$DOTFILES_DIR/starship/starship.toml" "$HOME/.config/starship.toml"
    check_link "$DOTFILES_DIR/ghostty"    "$HOME/.config/ghostty"
    check_link "$DOTFILES_DIR/nvim"       "$HOME/.config/nvim"
    check_link "$DOTFILES_DIR/bat"        "$HOME/.config/bat"
    check_link "$DOTFILES_DIR/fastfetch"  "$HOME/.config/fastfetch"
    check_link "$DOTFILES_DIR/git/delta"  "$HOME/.config/delta"
    check_link "$DOTFILES_DIR/git/ignore" "$HOME/.config/git/ignore"
    check_link "$DOTFILES_DIR/gh"         "$HOME/.config/gh"
    check_link "$DOTFILES_DIR/k9s"        "$HOME/.config/k9s"
    check_link "$DOTFILES_DIR/lazygit"    "$HOME/.config/lazygit"
    check_link "$DOTFILES_DIR/aerospace/aerospace.toml" "$HOME/.aerospace.toml"
    check_link "$DOTFILES_DIR/sketchybar"  "$HOME/.config/sketchybar"
    check_link "$DOTFILES_DIR/borders"     "$HOME/.config/borders"
    check_link "$DOTFILES_DIR/btop"        "$HOME/.config/btop"
    check_link "$DOTFILES_DIR/yazi"        "$HOME/.config/yazi"
    check_link "$DOTFILES_DIR/cava"        "$HOME/.config/cava"
    check_link "$DOTFILES_DIR/glow"        "$HOME/.config/glow"
    check_link "$DOTFILES_DIR/tmux/tmux.conf" "$HOME/.tmux.conf"

    log_info "Checking brew packages..."
    if command -v brew &>/dev/null; then
        local missing=0
        while IFS= read -r pkg; do
            pkg="$(echo "$pkg" | sed 's/^brew "//;s/"$//')"
            if ! brew list --formula "$pkg" &>/dev/null; then
                log_error "Missing brew package: $pkg"
                missing=$((missing + 1))
            fi
        done < <(grep '^brew ' "$DOTFILES_DIR/Brewfile")
        if [[ $missing -eq 0 ]]; then
            log_success "All brew packages installed"
        fi
        errors=$((errors + missing))
    else
        log_warn "Homebrew not installed"
        errors=$((errors + 1))
    fi

    log_info "Checking Oh My Zsh..."
    if [[ -d "$HOME/.oh-my-zsh" ]]; then
        log_success "Oh My Zsh installed"
    else
        log_error "Oh My Zsh not found"
        errors=$((errors + 1))
    fi

    echo ""
    if [[ $errors -eq 0 ]]; then
        log_success "All checks passed"
    else
        log_error "$errors issue(s) found"
    fi
    return $errors
}

# ── Main ────────────────────────────────────────
main() {
    echo ""
    echo "  dotfiles installer"
    echo "  ──────────────────"
    echo "  repo: $DOTFILES_DIR"
    echo ""

    if $CHECK_ONLY; then
        check_symlinks
        return $?
    fi

    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_homebrew
        install_brew_packages
    fi

    install_omz
    install_omz_plugins
    install_vim_plug
    install_tpm
    install_retrovisor
    create_symlinks

    # Build bat cache (custom themes)
    if command -v bat &>/dev/null; then
        log_info "Building bat theme cache..."
        bat cache --build 2>/dev/null
    fi

    echo ""
    if $BACKUP_CREATED; then
        log_warn "Backups saved to: $BACKUP_DIR"
    fi
    log_success "Done! Restart your shell or run: source ~/.zshrc"
    echo ""
}

main "$@"
