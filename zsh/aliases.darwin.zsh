#!/usr/bin/env zsh
# ── macOS-specific aliases ── Snazzy cyberpunk edition ──
# Sourced only on Darwin via zshrc OSTYPE guard.

# ── Clipboard / opener polyfills ─────────────
# (macOS has these natively — no aliases needed)

# ── System information ────────────────────────
alias top='top -o cpu'
alias mem='memory_pressure'
alias cpu='top -l 1 | grep "CPU usage"'

# ── Trash ────────────────────────────────────
alias emptytrash='rm -rf ~/.Trash/*'

# ── Window management ────────────────────────
alias aero-reload='aerospace reload-config'
alias bar-reload='sketchybar --reload'

# ── CRT / RetroVisor (ScreenCaptureKit) ──────
alias crt-on='open -a RetroVisor'
alias crt-off='pkill -x RetroVisor'
alias crt-toggle='pgrep -x RetroVisor && pkill -x RetroVisor || open -a RetroVisor'

# ── Peekaboo — screen capture for visual review ──
alias peek='peekaboo'

# ── Shader auto-rotate via launchd ───────────
shader-auto() {
  local plist="$HOME/Library/LaunchAgents/com.dotfiles.shader-rotate.plist"
  local label="com.dotfiles.shader-rotate"
  case "${1:-status}" in
    start)
      local interval=$(( ${2:-30} * 60 ))
      local tmp; tmp="$(mktemp "${plist}.XXXXXX")"
      sed "s|<integer>[0-9]*</integer>|<integer>${interval}</integer>|" "$plist" > "$tmp"
      mv -f "$tmp" "$plist"
      launchctl unload "$plist" 2>/dev/null
      launchctl load "$plist"
      echo "Shader auto-rotate started (every ${2:-30} minutes)"
      ;;
    stop)
      launchctl unload "$plist" 2>/dev/null
      echo "Shader auto-rotate stopped"
      ;;
    status)
      if launchctl list "$label" &>/dev/null; then
        echo "Shader auto-rotate: running"
      else
        echo "Shader auto-rotate: stopped"
      fi
      ;;
    *)
      echo "Usage: shader-auto {start [minutes]|stop|status}"
      ;;
  esac
}
