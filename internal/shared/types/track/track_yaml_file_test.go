package track

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTrackYamlFileRoundTrip(t *testing.T) {
	input := []byte(`{
		"trackData": {
			"title": "Song", "artist": "Artist", "albumartist": "Artist",
			"album": "Album", "genre": "Rock", "year": 2026
		},
		"trackIdData": {
			"streamingData": [{
				"streamingService": "spotify", "trackUrl": "https://open.spotify.com/track/example",
				"trackId": "example", "apiData": {"tracks": {"items": []}}
			}]
		},
		"mediaFileData": [{
			"fpcalcData": {"duration": 263.62, "fingerprint": "example"},
			"acoustidData": {
				"results": [{
					"id": "result", "score": 0.98905593,
					"recordings": [{
						"id": "recording", "title": "Bingo Bango (radio mix)",
						"length": 225493, "video": false, "disambiguation": "",
						"firstReleaseDate": "2000-04-03",
						"isrcs": ["GBBKS0000003", "GBBKS0000042"]
					}, {"id": "minimal"}]
				}],
				"status": "ok"
			}
		}, {
			"path": "song.mp3", "size": 0, "format": "mp3", "title": "",
			"artist": "Artist", "albumartist": "Artist", "album": "Album",
			"genre": "Rock", "year": 0, "duration": 0
		}]
	}`)
	for _, codec := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"JSON", json.Marshal, json.Unmarshal},
		{"YAML", yaml.Marshal, yaml.Unmarshal},
	} {
		t.Run(codec.name, func(t *testing.T) {
			var data TrackYamlFile
			if err := codec.unmarshal(input, &data); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(data.MediaFileData) != 2 || data.TrackIDData == nil ||
				len(data.TrackIDData.StreamingData) != 1 {
				t.Fatalf("unexpected track data: %+v", data)
			}
			media := data.MediaFileData[0]
			if media.Path != nil || media.FpCalcData == nil ||
				media.FpCalcData.Duration == nil || *media.FpCalcData.Duration != 263.62 ||
				media.AcoustIDData == nil || len(media.AcoustIDData.Results) != 1 {
				t.Fatalf("unexpected partial media-file entry: %+v", media)
			}
			recordings := media.AcoustIDData.Results[0].Recordings
			if len(recordings) != 2 {
				t.Fatalf("unexpected recordings: %+v", recordings)
			}
			full, minimal := recordings[0], recordings[1]
			if full.Video == nil || *full.Video || full.Length == nil || *full.Length != 225493 ||
				full.Title == nil || *full.Title != "Bingo Bango (radio mix)" ||
				full.Disambiguation == nil || *full.Disambiguation != "" ||
				full.FirstReleaseDate == nil || *full.FirstReleaseDate != "2000-04-03" {
				t.Errorf("unexpected recording: %+v", full)
			}
			if minimal.Title != nil || minimal.Length != nil || minimal.Video != nil ||
				minimal.Disambiguation != nil || minimal.FirstReleaseDate != nil || minimal.ISRCs != nil {
				t.Errorf("expected absent optional fields: %+v", minimal)
			}
			out, err := codec.marshal(data)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var want, got map[string]any
			if err := codec.unmarshal(input, &want); err != nil {
				t.Fatal(err)
			}
			if err := codec.unmarshal(out, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("track data changed:\ngot: %s\nwant: %s", out, input)
			}
		})
	}
}

func TestTrackYamlFileOmitsOptionalSections(t *testing.T) {
	for _, input := range []string{
		`{"trackData":{"title":"","artist":"","albumartist":"","album":"","genre":"","year":0}}`,
		`{"trackData":{"title":"","artist":"","albumartist":"","album":"","genre":"","year":0},"trackIdData":null,"mediaFileData":null}`,
		`{"trackData":{"title":"","artist":"","albumartist":"","album":"","genre":"","year":0},"mediaFileData":[]}`,
	} {
		var data TrackYamlFile
		if err := yaml.Unmarshal([]byte(input), &data); err != nil {
			t.Fatal(err)
		}

		for _, marshal := range []func(any) ([]byte, error){json.Marshal, yaml.Marshal} {
			out, err := marshal(data)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]any
			if err := yaml.Unmarshal(out, &fields); err != nil {
				t.Fatal(err)
			}
			if len(fields) != 1 || fields["trackData"] == nil {
				t.Errorf("expected only trackData, got %s", out)
			}
		}
	}
}

func TestTrackYamlFileTrackIDDataObject(t *testing.T) {
	const core = `"trackData":{"title":"","artist":"","albumartist":"","album":"","genre":"","year":0}`
	for _, codec := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"JSON", json.Marshal, json.Unmarshal},
		{"YAML", yaml.Marshal, yaml.Unmarshal},
	} {
		t.Run(codec.name, func(t *testing.T) {
			for _, value := range []string{`{}`, `{"streamingData":[]}`} {
				var data TrackYamlFile
				if err := codec.unmarshal([]byte(`{`+core+`,"trackIdData":`+value+`}`), &data); err != nil {
					t.Fatal(err)
				}
				if data.TrackIDData == nil {
					t.Fatal("expected an empty object to remain present")
				}
				out, err := codec.marshal(data)
				if err != nil {
					t.Fatal(err)
				}
				var fields map[string]any
				if err := codec.unmarshal(out, &fields); err != nil {
					t.Fatal(err)
				}
				if _, ok := fields["trackIdData"].(map[string]any); !ok {
					t.Fatalf("expected trackIdData object, got %s", out)
				}
			}
			for _, value := range []string{`[]`, `[{}]`} {
				var data TrackYamlFile
				if err := codec.unmarshal([]byte(`{`+core+`,"trackIdData":`+value+`}`), &data); err == nil {
					t.Fatalf("expected legacy list %s to be rejected", value)
				}
			}
		})
	}
}
