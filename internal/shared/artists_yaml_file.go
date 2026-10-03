package shared

// ArtistsYamlFile is the data model for the artists.yml file output by mediascan cmd/scan-to-artists-yaml
// keep me in sync with moongas-mediascan-python/src/mediascan/artists_yaml_file.py
type ArtistsYamlFile struct {
	Artists []ArtistsYamlArtistData `yaml:"artists"`
}
