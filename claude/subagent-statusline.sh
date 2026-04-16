#!/bin/bash
# subagentStatusLine — per-agent row in the agent panel
input=$(cat)

mapfile -t F < <(echo "$input" | jq -r '
  (.agent.name // ""),
  (.cwd // ""),
  (.context_window.used_percentage // ""),
  (.cost.total_cost_usd // "")
')
AGENT="${F[0]}"; CWD="${F[1]}"; PCT="${F[2]}"; COST="${F[3]}"

C_RESET=$'\033[0m'
C_DIM=$'\033[2m'
rgb() { printf '\033[38;2;%d;%d;%dm' "$1" "$2" "$3"; }

SLUG="${CWD##*/}"

# Per-agent icon and color
case "$AGENT" in
  *explore*|*Explore*)  ICON=""; CLR=$(rgb 87 199 255) ;;   # cyan
  *plan*|*Plan*)        ICON=""; CLR=$(rgb 243 249 157) ;; # yellow
  *review*|*Review*)    ICON=""; CLR=$(rgb 255 106 193) ;; # magenta
  *code*|*Code*)        ICON=""; CLR=$(rgb 90 247 142) ;;  # green
  *)                    ICON=""; CLR=$(rgb 154 237 254) ;; # light cyan
esac

OUT="${CLR}${ICON} ${AGENT:-agent}${C_RESET}"
[[ -n "$SLUG" ]] && OUT="${OUT} ${C_DIM}${SLUG}${C_RESET}"
[[ -n "$PCT" ]] && OUT="${OUT} ${C_DIM}${PCT%%.*}%${C_RESET}"
[[ -n "$COST" && "$COST" != "0" ]] && OUT="${OUT} ${C_DIM}\$${COST}${C_RESET}"

echo "$OUT"
