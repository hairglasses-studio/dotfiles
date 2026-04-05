#!/usr/bin/env bash
# manjaro/install.sh — Manjaro Linux installer for dotfiles
# Feature flags in dotfiles.toml — toggle components on/off
# Usage: ./manjaro/install.sh [--check] [--profile minimal|dev|full|cyberpunk]
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"
BACKUP_DIR="$HOME/.dotfiles-backup-$(date +%Y%m%d-%H%M%S)"
BACKUP_CREATED=false
CHECK_ONLY=false
PROFILE=""

# Parse args
while [[ $# -gt 0 ]]; do
    case "$1" in
        --check) CHECK_ONLY=true; shift ;;
        --profile) PROFILE="${2:-}"; shift 2 ;;
        *) shift ;;
    esac
done

# ── Feature flag system ──────────────────────────────────────
declare -A FEATURES
_parse_features() {
    local toml="$DOTFILES/dotfiles.toml"
    [[ -f "$toml" ]] || return 0
    local section=""
    while IFS= read -r line; do
        line="${line%%#*}"          # strip comments
        line="${line//[[:space:]]/}" # strip whitespace
        [[ -z "$line" ]] && continue
        if [[ "$line" =~ ^\[(.+)\]$ ]]; then
            section="${BASH_REMATCH[1]}"
        elif [[ "$line" =~ ^([a-zA-Z0-9_-]+)=(true|false)$ ]]; then
            FEATURES["${BASH_REMATCH[1]}"]="${BASH_REMATCH[2]}"
        fi
    done < "$toml"

    # Apply profile override if specified
    if [[ -n "$PROFILE" ]]; then
        local in_profile=false
        while IFS= read -r line; do
            line="${line%%#*}"
            [[ "$line" =~ ^\[profiles\.$PROFILE\] ]] && in_profile=true && continue
            [[ "$line" =~ ^\[ ]] && in_profile=false
            if $in_profile && [[ "$line" =~ enable.*=.*\[(.+)\] ]]; then
                local items="${BASH_REMATCH[1]}"
                if [[ "$items" == *'"*"'* ]]; then
                    # Wildcard — enable everything
                    for key in "${!FEATURES[@]}"; do FEATURES["$key"]="true"; done
                else
                    # Disable all, then enable listed
                    for key in "${!FEATURES[@]}"; do FEATURES["$key"]="false"; done
                    items="${items//\"/}"
                    IFS=',' read -ra arr <<< "$items"
                    for item in "${arr[@]}"; do
                        item="${item//[[:space:]]/}"
                        FEATURES["$item"]="true"
                    done
                fi
            fi
        done < "$toml"
    fi
}

is_enabled() { [[ "${FEATURES[${1}]:-true}" == "true" ]]; }

_parse_features

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
        yay -S --needed --noconfirm ttf-maple-nerd-font logiops makima-bin
    else
        warn "yay not found — skipping AUR packages (ttf-maple-nerd-font, logiops, makima-bin)"
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

# ── Input devices ────────────────────────────────────────────
setup_input_devices() {
    info "Setting up input device configs..."

    # Solaar — copy (not symlink) because solaar writes runtime state
    if is_enabled solaar && [[ -f "$DOTFILES/solaar/config.yaml" ]]; then
        local solaar_dir="$HOME/.config/solaar"
        mkdir -p "$solaar_dir"
        if [[ -f "$solaar_dir/config.yaml" ]]; then
            backup_file "$solaar_dir/config.yaml"
        fi
        cp "$DOTFILES/solaar/config.yaml" "$solaar_dir/config.yaml"
        success "Copied solaar config"
    fi

    # makima — symlink the config directory
    if is_enabled makima && [[ -d "$DOTFILES/makima" ]]; then
        link_file "$DOTFILES/makima" "$HOME/.config/makima"
    fi

    # logiops — remind about deploy script (needs sudo)
    if is_enabled logiops && [[ -f "$DOTFILES/logiops/logid.cfg" ]]; then
        warn "logiops config tracked in dotfiles/logiops/logid.cfg"
        warn "Deploy to /etc/: ./scripts/logiops-deploy.sh (needs sudo)"
    fi

    # udev rules (Keychron USB power, etc.)
    if [[ -d "$DOTFILES/udev" ]]; then
        for rule in "$DOTFILES/udev/"*.rules; do
            [[ -f "$rule" ]] || continue
            local dest="/etc/udev/rules.d/$(basename "$rule")"
            if ! diff -q "$rule" "$dest" &>/dev/null; then
                sudo cp "$rule" "$dest"
                info "  Installed udev rule: $(basename "$rule")"
            fi
        done
        sudo udevadm control --reload-rules
    fi
}

# ── MCP server registration ──────────────────────────────────
setup_mcp() {
    # Symlink .mcp.json to parent workspace so Claude Code finds it
    local workspace
    workspace="$(dirname "$DOTFILES")"
    if [[ -f "$DOTFILES/.mcp.json" ]] && [[ -d "$workspace" ]]; then
        link_file "$DOTFILES/.mcp.json" "$workspace/.mcp.json"
    fi
}

# ── Symlinks ──────────────────────────────────────────────────
create_symlinks() {
    info "Creating symlinks (profile: ${PROFILE:-dotfiles.toml})..."

    # ── Core ──
    is_enabled zsh      && link_file "$DOTFILES/zsh/zshrc"              "$HOME/.zshrc"
    is_enabled zsh      && link_file "$DOTFILES/zsh/p10k.zsh"           "$HOME/.p10k.zsh"
    is_enabled zsh      && link_file "$DOTFILES/zsh/zshenv"             "$HOME/.zshenv"
    is_enabled git      && link_file "$DOTFILES/git/gitconfig"          "$HOME/.gitconfig"
    is_enabled git      && link_file "$DOTFILES/git/delta"              "$HOME/.config/delta"
    is_enabled git      && link_file "$DOTFILES/git/ignore"             "$HOME/.config/git/ignore"
    is_enabled ssh      && link_file "$DOTFILES/ssh/config"             "$HOME/.ssh/config"
    is_enabled starship && link_file "$DOTFILES/starship/starship.toml" "$HOME/.config/starship.toml"
    is_enabled tmux     && link_file "$DOTFILES/tmux/tmux.conf"         "$HOME/.tmux.conf"
    is_enabled nvim     && link_file "$DOTFILES/nvim"                   "$HOME/.config/nvim"

    # ── Terminal ──
    is_enabled ghostty   && link_file "$DOTFILES/ghostty"   "$HOME/.config/ghostty"
    is_enabled foot      && link_file "$DOTFILES/foot"      "$HOME/.config/foot"
    is_enabled bat       && link_file "$DOTFILES/bat"       "$HOME/.config/bat"
    is_enabled fastfetch && link_file "$DOTFILES/fastfetch" "$HOME/.config/fastfetch"

    # ── TUI theming ──
    is_enabled gh      && link_file "$DOTFILES/gh"      "$HOME/.config/gh"
    is_enabled k9s     && link_file "$DOTFILES/k9s"     "$HOME/.config/k9s"
    is_enabled lazygit && link_file "$DOTFILES/lazygit"  "$HOME/.config/lazygit"
    is_enabled btop    && link_file "$DOTFILES/btop"     "$HOME/.config/btop"
    is_enabled yazi    && link_file "$DOTFILES/yazi"     "$HOME/.config/yazi"
    is_enabled cava    && link_file "$DOTFILES/cava"     "$HOME/.config/cava"
    is_enabled glow    && link_file "$DOTFILES/glow"     "$HOME/.config/glow"

    # ── Desktop (Linux) ──
    is_enabled sway     && [[ -d "$DOTFILES/sway" ]]     && link_file "$DOTFILES/sway"     "$HOME/.config/sway"
    is_enabled waybar   && [[ -d "$DOTFILES/waybar" ]]   && link_file "$DOTFILES/waybar"   "$HOME/.config/waybar"
    is_enabled hyprland && [[ -d "$DOTFILES/hyprland" ]] && link_file "$DOTFILES/hyprland" "$HOME/.config/hypr"
    is_enabled eww      && [[ -d "$DOTFILES/eww" ]]      && link_file "$DOTFILES/eww"      "$HOME/.config/eww"
    is_enabled mako     && [[ -d "$DOTFILES/mako" ]]     && link_file "$DOTFILES/mako"     "$HOME/.config/mako"
    is_enabled wofi     && [[ -d "$DOTFILES/wofi" ]]     && link_file "$DOTFILES/wofi"     "$HOME/.config/wofi"
    is_enabled wlogout  && [[ -d "$DOTFILES/wlogout" ]]  && link_file "$DOTFILES/wlogout"  "$HOME/.config/wlogout"

    # ── GTK ──
    if is_enabled gtk && [[ -f "$DOTFILES/gtk/settings.ini" ]]; then
        mkdir -p "$HOME/.config/gtk-3.0"
        link_file "$DOTFILES/gtk/settings.ini" "$HOME/.config/gtk-3.0/settings.ini"
    fi
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

        # Enable makima system service (provided by makima-bin package)
        if is_enabled makima && [[ -f "/usr/lib/systemd/system/makima.service" ]]; then
            sudo systemctl enable makima.service
            info "  makima service enabled"
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

    # ── Input devices ──
    check_link "$DOTFILES/makima" "$HOME/.config/makima"

    if systemctl is-active logid.service &>/dev/null; then
        success "logid service active"
    else
        error "logid service not active"
        errors=$((errors + 1))
    fi

    if systemctl is-enabled makima.service &>/dev/null; then
        success "makima service enabled"
    else
        warn "makima service not enabled"
    fi

    if [[ -f "$HOME/.config/solaar/config.yaml" ]]; then
        success "solaar config present"
    else
        warn "solaar config missing"
    fi

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
    setup_input_devices
    setup_mcp
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
