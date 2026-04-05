#!/usr/bin/env bash
# mod-shader.sh — hg shader module
# Wraps the existing shader management scripts in ghostty/shaders/bin/

_SHADER_BIN="$HG_DOTFILES/ghostty/shaders/bin"
_SHADER_DIR="$HOME/.config/ghostty/shaders"

source "$HG_DOTFILES/scripts/lib/ghostty-config.sh" 2>/dev/null

shader_description() {
  echo "132+ GLSL shaders — browse, pick, cycle, test"
}

shader_commands() {
  cat <<'CMDS'
current	Show currently active shader
set	Set active shader by name
random	Pick a random shader
cycle	Cycle through curated shader list [--prev|--show]
next	Advance shader playlist
pick	Interactive fzf shader picker
list	List shaders [--category X] [--cost Y]
info	Show shader metadata
audit	Audition shaders one-by-one [--category X] [--resume]
test	Run shader compilation tests [name]
build	Inline shared GLSL libraries [--all|--check]
status	Show shader state, playlist, and timer info
CMDS
}

# ── Subcommand implementations ─────────────────

_shader_current() {
  local name
  name="$(ghostty_get_shader_name)"
  if [[ -z "$name" ]]; then
    echo "${HG_DIM}no shader active${HG_RESET}"
  else
    local anim
    anim="$(ghostty_get_shader_animation)"
    printf "%s%s%s" "$HG_CYAN" "$name" "$HG_RESET"
    [[ "$anim" == "true" ]] && printf " %s(animated)%s" "$HG_DIM" "$HG_RESET"
    printf "\n"
  fi
}

_shader_set() {
  local name="${1:-}"
  [[ -n "$name" ]] || hg_die "Usage: hg shader set <name>"

  # Resolve path
  local shader_path=""
  if [[ -f "$_SHADER_DIR/${name}.glsl" ]]; then
    shader_path="$_SHADER_DIR/${name}.glsl"
  elif [[ -f "$_SHADER_DIR/$name" ]]; then
    shader_path="$_SHADER_DIR/$name"
  else
    hg_die "Shader not found: $name"
  fi

  # Detect animation
  local anim="false"
  grep -qE '(ghostty_time|iTime|u_time)' "$shader_path" 2>/dev/null && anim="true"

  # Atomic update
  local tmp
  tmp="$(mktemp "${_GHOSTTY_CONFIG}.XXXXXX")" || hg_die "mktemp failed"
  sed -e "s|^#* *custom-shader = .*|custom-shader = $shader_path|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = $anim|" \
      "$_GHOSTTY_CONFIG" > "$tmp"
  mv -f "$tmp" "$_GHOSTTY_CONFIG"

  local display_name
  display_name="$(basename "$shader_path" .glsl)"
  printf "%s→%s %s%s%s" "$HG_GREEN" "$HG_RESET" "$HG_CYAN" "$display_name" "$HG_RESET"
  [[ "$anim" == "true" ]] && printf " %s(animated)%s" "$HG_DIM" "$HG_RESET"
  printf "\n"
}

_shader_info() {
  local name="${1:-}"
  [[ -n "$name" ]] || hg_die "Usage: hg shader info <name>"

  local meta="$_SHADER_BIN/shader-meta.sh"
  [[ -x "$meta" ]] || hg_die "shader-meta.sh not found"

  local category cost source desc
  category="$(bash "$meta" get "$name" category 2>/dev/null || echo "—")"
  cost="$(bash "$meta" get "$name" cost 2>/dev/null || echo "—")"
  source="$(bash "$meta" get "$name" source 2>/dev/null || echo "—")"
  desc="$(bash "$meta" get "$name" description 2>/dev/null || echo "—")"

  printf "\n %s%s%s\n" "$HG_BOLD" "$name" "$HG_RESET"
  printf " %s%-12s%s %s\n" "$HG_DIM" "category" "$HG_RESET" "$category"
  printf " %s%-12s%s %s\n" "$HG_DIM" "cost" "$HG_RESET" "$cost"
  printf " %s%-12s%s %s\n" "$HG_DIM" "source" "$HG_RESET" "$source"
  printf " %s%-12s%s %s\n\n" "$HG_DIM" "description" "$HG_RESET" "$desc"
}

_shader_pick() {
  hg_require fzf
  local meta="$_SHADER_BIN/shader-meta.sh"
  [[ -x "$meta" ]] || hg_die "shader-meta.sh not found"

  local pick
  pick="$(bash "$meta" fzf-lines | fzf \
    --ansi --reverse --height=80% \
    --preview="head -40 '$_SHADER_DIR/{1}.glsl' 2>/dev/null" \
    --preview-window=right:40% \
    --header="Pick a shader (enter to apply)" \
    | awk '{print $1}')"

  [[ -n "$pick" ]] || return 0
  _shader_set "$pick"
}

_shader_next() {
  local playlist_script="$_SHADER_BIN/shader-playlist.sh"
  [[ -f "$playlist_script" ]] || hg_die "shader-playlist.sh not found"

  local active_pl
  active_pl="$(cat "$HOME/.local/state/ghostty/auto-rotate-playlist" 2>/dev/null || echo "low-intensity")"
  zsh -c "source '$playlist_script' && shader-playlist-next '$active_pl'"
}

_shader_status() {
  printf "\n %s%sshader status%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # Active shader
  printf " %s%-14s%s " "$HG_DIM" "active" "$HG_RESET"
  _shader_current

  # Cycle index
  local cycle_idx
  cycle_idx="$(cat "$HOME/.local/state/ghostty/cycle-index" 2>/dev/null || echo "0")"
  printf " %s%-14s%s %s\n" "$HG_DIM" "cycle index" "$HG_RESET" "$cycle_idx"

  # Active playlist
  local playlist
  playlist="$(cat "$HOME/.local/state/ghostty/auto-rotate-playlist" 2>/dev/null || echo "low-intensity")"
  printf " %s%-14s%s %s\n" "$HG_DIM" "playlist" "$HG_RESET" "$playlist"

  # Timer status (systemd on Linux, launchd on macOS)
  if [[ "$(uname)" == "Darwin" ]]; then
    if launchctl list com.dotfiles.shader-rotate &>/dev/null 2>&1; then
      printf " %s%-14s%s %srunning%s\n" "$HG_DIM" "auto-rotate" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf " %s%-14s%s %sstopped%s\n" "$HG_DIM" "auto-rotate" "$HG_RESET" "$HG_DIM" "$HG_RESET"
    fi
  else
    if systemctl --user is-active shader-rotate.timer &>/dev/null 2>&1; then
      printf " %s%-14s%s %srunning%s\n" "$HG_DIM" "auto-rotate" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf " %s%-14s%s %sstopped%s\n" "$HG_DIM" "auto-rotate" "$HG_RESET" "$HG_DIM" "$HG_RESET"
    fi
  fi
  printf "\n"
}

# ── Dispatch ───────────────────────────────────

shader_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    current)  _shader_current ;;
    set)      _shader_set "$@" ;;
    random)   exec bash "$_SHADER_BIN/shader-random.sh" ;;
    cycle)    exec bash "$_SHADER_BIN/shader-cycle.sh" "$@" ;;
    next)     _shader_next ;;
    pick)     _shader_pick ;;
    list)     exec bash "$_SHADER_BIN/shader-meta.sh" list "$@" ;;
    info)     _shader_info "$@" ;;
    audit)    exec bash "$_SHADER_BIN/shader-audit.sh" "$@" ;;
    test)     exec bash "$_SHADER_BIN/shader-test.sh" "$@" ;;
    build)    exec bash "$_SHADER_BIN/shader-build.sh" "$@" ;;
    status)   _shader_status ;;
    *)        hg_die "Unknown shader command: $cmd. Run 'hg shader --help'." ;;
  esac
}
