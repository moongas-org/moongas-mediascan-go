package shared

// TrackData contains core (mandatory) track metadata,
// It has a lot of overlap with MediaFileData, naturally, but excludes fields
// specific to the local media file itself, e.g. 'format'
// keep me in sync with moongas-mediascan-python/src/mediascan/track_data.py
type TrackData struct {
	Title       string `yaml:"title"`
	Artist      string `yaml:"artist"`
	AlbumArtist string `yaml:"albumartist"`
	Album       string `yaml:"album"`
	Genre       string `yaml:"genre"`
	Year        int    `yaml:"year"`
}
