#!/usr/bin/env bash
# hg-core.sh — Shared framework for the hg CLI
# Source this file: source "$(dirname "$0")/lib/hg-core.sh"

# ── Hairglasses Neon palette (24-bit true color) ─────────
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

hg_env_file_value() {
  local env_file="${1:-}"
  local key="${2:-}"
  [[ -n "$env_file" ]] && [[ -n "$key" ]] && [[ -f "$env_file" ]] || return 0

  local line raw value
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ "$line" =~ ^[[:space:]]*# ]] && continue
    [[ "$line" =~ ^[[:space:]]*$ ]] && continue
    if [[ "$line" =~ ^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*=(.*)$ ]]; then
      [[ "${BASH_REMATCH[2]}" == "$key" ]] || continue
      raw="${BASH_REMATCH[3]}"
      raw="${raw#"${raw%%[![:space:]]*}"}"
      raw="${raw%"${raw##*[![:space:]]}"}"
      if [[ "$raw" =~ ^\"(.*)\"$ ]]; then
        value="${BASH_REMATCH[1]}"
      elif [[ "$raw" =~ ^\'(.*)\'$ ]]; then
        value="${BASH_REMATCH[1]}"
      else
        value="${raw%%[[:space:]]#*}"
      fi
      printf '%s\n' "$value"
      return 0
    fi
  done <"$env_file"
}

hg_user_home() {
  local user="${1:-}"
  [[ -n "$user" ]] || return 1
  getent passwd "$user" | cut -d: -f6
}

hg_user_uid() {
  local user="${1:-}"
  [[ -n "$user" ]] || return 1
  id -u "$user"
}

hg_user_gid() {
  local user="${1:-}"
  [[ -n "$user" ]] || return 1
  id -g "$user"
}

hg_user_runtime_dir() {
  local user="${1:-}"
  local uid
  [[ -n "$user" ]] || return 1
  uid="$(hg_user_uid "$user")" || return 1
  printf '/run/user/%s\n' "$uid"
}

hg_user_bus_address() {
  local user="${1:-}"
  local runtime
  [[ -n "$user" ]] || return 1
  runtime="$(hg_user_runtime_dir "$user")" || return 1
  printf 'unix:path=%s/bus\n' "$runtime"
}

hg_run_as_user() {
  local user="${1:-}"
  shift || true
  [[ -n "$user" ]] || hg_die "hg_run_as_user requires a target user"
  [[ $# -gt 0 ]] || hg_die "hg_run_as_user requires a command"

  local owner_home runtime bus uid gid current_uid current_user
  owner_home="$(hg_user_home "$user")"
  [[ -n "$owner_home" ]] || hg_die "Failed to resolve home directory for user: $user"
  runtime="$(hg_user_runtime_dir "$user")"
  bus="$(hg_user_bus_address "$user")"
  uid="$(hg_user_uid "$user")"
  gid="$(hg_user_gid "$user")"
  current_uid="${EUID:-$(id -u)}"
  current_user="$(id -un)"

  if [[ "$current_user" == "$user" ]]; then
    env \
      HOME="$owner_home" \
      USER="$user" \
      LOGNAME="$user" \
      XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-$runtime}" \
      DBUS_SESSION_BUS_ADDRESS="${DBUS_SESSION_BUS_ADDRESS:-$bus}" \
      "$@"
    return $?
  fi

  # Prefer setpriv for root->user hops: keeps env explicit without creating the
  # extra PAM/audit session noise that runuser/sudo can trigger under logind.
  if [[ "$current_uid" -eq 0 ]] && command -v setpriv >/dev/null 2>&1; then
    env \
      HOME="$owner_home" \
      USER="$user" \
      LOGNAME="$user" \
      XDG_RUNTIME_DIR="$runtime" \
      DBUS_SESSION_BUS_ADDRESS="$bus" \
      setpriv --reuid "$uid" --regid "$gid" --init-groups "$@"
    return $?
  fi

  if [[ "$current_uid" -eq 0 ]] && command -v runuser >/dev/null 2>&1; then
    runuser -u "$user" -- env \
      HOME="$owner_home" \
      USER="$user" \
      LOGNAME="$user" \
      XDG_RUNTIME_DIR="$runtime" \
      DBUS_SESSION_BUS_ADDRESS="$bus" \
      "$@"
    return $?
  fi

  if command -v sudo >/dev/null 2>&1; then
    sudo -u "$user" -H env \
      HOME="$owner_home" \
      USER="$user" \
      LOGNAME="$user" \
      XDG_RUNTIME_DIR="$runtime" \
      DBUS_SESSION_BUS_ADDRESS="$bus" \
      "$@"
    return $?
  fi

  hg_die "Cannot switch to user $user: need setpriv, runuser, or sudo"
}

hg_systemctl_user() {
  local user="${1:-}"
  shift || true
  [[ -n "$user" ]] || hg_die "hg_systemctl_user requires a target user"
  hg_run_as_user "$user" systemctl --user "$@"
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

hg_workspace_owner() {
  local root="${1:-${HG_STUDIO_ROOT:-}}"
  if [[ -n "$root" ]] && [[ -e "$root" ]]; then
    stat -c '%U' "$root"
    return 0
  fi
  id -un
}

hg_workspace_owner_home() {
  local owner="${1:-}"
  owner="${owner:-$(hg_workspace_owner "${2:-${HG_STUDIO_ROOT:-}}")}"

  local home=""
  home="$(hg_user_home "$owner" 2>/dev/null || true)"
  if [[ -n "$home" ]]; then
    printf '%s\n' "$home"
    return 0
  fi

  printf '%s\n' "$HOME"
}

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
