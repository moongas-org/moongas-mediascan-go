package track

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track/streaming"
	"gopkg.in/yaml.v3"
)

func TestRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		new    func() any
		input  string
		fields []string
	}{
		{
			"track", func() any { return &TrackData{} },
			`{"title":"","artist":"","albumartist":"","album":"","genre":"","year":0}`,
			[]string{"title", "artist", "albumartist", "album", "genre", "year"},
		},
		{
			"track file", func() any { return &TrackYamlFile{} },
			`{"trackData":{"title":"","artist":"","albumartist":"","album":"","genre":"","year":0}}`,
			[]string{"trackData"},
		},
		{
			"streaming", func() any { return &streaming.StreamingData[any]{} },
			`{"streamingService":"custom","trackUrl":"","trackId":"","apiData":null}`,
			[]string{"streamingService", "trackUrl", "trackId", "apiData"},
		},
		{
			"Spotify", func() any { return &streaming.StreamingSpotifyData{} },
			`{"streamingService":"spotify","trackUrl":"","trackId":"","apiData":{}}`,
			[]string{"streamingService", "trackUrl", "trackId", "apiData"},
		},
		{
			"AcoustID response", func() any { return &TrackAcoustIDData{} },
			`{"results":[],"status":""}`, []string{"results", "status"},
		},
		{
			"AcoustID result", func() any { return &TrackAcoustIDResult{} },
			`{"id":"","recordings":[],"score":0}`, []string{"id", "recordings", "score"},
		},
		{
			"AcoustID recording", func() any { return &TrackAcoustIDRecording{} },
			`{"id":""}`, []string{"id"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, codec := range []struct {
				name      string
				marshal   func(any) ([]byte, error)
				unmarshal func([]byte, any) error
			}{
				{"JSON", json.Marshal, json.Unmarshal},
				{"YAML", yaml.Marshal, yaml.Unmarshal},
			} {
				t.Run(codec.name, func(t *testing.T) {
					var fields map[string]json.RawMessage
					if err := json.Unmarshal([]byte(tt.input), &fields); err != nil {
						t.Fatal(err)
					}
					encode := func(fields map[string]json.RawMessage) []byte {
						t.Helper()
						data, err := json.Marshal(fields)
						if err != nil {
							t.Fatal(err)
						}
						// Produce actual YAML as well as JSON inputs.
						if codec.name == "YAML" {
							var object map[string]any
							if err := yaml.Unmarshal(data, &object); err != nil {
								t.Fatal(err)
							}
							data, err = codec.marshal(object)
							if err != nil {
								t.Fatal(err)
							}
						}
						return data
					}
					if err := codec.unmarshal(encode(fields), tt.new()); err != nil {
						t.Fatalf("valid zero values or omitted optional fields rejected: %v", err)
					}
					for _, field := range tt.fields {
						t.Run("missing "+field, func(t *testing.T) {
							value := fields[field]
							delete(fields, field)
							err := codec.unmarshal(encode(fields), tt.new())
							fields[field] = value
							if err == nil || !strings.Contains(err.Error(), field) {
								t.Fatalf("expected missing %s error, got %v", field, err)
							}
						})
					}
				})
			}
		})
	}
}

func TestRequiredFieldsInNestedData(t *testing.T) {
	for _, input := range []string{
		`{"trackData":{}}`,
		`{"streamingData":[{"trackUrl":"","trackId":"","apiData":null}]}`,
		`{"acoustidData":{"results":[{"id":"x","score":0}],"status":"ok"}}`,
		`{"acoustidData":{"results":[{"id":"x","score":0,"recordings":[{}]}],"status":"ok"}}`,
	} {
		for _, unmarshal := range []func([]byte, any) error{json.Unmarshal, yaml.Unmarshal} {
			var target any
			switch {
			case strings.Contains(input, "trackData"):
				target = &TrackYamlFile{}
			case strings.Contains(input, "streamingData"):
				target = &TrackIDData{}
			default:
				target = &TrackMediaFileData{}
			}
			if err := unmarshal([]byte(input), target); err == nil {
				t.Errorf("expected nested required-field error for %s", input)
			}
		}
	}
}
