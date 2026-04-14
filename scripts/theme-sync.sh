#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

QUIET=false
USE_WALLPAPER=true

while [[ $# -gt 0 ]]; do
  case "$1" in
    --quiet) QUIET=true ;;
    --no-wallpaper) USE_WALLPAPER=false ;;
    *)
      hg_die "Usage: theme-sync.sh [--quiet] [--no-wallpaper]"
      ;;
  esac
  shift
done

PALETTE_FILE="$HG_DOTFILES/theme/palette.env"
[[ -f "$PALETTE_FILE" ]] || hg_die "Missing palette file: $PALETTE_FILE"
# shellcheck disable=SC1090
source "$PALETTE_FILE"

CURRENT_WALLPAPER="${XDG_STATE_HOME:-$HOME/.local/state}/swww/current"
WALLPAPER_PATH=""
if $USE_WALLPAPER && [[ -f "$CURRENT_WALLPAPER" ]]; then
  WALLPAPER_PATH="$(head -1 "$CURRENT_WALLPAPER" 2>/dev/null || true)"
fi

theme_primary="$THEME_PRIMARY"
theme_secondary="$THEME_SECONDARY"
theme_tertiary="$THEME_TERTIARY"

if $USE_WALLPAPER && [[ -n "$WALLPAPER_PATH" ]] && [[ -f "$WALLPAPER_PATH" ]] && command -v matugen >/dev/null 2>&1 && command -v jq >/dev/null 2>&1; then
  if palette_json="$(matugen color -j hex "$WALLPAPER_PATH" 2>/dev/null)" && [[ -n "$palette_json" ]]; then
    theme_primary="$(printf '%s' "$palette_json" | jq -r '.colors.dark.primary // empty' | sed 's/^#//')"
    theme_secondary="$(printf '%s' "$palette_json" | jq -r '.colors.dark.secondary // empty' | sed 's/^#//')"
    theme_tertiary="$(printf '%s' "$palette_json" | jq -r '.colors.dark.tertiary // empty' | sed 's/^#//')"
    [[ -n "$theme_primary" ]] || theme_primary="$THEME_PRIMARY"
    [[ -n "$theme_secondary" ]] || theme_secondary="$THEME_SECONDARY"
    [[ -n "$theme_tertiary" ]] || theme_tertiary="$THEME_TERTIARY"
  fi
fi

hex_to_rgb() {
  local hex="${1#\#}"
  printf '%d, %d, %d' "0x${hex:0:2}" "0x${hex:2:2}" "0x${hex:4:2}"
}

hex_css() {
  printf '#%s' "${1#\#}"
}

rgba_css() {
  local rgb
  rgb="$(hex_to_rgb "$1")"
  printf 'rgba(%s, %s)' "$rgb" "$2"
}

mkdir -p \
  "$HOME/.config/hyprshell" \
  "$HOME/.config/swaync" \
  "$HOME/.config/wofi" \
  "$HOME/.config/wlogout"

write_gtk_theme() {
  local target="$1"
  cat > "$target" <<EOF
@define-color theme_bg $(hex_css "$THEME_BG");
@define-color theme_surface $(hex_css "$THEME_SURFACE");
@define-color theme_surface_alt $(hex_css "$THEME_SURFACE_ALT");
@define-color theme_panel $(hex_css "$THEME_PANEL");
@define-color theme_panel_strong $(hex_css "$THEME_PANEL_STRONG");
@define-color theme_fg $(hex_css "$THEME_FG");
@define-color theme_muted $(hex_css "$THEME_MUTED");
@define-color theme_primary $(hex_css "$theme_primary");
@define-color theme_secondary $(hex_css "$theme_secondary");
@define-color theme_tertiary $(hex_css "$theme_tertiary");
@define-color theme_warning $(hex_css "$THEME_WARNING");
@define-color theme_danger $(hex_css "$THEME_DANGER");
@define-color theme_border $(hex_css "$THEME_BORDER");
@define-color theme_border_strong $(hex_css "$THEME_BORDER_STRONG");
EOF
}

for target in \
  "$HOME/.config/hyprshell/theme.generated.css" \
  "$HOME/.config/swaync/theme.generated.css" \
  "$HOME/.config/wofi/theme.generated.css" \
  "$HOME/.config/wlogout/theme.generated.css"
do
  write_gtk_theme "$target"
done

apply_runtime_theme_preferences() {
  local desktop_marker

  export QT_QPA_PLATFORMTHEME="${QT_QPA_PLATFORMTHEME:-qt6ct}"
  export QT_QUICK_CONTROLS_MATERIAL_THEME="${QT_QUICK_CONTROLS_MATERIAL_THEME:-Dark}"
  export QT_QUICK_CONTROLS_UNIVERSAL_THEME="${QT_QUICK_CONTROLS_UNIVERSAL_THEME:-Dark}"
  unset QT_STYLE_OVERRIDE || true
  desktop_marker="$(printf '%s:%s' "${XDG_CURRENT_DESKTOP:-}" "${DESKTOP_SESSION:-}" | tr '[:upper:]' '[:lower:]')"

  if command -v gsettings >/dev/null 2>&1; then
    gsettings set org.gnome.desktop.interface gtk-theme "Adwaita-dark" >/dev/null 2>&1 || true
    gsettings set org.gnome.desktop.interface icon-theme "Papirus-Dark" >/dev/null 2>&1 || true
    gsettings set org.gnome.desktop.interface cursor-theme "Bibata-Modern-Classic" >/dev/null 2>&1 || true
    gsettings set org.gnome.desktop.interface color-scheme "prefer-dark" >/dev/null 2>&1 || true
  fi

  if command -v xfconf-query >/dev/null 2>&1; then
    xfconf-query -c xsettings -p /Net/ThemeName -n -t string -s "Adwaita-dark" >/dev/null 2>&1 || true
    xfconf-query -c xsettings -p /Net/IconThemeName -n -t string -s "Papirus-Dark" >/dev/null 2>&1 || true
    xfconf-query -c xsettings -p /Gtk/CursorThemeName -n -t string -s "Bibata-Modern-Classic" >/dev/null 2>&1 || true
  fi

  if [[ -n "${DBUS_SESSION_BUS_ADDRESS:-}" && ( "$desktop_marker" == *plasma* || "$desktop_marker" == *kde* ) ]] && command -v plasma-apply-colorscheme >/dev/null 2>&1; then
    plasma-apply-colorscheme BreezeDark >/dev/null 2>&1 || true
  fi
  if [[ -n "${DBUS_SESSION_BUS_ADDRESS:-}" && ( "$desktop_marker" == *plasma* || "$desktop_marker" == *kde* ) ]] && command -v plasma-apply-desktoptheme >/dev/null 2>&1; then
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
