package track

import "github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track/streaming"

// TrackIDData contains streaming service data for an abstract track.
// keep me in sync with moongas-mediascan-python/src/mediascan/types/track/track_id_data.py
type TrackIDData struct {
	StreamingData []streaming.StreamingData[any] `yaml:"streamingData,omitempty" json:"streamingData,omitempty"`
}
