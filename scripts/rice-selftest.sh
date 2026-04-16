#!/usr/bin/env bash
set -euo pipefail
# rice-selftest.sh — Comprehensive self-test for all dotfiles rice components
# Outputs structured JSON for MCP/agent consumption, human-readable summary to stderr
# Usage: rice-selftest.sh [--json] [--section SECTION]
# Sections: config, keybinds, plugins, services, fonts, symlinks, palette, tools, shader, persistence, all

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh" 2>/dev/null || true
source "$SCRIPT_DIR/lib/kitty-config.sh" 2>/dev/null || true
source "$SCRIPT_DIR/lib/runtime-desktop-env.sh" 2>/dev/null || true

JSON_MODE=false
SECTION="all"
JQ_AVAILABLE=true
while [[ $# -gt 0 ]]; do
  case "$1" in
    --json) JSON_MODE=true; shift ;;
    --section) SECTION="$2"; shift 2 ;;
    *) SECTION="$1"; shift ;;
  esac
done

if ! command -v jq >/dev/null 2>&1; then
  JQ_AVAILABLE=false
fi

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

hypr_configerror_lines() {
  local raw="$1"

  if [[ -z "${raw//[[:space:]]/}" ]]; then
    return 0
  fi

  if $JQ_AVAILABLE && printf '%s' "$raw" | jq -e 'type == "array"' >/dev/null 2>&1; then
    printf '%s' "$raw" | jq -r '.[]? | strings | select((gsub("\\s+"; "")) | length > 0)'
    return 0
  fi

  if grep -qi "no errors" <<<"$raw"; then
    return 0
  fi

  printf '%s\n' "$raw" | sed '/^[[:space:]]*$/d'
}

# ── Section: Hyprland config ───────────────────────
test_config() {
  echo "── Config Validation ──" >&2
  if ! $JQ_AVAILABLE; then
    add_result config "jq_dependency" fail "jq not found"
    return
  fi
  refresh_desktop_runtime_env 2>/dev/null || true
  local raw errs stderr_raw stderr_file hypr_status
  stderr_file="$(mktemp)"
  hypr_status=0
  if raw="$(hyprctl -j configerrors 2>"$stderr_file")"; then
    hypr_status=0
  else
    hypr_status=$?
  fi
  stderr_raw="$(cat "$stderr_file")"
  rm -f "$stderr_file"

  errs="$(hypr_configerror_lines "$raw")"
  if [[ -z "$errs" && -z "${raw//[[:space:]]/}" && -n "${stderr_raw//[[:space:]]/}" ]]; then
    errs="$(hypr_configerror_lines "$stderr_raw")"
  fi
  if [[ -z "$errs" && $hypr_status -ne 0 && -n "${stderr_raw//[[:space:]]/}" ]]; then
    errs="$(printf '%s\n' "$stderr_raw" | sed '/^[[:space:]]*$/d')"
  fi
  if [[ -z "$errs" ]]; then
    add_result config "hyprland_config" pass "zero errors"
  else
    add_result config "hyprland_config" fail "$errs"
  fi
}

# ── Section: Keybinds ─────────────────────────────────
test_keybinds() {
  echo "── Keybind Validation ──" >&2
  if ! $JQ_AVAILABLE; then
    add_result keybinds "jq_dependency" fail "jq not found"
    return
  fi
  refresh_desktop_runtime_env 2>/dev/null || true
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
  refresh_desktop_runtime_env 2>/dev/null || true
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
  refresh_desktop_runtime_env 2>/dev/null || true
  for svc in ironbar swaync swww-daemon pypr swayosd-server; do
    if [[ "$svc" == "pypr" ]]; then
      if command -v pypr >/dev/null 2>&1 && pypr version >/dev/null 2>&1; then
        add_result services "pypr" pass "daemon responding"
      else
        add_result services "pypr" warn "daemon not responding"
      fi
      continue
    fi
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

  # Kitty active font
  local font
  local match
  font="$(kitty_get_font 2>/dev/null)"
  if [[ -n "$font" ]]; then
    add_result fonts "kitty_font" pass "$font"
    match="$(fc-match -f '%{family}\n' "$font" 2>/dev/null | head -1 || true)"
    if [[ -n "$match" && "$match" == *"$font"* ]]; then
      add_result fonts "kitty_font_available" pass "installed"
    else
      add_result fonts "kitty_font_available" fail "missing — active kitty font resolved to ${match:-unknown}"
    fi
  else
    add_result fonts "kitty_font" warn "could not query"
  fi

  # Monaspace availability for the default Kitty profile
  match="$(fc-match -f '%{family}\n' 'Monaspace Neon' 2>/dev/null | head -1 || true)"
  if [[ -n "$match" && "$match" == *"Monaspace Neon"* ]]; then
    add_result fonts "monaspace_neon" pass "installed"
  else
    add_result fonts "monaspace_neon" fail "missing — default Kitty font family resolved to ${match:-unknown}"
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
  echo "── Theme Surface ──" >&2

  local scan_paths=(
    "$SCRIPT_DIR/../ironbar"
    "$SCRIPT_DIR/../hyprshell"
    "$SCRIPT_DIR/../swaync"
    "$SCRIPT_DIR/../wofi"
    "$SCRIPT_DIR/../wlogout"
    "$SCRIPT_DIR/../hyprland"
    "$SCRIPT_DIR/../kitty/kitty.conf"
    "$SCRIPT_DIR/../fontconfig/conf.d/51-monospace.conf"
    "$SCRIPT_DIR/../README.md"
    "$SCRIPT_DIR/../docs/SPRINT-NEXT.md"
  )

  local hack_count matcha_count jetbrains_count
  hack_count="$(rg -l 'Hack Nerd Font' "${scan_paths[@]}" 2>/dev/null | wc -l || true)"
  matcha_count="$(rg -li 'Matcha-dark-sea' "${scan_paths[@]}" 2>/dev/null | wc -l || true)"
  jetbrains_count="$(rg -li 'JetBrains Mono' "${scan_paths[@]}" 2>/dev/null | wc -l || true)"

  if [[ $hack_count -gt 0 ]]; then
    add_result palette "legacy_hack_refs" fail "$hack_count files still reference Hack Nerd Font"
  else
    add_result palette "legacy_hack_refs" pass "zero references"
  fi

  if [[ $matcha_count -gt 0 ]]; then
    add_result palette "legacy_matcha_refs" warn "$matcha_count files still reference Matcha-dark-sea"
  else
    add_result palette "legacy_matcha_refs" pass "zero references"
  fi

  if [[ $jetbrains_count -gt 0 ]]; then
    add_result palette "legacy_jetbrains_refs" warn "$jetbrains_count files still reference JetBrains Mono"
  else
    add_result palette "legacy_jetbrains_refs" pass "zero references"
  fi

  if [[ -f "$SCRIPT_DIR/../theme/palette.env" ]] && [[ -x "$SCRIPT_DIR/theme-sync.sh" ]]; then
    add_result palette "theme_sync_pipeline" pass "palette + sync script present"
  else
    add_result palette "theme_sync_pipeline" fail "missing theme/palette.env or scripts/theme-sync.sh"
  fi

  if [[ -x "$SCRIPT_DIR/hyprpm-bootstrap.sh" ]]; then
    add_result palette "hyprpm_bootstrap" pass "repo-managed plugin bootstrap present"
  else
    add_result palette "hyprpm_bootstrap" warn "scripts/hyprpm-bootstrap.sh missing or not executable"
  fi
}

# ── Section: Tools ─────────────────────────────────
test_tools() {
  echo "── CLI Tools ──" >&2
  for tool in eza fd bat rg zoxide fzf starship atuin duf dust procs lazygit hyprpicker wayshot hyprshade wallust wl-screenrec wofi-emoji cliphist kanshi ddcutil wluma hyprlax papertoy; do
    if command -v "$tool" &>/dev/null; then
      add_result tools "$tool" pass "installed"
    else
      add_result tools "$tool" warn "not found"
    fi
  done
}

# ── Section: Shader ────────────────────────────────
test_shader() {
  echo "── Visual Status ──" >&2
  local opacity tab_style
  opacity="$(kitty_get_opacity 2>/dev/null || echo "unknown")"
  tab_style="$(kitty_get_tab_style 2>/dev/null || echo "unknown")"
  add_result shader "background_opacity" pass "${opacity}"
  add_result shader "tab_bar_style" pass "${tab_style}"

  # Cursor trail check
  local trail
  trail="$(grep '^cursor_trail ' "$HOME/.config/kitty/kitty.conf" 2>/dev/null | awk '{print $2}' || echo "0")"
  add_result shader "cursor_trail" pass "${trail:-disabled}"
}

# ── Section: Tmux persistence ─────────────────────
test_persistence() {
  echo "── Tmux Persistence ──" >&2
  if ! $JQ_AVAILABLE; then
    add_result persistence "jq_dependency" fail "jq not found"
    return
  fi

  local health_script output
  health_script="$SCRIPT_DIR/tmux-persistence-health.sh"
  if [[ ! -x "$health_script" ]]; then
    add_result persistence "tmux_persistence_health" fail "missing $health_script"
    return
  fi

  output="$("$health_script" --json 2>/dev/null || true)"
  if [[ -z "$output" ]] || ! printf '%s' "$output" | jq -e '.results | type == "array"' >/dev/null 2>&1; then
    add_result persistence "tmux_persistence_health" fail "invalid health output"
    return
  fi

  while IFS=$'\t' read -r check status detail; do
    [[ -n "$check" ]] || continue
    add_result persistence "$check" "$status" "$detail"
  done < <(printf '%s' "$output" | jq -r '.results[] | "\(.check)\t\(.status)\t\(.detail)"')
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
  persistence) test_persistence ;;
  all)
    test_config
    test_keybinds
    test_plugins
    test_services
    test_fonts
    test_shader
    test_persistence
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
