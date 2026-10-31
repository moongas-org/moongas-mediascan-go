package track

// Track contains track YAML metadata and the path to the YAML file itself.
type Track struct {
	TrackYamlFile TrackYamlFile `yaml:"trackYamlFile" json:"trackYamlFile"`
	Path          string        `yaml:"path" json:"path"`
}
