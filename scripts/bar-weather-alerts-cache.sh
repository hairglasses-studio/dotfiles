#!/usr/bin/env bash
# bar-weather-alerts-cache.sh — NWS severe weather alerts cache for the
# ticker weather-alerts stream.
#
# Resolves lat/lon from wttr.in (same source the weather stream uses) and
# queries api.weather.gov/alerts/active?point=<lat>,<lon>. Writes:
#   Line 1: <severity> · <N> active
#   Lines 2+: <event>  <headline-first-sentence>
# Paired with bar-weather-alerts.timer (15-minute interval).
#
# NWS covers the United States only. Users outside the US will see an
# empty cache — the stream's _empty() sentinel advances via backoff.

set -uo pipefail

CACHE_FILE="/tmp/bar-weather-alerts.txt"
COORDS_CACHE="/tmp/bar-weather-coords.txt"
TMPFILE="$(mktemp /tmp/bar-weather-alerts.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

if ! command -v curl >/dev/null || ! command -v jq >/dev/null; then
  printf 'curl/jq missing\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

# Resolve coords. Refresh if cache missing or older than a week.
needs_refresh=1
if [[ -f "$COORDS_CACHE" ]]; then
  mtime=$(stat -c %Y "$COORDS_CACHE" 2>/dev/null || echo 0)
  now=$(date +%s)
  (( now - mtime < 604800 )) && needs_refresh=0
fi
if (( needs_refresh )); then
  wttr="$(curl -fsSL --max-time 5 "https://wttr.in/?format=j1" 2>/dev/null)" || exit 0
  lat="$(printf '%s' "$wttr" | jq -r '.nearest_area[0].latitude // empty' 2>/dev/null)"
  lon="$(printf '%s' "$wttr" | jq -r '.nearest_area[0].longitude // empty' 2>/dev/null)"
  [[ -n "$lat" && -n "$lon" ]] || exit 0
  printf '%s %s\n' "$lat" "$lon" > "$COORDS_CACHE"
fi
read -r lat lon <"$COORDS_CACHE"
[[ -n "${lat:-}" && -n "${lon:-}" ]] || exit 0

# Query NWS. `User-Agent` header is required by the API per their ToS.
nws="$(curl -fsSL --max-time 6 \
  -H "User-Agent: hairglasses-ticker (mixellburk@gmail.com)" \
  "https://api.weather.gov/alerts/active?point=${lat},${lon}" 2>/dev/null)" || exit 0

count="$(printf '%s' "$nws" | jq -r '.features | length' 2>/dev/null)"
[[ -z "$count" ]] && count=0

if (( count == 0 )); then
  # Empty file — ticker treats as _empty() and skips via backoff.
  : > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

# Pick the worst severity across all alerts.
worst="$(printf '%s' "$nws" | jq -r '
  [.features[].properties.severity] | unique |
  (if contains(["Extreme"])  then "Extreme"
   elif contains(["Severe"])  then "Severe"
   elif contains(["Moderate"]) then "Moderate"
   elif contains(["Minor"])    then "Minor"
   else "Unknown" end)' 2>/dev/null)"
worst="${worst:-Unknown}"

{
  printf '%s · %s active\n' "$worst" "$count"
  printf '%s' "$nws" | jq -r '
    .features[] | .properties |
    "\(.event)  \(.headline // .description // "" | .[0:80])"' 2>/dev/null \
    | head -5
} > "$TMPFILE"

mv "$TMPFILE" "$CACHE_FILE"
