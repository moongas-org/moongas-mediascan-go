package media_file

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/utils"
	"gopkg.in/yaml.v3"
)

func TestMediaFileDataRoundTrip(t *testing.T) {
	for _, codec := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{"JSON", json.Marshal, json.Unmarshal},
		{"YAML", yaml.Marshal, yaml.Unmarshal},
	} {
		t.Run(codec.name, func(t *testing.T) {
			for _, want := range []MediaFileData{
				{},
				{
					FpCalcData: &MediaFileFpCalcData{
						Duration: utils.Pointer(263.62), Fingerprint: utils.Pointer("fingerprint"),
					},
					AcoustIDData: &MediaFileAcoustIDData{
						Results: []MediaFileAcoustIDResult{}, Status: "ok",
					},
				},
				{
					Path: utils.Pointer("song.mp3"),
					Size: utils.Pointer(int64(0)), Format: utils.Pointer(""), Title: utils.Pointer(""),
					Artist: utils.Pointer("Artist"), AlbumArtist: utils.Pointer("Album Artist"),
					Album: utils.Pointer("Album"), Genre: utils.Pointer("Rock"), Year: utils.Pointer(0),
					Duration: utils.Pointer(12.75), ModTime: utils.Pointer(time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)),
				},
			} {
				out, err := codec.marshal(want)
				if err != nil {
					t.Fatal(err)
				}

				var got MediaFileData
				if err := codec.unmarshal(out, &got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("round trip changed media data: got %+v, want %+v", got, want)
				}
				var fields map[string]any
				if err := codec.unmarshal(out, &fields); err != nil {
					t.Fatal(err)
				}
				if want.Path == nil {
					for _, field := range []string{"path", "size", "duration", "modTime"} {
						if _, exists := fields[field]; exists {
							t.Errorf("absent field %s should be omitted in %s", field, out)
						}
					}
				} else {
					for _, field := range []string{"path", "size", "format", "title", "year", "duration", "modTime"} {
						if _, exists := fields[field]; !exists {
							t.Errorf("explicit field %s should be preserved in %s", field, out)
						}
					}
				}
			}
		})
	}
}

func TestMediaFileDataPythonModTime(t *testing.T) {
	for _, input := range []string{
		"2024-09-16 00:34:15.638511+00:00",
		"'2024-09-16 00:34:15.638511+00:00'",
		"2024-09-16 02:34:15.638511+02:00",
		"2024-09-16T00:34:15.638511Z",
	} {
		t.Run(input, func(t *testing.T) {
			var got MediaFileData
			if err := yaml.Unmarshal([]byte("modTime: "+input+"\n"), &got); err != nil {
				t.Fatal(err)
			}
			want := time.Date(2024, 9, 16, 0, 34, 15, 638511000, time.UTC)
			if got.ModTime == nil || !got.ModTime.Equal(want) {
				t.Fatalf("modTime = %v, want %v", got.ModTime, want)
			}
		})
	}
}

func TestMediaFileDataModTimeNullAndInvalid(t *testing.T) {
	var got MediaFileData
	if err := yaml.Unmarshal([]byte("modTime: null\n"), &got); err != nil {
		t.Fatal(err)
	}
	if got.ModTime != nil {
		t.Fatal("null modTime must remain nil")
	}
	for _, input := range []string{"not-a-date", "2024-99-16 00:34:15+00:00"} {
		if err := yaml.Unmarshal([]byte("modTime: "+input+"\n"), &got); err == nil {
			t.Fatalf("expected error for invalid modTime %q", input)
		}
	}
}
