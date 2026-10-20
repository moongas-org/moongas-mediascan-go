package shared

// AlbumData contains the contents of the album_data block of an album.yml file.
type AlbumData struct {
	Name          string `yaml:"name"`
	DateReleased  Date   `yaml:"dateReleased"`
	ExplicitCover bool   `yaml:"explicitCover"`
}
