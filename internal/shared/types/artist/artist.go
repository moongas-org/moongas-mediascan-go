package artist

// Artist contains artist.yml metadata and the directory containing the file.
type Artist struct {
	ArtistYamlFile ArtistYamlFile `yaml:"artistYamlFile"`
	Path           string         `yaml:"path"`
}
