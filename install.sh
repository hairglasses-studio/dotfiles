#!/usr/bin/env bash
set -euo pipefail

# ── Dotfiles Installer ─────────────────────────
# Symlinks configs from this repo to their expected locations.
# Idempotent — safe to run multiple times.

DOTFILES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OS="$(uname -s)"
BACKUP_DIR="$HOME/.dotfiles-backup-$(date +%Y%m%d-%H%M%S)"
BACKUP_CREATED=false
CHECK_ONLY=false

if [[ "${1:-}" == "--check" ]]; then
    CHECK_ONLY=true
fi

# ── Logging ─────────────────────────────────────
_has_tte() { command -v tte &>/dev/null && [[ -t 1 ]]; }
log_info()    { printf "\033[38;2;87;199;255m[INFO]\033[0m  %s\n" "$1"; }
log_success() { printf "\033[38;2;90;247;142m[OK]\033[0m    %s\n" "$1"; }
log_warn()    { printf "\033[38;2;243;249;157m[WARN]\033[0m  %s\n" "$1"; }
log_error()   { printf "\033[38;2;255;92;87m[ERR]\033[0m   %s\n" "$1"; }
log_phase() {
  if _has_tte; then
    echo "$1" | tte decrypt --typing-speed 6 --ciphertext-colors 57c7ff ff6ac1 2>/dev/null
  else
    printf "\n\033[38;2;255;106;193m── %s ──\033[0m\n\n" "$1"
  fi
}

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

copy_mutable_file() {
    local src="$1" dst="$2"

    # Leave existing live config alone; apps may mutate it locally.
    if [[ -f "$dst" ]]; then
        log_success "Already present: $dst"
        return 0
    fi

    if [[ -e "$dst" ]] || [[ -L "$dst" ]]; then
        backup_file "$dst"
    fi

    mkdir -p "$(dirname "$dst")"
    command cp -f "$src" "$dst"
    chmod 600 "$dst"
    log_success "Seeded copy: $dst"
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

# ── Linux Package Management (yay/paru/metapac) ─
install_yay() {
    if command -v yay &>/dev/null; then
        log_success "yay already installed"
        return 0
    fi
    log_info "Installing yay AUR helper..."
    if ! command -v pacman &>/dev/null; then
        log_error "pacman not found — not an Arch-based system?"
        return 1
    fi
    sudo pacman -S --needed --noconfirm base-devel git
    local tmp_dir
    tmp_dir="$(mktemp -d)"
    git clone https://aur.archlinux.org/yay.git "$tmp_dir/yay"
    (cd "$tmp_dir/yay" && makepkg -si --noconfirm)
    rm -rf "$tmp_dir"
    if command -v yay &>/dev/null; then
        log_success "yay installed"
    else
        log_error "yay installation failed"
        return 1
    fi
}

install_paru() {
    if command -v paru &>/dev/null; then
        log_success "paru already installed"
        return 0
    fi
    log_info "Installing paru AUR helper..."
    local helper="yay"
    command -v yay &>/dev/null || helper="sudo pacman"
    $helper -S --needed --noconfirm paru 2>&1 | while read -r line; do
        log_info "  $line"
    done
    if command -v paru &>/dev/null; then
        log_success "paru installed"
    else
        log_error "paru installation failed"
        return 1
    fi
}

install_metapac() {
    if command -v metapac &>/dev/null; then
        log_success "metapac already installed"
        return 0
    fi
    log_info "Installing metapac..."
    local helper="paru"
    command -v paru &>/dev/null || helper="yay"
    $helper -S --needed --noconfirm metapac 2>&1 | while read -r line; do
        log_info "  $line"
    done
    if command -v metapac &>/dev/null; then
        log_success "metapac installed"
    else
        log_error "metapac installation failed"
        return 1
    fi
}

install_linux_packages() {
    # Prefer metapac if installed (declarative, multi-backend)
    if command -v metapac &>/dev/null; then
        log_info "Syncing packages via metapac..."
        metapac sync --no-confirm 2>&1 | while read -r line; do
            log_info "  $line"
        done
        return 0
    fi

    # Fallback to Pacfile for bootstrap (metapac not yet installed)
    local pacfile="$DOTFILES_DIR/Pacfile"
    if [[ ! -f "$pacfile" ]]; then
        log_warn "Neither metapac nor Pacfile available — skipping packages"
        return 0
    fi
    log_info "metapac not found — falling back to Pacfile..."
    grep -v '^\s*#' "$pacfile" | grep -v '^\s*$' | yay -S --needed --noconfirm - 2>&1 | while read -r line; do
        log_info "  $line"
    done
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
    # Release may be .dmg or .zip — try both
    latest_url=$(curl -s https://api.github.com/repos/dirkwhoffmann/RetroVisor/releases/latest \
        | grep "browser_download_url" | grep -E '\.(dmg|zip)"' | head -1 | cut -d '"' -f 4)
    if [[ -z "$latest_url" ]]; then
        log_warn "Could not find RetroVisor release — install manually from https://github.com/dirkwhoffmann/RetroVisor/releases"
        return 0
    fi
    local ext="${latest_url##*.}"
    local tmp_path="/tmp/RetroVisor.$ext"
    curl -fsSL "$latest_url" -o "$tmp_path"
    if [[ "$ext" == "dmg" ]]; then
        local mount_point
        mount_point=$(hdiutil attach "$tmp_path" -nobrowse -quiet | tail -1 | awk '{print $3}')
        cp -R "$mount_point/RetroVisor.app" /Applications/ 2>/dev/null || true
        hdiutil detach "$mount_point" -quiet 2>/dev/null || true
    elif [[ "$ext" == "zip" ]]; then
        unzip -q -o "$tmp_path" -d /tmp/RetroVisor_extract 2>/dev/null
        cp -R /tmp/RetroVisor_extract/RetroVisor.app /Applications/ 2>/dev/null || true
        rm -rf /tmp/RetroVisor_extract
    fi
    rm -f "$tmp_path"
    if [[ -d "$app_path" ]]; then
        log_success "RetroVisor installed to /Applications"
    else
        log_warn "RetroVisor install may have failed — install manually from https://github.com/dirkwhoffmann/RetroVisor/releases"
    fi
}

# ── Tattoy shader symlink ────────────────────
setup_tattoy_shaders() {
    local tattoy_shaders
    if [[ "$OS" == "Darwin" ]]; then
        tattoy_shaders="$HOME/Library/Application Support/tattoy/shaders"
    else
        tattoy_shaders="${XDG_CONFIG_HOME:-$HOME/.config}/tattoy/shaders"
    fi
    local ghostty_shaders="$DOTFILES_DIR/ghostty/shaders"
    if [[ -L "$tattoy_shaders" ]] && [[ "$(readlink "$tattoy_shaders")" == "$ghostty_shaders" ]]; then
        log_success "Tattoy shaders already linked"
    elif [[ -d "$tattoy_shaders" ]]; then
        log_warn "Tattoy shaders dir exists, skipping symlink"
    else
        mkdir -p "$(dirname "$tattoy_shaders")"
        ln -sf "$ghostty_shaders" "$tattoy_shaders"
        log_success "Linked Tattoy shaders to Ghostty shader collection"
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
    link_file "$DOTFILES_DIR/zsh/profile"            "$HOME/.profile"
    link_file "$DOTFILES_DIR/git/gitconfig"          "$HOME/.gitconfig"
    link_file "$DOTFILES_DIR/ssh/config"             "$HOME/.ssh/config"
    link_file "$DOTFILES_DIR/starship/starship.toml" "$HOME/.config/starship.toml"
    copy_mutable_file "$DOTFILES_DIR/codex/config.toml" "$HOME/.codex/config.toml"

    # Directory symlinks
    link_file "$DOTFILES_DIR/kitty"      "$HOME/.config/kitty"
    # link_file "$DOTFILES_DIR/ghostty"    "$HOME/.config/ghostty"  # kept for shader pipeline
    link_file "$DOTFILES_DIR/nvim"       "$HOME/.config/nvim"
    link_file "$DOTFILES_DIR/bat"        "$HOME/.config/bat"
    link_file "$DOTFILES_DIR/fastfetch"  "$HOME/.config/fastfetch"
    link_file "$DOTFILES_DIR/git/delta"  "$HOME/.config/delta"
    link_file "$DOTFILES_DIR/git/ignore" "$HOME/.config/git/ignore"
    link_file "$DOTFILES_DIR/gh"         "$HOME/.config/gh"
    link_file "$DOTFILES_DIR/k9s"        "$HOME/.config/k9s"
    link_file "$DOTFILES_DIR/lazygit"   "$HOME/.config/lazygit"

    # Desktop rice (platform-specific)
    if [[ "$OS" == "Darwin" ]]; then
        link_file "$DOTFILES_DIR/aerospace/aerospace.toml" "$HOME/.aerospace.toml"
        link_file "$DOTFILES_DIR/sketchybar"  "$HOME/.config/sketchybar"
        link_file "$DOTFILES_DIR/borders"     "$HOME/.config/borders"
    elif [[ "$OS" == "Linux" ]]; then
        # waybar quarantined — eww is primary bar
        # mako quarantined — swaync is primary notification daemon
        link_file "$DOTFILES_DIR/swaync/config.json" "$HOME/.config/swaync/config.json"
        link_file "$DOTFILES_DIR/swaync/style.css" "$HOME/.config/swaync/style.css"
        link_file "$DOTFILES_DIR/wofi/config" "$HOME/.config/wofi/config"
        link_file "$DOTFILES_DIR/wofi/style.css" "$HOME/.config/wofi/style.css"
        link_file "$DOTFILES_DIR/foot/foot.ini" "$HOME/.config/foot/foot.ini"
        link_file "$DOTFILES_DIR/hyprland" "$HOME/.config/hypr"
        link_file "$DOTFILES_DIR/pypr/config.toml" "$HOME/.config/pypr/config.toml"
        link_file "$DOTFILES_DIR/eww" "$HOME/.config/eww"
        link_file "$DOTFILES_DIR/ironbar" "$HOME/.config/ironbar"
        link_file "$DOTFILES_DIR/helix/config.toml" "$HOME/.config/helix/config.toml"
        link_file "$DOTFILES_DIR/solaar/config.yaml" "$HOME/.config/solaar/config.yaml"
        link_file "$DOTFILES_DIR/environment.d/ralphglasses.conf" "$HOME/.config/environment.d/ralphglasses.conf"
        link_file "$DOTFILES_DIR/fontconfig/conf.d/51-monospace.conf" "$HOME/.config/fontconfig/conf.d/51-monospace.conf"
        link_file "$DOTFILES_DIR/metapac" "$HOME/.config/metapac"
        link_file "$DOTFILES_DIR/paru/paru.conf" "$HOME/.config/paru/paru.conf"
        link_file "$DOTFILES_DIR/topgrade/topgrade.toml" "$HOME/.config/topgrade.toml"
        link_file "$DOTFILES_DIR/wlogout/layout" "$HOME/.config/wlogout/layout"
        link_file "$DOTFILES_DIR/wlogout/style.css" "$HOME/.config/wlogout/style.css"
        link_file "$DOTFILES_DIR/gtk/settings.ini" "$HOME/.config/gtk-3.0/settings.ini"
        link_file "$DOTFILES_DIR/xdg-desktop-portal/portals.conf" "$HOME/.config/xdg-desktop-portal/portals.conf"
    fi
    link_file "$DOTFILES_DIR/btop"        "$HOME/.config/btop"
    link_file "$DOTFILES_DIR/yazi"        "$HOME/.config/yazi"
    link_file "$DOTFILES_DIR/cava"        "$HOME/.config/cava"
    link_file "$DOTFILES_DIR/clipse/config.json" "$HOME/.config/clipse/config.json"
    link_file "$DOTFILES_DIR/clipse/custom_theme.json" "$HOME/.config/clipse/custom_theme.json"
    link_file "$DOTFILES_DIR/glow"        "$HOME/.config/glow"

    # Individual file symlinks (non-XDG)
    link_file "$DOTFILES_DIR/tmux/tmux.conf" "$HOME/.tmux.conf"

    # Tattoy (terminal shader compositor)
    if [[ "$OS" == "Darwin" ]]; then
        link_file "$DOTFILES_DIR/tattoy/tattoy.toml" "$HOME/Library/Application Support/tattoy/tattoy.toml"
    else
        link_file "$DOTFILES_DIR/tattoy/tattoy.toml" "${XDG_CONFIG_HOME:-$HOME/.config}/tattoy/tattoy.toml"
    fi

    # Platform-specific service management
    if [[ "$OS" == "Darwin" ]]; then
        # RetroVisor auto-launch
        link_file "$DOTFILES_DIR/retrovisor/com.dirkwhoffmann.RetroVisor.plist" \
            "$HOME/Library/LaunchAgents/com.dirkwhoffmann.RetroVisor.plist"

        # Shader auto-rotation (disabled by default — enable with: shader-auto start)
        link_file "$DOTFILES_DIR/ghostty/com.dotfiles.shader-rotate.plist" \
            "$HOME/Library/LaunchAgents/com.dotfiles.shader-rotate.plist"
    elif [[ "$OS" == "Linux" ]]; then
        log_info "Installing systemd user services..."
        mkdir -p "$HOME/.config/systemd/user"
        link_file "$DOTFILES_DIR/systemd/shader-rotate.timer" "$HOME/.config/systemd/user/shader-rotate.timer"
        link_file "$DOTFILES_DIR/systemd/shader-rotate.service" "$HOME/.config/systemd/user/shader-rotate.service"
        link_file "$DOTFILES_DIR/systemd/tmux.service" "$HOME/.config/systemd/user/tmux.service"
        link_file "$DOTFILES_DIR/systemd/eww-calendar-sync.service" "$HOME/.config/systemd/user/eww-calendar-sync.service"
        link_file "$DOTFILES_DIR/systemd/eww-calendar-sync.timer" "$HOME/.config/systemd/user/eww-calendar-sync.timer"
        link_file "$DOTFILES_DIR/systemd/mx-battery-notify.service" "$HOME/.config/systemd/user/mx-battery-notify.service"
        link_file "$DOTFILES_DIR/systemd/mx-battery-notify.timer" "$HOME/.config/systemd/user/mx-battery-notify.timer"
        link_file "$DOTFILES_DIR/systemd/rg-status-bar.service" "$HOME/.config/systemd/user/rg-status-bar.service"
        link_file "$DOTFILES_DIR/systemd/rg-status-bar.timer" "$HOME/.config/systemd/user/rg-status-bar.timer"
        link_file "$DOTFILES_DIR/systemd/rclone-gdrive.service" "$HOME/.config/systemd/user/rclone-gdrive.service"
        link_file "$DOTFILES_DIR/systemd/rclone-mega.service" "$HOME/.config/systemd/user/rclone-mega.service"
        link_file "$DOTFILES_DIR/systemd/stash.service" "$HOME/.config/systemd/user/stash.service"
        link_file "$DOTFILES_DIR/systemd/stash-ai-server.service" "$HOME/.config/systemd/user/stash-ai-server.service"
        link_file "$DOTFILES_DIR/systemd/stash-healthcheck.service" "$HOME/.config/systemd/user/stash-healthcheck.service"
        link_file "$DOTFILES_DIR/systemd/stash-healthcheck.timer" "$HOME/.config/systemd/user/stash-healthcheck.timer"
        link_file "$DOTFILES_DIR/systemd/stash-maintenance.service" "$HOME/.config/systemd/user/stash-maintenance.service"
        link_file "$DOTFILES_DIR/systemd/stash-maintenance.timer" "$HOME/.config/systemd/user/stash-maintenance.timer"
        link_file "$DOTFILES_DIR/systemd/nsfw-ai-model-server.service" "$HOME/.config/systemd/user/nsfw-ai-model-server.service"
        link_file "$DOTFILES_DIR/systemd/rg-marathon@.service" "$HOME/.config/systemd/user/rg-marathon@.service"
        link_file "$DOTFILES_DIR/systemd/makima.service" "$HOME/.config/systemd/user/makima.service"
        systemctl --user daemon-reload

        # Validate cross-repo symlinks (warn if sibling repos are missing)
        local cross_repo_links=(
            "$DOTFILES_DIR/scripts/rg-status-bar.sh"
        )
        for link in "${cross_repo_links[@]}"; do
            if [[ -L "$link" ]] && [[ ! -e "$link" ]]; then
                log_warn "Broken cross-repo symlink: $link -> $(readlink "$link")"
            fi
        done
    fi
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
    check_link "$DOTFILES_DIR/zsh/profile"            "$HOME/.profile"
    check_link "$DOTFILES_DIR/git/gitconfig"          "$HOME/.gitconfig"
    check_link "$DOTFILES_DIR/ssh/config"             "$HOME/.ssh/config"
    check_link "$DOTFILES_DIR/starship/starship.toml" "$HOME/.config/starship.toml"
    if [[ -f "$HOME/.codex/config.toml" ]]; then
        log_success "OK (copy): $HOME/.codex/config.toml"
    else
        log_error "Missing: $HOME/.codex/config.toml"
        errors=$((errors + 1))
    fi
    check_link "$DOTFILES_DIR/ghostty"    "$HOME/.config/ghostty"
    check_link "$DOTFILES_DIR/nvim"       "$HOME/.config/nvim"
    check_link "$DOTFILES_DIR/bat"        "$HOME/.config/bat"
    check_link "$DOTFILES_DIR/fastfetch"  "$HOME/.config/fastfetch"
    check_link "$DOTFILES_DIR/git/delta"  "$HOME/.config/delta"
    check_link "$DOTFILES_DIR/git/ignore" "$HOME/.config/git/ignore"
    check_link "$DOTFILES_DIR/gh"         "$HOME/.config/gh"
    check_link "$DOTFILES_DIR/k9s"        "$HOME/.config/k9s"
    check_link "$DOTFILES_DIR/lazygit"    "$HOME/.config/lazygit"
    if [[ "$OS" == "Darwin" ]]; then
        check_link "$DOTFILES_DIR/aerospace/aerospace.toml" "$HOME/.aerospace.toml"
        check_link "$DOTFILES_DIR/sketchybar"  "$HOME/.config/sketchybar"
        check_link "$DOTFILES_DIR/borders"     "$HOME/.config/borders"
    elif [[ "$OS" == "Linux" ]]; then
        # waybar, mako quarantined — replaced by eww + swaync
        check_link "$DOTFILES_DIR/swaync/config.json" "$HOME/.config/swaync/config.json"
        check_link "$DOTFILES_DIR/swaync/style.css" "$HOME/.config/swaync/style.css"
        check_link "$DOTFILES_DIR/wofi/config" "$HOME/.config/wofi/config"
        check_link "$DOTFILES_DIR/wofi/style.css" "$HOME/.config/wofi/style.css"
        check_link "$DOTFILES_DIR/foot/foot.ini" "$HOME/.config/foot/foot.ini"
        check_link "$DOTFILES_DIR/hyprland" "$HOME/.config/hypr"
        check_link "$DOTFILES_DIR/pypr/config.toml" "$HOME/.config/pypr/config.toml"
        check_link "$DOTFILES_DIR/eww" "$HOME/.config/eww"
        check_link "$DOTFILES_DIR/helix/config.toml" "$HOME/.config/helix/config.toml"
        check_link "$DOTFILES_DIR/solaar/config.yaml" "$HOME/.config/solaar/config.yaml"
        check_link "$DOTFILES_DIR/environment.d/ralphglasses.conf" "$HOME/.config/environment.d/ralphglasses.conf"
        check_link "$DOTFILES_DIR/fontconfig/conf.d/51-monospace.conf" "$HOME/.config/fontconfig/conf.d/51-monospace.conf"
        check_link "$DOTFILES_DIR/metapac" "$HOME/.config/metapac"
        check_link "$DOTFILES_DIR/paru/paru.conf" "$HOME/.config/paru/paru.conf"
        check_link "$DOTFILES_DIR/topgrade/topgrade.toml" "$HOME/.config/topgrade.toml"
        check_link "$DOTFILES_DIR/wlogout/layout" "$HOME/.config/wlogout/layout"
        check_link "$DOTFILES_DIR/wlogout/style.css" "$HOME/.config/wlogout/style.css"
        check_link "$DOTFILES_DIR/gtk/settings.ini" "$HOME/.config/gtk-3.0/settings.ini"
        check_link "$DOTFILES_DIR/xdg-desktop-portal/portals.conf" "$HOME/.config/xdg-desktop-portal/portals.conf"
    fi
    check_link "$DOTFILES_DIR/btop"        "$HOME/.config/btop"
    check_link "$DOTFILES_DIR/yazi"        "$HOME/.config/yazi"
    check_link "$DOTFILES_DIR/cava"        "$HOME/.config/cava"
    check_link "$DOTFILES_DIR/glow"        "$HOME/.config/glow"
    check_link "$DOTFILES_DIR/clipse/config.json" "$HOME/.config/clipse/config.json"
    check_link "$DOTFILES_DIR/clipse/custom_theme.json" "$HOME/.config/clipse/custom_theme.json"
    check_link "$DOTFILES_DIR/tmux/tmux.conf" "$HOME/.tmux.conf"
    if [[ "$OS" == "Darwin" ]]; then
        check_link "$DOTFILES_DIR/tattoy/tattoy.toml" "$HOME/Library/Application Support/tattoy/tattoy.toml"
        check_link "$DOTFILES_DIR/retrovisor/com.dirkwhoffmann.RetroVisor.plist" \
            "$HOME/Library/LaunchAgents/com.dirkwhoffmann.RetroVisor.plist"
        check_link "$DOTFILES_DIR/ghostty/com.dotfiles.shader-rotate.plist" \
            "$HOME/Library/LaunchAgents/com.dotfiles.shader-rotate.plist"
    else
        # tattoy.toml is intentionally a copy (app writes to it at runtime)
        if [[ -f "${XDG_CONFIG_HOME:-$HOME/.config}/tattoy/tattoy.toml" ]]; then
            log_success "OK (copy): ${XDG_CONFIG_HOME:-$HOME/.config}/tattoy/tattoy.toml"
        else
            log_error "Missing: ${XDG_CONFIG_HOME:-$HOME/.config}/tattoy/tattoy.toml"
            errors=$((errors + 1))
        fi

        log_info "Checking systemd user services..."
        local svc_dir="$HOME/.config/systemd/user"
        check_link "$DOTFILES_DIR/systemd/shader-rotate.timer" "$svc_dir/shader-rotate.timer"
        check_link "$DOTFILES_DIR/systemd/shader-rotate.service" "$svc_dir/shader-rotate.service"
        check_link "$DOTFILES_DIR/systemd/tmux.service" "$svc_dir/tmux.service"
        check_link "$DOTFILES_DIR/systemd/eww-calendar-sync.service" "$svc_dir/eww-calendar-sync.service"
        check_link "$DOTFILES_DIR/systemd/eww-calendar-sync.timer" "$svc_dir/eww-calendar-sync.timer"
        check_link "$DOTFILES_DIR/systemd/mx-battery-notify.service" "$svc_dir/mx-battery-notify.service"
        check_link "$DOTFILES_DIR/systemd/mx-battery-notify.timer" "$svc_dir/mx-battery-notify.timer"
        check_link "$DOTFILES_DIR/systemd/rg-status-bar.service" "$svc_dir/rg-status-bar.service"
        check_link "$DOTFILES_DIR/systemd/rg-status-bar.timer" "$svc_dir/rg-status-bar.timer"
        check_link "$DOTFILES_DIR/systemd/rclone-gdrive.service" "$svc_dir/rclone-gdrive.service"
        check_link "$DOTFILES_DIR/systemd/rclone-mega.service" "$svc_dir/rclone-mega.service"
        check_link "$DOTFILES_DIR/systemd/stash.service" "$svc_dir/stash.service"
        check_link "$DOTFILES_DIR/systemd/stash-ai-server.service" "$svc_dir/stash-ai-server.service"
        check_link "$DOTFILES_DIR/systemd/stash-healthcheck.service" "$svc_dir/stash-healthcheck.service"
        check_link "$DOTFILES_DIR/systemd/stash-healthcheck.timer" "$svc_dir/stash-healthcheck.timer"
        check_link "$DOTFILES_DIR/systemd/stash-maintenance.service" "$svc_dir/stash-maintenance.service"
        check_link "$DOTFILES_DIR/systemd/stash-maintenance.timer" "$svc_dir/stash-maintenance.timer"
        check_link "$DOTFILES_DIR/systemd/nsfw-ai-model-server.service" "$svc_dir/nsfw-ai-model-server.service"
        check_link "$DOTFILES_DIR/systemd/rg-marathon@.service" "$svc_dir/rg-marathon@.service"
        check_link "$DOTFILES_DIR/systemd/makima.service" "$svc_dir/makima.service"
    fi

    if [[ "$OS" == "Darwin" ]]; then
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
    fi

    if [[ "$OS" == "Linux" ]]; then
        if command -v metapac &>/dev/null; then
            log_info "Checking metapac packages..."
            local unmanaged
            unmanaged="$(metapac unmanaged 2>&1)"
            if echo "$unmanaged" | grep -q "no unmanaged packages"; then
                log_success "All packages managed by metapac"
            else
                log_warn "Unmanaged packages found — run: metapac unmanaged"
            fi
        elif command -v yay &>/dev/null && [[ -f "$DOTFILES_DIR/Pacfile" ]]; then
            log_info "Checking pacman/AUR packages (legacy Pacfile)..."
            local missing=0
            while IFS= read -r pkg; do
                if ! yay -Qi "$pkg" &>/dev/null; then
                    log_error "Missing package: $pkg"
                    missing=$((missing + 1))
                fi
            done < <(grep -v '^\s*#' "$DOTFILES_DIR/Pacfile" | grep -v '^\s*$')
            if [[ $missing -eq 0 ]]; then
                log_success "All Pacfile packages installed"
            fi
            errors=$((errors + missing))
        else
            log_warn "Neither metapac nor yay installed"
            errors=$((errors + 1))
        fi
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
    if _has_tte; then
        echo "DOTFILES INSTALLER" | tte synthgrid \
            --grid-gradient-stops 57c7ff ff6ac1 \
            --text-gradient-stops 57c7ff 5af78e \
            --max-active-blocks 0.1 2>/dev/null
    else
        echo "  dotfiles installer"
        echo "  ──────────────────"
    fi
    log_info "repo: $DOTFILES_DIR"
    echo ""

    if $CHECK_ONLY; then
        check_symlinks
        return $?
    fi

    log_phase "PACKAGE MANAGEMENT"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        install_homebrew
        install_brew_packages
    elif [[ "$OS" == "Linux" ]]; then
        install_yay
        install_paru
        install_metapac
        # Symlink metapac config before sync so it reads group files
        if [[ -d "$DOTFILES_DIR/metapac" ]]; then
            link_file "$DOTFILES_DIR/metapac" "$HOME/.config/metapac"
        fi
        install_linux_packages
    fi

    log_phase "SHELL ENVIRONMENT"
    install_omz
    install_omz_plugins

    log_phase "EDITOR PLUGINS"
    install_vim_plug
    install_tpm

    if [[ "$OS" == "Darwin" ]]; then
        install_retrovisor
    fi

    log_phase "SYMLINK CONFIGURATION"
    create_symlinks
    setup_tattoy_shaders

    # Build bat cache (custom themes)
    if command -v bat &>/dev/null; then
        log_info "Building bat theme cache..."
        bat cache --build 2>/dev/null
    fi

    echo ""
    if $BACKUP_CREATED; then
        log_warn "Backups saved to: $BACKUP_DIR"
    fi

    if _has_tte; then
        echo "INSTALLATION COMPLETE" | tte fireworks \
            --firework-colors 5af78e 57c7ff ff6ac1 \
            --final-gradient-stops 5af78e 57c7ff \
            --explode-anywhere 2>/dev/null
    else
        log_success "Done! Restart your shell or run: source ~/.zshrc"
    fi
    echo ""
}

main "$@"
