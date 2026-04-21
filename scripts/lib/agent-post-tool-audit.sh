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
  */swaync/*)               config_reload_service swaync   2>/dev/null || errors+=("[reload] swaync reload failed") ;;
  */ironbar/*)              config_reload_service ironbar  2>/dev/null || errors+=("[reload] ironbar reload failed") ;;
  */tmux/*)                 config_reload_service tmux     2>/dev/null || errors+=("[reload] tmux reload failed") ;;
  */metapac/*)
    # Validate TOML and check for unmanaged packages
    if command -v metapac &>/dev/null; then
      unmanaged="$(metapac unmanaged 2>&1)"
      if ! echo "$unmanaged" | grep -q "no unmanaged packages"; then
        errors+=("[metapac] Unmanaged packages detected after edit — run: metapac sync")
      fi
    fi
    ;;
esac

# Brief settle for reload (100ms — down from 500ms)
sleep 0.1

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

  */ironbar/*)
    if command -v ironbar &>/dev/null; then
      if ! ironbar ping >/dev/null 2>&1; then
        errors+=("[ironbar] ironbar is not responding after reload")
      fi
    fi
    if command -v systemctl &>/dev/null && systemctl --user --quiet is-failed ironbar.service; then
      errors+=("[ironbar] ironbar.service is failed")
      while IFS= read -r line; do
        [[ -n "$line" ]] || continue
        errors+=("  $line")
      done < <(journalctl --user -u ironbar.service -n 10 --no-pager 2>/dev/null || true)
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

# ── Phase 3: Theme check (non-blocking, warnings only) ──
palette_warnings=()
case "$file_path" in
  *.conf | *.toml | *.scss | *.yuck | *.ini | *.css | *.json)
    if [[ -f "$file_path" ]]; then
      for ref in \
        "Hack Nerd Font|legacy UI font" \
        "Matcha-dark-sea|legacy GTK theme" \
        "JetBrains Mono|legacy font"; do
        IFS='|' read -r pattern label <<<"$ref"
        if grep -qi "$pattern" "$file_path" 2>/dev/null; then
          palette_warnings+=("[theme] ${label} reference in $file_path")
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
