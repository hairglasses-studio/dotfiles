#!/bin/bash
# Sync likes/follows for a Beatport playlist using API + Playwright
# Usage: ./sync-likes.sh <playlist-id> [--tracks-only|--artists-only|--resume]
set -e

cd "$(dirname "$0")"
PLAYLIST_ID="${1:-4014989}"
MODE="${2:-both}"

export PLAYWRIGHT_BROWSERS_PATH=./node_modules/playwright-core/.local-browsers
export BEATPORT_USERNAME="${BEATPORT_USERNAME:-hairglasses}"
export BEATPORT_PASSWORD="${BEATPORT_PASSWORD:-M@kawel1M@kawel1!}"

echo "=== Beatport Playlist Sync ==="
echo "Playlist: $PLAYLIST_ID"
echo "Mode: $MODE"
echo "Time: $(date)"
echo ""

# Show current progress
echo "Current progress:"
node index.js status --playlist="$PLAYLIST_ID" 2>/dev/null || echo "No previous progress"
echo ""

echo "=== Fetching playlist $PLAYLIST_ID via API ==="
IDS=$(../../bin/beatport-likes "$PLAYLIST_ID")
TRACKS=$(echo "$IDS" | grep '^TRACKS=' | cut -d= -f2)
ARTISTS=$(echo "$IDS" | grep '^ARTISTS=' | cut -d= -f2)
echo "$IDS" | grep '^#'

if [[ "$MODE" != "--artists-only" ]]; then
  echo ""
  echo "=== Liking tracks ==="
  node index.js like-tracks "$TRACKS" --playlist="$PLAYLIST_ID" 2>&1 | tee -a /tmp/beatport-likes-tracks.log
fi

if [[ "$MODE" != "--tracks-only" ]]; then
  echo ""
  echo "=== Following artists ==="
  node index.js follow-artists "$ARTISTS" --playlist="$PLAYLIST_ID" 2>&1 | tee -a /tmp/beatport-likes-artists.log
fi

echo ""
echo "=== Final Summary ==="
node index.js status --playlist="$PLAYLIST_ID"
echo ""
echo "Completed at: $(date)"
