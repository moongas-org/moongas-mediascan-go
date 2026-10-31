// Package media_file defines metadata types for individual media files.
// It includes scanned metadata, fingerprints, and AcoustID results.
package media_file

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// MediaFileData permits partial entries containing only fingerprints or API results.
type MediaFileData struct {
	Path         *string                `yaml:"path,omitempty" json:"path,omitempty"`
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

func (m *MediaFileData) UnmarshalYAML(node *yaml.Node) error {
	type mediaFileData MediaFileData
	var decoded mediaFileData
	normalized := *node
	if node.Kind == yaml.MappingNode {
		normalized.Content = append([]*yaml.Node(nil), node.Content...)
		for i := 0; i+1 < len(node.Content); i += 2 {
			value := node.Content[i+1]
			if node.Content[i].Value != "modTime" || value.Tag == "!!null" {
				continue
			}
			var timestamp time.Time
			if err := value.Decode(&timestamp); err != nil {
				// Python datetime strings use a space instead of RFC3339's T.
				parsed, parseErr := time.Parse("2006-01-02 15:04:05Z07:00", value.Value)
				if parseErr != nil {
					return fmt.Errorf("decode mediaFileData.modTime: %w", err)
				}
				replacement := *value
				replacement.Tag = "!!timestamp"
				replacement.Value = parsed.Format(time.RFC3339Nano)
				normalized.Content[i+1] = &replacement
			}
		}
	}
	if err := normalized.Decode(&decoded); err != nil {
		return err
	}
	*m = MediaFileData(decoded)
	return nil
}
