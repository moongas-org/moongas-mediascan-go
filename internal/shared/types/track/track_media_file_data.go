package track

// TrackMediaFileData represents optional local file metadata in track YAML.
// Unlike the scanner's MediaFileData, it permits fingerprint-only entries.
type TrackMediaFileData struct {
	Path         *string            `yaml:"path,omitempty" json:"path,omitempty"`
	Size         *int64             `yaml:"size,omitempty" json:"size,omitempty"`
	Format       *string            `yaml:"format,omitempty" json:"format,omitempty"`
	Title        *string            `yaml:"title,omitempty" json:"title,omitempty"`
	Artist       *string            `yaml:"artist,omitempty" json:"artist,omitempty"`
	AlbumArtist  *string            `yaml:"albumartist,omitempty" json:"albumartist,omitempty"`
	Album        *string            `yaml:"album,omitempty" json:"album,omitempty"`
	Genre        *string            `yaml:"genre,omitempty" json:"genre,omitempty"`
	Year         *int               `yaml:"year,omitempty" json:"year,omitempty"`
	Duration     *int               `yaml:"duration,omitempty" json:"duration,omitempty"`
	FpCalcData   *TrackFpCalcData   `yaml:"fpcalcData,omitempty" json:"fpcalcData,omitempty"`
	AcoustIDData *TrackAcoustIDData `yaml:"acoustidData,omitempty" json:"acoustidData,omitempty"`
}
