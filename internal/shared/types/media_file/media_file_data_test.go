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
