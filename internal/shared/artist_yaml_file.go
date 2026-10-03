package shared

// ArtistYamlFile represents an artist.yml file.
// keep me in sync with moongas-mediascan-python/src/mediascan/artist_yaml_file.py
type ArtistYamlFile struct {
	ArtistData ArtistData `yaml:"artistData"`
}
