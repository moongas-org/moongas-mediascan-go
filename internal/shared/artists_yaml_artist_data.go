// Package shared contains types shared across the mediascan application.
package shared

// ArtistsYamlArtistData (DEPRECATED) includes the directory path and the data read
// from the artist.yml file in that directory.
// keep me in sync with moongas-mediascan-python/src/mediascan/artists_yaml_artist_data.py
type ArtistsYamlArtistData struct {
	ArtistData ArtistData `yaml:"artistData"`
	Path       string     `yaml:"path"`
}
