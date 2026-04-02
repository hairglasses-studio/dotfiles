#!/usr/bin/env bash
# eww-updates.sh — Check for available system updates (Arch/Manjaro)
# Outputs update count, or empty string if none available

count=$(checkupdates 2>/dev/null | wc -l)
aur_count=$(yay -Qua 2>/dev/null | wc -l)
total=$((count + aur_count))

if [[ $total -gt 0 ]]; then
  echo "$total"
else
  echo ""
fi
