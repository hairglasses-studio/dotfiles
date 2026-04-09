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
PRINT_LINK_SPECS=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --check) CHECK_ONLY=true ;;
        --print-link-specs) PRINT_LINK_SPECS=true ;;
        *) printf 'Unknown option: %s\n' "$1" >&2; exit 2 ;;
    esac
    shift
done

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

source_exists() {
    local src="$1"
    [[ -e "$src" || -L "$src" ]]
}

link_if_present() {
    local src="$1" dst="$2"
    if source_exists "$src"; then
        link_file "$src" "$dst"
    else
        log_warn "Skipping missing source: $src"
    fi
}

print_common_link_specs() {
    cat <<EOF
$DOTFILES_DIR/zsh/zshrc|$HOME/.zshrc
$DOTFILES_DIR/zsh/p10k.zsh|$HOME/.p10k.zsh
$DOTFILES_DIR/zsh/zshenv|$HOME/.zshenv
$DOTFILES_DIR/zsh/profile|$HOME/.profile
$DOTFILES_DIR/git/gitconfig|$HOME/.gitconfig
$DOTFILES_DIR/ssh/config|$HOME/.ssh/config
$DOTFILES_DIR/starship/starship.toml|$HOME/.config/starship.toml
$DOTFILES_DIR/ghostty|$HOME/.config/ghostty
$DOTFILES_DIR/kitty|$HOME/.config/kitty
$DOTFILES_DIR/nvim|$HOME/.config/nvim
$DOTFILES_DIR/bat|$HOME/.config/bat
$DOTFILES_DIR/fastfetch|$HOME/.config/fastfetch
$DOTFILES_DIR/git/delta|$HOME/.config/delta
$DOTFILES_DIR/git/ignore|$HOME/.config/git/ignore
$DOTFILES_DIR/gh|$HOME/.config/gh
$DOTFILES_DIR/k9s|$HOME/.config/k9s
$DOTFILES_DIR/lazygit|$HOME/.config/lazygit
$DOTFILES_DIR/btop|$HOME/.config/btop
$DOTFILES_DIR/yazi|$HOME/.config/yazi
$DOTFILES_DIR/cava|$HOME/.config/cava
$DOTFILES_DIR/clipse/config.json|$HOME/.config/clipse/config.json
$DOTFILES_DIR/clipse/custom_theme.json|$HOME/.config/clipse/custom_theme.json
$DOTFILES_DIR/glow|$HOME/.config/glow
$DOTFILES_DIR/tmux/tmux.conf|$HOME/.tmux.conf
$DOTFILES_DIR/scripts/hg-codex-launch.sh|$HOME/.local/bin/hg-codex-launch.sh
$DOTFILES_DIR/scripts/hg-claude-launch.sh|$HOME/.local/bin/hg-claude-launch.sh
$DOTFILES_DIR/scripts/hg-gemini-launch.sh|$HOME/.local/bin/hg-gemini-launch.sh
$DOTFILES_DIR/scripts/hg-codex-worktree-prune.sh|$HOME/.local/bin/hg-codex-worktree-prune.sh
$DOTFILES_DIR/scripts/hg-agent-home-sync.sh|$HOME/.local/bin/hg-agent-home-sync.sh
EOF
}

print_darwin_link_specs() {
    cat <<EOF
$DOTFILES_DIR/aerospace/aerospace.toml|$HOME/.aerospace.toml
$DOTFILES_DIR/sketchybar|$HOME/.config/sketchybar
$DOTFILES_DIR/borders|$HOME/.config/borders
$DOTFILES_DIR/tattoy/tattoy.toml|$HOME/Library/Application Support/tattoy/tattoy.toml
$DOTFILES_DIR/retrovisor/com.dirkwhoffmann.RetroVisor.plist|$HOME/Library/LaunchAgents/com.dirkwhoffmann.RetroVisor.plist
$DOTFILES_DIR/ghostty/com.dotfiles.shader-rotate.plist|$HOME/Library/LaunchAgents/com.dotfiles.shader-rotate.plist
EOF
}

print_linux_link_specs() {
    cat <<EOF
$DOTFILES_DIR/swaync/config.json|$HOME/.config/swaync/config.json
$DOTFILES_DIR/swaync/style.css|$HOME/.config/swaync/style.css
$DOTFILES_DIR/hyprshell|$HOME/.config/hyprshell
$DOTFILES_DIR/hypr-dock|$HOME/.config/hypr-dock
$DOTFILES_DIR/hyprdynamicmonitors|$HOME/.config/hyprdynamicmonitors
$DOTFILES_DIR/hyprland-autoname-workspaces|$HOME/.config/hyprland-autoname-workspaces
$DOTFILES_DIR/wofi/config|$HOME/.config/wofi/config
$DOTFILES_DIR/wofi/style.css|$HOME/.config/wofi/style.css
$DOTFILES_DIR/hyprland|$HOME/.config/hypr
$DOTFILES_DIR/pypr/config.toml|$HOME/.config/pypr/config.toml
$DOTFILES_DIR/eww|$HOME/.config/eww
$DOTFILES_DIR/helix/config.toml|$HOME/.config/helix/config.toml
$DOTFILES_DIR/makima|$HOME/.config/makima
$DOTFILES_DIR/environment.d/ralphglasses.conf|$HOME/.config/environment.d/ralphglasses.conf
$DOTFILES_DIR/fontconfig/conf.d/51-monospace.conf|$HOME/.config/fontconfig/conf.d/51-monospace.conf
$DOTFILES_DIR/metapac|$HOME/.config/metapac
$DOTFILES_DIR/paru/paru.conf|$HOME/.config/paru/paru.conf
$DOTFILES_DIR/topgrade/topgrade.toml|$HOME/.config/topgrade.toml
$DOTFILES_DIR/wlogout/layout|$HOME/.config/wlogout/layout
$DOTFILES_DIR/wlogout/style.css|$HOME/.config/wlogout/style.css
$DOTFILES_DIR/gtk/settings.ini|$HOME/.config/gtk-3.0/settings.ini
$DOTFILES_DIR/gtk-4.0/settings.ini|$HOME/.config/gtk-4.0/settings.ini
$DOTFILES_DIR/xdg-desktop-portal/portals.conf|$HOME/.config/xdg-desktop-portal/portals.conf
$DOTFILES_DIR/tattoy/tattoy.toml|${XDG_CONFIG_HOME:-$HOME/.config}/tattoy/tattoy.toml
$DOTFILES_DIR/scripts/kitty-shader-playlist.sh|$HOME/.local/bin/kitty-shader-playlist
$DOTFILES_DIR/scripts/kitty-dev-launch.sh|$HOME/.local/bin/kitty-dev-launch
$DOTFILES_DIR/scripts/kitty-visual-launch.sh|$HOME/.local/bin/kitty-visual-launch
$DOTFILES_DIR/scripts/app-launcher.sh|$HOME/.local/bin/app-launcher
$DOTFILES_DIR/scripts/app-switcher.sh|$HOME/.local/bin/app-switcher
$DOTFILES_DIR/scripts/juhradial-mx.sh|$HOME/.local/bin/juhradial-mx
$DOTFILES_DIR/scripts/juhradial-settings.sh|$HOME/.local/bin/juhradial-settings
EOF
}

print_linux_systemd_link_specs() {
    local src
    for src in "$DOTFILES_DIR"/systemd/*; do
        [[ -f "$src" ]] || continue
        [[ "$(basename "$src")" == "makima.service" ]] && continue
        printf '%s|%s\n' "$src" "$HOME/.config/systemd/user/$(basename "$src")"
    done
}

print_link_specs() {
    print_common_link_specs
    if [[ "$OS" == "Darwin" ]]; then
        print_darwin_link_specs
    elif [[ "$OS" == "Linux" ]]; then
        print_linux_link_specs
        print_linux_systemd_link_specs
    fi
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

# ── Tattoy shader symlink (Ghostty-only, skipped for kitty) ──
setup_tattoy_shaders() {
    log_info "Tattoy shaders skipped (kitty uses CRTty/DarkWindow pipeline)"
}

install_juhradial_stack() {
    if [[ "$OS" != "Linux" ]]; then
        return 0
    fi

    local script="$DOTFILES_DIR/scripts/juhradial-install.sh"
    if [[ ! -x "$script" ]]; then
        log_warn "Skipping juhradial install — missing executable: $script"
        return 0
    fi

    log_info "Installing juhradial-mx stack..."
    if "$script" --quiet; then
        log_success "juhradial-mx installed"
    else
        log_warn "juhradial-mx install reported an error — rerun: $script"
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
    if [[ "$OS" == "Linux" ]]; then
        mkdir -p "$HOME/.config/systemd/user"
    fi

    local src dst
    while IFS='|' read -r src dst; do
        [[ -n "$src" ]] || continue
        link_if_present "$src" "$dst"
    done < <(print_link_specs)

    # Platform-specific service management
    if [[ "$OS" == "Darwin" ]]; then
        :
    elif [[ "$OS" == "Linux" ]]; then
        log_info "Installing systemd user services..."
        systemctl --user daemon-reload

        mkdir -p "$HOME/.local/state/hypr"
        if [[ ! -f "$HOME/.local/state/hypr/monitors.dynamic.conf" ]]; then
            printf '# Generated by hyprdynamicmonitors.\n' > "$HOME/.local/state/hypr/monitors.dynamic.conf"
        fi
        mkdir -p "$HOME/.local/state/kitty/sessions"

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

    check_link_if_present() {
        local src="$1" dst="$2"
        if source_exists "$src"; then
            check_link "$src" "$dst"
        elif [[ -e "$dst" || -L "$dst" ]]; then
            log_warn "Source missing in repo, destination left as-is: $dst"
        else
            log_info "Skipping missing source: $src"
        fi
    }

    log_info "Checking symlinks..."
    local src dst
    while IFS='|' read -r src dst; do
        [[ -n "$src" ]] || continue
        check_link_if_present "$src" "$dst"
    done < <(print_link_specs)

    if [[ "$OS" == "Linux" ]]; then
        log_info "Checking systemd user services..."
    fi

    if [[ "$OS" == "Darwin" ]]; then
        log_info "Checking brew packages..."
        if command -v brew &>/dev/null; then
            if [[ -f "$DOTFILES_DIR/Brewfile" ]]; then
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
                log_info "Skipping brew package check (Brewfile not present)"
            fi
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
    if $PRINT_LINK_SPECS; then
        print_link_specs
        return 0
    fi

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
    install_juhradial_stack

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
