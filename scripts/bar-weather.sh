#!/usr/bin/env bash
# bar-weather.sh — Weather data for the desktop menubar via wttr.in
set -euo pipefail

declare -A ICONS=(
  [113]="" [116]="" [119]="" [122]="" [143]="" [176]="" [179]="" [182]=""
  [185]="" [200]="" [227]="" [230]="" [248]="" [260]="" [263]="" [266]=""
  [281]="" [284]="" [293]="" [296]="" [299]="" [302]="" [305]="" [308]=""
  [311]="" [314]="" [317]="" [320]="" [323]="" [326]="" [329]="" [332]=""
  [335]="" [338]="" [350]="" [353]="" [356]="" [359]="" [362]="" [365]=""
  [368]="" [371]="" [374]="" [377]="" [386]="" [389]="" [392]="" [395]=""
)

weather=""
for attempt in 1 2 3; do
  weather=$(curl -sf --connect-timeout 5 --max-time 10 "https://wttr.in/?format=j1" 2>/dev/null) && break
  weather=""
  (( attempt < 3 )) && sleep $(( attempt * 2 ))
done
[[ -n "$weather" ]] || exit 0

temp=$(printf '%s' "$weather" | jq -r '.current_condition[0].temp_C // empty' 2>/dev/null) || exit 0
code=$(printf '%s' "$weather" | jq -r '.current_condition[0].weatherCode // empty' 2>/dev/null) || exit 0
desc=$(printf '%s' "$weather" | jq -r '.current_condition[0].weatherDesc[0].value // empty' 2>/dev/null) || exit 0

icon="${ICONS[$code]:-}"
printf '%s\n' "${icon} ${temp}° ${desc}"
