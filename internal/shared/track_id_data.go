package shared

// TrackIDData contains identifiers and URLs for a track across various streaming services.
// keep me in sync with moongas-mediascan-python/src/track_id_data.py
type TrackIDData struct {
	// All fields are optional; optional fields use pointers and 'omitempty'
	// for clean re-marshaling

	// Industry standard key for cross-platform matching
	ISRC *string `yaml:"isrc,omitempty"`

	// Spotify
	SpotifyID  *string `yaml:"spotifyId,omitempty"`
	SpotifyURI *string `yaml:"spotifyUri,omitempty"`
	SpotifyURL *string `yaml:"spotifyUrl,omitempty"`

	// Apple Music
	AppleMusicID  *string `yaml:"appleMusicId,omitempty"`
	AppleMusicURL *string `yaml:"appleMusicUrl,omitempty"`

	// YouTube & YouTube Music
	YouTubeMusicID  *string `yaml:"youtubeMusicId,omitempty"` // Explicitly for ://youtube.com
	YouTubeMusicURL *string `yaml:"youtubeMusicUrl,omitempty"`
	YouTubeVideoID  *string `yaml:"youtubeVideoId,omitempty"` // Standard video streaming ID
	YouTubeVideoURL *string `yaml:"youtubeVideoUrl,omitempty"`

	// Amazon Music
	AmazonMusicID  *string `yaml:"amazonMusicId,omitempty"`
	AmazonMusicURL *string `yaml:"amazonMusicUrl,omitempty"`

	// Tidal
	TidalID  *string `yaml:"tidalId,omitempty"`
	TidalURL *string `yaml:"tidalUrl,omitempty"`

	// Deezer
	DeezerID  *string `yaml:"deezerId,omitempty"`
	DeezerURL *string `yaml:"deezerUrl,omitempty"`
}
