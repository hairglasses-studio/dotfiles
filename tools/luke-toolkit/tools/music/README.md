# Music & DJ Tools

Tools for managing tracks, setlists, and music production.

## Available Scripts

| Script | Description |
|--------|-------------|
| `analyze-bpm.sh` | Analyze BPM of audio files |
| `organize-tracks.sh` | Sort tracks into folders by genre/BPM |
| `setlist.sh` | Create and manage setlists |

## Quick Start

### Analyze Track BPM

```bash
./analyze-bpm.sh track.mp3
```

### Create a Setlist

```bash
./setlist.sh create "Friday Night Set"
./setlist.sh add "track1.mp3" "track2.mp3"
./setlist.sh show
```

### Organize Your Library

```bash
./organize-tracks.sh ~/Music/DJ
```

## Using with Claude

Just ask:
```
"Analyze the BPM of all tracks in ~/Music/new"
"Create a setlist for a 2-hour house set"
"Find all tracks between 120-128 BPM"
```
