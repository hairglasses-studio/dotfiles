#!/usr/bin/env bash
# theme-sync.sh — Apply Hairglasses Neon runtime theme preferences.
#
# Delegates palette propagation to scripts/palette-propagate.sh (which renders
# all template-backed consumers). This script additionally applies runtime
# theme state that cannot be expressed as a file: gsettings, Qt env vars,
# Plasma color/desktop theme when running under KDE, and dbus/systemd
# environment activation for GUI apps.
#
# Usage:
#   theme-sync.sh                # render palette + apply runtime preferences
#   theme-sync.sh --quiet        # silent operation
#   theme-sync.sh --no-wallpaper # skip optional wallpaper-derived accent extraction
#   theme-sync.sh --wallpaper    # force wallpaper extraction via matugen

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/hg-core.sh
source "$SCRIPT_DIR/lib/hg-core.sh"

QUIET=false
USE_WALLPAPER=""   # unset: default mode; true/false: explicit
while [[ $# -gt 0 ]]; do
    case "$1" in
        --quiet)         QUIET=true ;;
        --no-wallpaper)  USE_WALLPAPER=false ;;
        --wallpaper)     USE_WALLPAPER=true ;;
        *)
            hg_die "Usage: theme-sync.sh [--quiet] [--wallpaper|--no-wallpaper]" ;;
    esac
    shift
done

# ── Step 1: render palette via palette-propagate.sh ───────────────
propagate_args=()
$QUIET && propagate_args+=(--no-reload)
[[ "$USE_WALLPAPER" == true ]] && propagate_args+=(--wallpaper)

if [[ -x "$SCRIPT_DIR/palette-propagate.sh" ]]; then
    if $QUIET; then
        "$SCRIPT_DIR/palette-propagate.sh" "${propagate_args[@]}" >/dev/null 2>&1 || true
    else
        "$SCRIPT_DIR/palette-propagate.sh" "${propagate_args[@]}"
    fi
else
    $QUIET || hg_warn "palette-propagate.sh missing — skipping template render"
fi

# ── Step 2: source palette for any runtime-level theme state ──────
PALETTE_FILE="$HG_DOTFILES/theme/palette.env"
[[ -f "$PALETTE_FILE" ]] || hg_die "Missing palette file: $PALETTE_FILE"
# shellcheck disable=SC1090
source "$PALETTE_FILE"

# ── Step 3: runtime theme preferences (gsettings, Qt, Plasma) ─────
apply_runtime_theme_preferences() {
    local desktop_marker should_apply_plasma=false

    export QT_QPA_PLATFORMTHEME="${QT_QPA_PLATFORMTHEME:-qt6ct}"
    export QT_QUICK_CONTROLS_MATERIAL_THEME="${QT_QUICK_CONTROLS_MATERIAL_THEME:-Dark}"
    export QT_QUICK_CONTROLS_UNIVERSAL_THEME="${QT_QUICK_CONTROLS_UNIVERSAL_THEME:-Dark}"
    unset QT_STYLE_OVERRIDE || true
    desktop_marker="$(printf '%s:%s' "${XDG_CURRENT_DESKTOP:-}" "${DESKTOP_SESSION:-}" | tr '[:upper:]' '[:lower:]')"

    if command -v gsettings >/dev/null 2>&1; then
        gsettings set org.gnome.desktop.interface gtk-theme "adw-gtk3-dark" >/dev/null 2>&1 || true
        gsettings set org.gnome.desktop.interface icon-theme "${THEME_ICON_THEME}" >/dev/null 2>&1 || true
        gsettings set org.gnome.desktop.interface cursor-theme "${THEME_CURSOR_THEME}" >/dev/null 2>&1 || true
        gsettings set org.gnome.desktop.interface color-scheme "prefer-dark" >/dev/null 2>&1 || true
    fi

    if command -v xfconf-query >/dev/null 2>&1; then
        xfconf-query -c xsettings -p /Net/ThemeName -n -t string -s "adw-gtk3-dark" >/dev/null 2>&1 || true
        xfconf-query -c xsettings -p /Net/IconThemeName -n -t string -s "${THEME_ICON_THEME}" >/dev/null 2>&1 || true
        xfconf-query -c xsettings -p /Gtk/CursorThemeName -n -t string -s "${THEME_CURSOR_THEME}" >/dev/null 2>&1 || true
    fi

    if [[ -n "${DBUS_SESSION_BUS_ADDRESS:-}" \
          && ( -n "${WAYLAND_DISPLAY:-}" || -n "${DISPLAY:-}" ) \
          && ( "$desktop_marker" == *plasma* || "$desktop_marker" == *kde* ) ]]; then
        should_apply_plasma=true
    fi

    if $should_apply_plasma && command -v plasma-apply-colorscheme >/dev/null 2>&1; then
        plasma-apply-colorscheme BreezeDark >/dev/null 2>&1 || true
    fi
    if $should_apply_plasma && command -v plasma-apply-desktoptheme >/dev/null 2>&1; then
        plasma-apply-desktoptheme breeze-dark >/dev/null 2>&1 || true
    fi

    if command -v systemctl >/dev/null 2>&1; then
        systemctl --user unset-environment QT_STYLE_OVERRIDE >/dev/null 2>&1 || true
        systemctl --user import-environment DISPLAY WAYLAND_DISPLAY XDG_CURRENT_DESKTOP QT_QPA_PLATFORMTHEME QT_QUICK_CONTROLS_MATERIAL_THEME QT_QUICK_CONTROLS_UNIVERSAL_THEME >/dev/null 2>&1 || true
    fi
    if command -v dbus-update-activation-environment >/dev/null 2>&1; then
        dbus-update-activation-environment --systemd DISPLAY WAYLAND_DISPLAY XDG_CURRENT_DESKTOP QT_QPA_PLATFORMTHEME QT_QUICK_CONTROLS_MATERIAL_THEME QT_QUICK_CONTROLS_UNIVERSAL_THEME >/dev/null 2>&1 || true
    fi
}

apply_runtime_theme_preferences

if ! $QUIET; then
    hg_ok "Theme synced: ${THEME_NAME}"
fi
