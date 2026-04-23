#!/usr/bin/env bash
# mcpkit-drift-check.sh — weekly check that the dotfiles-mcp go.mod pin
# of github.com/hairglasses-studio/mcpkit is at the latest tagged version.
# Emits a `{type: "mcpkit_version_drift"}` event on mismatch so /heal
# picks it up and the user can run dotfiles_mcpkit_version_sync.
#
# Design choice: this script does NOT apply the fix itself. `go get`
# against mcpkit mutates go.mod + go.sum, runs `go mod tidy`, and
# ideally lands alongside a PR — the remediation path belongs to a
# human in the loop, not a weekly systemd timer. Detection only.
#
# Run: scripts/mcpkit-drift-check.sh
# Enable: systemctl --user enable --now dotfiles-mcpkit-drift.timer

set -euo pipefail

DOTFILES_DIR="${DOTFILES_DIR:-$HOME/hairglasses-studio/dotfiles}"
EVENTS_LOG="${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/events.jsonl"
MOD_PATH="$DOTFILES_DIR/mcp/dotfiles-mcp/go.mod"

command -v go >/dev/null 2>&1 || exit 0
[[ -f "$MOD_PATH" ]] || exit 0

now_rfc() { date -u +'%Y-%m-%dT%H:%M:%SZ'; }

json_string() {
  if command -v jq >/dev/null 2>&1; then
    jq -Rs . <<< "$1"
  else
    printf '"%s"' "$(printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g')"
  fi
}

# Current pinned version — second field on the line that references mcpkit.
pinned="$(grep 'github.com/hairglasses-studio/mcpkit' "$MOD_PATH" \
          | awk '{print $NF}' \
          | head -1 || true)"
[[ -n "$pinned" ]] || exit 0

# Latest published — `go list -m -versions` sorts low→high; last field is newest.
# `go list` needs to run inside a module, so cd into dotfiles-mcp.
latest="$(cd "$DOTFILES_DIR/mcp/dotfiles-mcp" \
          && go list -m -versions github.com/hairglasses-studio/mcpkit 2>/dev/null \
          | awk '{print $NF}' || true)"
[[ -n "$latest" ]] || exit 0

if [[ "$pinned" == "$latest" ]]; then
  # No drift — exit silently. The watchdog's liveness signal is the
  # systemd timer firing; there's no positive heartbeat to write.
  exit 0
fi

mkdir -p "$(dirname "$EVENTS_LOG")"
fingerprint_json="$(json_string "pinned=$pinned latest=$latest")"
payload=$(printf '{"type":"mcpkit_version_drift","at":"%s","severity":"low","rule":"weekly drift check","source":"mcpkit-drift-check","error_code":"mcpkit_version_drift","fingerprint":%s}' \
             "$(now_rfc)" "$fingerprint_json")
printf '%s\n' "$payload" >> "$EVENTS_LOG"
