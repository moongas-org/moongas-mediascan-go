// Package track defines track metadata and identifiers.
package track

import (
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types"
	"gopkg.in/yaml.v3"
)

// TrackData contains core (mandatory) track metadata,
// It has a lot of overlap with MediaFileData, naturally, but excludes fields
// specific to the local media file itself, e.g. 'format'
// keep me in sync with moongas-mediascan-python/src/mediascan/track_data.py
type TrackData struct {
	Title       string `yaml:"title" json:"title"`
	Artist      string `yaml:"artist" json:"artist"`
	AlbumArtist string `yaml:"albumartist" json:"albumartist"`
	Album       string `yaml:"album" json:"album"`
	Genre       string `yaml:"genre" json:"genre"`
	Year        int    `yaml:"year" json:"year"`
}

func (t *TrackData) UnmarshalJSON(data []byte) error {
	type trackData TrackData
	var decoded trackData
	if err := types.DecodeRequiredJSON(data, &decoded, "title", "artist", "albumartist", "album", "genre", "year"); err != nil {
		return err
	}
	*t = TrackData(decoded)
	return nil
}

func (t *TrackData) UnmarshalYAML(node *yaml.Node) error {
	type trackData TrackData
	var decoded trackData
	if err := types.DecodeRequiredYAML(node, &decoded, "title", "artist", "albumartist", "album", "genre", "year"); err != nil {
		return err
	}
	*t = TrackData(decoded)
	return nil
}
