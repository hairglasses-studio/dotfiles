#!/usr/bin/env bash
# hypr-config-snapshot.sh — capture the current ~/.config/hypr/ config to a
# timestamped directory under $XDG_STATE_HOME/dotfiles/desktop-control/hypr/
# config-snapshots/. Shell-side companion to the hypr_config_snapshot MCP tool,
# suitable for invocation from PreToolUse hooks where firing the MCP server is
# overkill. Produces the same directory layout + meta.json shape so both paths
# are interchangeable from the rollback perspective.
#
# Usage:
#   hypr-config-snapshot.sh [label]
#
# label defaults to "auto". Exit 0 on success, non-zero on write failure.
# Best-effort on metadata capture — missing git / nvidia / hyprctl produce
# empty fields rather than failing the whole snapshot.

set -euo pipefail

label="${1:-auto}"
# sanitize: allow only [A-Za-z0-9._-], collapse runs of other chars to single dash
sanitized="$(printf '%s' "$label" | tr -c 'A-Za-z0-9._-' '-' | sed -E 's/-+/-/g; s/^-+|-+$//g')"
: "${sanitized:=auto}"

xdg_state="${XDG_STATE_HOME:-$HOME/.local/state}"
root="$xdg_state/dotfiles/desktop-control/hypr/config-snapshots"
mkdir -p "$root"

now_utc="$(date -u +'%Y%m%d-%H%M%S')"
now_rfc="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
dir_name="${now_utc}_${sanitized}"
snap_dir="$root/$dir_name"
mkdir -p "$snap_dir"

live="$HOME/.config/hypr"
# Files of interest — matches configSnapshotFiles in hypr_config_snapshot.go.
files=(
  hyprland.conf
  monitors.conf
  local.conf
  colors.conf
  darkwindow-shaders.conf
  plugin-binds.conf
  hyprshade.toml
  wallpaper.env
)

captured=()
for f in "${files[@]}"; do
  src="$live/$f"
  if [[ -f "$src" ]]; then
    # Use install so we get an atomic-ish write (creates a temp then renames).
    install -m 0644 "$src" "$snap_dir/$f"
    captured+=("$f")
  fi
done

# Best-effort metadata capture.
git_sha=""
if command -v git >/dev/null 2>&1; then
  git_sha="$(git -C "$HOME/hairglasses-studio/dotfiles" rev-parse --short HEAD 2>/dev/null || true)"
fi

kernel=""
if [[ -r /proc/sys/kernel/osrelease ]]; then
  kernel="$(tr -d '\n' < /proc/sys/kernel/osrelease)"
fi

driver=""
if [[ -r /proc/driver/nvidia/version ]]; then
  driver="$(head -n1 /proc/driver/nvidia/version 2>/dev/null \
              | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -n1 || true)"
fi

hypr_version=""
if command -v hyprctl >/dev/null 2>&1; then
  hypr_version="$(hyprctl version 2>/dev/null | head -n1 | tr -d '\r' || true)"
fi

# Emit meta.json matching the Go struct layout (field names + JSON tags).
# jq is present on the workstation; if it ever disappears, fall back to a
# minimal hand-rolled JSON.
if command -v jq >/dev/null 2>&1; then
  jq -n \
    --arg name "$dir_name" \
    --arg label "$sanitized" \
    --arg path "$snap_dir" \
    --arg saved_at "$now_rfc" \
    --arg git_sha "$git_sha" \
    --arg kernel "$kernel" \
    --arg driver "$driver" \
    --arg hypr_version "$hypr_version" \
    --argjson files "$(printf '%s\n' "${captured[@]}" | jq -R . | jq -s .)" \
    '{name:$name, label:$label, path:$path, saved_at:$saved_at, files:$files,
      git_sha:$git_sha, kernel:$kernel, driver:$driver, hypr_version:$hypr_version}' \
    > "$snap_dir/meta.json"
else
  printf '{"name":"%s","label":"%s","path":"%s","saved_at":"%s"}\n' \
    "$dir_name" "$sanitized" "$snap_dir" "$now_rfc" > "$snap_dir/meta.json"
fi

printf '%s\n' "$snap_dir"
