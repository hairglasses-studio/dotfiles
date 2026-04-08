#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/notify.sh"

EXTRAS_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/kitty/extras"

sync_repo() {
  local name="$1" url="$2" branch="$3"
  local dir="$EXTRAS_DIR/$name"

  if [[ -d "$dir/.git" ]]; then
    hg_info "Updating $name"
    git -C "$dir" pull --ff-only origin "$branch"
  else
    hg_info "Cloning $name"
    git clone --depth=1 --branch "$branch" "$url" "$dir"
  fi
}

main() {
  mkdir -p "$EXTRAS_DIR"

  sync_repo "kitty-smart-scroll" "https://github.com/yurikhan/kitty-smart-scroll.git" "master"
  sync_repo "kitty-save-session" "https://github.com/dflock/kitty-save-session.git" "main"

  hg_ok "Kitty extras synced to $EXTRAS_DIR"
  hg_notify_low "Kitty" "Extras synced to $(basename "$EXTRAS_DIR")"
  printf '\nUncomment the sample mappings in ~/.config/kitty/kitty.extras.conf to enable smart scrolling.\n'
}

main "$@"
