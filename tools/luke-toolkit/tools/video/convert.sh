#!/bin/bash
# Video Converter - Simple FFmpeg wrapper
# Usage: convert.sh input output [preset]

input="$1"
output="$2"
preset="${3:-web}"

[ -z "$input" ] || [ -z "$output" ] && { echo "Usage: convert.sh input output [preset]"; exit 1; }
[ ! -f "$input" ] && { echo "File not found: $input"; exit 1; }

case "$preset" in
    web)      ffmpeg -i "$input" -c:v libx264 -preset medium -crf 23 -c:a aac -b:a 128k "$output" ;;
    archive)  ffmpeg -i "$input" -c:v prores_ks -profile:v 3 -c:a pcm_s16le "$output" ;;
    social)   ffmpeg -i "$input" -c:v libx264 -preset fast -crf 26 -vf "scale=-2:720" -c:a aac -b:a 96k "$output" ;;
    projector) ffmpeg -i "$input" -c:v libx264 -preset slow -crf 18 -vf "scale=-2:2160" -c:a aac -b:a 192k "$output" ;;
    *)        echo "Unknown preset: $preset (web|archive|social|projector)"; exit 1 ;;
esac

echo "Converted: $input -> $output ($preset)"
