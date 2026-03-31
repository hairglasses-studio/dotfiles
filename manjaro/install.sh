#!/usr/bin/env bash
# manjaro/install.sh — Manjaro Linux installer for dotfiles
# Counterpart to install.sh (macOS)
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="$HOME/.dotfiles-backup-$(date +%Y%m%d-%H%M%S)"
BACKUP_CREATED=false
CHECK_ONLY=false

if [[ "${1:-}" == "--check" ]]; then
    CHECK_ONLY=true
fi

# ── Logging ───────────────────────────────────────────────────
info()    { printf '\033[1;34m[INFO]\033[0m  %s\n' "$1"; }
success() { printf '\033[1;32m[OK]\033[0m    %s\n' "$1"; }
warn()    { printf '\033[1;33m[WARN]\033[0m  %s\n' "$1"; }
error()   { printf '\033[1;31m[ERR]\033[0m   %s\n' "$1"; }

# ── Backup ────────────────────────────────────────────────────
backup_file() {
    local target="$1"
    if [[ ! -d "$BACKUP_DIR" ]]; then
        mkdir -p "$BACKUP_DIR"
        BACKUP_CREATED=true
        info "Backup directory: $BACKUP_DIR"
    fi
    local rel="${target#$HOME/}"
    mkdir -p "$BACKUP_DIR/$(dirname "$rel")"
    mv "$target" "$BACKUP_DIR/$rel"
    warn "Backed up: $target"
}

# ── Core symlink function ─────────────────────────────────────
link_file() {
    local src="$1" dst="$2"

    # Already correctly linked
    if [[ -L "$dst" ]] && [[ "$(readlink "$dst")" == "$src" ]]; then
        success "Already linked: $dst"
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
    success "Linked: $dst -> $src"
}

# ── Package installation ──────────────────────────────────────
install_packages() {
    info "Updating system and installing pacman packages..."
    local pkgs
    pkgs=$(grep -v '^#' "$DOTFILES/manjaro/packages.txt" | grep -v '^$')
    sudo pacman -Syu --needed --noconfirm $pkgs

    if command -v yay &>/dev/null; then
        info "Installing AUR packages..."
        yay -S --needed --noconfirm ttf-jetbrains-mono-nerd
    else
        warn "yay not found — skipping AUR packages (ttf-jetbrains-mono-nerd)"
        warn "Install yay: https://github.com/Jguer/yay"
    fi
}

# ── Oh My Zsh ─────────────────────────────────────────────────
install_omz() {
    if [[ ! -d "$HOME/.oh-my-zsh" ]]; then
        info "Installing Oh My Zsh..."
        sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended --keep-zshrc
    else
        success "Oh My Zsh already installed"
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
            success "Plugin already installed: $name"
        else
            info "Installing plugin: $name"
            git clone --depth=1 "$url" "$dir" 2>/dev/null
        fi
    done

    # Powerlevel10k theme
    local p10k_dir="$ZSH_CUSTOM/themes/powerlevel10k"
    if [[ -d "$p10k_dir" ]]; then
        success "Theme already installed: powerlevel10k"
    else
        info "Installing theme: powerlevel10k"
        git clone --depth=1 https://github.com/romkatv/powerlevel10k.git "$p10k_dir" 2>/dev/null
    fi
}

# ── Neovim ────────────────────────────────────────────────────
install_vim_plug() {
    local plug_path="${XDG_DATA_HOME:-$HOME/.local/share}/nvim/site/autoload/plug.vim"
    if [[ -f "$plug_path" ]]; then
        success "vim-plug already installed"
    else
        info "Installing vim-plug..."
        curl -fLo "$plug_path" --create-dirs \
            https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim 2>/dev/null
    fi

    # Create nvim persistence directories
    mkdir -p "$HOME/.local/share/nvim/backup"
    mkdir -p "$HOME/.local/share/nvim/undo"
    mkdir -p "$HOME/.local/share/nvim/swap"
}

# ── Tmux Plugin Manager ──────────────────────────────────────
install_tpm() {
    local tpm_dir="$HOME/.tmux/plugins/tpm"
    if [[ -d "$tpm_dir" ]]; then
        success "TPM already installed"
    else
        info "Installing TPM..."
        git clone --depth=1 https://github.com/tmux-plugins/tpm "$tpm_dir" 2>/dev/null
    fi
}

# ── Tattoy shader symlink ────────────────────────────────────
setup_tattoy_shaders() {
    local tattoy_shaders="$HOME/.config/tattoy/shaders"
    local ghostty_shaders="$DOTFILES/ghostty/shaders"
    if [[ -L "$tattoy_shaders" ]] && [[ "$(readlink "$tattoy_shaders")" == "$ghostty_shaders" ]]; then
        success "Tattoy shaders already linked"
    elif [[ -d "$tattoy_shaders" ]]; then
        warn "Tattoy shaders dir exists, skipping symlink"
    else
        mkdir -p "$(dirname "$tattoy_shaders")"
        ln -sf "$ghostty_shaders" "$tattoy_shaders"
        success "Linked Tattoy shaders to Ghostty shader collection"
    fi
}

# ── Symlinks ──────────────────────────────────────────────────
create_symlinks() {
    info "Creating symlinks..."

    # Individual files
    link_file "$DOTFILES/zsh/zshrc"              "$HOME/.zshrc"
    link_file "$DOTFILES/zsh/p10k.zsh"           "$HOME/.p10k.zsh"
    link_file "$DOTFILES/zsh/zshenv"             "$HOME/.zshenv"
    link_file "$DOTFILES/git/gitconfig"          "$HOME/.gitconfig"
    link_file "$DOTFILES/ssh/config"             "$HOME/.ssh/config"
    link_file "$DOTFILES/starship/starship.toml" "$HOME/.config/starship.toml"

    # Directory symlinks (cross-platform)
    link_file "$DOTFILES/ghostty"    "$HOME/.config/ghostty"
    link_file "$DOTFILES/nvim"       "$HOME/.config/nvim"
    link_file "$DOTFILES/bat"        "$HOME/.config/bat"
    link_file "$DOTFILES/fastfetch"  "$HOME/.config/fastfetch"
    link_file "$DOTFILES/git/delta"  "$HOME/.config/delta"
    link_file "$DOTFILES/git/ignore" "$HOME/.config/git/ignore"
    link_file "$DOTFILES/gh"         "$HOME/.config/gh"
    link_file "$DOTFILES/k9s"        "$HOME/.config/k9s"
    link_file "$DOTFILES/lazygit"    "$HOME/.config/lazygit"
    link_file "$DOTFILES/btop"       "$HOME/.config/btop"
    link_file "$DOTFILES/yazi"       "$HOME/.config/yazi"
    link_file "$DOTFILES/cava"       "$HOME/.config/cava"
    link_file "$DOTFILES/glow"       "$HOME/.config/glow"

    # Tmux
    link_file "$DOTFILES/tmux/tmux.conf" "$HOME/.tmux.conf"

    # Tattoy (Linux XDG path, not macOS ~/Library)
    link_file "$DOTFILES/tattoy/tattoy.toml" "$HOME/.config/tattoy/tattoy.toml"

    # Sway + Waybar (if configs exist in dotfiles)
    if [[ -d "$DOTFILES/sway" ]]; then
        link_file "$DOTFILES/sway" "$HOME/.config/sway"
    fi
    if [[ -d "$DOTFILES/waybar" ]]; then
        link_file "$DOTFILES/waybar" "$HOME/.config/waybar"
    fi

    # Hyprland + tools
    if [[ -d "$DOTFILES/hyprland" ]]; then
        link_file "$DOTFILES/hyprland" "$HOME/.config/hypr"
    fi
    for dir in eww mako wofi wlogout; do
        if [[ -d "$DOTFILES/$dir" ]]; then
            link_file "$DOTFILES/$dir" "$HOME/.config/$dir"
        fi
    done
}

# ── Systemd user services ────────────────────────────────────
install_systemd_services() {
    if [[ -d "$DOTFILES/systemd" ]]; then
        info "Installing systemd user services..."
        mkdir -p "$HOME/.config/systemd/user"
        for unit in "$DOTFILES/systemd/"*.{timer,service} ; do
            [[ -f "$unit" ]] || continue
            cp "$unit" "$HOME/.config/systemd/user/"
            info "  Installed $(basename "$unit")"
        done
        systemctl --user daemon-reload

        # Enable shader-rotate timer if present
        if [[ -f "$HOME/.config/systemd/user/shader-rotate.timer" ]]; then
            systemctl --user enable --now shader-rotate.timer
            info "  shader-rotate timer enabled"
        fi
    fi
}

# ── Ghostty shaders ──────────────────────────────────────────
setup_shaders() {
    info "Setting up Ghostty shader pipeline..."
    if [[ -d "$DOTFILES/ghostty/shaders" ]]; then
        chmod +x "$DOTFILES/ghostty/shaders/bin/"*.sh 2>/dev/null || true
        info "  Shader scripts ready"
    fi
}

# ── Check mode ────────────────────────────────────────────────
check_install() {
    local errors=0

    check_link() {
        local src="$1" dst="$2"
        if [[ -L "$dst" ]] && [[ "$(readlink "$dst")" == "$src" ]]; then
            success "OK: $dst"
        elif [[ -L "$dst" ]]; then
            error "Wrong target: $dst -> $(readlink "$dst") (expected $src)"
            errors=$((errors + 1))
        elif [[ -e "$dst" ]]; then
            warn "Exists but not a symlink: $dst"
            errors=$((errors + 1))
        else
            error "Missing: $dst"
            errors=$((errors + 1))
        fi
    }

    info "Checking symlinks..."
    check_link "$DOTFILES/zsh/zshrc"              "$HOME/.zshrc"
    check_link "$DOTFILES/zsh/p10k.zsh"           "$HOME/.p10k.zsh"
    check_link "$DOTFILES/zsh/zshenv"             "$HOME/.zshenv"
    check_link "$DOTFILES/git/gitconfig"          "$HOME/.gitconfig"
    check_link "$DOTFILES/ssh/config"             "$HOME/.ssh/config"
    check_link "$DOTFILES/starship/starship.toml" "$HOME/.config/starship.toml"
    check_link "$DOTFILES/ghostty"    "$HOME/.config/ghostty"
    check_link "$DOTFILES/nvim"       "$HOME/.config/nvim"
    check_link "$DOTFILES/bat"        "$HOME/.config/bat"
    check_link "$DOTFILES/fastfetch"  "$HOME/.config/fastfetch"
    check_link "$DOTFILES/git/delta"  "$HOME/.config/delta"
    check_link "$DOTFILES/git/ignore" "$HOME/.config/git/ignore"
    check_link "$DOTFILES/gh"         "$HOME/.config/gh"
    check_link "$DOTFILES/k9s"        "$HOME/.config/k9s"
    check_link "$DOTFILES/lazygit"    "$HOME/.config/lazygit"
    check_link "$DOTFILES/btop"       "$HOME/.config/btop"
    check_link "$DOTFILES/yazi"       "$HOME/.config/yazi"
    check_link "$DOTFILES/cava"       "$HOME/.config/cava"
    check_link "$DOTFILES/glow"       "$HOME/.config/glow"
    check_link "$DOTFILES/tmux/tmux.conf" "$HOME/.tmux.conf"
    check_link "$DOTFILES/tattoy/tattoy.toml" "$HOME/.config/tattoy/tattoy.toml"

    info "Checking pacman packages..."
    local missing=0
    while IFS= read -r pkg; do
        if ! pacman -Qi "$pkg" &>/dev/null; then
            error "Missing package: $pkg"
            missing=$((missing + 1))
        fi
    done < <(grep -v '^#' "$DOTFILES/manjaro/packages.txt" | grep -v '^$')
    if [[ $missing -eq 0 ]]; then
        success "All pacman packages installed"
    fi
    errors=$((errors + missing))

    info "Checking Oh My Zsh..."
    if [[ -d "$HOME/.oh-my-zsh" ]]; then
        success "Oh My Zsh installed"
    else
        error "Oh My Zsh not found"
        errors=$((errors + 1))
    fi

    echo ""
    if [[ $errors -eq 0 ]]; then
        success "All checks passed"
    else
        error "$errors issue(s) found"
    fi
    return $errors
}

# ── Main ──────────────────────────────────────────────────────
main() {
    echo ""
    echo "  Manjaro dotfiles installer"
    echo "  ──────────────────────────"
    echo "  repo: $DOTFILES"
    echo ""

    if $CHECK_ONLY; then
        check_install
        return $?
    fi

    install_packages
    install_omz
    install_omz_plugins
    install_vim_plug
    install_tpm
    create_symlinks
    setup_tattoy_shaders
    install_systemd_services
    setup_shaders

    # Build bat cache (custom themes)
    if command -v bat &>/dev/null; then
        info "Building bat theme cache..."
        bat cache --build 2>/dev/null
    fi

    echo ""
    if $BACKUP_CREATED; then
        warn "Backups saved to: $BACKUP_DIR"
    fi
    success "Done! Restart your shell or run: source ~/.zshrc"
    echo ""
}

main "$@"
