// Package artists defines metadata types for collections of artists.
package artists

import "github.com/moongas-org/moongas-mediascan-go/internal/shared/types/artist"

// ArtistsYamlArtistData (DEPRECATED) includes the directory path and the data read
// from the artist.yml file in that directory.
// keep me in sync with moongas-mediascan-python/src/mediascan/artists_yaml_artist_data.py
type ArtistsYamlArtistData struct {
	ArtistData artist.ArtistData `yaml:"artistData"`
	Path       string            `yaml:"path"`
}
