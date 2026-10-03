package shared

import (
	"time"
)

type MediaFileData struct {
	Path        string // the path to the mediafile
	AlbumPath   string // the parent dir of the mediafile
	ArtistPath  string // i.e. two dir levels up from mediafile path
	Size        int64
	Format      string
	Title       string
	Artist      string
	AlbumArtist string
	Album       string
	Genre       string
	Year        int
	Duration    float64
	ModTime     time.Time
}
