package streaming

// StreamingSpotifyData contains Spotify track identifiers and its search response.
type StreamingSpotifyData struct {
	StreamingData[SpotifyAPIData] `yaml:",inline"`
}

// SpotifyAPIData represents a Spotify track search response.
type SpotifyAPIData struct {
	Tracks SpotifyTrackPage `yaml:"tracks" json:"tracks"`
}

// SpotifyTrackPage contains search pagination and the matching tracks.
type SpotifyTrackPage struct {
	Href     string         `yaml:"href" json:"href"`
	Limit    int            `yaml:"limit" json:"limit"`
	Next     *string        `yaml:"next" json:"next"`
	Offset   int            `yaml:"offset" json:"offset"`
	Previous *string        `yaml:"previous" json:"previous"`
	Total    int            `yaml:"total" json:"total"`
	Items    []SpotifyTrack `yaml:"items" json:"items"`
}

// SpotifyTrack represents a track returned by Spotify.
type SpotifyTrack struct {
	Album        SpotifyAlbum        `yaml:"album" json:"album"`
	Artists      []SpotifyArtist     `yaml:"artists" json:"artists"`
	DiscNumber   int                 `yaml:"disc_number" json:"disc_number"`
	DurationMS   int                 `yaml:"duration_ms" json:"duration_ms"`
	Explicit     bool                `yaml:"explicit" json:"explicit"`
	ExternalIDs  SpotifyExternalIDs  `yaml:"external_ids" json:"external_ids"`
	ExternalURLs SpotifyExternalURLs `yaml:"external_urls" json:"external_urls"`
	Href         string              `yaml:"href" json:"href"`
	ID           string              `yaml:"id" json:"id"`
	IsLocal      bool                `yaml:"is_local" json:"is_local"`
	IsPlayable   bool                `yaml:"is_playable" json:"is_playable"`
	Name         string              `yaml:"name" json:"name"`
	TrackNumber  int                 `yaml:"track_number" json:"track_number"`
	Type         string              `yaml:"type" json:"type"`
	URI          string              `yaml:"uri" json:"uri"`
}

// SpotifyAlbum represents the album metadata included with a track.
type SpotifyAlbum struct {
	AlbumType            string              `yaml:"album_type" json:"album_type"`
	Artists              []SpotifyArtist     `yaml:"artists" json:"artists"`
	ExternalURLs         SpotifyExternalURLs `yaml:"external_urls" json:"external_urls"`
	Href                 string              `yaml:"href" json:"href"`
	ID                   string              `yaml:"id" json:"id"`
	Images               []SpotifyImage      `yaml:"images" json:"images"`
	IsPlayable           bool                `yaml:"is_playable" json:"is_playable"`
	Name                 string              `yaml:"name" json:"name"`
	ReleaseDate          string              `yaml:"release_date" json:"release_date"`
	ReleaseDatePrecision string              `yaml:"release_date_precision" json:"release_date_precision"`
	TotalTracks          int                 `yaml:"total_tracks" json:"total_tracks"`
	Type                 string              `yaml:"type" json:"type"`
	URI                  string              `yaml:"uri" json:"uri"`
}

// SpotifyArtist represents an artist credited on an album or track.
type SpotifyArtist struct {
	ExternalURLs SpotifyExternalURLs `yaml:"external_urls" json:"external_urls"`
	Href         string              `yaml:"href" json:"href"`
	ID           string              `yaml:"id" json:"id"`
	Name         string              `yaml:"name" json:"name"`
	Type         string              `yaml:"type" json:"type"`
	URI          string              `yaml:"uri" json:"uri"`
}

// SpotifyImage represents album artwork; Spotify may return null dimensions.
type SpotifyImage struct {
	Height *int   `yaml:"height" json:"height"`
	Width  *int   `yaml:"width" json:"width"`
	URL    string `yaml:"url" json:"url"`
}

// SpotifyExternalURLs contains the Spotify web link for an API object.
type SpotifyExternalURLs struct {
	Spotify string `yaml:"spotify" json:"spotify"`
}

// SpotifyExternalIDs contains a track's industry identifier.
type SpotifyExternalIDs struct {
	ISRC string `yaml:"isrc" json:"isrc"`
}
