// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/cache"
	"github.com/hairglasses-studio/hg-mcp/pkg/httpclient"
)

// Spotify caches — reduce API calls for frequently-polled data
var (
	spotifyGenresCache = cache.New[[]string](5 * time.Minute)        // Genre list rarely changes
	spotifyStatusCache = cache.New[*SpotifyStatus](15 * time.Second) // Status polled by dashboards
)

// SpotifyClient provides access to Spotify Web API
type SpotifyClient struct {
	clientID     string
	clientSecret string
	accessToken  string
	tokenExpiry  time.Time
	httpClient   *http.Client
	mu           sync.RWMutex
}

// SpotifyStatus represents API connection status
type SpotifyStatus struct {
	Connected   bool   `json:"connected"`
	HasToken    bool   `json:"has_token"`
	TokenExpiry string `json:"token_expiry,omitempty"`
}

// SpotifyTrack represents a track
type SpotifyTrack struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Artists     []SpotifyArtist `json:"artists"`
	Album       SpotifyAlbum    `json:"album"`
	DurationMS  int             `json:"duration_ms"`
	Popularity  int             `json:"popularity"`
	PreviewURL  string          `json:"preview_url,omitempty"`
	ExternalURL string          `json:"external_url,omitempty"`
	URI         string          `json:"uri"`
}

// SpotifyArtist represents an artist
type SpotifyArtist struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Genres      []string `json:"genres,omitempty"`
	Popularity  int      `json:"popularity,omitempty"`
	Followers   int      `json:"followers,omitempty"`
	ExternalURL string   `json:"external_url,omitempty"`
	URI         string   `json:"uri"`
}

// SpotifyAlbum represents an album
type SpotifyAlbum struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Artists     []SpotifyArtist `json:"artists,omitempty"`
	ReleaseDate string          `json:"release_date,omitempty"`
	TotalTracks int             `json:"total_tracks,omitempty"`
	Images      []SpotifyImage  `json:"images,omitempty"`
	ExternalURL string          `json:"external_url,omitempty"`
	URI         string          `json:"uri"`
}

// SpotifyImage represents an image
type SpotifyImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// SpotifyPlaylist represents a playlist
type SpotifyPlaylist struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Owner       SpotifyUser    `json:"owner"`
	Tracks      SpotifyTracks  `json:"tracks"`
	Public      bool           `json:"public"`
	Images      []SpotifyImage `json:"images,omitempty"`
	ExternalURL string         `json:"external_url,omitempty"`
	URI         string         `json:"uri"`
}

// SpotifyUser represents a user
type SpotifyUser struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
	URI         string `json:"uri"`
}

// SpotifyTracks represents tracks in a playlist
type SpotifyTracks struct {
	Total int `json:"total"`
}

// SpotifyAudioFeatures represents audio features for a track
type SpotifyAudioFeatures struct {
	ID               string  `json:"id"`
	Tempo            float64 `json:"tempo"`
	Key              int     `json:"key"`
	Mode             int     `json:"mode"`
	TimeSignature    int     `json:"time_signature"`
	Danceability     float64 `json:"danceability"`
	Energy           float64 `json:"energy"`
	Valence          float64 `json:"valence"`
	Loudness         float64 `json:"loudness"`
	Speechiness      float64 `json:"speechiness"`
	Acousticness     float64 `json:"acousticness"`
	Instrumentalness float64 `json:"instrumentalness"`
	Liveness         float64 `json:"liveness"`
}

// SpotifySearchResult represents search results
type SpotifySearchResult struct {
	Tracks  []SpotifyTrack  `json:"tracks,omitempty"`
	Artists []SpotifyArtist `json:"artists,omitempty"`
	Albums  []SpotifyAlbum  `json:"albums,omitempty"`
}

// SpotifyHealth represents health status
type SpotifyHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	HasCredentials  bool     `json:"has_credentials"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

var (
	spotifyClientSingleton *SpotifyClient
	spotifyClientOnce      sync.Once
	spotifyClientErr       error

	// TestOverrideSpotifyClient, when non-nil, is returned by GetSpotifyClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideSpotifyClient *SpotifyClient
)

// GetSpotifyClient returns the singleton Spotify client.
func GetSpotifyClient() (*SpotifyClient, error) {
	if TestOverrideSpotifyClient != nil {
		return TestOverrideSpotifyClient, nil
	}
	spotifyClientOnce.Do(func() {
		spotifyClientSingleton, spotifyClientErr = NewSpotifyClient()
	})
	return spotifyClientSingleton, spotifyClientErr
}

// NewTestSpotifyClient creates an in-memory test client.
func NewTestSpotifyClient() *SpotifyClient {
	return &SpotifyClient{
		clientID:     "test-client-id",
		clientSecret: "test-client-secret",
		httpClient:   httpclient.Fast(),
	}
}

// NewSpotifyClient creates a new Spotify client
func NewSpotifyClient() (*SpotifyClient, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET required")
	}

	return &SpotifyClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   httpclient.Standard(),
	}, nil
}

// authenticate gets an access token using client credentials flow
func (c *SpotifyClient) authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if token is still valid
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	// Request new token
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed: %s", resp.Status)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return nil
}

// doRequest makes an authenticated API request
func (c *SpotifyClient) doRequest(ctx context.Context, method, endpoint string) (*http.Response, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, err
	}

	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, method, "https://api.spotify.com/v1"+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// GetStatus returns API connection status (cached 15s)
func (c *SpotifyClient) GetStatus(ctx context.Context) (*SpotifyStatus, error) {
	return spotifyStatusCache.GetOrFetch(ctx, func(ctx context.Context) (*SpotifyStatus, error) {
		status := &SpotifyStatus{
			Connected: false,
			HasToken:  false,
		}

		if err := c.authenticate(ctx); err != nil {
			return status, nil
		}

		c.mu.RLock()
		status.HasToken = c.accessToken != ""
		if !c.tokenExpiry.IsZero() {
			status.TokenExpiry = c.tokenExpiry.Format(time.RFC3339)
		}
		c.mu.RUnlock()

		// Test API connection
		resp, err := c.doRequest(ctx, "GET", "/browse/new-releases?limit=1")
		if err != nil {
			return status, nil
		}
		defer resp.Body.Close()

		status.Connected = resp.StatusCode == http.StatusOK

		return status, nil
	})
}

// Search searches for tracks, artists, and albums
func (c *SpotifyClient) Search(ctx context.Context, query string, types []string, limit int) (*SpotifySearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	typeStr := "track"
	if len(types) > 0 {
		typeStr = strings.Join(types, ",")
	}

	endpoint := fmt.Sprintf("/search?q=%s&type=%s&limit=%d", url.QueryEscape(query), typeStr, limit)

	resp, err := c.doRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed: %s", resp.Status)
	}

	var result struct {
		Tracks struct {
			Items []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				DurationMS int    `json:"duration_ms"`
				Popularity int    `json:"popularity"`
				PreviewURL string `json:"preview_url"`
				URI        string `json:"uri"`
				Artists    []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					URI  string `json:"uri"`
				} `json:"artists"`
				Album struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					ReleaseDate string `json:"release_date"`
					URI         string `json:"uri"`
				} `json:"album"`
				ExternalURLs struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"items"`
		} `json:"tracks"`
		Artists struct {
			Items []struct {
				ID         string   `json:"id"`
				Name       string   `json:"name"`
				Genres     []string `json:"genres"`
				Popularity int      `json:"popularity"`
				URI        string   `json:"uri"`
				Followers  struct {
					Total int `json:"total"`
				} `json:"followers"`
				ExternalURLs struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"items"`
		} `json:"artists"`
		Albums struct {
			Items []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				ReleaseDate string `json:"release_date"`
				TotalTracks int    `json:"total_tracks"`
				URI         string `json:"uri"`
				Artists     []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					URI  string `json:"uri"`
				} `json:"artists"`
				ExternalURLs struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"items"`
		} `json:"albums"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	searchResult := &SpotifySearchResult{}

	for _, t := range result.Tracks.Items {
		track := SpotifyTrack{
			ID:          t.ID,
			Name:        t.Name,
			DurationMS:  t.DurationMS,
			Popularity:  t.Popularity,
			PreviewURL:  t.PreviewURL,
			ExternalURL: t.ExternalURLs.Spotify,
			URI:         t.URI,
		}
		for _, a := range t.Artists {
			track.Artists = append(track.Artists, SpotifyArtist{ID: a.ID, Name: a.Name, URI: a.URI})
		}
		track.Album = SpotifyAlbum{ID: t.Album.ID, Name: t.Album.Name, ReleaseDate: t.Album.ReleaseDate, URI: t.Album.URI}
		searchResult.Tracks = append(searchResult.Tracks, track)
	}

	for _, a := range result.Artists.Items {
		searchResult.Artists = append(searchResult.Artists, SpotifyArtist{
			ID:          a.ID,
			Name:        a.Name,
			Genres:      a.Genres,
			Popularity:  a.Popularity,
			Followers:   a.Followers.Total,
			ExternalURL: a.ExternalURLs.Spotify,
			URI:         a.URI,
		})
	}

	for _, a := range result.Albums.Items {
		album := SpotifyAlbum{
			ID:          a.ID,
			Name:        a.Name,
			ReleaseDate: a.ReleaseDate,
			TotalTracks: a.TotalTracks,
			ExternalURL: a.ExternalURLs.Spotify,
			URI:         a.URI,
		}
		for _, ar := range a.Artists {
			album.Artists = append(album.Artists, SpotifyArtist{ID: ar.ID, Name: ar.Name, URI: ar.URI})
		}
		searchResult.Albums = append(searchResult.Albums, album)
	}

	return searchResult, nil
}

// GetTrack gets track details
func (c *SpotifyClient) GetTrack(ctx context.Context, trackID string) (*SpotifyTrack, error) {
	resp, err := c.doRequest(ctx, "GET", "/tracks/"+trackID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get track failed: %s", resp.Status)
	}

	var result struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		DurationMS int    `json:"duration_ms"`
		Popularity int    `json:"popularity"`
		PreviewURL string `json:"preview_url"`
		URI        string `json:"uri"`
		Artists    []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			URI  string `json:"uri"`
		} `json:"artists"`
		Album struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			ReleaseDate string `json:"release_date"`
			TotalTracks int    `json:"total_tracks"`
			URI         string `json:"uri"`
			Images      []struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"images"`
		} `json:"album"`
		ExternalURLs struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	track := &SpotifyTrack{
		ID:          result.ID,
		Name:        result.Name,
		DurationMS:  result.DurationMS,
		Popularity:  result.Popularity,
		PreviewURL:  result.PreviewURL,
		ExternalURL: result.ExternalURLs.Spotify,
		URI:         result.URI,
	}

	for _, a := range result.Artists {
		track.Artists = append(track.Artists, SpotifyArtist{ID: a.ID, Name: a.Name, URI: a.URI})
	}

	track.Album = SpotifyAlbum{
		ID:          result.Album.ID,
		Name:        result.Album.Name,
		ReleaseDate: result.Album.ReleaseDate,
		TotalTracks: result.Album.TotalTracks,
		URI:         result.Album.URI,
	}
	for _, img := range result.Album.Images {
		track.Album.Images = append(track.Album.Images, SpotifyImage{URL: img.URL, Width: img.Width, Height: img.Height})
	}

	return track, nil
}

// GetAudioFeatures gets audio features for a track
func (c *SpotifyClient) GetAudioFeatures(ctx context.Context, trackID string) (*SpotifyAudioFeatures, error) {
	resp, err := c.doRequest(ctx, "GET", "/audio-features/"+trackID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get audio features failed: %s", resp.Status)
	}

	var features SpotifyAudioFeatures
	if err := json.NewDecoder(resp.Body).Decode(&features); err != nil {
		return nil, err
	}

	return &features, nil
}

// GetArtist gets artist details
func (c *SpotifyClient) GetArtist(ctx context.Context, artistID string) (*SpotifyArtist, error) {
	resp, err := c.doRequest(ctx, "GET", "/artists/"+artistID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get artist failed: %s", resp.Status)
	}

	var result struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		Genres     []string `json:"genres"`
		Popularity int      `json:"popularity"`
		URI        string   `json:"uri"`
		Followers  struct {
			Total int `json:"total"`
		} `json:"followers"`
		ExternalURLs struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &SpotifyArtist{
		ID:          result.ID,
		Name:        result.Name,
		Genres:      result.Genres,
		Popularity:  result.Popularity,
		Followers:   result.Followers.Total,
		ExternalURL: result.ExternalURLs.Spotify,
		URI:         result.URI,
	}, nil
}

// GetArtistTopTracks gets an artist's top tracks
func (c *SpotifyClient) GetArtistTopTracks(ctx context.Context, artistID string) ([]SpotifyTrack, error) {
	resp, err := c.doRequest(ctx, "GET", "/artists/"+artistID+"/top-tracks?market=US")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get top tracks failed: %s", resp.Status)
	}

	var result struct {
		Tracks []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			DurationMS int    `json:"duration_ms"`
			Popularity int    `json:"popularity"`
			PreviewURL string `json:"preview_url"`
			URI        string `json:"uri"`
			Artists    []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				URI  string `json:"uri"`
			} `json:"artists"`
			Album struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				URI  string `json:"uri"`
			} `json:"album"`
			ExternalURLs struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var tracks []SpotifyTrack
	for _, t := range result.Tracks {
		track := SpotifyTrack{
			ID:          t.ID,
			Name:        t.Name,
			DurationMS:  t.DurationMS,
			Popularity:  t.Popularity,
			PreviewURL:  t.PreviewURL,
			ExternalURL: t.ExternalURLs.Spotify,
			URI:         t.URI,
		}
		for _, a := range t.Artists {
			track.Artists = append(track.Artists, SpotifyArtist{ID: a.ID, Name: a.Name, URI: a.URI})
		}
		track.Album = SpotifyAlbum{ID: t.Album.ID, Name: t.Album.Name, URI: t.Album.URI}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// GetRelatedArtists gets related artists
func (c *SpotifyClient) GetRelatedArtists(ctx context.Context, artistID string) ([]SpotifyArtist, error) {
	resp, err := c.doRequest(ctx, "GET", "/artists/"+artistID+"/related-artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get related artists failed: %s", resp.Status)
	}

	var result struct {
		Artists []struct {
			ID         string   `json:"id"`
			Name       string   `json:"name"`
			Genres     []string `json:"genres"`
			Popularity int      `json:"popularity"`
			URI        string   `json:"uri"`
			Followers  struct {
				Total int `json:"total"`
			} `json:"followers"`
			ExternalURLs struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
		} `json:"artists"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var artists []SpotifyArtist
	for _, a := range result.Artists {
		artists = append(artists, SpotifyArtist{
			ID:          a.ID,
			Name:        a.Name,
			Genres:      a.Genres,
			Popularity:  a.Popularity,
			Followers:   a.Followers.Total,
			ExternalURL: a.ExternalURLs.Spotify,
			URI:         a.URI,
		})
	}

	return artists, nil
}

// GetAlbum gets album details
func (c *SpotifyClient) GetAlbum(ctx context.Context, albumID string) (*SpotifyAlbum, error) {
	resp, err := c.doRequest(ctx, "GET", "/albums/"+albumID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get album failed: %s", resp.Status)
	}

	var result struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		ReleaseDate string `json:"release_date"`
		TotalTracks int    `json:"total_tracks"`
		URI         string `json:"uri"`
		Artists     []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			URI  string `json:"uri"`
		} `json:"artists"`
		Images []struct {
			URL    string `json:"url"`
			Width  int    `json:"width"`
			Height int    `json:"height"`
		} `json:"images"`
		ExternalURLs struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	album := &SpotifyAlbum{
		ID:          result.ID,
		Name:        result.Name,
		ReleaseDate: result.ReleaseDate,
		TotalTracks: result.TotalTracks,
		ExternalURL: result.ExternalURLs.Spotify,
		URI:         result.URI,
	}

	for _, a := range result.Artists {
		album.Artists = append(album.Artists, SpotifyArtist{ID: a.ID, Name: a.Name, URI: a.URI})
	}

	for _, img := range result.Images {
		album.Images = append(album.Images, SpotifyImage{URL: img.URL, Width: img.Width, Height: img.Height})
	}

	return album, nil
}

// GetNewReleases gets new album releases
func (c *SpotifyClient) GetNewReleases(ctx context.Context, limit int) ([]SpotifyAlbum, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/browse/new-releases?limit=%d", limit))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get new releases failed: %s", resp.Status)
	}

	var result struct {
		Albums struct {
			Items []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				ReleaseDate string `json:"release_date"`
				TotalTracks int    `json:"total_tracks"`
				URI         string `json:"uri"`
				Artists     []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					URI  string `json:"uri"`
				} `json:"artists"`
				Images []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"images"`
				ExternalURLs struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
			} `json:"items"`
		} `json:"albums"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var albums []SpotifyAlbum
	for _, a := range result.Albums.Items {
		album := SpotifyAlbum{
			ID:          a.ID,
			Name:        a.Name,
			ReleaseDate: a.ReleaseDate,
			TotalTracks: a.TotalTracks,
			ExternalURL: a.ExternalURLs.Spotify,
			URI:         a.URI,
		}
		for _, ar := range a.Artists {
			album.Artists = append(album.Artists, SpotifyArtist{ID: ar.ID, Name: ar.Name, URI: ar.URI})
		}
		for _, img := range a.Images {
			album.Images = append(album.Images, SpotifyImage{URL: img.URL, Width: img.Width, Height: img.Height})
		}
		albums = append(albums, album)
	}

	return albums, nil
}

// GetRecommendations gets track recommendations based on seed tracks/artists
func (c *SpotifyClient) GetRecommendations(ctx context.Context, seedTracks, seedArtists []string, limit int) ([]SpotifyTrack, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))

	if len(seedTracks) > 0 {
		params.Set("seed_tracks", strings.Join(seedTracks, ","))
	}
	if len(seedArtists) > 0 {
		params.Set("seed_artists", strings.Join(seedArtists, ","))
	}

	if len(seedTracks) == 0 && len(seedArtists) == 0 {
		return nil, fmt.Errorf("at least one seed track or artist is required")
	}

	resp, err := c.doRequest(ctx, "GET", "/recommendations?"+params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get recommendations failed: %s", resp.Status)
	}

	var result struct {
		Tracks []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			DurationMS int    `json:"duration_ms"`
			Popularity int    `json:"popularity"`
			PreviewURL string `json:"preview_url"`
			URI        string `json:"uri"`
			Artists    []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				URI  string `json:"uri"`
			} `json:"artists"`
			Album struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				URI  string `json:"uri"`
			} `json:"album"`
			ExternalURLs struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var tracks []SpotifyTrack
	for _, t := range result.Tracks {
		track := SpotifyTrack{
			ID:          t.ID,
			Name:        t.Name,
			DurationMS:  t.DurationMS,
			Popularity:  t.Popularity,
			PreviewURL:  t.PreviewURL,
			ExternalURL: t.ExternalURLs.Spotify,
			URI:         t.URI,
		}
		for _, a := range t.Artists {
			track.Artists = append(track.Artists, SpotifyArtist{ID: a.ID, Name: a.Name, URI: a.URI})
		}
		track.Album = SpotifyAlbum{ID: t.Album.ID, Name: t.Album.Name, URI: t.Album.URI}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// GetGenres gets available genre seeds for recommendations (cached 5 min)
func (c *SpotifyClient) GetGenres(ctx context.Context) ([]string, error) {
	return spotifyGenresCache.GetOrFetch(ctx, func(ctx context.Context) ([]string, error) {
		resp, err := c.doRequest(ctx, "GET", "/recommendations/available-genre-seeds")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("get genres failed: %s", resp.Status)
		}

		var result struct {
			Genres []string `json:"genres"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		return result.Genres, nil
	})
}

// GetHealth returns health status
func (c *SpotifyClient) GetHealth(ctx context.Context) (*SpotifyHealth, error) {
	health := &SpotifyHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check credentials
	if c.clientID == "" || c.clientSecret == "" {
		health.Score -= 50
		health.HasCredentials = false
		health.Issues = append(health.Issues, "Missing Spotify API credentials")
		health.Recommendations = append(health.Recommendations, "Set SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables")
	} else {
		health.HasCredentials = true
	}

	// Test connection
	status, _ := c.GetStatus(ctx)
	health.Connected = status != nil && status.Connected

	if !health.Connected && health.HasCredentials {
		health.Score -= 30
		health.Issues = append(health.Issues, "Cannot connect to Spotify API")
		health.Recommendations = append(health.Recommendations, "Check API credentials are valid")
	}

	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}
