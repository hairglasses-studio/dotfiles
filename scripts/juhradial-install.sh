#!/usr/bin/env bash
# juhradial-install.sh — install juhradial-mx from a pinned upstream commit
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

REPO_URL="https://github.com/JuhLabs/juhradial-mx.git"
PINNED_COMMIT="db83da0a0117fd63081c557d0f1da4d384b1d255"

quiet=false

for arg in "$@"; do
  case "$arg" in
    --quiet) quiet=true ;;
    *)
      printf 'Unknown option: %s\n' "$arg" >&2
      exit 2
      ;;
  esac
done

log() {
  $quiet || printf '[juhradial-install] %s\n' "$*"
}

require_cmd() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1 || {
    printf 'Missing required command: %s\n' "$cmd" >&2
    exit 1
  }
}

sync_tree() {
  local src="$1"
  local dst="$2"

  if command -v rsync >/dev/null 2>&1; then
    mkdir -p "$dst"
    rsync -a --delete "$src"/ "$dst"/
    return 0
  fi

  rm -rf "$dst"
  mkdir -p "$(dirname "$dst")"
  cp -a "$src" "$dst"
}

apply_repo_patches() {
  local patch_dir patch
  patch_dir="$(juhradial_patch_dir)"
  [[ -d "$patch_dir" ]] || return 0

  while IFS= read -r patch; do
    [[ -n "$patch" ]] || continue
    if git -C "$src_dir" apply --reverse --check "$patch" >/dev/null 2>&1; then
      log "Patch already applied: $(basename "$patch")"
      continue
    fi

    log "Applying patch $(basename "$patch")"
    git -C "$src_dir" apply "$patch"
  done < <(find "$patch_dir" -maxdepth 1 -type f -name '*.patch' | sort)
}

require_cmd git
require_cmd cargo
require_cmd python3
require_cmd sudo

src_dir="$(juhradial_source_dir)"
install_dir="$(juhradial_install_dir)"
bin_dir="$HOME/.local/bin"

mkdir -p "$(dirname "$src_dir")" "$bin_dir" "$(dirname "$install_dir")"

if [[ -d "$src_dir/.git" ]]; then
  log "Updating managed source checkout at $src_dir"
  git -C "$src_dir" fetch --tags --force origin
else
  log "Cloning juhradial-mx into $src_dir"
  rm -rf "$src_dir"
  git clone "$REPO_URL" "$src_dir"
fi

git -C "$src_dir" checkout --detach "$PINNED_COMMIT" >/dev/null
apply_repo_patches

log "Building juhradiald from $PINNED_COMMIT"
cargo build --manifest-path "$src_dir/daemon/Cargo.toml" --release >/dev/null

install -Dm755 "$src_dir/daemon/target/release/juhradiald" "$bin_dir/juhradiald"
ln -sf "$DOTFILES_DIR/scripts/juhradial-mx.sh" "$bin_dir/juhradial-mx"
ln -sf "$DOTFILES_DIR/scripts/juhradial-settings.sh" "$bin_dir/juhradial-settings"
ln -sf "$DOTFILES_DIR/scripts/hyprshell-trigger.sh" "$bin_dir/juhradial-hyprshell-trigger"
ln -sf "$DOTFILES_DIR/scripts/kitty-clipboard-action.sh" "$bin_dir/juhradial-kitty-clipboard"
ln -sf "$DOTFILES_DIR/scripts/kitty-font-wheel.sh" "$bin_dir/juhradial-kitty-font-wheel"
sync_tree "$src_dir/overlay" "$install_dir/overlay"
sync_tree "$src_dir/assets" "$install_dir/assets"

if [[ "${JUHRADIAL_INSTALL_SKIP_SYNC:-0}" != "1" ]]; then
  "$SCRIPT_DIR/juhradial-sync.sh" --quiet
fi

install -Dm644 \
  "$DOTFILES_DIR/systemd/juhradialmx-daemon.service" \
  "$HOME/.config/systemd/user/juhradialmx-daemon.service"

log "Deploying udev rules"
sudo install -Dm644 "$DOTFILES_DIR/udev/99-juhradialmx.rules" /etc/udev/rules.d/99-juhradialmx.rules
sudo install -Dm644 "$DOTFILES_DIR/udev/60-ydotool-uinput.rules" /etc/udev/rules.d/60-ydotool-uinput.rules
sudo udevadm control --reload-rules
sudo udevadm trigger

if ! id -nG "$USER" | tr ' ' '\n' | grep -qx input; then
  log "User is not in the input group; uaccess should cover the active session, but add the group manually if hidraw access fails"
fi

log "Reloading user services"
juhradial_systemctl daemon-reload
juhradial_systemctl enable --now ydotool.service >/dev/null
juhradial_systemctl enable --now juhradialmx-daemon.service >/dev/null
juhradial_systemctl restart juhradialmx-daemon.service >/dev/null

if [[ -n "${WAYLAND_DISPLAY:-}${DISPLAY:-}" ]]; then
  "$SCRIPT_DIR/juhradial-mx.sh" --overlay-only --quiet || true
fi

log "juhradial-mx installed at $install_dir"
