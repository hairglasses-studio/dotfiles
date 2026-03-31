#!/usr/bin/env bash
# cyberboot.sh — Cyberpunk terminal boot sequence
# Sources into zsh on interactive shell startup

CYBER_FLAG="/tmp/.cyberboot-${USER}"

# ── Typewriter effect ───────────────────────────
_typewrite() {
  local text="$1" delay="${2:-0.008}"
  for (( i=0; i<${#text}; i++ )); do
    printf '%s' "${text:$i:1}"
    sleep "$delay"
  done
  echo
}

# ── Banner ──────────────────────────────────────
if command -v toilet &>/dev/null && command -v lolcat &>/dev/null; then
  toilet -f future "CYBERNET" --filter border 2>/dev/null | lolcat -f -S 50 -p 1
elif command -v figlet &>/dev/null && command -v lolcat &>/dev/null; then
  figlet -f slant "CYBERNET" | lolcat -f -S 50
elif command -v figlet &>/dev/null; then
  figlet -f slant "CYBERNET"
fi

# ── Init sequence (first terminal only) ────────
if [[ ! -f "$CYBER_FLAG" ]]; then
  touch "$CYBER_FLAG"

  printf '\033[38;2;87;199;255m'  # cyan
  _typewrite "> NEURAL_LINK .............. established"
  printf '\033[38;2;90;247;142m'  # green
  _typewrite "> MESH_NET ................. synchronized"
  printf '\033[38;2;255;106;193m' # magenta
  _typewrite "> ENCRYPTION AES-256 ....... active"
  printf '\033[38;2;243;249;157m' # yellow
  _typewrite "> FIREWALL ................. armed"
  printf '\033[0m'
  echo

  # Sneakers decrypt reveal for system info
  if command -v nms &>/dev/null && command -v fastfetch &>/dev/null; then
    fastfetch 2>/dev/null | nms -a -f cyan 2>/dev/null
  elif command -v fastfetch &>/dev/null; then
    fastfetch
  fi
else
  # Subsequent terminals: fast output
  command -v fastfetch &>/dev/null && fastfetch
fi
