"""
CR8 Audio Analysis Lambda Handler

Processes tracks from the DynamoDB analysis queue:
1. Claims pending items from cr8_sync_queue
2. Downloads audio from S3
3. Analyzes with librosa (BPM, key, energy)
4. Updates cr8_tracks with results
5. Marks queue item as completed
"""

import boto3
import json
import logging
import os
import tempfile
from datetime import datetime, timezone
from decimal import Decimal
from typing import Any, Dict, List, Optional, Tuple

import librosa
import numpy as np

logger = logging.getLogger()
logger.setLevel(logging.INFO)

# Configuration from environment
S3_BUCKET = os.environ.get("S3_BUCKET", "cr8-music-storage")
TRACKS_TABLE = os.environ.get("TRACKS_TABLE", "cr8_tracks")
QUEUE_TABLE = os.environ.get("QUEUE_TABLE", "cr8_sync_queue")
BATCH_SIZE = int(os.environ.get("BATCH_SIZE", "20"))
ANALYSIS_DURATION = int(os.environ.get("ANALYSIS_DURATION", "60"))

# Key mapping for Camelot notation
KEY_TO_CAMELOT = {
    'G#m': '1A', 'Abm': '1A', 'D#m': '2A', 'Ebm': '2A', 'A#m': '3A', 'Bbm': '3A',
    'Fm': '4A', 'Cm': '5A', 'Gm': '6A', 'Dm': '7A', 'Am': '8A', 'Em': '9A',
    'Bm': '10A', 'F#m': '11A', 'Gbm': '11A', 'C#m': '12A', 'Dbm': '12A',
    'B': '1B', 'Cb': '1B', 'F#': '2B', 'Gb': '2B', 'C#': '3B', 'Db': '3B',
    'G#': '4B', 'Ab': '4B', 'D#': '5B', 'Eb': '5B', 'A#': '6B', 'Bb': '6B',
    'F': '7B', 'C': '8B', 'G': '9B', 'D': '10B', 'A': '11B', 'E': '12B',
}

CHROMA_TO_KEY = {
    0: ('C', 'Am'), 1: ('C#', 'A#m'), 2: ('D', 'Bm'), 3: ('D#', 'Cm'),
    4: ('E', 'C#m'), 5: ('F', 'Dm'), 6: ('F#', 'D#m'), 7: ('G', 'Em'),
    8: ('G#', 'Fm'), 9: ('A', 'F#m'), 10: ('A#', 'Gm'), 11: ('B', 'G#m'),
}

# AWS clients
dynamodb = boto3.resource("dynamodb")
s3 = boto3.client("s3")
tracks_table = dynamodb.Table(TRACKS_TABLE)
queue_table = dynamodb.Table(QUEUE_TABLE)


def get_pending_items(limit: int) -> List[Dict]:
    """Get pending analysis queue items."""
    items = []
    last_key = None

    while len(items) < limit:
        kwargs = {
            "FilterExpression": "queue_type = :qt AND #status = :s",
            "ExpressionAttributeNames": {"#status": "status"},
            "ExpressionAttributeValues": {":qt": "analysis", ":s": "pending"},
            "Limit": min(limit - len(items), 100),
        }
        if last_key:
            kwargs["ExclusiveStartKey"] = last_key

        response = queue_table.scan(**kwargs)
        items.extend(response.get("Items", []))

        last_key = response.get("LastEvaluatedKey")
        if not last_key:
            break

    return items[:limit]


def claim_item(item_id: str) -> bool:
    """Claim a queue item for processing."""
    try:
        queue_table.update_item(
            Key={"id": item_id},
            UpdateExpression="SET #status = :s, started_at = :t",
            ExpressionAttributeNames={"#status": "status"},
            ExpressionAttributeValues={
                ":s": "processing",
                ":t": datetime.now(timezone.utc).isoformat(),
                ":pending": "pending",
            },
            ConditionExpression="#status = :pending",
        )
        return True
    except dynamodb.meta.client.exceptions.ConditionalCheckFailedException:
        return False
    except Exception as e:
        logger.warning(f"Failed to claim item {item_id}: {e}")
        return False


def complete_item(item_id: str, success: bool, error: str = None):
    """Mark queue item as completed or failed."""
    status = "completed" if success else "failed"
    update_expr = "SET #status = :s, completed_at = :t"
    expr_values = {
        ":s": status,
        ":t": datetime.now(timezone.utc).isoformat(),
    }

    if error:
        update_expr += ", error_message = :e"
        expr_values[":e"] = error

    try:
        queue_table.update_item(
            Key={"id": item_id},
            UpdateExpression=update_expr,
            ExpressionAttributeNames={"#status": "status"},
            ExpressionAttributeValues=expr_values,
        )
    except Exception as e:
        logger.error(f"Failed to complete item {item_id}: {e}")


def download_from_s3(s3_path: str, local_path: str) -> bool:
    """Download file from S3."""
    try:
        if s3_path.startswith("s3://"):
            s3_path = s3_path.replace(f"s3://{S3_BUCKET}/", "")
        elif s3_path.startswith("/"):
            s3_path = s3_path[1:]

        logger.info(f"Downloading s3://{S3_BUCKET}/{s3_path}")
        s3.download_file(S3_BUCKET, s3_path, local_path)
        return True
    except Exception as e:
        logger.error(f"S3 download failed: {e}")
        return False


def analyze_audio(audio_path: str, duration: int = 60) -> Dict[str, Any]:
    """Analyze audio file for BPM, key, energy."""
    try:
        logger.info(f"Analyzing: {audio_path}")
        y, sr = librosa.load(audio_path, sr=22050, duration=duration, mono=True)

        if len(y) == 0:
            return {"success": False, "error": "Empty audio file"}

        # BPM detection
        onset_env = librosa.onset.onset_strength(y=y, sr=sr)
        tempo, beats = librosa.beat.beat_track(onset_envelope=onset_env, sr=sr)
        if isinstance(tempo, np.ndarray):
            tempo = float(tempo[0]) if len(tempo) > 0 else 120.0
        else:
            tempo = float(tempo)

        # Key detection
        chroma = librosa.feature.chroma_cqt(y=y, sr=sr)
        chroma_mean = np.mean(chroma, axis=1)
        pitch_class = int(np.argmax(chroma_mean))

        major_profile = np.array([6.35, 2.23, 3.48, 2.33, 4.38, 4.09, 2.52, 5.19, 2.39, 3.66, 2.29, 2.88])
        minor_profile = np.array([6.33, 2.68, 3.52, 5.38, 2.60, 3.53, 2.54, 4.75, 3.98, 2.69, 3.34, 3.17])

        major_corr = np.correlate(chroma_mean, np.roll(major_profile, pitch_class))[0]
        minor_corr = np.correlate(chroma_mean, np.roll(minor_profile, pitch_class))[0]

        is_major = major_corr > minor_corr
        major_key, minor_key = CHROMA_TO_KEY[pitch_class]
        key = major_key if is_major else minor_key
        camelot = KEY_TO_CAMELOT.get(key)

        # Energy
        rms = librosa.feature.rms(y=y)
        energy = min(1.0, float(np.mean(rms)) * 2)

        # Duration
        duration_ms = int(librosa.get_duration(y=y, sr=sr) * 1000)

        return {
            "success": True,
            "bpm": round(tempo, 1),
            "key": key,
            "camelot_key": camelot,
            "energy": round(energy, 3),
            "duration_ms": duration_ms,
        }

    except Exception as e:
        logger.error(f"Analysis failed: {e}")
        return {"success": False, "error": str(e)}


def update_track(track_id: str, analysis: Dict[str, Any]) -> bool:
    """Update track with analysis results."""
    try:
        update_expr = "SET bpm = :bpm, #key = :key, camelot_key = :camelot, energy = :energy, analyzed_at = :t"
        expr_values = {
            ":bpm": Decimal(str(analysis["bpm"])),
            ":key": analysis["key"],
            ":camelot": analysis.get("camelot_key"),
            ":energy": Decimal(str(analysis["energy"])),
            ":t": datetime.now(timezone.utc).isoformat(),
        }

        if analysis.get("duration_ms"):
            update_expr += ", duration_ms = :dur"
            expr_values[":dur"] = analysis["duration_ms"]

        tracks_table.update_item(
            Key={"id": track_id},
            UpdateExpression=update_expr,
            ExpressionAttributeNames={"#key": "key"},
            ExpressionAttributeValues=expr_values,
        )
        return True
    except Exception as e:
        logger.error(f"Failed to update track {track_id}: {e}")
        return False


def process_item(item: Dict) -> Tuple[bool, str]:
    """Process a single queue item."""
    item_id = item["id"]
    track_id = item.get("track_id")
    track_path = item.get("track_path")

    if not track_id:
        return False, "Missing track_id"

    if not track_path:
        # Try to get from tracks table
        try:
            response = tracks_table.get_item(Key={"id": track_id})
            track = response.get("Item")
            if track:
                track_path = track.get("storage_path") or track.get("file_path")
        except Exception as e:
            return False, f"Failed to get track: {e}"

    if not track_path:
        return False, "No storage path for track"

    # Download and analyze - preserve original extension for correct decoder
    ext = os.path.splitext(track_path)[1] or ".mp3"
    with tempfile.NamedTemporaryFile(suffix=ext, delete=False) as tmp:
        tmp_path = tmp.name

    try:
        if not download_from_s3(track_path, tmp_path):
            return False, "S3 download failed"

        analysis = analyze_audio(tmp_path, ANALYSIS_DURATION)

        if not analysis.get("success"):
            return False, analysis.get("error", "Analysis failed")

        if not update_track(track_id, analysis):
            return False, "Failed to update track"

        logger.info(f"Analyzed track {track_id}: BPM={analysis['bpm']}, Key={analysis['key']} ({analysis.get('camelot_key')})")
        return True, None

    finally:
        if os.path.exists(tmp_path):
            os.remove(tmp_path)


def lambda_handler(event: Dict, context) -> Dict:
    """Lambda entry point."""
    batch_size = event.get("batch_size", BATCH_SIZE)
    source = event.get("source", "manual")

    logger.info(f"Starting analysis worker (batch={batch_size}, source={source})")

    total_processed = 0
    total_success = 0
    total_failed = 0

    items = get_pending_items(batch_size)

    if not items:
        logger.info("Queue empty, nothing to process")
        return {
            "statusCode": 200,
            "body": json.dumps({
                "message": "Queue empty",
                "processed": 0,
                "success": 0,
                "failed": 0,
            })
        }

    logger.info(f"Processing {len(items)} items...")

    for item in items:
        item_id = item["id"]
        track_id = item.get("track_id", "unknown")

        if not claim_item(item_id):
            logger.warning(f"Could not claim item {item_id}, skipping")
            continue

        success, error = process_item(item)
        complete_item(item_id, success, error)

        total_processed += 1
        if success:
            total_success += 1
        else:
            total_failed += 1
            logger.warning(f"Failed track {track_id}: {error}")

    result = {
        "processed": total_processed,
        "success": total_success,
        "failed": total_failed,
    }

    logger.info(f"Worker finished: {total_processed} processed, {total_success} success, {total_failed} failed")

    return {
        "statusCode": 200,
        "body": json.dumps(result)
    }
