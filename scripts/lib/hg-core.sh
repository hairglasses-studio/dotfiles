#!/usr/bin/env bash
# hg-core.sh — Shared framework for the hg CLI
# Source this file: source "$(dirname "$0")/lib/hg-core.sh"

# ── Snazzy palette (24-bit true color) ─────────
HG_CYAN=$'\033[38;2;87;199;255m'
HG_GREEN=$'\033[38;2;90;247;142m'
HG_MAGENTA=$'\033[38;2;255;106;193m'
HG_YELLOW=$'\033[38;2;243;249;157m'
HG_RED=$'\033[38;2;255;92;87m'
HG_DIM=$'\033[38;5;243m'
HG_BOLD=$'\033[1m'
HG_RESET=$'\033[0m'

# ── Formatted output ──────────────────────────
hg_info()  { printf "%s[info]%s  %s\n" "$HG_CYAN"   "$HG_RESET" "$1"; }
hg_ok()    { printf "%s[ok]%s    %s\n" "$HG_GREEN"  "$HG_RESET" "$1"; }
hg_warn()  { printf "%s[warn]%s  %s\n" "$HG_YELLOW" "$HG_RESET" "$1"; }
hg_error() { printf "%s[err]%s   %s\n" "$HG_RED"    "$HG_RESET" "$1" >&2; }
hg_die()   { hg_error "$1"; exit "${2:-1}"; }

# ── Require commands ───────────────────────────
hg_require() {
  for cmd in "$@"; do
    command -v "$cmd" &>/dev/null || hg_die "$cmd is required but not installed"
  done
}

# ── Paths ──────────────────────────────────────
_hg_core_path="${BASH_SOURCE[0]:-$0}"
_hg_core_dir="$(cd "$(dirname "$_hg_core_path")" && pwd)"
_hg_core_dotfiles="$(cd "$_hg_core_dir/../.." && pwd)"
_hg_core_studio="$(cd "$_hg_core_dotfiles/.." && pwd)"

_hg_core_dotfiles_is_valid() {
  local root="${1:-}"
  [[ -n "$root" ]] && [[ -d "$root/scripts" ]] && [[ -f "$root/AGENTS.md" ]]
}

_hg_core_studio_is_valid() {
  local root="${1:-}"
  _hg_core_dotfiles_is_valid "$root/dotfiles"
}

if [[ -n "${HG_STUDIO_ROOT:-}" ]] && _hg_core_studio_is_valid "${HG_STUDIO_ROOT}"; then
  HG_STUDIO_ROOT="$(cd "${HG_STUDIO_ROOT}" && pwd)"
  HG_DOTFILES="${HG_STUDIO_ROOT}/dotfiles"
elif [[ -n "${DOTFILES_DIR:-}" ]] && _hg_core_dotfiles_is_valid "${DOTFILES_DIR}"; then
  HG_DOTFILES="$(cd "${DOTFILES_DIR}" && pwd)"
  HG_STUDIO_ROOT="$(cd "${HG_DOTFILES}/.." && pwd)"
elif _hg_core_dotfiles_is_valid "$_hg_core_dotfiles"; then
  HG_STUDIO_ROOT="$_hg_core_studio"
  HG_DOTFILES="$_hg_core_dotfiles"
else
  HG_STUDIO_ROOT="$HOME/hairglasses-studio"
  HG_DOTFILES="$HG_STUDIO_ROOT/dotfiles"
fi

HG_STATE_DIR="$HOME/.local/state/hg"
mkdir -p "$HG_STATE_DIR" 2>/dev/null || true

hg_gemini_builtin_command_names() {
  cat <<'EOF'
about
agents
auth
bug
chat
clear
commands
compress
copy
directory
dir
docs
editor
extensions
help
hooks
ide
init
mcp
memory
model
oncall
permissions
plan
policies
privacy
quit
exit
restore
rewind
resume
settings
shells
bashes
setup-github
skills
stats
terminal-setup
theme
tools
upgrade
vim
EOF
}

hg_gemini_name_is_builtin() {
  local name="${1:-}"
  local normalized builtin
  normalized="$(printf '%s' "$name" | tr '[:upper:]' '[:lower:]')"

  while IFS= read -r builtin; do
    [[ -n "$builtin" ]] || continue
    if [[ "$normalized" == "$builtin" ]]; then
      return 0
    fi
  done < <(hg_gemini_builtin_command_names)

  return 1
}
