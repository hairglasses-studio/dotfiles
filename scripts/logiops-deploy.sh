#!/usr/bin/env bash
# logiops-deploy.sh — Deploy repo-managed logid config to /etc and restart logid
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SRC="$DOTFILES_DIR/logiops/logid.cfg"
DST="/etc/logid.cfg"

quiet=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --quiet) quiet=true ;;
    *)
      printf 'Unknown option: %s\n' "$1" >&2
      exit 2
      ;;
  esac
  shift
done

info() {
  $quiet && return 0
  printf '\033[0;36m:: %s\033[0m\n' "$*"
}

ok() {
  $quiet && return 0
  printf '\033[0;32m   %s\033[0m\n' "$*"
}

warn() {
  $quiet && return 0
  printf '\033[0;33m   %s\033[0m\n' "$*"
}

[[ -f "$SRC" ]] || {
  printf 'Tracked logiops config missing: %s\n' "$SRC" >&2
  exit 1
}

if [[ -f "$DST" ]] && cmp -s "$SRC" "$DST"; then
  info "logid config already in sync"
else
  if [[ -f "$DST" ]]; then
    backup="${DST}.bak.$(date +%Y%m%d-%H%M%S)"
    info "Backing up current /etc/logid.cfg to $backup"
    sudo cp "$DST" "$backup"
  fi
  info "Deploying $SRC to $DST"
  sudo install -Dm644 "$SRC" "$DST"
  ok "logid config deployed"
fi

info "Restarting logid.service"
sudo systemctl restart logid.service
ok "logid restarted"
