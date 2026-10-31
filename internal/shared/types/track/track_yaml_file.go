package track

import (
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types"
	"gopkg.in/yaml.v3"
)

// TrackYamlFile contains abstract track metadata, streaming identifiers,
// and optional local media-file metadata and fingerprints.
// keep me in sync with moongas-mediascan-python/src/mediascan/types/track/track_yaml_file.py
type TrackYamlFile struct {
	TrackData     TrackData            `yaml:"trackData" json:"trackData"`
	TrackIDData   *TrackIDData         `yaml:"trackIdData,omitempty" json:"trackIdData,omitempty"`
	MediaFileData []TrackMediaFileData `yaml:"mediaFileData,omitempty" json:"mediaFileData,omitempty"`
}

func (t *TrackYamlFile) UnmarshalJSON(data []byte) error {
	type trackYamlFile TrackYamlFile
	var decoded trackYamlFile
	if err := types.DecodeRequiredJSON(data, &decoded, "trackData"); err != nil {
		return err
	}
	*t = TrackYamlFile(decoded)
	return nil
}

func (t *TrackYamlFile) UnmarshalYAML(node *yaml.Node) error {
	type trackYamlFile TrackYamlFile
	var decoded trackYamlFile
	if err := types.DecodeRequiredYAML(node, &decoded, "trackData"); err != nil {
		return err
	}
	*t = TrackYamlFile(decoded)
	return nil
}
