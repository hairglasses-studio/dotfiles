#!/usr/bin/env bash
# bar-tokens-cache.sh — Claude token-burn cache for the ticker token-burn stream
#
# Walks today's Claude session JSONL files under ~/.claude/projects/*/*.jsonl
# and sums input + output tokens. Writes compact summary to /tmp/bar-tokens.txt.
# Paired with bar-tokens.timer (60-second interval).
#
# Output format:
#   Line 1: TODAY  in:<n>k out:<n>k  sessions:<n>
#   Lines 2+: <session-prefix>  in:<n>k out:<n>k

set -euo pipefail

CACHE_FILE="/tmp/bar-tokens.txt"
TMPFILE="$(mktemp /tmp/bar-tokens.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

if ! command -v jq >/dev/null; then
  printf 'jq missing\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

today_epoch="$(date -d 'today 00:00' +%s)"
projects_dir="$HOME/.claude/projects"
[[ -d "$projects_dir" ]] || { : > "$TMPFILE"; mv "$TMPFILE" "$CACHE_FILE"; exit 0; }

total_in=0
total_out=0
session_count=0
declare -a per_session

while IFS= read -r -d '' f; do
  mtime=$(stat -c %Y "$f" 2>/dev/null || echo 0)
  (( mtime >= today_epoch )) || continue
  # Sum .message.usage.input_tokens + .message.usage.output_tokens per line
  counts=$(jq -r '
      select(.message.usage != null)
      | [(.message.usage.input_tokens // 0),
         (.message.usage.output_tokens // 0)]
      | @tsv' "$f" 2>/dev/null \
      | awk '{i+=$1; o+=$2} END {printf "%d %d", i+0, o+0}')
  read -r in_t out_t <<<"$counts"
  (( in_t > 0 || out_t > 0 )) || continue
  session_count=$((session_count + 1))
  total_in=$((total_in + in_t))
  total_out=$((total_out + out_t))
  proj_name=$(basename "$(dirname "$f")" | sed 's/^-home-hg-hairglasses-studio-//;s/^-home-hg-//;s/^-//')
  per_session+=("$(printf '%s\tin:%s\tout:%s' "${proj_name:0:24}" "$in_t" "$out_t")")
done < <(find "$projects_dir" -maxdepth 2 -name '*.jsonl' -print0 2>/dev/null)

humanize() {
  local n=$1
  if (( n >= 1000000 )); then printf '%.1fM' "$(echo "scale=1; $n/1000000" | bc -l)"
  elif (( n >= 1000 )); then printf '%dk' $((n / 1000))
  else printf '%d' "$n"
  fi
}

{
  printf 'TODAY  in:%s out:%s  sessions:%d\n' \
    "$(humanize "$total_in")" "$(humanize "$total_out")" "$session_count"
  # Top 8 sessions by total tokens
  for row in "${per_session[@]}"; do
    printf '%s\n' "$row"
  done | awk -F '\t' '
    { inp=$2; outp=$3; sub("in:","",inp); sub("out:","",outp);
      printf "%d\t%s\n", inp+outp, $0 }' \
    | sort -rn | head -8 \
    | awk -F '\t' '{printf "%s  %s  %s\n", $2, $3, $4}'
} > "$TMPFILE"

mv "$TMPFILE" "$CACHE_FILE"
