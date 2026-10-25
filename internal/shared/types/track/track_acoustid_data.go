package track

// TrackAcoustIDData represents a response from the AcoustID API.
type TrackAcoustIDData struct {
	Results []TrackAcoustIDResult `yaml:"results"`
	Status  string                `yaml:"status"`
}

// TrackAcoustIDResult contains an AcoustID match and its recordings.
type TrackAcoustIDResult struct {
	ID         string                   `yaml:"id"`
	Recordings []TrackAcoustIDRecording `yaml:"recordings"`
	Score      float64                  `yaml:"score"`
}

// TrackAcoustIDRecording identifies a recording associated with an AcoustID match.
type TrackAcoustIDRecording struct {
	ID    string   `yaml:"id"`
	ISRCs []string `yaml:"isrcs"`
}
