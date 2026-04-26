#!/usr/bin/env bash
# palette-propagate.sh — Render the Hairglasses Neon palette to every consumer.
#
# Reads theme/palette.env, envsubst-renders every template in matugen/templates/,
# writes to each consumer's real config path, then fires per-app reload hooks.
#
# Usage:
#   palette-propagate.sh               # fixed-mode render (envsubst from palette.env)
#   palette-propagate.sh --wallpaper   # also run matugen for wallpaper-derived accents
#   palette-propagate.sh --dry-run     # preview targets without writing
#   palette-propagate.sh --no-reload   # render but skip post-hooks

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/lib/shell-stack.sh"

DRY_RUN=false
DO_RELOAD=true
USE_WALLPAPER=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run)    DRY_RUN=true ;;
        --no-reload)  DO_RELOAD=false ;;
        --wallpaper)  USE_WALLPAPER=true ;;
        -h|--help)
            sed -n '3,11p' "$0" | sed 's/^# \{0,1\}//'
            exit 0 ;;
        *)
            echo "palette-propagate.sh: unknown option: $1" >&2
            exit 2 ;;
    esac
    shift
done

# shellcheck source=../theme/palette.env
source "$DOTFILES_DIR/theme/palette.env"
shell_stack_load
export THEME_NAME THEME_MODE THEME_UI_FONT THEME_CODE_FONT THEME_ICON_FONT
export THEME_BG THEME_SURFACE THEME_SURFACE_ALT THEME_PANEL THEME_PANEL_STRONG
export THEME_BORDER THEME_BORDER_STRONG THEME_FG THEME_MUTED
export THEME_PRIMARY THEME_SECONDARY THEME_TERTIARY
export THEME_WARNING THEME_DANGER THEME_BLUE
export THEME_RADIUS_SM THEME_RADIUS_MD THEME_RADIUS_LG

_c_info() { printf '\033[38;2;41;240;255m[palette]\033[0m %s\n' "$1"; }
_c_ok()   { printf '\033[38;2;61;255;181m[palette]\033[0m %s\n' "$1"; }
_c_warn() { printf '\033[38;2;255;228;94m[palette]\033[0m %s\n' "$1" >&2; }

TEMPLATES_DIR="$DOTFILES_DIR/matugen/templates"
[[ -d "$TEMPLATES_DIR" ]] || { _c_warn "Templates dir missing: $TEMPLATES_DIR"; exit 1; }

_c_info "Palette: $THEME_NAME (#$THEME_PRIMARY / #$THEME_SECONDARY / #$THEME_TERTIARY)"

# ── Optional wallpaper-derived accent override ────────────────────
if $USE_WALLPAPER; then
    wall_state="${XDG_STATE_HOME:-$HOME/.local/state}/swww/current"
    wall_path=""
    if [[ -f "$wall_state" ]]; then
        wall_path="$(head -1 "$wall_state" 2>/dev/null || true)"
    fi
    if [[ -n "$wall_path" && -f "$wall_path" ]] && command -v matugen >/dev/null && command -v jq >/dev/null; then
        _c_info "Extracting accents from: $wall_path"
        if palette_json="$(matugen color -j hex "$wall_path" 2>/dev/null)" && [[ -n "$palette_json" ]]; then
            p_primary="$(echo "$palette_json" | jq -r '.colors.dark.primary // empty' | sed 's/^#//')"
            p_secondary="$(echo "$palette_json" | jq -r '.colors.dark.secondary // empty' | sed 's/^#//')"
            p_tertiary="$(echo "$palette_json" | jq -r '.colors.dark.tertiary // empty' | sed 's/^#//')"
            [[ -n "$p_primary"   ]] && export THEME_PRIMARY="$p_primary"
            [[ -n "$p_secondary" ]] && export THEME_SECONDARY="$p_secondary"
            [[ -n "$p_tertiary"  ]] && export THEME_TERTIARY="$p_tertiary"
            _c_info "Accents overridden: #$THEME_PRIMARY / #$THEME_SECONDARY / #$THEME_TERTIARY"
        else
            _c_warn "matugen extract failed; keeping fixed palette"
        fi
    else
        _c_warn "Wallpaper extract requested but wallpaper/matugen/jq missing"
    fi
fi

# ── Render helper ──────────────────────────────────────────────────
# Substitute only known THEME_* vars to avoid expanding unrelated shell vars
# that might appear in a template.
_render() {
    local template="$1" target="$2"
    if $DRY_RUN; then
        printf '  [dry] %-68s → %s\n' "${template##*/}" "$target"
        return 0
    fi
    mkdir -p "$(dirname "$target")"
    envsubst \
        '$THEME_BG $THEME_SURFACE $THEME_SURFACE_ALT $THEME_PANEL $THEME_PANEL_STRONG
         $THEME_BORDER $THEME_BORDER_STRONG $THEME_FG $THEME_MUTED
         $THEME_PRIMARY $THEME_SECONDARY $THEME_TERTIARY
         $THEME_WARNING $THEME_DANGER $THEME_BLUE' \
        < "$template" > "$target"
    printf '  [write] %-66s → %s\n' "${template##*/}" "$target"
}

# ── Consumer registry ──────────────────────────────────────────────
# Each entry: <template> <target-path> <post-hook>
# Post-hooks are shell commands executed after write (use "" to skip).

_handle() {
    local template="$1" target="$2" hook="${3:-}"
    [[ -f "$TEMPLATES_DIR/$template" ]] || { _c_warn "Missing template: $template"; return 0; }
    _render "$TEMPLATES_DIR/$template" "$target"
    if $DO_RELOAD && [[ -n "$hook" ]] && ! $DRY_RUN; then
        # shellcheck disable=SC2086
        eval "$hook" >/dev/null 2>&1 || _c_warn "post-hook failed: $hook"
    fi
}

# GTK CSS consumers — share the same template
for dir in swaync wofi wlogout; do
    _handle gtk-colors.css "$HOME/.config/$dir/theme.generated.css" ""
done

# Quickshell consumes generated QML tokens instead of GTK CSS. Render into the
# repo tree so both direct `--path` runs and installed ~/.config symlinks see it.
_handle qt-colors.qml.template "$DOTFILES_DIR/quickshell/styles/Colors.qml" ""

# Toolkit theming — libadwaita @define-color overrides for GTK3/4 apps and
# qt5ct/qt6ct color scheme INI for Qt apps. The static gtk.css files in
# ~/.config/gtk-{3,4}.0/ @import the colors.css rendered here.
_handle gtk-libadwaita-overrides.css "$HOME/.config/gtk-4.0/colors.css" ''
_handle gtk-libadwaita-overrides.css "$HOME/.config/gtk-3.0/colors.css" ''
_handle qt-colorscheme.conf          "$HOME/.config/qt5ct/colors/hairglasses.conf" ''
_handle qt-colorscheme.conf          "$HOME/.config/qt6ct/colors/hairglasses.conf" ''

# Terminals + visual apps
_handle kitty-colors.conf    "$HOME/.config/kitty/cyberpunk-neon.conf"                   'pkill -SIGUSR1 kitty || true'
_handle hyprland-colors.conf "$HOME/.config/hypr/colors.conf"                             ''
_handle hyprlock-colors.conf "$HOME/.config/hypr/hyprlock-colors.conf"                    ''
_handle btop-theme.theme     "$HOME/.config/btop/themes/hairglasses-neon.theme"           ''
_handle yazi-theme.toml      "$HOME/.config/yazi/theme.toml"                              ''
_handle zsh-fzf-colors.sh    "$HOME/.config/fzf/hg-colors.sh"                             ''

# Cava ini format doesn't support includes — render the whole config.
# Audio settings are stable; change in the template if you need to adjust them.
_handle cava-colors.conf     "$HOME/.config/cava/config"                                  'pkill -USR2 cava 2>/dev/null || true'

# ── Reload hooks (after all targets are written) ─────────────────
# Quickshell owns every desktop surface (bar, ticker, dock, menus,
# notifications, companion overlays). One restart picks up the new
# palette across the whole stack. swaync gets a config refresh only
# when the user is intentionally on rollback (Quickshell stopped).
if $DO_RELOAD && ! $DRY_RUN; then
    if shell_stack_quickshell_wanted || systemctl --user is-active dotfiles-quickshell.service >/dev/null 2>&1; then
        systemctl --user restart dotfiles-quickshell.service 2>/dev/null || true
    else
        swaync-client -rs 2>/dev/null || true
    fi
fi

$DRY_RUN && _c_info "Dry-run complete — no files written" || _c_ok "Palette propagation complete"
