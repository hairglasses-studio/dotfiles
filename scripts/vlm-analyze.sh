#!/usr/bin/env bash
# vlm-analyze.sh — Analyze a screenshot using Claude's vision API
# Usage: vlm-analyze.sh <image_path> [prompt]
# Requires: ANTHROPIC_API_KEY env var, curl, base64
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

hg_require curl base64

# ── Args ──────────────────────────────────────────
IMAGE_PATH="${1:-}"
PROMPT="${2:-Describe this desktop screenshot. Evaluate the visual quality, color consistency with a cyberpunk neon theme (Snazzy palette: cyan #57c7ff, magenta #ff6ac1, green #5af78e, yellow #f3f99d, red #ff5c57 on black #000000), and any visual issues or improvements.}"

if [[ -z "$IMAGE_PATH" ]]; then
  hg_die "Usage: vlm-analyze.sh <image_path> [prompt]"
fi

if [[ ! -f "$IMAGE_PATH" ]]; then
  hg_die "Image not found: $IMAGE_PATH"
fi

if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
  hg_die "ANTHROPIC_API_KEY environment variable is not set"
fi

# ── Detect media type ─────────────────────────────
case "${IMAGE_PATH,,}" in
  *.png)  MEDIA_TYPE="image/png" ;;
  *.jpg|*.jpeg) MEDIA_TYPE="image/jpeg" ;;
  *.gif)  MEDIA_TYPE="image/gif" ;;
  *.webp) MEDIA_TYPE="image/webp" ;;
  *)      MEDIA_TYPE="image/png" ;;
esac

# ── Base64-encode the image ───────────────────────
hg_info "Encoding image: $IMAGE_PATH"
IMAGE_B64=$(base64 -w 0 "$IMAGE_PATH" 2>/dev/null || base64 "$IMAGE_PATH" 2>/dev/null)

# ── Build JSON payload ────────────────────────────
# Use a temp file to avoid shell escaping issues with large base64 strings
PAYLOAD_FILE=$(mktemp /tmp/vlm-payload.XXXXXX.json)
trap 'rm -f "$PAYLOAD_FILE"' EXIT

cat > "$PAYLOAD_FILE" <<JSONEOF
{
  "model": "claude-sonnet-4-20250514",
  "max_tokens": 1024,
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "image",
          "source": {
            "type": "base64",
            "media_type": "$MEDIA_TYPE",
            "data": "$IMAGE_B64"
          }
        },
        {
          "type": "text",
          "text": $(printf '%s' "$PROMPT" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read()))')
        }
      ]
    }
  ]
}
JSONEOF

# ── Send request ──────────────────────────────────
hg_info "Sending to Claude API..."

RESPONSE=$(curl -s -w "\n%{http_code}" \
  https://api.anthropic.com/v1/messages \
  -H "content-type: application/json" \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -d @"$PAYLOAD_FILE")

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [[ "$HTTP_CODE" != "200" ]]; then
  hg_error "API request failed (HTTP $HTTP_CODE)"
  echo "$BODY" >&2
  exit 1
fi

# ── Extract text from response ────────────────────
ANALYSIS=$(echo "$BODY" | python3 -c '
import json, sys
data = json.load(sys.stdin)
for block in data.get("content", []):
    if block.get("type") == "text":
        print(block["text"])
' 2>/dev/null)

if [[ -z "$ANALYSIS" ]]; then
  hg_error "Failed to parse API response"
  echo "$BODY" >&2
  exit 1
fi

echo ""
echo "$ANALYSIS"
