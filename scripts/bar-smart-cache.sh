#!/usr/bin/env bash
# bar-smart-cache.sh — SMART disk health cache for the ticker smart-disk stream
#
# Runs `sudo -n smartctl -H -j` against each block device and writes a line
# per device to /tmp/bar-smart.txt. Paired with bar-smart.timer (hourly).
#
# Output: one line per device, e.g. "nvme0n1  PASSED (37°C)".

set -euo pipefail

CACHE_FILE="/tmp/bar-smart.txt"
TMPFILE="$(mktemp /tmp/bar-smart.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

if ! command -v smartctl >/dev/null || ! command -v jq >/dev/null; then
  printf 'smartctl/jq missing\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

# Pick block devices we can actually query: NVMe + SATA physical disks.
devices=()
for d in /dev/nvme?n? /dev/sd?; do
  [[ -b "$d" ]] || continue
  # Skip partitions (nvme0n1p1 etc)
  [[ "$d" =~ p[0-9]+$ ]] && continue
  devices+=("$d")
done

if (( ${#devices[@]} == 0 )); then
  printf 'no block devices found\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

for dev in "${devices[@]}"; do
  json="$(sudo -n smartctl -H -A -j "$dev" 2>/dev/null || true)"
  if [[ -z "$json" ]]; then
    printf '%s  UNAVAILABLE\n' "$(basename "$dev")" >> "$TMPFILE"
    continue
  fi
  passed="$(jq -r '.smart_status.passed // empty' <<<"$json")"
  temp="$(jq -r '.temperature.current // empty' <<<"$json")"
  label="FAILED"
  [[ "$passed" == "true" ]] && label="PASSED"
  temp_fmt=""
  [[ -n "$temp" ]] && temp_fmt=" (${temp}°C)"
  printf '%s  %s%s\n' "$(basename "$dev")" "$label" "$temp_fmt" >> "$TMPFILE"
done

mv "$TMPFILE" "$CACHE_FILE"
