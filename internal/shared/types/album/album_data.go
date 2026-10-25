// Package album defines album metadata types and YAML representations.
package album

import "github.com/moongas-org/moongas-mediascan-go/internal/shared/types"

// AlbumData contains the contents of the album_data block of an album.yml file.
type AlbumData struct {
	Name          string     `yaml:"name"`
	DateReleased  types.Date `yaml:"dateReleased"`
	ExplicitCover bool       `yaml:"explicitCover"`
}
