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

# ── Shader auto-rotate ─────────────────────────
# macOS launchd timer removed — Linux systemd timer is in aliases.linux.zsh
