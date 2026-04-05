#!/bin/bash
# Beatport Sync Status Report
# Shows status of all running sync operations

cd "$(dirname "$0")"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║              BEATPORT SYNC STATUS REPORT                     ║"
echo "╠══════════════════════════════════════════════════════════════╣"
echo "║ Time: $(date '+%Y-%m-%d %H:%M:%S')                                    ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Download sync status
echo "📥 DOWNLOAD SYNC"
echo "────────────────────────────────────────────────────────────────"
if [[ -f /tmp/beatport-sync.log ]]; then
  LAST_LINE=$(tail -1 /tmp/beatport-sync.log | tr '\r' '\n' | tail -1)
  echo "$LAST_LINE"
else
  echo "No download sync in progress"
fi
echo ""

# Likes/follows sync status
echo "❤️  LIKES/FOLLOWS SYNC"
echo "────────────────────────────────────────────────────────────────"
if [[ -f /tmp/beatport-likes-full.log ]]; then
  LIKED=$(grep -c '✓ Liked track' /tmp/beatport-likes-full.log 2>/dev/null) || LIKED=0
  FOLLOWED=$(grep -c '✓ Followed artist' /tmp/beatport-likes-full.log 2>/dev/null) || FOLLOWED=0
  ALREADY_LIKED=$(grep -c 'already liked' /tmp/beatport-likes-full.log 2>/dev/null) || ALREADY_LIKED=0
  ALREADY_FOLLOWING=$(grep -c 'Already following' /tmp/beatport-likes-full.log 2>/dev/null) || ALREADY_FOLLOWING=0
  ERRORS=$(grep -c '✗ Error' /tmp/beatport-likes-full.log 2>/dev/null) || ERRORS=0

  echo "Tracks liked:      $LIKED (already: $ALREADY_LIKED)"
  echo "Artists followed:  $FOLLOWED (already: $ALREADY_FOLLOWING)"
  echo "Errors:            $ERRORS"
  echo ""
  echo "Last activity:"
  tail -3 /tmp/beatport-likes-full.log 2>/dev/null | head -2
else
  echo "No likes/follows sync in progress"
fi
echo ""

# Progress file status
echo "💾 PROGRESS TRACKING"
echo "────────────────────────────────────────────────────────────────"
if [[ -f .beatport-progress.json ]]; then
  cat .beatport-progress.json | python3 -c "
import json, sys
data = json.load(sys.stdin)
for playlist, stats in data.items():
    liked = len(stats.get('liked', []))
    followed = len(stats.get('followed', []))
    failed = len(stats.get('failed', {}))
    print(f'Playlist {playlist}: {liked} liked, {followed} followed, {failed} failed')
" 2>/dev/null || echo "Progress file exists but couldn't parse"
else
  echo "No progress file yet"
fi
echo ""

# Running processes
echo "🔄 RUNNING PROCESSES"
echo "────────────────────────────────────────────────────────────────"
ps aux | grep -E "(beatport-sync|sync-likes|playwright)" | grep -v grep | awk '{print $2, $11, $12}' || echo "No sync processes running"
echo ""

echo "────────────────────────────────────────────────────────────────"
echo "Monitor logs:"
echo "  Downloads:      tail -f /tmp/beatport-sync.log"
echo "  Likes/Follows:  tail -f /tmp/beatport-likes-full.log"
