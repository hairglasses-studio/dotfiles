#!/usr/bin/env python3
"""Rekordbox library query script using pyrekordbox."""

import argparse
import json
import os
import sys
from datetime import datetime, timedelta

try:
    from pyrekordbox import Rekordbox6Database
except ImportError:
    print(json.dumps({"error": "pyrekordbox not installed. Run: pip install pyrekordbox"}))
    sys.exit(1)


def get_db():
    """Connect to Rekordbox database."""
    db_path = os.path.expanduser("~/Library/Pioneer/rekordbox/master.db")
    return Rekordbox6Database(db_path)


def format_duration(seconds):
    """Format seconds to mm:ss."""
    if not seconds:
        return ""
    seconds = int(seconds)
    return f"{seconds // 60}:{seconds % 60:02d}"


def normalize_bpm(raw_bpm):
    """Convert raw BPM (stored as BPM*100) to actual BPM."""
    if not raw_bpm:
        return 0.0
    return raw_bpm / 100.0


def get_stats(args):
    """Get library statistics."""
    db = get_db()

    tracks = list(db.get_content())
    playlists = [p for p in db.get_playlist() if not p.is_folder]
    folders = [p for p in db.get_playlist() if p.is_folder]

    # Count cue points
    cue_count = 0
    hot_cue_count = 0
    loop_count = 0

    for track in tracks:
        try:
            cues = list(db.get_cue(ContentID=track.ID))
            for cue in cues:
                if cue.Kind == 0:
                    hot_cue_count += 1
                elif cue.Kind == 4:
                    loop_count += 1
                else:
                    cue_count += 1
        except:
            pass

    return {
        "tracks": len(tracks),
        "playlists": len(playlists),
        "folders": len(folders),
        "cue_points": cue_count,
        "hot_cues": hot_cue_count,
        "loops": loop_count,
        "db_path": os.path.expanduser("~/Library/Pioneer/rekordbox/master.db"),
    }


def get_playlists(args):
    """List all playlists with hierarchy."""
    db = get_db()
    tree = args.get("tree", True)

    playlists = list(db.get_playlist())

    def build_tree(parent_id="root", depth=0):
        result = []
        for pl in playlists:
            if str(pl.ParentID) == str(parent_id):
                track_count = 0
                if not pl.is_folder:
                    try:
                        track_count = len(list(db.get_playlist_songs(PlaylistID=pl.ID)))
                    except:
                        pass

                result.append({
                    "id": str(pl.ID),
                    "name": pl.Name,
                    "is_folder": pl.is_folder,
                    "depth": depth,
                    "track_count": track_count,
                })

                if pl.is_folder and tree:
                    result.extend(build_tree(pl.ID, depth + 1))

        return result

    return {"playlists": build_tree()}


def search_tracks(args):
    """Search tracks by various criteria."""
    db = get_db()

    query = args.get("query", "").lower()
    bpm_min = args.get("bpm_min", 0)
    bpm_max = args.get("bpm_max", 0)
    key_filter = args.get("key", "").upper()
    limit = args.get("limit", 20)

    results = []

    for track in db.get_content():
        # Text search
        if query:
            title = (track.Title or "").lower()
            artist = (track.ArtistName or "").lower()
            if query not in title and query not in artist:
                continue

        # BPM filter
        bpm = normalize_bpm(track.BPM)
        if bpm_min > 0 and bpm < bpm_min:
            continue
        if bpm_max > 0 and bpm > bpm_max:
            continue

        # Key filter
        if key_filter:
            track_key = track.KeyName or ""
            if key_filter not in track_key.upper():
                continue

        results.append({
            "id": track.ID,
            "title": track.Title or "",
            "artist": track.ArtistName or "",
            "bpm": round(bpm, 1),
            "key": track.KeyName or "",
            "duration": format_duration(track.Length),
        })

        if len(results) >= limit:
            break

    return {"tracks": results}


def get_track_info(args):
    """Get detailed track information."""
    db = get_db()

    track_id = args.get("track_id", 0)
    file_path = args.get("path", "")

    track = None

    if track_id:
        for t in db.get_content():
            if t.ID == track_id:
                track = t
                break
    elif file_path:
        for t in db.get_content():
            if file_path in str(t.FolderPath or ""):
                track = t
                break

    if not track:
        return {"error": "Track not found"}

    # Get cue points
    cues = []
    try:
        for cue in db.get_cue(ContentID=track.ID):
            cue_type = "Memory"
            if cue.Kind == 0:
                cue_type = "Hot Cue"
            elif cue.Kind == 4:
                cue_type = "Loop"

            cues.append({
                "type": cue_type,
                "name": cue.Comment or "",
                "time": format_duration(cue.InMsec),
            })
    except:
        pass

    return {
        "id": track.ID,
        "title": track.Title or "",
        "artist": track.ArtistName or "",
        "album": track.AlbumName or "",
        "bpm": round(normalize_bpm(track.BPM), 1),
        "key": track.KeyName or "",
        "duration": format_duration(track.Length),
        "path": str(track.FolderPath or ""),
        "cues": cues,
        "rating": track.Rating or 0,
        "color": track.ColorID or 0,
    }


def get_recent(args):
    """Get recently added tracks."""
    db = get_db()

    days = args.get("days", 7)
    limit = args.get("limit", 50)

    cutoff = datetime.now() - timedelta(days=days)
    results = []

    for track in db.get_content():
        try:
            date_added = track.DateCreated
            if date_added and date_added >= cutoff:
                results.append({
                    "id": track.ID,
                    "title": track.Title or "",
                    "artist": track.ArtistName or "",
                    "bpm": round(normalize_bpm(track.BPM), 1),
                    "date_added": date_added.strftime("%Y-%m-%d"),
                })
        except:
            pass

        if len(results) >= limit:
            break

    results.sort(key=lambda x: x.get("date_added", ""), reverse=True)
    return {"tracks": results[:limit]}


def export_playlist(args):
    """Export a playlist to XML format for sharing."""
    import xml.etree.ElementTree as ET
    from xml.dom import minidom

    db = get_db()
    playlist_name = args.get("playlist", "")

    if not playlist_name:
        return {"error": "playlist name required"}

    # Find the playlist
    target_playlist = None
    for pl in db.get_playlist():
        if pl.Name == playlist_name and not pl.is_folder:
            target_playlist = pl
            break

    if not target_playlist:
        return {"error": f"Playlist '{playlist_name}' not found"}

    # Get tracks from playlist
    tracks = []
    try:
        for song in db.get_playlist_songs(PlaylistID=target_playlist.ID):
            content = song.Content
            if content:
                tracks.append({
                    "id": content.ID,
                    "title": content.Title or "",
                    "artist": content.ArtistName or "",
                    "album": content.AlbumName or "",
                    "bpm": normalize_bpm(content.BPM),
                    "key": content.KeyName or "",
                    "duration": content.Length or 0,
                    "path": str(content.FolderPath or ""),
                })
    except Exception as e:
        return {"error": f"Failed to get playlist tracks: {e}"}

    # Create XML
    root = ET.Element("rekordbox_playlist")
    root.set("name", playlist_name)
    root.set("track_count", str(len(tracks)))

    for track in tracks:
        track_elem = ET.SubElement(root, "track")
        for key, value in track.items():
            track_elem.set(key, str(value))

    # Write to file
    cache_dir = os.path.expanduser("~/.cache/aftrs/rekordbox/exports")
    os.makedirs(cache_dir, exist_ok=True)

    safe_name = "".join(c if c.isalnum() or c in "- _" else "_" for c in playlist_name)
    xml_path = os.path.join(cache_dir, f"{safe_name}.xml")

    xmlstr = minidom.parseString(ET.tostring(root)).toprettyxml(indent="  ")
    with open(xml_path, "w") as f:
        f.write(xmlstr)

    return {
        "xml_path": xml_path,
        "track_count": len(tracks),
        "playlist_name": playlist_name,
    }


def import_playlist(args):
    """Import a playlist from XML format."""
    import xml.etree.ElementTree as ET

    db = get_db()
    xml_path = args.get("xml_path", "")
    folder_name = args.get("folder_name", "Shared")

    if not xml_path or not os.path.exists(xml_path):
        return {"error": f"XML file not found: {xml_path}"}

    # Parse XML
    try:
        tree = ET.parse(xml_path)
        root = tree.getroot()
    except Exception as e:
        return {"error": f"Failed to parse XML: {e}"}

    playlist_name = root.get("name", os.path.basename(xml_path).replace(".xml", ""))

    # Find or create folder structure
    folder_parts = folder_name.split("/")
    parent_folder = None

    for part in folder_parts:
        found = False
        for pl in db.get_playlist():
            if pl.Name == part and pl.is_folder:
                if parent_folder is None and str(pl.ParentID) == "root":
                    parent_folder = pl
                    found = True
                    break
                elif parent_folder and str(pl.ParentID) == str(parent_folder.ID):
                    parent_folder = pl
                    found = True
                    break

        if not found:
            parent_folder = db.create_playlist(part, parent=parent_folder, is_folder=True)

    # Find or create playlist
    target_playlist = None
    for pl in db.get_playlist():
        if pl.Name == playlist_name and not pl.is_folder:
            if parent_folder and str(pl.ParentID) == str(parent_folder.ID):
                target_playlist = pl
                break

    if not target_playlist:
        target_playlist = db.create_playlist(playlist_name, parent=parent_folder)

    # Get existing track IDs in playlist
    existing_ids = set()
    try:
        for song in db.get_playlist_songs(PlaylistID=target_playlist.ID):
            existing_ids.add(song.ContentID)
    except:
        pass

    # Import tracks
    imported = 0
    skipped = 0

    for track_elem in root.findall("track"):
        title = track_elem.get("title", "")
        artist = track_elem.get("artist", "")
        path = track_elem.get("path", "")

        # Find track in library by path or title+artist
        found_track = None
        for content in db.get_content():
            if path and path in str(content.FolderPath or ""):
                found_track = content
                break
            if content.Title == title and content.ArtistName == artist:
                found_track = content
                break

        if found_track:
            if found_track.ID in existing_ids:
                skipped += 1
            else:
                try:
                    db.add_to_playlist(target_playlist, found_track)
                    imported += 1
                except:
                    skipped += 1
        else:
            skipped += 1

    # Commit changes
    if imported > 0:
        try:
            db.commit()
        except Exception as e:
            return {"error": f"Failed to commit: {e}"}

    return {
        "imported": imported,
        "skipped": skipped,
        "playlist_name": playlist_name,
        "folder": folder_name,
    }


def main():
    parser = argparse.ArgumentParser(description="Rekordbox library query tool")
    parser.add_argument("--action", required=True, choices=[
        "stats", "playlists", "search", "track_info", "recent",
        "export_playlist", "import_playlist"
    ])
    parser.add_argument("--args", default="{}", help="JSON arguments for the action")
    args = parser.parse_args()

    try:
        action_args = json.loads(args.args)
    except json.JSONDecodeError:
        print(json.dumps({"error": "Invalid JSON arguments"}))
        sys.exit(1)

    actions = {
        "stats": get_stats,
        "playlists": get_playlists,
        "search": search_tracks,
        "track_info": get_track_info,
        "recent": get_recent,
        "export_playlist": export_playlist,
        "import_playlist": import_playlist,
    }

    try:
        result = actions[args.action](action_args)
        print(json.dumps(result))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)


if __name__ == "__main__":
    main()
