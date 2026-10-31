// Package streaming defines streaming service track data.
package streaming

import (
	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types"
	"gopkg.in/yaml.v3"
)

// StreamingData contains a streaming service's track identifiers and API response.
// T defines the service-specific response structure.
type StreamingData[T any] struct {
	StreamingService StreamingService `yaml:"streamingService" json:"streamingService"`
	// Industry standard key for cross-platform matching
	ISRC     *string `yaml:"isrc,omitempty" json:"isrc,omitempty"`
	TrackURL string  `yaml:"trackUrl" json:"trackUrl"`
	TrackID  string  `yaml:"trackId" json:"trackId"`
	APIData  T       `yaml:"apiData" json:"apiData"`
}

func (s *StreamingData[T]) UnmarshalJSON(data []byte) error {
	type streamingData StreamingData[T]
	var decoded streamingData
	if err := types.DecodeRequiredJSON(data, &decoded, "streamingService", "trackUrl", "trackId", "apiData"); err != nil {
		return err
	}
	*s = StreamingData[T](decoded)
	return nil
}

func (s *StreamingData[T]) UnmarshalYAML(node *yaml.Node) error {
	type streamingData StreamingData[T]
	var decoded streamingData
	if err := types.DecodeRequiredYAML(node, &decoded, "streamingService", "trackUrl", "trackId", "apiData"); err != nil {
		return err
	}
	*s = StreamingData[T](decoded)
	return nil
}
