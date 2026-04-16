#!/bin/bash
# Living statusline — spinner, time-of-day palette, gradient bar, git, cost
input=$(cat)

# ── Single jq parse (one field per line) ─────────
mapfile -t F < <(echo "$input" | jq -r '
  (.cwd // ""),
  (.session_name // ""),
  (.context_window.used_percentage // ""),
  (.cost.total_cost_usd // ""),
  (.output_style.name // "")
')
CWD="${F[0]}"; SESSION="${F[1]}"; PCT="${F[2]}"; COST="${F[3]}"; STYLE="${F[4]}"

DIR="${CWD/#"$HOME"/\~}"

# ── ANSI helpers ─────────────────────────────────
C_RESET=$'\033[0m'
C_DIM=$'\033[2m'
rgb() { printf '\033[38;2;%d;%d;%dm' "$1" "$2" "$3"; }

C_CYAN=$(rgb 87 199 255)
C_MAGENTA=$(rgb 255 106 193)
C_GREEN=$(rgb 90 247 142)
C_YELLOW=$(rgb 243 249 157)
C_RED=$(rgb 255 92 87)

# ── Time-of-day palette ─────────────────────────
HOUR=$(date +%H); HOUR=${HOUR#0}
if (( HOUR < 9 )); then
  C_ACCENT=$(rgb 98 114 164)
elif (( HOUR >= 18 )); then
  C_ACCENT=$(rgb 255 184 108)
else
  C_ACCENT="$C_CYAN"
fi

# ── Braille spinner ──────────────────────────────
FRAMES=(⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏)
SPINNER="${C_ACCENT}${FRAMES[$(( $(date +%s) % 10 ))]}${C_RESET}"

# ── Session name ─────────────────────────────────
SESSION_SEG=""
[[ -n "$SESSION" ]] && SESSION_SEG="${C_MAGENTA}󱄅 ${SESSION}${C_RESET}"

# ── CWD ──────────────────────────────────────────
CWD_SEG="${C_CYAN}${DIR}${C_RESET}"

# ── Git (cached 5s) ──────────────────────────────
GIT_SEG=""
if [[ -n "$CWD" ]] && git -C "$CWD" rev-parse --is-inside-work-tree &>/dev/null; then
  GIT_DIR=$(git -C "$CWD" rev-parse --absolute-git-dir 2>/dev/null)

  # Operation state (cheap file checks, no cache)
  OP=""
  if [[ -d "$GIT_DIR/rebase-merge" ]]; then
    OP="⚡REBASE $(cat "$GIT_DIR/rebase-merge/msgnum" 2>/dev/null)/$(cat "$GIT_DIR/rebase-merge/end" 2>/dev/null)"
  elif [[ -d "$GIT_DIR/rebase-apply" ]]; then
    OP="⚡REBASE $(cat "$GIT_DIR/rebase-apply/next" 2>/dev/null)/$(cat "$GIT_DIR/rebase-apply/last" 2>/dev/null)"
  elif [[ -f "$GIT_DIR/MERGE_HEAD" ]]; then
    OP="⚡MERGE"
  elif [[ -f "$GIT_DIR/CHERRY_PICK_HEAD" ]]; then
    OP="⚡PICK"
  fi

  # Cached branch + ahead/behind + stash
  CACHE_KEY=$(printf '%s' "$CWD" | md5sum | cut -c1-8)
  GIT_CACHE="/tmp/.claude-sl-git-${CACHE_KEY}"
  NOW=$(date +%s)
  REFRESH=1

  if [[ -f "$GIT_CACHE" ]]; then
    IFS=$'\t' read -r CACHED_TIME BRANCH AHEAD BEHIND STASH < "$GIT_CACHE"
    (( NOW - CACHED_TIME <= 5 )) && REFRESH=0
  fi

  if (( REFRESH )); then
    BRANCH=$(git -C "$CWD" branch --show-current 2>/dev/null)
    [[ -z "$BRANCH" ]] && BRANCH="HEAD"
    AHEAD=0; BEHIND=0
    if AB=$(git -C "$CWD" rev-list --left-right --count HEAD...@{upstream} 2>/dev/null); then
      read -r AHEAD BEHIND <<< "$AB"
    fi
    STASH=$(git -C "$CWD" stash list 2>/dev/null | wc -l)
    printf '%s\t%s\t%s\t%s\t%s\n' "$NOW" "$BRANCH" "$AHEAD" "$BEHIND" "$STASH" > "$GIT_CACHE"
  fi

  if [[ -n "$OP" ]]; then
    GIT_SEG="${C_RED}${OP}${C_RESET}"
  else
    (( AHEAD > 0 || BEHIND > 0 )) && GC="$C_YELLOW" || GC="$C_GREEN"
    GIT_SEG="${GC} ${BRANCH}${C_RESET}"
    (( AHEAD > 0 )) && GIT_SEG="${GIT_SEG}${C_GREEN}↑${AHEAD}${C_RESET}"
    (( BEHIND > 0 )) && GIT_SEG="${GIT_SEG}${C_RED}↓${BEHIND}${C_RESET}"
    (( STASH > 0 )) && GIT_SEG="${GIT_SEG}${C_DIM}≡${STASH}${C_RESET}"
  fi
fi

# ── RGB gradient context bar ─────────────────────
CTX_SEG=""
if [[ -n "$PCT" ]]; then
  P=${PCT%%.*}
  FILLED=$(( P * 20 / 100 ))
  (( FILLED > 20 )) && FILLED=20

  _gradient_rgb() {
    local pos=$1
    if (( pos <= 50 )); then
      local t=$(( pos * 2 ))
      R=$(( 90 + 153 * t / 100 ))
      G=$(( 247 + 2 * t / 100 ))
      B=$(( 142 + 15 * t / 100 ))
    else
      local t=$(( (pos - 50) * 2 ))
      R=$(( 243 + 12 * t / 100 ))
      G=$(( 249 - 157 * t / 100 ))
      B=$(( 157 - 70 * t / 100 ))
    fi
  }

  BAR=""
  for (( i=0; i<20; i++ )); do
    POS=$(( i * 100 / 20 ))
    _gradient_rgb "$POS"
    if (( i < FILLED )); then
      BAR="${BAR}$(rgb $R $G $B)█"
    else
      BAR="${BAR}${C_DIM}░"
    fi
  done
  BAR="${BAR}${C_RESET}"

  _gradient_rgb "$P"
  CTX_SEG="${BAR} $(rgb $R $G $B)${P}%${C_RESET}"
fi

# ── Cost + burn rate ─────────────────────────────
COST_SEG=""
if [[ -n "$COST" && "$COST" != "0" ]]; then
  COST_SEG="${C_DIM}$(printf '$%.2f' "$COST")${C_RESET}"

  BURN_CACHE="/tmp/.claude-sl-burn"
  NOW_B=$(date +%s)
  if [[ -f "$BURN_CACHE" ]]; then
    read -r PREV_COST PREV_TIME < "$BURN_CACHE"
    ELAPSED=$(( NOW_B - PREV_TIME ))
    if (( ELAPSED > 0 && ELAPSED < 300 )); then
      RATE=$(awk "BEGIN{r=($COST-$PREV_COST)/$ELAPSED*3600; printf \"%.2f\",r}")
      if [[ "$RATE" != "0.00" && "${RATE:0:1}" != "-" ]]; then
        RI=${RATE%%.*}
        if (( RI >= 10 )); then RC="$C_RED"
        elif (( RI >= 5 )); then RC="$C_YELLOW"
        else RC="$C_DIM"; fi
        COST_SEG="${COST_SEG} ${RC}\$${RATE}/hr${C_RESET}"
      fi
    fi
  fi
  echo "$COST $NOW_B" > "$BURN_CACHE"
fi

# ── Assemble ─────────────────────────────────────
SEP=" ${C_DIM}│${C_RESET} "
OUT="${SPINNER}"
[[ -n "$SESSION_SEG" ]] && OUT="${OUT} ${SESSION_SEG}"
OUT="${OUT} ${CWD_SEG}"
[[ -n "$GIT_SEG" ]] && OUT="${OUT}${SEP}${GIT_SEG}"
[[ -n "$CTX_SEG" ]] && OUT="${OUT}${SEP}${CTX_SEG}"
[[ -n "$COST_SEG" ]] && OUT="${OUT}${SEP}${COST_SEG}"
STYLE_UP=$(echo "$STYLE" | tr '[:lower:]' '[:upper:]')
[[ -n "$STYLE" && "$STYLE" != "Default" && "$STYLE" != "default" ]] && OUT="${OUT}${SEP}${C_DIM}[${STYLE_UP}]${C_RESET}"

# Propagate window title for Hyprland matching
_slug="${CWD##*/}"
if [[ -n "$SESSION" ]]; then
  _title="────${SESSION}────${_slug}"
else
  _title="────${_slug}────"
fi
printf '\033]0;%s\007' "$_title" > /proc/$PPID/fd/0 2>/dev/null

echo "$OUT"
