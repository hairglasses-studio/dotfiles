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
DOTFILES_TOPLEVEL="$DOTFILES_DIR"
DOTFILES_GIT_COMMON_DIR=""

_print_help() {
    cat <<'EOF'
install.sh — Hairglasses dotfiles installer

Symlinks configs from this repo to $HOME and enables the systemd
user units that drive the Hyprland desktop (ticker, overlays, cache
timers, notification daemon, etc).

Idempotent — safe to rerun after edits; re-applies the symlinks that
differ and leaves untouched ones alone.

USAGE:
  install.sh                 apply symlinks + enable services
  install.sh --check         print what would change, don't modify anything
  install.sh --print-link-specs
                             dump the link-spec catalogue and exit
  install.sh --list-services list the systemd user units this script enables
  install.sh -h | --help     this text

The service list is hardcoded in the `desktop_service_units` and
`desktop_passive_units` arrays near line 700 — use --list-services
for the current set without scanning the source.
EOF
}

LIST_SERVICES=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --check) CHECK_ONLY=true ;;
        --print-link-specs) PRINT_LINK_SPECS=true ;;
        --list-services) LIST_SERVICES=true ;;
        -h|--help) _print_help; exit 0 ;;
        *) printf 'Unknown option: %s\n' "$1" >&2
           printf 'Try: install.sh --help\n' >&2
           exit 2 ;;
    esac
    shift
done

if $LIST_SERVICES; then
    # Catalogue extraction: parse the arrays from this file itself so
    # the listing stays in lockstep even when someone bumps the arrays
    # but forgets to update a separate doc.
    # Extract unit names from the arrays by slicing between the opening
    # `local <name>=(` and the closing `)`, then keep just the unit-name
    # lines (strip leading whitespace, trailing comments, blanks).
    _extract_units() {
        awk -v start_re="local $1=\\\\(" '
            $0 ~ start_re { in_block=1; next }
            in_block && /^\s*\)/ { in_block=0; exit }
            in_block && NF > 0 {
                sub(/^[ \t]+/, "")
                sub(/[ \t]*#.*$/, "")
                if ($0 != "") print "  " $0
            }
        ' "${BASH_SOURCE[0]}" | sort -u
    }
    printf 'desktop_service_units:\n';  _extract_units desktop_service_units
    printf '\ndesktop_passive_units:\n'; _extract_units desktop_passive_units
    exit 0
fi

if git -C "$DOTFILES_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    DOTFILES_TOPLEVEL="$(git -C "$DOTFILES_DIR" rev-parse --show-toplevel 2>/dev/null || printf '%s' "$DOTFILES_DIR")"
    _dotfiles_common_dir="$(git -C "$DOTFILES_DIR" rev-parse --path-format=absolute --git-common-dir 2>/dev/null || true)"
    if [[ -n "$_dotfiles_common_dir" ]]; then
        DOTFILES_GIT_COMMON_DIR="$(cd "$DOTFILES_DIR" && cd "$_dotfiles_common_dir" 2>/dev/null && pwd -P || true)"
    fi
    unset _dotfiles_common_dir
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

copy_file() {
    local src="$1" dst="$2"

    if [[ -f "$dst" ]] && cmp -s "$src" "$dst"; then
        log_success "Already copied: $dst"
        return 0
    fi

    if [[ -e "$dst" ]] || [[ -L "$dst" ]]; then
        backup_file "$dst"
    fi

    mkdir -p "$(dirname "$dst")"
    install -m644 "$src" "$dst"
    log_success "Copied: $dst"
}

copy_file_if_present() {
    local src="$1" dst="$2"
    if source_exists "$src"; then
        copy_file "$src" "$dst"
    else
        log_warn "Skipping missing source: $src"
    fi
}

print_common_link_specs() {
    cat <<EOF
$DOTFILES_DIR/zsh/zshrc|$HOME/.zshrc
$DOTFILES_DIR/zsh/zshenv|$HOME/.zshenv
$DOTFILES_DIR/zsh/profile|$HOME/.profile
$DOTFILES_DIR/git/gitconfig|$HOME/.gitconfig
$DOTFILES_DIR/ssh/config|$HOME/.ssh/config
$DOTFILES_DIR/starship/starship.toml|$HOME/.config/starship.toml
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
$DOTFILES_DIR/scripts/theme-sync.sh|$HOME/.local/bin/theme-sync
$DOTFILES_DIR/scripts/hyprpm-bootstrap.sh|$HOME/.local/bin/hyprpm-bootstrap
EOF
}

print_darwin_link_specs() {
    cat <<EOF
$DOTFILES_DIR/tattoy/tattoy.toml|$HOME/Library/Application Support/tattoy/tattoy.toml
EOF
}

print_linux_link_specs() {
    cat <<EOF
$DOTFILES_DIR/swaync/config.json|$HOME/.config/swaync/config.json
$DOTFILES_DIR/swaync/style.css|$HOME/.config/swaync/style.css
$DOTFILES_DIR/ironbar|$HOME/.config/ironbar
$DOTFILES_DIR/hyprshell/config.toml|$HOME/.config/hyprshell/config.toml
$DOTFILES_DIR/hyprshell/styles.css|$HOME/.config/hyprshell/styles.css
$DOTFILES_DIR/hypr-dock|$HOME/.config/hypr-dock
$DOTFILES_DIR/hyprdynamicmonitors|$HOME/.config/hyprdynamicmonitors
$DOTFILES_DIR/hyprland-autoname-workspaces|$HOME/.config/hyprland-autoname-workspaces
$DOTFILES_DIR/wofi/config|$HOME/.config/wofi/config
$DOTFILES_DIR/wofi/style.css|$HOME/.config/wofi/style.css
$DOTFILES_DIR/hyprland|$HOME/.config/hypr
$DOTFILES_DIR/pypr/config.toml|$HOME/.config/pypr/config.toml
$DOTFILES_DIR/helix/config.toml|$HOME/.config/helix/config.toml
$DOTFILES_DIR/environment.d/theme.conf|$HOME/.config/environment.d/theme.conf
$DOTFILES_DIR/environment.d/ralphglasses.conf|$HOME/.config/environment.d/ralphglasses.conf
$DOTFILES_DIR/fontconfig/conf.d/51-monospace.conf|$HOME/.config/fontconfig/conf.d/51-monospace.conf
$DOTFILES_DIR/metapac|$HOME/.config/metapac
$DOTFILES_DIR/paru/paru.conf|$HOME/.config/paru/paru.conf
$DOTFILES_DIR/qt5ct/qt5ct.conf|$HOME/.config/qt5ct/qt5ct.conf
$DOTFILES_DIR/qt6ct/qt6ct.conf|$HOME/.config/qt6ct/qt6ct.conf
$DOTFILES_DIR/Kvantum/kvantum.kvconfig|$HOME/.config/Kvantum/kvantum.kvconfig
$DOTFILES_DIR/kdeglobals|$HOME/.config/kdeglobals
$DOTFILES_DIR/kcminputrc|$HOME/.config/kcminputrc
$DOTFILES_DIR/topgrade/topgrade.toml|$HOME/.config/topgrade.toml
$DOTFILES_DIR/wlogout/layout|$HOME/.config/wlogout/layout
$DOTFILES_DIR/wlogout/style.css|$HOME/.config/wlogout/style.css
$DOTFILES_DIR/gtk/settings.ini|$HOME/.config/gtk-3.0/settings.ini
$DOTFILES_DIR/gtk-4.0/settings.ini|$HOME/.config/gtk-4.0/settings.ini
$DOTFILES_DIR/xdg-desktop-portal/portals.conf|$HOME/.config/xdg-desktop-portal/portals.conf
$DOTFILES_DIR/tattoy/tattoy.toml|${XDG_CONFIG_HOME:-$HOME/.config}/tattoy/tattoy.toml
$DOTFILES_DIR/scripts/kitty-shader-playlist.sh|$HOME/.local/bin/kitty-shader-playlist
$DOTFILES_DIR/scripts/kitty-shell-launch.sh|$HOME/.local/bin/kitty-shell-launch
$DOTFILES_DIR/scripts/kitty-dev-launch.sh|$HOME/.local/bin/kitty-dev-launch
$DOTFILES_DIR/scripts/kitty-visual-launch.sh|$HOME/.local/bin/kitty-visual-launch
$DOTFILES_DIR/scripts/jellyfin-stack-boot.sh|$HOME/.local/bin/jellyfin-stack-boot.sh
$DOTFILES_DIR/scripts/app-launcher.sh|$HOME/.local/bin/app-launcher
$DOTFILES_DIR/scripts/app-switcher.sh|$HOME/.local/bin/app-switcher
$DOTFILES_DIR/scripts/keybind-ticker.py|$HOME/.local/bin/keybind-ticker
$DOTFILES_DIR/scripts/dev-console.sh|$HOME/.local/bin/dev-console
$DOTFILES_DIR/scripts/pin-dev-console-session.sh|$HOME/.local/bin/pin-dev-console-session
$DOTFILES_DIR/scripts/palette-playlist.sh|$HOME/.local/bin/palette-playlist
$DOTFILES_DIR/scripts/kitty-playlist-validate.sh|$HOME/.local/bin/kitty-playlist-validate
$DOTFILES_DIR/scripts/retroarch-archive-homebrew-manifest.py|$HOME/.local/bin/retroarch-archive-homebrew-manifest
$DOTFILES_DIR/scripts/retroarch-archive-homebrew-fetch.py|$HOME/.local/bin/retroarch-archive-homebrew-fetch
$DOTFILES_DIR/scripts/retroarch-archive-homebrew-review.py|$HOME/.local/bin/retroarch-archive-homebrew-review
$DOTFILES_DIR/scripts/retroarch-archive-homebrew-import.py|$HOME/.local/bin/retroarch-archive-homebrew-import
$DOTFILES_DIR/scripts/retroarch-archive-homebrew-playlists.py|$HOME/.local/bin/retroarch-archive-homebrew-playlists
$DOTFILES_DIR/scripts/retroarch-archive-homebrew-sync.py|$HOME/.local/bin/retroarch-archive-homebrew-sync
$DOTFILES_DIR/scripts/retroarch-workstation-audit.py|$HOME/.local/bin/retroarch-workstation-audit
$DOTFILES_DIR/scripts/retroarch-install-workstation-cores.sh|$HOME/.local/bin/retroarch-install-workstation-cores
$DOTFILES_DIR/scripts/retroarch-bios-apply.py|$HOME/.local/bin/retroarch-bios-apply
$DOTFILES_DIR/scripts/retroarch-apply-network-cmd.py|$HOME/.local/bin/retroarch-apply-network-cmd
$DOTFILES_DIR/scripts/retroarch-build-libretro-cores.sh|$HOME/.local/bin/retroarch-build-libretro-cores
$DOTFILES_DIR/scripts/retroarch-install-widescreen-cores.sh|$HOME/.local/bin/retroarch-install-widescreen-cores
$DOTFILES_DIR/scripts/retroarch-dolphin-sync-sys.sh|$HOME/.local/bin/retroarch-dolphin-sync-sys
$DOTFILES_DIR/scripts/retroarch-next-widescreen-setup.sh|$HOME/.local/bin/retroarch-next-widescreen-setup
$DOTFILES_DIR/scripts/retroarch-flycast-apply-widescreen-defaults.py|$HOME/.local/bin/retroarch-flycast-apply-widescreen-defaults
$DOTFILES_DIR/scripts/retroarch-dolphin-apply-widescreen-defaults.py|$HOME/.local/bin/retroarch-dolphin-apply-widescreen-defaults
$DOTFILES_DIR/scripts/retroarch-flycast-widescreen-audit.py|$HOME/.local/bin/retroarch-flycast-widescreen-audit
$DOTFILES_DIR/scripts/retroarch-dolphin-widescreen-audit.py|$HOME/.local/bin/retroarch-dolphin-widescreen-audit
$DOTFILES_DIR/hyprland/hyprshade.toml|$HOME/.config/hyprshade/config.toml
$DOTFILES_DIR/wluma/config.toml|$HOME/.config/wluma/config.toml
$DOTFILES_DIR/kanshi/config|$HOME/.config/kanshi/config
$DOTFILES_DIR/glshell|$HOME/.config/glshell
EOF
}

print_linux_systemd_link_specs() {
    local src rel base
    while IFS= read -r -d '' src; do
        rel="${src#"$DOTFILES_DIR/systemd/"}"
        base="$(basename "$src")"
        printf '%s|%s\n' "$src" "$HOME/.config/systemd/user/$rel"
    done < <(find "$DOTFILES_DIR/systemd" -type f -print0 | sort -z)
}

print_linux_systemd_copy_specs() {
    :
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

print_linux_writable_config_dirs() {
    cat <<EOF
$HOME/.config/hyprshell|hyprshell writable config dir
EOF
}

print_linux_runtime_specs() {
    cat <<EOF
$HOME/.local/state/hypr/monitors.dynamic.conf|generated Hyprland monitor include|file|required
$HOME/.local/state/kitty/sessions|kitty session state directory|dir|required
$HOME/.local/state/dotfiles/desktop-control|desktop control state root|dir|optional
$HOME/.local/state/dotfiles/desktop-control/notifications/history.jsonl|desktop notification history log|file|optional
EOF
}

linux_live_wayland_session_ready() {
    [[ -n "${XDG_RUNTIME_DIR:-}" ]] || return 1
    [[ -n "${WAYLAND_DISPLAY:-}" ]] || return 1
    [[ -S "${XDG_RUNTIME_DIR%/}/${WAYLAND_DISPLAY}" ]]
}

refresh_linux_desktop_user_environment() {
    local env_vars=()
    local var
    for var in XDG_RUNTIME_DIR WAYLAND_DISPLAY HYPRLAND_INSTANCE_SIGNATURE DBUS_SESSION_BUS_ADDRESS; do
        [[ -n "${!var:-}" ]] && env_vars+=("$var")
    done
    if [[ "${#env_vars[@]}" -eq 0 ]]; then
        return 0
    fi
    systemctl --user import-environment "${env_vars[@]}" >/dev/null 2>&1 || true
    if command -v dbus-update-activation-environment >/dev/null 2>&1; then
        dbus-update-activation-environment --systemd "${env_vars[@]}" >/dev/null 2>&1 || true
    fi
}

start_linux_desktop_units_if_live() {
    local service_units=("$@")
    local passive_units=(rg-status-bar.timer)

    if ! linux_live_wayland_session_ready; then
        log_info "Skipping live desktop unit start; no active Wayland session detected in current shell"
        return 0
    fi

    refresh_linux_desktop_user_environment

    if [[ "${#passive_units[@]}" -gt 0 ]]; then
        if systemctl --user start "${passive_units[@]}" >/dev/null 2>&1; then
            log_success "Started passive desktop units in live session"
        else
            log_warn "Failed to start one or more passive desktop units in live session"
        fi
    fi

    if [[ "${#service_units[@]}" -gt 0 ]]; then
        if systemctl --user reload-or-restart "${service_units[@]}" >/dev/null 2>&1; then
            log_success "Reloaded desktop services for current live session"
        else
            log_warn "Failed to reload one or more desktop services; run 'systemctl --user reload-or-restart ${service_units[*]}' inside your desktop session"
        fi
    fi
}

resolve_physical_path() {
    local path="$1"
    if command -v realpath >/dev/null 2>&1; then
        realpath "$path" 2>/dev/null
        return $?
    fi
    perl -MCwd=abs_path -e '
        my $resolved = abs_path(shift);
        exit 1 unless defined $resolved;
        print "$resolved\n";
    ' "$path" 2>/dev/null
}

path_relative_to_root() {
    local root="$1" path="$2"
    case "$path" in
        "$root") printf '.\n' ;;
        "$root"/*) printf '%s\n' "${path#$root/}" ;;
        *) return 1 ;;
    esac
}

same_repo_managed_target() {
    local src="$1" dst="$2"
    [[ -L "$dst" ]] || return 1
    [[ -n "$DOTFILES_GIT_COMMON_DIR" ]] || return 1

    local src_abs dst_abs src_rel dst_repo_root dst_rel dst_common_raw dst_common
    src_abs="$(resolve_physical_path "$src")" || return 1
    dst_abs="$(resolve_physical_path "$dst")" || return 1
    src_rel="$(path_relative_to_root "$DOTFILES_TOPLEVEL" "$src_abs")" || return 1
    dst_repo_root="$(git -C "$(dirname "$dst_abs")" rev-parse --show-toplevel 2>/dev/null)" || return 1
    dst_common_raw="$(git -C "$dst_repo_root" rev-parse --path-format=absolute --git-common-dir 2>/dev/null)" || return 1
    dst_common="$(cd "$dst_repo_root" && cd "$dst_common_raw" 2>/dev/null && pwd -P)" || return 1
    [[ "$dst_common" == "$DOTFILES_GIT_COMMON_DIR" ]] || return 1
    dst_rel="$(path_relative_to_root "$dst_repo_root" "$dst_abs")" || return 1
    [[ "$src_rel" == "$dst_rel" ]]
}

linux_dir_allows_real_path() {
    local path="$1"
    case "$path" in
        "$HOME/.config/hypr-dock"|\
        "$HOME/.config/hyprdynamicmonitors"|\
        "$HOME/.config/hyprland-autoname-workspaces")
            return 0
            ;;
    esac
    return 1
}

path_is_package_managed() {
    local path="$1"
    command -v pacman >/dev/null 2>&1 || return 1

    if pacman -Qo "$path" >/dev/null 2>&1; then
        return 0
    fi
    [[ -d "$path" ]] || return 1

    local child
    while IFS= read -r child; do
        pacman -Qo "$child" >/dev/null 2>&1 && return 0
    done < <(find "$path" -maxdepth 2 -mindepth 1 2>/dev/null | head -n 64)
    return 1
}

linux_path_owner_details() {
    local path="$1"
    if stat -Lc 'owner=%U group=%G mode=%a' "$path" >/dev/null 2>&1; then
        stat -Lc 'owner=%U group=%G mode=%a' "$path"
        return 0
    fi
    printf 'owner=unknown group=unknown mode=unknown\n'
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
        "auto-notify|https://github.com/MichaelAquilina/zsh-auto-notify.git"
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

# ── Tattoy shader symlink ──
setup_tattoy_shaders() {
    log_info "Tattoy shaders skipped (kitty uses DarkWindow pipeline)"
}

deploy_linux_etc_configs() {
    if [[ "$OS" != "Linux" ]]; then
        return 0
    fi

    local script="$DOTFILES_DIR/scripts/etc-deploy.sh"
    if [[ ! -x "$script" ]]; then
        log_warn "Skipping /etc deploy — missing executable: $script"
        return 0
    fi

    log_info "Deploying tracked /etc configs..."
    if "$script"; then
        log_success "Tracked /etc configs deployed"
    else
        log_warn "Tracked /etc deploy reported an error — rerun: $script"
    fi
}

normalize_writable_config_dir() {
    local dir="$1"
    if [[ -L "$dir" ]]; then
        backup_file "$dir"
    fi
    mkdir -p "$dir"
}

sync_visual_theme() {
    if [[ "$OS" != "Linux" ]]; then
        return 0
    fi
    local script="$DOTFILES_DIR/scripts/theme-sync.sh"
    if [[ ! -x "$script" ]]; then
        log_warn "Skipping theme sync — missing executable: $script"
        return 0
    fi
    log_info "Syncing generated theme assets..."
    if "$script" --quiet; then
        log_success "Theme assets synced"
    else
        log_warn "Theme sync reported an error — rerun: $script"
    fi
}

bootstrap_hyprland_plugins() {
    if [[ "$OS" != "Linux" ]]; then
        return 0
    fi
    local script="$DOTFILES_DIR/scripts/hyprpm-bootstrap.sh"
    if [[ ! -x "$script" ]]; then
        log_warn "Skipping hyprpm bootstrap — missing executable: $script"
        return 0
    fi
    log_info "Bootstrapping Hyprland plugins..."
    if "$script" --quiet; then
        log_success "Hyprland plugins bootstrapped"
    else
        log_warn "Hyprland plugin bootstrap reported an error — rerun: $script"
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
        normalize_writable_config_dir "$HOME/.config/hyprshell"
    fi

    local src dst
    while IFS='|' read -r src dst; do
        [[ -n "$src" ]] || continue
        link_if_present "$src" "$dst"
    done < <(print_link_specs)

    if [[ "$OS" == "Linux" ]]; then
        while IFS='|' read -r src dst; do
            [[ -n "$src" ]] || continue
            copy_file_if_present "$src" "$dst"
        done < <(print_linux_systemd_copy_specs)
    fi

    # Platform-specific service management
    if [[ "$OS" == "Darwin" ]]; then
        :
    elif [[ "$OS" == "Linux" ]]; then
        log_info "Installing systemd user services..."
        local retired_units=(
            waybar.service
            foot-server.socket
            foot-server.service
        )
        local obsolete_audio_compat_units=(
            pulseaudio.service
            pulseaudio.socket
            pipewire-media-session.service
        )
        systemctl --user disable --now "${retired_units[@]}" >/dev/null 2>&1 || true
        systemctl --user mask "${obsolete_audio_compat_units[@]}" >/dev/null 2>&1 || true
        rm -f "$HOME/.config/systemd/user/foot-server.socket"
        systemctl --user daemon-reload

        mkdir -p "$HOME/.local/state/hypr"
        if [[ ! -f "$HOME/.local/state/hypr/monitors.dynamic.conf" ]]; then
            printf '# Generated by hyprdynamicmonitors.\n' > "$HOME/.local/state/hypr/monitors.dynamic.conf"
        fi
        mkdir -p "$HOME/.local/state/kitty/sessions"

        local desktop_service_units=(
            ironbar.service
            dotfiles-hyprshell.service
            dotfiles-hypr-dock.service
            dotfiles-hyprdynamicmonitors.service
            dotfiles-hyprland-autoname-workspaces.service
            dotfiles-notification-history.service
            dotfiles-keybind-ticker.service
            dotfiles-keybind-ticker@DP-3_focus.service
            dotfiles-ticker-lockwatch.service
            dotfiles-ticker-recordwatch.service
            dotfiles-lyrics-ticker.service
            dotfiles-hypr-monitor-watch.service
            dotfiles-window-label.service
            dotfiles-fleet-sparkline.service
            dotfiles-cliphist.service
        )
        local desktop_passive_units=(
            rg-status-bar.timer
            bar-gpu.timer
            bar-updates.timer
            bar-mx.timer
            bar-weather.timer
            bar-ci.timer
            bar-calendar.timer
            bar-tokens.timer
            bar-dirty.timer
            bar-archnews.timer
            bar-smart.timer
            bar-hn.timer
            bar-prs.timer
            bar-weather-alerts.timer
            bar-cve.timer
        )
        local desktop_units=("${desktop_service_units[@]}" "${desktop_passive_units[@]}")
        if systemctl --user enable "${desktop_units[@]}" >/dev/null 2>&1; then
            log_success "Enabled desktop systemd user units"
        else
            log_warn "Failed to enable one or more desktop user units; rerun 'systemctl --user enable ${desktop_units[*]}' inside your desktop session"
        fi
        start_linux_desktop_units_if_live "${desktop_service_units[@]}"

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
        elif same_repo_managed_target "$src" "$dst"; then
            log_success "OK (same repo): $dst"
        elif [[ -L "$dst" ]]; then
            log_error "Wrong target: $dst -> $(readlink "$dst") (expected $src)"
            errors=$((errors + 1))
        elif [[ -d "$dst" ]] && linux_dir_allows_real_path "$dst" && path_is_package_managed "$dst"; then
            log_success "OK (package-managed dir): $dst"
        elif [[ -e "$dst" ]]; then
            log_warn "Exists but not a symlink: $dst"
            errors=$((errors + 1))
        else
            log_error "Missing: $dst"
            errors=$((errors + 1))
        fi
    }

    check_linux_writable_dirs() {
        local path label details
        log_info "Checking writable config roots..."
        while IFS='|' read -r path label; do
            [[ -n "$path" ]] || continue
            if [[ -L "$path" ]]; then
                log_error "Expected writable dir, found symlink: $path -> $(readlink "$path")"
                errors=$((errors + 1))
            elif [[ -d "$path" ]] && [[ -w "$path" ]] && [[ -x "$path" ]]; then
                log_success "OK ($label): $path"
            elif [[ -d "$path" ]]; then
                details="$(linux_path_owner_details "$path")"
                log_error "Writable dir not writable by current user ($label): $path [$details]"
                errors=$((errors + 1))
            else
                log_warn "Missing writable dir: $path"
            fi
        done < <(print_linux_writable_config_dirs)
    }

    check_linux_runtime_paths() {
        local path label kind requirement
        log_info "Checking generated runtime state..."
        while IFS='|' read -r path label kind requirement; do
            [[ -n "$path" ]] || continue
            if [[ "$kind" == "dir" && -d "$path" ]] || [[ "$kind" == "file" && -f "$path" ]]; then
                log_success "OK ($label): $path"
            elif [[ "$requirement" == "required" ]]; then
                log_warn "Missing generated runtime $kind: $path"
            else
                log_info "Generated runtime $kind not created yet: $path"
            fi
        done < <(print_linux_runtime_specs)
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

    check_copied_file() {
        local src="$1" dst="$2"
        if [[ -f "$dst" ]] && cmp -s "$src" "$dst"; then
            log_success "OK (copied): $dst"
        elif [[ -L "$dst" ]]; then
            log_error "Expected copied file, found symlink: $dst -> $(readlink "$dst")"
            errors=$((errors + 1))
        elif [[ -e "$dst" ]]; then
            log_error "Copied file drift: $dst"
            errors=$((errors + 1))
        else
            log_error "Missing copied file: $dst"
            errors=$((errors + 1))
        fi
    }

    check_copied_file_if_present() {
        local src="$1" dst="$2"
        if source_exists "$src"; then
            check_copied_file "$src" "$dst"
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
        while IFS='|' read -r src dst; do
            [[ -n "$src" ]] || continue
            check_copied_file_if_present "$src" "$dst"
        done < <(print_linux_systemd_copy_specs)
    fi

    if [[ "$OS" == "Linux" ]]; then
        log_info "Checking systemd user services..."
        check_linux_writable_dirs
        check_linux_runtime_paths
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

    if [[ "$OS" == "Linux" ]]; then
        log_phase "SYSTEM SETUP"
        deploy_linux_etc_configs
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
        deploy_linux_etc_configs
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
    sync_visual_theme
    bootstrap_hyprland_plugins
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
