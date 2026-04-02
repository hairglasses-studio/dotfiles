#!/bin/bash
input=$(cat)

CWD=$(echo "$input" | jq -r '.cwd // empty')
SLUG="${CWD##*/}"
PCT=$(echo "$input" | jq -r '.context_window.used_percentage // 0' | cut -d. -f1)
SESSION=$(echo "$input" | jq -r '.session_name // empty')

# Snazzy palette ANSI (true-color)
C_RESET=$'\033[0m'
C_BOLD=$'\033[1m'
C_CYAN=$'\033[38;2;87;199;255m'     # #57c7ff
C_MAGENTA=$'\033[38;2;255;106;193m' # #ff6ac1
C_GREEN=$'\033[38;2;90;247;142m'    # #5af78e
C_YELLOW=$'\033[38;2;243;249;157m'  # #f3f99d
C_RED=$'\033[38;2;255;92;87m'       # #ff5c57
C_DIM=$'\033[2m'

# Context color: green < 50%, yellow < 80%, red >= 80%
if (( PCT >= 80 )); then
  C_CTX="$C_RED"
elif (( PCT >= 50 )); then
  C_CTX="$C_YELLOW"
else
  C_CTX="$C_GREEN"
fi

# Nerd Font glyphs
ICON_DIR=""       # nf-cod-folder
ICON_CTX="󰾅"      # nf-md-chart_donut
ICON_SESSION="󱄅"   # nf-md-tag
SEP=""            # powerline thin separator

# COLUMNS may not be set in non-interactive context; default wide
W="${COLUMNS:-120}"

# ── Adaptive path display ──────────────────────
# Abbreviate home prefix
TILDE_CWD="${CWD/#"$HOME"/\~}"

# Smart-abbreviate: compress intermediate path segments
# ~/hairglasses-studio/dotfiles -> ~/h-s/dotfiles
_abbrev_path() {
  local p="$1"
  IFS='/' read -ra parts <<< "$p"
  local last="${parts[-1]}"
  local result=""
  for (( i=0; i<${#parts[@]}-1; i++ )); do
    local seg="${parts[$i]}"
    if [[ "$seg" == "~" ]]; then
      result="~"
    elif [[ -z "$seg" ]]; then
      result=""
    else
      # Abbreviate: keep first char + chars after hyphens/underscores
      # e.g. hairglasses-studio -> h-s, my_project -> m_p
      local abbr="${seg:0:1}"
      local j
      for (( j=1; j<${#seg}; j++ )); do
        local ch="${seg:$j:1}"
        if [[ "$ch" == "-" || "$ch" == "_" || "$ch" == "." ]]; then
          abbr="${abbr}${ch}${seg:$((j+1)):1}"
        fi
      done
      result="${result}/${abbr}"
    fi
  done
  echo "${result}/${last}"
}

ABBREV_CWD=$(_abbrev_path "$TILDE_CWD")

# Pick path display based on width
if (( W > 100 )); then
  PATH_DISPLAY="$TILDE_CWD"
elif (( W > 60 )); then
  PATH_DISPLAY="$ABBREV_CWD"
elif (( W > 40 )); then
  # Parent slug + basename
  parent="${CWD%/*}"
  parent="${parent##*/}"
  PATH_DISPLAY="${parent}/${SLUG}"
else
  PATH_DISPLAY="$SLUG"
fi

# ── Build segments ─────────────────────────────
session_seg=""
if [[ -n "$SESSION" ]]; then
  session_seg="${C_MAGENTA}${ICON_SESSION} ${SESSION}${C_RESET} "
fi

slug_seg="${C_BOLD}${C_CYAN}${ICON_DIR} ${PATH_DISPLAY:-?}${C_RESET}"
ctx_seg="${C_DIM}${SEP}${C_RESET} ${C_CTX}${ICON_CTX} ${PCT}%${C_RESET}"

if (( W < 40 )); then
  echo "${session_seg}${slug_seg}"
else
  echo "${session_seg}${slug_seg} ${ctx_seg}"
fi
