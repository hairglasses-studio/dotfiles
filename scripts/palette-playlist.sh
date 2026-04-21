#!/usr/bin/env bash
set -euo pipefail

# palette-playlist.sh — Swap the active Hairglasses palette.
#
# The palette pipeline already renders every consumer from theme/palette.env.
# This script keeps palette.env as a symlink into theme/palettes/<name>.env and
# re-fires palette-propagate.sh after each swap. Downstream scripts, templates,
# and the matugen pipeline need no changes — the source file is the same path;
# only the bytes behind it rotate.
#
# Commands:
#   list                 Print available palettes
#   current              Print the active palette name
#   set <name>           Swap to <name> and propagate
#   next                 Advance through palettes/ alphabetically
#   prev                 Step back through palettes/ alphabetically
#   random               Pick a random non-current palette
#   reset                Restore hairglasses-neon
#   preview <name>       Dry-run: show what palette-propagate would write
#   status               Print palette + position
#
# Flags on set/next/prev/random/reset/preview:
#   --no-reload          Skip consumer reload hooks (render only)
#   --quiet              Suppress notify-send
#
# State: ~/.local/state/palette/ holds current + history; the symlink at
# theme/palette.env is authoritative.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/notify.sh"

_dotfiles="$HG_DOTFILES"
_palettes_dir="$_dotfiles/theme/palettes"
_active_symlink="$_dotfiles/theme/palette.env"
_propagate="$SCRIPT_DIR/palette-propagate.sh"

_state_dir="${XDG_STATE_HOME:-$HOME/.local/state}/palette"
_current_file="$_state_dir/current"
_history_file="$_state_dir/history"
mkdir -p "$_state_dir"

_available() {
  local file name
  [[ -d "$_palettes_dir" ]] || return 1
  while IFS= read -r -d '' file; do
    name="${file##*/}"
    name="${name%.env}"
    printf '%s\n' "$name"
  done < <(find "$_palettes_dir" -maxdepth 1 -type f -name '*.env' -print0 | sort -z)
}

_exists() {
  [[ -f "$_palettes_dir/$1.env" ]]
}

_active_name() {
  if [[ -L "$_active_symlink" ]]; then
    local target
    target="$(readlink "$_active_symlink")"
    target="${target##*/}"
    printf '%s' "${target%.env}"
    return 0
  fi
  if [[ -f "$_current_file" ]]; then
    tr -d '\n' < "$_current_file"
    return 0
  fi
  # Fall back to hairglasses-neon if palette.env is a regular file and matches
  printf 'hairglasses-neon'
}

_record_history() {
  local name="$1"
  printf '%s\t%s\n' "$(date -Iseconds)" "$name" >> "$_history_file"
  printf '%s' "$name" > "$_current_file"
}

_swap_symlink() {
  local target_name="$1"
  local target_path="$_palettes_dir/${target_name}.env"
  [[ -f "$target_path" ]] || hg_die "Palette not found: $target_name"

  # If palette.env is a regular file (first-time migration), back it up before
  # replacing it with a symlink.
  if [[ -e "$_active_symlink" && ! -L "$_active_symlink" ]]; then
    local backup="$_active_symlink.pre-playlist.bak"
    if [[ ! -f "$backup" ]]; then
      cp "$_active_symlink" "$backup"
      hg_info "Backed up original palette.env to ${backup##*/}"
    fi
  fi

  # Atomic replace via tmp symlink + mv.
  local tmp="$_active_symlink.tmp.$$"
  ln -s "palettes/${target_name}.env" "$tmp"
  mv -f "$tmp" "$_active_symlink"
}

_apply() {
  local name="$1" do_reload="$2"
  _exists "$name" || hg_die "Palette not found: $name (run 'palette-playlist list')"
  _swap_symlink "$name"
  _record_history "$name"

  local -a args=()
  $do_reload || args+=(--no-reload)
  "$_propagate" "${args[@]}"
}

_list_names() {
  _available
}

cmd_list() {
  local active current
  active="$(_active_name)"
  hg_info "palettes dir: $_palettes_dir"
  while IFS= read -r current; do
    if [[ "$current" == "$active" ]]; then
      printf '  * %s\n' "$current"
    else
      printf '    %s\n' "$current"
    fi
  done < <(_list_names)
}

cmd_current() {
  _active_name
  printf '\n'
}

cmd_status() {
  local active
  active="$(_active_name)"
  hg_info "palette: $active"
  if [[ -L "$_active_symlink" ]]; then
    hg_info "target:  $(readlink "$_active_symlink")"
  else
    hg_warn "palette.env is not a symlink — run: palette-playlist set $active"
  fi
  local total
  total="$(_list_names | wc -l | tr -d ' ')"
  hg_info "total:   $total palette(s) in $_palettes_dir"
}

cmd_set() {
  local name="${1:-}" do_reload=true
  shift || true
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --no-reload) do_reload=false ;;
      --quiet) true ;;  # no-op (propagate is already quiet-ish)
      *) hg_die "Unknown flag: $1" ;;
    esac
    shift
  done
  [[ -n "$name" ]] || hg_die "Usage: palette-playlist set <name> [--no-reload]"
  _apply "$name" "$do_reload"
  hg_notify_low "Palette" "$name"
  hg_ok "Active palette: $name"
}

_nth_neighbor() {
  local direction="$1"
  local -a names
  mapfile -t names < <(_list_names)
  local n=${#names[@]}
  (( n > 0 )) || hg_die "No palettes in $_palettes_dir"

  local active idx=0 i
  active="$(_active_name)"
  for (( i = 0; i < n; i++ )); do
    if [[ "${names[i]}" == "$active" ]]; then
      idx=$i
      break
    fi
  done

  case "$direction" in
    next) idx=$(( (idx + 1) % n )) ;;
    prev) idx=$(( (idx - 1 + n) % n )) ;;
    *)    hg_die "internal: unknown direction $direction" ;;
  esac
  printf '%s\n' "${names[idx]}"
}

cmd_next() {
  local do_reload=true name
  while [[ $# -gt 0 ]]; do
    case "$1" in --no-reload) do_reload=false ;; *) ;; esac
    shift
  done
  name="$(_nth_neighbor next)"
  _apply "$name" "$do_reload"
  hg_notify_low "Palette" "$name"
  hg_ok "$name"
}

cmd_prev() {
  local do_reload=true name
  while [[ $# -gt 0 ]]; do
    case "$1" in --no-reload) do_reload=false ;; *) ;; esac
    shift
  done
  name="$(_nth_neighbor prev)"
  _apply "$name" "$do_reload"
  hg_notify_low "Palette" "$name"
  hg_ok "$name"
}

cmd_random() {
  local do_reload=true name active
  while [[ $# -gt 0 ]]; do
    case "$1" in --no-reload) do_reload=false ;; *) ;; esac
    shift
  done
  local -a names
  mapfile -t names < <(_list_names)
  local n=${#names[@]}
  (( n > 1 )) || hg_die "Need at least 2 palettes for random; have $n"
  active="$(_active_name)"
  local attempts=0
  while (( attempts < n * 2 )); do
    name="${names[$(( RANDOM % n ))]}"
    [[ "$name" != "$active" ]] && break
    attempts=$((attempts + 1))
  done
  _apply "$name" "$do_reload"
  hg_notify_low "Palette" "$name"
  hg_ok "$name"
}

cmd_reset() {
  local do_reload=true
  while [[ $# -gt 0 ]]; do
    case "$1" in --no-reload) do_reload=false ;; *) ;; esac
    shift
  done
  _apply hairglasses-neon "$do_reload"
  hg_notify_low "Palette" "hairglasses-neon (reset)"
  hg_ok "Reset to hairglasses-neon"
}

cmd_preview() {
  local name="${1:-}"
  [[ -n "$name" ]] || hg_die "Usage: palette-playlist preview <name>"
  _exists "$name" || hg_die "Palette not found: $name"
  hg_info "Preview — would render from $_palettes_dir/$name.env"
  # palette-propagate.sh always re-sources palette.env itself, so the only
  # reliable way to preview is to swap the symlink briefly then restore it.
  local saved=""
  if [[ -L "$_active_symlink" ]]; then
    saved="$(readlink "$_active_symlink")"
  else
    hg_warn "palette.env is not a symlink — run 'set' once before preview"
    return 1
  fi
  local tmp="$_active_symlink.tmp.$$"
  ln -s "palettes/${name}.env" "$tmp"
  mv -f "$tmp" "$_active_symlink"
  # Always restore, even on error.
  trap 'ln -snf "$saved" "$_active_symlink"; trap - EXIT' EXIT INT TERM
  "$_propagate" --dry-run --no-reload || true
  ln -snf "$saved" "$_active_symlink"
  trap - EXIT INT TERM
}

cmd_help() {
  cat <<'EOF'
Usage: palette-playlist <command> [args]

Commands:
  list                  List available palettes (* marks active)
  current               Print the active palette name
  status                Verbose status (symlink target + count)
  set <name>            Swap to <name> and propagate to every consumer
  next [--no-reload]    Advance through palettes/ alphabetically
  prev [--no-reload]    Step back
  random [--no-reload]  Pick a random non-current palette
  reset [--no-reload]   Restore hairglasses-neon
  preview <name>        Dry-run: show targets without writing

Flags:
  --no-reload           Render files but skip service reload hooks

Files:
  theme/palette.env          Symlink → theme/palettes/<active>.env
  theme/palettes/*.env       Available palettes
  ~/.local/state/palette/    current + history
EOF
}

main() {
  local cmd="${1:-help}"
  shift || true
  case "$cmd" in
    list) cmd_list "$@" ;;
    current) cmd_current "$@" ;;
    status) cmd_status "$@" ;;
    set) cmd_set "$@" ;;
    next) cmd_next "$@" ;;
    prev) cmd_prev "$@" ;;
    random) cmd_random "$@" ;;
    reset) cmd_reset "$@" ;;
    preview) cmd_preview "$@" ;;
    help|-h|--help) cmd_help ;;
    *) hg_die "Unknown command: $cmd (run 'palette-playlist help')" ;;
  esac
}

main "$@"
