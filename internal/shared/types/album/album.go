package album

// Album contains album.yml metadata and the directory containing the file.
type Album struct {
	AlbumYamlFile AlbumYamlFile `yaml:"albumYamlFile"`
	Path          string        `yaml:"path"`
}
