# spotify

> Spotify Web API integration for music discovery and track metadata

**14 tools**

## Tools

- [`aftrs_spotify_album`](#aftrs-spotify-album)
- [`aftrs_spotify_artist`](#aftrs-spotify-artist)
- [`aftrs_spotify_artist_top`](#aftrs-spotify-artist-top)
- [`aftrs_spotify_features`](#aftrs-spotify-features)
- [`aftrs_spotify_genres`](#aftrs-spotify-genres)
- [`aftrs_spotify_health`](#aftrs-spotify-health)
- [`aftrs_spotify_key`](#aftrs-spotify-key)
- [`aftrs_spotify_new_releases`](#aftrs-spotify-new-releases)
- [`aftrs_spotify_recommendations`](#aftrs-spotify-recommendations)
- [`aftrs_spotify_related`](#aftrs-spotify-related)
- [`aftrs_spotify_search`](#aftrs-spotify-search)
- [`aftrs_spotify_status`](#aftrs-spotify-status)
- [`aftrs_spotify_track`](#aftrs-spotify-track)
- [`aftrs_spotify_track_analysis`](#aftrs-spotify-track-analysis)

---

## aftrs_spotify_album

Get detailed information about a Spotify album

**Complexity:** simple

**Tags:** `spotify`, `music`, `album`, `metadata`

**Use Cases:**
- Get album details
- View album artwork

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `album_id` | string | Yes | Spotify album ID |

### Example

```json
{
  "album_id": "example"
}
```

---

## aftrs_spotify_artist

Get detailed information about a Spotify artist

**Complexity:** simple

**Tags:** `spotify`, `music`, `artist`, `metadata`

**Use Cases:**
- Get artist details
- View artist genres

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `artist_id` | string | Yes | Spotify artist ID |

### Example

```json
{
  "artist_id": "example"
}
```

---

## aftrs_spotify_artist_top

Get an artist's top tracks

**Complexity:** simple

**Tags:** `spotify`, `music`, `artist`, `top-tracks`

**Use Cases:**
- Discover artist's best tracks
- Build playlists

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `artist_id` | string | Yes | Spotify artist ID |

### Example

```json
{
  "artist_id": "example"
}
```

---

## aftrs_spotify_features

Get audio features for a track (BPM, key, energy, danceability)

**Complexity:** simple

**Tags:** `spotify`, `music`, `audio-features`, `bpm`, `key`

**Use Cases:**
- Get BPM and key
- Analyze track mood
- DJ preparation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `track_id` | string | Yes | Spotify track ID |

### Example

```json
{
  "track_id": "example"
}
```

---

## aftrs_spotify_genres

Get available genre seeds for recommendations

**Complexity:** simple

**Tags:** `spotify`, `music`, `genres`, `discovery`

**Use Cases:**
- List available genres
- Find genre categories

---

## aftrs_spotify_health

Check Spotify API health and configuration

**Complexity:** simple

**Tags:** `spotify`, `health`, `diagnostics`

**Use Cases:**
- Diagnose API issues
- Check configuration

---

## aftrs_spotify_key

Convert Spotify key number to musical notation

**Complexity:** simple

**Tags:** `spotify`, `music`, `key`, `utility`

**Use Cases:**
- Convert key notation
- DJ preparation

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | integer | Yes | Spotify key number (0-11) |
| `mode` | integer | Yes | Mode (0 = minor, 1 = major) |

### Example

```json
{
  "key": 0,
  "mode": 0
}
```

---

## aftrs_spotify_new_releases

Get new album releases

**Complexity:** simple

**Tags:** `spotify`, `music`, `new-releases`, `discovery`

**Use Cases:**
- Discover new music
- Stay current with releases

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum results (default: 20, max: 50) |

### Example

```json
{
  "limit": 0
}
```

---

## aftrs_spotify_recommendations

Get track recommendations based on seed tracks or artists

**Complexity:** simple

**Tags:** `spotify`, `music`, `recommendations`, `discovery`

**Use Cases:**
- Get similar tracks
- Build playlists
- Discover new music

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum results (default: 20, max: 100) |
| `seed_artists` | string |  | Comma-separated Spotify artist IDs (up to 5) |
| `seed_tracks` | string |  | Comma-separated Spotify track IDs (up to 5) |

### Example

```json
{
  "limit": 0,
  "seed_artists": "example",
  "seed_tracks": "example"
}
```

---

## aftrs_spotify_related

Get artists related to a given artist

**Complexity:** simple

**Tags:** `spotify`, `music`, `discovery`, `related-artists`

**Use Cases:**
- Discover similar artists
- Expand music library

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `artist_id` | string | Yes | Spotify artist ID |

### Example

```json
{
  "artist_id": "example"
}
```

---

## aftrs_spotify_search

Search Spotify for tracks, artists, or albums

**Complexity:** simple

**Tags:** `spotify`, `music`, `search`, `discovery`

**Use Cases:**
- Find tracks by name
- Search for artists
- Discover new music

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | integer |  | Maximum results (default: 20, max: 50) |
| `query` | string | Yes | Search query |
| `type` | string |  | Search type: track, artist, album, or comma-separated (default: track) |

### Example

```json
{
  "limit": 0,
  "query": "example",
  "type": "example"
}
```

---

## aftrs_spotify_status

Get Spotify API connection status and token info

**Complexity:** simple

**Tags:** `spotify`, `music`, `status`, `api`

**Use Cases:**
- Check Spotify API connectivity
- Verify authentication

---

## aftrs_spotify_track

Get detailed information about a Spotify track

**Complexity:** simple

**Tags:** `spotify`, `music`, `track`, `metadata`

**Use Cases:**
- Get track details
- View track metadata

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `track_id` | string | Yes | Spotify track ID |

### Example

```json
{
  "track_id": "example"
}
```

---

## aftrs_spotify_track_analysis

Get combined track details and audio features in one call

**Complexity:** simple

**Tags:** `spotify`, `music`, `analysis`, `bpm`, `key`

**Use Cases:**
- Full track analysis
- DJ preparation
- Sync visuals to music

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `track_id` | string | Yes | Spotify track ID |

### Example

```json
{
  "track_id": "example"
}
```

---

