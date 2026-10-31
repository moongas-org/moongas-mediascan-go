package track

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTrackRoundTrip(t *testing.T) {
	want := Track{
		TrackYamlFile: TrackYamlFile{
			TrackData: TrackData{
				Title: "Song", Artist: "Artist", AlbumArtist: "Artist",
				Album: "Album", Genre: "Rock", Year: 2026,
			},
		},
		Path: "/music/Artist/Album/song.yml",
	}
	for _, codec := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"JSON", json.Marshal, json.Unmarshal},
		{"YAML", yaml.Marshal, yaml.Unmarshal},
	} {
		t.Run(codec.name, func(t *testing.T) {
			out, err := codec.marshal(want)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got Track
			if err := codec.unmarshal(out, &got); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("track = %+v, want %+v", got, want)
			}
		})
	}
}
