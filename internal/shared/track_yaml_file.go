package shared

// TrackYamlFile contains abstract track metadata,
// It has a lot of overlap with MediaFileData, naturally, but includes additional fields
// such as track identifiers for various streaming services
// whereas MediaFileData contains info specific to a local media file,
// e.g. 'format', which are omitted from TrackYamlFile.
// keep me in sync with moongas-mediascan-python/src/mediascan/track_yaml_file.py
type TrackYamlFile struct {
	TrackData TrackData `yaml:"trackData"` // Mandatory core track metadata
	// Optional track identifiers for various streaming services
	TrackIDData []TrackIDData `yaml:"trackIdData,omitempty"`
}
