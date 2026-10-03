package shared

type MediascanConf struct {
	MediaDirs          []string `yaml:"mediaDirs"`
	MediaExts          []string `yaml:"mediaExts"`
	ExcludePaths       []string `yaml:"excludePaths"`
	ExcludeTitle       []string `yaml:"excludeTitle"`
	ExcludeArtist      []string `yaml:"excludeArtist"`
	ExcludeAlbumArtist []string `yaml:"excludeAlbumArtist"`
	ExcludeAlbum       []string `yaml:"excludeAlbum"`
	ExcludeGenre       []string `yaml:"excludeGenre"`
	SortBy             string   `yaml:"sortBy"`
	GetMp3Duration     bool     `yaml:"getMp3Duration"`
}
