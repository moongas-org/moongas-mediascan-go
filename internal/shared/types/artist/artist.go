package artist

// Artist contains an artist directory and any available artist.yml metadata.
type Artist struct {
	ArtistYamlFile ArtistYamlFile `yaml:"artistYamlFile"`
	Path           string         `yaml:"path"`
	HasArtistYaml  bool           `yaml:"hasArtistYaml"`
}
