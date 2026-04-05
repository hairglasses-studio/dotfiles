#!/usr/bin/env python3
"""Rekordbox playlist sync script using pyrekordbox."""

import argparse
import json
import os
import sys
from pathlib import Path
from typing import Optional

try:
    from pyrekordbox import Rekordbox6Database
except ImportError:
    print(json.dumps({"error": "pyrekordbox not installed. Run: pip install pyrekordbox"}))
    sys.exit(1)


def find_or_create_folder(db, name: str, parent=None):
    """Find or create a folder in Rekordbox."""
    for pl in db.get_playlist():
        if pl.Name == name and pl.is_folder:
            if parent is None and pl.ParentID == 'root':
                return pl
            elif parent and str(pl.ParentID) == str(parent.ID):
                return pl

    # Create folder
    return db.create_playlist(name, parent=parent, is_folder=True)


def find_or_create_playlist(db, name: str, parent_folder):
    """Find or create a playlist under a folder."""
    for pl in db.get_playlist():
        if pl.Name == name and not pl.is_folder:
            if str(pl.ParentID) == str(parent_folder.ID):
                return pl

    # Create playlist
    return db.create_playlist(name, parent=parent_folder)


def get_playlist_track_ids(db, playlist) -> set:
    """Get track IDs already in a playlist."""
    track_ids = set()
    try:
        for song in db.get_playlist_songs(PlaylistID=playlist.ID):
            track_ids.add(song.ContentID)
    except Exception:
        pass
    return track_ids


def find_track_by_path(db, file_path: str):
    """Find a track in Rekordbox by file path."""
    for track in db.get_content():
        track_path = str(track.FolderPath or '')
        if file_path in track_path or track_path.endswith(os.path.basename(file_path)):
            return track
    return None


def import_track(db, file_path: str):
    """Import a track into Rekordbox collection."""
    path = Path(file_path)
    if not path.exists():
        return None

    # Check if already in collection
    existing = find_track_by_path(db, file_path)
    if existing:
        return existing

    # Import new track
    try:
        return db.add_content(str(path), Title=path.stem)
    except Exception as e:
        print(f"Failed to import {file_path}: {e}", file=sys.stderr)
        return None


def sync_mapping(db, mapping: dict, dry_run: bool = False) -> dict:
    """Sync a single playlist mapping."""
    result = {"imported": 0, "skipped": 0, "failed": 0, "total": 0}

    local_path = Path(mapping["local_path"])
    if not local_path.exists():
        return result

    # Get audio files
    audio_files = []
    for ext in ["*.aiff", "*.mp3", "*.wav", "*.flac"]:
        audio_files.extend(local_path.glob(ext))

    result["total"] = len(audio_files)

    if dry_run:
        result["skipped"] = len(audio_files)
        return result

    # Find or create parent folder and playlist
    parent_folder = find_or_create_folder(db, mapping["parent_folder"])
    playlist = find_or_create_playlist(db, mapping["playlist_name"], parent_folder)

    # Get existing tracks in playlist
    existing_ids = get_playlist_track_ids(db, playlist)

    # Import and add tracks
    for audio_file in audio_files:
        try:
            track = import_track(db, str(audio_file))
            if track is None:
                result["failed"] += 1
                continue

            if track.ID in existing_ids:
                result["skipped"] += 1
                continue

            db.add_to_playlist(playlist, track)
            result["imported"] += 1
        except Exception as e:
            print(f"Error processing {audio_file}: {e}", file=sys.stderr)
            result["failed"] += 1

    return result


def main():
    parser = argparse.ArgumentParser(description="Sync playlists to Rekordbox")
    parser.add_argument("--mappings", required=True, help="JSON array of playlist mappings")
    parser.add_argument("--dry-run", action="store_true", help="Preview without changes")
    parser.add_argument("--db-path", help="Custom Rekordbox database path")
    args = parser.parse_args()

    try:
        mappings = json.loads(args.mappings)
    except json.JSONDecodeError as e:
        print(json.dumps({"error": f"Invalid mappings JSON: {e}"}))
        sys.exit(1)

    # Connect to Rekordbox
    try:
        db_path = args.db_path or os.path.expanduser("~/Library/Pioneer/rekordbox/master.db")
        db = Rekordbox6Database(db_path)
    except Exception as e:
        print(json.dumps({"error": f"Failed to connect to Rekordbox: {e}"}))
        sys.exit(1)

    # Process each mapping
    totals = {"total": 0, "imported": 0, "skipped": 0, "failed": 0}

    for mapping in mappings:
        result = sync_mapping(db, mapping, args.dry_run)
        totals["total"] += result["total"]
        totals["imported"] += result["imported"]
        totals["skipped"] += result["skipped"]
        totals["failed"] += result["failed"]

    # Commit changes
    if not args.dry_run and totals["imported"] > 0:
        try:
            db.commit()
        except Exception as e:
            totals["error"] = f"Failed to commit: {e}"

    print(json.dumps(totals))


if __name__ == "__main__":
    main()
