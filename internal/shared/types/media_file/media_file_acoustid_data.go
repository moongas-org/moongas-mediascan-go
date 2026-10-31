package media_file

import (
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types"
	"gopkg.in/yaml.v3"
)

// MediaFileAcoustIDData represents a response from the AcoustID API.
type MediaFileAcoustIDData struct {
	Results []MediaFileAcoustIDResult `yaml:"results" json:"results"`
	Status  string                    `yaml:"status" json:"status"`
}

// MediaFileAcoustIDResult contains an AcoustID match and its recordings.
type MediaFileAcoustIDResult struct {
	ID         string                       `yaml:"id" json:"id"`
	Recordings []MediaFileAcoustIDRecording `yaml:"recordings" json:"recordings"`
	Score      float64                      `yaml:"score" json:"score"`
}

// MediaFileAcoustIDRecording identifies a recording associated with an AcoustID match.
type MediaFileAcoustIDRecording struct {
	ID               string    `yaml:"id" json:"id"`
	Title            *string   `yaml:"title,omitempty" json:"title,omitempty"`
	Length           *int      `yaml:"length,omitempty" json:"length,omitempty"`
	Video            *bool     `yaml:"video,omitempty" json:"video,omitempty"`
	Disambiguation   *string   `yaml:"disambiguation,omitempty" json:"disambiguation,omitempty"`
	FirstReleaseDate *string   `yaml:"firstReleaseDate,omitempty" json:"firstReleaseDate,omitempty"`
	ISRCs            *[]string `yaml:"isrcs,omitempty" json:"isrcs,omitempty"`
}

func (a *MediaFileAcoustIDData) UnmarshalJSON(data []byte) error {
	type acoustIDData MediaFileAcoustIDData
	var decoded acoustIDData
	if err := types.DecodeRequiredJSON(data, &decoded, "results", "status"); err != nil {
		return err
	}
	*a = MediaFileAcoustIDData(decoded)
	return nil
}

func (a *MediaFileAcoustIDData) UnmarshalYAML(node *yaml.Node) error {
	type acoustIDData MediaFileAcoustIDData
	var decoded acoustIDData
	if err := types.DecodeRequiredYAML(node, &decoded, "results", "status"); err != nil {
		return err
	}
	*a = MediaFileAcoustIDData(decoded)
	return nil
}

func (a *MediaFileAcoustIDResult) UnmarshalJSON(data []byte) error {
	type acoustIDResult MediaFileAcoustIDResult
	var decoded acoustIDResult
	if err := types.DecodeRequiredJSON(data, &decoded, "id", "recordings", "score"); err != nil {
		return err
	}
	*a = MediaFileAcoustIDResult(decoded)
	return nil
}

func (a *MediaFileAcoustIDResult) UnmarshalYAML(node *yaml.Node) error {
	type acoustIDResult MediaFileAcoustIDResult
	var decoded acoustIDResult
	if err := types.DecodeRequiredYAML(node, &decoded, "id", "recordings", "score"); err != nil {
		return err
	}
	*a = MediaFileAcoustIDResult(decoded)
	return nil
}

func (a *MediaFileAcoustIDRecording) UnmarshalJSON(data []byte) error {
	type acoustIDRecording MediaFileAcoustIDRecording
	var decoded acoustIDRecording
	if err := types.DecodeRequiredJSON(data, &decoded, "id"); err != nil {
		return err
	}
	*a = MediaFileAcoustIDRecording(decoded)
	return nil
}

func (a *MediaFileAcoustIDRecording) UnmarshalYAML(node *yaml.Node) error {
	type acoustIDRecording MediaFileAcoustIDRecording
	var decoded acoustIDRecording
	if err := types.DecodeRequiredYAML(node, &decoded, "id"); err != nil {
		return err
	}
	*a = MediaFileAcoustIDRecording(decoded)
	return nil
}
