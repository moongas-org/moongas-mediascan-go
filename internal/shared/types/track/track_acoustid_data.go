package track

import (
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types"
	"gopkg.in/yaml.v3"
)

// TrackAcoustIDData represents a response from the AcoustID API.
type TrackAcoustIDData struct {
	Results []TrackAcoustIDResult `yaml:"results" json:"results"`
	Status  string                `yaml:"status" json:"status"`
}

// TrackAcoustIDResult contains an AcoustID match and its recordings.
type TrackAcoustIDResult struct {
	ID         string                   `yaml:"id" json:"id"`
	Recordings []TrackAcoustIDRecording `yaml:"recordings" json:"recordings"`
	Score      float64                  `yaml:"score" json:"score"`
}

// TrackAcoustIDRecording identifies a recording associated with an AcoustID match.
type TrackAcoustIDRecording struct {
	ID               string    `yaml:"id" json:"id"`
	Title            *string   `yaml:"title,omitempty" json:"title,omitempty"`
	Length           *int      `yaml:"length,omitempty" json:"length,omitempty"`
	Video            *bool     `yaml:"video,omitempty" json:"video,omitempty"`
	Disambiguation   *string   `yaml:"disambiguation,omitempty" json:"disambiguation,omitempty"`
	FirstReleaseDate *string   `yaml:"firstReleaseDate,omitempty" json:"firstReleaseDate,omitempty"`
	ISRCs            *[]string `yaml:"isrcs,omitempty" json:"isrcs,omitempty"`
}

func (a *TrackAcoustIDData) UnmarshalJSON(data []byte) error {
	type acoustIDData TrackAcoustIDData
	var decoded acoustIDData
	if err := types.DecodeRequiredJSON(data, &decoded, "results", "status"); err != nil {
		return err
	}
	*a = TrackAcoustIDData(decoded)
	return nil
}

func (a *TrackAcoustIDData) UnmarshalYAML(node *yaml.Node) error {
	type acoustIDData TrackAcoustIDData
	var decoded acoustIDData
	if err := types.DecodeRequiredYAML(node, &decoded, "results", "status"); err != nil {
		return err
	}
	*a = TrackAcoustIDData(decoded)
	return nil
}

func (a *TrackAcoustIDResult) UnmarshalJSON(data []byte) error {
	type acoustIDResult TrackAcoustIDResult
	var decoded acoustIDResult
	if err := types.DecodeRequiredJSON(data, &decoded, "id", "recordings", "score"); err != nil {
		return err
	}
	*a = TrackAcoustIDResult(decoded)
	return nil
}

func (a *TrackAcoustIDResult) UnmarshalYAML(node *yaml.Node) error {
	type acoustIDResult TrackAcoustIDResult
	var decoded acoustIDResult
	if err := types.DecodeRequiredYAML(node, &decoded, "id", "recordings", "score"); err != nil {
		return err
	}
	*a = TrackAcoustIDResult(decoded)
	return nil
}

func (a *TrackAcoustIDRecording) UnmarshalJSON(data []byte) error {
	type acoustIDRecording TrackAcoustIDRecording
	var decoded acoustIDRecording
	if err := types.DecodeRequiredJSON(data, &decoded, "id"); err != nil {
		return err
	}
	*a = TrackAcoustIDRecording(decoded)
	return nil
}

func (a *TrackAcoustIDRecording) UnmarshalYAML(node *yaml.Node) error {
	type acoustIDRecording TrackAcoustIDRecording
	var decoded acoustIDRecording
	if err := types.DecodeRequiredYAML(node, &decoded, "id"); err != nil {
		return err
	}
	*a = TrackAcoustIDRecording(decoded)
	return nil
}
