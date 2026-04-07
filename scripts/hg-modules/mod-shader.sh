#!/usr/bin/env bash
# mod-shader.sh — hg shader module
# Uses kitty-shader-playlist.sh for shader management (CRTty/DarkWindow backends).
# Source GLSL shaders remain in ghostty/shaders/ (transpiled to kitty/shaders/).

_SHADER_BIN="$HG_DOTFILES/ghostty/shaders/bin"
_SHADER_DIR="$HG_DOTFILES/ghostty/shaders"
_KITTY_PLAYLIST="$HG_DOTFILES/scripts/kitty-shader-playlist.sh"

if [[ ! -f "$_KITTY_PLAYLIST" ]]; then
  shader_description() { echo "Shader pipeline (kitty-shader-playlist.sh not found)"; }
  shader_commands() { echo "status	Shader pipeline requires kitty-shader-playlist.sh"; }
  return 0 2>/dev/null || exit 0
fi

shader_description() {
  echo "131+ GLSL shaders — browse, pick, cycle, test"
}

shader_commands() {
  cat <<'CMDS'
current	Show currently active shader
set	Set active shader by name
random	Pick a random shader
next	Advance shader playlist
pick	Interactive fzf shader picker
list	List shaders [--category X] [--cost Y]
info	Show shader metadata
test	Run shader compilation tests [name]
build	Inline shared GLSL libraries [--all|--check]
status	Show shader state, playlist, and timer info
CMDS
}

# ── Subcommand implementations ─────────────────

_shader_current() {
  local current
  current="$(bash "$_KITTY_PLAYLIST" current 2>/dev/null)"
  if [[ -z "$current" ]]; then
    echo "${HG_DIM}no shader active${HG_RESET}"
  else
    printf "%s%s%s\n" "$HG_CYAN" "$current" "$HG_RESET"
  fi
}

_shader_set() {
  local name="${1:-}"
  [[ -n "$name" ]] || hg_die "Usage: hg shader set <name>"
  bash "$_KITTY_PLAYLIST" set "$name" || hg_die "Failed to set shader: $name"
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

_shader_status() {
  printf "\n %s%sshader status%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # Active shader
  printf " %s%-14s%s " "$HG_DIM" "active" "$HG_RESET"
  _shader_current

  # Timer status
  if systemctl --user is-active shader-rotate.timer &>/dev/null 2>&1; then
    printf " %s%-14s%s %srunning%s\n" "$HG_DIM" "auto-rotate" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf " %s%-14s%s %sstopped%s\n" "$HG_DIM" "auto-rotate" "$HG_RESET" "$HG_DIM" "$HG_RESET"
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
    random)   exec bash "$_KITTY_PLAYLIST" random ;;
    next)     exec bash "$_KITTY_PLAYLIST" next ambient ;;
    pick)     _shader_pick ;;
    list)     exec bash "$_SHADER_BIN/shader-meta.sh" list "$@" ;;
    info)     _shader_info "$@" ;;
    test)     exec bash "$_SHADER_BIN/shader-test.sh" "$@" ;;
    build)    exec bash "$_SHADER_BIN/shader-build.sh" "$@" ;;
    status)   _shader_status ;;
    *)        hg_die "Unknown shader command: $cmd. Run 'hg shader --help'." ;;
  esac
}
