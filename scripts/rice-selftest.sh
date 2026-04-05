#!/usr/bin/env bash
set -euo pipefail
# rice-selftest.sh — Comprehensive self-test for all dotfiles rice components
# Outputs structured JSON for MCP/agent consumption, human-readable summary to stderr
# Usage: rice-selftest.sh [--json] [--section SECTION]
# Sections: config, keybinds, plugins, services, fonts, symlinks, palette, tools, shader, all

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh" 2>/dev/null || true

JSON_MODE=false
SECTION="all"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --json) JSON_MODE=true; shift ;;
    --section) SECTION="$2"; shift 2 ;;
    *) SECTION="$1"; shift ;;
  esac
done

errors=0
warnings=0
results=()

add_result() {
  local section="$1" check="$2" status="$3" detail="${4:-}"
  results+=("{\"section\":\"$section\",\"check\":\"$check\",\"status\":\"$status\",\"detail\":\"$detail\"}")
  if [[ "$status" == "fail" ]]; then
    errors=$((errors + 1))
    echo "  [FAIL] $check: $detail" >&2
  elif [[ "$status" == "warn" ]]; then
    warnings=$((warnings + 1))
    echo "  [WARN] $check: $detail" >&2
  else
    echo "  [OK]   $check${detail:+: $detail}" >&2
  fi
}

# ── Section: Hyprland config ───────────────────────
test_config() {
  echo "── Config Validation ──" >&2
  local errs
  errs="$(hyprctl configerrors 2>&1)"
  if [[ -z "$errs" ]]; then
    add_result config "hyprland_config" pass "zero errors"
  else
    add_result config "hyprland_config" fail "$errs"
  fi
}

# ── Section: Keybinds ─────────────────────────────────
test_keybinds() {
  echo "── Keybind Validation ──" >&2
  local binds
  binds="$(hyprctl binds -j 2>/dev/null)"

  # Check critical keybinds are registered (dispatcher|key|label)
  for check in \
    "exec|X|Dropdown terminal" \
    "exec|E|Emoji picker" \
    "hyprexpo:expo|Tab|Workspace overview" \
    "split-workspace|1|Workspace 1" \
  ; do
    IFS='|' read -r dispatcher key label <<< "$check"
    if echo "$binds" | jq -e ".[] | select(.dispatcher == \"$dispatcher\" and .key == \"$key\")" &>/dev/null; then
      add_result keybinds "$label" pass "registered"
    else
      add_result keybinds "$label" warn "not found ($dispatcher $key)"
    fi
  done

  # Check for duplicate keybinds (same mod+key bound twice)
  local dupes
  dupes="$(echo "$binds" | jq -r '[.[] | select(.submap == "") | {mod: .modmask, key: .key}] | group_by(.mod, .key) | map(select(length > 1)) | length')"
  if [[ "$dupes" == "0" ]]; then
    add_result keybinds "no_duplicates" pass "zero conflicts"
  else
    add_result keybinds "no_duplicates" fail "$dupes duplicate keybind(s)"
  fi
}

# ── Section: Plugins ───────────────────────────────
test_plugins() {
  echo "── Hyprland Plugins ──" >&2
  local plugins
  plugins="$(hyprctl plugins list 2>/dev/null | grep "^Plugin" | sed 's/Plugin //' | sed 's/ by.*//')"
  local count
  count="$(echo "$plugins" | wc -l)"

  for p in borders-plus-plus hyprbars hyprexpo hyprfocus hyprwinwrap split-monitor-workspaces dynamic-cursors; do
    if echo "$plugins" | grep -q "$p"; then
      add_result plugins "$p" pass "loaded"
    else
      add_result plugins "$p" warn "not loaded"
    fi
  done
  add_result plugins "total_count" pass "$count plugins"
}

# ── Section: Services ──────────────────────────────
test_services() {
  echo "── Running Services ──" >&2
  for svc in eww swaync swww-daemon hypridle pypr swayosd-server; do
    if pgrep -x "$svc" &>/dev/null; then
      add_result services "$svc" pass "running"
    else
      add_result services "$svc" warn "not running"
    fi
  done

  # Compositor check (Hyprland is the session, not a pgrep-able service name)
  if [[ -n "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
    add_result services "hyprland" pass "active (sig: ${HYPRLAND_INSTANCE_SIGNATURE:0:8}...)"
  else
    add_result services "hyprland" fail "HYPRLAND_INSTANCE_SIGNATURE not set"
  fi
}

# ── Section: Fonts ─────────────────────────────────
test_fonts() {
  echo "── Font Status ──" >&2

  # Ghostty active font
  local font
  font="$(ghostty +show-config 2>/dev/null | grep '^font-family ' | head -1 | sed 's/font-family = //')"
  if [[ -n "$font" ]]; then
    add_result fonts "ghostty_font" pass "$font"
  else
    add_result fonts "ghostty_font" warn "could not query"
  fi

  # Check standard font (Maple Mono NF CN) is installed
  if fc-list "Maple Mono NF CN" | grep -q .; then
    add_result fonts "maple_mono_nf_cn" pass "installed"
  else
    add_result fonts "maple_mono_nf_cn" fail "missing — standard font not found"
  fi
}

# ── Section: Symlinks ──────────────────────────────
test_symlinks() {
  echo "── Symlink Health ──" >&2
  local total=0 ok=0 broken=0
  # Run install.sh --check and parse output
  local output
  output="$(bash "$SCRIPT_DIR/../install.sh" --check 2>&1)" || true
  ok=$(echo "$output" | grep -c "\[OK\]" || true)
  broken=$(echo "$output" | grep -c "\[ERR\]" || true)
  total=$((ok + broken))
  add_result symlinks "total" pass "$ok/$total healthy"
  if [[ $broken -gt 0 ]]; then
    local failures
    failures="$(echo "$output" | grep "\[ERR\]" | head -3)"
    add_result symlinks "broken" fail "$failures"
  fi
}

# ── Section: Palette ───────────────────────────────
test_palette() {
  echo "── Palette Compliance ──" >&2
  local violations=0
  # Check for known non-Snazzy colors in tracked config files
  for color in "#d4d4d4" "#282a36" "#f8f8f2" "#264f78"; do
    local count
    count="$(grep -rl "$color" --include="*.conf" --include="*.toml" --include="*.css" --include="*.scss" --include="*.ini" --include="*.yml" --include="*.yaml" --include="*.theme" --include="*.rasi" "$SCRIPT_DIR/.." 2>/dev/null | grep -v ".git" | wc -l || true)"
    if [[ $count -gt 0 ]]; then
      add_result palette "non_snazzy_$color" fail "$count files"
      violations=$((violations + count))
    fi
  done
  if [[ $violations -eq 0 ]]; then
    add_result palette "snazzy_compliance" pass "zero non-Snazzy colors"
  fi

  # Check for JetBrains font refs
  local jb_count
  jb_count="$(grep -ril "jetbrains" --include="*.conf" --include="*.toml" --include="*.css" --include="*.scss" --include="*.ini" --include="*.yml" --include="*.rasi" "$SCRIPT_DIR/.." 2>/dev/null | grep -v ".git" | wc -l || true)"
  if [[ $jb_count -gt 0 ]]; then
    add_result palette "jetbrains_font_refs" warn "$jb_count files"
  else
    add_result palette "jetbrains_font_refs" pass "zero references"
  fi
}

# ── Section: Tools ─────────────────────────────────
test_tools() {
  echo "── CLI Tools ──" >&2
  for tool in eza fd bat rg zoxide fzf starship atuin duf dust procs lazygit hyprpicker wayshot hyprshade wallust wl-screenrec wofi-emoji; do
    if command -v "$tool" &>/dev/null; then
      add_result tools "$tool" pass "installed"
    else
      add_result tools "$tool" warn "not found"
    fi
  done
}

# ── Section: Shader ────────────────────────────────
test_shader() {
  echo "── Shader Status ──" >&2
  local shader
  shader="$(grep '^custom-shader = ' "$HOME/.config/ghostty/config" 2>/dev/null | sed 's/custom-shader = //' || true)"
  local anim
  anim="$(grep '^custom-shader-animation = ' "$HOME/.config/ghostty/config" 2>/dev/null | sed 's/custom-shader-animation = //' || true)"
  add_result shader "active_shader" pass "${shader:-none}"
  add_result shader "animation" pass "${anim:-false}"

  # Shader count
  local count
  count="$(find "$HOME/.config/ghostty/shaders" -maxdepth 2 -name "*.glsl" 2>/dev/null | wc -l || true)"
  add_result shader "shader_count" pass "$count shaders"
}

# ── Run sections ───────────────────────────────────
case "$SECTION" in
  config)    test_config ;;
  keybinds)  test_keybinds ;;
  plugins)   test_plugins ;;
  services)  test_services ;;
  fonts)     test_fonts ;;
  symlinks)  test_symlinks ;;
  palette)   test_palette ;;
  tools)     test_tools ;;
  shader)    test_shader ;;
  all)
    test_config
    test_keybinds
    test_plugins
    test_services
    test_fonts
    test_shader
    test_palette
    test_tools
    test_symlinks
    ;;
esac

# ── Output ─────────────────────────────────────────
echo "" >&2
echo "── Summary: $errors errors, $warnings warnings ──" >&2

if $JSON_MODE; then
  printf '{"errors":%d,"warnings":%d,"results":[%s]}\n' "$errors" "$warnings" "$(IFS=,; echo "${results[*]}")"
fi

exit "$errors"
