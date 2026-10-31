package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moongas-org/moongas-mediascan-go/internal/shared/types/track"
	"gopkg.in/yaml.v3"
)

func TestRunFpcalcOutputParsing(t *testing.T) {
	var output fpcalcOutput
	if err := json.Unmarshal([]byte(`{"duration":12.5,"fingerprint":"abc"}`), &output); err != nil {
		t.Fatal(err)
	}
	if output.Duration == nil || *output.Duration != 12.5 || output.Fingerprint == nil || *output.Fingerprint != "abc" {
		t.Fatalf("unexpected fpcalc output: %#v", output)
	}
}

func TestNeedsFpcalcData(t *testing.T) {
	for _, tt := range []struct {
		name string
		yaml string
		want bool
	}{
		{
			name: "missing media-file metadata",
			yaml: "trackData:\n  title: Song\n",
			want: true,
		},
		{
			name: "missing fpcalc data",
			yaml: "trackData:\n  title: Song\nmediaFileData:\n  - path: song.mp3\n",
			want: true,
		},
		{
			name: "missing fingerprint",
			yaml: "mediaFileData:\n  - fpcalcData:\n      duration: 12.5\n",
			want: true,
		},
		{
			name: "complete fpcalc data",
			yaml: "mediaFileData:\n  - fpcalcData:\n      duration: 12.5\n      fingerprint: abc\n",
			want: false,
		},
		{
			name: "null media-file metadata",
			yaml: "mediaFileData: null\n",
			want: true,
		},
		{
			name: "empty media-file metadata",
			yaml: "mediaFileData: []\n",
			want: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := needsFpcalcData([]byte(tt.yaml))
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("needsFpcalcData = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestUpdateTrackYamlFilePreservesOtherDataAndBacksUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "song.yml")
	original := []byte(`trackData:
  title: Song
trackIdData:
  streamingData:
    - streamingService: spotify
      isrc: ABC
      trackId: track-id
      trackUrl: https://open.spotify.com/track/track-id
      apiData:
        tracks:
          items: []
mediaFileData:
  - path: song.mp3
    acoustidData:
      results: []
      status: ok
`)
	if err := os.WriteFile(path, original, 0o640); err != nil {
		t.Fatal(err)
	}

	duration := 12.5
	fingerprint := "new-fingerprint"
	updated, err := updateTrackYamlFile(path, original, track.TrackFpCalcData{
		Duration:    &duration,
		Fingerprint: &fingerprint,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("expected YAML to be updated")
	}

	backup, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != string(original) {
		t.Fatalf("backup differs from original: %q", backup)
	}
	updatedYaml, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updatedYaml), "streamingData:") ||
		!strings.Contains(string(updatedYaml), "trackId: track-id") ||
		!strings.Contains(string(updatedYaml), "apiData:") ||
		!strings.Contains(string(updatedYaml), "isrc: ABC") ||
		!strings.Contains(string(updatedYaml), "title: Song") {
		t.Fatalf("unrelated YAML fields were not preserved:\n%s", updatedYaml)
	}
	needsUpdate, err := needsFpcalcData(updatedYaml)
	if err != nil {
		t.Fatal(err)
	}
	if needsUpdate {
		t.Fatal("updated YAML is still missing fpcalc data")
	}
	var typed struct {
		MediaFileData []track.TrackMediaFileData `yaml:"mediaFileData"`
	}
	if err := yaml.Unmarshal(updatedYaml, &typed); err != nil {
		t.Fatal(err)
	}
	if len(typed.MediaFileData) != 1 || typed.MediaFileData[0].FpCalcData == nil ||
		typed.MediaFileData[0].AcoustIDData == nil || typed.MediaFileData[0].Path == nil ||
		*typed.MediaFileData[0].Path != "song.mp3" {
		t.Fatalf("expected fingerprint on existing media-file entry: %+v", typed.MediaFileData)
	}
}

func TestUpdateTrackYamlFileUpdatesExistingFpcalcData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "song.yml")
	original := []byte(`trackData:
  title: Song
mediaFileData:
  - path: song.mp3
    fpcalcData:
      duration: 1
      fingerprint: old
`)
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}

	duration := 2.25
	fingerprint := "replacement"
	updated, err := updateTrackYamlFile(path, original, track.TrackFpCalcData{
		Duration:    &duration,
		Fingerprint: &fingerprint,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated {
		t.Fatal("expected fpcalc data to be replaced")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]interface{}
	if err := yaml.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	mediaFiles := root["mediaFileData"].([]interface{})
	entry := mediaFiles[0].(map[string]interface{})
	fpcalcData := entry["fpcalcData"].(map[string]interface{})
	if fpcalcData["duration"] != 2.25 || fpcalcData["fingerprint"] != "replacement" || entry["path"] != "song.mp3" {
		t.Fatalf("unexpected updated YAML: %#v", root)
	}
}

func TestUpdateTrackYamlFileDoesNotWriteWhenUnchanged(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "song.yml")
	original := []byte("trackData:\n  title: Song\n")
	if err := os.WriteFile(path, original, 0o644); err != nil {
		t.Fatal(err)
	}

	duration := 12.5
	fingerprint := "abc"
	result := track.TrackFpCalcData{Duration: &duration, Fingerprint: &fingerprint}
	if _, err := updateTrackYamlFile(path, original, result); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected backup to be created: %v", err)
	}
	if changed, err := updateTrackYamlFile(path, updated, result); err != nil {
		t.Fatal(err)
	} else if changed {
		t.Fatal("expected identical fpcalc data not to modify YAML")
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("expected existing backup to remain: %v", err)
	}
}

func TestFpcalcRejectsInvalidMediaFileData(t *testing.T) {
	duration := 12.5
	fingerprint := "abc"
	for _, input := range []string{
		"mediaFileData: {}\n",
		"mediaFileData: text\n",
		"mediaFileData:\n  - text\n",
	} {
		t.Run(input, func(t *testing.T) {
			if _, err := needsFpcalcData([]byte(input)); err == nil {
				t.Fatal("expected schema validation error")
			}
			path := filepath.Join(t.TempDir(), "song.yml")
			if _, err := updateTrackYamlFile(path, []byte(input), track.TrackFpCalcData{
				Duration: &duration, Fingerprint: &fingerprint,
			}); err == nil {
				t.Fatal("expected schema validation error during update")
			}
			if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
				t.Fatalf("invalid input must not create backup: %v", err)
			}
		})
	}
}

func TestUpdateTrackYamlFileCreatesOptionalFpcalcData(t *testing.T) {
	duration := 12.5
	fingerprint := "abc"
	for _, input := range []string{
		"trackData:\n  title: Song\n",
		"mediaFileData: null\n",
		"mediaFileData: []\n",
		"mediaFileData:\n  - fpcalcData: null\n",
	} {
		t.Run(input, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "song.yml")
			if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
				t.Fatal(err)
			}
			if changed, err := updateTrackYamlFile(path, []byte(input), track.TrackFpCalcData{
				Duration: &duration, Fingerprint: &fingerprint,
			}); err != nil || !changed {
				t.Fatalf("expected update, got changed=%t, err=%v", changed, err)
			}
			out, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var data struct {
				MediaFileData []track.TrackMediaFileData `yaml:"mediaFileData"`
			}
			if err := yaml.Unmarshal(out, &data); err != nil {
				t.Fatal(err)
			}
			if len(data.MediaFileData) != 1 || data.MediaFileData[0].FpCalcData == nil ||
				data.MediaFileData[0].FpCalcData.Fingerprint == nil ||
				*data.MediaFileData[0].FpCalcData.Fingerprint != fingerprint {
				t.Fatalf("unexpected media-file data: %+v", data.MediaFileData)
			}
			if strings.Contains(string(out), "trackIdData:") {
				t.Fatalf("fingerprint must not create track identifiers:\n%s", out)
			}
		})
	}
}
