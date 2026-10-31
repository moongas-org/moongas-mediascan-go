// Package media_file defines metadata types for individual media files.
// It includes scanned metadata, fingerprints, and AcoustID results.
package media_file

import (
	"time"
)

// MediaFileData permits partial entries containing only fingerprints or API results.
type MediaFileData struct {
	Path         *string                `yaml:"path,omitempty" json:"path,omitempty"`
	AlbumPath    string                 `yaml:"albumPath,omitempty" json:"albumPath,omitempty"`
	ArtistPath   string                 `yaml:"artistPath,omitempty" json:"artistPath,omitempty"`
	Size         *int64                 `yaml:"size,omitempty" json:"size,omitempty"`
	Format       *string                `yaml:"format,omitempty" json:"format,omitempty"`
	Title        *string                `yaml:"title,omitempty" json:"title,omitempty"`
	Artist       *string                `yaml:"artist,omitempty" json:"artist,omitempty"`
	AlbumArtist  *string                `yaml:"albumartist,omitempty" json:"albumartist,omitempty"`
	Album        *string                `yaml:"album,omitempty" json:"album,omitempty"`
	Genre        *string                `yaml:"genre,omitempty" json:"genre,omitempty"`
	Year         *int                   `yaml:"year,omitempty" json:"year,omitempty"`
	Duration     *float64               `yaml:"duration,omitempty" json:"duration,omitempty"`
	ModTime      *time.Time             `yaml:"modTime,omitempty" json:"modTime,omitempty"`
	FpCalcData   *MediaFileFpCalcData   `yaml:"fpcalcData,omitempty" json:"fpcalcData,omitempty"`
	AcoustIDData *MediaFileAcoustIDData `yaml:"acoustidData,omitempty" json:"acoustidData,omitempty"`
}
