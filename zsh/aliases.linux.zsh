#!/usr/bin/env zsh
# ── Linux-specific aliases ── Snazzy cyberpunk edition ──
# Sourced only on Linux via zshrc OSTYPE guard.

# ── Clipboard / opener polyfills ─────────────
# Let macOS-written scripts use pbcopy/pbpaste seamlessly
alias pbcopy='wl-copy'
alias pbpaste='wl-paste'
alias open='xdg-open'

# ── System information ────────────────────────
alias mem='free -h'
alias cpu='grep "cpu " /proc/stat | awk "{usage=(\$2+\$4)*100/(\$2+\$3+\$4+\$5)} END {print usage \"%\"}"'

# ── Trash ────────────────────────────────────
alias emptytrash='rm -rf ~/.local/share/Trash/files/* ~/.local/share/Trash/info/*'

# ── Package management wrapper ────────────────
yay() {
  if [[ -t 1 ]] && cmd_exists tte && [[ "$1" == "-S" || "$1" == "-Syu" ]]; then
    echo "PACKAGE ACQUISITION // $*" | tte beams \
      --beam-delay 2 --beam-gradient-stops 57c7ff 5af78e \
      --final-gradient-stops 57c7ff 5af78e 2>/dev/null
  fi
  command yay "$@"
  local rc=$?
  if (( rc == 0 )) && [[ -t 1 ]] && cmd_exists tte && [[ "$1" == "-S" || "$1" == "-Syu" ]]; then
    echo "PACKAGES SYNCHRONIZED" | tte slide \
      --movement-speed 1.5 --final-gradient-stops 5af78e 57c7ff 2>/dev/null
  fi
  return $rc
}

# ── Shader auto-rotate via systemd timer ─────
shader-auto() {
  local timer="shader-rotate.timer"
  local service="shader-rotate.service"
  case "${1:-status}" in
    start)
      local interval="${2:-30}"
      # Create/update the systemd timer unit
      local timer_dir="$HOME/.config/systemd/user"
      mkdir -p "$timer_dir"
      cat > "$timer_dir/$timer" <<UNIT
[Unit]
Description=Shader auto-rotate timer

[Timer]
OnActiveSec=${interval}min
OnUnitActiveSec=${interval}min
AccuracySec=10s

[Install]
WantedBy=timers.target
UNIT
      cat > "$timer_dir/$service" <<UNIT
[Unit]
Description=Rotate Ghostty shader

[Service]
Type=oneshot
ExecStart=%h/.config/ghostty/shaders/bin/shader-auto-rotate.sh
Environment=PATH=%h/.local/bin:%h/bin:/usr/local/bin:/usr/bin:/bin
UNIT
      systemctl --user daemon-reload
      systemctl --user enable --now "$timer"
      echo "Shader auto-rotate started (every ${interval} minutes)"
      ;;
    stop)
      systemctl --user disable --now "$timer" 2>/dev/null
      echo "Shader auto-rotate stopped"
      ;;
    status)
      if systemctl --user is-active "$timer" &>/dev/null; then
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
