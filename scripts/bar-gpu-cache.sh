#!/usr/bin/env bash
# bar-gpu-cache.sh — Cache writer for GPU telemetry
#
# Writes two caches from a single nvidia-smi call:
#   /tmp/bar-gpu.txt      — plain power label (e.g. "124W") consumed by the
#                           Ironbar GPU widget (cat'd directly).
#   /tmp/bar-gpu-full.txt — pipe-separated snapshot consumed by the ticker's
#                           `gpu` stream:
#                             POWER_W|UTIL_PCT|TEMP_C|MEM_USED|MEM_TOTAL|NAME

set -euo pipefail

CACHE_POWER="/tmp/bar-gpu.txt"
CACHE_FULL="/tmp/bar-gpu-full.txt"
TMP_POWER="$(mktemp /tmp/bar-gpu.XXXXXX)"
TMP_FULL="$(mktemp /tmp/bar-gpu-full.XXXXXX)"
trap 'rm -f "$TMP_POWER" "$TMP_FULL"' EXIT

row=$(nvidia-smi \
  --query-gpu=power.draw,utilization.gpu,temperature.gpu,memory.used,memory.total,name \
  --format=csv,noheader,nounits 2>/dev/null \
  | awk 'NR==1')

if [[ -n "$row" ]]; then
  IFS=',' read -r power util temp mem_used mem_total name <<< "$row"
  power=$(echo "${power:-0}" | awk '{printf "%.0f", $1}')
  util=$(echo "${util:-0}" | awk '{printf "%.0f", $1}')
  temp=$(echo "${temp:-0}" | awk '{printf "%.0f", $1}')
  mem_used=$(echo "${mem_used:-0}" | awk '{printf "%.0f", $1}')
  mem_total=$(echo "${mem_total:-0}" | awk '{printf "%.0f", $1}')
  name=$(echo "${name:-GPU}" | sed -e 's/^ *//' -e 's/ *$//')
  printf '%sW\n' "$power" > "$TMP_POWER"
  printf '%s|%s|%s|%s|%s|%s\n' "$power" "$util" "$temp" "$mem_used" "$mem_total" "$name" > "$TMP_FULL"
else
  printf '' > "$TMP_POWER"
  printf '' > "$TMP_FULL"
fi
mv "$TMP_POWER" "$CACHE_POWER"
mv "$TMP_FULL" "$CACHE_FULL"
