#!/usr/bin/env zsh
# shader-pick.sh — fzf-powered shader picker for Tattoy
# Source this from aliases.zsh; call tattoy-pick or tattoy-pick-cursor

_shader_pick_dir="${0:A:h:h}"
if [[ "$(uname)" == "Darwin" ]]; then
    _shader_pick_config="$HOME/Library/Application Support/tattoy/tattoy.toml"
else
    _shader_pick_config="${XDG_CONFIG_HOME:-$HOME/.config}/tattoy/tattoy.toml"
fi

# Pick a background shader with fzf and apply it to tattoy.toml
tattoy-pick() {
  local shader_dir="$_shader_pick_dir"
  local playlist="$_shader_pick_dir/playlists/tattoo-background.txt"

  # Gather shader names: use playlist if it exists, else glob .glsl files
  local -a shaders
  if [[ -f "$_shader_pick_dir/playlists/tattoy-background.txt" ]]; then
    shaders=("${(@f)$(< "$_shader_pick_dir/playlists/tattoy-background.txt")}")
  else
    shaders=("${(@f)$(ls "$shader_dir"/*.glsl 2>/dev/null | xargs -I{} basename {})}")
  fi

  [[ ${#shaders} -eq 0 ]] && { echo "No shaders found." >&2; return 1; }

  # Get current shader for highlighting
  local current=""
  if [[ -f "$_shader_pick_config" ]]; then
    current="$(sed -n 's/^path = "shaders\/\(.*\)"  # shader-path/\1/p' "$_shader_pick_config")"
  fi

  # Run fzf
  local pick
  pick="$(printf '%s\n' "${shaders[@]}" | fzf \
    --prompt="bg shader> " \
    --header="Current: ${current:-none}" \
    --preview="head -30 '$shader_dir/{}'" \
    --preview-window=right:40%:wrap \
    --height=80% \
    --reverse)"

  [[ -z "$pick" ]] && return 0

  # Apply to tattoy.toml
  if [[ -f "$_shader_pick_config" ]]; then
    local tmp
    tmp="$(mktemp "${_shader_pick_config}.XXXXXX")"
    sed "s|^path = \"shaders/.*\"  # shader-path|path = \"shaders/$pick\"  # shader-path|" \
      "$_shader_pick_config" > "$tmp"
    command mv -f "$tmp" "$_shader_pick_config"
    echo "Background shader → $pick"
  else
    echo "Config not found: $_shader_pick_config" >&2
    return 1
  fi
}

# Pick a cursor shader with fzf and apply it to tattoy.toml
tattoy-pick-cursor() {
  local shader_dir="$_shader_pick_dir"

  local -a shaders
  if [[ -f "$_shader_pick_dir/playlists/tattoy-cursor.txt" ]]; then
    shaders=("${(@f)$(< "$_shader_pick_dir/playlists/tattoy-cursor.txt")}")
  else
    shaders=("${(@f)$(ls "$shader_dir"/cursor_*.glsl 2>/dev/null | xargs -I{} basename {})}")
  fi

  [[ ${#shaders} -eq 0 ]] && { echo "No cursor shaders found." >&2; return 1; }

  local current=""
  if [[ -f "$_shader_pick_config" ]]; then
    current="$(sed -n 's/^path = "shaders\/\(.*\)"  # cursor-path/\1/p' "$_shader_pick_config")"
  fi

  local pick
  pick="$(printf '%s\n' "${shaders[@]}" | fzf \
    --prompt="cursor shader> " \
    --header="Current: ${current:-none}" \
    --preview="head -30 '$shader_dir/{}'" \
    --preview-window=right:40%:wrap \
    --height=80% \
    --reverse)"

  [[ -z "$pick" ]] && return 0

  if [[ -f "$_shader_pick_config" ]]; then
    local tmp
    tmp="$(mktemp "${_shader_pick_config}.XXXXXX")"
    sed "s|^path = \"shaders/.*\"  # cursor-path|path = \"shaders/$pick\"  # cursor-path|" \
      "$_shader_pick_config" > "$tmp"
    command mv -f "$tmp" "$_shader_pick_config"
    echo "Cursor shader → $pick"
  else
    echo "Config not found: $_shader_pick_config" >&2
    return 1
  fi
}
