#!/usr/bin/env bash
# eww-weather.sh — Weather data for eww bar via wttr.in
set -euo pipefail

# Weather code to Nerd Font icon mapping
declare -A ICONS=(
  [113]=""   # Clear/Sunny
  [116]=""   # Partly cloudy
  [119]=""   # Cloudy
  [122]=""   # Overcast
  [143]=""   # Mist
  [176]=""   # Light rain
  [179]=""   # Light snow
  [182]=""   # Light sleet
  [185]=""   # Light sleet
  [200]=""   # Thunder
  [227]=""   # Blowing snow
  [230]=""   # Blizzard
  [248]=""   # Fog
  [260]=""   # Freezing fog
  [263]=""   # Light drizzle
  [266]=""   # Light drizzle
  [281]=""   # Freezing drizzle
  [284]=""   # Heavy freezing drizzle
  [293]=""   # Light rain
  [296]=""   # Light rain
  [299]=""   # Moderate rain
  [302]=""   # Heavy rain
  [305]=""   # Heavy rain
  [308]=""   # Heavy rain
  [311]=""   # Freezing rain
  [314]=""   # Heavy freezing rain
  [317]=""   # Light sleet
  [320]=""   # Moderate sleet
  [323]=""   # Light snow
  [326]=""   # Light snow
  [329]=""   # Heavy snow
  [332]=""   # Heavy snow
  [335]=""   # Heavy snow
  [338]=""   # Heavy snow
  [350]=""   # Ice pellets
  [353]=""   # Light showers
  [356]=""   # Heavy showers
  [359]=""   # Torrential rain
  [362]=""   # Light sleet showers
  [365]=""   # Heavy sleet showers
  [368]=""   # Light snow showers
  [371]=""   # Heavy snow showers
  [374]=""   # Ice pellets
  [377]=""   # Ice pellets
  [386]=""   # Thunder + light rain
  [389]=""   # Thunder + heavy rain
  [392]=""   # Thunder + light snow
  [395]=""   # Thunder + heavy snow
)

weather=$(curl -sf "https://wttr.in/?format=j1" 2>/dev/null) || exit 0

temp=$(echo "$weather" | jq -r '.current_condition[0].temp_C // empty' 2>/dev/null) || exit 0
code=$(echo "$weather" | jq -r '.current_condition[0].weatherCode // empty' 2>/dev/null) || exit 0
desc=$(echo "$weather" | jq -r '.current_condition[0].weatherDesc[0].value // empty' 2>/dev/null) || exit 0

icon="${ICONS[$code]:-}"
echo "${icon} ${temp}°"
