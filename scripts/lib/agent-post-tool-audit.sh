#!/usr/bin/env bash
# agent-post-tool-audit.sh — Post-edit reload + error audit hook.
# Reads JSON from stdin, reloads affected services, then checks all relevant
# error sources. Exits non-zero with stderr report if errors found — this
# makes the feedback visible to Claude via the hook mechanism.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh" 2>/dev/null || exit 0

# ── Extract file path from tool JSON ──────────────────
extract_file_path() {
  local input="$1"
  local match
  match="$(
    printf '%s' "$input" |
      tr '\n' ' ' |
      grep -oE '"(file_path|path)"[[:space:]]*:[[:space:]]*"[^"]*"' |
      head -1 || true
  )"
  [[ -n "$match" ]] || return 1
  printf '%s\n' "$match" | sed -E 's/^"(file_path|path)"[[:space:]]*:[[:space:]]*"//; s/"$//'
}

input="$(cat)"
file_path="$(extract_file_path "$input" || true)"
[[ -n "$file_path" ]] || exit 0

errors=()

# ── Phase 1: Reload affected service ─────────────────
case "$file_path" in
  */hyprland/* | */hypr/*)  config_reload_service hyprland 2>/dev/null || errors+=("[reload] hyprland reload failed") ;;
  */mako/*)                 config_reload_service mako     2>/dev/null || errors+=("[reload] mako reload failed") ;;
  */eww/*)                  config_reload_service eww      2>/dev/null || errors+=("[reload] eww reload failed") ;;
  */waybar/*)               config_reload_service waybar   2>/dev/null || errors+=("[reload] waybar reload failed") ;;
  */tmux/*)                 config_reload_service tmux     2>/dev/null || errors+=("[reload] tmux reload failed") ;;
esac

# Brief settle time for reload to take effect
sleep 0.5

# ── Phase 2: Targeted audit ──────────────────────────
case "$file_path" in
  */hyprland/* | */hypr/*)
    # Check config parse errors
    if command -v hyprctl &>/dev/null; then
      cfg_errors="$(hyprctl -j configerrors 2>/dev/null || echo '[]')"
      # Filter out empty results: [], [""], [" "]
      if [[ "$cfg_errors" != "[]" ]] && ! printf '%s' "$cfg_errors" | grep -qE '^\[\s*""\s*\]$'; then
        # Extract error strings from JSON array
        parsed="$(printf '%s' "$cfg_errors" | tr -d '[]"' | sed 's/,/\n/g' | sed '/^$/d')"
        if [[ -n "$parsed" ]]; then
          while IFS= read -r line; do
            errors+=("[hyprland] $line")
          done <<< "$parsed"
        fi
      fi

      # Check recent log for ERR/WARN (last 20 lines after reload)
      hypr_log="$(find /run/user/"$(id -u)"/hypr/ -name hyprland.log 2>/dev/null | head -1)"
      if [[ -n "$hypr_log" && -f "$hypr_log" ]]; then
        log_errors="$(tail -20 "$hypr_log" | grep -E '\[ERR\]' || true)"
        if [[ -n "$log_errors" ]]; then
          errors+=("[hyprland-log] Recent errors in hyprland.log:")
          while IFS= read -r line; do
            errors+=("  $line")
          done <<< "$log_errors"
        fi
      fi
    fi
    ;;

  */eww/*)
    # Check eww log for errors
    if command -v eww &>/dev/null; then
      eww_errors="$(eww logs 2>&1 | tail -10 | grep -iE 'error|warn|failed' || true)"
      if [[ -n "$eww_errors" ]]; then
        errors+=("[eww] Errors in eww logs:")
        while IFS= read -r line; do
          errors+=("  $line")
        done <<< "$eww_errors"
      fi
    fi
    ;;

  */systemd/*)
    # Check for failed user services after unit file changes
    failed="$(systemctl --user --failed --no-pager 2>&1 | grep -E '●|failed' || true)"
    if [[ -n "$failed" ]]; then
      errors+=("[systemd] Failed user services:")
      while IFS= read -r line; do
        errors+=("  $line")
      done <<< "$failed"
    fi
    ;;
esac

# ── Phase 3: Palette check (non-blocking, warnings only) ──
palette_warnings=()
case "$file_path" in
  *.conf | *.toml | *.scss | *.yuck | *.ini)
    if [[ -f "$file_path" ]]; then
      snazzy_palette="000000 1a1a1a 1a1b26 ff5c57 5af78e f3f99d 57c7ff ff6ac1 9aedfe f1f1f0 686868 eff0eb d4d4d4 264f78 ffffff f8f8f2 282a36"
      hex_colors="$(grep -oiE '#[0-9a-fA-F]{6}' "$file_path" 2>/dev/null | sed 's/#//' | tr '[:upper:]' '[:lower:]' | sort -u || true)"
      for color in $hex_colors; do
        match=false
        for ok in $snazzy_palette; do
          if [[ "$color" == "$ok" ]]; then
            match=true
            break
          fi
        done
        if [[ "$match" == false ]]; then
          palette_warnings+=("[palette] non-Snazzy color #$color in $file_path")
        fi
      done
    fi
    ;;
esac

# ── Output ────────────────────────────────────────────
# Palette warnings go to stderr but don't affect exit code
for warn in "${palette_warnings[@]+"${palette_warnings[@]}"}"; do
  echo "$warn" >&2
done

# Real errors: report and exit non-zero so Claude sees them
if (( ${#errors[@]} > 0 )); then
  for err in "${errors[@]}"; do
    echo "$err" >&2
  done
  exit 1
fi

exit 0
