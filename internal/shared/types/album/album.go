package album

// Album contains an album directory and any available album.yml metadata.
type Album struct {
	AlbumYamlFile AlbumYamlFile `yaml:"albumYamlFile"`
	Path          string        `yaml:"path"`
	HasAlbumYaml  bool          `yaml:"hasAlbumYaml"`
}
